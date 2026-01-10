package compose

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"red-cloud/mod/gologger"
	"red-cloud/utils/sshutil"

	"github.com/hashicorp/terraform-exec/tfexec"
)

// RunComposeUp 编排入口
func RunComposeUp(opts ComposeOptions) error {
	// 1. 初始化 (调用 Core)
	ctx, err := NewComposeContext(opts)
	if err != nil {
		return err
	}
	if err := VerifyTemplates(ctx); err != nil {
		return err
	}

	// 2. 编排循环
	pendingCount := len(ctx.RuntimeSvcs)
	for pendingCount > 0 {
		deployedInThisLoop := 0

		// 使用排序后的 Keys 遍历
		for _, name := range ctx.SortedSvcKeys {
			svc := ctx.RuntimeSvcs[name]

			if svc.IsDeployed {
				continue
			}

			if canDeploy(svc, ctx.RuntimeSvcs) {
				gologger.Info().Msgf("🚀 开始部署服务: %s (Type: %s)", svc.Name, svc.Spec.Image)

				if err := processServiceUp(svc, ctx); err != nil {
					return fmt.Errorf("部署服务 [%s] 失败: %v", svc.Name, err)
				}

				svc.IsDeployed = true
				deployedInThisLoop++
				pendingCount--
			}
		}

		if deployedInThisLoop == 0 && pendingCount > 0 {
			return fmt.Errorf("编排死锁: 存在循环依赖，或依赖的服务被 Profile 过滤未启动")
		}
	}

	// 3. 执行 Setup
	if len(ctx.ConfigRaw.Setup) > 0 {
		gologger.Info().Msg("⚙️ 开始执行 Setup 后置任务...")
		if err := runSetupTasks(ctx.ConfigRaw.Setup, ctx.RuntimeSvcs, ctx.LogMgr); err != nil {
			return err
		}
	}

	return nil
}

// RunComposeDown 销毁入口
func RunComposeDown(opts ComposeOptions) error {
	ctx, err := NewComposeContext(opts)
	if err != nil {
		return err
	}

	// 状态回填
	pendingCount := 0
	for _, name := range ctx.SortedSvcKeys {
		svc := ctx.RuntimeSvcs[name]
		c, err := ctx.Project.GetCase(svc.Name)
		if err != nil {
			svc.IsDeployed = false
			continue
		}
		svc.CaseRef = c
		svc.IsDeployed = true
		pendingCount++

		if rawOut, err := c.TfOutput(); err == nil {
			svc.Outputs = parseTfOutput(rawOut)
		}
	}

	// 逆序销毁
	for pendingCount > 0 {
		destroyedInThisLoop := 0
		// 倒序遍历建议
		for i := len(ctx.SortedSvcKeys) - 1; i >= 0; i-- {
			svc := ctx.RuntimeSvcs[ctx.SortedSvcKeys[i]]

			if !svc.IsDeployed {
				continue
			}

			if canDestroy(svc, ctx.RuntimeSvcs) {
				gologger.Info().Msgf("🔥 正在销毁服务: %s", svc.Name)
				if err := svc.CaseRef.TfDestroy(); err != nil {
					gologger.Error().Msgf("销毁服务 [%s] 失败: %v", svc.Name, err)
				}

				svc.IsDeployed = false
				destroyedInThisLoop++
				pendingCount--
			}
		}

		if destroyedInThisLoop == 0 && pendingCount > 0 {
			return fmt.Errorf("销毁死锁: 存在循环依赖")
		}
	}
	return nil
}

