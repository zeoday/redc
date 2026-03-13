<script>

  import { onMount, onDestroy } from 'svelte';
  import { FetchRegistryTemplates, PullTemplate, ListTemplates, FetchTemplateReadme, GetLanguage } from '../../../wailsjs/go/main/App.js';
  import { normalizeVersion, compareVersions, hasUpdate } from '../../utils/version.js';

  // Registry state
let { t } = $props();
  let registryTemplates = $state([]);
  let registryLoading = $state(false);
  let registryError = $state('');
  let registrySearch = $state('');
  let pullingTemplates = $state({});
  let registryNotice = $state({ type: '', message: '' });
  let registryNoticeTimer = null;
  let templates = $state([]);

  // Readme modal state
  let readmeModal = $state({ show: false, content: '', html: '', loading: false, templateName: '' });

  // Simple markdown to HTML converter
  function parseMarkdown(md) {
    if (!md) return '';
    
    // First escape HTML (but preserve code blocks placeholder)
    const codeBlocks = [];
    let idx = 0;
    
    // Replace code blocks with placeholders to protect them
    md = md.replace(/```[\s\S]*?```/g, (match) => {
      const placeholder = `__CODEBLOCK_${idx}__`;
      codeBlocks.push(match);
      idx++;
      return placeholder;
    });
    
    // Now escape HTML in the rest
    md = md.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    
    // Step 3: Restore code blocks with escaped content
    codeBlocks.forEach((block, i) => {
      // Extract code content (remove ``` and optional language)
      let code = block.replace(/^```[a-z]*\n?/, '').replace(/```$/, '').trim();
      // Escape any remaining markdown characters in code
      code = code.replace(/^# /gm, '&#35; ').replace(/^\* /gm, '&#42; ').replace(/^- /gm, '&#45; ');
      const codeHtml = `<pre class="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto my-3 text-[12px] font-mono leading-relaxed"><code>${code}</code></pre>`;
      md = md.replace(`__CODEBLOCK_${i}__`, codeHtml);
    });
    
    // Process inline code
    md = md.replace(/`([^`]+)`/g, '<code class="bg-gray-100 px-1.5 py-0.5 rounded text-[12px] font-mono text-pink-600">$1</code>');
    
    // Process headers (only at line start)
    md = md.replace(/^#### (.*$)/gm, '<h4 class="text-sm font-semibold mt-5 mb-2 text-gray-800">$1</h4>');
    md = md.replace(/^### (.*$)/gm, '<h3 class="text-sm font-semibold mt-5 mb-2 text-gray-800">$1</h3>');
    md = md.replace(/^## (.*$)/gm, '<h2 class="text-base font-bold mt-6 mb-3 text-gray-900">$1</h2>');
    md = md.replace(/^# (.*$)/gm, '<h1 class="text-lg font-bold mt-6 mb-3 text-gray-900">$1</h1>');
    
    // Process bold and italic
    md = md.replace(/\*\*\*(.*?)\*\*\*/g, '<strong><em>$1</em></strong>');
    md = md.replace(/\*\*(.*?)\*\*/g, '<strong class="font-semibold">$1</strong>');
    md = md.replace(/\*(.*?)\*/g, '<em>$1</em>');
    md = md.replace(/___(.*?)___/g, '<strong><em>$1</em></strong>');
    md = md.replace(/__(.*?)__/g, '<strong class="font-semibold">$1</strong>');
    md = md.replace(/_(.*?)_/g, '<em>$1</em>');
    
    // Process links
    md = md.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" class="text-blue-600 hover:text-blue-800 hover:underline underline-offset-2" target="_blank" rel="noopener">$1</a>');
    
    // Process blockquotes
    md = md.replace(/^> (.*$)/gm, '<blockquote class="border-l-4 border-gray-300 pl-4 py-1 my-3 text-gray-600 italic">$1</blockquote>');
    
    // Process horizontal rules
    md = md.replace(/^---$/gm, '<hr class="my-6 border-gray-200">');
    md = md.replace(/^\*\*\*$/gm, '<hr class="my-6 border-gray-200">');
    
    // Process unordered lists - more specific pattern to avoid matching code
    md = md.replace(/^(\* |-)(?!\*)(.*$)/gm, '<li class="ml-4 list-disc text-gray-700">$2</li>');
    
    // Process ordered lists
    md = md.replace(/^\d+\.(?!\.)(.*$)/gm, '<li class="ml-4 list-decimal text-gray-700">$1</li>');
    
    // Remove newlines between list items to allow proper grouping
    md = md.replace(/<\/li>\n<li/g, '</li><li');
    md = md.replace(/<\/li>\s*<br>/g, '</li>');
    md = md.replace(/<br>\s*<li/g, '<li');
    
    // Wrap consecutive list items in ul/ol tags
    md = md.replace(/(<li[^>]*>[^<]*<\/li>)+/g, (match) => {
      // Clean up any remaining <br> tags
      match = match.replace(/<br\s*\/?>/g, '');
      if (match.includes('list-disc')) {
        return `<ul class="my-2">${match}</ul>`;
      } else {
        return `<ol class="my-2 list-inside">${match}</ol>`;
      }
    });
    
    // Process paragraphs - split by double newlines
    const paragraphs = md.split(/\n\n+/);
    let result = paragraphs.map(p => {
      p = p.trim();
      if (!p) return '';
      // Skip if already wrapped in HTML tags (including lists)
      if (p.match(/^<(h[1-4]|ul|ol|pre|blockquote|hr)/i)) return p;
      // Wrap in paragraph
      return `<p class="my-2 text-gray-700 leading-relaxed">${p.replace(/\n/g, '<br>')}</p>`;
    }).join('\n');
    
    return result;
  }

  // Batch operation state
  let selectedTemplates = $state(new Set());
  let batchOperating = $state(false);
  let batchPullConfirm = $state({ show: false, count: 0 });
  let batchUpdateConfirm = $state({ show: false, count: 0 });

  let filteredRegistryTemplates = $derived(registryTemplates
    .filter(t => 
      !registrySearch || 
      t.name.toLowerCase().includes(registrySearch.toLowerCase()) ||
      (t.author && t.author.toLowerCase().includes(registrySearch.toLowerCase())) ||
      (t.description && t.description.toLowerCase().includes(registrySearch.toLowerCase())) ||
      (t.tags && t.tags.some(tag => tag.toLowerCase().includes(registrySearch.toLowerCase())))
    )
    .sort((a, b) => {
      // Installed templates first
      if (a.installed && !b.installed) return -1;
      if (!a.installed && b.installed) return 1;
      // Then sort by name alphabetically
      return a.name.localeCompare(b.name);
    }));

  let allSelected = $derived(filteredRegistryTemplates.length > 0 && selectedTemplates.size === filteredRegistryTemplates.length);

  let someSelected = $derived(selectedTemplates.size > 0 && selectedTemplates.size < filteredRegistryTemplates.length);

  let hasSelection = $derived(selectedTemplates.size > 0);


  // Get templates that can be pulled (not installed)
  let canPullTemplates = $derived(Array.from(selectedTemplates).filter(name => {
    const tmpl = registryTemplates.find(t => t.name === name);
    return tmpl && !tmpl.installed;
  }));


  // Get templates that can be updated (installed and has update)
  let canUpdateTemplates = $derived(Array.from(selectedTemplates).filter(name => {
    const tmpl = registryTemplates.find(t => t.name === name);
    return tmpl && tmpl.installed && hasUpdate(tmpl);
  }));


  function setRegistryNotice(type, message, autoClear = true) {
    registryNotice = { type, message };
    if (registryNoticeTimer) {
      clearTimeout(registryNoticeTimer);
      registryNoticeTimer = null;
    }
    if (autoClear && message) {
      registryNoticeTimer = setTimeout(() => {
        registryNotice = { type: '', message: '' };
      }, 3000);
    }
  }

  async function loadRegistryTemplates() {
    registryLoading = true;
    registryError = '';
    try {
      registryTemplates = await FetchRegistryTemplates('');
    } catch (e) {
      registryError = e.message || String(e);
      registryTemplates = [];
    } finally {
      registryLoading = false;
    }
  }

  async function syncLocalTemplates() {
    try {
      const list = await ListTemplates();
      templates = list || [];
    } catch (e) {
      console.error('Failed to sync local templates:', e);
    }
  }

  async function handlePullTemplate(templateName, force = false) {
    pullingTemplates[templateName] = true;
    pullingTemplates = pullingTemplates;
    setRegistryNotice('info', `${t.pulling} ${templateName}`, false);
    try {
      await PullTemplate(templateName, force);
      // Refresh registry and local templates after successful pull
      await loadRegistryTemplates();
      await syncLocalTemplates();
      registryTemplates = (registryTemplates || []).map((tmpl) => {
        if (tmpl.name !== templateName) return tmpl;
        const latest = tmpl.latest || tmpl.localVersion;
        return {
          ...tmpl,
          installed: true,
          localVersion: latest || tmpl.localVersion
        };
      });
      setRegistryNotice('success', `${t.pullSuccess}: ${templateName}`);
    } catch (e) {
      setRegistryNotice('error', `${t.pullFailed}: ${templateName}`);
    } finally {
      pullingTemplates[templateName] = false;
      pullingTemplates = pullingTemplates;
    }
  }

  async function handleShowReadme(templateName) {
    readmeModal = { show: true, content: '', html: '', loading: true, templateName };
    try {
      const lang = await GetLanguage();
      const content = await FetchTemplateReadme(templateName, lang || 'zh');
      const html = parseMarkdown(content);
      readmeModal = { ...readmeModal, content, html, loading: false };
    } catch (e) {
      readmeModal = { ...readmeModal, content: e.message || String(e), html: `<p class="text-red-500">${e.message || String(e)}</p>`, loading: false };
    }
  }

  function closeReadmeModal() {
    readmeModal = { show: false, content: '', html: '', loading: false, templateName: '' };
  }

  // Listen for refresh events to update pulling status
  $effect(() => {
    if (registryTemplates.length > 0) {
      // Reset pulling status when templates are refreshed
      for (const t of registryTemplates) {
        if (t.installed && pullingTemplates[t.name]) {
          pullingTemplates[t.name] = false;
        }
      }
    }
  });

  // ============================================================================
  // Batch Operation Functions
  // ============================================================================

  function toggleSelectAll() {
    if (allSelected) {
      selectedTemplates = new Set();
    } else {
      selectedTemplates = new Set(filteredRegistryTemplates.map(t => t.name));
    }
  }

  function toggleSelectTemplate(templateName) {
    const newSet = new Set(selectedTemplates);
    if (newSet.has(templateName)) {
      newSet.delete(templateName);
    } else {
      newSet.add(templateName);
    }
    selectedTemplates = newSet;
  }

  function showBatchPullConfirm() {
    batchPullConfirm = { show: true, count: canPullTemplates.length };
  }

  function cancelBatchPull() {
    batchPullConfirm = { show: false, count: 0 };
  }

  async function confirmBatchPull() {
    batchPullConfirm = { show: false, count: 0 };
    batchOperating = true;

    try {
      await Promise.all(canPullTemplates.map(name => handlePullTemplate(name, false)));
      selectedTemplates = new Set();
    } catch (e) {
      setRegistryNotice('error', e.message || String(e));
    } finally {
      batchOperating = false;
      await loadRegistryTemplates();
    }
  }

  function showBatchUpdateConfirm() {
    batchUpdateConfirm = { show: true, count: canUpdateTemplates.length };
  }

  function cancelBatchUpdate() {
    batchUpdateConfirm = { show: false, count: 0 };
  }

  async function confirmBatchUpdate() {
    batchUpdateConfirm = { show: false, count: 0 };
    batchOperating = true;

    try {
      await Promise.all(canUpdateTemplates.map(name => handlePullTemplate(name, true)));
      selectedTemplates = new Set();
    } catch (e) {
      setRegistryNotice('error', e.message || String(e));
    } finally {
      batchOperating = false;
      await loadRegistryTemplates();
    }
  }

  onMount(() => {
    loadRegistryTemplates();
  });

  onDestroy(() => {
    if (registryNoticeTimer) {
      clearTimeout(registryNoticeTimer);
      registryNoticeTimer = null;
    }
  });

  // Export refresh function for parent component
  export function refresh() {
    loadRegistryTemplates();
  }

</script>

<div class="space-y-5">
  <!-- Search -->
  <div class="bg-white rounded-xl border border-gray-100 p-5">
    <div class="flex items-center gap-4">
      <div class="flex-1 relative">
        <svg class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input 
          type="text" 
          placeholder={t.search}
          class="w-full h-10 pl-10 pr-4 text-[13px] bg-gray-50 border-0 rounded-lg text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-gray-900 focus:ring-offset-1 transition-shadow"
          bind:value={registrySearch} 
        />
      </div>
      <button 
        class="h-10 px-5 bg-blue-600 text-white text-[13px] font-medium rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
        onclick={toggleSelectAll}
        disabled={registryLoading || filteredRegistryTemplates.length === 0}
      >
        {allSelected ? t.clearSelection : t.selectAll || '全选'}
      </button>
      <button 
        class="h-10 px-5 text-gray-700 bg-white border border-gray-300 hover:bg-gray-50 text-[13px] font-medium rounded-lg transition-colors disabled:opacity-50"
        onclick={loadRegistryTemplates}
        disabled={registryLoading}
      >
        {registryLoading ? t.loading : t.refreshRepo}
      </button>
    </div>
    {#if registryNotice.message}
      <div class="mt-3 flex items-center gap-2 rounded-lg border px-3 py-2 text-[12px]
        {registryNotice.type === 'success' ? 'bg-emerald-50 border-emerald-100 text-emerald-700' : registryNotice.type === 'error' ? 'bg-red-50 border-red-100 text-red-700' : 'bg-amber-50 border-amber-100 text-amber-700'}">
        {#if registryNotice.type === 'info'}
          <div class="w-3.5 h-3.5 border-2 border-amber-200 border-t-amber-600 rounded-full animate-spin"></div>
        {/if}
        <span class="flex-1 truncate">{registryNotice.message}</span>
        <button class="text-gray-400 hover:text-gray-600 transition-colors cursor-pointer" onclick={() => setRegistryNotice('', '')} aria-label="关闭通知">
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    {/if}
  </div>

  {#if registryError}
    <div class="flex items-center gap-3 px-4 py-3 bg-red-50 border border-red-100 rounded-lg">
      <svg class="w-4 h-4 text-red-500 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
      </svg>
      <span class="text-[13px] text-red-700 flex-1">{registryError}</span>
      <button class="text-red-400 hover:text-red-600 cursor-pointer" onclick={() => registryError = ''} aria-label="关闭错误">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  {/if}

  {#if registryLoading}
    <div class="flex items-center justify-center h-64">
      <div class="w-6 h-6 border-2 border-gray-100 border-t-gray-900 rounded-full animate-spin"></div>
    </div>
  {:else}
    <!-- Batch Operations Bar -->
    {#if hasSelection}
      <div class="bg-white rounded-xl border border-gray-100 p-5 mb-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <span class="text-[13px] font-medium text-blue-900">
              {t.selected} {selectedTemplates.size} {t.items}
            </span>
            <button
              class="text-[12px] text-blue-600 hover:text-blue-800 underline"
              onclick={() => { selectedTemplates = new Set(); }}
            >
              {t.clearSelection}
            </button>
          </div>
          <div class="flex items-center gap-2">
            {#if canPullTemplates.length > 0}
              <button
                class="px-3 py-1.5 text-[12px] font-medium text-white bg-gray-900 rounded-md hover:bg-gray-800 transition-colors disabled:opacity-50"
                onclick={showBatchPullConfirm}
                disabled={batchOperating}
              >
                {t.batchPull} ({canPullTemplates.length})
              </button>
            {/if}
            {#if canUpdateTemplates.length > 0}
              <button
                class="px-3 py-1.5 text-[12px] font-medium text-blue-700 bg-blue-50 rounded-md hover:bg-blue-100 transition-colors disabled:opacity-50"
                onclick={showBatchUpdateConfirm}
                disabled={batchOperating}
              >
                {t.batchUpdate} ({canUpdateTemplates.length})
              </button>
            {/if}
          </div>
        </div>
      </div>
    {/if}

    <!-- Template Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {#each filteredRegistryTemplates as tmpl}
        <div class="bg-white rounded-xl border border-gray-100 p-5 hover:shadow-md transition-shadow relative">
          <!-- Checkbox -->
          <div class="absolute top-4 left-4">
            <input
              type="checkbox"
              class="w-4 h-4 rounded border-gray-300 text-gray-900 focus:ring-2 focus:ring-gray-900 focus:ring-offset-1 cursor-pointer"
              checked={selectedTemplates.has(tmpl.name)}
              onchange={() => toggleSelectTemplate(tmpl.name)}
              onclick={(e) => e.stopPropagation()}
            />
          </div>
          
          <div class="pl-6">
            <div class="flex items-start justify-between mb-3">
              <div class="flex-1 min-w-0">
                <h3 class="text-[14px] font-semibold text-gray-900 truncate">{tmpl.name}</h3>
                <p class="text-[12px] text-gray-500 mt-0.5">v{tmpl.latest}</p>
              </div>
              {#if tmpl.installed}
                <span class="inline-flex items-center gap-1 px-2 py-0.5 bg-emerald-50 text-emerald-600 text-[11px] font-medium rounded-full">
                  <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                  {t.installed}
                </span>
              {/if}
            </div>
            
            {#if tmpl.description}
              <p class="text-[12px] text-gray-600 mb-3 line-clamp-2">{tmpl.description}</p>
            {/if}
            
            {#if tmpl.tags && tmpl.tags.length > 0}
              <div class="flex flex-wrap gap-1 mb-3">
                {#each tmpl.tags.slice(0, 3) as tag}
                  <span class="px-2 py-0.5 bg-gray-100 text-gray-600 text-[10px] rounded-full">{tag}</span>
                {/each}
                {#if tmpl.tags.length > 3}
                  <span class="px-2 py-0.5 bg-gray-100 text-gray-500 text-[10px] rounded-full">+{tmpl.tags.length - 3}</span>
                {/if}
              </div>
            {/if}
            
            <div class="flex items-center justify-between pt-3 border-t border-gray-100">
              <div class="text-[11px] text-gray-500">
                {#if tmpl.author}by {tmpl.author}{/if}
                {#if tmpl.installed && hasUpdate(tmpl)}
                  <span class="ml-2 text-amber-600">
                    v{tmpl.localVersion} → v{tmpl.latest}
                  </span>
                {/if}
              </div>
              {#if pullingTemplates[tmpl.name]}
                <span class="inline-flex items-center gap-2 px-3 py-1.5 text-[12px] font-medium text-amber-600">
                  <span class="w-3 h-3 border-2 border-amber-200 border-t-amber-600 rounded-full animate-spin"></span>
                  {t.pulling}
                </span>
              {:else}
                <div class="flex gap-2">
                  <button 
                    class="px-3 py-1.5 text-[12px] font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
                    onclick={() => handleShowReadme(tmpl.name)}
                  >{t.viewReadme || '查看'}</button>
                  {#if tmpl.installed && hasUpdate(tmpl)}
                    <button 
                      class="px-3 py-1.5 text-[12px] font-medium text-blue-600 bg-blue-50 rounded-lg hover:bg-blue-100 transition-colors"
                      onclick={() => handlePullTemplate(tmpl.name, true)}
                    >{t.update}</button>
                  {:else if !tmpl.installed}
                    <button 
                      class="px-3 py-1.5 text-[12px] font-medium text-white bg-gray-900 rounded-lg hover:bg-gray-800 transition-colors"
                      onclick={() => handlePullTemplate(tmpl.name, false)}
                    >{t.pull}</button>
                  {/if}
                </div>
              {/if}
            </div>
          </div>
        </div>
      {:else}
        <div class="col-span-full py-16 text-center">
          <svg class="w-10 h-10 mx-auto mb-3 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
            <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 7.5l-.625 10.632a2.25 2.25 0 01-2.247 2.118H6.622a2.25 2.25 0 01-2.247-2.118L3.75 7.5M10 11.25h4M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z" />
          </svg>
          <p class="text-[13px] text-gray-500">
            {#if registrySearch}
              {t.noMatch}
            {:else}
              {t.clickRefresh}
            {/if}
          </p>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Batch Pull Confirmation Modal -->
{#if batchPullConfirm.show}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 overflow-visible" onclick={cancelBatchPull}>
    <div class="bg-white rounded-xl border border-gray-200 shadow-xl max-w-sm w-full mx-4 overflow-hidden" onclick={(e) => e.stopPropagation()}>
      <div class="px-6 py-5">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center">
            <svg class="w-5 h-5 text-gray-900" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
          </div>
          <div>
            <h3 class="text-[15px] font-semibold text-gray-900">{t.confirmBatchPull}</h3>
            <p class="text-[13px] text-gray-500">{t.pulling}</p>
          </div>
        </div>
        <p class="text-[13px] text-gray-600">
          {t.confirmBatchPullMessage} <span class="font-medium text-gray-900">{batchPullConfirm.count}</span> {t.templates}?
        </p>
      </div>
      <div class="px-6 py-4 bg-gray-50 flex justify-end gap-2">
        <button 
          class="px-4 py-2 text-[13px] font-medium text-gray-700 bg-white border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors"
          onclick={cancelBatchPull}
        >{t.cancel}</button>
        <button 
          class="px-4 py-2 text-[13px] font-medium text-white bg-gray-900 rounded-lg hover:bg-gray-800 transition-colors"
          onclick={confirmBatchPull}
        >{t.pull}</button>
      </div>
    </div>
  </div>
{/if}

<!-- Batch Update Confirmation Modal -->
{#if batchUpdateConfirm.show}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 overflow-visible" onclick={cancelBatchUpdate}>
    <div class="bg-white rounded-xl border border-gray-200 shadow-xl max-w-sm w-full mx-4 overflow-hidden" onclick={(e) => e.stopPropagation()}>
      <div class="px-6 py-5">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center">
            <svg class="w-5 h-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </div>
          <div>
            <h3 class="text-[15px] font-semibold text-gray-900">{t.confirmBatchUpdate}</h3>
            <p class="text-[13px] text-gray-500">{t.update}</p>
          </div>
        </div>
        <p class="text-[13px] text-gray-600">
          {t.confirmBatchUpdateMessage} <span class="font-medium text-gray-900">{batchUpdateConfirm.count}</span> {t.templates}?
        </p>
      </div>
      <div class="px-6 py-4 bg-gray-50 flex justify-end gap-2">
        <button 
          class="px-4 py-2 text-[13px] font-medium text-gray-700 bg-white border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors"
          onclick={cancelBatchUpdate}
        >{t.cancel}</button>
        <button 
          class="px-4 py-2 text-[13px] font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors"
          onclick={confirmBatchUpdate}
        >{t.update}</button>
      </div>
    </div>
  </div>
{/if}

{#if readmeModal.show}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 overflow-visible" onclick={closeReadmeModal} role="dialog" aria-modal="true" aria-labelledby="readme-modal-title" tabindex="-1">
    <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
    <div class="bg-white rounded-xl border border-gray-200 shadow-xl max-w-3xl w-full mx-4 max-h-[80vh] flex flex-col" onclick={(e) => e.stopPropagation()}>
      <div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
        <div>
          <h2 id="readme-modal-title" class="text-[15px] font-medium text-gray-900">{t.readme || 'README'}</h2>
          <p class="text-[12px] text-gray-500">{readmeModal.templateName}</p>
        </div>
        <button class="text-gray-400 hover:text-gray-600 transition-colors cursor-pointer" onclick={closeReadmeModal} aria-label={t.close || '关闭'}>
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <div class="px-6 py-4 overflow-auto flex-1">
        {#if readmeModal.loading}
          <div class="flex items-center justify-center py-8">
            <svg class="animate-spin h-6 w-6 text-gray-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          </div>
        {:else}
          <div class="text-[13px] text-gray-700">
            {@html readmeModal.html || readmeModal.content}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}
