package mod

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestIsBaseTemplate 测试 IsBaseTemplate 方法
func TestIsBaseTemplate(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	
	tests := []struct {
		name           string
		caseJSON       map[string]interface{}
		expectedResult bool
		expectError    bool
	}{
		{
			name: "基础模板",
			caseJSON: map[string]interface{}{
				"name":             "test-template",
				"description":      "Test template",
				"user":             "system",
				"version":          "1.0.0",
				"is_base_template": true,
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name: "非基础模板",
			caseJSON: map[string]interface{}{
				"name":             "test-template",
				"description":      "Test template",
				"user":             "system",
				"version":          "1.0.0",
				"is_base_template": false,
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name: "缺少 is_base_template 字段",
			caseJSON: map[string]interface{}{
				"name":        "test-template",
				"description": "Test template",
				"user":        "system",
				"version":     "1.0.0",
			},
			expectedResult: false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试模板目录
			templateDir := filepath.Join(tempDir, tt.name)
			if err := os.MkdirAll(templateDir, 0755); err != nil {
				t.Fatalf("Failed to create template directory: %v", err)
			}

			// 写入 case.json
			caseFilePath := filepath.Join(templateDir, TmplCaseFile)
			data, err := json.Marshal(tt.caseJSON)
			if err != nil {
				t.Fatalf("Failed to marshal case.json: %v", err)
			}
			if err := os.WriteFile(caseFilePath, data, 0644); err != nil {
				t.Fatalf("Failed to write case.json: %v", err)
			}

			// 创建 TemplateManager
			tm := &TemplateManager{templateDir: tempDir}

			// 测试 IsBaseTemplate
			result, err := tm.IsBaseTemplate(templateDir)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expectedResult {
					t.Errorf("Expected %v but got %v", tt.expectedResult, result)
				}
			}
		})
	}
}

// TestScanBaseTemplates 测试 ScanBaseTemplates 方法
func TestScanBaseTemplates(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试模板结构
	templates := []struct {
		path           string
		isBase         bool
		name           string
		description    string
		version        string
		providers      []string
	}{
		{
			path:        "provider1/template1",
			isBase:      true,
			name:        "template1",
			description: "Base template 1",
			version:     "1.0.0",
			providers:   []string{"alicloud", "tencentcloud"},
		},
		{
			path:        "provider1/template2",
			isBase:      false,
			name:        "template2",
			description: "Predefined template",
			version:     "1.0.0",
			providers:   []string{},
		},
		{
			path:        "provider2/template3",
			isBase:      true,
			name:        "template3",
			description: "Base template 2",
			version:     "2.0.0",
			providers:   []string{"aws", "volcengine"},
		},
	}

	for _, tmpl := range templates {
		templateDir := filepath.Join(tempDir, tmpl.path)
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			t.Fatalf("Failed to create template directory: %v", err)
		}

		caseData := map[string]interface{}{
			"name":               tmpl.name,
			"description":        tmpl.description,
			"user":               "system",
			"version":            tmpl.version,
			"is_base_template":   tmpl.isBase,
			"supported_providers": tmpl.providers,
		}

		caseFilePath := filepath.Join(templateDir, TmplCaseFile)
		data, err := json.Marshal(caseData)
		if err != nil {
			t.Fatalf("Failed to marshal case.json: %v", err)
		}
		if err := os.WriteFile(caseFilePath, data, 0644); err != nil {
			t.Fatalf("Failed to write case.json: %v", err)
		}
	}

	// 创建 TemplateManager
	tm := &TemplateManager{templateDir: tempDir}

	// 测试 ScanBaseTemplates
	baseTemplates, err := tm.ScanBaseTemplates()
	if err != nil {
		t.Fatalf("ScanBaseTemplates failed: %v", err)
	}

	// 验证结果
	expectedCount := 2 // 只有 template1 和 template3 是基础模板
	if len(baseTemplates) != expectedCount {
		t.Errorf("Expected %d base templates but got %d", expectedCount, len(baseTemplates))
	}

	// 验证基础模板的属性
	for _, tmpl := range baseTemplates {
		if !tmpl.IsBase {
			t.Errorf("Template %s should be a base template", tmpl.Name)
		}
		if tmpl.Name == "" {
			t.Errorf("Template name should not be empty")
		}
		if tmpl.Description == "" {
			t.Errorf("Template description should not be empty")
		}
		if tmpl.Version == "" {
			t.Errorf("Template version should not be empty")
		}
	}
}

