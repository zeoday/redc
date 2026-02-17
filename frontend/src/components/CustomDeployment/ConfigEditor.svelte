<script lang="ts">
  import { onMount } from 'svelte';
  import { ValidateDeploymentConfig } from '../../../wailsjs/go/main/App.js';
  import ProviderSelector from './ProviderSelector.svelte';
  import RegionSelector from './RegionSelector.svelte';
  import InstanceTypeSelector from './InstanceTypeSelector.svelte';
  import UserdataEditor from './UserdataEditor.svelte';
  
  let { 
    t, 
    template, 
    config = {}, 
    validation = { valid: true, errors: [], warnings: [] },
    onConfigUpdate = () => {},
    onValidationUpdate = () => {}
  } = $props();

  // Local state for form fields
  let formData = $state({
    name: '',
    provider: '',
    region: '',
    instanceType: '',
    userdata: '',
    variables: {}
  });

  // Validation state
  let isValidating = $state(false);
  let validationTimeout = null;

  // Watch for external config changes
  $effect(() => {
    if (config) {
      formData = {
        name: config.name || '',
        provider: config.provider || '',
        region: config.region || '',
        instanceType: config.instanceType || '',
        userdata: config.userdata || '',
        variables: { ...(config.variables || {}) }
      };
    }
  });

  // Auto-set provider from template if available
  $effect(() => {
    if (template && template.provider && !formData.provider) {
      formData.provider = template.provider;
      updateConfig();
    }
  });

  // Initialize variables from template
  onMount(() => {
    if (template && template.variables) {
      const vars = {};
      template.variables.forEach(v => {
        vars[v.name] = formData.variables[v.name] || v.defaultValue || '';
      });
      formData.variables = vars;
      updateConfig();
    }
  });

  // Auto-sync top-level fields to template variables
  $effect(() => {
    if (template && template.variables) {
      const needsSync = template.variables.some(v => 
        v.name === 'cloud_provider' || v.name === 'region' || v.name === 'instance_type'
      );
      
      if (needsSync) {
        let updated = false;
        
        // Sync cloud_provider
        if (formData.provider && formData.variables['cloud_provider'] !== formData.provider) {
          formData.variables['cloud_provider'] = formData.provider;
          updated = true;
        }
        // Sync region
        if (formData.region && formData.variables['region'] !== formData.region) {
          formData.variables['region'] = formData.region;
          updated = true;
        }
        // Sync instance_type
        if (formData.instanceType && formData.variables['instance_type'] !== formData.instanceType) {
          formData.variables['instance_type'] = formData.instanceType;
          updated = true;
        }
        
        // Only trigger update if something changed
        if (updated) {
          updateConfig();
        }
      }
    }
  });

  function updateConfig() {
    onConfigUpdate({ ...formData });
    // Trigger validation after config update
    scheduleValidation();
  }

  function handleFieldChange(field, value) {
    formData[field] = value;
    updateConfig();
  }

  function handleVariableChange(varName, value) {
    formData.variables[varName] = value;
    updateConfig();
  }

  // Schedule validation with debounce
  function scheduleValidation() {
    if (validationTimeout) {
      clearTimeout(validationTimeout);
    }
    
    validationTimeout = setTimeout(() => {
      validateConfig();
    }, 500); // Debounce for 500ms
  }

  // Validate configuration
  async function validateConfig() {
    // Don't validate if required fields are empty
    if (!formData.name || !formData.provider || !formData.region || !formData.instanceType) {
      // Clear validation errors when fields are empty
      onValidationUpdate({
        valid: true,
        errors: [],
        warnings: []
      });
      return;
    }

    isValidating = true;
    
    try {
      const configToValidate: any = {
        name: formData.name,
        template_name: template?.name || '',
        provider: formData.provider,
        region: formData.region,
        instance_type: formData.instanceType,
        userdata: formData.userdata,
        variables: formData.variables,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      };

      const result = await ValidateDeploymentConfig(configToValidate);
      
      if (result) {
        onValidationUpdate(result);
      }
    } catch (error: any) {
      console.error('Validation error:', error);
      // Show error in validation
      onValidationUpdate({
        valid: false,
        errors: [{
          field: 'general',
          message: error.message || String(error),
          code: 'VALIDATION_ERROR'
        }],
        warnings: []
      });
    } finally {
      isValidating = false;
    }
  }

  // Get required variables from template (excluding already configured ones)
  let requiredVariables = $derived(() => {
    if (!template || !template.variables) return [];
    return template.variables.filter(v => 
      v.required && 
      v.name !== 'cloud_provider' && 
      v.name !== 'region' && 
      v.name !== 'instance_type'
    );
  });

  // Get optional variables from template (excluding already configured ones)
  let optionalVariables = $derived(() => {
    if (!template || !template.variables) return [];
    return template.variables.filter(v => 
      !v.required && 
      v.name !== 'cloud_provider' && 
      v.name !== 'region' && 
      v.name !== 'instance_type'
    );
  });