// processServiceUp 单个服务部署逻辑
func processServiceUp(svc *RuntimeService, ctx *ComposeContext) error {
	tfVars := make(map[string]string)

	// Configs
	for _, cfgStr := range svc.Spec.Configs {
		parts := strings.SplitN(cfgStr, "=", 2)
		if len(parts) == 2 {
			tfName, cfgKey := parts[0], parts[1]
			if val, ok := ctx.GlobalConfigs[cfgKey]; ok {
				tfVars[tfName] = val
			} else {
				gologger.Error().Msgf("[%s] Config key '%s' not found", svc.Name, cfgKey)
			}
		}
	}

	// Environment
	for _, envStr := range svc.Spec.Environment {
		parts := strings.SplitN(envStr, "=", 2)
		if len(parts) == 2 {
			key, rawVal := parts[0], parts[1]
			vals, err := expandVariable(rawVal, ctx.RuntimeSvcs, svc)
			if err != nil {
				return fmt.Errorf("Environment parse error: %v", err)
			}
			tfVars[key] = strings.Join(vals, ",")
		}
	}

	// Provider Alias
	if pStr, ok := svc.Spec.Provider.(string); ok && pStr != "" && pStr != "default" {
		tfVars["provider_alias"] = pStr
	}

	// TF Apply
	p := ctx.Project
	c, err := p.GetCase(svc.Name)
	if err != nil {
		c, err = p.CaseCreate(svc.Spec.Image, p.User, svc.Name, tfVars)
		if err != nil {
			return fmt.Errorf("CaseCreate fail: %v", err)
		}
	}
	if err := c.TfApply(); err != nil {
		return fmt.Errorf("Terraform Apply fail: %v", err)
	}
	svc.CaseRef = c

	// Output Cache
	rawOut, err := c.TfOutput()
	if err == nil {
		svc.Outputs = parseTfOutput(rawOut)
	}

	// SSH Actions
	return runSSHActions(svc, ctx.LogMgr)
}

// --- 运行时辅助函数 ---

func runSSHActions(svc *RuntimeService, logMgr *gologger.LogManager) error {
	if svc.Spec.Command == "" && len(svc.Spec.Volumes) == 0 && len(svc.Spec.Downloads) == 0 {
		return nil
	}

	sshConf, err := svc.CaseRef.GetSSHConfig()
	if err != nil {
		gologger.Debug().Msgf("[%s] Skipping SSH actions: %v", svc.Name, err)
		return nil
	}

	client, err := sshutil.NewClient(sshConf)
	if err != nil {
		gologger.Error().Msgf("[%s] SSH Connect Fail: %v", svc.Name, err)
		return nil
	}
	defer client.Close()

	logger, _ := logMgr.NewServiceLogger(svc.Name)
	var writer io.Writer = os.Stdout
	if logger != nil {
		defer logger.Close()
		writer = logger
	}

	// Volumes
	for _, vol := range svc.Spec.Volumes {
		parts := strings.Split(vol, ":")
		if len(parts) == 2 {
			localPath, remotePath := parts[0], parts[1]
			gologger.Info().Msgf("[%s] Uploading %s -> %s", svc.Name, localPath, remotePath)
			if err := client.Upload(localPath, remotePath); err != nil {
				gologger.Error().Msgf("[%s] Upload failed: %v", svc.Name, err)
			}
		}
	}

	// Command
	if svc.Spec.Command != "" {
		gologger.Info().Msgf("[%s] Running init command...", svc.Name)
		if err := client.RunCommandWithLogger(svc.Spec.Command, writer); err != nil {
			gologger.Error().Msgf("[%s] Command failed: %v", svc.Name, err)
		}
	}

	// Downloads
	for _, dl := range svc.Spec.Downloads {
		parts := strings.Split(dl, ":")
		if len(parts) == 2 {
			remotePath, localPath := parts[0], parts[1]
			gologger.Info().Msgf("[%s] Downloading %s -> %s", svc.Name, remotePath, localPath)
			if err := client.Download(remotePath, localPath); err != nil {
				gologger.Error().Msgf("[%s] Download failed: %v", svc.Name, err)
			}
		}
	}
	return nil
}

