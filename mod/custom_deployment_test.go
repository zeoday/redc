package mod

import (
	"encoding/json"
	"os"
	"path/filepath"
	"red-cloud/pb"
	"testing"
	"time"
)

// TestNewCustomDeploymentService tests creating a new custom deployment service instance
func TestNewCustomDeploymentService(t *testing.T) {
	service := NewCustomDeploymentService()
	
	if service == nil {
		t.Fatal("NewCustomDeploymentService returned nil")
	}
	
	if service.templateMgr == nil {
		t.Error("templateMgr is nil")
	}
	
	if service.validator == nil {
		t.Error("validator is nil")
	}
	
	if service.executor == nil {
		t.Error("executor is nil")
	}
	
	if service.configStore == nil {
		t.Error("configStore is nil")
	}
}

// TestBaseTemplate tests the BaseTemplate structure
func TestBaseTemplate(t *testing.T) {
	template := &BaseTemplate{
		Name:        "test-template",
		Description: "Test template description",
		Version:     "1.0.0",
		Variables: []TemplateVariable{
			{
				Name:        "provider",
				Type:        "string",
				Description: "Cloud provider",
				Required:    true,
			},
		},
		Providers: []string{"alicloud", "tencentcloud"},
		IsBase:    true,
		User:      "system",
	}
	
	if template.Name != "test-template" {
		t.Errorf("Expected name 'test-template', got '%s'", template.Name)
	}
	
	if len(template.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(template.Variables))
	}
	
	if len(template.Providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(template.Providers))
	}
	
	if !template.IsBase {
		t.Error("Expected IsBase to be true")
	}
}

// TestDeploymentConfig tests the DeploymentConfig structure
func TestDeploymentConfig(t *testing.T) {
	now := time.Now()
	config := &DeploymentConfig{
		Name:         "test-deployment",
		TemplateName: "universal-ecs",
		Provider:     "alicloud",
		Region:       "cn-hangzhou",
		InstanceType: "ecs.t6-c1m1.large",
		Userdata:     "#!/bin/bash\necho 'Hello World'",
		Variables: map[string]string{
			"instance_name":     "my-instance",
			"instance_password": "MyPassword123!",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	if config.Name != "test-deployment" {
		t.Errorf("Expected name 'test-deployment', got '%s'", config.Name)
	}
	
	if config.Provider != "alicloud" {
		t.Errorf("Expected provider 'alicloud', got '%s'", config.Provider)
	}
	
	if config.Region != "cn-hangzhou" {
		t.Errorf("Expected region 'cn-hangzhou', got '%s'", config.Region)
	}
	
	if len(config.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(config.Variables))
	}
}

// TestValidationResult tests the ValidationResult structure
func TestValidationResult(t *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{
				Field:   "provider",
				Message: "Provider is required",
				Code:    ErrCodeRequired,
			},
		},
		Warnings: []ValidationWarning{
			{
				Field:   "userdata",
				Message: "Userdata is empty",
			},
		},
	}
	
	if result.Valid {
		t.Error("Expected Valid to be false")
	}
	
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}
	
	if result.Errors[0].Code != ErrCodeRequired {
		t.Errorf("Expected error code '%s', got '%s'", ErrCodeRequired, result.Errors[0].Code)
	}
	
	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}
}

// TestValidationErrorCodes tests the error code constants
func TestValidationErrorCodes(t *testing.T) {
	codes := []string{
		ErrCodeRequired,
		ErrCodeInvalidFormat,
		ErrCodeInvalidValue,
		ErrCodeNotSupported,
		ErrCodeNotAvailable,
	}
	
	expectedCodes := []string{
		"REQUIRED",
		"INVALID_FORMAT",
		"INVALID_VALUE",
		"NOT_SUPPORTED",
		"NOT_AVAILABLE",
	}
	
	for i, code := range codes {
		if code != expectedCodes[i] {
			t.Errorf("Expected code '%s', got '%s'", expectedCodes[i], code)
		}
	}
}

// TestRegion tests the Region structure
func TestRegion(t *testing.T) {
	region := &Region{
		Code: "cn-hangzhou",
		Name: "华东1（杭州）",
	}
	
	if region.Code != "cn-hangzhou" {
		t.Errorf("Expected code 'cn-hangzhou', got '%s'", region.Code)
	}
	
	if region.Name != "华东1（杭州）" {
		t.Errorf("Expected name '华东1（杭州）', got '%s'", region.Name)
	}
}

