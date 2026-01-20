package mod

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"red-cloud/mod/gologger"

	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

var commonCachePath = "./tf-plugin-cache" // Provider 插件缓存目录
var ProjectPath = "./redc-taskresult"
var ProjectFile = "project.json"
var planPath = "case.tfplan"

// Config 结构体用于解析 YAML
type Config struct {
	Providers struct {
		Aws struct {
			AccessKey string `yaml:"access_key"`
			SecretKey string `yaml:"secret_key"`
			Region    string `yaml:"region"`
		} `yaml:"aws"`
		Alicloud struct {
			AccessKey string `yaml:"access_key"`
			SecretKey string `yaml:"secret_key"`
			Region    string `yaml:"region"`
		} `yaml:"aliyun"`
		Tencentcloud struct {
			SecretId  string `yaml:"secret_id"`
			SecretKey string `yaml:"secret_key"`
			Region    string `yaml:"region"`
		} `yaml:"tencentcloud"`
	} `yaml:"providers"`
}

// LoadConfig 将配置写入环境变量，Terraform Provider 会自动读取
func LoadConfig(path string) error {
	if path != "" {
		viper.SetConfigFile(path)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".") // 当前目录
		home, _ := os.UserHomeDir()
		viper.AddConfigPath(home + "/.config/redc")   // 用户目录
		viper.AddConfigPath("/opt/homebrew/etc/redc") // Brew (M1/M2/M3)
		viper.AddConfigPath("/usr/local/etc/redc")    // Brew (Intel)
		viper.AddConfigPath("/etc/redc")              // Linux 系统级
	}
	// 读取配置
	// 如果指定了 cfgFile 但文件不存在，这里会直接报错
	if err := viper.ReadInConfig(); err != nil {
		// 处理错误：如果是“未找到配置文件”，且并不是用户强制指定的，通常可以选择忽略或使用默认值
		// 但如果是用户用 --config 指定的，文件不存在必须报错
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			gologger.Info().Msgf("未发现配置文件，将使用环境变量配置！")
		} else {
			gologger.Error().Msgf("配置文件加载失败: %s", err)
		}
	}

	var conf Config
	if err := viper.Unmarshal(&conf); err != nil {
		return err
	}
	// 设置标准 Terraform 环境变量
	usr, err := user.Current()
	err = os.Setenv("TF_PLUGIN_CACHE_DIR", filepath.Join(usr.HomeDir, ".terraform.d", "plugin-cache"))
	if err != nil {
		gologger.Error().Msgf("设置环境变量失败: %s", err)
	}
	os.Setenv("AWS_ACCESS_KEY_ID", conf.Providers.Aws.AccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", conf.Providers.Aws.SecretKey)

	os.Setenv("ALICLOUD_ACCESS_KEY", conf.Providers.Alicloud.AccessKey)
	os.Setenv("ALICLOUD_SECRET_KEY", conf.Providers.Alicloud.SecretKey)

	os.Setenv("TENCENTCLOUD_SECRET_ID", conf.Providers.Tencentcloud.SecretId)
	os.Setenv("TENCENTCLOUD_SECRET_KEY", conf.Providers.Tencentcloud.SecretKey)

	return nil
}

// ParseConfig 解析配置文件
func ParseConfig(path string) (string, string) {
	cfg, err := ini.Load(path)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	ALICLOUD_ACCESS_KEY := cfg.Section("").Key("ALICLOUD_ACCESS_KEY").String()
	ALICLOUD_SECRET_KEY := cfg.Section("").Key("ALICLOUD_SECRET_KEY").String()

	return ALICLOUD_ACCESS_KEY, ALICLOUD_SECRET_KEY
}
