package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	redc "red-cloud/mod"
	"red-cloud/mod/gologger"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx         context.Context
	project     *redc.RedcProject
	mu          sync.Mutex
	initError   string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Set default values (same as CLI defaults)
	if redc.Project == "" {
		redc.Project = "default"
	}
	if redc.U == "" {
		redc.U = "system" // Use "system" to match CLI default and bypass permission check
	}

	// Initialize config using same path detection as CLI
	if err := redc.LoadConfig(""); err != nil {
		a.initError = fmt.Sprintf("配置加载失败: %v", err)
		runtime.LogErrorf(ctx, a.initError)
		return
	}

	runtime.LogInfof(ctx, "配置加载成功 - RedcPath: %s, ProjectPath: %s, TemplateDir: %s", 
		redc.RedcPath, redc.ProjectPath, redc.TemplateDir)

	// Load default project
	if p, err := redc.ProjectParse(redc.Project, redc.U); err == nil {
		a.project = p
		runtime.LogInfof(ctx, "项目加载成功: %s", a.project.ProjectName)
	} else {
		a.initError = fmt.Sprintf("项目加载失败: %v", err)
		runtime.LogErrorf(ctx, a.initError)
	}
}

// emitLog sends a log message to the frontend
func (a *App) emitLog(message string) {
	runtime.EventsEmit(a.ctx, "log", message)
}

// emitRefresh notifies the frontend to refresh data
func (a *App) emitRefresh() {
	runtime.EventsEmit(a.ctx, "refresh", nil)
}

// createLogWriter creates an io.Writer that emits logs to the frontend
func (a *App) createLogWriter(prefix string) io.Writer {
	return gologger.NewEventWriter(a.emitLog, prefix)
}

// CaseInfo represents case information for frontend display
type CaseInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	State      string `json:"state"`
	StateTime  string `json:"stateTime"`
	CreateTime string `json:"createTime"`
	Operator   string `json:"operator"`
}

// ConfigInfo represents the configuration for frontend display
type ConfigInfo struct {
	RedcPath    string `json:"redcPath"`
	ProjectPath string `json:"projectPath"`
	HttpProxy   string `json:"httpProxy"`
	HttpsProxy  string `json:"httpsProxy"`
	NoProxy     string `json:"noProxy"`
}

// GetConfig returns current configuration
func (a *App) GetConfig() ConfigInfo {
	return ConfigInfo{
		RedcPath:    redc.RedcPath,
		ProjectPath: redc.ProjectPath,
		HttpProxy:   os.Getenv("HTTP_PROXY"),
		HttpsProxy:  os.Getenv("HTTPS_PROXY"),
		NoProxy:     os.Getenv("NO_PROXY"),
	}
}

// SaveProxyConfig saves proxy configuration to environment variables
func (a *App) SaveProxyConfig(httpProxy, httpsProxy, noProxy string) error {
	// Set environment variables for current process
	if httpProxy != "" {
		os.Setenv("HTTP_PROXY", httpProxy)
		os.Setenv("http_proxy", httpProxy)
	} else {
		os.Unsetenv("HTTP_PROXY")
		os.Unsetenv("http_proxy")
	}

	if httpsProxy != "" {
		os.Setenv("HTTPS_PROXY", httpsProxy)
		os.Setenv("https_proxy", httpsProxy)
	} else {
		os.Unsetenv("HTTPS_PROXY")
		os.Unsetenv("https_proxy")
	}

	if noProxy != "" {
		os.Setenv("NO_PROXY", noProxy)
		os.Setenv("no_proxy", noProxy)
	} else {
		os.Unsetenv("NO_PROXY")
		os.Unsetenv("no_proxy")
	}

	a.emitLog(fmt.Sprintf("代理配置已更新 - HTTP: %s, HTTPS: %s, NO_PROXY: %s", httpProxy, httpsProxy, noProxy))
	return nil
}

