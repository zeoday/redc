package mod

import (
    "bufio"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "red-cloud/mod/gologger"
)

// moduleRegistry 注册所有可用的模块钩子
var moduleRegistry = map[string]func(*Case) error{
    "gen_clash_config": genClashConfig,
    "upload_r2":       uploadR2,
}

// genClashConfig 根据模板 outputs 与 tfvars 生成 Clash 配置并上传 R2
func genClashConfig(c *Case) error {
    // 确保 output 已加载
    if c.output == nil {
        if _, err := c.TfOutput(); err != nil {
            return fmt.Errorf("获取场景 output 失败: %w", err)
        }
    }

    ips, err := collectIPs(c)
    if err != nil {
        return err
    }
    if len(ips) == 0 {
        return fmt.Errorf("未获取到任何节点 IP")
    }

    vars, err := loadTfVars(filepath.Join(c.Path, "terraform.tfvars"))
    if err != nil {
        return err
    }
    ssPort, okP := vars["port"]
    ssPass, okS := vars["password"]
    if !okP || !okS {
        return fmt.Errorf("terraform.tfvars 缺少 port 或 password")
    }
    fileName := vars["filename"]
    if fileName == "" {
        fileName = "default-config.yaml"
    }

    configContent, err := buildClashConfig(ips, ssPort, ssPass)
    if err != nil {
        return err
    }

    // 写入本地文件
    configPath := filepath.Join(c.Path, "config.yaml")
    if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
        return fmt.Errorf("写入 config.yaml 失败: %w", err)
    }

    // 生成上传副本（文件名可配置）
    uploadName := filepath.Join(c.Path, fileName)
    if err := os.WriteFile(uploadName, []byte(configContent), 0644); err != nil {
        return fmt.Errorf("写入 %s 失败: %w", fileName, err)
    }

    gologger.Info().Msgf("Clash 配置生成完成，文件位置: %s", configPath)
    return nil
}

func collectIPs(c *Case) ([]string, error) {
    // 优先用 Terraform Output 中的 ecs_ip（list），兼容 public_ip（string/array）
    tryKeys := []string{"ecs_ip", "public_ip"}
    for _, key := range tryKeys {
        if meta, ok := c.output[key]; ok {
            var arr []string
            if err := json.Unmarshal(meta.Value, &arr); err == nil {
                return arr, nil
            }
            var single string
            if err := json.Unmarshal(meta.Value, &single); err == nil && single != "" {
                return []string{single}, nil
            }
        }
    }
    return nil, fmt.Errorf("未在 Terraform Output 中找到 ecs_ip/public_ip")
}

func loadTfVars(path string) (map[string]string, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("读取 terraform.tfvars 失败: %w", err)
    }
    defer f.Close()

    res := make(map[string]string)
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
            continue
        }
        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }
        key := strings.TrimSpace(parts[0])
        val := strings.Trim(strings.TrimSpace(parts[1]), "\"')")
        val = strings.Trim(val, "\"")
        res[key] = val
    }
    if err := scanner.Err(); err != nil {
        return nil, err
    }
    return res, nil
}