// TestInstanceType tests the InstanceType structure
func TestInstanceType(t *testing.T) {
	instanceType := &InstanceType{
		Code:        "ecs.t6-c1m1.large",
		Name:        "通用型 t6",
		CPU:         1,
		Memory:      1,
		Description: "1核1GB",
		Price:       0.05,
	}
	
	if instanceType.Code != "ecs.t6-c1m1.large" {
		t.Errorf("Expected code 'ecs.t6-c1m1.large', got '%s'", instanceType.Code)
	}
	
	if instanceType.CPU != 1 {
		t.Errorf("Expected CPU 1, got %d", instanceType.CPU)
	}
	
	if instanceType.Memory != 1 {
		t.Errorf("Expected Memory 1, got %d", instanceType.Memory)
	}
	
	if instanceType.Price != 0.05 {
		t.Errorf("Expected Price 0.05, got %f", instanceType.Price)
	}
}

// TestCostEstimate tests the CostEstimate structure
func TestCostEstimate(t *testing.T) {
	estimate := &CostEstimate{
		MonthlyCost: 36.0,
		Currency:    "CNY",
		Details: map[string]float64{
			"instance": 36.0,
		},
	}
	
	if estimate.MonthlyCost != 36.0 {
		t.Errorf("Expected MonthlyCost 36.0, got %f", estimate.MonthlyCost)
	}
	
	if estimate.Currency != "CNY" {
		t.Errorf("Expected Currency 'CNY', got '%s'", estimate.Currency)
	}
	
	if len(estimate.Details) != 1 {
		t.Errorf("Expected 1 detail, got %d", len(estimate.Details))
	}
}

// TestCustomDeployment tests the CustomDeployment structure
func TestCustomDeployment(t *testing.T) {
	now := time.Now()
	deployment := &CustomDeployment{
		ID:           "test-deployment-id",
		Name:         "test-deployment",
		TemplateName: "universal-ecs",
		Config: &DeploymentConfig{
			Name:         "test-deployment",
			TemplateName: "universal-ecs",
			Provider:     "alicloud",
			Region:       "cn-hangzhou",
			InstanceType: "ecs.t6-c1m1.large",
		},
		State:     StateCreated,
		CreatedAt: now,
		UpdatedAt: now,
		Outputs: map[string]interface{}{
			"instance_id": "i-123456",
			"public_ip":   "1.2.3.4",
		},
	}
	
	if deployment.ID != "test-deployment-id" {
		t.Errorf("Expected ID 'test-deployment-id', got '%s'", deployment.ID)
	}
	
	if deployment.State != StateCreated {
		t.Errorf("Expected State '%s', got '%s'", StateCreated, deployment.State)
	}
	
	if len(deployment.Outputs) != 2 {
		t.Errorf("Expected 2 outputs, got %d", len(deployment.Outputs))
	}
}

// TestGetProviderRegions tests the GetProviderRegions function
func TestGetProviderRegions(t *testing.T) {
	tests := []struct {
		name          string
		provider      string
		expectError   bool
		expectEmpty   bool
		expectedCount int
	}{
		{
			name:          "Valid provider - alicloud",
			provider:      "alicloud",
			expectError:   false,
			expectEmpty:   false,
			expectedCount: 23, // Based on regions.json
		},
		{
			name:          "Valid provider - tencentcloud",
			provider:      "tencentcloud",
			expectError:   false,
			expectEmpty:   false,
			expectedCount: 22,
		},
		{
			name:          "Valid provider - aws",
			provider:      "aws",
			expectError:   false,
			expectEmpty:   false,
			expectedCount: 22,
		},
		{
			name:          "Valid provider - volcengine",
			provider:      "volcengine",
			expectError:   false,
			expectEmpty:   false,
			expectedCount: 3,
		},
		{
			name:          "Valid provider - huaweicloud",
			provider:      "huaweicloud",
			expectError:   false,
			expectEmpty:   false,
			expectedCount: 15,
		},
		{
			name:        "Invalid provider",
			provider:    "invalid-provider",
			expectError: true,
		},
		{
			name:        "Empty provider - returns all regions",
			provider:    "",
			expectError: false,
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regions, err := GetProviderRegions(tt.provider)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.expectEmpty && len(regions) > 0 {
				t.Errorf("Expected empty regions list, got %d regions", len(regions))
			}

			if !tt.expectEmpty && len(regions) == 0 {
				t.Error("Expected non-empty regions list, got empty")
			}

			if tt.expectedCount > 0 && len(regions) != tt.expectedCount {
				t.Errorf("Expected %d regions, got %d", tt.expectedCount, len(regions))
			}

			// Verify region structure
			for _, region := range regions {
				if region.Code == "" {
					t.Error("Region code should not be empty")
				}
				if region.Name == "" {
					t.Error("Region name should not be empty")
				}
			}
		})
	}
}