func runSetupTasks(tasks []SetupTask, svcs map[string]*RuntimeService, logMgr *gologger.LogManager) error {
	for _, task := range tasks {
		targetSvc, ok := svcs[task.Service]
		if !ok {
			gologger.Debug().Msgf("Task [%s] skipped: Target %s not active", task.Name, task.Service)
			continue
		}

		cmds, err := expandVariable(task.Command, svcs, nil)
		if err != nil {
			gologger.Error().Msgf("Task [%s] var error: %v", task.Name, err)
			continue
		}

		sshConf, err := targetSvc.CaseRef.GetSSHConfig()
		if err != nil {
			gologger.Error().Msgf("Task [%s] SSH config error: %v", task.Name, err)
			continue
		}

		func() {
			client, err := sshutil.NewClient(sshConf)
			if err != nil {
				gologger.Error().Msgf("Task [%s] SSH connect failed: %v", task.Name, err)
				return
			}
			defer client.Close()

			logger, _ := logMgr.NewServiceLogger("setup")
			if logger != nil {
				logger.ServiceName = "setup"
				defer logger.Close()
			}

			for _, cmd := range cmds {
				gologger.Info().Msgf("[setup] Task: %s | Cmd: %s", task.Name, cmd)
				if err := client.RunCommandWithLogger(cmd, logger); err != nil {
					gologger.Error().Msgf("[setup] Task failed: %v", err)
				}
			}
		}()
	}
	return nil
}

func expandVariable(raw string, ctx map[string]*RuntimeService, currentSvc *RuntimeService) ([]string, error) {
	re := regexp.MustCompile(`\$\{(.+?)\}`)
	matches := re.FindAllStringSubmatch(raw, -1)

	if len(matches) == 0 {
		return []string{raw}, nil
	}

	fullExpr := matches[0][0]
	innerContent := matches[0][1]
	parts := strings.Split(innerContent, ".")

	if len(parts) != 3 || parts[1] != "outputs" {
		return []string{raw}, nil
	}

	refName, outputKey := parts[0], parts[2]
	var candidates []*RuntimeService

	// 1. 精确
	if s, ok := ctx[refName]; ok {
		candidates = append(candidates, s)
	}

	// 2. 上下文
	if len(candidates) == 0 && currentSvc != nil {
		suffix := strings.TrimPrefix(currentSvc.Name, currentSvc.RawName)
		if suffix != "" {
			guessedName := refName + suffix
			if s, ok := ctx[guessedName]; ok && s.RawName == refName {
				candidates = append(candidates, s)
			}
		}
	}

	// 3. 广播
	if len(candidates) == 0 {
		for _, s := range ctx {
			if s.RawName == refName {
				candidates = append(candidates, s)
			}
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("referenced service '%s' not found or not active", refName)
	}

	var results []string
	for _, target := range candidates {
		if !target.IsDeployed {
			return nil, fmt.Errorf("referenced service '%s' is not deployed", target.Name)
		}
		val, ok := target.Outputs[outputKey]
		if !ok {
			return nil, fmt.Errorf("output key '%s' missing in %s", outputKey, target.Name)
		}
		newStr := strings.ReplaceAll(raw, fullExpr, fmt.Sprint(val))
		results = append(results, newStr)
	}
	return results, nil
}

func canDeploy(svc *RuntimeService, all map[string]*RuntimeService) bool {
	for _, depName := range svc.Spec.DependsOn {
		foundAny := false
		for _, rtSvc := range all {
			if rtSvc.RawName == depName {
				foundAny = true
				if !rtSvc.IsDeployed {
					return false
				}
			}
		}
		if !foundAny {
			continue
		}
	}
	return true
}

func canDestroy(target *RuntimeService, all map[string]*RuntimeService) bool {
	for _, other := range all {
		if !other.IsDeployed {
			continue
		}
		for _, dep := range other.Spec.DependsOn {
			if dep == target.RawName {
				return false
			}
		}
	}
	return true
}

func parseTfOutput(outputs map[string]tfexec.OutputMeta) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range outputs {
		var val interface{}
		if jsonErr := json.Unmarshal(v.Value, &val); jsonErr != nil {
			res[k] = string(v.Value)
		} else {
			res[k] = val
		}
	}
	return res
}
