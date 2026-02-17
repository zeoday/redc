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
  }
];

export function getTemplatesByType(type) {
  return userdataTemplates.filter(t => t.type === type);
}