// ListCases returns all cases for the current project
func (a *App) ListCases() ([]CaseInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.project == nil {
		if a.initError != "" {
			return nil, fmt.Errorf(a.initError)
		}
		return nil, fmt.Errorf("项目未加载")
	}

	cases, err := redc.LoadProjectCases(a.project.ProjectName)
	if err != nil {
		return nil, err
	}

	result := make([]CaseInfo, 0, len(cases))
	for _, c := range cases {
		result = append(result, CaseInfo{
			ID:         c.Id,
			Name:       c.Name,
			Type:       c.Type,
			State:      c.State,
			StateTime:  c.StateTime,
			CreateTime: c.CreateTime,
			Operator:   c.Operator,
		})
	}
	return result, nil
}

// TemplateInfo represents template information for frontend display
type TemplateInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	User        string `json:"user"`
	Module      string `json:"module"`
}

// ListTemplates returns available templates
func (a *App) ListTemplates() ([]TemplateInfo, error) {
	templates, err := redc.ListLocalTemplates()
	if err != nil {
		return nil, err
	}
	result := make([]TemplateInfo, 0, len(templates))
	for _, t := range templates {
		result = append(result, TemplateInfo{
			Name:        t.Name,
			Description: t.Description,
			Version:     t.Version,
			User:        t.User,
			Module:      t.RedcModule,
		})
	}
	return result, nil
}

// TemplateVariable represents a variable definition from terraform
type TemplateVariable struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Description  string `json:"description"`
	DefaultValue string `json:"defaultValue"`
	Required     bool   `json:"required"`
}

// GetTemplateVariables parses variables.tf and terraform.tfvars to get template variables
func (a *App) GetTemplateVariables(templateName string) ([]TemplateVariable, error) {
	templatePath := filepath.Join(redc.TemplateDir, templateName)
	
	// Parse variables.tf to get variable definitions
	variablesFile := filepath.Join(templatePath, "variables.tf")
	tfvarsFile := filepath.Join(templatePath, "terraform.tfvars")
	
	variables := make(map[string]*TemplateVariable)
	
	// Parse variables.tf
	if _, err := os.Stat(variablesFile); err == nil {
		vars, err := parseVariablesTf(variablesFile)
		if err != nil {
			return nil, fmt.Errorf("解析 variables.tf 失败: %v", err)
		}
		for _, v := range vars {
			variables[v.Name] = v
		}
	}
	
	// Parse terraform.tfvars for default values
	if _, err := os.Stat(tfvarsFile); err == nil {
		defaults, err := parseTfvars(tfvarsFile)
		if err != nil {
			return nil, fmt.Errorf("解析 terraform.tfvars 失败: %v", err)
		}
		for name, value := range defaults {
			if v, ok := variables[name]; ok {
				v.DefaultValue = value
				v.Required = false
			}
		}
	}
	
	// Convert map to slice
	result := make([]TemplateVariable, 0, len(variables))
	for _, v := range variables {
		result = append(result, *v)
	}
	return result, nil
}

