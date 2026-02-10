<script>
  import { onMount, onDestroy } from 'svelte';
  import { StartSSHTerminal, WriteToTerminal, ResizeTerminal, CloseTerminal } from '../../../wailsjs/go/main/App.js';
  import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime.js';

  let { t, caseId, caseName, onClose } = $props();

  let terminalContainer = $state(null);
  let terminal = $state(null);
  let fitAddon = $state(null);
  let sessionId = $state(null);
  let connected = $state(false);
  let connecting = $state(false);
  let error = $state('');

  onMount(async () => {
    // 动态导入 xterm
    try {
      const { Terminal } = await import('xterm');
      const { FitAddon } = await import('xterm-addon-fit');
      
      // 创建终端实例
      terminal = new Terminal({
        cursorBlink: true,
        fontSize: 14,
        fontFamily: 'Menlo, Monaco, "Courier New", monospace',
        theme: {
          background: '#1a1b26',
          foreground: '#a9b1d6',
          cursor: '#c0caf5',
          black: '#32344a',
          red: '#f7768e',
          green: '#9ece6a',
          yellow: '#e0af68',
          blue: '#7aa2f7',
          magenta: '#ad8ee6',
          cyan: '#449dab',
          white: '#787c99',
          brightBlack: '#444b6a',
          brightRed: '#ff7a93',
          brightGreen: '#b9f27c',
          brightYellow: '#ff9e64',
          brightBlue: '#7da6ff',
          brightMagenta: '#bb9af7',
          brightCyan: '#0db9d7',
          brightWhite: '#acb0d0',
        },
        allowProposedApi: true,
      });

      // 添加 fit addon
      fitAddon = new FitAddon();
      terminal.loadAddon(fitAddon);

      // 挂载到 DOM
      terminal.open(terminalContainer);
      fitAddon.fit();

      // 监听用户输入
      terminal.onData((data) => {
        if (sessionId && connected) {
          WriteToTerminal(sessionId, data).catch(err => {
            console.error('写入终端失败:', err);
          });
        }
      });

      // 监听窗口大小变化
      const resizeObserver = new ResizeObserver(() => {
        if (fitAddon && terminal) {
          fitAddon.fit();
          if (sessionId && connected) {
            ResizeTerminal(sessionId, terminal.rows, terminal.cols).catch(err => {
              console.error('调整终端大小失败:', err);
            });
          }
        }
      });
      resizeObserver.observe(terminalContainer);

      // 连接到 SSH
      await connectSSH();

      return () => {
        resizeObserver.disconnect();
      };
    } catch (err) {
      error = `加载终端失败: ${err.message}`;
      console.error('加载终端失败:', err);
    }
  });

  onDestroy(() => {
    cleanup();
  });

  async function connectSSH() {
    if (connecting || connected) return;

    connecting = true;
    error = '';
    terminal?.writeln('\x1b[1;34m正在连接 SSH...\x1b[0m');

    try {
      // 启动 SSH 终端会话
      sessionId = await StartSSHTerminal(caseId, terminal.rows, terminal.cols);
      
      // 监听终端输出
      EventsOn(`terminal-output-${sessionId}`, (data) => {
        terminal?.write(data);
      });

      // 监听终端错误
      EventsOn(`terminal-error-${sessionId}`, (err) => {
        terminal?.writeln(`\r\n\x1b[1;31m错误: ${err}\x1b[0m`);
        error = err;
      });

      // 监听终端关闭
      EventsOn(`terminal-closed-${sessionId}`, () => {
        terminal?.writeln('\r\n\x1b[1;33m连接已关闭\x1b[0m');
        connected = false;
      });

      connected = true;
      terminal?.writeln('\x1b[1;32m已连接\x1b[0m\r\n');
    } catch (err) {
      error = err.message || String(err);
      terminal?.writeln(`\r\n\x1b[1;31m连接失败: ${error}\x1b[0m`);
    } finally {
      connecting = false;
    }
  }

  function cleanup() {
    if (sessionId) {
      EventsOff(`terminal-output-${sessionId}`);
      EventsOff(`terminal-error-${sessionId}`);
      EventsOff(`terminal-closed-${sessionId}`);
      CloseTerminal(sessionId).catch(err => {
        console.error('关闭终端失败:', err);
      });
    }
    terminal?.dispose();
  }

  function handleClose() {
    cleanup();
    onClose();
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      handleClose();
    }
  }

  function handleBackdropClick(e) {
    if (e.target === e.currentTarget) {
      handleClose();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<!-- 导入 xterm 样式 -->
<svelte:head>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css" />
</svelte:head>

<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onclick={handleBackdropClick}>
  <div class="bg-gray-900 rounded-xl shadow-xl w-full max-w-5xl h-[80vh] overflow-hidden flex flex-col" onclick={(e) => e.stopPropagation()}>
    <!-- Header -->
    <div class="px-5 py-4 border-b border-gray-700 flex items-center justify-between flex-shrink-0">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center">
          <svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 7.5l3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0021 18V6a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 6v12a2.25 2.25 0 002.25 2.25z" />
          </svg>
        </div>
        <div>
          <h3 class="text-[15px] font-semibold text-white">{t.webTerminal || 'Web 终端'}</h3>
          <p class="text-[12px] text-gray-400 mt-0.5">{caseName} <span class="text-gray-500">({caseId?.substring(0, 8)})</span></p>
        </div>
      </div>
      <div class="flex items-center gap-2">
        {#if connected}
          <span class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-emerald-500/20 text-emerald-400 text-[11px] font-medium rounded-full">
            <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
            {t.connected || '已连接'}
          </span>
        {:else if connecting}
          <span class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-amber-500/20 text-amber-400 text-[11px] font-medium rounded-full">
            <svg class="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            {t.connecting || '连接中...'}
          </span>
        {:else}
          <span class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-gray-700 text-gray-400 text-[11px] font-medium rounded-full">
            <span class="w-1.5 h-1.5 rounded-full bg-gray-500"></span>
            {t.disconnected || '未连接'}
          </span>
        {/if}
        <button
          class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-gray-800 text-gray-400 hover:text-gray-200 transition-colors"
          onclick={handleClose}
          aria-label="关闭"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Terminal Container -->
    <div class="flex-1 overflow-hidden p-4">
      <div bind:this={terminalContainer} class="w-full h-full rounded-lg overflow-hidden"></div>
    </div>

    <!-- Error Display -->
    {#if error}
      <div class="px-5 py-3 border-t border-gray-700 bg-red-500/10">
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4 text-red-400 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <span class="text-[12px] text-red-400">{error}</span>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  :global(.xterm) {
    height: 100%;
    padding: 8px;
  }
  
  :global(.xterm-viewport) {
    overflow-y: auto;
  }
</style>