// TestScanBaseTemplates_EmptyDirectory 测试空目录的情况
func TestScanBaseTemplates_EmptyDirectory(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建 TemplateManager
	tm := &TemplateManager{templateDir: tempDir}

	// 测试 ScanBaseTemplates
	baseTemplates, err := tm.ScanBaseTemplates()
	if err != nil {
		t.Fatalf("ScanBaseTemplates failed: %v", err)
	}

	// 验证结果
	if len(baseTemplates) != 0 {
		t.Errorf("Expected 0 base templates but got %d", len(baseTemplates))
	}
}

// TestScanBaseTemplates_NonExistentDirectory 测试不存在的目录
func TestScanBaseTemplates_NonExistentDirectory(t *testing.T) {
	// 使用不存在的目录
	nonExistentDir := filepath.Join(t.TempDir(), "nonexistent")

	// 创建 TemplateManager
	tm := &TemplateManager{templateDir: nonExistentDir}

	// 测试 ScanBaseTemplates
	baseTemplates, err := tm.ScanBaseTemplates()
	if err != nil {
		t.Fatalf("ScanBaseTemplates failed: %v", err)
	}

	// 验证结果 - 应该返回空列表而不是错误
	if baseTemplates != nil && len(baseTemplates) != 0 {
		t.Errorf("Expected nil or empty list but got %d templates", len(baseTemplates))
	}
}