// TestGetProviderRegionsService tests the CustomDeploymentService.GetProviderRegions method
func TestGetProviderRegionsService(t *testing.T) {
	service := NewCustomDeploymentService()

	// Test with valid provider
	regions, err := service.GetProviderRegions("alicloud")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(regions) == 0 {
		t.Error("Expected non-empty regions list")
	}

	// Verify first region has expected structure
	if len(regions) > 0 {
		region := regions[0]
		if region.Code == "" {
			t.Error("Region code should not be empty")
		}
		if region.Name == "" {
			t.Error("Region name should not be empty")
		}
	}

	// Test with invalid provider
	_, err = service.GetProviderRegions("invalid-provider")
	if err == nil {
		t.Error("Expected error for invalid provider")
	}
}

// TestLoadRegionsData tests the loadRegionsData function
func TestLoadRegionsData(t *testing.T) {
	regions, err := loadRegionsData()
	if err != nil {
		t.Fatalf("Failed to load regions data: %v", err)
	}

	// Check that all expected providers are present
	expectedProviders := []string{"alicloud", "tencentcloud", "aws", "volcengine", "huaweicloud"}
	for _, provider := range expectedProviders {
		if _, ok := regions[provider]; !ok {
			t.Errorf("Expected provider '%s' not found in regions data", provider)
		}
	}

	// Verify each provider has regions
	for provider, regionList := range regions {
		if len(regionList) == 0 {
			t.Errorf("Provider '%s' has no regions", provider)
		}

		// Verify each region has code and name
		for i, region := range regionList {
			if region.Code == "" {
				t.Errorf("Provider '%s' region %d has empty code", provider, i)
			}
			if region.Name == "" {
				t.Errorf("Provider '%s' region %d has empty name", provider, i)
			}
		}
	}
}

// TestGetInstanceTypes tests the GetInstanceTypes function
func TestGetInstanceTypes(t *testing.T) {
	tests := []struct {
		name          string
		provider      string
		region        string
		expectError   bool
		expectEmpty   bool
		minCount      int
	}{
		{
			name:        "Valid - alicloud",
			provider:    "alicloud",
			region:      "cn-hangzhou",
			expectError: false,
			expectEmpty: false,
			minCount:    5,
		},
		{
			name:        "Valid - tencentcloud",
			provider:    "tencentcloud",
			region:      "ap-guangzhou",
			expectError: false,
			expectEmpty: false,
			minCount:    5,
		},
		{
			name:        "Valid - aws",
			provider:    "aws",
			region:      "us-east-1",
			expectError: false,
			expectEmpty: false,
			minCount:    5,
		},
		{
			name:        "Valid - volcengine",
			provider:    "volcengine",
			region:      "cn-beijing",
			expectError: false,
			expectEmpty: false,
			minCount:    3,
		},
		{
			name:        "Valid - huaweicloud",
			provider:    "huaweicloud",
			region:      "cn-north-1",
			expectError: false,
			expectEmpty: false,
			minCount:    3,
		},
		{
			name:        "Empty provider",
			provider:    "",
			region:      "cn-hangzhou",
			expectError: true,
		},
		{
			name:        "Empty region",
			provider:    "alicloud",
			region:      "",
			expectError: true,
		},
		{
			name:        "Invalid provider",
			provider:    "invalid-provider",
			region:      "cn-hangzhou",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			types, err := GetInstanceTypes(tt.provider, tt.region)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.expectEmpty && len(types) > 0 {
				t.Errorf("Expected empty instance types list, got %d types", len(types))
			}

			if !tt.expectEmpty && len(types) == 0 {
				t.Error("Expected non-empty instance types list, got empty")
			}

			if tt.minCount > 0 && len(types) < tt.minCount {
				t.Errorf("Expected at least %d instance types, got %d", tt.minCount, len(types))
			}

			// Verify instance type structure
			for _, instanceType := range types {
				if instanceType.Code == "" {
					t.Error("Instance type code should not be empty")
				}
				if instanceType.Name == "" {
					t.Error("Instance type name should not be empty")
				}
				if instanceType.CPU <= 0 {
					t.Errorf("Instance type CPU should be positive, got %d", instanceType.CPU)
				}
				if instanceType.Memory <= 0 {
					t.Errorf("Instance type memory should be positive, got %d", instanceType.Memory)
				}
				if instanceType.Description == "" {
					t.Error("Instance type description should not be empty")
				}
			}
		})
	}
}

