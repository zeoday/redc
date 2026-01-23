# Redc Compose Example: Alibaba Cloud + Volcengine ECS Deployment

This is a simple redc compose orchestration example for deploying ECS instances on both Alibaba Cloud and Volcengine simultaneously.

## Prerequisites

### 1. Ensure redc Configuration File Exists

Configuration file location: `~/redc/config.yaml`

```yaml
providers:
  aliyun:
    ALICLOUD_ACCESS_KEY: "Your Alibaba Cloud AccessKey"
    ALICLOUD_SECRET_KEY: "Your Alibaba Cloud SecretKey"
    region: "cn-hangzhou"
  
  volcengine:
    VOLCENGINE_ACCESS_KEY: "Your Volcengine AccessKey"
    VOLCENGINE_SECRET_KEY: "Your Volcengine SecretKey"
    region: "cn-beijing"
```

### 2. Download Templates

```bash
# Download Alibaba Cloud ECS template
redc pull aliyun/ecs

# Download Volcengine ECS template (if available)
redc pull volcengine/ecs

# Initialize templates
redc init
```

### 3. Modify Configuration File

Modify the following in `redc-compose.yaml` according to your needs:
- Instance password (`password`)
- Instance type (`instance_type`)
- Image ID (`image_id`)
- Region (`region`)

## Usage

### Preview Configuration

Before actual deployment, preview the orchestration plan:

```bash
redc compose config redc-compose.yaml
```

This will display:
- List of services to be created
- Configuration variables for each service
- Dependencies
- Post-deployment tasks

### Start Orchestration

```bash
# Execute in the directory containing redc-compose.yaml
redc compose up redc-compose.yaml
```

Execution process:
1. Create Alibaba Cloud ECS instance
2. Create Volcengine ECS instance
3. Wait for instances to start
4. Execute initialization commands within instances
5. Execute setup post-deployment tasks

### Check Status

```bash
# View all instance status
redc ps
```

### Connect to Instances

```bash
# Connect to Alibaba Cloud instance
redc exec <aliyun_caseid> -t bash

# Connect to Volcengine instance
redc exec <volcengine_caseid> -t bash
```

### Execute Commands

```bash
# Execute command on Alibaba Cloud instance
redc exec <aliyun_caseid> "whoami"

# Execute command on Volcengine instance
redc exec <volcengine_caseid> "uname -a"
```

### Destroy Environment

```bash
# Destroy all instances
redc compose down redc-compose.yaml
```

## Advanced Usage

### 1. Use Profiles to Control Environments

Modify the configuration file to add profiles to services:

```yaml
services:
  aliyun_server:
    profiles:
      - prod
      - dev
    # ... other configurations

  volcengine_server:
    profiles:
      - prod
    # ... other configurations
```

Start specific environments only:

```bash
# Start only prod environment services
redc compose up -f redc-compose.yaml -p prod

# Start dev environment services
redc compose up -f redc-compose.yaml -p dev
```

### 2. File Upload

Add volumes to service configuration:

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

### 3. File Download

Add downloads to service configuration:

```yaml
services:
  aliyun_server:
    downloads:
      - /var/log/app.log:./logs/aliyun_app.log
      - /root/.ssh/id_rsa.pub:./keys/aliyun_key.pub
```

### 4. Service Dependencies

```yaml
services:
  database:
    image: aliyun/ecs
    # ... configuration
  
  app_server:
    image: volcengine/ecs
    depends_on:
      - database
    environment:
      - db_host=${database.outputs.private_ip}
    # ... configuration
```

### 5. Multiple Replicas Deployment

```yaml
services:
  worker_nodes:
    image: aliyun/ecs
    deploy:
      replicas: 3  # Create 3 instances
    # ... configuration
```

This will automatically create: worker_nodes_1, worker_nodes_2, worker_nodes_3

## Common Issues

### Q1: Template not found?

Ensure templates are downloaded and initialized:
```bash
redc pull aliyun/ecs
redc pull volcengine/ecs
redc init
```

### Q2: Authentication failed?

Check if AccessKey and SecretKey in `~/redc/config.yaml` are correct.

### Q3: How to view detailed logs?

Add `--debug` parameter:
```bash
redc compose up -f redc-compose.yaml --debug
```

### Q4: Instance startup failed?

1. Check if instance type is available in the selected region
2. Check if image ID is correct
3. Check if account balance is sufficient
4. Use `redc compose config` to preview configuration

## References

- [Redc Official Documentation](https://github.com/wgpsec/redc)
- [Template Repository](https://github.com/wgpsec/redc-template)
- [Online Template Browser](https://redc.wgpsec.org/)
