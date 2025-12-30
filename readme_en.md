# redc

[中文](README.md) | English

---

## Build

> Before building, copy the scenarios you need from the https://github.com/wgpsec/redc-template repository to the redc/utils/redc-templates/ path as needed !!!

After copying, compile using goreleaser

**goreleaser**
```
brew install goreleaser

goreleaser --snapshot --clean
```

## Local Dependency Tool Installation

**mac**

aliyun-cli installation
```bash
brew install aliyun-cli
```

terraform installation
```bash
brew tap hashicorp/tap
brew install hashicorp/tap/terraform
brew upgrade hashicorp/tap/terraform
```

jq installation
```bash
brew install jq
```

aws-cli installation
- https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html

**linux**
```bash
# terraform
mkdir -p /tmp/terraform && cd /tmp/terraform && wget -O terraform_1.6.6_linux_amd64.zip 'https://releases.hashicorp.com/terraform/1.6.6/terraform_1.6.6_linux_amd64.zip'
unzip terraform_1.6.6_linux_amd64.zip
mv --force terraform /usr/local/bin/terraform > /dev/null 2>&1 && chmod +x /usr/local/bin/terraform
rm -rf /tmp/terraform

# For arm64 machines
# mkdir -p /tmp/terraform && cd /tmp/terraform && wget -O terraform_1.6.6_linux_arm64.zip 'https://releases.hashicorp.com/terraform/1.6.6/terraform_1.6.6_linux_arm64.zip'
# unzip terraform_1.6.6_linux_arm64.zip
# mv --force terraform /usr/local/bin/terraform > /dev/null 2>&1 && chmod +x /usr/local/bin/terraform
# rm -rf /tmp/terraform

cd /tmp
terraform -version

# aliyun
mkdir -p /tmp/aliyuncli && cd /tmp/aliyuncli && wget -O aliyun-cli-linux-latest-amd64.tgz 'https://aliyuncli.alicdn.com/aliyun-cli-linux-latest-amd64.tgz?file=aliyun-cli-linux-latest-amd64.tgz'
tar -xzvf aliyun-cli-linux-latest-amd64.tgz
mv --force aliyun /usr/local/bin/aliyun > /dev/null 2>&1 && chmod +x /usr/local/bin/aliyun
rm -rf /tmp/aliyuncli

apt install jq || yum install jq

# aws
https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html
aws configure
```

## CLI Dependency Configuration

**aliyun cli configuration**
```bash
aliyun configure set --profile cloud-tool --mode AK --region cn-beijing --access-key-id xxxxxxxxxxxxxx --access-key-secret xxxxxxxxxxxxxx
```

**aws-cli configuration**
```
aws configure
AKIAXXXXXXXXXXXXX
XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
ap-east-1
text
```

**Configure rclone** (Configure this if you use proxy pool to transfer to r2, otherwise it's optional)
```
rclone config
s3
Cloudflare R2 Storage
XXXXXXXXXXXXXXXXXXX
XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
https://XXXXXXXXXXXXXXXXXXXXXXXXXXX.r2.cloudflarestorage.com
auto

rclone lsf r2:test
```

Configure tf plugin cache path
```bash
echo 'plugin_cache_dir = "$HOME/.terraform.d/plugin-cache"' > ~/.terraformrc
```

Initialize redc before use, it will automatically download tf module dependencies (If you repackage the template, you need to initialize again after recompilation)
```
./redc -init
```

## Interactive Usage

```bash
# Start awvs scenario
redc -start awvs -u zhangsan
.........
.........
Scenario uuid:xxxxxxxxx

# Check the status of a specified scenario
redc -status [uuid]

# Stop a specified scenario
redc -stop [uuid]

# View all scenarios
redc -list
uuid        type    createtime      operator
xxxxxxxxx   awvs    2022.02.22      system
bbbbbbbbb   file    2022.02.22      system
```

Scenario name - corresponds to the template repository https://github.com/wgpsec/redc-template

Use the "folder name" you placed in the redc/utils/redc-templates/ path

For specific usage and commands for each scenario, please check the readme of the specific scenario in the template repository https://github.com/wgpsec/redc-template

---

## Design Plan

1. Create a new project first
2. Creating a scenario under a specified project will copy a scenario folder from the scenario library to the project folder
3. Creating the same scenario under different projects will not interfere with each other
4. Creating the same scenario under the same project will not interfere with each other
5. Multiple user operations will not interfere with each other (local authentication is done, but this should actually be done on the platform)

- redc configuration file (.redc.ini)
- Project1 (./project1)
    - Scenario1 (./project1/[uuid1])
        - main.tf
        - version.tf
        - output.tf
    - Scenario2 (./project1/[uuid2])
        - main.tf
        - version.tf
        - output.tf
    - Project status file (project.ini)
- Project2 (./project2)
    - Scenario1 (./project2/[uuid1])
        - main.tf
        - version.tf
        - output.tf
    - Scenario2 (./project2/[uuid2])
        - ...
    - Project status file (project.ini)
- Project3 (./project3)
    - ...

## Article Introduction

- https://mp.weixin.qq.com/s/JH-IlL_GFgZp3xXeOFzZeQ
