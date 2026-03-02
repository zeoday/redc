<script lang="ts">
  import { SetShowWelcomeDialog } from '../../../wailsjs/wailsjs/go/main/App.js';

  let { t, show = false, onClose } = $props();
  
  let dontShowAgain = $state(false);

  $effect(() => {
    if (show) {
      dontShowAgain = false;
    }
  });

  async function handleClose() {
    if (dontShowAgain) {
      try {
        await SetShowWelcomeDialog(false);
      } catch (e) {
        console.error('Failed to save welcome dialog setting:', e);
      }
    }
    onClose?.();
  }
</script>

{#if show}
  <div class="fixed inset-0 z-50 flex items-center justify-center">
    <!-- Backdrop -->
    <div 
      class="absolute inset-0 bg-black/40"
      onclick={handleClose}
      onkeydown={(e) => e.key === 'Escape' && handleClose()}
      role="button"
      tabindex="0"
    ></div>
    
    <!-- Dialog -->
    <div class="relative bg-white rounded-xl border border-gray-100 max-w-md w-full mx-4 overflow-hidden">
      <!-- Header -->
      <div class="px-5 py-4 border-b border-gray-100">
        <h2 class="text-[15px] font-medium text-gray-900">
          {t.welcomeTitle || '欢迎使用 RedC'}
        </h2>
      </div>
      
      <!-- Content -->
      <div class="px-5 py-4 space-y-3">
        <!-- Feature 1 -->
        <div class="flex gap-3">
          <div class="flex-shrink-0 w-8 h-8 bg-blue-50 rounded-lg flex items-center justify-center">
            <svg class="w-4 h-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <div>
            <h3 class="text-[13px] font-medium text-gray-900">{t.welcomeFeature1 || '一键部署云资源'}</h3>
            <p class="text-[12px] text-gray-500">{t.welcomeFeature1Desc || '支持多种云厂商，快速创建和管理云资源场景'}</p>
          </div>
        </div>
        
        <!-- Feature 2 -->
        <div class="flex gap-3">
          <div class="flex-shrink-0 w-8 h-8 bg-emerald-50 rounded-lg flex items-center justify-center">
            <svg class="w-4 h-4 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
          </div>
          <div>
            <h3 class="text-[13px] font-medium text-gray-900">{t.welcomeFeature2 || '成本优化分析'}</h3>
            <p class="text-[12px] text-gray-500">{t.welcomeFeature2Desc || 'AI 智能分析资源使用，提供成本优化建议'}</p>
          </div>
        </div>
        
        <!-- Feature 3 -->
        <div class="flex gap-3">
          <div class="flex-shrink-0 w-8 h-8 bg-gray-50 rounded-lg flex items-center justify-center">
            <svg class="w-4 h-4 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
            </svg>
          </div>
          <div>
            <h3 class="text-[13px] font-medium text-gray-900">{t.welcomeFeature3 || '本地模板管理'}</h3>
            <p class="text-[12px] text-gray-500">{t.welcomeFeature3Desc || '支持自定义模板，满足个性化需求'}</p>
          </div>
        </div>
      </div>
      
      <!-- Footer -->
      <div class="px-5 py-3 border-t border-gray-100 flex items-center justify-between">
        <label class="flex items-center gap-2 cursor-pointer">
          <input 
            type="checkbox" 
            bind:checked={dontShowAgain}
            class="w-4 h-4 rounded border-gray-300 text-gray-900 focus:ring-2 focus:ring-gray-900 focus:ring-offset-1"
          />
          <span class="text-[12px] text-gray-500">{t.welcomeDontShow || '下次不显示'}</span>
        </label>
        <button
          class="h-9 px-4 bg-gray-900 text-white text-[13px] font-medium rounded-lg hover:bg-gray-800 transition-colors"
          onclick={handleClose}
        >
          {t.welcomeGotIt || '知道了'}
        </button>
      </div>
    </div>
  </div>
{/if}