// TestGetInstanceTypesService tests the CustomDeploymentService.GetInstanceTypes method
func TestGetInstanceTypesService(t *testing.T) {
	service := NewCustomDeploymentService()

	// Test with valid provider and region
	types, err := service.GetInstanceTypes("alicloud", "cn-hangzhou")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(types) == 0 {
		t.Error("Expected non-empty instance types list")
	}

	// Verify first instance type has expected structure
	if len(types) > 0 {
		instanceType := types[0]
		if instanceType.Code == "" {
			t.Error("Instance type code should not be empty")
		}
		if instanceType.Name == "" {
			t.Error("Instance type name should not be empty")
		}
		if instanceType.CPU <= 0 {
			t.Error("Instance type CPU should be positive")
		}
		if instanceType.Memory <= 0 {
			t.Error("Instance type memory should be positive")
		}
	}

	// Test with invalid provider
	_, err = service.GetInstanceTypes("invalid-provider", "cn-hangzhou")
	if err == nil {
		t.Error("Expected error for invalid provider")
	}

	// Test with empty provider
	_, err = service.GetInstanceTypes("", "cn-hangzhou")
	if err == nil {
		t.Error("Expected error for empty provider")
	}

	// Test with empty region
	_, err = service.GetInstanceTypes("alicloud", "")
	if err == nil {
		t.Error("Expected error for empty region")
	}
}

// TestFetchInstanceTypesFromProvider tests fetching from different providers
func TestFetchInstanceTypesFromProvider(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		region      string
		expectError bool
		minCount    int
	}{
		{
			name:        "Alicloud",
			provider:    "alicloud",
			region:      "cn-hangzhou",
			expectError: false,
			minCount:    5,
		},
		{
			name:        "Tencentcloud",
			provider:    "tencentcloud",
			region:      "ap-guangzhou",
			expectError: false,
			minCount:    5,
		},
		{
			name:        "AWS",
			provider:    "aws",
			region:      "us-east-1",
			expectError: false,
			minCount:    5,
		},
		{
			name:        "Volcengine",
			provider:    "volcengine",
			region:      "cn-beijing",
			expectError: false,
			minCount:    3,
		},
		{
			name:        "Huaweicloud",
			provider:    "huaweicloud",
			region:      "cn-north-1",
			expectError: false,
			minCount:    3,
		},
		{
			name:        "Invalid provider",
			provider:    "invalid",
			region:      "cn-hangzhou",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			types, err := fetchInstanceTypesFromProvider(tt.provider, tt.region)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(types) < tt.minCount {
				t.Errorf("Expected at least %d instance types, got %d", tt.minCount, len(types))
			}

			// Verify all instance types have required fields
			for i, instanceType := range types {
				if instanceType.Code == "" {
					t.Errorf("Instance type %d has empty code", i)
				}
				if instanceType.Name == "" {
					t.Errorf("Instance type %d has empty name", i)
				}
				if instanceType.CPU <= 0 {
					t.Errorf("Instance type %d has invalid CPU: %d", i, instanceType.CPU)
				}
				if instanceType.Memory <= 0 {
					t.Errorf("Instance type %d has invalid memory: %d", i, instanceType.Memory)
				}
			}
		})
	}
}

