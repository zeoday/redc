<script>
  import { onMount } from 'svelte';
  import { loadUserdataTemplates, getTemplatesByCategory, getAIScenarios, getVulhubScenarios, getC2Scenarios } from '../../lib/userdataTemplates.js';
  
  let { t, onTabChange } = $props();
  let specialModuleTab = $state('vulhub');
  let selectedAIScenario = $state(null);
  let selectedVulhubScenario = $state(null);
  let selectedC2Scenario = $state(null);
  let copied = $state(false);
  let aiSearchQuery = $state('');
  let vulhubSearchQuery = $state('');
  let c2SearchQuery = $state('');
  let templates = $state([]);
  let templatesLoading = $state(true);
  
  onMount(async () => {
    templates = await loadUserdataTemplates();
    templatesLoading = false;
  });
  
  let aiScenarios = $derived(() => {
    let scenarios = getAIScenarios(templates);
    if (aiSearchQuery) {
      const query = aiSearchQuery.toLowerCase();
      scenarios = scenarios.filter(s => 
        (s.nameZh || s.name).toLowerCase().includes(query) ||
        s.name.toLowerCase().includes(query)
      );
    }
    return scenarios;
  });
  
  let vulhubScenarios = $derived(() => {
    let scenarios = getVulhubScenarios(templates);
    if (vulhubSearchQuery) {
      const query = vulhubSearchQuery.toLowerCase();
      scenarios = scenarios.filter(s => 
        (s.nameZh || s.name).toLowerCase().includes(query) ||
        (s.cveId || '').toLowerCase().includes(query) ||
        s.name.toLowerCase().includes(query)
      );
    }
    return scenarios;
  });
  
  let c2Scenarios = $derived(() => {
    let scenarios = getC2Scenarios(templates);
    if (c2SearchQuery) {
      const query = c2SearchQuery.toLowerCase();
      scenarios = scenarios.filter(s => 
        (s.nameZh || s.name).toLowerCase().includes(query) ||
        s.name.toLowerCase().includes(query)
      );
    }
    return scenarios;
  });
  
  function selectAIScenario(scenario) {
    selectedAIScenario = scenario;
  }
  
  function selectVulhubScenario(scenario) {
    selectedVulhubScenario = scenario;
  }
  
  function selectC2Scenario(scenario) {
    selectedC2Scenario = scenario;
  }
  
  async function copyToClipboard() {
    if (!selectedVulhubScenario) return;
    try {
      await navigator.clipboard.writeText(selectedVulhubScenario.script);
      copied = true;
      setTimeout(() => copied = false, 2000);
    } catch (e) {
      console.error('Failed to copy:', e);
    }
  }
</script>