// TestGetTemplateVariables 测试 GetTemplateVariables 方法
func TestGetTemplateVariables(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		variablesTf   string
		expectedVars  []TemplateVariable
		expectError   bool
	}{
		{
			name: "基本变量定义",
			variablesTf: `
variable "provider" {
  description = "云厂商"
  type        = string
}

variable "region" {
  description = "地域"
  type        = string
  default     = "cn-hangzhou"
}

variable "instance_type" {
  description = "实例规格"
  type        = string
}
`,
			expectedVars: []TemplateVariable{
				{
					Name:        "provider",
					Type:        "string",
					Description: "云厂商",
					Required:    true,
				},
				{
					Name:         "region",
					Type:         "string",
					Description:  "地域",
					Required:     false,
					DefaultValue: "cn-hangzhou",
				},
				{
					Name:        "instance_type",
					Type:        "string",
					Description: "实例规格",
					Required:    true,
				},
			},
			expectError: false,
		},
		{
			name: "带验证规则的变量",
			variablesTf: `
variable "provider" {
  description = "云厂商"
  type        = string
  validation {
    condition     = contains(["alicloud", "tencentcloud", "aws"], var.provider)
    error_message = "不支持的云厂商"
  }
}
`,
			expectedVars: []TemplateVariable{
				{
					Name:        "provider",
					Type:        "string",
					Description: "云厂商",
					Required:    true,
					Validation: &VariableValidation{
						AllowedValues: []string{"alicloud", "tencentcloud", "aws"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "不同类型的变量",
			variablesTf: `
variable "instance_count" {
  description = "实例数量"
  type        = number
  default     = 1
}

variable "enable_monitoring" {
  description = "启用监控"
  type        = bool
  default     = true
}
`,
			expectedVars: []TemplateVariable{
				{
					Name:         "instance_count",
					Type:         "number",
					Description:  "实例数量",
					Required:     false,
					DefaultValue: "1",
				},
				{
					Name:         "enable_monitoring",
					Type:         "bool",
					Description:  "启用监控",
					Required:     false,
					DefaultValue: "true",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试模板目录
			templateDir := filepath.Join(tempDir, tt.name)
			if err := os.MkdirAll(templateDir, 0755); err != nil {
				t.Fatalf("Failed to create template directory: %v", err)
			}

			// 写入 variables.tf
			variablesFilePath := filepath.Join(templateDir, "variables.tf")
			if err := os.WriteFile(variablesFilePath, []byte(tt.variablesTf), 0644); err != nil {
				t.Fatalf("Failed to write variables.tf: %v", err)
			}

			// 创建 TemplateManager
			tm := &TemplateManager{templateDir: tempDir}

			// 测试 GetTemplateVariables
			variables, err := tm.GetTemplateVariables(tt.name)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// 验证变量数量
			if len(variables) != len(tt.expectedVars) {
				t.Errorf("Expected %d variables but got %d", len(tt.expectedVars), len(variables))
			}

			// 验证每个变量的属性
			for i, expectedVar := range tt.expectedVars {
				if i >= len(variables) {
					break
				}
				actualVar := variables[i]

				if actualVar.Name != expectedVar.Name {
					t.Errorf("Variable %d: expected name %s but got %s", i, expectedVar.Name, actualVar.Name)
				}
				if actualVar.Type != expectedVar.Type {
					t.Errorf("Variable %s: expected type %s but got %s", actualVar.Name, expectedVar.Type, actualVar.Type)
				}
				if actualVar.Description != expectedVar.Description {
					t.Errorf("Variable %s: expected description %s but got %s", actualVar.Name, expectedVar.Description, actualVar.Description)
				}
				if actualVar.Required != expectedVar.Required {
					t.Errorf("Variable %s: expected required %v but got %v", actualVar.Name, expectedVar.Required, actualVar.Required)
				}
				if actualVar.DefaultValue != expectedVar.DefaultValue {
					t.Errorf("Variable %s: expected default value %s but got %s", actualVar.Name, expectedVar.DefaultValue, actualVar.DefaultValue)
				}

				// 验证验证规则
				if expectedVar.Validation != nil {
					if actualVar.Validation == nil {
						t.Errorf("Variable %s: expected validation rules but got none", actualVar.Name)
					} else {
						if len(actualVar.Validation.AllowedValues) != len(expectedVar.Validation.AllowedValues) {
							t.Errorf("Variable %s: expected %d allowed values but got %d", 
								actualVar.Name, len(expectedVar.Validation.AllowedValues), len(actualVar.Validation.AllowedValues))
						}
					}
				}
			}
		})
	}
}

// TestGetTemplateVariables_NoVariablesFile 测试没有 variables.tf 文件的情况
func TestGetTemplateVariables_NoVariablesFile(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试模板目录（但不创建 variables.tf）
	templateDir := filepath.Join(tempDir, "test-template")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("Failed to create template directory: %v", err)
	}

	// 创建 TemplateManager
	tm := &TemplateManager{templateDir: tempDir}

	// 测试 GetTemplateVariables
	variables, err := tm.GetTemplateVariables("test-template")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 验证结果 - 应该返回空列表
	if len(variables) != 0 {
		t.Errorf("Expected 0 variables but got %d", len(variables))
	}
}

// TestGetTemplateVariables_NonExistentTemplate 测试不存在的模板
func TestGetTemplateVariables_NonExistentTemplate(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建 TemplateManager
	tm := &TemplateManager{templateDir: tempDir}

	// 测试 GetTemplateVariables
	_, err := tm.GetTemplateVariables("nonexistent-template")
	if err == nil {
		t.Errorf("Expected error for nonexistent template but got none")
	}
}

// TestGetTemplateVariables_InvalidHCL 测试无效的 HCL 语法
func TestGetTemplateVariables_InvalidHCL(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试模板目录
	templateDir := filepath.Join(tempDir, "invalid-template")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("Failed to create template directory: %v", err)
	}

	// 写入无效的 variables.tf
	variablesFilePath := filepath.Join(templateDir, "variables.tf")
	invalidHCL := `
variable "test" {
  description = "Test variable"
  type = string
  # 缺少闭合括号
`
	if err := os.WriteFile(variablesFilePath, []byte(invalidHCL), 0644); err != nil {
		t.Fatalf("Failed to write variables.tf: %v", err)
	}

	// 创建 TemplateManager
	tm := &TemplateManager{templateDir: tempDir}

	// 测试 GetTemplateVariables
	_, err := tm.GetTemplateVariables("invalid-template")
	if err == nil {
		t.Errorf("Expected error for invalid HCL but got none")
	}
}