// parseVariablesTf parses a variables.tf file
func parseVariablesTf(filePath string) ([]*TemplateVariable, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var variables []*TemplateVariable
	scanner := bufio.NewScanner(file)
	
	varNameRegex := regexp.MustCompile(`^variable\s+"([^"]+)"`)
	typeRegex := regexp.MustCompile(`^\s*type\s*=\s*(.+)`)
	descRegex := regexp.MustCompile(`^\s*description\s*=\s*"([^"]*)"`)
	defaultRegex := regexp.MustCompile(`^\s*default\s*=\s*(.+)`)
	
	var currentVar *TemplateVariable
	braceCount := 0
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Check for variable declaration
		if matches := varNameRegex.FindStringSubmatch(line); len(matches) > 1 {
			if currentVar != nil {
				variables = append(variables, currentVar)
			}
			currentVar = &TemplateVariable{
				Name:     matches[1],
				Required: true,
				Type:     "string",
			}
			braceCount = 1
			continue
		}
		
		if currentVar == nil {
			continue
		}
		
		// Count braces
		braceCount += strings.Count(line, "{") - strings.Count(line, "}")
		
		// Parse type
		if matches := typeRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentVar.Type = strings.TrimSpace(matches[1])
		}
		
		// Parse description
		if matches := descRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentVar.Description = matches[1]
		}
		
		// Parse default
		if matches := defaultRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentVar.DefaultValue = strings.Trim(strings.TrimSpace(matches[1]), `"`)
			currentVar.Required = false
		}
		
		// End of variable block
		if braceCount <= 0 && currentVar != nil {
			variables = append(variables, currentVar)
			currentVar = nil
		}
	}
	
	// Add last variable if exists
	if currentVar != nil {
		variables = append(variables, currentVar)
	}
	
	return variables, scanner.Err()
}

// parseTfvars parses a terraform.tfvars file
func parseTfvars(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	defaults := make(map[string]string)
	scanner := bufio.NewScanner(file)
	
	// Pattern: name = "value" or name = value
	lineRegex := regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*"?([^"]*)"?`)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		if matches := lineRegex.FindStringSubmatch(line); len(matches) > 2 {
			defaults[matches[1]] = matches[2]
		}
	}
	
	return defaults, scanner.Err()
}

// StartCase starts a case by ID
func (a *App) StartCase(caseID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.project == nil {
		return fmt.Errorf("项目未加载")
	}

	c, err := a.project.GetCase(caseID)
	if err != nil {
		return fmt.Errorf("获取场景失败: %v", err)
	}

	if c == nil {
		return fmt.Errorf("场景对象为 nil")
	}

	// Validate case before starting
	if c.Path == "" {
		return fmt.Errorf("场景路径为空")
	}

	caseName := c.Name
	casePath := c.Path
	caseState := c.State

	a.emitLog(fmt.Sprintf("准备启动场景: %s, 路径: %s, 当前状态: %s", caseName, casePath, caseState))

	// Run in goroutine to avoid blocking GUI
	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.emitLog(fmt.Sprintf("启动场景时发生错误: %v", r))
			}
			a.emitRefresh() // 操作完成后刷新仪表盘
		}()
		
		a.emitLog(fmt.Sprintf("正在启动场景: %s", caseName))
		if err := c.TfApply(); err != nil {
			a.emitLog(fmt.Sprintf("启动失败: %v", err))
			return
		}
		a.emitLog(fmt.Sprintf("场景启动成功: %s", caseName))
		
		// 获取并显示 outputs
		if outputs, err := c.TfOutput(); err == nil {
			for name, meta := range outputs {
				a.emitLog(fmt.Sprintf("  %s = %s", name, string(meta.Value)))
			}
		}
	}()

	return nil
}

// StopCase stops a case by ID
func (a *App) StopCase(caseID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.project == nil {
		return fmt.Errorf("项目未加载")
	}

	c, err := a.project.GetCase(caseID)
	if err != nil {
		return err
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.emitLog(fmt.Sprintf("停止场景时发生错误: %v", r))
			}
			a.emitRefresh() // 操作完成后刷新仪表盘
		}()
		
		a.emitLog(fmt.Sprintf("正在停止场景: %s", c.Name))
		if err := c.Stop(); err != nil {
			a.emitLog(fmt.Sprintf("停止失败: %v", err))
			return
		}
		a.emitLog(fmt.Sprintf("场景停止成功: %s", c.Name))
	}()

	return nil
}