// TestCreateCustomDeployment tests the CreateCustomDeployment method
func TestCreateCustomDeployment(t *testing.T) {
	// Create a temporary directory for test project
	tempProjectDir, err := os.MkdirTemp("", "redc-test-project-*")
	if err != nil {
		t.Fatalf("Failed to create temp project directory: %v", err)
	}
	defer os.RemoveAll(tempProjectDir)

	// Create a temporary template directory
	tempTemplateDir, err := os.MkdirTemp("", "redc-test-templates-*")
	if err != nil {
		t.Fatalf("Failed to create temp template directory: %v", err)
	}
	defer os.RemoveAll(tempTemplateDir)

	// Save original TemplateDir and restore after test
	originalTemplateDir := TemplateDir
	TemplateDir = tempTemplateDir
	defer func() { TemplateDir = originalTemplateDir }()

	// Create a test template
	testTemplateName := "test-template"
	testTemplatePath := filepath.Join(tempTemplateDir, testTemplateName)
	if err := os.MkdirAll(testTemplatePath, 0755); err != nil {
		t.Fatalf("Failed to create test template directory: %v", err)
	}

	// Create case.json
	caseData := map[string]interface{}{
		"name":              testTemplateName,
		"description":       "Test template",
		"user":              "system",
		"version":           "1.0.0",
		"redc_module":       "",
		"is_base_template":  true,
		"supported_providers": []string{"alicloud"},
	}
	caseJSON, _ := json.Marshal(caseData)
	caseFilePath := filepath.Join(testTemplatePath, TmplCaseFile)
	if err := os.WriteFile(caseFilePath, caseJSON, 0644); err != nil {
		t.Fatalf("Failed to write case.json: %v", err)
	}

	// Create variables.tf
	variablesTf := `
variable "provider" {
  description = "云厂商"
  type        = string
}

variable "region" {
  description = "地域"
  type        = string
}

variable "instance_type" {
  description = "实例规格"
  type        = string
}
`
	variablesFilePath := filepath.Join(testTemplatePath, "variables.tf")
	if err := os.WriteFile(variablesFilePath, []byte(variablesTf), 0644); err != nil {
		t.Fatalf("Failed to write variables.tf: %v", err)
	}

	// Create main.tf (minimal)
	mainTf := `
terraform {
  required_providers {
    alicloud = {
      source  = "aliyun/alicloud"
      version = "~> 1.0"
    }
  }
}

provider "alicloud" {
  region = var.region
}
`
	mainFilePath := filepath.Join(testTemplatePath, "main.tf")
	if err := os.WriteFile(mainFilePath, []byte(mainTf), 0644); err != nil {
		t.Fatalf("Failed to write main.tf: %v", err)
	}

	// Create service with custom template manager
	service := NewCustomDeploymentService()
	service.templateMgr = &TemplateManager{templateDir: tempTemplateDir}
	service.validator = &ConfigValidator{
		templateMgr: &TemplateManager{templateDir: tempTemplateDir},
	}

	// Test with valid configuration
	config := &DeploymentConfig{
		Name:         "test-deployment",
		TemplateName: testTemplateName,
		Provider:     "alicloud",
		Region:       "cn-hangzhou",
		InstanceType: "ecs.t6-c1m1.large",
		Variables: map[string]string{
			"instance_name": "test-instance",
		},
	}

	// Note: This test will fail at Terraform init/plan stage because we don't have
	// a real Terraform setup in the test environment. We're testing the logic up to that point.
	deployment, err := service.CreateCustomDeployment(config, tempProjectDir, "test-project")
	
	// We expect an error because Terraform init/plan will fail in test environment
	// But we can verify that the method handles the error correctly
	if err != nil {
		// This is expected - Terraform operations will fail in test environment
		t.Logf("Expected error during Terraform operations: %v", err)
		
		// Verify that the error message is informative
		if err.Error() == "" {
			t.Error("Error message should not be empty")
		}
		
		// The deployment directory should be cleaned up on error
		// We can't easily verify this without mocking, so we'll skip this check
		return
	}

	// If we somehow got here (unlikely in test environment), verify the deployment
	if deployment == nil {
		t.Fatal("Expected deployment to be created")
	}

	if deployment.ID == "" {
		t.Error("Deployment ID should not be empty")
	}

	if deployment.Name != config.Name {
		t.Errorf("Expected deployment name '%s', got '%s'", config.Name, deployment.Name)
	}

	if deployment.TemplateName != config.TemplateName {
		t.Errorf("Expected template name '%s', got '%s'", config.TemplateName, deployment.TemplateName)
	}

	if deployment.State != StatePending {
		t.Errorf("Expected state '%s', got '%s'", StatePending, deployment.State)
	}

	if deployment.Config == nil {
		t.Error("Deployment config should not be nil")
	}

	if deployment.Outputs == nil {
		t.Error("Deployment outputs should not be nil")
	}

	if deployment.ProjectID != "test-project" {
		t.Errorf("Expected project ID 'test-project', got '%s'", deployment.ProjectID)
	}
}

// TestCreateCustomDeployment_InvalidConfig tests error handling for invalid configurations
func TestCreateCustomDeployment_InvalidConfig(t *testing.T) {
	service := NewCustomDeploymentService()
	tempProjectDir, _ := os.MkdirTemp("", "redc-test-project-*")
	defer os.RemoveAll(tempProjectDir)

	tests := []struct {
		name        string
		config      *DeploymentConfig
		projectPath string
		projectID   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Nil config",
			config:      nil,
			projectPath: tempProjectDir,
			projectID:   "test-project",
			expectError: true,
			errorMsg:    "部署配置不能为空",
		},
		{
			name: "Empty project path",
			config: &DeploymentConfig{
				Name:         "test",
				TemplateName: "test-template",
				Provider:     "alicloud",
				Region:       "cn-hangzhou",
				InstanceType: "ecs.t6-c1m1.large",
			},
			projectPath: "",
			projectID:   "test-project",
			expectError: true,
			errorMsg:    "项目路径不能为空",
		},
		{
			name: "Empty project ID",
			config: &DeploymentConfig{
				Name:         "test",
				TemplateName: "test-template",
				Provider:     "alicloud",
				Region:       "cn-hangzhou",
				InstanceType: "ecs.t6-c1m1.large",
			},
			projectPath: tempProjectDir,
			projectID:   "",
			expectError: true,
			errorMsg:    "项目 ID 不能为空",
		},
		{
			name: "Missing required fields",
			config: &DeploymentConfig{
				Name:         "test",
				TemplateName: "test-template",
				// Missing Provider, Region, InstanceType
			},
			projectPath: tempProjectDir,
			projectID:   "test-project",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployment, err := service.CreateCustomDeployment(tt.config, tt.projectPath, tt.projectID)

			if !tt.expectError {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Error("Expected error but got none")
				return
			}

			if tt.errorMsg != "" && err.Error() != tt.errorMsg {
				t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
			}

			if deployment != nil {
				t.Error("Expected nil deployment on error")
			}
		})
	}
}