<div class="w-full">
  <!-- Tabs -->
    <button
      class="px-4 py-2 text-[13px] font-medium transition-colors {specialModuleTab === 'vulhub' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-600 hover:text-gray-900'}"
      onclick={() => specialModuleTab = 'vulhub'}
    >
      {t.vulhubSupport}
    </button>
    <button
      class="px-4 py-2 text-[13px] font-medium transition-colors {specialModuleTab === 'c2' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-600 hover:text-gray-900'}"
      onclick={() => specialModuleTab = 'c2'}
    >
      {t.c2Scenes}
    </button>
    <button
      class="px-4 py-2 text-[13px] font-medium transition-colors {specialModuleTab === 'ai' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-600 hover:text-gray-900'}"
      onclick={() => specialModuleTab = 'ai'}
    >
      {t.aiScenes}
    </button>
    <button
      class="px-4 py-2 text-[13px] font-medium transition-colors {specialModuleTab === 'redcModules' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-600 hover:text-gray-900'}"
      onclick={() => specialModuleTab = 'redcModules'}
    >
      {t.redcModules}
    </button>
  </div>

  <!-- Content -->
  <div class="w-full">

  {#if specialModuleTab === 'vulhub'}
    <div class="bg-white rounded-xl border border-gray-100 p-4 sm:p-6 md:p-8">
      {#if templatesLoading}
        <div class="flex items-center justify-center h-32">
          <div class="w-6 h-6 border-2 border-gray-100 border-t-orange-500 rounded-full animate-spin"></div>
        </div>
      {:else if templates.length === 0}
        <div class="bg-blue-50 border border-blue-100 rounded-xl p-5 mb-4">
          <div class="flex items-start gap-3">
            <svg class="w-5 h-5 text-blue-500 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
            </svg>
            <div class="flex-1">
              <p class="text-[13px] text-blue-700">{t.noUserdataTemplatesHint}</p>
              <button 
                class="mt-3 h-8 px-4 bg-blue-500 text-white text-[12px] font-medium rounded-lg hover:bg-blue-600 transition-colors cursor-pointer"
                onclick={() => onTabChange && onTabChange('registry')}
              >
                {t.noUserdataTemplatesHintButton}
              </button>
            </div>
          </div>
        </div>
      {:else if getVulhubScenarios(templates).length > 0}
        <div class="mb-4">
          <input
            type="text"
            placeholder={t.search || '搜索漏洞环境...'}
            class="w-full h-10 px-3 text-[13px] bg-gray-50 border-0 rounded-lg text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:ring-offset-1 transition-shadow"
            bind:value={vulhubSearchQuery}
          />
          {#if vulhubScenarios().length > 0}
            <div class="flex flex-wrap gap-2 mt-3">
              {#each vulhubScenarios() as scenario}
                <button
                  class="px-3 py-2 text-[12px] font-medium rounded-lg transition-all {selectedVulhubScenario?.name === scenario.name ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}"
                  onclick={() => selectVulhubScenario(scenario)}
                >
                  {scenario.nameZh || scenario.name}
                </button>
              {/each}
            </div>
          {:else}
            <p class="text-[12px] text-gray-500 mt-3 text-center">未找到匹配的漏洞环境</p>
          {/if}
        </div>
      {:else}
        <div class="text-center py-8">
          <p class="text-[13px] text-gray-500">暂无可用的 Vulhub 漏洞环境</p>
          <p class="text-[12px] text-gray-400 mt-1">请确保已加载 redc-template 模板库</p>
        </div>
      {/if}
      
      {#if selectedVulhubScenario}
        <div class="mt-6">
          <div class="flex items-center justify-between mb-2">
            <div>
              <p class="text-[13px] font-medium text-gray-700">
                {selectedVulhubScenario.nameZh || selectedVulhubScenario.name}
              </p>
              {#if selectedVulhubScenario.cveId}
                <p class="text-[11px] text-orange-600 font-medium">
                  {selectedVulhubScenario.cveId}
                </p>
              {/if}
            </div>
            <button
              class="px-3 py-1.5 text-[12px] font-medium rounded-lg transition-all {copied ? 'bg-green-600 text-white' : 'bg-blue-600 text-white hover:bg-blue-700'}"
              onclick={copyToClipboard}
            >
              {copied ? (t.copiedSuccess || '已复制!') : (t.copyScript || '复制脚本')}
            </button>
          </div>
          {#if selectedVulhubScenario.level}
            <div class="mb-2">
              <span class="inline-block px-2 py-0.5 text-[11px] font-medium rounded {selectedVulhubScenario.level === 'critical' ? 'bg-red-100 text-red-700' : 'bg-orange-100 text-orange-700'}">
                {selectedVulhubScenario.level === 'critical' ? (t.severityCritical || '严重') : (t.severityHigh || '高危')}
              </span>
            </div>
          {/if}
          {#if selectedVulhubScenario.description}
            <p class="text-[12px] text-gray-500 mb-3">{selectedVulhubScenario.description}</p>
          {/if}
          {#if selectedVulhubScenario.environment}
            <div class="text-[12px] text-gray-500 mb-3">
              {#if selectedVulhubScenario.environment.port}
                <span class="mr-3">端口: {selectedVulhubScenario.environment.port}</span>
              {/if}
              {#if selectedVulhubScenario.environment.image}
                <span>镜像: {selectedVulhubScenario.environment.image}</span>
              {/if}
            </div>
          {/if}
          <pre class="bg-gray-900 text-gray-100 text-[12px] p-4 rounded-lg overflow-x-auto max-h-96 overflow-y-auto font-mono">{selectedVulhubScenario.script}</pre>
        </div>
      {:else}
        <div class="bg-blue-50 rounded-lg p-4 sm:p-6 text-center">
          <p class="text-[13px] text-blue-800 mb-2">选择一个漏洞环境查看部署脚本</p>
          <p class="text-[12px] text-blue-600">脚本将在云服务器上自动安装 Docker 并启动漏洞环境</p>
        </div>
      {/if}
    </div>
  {:else if specialModuleTab === 'c2'}
    <div class="bg-white rounded-xl border border-gray-100 p-4 sm:p-6 md:p-8">
      {#if templatesLoading}
        <div class="flex items-center justify-center h-32">
          <div class="w-6 h-6 border-2 border-gray-100 border-t-orange-500 rounded-full animate-spin"></div>
        </div>
      {:else if templates.length === 0}
        <div class="bg-blue-50 border border-blue-100 rounded-xl p-5 mb-4">
          <div class="flex items-start gap-3">
            <svg class="w-5 h-5 text-blue-500 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
            </svg>
            <div class="flex-1">
              <p class="text-[13px] text-blue-700">{t.noUserdataTemplatesHint}</p>
              <button 
                class="mt-3 h-8 px-4 bg-blue-500 text-white text-[12px] font-medium rounded-lg hover:bg-blue-600 transition-colors cursor-pointer"
                onclick={() => onTabChange && onTabChange('registry')}
              >
                {t.noUserdataTemplatesHintButton}
              </button>
            </div>
          </div>
        </div>
      {:else}
        <div class="mb-4">
          <input
            type="text"
            placeholder={t.searchPlaceholder || '搜索...'}
            class="w-full h-10 px-3 text-[13px] bg-gray-50 border-0 rounded-lg text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-gray-900 focus:ring-offset-1 transition-shadow"
            bind:value={c2SearchQuery}
          />
          {#if c2Scenarios().length > 0}
            <div class="flex flex-wrap gap-2 mt-3">
              {#each c2Scenarios() as scenario}
                <button
                  class="px-3 py-2 text-[12px] font-medium rounded-lg transition-all {selectedC2Scenario?.name === scenario.name ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}"
                  onclick={() => selectC2Scenario(scenario)}
                >
                  {scenario.nameZh || scenario.name}
                </button>
              {/each}
            </div>
          {:else}
            <p class="text-[12px] text-gray-500 mt-3 text-center">未找到匹配的 C2 场景</p>
          {/if}
        </div>
        
        {#if selectedC2Scenario}
          <div class="border-t border-gray-100 pt-4">
            {#if selectedC2Scenario.description}
              <p class="text-[13px] text-gray-700 mb-3">{selectedC2Scenario.description}</p>
            {/if}
            {#if selectedC2Scenario.script}
              <pre class="bg-gray-900 text-gray-100 text-[12px] p-4 rounded-lg overflow-x-auto max-h-96 overflow-y-auto font-mono">{selectedC2Scenario.script}</pre>
            {/if}
          </div>
        {:else}
          <div class="bg-blue-50 rounded-lg p-4 sm:p-6 text-center">
            <p class="text-[13px] text-blue-800 mb-2">选择一个 C2 场景查看部署脚本</p>
            <p class="text-[12px] text-blue-600">脚本将在云服务器上自动安装 C2 工具</p>
          </div>
        {/if}
      {/if}
    </div>
  {:else if specialModuleTab === 'ai'}
    <div class="bg-white rounded-xl border border-gray-100 p-4 sm:p-6 md:p-8">
      {#if templates.length === 0}
        <div class="bg-blue-50 border border-blue-100 rounded-xl p-5 mb-4">
          <div class="flex items-start gap-3">
            <svg class="w-5 h-5 text-blue-500 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
            </svg>
            <div class="flex-1">
              <p class="text-[13px] text-blue-700">{t.noUserdataTemplatesHint}</p>
              <button 
                class="mt-3 h-8 px-4 bg-blue-500 text-white text-[12px] font-medium rounded-lg hover:bg-blue-600 transition-colors cursor-pointer"
                onclick={() => onTabChange && onTabChange('registry')}
              >
                {t.noUserdataTemplatesHintButton}
              </button>
            </div>
          </div>
        </div>
      {:else if getTemplatesByCategory(templates, 'ai').length > 0}
        <div class="mb-4">
          <input
            type="text"
            placeholder={t.search || '搜索场景...'}
            class="w-full h-10 px-3 text-[13px] bg-gray-50 border-0 rounded-lg text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:ring-offset-1 transition-shadow"
            bind:value={aiSearchQuery}
          />
          {#if aiScenarios().length > 0}
            <div class="flex flex-wrap gap-2 mt-3">
              {#each aiScenarios() as scenario}
                <button
                  class="px-3 py-2 text-[12px] font-medium rounded-lg transition-all {selectedAIScenario?.name === scenario.name ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}"
                  onclick={() => selectAIScenario(scenario)}
                >
                  {scenario.nameZh || scenario.name}
                </button>
              {/each}
            </div>
          {:else}
            <p class="text-[12px] text-gray-500 mt-3 text-center">未找到匹配的场景</p>
          {/if}
        </div>
      {:else}
        <div class="text-center py-8">
          <p class="text-[13px] text-gray-500">暂无可用的 AI 场景</p>
          <p class="text-[12px] text-gray-400 mt-1">请确保已加载 redc-template 模板库</p>
        </div>
      {/if}
      
      {#if selectedAIScenario}
        <div class="mt-6">
          <div class="flex items-center justify-between mb-2">
            <div>
              <p class="text-[13px] font-medium text-gray-700">
                {selectedAIScenario.nameZh || selectedAIScenario.name}
              </p>
              {#if selectedAIScenario.url}
                <p class="text-[11px] text-gray-500">
                  {selectedAIScenario.url}
                </p>
              {/if}
            </div>
            <button
              class="px-3 py-1.5 text-[12px] font-medium rounded-lg transition-all {copied ? 'bg-green-600 text-white' : 'bg-blue-600 text-white hover:bg-blue-700'}"
              onclick={copyToClipboard}
            >
              {copied ? (t.copiedSuccess || '已复制!') : (t.copyScript || '复制脚本')}
            </button>
          </div>
          {#if selectedAIScenario.description}
            <p class="text-[12px] text-gray-500 mb-3">{selectedAIScenario.description}</p>
          {/if}
          <pre class="bg-gray-900 text-gray-100 text-[12px] p-4 rounded-lg overflow-x-auto max-h-96 overflow-y-auto font-mono">{selectedAIScenario.script}</pre>
          {#if selectedAIScenario.installNotes}
            <div class="mt-3 p-3 bg-blue-50 border border-blue-200 rounded-lg">
              <p class="text-[12px] text-blue-800 whitespace-pre-line">{selectedAIScenario.installNotes}</p>
            </div>
          {/if}
        </div>
      {:else}
        <div class="bg-blue-50 rounded-lg p-4 sm:p-6 text-center">
          <p class="text-[13px] text-blue-800 mb-2">选择一个 AI 场景查看部署脚本</p>
          <p class="text-[12px] text-blue-600">脚本将在云服务器上自动安装并启动 AI 服务</p>
        </div>
      {/if}
    </div>
  {:else if specialModuleTab === 'redcModules'}
    <div class="bg-white rounded-xl border border-gray-100 p-4 sm:p-6 md:p-8">
      <div class="text-center py-4 mb-6">
        <h3 class="text-[15px] font-medium text-gray-800 mb-2">RedC 增强模块</h3>
        <p class="text-[12px] text-gray-500">场景启动时自动执行的增强功能</p>
      </div>

      <div class="space-y-4">
        <!-- gen_clash_config -->
        <div class="border border-gray-100 rounded-lg p-4 hover:border-blue-200 transition-colors">
          <div class="flex items-center gap-3 mb-2">
            <div class="w-8 h-8 rounded-lg bg-blue-100 flex items-center justify-center">
              <svg class="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
              </svg>
            </div>
            <div>
              <h4 class="text-[13px] font-medium text-gray-800">gen_clash_config</h4>
              <p class="text-[11px] text-gray-500">生成 Clash 代理配置</p>
            </div>
          </div>
          <p class="text-[12px] text-gray-600 mb-3">根据场景输出的节点 IP 自动生成 Clash 配置文件，支持 Shadowsocks 协议，方便客户端直接导入使用。</p>
          <div class="bg-gray-50 rounded-lg p-3 mb-3">
            <p class="text-[11px] text-gray-500 mb-1">使用场景：</p>
            <p class="text-[11px] text-gray-700">多节点代理场景 (aws/proxy, aliyun/proxy 等)</p>
          </div>
          <div class="text-[11px] text-blue-600">
            <span class="font-medium">配置示例 (terraform.tfvars):</span>
            <pre class="mt-1 bg-gray-900 text-gray-100 p-2 rounded overflow-x-auto">port = "64277"
password = "your_password"
filename = "proxy-config.yaml"</pre>
          </div>
        </div>

        <!-- upload_r2 -->
        <div class="border border-gray-100 rounded-lg p-4 hover:border-blue-200 transition-colors">
          <div class="flex items-center gap-3 mb-2">
            <div class="w-8 h-8 rounded-lg bg-purple-100 flex items-center justify-center">
              <svg class="w-4 h-4 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"></path>
              </svg>
            </div>
            <div>
              <h4 class="text-[13px] font-medium text-gray-800">upload_r2</h4>
              <p class="text-[11px] text-gray-500">上传配置到 Cloudflare R2</p>
            </div>
          </div>
          <p class="text-[12px] text-gray-600 mb-3">将生成的配置文件自动上传到 Cloudflare R2 存储桶，并输出可直接访问的下载链接。</p>
          <div class="bg-gray-50 rounded-lg p-3 mb-3">
            <p class="text-[11px] text-gray-500 mb-1">使用场景：</p>
            <p class="text-[11px] text-gray-700">需要远程下载配置文件的场景</p>
          </div>
          <div class="text-[11px] text-purple-600">
            <span class="font-medium">配置示例 (terraform.tfvars):</span>
            <pre class="mt-1 bg-gray-900 text-gray-100 p-2 rounded overflow-x-auto">filename = "proxy-config.yaml"
buckets_name = "your-bucket"
buckets_path = "proxy"</pre>
          </div>
        </div>

        <!-- chang_dns -->
        <div class="border border-gray-100 rounded-lg p-4 hover:border-blue-200 transition-colors">
          <div class="flex items-center gap-3 mb-2">
            <div class="w-8 h-8 rounded-lg bg-green-100 flex items-center justify-center">
              <svg class="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
            </div>
            <div>
              <h4 class="text-[13px] font-medium text-gray-800">chang_dns</h4>
              <p class="text-[11px] text-gray-500">自动更新 Cloudflare DNS</p>
            </div>
          </div>
          <p class="text-[12px] text-gray-600 mb-3">场景启动后自动将实例公网 IP 更新到 Cloudflare DNS 记录，支持动态 IP 场景。</p>
          <div class="bg-gray-50 rounded-lg p-3 mb-3">
            <p class="text-[11px] text-gray-500 mb-1">使用场景：</p>
            <p class="text-[11px] text-gray-700">DNSLog、Interactsh 等需要域名解析的场景</p>
          </div>
          <div class="text-[11px] text-green-600">
            <span class="font-medium">配置方式：</span>
            <p class="mt-1 text-gray-600">在启动参数中传入 <code class="bg-gray-100 px-1 rounded">domain=your.dnslog.com</code></p>
            <p class="mt-1 text-gray-500">需要在配置文件中配置 Cloudflare API 凭证</p>
          </div>
        </div>
      </div>

      <div class="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
        <h4 class="text-[13px] font-medium text-blue-800 mb-2">如何在场景中使用</h4>
        <p class="text-[12px] text-blue-700 mb-2">在场景的 <code class="bg-white px-1 rounded">case.json</code> 文件中添加 <code class="bg-white px-1 rounded">redc_module</code> 字段：</p>
        <pre class="bg-gray-900 text-gray-100 text-[11px] p-3 rounded overflow-x-auto">{'{'}"name": "proxy",
  "redc_module": "gen_clash_config,upload_r2",
  "description": "多节点代理场景"{'}'}</pre>
        <p class="text-[11px] text-blue-600 mt-2">多个模块用逗号分隔，顺序执行。</p>
      </div>
    </div>
  {/if}
</div>
