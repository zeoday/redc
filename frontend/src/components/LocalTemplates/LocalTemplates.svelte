<script>

  import { onMount } from 'svelte';
  import { ListAllTemplates, GetTemplateVariables, RemoveTemplate, CopyTemplate, GetTemplateFiles, SaveTemplateFiles, CopyFileTo, ExportTemplates, ImportTemplates, CreateLocalTemplate, DeleteTemplateFile, ValidateTemplate } from '../../../wailsjs/go/main/App.js';
  import { selectFile, selectSaveFile } from '../../lib/file-dialog.js';
  import CodeEditor from '../CodeEditor/CodeEditor.svelte';

  // Translation object passed from parent component

  // ============================================================================
  // State Management
  // ============================================================================
  
  // Tab state for template categories
  let templateTab = $state('all');
  
  // Local templates list and loading state
  let { t } = $props();
  let localTemplates = $state([]);
  let localTemplatesLoading = $state(false);
  let localTemplatesSearch = $state('');
  
  // Template detail drawer state
  let localTemplateDetail = $state(null);
  let localTemplateVars = $state([]);
  let localTemplateVarsLoading = $state(false);
  
  // Delete confirmation modal state
  let deleteTemplateConfirm = $state({ show: false, name: '' });
  let deletingTemplate = $state({});
  
  // Batch operation state
  let selectedTemplates = $state(new Set());
  let batchOperating = $state(false);
  let batchDeleteConfirm = $state({ show: false, count: 0 });
  
  // Clone template modal state
  let cloneTemplateModal = $state({ show: false, source: '', target: '' });
  
  // Template editor modal state
  // - show: Whether the editor modal is visible
  // - name: The template name being edited
  // - files: Object mapping filename to content { [filename]: content }
  // - active: Currently selected filename in the editor
  // - saving: Whether a save operation is in progress
  // - error: Error message to display (if any)
  let templateEditor = $state({ show: false, name: '', files: {}, active: '', saving: false, error: '' });
  
  // Global error message
  let error = $state('');

  // ============================================================================
  // Template List Functions
  // ============================================================================

  /**
   * Load the list of local templates from the backend
   */
  async function loadLocalTemplates() {
    localTemplatesLoading = true;
    try {
      localTemplates = await ListAllTemplates() || [];
    } catch (e) {
      error = e.message || String(e);
      localTemplates = [];
    } finally {
      localTemplatesLoading = false;
    }
  }

  // ============================================================================
  // Template Detail Functions
  // ============================================================================

  /**
   * Show template detail drawer with variables
   * @param {Object} tmpl - The template object to show details for
   */
  async function showTemplateDetail(tmpl) {
    localTemplateDetail = tmpl;
    localTemplateVars = [];
    localTemplateVarsLoading = true;
    try {
      const vars = await GetTemplateVariables(tmpl.name);
      localTemplateVars = vars || [];
    } catch (e) {
      console.error('Failed to load template variables:', e);
      localTemplateVars = [];
    } finally {
      localTemplateVarsLoading = false;
    }
  }

  /**
   * Close the template detail drawer
   */
  function closeTemplateDetail() {
    localTemplateDetail = null;
    localTemplateVars = [];
  }

  // ============================================================================
  // Delete Template Functions
  // ============================================================================

  /**
   * Show delete confirmation modal
   * @param {string} name - The template name to delete
   */
  function showDeleteTemplateConfirm(name) {
    deleteTemplateConfirm = { show: true, name };
  }

  /**
   * Cancel delete operation and close confirmation modal
   */
  function cancelDeleteTemplate() {
    deleteTemplateConfirm = { show: false, name: '' };
  }

  /**
   * Confirm and execute template deletion
   */
  async function confirmDeleteTemplate() {
    const name = deleteTemplateConfirm.name;
    deleteTemplateConfirm = { show: false, name: '' };
    deletingTemplate[name] = true;
    deletingTemplate = deletingTemplate; // Trigger reactivity
    try {
      await RemoveTemplate(name);
      await loadLocalTemplates();
    } catch (e) {
      error = e.message || String(e);
    } finally {
      deletingTemplate[name] = false;
      deletingTemplate = deletingTemplate; // Trigger reactivity
    }
  }

  // ============================================================================
  // Clone Template Functions
  // ============================================================================

  /**
   * Show clone template modal
   * @param {Object} tmpl - The template object to clone
   */
  async function handleCloneTemplate(tmpl) {
    cloneTemplateModal = { show: true, source: tmpl.name, target: `${tmpl.name}-copy` };
  }

  /**
   * Cancel clone operation and close modal
   */
  function cancelCloneTemplate() {
    cloneTemplateModal = { show: false, source: '', target: '' };
  }

  /**
   * Confirm and execute template cloning
   */
  async function confirmCloneTemplate() {
    const targetName = cloneTemplateModal.target.trim();
    const sourceName = cloneTemplateModal.source;
    cloneTemplateModal = { show: false, source: '', target: '' };
    if (!targetName) return;
    try {
      await CopyTemplate(sourceName, targetName);
      await loadLocalTemplates();
    } catch (e) {
      error = e.message || String(e);
    }
  }

  // ============================================================================
  // Template Editor Functions
  // ============================================================================

  /**
   * Open the template editor modal and load template files
   * @param {Object} tmpl - The template object to edit
   * 
   * This function:
   * 1. Opens the editor modal
   * 2. Loads all template files from the backend
   * 3. Selects the first file as active
   * 4. Handles errors gracefully without closing the modal
   */
  async function openTemplateEditor(tmpl) {
    templateEditor = { show: true, name: tmpl.name, files: {}, active: '', saving: false, error: '' };
    try {
      const files = await GetTemplateFiles(tmpl.name);
      const names = Object.keys(files || {});
      templateEditor = {
        ...templateEditor,
        files: files || {},
        active: names.length > 0 ? names[0] : '',
      };
    } catch (e) {
      templateEditor = { ...templateEditor, error: e.message || String(e) };
    }
  }

  /**
   * Close the template editor modal
   * Note: This discards any unsaved changes
   */
  function closeTemplateEditor() {
    templateEditor = { show: false, name: '', files: {}, active: '', saving: false, error: '' };
  }

  /**
   * Save all template files to the backend
   * 
   * This function:
   * 1. Validates that a template name exists
   * 2. Sets saving state to show loading indicator
   * 3. Calls SaveTemplateFiles API with all file contents
   * 4. Handles errors without closing the modal (allows retry)
   * 5. Resets saving state when complete
   */
  async function saveTemplateEditor() {
    if (!templateEditor.name) return;
    templateEditor = { ...templateEditor, saving: true, error: '' };
    try {
      await SaveTemplateFiles(templateEditor.name, templateEditor.files);
      templateEditor = { ...templateEditor, saving: false };
    } catch (e) {
      templateEditor = { ...templateEditor, saving: false, error: e.message || String(e) };
    }
  }

  // ============================================================================
  // Batch Operation Functions
  // ============================================================================

  function toggleSelectAll() {
    if (allSelected) {
      selectedTemplates = new Set();
    } else {
      selectedTemplates = new Set(filteredLocalTemplates.map(t => t.name));
    }
  }

  function toggleSelectTemplate(name) {
    const newSet = new Set(selectedTemplates);
    if (newSet.has(name)) {
      newSet.delete(name);
    } else {
      newSet.add(name);
    }
    selectedTemplates = newSet;
  }

  function showBatchDeleteConfirm() {
    batchDeleteConfirm = { show: true, count: selectedTemplates.size };
  }

  function cancelBatchDelete() {
    batchDeleteConfirm = { show: false, count: 0 };
  }

  async function confirmBatchDelete() {
    batchDeleteConfirm = { show: false, count: 0 };
    batchOperating = true;
    
    const templateNames = Array.from(selectedTemplates);
    
    try {
      // Execute deletions in parallel
      await Promise.all(templateNames.map(name => RemoveTemplate(name)));
      selectedTemplates = new Set();
    } catch (e) {
      error = e.message || String(e);
    } finally {
      batchOperating = false;
      await loadLocalTemplates();
    }
  }

  // ============================================================================
  // Reactive Statements
  // ============================================================================

  /**
   * Determine template type based on path and properties
   */
  function getTemplateType(tmpl) {
    // Prefer the template type field from case.json
    if (tmpl.template) {
      const t2 = tmpl.template;
      if (t2 === 'base') return 'custom';
      if (t2 === 'userdata') return 'userdata';
      if (t2 === 'compose') return 'compose';
      if (t2 === 'preset') return 'preset';
    }
    // Fallback: detect from path name
    const name = tmpl.name || '';
    if (name.includes('base-templates/')) return 'custom';
    if (name.includes('userdata-templates/')) return 'userdata';
    if (name.includes('compose-templates/')) return 'compose';
    return 'preset';
  }

  /**
   * Filter and sort local templates based on search query and selected tab
   * Searches in: name, description, and module fields
   */
  let filteredLocalTemplates = $derived(localTemplates
    .filter(t => {
      // Filter by tab
      if (templateTab !== 'all') {
        const type = getTemplateType(t);
        if (templateTab !== type) return false;
      }
      // Filter by search query
      if (localTemplatesSearch) {
        const search = localTemplatesSearch.toLowerCase();
        return t.name.toLowerCase().includes(search) ||
          (t.description && t.description.toLowerCase().includes(search)) ||
          (t.module && t.module.toLowerCase().includes(search));
      }
      return true;
    })
    .sort((a, b) => a.name.localeCompare(b.name)));

  let allSelected = $derived(filteredLocalTemplates.length > 0 && selectedTemplates.size === filteredLocalTemplates.length);

  let someSelected = $derived(selectedTemplates.size > 0 && selectedTemplates.size < filteredLocalTemplates.length);

  let hasSelection = $derived(selectedTemplates.size > 0);

  // Export/Import state
  let exporting = $state(false);
  let importing = $state(false);
  let exportMessage = $state('');

  // Create template dialog state
  let createTemplateDialog = $state({ show: false, name: '', scaffold: 'preset', loading: false, error: '' });

  // File add inline input state (in templateEditor sidebar)
  let addingFile = $state({ show: false, name: '' });

  // Delete file confirmation state
  let deleteFileConfirm = $state({ show: false, fileName: '' });

  // Template validation state
  let validateResult = $state({ show: false, loading: false, templateName: '', result: null });
  
  // Export selected templates
  async function handleExportTemplates() {
    if (selectedTemplates.size === 0) return;
    
    exporting = true;
    exportMessage = '';
    try {
      const templateNames = Array.from(selectedTemplates);
      const savePath = await selectSaveFile(
        t.exportTemplates || '导出模板',
        'templates.zip'
      );
      
      if (!savePath) {
        exporting = false;
        return;
      }
      
      const zipPath = await ExportTemplates(templateNames);
      if (zipPath) {
        await CopyFileTo(zipPath, savePath);
        exportMessage = t.exportSuccess || '模板导出成功';
        setTimeout(() => { exportMessage = ''; }, 3000);
      }
    } catch (e) {
      console.error('Export failed:', e);
      exportMessage = '导出失败: ' + String(e);
    } finally {
      exporting = false;
    }
  }

  // Import templates
  async function handleImportTemplates() {
    importing = true;
    exportMessage = '';
    try {
      const filePath = await selectFile(t.importTemplates || '导入模板');
      
      if (!filePath) {
        importing = false;
        return;
      }
      
      const imported = await ImportTemplates(filePath);
      if (imported && imported.length > 0) {
        exportMessage = `${t.importSuccess || '模板导入成功'}: ${imported.join(', ')}`;
        // Refresh templates
        await loadLocalTemplates();
      }
      setTimeout(() => { exportMessage = ''; }, 5000);
    } catch (e) {
      console.error('Import failed:', e);
      exportMessage = '导入失败: ' + String(e);
    } finally {
      importing = false;
    }
  }


  // ============================================================================
  // Create Template Functions
  // ============================================================================

  function showCreateTemplateDialog() {
    createTemplateDialog = { show: true, name: '', scaffold: 'preset', loading: false, error: '' };
  }

  function cancelCreateTemplate() {
    createTemplateDialog = { show: false, name: '', scaffold: 'preset', loading: false, error: '' };
  }

  async function confirmCreateTemplate() {
    const name = createTemplateDialog.name.trim();
    if (!name) {
      createTemplateDialog = { ...createTemplateDialog, error: t.templateNameRequired || '请输入模板名称' };
      return;
    }
    // Validate name: allow letters, digits, -, _, /
    if (!/^[a-zA-Z0-9][a-zA-Z0-9_\-/]*$/.test(name)) {
      createTemplateDialog = { ...createTemplateDialog, error: t.templateNameInvalid || '名称只能包含字母、数字、-、_、/' };
      return;
    }
    if (name.includes('..')) {
      createTemplateDialog = { ...createTemplateDialog, error: t.templateNameInvalid || '名称包含非法字符' };
      return;
    }
    createTemplateDialog = { ...createTemplateDialog, loading: true, error: '' };
    try {
      await CreateLocalTemplate(name, createTemplateDialog.scaffold);
      createTemplateDialog = { show: false, name: '', scaffold: 'preset', loading: false, error: '' };
      await loadLocalTemplates();
      // Auto-open editor for the new template
      const newTmpl = localTemplates.find(t => t.name === name);
      if (newTmpl) {
        openTemplateEditor(newTmpl);
      }
    } catch (e) {
      createTemplateDialog = { ...createTemplateDialog, loading: false, error: e.message || String(e) };
    }
  }

  // ============================================================================
  // Template Editor File Management
  // ============================================================================

  function showAddFile() {
    addingFile = { show: true, name: '' };
  }

  function cancelAddFile() {
    addingFile = { show: false, name: '' };
  }

  function confirmAddFile() {
    const fname = addingFile.name.trim();
    if (!fname) return;
    // Prevent duplicates
    if (templateEditor.files[fname] !== undefined) {
      addingFile = { show: false, name: '' };
      return;
    }
    templateEditor.files[fname] = '';
    templateEditor = { ...templateEditor, active: fname };
    addingFile = { show: false, name: '' };
  }

  function showDeleteFileConfirm(fileName) {
    deleteFileConfirm = { show: true, fileName };
  }

  function cancelDeleteFile() {
    deleteFileConfirm = { show: false, fileName: '' };
  }

  async function confirmDeleteFile() {
    const fname = deleteFileConfirm.fileName;
    deleteFileConfirm = { show: false, fileName: '' };
    if (!fname) return;

    try {
      await DeleteTemplateFile(templateEditor.name, fname);
      const newFiles = { ...templateEditor.files };
      delete newFiles[fname];
      const names = Object.keys(newFiles);
      templateEditor = {
        ...templateEditor,
        files: newFiles,
        active: templateEditor.active === fname ? (names[0] || '') : templateEditor.active,
      };
    } catch (e) {
      templateEditor = { ...templateEditor, error: e.message || String(e) };
    }
  }

  // ============================================================================
  // Template Validation Functions
  // ============================================================================

  async function handleValidateTemplate(tmpl) {
    validateResult = { show: true, loading: true, templateName: tmpl.name, result: null };
    try {
      const result = await ValidateTemplate(tmpl.name);
      validateResult = { show: true, loading: false, templateName: tmpl.name, result };
    } catch (e) {
      validateResult = {
        show: true, loading: false, templateName: tmpl.name,
        result: { valid: false, error_count: 1, warning_count: 0, diagnostics: [{ severity: 'error', summary: e.message || String(e), detail: '', filename: '', line: 0 }] }
      };
    }
  }

  function closeValidateResult() {
    validateResult = { show: false, loading: false, templateName: '', result: null };
  }

  // ============================================================================
  // Lifecycle
  // ============================================================================

  /**
   * Load templates when component mounts
   */
  onMount(() => {
    loadLocalTemplates();
  });

  /**
   * Export refresh function for parent component to call
   * This allows parent components to trigger a template list refresh
   */
  export function refresh() {
    loadLocalTemplates();
  }