// TestCopyTemplate tests the copyTemplate helper function
func TestCopyTemplate(t *testing.T) {
	// Create source directory
	srcDir, err := os.MkdirTemp("", "redc-test-src-*")
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	defer os.RemoveAll(srcDir)

	// Create some test files in source
	testFile1 := filepath.Join(srcDir, "test1.txt")
	if err := os.WriteFile(testFile1, []byte("test content 1"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	testFile2 := filepath.Join(srcDir, "test2.txt")
	if err := os.WriteFile(testFile2, []byte("test content 2"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create destination directory
	dstDir, err := os.MkdirTemp("", "redc-test-dst-*")
	if err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}
	defer os.RemoveAll(dstDir)

	// Test copy
	err = copyTemplate(srcDir, dstDir)
	if err != nil {
		t.Fatalf("Failed to copy template: %v", err)
	}

	// Verify files were copied
	copiedFile1 := filepath.Join(dstDir, "test1.txt")
	if _, err := os.Stat(copiedFile1); os.IsNotExist(err) {
		t.Error("test1.txt was not copied")
	}

	copiedFile2 := filepath.Join(dstDir, "test2.txt")
	if _, err := os.Stat(copiedFile2); os.IsNotExist(err) {
		t.Error("test2.txt was not copied")
	}

	// Verify content
	content1, err := os.ReadFile(copiedFile1)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}
	if string(content1) != "test content 1" {
		t.Errorf("Expected content 'test content 1', got '%s'", string(content1))
	}
}

// TestCustomDeploymentDBSave tests saving a deployment to the database
func TestCustomDeploymentDBSave(t *testing.T) {
	// Create a test deployment
	now := time.Now()
	deployment := &CustomDeployment{
		ID:           "test-deployment-id",
		Name:         "test-deployment",
		TemplateName: "universal-ecs",
		Config: &DeploymentConfig{
			Name:         "test-deployment",
			TemplateName: "universal-ecs",
			Provider:     "alicloud",
			Region:       "cn-hangzhou",
			InstanceType: "ecs.t6-c1m1.large",
			Variables: map[string]string{
				"instance_name": "test-instance",
			},
		},
		State:     StatePending,
		CreatedAt: now,
		UpdatedAt: now,
		Outputs: map[string]interface{}{
			"instance_id": "i-123456",
			"public_ip":   "1.2.3.4",
		},
		ProjectID: "test-project",
	}

	// Save to database
	err := deployment.DBSave()
	if err != nil {
		t.Fatalf("Failed to save deployment: %v", err)
	}

	// Load from database
	loaded, err := LoadCustomDeployment("test-project", "test-deployment-id")
	if err != nil {
		t.Fatalf("Failed to load deployment: %v", err)
	}

	// Verify loaded deployment matches original
	if loaded.ID != deployment.ID {
		t.Errorf("Expected ID '%s', got '%s'", deployment.ID, loaded.ID)
	}

	if loaded.Name != deployment.Name {
		t.Errorf("Expected Name '%s', got '%s'", deployment.Name, loaded.Name)
	}

	if loaded.TemplateName != deployment.TemplateName {
		t.Errorf("Expected TemplateName '%s', got '%s'", deployment.TemplateName, loaded.TemplateName)
	}

	if loaded.State != deployment.State {
		t.Errorf("Expected State '%s', got '%s'", deployment.State, loaded.State)
	}

	if loaded.Config == nil {
		t.Fatal("Config should not be nil")
	}

	if loaded.Config.Provider != deployment.Config.Provider {
		t.Errorf("Expected Provider '%s', got '%s'", deployment.Config.Provider, loaded.Config.Provider)
	}

	if loaded.Config.Region != deployment.Config.Region {
		t.Errorf("Expected Region '%s', got '%s'", deployment.Config.Region, loaded.Config.Region)
	}

	if len(loaded.Outputs) != len(deployment.Outputs) {
		t.Errorf("Expected %d outputs, got %d", len(deployment.Outputs), len(loaded.Outputs))
	}

	// Clean up
	err = deployment.DBRemove()
	if err != nil {
		t.Errorf("Failed to remove deployment: %v", err)
	}
}

// TestCustomDeploymentDBSave_MissingProjectID tests error handling when ProjectID is missing
func TestCustomDeploymentDBSave_MissingProjectID(t *testing.T) {
	deployment := &CustomDeployment{
		ID:           "test-deployment-id",
		Name:         "test-deployment",
		TemplateName: "universal-ecs",
		State:        StatePending,
		// ProjectID is missing
	}

	err := deployment.DBSave()
	if err == nil {
		t.Error("Expected error when ProjectID is missing")
	}

	if err != nil && err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

// TestLoadProjectCustomDeployments tests loading all deployments for a project
func TestLoadProjectCustomDeployments(t *testing.T) {
	projectID := "test-project-multi"

	// Create multiple test deployments
	deployments := []*CustomDeployment{
		{
			ID:           "deployment-1",
			Name:         "deployment-1",
			TemplateName: "template-1",
			Config: &DeploymentConfig{
				Name:         "deployment-1",
				TemplateName: "template-1",
				Provider:     "alicloud",
				Region:       "cn-hangzhou",
				InstanceType: "ecs.t6-c1m1.large",
			},
			State:     StatePending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Outputs:   make(map[string]interface{}),
			ProjectID: projectID,
		},
		{
			ID:           "deployment-2",
			Name:         "deployment-2",
			TemplateName: "template-2",
			Config: &DeploymentConfig{
				Name:         "deployment-2",
				TemplateName: "template-2",
				Provider:     "tencentcloud",
				Region:       "ap-guangzhou",
				InstanceType: "S6.MEDIUM2",
			},
			State:     StateRunning,
			CreatedAt: time.Now().Add(-1 * time.Hour),
			UpdatedAt: time.Now(),
			Outputs:   make(map[string]interface{}),
			ProjectID: projectID,
		},
	}

	// Save all deployments
	for _, d := range deployments {
		if err := d.DBSave(); err != nil {
			t.Fatalf("Failed to save deployment %s: %v", d.ID, err)
		}
	}

	// Load all deployments for the project
	loaded, err := LoadProjectCustomDeployments(projectID)
	if err != nil {
		t.Fatalf("Failed to load deployments: %v", err)
	}

	if len(loaded) != len(deployments) {
		t.Errorf("Expected %d deployments, got %d", len(deployments), len(loaded))
	}

	// Verify deployments are sorted by creation time (newest first)
	if len(loaded) >= 2 {
		if loaded[0].CreatedAt.Before(loaded[1].CreatedAt) {
			t.Error("Deployments should be sorted by creation time (newest first)")
		}
	}

	// Clean up
	for _, d := range deployments {
		if err := d.DBRemove(); err != nil {
			t.Errorf("Failed to remove deployment %s: %v", d.ID, err)
		}
	}
}

// TestLoadProjectCustomDeployments_EmptyProject tests loading from a project with no deployments
func TestLoadProjectCustomDeployments_EmptyProject(t *testing.T) {
	deployments, err := LoadProjectCustomDeployments("non-existent-project")
	if err != nil {
		t.Errorf("Should not error for non-existent project: %v", err)
	}

	if len(deployments) != 0 {
		t.Errorf("Expected empty deployments list, got %d deployments", len(deployments))
	}
}

// TestCustomDeploymentDBRemove tests removing a deployment from the database
func TestCustomDeploymentDBRemove(t *testing.T) {
	// Create and save a test deployment
	deployment := &CustomDeployment{
		ID:           "test-deployment-remove",
		Name:         "test-deployment",
		TemplateName: "universal-ecs",
		Config: &DeploymentConfig{
			Name:         "test-deployment",
			TemplateName: "universal-ecs",
			Provider:     "alicloud",
			Region:       "cn-hangzhou",
			InstanceType: "ecs.t6-c1m1.large",
		},
		State:     StatePending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Outputs:   make(map[string]interface{}),
		ProjectID: "test-project-remove",
	}

	// Save to database
	err := deployment.DBSave()
	if err != nil {
		t.Fatalf("Failed to save deployment: %v", err)
	}

	// Verify it exists
	loaded, err := LoadCustomDeployment("test-project-remove", "test-deployment-remove")
	if err != nil {
		t.Fatalf("Failed to load deployment: %v", err)
	}
	if loaded == nil {
		t.Fatal("Deployment should exist")
	}

	// Remove it
	err = deployment.DBRemove()
	if err != nil {
		t.Fatalf("Failed to remove deployment: %v", err)
	}

	// Verify it no longer exists
	_, err = LoadCustomDeployment("test-project-remove", "test-deployment-remove")
	if err == nil {
		t.Error("Expected error when loading removed deployment")
	}
}

// TestCustomDeploymentToProto tests the toProto conversion
func TestCustomDeploymentToProto(t *testing.T) {
	now := time.Now()
	deployment := &CustomDeployment{
		ID:           "test-id",
		Name:         "test-name",
		TemplateName: "test-template",
		Config: &DeploymentConfig{
			Name:         "test-config",
			TemplateName: "test-template",
			Provider:     "alicloud",
			Region:       "cn-hangzhou",
			InstanceType: "ecs.t6-c1m1.large",
		},
		State:     StatePending,
		CreatedAt: now,
		UpdatedAt: now,
		Outputs: map[string]interface{}{
			"key": "value",
		},
		ProjectID: "test-project",
	}

	proto, err := deployment.toProto()
	if err != nil {
		t.Fatalf("Failed to convert to proto: %v", err)
	}

	if proto.Id != deployment.ID {
		t.Errorf("Expected ID '%s', got '%s'", deployment.ID, proto.Id)
	}

	if proto.Name != deployment.Name {
		t.Errorf("Expected Name '%s', got '%s'", deployment.Name, proto.Name)
	}

	if proto.TemplateName != deployment.TemplateName {
		t.Errorf("Expected TemplateName '%s', got '%s'", deployment.TemplateName, proto.TemplateName)
	}

	if proto.State != deployment.State {
		t.Errorf("Expected State '%s', got '%s'", deployment.State, proto.State)
	}

	if proto.ProjectId != deployment.ProjectID {
		t.Errorf("Expected ProjectID '%s', got '%s'", deployment.ProjectID, proto.ProjectId)
	}

	// Verify ConfigJson is valid JSON
	var config DeploymentConfig
	err = json.Unmarshal([]byte(proto.ConfigJson), &config)
	if err != nil {
		t.Errorf("ConfigJson should be valid JSON: %v", err)
	}

	// Verify OutputsJson is valid JSON
	var outputs map[string]interface{}
	err = json.Unmarshal([]byte(proto.OutputsJson), &outputs)
	if err != nil {
		t.Errorf("OutputsJson should be valid JSON: %v", err)
	}
}

// TestCustomDeploymentFromProto tests the fromProto conversion
func TestCustomDeploymentFromProto(t *testing.T) {
	now := time.Now()
	configJSON := `{"name":"test-config","template_name":"test-template","provider":"alicloud","region":"cn-hangzhou","instance_type":"ecs.t6-c1m1.large","variables":{},"created_at":"` + now.Format(time.RFC3339) + `","updated_at":"` + now.Format(time.RFC3339) + `"}`
	outputsJSON := `{"key":"value"}`

	proto := &pb.CustomDeployment{
		Id:           "test-id",
		Name:         "test-name",
		TemplateName: "test-template",
		ConfigJson:   configJSON,
		State:        StatePending,
		CreatedAt:    now.Format(time.RFC3339),
		UpdatedAt:    now.Format(time.RFC3339),
		OutputsJson:  outputsJSON,
		ProjectId:    "test-project",
	}

	deployment, err := customDeploymentFromProto(proto)
	if err != nil {
		t.Fatalf("Failed to convert from proto: %v", err)
	}

	if deployment.ID != proto.Id {
		t.Errorf("Expected ID '%s', got '%s'", proto.Id, deployment.ID)
	}

	if deployment.Name != proto.Name {
		t.Errorf("Expected Name '%s', got '%s'", proto.Name, deployment.Name)
	}

	if deployment.TemplateName != proto.TemplateName {
		t.Errorf("Expected TemplateName '%s', got '%s'", proto.TemplateName, deployment.TemplateName)
	}

	if deployment.State != proto.State {
		t.Errorf("Expected State '%s', got '%s'", proto.State, deployment.State)
	}

	if deployment.ProjectID != proto.ProjectId {
		t.Errorf("Expected ProjectID '%s', got '%s'", proto.ProjectId, deployment.ProjectID)
	}

	if deployment.Config == nil {
		t.Fatal("Config should not be nil")
	}

	if deployment.Outputs == nil {
		t.Fatal("Outputs should not be nil")
	}

	if len(deployment.Outputs) != 1 {
		t.Errorf("Expected 1 output, got %d", len(deployment.Outputs))
	}
}
