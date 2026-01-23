# Redc Skills - AI 运维指南

基于 Terraform 的多云红队基础设施自动化工具。

## 快速参考

**配置:** `~/redc/config.yaml` | **模板:** `~/redc/redc-templates/` | **文档:** https://github.com/wgpsec/redc

## 配置

```yaml
providers:
  aws:
    AWS_ACCESS_KEY_ID: "KEY"
    AWS_SECRET_ACCESS_KEY: "SECRET"
    region: "us-east-1"
  aliyun:
    ALICLOUD_ACCESS_KEY: "KEY"
    ALICLOUD_SECRET_KEY: "SECRET"
    region: "cn-hangzhou"
  tencentcloud:
    TENCENTCLOUD_SECRET_ID: "ID"
    TENCENTCLOUD_SECRET_KEY: "KEY"
    region: "ap-guangzhou"
```

**环境变量:** 如 config.yaml 不可用，设置 `AWS_ACCESS_KEY_ID`、`ALICLOUD_ACCESS_KEY`、`TENCENTCLOUD_SECRET_ID` 等。

## 全局参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--config` | `~/redc/config.yaml` | 配置文件路径 |
| `-u, --user` | `system` | 用户标识 |
| `--project` | `default` | 项目名称 |
| `--debug` | `false` | 调试模式 |

## 命令

### 初始化
```bash
redc init                           # 初始化模板
redc pull <镜像>                    # 下载模板（如 aliyun/ecs）
redc image ls                       # 列出本地模板
```

### 部署
```bash
redc run <模板> -n <名称> -e key=val    # 规划并启动（创建基础设施）
redc plan <模板> -n <名称>             # 仅规划（预览而不创建）
redc start <case-id>                    # 启动场景（创建基础设施）
```

### 管理
```bash
redc ps                             # 列出所有场景
redc status <case-id>               # 检查状态
redc stop <case-id>                 # 停止基础设施
redc kill <case-id>                 # 停止并删除
redc rm <case-id>                   # 删除场景
```

### 执行
```bash
redc exec <case-id> <命令>          # 运行命令
redc exec -t <case-id> bash         # 交互式 shell
redc cp <源> <case-id>:<目标>       # 上传文件
redc cp <case-id>:<源> <目标>       # 下载文件
```

## 工作流

**快速部署:**
```bash
redc pull aliyun/ecs && redc init
redc run aliyun/ecs -n myserver -e password=Pass123
# 返回 case-id（如 8a57078ee856）
redc exec 8a57078ee856 whoami
```

**受控部署:**
```bash
redc plan aws/ec2 -n staging
# 审查规划的资源后:
redc start <case-id>
redc cp deploy.sh <case-id>:/root/
redc exec <case-id> "bash /root/deploy.sh"
```

**清理:**
```bash
redc stop <case-id> && redc rm <case-id>
# 或: redc kill <case-id>
```

## 自动化

```python
import subprocess, re

def redc_run(template, name, env=None):
    cmd = ["redc", "run", template, "-n", name]
    if env:
        for k, v in env.items():
            cmd.extend(["-e", f"{k}={v}"])
    result = subprocess.run(cmd, capture_output=True, text=True, check=True)
    match = re.search(r'[a-f0-9]{12,64}', result.stdout)
    return match.group(0) if match else None

# 使用
case_id = redc_run("aliyun/ecs", "auto_deploy", {"password": "Secure123"})
```

## 输出模式

- **成功:** 输出中有 `✅`
- **错误:** 输出中有 `❌`
- **Case ID:** `[a-f0-9]{64}`（使用前12个字符）
- **状态:** `running`、`stopped`、`created`、`error`

## 错误处理

| 错误 | 解决方案 |
|------|----------|
| 配置未找到 | 创建 `~/redc/config.yaml` |
| 模板未找到 | 运行 `redc pull <模板>` |
| Case ID 未找到 | 检查 `redc ps` |
| SSH 超时 | 验证实例运行中，安全组设置 |
| 初始化失败 | 检查网络，配置 Terraform 镜像 |

## JSON Schema

**场景:**
```json
{
  "id": "string[64]",
  "name": "string",
  "template": "string",
  "status": "running|stopped|created|error",
  "outputs": {"public_ip": "string"}
}
```

**配置:**
```json
{
  "providers": {
    "aws": {"AWS_ACCESS_KEY_ID": "string", "region": "string"}
  }
}
```

## 最佳实践

- 使用短 ID（前12个字符）
- 分配有意义的名称: `<项目>_<用途>_<环境>`
- 总是清理: `redc stop` → `redc rm`
- 使用 `--debug` 进行故障排除
- 永远不要将 config.yaml 提交到版本控制
- 使用 `redc ps` 监控成本

## 资源

- 仓库: https://github.com/wgpsec/redc
- 模板: https://github.com/wgpsec/redc-template
- 在线: https://redc.wgpsec.org/

---
**版本:** 1.0.0 | **许可证:** Apache 2.0
