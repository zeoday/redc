package mod

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestE2E_CompleteDeploymentFlow tests the complete deployment flow from configuration to deployment
func TestE2E_CompleteDeploymentFlow(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	configDir := filepath.Join(tmpDir, "configs")
	projectDir := filepath.Join(tmpDir, "project")

	// Create directories
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("Failed to create template dir: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create a test base template
	testTemplateName := "test-base-template"
	testTemplateDir := filepath.Join(templateDir, testTemplateName)
	if err := os.MkdirAll(testTemplateDir, 0755); err != nil {
		t.Fatalf("Failed to create test template dir: %v", err)
	}

	// Create case.json
	caseJSON := map[string]interface{}{
		"name":                testTemplateName,
		"description":         "Test base template",
		"is_base_template":    true,
		"supported_providers": []string{"alicloud", "tencentcloud"},
	}
	caseData, _ := json.Marshal(caseJSON)
	if err := os.WriteFile(filepath.Join(testTemplateDir, "case.json"), caseData, 0644); err != nil {
		t.Fatalf("Failed to create case.json: %v", err)
	}

	// Create variables.tf
	variablesTF := `variable "provider" {
  description = "Cloud provider"
  type        = string
}

variable "region" {
  description = "Region"
  type        = string
}

variable "instance_type" {
  description = "Instance type"
  type        = string
}

variable "userdata" {
  description = "Userdata script"
  type        = string
  default     = ""
}
`
	if err := os.WriteFile(filepath.Join(testTemplateDir, "variables.tf"), []byte(variablesTF), 0644); err != nil {
		t.Fatalf("Failed to create variables.tf: %v", err)
	}

	// Initialize services
	templateMgr := NewTemplateManager()
	templateMgr.templateDir = templateDir
	validator := &ConfigValidator{templateMgr: templateMgr}
	configStore := &ConfigStore{configDir: configDir}

	// Step 1: Get base templates
	t.Run("Step1_GetBaseTemplates", func(t *testing.T) {
		templates, err := templateMgr.ScanBaseTemplates()
		if err != nil {
			t.Fatalf("Failed to get base templates: %v", err)
		}
		if len(templates) == 0 {
			t.Fatal("Expected at least one base template")
		}
		found := false
		for _, tmpl := range templates {
			if tmpl.Name == testTemplateName {
				found = true
				if !tmpl.IsBase {
					t.Error("Template should be marked as base template")
				}
				if len(tmpl.Providers) != 2 {
					t.Errorf("Expected 2 providers, got %d", len(tmpl.Providers))
				}
			}
		}
		if !found {
			t.Error("Test template not found in base templates")
		}
	})

	// Step 2: Get template variables
	t.Run("Step2_GetTemplateVariables", func(t *testing.T) {
		vars, err := templateMgr.GetTemplateVariables(testTemplateName)
		if err != nil {
			t.Fatalf("Failed to get template variables: %v", err)
		}
		if len(vars) == 0 {
			t.Fatal("Expected template variables")
		}
		// Verify we have the expected variables
		expectedVars := map[string]bool{
			"provider":      false,
			"region":        false,
			"instance_type": false,
			"userdata":      false,
		}
		for _, v := range vars {
			if _, exists := expectedVars[v.Name]; exists {
				expectedVars[v.Name] = true
			}
		}
		for varName, found := range expectedVars {
			if !found {
				t.Errorf("Expected variable %s not found", varName)
			}
		}
	})

	// Step 3: Create deployment configuration
	config := &DeploymentConfig{
		Name:         "test-deployment",
		TemplateName: testTemplateName,
		Provider:     "alicloud",
		Region:       "cn-hangzhou",
		InstanceType: "ecs.t6-c1m1.large",
		Userdata:     "#!/bin/bash\necho 'Hello World'",
		Variables: map[string]string{
			"provider":      "alicloud",
			"region":        "cn-hangzhou",
			"instance_type": "ecs.t6-c1m1.large",
			"userdata":      "#!/bin/bash\necho 'Hello World'",
		},
	}

	// Step 4: Validate configuration
	t.Run("Step4_ValidateConfiguration", func(t *testing.T) {
		result, err := validator.ValidateDeploymentConfig(config)
		if err != nil {
			t.Fatalf("Validation failed: %v", err)
		}
		if !result.Valid {
			t.Errorf("Configuration should be valid, errors: %v", result.Errors)
		}
	})

	// Step 5: Save configuration as template
	configTemplateName := "test-config-template"
	t.Run("Step5_SaveConfigTemplate", func(t *testing.T) {
		err := configStore.SaveConfigTemplate(configTemplateName, config)
		if err != nil {
			t.Fatalf("Failed to save config template: %v", err)
		}
	})

	// Step 6: Load configuration template
	t.Run("Step6_LoadConfigTemplate", func(t *testing.T) {
		loadedConfig, err := configStore.LoadConfigTemplate(configTemplateName)
		if err != nil {
			t.Fatalf("Failed to load config template: %v", err)
		}
		if loadedConfig.Name != config.Name {
			t.Errorf("Expected name %s, got %s", config.Name, loadedConfig.Name)
		}
		if loadedConfig.Provider != config.Provider {
			t.Errorf("Expected provider %s, got %s", config.Provider, loadedConfig.Provider)
		}
	})

	// Step 7: List configuration templates
	t.Run("Step7_ListConfigTemplates", func(t *testing.T) {
		templates, err := configStore.ListConfigTemplates()
		if err != nil {
			t.Fatalf("Failed to list config templates: %v", err)
		}
		found := false
		for _, tmpl := range templates {
			if tmpl == configTemplateName {
				found = true
			}
		}
		if !found {
			t.Error("Saved config template not found in list")
		}
	})

	// Step 8: Export configuration template
	exportPath := filepath.Join(tmpDir, "exported-config.json")
	t.Run("Step8_ExportConfigTemplate", func(t *testing.T) {
		err := configStore.ExportConfigTemplate(configTemplateName, exportPath)
		if err != nil {
			t.Fatalf("Failed to export config template: %v", err)
		}
		if _, err := os.Stat(exportPath); os.IsNotExist(err) {
			t.Error("Exported file does not exist")
		}
	})

	// Step 9: Import configuration template
	importedTemplateName := "imported-config-template"
	t.Run("Step9_ImportConfigTemplate", func(t *testing.T) {
		err := configStore.ImportConfigTemplate(importedTemplateName, exportPath)
		if err != nil {
			t.Fatalf("Failed to import config template: %v", err)
		}
		// Verify imported template
		loadedConfig, err := configStore.LoadConfigTemplate(importedTemplateName)
		if err != nil {
			t.Fatalf("Failed to load imported config template: %v", err)
		}
		if loadedConfig.Provider != config.Provider {
			t.Errorf("Imported config provider mismatch: expected %s, got %s", config.Provider, loadedConfig.Provider)
		}
	})

	// Step 10: Delete configuration template
	t.Run("Step10_DeleteConfigTemplate", func(t *testing.T) {
		err := configStore.DeleteConfigTemplate(configTemplateName)
		if err != nil {
			t.Fatalf("Failed to delete config template: %v", err)
		}
		// Verify deletion
		_, err = configStore.LoadConfigTemplate(configTemplateName)
		if err == nil {
			t.Error("Expected error when loading deleted template")
		}
	})
}

// TestE2E_MultiCloudDeployment tests deployment across different cloud providers
func TestE2E_MultiCloudDeployment(t *testing.T) {
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	configDir := filepath.Join(tmpDir, "configs")

	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("Failed to create template dir: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create multi-cloud template
	testTemplateName := "multi-cloud-template"
	testTemplateDir := filepath.Join(templateDir, testTemplateName)
	if err := os.MkdirAll(testTemplateDir, 0755); err != nil {
		t.Fatalf("Failed to create test template dir: %v", err)
	}

	caseJSON := map[string]interface{}{
		"name":                testTemplateName,
		"description":         "Multi-cloud template",
		"is_base_template":    true,
		"supported_providers": []string{"alicloud", "tencentcloud", "aws"},
	}
	caseData, _ := json.Marshal(caseJSON)
	if err := os.WriteFile(filepath.Join(testTemplateDir, "case.json"), caseData, 0644); err != nil {
		t.Fatalf("Failed to create case.json: %v", err)
	}

	variablesTF := `variable "provider" {
  type = string
}
variable "region" {
  type = string
}
variable "instance_type" {
  type = string
}
`
	if err := os.WriteFile(filepath.Join(testTemplateDir, "variables.tf"), []byte(variablesTF), 0644); err != nil {
		t.Fatalf("Failed to create variables.tf: %v", err)
	}

	templateMgr := NewTemplateManager()
	templateMgr.templateDir = templateDir
	validator := &ConfigValidator{templateMgr: templateMgr}
	configStore := &ConfigStore{configDir: configDir}

	// Test configurations for different providers
	providers := []struct {
		provider     string
		region       string
		instanceType string
	}{
		{"alicloud", "cn-hangzhou", "ecs.t6-c1m1.large"},
		{"tencentcloud", "ap-guangzhou", "S5.MEDIUM2"},
		{"aws", "us-east-1", "t3.micro"},
	}

	for _, p := range providers {
		t.Run("Provider_"+p.provider, func(t *testing.T) {
			config := &DeploymentConfig{
				Name:         "test-" + p.provider,
				TemplateName: testTemplateName,
				Provider:     p.provider,
				Region:       p.region,
				InstanceType: p.instanceType,
				Variables: map[string]string{
					"provider":      p.provider,
					"region":        p.region,
					"instance_type": p.instanceType,
				},
			}

			// Validate configuration
			result, err := validator.ValidateDeploymentConfig(config)
			if err != nil {
				t.Fatalf("Validation failed for %s: %v", p.provider, err)
			}
			if !result.Valid {
				t.Errorf("Configuration for %s should be valid, errors: %v", p.provider, result.Errors)
			}

			// Save configuration
			err = configStore.SaveConfigTemplate("config-"+p.provider, config)
			if err != nil {
				t.Fatalf("Failed to save config for %s: %v", p.provider, err)
			}
		})
	}
}

// TestE2E_ErrorHandlingAndRecovery tests error handling throughout the deployment flow
func TestE2E_ErrorHandlingAndRecovery(t *testing.T) {
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	configDir := filepath.Join(tmpDir, "configs")

	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("Failed to create template dir: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	templateMgr := NewTemplateManager()
	templateMgr.templateDir = templateDir
	validator := &ConfigValidator{templateMgr: templateMgr}
	configStore := &ConfigStore{configDir: configDir}

	// Test 1: Invalid template name
	t.Run("Error_InvalidTemplateName", func(t *testing.T) {
		_, err := templateMgr.GetTemplateVariables("non-existent-template")
		if err == nil {
			t.Error("Expected error for non-existent template")
		}
	})

	// Test 2: Missing required fields
	t.Run("Error_MissingRequiredFields", func(t *testing.T) {
		config := &DeploymentConfig{
			Name:         "",
			TemplateName: "",
			Provider:     "",
			Region:       "",
			InstanceType: "",
		}
		result, err := validator.ValidateDeploymentConfig(config)
		// Validation will fail because template doesn't exist
		if err == nil {
			// If no error, check that validation failed
			if result.Valid {
				t.Error("Configuration with missing fields should be invalid")
			}
			if len(result.Errors) == 0 {
				t.Error("Expected validation errors for missing fields")
			}
		}
		// If error occurred (template not found), that's also acceptable
	})

	// Test 3: Invalid provider (using ValidateProvider directly)
	t.Run("Error_InvalidProvider", func(t *testing.T) {
		err := validator.ValidateProvider("invalid-provider")
		if err == nil {
			t.Error("Expected error when validating invalid provider")
		}
	})

	// Test 4: Load non-existent config template
	t.Run("Error_LoadNonExistentConfig", func(t *testing.T) {
		_, err := configStore.LoadConfigTemplate("non-existent-config")
		if err == nil {
			t.Error("Expected error when loading non-existent config")
		}
	})

	// Test 5: Delete non-existent config template
	t.Run("Error_DeleteNonExistentConfig", func(t *testing.T) {
		err := configStore.DeleteConfigTemplate("non-existent-config")
		if err == nil {
			t.Error("Expected error when deleting non-existent config")
		}
	})

	// Test 6: Import invalid file
	t.Run("Error_ImportInvalidFile", func(t *testing.T) {
		invalidFile := filepath.Join(tmpDir, "invalid.json")
		if err := os.WriteFile(invalidFile, []byte("invalid json"), 0644); err != nil {
			t.Fatalf("Failed to create invalid file: %v", err)
		}
		err := configStore.ImportConfigTemplate("test", invalidFile)
		if err == nil {
			t.Error("Expected error when importing invalid file")
		}
	})

	// Test 7: Export non-existent config
	t.Run("Error_ExportNonExistentConfig", func(t *testing.T) {
		exportPath := filepath.Join(tmpDir, "export.json")
		err := configStore.ExportConfigTemplate("non-existent-config", exportPath)
		if err == nil {
			t.Error("Expected error when exporting non-existent config")
		}
	})
}

// TestE2E_ConcurrentOperations tests concurrent operations on the deployment system
func TestE2E_ConcurrentOperations(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "configs")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configStore := &ConfigStore{configDir: configDir}

	// Create multiple configurations concurrently
	numConfigs := 10
	done := make(chan bool, numConfigs)

	for i := 0; i < numConfigs; i++ {
		go func(index int) {
			config := &DeploymentConfig{
				Name:         "concurrent-test",
				TemplateName: "test-template",
				Provider:     "alicloud",
				Region:       "cn-hangzhou",
				InstanceType: "ecs.t6-c1m1.large",
			}
			configName := "config-" + string(rune('0'+index))
			err := configStore.SaveConfigTemplate(configName, config)
			if err != nil {
				t.Errorf("Failed to save config %s: %v", configName, err)
			}
			done <- true
		}(i)
	}

	// Wait for all operations to complete
	timeout := time.After(5 * time.Second)
	for i := 0; i < numConfigs; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}

	// Verify all configurations were saved
	templates, err := configStore.ListConfigTemplates()
	if err != nil {
		t.Fatalf("Failed to list config templates: %v", err)
	}
	if len(templates) != numConfigs {
		t.Errorf("Expected %d config templates, got %d", numConfigs, len(templates))
	}
}
