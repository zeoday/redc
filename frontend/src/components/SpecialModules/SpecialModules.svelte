<script>
  import { onMount } from 'svelte';
  import { loadUserdataTemplates, getTemplatesByCategory, getAIScenarios, getVulhubScenarios } from '../../lib/userdataTemplates.js';
  
  let { t } = $props();
  let specialModuleTab = $state('vulhub');
  let selectedAIScenario = $state(null);
  let selectedVulhubScenario = $state(null);
  let copied = $state(false);
  let aiSearchQuery = $state('');
  let vulhubSearchQuery = $state('');
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
  
  function selectAIScenario(scenario) {
    selectedAIScenario = scenario;
  }
  
  function selectVulhubScenario(scenario) {
    selectedVulhubScenario = scenario;
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
  </div>

  <!-- Content -->
  <div class="w-full">

  {#if specialModuleTab === 'vulhub'}
    <div class="bg-white rounded-xl border border-gray-100 p-4 sm:p-6 md:p-8">
      {#if templatesLoading}
        <div class="flex items-center justify-center h-32">
          <div class="w-6 h-6 border-2 border-gray-100 border-t-orange-500 rounded-full animate-spin"></div>
        </div>
      {:else if templates.length > 0 && getVulhubScenarios(templates).length > 0}
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
              {copied ? '已复制!' : '复制脚本'}
            </button>
          </div>
          {#if selectedVulhubScenario.level}
            <div class="mb-2">
              <span class="inline-block px-2 py-0.5 text-[11px] font-medium rounded {selectedVulhubScenario.level === 'critical' ? 'bg-red-100 text-red-700' : 'bg-orange-100 text-orange-700'}">
                {selectedVulhubScenario.level === 'critical' ? '严重' : '高危'}
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
      <div class="text-center py-8">
        <p class="text-[13px] text-gray-500">暂无可用的 C2 场景</p>
        <p class="text-[12px] text-gray-400 mt-1">此模块功能开发中...</p>
      </div>
      <div class="bg-blue-50 rounded-lg p-4 sm:p-6 text-center">
        <p class="text-[13px] text-blue-800 mb-2">C2 场景管理模块</p>
        <p class="text-[12px] text-blue-600">此功能正在开发中，敬请期待</p>
      </div>
    </div>
  {:else if specialModuleTab === 'ai'}
    <div class="bg-white rounded-xl border border-gray-100 p-4 sm:p-6 md:p-8">
      {#if templates.length > 0 && getTemplatesByCategory(templates, 'ai').length > 0}
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
              {copied ? '已复制!' : '复制脚本'}
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
  {/if}
</div>
