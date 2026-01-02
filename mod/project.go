package mod

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"red-cloud/mod/gologger"
	"time"

	"gopkg.in/ini.v1"
)

// RedcProject 项目结构体
type RedcProject struct {
	ProjectName string `json:"project_name"`
	ProjectPath string `json:"project_path"`
	CreateTime  string `json:"create_time"`
	User        string `json:"user"`
	Case        []Case `json:"case"`
}

// Case 项目信息
type Case struct {
	// Id uuid
	Id       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Operator string `json:"operator"`
	Path     string `json:"path"`
	// Node 节点数量
	Node          int      `json:"node"`
	Domain        string   `json:"domain"`
	Domain2       string   `json:"domain2"`
	CreateTime    string   `json:"create_time"`
	Base64Command string   `json:"base64_command"`
	Parameter     []string `json:"parameter"`
	Plan          string   `json:"plan"`
}

func (c *Case) Apply() error {
	// 部分场景单独处理
	if c.Type == "cs-49" || c.Type == "c2-new" || c.Type == "snowc2" {
		c.Parameter = RVar(
			fmt.Sprintf("node_count=%d", Node),
			fmt.Sprintf("domain=%s", Domain),
		)

		C2Apply(c.Path)
	} else if c.Type == "aws-proxy" || c.Type == "aliyun-proxy" || c.Type == "asm" || c.Type == "asm-node" {
		c.Parameter = RVar(fmt.Sprintf("node_count=%d", Node))
	} else if c.Type == "dnslog" || c.Type == "xraydnslog" || c.Type == "interactsh" {
		if Domain == "360.com" {
			return fmt.Errorf("创建 dnslog 时,域名不可为默认值")
		}
		c.Parameter = RVar(fmt.Sprintf("domain=%s", Domain))
	} else if c.Type == "pss5" || c.Type == "frp" || c.Type == "frp-loki" || c.Type == "nps" {
		c.Parameter = []string{fmt.Sprintf("base64_command=%s", Base64Command)}
	} else if c.Type == "asm" {
		c.Parameter = RVar(
			fmt.Sprintf("node_count=%d", Node),
		)
	} else if c.Type == "asm-node" {
		c.Parameter = RVar(
			fmt.Sprintf("node_count=%d", Node),
			fmt.Sprintf("domain2=%s", Domain2),
			fmt.Sprintf("doamin=%s", Domain),
		)
	}
	err := TfApply(c.Path, c.Parameter...)
	if err != nil {
		return err
	}

	return nil
}

func ProjectParse(name string, user string) error {
	// 确认项目文件夹是否存在,不存在就创建
	path := filepath.Join(ProjectPath, name)
	_, err := os.Stat(path)
	if err != nil {
		// 创建项目目录
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("项目目录创建失败！\n%s", err.Error())
		}
		gologger.Info().Msgf("项目目录「%s」创建成功！", name)
		// 创建项目状态文件
		currentTime := time.Now().Format("2006-01-02 15:04:05")
		project := &RedcProject{
			ProjectName: name,
			ProjectPath: path,
			CreateTime:  currentTime,
			User:        user,
		}

		err := project.SaveProject()
		if err != nil {
			return err
		}

		gologger.Info().Msgf("项目状态文件「%s」创建成功！", ProjectFile)

	}
	return nil
}

// ProjectByName 读取项目配置
func ProjectByName(name string) (*RedcProject, error) {
	path := filepath.Join(ProjectPath, name, ProjectFile)
	// 读取 JSON 文件
	data, err := os.ReadFile(path)
	if err != nil {
		gologger.Debug().Msgf("项目文件读取失败 %s", err.Error())
		return nil, fmt.Errorf("项目文件读取失败: %v", err)
	}

	// 解析 JSON 数据
	var project RedcProject
	err = json.Unmarshal(data, &project)
	if err != nil {
		return nil, fmt.Errorf("解析数据失败: %v", err)
	}

	return &project, nil
}

// GetCaseByUid 从项目中匹配 case
func (p *RedcProject) GetCaseByUid(uid string) (*Case, error) {
	for i, caseInfo := range p.Case {
		if caseInfo.Id == uid {
			return &p.Case[i], nil
		}
	}
	return nil, fmt.Errorf("项目 %s ,未找到uid为 %s 的case", p.ProjectName, uid)
}

// HandleCase 删除指定uid的case
func (p *RedcProject) HandleCase(uid string) error {
	found := false
	for i, caseInfo := range p.Case {
		if caseInfo.Id == uid {
			// 执行删除逻辑：将 i 之后的所有元素前移
			p.Case = append(p.Case[:i], p.Case[i+1:]...)
			found = true
			break // 找到并删除后立即退出循环
		}
	}

	if !found {
		return fmt.Errorf("未找到 UID 为 %s 的 case，无需删除", uid)
	}

	// 3. 将修改后的 project 写回文件
	err := p.SaveProject()
	if err != nil {
		return fmt.Errorf("更新项目文件失败: %v", err)
	}

	return nil
}

func (p *RedcProject) AddCase(c *Case) error {
	p.Case = append(p.Case, *c)
	return nil
}

// SaveProject 将修改后的项目配置写回 JSON 文件
func (p *RedcProject) SaveProject() error {
	// 1. 确定文件路径（建议与 ProjectByName 逻辑保持一致）
	// 注意：ProjectPath, ProjectFile 应该是你全局定义的变量
	path := filepath.Join(ProjectPath, p.ProjectName, ProjectFile)

	// 2. 序列化数据
	// MarshalIndent 会生成带缩进的 JSON，方便人类阅读；如果追求体积小，可用 json.Marshal
	data, err := json.MarshalIndent(p, "", "    ")
	if err != nil {
		return fmt.Errorf("序列化失败: %v", err)
	}

	// 3. 写入文件
	// 0644 表示：所有者有读写权限，其他人只读
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

func ProjectConfigParse(path string) {
	cfg, err := ini.Load(path)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(3)
	}
	fmt.Println("项目名称:", cfg.Section("Global").Key("ProjectName").String())
	fmt.Println("项目路径:", cfg.Section("Global").Key("ProjectPath").String())
	fmt.Println("创建时间:", cfg.Section("Global").Key("CreateTime").String())
	fmt.Println("创建人员:", cfg.Section("Global").Key("Operator").String())
}
