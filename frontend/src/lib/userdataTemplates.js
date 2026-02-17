export const userdataTemplates = [
  {
    name: 'Basic Setup',
    nameZh: '基础设置',
    type: 'bash',
    script: `#!/bin/bash
# Update system packages
apt-get update -y
apt-get upgrade -y

# Install common tools
apt-get install -y curl wget git vim

echo "Setup completed!"
`
  },
  {
    name: 'Docker Installation',
    nameZh: 'Docker 安装',
    type: 'bash',
    script: `#!/bin/bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Start Docker service
systemctl start docker
systemctl enable docker

# Add user to docker group
usermod -aG docker ubuntu

echo "Docker installed successfully!"
`
  },
  {
    name: 'Nginx Installation',
    nameZh: 'Nginx 安装',
    type: 'bash',
    script: `#!/bin/bash
# Install Nginx
apt-get update -y
apt-get install -y nginx

# Start Nginx service
systemctl start nginx
systemctl enable nginx

echo "Nginx installed and started!"
`
  },
  {
    name: 'Docker Compose',
    nameZh: 'Docker Compose 安装',
    type: 'bash',
    script: `#!/bin/bash
# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Verify installation
docker-compose --version

echo "Docker Compose installed successfully!"
`
  },
  {
    name: 'MySQL Client',
    nameZh: 'MySQL 客户端安装',
    type: 'bash',
    script: `#!/bin/bash
# Install MySQL client
apt-get update -y
apt-get install -y default-mysql-client

echo "MySQL client installed successfully!"
`
  },
  {
    name: 'PostgreSQL Client',
    nameZh: 'PostgreSQL 客户端安装',
    type: 'bash',
    script: `#!/bin/bash
# Install PostgreSQL client
apt-get update -y
apt-get install -y postgresql-client

echo "PostgreSQL client installed successfully!"
`
  },
  {
    name: 'Node.js Installation',
    nameZh: 'Node.js 安装',
    type: 'bash',
    script: `#!/bin/bash
# Install Node.js 18.x
curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
apt-get install -y nodejs

# Verify installation
node --version
npm --version

echo "Node.js installed successfully!"
`
  },
  {
    name: 'Python Installation',
    nameZh: 'Python 安装',
    type: 'bash',
    script: `#!/bin/bash
# Install Python and pip
apt-get update -y
apt-get install -y python3 python3-pip python3-venv

# Verify installation
python3 --version
pip3 --version

echo "Python installed successfully!"
`
  },
  {
    name: 'Basic Setup',
    nameZh: '基础设置',
    type: 'powershell',
    script: `<powershell>
# Set execution policy
Set-ExecutionPolicy Bypass -Scope Process -Force

# Install Chocolatey
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))

Write-Host "Setup completed!"
</powershell>
`
  },
  {
    name: 'IIS Installation',
    nameZh: 'IIS 安装',
    type: 'powershell',
    script: `<powershell>
# Install IIS
Install-WindowsFeature -name Web-Server -IncludeManagementTools

Write-Host "IIS installed successfully!"
</powershell>
`
  },
  {
    name: 'Docker Desktop',
    nameZh: 'Docker Desktop 安装',
    type: 'powershell',
    script: `<powershell>
# Install Docker Desktop via Chocolatey
choco install docker-desktop -y

# Start Docker Desktop
Start-Process 'C:\Program Files\Docker\Docker\Docker Desktop.exe'

Write-Host "Docker Desktop installation completed!"
</powershell>
`
  },
  {
    name: 'VSCode Installation',
    nameZh: 'VSCode 安装',
    type: 'powershell',
    script: `<powershell>
# Install VSCode via Chocolatey
choco install vscode -y

Write-Host "VSCode installed successfully!"
</powershell>
`
  },
  {
    name: 'Git Installation',
    nameZh: 'Git 安装',
    type: 'powershell',
    script: `<powershell>
# Install Git via Chocolatey
choco install git -y

# Verify installation
git --version

Write-Host "Git installed successfully!"
</powershell>
`
  },
  {
    name: 'OpenClaw',
    nameZh: 'OpenClaw',
    type: 'bash',
    category: 'ai',
    url: 'https://openclaw.ai/',
    description: '一款开源个人 AI 智能助理。',
    installNotes: '安装完成后，输入 openclaw onboard --install-daemon 命令配置',
    script: `#!/bin/bash
# OpenClaw 自动安装脚本
curl -fsSL https://openclaw.ai/install.sh | bash
`
  },
  {
    name: 'n8n',
    nameZh: 'n8n',
    type: 'bash',
    category: 'ai',
    url: 'https://n8n.io/',
    description: '开源工作流自动化工具，支持 AI 节点集成',
    installNotes: '启动命令: n8n\n环境变量: N8N_BASIC_AUTH_ACTIVE=true 设置登录账号密码',
    script: `#!/bin/bash
# n8n 自动安装脚本
npm install -g n8n

# 启动 n8n
# n8n
# 默认端口: 5678
`
  },
  {
    name: 'Dify',
    nameZh: 'Dify',
    type: 'bash',
    category: 'ai',
    url: 'https://dify.ai/',
    description: '开源 LLM 应用开发平台，支持 RAG 和 AI Agent',
    installNotes: '访问 http://localhost:3000 进行初始化配置\n默认管理员邮箱: admin@example.com',
    script: `#!/bin/bash
# Dify Docker 部署
# 请确保已安装 Docker 和 Docker Compose

git clone https://github.com/langgenius/dify.git
cd dify/docker
docker-compose up -d

# 默认端口: 3000
`
  }
];

export function getTemplatesByType(type) {
  return userdataTemplates.filter(t => t.type === type);
}

export function getTemplatesByCategory(category) {
  return userdataTemplates.filter(t => t.category === category);
}
