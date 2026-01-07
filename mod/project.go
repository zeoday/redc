package mod

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"red-cloud/mod/gologger"
	"text/tabwriter"
	"time"
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
	Id         string   `json:"id"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Operator   string   `json:"operator"`
	Path       string   `json:"path"`
	Node       int      `json:"node"`
	CreateTime string   `json:"create_time"`
	Parameter  []string `json:"parameter"`
}

// CaseScene 场景参数判断
func CaseScene(t string) ([]string, error) {
	var par []string
	switch t {
	case "cs-49", "c2-new", "snowc2":
		par = RVar(
			fmt.Sprintf("node_count=%d", Node),
			fmt.Sprintf("domain=%s", Domain),
		)
	case "aws-proxy", "aliyun-proxy", "asm":
		par = RVar(fmt.Sprintf("node_count=%d", Node))
	case "dnslog", "xraydnslog", "interactsh":
		if Domain == "360.com" {
			return par, fmt.Errorf("创建 dnslog 时,域名不可为默认值")
		}
		par = RVar(fmt.Sprintf("domain=%s", Domain))
	case "pss5", "frp", "frp-loki", "nps":
		par = []string{fmt.Sprintf("base64_command=%s", Base64Command)}
	case "asm-node":
		par = RVar(
			fmt.Sprintf("node_count=%d", Node),
			fmt.Sprintf("domain2=%s", Domain2),
			fmt.Sprintf("doamin=%s", Domain),
		)
	}
	return par, nil
}

// NewProjectConfig 创建项目配置文件
func NewProjectConfig(name string, user string) (*RedcProject, error) {
	path := filepath.Join(ProjectPath, name)
	// 创建项目目录
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("创建项目目录失败: %w", err)
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

	if err := project.SaveProject(); err != nil {
		// 如果保存失败，应该清理目录吗？视业务逻辑而定，这里暂时只返回错误
		return nil, fmt.Errorf("保存项目状态文件失败: %w", err)
	}
	gologger.Info().Msgf("项目状态文件「%s」创建成功！", ProjectFile)
	return project, nil

}

func ProjectParse(name string, user string) (*RedcProject, error) {
	// 尝试直接读取项目
	if p, err := ProjectByName(name); err == nil {
		// 项目鉴权
		if p.User != user && user != "system" {
			return nil, fmt.Errorf("当前用户「%s」无权限访问项目「%s」", user, name)
		}
		// 读取成功，直接返回
		return p, nil
	}
	path := filepath.Join(ProjectPath, name)
	// 检查目录是否存在，或者直接尝试创建
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		gologger.Info().Msgf("项目不存在，正在创建新项目: %s", name)
		return NewProjectConfig(name, user)
	} else if statErr != nil {
		// 目录存在但有其他错误（如权限不足）
		return nil, statErr
	}
	return NewProjectConfig(name, user)
}

// ProjectByName 读取项目配置
func ProjectByName(name string) (*RedcProject, error) {
	path := filepath.Join(ProjectPath, name, ProjectFile)
	data, err := os.ReadFile(path)
	if err != nil {
		gologger.Debug().Msgf("读取项目文件失败 [%s]: %v", name, err)
		return nil, err
	}

	var project RedcProject
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("解析项目配置失败: %w", err)
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
func (p *RedcProject) HandleCase(c *Case) error {
	uid := c.Id
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

func (p *RedcProject) CaseList() {
	// 使用 tabwriter 创建表格输出
	w := tabwriter.NewWriter(os.Stdout, 15, 0, 1, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "UUID\tType\tName\tOperator\tCreateTime\t")

	// 遍历项目中的所有 Case
	for _, c := range p.Case {
		// 鉴权：只显示当前用户或 system 用户的 Case
		if c.Operator == U || U == "system" {
			// 从 Case 结构中获取 Type（从 Name 字段获取类型信息，或者需要额外的 Type 字段）
			caseType := c.Name // 假设 Name 包含类型信息，或者需要在 Case 结构中添加 Type 字段
			fmt.Fprintln(w, c.Id, "\t", caseType, "\t", c.Name, "\t", c.Operator, "\t", c.CreateTime)
		}
	}

	if err := w.Flush(); err != nil {
		gologger.Fatal().Msgf("表格输出失败:/%s", err.Error())
	}

}

// SaveProject 将修改后的项目配置写回 JSON 文件
func (p *RedcProject) SaveProject() error {
	dirPath := p.ProjectPath
	path := filepath.Join(ProjectPath, p.ProjectName, ProjectFile)
	// 2. 防御性编程：确保目录存在
	// 防止用户手动删除了目录，导致保存文件失败
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("无法恢复项目目录: %w", err)
	}
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