</script>

<div class="space-y-5">
  <!-- Search and Actions -->
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
          bind:value={localTemplatesSearch} 
        />
      </div>
      <button 
        class="h-10 px-5 text-gray-700 bg-white border border-gray-300 hover:bg-gray-50 text-[13px] font-medium rounded-lg transition-colors disabled:opacity-50"
        onclick={loadLocalTemplates}
        disabled={localTemplatesLoading}
      >
        {localTemplatesLoading ? t.loading : t.refresh}
      </button>
      <button 
        class="h-10 px-5 text-gray-700 bg-white border border-gray-300 hover:bg-gray-50 text-[13px] font-medium rounded-lg transition-colors disabled:opacity-50"
        onclick={handleImportTemplates}
        disabled={importing}
      >
        {importing ? t.loading : (t.importTemplates || '导入')}
      </button>
      <button 
        class="h-10 px-5 text-gray-700 bg-white border border-gray-300 hover:bg-gray-50 text-[13px] font-medium rounded-lg transition-colors disabled:opacity-50"
        onclick={handleExportTemplates}
        disabled={exporting || !hasSelection}
      >
        {exporting ? t.loading : (t.exportTemplates || '导出')}
      </button>
      <button 
        class="h-10 px-5 text-white bg-emerald-600 hover:bg-emerald-700 text-[13px] font-medium rounded-lg transition-colors"
        onclick={showCreateTemplateDialog}
      >
        {t.createTemplate || '新建模板'}
      </button>
    </div>
  </div>

  {#if exportMessage}
    <div class="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-lg text-[13px] text-blue-700">
      {exportMessage}
    </div>
  {/if}

  <!-- Tabs -->
  <div class="flex gap-2 border-b border-gray-100 mb-4">
    <button
      class="px-4 py-2 text-[13px] font-medium transition-colors {templateTab === 'all' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-600 hover:text-gray-900'}"
      onclick={() => templateTab = 'all'}
    >
      {t.allTemplates || '全部'}
    </button>
    <button
      class="px-4 py-2 text-[13px] font-medium transition-colors {templateTab === 'preset' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-600 hover:text-gray-900'}"
      onclick={() => templateTab = 'preset'}
    >
      {t.presetTemplates || '预定义模板'}
    </button>
    <button
      class="px-4 py-2 text-[13px] font-medium transition-colors {templateTab === 'custom' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-600 hover:text-gray-900'}"
      onclick={() => templateTab = 'custom'}
    >
      {t.customTemplates || '自定义模板'}
    </button>
    <button
      class="px-4 py-2 text-[13px] font-medium transition-colors {templateTab === 'userdata' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-600 hover:text-gray-900'}"
      onclick={() => templateTab = 'userdata'}
    >
      {t.userdataTemplates || 'Userdata模板'}
    </button>
    <button
      class="px-4 py-2 text-[13px] font-medium transition-colors {templateTab === 'compose' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-600 hover:text-gray-900'}"
      onclick={() => templateTab = 'compose'}
    >
      {t.composeTemplates || 'Compose模板'}
    </button>
  </div>

  {#if localTemplatesLoading}
    <div class="flex items-center justify-center h-64">
      <div class="w-6 h-6 border-2 border-gray-100 border-t-gray-900 rounded-full animate-spin"></div>
    </div>
  {:else}
    <!-- Template Table -->
    <div class="bg-white rounded-xl border border-gray-100 overflow-hidden">
      <!-- Batch Operations Bar -->
      {#if hasSelection}
        <div class="px-5 py-3 bg-blue-50 border-b border-blue-100 flex items-center justify-between">
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
            <button
              class="px-3 py-1.5 text-[12px] font-medium text-red-700 bg-red-50 rounded-md hover:bg-red-100 transition-colors disabled:opacity-50"
              onclick={showBatchDeleteConfirm}
              disabled={batchOperating}
            >
              {t.batchDelete}
            </button>
          </div>
        </div>
      {/if}
      
      <table class="w-full table-auto">
        <thead>
          <tr class="border-b border-gray-100">
            <th class="text-left pl-4 pr-1 py-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wide w-6">
              <input
                type="checkbox"
                class="w-4 h-4 rounded border-gray-300 text-gray-900 focus:ring-2 focus:ring-gray-900 focus:ring-offset-1 cursor-pointer"
                checked={allSelected}
                indeterminate={someSelected}
                onchange={toggleSelectAll}
              />
            </th>
            <th class="text-left px-4 py-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wide w-[140px]">{t.name}</th>
            <th class="text-left px-3 py-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wide w-[60px]">{t.version}</th>
            <th class="text-left px-3 py-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wide w-[140px]">{t.author}</th>
            <th class="text-left px-3 py-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wide w-[180px]">{t.module}</th>
            <th class="text-left px-3 py-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wide w-[320px]">{t.description}</th>
            <th class="text-right pl-4 pr-6 py-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wide w-[220px]">{t.actions}</th>
          </tr>
        </thead>
        <tbody>
          {#each filteredLocalTemplates as tmpl}
            <tr class="border-b border-gray-50 hover:bg-gray-50/50 transition-colors">
              <td class="pl-4 pr-1 py-3.5" onclick={(e) => e.stopPropagation()}>
                <input
                  type="checkbox"
                  class="w-4 h-4 rounded border-gray-300 text-gray-900 focus:ring-2 focus:ring-gray-900 focus:ring-offset-1 cursor-pointer"
                  checked={selectedTemplates.has(tmpl.name)}
                  onchange={() => toggleSelectTemplate(tmpl.name)}
                />
              </td>
              <td class="px-4 py-3.5">
                <span class="text-[13px] font-medium text-gray-900 break-all">{tmpl.name}</span>
              </td>
              <td class="px-3 py-3.5">
                <span class="text-[13px] text-gray-600">{tmpl.version || '-'}</span>
              </td>
              <td class="px-3 py-3.5">
                <span class="text-[13px] text-gray-600 break-words whitespace-normal block" title={tmpl.user || '-'}>{tmpl.user || '-'}</span>
              </td>
              <td class="px-3 py-3.5">
                {#if tmpl.module}
                  <span class="px-2 py-0.5 bg-blue-50 text-blue-600 text-[11px] font-medium rounded-full inline-block break-words whitespace-normal max-w-full" title={tmpl.module}>{tmpl.module}</span>
                {:else}
                  <span class="text-[13px] text-gray-400">-</span>
                {/if}
              </td>
              <td class="px-3 py-3.5 w-[320px]">
                <span class="text-[12px] text-gray-500 break-words whitespace-normal" title={tmpl.description}>{tmpl.description || '-'}</span>
              </td>
              <td class="pl-4 pr-6 py-3.5 text-right w-[220px]">
                <div class="flex flex-col gap-1.5 items-end">
                  <div class="flex items-center gap-1.5">
                    <!-- Icon buttons: clone, validate, delete -->
                    <button 
                      class="w-7 h-7 flex items-center justify-center text-gray-500 bg-gray-100 rounded-md hover:bg-gray-200 hover:text-gray-700 transition-colors"
                      onclick={() => handleCloneTemplate(tmpl)}
                      title={t.cloneTemplate}
                    >
                      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M8 16h8M8 12h8m-6 8h6a2 2 0 002-2V8l-6-6H8a2 2 0 00-2 2v16a2 2 0 002 2z" />
                      </svg>
                    </button>
                    <button 
                      class="w-7 h-7 flex items-center justify-center text-emerald-600 bg-emerald-50 rounded-md hover:bg-emerald-100 hover:text-emerald-700 transition-colors"
                      onclick={() => handleValidateTemplate(tmpl)}
                      title={t.validateTemplate || '语法检查'}
                    >
                      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </button>
                    {#if deletingTemplate[tmpl.name]}
                      <span class="w-7 h-7 flex items-center justify-center">
                        <div class="w-3.5 h-3.5 border-2 border-gray-200 border-t-amber-500 rounded-full animate-spin"></div>
                      </span>
                    {:else}
                      <button 
                        class="w-7 h-7 flex items-center justify-center text-red-500 bg-red-50 rounded-md hover:bg-red-100 hover:text-red-700 transition-colors"
                        onclick={() => showDeleteTemplateConfirm(tmpl.name)}
                        title={t.delete}
                      >
                        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </button>
                    {/if}
                    <!-- Text buttons: view params, edit -->
                    <button 
                      class="px-2.5 py-1 text-[12px] font-medium text-blue-700 bg-blue-50 rounded-md hover:bg-blue-100 transition-colors whitespace-nowrap"
                      onclick={() => showTemplateDetail(tmpl)}
                    >{t.viewParams}</button>
                    <button 
                      class="px-2.5 py-1 text-[12px] font-medium text-blue-700 bg-blue-50 rounded-md hover:bg-blue-100 transition-colors whitespace-nowrap"
                      onclick={() => openTemplateEditor(tmpl)}
                    >{t.editTemplate}</button>
                  </div>
                </div>
              </td>
            </tr>
          {:else}
            <tr>
              <td colspan="7" class="py-16">
                <div class="flex flex-col items-center justify-center text-gray-400">
                  <svg class="w-10 h-10 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12.75V12A2.25 2.25 0 014.5 9.75h15A2.25 2.25 0 0121.75 12v.75m-8.69-6.44l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44z" />
                  </svg>
                  <p class="text-[13px]">{t.noLocalTemplates}</p>
                  <button 
                    class="mt-2 text-[12px] text-blue-600 hover:underline"
                    onclick={() => { window.dispatchEvent(new CustomEvent('switchTab', { detail: 'registry' })); }}
                  >{t.goToRegistry}</button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}

  {#if error}
    <div class="flex items-center gap-3 px-4 py-3 bg-red-50 border border-red-100 rounded-lg">
      <svg class="w-4 h-4 text-red-500 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
      </svg>
      <span class="text-[13px] text-red-700 flex-1">{error}</span>
      <button class="text-red-400 hover:text-red-600" onclick={() => error = ''} aria-label="关闭错误">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  {/if}
</div>

<!-- Batch Delete Confirmation Modal -->
{#if batchDeleteConfirm.show}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 overflow-visible" onclick={cancelBatchDelete}>
    <div class="bg-white rounded-xl border border-gray-200 max-w-sm w-full mx-4 overflow-hidden" onclick={(e) => e.stopPropagation()}>
      <div class="px-6 py-5">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center">
            <svg class="w-5 h-5 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <div>
            <h3 class="text-[15px] font-semibold text-gray-900">{t.confirmBatchDelete}</h3>
            <p class="text-[13px] text-gray-500">{t.deleteWarning}</p>
          </div>
        </div>
        <p class="text-[13px] text-gray-600">
          {t.confirmBatchDeleteMessage} <span class="font-medium text-gray-900">{batchDeleteConfirm.count}</span> {t.templates}?
        </p>
      </div>
      <div class="px-6 py-4 bg-gray-50 flex justify-end gap-2">
        <button 
          class="px-4 py-2 text-[13px] font-medium text-gray-700 bg-white border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors"
          onclick={cancelBatchDelete}
        >{t.cancel}</button>
        <button 
          class="px-4 py-2 text-[13px] font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors"
          onclick={confirmBatchDelete}
        >{t.delete}</button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete Template Confirmation Modal -->
{#if deleteTemplateConfirm.show}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 overflow-visible" onclick={cancelDeleteTemplate}>
    <div class="bg-white rounded-xl border border-gray-200 max-w-sm w-full mx-4 overflow-hidden" onclick={(e) => e.stopPropagation()}>
      <div class="px-6 py-5">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center">
            <svg class="w-5 h-5 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <div>
            <h3 class="text-[15px] font-semibold text-gray-900">{t.confirmDelete}</h3>
            <p class="text-[13px] text-gray-500">{t.deleteWarning}</p>
          </div>
        </div>
        <p class="text-[13px] text-gray-600">
          {t.confirmDeleteTemplate} <span class="font-medium text-gray-900">"{deleteTemplateConfirm.name}"</span>?
        </p>
      </div>
      <div class="px-6 py-4 bg-gray-50 flex justify-end gap-2">
        <button 
          class="px-4 py-2 text-[13px] font-medium text-gray-700 bg-white border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors"
          onclick={cancelDeleteTemplate}
        >{t.cancel}</button>
        <button 
          class="px-4 py-2 text-[13px] font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors"
          onclick={confirmDeleteTemplate}
        >{t.delete}</button>
      </div>
    </div>
  </div>
{/if}

<!-- Clone Template Modal -->
{#if cloneTemplateModal.show}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 overflow-visible" onclick={cancelCloneTemplate}>
    <div class="bg-white rounded-xl border border-gray-200 max-w-sm w-full mx-4 overflow-hidden" onclick={(e) => e.stopPropagation()}>
      <div class="px-6 py-5">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center">
            <svg class="w-5 h-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16h8M8 12h8m-6 8h6a2 2 0 002-2V8a2 2 0 00-2-2h-2l-2-2H8a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
          <div>
            <h3 class="text-[15px] font-semibold text-gray-900">{t.cloneTitle}</h3>
            <p class="text-[13px] text-gray-500">{t.cloneHint}</p>
          </div>
        </div>
        <label for="cloneName" class="block text-[12px] font-medium text-gray-500 mb-1.5">{t.cloneName}</label>
        <input
          id="cloneName"
          type="text"
          class="w-full h-10 px-3 text-[13px] bg-gray-50 border-0 rounded-lg text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-gray-900 focus:ring-offset-1 transition-shadow"
          bind:value={cloneTemplateModal.target}
        />
      </div>
      <div class="px-6 py-4 bg-gray-50 flex justify-end gap-2">
        <button 
          class="px-4 py-2 text-[13px] font-medium text-gray-700 bg-white border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors"
          onclick={cancelCloneTemplate}
        >{t.cancel}</button>
        <button 
          class="px-4 py-2 text-[13px] font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors"
          onclick={confirmCloneTemplate}
        >{t.cloneTemplate}</button>
      </div>
    </div>
  </div>
{/if}

<!-- Template Detail Drawer -->
{#if localTemplateDetail}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/50 flex justify-end z-50" onclick={closeTemplateDetail}>
    <div class="w-full max-w-2xl bg-white h-full overflow-auto shadow-xl" onclick={(e) => e.stopPropagation()}>
      <div class="sticky top-0 bg-white border-b border-gray-100 px-6 py-4 flex items-center justify-between">
        <div>
          <h2 class="text-[16px] font-semibold text-gray-900">{localTemplateDetail.name}</h2>
          <p class="text-[12px] text-gray-500 mt-0.5">v{localTemplateDetail.version || '-'}</p>
        </div>
        <button 
          class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-gray-100 text-gray-400 hover:text-gray-600 transition-colors"
          onclick={closeTemplateDetail}
          aria-label="关闭详情"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      
      <div class="p-6 space-y-6">
        <!-- Template Info -->
        <div class="space-y-3">
          {#if localTemplateDetail.description}
            <div>
              <div class="text-[11px] font-medium text-gray-500 uppercase tracking-wide mb-1">{t.description}</div>
              <p class="text-[13px] text-gray-700">{localTemplateDetail.description}</p>
            </div>
          {/if}
          <div class="grid grid-cols-2 gap-4">
            <div>
              <div class="text-[11px] font-medium text-gray-500 uppercase tracking-wide mb-1">{t.author}</div>
              <p class="text-[13px] text-gray-900">{localTemplateDetail.user || '-'}</p>
            </div>
            <div>
              <div class="text-[11px] font-medium text-gray-500 uppercase tracking-wide mb-1">{t.module}</div>
              {#if localTemplateDetail.module}
                <span class="px-2 py-0.5 bg-blue-50 text-blue-600 text-[12px] font-medium rounded-full">{localTemplateDetail.module}</span>
              {:else}
                <p class="text-[13px] text-gray-400">-</p>
              {/if}
            </div>
          </div>
        </div>

        <!-- Template Parameters -->
        <div>
          <div class="text-[14px] font-semibold text-gray-900 mb-3">{t.templateParams}</div>
          {#if localTemplateVarsLoading}
            <div class="flex items-center justify-center py-8">
              <div class="w-5 h-5 border-2 border-gray-100 border-t-gray-900 rounded-full animate-spin"></div>
              <span class="ml-2 text-[13px] text-gray-500">{t.loadingParams}</span>
            </div>
          {:else if localTemplateVars.length === 0}
            <div class="py-8 text-center text-[13px] text-gray-400">
              {t.noParams}
            </div>
          {:else}
            <div class="border border-gray-100 rounded-lg overflow-x-auto">
              <table class="w-full text-[12px] min-w-[520px]">
                <thead>
                  <tr class="bg-gray-50 border-b border-gray-100">
                    <th class="text-left px-4 py-2.5 font-semibold text-gray-600">{t.paramName}</th>
                    <th class="text-left px-4 py-2.5 font-semibold text-gray-600">{t.paramType}</th>
                    <th class="text-left px-4 py-2.5 font-semibold text-gray-600">{t.paramDefault}</th>
                    <th class="text-center px-4 py-2.5 font-semibold text-gray-600">{t.paramRequired}</th>
                  </tr>
                </thead>
                <tbody>
                  {#each localTemplateVars as v}
                    <tr class="border-b border-gray-50 hover:bg-gray-50/50">
                      <td class="px-4 py-3">
                        <div class="font-medium text-gray-900">{v.name}</div>
                        {#if v.description}
                          <div class="text-[11px] text-gray-500 mt-0.5">{v.description}</div>
                        {/if}
                      </td>
                      <td class="px-4 py-3">
                        <code class="px-1.5 py-0.5 bg-gray-100 text-gray-700 rounded text-[11px]">{v.type}</code>
                      </td>
                      <td class="px-4 py-3">
                        {#if v.defaultValue}
                          <code class="text-gray-600">{v.defaultValue}</code>
                        {:else}
                          <span class="text-gray-400">-</span>
                        {/if}
                      </td>
                      <td class="px-4 py-3 text-center">
                        {#if v.required}
                          <span class="inline-flex items-center justify-center w-5 h-5 bg-emerald-100 text-emerald-600 rounded-full">
                            <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                            </svg>
                          </span>
                        {:else}
                          <span class="inline-flex items-center justify-center w-5 h-5 bg-gray-100 text-gray-400 rounded-full">
                            <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14" />
                            </svg>
                          </span>
                        {/if}
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {/if}
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Template Editor Modal -->
{#if templateEditor.show}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onclick={closeTemplateEditor}>
    <div class="bg-white rounded-xl border border-gray-200 max-w-6xl w-full h-[85vh] overflow-hidden" onclick={(e) => e.stopPropagation()}>
      <div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
        <div>
          <h3 class="text-[15px] font-semibold text-gray-900">{t.editTemplate}</h3>
          <p class="text-[12px] text-gray-500">{templateEditor.name}</p>
        </div>
        <div class="flex items-center gap-2">
          <button
            class="px-3 py-1.5 text-[12px] font-medium text-gray-600 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
            onclick={closeTemplateEditor}
          >{t.close}</button>
          <button
            class="px-3 py-1.5 text-[12px] font-medium text-amber-700 bg-amber-50 rounded-md hover:bg-amber-100 transition-colors"
            onclick={() => handleValidateTemplate({ name: templateEditor.name })}
          >{t.validateTemplate || '语法检查'}</button>
          <button
            class="px-3 py-1.5 text-[12px] font-medium text-white bg-emerald-500 rounded-md hover:bg-emerald-600 transition-colors disabled:opacity-50"
            onclick={saveTemplateEditor}
            disabled={templateEditor.saving}
          >{templateEditor.saving ? t.saving : t.saveTemplate}</button>
        </div>
      </div>
      <div class="flex h-[calc(100%-73px)]">
        <div class="w-64 border-r border-gray-100 overflow-auto flex flex-col">
          <div class="px-4 py-3 text-[12px] font-semibold text-gray-600">{t.templateFiles}</div>
          <div class="flex-1 overflow-auto">
            {#each Object.keys(templateEditor.files) as fname}
              <div class="group flex items-center">
                <button
                  class="flex-1 text-left px-4 py-2 text-[12px] transition-colors truncate {templateEditor.active === fname ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-50'}"
                  onclick={() => templateEditor = { ...templateEditor, active: fname }}
                  title={fname}
                >{fname}</button>
                <button
                  class="w-6 h-6 flex items-center justify-center text-gray-400 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity mr-1 flex-shrink-0"
                  onclick={() => showDeleteFileConfirm(fname)}
                  title={t.deleteFile || '删除文件'}
                >
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            {/each}
          </div>
          <!-- Add file input or button -->
          {#if addingFile.show}
            <div class="px-3 py-2 border-t border-gray-100">
              <input
                type="text"
                class="w-full h-8 px-2 text-[12px] bg-gray-50 border border-gray-200 rounded text-gray-900 placeholder-gray-400 focus:ring-1 focus:ring-emerald-500 focus:border-emerald-500"
                placeholder={t.newFileName || '文件名（如 main.tf）'}
                bind:value={addingFile.name}
                onkeydown={(e) => { if (e.key === 'Enter') confirmAddFile(); if (e.key === 'Escape') cancelAddFile(); }}
              />
              <div class="flex gap-1 mt-1">
                <button class="flex-1 h-6 text-[11px] text-gray-600 bg-gray-100 rounded hover:bg-gray-200 transition-colors" onclick={cancelAddFile}>{t.cancel}</button>
                <button class="flex-1 h-6 text-[11px] text-white bg-emerald-500 rounded hover:bg-emerald-600 transition-colors" onclick={confirmAddFile}>{t.confirm || '确认'}</button>
              </div>
            </div>
          {:else}
            <button
              class="mx-3 my-2 h-8 flex items-center justify-center gap-1 text-[12px] text-gray-500 border border-dashed border-gray-300 rounded hover:border-emerald-400 hover:text-emerald-600 transition-colors"
              onclick={showAddFile}
            >
              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
              </svg>
              {t.addFile || '新建文件'}
            </button>
          {/if}
        </div>
        <div class="flex-1 p-4 flex flex-col overflow-hidden">
          {#if templateEditor.error}
            <div class="text-[12px] text-red-500 mb-2 flex-shrink-0">{templateEditor.error}</div>
          {/if}
          {#if templateEditor.active}
            <!-- 
              CodeEditor Component Integration
              - filename: Current file name (used for syntax detection)
              - value: Current file content
              - on:change: Handle content changes
              
              Important: Must reassign templateEditor object to trigger Svelte reactivity
              after updating nested files object
            -->
            <div class="flex-1 min-h-0">
              <CodeEditor
                filename={templateEditor.active}
                value={templateEditor.files[templateEditor.active]}
                onchange={(newContent) => {
                  templateEditor.files[templateEditor.active] = newContent;
                  templateEditor = templateEditor; // Trigger reactivity
                }}
              />
            </div>
          {:else}
            <div class="text-[12px] text-gray-400">{t.noParams}</div>
          {/if}
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Create Template Dialog -->
{#if createTemplateDialog.show}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 overflow-visible" onclick={cancelCreateTemplate}>
    <div class="bg-white rounded-xl border border-gray-200 max-w-md w-full mx-4 overflow-hidden" onclick={(e) => e.stopPropagation()}>
      <div class="px-6 py-5">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 rounded-full bg-emerald-100 flex items-center justify-center">
            <svg class="w-5 h-5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
          </div>
          <div>
            <h3 class="text-[15px] font-semibold text-gray-900">{t.createTemplate || '新建模板'}</h3>
            <p class="text-[13px] text-gray-500">{t.createTemplateHint || '创建一个新的 Terraform 模板'}</p>
          </div>
        </div>

        <label for="newTemplateName" class="block text-[12px] font-medium text-gray-500 mb-1.5">{t.templateName || '模板名称'}</label>
        <input
          id="newTemplateName"
          type="text"
          class="w-full h-10 px-3 text-[13px] bg-gray-50 border-0 rounded-lg text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-emerald-500 focus:ring-offset-1 transition-shadow mb-4"
          placeholder={t.templateNamePlaceholder || '例如: my-template 或 myteam/ecs'}
          bind:value={createTemplateDialog.name}
          onkeydown={(e) => { if (e.key === 'Enter') confirmCreateTemplate(); }}
        />

        <label class="block text-[12px] font-medium text-gray-500 mb-2">{t.scaffoldType || '模板类型'}</label>
        <div class="grid grid-cols-2 gap-2">
          <button
            class="px-3 py-2.5 text-left rounded-lg border transition-colors {createTemplateDialog.scaffold === 'preset' ? 'border-emerald-500 bg-emerald-50 text-emerald-700' : 'border-gray-200 text-gray-600 hover:bg-gray-50'}"
            onclick={() => createTemplateDialog = { ...createTemplateDialog, scaffold: 'preset' }}
          >
            <div class="text-[13px] font-medium">{t.scaffoldPreset || '预定义模板'}</div>
            <div class="text-[11px] mt-0.5 opacity-70">{t.scaffoldPresetDesc || 'Terraform 基础骨架'}</div>
          </button>
          <button
            class="px-3 py-2.5 text-left rounded-lg border transition-colors {createTemplateDialog.scaffold === 'preset-userdata' ? 'border-emerald-500 bg-emerald-50 text-emerald-700' : 'border-gray-200 text-gray-600 hover:bg-gray-50'}"
            onclick={() => createTemplateDialog = { ...createTemplateDialog, scaffold: 'preset-userdata' }}
          >
            <div class="text-[13px] font-medium">{t.scaffoldPresetUserdata || '预定义 + Userdata'}</div>
            <div class="text-[11px] mt-0.5 opacity-70">{t.scaffoldPresetUserdataDesc || '含初始化脚本文件'}</div>
          </button>
          <button
            class="px-3 py-2.5 text-left rounded-lg border transition-colors {createTemplateDialog.scaffold === 'base' ? 'border-emerald-500 bg-emerald-50 text-emerald-700' : 'border-gray-200 text-gray-600 hover:bg-gray-50'}"
            onclick={() => createTemplateDialog = { ...createTemplateDialog, scaffold: 'base' }}
          >
            <div class="text-[13px] font-medium">{t.scaffoldBase || '自定义模板'}</div>
            <div class="text-[11px] mt-0.5 opacity-70">{t.scaffoldBaseDesc || '自定义部署场景'}</div>
          </button>
          <button
            class="px-3 py-2.5 text-left rounded-lg border transition-colors {createTemplateDialog.scaffold === 'userdata' ? 'border-emerald-500 bg-emerald-50 text-emerald-700' : 'border-gray-200 text-gray-600 hover:bg-gray-50'}"
            onclick={() => createTemplateDialog = { ...createTemplateDialog, scaffold: 'userdata' }}
          >
            <div class="text-[13px] font-medium">{t.scaffoldUserdata || 'Userdata 模板'}</div>
            <div class="text-[11px] mt-0.5 opacity-70">{t.scaffoldUserdataDesc || '仅含初始化脚本'}</div>
          </button>
          <button
            class="px-3 py-2.5 text-left rounded-lg border transition-colors col-span-2 {createTemplateDialog.scaffold === 'compose' ? 'border-emerald-500 bg-emerald-50 text-emerald-700' : 'border-gray-200 text-gray-600 hover:bg-gray-50'}"
            onclick={() => createTemplateDialog = { ...createTemplateDialog, scaffold: 'compose' }}
          >
            <div class="text-[13px] font-medium">{t.scaffoldCompose || 'Compose 模板'}</div>
            <div class="text-[11px] mt-0.5 opacity-70">{t.scaffoldComposeDesc || '多云编排部署'}</div>
          </button>
        </div>

        {#if createTemplateDialog.error}
          <div class="mt-3 text-[12px] text-red-600 bg-red-50 rounded-lg px-3 py-2">{createTemplateDialog.error}</div>
        {/if}
      </div>
      <div class="px-6 py-4 bg-gray-50 flex justify-end gap-2">
        <button
          class="px-4 py-2 text-[13px] font-medium text-gray-700 bg-white border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors"
          onclick={cancelCreateTemplate}
          disabled={createTemplateDialog.loading}
        >{t.cancel}</button>
        <button
          class="px-4 py-2 text-[13px] font-medium text-white bg-emerald-600 rounded-lg hover:bg-emerald-700 transition-colors disabled:opacity-50"
          onclick={confirmCreateTemplate}
          disabled={createTemplateDialog.loading}
        >{createTemplateDialog.loading ? t.loading : (t.create || '创建')}</button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete File Confirmation Modal -->
{#if deleteFileConfirm.show}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-[60] overflow-visible" onclick={cancelDeleteFile}>
    <div class="bg-white rounded-xl border border-gray-200 max-w-sm w-full mx-4 overflow-hidden" onclick={(e) => e.stopPropagation()}>
      <div class="px-6 py-5">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center">
            <svg class="w-5 h-5 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </div>
          <div>
            <h3 class="text-[15px] font-semibold text-gray-900">{t.deleteFileTitle || '删除文件'}</h3>
            <p class="text-[13px] text-gray-500">{t.deleteFileHint || '确定要删除此文件吗？此操作不可撤销。'}</p>
          </div>
        </div>
        <div class="px-3 py-2 bg-gray-50 rounded-lg">
          <code class="text-[13px] text-gray-800">{deleteFileConfirm.fileName}</code>
        </div>
      </div>
      <div class="px-6 py-4 bg-gray-50 flex justify-end gap-2">
        <button
          class="px-4 py-2 text-[13px] font-medium text-gray-700 bg-white border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors"
          onclick={cancelDeleteFile}
        >{t.cancel}</button>
        <button
          class="px-4 py-2 text-[13px] font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors"
          onclick={confirmDeleteFile}
        >{t.delete}</button>
      </div>
    </div>
  </div>
{/if}

<!-- Template Validation Result Modal -->
{#if validateResult.show}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-[60] overflow-visible" onclick={closeValidateResult}>
    <div class="bg-white rounded-xl border border-gray-200 max-w-lg w-full mx-4 overflow-hidden" onclick={(e) => e.stopPropagation()}>
      <div class="px-6 py-5">
        <div class="flex items-center gap-3 mb-4">
          {#if validateResult.loading}
            <div class="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center">
              <div class="w-5 h-5 border-2 border-gray-300 border-t-gray-700 rounded-full animate-spin"></div>
            </div>
            <div>
              <h3 class="text-[15px] font-semibold text-gray-900">{t.validating || '验证中...'}</h3>
              <p class="text-[13px] text-gray-500">{validateResult.templateName}</p>
            </div>
          {:else if validateResult.result?.valid}
            <div class="w-10 h-10 rounded-full bg-emerald-100 flex items-center justify-center">
              <svg class="w-5 h-5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <div>
              <h3 class="text-[15px] font-semibold text-emerald-700">{t.validatePassed || '验证通过'}</h3>
              <p class="text-[13px] text-gray-500">{validateResult.templateName}</p>
            </div>
          {:else}
            <div class="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center">
              <svg class="w-5 h-5 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
              </svg>
            </div>
            <div>
              <h3 class="text-[15px] font-semibold text-red-700">{t.validateFailed || '验证失败'}</h3>
              <p class="text-[13px] text-gray-500">
                {validateResult.templateName}
                {#if validateResult.result}
                  — {validateResult.result.error_count} {t.errors || '个错误'}{#if validateResult.result.warning_count > 0}, {validateResult.result.warning_count} {t.warnings || '个警告'}{/if}
                {/if}
              </p>
            </div>
          {/if}
        </div>

        {#if !validateResult.loading && validateResult.result}
          {#if validateResult.result.valid}
            <div class="px-4 py-3 bg-emerald-50 rounded-lg text-[13px] text-emerald-700">
              {t.validatePassedMsg || '模板语法和参数配置正确，可以正常使用。'}
            </div>
          {:else if validateResult.result.diagnostics?.length > 0}
            <div class="space-y-2 max-h-[300px] overflow-auto">
              {#each validateResult.result.diagnostics as diag}
                <div class="px-4 py-3 rounded-lg text-[12px] {diag.severity === 'error' ? 'bg-red-50 border border-red-100' : 'bg-amber-50 border border-amber-100'}">
                  <div class="flex items-start gap-2">
                    {#if diag.severity === 'error'}
                      <span class="px-1.5 py-0.5 bg-red-100 text-red-700 rounded text-[10px] font-bold flex-shrink-0 mt-0.5">ERROR</span>
                    {:else}
                      <span class="px-1.5 py-0.5 bg-amber-100 text-amber-700 rounded text-[10px] font-bold flex-shrink-0 mt-0.5">WARN</span>
                    {/if}
                    <div class="flex-1 min-w-0">
                      <div class="font-medium text-gray-900">{diag.summary}</div>
                      {#if diag.detail}
                        <div class="text-gray-600 mt-1 break-words whitespace-pre-wrap">{diag.detail}</div>
                      {/if}
                      {#if diag.filename}
                        <div class="text-gray-400 mt-1">
                          📄 {diag.filename}{#if diag.line > 0}:{diag.line}{/if}
                        </div>
                      {/if}
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        {/if}
      </div>
      {#if !validateResult.loading}
        <div class="px-6 py-4 bg-gray-50 flex justify-end">
          <button
            class="px-4 py-2 text-[13px] font-medium text-gray-700 bg-white border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors"
            onclick={closeValidateResult}
          >{t.close}</button>
        </div>
      {/if}
    </div>
  </div>
{/if}
