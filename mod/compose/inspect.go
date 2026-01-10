package compose

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
)

// InspectConfig 解析并打印编排计划
func InspectConfig(opts ComposeOptions) error {
	ctx, err := NewComposeContext(opts)
	if err != nil {
		return err
	}

	fmt.Printf("\n📋 编排计划预览 (Project: %s)\n", ctx.Project.ProjectName)
	fmt.Printf("检测到配置文件: %s\n", opts.File)
	fmt.Printf("激活 Profile: %v\n", opts.Profiles)
	fmt.Println(strings.Repeat("-", 60))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	for _, name := range ctx.SortedSvcKeys {
		svc := ctx.RuntimeSvcs[name]
		vars := previewTfVars(svc, ctx)

		fmt.Fprintf(w, "Service:\t%s\n", svc.Name)
		fmt.Fprintf(w, "Template:\t%s\n", svc.Spec.Image)
		if svc.RawName != svc.Name {
			fmt.Fprintf(w, "Based On:\t%s (Provider: %v)\n", svc.RawName, svc.Spec.Provider)
		}

		if len(vars) > 0 {
			fmt.Fprintln(w, "Variables:")
			for k, v := range vars {
				if len(v) > 50 && !strings.Contains(v, "<computed") {
					v = v[:47] + "..."
				}
				fmt.Fprintf(w, "  - %s:\t%s\n", k, v)
			}
		} else {
			fmt.Fprintln(w, "Variables:\t(None)")
		}

		if len(svc.Spec.DependsOn) > 0 {
			fmt.Fprintf(w, "Depends On:\t%v\n", svc.Spec.DependsOn)
		}
		fmt.Fprintln(w, strings.Repeat("-", 60))
	}

	w.Flush()
	fmt.Printf("\n总计将创建/管理 %d 个服务实例。\n", len(ctx.RuntimeSvcs))
	return nil
}

func previewTfVars(svc *RuntimeService, ctx *ComposeContext) map[string]string {
	tfVars := make(map[string]string)

	// Configs
	for _, cfgStr := range svc.Spec.Configs {
		parts := strings.SplitN(cfgStr, "=", 2)
		if len(parts) == 2 {
			tfName, cfgKey := parts[0], parts[1]
			if _, ok := ctx.GlobalConfigs[cfgKey]; ok {
				tfVars[tfName] = fmt.Sprintf("<File/Config Content: %s>", cfgKey)
			} else {
				tfVars[tfName] = "<Error: Config Not Found>"
			}
		}
	}

	// Environment
	for _, envStr := range svc.Spec.Environment {
		parts := strings.SplitN(envStr, "=", 2)
		if len(parts) == 2 {
			key, rawVal := parts[0], parts[1]
			vals := previewExpandVariable(rawVal, ctx.RuntimeSvcs, svc)
			tfVars[key] = strings.Join(vals, ",")
		}
	}

	// Provider Alias
	if pStr, ok := svc.Spec.Provider.(string); ok && pStr != "" && pStr != "default" {
		tfVars["provider_alias"] = pStr
	}

	return tfVars
}

func previewExpandVariable(raw string, ctx map[string]*RuntimeService, currentSvc *RuntimeService) []string {
	re := regexp.MustCompile(`\$\{(.+?)\}`)
	matches := re.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return []string{raw}
	}

	fullExpr := matches[0][0]
	innerContent := matches[0][1]
	parts := strings.Split(innerContent, ".")

	if len(parts) != 3 || parts[1] != "outputs" {
		return []string{raw}
	}

	refName, outputKey := parts[0], parts[2]

	// 简单检查是否存在
	found := false
	if _, ok := ctx[refName]; ok {
		found = true
	}
	if !found {
		for _, s := range ctx {
			if s.RawName == refName {
				found = true
				break
			}
		}
	}

	if !found {
		return []string{fmt.Sprintf("<Error: Svc '%s' Not Found>", refName)}
	}

	// 返回模拟值
	placeholder := fmt.Sprintf("<Computed: %s.%s>", refName, outputKey)
	return []string{strings.ReplaceAll(raw, fullExpr, placeholder)}
}