</script>

<div class="bg-white rounded-xl border border-gray-100 p-5">
  <div class="flex items-center justify-between mb-4">
    <h2 class="text-[15px] font-semibold text-gray-900">
      {t.deploymentConfig || '部署配置'}
    </h2>
    
    <!-- Validation Status Indicator -->
    {#if isValidating}
      <div class="flex items-center gap-2 text-[11px] text-gray-500">
        <div class="w-3 h-3 border-2 border-gray-300 border-t-gray-600 rounded-full animate-spin"></div>
        <span>{t.validating || '验证中...'}</span>
      </div>
    {:else if validation && validation.valid}
      <div class="flex items-center gap-1.5 text-[11px] text-green-600">
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
        </svg>
        <span>{t.validationPassed || '验证通过'}</span>
      </div>
    {/if}
  </div>

  <!-- Validation Errors -->
  {#if validation && !validation.valid && validation.errors.length > 0}
    <div class="mb-4 p-3 bg-red-50 border border-red-100 rounded-lg">
      <div class="flex items-start gap-2">
        <svg class="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
        </svg>
        <div class="flex-1">
          <p class="text-[12px] font-medium text-red-900 mb-1">{t.validationErrors || '验证错误'}</p>
          <ul class="text-[11px] text-red-700 space-y-0.5">
            {#each validation.errors as error}
              <li>• {error.field}: {error.message}</li>
            {/each}
          </ul>
        </div>
      </div>
    </div>
  {/if}

  <!-- Validation Warnings -->
  {#if validation && validation.warnings && validation.warnings.length > 0}
    <div class="mb-4 p-3 bg-amber-50 border border-amber-100 rounded-lg">
      <div class="flex items-start gap-2">
        <svg class="w-4 h-4 text-amber-500 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
        </svg>
        <div class="flex-1">
          <p class="text-[12px] font-medium text-amber-900 mb-1">{t.warnings || '警告'}</p>
          <ul class="text-[11px] text-amber-700 space-y-0.5">
            {#each validation.warnings as warning}
              <li>• {warning.message}</li>
            {/each}
          </ul>
        </div>
      </div>
    </div>
  {/if}

  <div class="space-y-4">
    <!-- Deployment Name -->
    <div>
      <label for="deployment-name" class="block text-[12px] font-medium text-gray-700 mb-1.5">
        {t.deploymentName || '部署名称'}
        <span class="text-gray-400 ml-1">({t.optional || '可选'})</span>
      </label>
      <input
        id="deployment-name"
        type="text"
        placeholder={t.deploymentNamePlaceholder || '输入部署名称...'}
        class="w-full h-10 px-3 text-[13px] bg-gray-50 border-0 rounded-lg text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-gray-900 focus:ring-offset-1 transition-shadow"
        value={formData.name}
        oninput={(e) => handleFieldChange('name', e.currentTarget.value)}
      />
    </div>

    <!-- Note: Provider, Region, InstanceType, and Userdata components will be integrated in task 15 -->
    <!-- For now, we'll use simple placeholders -->
    
    <!-- Cloud Provider - Only show if template doesn't specify a provider -->
    {#if !template || !template.provider}
      <ProviderSelector
        {t}
        value={formData.provider}
        onChange={(value) => handleFieldChange('provider', value)}
      />
    {:else}
      <!-- Show provider as read-only info -->
      <div>
        <label class="block text-[12px] font-medium text-gray-700 mb-1.5">
          {t.cloudProvider || '云厂商'}
        </label>
        <div class="w-full h-10 px-3 text-[13px] bg-gray-100 border-0 rounded-lg text-gray-600 flex items-center">
          {template.provider === 'alicloud' ? '阿里云 (Alibaba Cloud)' :
           template.provider === 'tencentcloud' ? '腾讯云 (Tencent Cloud)' :
           template.provider === 'aws' ? 'AWS (Amazon Web Services)' :
           template.provider === 'volcengine' ? '火山引擎 (Volcengine)' :
           template.provider === 'huaweicloud' ? '华为云 (Huawei Cloud)' :
           template.provider}
        </div>
      </div>
    {/if}

    <!-- Region -->
    <RegionSelector
      {t}
      provider={formData.provider}
      value={formData.region}
      onChange={(value) => handleFieldChange('region', value)}
    />

    <!-- Instance Type -->
    <InstanceTypeSelector
      {t}
      provider={formData.provider}
      region={formData.region}
      value={formData.instanceType}
      onChange={(value) => handleFieldChange('instanceType', value)}
    />

    <!-- Userdata -->
    <UserdataEditor
      {t}
      value={formData.userdata}
      onChange={(value) => handleFieldChange('userdata', value)}
    />

    <!-- Template Variables -->
    {#if template && template.variables && template.variables.length > 0}
      <div class="border-t border-gray-100 pt-4">
        <h3 class="text-[13px] font-medium text-gray-900 mb-3">
          {t.templateParams || '模板参数'}
        </h3>

        <!-- Required Variables -->
        {#if requiredVariables().length > 0}
          <div class="mb-4">
            <p class="text-[11px] font-medium text-gray-500 uppercase tracking-wide mb-2">
              {t.requiredParams || '必需参数'}
            </p>
            <div class="grid grid-cols-2 gap-3">
              {#each requiredVariables() as variable, index}
                <div>
                  <label for="required-var-{index}" class="block text-[11px] text-gray-700 mb-1">
                    {variable.name}
                    <span class="text-red-500">*</span>
                    {#if variable.description}
                      <span class="text-gray-400 ml-1">({variable.description})</span>
                    {/if}
                  </label>
                  <input
                    id="required-var-{index}"
                    type="text"
                    placeholder={variable.defaultValue || ''}
                    class="w-full h-9 px-3 text-[12px] bg-gray-50 border-0 rounded-lg text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-gray-900 focus:ring-offset-1 transition-shadow"
                    value={formData.variables[variable.name] || ''}
                    oninput={(e) => handleVariableChange(variable.name, e.currentTarget.value)}
                  />
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Optional Variables -->
        {#if optionalVariables().length > 0}
          <div>
            <p class="text-[11px] font-medium text-gray-500 uppercase tracking-wide mb-2">
              {t.optionalParams || '可选参数'}
            </p>
            <div class="grid grid-cols-2 gap-3">
              {#each optionalVariables() as variable, index}
                <div>
                  <label for="optional-var-{index}" class="block text-[11px] text-gray-700 mb-1">
                    {variable.name}
                    {#if variable.description}
                      <span class="text-gray-400 ml-1">({variable.description})</span>
                    {/if}
                  </label>
                  <input
                    id="optional-var-{index}"
                    type="text"
                    placeholder={variable.defaultValue || ''}
                    class="w-full h-9 px-3 text-[12px] bg-gray-50 border-0 rounded-lg text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-gray-900 focus:ring-offset-1 transition-shadow"
                    value={formData.variables[variable.name] || ''}
                    oninput={(e) => handleVariableChange(variable.name, e.currentTarget.value)}
                  />
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>

<style>
  /* Component-specific styles if needed */
</style>
