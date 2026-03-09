<script>

  import { ComposePreview, ComposeUp, ComposeDown, SelectComposeFile } from '../../../wailsjs/go/main/App.js';
  import { loadComposeTemplates } from '../../lib/composeTemplates.js';
  import { onMount } from 'svelte';
  import ELK from 'elkjs/lib/elk.bundled.js';

  let { t, onTabChange } = $props();

  let composeTemplates = $state([]);
  let selectedTemplate = $state(null);
  let templatesLoading = $state(true);

  onMount(async () => {
    composeTemplates = await loadComposeTemplates();
    templatesLoading = false;
  });

  // State
  let composeFilePath = $state('');
  let composeProfiles = $state('');
  let composeSummary = $state(null);
  let composeLoading = $state(false);
  let composeActionLoading = $state(false);
  let composeError = $state('');
  let hasManuallyPreviewed = $state(false);
  let lastPreviewedPath = $state('');

  // Topology state
  let composeTopoModal = $state(false);
  let elkNodes = $state([]);
  let elkEdges = $state([]);
  let svgViewBox = $state('0 0 800 400');
  let topoZoom = $state(1);

  function getProviderColor(provider) {
    const p = (provider || '').toLowerCase();
    if (p.includes('aliyun') || p.includes('alicloud')) return { border: '#f97316', text: '#ea580c', label: 'Aliyun' };
    if (p.includes('aws') || p.includes('amazon')) return { border: '#f59e0b', text: '#d97706', label: 'AWS' };
    if (p.includes('tencent') || p.includes('tencentcloud')) return { border: '#3b82f6', text: '#2563eb', label: 'Tencent' };
    if (p.includes('volcengine') || p.includes('volcano')) return { border: '#10b981', text: '#059669', label: 'Volcano' };
    if (p.includes('huawei') || p.includes('hcloud')) return { border: '#ef4444', text: '#dc2626', label: 'Huawei' };
    if (p.includes('azure')) return { border: '#0ea5e9', text: '#0284c7', label: 'Azure' };
    if (p.includes('gcp') || p.includes('google')) return { border: '#8b5cf6', text: '#7c3aed', label: 'GCP' };
    return { border: '#6b7280', text: '#4b5563', label: 'Cloud' };
  }

  async function openComposeTopology() {
    composeTopoModal = true;
    elkNodes = [];
    elkEdges = [];
    topoZoom = 1;
    await layoutComposeTopology();
  }

  async function layoutComposeTopology() {
    if (!composeSummary?.services?.length) return;
    try {
      const elk = new ELK();
      const services = composeSummary.services;
      const NODE_W = 200;
      const NODE_H = 58;

      // Build name set for edge filtering (using rawName and name)
      const nameSet = new Set();
      services.forEach(s => { nameSet.add(s.name); nameSet.add(s.rawName); });

      // Build edges from dependsOn
      const rawEdges = [];
      services.forEach(s => {
        if (s.dependsOn && s.dependsOn.length > 0) {
          s.dependsOn.forEach(dep => {
            if (nameSet.has(dep)) {
              rawEdges.push({ from: s.name, to: dep });
            }
          });
        }
      });

      const nodeCount = services.length;
      const nodeSpacing = nodeCount > 8 ? '25' : '35';
      const layerSpacing = nodeCount > 8 ? '40' : '55';

      const graph = {
        id: 'root',
        layoutOptions: {
          'elk.algorithm': 'layered',
          'elk.direction': 'DOWN',
          'elk.spacing.nodeNode': nodeSpacing,
          'elk.layered.spacing.nodeNodeBetweenLayers': layerSpacing,
          'elk.padding': '[top=15,left=15,bottom=15,right=15]',
          'elk.layered.nodePlacement.strategy': 'BRANDES_KOEPF',
        },
        children: services.map(s => ({ id: s.name, width: NODE_W, height: NODE_H })),
        edges: rawEdges.map((e, i) => ({ id: `e${i}`, sources: [e.from], targets: [e.to] })),
      };

      const layout = /** @type {any} */ (await elk.layout(graph));

      const svcMap = {};
      services.forEach(s => { svcMap[s.name] = s; });

      const newNodes = (layout.children || []).map(n => {
        const svc = svcMap[n.id] || {};
        const color = getProviderColor(svc.provider);
        return {
          id: n.id,
          x: n.x, y: n.y, w: n.width, h: n.height,
          svc,
          color,
          label: svc.name || n.id,
          sublabel: svc.template || '',
          replicas: svc.replicas || 1,
        };
      });

      const newEdges = (layout.edges || []).map(e => {
        const sections = e.sections || [];
        if (sections.length > 0) {
          const s = sections[0];
          return { id: e.id, startPoint: s.startPoint, endPoint: s.endPoint, bendPoints: s.bendPoints || [] };
        }
        const src = newNodes.find(n => n.id === (e.sources?.[0]));
        const tgt = newNodes.find(n => n.id === (e.targets?.[0]));
        if (src && tgt) {
          return { id: e.id, startPoint: { x: src.x + src.w / 2, y: src.y + src.h }, endPoint: { x: tgt.x + tgt.w / 2, y: tgt.y }, bendPoints: [] };
        }
        return null;
      }).filter(Boolean);

      const padding = 40;
      const maxX = Math.max(...newNodes.map(n => n.x + n.w), 400) + padding * 2;
      const maxY = Math.max(...newNodes.map(n => n.y + n.h), 200) + padding * 2;

      elkNodes = newNodes;
      elkEdges = newEdges;
      svgViewBox = `0 0 ${maxX} ${maxY}`;
    } catch (e) {
      console.error('ELK layout failed:', e);
    }
  }

  function edgePath(edge) {
    const { startPoint, endPoint, bendPoints } = edge;
    if (bendPoints.length === 0) {
      const midY = (startPoint.y + endPoint.y) / 2;
      return `M ${startPoint.x} ${startPoint.y} C ${startPoint.x} ${midY}, ${endPoint.x} ${midY}, ${endPoint.x} ${endPoint.y}`;
    }
    let d = `M ${startPoint.x} ${startPoint.y}`;
    const pts = [startPoint, ...bendPoints, endPoint];
    for (let i = 1; i < pts.length; i++) {
      const prev = pts[i - 1]; const curr = pts[i];
      const midY = (prev.y + curr.y) / 2;
      d += ` C ${prev.x} ${midY}, ${curr.x} ${midY}, ${curr.x} ${curr.y}`;
    }
    return d;
  }

  // Functions
  function handleSelectTemplate() {
    if (!selectedTemplate) {
      composeFilePath = '';
      return;
    }
    const template = composeTemplates.find(t => t.name === selectedTemplate);
    if (template && template.path) {
      composeFilePath = template.path + '/redc-compose.yaml';
      hasManuallyPreviewed = true;
      previewCompose();
    }
  }

  async function handleBrowseFile() {
    try {
      const selectedPath = await SelectComposeFile();
      if (selectedPath) {
        composeFilePath = selectedPath;
        hasManuallyPreviewed = true;
        previewCompose();
      }
    } catch (e) {
      console.error('Failed to select file:', e);
    }
  }

  function parseComposeProfiles(value) {
    if (!value) return [];
    return value
      .split(',')
      .map(v => v.trim())
      .filter(Boolean);
  }

  // Auto-preview when file path or profiles change (only after first manual preview)
  let timer = null;
  
  async function previewCompose() {
    if (!composeFilePath) {
      composeError = '';
      composeSummary = null;
      return;
    }
    
    hasManuallyPreviewed = true;
    composeLoading = true;
    composeError = '';
    try {
      composeSummary = await ComposePreview(composeFilePath, parseComposeProfiles(composeProfiles));
    } catch (e) {
      composeError = e.message || String(e);
      composeSummary = null;
    } finally {
      composeLoading = false;
    }
  }

  $effect(() => {
    if (hasManuallyPreviewed && composeFilePath && composeFilePath !== lastPreviewedPath) {
      if (timer) clearTimeout(timer);
      timer = setTimeout(() => {
        lastPreviewedPath = composeFilePath;
        previewCompose();
      }, 500);
    }
  });

  export async function handleComposeUp() {
    if (!composeFilePath) {
      composeError = t.composeFile + ' ' + t.paramRequired;
      return;
    }
    
    composeActionLoading = true;
    composeError = '';
    try {
      await ComposeUp(composeFilePath, parseComposeProfiles(composeProfiles));
    } catch (e) {
      composeError = e.message || String(e);
    } finally {
      composeActionLoading = false;
    }
  }

  export async function handleComposeDown() {
    if (!composeFilePath) {
      composeError = t.composeFile + ' ' + t.paramRequired;
      return;
    }
    
    composeActionLoading = true;
    composeError = '';
    try {
      await ComposeDown(composeFilePath, parseComposeProfiles(composeProfiles));
    } catch (e) {
      composeError = e.message || String(e);
    } finally {
      composeActionLoading = false;
    }
  }

