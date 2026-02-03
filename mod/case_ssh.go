package mod

import (
	"fmt"
	"path/filepath"
	"red-cloud/utils/sshutil"
	"strings"
	"time"
)

// InstanceInfo 结构体
type InstanceInfo struct {
	ID       string
	IP       string
	User     string
	Password string
	KeyPath  string
	Port     int
}

// GetSSHConfig 统一获取 SSH 连接配置
// 自动尝试多种常见的 Output Key (兼容 ecs_ip/public_ip, password/ecs_password)
func (c *Case) GetSSHConfig() (*sshutil.SSHConfig, error) {
	if c == nil {
		return nil, fmt.Errorf("case instance is nil")
	}

	// 1. 尝试获取 IP
	// 优先顺序: public_ip -> ecs_ip -> ip
	ipKeys := []string{"public_ip", "ecs_ip", "ip", "main_ip"}
	var ip string
	var err error

	for _, key := range ipKeys {
		ip, err = c.GetInstanceInfo(key)
		if err == nil && ip != "" {
			break
		}
	}
	if ip == "" {
		return nil, fmt.Errorf("无法获取实例 IP (尝试了: %v)", ipKeys)
	}

	// 2. 尝试获取密码
	// 优先顺序: password -> ecs_password -> root_password
	pwdKeys := []string{"password", "ecs_password", "root_password"}
	var pwd string
	for _, key := range pwdKeys {
		pwd, _ = c.GetInstanceInfo(key)
		if pwd != "" {
			break
		}
	}

	// 3. 尝试获取 SSH 私钥路径
	// 优先顺序: ssh_key_path -> private_key_path -> key_path
	keyPathKeys := []string{"ssh_key_path", "private_key_path", "key_path"}
	var keyPath string
	for _, key := range keyPathKeys {
		keyPath, _ = c.GetInstanceInfo(key)
		if keyPath != "" {
			break
		}
	}

	// 3.1 如果是相对路径，转换为相对于 Case 目录的绝对路径
	if keyPath != "" && c.Path != "" {
		if strings.HasPrefix(keyPath, "./") || strings.HasPrefix(keyPath, "../") {
			keyPath = filepath.Join(c.Path, keyPath)
		}
	}

	// 4. 返回标准配置
	return &sshutil.SSHConfig{
		Host:     ip,
		Port:     22,
		User:     "root",
		Password: pwd,
		KeyPath:  keyPath,
		Timeout:  5 * time.Second,
	}, nil
}
func (c *Case) getSSHClient() (*sshutil.Client, error) {
	info, err := c.GetSSHConfig()
	if err != nil {
		return nil, err
	}
	return sshutil.NewClient(info)
}