// RemoveCase removes a case by ID
func (a *App) RemoveCase(caseID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.project == nil {
		return fmt.Errorf("项目未加载")
	}

	c, err := a.project.GetCase(caseID)
	if err != nil {
		return err
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.emitLog(fmt.Sprintf("删除场景时发生错误: %v", r))
			}
			a.emitRefresh() // 操作完成后刷新仪表盘
		}()
		
		a.emitLog(fmt.Sprintf("正在删除场景: %s", c.Name))
		if err := c.Remove(); err != nil {
			a.emitLog(fmt.Sprintf("删除失败: %v", err))
			return
		}
		a.emitLog(fmt.Sprintf("场景删除成功: %s", c.Name))
	}()

	return nil
}

// CreateCase creates a new case from a template (async)
func (a *App) CreateCase(templateName string, name string, vars map[string]string) error {
	a.mu.Lock()
	if a.project == nil {
		a.mu.Unlock()
		return fmt.Errorf("项目未加载")
	}
	project := a.project
	a.mu.Unlock()

	a.emitLog(fmt.Sprintf("正在创建场景: %s (模板: %s)", name, templateName))

	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.emitLog(fmt.Sprintf("创建场景时发生错误: %v", r))
			}
			a.emitRefresh()
		}()

		c, err := project.CaseCreate(templateName, redc.U, name, vars)
		if err != nil {
			a.emitLog(fmt.Sprintf("场景创建失败: %v", err))
			return
		}
		a.emitLog(fmt.Sprintf("场景创建成功: %s (%s)", c.Name, c.GetId()))
	}()

	return nil
}

// CreateAndRunCase creates a new case and immediately starts it (like CLI "run" command)
func (a *App) CreateAndRunCase(templateName string, name string, vars map[string]string) error {
	a.mu.Lock()
	if a.project == nil {
		a.mu.Unlock()
		return fmt.Errorf("项目未加载")
	}
	project := a.project
	a.mu.Unlock()

	a.emitLog(fmt.Sprintf("正在创建并运行场景: %s (模板: %s)", name, templateName))

	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.emitLog(fmt.Sprintf("创建场景时发生错误: %v", r))
			}
			a.emitRefresh()
		}()

		// Step 1: Create the case (same as planLogic in CLI)
		c, err := project.CaseCreate(templateName, redc.U, name, vars)
		if err != nil {
			a.emitLog(fmt.Sprintf("场景创建失败: %v", err))
			return
		}
		a.emitLog(fmt.Sprintf("场景创建成功: %s (%s)", c.Name, c.GetId()))

		// Step 2: Start the case immediately (same as runCmd in CLI)
		a.emitLog(fmt.Sprintf("正在启动场景: %s", c.Name))

		// Run terraform apply
		if err := c.TfApply(); err != nil {
			a.emitLog(fmt.Sprintf("启动失败: %v", err))
			return
		}

		a.emitLog(fmt.Sprintf("场景启动成功: %s", c.Name))

		// Get and display outputs
		if outputs, err := c.TfOutput(); err == nil {
			for key, meta := range outputs {
				a.emitLog(fmt.Sprintf("%s = %s", key, string(meta.Value)))
			}
		}

		a.emitRefresh()
	}()

	return nil
}

// DeployCase creates and immediately starts a case (deprecated - use CreateCase then StartCase)
func (a *App) DeployCase(templateName string, name string, vars map[string]string) error {
	// CreateCase is now async, so this method just creates the case
	// User should manually start it after creation
	return a.CreateCase(templateName, name, vars)
}

// GetCaseOutputs returns the terraform outputs for a case
func (a *App) GetCaseOutputs(caseID string) (map[string]string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.project == nil {
		return nil, fmt.Errorf("项目未加载")
	}

	c, err := a.project.GetCase(caseID)
	if err != nil {
		return nil, err
	}

	// Only get outputs for running cases
	if c.State != "running" {
		return nil, nil
	}

	outputs, err := c.TfOutput()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for name, meta := range outputs {
		// Remove quotes from JSON string values
		value := string(meta.Value)
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		result[name] = value
	}
	return result, nil
}

