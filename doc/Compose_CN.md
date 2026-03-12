# Redc Compose 示例：阿里云 + 火山云 ECS 部署

这是一个简单的 redc compose 编排示例，用于同时部署阿里云和火山云的 ECS 实例。

## 前置要求

### 1. 确保 redc 配置文件存在

配置文件位置：`~/redc/config.yaml`

```yaml
providers:
  aliyun:
    ALICLOUD_ACCESS_KEY: "你的阿里云 AccessKey"
    ALICLOUD_SECRET_KEY: "你的阿里云 SecretKey"
    region: "cn-hangzhou"
  
  volcengine:
    VOLCENGINE_ACCESS_KEY: "你的火山云 AccessKey"
    VOLCENGINE_SECRET_KEY: "你的火山云 SecretKey"
    region: "cn-beijing"
```

### 2. 下载模板

```bash
# 下载阿里云 ECS 模板
redc pull aliyun/ecs

# 下载火山云 ECS 模板（如果存在）
redc pull volcengine/ecs

# 初始化模板
redc init
```

### 3. 修改配置文件

根据实际需求修改 `redc-compose.yaml` 中的：
- 实例密码（`password`）
- 实例规格（`instance_type`）
- 镜像 ID（`image_id`）
- 区域（`region`）

## 使用步骤

### 预览配置

在实际部署前，可以先预览编排计划：

```bash
redc compose config redc-compose.yaml
```

这会显示：
- 将要创建的服务列表
- 每个服务的配置变量
- 依赖关系
- 后置任务

### 启动编排

```bash
# 在 redc-compose.yaml 所在目录执行
redc compose up redc-compose.yaml
```

执行过程：
1. 创建阿里云 ECS 实例
2. 创建火山云 ECS 实例
3. 等待实例启动完成
4. 执行实例内的初始化命令
5. 执行 setup 后置任务

### 查看状态

```bash
# 查看所有实例状态
redc ps
```

### 连接实例

```bash
# 连接阿里云实例
redc exec <aliyun_caseid> -t bash

# 连接火山云实例
redc exec <volcengine_caseid> -t bash
```

### 执行命令

```bash
# 在阿里云实例执行命令
redc exec <aliyun_caseid> "whoami"

# 在火山云实例执行命令
redc exec <volcengine_caseid> "uname -a"
```

### 销毁环境

```bash
# 销毁所有实例
redc compose down -f redc-compose.yaml
```

## 高级用法

### 1. 使用 Profile 控制环境

Profile 是一种**服务分组/环境过滤**机制，允许你在同一个 compose 文件中定义多组服务，按需选择性部署。

**工作原理：**

- 每个服务可以通过 `profiles` 字段标记属于哪些分组（如 `prod`、`dev`、`attack`）
- 执行 `compose up` 时指定 `-p <profile>` 参数，**只有属于该 profile 的服务会被部署**
- **没有 `profiles` 字段的服务**视为默认服务，在不指定任何 profile 时会被部署
- 不指定 `-p` 参数时，所有没有 `profiles` 字段的服务都会启动

**适用场景举例：**

| 场景 | 说明 |
|------|------|
| 多环境隔离 | 同一文件定义 prod 和 dev 环境的服务器，按需部署其中一组 |
| 按需扩展 | 将压测节点标记为 `attack` profile，仅在需要压测时启动 |
| 渐进部署 | 先部署 `base` 基础设施，再部署 `app` 应用层 |

> **提示：** 如果你的所有服务都需要一起部署，不需要设置 `profiles` 字段，也不需要指定 `-p` 参数，此时该功能可以忽略。在 GUI 的编排管理页面中，该选项位于"高级选项"面板内。

修改配置文件，为服务添加 profile：

```yaml
services:
  aliyun_server:
    profiles:
      - prod
      - dev
    # ... 其他配置

  volcengine_server:
    profiles:
      - prod
    # ... 其他配置

  monitor_server:
    # 没有 profiles 字段 → 不指定 -p 时默认部署
    image: ./templates/monitor
```

只启动特定环境：

```bash
# 只启动 prod 环境的服务（aliyun_server + volcengine_server）
redc compose up -f redc-compose.yaml -p prod

# 只启动 dev 环境的服务（aliyun_server）
redc compose up -f redc-compose.yaml -p dev

# 不指定 profile，部署所有无 profiles 字段的服务（monitor_server）
redc compose up -f redc-compose.yaml
```

### 2. 文件上传

在服务配置中添加 volumes：

```yaml
services:
  aliyun_server:
    volumes:
      - ./scripts/init.sh:/root/init.sh
      - ./config/app.conf:/etc/app/config.conf
    command: |
      chmod +x /root/init.sh
      bash /root/init.sh
```

### 3. 文件下载

在服务配置中添加 downloads：

```yaml
services:
  aliyun_server:
    downloads:
      - /var/log/app.log:./logs/aliyun_app.log
      - /root/.ssh/id_rsa.pub:./keys/aliyun_key.pub
```

### 4. 服务依赖

```yaml
services:
  database:
    image: aliyun/ecs
    # ... 配置
  
  app_server:
    image: volcengine/ecs
    depends_on:
      - database
    environment:
      - db_host=${database.outputs.private_ip}
    # ... 配置
```

### 5. 多副本部署

```yaml
services:
  worker_nodes:
    image: aliyun/ecs
    deploy:
      replicas: 3  # 创建 3 个实例
    # ... 配置
```

会自动创建：worker_nodes_1, worker_nodes_2, worker_nodes_3

## 常见问题

### Q1: 模板找不到？

确保已经下载并初始化模板：
```bash
redc pull aliyun/ecs
redc pull volcengine/ecs
redc init
```

### Q2: 认证失败？

检查 `~/redc/config.yaml` 中的 AccessKey 和 SecretKey 是否正确。

### Q3: 如何查看详细日志？

添加 `--debug` 参数：
```bash
redc compose up -f redc-compose.yaml --debug
```

### Q4: 实例启动失败？

1. 检查实例规格是否在所选区域可用
2. 检查镜像 ID 是否正确
3. 检查账户余额是否充足
4. 使用 `redc compose config` 预览配置

## 参考资料

- [Redc 官方文档](https://github.com/wgpsec/redc)
- [模板仓库](https://github.com/wgpsec/redc-template)
- [在线模板浏览](https://redc.wgpsec.org/)
