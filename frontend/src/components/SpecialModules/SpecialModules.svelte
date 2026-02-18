<script>
  import { onMount } from 'svelte';
  import { loadUserdataTemplates, getTemplatesByCategory,getAIScenarios } from '../../lib/userdataTemplates.js';
  
  let { t } = $props();
  let specialModuleTab = $state('vulhub');
  let selectedAIScenario = $state(null);
  let copied = $state(false);
  let aiSearchQuery = $state('');
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
  
  function selectAIScenario(scenario) {
    selectedAIScenario = scenario;
  }
  
  async function copyToClipboard() {
    if (!selectedAIScenario) return;
    try {
      await navigator.clipboard.writeText(selectedAIScenario.script);
      copied = true;
      setTimeout(() => copied = false, 2000);
    } catch (e) {
      console.error('Failed to copy:', e);
    }
  }
</script>

<div class="w-full">
  <!-- Tabs -->
  <div class="flex gap-2 border-b border-gray-200 mb-6">
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
    <div class="bg-white rounded-xl border border-gray-100 p-4 sm:p-6 md:p-8 text-center">
      <div class="w-14 h-14 sm:w-16 sm:h-16 mx-auto mb-3 sm:mb-4 rounded-full bg-gradient-to-br from-orange-400 to-red-500 flex items-center justify-center">
        <svg class="w-7 h-7 sm:w-8 sm:h-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 2.754 1.626 2.754H9.75c0 1.427.502 2.672 1.327 3.376.825.704 1.626 1.626 1.626 1.626H19.5c1.409 0 2.5-1.122 2.5-2.5 0-.826-.217-1.626-1.626-2.5l-9.303-3.376zM12 15.75h.007v.008H12v-.008zM7.5 15.75h.007v.008H7.5v-.008zm4.5 0h.007v.008h-.007v-.008zm4.5 0h.007v.008h-.007v-.008z" />
        </svg>
      </div>
      <h3 class="text-[16px] sm:text-[18px] font-semibold text-gray-900 mb-2">{t.vulhubSupport}</h3>
      <p class="text-[13px] sm:text-[14px] text-gray-500 mb-4 sm:mb-6">Vulhub 漏洞环境模块</p>
      <div class="bg-gray-50 rounded-lg p-4 sm:p-6 text-left">
        <p class="text-[12px] sm:text-[13px] text-gray-600">此模块功能开发中...</p>
      </div>
    </div>
  {:else if specialModuleTab === 'c2'}
    <div class="bg-white rounded-xl border border-gray-100 p-4 sm:p-6 md:p-8 text-center">
      <div class="w-14 h-14 sm:w-16 sm:h-16 mx-auto mb-3 sm:mb-4 rounded-full bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center">
        <svg class="w-7 h-7 sm:w-8 sm:h-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25a3 3 0 013 3m3 0a6 6 0 01-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1121.75 8.25z" />
        </svg>
      </div>
      <h3 class="text-[16px] sm:text-[18px] font-semibold text-gray-900 mb-2">{t.c2Scenes}</h3>
      <p class="text-[13px] sm:text-[14px] text-gray-500 mb-4 sm:mb-6">C2 场景管理模块</p>
      <div class="bg-gray-50 rounded-lg p-4 sm:p-6 text-left">
        <p class="text-[12px] sm:text-[13px] text-gray-600">此模块功能开发中...</p>
      </div>
    </div>
  {:else if specialModuleTab === 'ai'}
    <div class="bg-white rounded-xl border border-gray-100 p-4 sm:p-6 md:p-8">
      {#if templates.length > 0 && getTemplatesByCategory(templates, 'ai').length > 0}
        <div class="mb-4">
          <input
            type="text"
            placeholder={t.search || '搜索场景...'}
            class="w-full h-10 px-3 text-[13px] bg-gray-50 border-0 rounded-lg text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-purple-500 focus:ring-offset-1 transition-shadow"
            bind:value={aiSearchQuery}
          />
          {#if aiScenarios().length > 0}
            <div class="flex flex-wrap gap-2 mt-3">
              {#each aiScenarios() as scenario}
                <button
                  class="px-3 py-2 text-[12px] font-medium rounded-lg transition-all {selectedAIScenario?.name === scenario.name ? 'bg-purple-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}"
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
              class="px-3 py-1.5 text-[12px] font-medium rounded-lg transition-all {copied ? 'bg-green-600 text-white' : 'bg-gray-900 text-white hover:bg-gray-800'}"
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
        <div class="bg-gray-50 rounded-lg p-6 text-center">
          <p class="text-[13px] text-gray-500">请选择一个场景查看安装脚本</p>
        </div>
      {/if}
    </div>
  {/if}
  </div>
</div>