// RegistryTemplate represents a template from the remote registry
type RegistryTemplate struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Latest      string   `json:"latest"`
	Versions    []string `json:"versions"`
	UpdatedAt   string   `json:"updatedAt"`
	Tags        []string `json:"tags"`
	Installed   bool     `json:"installed"`
	LocalVer    string   `json:"localVersion"`
}

// remoteTemplateInfo matches a single template in the registry
type remoteTemplateInfo struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Slug     string `json:"slug"`
	Latest   string `json:"latest"`
	Versions map[string]struct {
		URL       string `json:"url"`
		SHA256    string `json:"sha256"`
		UpdatedAt string `json:"updated_at"`
	} `json:"versions"`
	Metadata struct {
		Name        string `json:"name"`
		Author      string `json:"author"`
		Description string `json:"description"`
		Readme      string `json:"readme"`
	} `json:"metadata"`
}

// remoteIndexResponse matches the index.json structure from the registry
type remoteIndexResponse struct {
	UpdatedAt string                        `json:"updated_at"`
	RepoName  string                        `json:"repo_name"`
	Templates map[string]remoteTemplateInfo `json:"templates"`
}

// FetchRegistryTemplates fetches templates from the remote registry
func (a *App) FetchRegistryTemplates(registryURL string) ([]RegistryTemplate, error) {
	if registryURL == "" {
		registryURL = "https://redc.wgpsec.org"
	}

	a.emitLog(fmt.Sprintf("正在连接仓库: %s", registryURL))

	// Fetch index.json
	indexURL := fmt.Sprintf("%s/index.json?t=%d", registryURL, time.Now().Unix())
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(indexURL)
	if err != nil {
		return nil, fmt.Errorf("连接仓库失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("仓库返回错误: %s", resp.Status)
	}

	var idx remoteIndexResponse
	if err := json.NewDecoder(resp.Body).Decode(&idx); err != nil {
		return nil, fmt.Errorf("解析仓库索引失败: %v", err)
	}

	// Build result list
	result := make([]RegistryTemplate, 0, len(idx.Templates))
	for templateID, t := range idx.Templates {
		// Use templateID (e.g. "aliyun/ecs") as the name
		name := templateID
		if name == "" {
			name = t.ID
		}
		
		// Check if installed locally
		installed, localVer, _ := redc.CheckLocalImage(name)
		
		// Get version list
		versions := make([]string, 0, len(t.Versions))
		var updatedAt string
		for v, info := range t.Versions {
			versions = append(versions, v)
			if v == t.Latest && info.UpdatedAt != "" {
				updatedAt = info.UpdatedAt
			}
		}

		// Extract tags from provider
		var tags []string
		if t.Provider != "" {
			tags = []string{t.Provider}
		}

		result = append(result, RegistryTemplate{
			Name:        name,
			Description: t.Metadata.Description,
			Author:      t.Metadata.Author,
			Latest:      t.Latest,
			Versions:    versions,
			UpdatedAt:   updatedAt,
			Tags:        tags,
			Installed:   installed,
			LocalVer:    localVer,
		})
	}

	a.emitLog(fmt.Sprintf("已获取 %d 个模板", len(result)))
	return result, nil
}

// PullTemplate pulls a template from the registry
func (a *App) PullTemplate(templateName string, force bool) error {
	a.emitLog(fmt.Sprintf("正在拉取模板: %s", templateName))

	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.emitLog(fmt.Sprintf("拉取模板时发生错误: %v", r))
			}
			a.emitRefresh()
		}()

		opts := redc.PullOptions{
			RegistryURL: "https://redc.wgpsec.org",
			Force:       force,
			Timeout:     120 * time.Second,
		}

		if err := redc.Pull(context.Background(), templateName, opts); err != nil {
			a.emitLog(fmt.Sprintf("拉取失败: %v", err))
			return
		}

		a.emitLog(fmt.Sprintf("模板拉取成功: %s", templateName))
	}()

	return nil
}