</script>

<div class="max-w-3xl lg:max-w-5xl xl:max-w-full space-y-5">
  {#if !templatesLoading && composeTemplates.length === 0}
    <div class="bg-blue-50 border border-blue-100 rounded-xl p-5">
      <div class="flex items-start gap-3">
        <svg class="w-5 h-5 text-blue-500 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
        </svg>
        <div class="flex-1">
          <p class="text-[13px] text-blue-700">{t.noComposeTemplatesHint}</p>
          <button 
            class="mt-3 h-8 px-4 bg-blue-500 text-white text-[12px] font-medium rounded-lg hover:bg-blue-600 transition-colors cursor-pointer"
            onclick={() => onTabChange && onTabChange('registry')}
          >
            {t.noComposeTemplatesHintButton}
          </button>
        </div>
      </div>
    </div>
  {/if}
  {#if composeTemplates.length > 0}
    <div class="bg-white rounded-xl border border-gray-100 p-5">
      <div class="flex items-center gap-4">
        <div class="flex-1">
          <label for="templateSelect" class="block text-[12px] font-medium text-gray-500 mb-1.5">{t.selectTemplate || '选择模板'}</label>
          <select
            id="templateSelect"
            class="w-full h-10 px-3 text-[13px] bg-gray-50 border-0 rounded-lg text-gray-900 focus:ring-2 focus:ring-gray-900 focus:ring-offset-1 transition-shadow"
            bind:value={selectedTemplate}
            onchange={handleSelectTemplate}
          >
            <option value={null}>{templatesLoading ? (t.loading || '加载中...') : (t.selectTemplate || '请选择模板')}</option>
            {#each composeTemplates as tmpl}
              <option value={tmpl.name}>{tmpl.nameZh || tmpl.name}</option>
            {/each}
          </select>
        </div>
        {#if selectedTemplate}
          {@const currentTemplate = composeTemplates.find(t => t.name === selectedTemplate)}
          {#if currentTemplate?.description}
            <div class="flex-1 text-[12px] text-gray-500">
              <span class="font-medium">{(t.description || '描述')}:</span> {currentTemplate.description}
            </div>
          {/if}
          {#if currentTemplate?.path}
            <div class="flex-1 text-[12px] text-gray-400 font-mono mt-1">
              <span class="font-medium">{(t.path || '路径')}:</span> {currentTemplate.path}
            </div>
          {/if}
        {/if}
      </div>
    </div>
  {/if}
  
  <div class="bg-white rounded-xl border border-gray-100 p-5">
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <label for="composeFile" class="block text-[12px] font-medium text-gray-500 mb-1.5">{t.composeFile}</label>
        <div class="flex gap-2">
          <input
            id="composeFile"
            type="text"
            placeholder="redc-compose.yaml"
            class="flex-1 h-10 px-3 text-[13px] bg-gray-50 border-0 rounded-lg text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-gray-900 focus:ring-offset-1 transition-shadow font-mono"
            bind:value={composeFilePath}
          />
          <button
            class="h-10 px-4 bg-gray-100 text-gray-700 text-[12px] font-medium rounded-lg hover:bg-gray-200 transition-colors"
            onclick={handleBrowseFile}
          >
            {t.browseFile}
          </button>
        </div>
      </div>
      <div>
        <label for="composeProfiles" class="block text-[12px] font-medium text-gray-500 mb-1.5">{t.composeProfiles}</label>
        <input
          id="composeProfiles"
          type="text"
          placeholder="prod,dev"
          class="w-full h-10 px-3 text-[13px] bg-gray-50 border-0 rounded-lg text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-gray-900 focus:ring-offset-1 transition-shadow"
          bind:value={composeProfiles}
        />
      </div>
    </div>
    <div class="mt-4 flex flex-wrap gap-2">
      <button
        class="h-9 px-4 text-gray-700 bg-white border border-gray-300 hover:bg-gray-50 text-[12px] font-medium rounded-lg transition-colors disabled:opacity-50"
        onclick={previewCompose}
        disabled={composeLoading}
      >
        {composeLoading ? t.loading : t.previewCompose}
      </button>
      <button
        class="h-9 px-4 bg-emerald-500 text-white text-[12px] font-medium rounded-lg hover:bg-emerald-600 transition-colors disabled:opacity-50"
        onclick={handleComposeUp}
        disabled={composeActionLoading}
      >
        {composeActionLoading ? t.processing : t.composeUp}
      </button>
      <button
        class="h-9 px-4 bg-red-500 text-white text-[12px] font-medium rounded-lg hover:bg-red-600 transition-colors disabled:opacity-50"
        onclick={handleComposeDown}
        disabled={composeActionLoading}
      >
        {composeActionLoading ? t.processing : t.composeDown}
      </button>
    </div>
    {#if composeError}
      <div class="mt-3 text-[12px] text-red-500">{composeError}</div>
    {/if}
  </div>

  <div class="bg-white rounded-xl border border-gray-100 p-5">
    <div class="flex items-center justify-between mb-4">
      <div class="text-[14px] font-semibold text-gray-900">{t.composePreview}</div>
      {#if composeSummary?.services?.length > 0}
        <button
          class="flex items-center gap-1.5 h-7 px-3 text-[11px] font-medium text-gray-600 bg-gray-50 hover:bg-gray-100 border border-gray-200 rounded-lg transition-colors cursor-pointer"
          onclick={openComposeTopology}
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3.75v4.5m0-4.5h4.5m-4.5 0L9 9M3.75 20.25v-4.5m0 4.5h4.5m-4.5 0L9 15M20.25 3.75h-4.5m4.5 0v4.5m0-4.5L15 9m5.25 11.25h-4.5m4.5 0v-4.5m0 4.5L15 15" />
          </svg>
          {t.composeTopology || '拓扑视图'}
        </button>
      {/if}
    </div>
    {#if composeLoading}
      <div class="flex items-center justify-center h-24">
        <div class="w-6 h-6 border-2 border-gray-100 border-t-gray-900 rounded-full animate-spin"></div>
      </div>
    {:else if composeSummary && composeSummary.services && composeSummary.services.length > 0}
      <div class="border border-gray-100 rounded-lg overflow-hidden">
        <table class="w-full text-[12px]">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="text-left px-4 py-2.5 font-semibold text-gray-600">{t.serviceName}</th>
              <th class="text-left px-4 py-2.5 font-semibold text-gray-600">{t.serviceTemplate}</th>
              <th class="text-left px-4 py-2.5 font-semibold text-gray-600">{t.serviceProvider}</th>
              <th class="text-left px-4 py-2.5 font-semibold text-gray-600">{t.serviceDepends}</th>
              <th class="text-right px-4 py-2.5 font-semibold text-gray-600">{t.serviceReplicas}</th>
            </tr>
          </thead>
          <tbody>
            {#each composeSummary.services as svc}
              <tr class="border-b border-gray-50">
                <td class="px-4 py-3 text-gray-700">{svc.name}</td>
                <td class="px-4 py-3 text-gray-700">{svc.template}</td>
                <td class="px-4 py-3 text-gray-700">{svc.provider || '-'}</td>
                <td class="px-4 py-3 text-gray-700">{(svc.dependsOn && svc.dependsOn.length > 0) ? svc.dependsOn.join(', ') : '-'}</td>
                <td class="px-4 py-3 text-right text-gray-700">{svc.replicas || 1}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <div class="py-12 text-center text-[12px] text-gray-400">{t.noScene}</div>
    {/if}
  </div>
</div>

{#if composeTopoModal}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onclick={() => composeTopoModal = false}>
    <div class="bg-white rounded-2xl shadow-2xl w-full max-w-[1100px] mx-4 max-h-[90vh] flex flex-col overflow-hidden" onclick={e => e.stopPropagation()}>
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-100">
        <div class="flex items-center gap-3">
          <svg class="w-5 h-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3.75v4.5m0-4.5h4.5m-4.5 0L9 9M3.75 20.25v-4.5m0 4.5h4.5m-4.5 0L9 15M20.25 3.75h-4.5m4.5 0v4.5m0-4.5L15 9m5.25 11.25h-4.5m4.5 0v-4.5m0 4.5L15 15" />
          </svg>
          <span class="text-[15px] font-semibold text-gray-900">{t.composeTopology || '编排服务拓扑'}</span>
          <span class="text-[13px] text-gray-400">— {composeFilePath || 'redc-compose.yaml'}</span>
        </div>
        <button class="p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded transition-colors" onclick={() => composeTopoModal = false}>
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>

      <div class="flex-1 overflow-y-auto px-6 py-4 space-y-4">
        <!-- Stats bar -->
        <div class="flex items-center gap-4 text-[12px] text-gray-500">
          <span class="font-medium text-gray-700">{composeSummary?.total || 0} {t.composeSvcCount || '个服务'}</span>
          {#if elkEdges.length > 0}
            <span>· {elkEdges.length} {t.composeDepsCount || '条依赖'}</span>
          {/if}
          {#if composeSummary?.services}
            {@const providers = [...new Set(composeSummary.services.map(s => (s.provider||'').split(/[,\s]/)[0]).filter(Boolean))]}
            {#if providers.length > 0}
              <span>·</span>
              {#each providers as p}
                {@const color = getProviderColor(p)}
                <span class="inline-flex items-center gap-1">
                  <span class="w-2 h-2 rounded-full inline-block" style="background:{color.border}"></span>
                  <span>{p}</span>
                </span>
              {/each}
            {/if}
          {/if}
        </div>

        <!-- Provider legend -->
        {#if composeSummary?.services}
          {@const providerGroups = composeSummary.services.reduce((acc, s) => {
            const p = (s.provider||'').split(/[,\s]/)[0] || 'unknown';
            if (!acc[p]) acc[p] = 0;
            acc[p]++;
            return acc;
          }, {})}
          <div class="flex flex-wrap gap-2">
            {#each Object.entries(providerGroups) as [p, cnt]}
              {@const color = getProviderColor(p)}
              <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[11px] font-medium border" style="border-color:{color.border}; color:{color.text}; background:{color.border}18">
                <span class="w-2 h-2 rounded-full inline-block" style="background:{color.border}"></span>
                {p} × {cnt}
              </span>
            {/each}
          </div>
        {/if}

        <!-- Topology SVG -->
        {#if elkNodes.length > 0}
          <div class="relative">
            <!-- Zoom controls -->
            <div class="absolute top-2 right-2 z-10 flex items-center gap-1 bg-white/90 backdrop-blur rounded-lg border border-gray-200 shadow-sm px-1 py-0.5">
              <button class="w-7 h-7 flex items-center justify-center text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded transition-colors cursor-pointer font-bold text-[14px]" onclick={() => topoZoom = Math.max(0.3, topoZoom - 0.1)}>−</button>
              <span class="text-[11px] text-gray-500 w-10 text-center">{Math.round(topoZoom * 100)}%</span>
              <button class="w-7 h-7 flex items-center justify-center text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded transition-colors cursor-pointer font-bold text-[14px]" onclick={() => topoZoom = Math.min(3, topoZoom + 0.1)}>+</button>
              <button class="w-7 h-7 flex items-center justify-center text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded transition-colors cursor-pointer" onclick={() => topoZoom = 1} title="重置">
                <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M9 9V4.5M9 9H4.5M9 9L3.75 3.75M9 15v4.5M9 15H4.5M9 15l-5.25 5.25M15 9h4.5M15 9V4.5M15 9l5.25-5.25M15 15h4.5M15 15v4.5m0-4.5l5.25 5.25" /></svg>
              </button>
            </div>
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div class="bg-gray-50 rounded-xl border border-gray-200 overflow-auto" style="max-height: 65vh;"
              onwheel={(e) => { e.preventDefault(); topoZoom = Math.min(3, Math.max(0.3, topoZoom + (e.deltaY > 0 ? -0.08 : 0.08))); }}>
              <svg viewBox={svgViewBox} preserveAspectRatio="xMidYMid meet" class="w-full" style="min-height: 300px; transform: scale({topoZoom}); transform-origin: center top;">
                <defs>
                  <marker id="compose-arrowhead" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto">
                    <polygon points="0 0, 8 3, 0 6" fill="#94a3b8" />
                  </marker>
                </defs>

                {#each elkEdges as edge}
                  <path d={edgePath(edge)} fill="none" stroke="#cbd5e1" stroke-width="1.5" marker-end="url(#compose-arrowhead)" class="transition-colors hover:stroke-blue-400" />
                {/each}

                {#each elkNodes as node}
                  <g transform="translate({node.x}, {node.y})">
                    <rect width={node.w} height={node.h} rx="8" ry="8"
                      fill="white" stroke={node.color.border} stroke-width="2" class="drop-shadow-sm" />
                    <circle cx="14" cy="22" r="4" fill={node.color.border} />
                    <text x="26" y="18" font-size="11" font-weight="600" fill="#374151" font-family="system-ui, sans-serif">
                      {node.label}
                    </text>
                    <text x="26" y="32" font-size="9" fill="#9ca3af" font-family="system-ui, sans-serif">
                      {node.sublabel}
                    </text>
                    <text x="26" y="46" font-size="8" fill={node.color.text} font-family="system-ui, sans-serif" font-style="italic">
                      {node.svc?.provider || ''}
                    </text>
                    {#if node.replicas > 1}
                      <rect x={node.w - 30} y="6" width="22" height="14" rx="7" fill={node.color.border} />
                      <text x={node.w - 19} y="17" font-size="9" font-weight="700" fill="white" text-anchor="middle" font-family="system-ui, sans-serif">
                        ×{node.replicas}
                      </text>
                    {/if}
                  </g>
                {/each}
              </svg>
            </div>
          </div>
        {:else if composeSummary?.services?.length > 0}
          <!-- Fallback: service list -->
          <div class="space-y-2">
            {#each composeSummary.services as svc}
              {@const color = getProviderColor(svc.provider)}
              <div class="flex items-center gap-3 px-3 py-2 rounded-lg border" style="border-color:{color.border}40; background:{color.border}08">
                <span class="w-2.5 h-2.5 rounded-full flex-shrink-0" style="background:{color.border}"></span>
                <span class="text-[12px] font-medium text-gray-700">{svc.name}</span>
                <span class="text-[11px] text-gray-400">{svc.template}</span>
                {#if svc.dependsOn?.length > 0}
                  <span class="text-[10px] text-gray-400">← {svc.dependsOn.join(', ')}</span>
                {/if}
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Footer -->
      <div class="px-6 py-4 border-t border-gray-100 flex justify-end">
        <button class="h-9 px-5 bg-gray-100 text-gray-700 text-[12px] font-medium rounded-lg hover:bg-gray-200 transition-colors" onclick={() => composeTopoModal = false}>
          {t.close || '关闭'}
        </button>
      </div>
    </div>
  </div>
{/if}