func buildClashConfig(ips []string, port string, password string) (string, error) {
    decode := func(s string) (string, error) {
        b, err := base64.StdEncoding.DecodeString(s)
        if err != nil {
            return "", err
        }
        return string(b), nil
    }

    part1, err := decode("bWl4ZWQtcG9ydDogNjQyNzcKYWxsb3ctbGFuOiB0cnVlCmJpbmQtYWRkcmVzczogJyonCm1vZGU6IHJ1bGUKbG9nLWxldmVsOiBpbmZvCmlwdjY6IGZhbHNlCmV4dGVybmFsLWNvbnRyb2xsZXI6IDEyNy4wLjAuMTo5MDkwCnNlY3JldDogdHh0dHh0eHQyc3h0eHR4dGRkeHR4dDExMTExMTEKcm91dGluZy1tYXJrOiA2NjY2Cmhvc3RzOgoKcHJvZmlsZToKICBzdG9yZS1zZWxlY3RlZDogZmFsc2UKICBzdG9yZS1mYWtlLWlwOiB0cnVlCgpkbnM6CiAgZW5hYmxlOiBmYWxzZQogIGxpc3RlbjogMC4wLjAuMDo1MwogIGRlZmF1bHQtbmFtZXNlcnZlcjoKICAgIC0gMjIzLjUuNS41CiAgICAtIDExOS4yOS4yOS4yOQogIGVuaGFuY2VkLW1vZGU6IGZha2UtaXAgIyBvciByZWRpci1ob3N0IChub3QgcmVjb21tZW5kZWQpCiAgZmFrZS1pcC1yYW5nZTogMTk4LjE4LjAuMS8xNiAjIEZha2UgSVAgYWRkcmVzc2VzIHBvb2wgQ0lEUgogIG5hbWVzZXJ2ZXI6CiAgICAtIDIyMy41LjUuNSAjIGRlZmF1bHQgdmFsdWUKICAgIC0gMTE5LjI5LjI5LjI5ICMgZGVmYXVsdCB2YWx1ZQogICAgLSB0bHM6Ly9kbnMucnVieWZpc2guY246ODUzICMgRE5TIG92ZXIgVExTCiAgICAtIGh0dHBzOi8vMS4xLjEuMS9kbnMtcXVlcnkgIyBETlMgb3ZlciBIVFRQUwogICAgLSBkaGNwOi8vZW4wICMgZG5zIGZyb20gZGhjcAogICAgIyAtICc4LjguOC44I2VuMCcKCnByb3hpZXM6Cg==")
    if err != nil {
        return "", err
    }

    part3, err := decode("cHJveHktZ3JvdXBzOgogIC0gbmFtZTogInRlc3QiCiAgICB0eXBlOiBsb2FkLWJhbGFuY2UKICAgIHByb3hpZXM6Cg==")
    if err != nil {
        return "", err
    }

    part5, err := decode("ICAgIHVybDogJ2h0dHA6Ly93d3cuZ3N0YXRpYy5jb20vZ2VuZXJhdGVfMjA0JwogICAgaW50ZXJ2YWw6IDI0MDAKICAgIHN0cmF0ZWd5OiByb3VuZC1yb2JpbgoKcnVsZXM6CiAgLSBET01BSU4tU1VGRklYLGdvb2dsZS5jb20sdGVzdAogIC0gRE9NQUlOLUtFWVdPUkQsZ29vZ2xlLHRlc3QKICAtIERPTUFJTixnb29nbGUuY29tLHRlc3QKICAtIEdFT0lQLENOLHRlc3QKICAtIE1BVENILHRlc3QKICAtIFNSQy1JUC1DSURSLDE5Mi4xNjguMS4yMDEvMzIsRElSRUNUCiAgLSBJUC1DSURSLDEyNy4wLjAuMC84LERJUkVDVAogIC0gRE9NQUlOLVNVRkZJWCxhZC5jb20sUkVKRUNUCg==")
    if err != nil {
        return "", err
    }

    var sb strings.Builder
    sb.WriteString(part1)

    // part2: proxies 列表
    for _, ip := range ips {
        sb.WriteString("  - name: \"")
        sb.WriteString(ip)
        sb.WriteString("\"\n    type: ss\n    server: ")
        sb.WriteString(ip)
        sb.WriteString("\n    port: ")
        sb.WriteString(port)
        sb.WriteString("\n    cipher: chacha20-ietf-poly1305\n    password: \"")
        sb.WriteString(password)
        sb.WriteString("\"\n\n")
    }

    // part3
    sb.WriteString(part3)

    // part4: proxy-groups 内容
    for _, ip := range ips {
        sb.WriteString("      - ")
        sb.WriteString(ip)
        sb.WriteString("\n")
    }

    // part5
    sb.WriteString(part5)

    return sb.String(), nil
}

func runRcloneUpload(workdir, filePath string) error {
    // 先删除，再上传
    delCmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && rclone deletefile r2:test/proxyfile/%s", workdir, filepath.Base(filePath)))
    if err := delCmd.Run(); err != nil {
        gologger.Debug().Msgf("rclone deletefile 失败: %v", err)
    }

    upCmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && rclone copy %s r2:test/proxyfile/", workdir, filepath.Base(filePath)))
    if out, err := upCmd.CombinedOutput(); err != nil {
        return fmt.Errorf("rclone copy 失败: %v, output: %s", err, string(out))
    }

    gologger.Info().Msgf("R2 上传完成: proxyfile/%s", filepath.Base(filePath))
    return nil
}

// uploadR2 仅负责将已生成的配置文件上传到 R2
func uploadR2(c *Case) error {
    vars, _ := loadTfVars(filepath.Join(c.Path, "terraform.tfvars"))
    fileName := vars["filename"]
    if fileName == "" {
        fileName = "default-config.yaml"
    }
    file := filepath.Join(c.Path, fileName)
    if _, err := os.Stat(file); err != nil {
        return fmt.Errorf("上传失败，未找到文件: %s", file)
    }
    return runRcloneUpload(c.Path, file)
}
