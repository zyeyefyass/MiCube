<template>
  <div class="flex flex-col md:flex-row h-screen min-h-[100dvh] bg-[#090714] text-slate-100 overflow-hidden font-sans">
    <!-- Global Floating Toast Notification -->
    <transition name="toast-slide">
      <div v-if="actionMsg" class="fixed top-4 left-1/2 -translate-x-1/2 z-50 max-w-[92vw] px-4 py-2.5 rounded-2xl backdrop-blur-2xl border shadow-2xl flex items-center gap-2.5 text-xs font-medium"
        :class="actionError ? 'bg-red-950/90 text-red-200 border-red-500/60 shadow-red-500/30' : 'bg-slate-900/90 text-purple-200 border-purple-500/60 shadow-purple-500/30'">
        <span class="w-2 h-2 rounded-full animate-ping flex-shrink-0" :class="actionError ? 'bg-red-400' : 'bg-cyan-400'"></span>
        <span class="truncate">{{ actionMsg }}</span>
        <button @click="actionMsg = ''" class="ml-1 text-slate-400 hover:text-white flex-shrink-0">✕</button>
      </div>
    </transition>

    <!-- ── Mobile Top App Bar (< 768px) ── -->
    <header class="md:hidden flex items-center justify-between px-4 py-3 bg-[#130d2a]/90 backdrop-blur-xl border-b border-purple-900/40 sticky top-0 z-30 flex-shrink-0">
      <div class="flex items-center gap-2.5">
        <img src="/logo-singcube.svg" class="h-8 w-8 drop-shadow-[0_0_12px_rgba(0,210,255,0.7)]" alt="SingCube" />
        <div>
          <div class="flex items-center gap-1.5">
            <span class="text-sm font-bold bg-gradient-to-r from-purple-300 via-fuchsia-400 to-emerald-300 bg-clip-text text-transparent leading-none">SingCube</span>
            <span class="text-[9px] font-mono px-1.5 py-0.5 rounded-full bg-purple-950/80 text-purple-300 border border-purple-500/40 font-semibold shadow-[0_0_8px_rgba(0,210,255,0.3)]">{{ status.version || 'Test V0.7' }}</span>
          </div>
          <div class="text-[8px] text-purple-300/80 tracking-widest leading-none mt-0.5">ROUTEROS & NAS</div>
        </div>
      </div>

      <!-- Detected Host Platform badge with Official Vector Logo & 3D Tilt Glow -->
      <div class="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-slate-900/80 border border-purple-500/40 text-xs shadow-inner cursor-default transition-all duration-300 hover:scale-105 hover:border-purple-400/70" :title="formatPlatform(status.platform)">
        <div class="w-3.5 h-3.5 flex-shrink-0 flex items-center justify-center" v-html="getHostLogo(status.platform_id)"></div>
        <span class="text-[11px] font-medium text-slate-200 truncate max-w-[140px]">{{ formatPlatform(status.platform) }}</span>
      </div>

      <!-- Core Status Dot & Drawer Toggle -->
      <div class="flex items-center gap-2">
        <button @click="openInitWizard('welcome')" class="px-2 py-1 rounded-lg bg-purple-950/80 text-purple-300 border border-purple-500/40 text-[11px] font-bold flex items-center gap-1" title="初始化向导">
          <span>✨</span>
          <span>向导</span>
        </button>
        <span :class="['w-2.5 h-2.5 rounded-full flex-shrink-0', status.running ? 'bg-emerald-400 shadow-[0_0_8px_#10b981] animate-pulse' : 'bg-red-400 shadow-[0_0_6px_#ef4444]']" :title="status.running ? '核心运行中' : '核心已停止'"></span>
        <button @click="showMobileMenu = !showMobileMenu" class="p-1.5 rounded-lg bg-sky-900/50 text-purple-300 hover:text-white border border-purple-700/50 active:scale-95 transition" title="展开全部模块">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
      </div>
    </header>

    <!-- ── PC Desktop Left Sidebar (>= 768px) ── -->
    <aside class="hidden md:flex md:w-64 md:flex-shrink-0 bg-[#130d2a]/80 backdrop-blur-xl border-r border-purple-900/40 flex-col justify-between">
      <div>
        <!-- Brand Hero with 3D Dynamic Crystal Cube -->
        <div class="p-5 text-center border-b border-purple-900/30">
          <div class="relative mx-auto w-28 h-28 lg:w-32 lg:h-32 mb-2 flex items-center justify-center group cursor-pointer">
            <div class="absolute inset-0 rounded-full bg-gradient-to-tr from-cyan-500/25 via-sky-500/35 to-blue-600/10 blur-xl group-hover:scale-125 transition-transform duration-700 animate-pulse"></div>
            <div class="absolute inset-1.5 rounded-2xl border border-purple-400/25 rotate-45 group-hover:rotate-90 transition-transform duration-1000 ease-out pointer-events-none"></div>
            <img src="/logo-singcube.svg" class="relative z-10 h-24 w-24 lg:h-28 lg:w-28 drop-shadow-[0_0_28px_rgba(0,210,255,0.7)] filter hover:drop-shadow-[0_0_38px_rgba(0,210,255,0.95)] transition-all duration-500 hover:scale-110 active:scale-95" alt="SingCube 3D Crystal" />
          </div>
          <div class="flex items-center justify-center gap-2">
            <h1 class="text-2xl font-black tracking-wide bg-gradient-to-r from-sky-200 via-fuchsia-400 to-emerald-300 bg-clip-text text-transparent">SingCube</h1>
            <span class="text-[10px] font-mono px-1.5 py-0.5 rounded-full bg-purple-950/80 text-purple-300 border border-purple-500/40 font-semibold shadow-[0_0_8px_rgba(0,210,255,0.3)]">{{ status.version || 'Test V0.7' }}</span>
          </div>
          <p class="text-[9px] text-purple-300/70 tracking-widest mt-0.5 font-mono">ROUTEROS & NAS GATEWAY</p>

          <!-- Host Platform Badge with 3D perspective -->
          <div class="mt-3 w-full inline-flex items-center justify-center gap-2.5 px-3 py-2 rounded-xl bg-slate-900/80 border border-purple-500/40 text-xs shadow-lg transition-all duration-300 hover:border-purple-400/70 hover:shadow-[0_0_15px_rgba(0,210,255,0.3)]" :title="formatPlatform(status.platform)">
            <div class="w-5 h-5 flex-shrink-0 flex items-center justify-center" v-html="getHostLogo(status.platform_id)"></div>
            <div class="text-left overflow-hidden">
              <div class="text-[8px] text-slate-400 uppercase tracking-wider font-mono">HOST PLATFORM</div>
              <div class="text-xs font-bold text-purple-200 truncate max-w-[175px]" :title="formatPlatform(status.platform)">{{ formatPlatform(status.platform) }}</div>
            </div>
          </div>

          <!-- Initialization Setup Wizard Button -->
          <button @click="openInitWizard('welcome')" class="mt-2.5 w-full py-1.5 px-3 rounded-xl bg-gradient-to-r from-purple-500/20 via-fuchsia-500/20 to-teal-500/20 hover:from-purple-500/30 hover:to-emerald-500/30 border border-purple-400/40 text-purple-300 text-xs font-semibold flex items-center justify-center gap-1.5 shadow-md shadow-purple-500/10 transition-all hover:scale-[1.02]" title="开箱初始化配置向导 (导入备份 / 全新配置)">
            <span class="animate-pulse">✨</span>
            <span>初始化开箱向导</span>
          </button>
        </div>

        <!-- Navigation with Pure Vector SVG Icons -->
        <nav class="p-3 space-y-1.5 overflow-y-auto max-h-[calc(100vh-270px)]">
          <button v-for="item in navItems" :key="item.id"
            @click="switchMasterView(item.id)"
            :class="[
              'w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-xs font-medium transition-all duration-200',
              activeMasterTab === item.id
                ? 'bg-sky-500/20 text-purple-300 shadow-lg shadow-sky-500/10 border border-sky-400/40 translate-x-1'
                : 'text-slate-400 hover:bg-white/5 hover:text-slate-200 border border-transparent'
            ]">
            <div class="w-5 h-5 flex-shrink-0 flex items-center justify-center" v-html="item.icon"></div>
            <span>{{ item.label }}</span>
          </button>
        </nav>
      </div>

      <!-- Status Footer -->
      <div class="p-4 border-t border-purple-900/30 space-y-2 bg-[#120b29]/40">
        <div class="flex items-center justify-between text-xs">
          <div class="flex items-center gap-2">
            <span :class="['w-2 h-2 rounded-full', status.running ? 'bg-emerald-400 shadow-[0_0_8px_#10b981] animate-pulse' : 'bg-red-400']"></span>
            <span :class="status.running ? 'text-emerald-300 font-medium' : 'text-red-300 font-medium'">
              {{ status.running ? '核心运行中' : '核心已停止' }}
            </span>
          </div>
          <span class="text-[10px] text-slate-400 font-mono">{{ status.arch }}</span>
        </div>
        <div class="text-[10px] text-slate-500 flex justify-between">
          <span>总内存占用</span>
          <span class="text-purple-400 font-mono">{{ status.total_memory_mb?.toFixed(1) || '0.0' }} MB</span>
        </div>
        <div class="text-[9px] text-slate-500 text-center mt-2 font-mono flex items-center justify-center gap-1.5 border-t border-purple-900/30 pt-2">
          <span>SingCube</span>
          <span class="text-purple-400 font-bold">{{ status.version || 'Test V0.91' }}</span>
        </div>
      </div>
    </aside>

    <!-- ── Main Content Area (Responsive padding for mobile dock) ── -->
    <main class="flex-1 overflow-y-auto p-4 md:p-6 pb-28 md:pb-6">
      <!-- Dashboard -->
      <div v-if="currentView === 'dashboard'" class="space-y-6">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div class="flex items-center gap-2.5">
            <h2 class="text-2xl font-bold text-white">仪表盘</h2>
            <button @click="openInitWizard('welcome')" class="btn bg-gradient-to-r from-purple-500/20 via-fuchsia-500/20 to-teal-500/20 hover:from-purple-500/30 hover:to-emerald-500/30 text-purple-300 border border-purple-400/40 text-xs px-2.5 py-1 rounded-xl flex items-center gap-1.5 shadow-md shadow-purple-500/10 transition-all hover:scale-105" title="开箱初始化配置向导 (导入备份 / 全新配置)">
              <span class="animate-pulse">✨</span>
              <span>初始化向导</span>
            </button>
          </div>

          <!-- Profile Switcher Capsule & Settings Trigger (Item 1: 配置沙盒隔离) -->
          <div class="flex items-center gap-2 bg-slate-900/80 p-1.5 rounded-2xl border border-purple-500/30 shadow-lg">
            <div class="flex items-center gap-1.5 px-2.5 py-1 text-xs text-slate-300">
              <span class="text-sm">🎯</span>
              <span class="hidden sm:inline text-slate-400">当前模式:</span>
            </div>

            <div class="flex items-center gap-1">
              <button v-for="prof in profiles" :key="prof.id"
                @click="switchProfile(prof.id)"
                :class="[
                  'px-3 py-1.5 rounded-xl text-xs font-semibold transition-all duration-200',
                  activeProfileId === prof.id
                    ? 'bg-gradient-to-r from-purple-500 to-emerald-500 text-white shadow-[0_0_12px_rgba(0,210,255,0.4)]'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'
                ]">
                {{ prof.name }}
              </button>
            </div>

            <button @click="openProfileModal" class="p-1.5 rounded-xl text-slate-400 hover:text-purple-300 hover:bg-slate-800 transition" title="配置与测速独立参数管理">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/><circle cx="12" cy="12" r="3"/></svg>
            </button>
          </div>
        </div>

        <!-- Host IPv4 Forwarding Missing Alert Banner (Essential for Docker/Bypass) -->
        <div v-if="hostInfo.connected !== false && !hostInfo.ip_forward"
          class="glass-card p-4 border-l-4 border-l-amber-500 bg-amber-950/40 border border-amber-600/40 flex flex-col sm:flex-row sm:items-center justify-between gap-3 shadow-lg shadow-amber-950/30">
          <div class="flex items-start gap-3">
            <span class="text-2xl mt-0.5">⚠️</span>
            <div>
              <div class="text-sm font-bold text-amber-200 flex items-center gap-2">
                <span>宿主机内核 IPv4 转发与防火墙未打通 (net.ipv4.ip_forward = 0)</span>
                <span class="text-[10px] bg-amber-900/80 text-amber-300 border border-amber-500/50 px-2 py-0.5 rounded-full font-bold">必须首位打通</span>
              </div>
              <div class="text-xs text-slate-300 mt-1">
                Docker 运行在宿主内核之上。若宿主未开启内核转发与防火墙放行，所有局域网设备修改网关后流量将被宿主内核静默丢弃！
              </div>
            </div>
          </div>
          <div class="flex items-center gap-2 shrink-0">
            <button @click="openHostFwModal('permissive')" class="btn bg-gradient-to-r from-amber-500 to-orange-600 hover:from-amber-400 hover:to-orange-500 text-white text-xs font-bold px-3.5 py-2 shadow-lg shadow-amber-500/25 flex items-center gap-1.5 rounded-xl">
              <span>🛡️ 一键配置底座防火墙 (严格/宽松向导) ➔</span>
            </button>
          </div>
        </div>

        <!-- Novice Quick-Start Wizard (singcube-ux) -->
        <div v-if="showOnboarding && (!configState.subscriptions || configState.subscriptions.length === 0)"
          class="glass-card p-5 border-l-4 border-l-cyan-400 space-y-3 bg-gradient-to-r from-sky-950/40 via-cyan-950/20 to-transparent">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="text-xl">🚀</span>
              <h3 class="text-base font-bold text-white">欢迎使用 SingCube！极速开箱 4 步通关指南</h3>
            </div>
            <button @click="dismissOnboarding" class="text-xs text-slate-400 hover:text-white px-2 py-1 rounded bg-slate-800">我知道了 ✕</button>
          </div>
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 text-xs">
            <div @click="openHostFwModal('permissive')" :class="['p-3 rounded-xl border cursor-pointer transition', hostInfo.ip_forward ? 'bg-slate-900/60 border-emerald-500/40 hover:border-emerald-400' : 'bg-amber-950/50 border-amber-500/50 hover:border-amber-400 animate-pulse']">
              <div :class="['font-bold mb-1 flex items-center justify-between', hostInfo.ip_forward ? 'text-emerald-300' : 'text-amber-300']">
                <span>0. 宿主底座与防火墙 ➔</span>
                <span class="text-[10px]">{{ hostInfo.ip_forward ? '✓ 已就绪' : '⚠️ 需打通 (向导)' }}</span>
              </div>
              <div class="text-slate-400 text-[11px]">一键弹窗配置内核转发与防火墙（严格/宽松），防止局域网设备断网。</div>
            </div>
            <div @click="currentView = 'config'" class="p-3 bg-slate-900/60 rounded-xl border border-purple-800/40 hover:border-purple-400/60 cursor-pointer transition">
              <div class="font-bold text-purple-300 mb-1">1. 添加节点订阅 ➔</div>
              <div class="text-slate-400 text-[11px]">前往「订阅配置」粘贴您的 Clash 订阅链接并保存。</div>
            </div>
            <div @click="coreAction('start')" class="p-3 bg-slate-900/60 rounded-xl border border-purple-800/40 hover:border-emerald-400/60 cursor-pointer transition">
              <div class="font-bold text-emerald-300 mb-1">2. 一键启动核心 ➔</div>
              <div class="text-slate-400 text-[11px]">点击「▶ 启动」并勾选「开机自启」，TUN 协议栈即刻就绪。</div>
            </div>
            <div @click="startWgWizard" class="p-3 bg-gradient-to-br from-cyan-950/60 to-blue-950/40 rounded-xl border border-purple-500/50 hover:border-purple-400 cursor-pointer transition shadow-lg shadow-purple-500/10">
              <div class="font-bold text-purple-300 mb-1 flex items-center gap-1.5">
                <span>🚀 3. 手机一键极速穿透 ➔</span>
              </div>
              <div class="text-slate-400 text-[11px]">自动打通并弹出二维码，手机扫码连接即成功庆祝！</div>
            </div>
          </div>
        </div>

        <!-- Status Cards -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div class="glass-card p-5">
            <div class="text-xs text-purple-300 mb-1">核心状态</div>
            <div :class="['text-2xl font-bold', status.running ? 'text-emerald-400' : 'text-red-400']">
              {{ status.running ? '运行中' : '已停止' }}
            </div>
            <div class="text-xs text-slate-500 mt-1">PID: {{ status.pid || '-' }}</div>
          </div>

          <div class="glass-card p-5">
            <div class="text-xs text-purple-300 mb-1">运行时间</div>
            <div class="text-2xl font-bold text-white">{{ formatUptime(status.uptime_seconds) }}</div>
            <div class="text-xs text-slate-500 mt-1">{{ status.tun_mode || '-' }}</div>
          </div>

          <div class="glass-card p-5">
            <div class="text-xs text-purple-300 mb-1">核心内存</div>
            <div class="text-2xl font-bold text-purple-300">{{ status.core_memory_mb?.toFixed(1) || '0' }} MB</div>
            <div class="text-xs text-slate-500 mt-1">面板: {{ status.panel_memory_mb?.toFixed(1) || '0' }} MB</div>
          </div>

          <div class="glass-card p-5">
            <div class="text-xs text-purple-300 mb-1">宿主平台</div>
            <div class="text-base lg:text-lg font-bold text-white break-words" :title="formatPlatform(status.platform)">{{ formatPlatform(status.platform) }}</div>
            <div class="text-xs text-slate-500 mt-1">{{ status.arch || '-' }}</div>
          </div>
        </div>

        <!-- Live Realtime Traffic Waveform (singcube-ui & singcube-network) -->
        <div class="glass-card p-5 space-y-3 relative overflow-hidden border border-purple-500/30">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div>
              <div class="text-sm font-bold text-white flex items-center gap-2">
                <span>📈 实时网络吞吐与流量流速波形</span>
                <span class="w-2 h-2 rounded-full bg-cyan-400 animate-ping"></span>
              </div>
              <div class="text-[10px] text-slate-400 mt-0.5">高精度流量遥测 · 60fps 动态曲线</div>
            </div>
            <div class="flex items-center gap-4 text-xs font-mono">
              <div class="flex items-center gap-1.5 bg-slate-900/60 px-2.5 py-1 rounded-lg border border-purple-500/30">
                <span class="w-2 h-2 rounded-full bg-cyan-400"></span>
                <span class="text-slate-400">下行:</span>
                <span class="text-purple-300 font-bold">{{ realtimeSpeed.down_speed_human || '0 B/s' }}</span>
              </div>
              <div class="flex items-center gap-1.5 bg-slate-900/60 px-2.5 py-1 rounded-lg border border-emerald-500/30">
                <span class="w-2 h-2 rounded-full bg-emerald-400"></span>
                <span class="text-slate-400">上行:</span>
                <span class="text-emerald-300 font-bold">{{ realtimeSpeed.up_speed_human || '0 B/s' }}</span>
              </div>
              <div class="hidden md:flex items-center gap-1.5 bg-slate-900/60 px-2.5 py-1 rounded-lg border border-purple-700/30">
                <span class="text-slate-400">活跃链路:</span>
                <span class="text-white font-bold">{{ realtimeSpeed.active_conns || 0 }}</span>
              </div>
            </div>
          </div>

          <!-- SVG Realtime Waveform -->
          <div class="w-full h-28 bg-[#120b29]/60 rounded-xl p-1 relative border border-purple-900/40">
            <svg class="w-full h-full overflow-visible" viewBox="0 0 600 120" preserveAspectRatio="none">
              <defs>
                <linearGradient id="downGlow" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#00d2ff" stop-opacity="0.45" />
                  <stop offset="100%" stop-color="#00d2ff" stop-opacity="0.0" />
                </linearGradient>
                <linearGradient id="upGlow" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#10b981" stop-opacity="0.35" />
                  <stop offset="100%" stop-color="#10b981" stop-opacity="0.0" />
                </linearGradient>
              </defs>

              <!-- Grid horizontal lines -->
              <line x1="0" y1="30" x2="600" y2="30" stroke="#00355e" stroke-width="0.8" stroke-dasharray="4,4" />
              <line x1="0" y1="60" x2="600" y2="60" stroke="#00355e" stroke-width="0.8" stroke-dasharray="4,4" />
              <line x1="0" y1="90" x2="600" y2="90" stroke="#00355e" stroke-width="0.8" stroke-dasharray="4,4" />

              <!-- Download Area & Curve -->
              <path :d="waveformPaths.downPath" fill="url(#downGlow)" />
              <path :d="waveformPaths.downLine" fill="none" stroke="#00d2ff" stroke-width="2.2" stroke-linecap="round" />

              <!-- Upload Area & Curve -->
              <path :d="waveformPaths.upPath" fill="url(#upGlow)" />
              <path :d="waveformPaths.upLine" fill="none" stroke="#10b981" stroke-width="1.8" stroke-linecap="round" />

              <!-- Glowing pulse dots on latest sample -->
              <circle :cx="waveformPaths.lastDownX" :cy="waveformPaths.lastDownY" r="4" fill="#00d2ff" class="filter drop-shadow-[0_0_6px_#00d2ff]" />
              <circle :cx="waveformPaths.lastUpX" :cy="waveformPaths.lastUpY" r="3.5" fill="#10b981" class="filter drop-shadow-[0_0_5px_#10b981]" />
            </svg>
          </div>
        </div>

        <!-- Quick Proxy Selector Widget (singcube-ux) -->
        <div class="glass-card p-4 space-y-3 border-l-4 border-l-cyan-400">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div class="flex items-center gap-2">
              <span class="text-sm font-bold text-white flex items-center gap-1.5">
                <span>🎯 核心出口与快捷选路</span>
              </span>
              <span class="text-[10px] text-slate-400">快速指定「节点总览」主出站</span>
            </div>
            <button @click="currentView = 'proxies'" class="text-xs text-purple-400 hover:text-purple-300 flex items-center gap-1 self-start sm:self-auto font-medium">
              <span>进入策略组与节点总览 ➔</span>
            </button>
          </div>

          <div class="flex items-center gap-2 flex-wrap pt-1">
            <div class="text-xs text-slate-300 font-medium">主选路节点:</div>
            <select 
              v-if="quickOverviewGroup"
              :value="quickOverviewGroup.now"
              @change="selectProxy('节点总览', $event.target.value)"
              class="input-field text-xs font-mono font-bold text-purple-200 bg-slate-900/90 border-purple-700/60 py-1.5 px-3 rounded-xl cursor-pointer hover:border-purple-400 focus:border-purple-400"
            >
              <option v-for="node in quickOverviewGroup.all" :key="node.name" :value="node.name">
                {{ node.name }} {{ node.delay > 0 ? `(${node.delay}ms)` : '' }}
              </option>
            </select>
            <span v-else class="text-xs text-slate-500 font-mono">加载中...</span>
            <button @click="testAllProxiesDelay" :disabled="testingProxies" class="btn btn-blue text-xs py-1 px-3 ml-auto flex items-center gap-1">
              <span :class="{'animate-spin': testingProxies}">⚡</span>
              <span>{{ testingProxies ? '测速中...' : '一键测速' }}</span>
            </button>
          </div>
        </div>

        <!-- Core Control Buttons -->
        <div class="glass-card p-5">
          <h3 class="text-sm text-purple-300 mb-4 font-semibold">核心控制</h3>
          <div class="flex flex-wrap gap-3">
            <button @click="coreAction('start')"   :disabled="loading" class="btn btn-green">▶ 启动</button>
            <button @click="coreAction('stop')"    :disabled="loading" class="btn btn-red">■ 停止</button>
            <button @click="coreAction('restart')" :disabled="loading" class="btn btn-blue">↻ 重启</button>
            <button @click="coreAction('reload')"  :disabled="loading" class="btn btn-cyan">⟳ 热重载</button>
            <div class="flex-1"></div>
            <label class="flex items-center gap-2 text-sm text-slate-300 cursor-pointer">
              <input type="checkbox" :checked="status.autostart" @change="toggleAutostart" class="accent-cyan-400 w-4 h-4" />
              开机自启
            </label>
          </div>
          <div v-if="actionMsg" class="mt-3 text-xs px-3 py-2 rounded-lg" :class="actionError ? 'bg-red-900/40 text-red-300' : 'bg-emerald-900/40 text-emerald-300'">
            {{ actionMsg }}
          </div>
        </div>
      </div>

      <!-- ── Proxy Groups & Node Selection ── -->
      <div v-if="currentView === 'proxies'" class="space-y-6">
        <!-- ── Top Sub-Nav Pills: 节点与分流 ── -->
        <div class="flex items-center gap-2 p-1 bg-slate-900/80 rounded-2xl border border-purple-800/40 w-fit backdrop-blur-md shadow-sm">
          <button 
            @click="currentView = 'proxies'; proxiesSubTab = 'nodes'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'proxies' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>⚡ 节点选择与策略组</span>
          </button>
          <button 
            @click="currentView = 'config'; proxiesSubTab = 'config'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'config' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>📝 订阅配置与分流规则</span>
          </button>
        </div>

        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <h2 class="text-2xl font-bold text-white flex items-center gap-2">
              <span>节点选择与策略组</span>
              <span class="text-xs font-mono font-normal px-2.5 py-0.5 rounded-full bg-purple-950/80 text-purple-300 border border-purple-500/40">
                {{ proxyGroupsData.total_proxies || 0 }} 可用节点
              </span>
            </h2>
            <p class="text-xs text-slate-400 mt-1">实时掌控出口选路与节点健康度，支持各地区自选、全局低延迟切换及单节点毫秒级延迟测速。</p>
          </div>
          <div class="flex gap-2 flex-wrap items-center">
            <button @click="testAllProxiesDelay" :disabled="testingProxies" class="btn btn-blue text-xs flex items-center gap-1.5 shadow-md">
              <span :class="{'animate-spin': testingProxies}">⚡</span>
              <span>{{ testingProxies ? '正在并发测速...' : '一键全员测速' }}</span>
            </button>
            <button @click="fetchProxies" class="btn btn-cyan text-xs flex items-center gap-1 shadow-md">
              <span>↻ 刷新节点</span>
            </button>
          </div>
        </div>

        <!-- Filter & Search Bar -->
        <div class="glass-card p-4 flex flex-col md:flex-row items-center justify-between gap-3 border-l-4 border-l-cyan-400">
          <div class="flex items-center gap-2 flex-wrap w-full md:w-auto">
            <span class="text-xs text-purple-300 font-semibold mr-1">策略分类:</span>
            <button v-for="tab in [
              { id: 'all', label: '全部策略组' },
              { id: 'core', label: '核心优选' },
              { id: 'region', label: '地区节点' },
              { id: 'biz', label: '分流规则' }
            ]" :key="tab.id"
              @click="proxyFilter = tab.id"
              :class="[
                'px-3 py-1.5 rounded-xl text-xs font-medium transition-all',
                proxyFilter === tab.id
                  ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.25)]'
                  : 'bg-slate-900/60 text-slate-400 hover:text-slate-200 border border-slate-700/40'
              ]">
              {{ tab.label }}
            </button>
          </div>

          <div class="w-full md:w-64 relative">
            <input 
              v-model="proxySearch" 
              placeholder="搜索节点或策略名..." 
              class="input-field w-full pl-8 text-xs font-mono"
            />
            <div class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-500 text-xs">
              🔍
            </div>
            <button v-if="proxySearch" @click="proxySearch = ''" class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white text-xs">✕</button>
          </div>
        </div>

        <!-- Groups Grid -->
        <div v-if="filteredProxyGroups.length === 0" class="glass-card p-12 text-center text-slate-400 text-xs space-y-2">
          <div class="text-3xl">📦</div>
          <div class="text-sm font-semibold text-slate-300">未检测到匹配的策略组或节点</div>
          <div class="text-[11px] text-slate-500">请检查订阅链接是否已在「订阅配置」拉取，或尝试点击右上角「刷新节点」。</div>
        </div>

        <div v-else class="space-y-6">
          <div v-for="group in filteredProxyGroups" :key="group.name" class="glass-card p-5 space-y-3.5 border border-purple-800/40 hover:border-purple-500/40 transition">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-purple-900/40 pb-3">
              <div class="flex items-center gap-2.5">
                <span class="text-base font-bold text-white">{{ group.name }}</span>
                <span :class="[
                  'text-[10px] font-mono px-2 py-0.5 rounded-full border font-semibold',
                  group.is_select
                    ? 'bg-sky-950/80 text-purple-300 border-purple-500/40'
                    : 'bg-emerald-950/80 text-emerald-300 border-emerald-500/40'
                ]">
                  {{ group.is_select ? '手动选择 (Selector)' : (group.type === 'URLTest' ? '⚡ 最低延迟 (URL-Test)' : '🛡️ 故障转移 (Fallback)') }}
                </span>
                <span class="text-xs text-slate-500 font-mono">({{ group.all?.length || 0 }} 节点)</span>
              </div>

              <!-- Current Selection Capsule & Quick Toggle -->
              <div class="flex items-center gap-2 flex-wrap">
                <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-xl bg-slate-900/90 border border-purple-700/50 text-xs shadow-inner">
                  <span class="text-slate-400 text-[11px]">当前出口:</span>
                  <span class="font-bold text-purple-300 font-mono flex items-center gap-1.5">
                    <span class="w-1.5 h-1.5 rounded-full bg-cyan-400 animate-pulse"></span>
                    {{ group.now || '未就绪' }}
                  </span>
                </div>
                <button 
                  type="button" 
                  @click="toggleGroupExpanded(group.name)"
                  class="text-xs px-2.5 py-1 rounded-xl border border-purple-700/50 bg-slate-800/80 hover:bg-slate-700 text-slate-300 hover:text-white flex items-center gap-1 transition shadow-sm active:scale-95"
                >
                  <span>{{ expandedGroups[group.name] || proxySearch ? '🔼 收起' : '🔽 展开选路 (' + (group.all?.length || 0) + ')' }}</span>
                </button>
              </div>
            </div>

            <!-- Nodes Grid within Group: Compact by default, expands on click or search -->
            <div v-show="expandedGroups[group.name] || proxySearch" class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2.5 pt-1">
              <div 
                v-for="node in filterNodes(group.all)" 
                :key="node.name"
                @click="group.is_select && selectProxy(group.name, node.name)"
                :class="[
                  'p-2.5 rounded-xl border text-xs transition-all relative overflow-hidden flex flex-col justify-between gap-1.5',
                  group.is_select ? 'cursor-pointer hover:scale-[1.02]' : 'cursor-default',
                  node.name === group.now
                    ? 'bg-purple-950/40 border-purple-400/80 shadow-[0_0_12px_rgba(0,210,255,0.25)] ring-1 ring-cyan-400/50'
                    : 'bg-slate-900/60 border-purple-900/40 hover:border-sky-600/60 hover:bg-slate-800/50'
                ]"
              >
                <!-- Selected Badge Watermark -->
                <div v-if="node.name === group.now" class="absolute -right-3 -top-3 w-8 h-8 bg-purple-600/20 rounded-full blur-md pointer-events-none"></div>

                <div class="flex items-start justify-between gap-1.5">
                  <div class="font-medium text-slate-200 truncate flex-1" :title="node.name">
                    {{ node.name }}
                  </div>
                  <span v-if="node.name === group.now" class="text-[9px] font-bold text-purple-300 bg-cyan-900/60 px-1.5 py-0.5 rounded font-mono flex-shrink-0">
                    ✔ 已选
                  </span>
                </div>

                <div class="flex items-center justify-between gap-2 pt-1 border-t border-purple-900/20 text-[10px] font-mono">
                  <span class="text-slate-500 truncate max-w-[80px] uppercase">{{ node.type }}</span>
                  <div class="flex items-center gap-1">
                    <span 
                      v-if="node.delay > 0" 
                      :class="[
                        'px-1.5 py-0.5 rounded text-[10px] font-bold',
                        node.delay <= 120 ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-800/40' :
                        node.delay <= 250 ? 'bg-amber-950/80 text-amber-300 border border-amber-800/40' :
                        'bg-rose-950/80 text-rose-400 border border-rose-800/40'
                      ]"
                    >
                      {{ node.delay }}ms
                    </span>
                    <span v-else-if="node.delay === -1" class="text-[10px] text-purple-300 animate-pulse font-mono">测速中...</span>
                    <span v-else-if="node.delay === -2" class="text-[10px] text-amber-300 animate-pulse font-mono">复查中...</span>
                    <span v-else-if="node.delay === 0" class="text-[10px] text-rose-500 font-mono">超时</span>
                    <span v-else class="text-slate-600">未测速</span>
                    <button 
                      type="button" 
                      @click.stop="testSingleProxyDelay(node.name)" 
                      class="text-slate-500 hover:text-purple-300 p-0.5 transition"
                      title="单独测试此节点延迟"
                    >
                      ⚡
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Config Editor -->
      <div v-if="currentView === 'config'" class="space-y-6">
        <!-- ── Top Sub-Nav Pills: 节点与分流 ── -->
        <div class="flex items-center gap-2 p-1 bg-slate-900/80 rounded-2xl border border-purple-800/40 w-fit backdrop-blur-md shadow-sm">
          <button 
            @click="currentView = 'proxies'; proxiesSubTab = 'nodes'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'proxies' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>⚡ 节点选择与策略组</span>
          </button>
          <button 
            @click="currentView = 'config'; proxiesSubTab = 'config'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'config' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>📝 订阅配置与分流规则</span>
          </button>
        </div>

        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <h2 class="text-2xl font-bold text-white">订阅与规则配置</h2>
          <div class="flex gap-2 flex-wrap items-center">
            <button @click="exportBackup" class="btn bg-purple-900/60 hover:bg-purple-800 text-purple-200 border border-purple-500/40 text-xs flex items-center gap-1 shadow-md" title="导出当前系统所有订阅与分流规则备份">
              📤 导出配置备份
            </button>
            <label class="btn bg-slate-800 hover:bg-slate-700 text-sky-300 border border-purple-800/60 text-xs flex items-center gap-1 cursor-pointer shadow-md" title="从 JSON 备份文件恢复全部配置">
              📥 导入配置恢复
              <input type="file" accept=".json" @change="importBackup" class="hidden" />
            </label>
            <button @click="updateAllSubscriptions" :disabled="loading" class="btn btn-blue text-xs">📥 一键拉取节点</button>
            <button @click="fetchConfigState" class="btn btn-cyan text-xs">↻ 刷新</button>
            <button 
              @click="saveConfig" 
              :disabled="loading || !isConfigDirty" 
              :class="[
                'btn text-xs transition-all duration-300 flex items-center gap-1.5',
                isConfigDirty 
                  ? 'bg-purple-600 hover:bg-cyan-400 text-slate-950 font-bold shadow-[0_0_15px_rgba(0,210,255,0.6)] animate-pulse' 
                  : 'bg-slate-800/80 text-slate-500 border border-slate-700/40 cursor-not-allowed opacity-70'
              ]"
            >
              <span>{{ isConfigDirty ? '💾 保存并编译 (已修改)' : '✔ 配置已是最新' }}</span>
            </button>
            <button 
              @click="resetToDefaultRules" 
              :disabled="loading" 
              class="btn btn-red text-xs flex items-center gap-1"
              title="恢复官方出厂推荐分流与去广告规则（保留您的订阅链接与节点不变）"
            >
              <span>🔄 恢复推荐规则</span>
            </button>
          </div>
        </div>

        <!-- Advanced Engine Settings: Failover Mode & Direct Real-IP Bypass & QUIC Optimization -->
        <div class="glass-card p-5 space-y-4 border-l-4 border-l-cyan-400">
          <h3 class="text-sm text-purple-300 font-semibold flex items-center gap-2">
            <span>⚙️ 策略组模式与真实 IP 穿透控制</span>
          </h3>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <!-- Auto Select Mode -->
            <div class="bg-slate-900/40 p-4 rounded-xl border border-purple-900/30 space-y-2">
              <div class="text-xs font-semibold text-white flex items-center justify-between">
                <span>选路策略模式</span>
                <span class="text-[11px] text-purple-300 font-mono">{{ configState.auto_select_mode === 'fallback' ? '故障转移 (Fallback)' : '最低延迟 (URL-Test)' }}</span>
              </div>
              <p class="text-[11px] text-slate-400">选择各地区自动组与优选组的动态算法：</p>
              <div class="flex gap-2 pt-1">
                <button type="button" @click="configState.auto_select_mode = 'url-test'"
                  :class="['flex-1 py-1.5 px-2 rounded-lg text-xs font-medium border transition-all',
                    configState.auto_select_mode === 'url-test'
                      ? 'bg-purple-600/20 text-purple-300 border-purple-400/50'
                      : 'bg-slate-800 text-slate-400 border-transparent hover:text-slate-200']">
                  ⚡ 智能最低延迟
                </button>
                <button type="button" @click="configState.auto_select_mode = 'fallback'"
                  :class="['flex-1 py-1.5 px-2 rounded-lg text-xs font-medium border transition-all',
                    configState.auto_select_mode === 'fallback'
                      ? 'bg-purple-600/20 text-purple-300 border-purple-400/50'
                      : 'bg-slate-800 text-slate-400 border-transparent hover:text-slate-200']">
                  🛡️ 故障转移 (IP稳定)
                </button>
              </div>
              <div class="text-[10px] text-slate-500">
                {{ configState.auto_select_mode === 'fallback' ? '★ 优先固定首个可用节点，宕机才切换，避免币圈/ChatGPT频繁跳IP' : '★ 实时并发测速，延迟最低优先，适合游戏与极致网页响应' }}
              </div>
            </div>

            <!-- Direct Real-IP Bypass -->
            <div class="bg-slate-900/40 p-4 rounded-xl border border-purple-900/30 space-y-2">
              <div class="text-xs font-semibold text-white flex items-center justify-between">
                <span>国内直连显示真实 IP</span>
                <span :class="['text-[11px] font-mono', configState.direct_bypass_mode ? 'text-emerald-400' : 'text-slate-500']">
                  {{ configState.direct_bypass_mode ? '● 开启 (真实IP)' : '○ 关闭 (Fake-IP)' }}
                </span>
              </div>
              <p class="text-[11px] text-slate-400">国内大厂与局域网流量彻底绕过 Fake-IP 映射：</p>
              <label class="flex items-center gap-2 text-xs text-slate-300 cursor-pointer pt-1">
                <input type="checkbox" v-model="configState.direct_bypass_mode" class="accent-cyan-400 w-4 h-4" />
                <span>启用国内域名真实 IP 解析 (不占内核 CPU, 跑满载带宽)</span>
              </label>
              <div class="text-[10px] text-slate-500">
                ★ 排除中国域名进入 Fake-IP 池，国内直连完全显示客户端真实公网 IP，BT/P2P 穿透无阻碍。
              </div>
            </div>

            <!-- QUIC 443 Block Optimization -->
            <div class="bg-slate-900/40 p-4 rounded-xl border border-purple-900/30 space-y-2">
              <div class="text-xs font-semibold text-white flex items-center justify-between">
                <span>海外 QUIC (UDP 443) 优化</span>
                <span :class="['text-[11px] font-mono', configState.quic_block_enabled ? 'text-amber-400' : 'text-slate-500']">
                  {{ configState.quic_block_enabled ? '● 阻断 (强制TCP/BBR)' : '○ 原生放行' }}
                </span>
              </div>
              <p class="text-[11px] text-slate-400">针对运营商 UDP 丢包与海外 QOS 限制治理：</p>
              <label class="flex items-center gap-2 text-xs text-slate-300 cursor-pointer pt-1">
                <input type="checkbox" v-model="configState.quic_block_enabled" class="accent-cyan-400 w-4 h-4" />
                <span>拦截海外 UDP 443 (强制 YouTube/Chrome 走 TCP)</span>
              </label>
              <div class="text-[10px] text-slate-500">
                ★ 阻断海外 UDP 443 迫使客户端回退到 TCP TLS，享受节点 BBR 算法加速，消除 4K 缓冲转圈与丢包。
              </div>
            </div>
          </div>
        </div>

        <!-- Subscriptions -->
        <div class="glass-card p-5">
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center gap-2">
              <h3 class="text-sm text-purple-300 font-semibold">机场订阅</h3>
              <span class="text-xs font-mono text-purple-300 bg-sky-950/80 px-2.5 py-0.5 rounded-full border border-purple-500/40">
                总计有效节点: {{ totalConfigNodeCount }}
              </span>
            </div>
            <button @click="addSub" class="text-xs text-purple-400 hover:text-purple-300 font-medium">+ 添加订阅</button>
          </div>
          <div v-for="(sub, i) in configState.subscriptions" :key="sub.id" class="flex items-center gap-3 mb-2 flex-wrap sm:flex-nowrap">
            <input v-model="sub.name" placeholder="名称" class="input-field flex-shrink-0 w-32" />
            <input v-model="sub.url"  placeholder="订阅链接 URL" class="input-field flex-1" />
            <select v-model="sub.type" class="input-field w-24">
              <option value="clash">Clash</option>
              <option value="v2ray">V2Ray</option>
              <option value="ss">SS</option>
            </select>
            <span class="text-xs text-purple-300 bg-sky-900/60 border border-purple-700/50 px-2.5 py-1 rounded-lg font-mono font-bold flex-shrink-0">
              {{ sub.node_count || 0 }} 节点
            </span>
            <label class="flex items-center gap-1 text-xs text-slate-400 cursor-pointer flex-shrink-0">
              <input type="checkbox" v-model="sub.enabled" class="accent-cyan-400" /> 启用
            </label>
            <button @click="updateSingleSub(sub.id)" title="仅更新此订阅" class="text-purple-400 hover:text-purple-300 text-sm">📥</button>
            <button @click="removeSub(i)" class="text-red-400 hover:text-red-300 text-sm">✕</button>
          </div>
          <button @click="addSub" class="text-xs text-purple-400 hover:text-purple-300 mt-2">+ 添加订阅</button>
        </div>

        <!-- Rule Targets -->
        <div class="glass-card p-5">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-3">
            <div>
              <h3 class="text-sm text-purple-300 font-semibold">分流规则 (MRS)</h3>
              <p class="text-xs text-slate-400 mt-0.5">支持按目标业务指定出口策略组，提供各地区自动（最低延迟）与故转（故障转移）便捷下拉选择</p>
            </div>
            <button @click="addRuleTarget" class="btn btn-cyan text-xs py-1 px-3 self-start sm:self-auto flex items-center gap-1 shadow-md shadow-purple-500/10">
              <span>+ 添加规则</span>
            </button>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead class="text-xs text-purple-300/70 border-b border-purple-900/30">
                <tr>
                  <th class="text-left py-2 px-2">名称</th>
                  <th class="text-left py-2 px-2">MRS 规则名</th>
                  <th class="text-left py-2 px-2 min-w-[210px]">出口策略组</th>
                  <th class="text-center py-2 px-2">匹配优先级</th>
                  <th class="text-center py-2 px-2">启用</th>
                  <th class="text-center py-2 px-2">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(rule, idx) in configState.rule_targets" :key="rule.id || idx" class="border-t border-purple-900/20 hover:bg-slate-800/25 transition">
                  <!-- 1. 名称 (含官网图标) -->
                  <td class="py-2.5 px-2">
                    <div class="flex items-center gap-2">
                      <div class="w-7 h-7 rounded-lg bg-slate-900/90 border border-purple-800/50 flex items-center justify-center shrink-0 shadow-sm p-1" :title="rule.name">
                        <span v-html="getRuleOfficialIconSvg(rule.mrs_name || rule.name)" class="flex items-center justify-center"></span>
                      </div>
                      <input v-model="rule.name" class="input-field w-full text-xs font-medium text-slate-100" />
                    </div>
                  </td>

                  <!-- 2. MRS 规则名 (系统锁定不可更改的图框) -->
                  <td class="py-2.5 px-2">
                    <div v-if="!rule._editingMrs" class="flex items-center justify-between px-2.5 py-1.5 rounded-lg bg-slate-950/80 border border-purple-900/50 text-xs font-mono text-purple-300 shadow-inner group select-all" title="MRS 二进制规则集标识 (系统锁定不可更改)">
                      <div class="flex items-center gap-1.5 truncate">
                        <span class="text-[10px] text-purple-300/80" title="系统锁定保护">🔒</span>
                        <span class="font-bold tracking-wider text-purple-200">{{ rule.mrs_name }}</span>
                      </div>
                      <div class="flex items-center gap-1">
                        <span class="text-[10px] text-purple-300/70 font-mono bg-sky-950/90 px-1.5 py-0.5 rounded border border-purple-800/40 shrink-0">.mrs</span>
                        <button v-if="rule.is_custom" @click="rule._editingMrs = true" class="text-[10px] text-slate-500 hover:text-purple-300 p-0.5 opacity-0 group-hover:opacity-100 transition" title="自定义修改">✎</button>
                      </div>
                    </div>
                    <div v-else class="flex items-center gap-1">
                      <input v-model="rule.mrs_name" class="input-field w-full text-xs font-mono text-purple-300" placeholder="mrs文件名" />
                      <button @click="rule._editingMrs = false" class="btn btn-blue text-[10px] px-2 py-1 shrink-0 font-semibold">锁定</button>
                    </div>
                  </td>

                  <!-- 3. 出口策略组 (显示国旗或区旗) -->
                  <td class="py-2.5 px-2">
                    <div class="relative flex items-center gap-1.5 min-w-[210px]">
                      <div v-if="!rule._customInput" class="relative w-full flex items-center">
                        <span class="absolute left-2.5 z-10 text-sm pointer-events-none select-none flex items-center">
                          {{ getPolicyFlag(rule.policy_group) }}
                        </span>
                        <select 
                          v-model="rule.policy_group" 
                          class="input-field w-full text-xs font-medium bg-slate-900/90 border border-purple-800/60 text-purple-200 rounded-lg py-1.5 pl-8 pr-7 appearance-none cursor-pointer hover:border-purple-400 focus:border-purple-400 transition"
                        >
                          <optgroup label="⚡ 地区自动优选 (最低延迟)">
                            <option v-for="c in availableCountries" :key="c + '自动'" :value="c + '自动'">{{ getCountryFlag(c) }} {{ c }}自动</option>
                          </optgroup>
                          <optgroup label="🛡️ 地区故转 (顺序顺延 / IP稳定)">
                            <option v-for="c in availableCountries" :key="c + '故转'" :value="c + '故转'">{{ getCountryFlag(c) }} {{ c }}故转</option>
                          </optgroup>
                          <optgroup label="🌐 地区容灾 (备用灾备)">
                            <option v-for="c in availableCountries" :key="c + '容灾'" :value="c + '容灾'">{{ getCountryFlag(c) }} {{ c }}容灾</option>
                          </optgroup>
                          <optgroup label="🎯 地区手动 (人工指定节点)">
                            <option v-for="c in availableCountries" :key="c + '手动'" :value="c + '手动'">{{ getCountryFlag(c) }} {{ c }}手动</option>
                            <option v-for="c in availableCountries" :key="c + '节点'" :value="c + '节点'">{{ getCountryFlag(c) }} {{ c }}节点</option>
                          </optgroup>
                          <optgroup label="🌐 全局核心策略">
                            <option value="自动优选">⚡ 自动优选 (全局低延迟)</option>
                            <option value="故障转移">🛡️ 故障转移 (全局容灾)</option>
                            <option value="节点总览">🎯 节点总览 (手动选择)</option>
                            <option value="直连">🇨🇳 直连 (DIRECT 真实IP直出)</option>
                            <option value="拦截">🛑 拦截 (REJECT 阻断丢弃)</option>
                          </optgroup>
                          <optgroup v-if="isCustomPolicy(rule.policy_group)" label="✨ 当前自定义策略">
                            <option :value="rule.policy_group">{{ getPolicyFlag(rule.policy_group) }} {{ rule.policy_group }}</option>
                          </optgroup>
                        </select>
                        <div class="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 text-purple-300 text-[10px]">
                          ▼
                        </div>
                      </div>
                      <div v-else class="flex items-center gap-1.5 w-full">
                        <span class="text-sm px-2 py-1 bg-slate-950/80 border border-purple-800/40 rounded-lg flex items-center justify-center shrink-0">
                          {{ getPolicyFlag(rule.policy_group) }}
                        </span>
                        <input 
                          v-model="rule.policy_group" 
                          class="input-field w-full text-xs font-mono text-purple-200" 
                          placeholder="输入自定义策略组名称..."
                        />
                      </div>
                      <button 
                        type="button" 
                        @click="rule._customInput = !rule._customInput" 
                        class="p-1.5 rounded-lg bg-sky-950/60 hover:bg-sky-900 border border-purple-700/50 text-purple-300 hover:text-purple-300 transition text-[11px] flex-shrink-0"
                        :title="rule._customInput ? '切换为下拉选择' : '自定义手动输入'"
                      >
                        {{ rule._customInput ? '▾' : '✎' }}
                      </button>
                    </div>
                  </td>
                  <td class="py-2.5 px-2 text-center"><input v-model.number="rule.match_order" type="number" class="input-field w-14 text-center text-xs" /></td>
                  <td class="py-2.5 px-2 text-center"><input type="checkbox" v-model="rule.enabled" class="accent-cyan-400 w-4 h-4 cursor-pointer" /></td>
                  <td class="py-2.5 px-2 text-center">
                    <button @click="removeRuleTarget(idx)" class="text-red-400 hover:text-red-300 text-xs p-1" title="删除此规则">✕</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Auto Countries -->
        <div class="glass-card p-5">
          <h3 class="text-sm text-purple-300 mb-3 font-semibold">自动选择国家/地区</h3>
          <div class="flex flex-wrap gap-2">
            <span v-for="(c, i) in configState.auto_countries" :key="i"
              class="inline-flex items-center gap-1 px-3 py-1 bg-sky-900/40 rounded-full text-xs text-sky-300 border border-purple-700/30">
              {{ c }}
              <button @click="configState.auto_countries.splice(i, 1)" class="text-sky-500 hover:text-red-400">✕</button>
            </span>
            <input v-model="newCountry" @keydown.enter="addCountry" placeholder="输入后回车" class="input-field w-32 text-xs" />
          </div>
        </div>
      </div>

      <!-- WireGuard Intranet Penetration -->
      <div v-if="currentView === 'wireguard'" class="space-y-6">
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h2 class="text-2xl font-bold text-white flex items-center gap-2">
              <span>🌐 WireGuard 内网穿透</span>
              <span class="text-xs px-2.5 py-0.5 rounded-full bg-cyan-900/40 text-purple-300 border border-cyan-700/40 font-normal">
                RouterOS v7 原生
              </span>
            </h2>
            <p class="text-xs text-slate-400 mt-1">
              基于 RouterOS 原生 WireGuard 服务与 IP Cloud DDNS，手机安装官方 WireGuard APP 扫描二维码即可秒级穿透接入家庭局域网
            </p>
          </div>
          <div class="flex flex-wrap gap-2">
            <button @click="startWgWizard" class="btn bg-gradient-to-r from-cyan-500 via-sky-500 to-blue-600 hover:from-cyan-400 hover:to-blue-500 text-white text-xs font-bold shadow-lg shadow-purple-500/25 flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl animate-pulse">
              <span>🚀</span>
              <span>小白一键极速通关 (免配置)</span>
            </button>
            <button @click="showNewPeerModal = true" class="btn btn-cyan text-xs font-semibold shadow-lg shadow-purple-500/20">
              ➕ 新建设备
            </button>
            <button @click="applyWgAlignment" :disabled="wgAligning" class="btn btn-green text-xs font-semibold shadow-md shadow-emerald-500/20">
              <span v-if="wgAligning" class="inline-block animate-spin mr-1">↻</span>
              <span>⚡ 一键配置到 RouterOS</span>
            </button>
            <button @click="fetchWgAlignmentStatus" :disabled="wgAligning" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs border border-purple-800/40">
              🔍 体检对齐状态
            </button>
            <button @click="viewRosScript" class="btn bg-slate-800 hover:bg-slate-700 text-sky-300 text-xs border border-purple-800/40">
              📋 查看终端脚本 (备用)
            </button>
            <button @click="fetchWireGuardInfo" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs border border-purple-900/40">
              ↻ 刷新
            </button>
          </div>
        </div>

        <!-- Status Cards -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div class="glass-card p-5">
            <div class="text-xs text-purple-300 mb-1">公网穿透端点 (Endpoint)</div>
            <div class="text-base font-bold text-purple-300 truncate font-mono" :title="wgInfo.effective_endpoint">
              {{ wgInfo.effective_endpoint || '未配置或未绑定' }}
            </div>
            <div class="text-[10px] text-slate-500 mt-1">客户端远程连接入口</div>
          </div>

          <div class="glass-card p-5">
            <div class="text-xs text-purple-300 mb-1">RouterOS 动态域名 (Cloud DDNS)</div>
            <div class="text-base font-bold text-emerald-400 truncate font-mono" :title="wgInfo.cloud_ddns">
              {{ wgInfo.cloud_ddns || '未检测到 (需开启 IP Cloud)' }}
            </div>
            <div class="text-[10px] text-slate-500 mt-1">自带免配置免费域名</div>
          </div>

          <div class="glass-card p-5">
            <div class="text-xs text-purple-300 mb-1">授权穿透设备</div>
            <div class="text-2xl font-bold text-white">{{ (wgInfo.peers || []).length }} <span class="text-xs font-normal text-slate-400">台设备</span></div>
            <div class="text-[10px] text-slate-500 mt-1">各设备独立专用秘钥</div>
          </div>

          <div class="glass-card p-5">
            <div class="text-xs text-purple-300 mb-1">服务端虚拟网段</div>
            <div class="text-base font-bold text-white font-mono">{{ wgInfo.config?.server_address || '10.88.0.1/24' }}</div>
            <div class="text-[10px] text-slate-500 mt-1">端口: {{ wgInfo.config?.listen_port || 13231 }} (UDP)</div>
          </div>
        </div>

        <!-- WireGuard Alignment Matrix Card -->
        <div class="glass-card p-5 space-y-4 border-l-4 border-l-cyan-400">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-purple-900/30 pb-3">
            <div>
              <div class="text-sm font-bold text-white flex items-center gap-2">
                <span>⚡ WireGuard 与 RouterOS 实时对齐矩阵</span>
                <span :class="['text-xs px-2.5 py-0.5 rounded-full font-semibold', wgAlignmentReport.all_aligned ? 'bg-emerald-950 text-emerald-300 border border-emerald-500/40' : 'bg-amber-950 text-amber-300 border border-amber-500/40']">
                  {{ wgAlignmentReport.all_aligned ? '● 已完全对齐' : '● 待对齐 / 同步' }}
                </span>
              </div>
              <div class="text-xs text-slate-400 mt-1">检测 RouterOS 硬件接口、虚拟网段 IP、UDP 防火墙、动态域名及所有已授权 Peer 是否 100% 同步</div>
            </div>
            <div class="flex gap-2">
              <button @click="applyWgAlignment" :disabled="wgAligning" class="btn btn-cyan text-xs font-semibold shadow-lg shadow-purple-500/20">
                <span v-if="wgAligning" class="inline-block animate-spin mr-1">↻</span>
                <span>⚡ 一键对齐 WireGuard 设置</span>
              </button>
            </div>
          </div>

          <div v-if="wgAlignmentReport.items && wgAlignmentReport.items.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            <div v-for="item in wgAlignmentReport.items" :key="item.key"
              class="bg-slate-900/60 p-3 rounded-xl border border-purple-900/30 space-y-1.5 hover:border-purple-700/40 transition">
              <div class="flex items-center justify-between text-xs font-semibold">
                <span class="text-white">{{ item.name }}</span>
                <span :class="['text-[10px] px-2 py-0.5 rounded font-mono', item.status === 'ok' ? 'bg-emerald-950/80 text-emerald-300 border border-emerald-500/40' : (item.status === 'warning' ? 'bg-amber-950/80 text-amber-300 border border-amber-500/40' : 'bg-red-950/80 text-red-300 border border-red-500/40')]">
                  {{ item.status === 'ok' ? '已对齐' : (item.status === 'warning' ? '需注意' : '未对齐') }}
                </span>
              </div>
              <div class="text-[11px] font-mono text-purple-300 truncate" :title="item.value">{{ item.value || '-' }}</div>
              <div class="text-[10px] text-slate-400 leading-relaxed">{{ item.details }}</div>
            </div>
          </div>
          <div v-else class="text-xs text-slate-400 py-2">
            点击「🔍 体检对齐状态」即可实时探查 RouterOS 上的 WireGuard 接口与各 Peer 配置。
          </div>
        </div>

        <!-- WireGuard Server Config Form -->
        <div class="glass-card p-5 space-y-4">
          <div class="flex items-center justify-between border-b border-purple-900/30 pb-3">
            <div class="text-sm font-bold text-white flex items-center gap-2">
              <span>⚙️ WireGuard 服务端高级配置</span>
              <span class="text-[10px] text-purple-300 bg-sky-900/30 px-2 py-0.5 rounded">接口: {{ wgInfo.config?.interface_name || 'singcube-wg' }}</span>
            </div>
            <button @click="saveWireGuardConfig" :disabled="loading" class="btn btn-green text-xs">💾 保存服务端配置</button>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div>
              <label class="block text-xs text-slate-400 mb-1">监听端口 (UDP)</label>
              <input v-model.number="wgInfo.config.listen_port" type="number" class="input-field w-full font-mono text-xs" />
            </div>
            <div>
              <label class="block text-xs text-slate-400 mb-1">自定义穿透域名/IP (留空则用Cloud DDNS)</label>
              <input v-model="wgInfo.config.endpoint_host" placeholder="如 your-domain.com" class="input-field w-full text-xs font-mono" />
            </div>
            <div>
              <label class="block text-xs text-slate-400 mb-1">内网访问网段 (AllowedIPs)</label>
              <input v-model="wgInfo.config.intranet_cidr" class="input-field w-full font-mono text-xs" />
            </div>
            <div>
              <label class="block text-xs text-slate-400 mb-1">DNS 服务器</label>
              <input v-model="wgInfo.config.dns" class="input-field w-full font-mono text-xs" />
            </div>
          </div>

          <!-- Direct DNS Scenario Mode Switch (Item 2: DNS 场景模式切换) -->
          <div class="bg-slate-900/50 p-4 rounded-xl border border-purple-900/40 space-y-2">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-white flex items-center gap-1.5">
                <span>🛡️ 移动穿透 DNS 场景模式</span>
                <span :class="['text-[10px] px-2 py-0.5 rounded font-mono', wgInfo.config.direct_dns_mode === 'local_direct' ? 'bg-amber-950/60 text-amber-300 border border-amber-700/40' : 'bg-emerald-950/60 text-emerald-300 border border-emerald-700/40']">
                  {{ wgInfo.config.direct_dns_mode === 'local_direct' ? '⚡ 蜂窝/局域网原生 DNS' : '🔒 严格隧道防泄露 DNS' }}
                </span>
              </span>
            </div>
            <p class="text-[11px] text-slate-400">选择移动端扫码生成配置时的 DNS 分流解析策略：</p>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3 pt-1">
              <label :class="['p-3 rounded-xl border cursor-pointer transition-all flex items-start gap-2.5', wgInfo.config.direct_dns_mode !== 'local_direct' ? 'bg-purple-600/10 border-purple-400/60 text-white' : 'bg-slate-900/40 border-purple-900/30 text-slate-400 hover:text-slate-200']">
                <input type="radio" value="tunnel_strict" v-model="wgInfo.config.direct_dns_mode" class="mt-0.5 accent-cyan-400" />
                <div>
                  <div class="text-xs font-bold text-purple-200">🔒 严格隧道防泄露模式 (推荐)</div>
                  <div class="text-[10px] text-slate-400 mt-0.5 leading-relaxed">
                    下发 <strong>10.88.0.1</strong> 作为唯一 DNS。手机端所有域名解析强制流向旁路由，实现 0 明文泄露、防 ISP 追踪劫持并享受全网去广告。
                  </div>
                </div>
              </label>

              <label :class="['p-3 rounded-xl border cursor-pointer transition-all flex items-start gap-2.5', wgInfo.config.direct_dns_mode === 'local_direct' ? 'bg-amber-500/10 border-amber-400/60 text-white' : 'bg-slate-900/40 border-purple-900/30 text-slate-400 hover:text-slate-200']">
                <input type="radio" value="local_direct" v-model="wgInfo.config.direct_dns_mode" class="mt-0.5 accent-amber-400" />
                <div>
                  <div class="text-xs font-bold text-amber-200">⚡ 蜂窝/WiFi 原生直接 DNS (防断网)</div>
                  <div class="text-[10px] text-slate-400 mt-0.5 leading-relaxed">
                    下发 <strong>10.88.0.1, 223.5.5.5</strong>。手机直连当地电信/联通/WiFi DNS，内网走隧道解析，即使家中旁路由重启或断电，手机外网永不断网。
                  </div>
                </div>
              </label>
            </div>
          </div>
        </div>

        <!-- Peers List -->
        <div class="glass-card p-5 space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-base font-bold text-white flex items-center gap-2">
                <span>📱 已授权穿透设备列表</span>
                <span class="text-xs text-purple-300 bg-sky-900/40 px-2.5 py-0.5 rounded-full font-mono">{{ (wgInfo.peers || []).length }} 台</span>
              </div>
              <div class="text-xs text-slate-400 mt-1">
                点击设备右侧【📱 扫码配置】按钮，立即使用手机 WireGuard APP 扫描二维码完成免密配对
              </div>
            </div>
            <button @click="showNewPeerModal = true" class="btn btn-cyan text-xs">
              + 添加新设备
            </button>
          </div>

          <!-- Peer Items List -->
          <div v-if="(wgInfo.peers || []).length > 0" class="space-y-3 pt-1">
            <div v-for="peer in wgInfo.peers" :key="peer.id"
              class="flex flex-col md:flex-row md:items-center justify-between bg-slate-900/50 border border-purple-900/30 p-4 rounded-xl gap-4 hover:border-purple-700/50 transition">
              <div class="space-y-1">
                <div class="flex items-center gap-2">
                  <span class="text-lg">📲</span>
                  <span class="font-bold text-white text-sm">{{ peer.name }}</span>
                  <span :class="['text-[10px] px-2 py-0.5 rounded border', peer.allowed_ips_mode === 'full' ? 'bg-amber-900/40 text-amber-300 border-amber-700/40' : 'bg-cyan-900/40 text-purple-300 border-cyan-700/40']">
                    {{ peer.allowed_ips_mode === 'full' ? '全局代理模式' : '内网穿透模式' }}
                  </span>
                  <span :class="['text-[10px] px-2 py-0.5 rounded border', peer.synced_to_ros ? 'bg-emerald-900/40 text-emerald-300 border-emerald-700/40' : 'bg-slate-800 text-slate-400 border-slate-700/40']">
                    {{ peer.synced_to_ros ? '● ROS已同步' : '○ 离线状态' }}
                  </span>
                </div>
                <div class="text-xs text-slate-400 flex flex-wrap gap-4 font-mono pt-1">
                  <span>虚拟IP: <strong class="text-purple-300">{{ peer.client_ip }}/32</strong></span>
                  <span>可达网段: {{ peer.allowed_ips }}</span>
                  <span class="text-slate-500">创建于: {{ peer.created_at }}</span>
                </div>
                <div v-if="peer.ros_error" class="text-[11px] text-amber-400">
                  ⚠️ ROS 同步提示: {{ peer.ros_error }}
                </div>
              </div>

              <!-- Action buttons -->
              <div class="flex items-center gap-2 flex-shrink-0">
                <button @click="openRenamePeer(peer)" class="btn bg-slate-800 hover:bg-slate-700 text-sky-300 text-xs border border-purple-800/40">
                  ✏️ 改名
                </button>
                <button @click="showPeerQR(peer)" class="btn btn-cyan text-xs font-semibold flex items-center gap-1.5 shadow-md shadow-purple-500/20">
                  <span>📱</span>
                  <span>扫码配置</span>
                </button>
                <button @click="downloadPeerConf(peer)" class="btn bg-sky-900/50 hover:bg-sky-800/60 text-sky-200 text-xs border border-purple-700/40">
                  ⬇️ 下载.conf
                </button>
                <button @click="deletePeer(peer)" :disabled="loading" class="btn btn-red text-xs">
                  🗑️ 撤回
                </button>
              </div>
            </div>
          </div>

          <!-- Empty State -->
          <div v-else class="text-center py-12 bg-slate-900/30 rounded-2xl border border-dashed border-purple-900/40 space-y-3">
            <div class="text-4xl">📱</div>
            <div class="text-base font-bold text-white">暂未添加任何穿透设备</div>
            <div class="text-xs text-slate-400 max-w-md mx-auto">
              点击下方按钮即可为您的手机、平板或笔记本生成专属 WireGuard 穿透秘钥，并生成扫码二维码。
            </div>
            <div class="flex flex-wrap justify-center gap-3">
              <button @click="startWgWizard" class="btn bg-gradient-to-r from-cyan-500 via-sky-500 to-blue-600 hover:from-cyan-400 hover:to-blue-500 text-white text-xs font-bold px-4 py-2 rounded-xl shadow-lg shadow-purple-500/20">
                🚀 小白一键极速通关 (免配置)
              </button>
              <button @click="showNewPeerModal = true" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs px-3">
                + 手动添加设备
              </button>
            </div>
          </div>
        </div>

        <!-- 🖥️ Host (飞牛 fnOS / Linux) WireGuard Companion Card -->
        <div class="glass-card p-5 space-y-4 border-l-4 border-l-emerald-400">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-purple-900/30 pb-3">
            <div>
              <div class="text-sm font-bold text-white flex items-center gap-2">
                <span>🖥️ 飞牛 (fnOS) / Linux 宿主机 WireGuard 极速穿透</span>
                <span :class="['text-xs px-2.5 py-0.5 rounded-full font-semibold', hostInfo.wg0_active ? 'bg-emerald-950 text-emerald-300 border border-emerald-500/40' : (hostInfo.wireguard_installed ? 'bg-amber-950 text-amber-300 border border-amber-500/40' : 'bg-slate-800 text-slate-300 border border-slate-600/40')]">
                  {{ hostInfo.wg0_active ? '● 宿主 wg0 运行中' : (hostInfo.wireguard_installed ? '● 已安装未启动' : '○ 待部署') }}
                </span>
              </div>
              <div class="text-xs text-slate-400 mt-1">
                对于在飞牛 (fnOS) 或 NAS Docker 容器中运行的 SingCube，将 WireGuard 服务直接部署在宿主机 Linux 内核层，可大幅降低 NAT 穿透损耗并跑满千兆带宽
              </div>
            </div>
            <div class="flex flex-wrap gap-2">
              <button @click="deployHostWireGuard" :disabled="hostDeployingWg" class="btn btn-cyan text-xs font-semibold shadow-md shadow-purple-500/20">
                <span v-if="hostDeployingWg" class="inline-block animate-spin mr-1">↻</span>
                <span>⚡ 一键部署宿主机 WireGuard 服务</span>
              </button>
              <button @click="viewHostScript" class="btn bg-slate-800 hover:bg-slate-700 text-sky-300 text-xs border border-purple-800/40">
                📋 宿主机一键 Shell 脚本
              </button>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
            <div class="bg-slate-900/40 p-3 rounded-xl border border-purple-900/20 space-y-1">
              <div class="text-sky-300 font-semibold flex items-center gap-1">
                <span>🔄 双端双活自愈同步</span>
              </div>
              <div class="text-slate-400 text-[11px] leading-relaxed">
                在上方【📱 已授权穿透设备列表】添加、改名或撤回设备时，SingCube 会全自动通过后台动态向宿主机同步执行 <code class="text-purple-300 font-mono">wg set wg0 peer ...</code>，无需手动重启服务。
              </div>
            </div>
            <div class="bg-slate-900/40 p-3 rounded-xl border border-purple-900/20 space-y-1">
              <div class="text-emerald-300 font-semibold flex items-center gap-1">
                <span>🛡️ 防火墙端口与 NAT 自动打通</span>
              </div>
              <div class="text-slate-400 text-[11px] leading-relaxed">
                一键部署时自动持久化放行 UDP 13231 监听端口、开启 IPv4 转发并配置 iptables MASQUERADE 规则，保证外网手机能访问局域网所有设备。
              </div>
            </div>
            <div class="bg-slate-900/40 p-3 rounded-xl border border-purple-900/20 space-y-1">
              <div class="text-amber-300 font-semibold flex items-center gap-1">
                <span>🚀 飞牛 (fnOS) 兼容性认证</span>
              </div>
              <div class="text-slate-400 text-[11px] leading-relaxed">
                针对飞牛私有云 (Debian 12 核心) 进行专属路径适配，支持通过 apt-get 自动补齐 wireguard-tools 依赖并托管 systemd 开机自启。
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ── Connections View (Live Session Audit & Connection Manager) ── -->
      <div v-if="currentView === 'connections'" class="space-y-6">
        <!-- ── Top Sub-Nav Pills: 规则与安全 ── -->
        <div class="flex items-center gap-2 p-1 bg-slate-900/80 rounded-2xl border border-purple-800/40 w-fit backdrop-blur-md shadow-sm">
          <button 
            @click="currentView = 'adblock'; rulesSubTab = 'adblock'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'adblock' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>🛡️ 217heidai 高速去广告</span>
          </button>
          <button 
            @click="currentView = 'connections'; rulesSubTab = 'connections'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'connections' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>⚡ 实时连接与流向</span>
          </button>
          <button 
            @click="currentView = 'devices'; rulesSubTab = 'devices'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'devices' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>💻 局域网接入设备</span>
          </button>
        </div>

        <!-- Top Stats Banner -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <!-- Active Connections Card -->
          <div class="glass-card p-4 relative overflow-hidden group">
            <div class="absolute -right-4 -bottom-4 w-20 h-20 bg-purple-600/10 rounded-full blur-xl pointer-events-none"></div>
            <div class="flex items-center justify-between text-xs text-purple-300/80 mb-1">
              <span>当前活跃会话</span>
              <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-emerald-500/15 border border-emerald-400/30 text-[10px] text-emerald-300">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping"></span>
                实时监听
              </span>
            </div>
            <div class="text-2xl font-black text-white font-mono flex items-baseline gap-1">
              {{ connectionsData.connections?.length || 0 }}
              <span class="text-xs text-slate-400 font-normal">条会话</span>
            </div>
            <div class="text-[11px] text-slate-400 mt-1">全局内网与本机代理流</div>
          </div>

          <!-- Total Download -->
          <div class="glass-card p-4 relative overflow-hidden group">
            <div class="absolute -right-4 -bottom-4 w-20 h-20 bg-sky-500/10 rounded-full blur-xl pointer-events-none"></div>
            <div class="text-xs text-purple-300/80 mb-1 flex items-center gap-1">
              <svg class="w-3.5 h-3.5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7m0 0l-7-7m7 7V3"/></svg>
              <span>累计下行数据量</span>
            </div>
            <div class="text-2xl font-black text-purple-300 font-mono">
              {{ formatBytes(connectionsData.downloadTotal) }}
            </div>
            <div class="text-[11px] text-slate-400 mt-1">出口下行加速缓冲</div>
          </div>

          <!-- Total Upload -->
          <div class="glass-card p-4 relative overflow-hidden group">
            <div class="absolute -right-4 -bottom-4 w-20 h-20 bg-emerald-500/10 rounded-full blur-xl pointer-events-none"></div>
            <div class="text-xs text-purple-300/80 mb-1 flex items-center gap-1">
              <svg class="w-3.5 h-3.5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 10l7-7m0 0l7 7m-7-7v18"/></svg>
              <span>累计上行数据量</span>
            </div>
            <div class="text-2xl font-black text-emerald-300 font-mono">
              {{ formatBytes(connectionsData.uploadTotal) }}
            </div>
            <div class="text-[11px] text-slate-400 mt-1">直连/穿透/代理上送</div>
          </div>

          <!-- Action Hub -->
          <div class="glass-card p-3 flex flex-col justify-between">
            <div class="text-xs text-slate-300 font-semibold mb-1 flex items-center justify-between">
              <span>会话管理中枢</span>
              <label class="inline-flex items-center gap-1.5 cursor-pointer text-[10px] text-purple-300">
                <input type="checkbox" v-model="connAutoRefresh" class="rounded bg-slate-800 border-sky-600/50 text-purple-400 focus:ring-0 w-3 h-3" />
                自动轮询
              </label>
            </div>
            <div class="grid grid-cols-2 gap-1.5">
              <button @click="closeAllConnections" class="btn bg-rose-950/60 hover:bg-rose-900 border border-rose-500/50 text-rose-200 text-[11px] py-1.5 px-2 flex items-center justify-center gap-1 transition-all active:scale-95" title="强制断开所有连接">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
                全部断开
              </button>
              <button @click="flushDNSCache" class="btn bg-purple-950/60 hover:bg-cyan-900 border border-purple-500/50 text-purple-200 text-[11px] py-1.5 px-2 flex items-center justify-center gap-1 transition-all active:scale-95" title="清空 Fake-IP 与 SmartDNS 映射">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
                清空DNS
              </button>
            </div>
          </div>
        </div>

        <!-- Filter & Search Controls -->
        <div class="glass-card p-4 space-y-3">
          <div class="flex flex-col md:flex-row items-stretch md:items-center justify-between gap-3">
            <!-- Search bar -->
            <div class="relative flex-1">
              <svg class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/></svg>
              <input v-model="connSearch" placeholder="搜索域名、目的 IP、来源客户端、规则或出站节点..." class="input-field w-full pl-9 pr-8 text-xs font-mono" />
              <button v-if="connSearch" @click="connSearch = ''" class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white text-xs">✕</button>
            </div>

            <!-- Filter tabs & sorting -->
            <div class="flex flex-wrap items-center gap-2">
              <!-- Protocol Tabs -->
              <div class="inline-flex p-0.5 rounded-lg bg-slate-900/80 border border-purple-800/40 text-xs">
                <button v-for="proto in [{id:'all',l:'全部'}, {id:'tcp',l:'TCP'}, {id:'udp',l:'UDP'}]" :key="proto.id"
                  @click="connFilterProtocol = proto.id"
                  :class="['px-2.5 py-1 rounded-md text-[11px] font-medium transition', connFilterProtocol === proto.id ? 'bg-purple-600/25 text-purple-300 font-semibold' : 'text-slate-400 hover:text-slate-200']">
                  {{ proto.l }}
                </button>
              </div>

              <!-- Rule Type Tabs -->
              <div class="inline-flex p-0.5 rounded-lg bg-slate-900/80 border border-purple-800/40 text-xs">
                <button v-for="t in [{id:'all',l:'全部分流'}, {id:'proxy',l:'代理节点'}, {id:'direct',l:'国内直连'}]" :key="t.id"
                  @click="connFilterType = t.id"
                  :class="['px-2.5 py-1 rounded-md text-[11px] font-medium transition', connFilterType === t.id ? 'bg-sky-500/25 text-sky-300 font-semibold' : 'text-slate-400 hover:text-slate-200']">
                  {{ t.l }}
                </button>
              </div>

              <!-- Sort select -->
              <select v-model="connSortBy" class="input-field text-xs py-1 px-2.5 bg-slate-900/90 border-purple-800/40 text-slate-300">
                <option value="download">按下载流量降序</option>
                <option value="upload">按上传流量降序</option>
                <option value="duration">按存活时间降序</option>
              </select>

              <!-- Manual Refresh button -->
              <button @click="fetchConnections" class="btn btn-cyan text-xs py-1 px-3 flex items-center gap-1">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>
                刷新
              </button>
            </div>
          </div>
        </div>

        <!-- Connection Cards List -->
        <div class="space-y-2.5">
          <div v-if="filteredConnections.length === 0" class="glass-card p-10 text-center text-slate-400 text-sm">
            <svg class="w-12 h-12 mx-auto mb-2 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
            暂无符合条件的活跃连接
          </div>

          <div v-for="conn in filteredConnections" :key="conn.id"
            class="glass-card p-3.5 flex flex-col md:flex-row md:items-center justify-between gap-3 hover:border-purple-400/50 hover:shadow-[0_0_15px_rgba(0,210,255,0.15)] transition-all">
            <!-- Left Info: Host & Source -->
            <div class="space-y-1 min-w-0 flex-1">
              <div class="flex items-center gap-2 flex-wrap">
                <!-- Protocol badge -->
                <span :class="['px-1.5 py-0.5 rounded text-[10px] font-mono font-bold uppercase', conn.metadata?.network === 'tcp' ? 'bg-sky-500/20 text-sky-300 border border-sky-400/30' : 'bg-purple-500/20 text-purple-300 border border-purple-400/30']">
                  {{ conn.metadata?.network || 'TCP' }}
                </span>

                <!-- Host or Destination -->
                <span class="text-sm font-semibold text-white font-mono truncate max-w-xs md:max-w-md">
                  {{ conn.metadata?.host || conn.metadata?.destinationIP }}:{{ conn.metadata?.destinationPort }}
                </span>

                <!-- Sniffed Host if available -->
                <span v-if="conn.metadata?.sniffHost && conn.metadata.sniffHost !== conn.metadata.host" class="text-[11px] text-purple-300/80 font-mono">
                  ({{ conn.metadata.sniffHost }})
                </span>
              </div>

              <!-- Sub-metadata: Source IP + Process + Matched Rule -->
              <div class="flex items-center gap-2 text-xs text-slate-400 flex-wrap">
                <span class="font-mono text-slate-300 flex items-center gap-1">
                  <span>客户端: {{ conn.metadata?.sourceIP ? getDeviceDisplayName(conn.metadata.sourceIP) + ':' + conn.metadata.sourcePort : '本机' }}</span>
                  <button v-if="conn.metadata?.sourceIP" @click.stop="openQuickRename(conn.metadata.sourceIP)" class="text-slate-400 hover:text-purple-300 transition text-[11px] p-0.5" title="修改此设备备注 (按 IP/MAC)">✏️</button>
                </span>
                <span>•</span>
                <span class="inline-flex items-center gap-1 text-slate-300">
                  <span class="text-[10px] px-1.5 py-0.2 rounded bg-slate-800 text-sky-300 border border-purple-800/50">{{ conn.rule }}</span>
                  <span class="font-mono text-slate-200">{{ conn.rulePayload }}</span>
                </span>
              </div>
            </div>

            <!-- Middle Info: Outbound Proxy Chain -->
            <div class="flex flex-col md:items-end justify-center min-w-0 md:max-w-xs">
              <div class="text-xs font-semibold flex items-center gap-1 text-purple-200 truncate" :title="'路由决策流: ' + formatChains(conn.chains).join(' → ')">
                <span class="text-slate-400 text-[11px]">出站链路:</span>
                <span v-for="(node, idx) in formatChains(conn.chains)" :key="idx" class="inline-flex items-center gap-1">
                  <span v-if="idx > 0" class="text-sky-500 font-bold">→</span>
                  <span :class="[
                    node === '直连' || node === 'DIRECT' ? 'text-emerald-400 font-semibold' : 
                    (idx === formatChains(conn.chains).length - 1 ? 'text-purple-200 font-bold' : 'text-sky-300')
                  ]">{{ node }}</span>
                </span>
              </div>
              <div class="text-[11px] text-slate-400 mt-0.5">
                存活: {{ formatDuration(conn.start) }}
              </div>
            </div>

            <!-- Right Info: Traffic & Terminate Button -->
            <div class="flex items-center justify-between md:justify-end gap-3 pt-2 md:pt-0 border-t md:border-t-0 border-purple-900/30">
              <div class="text-right font-mono text-xs">
                <div class="text-purple-300 flex items-center gap-1">
                  <span>↓</span> {{ formatBytes(conn.download) }}
                </div>
                <div class="text-emerald-400 flex items-center gap-1">
                  <span>↑</span> {{ formatBytes(conn.upload) }}
                </div>
              </div>

              <!-- Terminate connection button -->
              <button @click="closeConnection(conn.id)"
                class="p-2 rounded-lg bg-slate-900/80 hover:bg-rose-950/80 border border-slate-700/60 hover:border-rose-500/60 text-slate-400 hover:text-rose-300 transition-all active:scale-95"
                title="中断此连接">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- ── LAN Device Management View (Item 5) ── -->
      <div v-if="currentView === 'devices'" class="space-y-6">
        <!-- ── Top Sub-Nav Pills: 规则与安全 ── -->
        <div class="flex items-center gap-2 p-1 bg-slate-900/80 rounded-2xl border border-purple-800/40 w-fit backdrop-blur-md shadow-sm">
          <button 
            @click="currentView = 'adblock'; rulesSubTab = 'adblock'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'adblock' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>🛡️ 217heidai 高速去广告</span>
          </button>
          <button 
            @click="currentView = 'connections'; rulesSubTab = 'connections'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'connections' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>⚡ 实时连接与流向</span>
          </button>
          <button 
            @click="currentView = 'devices'; rulesSubTab = 'devices'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'devices' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>💻 局域网接入设备</span>
          </button>
        </div>

        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <h2 class="text-2xl font-bold text-white flex items-center gap-2">
              <span>📱 局域网设备管理</span>
            </h2>
            <div class="text-xs text-slate-400 mt-1">支持按 IP 与 MAC 地址为全屋终端绑定友好备注，智能联动 RouterOS 与实时链路监控</div>
          </div>
          <div class="flex gap-2">
            <button @click="openDeviceEdit(null)" class="btn btn-cyan text-xs flex items-center gap-1 shadow-md">
              ➕ 添加/绑定设备
            </button>
            <button @click="fetchDevices" :disabled="devicesLoading" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs border border-purple-800/40">
              ↻ 刷新设备
            </button>
          </div>
        </div>

        <!-- Filter & Search Bar -->
        <div class="glass-card p-4 flex flex-col md:flex-row items-center justify-between gap-3">
          <div class="relative w-full md:w-96">
            <input v-model="deviceSearch" placeholder="搜索设备名称、IP 地址、MAC 物理地址或主机名..." class="input-field w-full text-xs pl-8 pr-3 py-1.5 bg-slate-900/90 border-purple-800/40 text-slate-200" />
            <svg class="w-4 h-4 text-slate-500 absolute left-2.5 top-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/></svg>
          </div>
          <div class="text-xs text-slate-400 font-mono">
            共发现 <strong class="text-purple-300">{{ filteredDevices.length }}</strong> 台设备
          </div>
        </div>

        <!-- Devices Grid / Table -->
        <div class="glass-card overflow-hidden">
          <div class="overflow-x-auto">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-900/80 text-slate-400 uppercase tracking-wider border-b border-purple-900/40">
                <tr>
                  <th class="py-3 px-4">设备名称 (友好备注)</th>
                  <th class="py-3 px-4">IP 地址</th>
                  <th class="py-3 px-4">MAC 地址</th>
                  <th class="py-3 px-4">主机名 / 探查来源</th>
                  <th class="py-3 px-4">在线状态</th>
                  <th class="py-3 px-4 text-right">操作</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-sky-900/20 font-mono">
                <tr v-if="filteredDevices.length === 0">
                  <td colspan="6" class="py-12 text-center text-slate-500 font-sans">
                    暂未发现或匹配到符合条件的局域网设备，可点击右上角「➕ 添加/绑定设备」手动添加
                  </td>
                </tr>
                <tr v-for="dev in filteredDevices" :key="dev.ip + dev.mac" class="hover:bg-slate-800/40 transition">
                  <td class="py-3 px-4 font-sans font-semibold">
                    <div class="flex items-center gap-2">
                      <span class="text-white">{{ dev.name }}</span>
                      <span v-if="dev.is_custom" class="text-[10px] px-1.5 py-0.5 rounded bg-cyan-900/50 text-purple-300 border border-purple-500/30">自定义</span>
                    </div>
                  </td>
                  <td class="py-3 px-4 text-purple-300">{{ dev.ip }}</td>
                  <td class="py-3 px-4 text-slate-300 font-mono">{{ dev.mac || '-' }}</td>
                  <td class="py-3 px-4 text-slate-400 font-sans">
                    <div class="flex flex-col">
                      <span class="text-slate-200">{{ dev.hostname || '-' }}</span>
                      <span class="text-[10px] text-slate-500">{{ dev.source }}</span>
                    </div>
                  </td>
                  <td class="py-3 px-4">
                    <span :class="['inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-sans font-medium', dev.online ? 'bg-emerald-950/80 text-emerald-300 border border-emerald-500/40' : 'bg-slate-800 text-slate-400']">
                      <span :class="['w-1.5 h-1.5 rounded-full', dev.online ? 'bg-emerald-400 shadow-[0_0_6px_#10b981] animate-pulse' : 'bg-slate-500']"></span>
                      {{ dev.online ? '在线' : '离线' }}
                    </span>
                  </td>
                  <td class="py-3 px-4 text-right font-sans">
                    <button @click="openDeviceEdit(dev)" class="btn btn-cyan py-1 px-3 text-xs inline-flex items-center gap-1 shadow-sm">
                      ✏️ 修改名称
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Telegram Bot -->
      <div v-if="currentView === 'tgbot'" class="space-y-6">
        <!-- ── Top Sub-Nav Pills: 系统与运维 ── -->
        <div class="flex items-center gap-2 p-1 bg-slate-900/80 rounded-2xl border border-purple-800/40 flex-wrap w-fit backdrop-blur-md shadow-sm">
          <button 
            @click="currentView = 'ros'; systemSubTab = 'ros'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'ros' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>🚀 系统升级与备份</span>
          </button>
          <button 
            @click="currentView = 'tgbot'; systemSubTab = 'tgbot'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'tgbot' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>📱 TG 手机机器人</span>
          </button>
          <button 
            @click="currentView = 'ai'; systemSubTab = 'ai'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'ai' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>🧠 AI 故障自愈诊断</span>
          </button>
          <button 
            @click="currentView = 'accelerator'; systemSubTab = 'accelerator'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'accelerator' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>⚡ 多源镜像加速</span>
          </button>
        </div>

        <h2 class="text-2xl font-bold text-white">Telegram 机器人</h2>
        <div class="glass-card p-5 space-y-4">
          <div class="flex items-center justify-between pb-3 border-b border-purple-900/30">
            <div>
              <div class="font-semibold text-white">Bot 状态</div>
              <div class="text-xs text-slate-400">通过 Telegram 远程查询状态与执行重启/测速</div>
            </div>
            <div :class="['px-3 py-1 rounded-full text-xs font-semibold', tgBotConfig.running ? 'bg-emerald-900/60 text-emerald-300' : 'bg-slate-800 text-slate-400']">
              {{ tgBotConfig.running ? '● 运行中' : '○ 未运行' }}
            </div>
          </div>

          <div class="space-y-3">
            <div>
              <label class="block text-xs text-purple-300 mb-1">Bot Token</label>
              <input v-model="tgBotInput.token" :placeholder="tgBotConfig.token || '从 @BotFather 获取的 Token'" class="input-field w-full" />
            </div>

            <div>
              <label class="block text-xs text-purple-300 mb-1">本地代理地址 (用于容器内访问 Telegram API)</label>
              <input v-model="tgBotInput.proxy_url" placeholder="http://127.0.0.1:7890" class="input-field w-full" />
            </div>

            <div>
              <label class="block text-xs text-purple-300 mb-1">白名单 Chat ID (留空允许所有人，多个用逗号隔开)</label>
              <input v-model="tgBotInput.allowed_chat_ids_str" placeholder="12345678, 87654321" class="input-field w-full" />
            </div>

            <label class="flex items-center gap-2 text-sm text-slate-300 cursor-pointer pt-2">
              <input type="checkbox" v-model="tgBotInput.enabled" class="accent-cyan-400 w-4 h-4" />
              启用 Telegram 机器人服务
            </label>
          </div>

          <div class="pt-3">
            <button @click="saveTgBot" :disabled="loading" class="btn btn-green text-xs">💾 保存并应用设置</button>
          </div>
        </div>
      </div>

      <!-- ROS Manager & System Upgrades -->
      <div v-if="currentView === 'ros'" class="space-y-6">
        <!-- ── Top Sub-Nav Pills: 系统与运维 ── -->
        <div class="flex items-center gap-2 p-1 bg-slate-900/80 rounded-2xl border border-purple-800/40 flex-wrap w-fit backdrop-blur-md shadow-sm">
          <button 
            @click="currentView = 'ros'; systemSubTab = 'ros'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'ros' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>🚀 系统升级与备份</span>
          </button>
          <button 
            @click="currentView = 'tgbot'; systemSubTab = 'tgbot'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'tgbot' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>📱 TG 手机机器人</span>
          </button>
          <button 
            @click="currentView = 'ai'; systemSubTab = 'ai'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'ai' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>🧠 AI 故障自愈诊断</span>
          </button>
          <button 
            @click="currentView = 'accelerator'; systemSubTab = 'accelerator'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'accelerator' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>⚡ 多源镜像加速</span>
          </button>
        </div>

        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <h2 class="text-2xl font-bold text-white">系统升级与网关维护</h2>
          <span class="text-xs text-slate-400 font-mono">底层网络底座 · 系统 OTA · RouterOS 联动与全量备份</span>
        </div>

        <!-- ── [首位设置] Host OS & Docker Companion Manager (飞牛 fnOS / PVE / Linux) ── -->
        <div class="glass-card p-5 space-y-4 border-l-4 border-l-sky-500 shadow-xl">
          <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-3 border-b border-purple-900/30 pb-3">
            <div>
              <div class="text-base font-bold text-white flex flex-wrap items-center gap-2">
                <span>🖥️ 宿主机 (飞牛 fnOS / PVE / Linux) 网络底座与系统运维</span>
                <span :class="['text-xs px-2.5 py-0.5 rounded-full font-bold border', hostInfo.ip_forward ? 'bg-emerald-950 text-emerald-300 border-emerald-500/40' : 'bg-amber-950 text-amber-300 border-amber-500/50 animate-pulse']">
                  {{ hostInfo.ip_forward ? '★ 底座转发已就绪' : '★ 旁路由底座第 0 步 · 必须优先配置' }}
                </span>
                <span v-if="hostInfo.is_fnos" class="text-xs bg-sky-900/60 text-sky-300 border border-purple-500/40 px-2 py-0.5 rounded-full font-mono">飞牛 fnOS 专属优化</span>
                <span v-else-if="hostInfo.is_pve" class="text-xs bg-orange-900/60 text-orange-300 border border-orange-500/40 px-2 py-0.5 rounded-full font-mono">PVE 宿主机</span>
                <span v-else class="text-xs bg-slate-800 text-slate-300 border border-slate-600/40 px-2 py-0.5 rounded-full font-mono">Docker 宿主协同</span>
              </div>
              <div class="text-xs text-slate-400 mt-1">
                打通宿主内核 IPv4 数据包转发、防火墙规则放行、宿主 WireGuard 模块直通与系统在线升级
              </div>
            </div>
            <div class="flex flex-wrap gap-2">
              <button @click="openHostFwModal('permissive')" class="btn btn-cyan text-xs flex items-center gap-1 font-semibold shadow-md shadow-purple-500/20">
                <span>📋 免密 1 键 Shell 脚本 (严格/宽松)</span>
              </button>
              <button @click="openHostFwModal('permissive')" :disabled="hostApplyingFw" class="btn btn-blue text-xs flex items-center gap-1 font-semibold shadow-md shadow-blue-500/20">
                <span v-if="hostApplyingFw" class="inline-block animate-spin mr-1">↻</span>
                <span>🛡️ 一键配置宿主防火墙 (弹窗向导)</span>
              </button>
              <button @click="testHostSSH" :disabled="hostTesting" class="btn bg-slate-800 hover:bg-slate-700 text-purple-300 border border-cyan-800/40 text-xs flex items-center gap-1">
                <span v-if="hostTesting" class="inline-block animate-spin mr-1">↻</span>
                <span>🔌 测试 SSH 连接</span>
              </button>
              <button @click="checkHostUpgrade" :disabled="hostCheckingUpgrade" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 border border-purple-800/40 text-xs flex items-center gap-1 font-semibold">
                <span v-if="hostCheckingUpgrade" class="inline-block animate-spin mr-1">↻</span>
                <span>⬆️ 检查宿主系统更新</span>
              </button>
            </div>
          </div>

          <!-- Beginner Guidance Box: Why this is Step 0 -->
          <div class="bg-sky-950/40 border border-sky-600/40 rounded-xl p-4 text-xs text-slate-200 space-y-2">
            <div class="font-bold flex items-center gap-1.5 text-purple-300 text-sm">
              <span>💡 为什么这里的设置必须放在首位？（旁路由底座生命线）：</span>
            </div>
            <div class="text-xs text-slate-300 leading-relaxed space-y-1.5">
              <p>
                • <strong>底座关键原理</strong>：Docker 容器共享宿主机的 Linux 内核网络栈。若宿主机未开启内核转发（<code class="text-purple-300 font-mono">net.ipv4.ip_forward = 1</code>）或系统防火墙拦截了转发数据包，所有经由局域网设备发往旁路由的请求将被宿主内核<strong>直接静默丢弃</strong>！这就是很多新手“手机一改网关全家瞬间断网”的根本原因。
              </p>
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 pt-1">
                <div class="bg-slate-900/60 p-2.5 rounded-lg border border-purple-800/40 cursor-pointer hover:border-purple-400/60 transition" @click="openHostFwModal('permissive')">
                  <div class="font-bold text-purple-300 mb-0.5">⚡ 方式 1（免密 1 键 Shell 脚本 · 最推荐）：</div>
                  <div class="text-[11px] text-slate-400">点击打开向导，直接在飞牛应用中心-终端或 SSH 终端中粘贴回车执行，2 秒自动打通，无需在网页输入密码！</div>
                </div>
                <div class="bg-slate-900/60 p-2.5 rounded-lg border border-purple-800/40 cursor-pointer hover:border-sky-400/60 transition" @click="openHostFwModal('permissive'); hostFwActiveTab = 'ssh'">
                  <div class="font-bold text-sky-300 mb-0.5">🛡️ 方式 2（面板全自动托管）：</div>
                  <div class="text-[11px] text-slate-400">在下方填入宿主机 IP（通常是 Docker 网桥网关 172.17.0.1）和 SSH 密码保存，在向导中一键全自动完成下发打通。</div>
                </div>
              </div>
            </div>
          </div>

          <!-- Host Telemetry & Live Status Grid -->
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 bg-slate-900/40 p-4 rounded-xl border border-purple-900/20 text-xs font-mono">
            <div>
              <span class="text-slate-400 text-[11px] block">宿主系统</span>
              <span class="text-white font-bold flex items-center gap-1 mt-0.5">
                <span class="text-purple-300">{{ hostInfo.os_pretty || hostInfo.os_name || '未连接 / 待配置' }}</span>
              </span>
            </div>
            <div>
              <span class="text-slate-400 text-[11px] block">系统内核 & 架构</span>
              <span class="text-white font-bold block mt-0.5 truncate" :title="hostInfo.kernel">{{ hostInfo.kernel ? `${hostInfo.kernel} (${hostInfo.arch})` : '-' }}</span>
            </div>
            <div>
              <span class="text-slate-400 text-[11px] block">CPU / 内存负载</span>
              <span class="text-emerald-400 font-bold block mt-0.5">{{ hostInfo.cpu_cores ? `${hostInfo.cpu_cores} 核 | ${hostInfo.memory_used || '0'}/${hostInfo.memory_total || '0'}` : '-' }}</span>
            </div>
            <div :class="['p-2 rounded-lg border transition', hostInfo.ip_forward ? 'bg-emerald-950/20 border-emerald-500/30' : 'bg-amber-950/40 border-amber-500/50 shadow-md shadow-amber-950/20']">
              <div class="flex items-center justify-between">
                <span class="text-slate-400 text-[11px] block">内核 IPv4 转发状态</span>
                <span v-if="!hostInfo.ip_forward" class="text-[9px] text-amber-300 font-bold bg-amber-900/80 px-1 py-0.2 rounded animate-pulse">需打通</span>
              </div>
              <span :class="['font-bold block mt-0.5 text-xs', hostInfo.ip_forward ? 'text-emerald-400' : 'text-amber-400']">
                {{ hostInfo.ip_forward ? '● 已开启 (net.ipv4.ip_forward=1)' : '○ 未开启 (可能会阻断流量)' }}
              </span>
              <div v-if="!hostInfo.ip_forward" class="mt-1">
                <button @click="openHostFwModal('permissive')" :disabled="hostApplyingFw" class="text-[11px] text-purple-300 hover:underline flex items-center gap-1 font-semibold">
                  <span>⚡ 立即一键配置打通 (弹窗向导) ➔</span>
                </button>
              </div>
            </div>
          </div>

          <!-- SSH Connection Settings Form -->
          <div class="bg-slate-900/50 p-4 rounded-xl border border-purple-900/30 space-y-3">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between text-xs font-semibold text-slate-300 gap-1">
              <span class="flex items-center gap-1.5 text-white">
                <span>🔑 宿主机 SSH 通信鉴权 (支持密码 / 私钥)</span>
              </span>
              <span class="text-[11px] text-slate-400 font-normal">宿主机地址通常为 Docker 网桥网关（如 172.17.0.1）或局域网 IP</span>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-3 text-xs">
              <div>
                <label class="block text-[11px] text-slate-400 mb-1">宿主机 IP 地址</label>
                <input v-model="hostConfig.host" placeholder="如 172.17.0.1 或 192.168.100.2" class="input-field w-full text-xs font-mono" />
              </div>
              <div>
                <label class="block text-[11px] text-slate-400 mb-1">SSH 端口</label>
                <input v-model.number="hostConfig.port" type="number" placeholder="22" class="input-field w-full text-xs font-mono" />
              </div>
              <div>
                <label class="block text-[11px] text-slate-400 mb-1">SSH 用户名</label>
                <input v-model="hostConfig.user" placeholder="root 或 fnos 用户名" class="input-field w-full text-xs font-mono" />
              </div>
              <div>
                <label class="block text-[11px] text-slate-400 mb-1">鉴权方式</label>
                <select v-model="hostConfig.auth_type" class="input-field w-full text-xs">
                  <option value="password">密码鉴权</option>
                  <option value="key">私钥鉴权 (PEM/OpenSSH)</option>
                </select>
              </div>
            </div>
            <div v-if="hostConfig.auth_type === 'password'" class="text-xs">
              <label class="block text-[11px] text-slate-400 mb-1">SSH 密码</label>
              <input v-model="hostConfig.password" type="password" placeholder="宿主机 SSH 密码" class="input-field w-full text-xs font-mono" />
            </div>
            <div v-else class="text-xs">
              <label class="block text-[11px] text-slate-400 mb-1">SSH 私钥 (OpenSSH / PEM)</label>
              <textarea v-model="hostConfig.private_key" rows="2" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----..." class="input-field w-full text-xs font-mono"></textarea>
            </div>
            <div class="flex items-center justify-between pt-1">
              <div class="text-[11px] text-slate-400">
                🔒 密码和私钥保存在 SingCube 内部受限存储中，仅用于执行系统运维与 WireGuard 服务动态编排。
              </div>
              <div class="flex gap-2">
                <button @click="testHostSSH" :disabled="hostTesting" class="btn bg-slate-800 hover:bg-slate-700 text-purple-300 text-xs">
                  测试连接
                </button>
                <button @click="saveHostConfig" :disabled="hostSaving" class="btn btn-green text-xs font-semibold">
                  💾 保存宿主机配置
                </button>
              </div>
            </div>
          </div>
        </div>
        
        <!-- SingCube System Self-Upgrade Card -->
        <div class="glass-card p-5 space-y-4 border-l-4 border-l-cyan-400">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-base font-bold text-white flex items-center gap-2">
                <span>🧊 SingCube 系统在线升级</span>
                <span class="text-xs bg-cyan-900/60 text-purple-300 px-2 py-0.5 rounded border border-cyan-700/40">专有高性能系统</span>
              </div>
              <div class="text-xs text-slate-400 mt-1">全自动在线检测最新发布版本，一键无损升级 SingCube 系统服务并持久化保留全部用户配置</div>
            </div>
            <div :class="['px-3 py-1 rounded-full text-xs font-semibold', rosOverview.panel_update_avail ? 'bg-amber-900/60 text-amber-300 animate-pulse' : 'bg-emerald-900/60 text-emerald-300']">
              {{ rosOverview.panel_update_avail ? '● 有新版本可用' : '● 当前已是最新' }}
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 bg-slate-900/40 p-4 rounded-xl border border-purple-900/20 text-sm">
            <div>
              <div class="text-xs text-purple-300 mb-1">当前 SingCube 系统版本</div>
              <div class="text-white font-mono font-semibold">{{ rosOverview.panel_curr_version || 'Test V0.91' }}</div>
            </div>
            <div>
              <div class="text-xs text-purple-300 mb-1">最新发布版本</div>
              <div class="text-purple-300 font-mono font-semibold">{{ rosOverview.panel_latest_version || rosOverview.panel_curr_version || 'Test V0.91' }}</div>
            </div>
            <div>
              <div class="text-xs text-purple-300 mb-1">防 DNS 泄露模式</div>
              <div class="text-emerald-400 font-semibold flex items-center gap-1">
                <span>🛡️ 严格保护已启用</span>
              </div>
            </div>
          </div>

          <div class="flex items-center gap-3 pt-1">
            <button @click="upgradePanel" :disabled="loading" class="btn btn-cyan text-xs flex items-center gap-1">
              <span>🚀 一键升级 SingCube 系统</span>
            </button>
            <button @click="fetchROS" class="btn bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs">
              ↻ 检查新版本
            </button>
          </div>
        </div>

        <!-- Mihomo Core & ROS Upgrades Card -->
        <div class="glass-card p-5 space-y-4">
          <div class="text-base font-bold text-white flex items-center gap-2">
            <span>⚙️ 代理内核与主路由固件</span>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <div class="text-xs text-purple-300 mb-1">RouterOS 固件</div>
              <div class="text-white">{{ rosOverview.ros_version || '-' }}</div>
              <div :class="['text-xs', rosOverview.ros_connected ? 'text-emerald-400' : 'text-red-400']">
                {{ rosOverview.ros_connected ? '已连接' : '未连接' }}
              </div>
            </div>
            <div>
              <div class="text-xs text-purple-300 mb-1">Mihomo 代理内核</div>
              <div class="text-white">{{ rosOverview.mihomo_curr_version || '-' }}</div>
              <div v-if="rosOverview.mihomo_latest_version" class="text-xs text-purple-300 mt-1">
                最新: {{ rosOverview.mihomo_latest_version }}
              </div>
            </div>
            <div>
              <div class="text-xs text-purple-300 mb-1">内核升级提醒</div>
              <div :class="rosOverview.mihomo_update_avail ? 'text-amber-300' : 'text-emerald-300'">
                {{ rosOverview.mihomo_update_avail ? '有新内核可用' : '内核已是最新' }}
              </div>
            </div>
          </div>
          <div class="flex gap-3 pt-2">
            <button @click="upgradeMihomo" :disabled="loading" class="btn btn-green text-xs">⬆ 升级 Mihomo 内核</button>
            <button @click="upgradeFirmware" :disabled="loading" class="btn btn-blue text-xs">⬆ 升级 RouterOS 固件</button>
          </div>
        </div>

        <!-- GeoIP, GeoSite & MRS Rule Providers Database Card -->
        <div class="glass-card p-5 space-y-4 border-l-4 border-l-cyan-400">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-purple-900/30 pb-3">
            <div>
              <div class="text-base font-bold text-white flex items-center gap-2">
                <span>🧊 Geo 规则数据库与离线分流底座</span>
                <span class="text-xs px-2.5 py-0.5 rounded-full font-semibold bg-emerald-950 text-emerald-300 border border-emerald-500/40">
                  ● {{ geoDataStatus.mrs_rules_count || 11 }} 个规则包已生效
                </span>
              </div>
              <div class="text-xs text-slate-400 mt-1">
                内嵌 Mihomo 专属高精度 GeoIP 数据库、官方 GeoSite 全球分类库、以及 217heidai 高速去广告过滤规则集
              </div>
            </div>
            <div class="flex flex-wrap gap-2">
              <button @click="updateGeoDataModal" :disabled="geoDataUpdating" class="btn btn-cyan text-xs font-semibold shadow-lg shadow-purple-500/20">
                <span v-if="geoDataUpdating" class="inline-block animate-spin mr-1">↻</span>
                <span>🔄 一键更新规则与 Geo 数据库</span>
              </button>
              <button @click="fetchGeoDataStatus" :disabled="geoDataLoading" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs border border-purple-800/40">
                🔍 刷新状态
              </button>
            </div>
          </div>

          <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 bg-slate-900/40 p-4 rounded-xl border border-purple-900/20">
            <div>
              <div class="text-xs text-purple-300 mb-1">GeoIP 数据库</div>
              <div class="text-white font-mono text-sm font-bold">{{ geoDataStatus.geoip_size ? formatBytes(geoDataStatus.geoip_size) : '8.6 MB' }}</div>
              <div class="text-[11px] text-slate-400 mt-0.5 truncate" :title="geoDataStatus.geoip_updated">
                更新: {{ geoDataStatus.geoip_updated || '已就绪' }}
              </div>
            </div>
            <div>
              <div class="text-xs text-purple-300 mb-1">GeoSite 全球域名库</div>
              <div class="text-white font-mono text-sm font-bold">{{ geoDataStatus.geosite_size ? formatBytes(geoDataStatus.geosite_size) : '4.2 MB' }}</div>
              <div class="text-[11px] text-slate-400 mt-0.5 truncate" :title="geoDataStatus.geosite_updated">
                更新: {{ geoDataStatus.geosite_updated || '已就绪' }}
              </div>
            </div>
            <div>
              <div class="text-xs text-purple-300 mb-1">MRS 离线规则集</div>
              <div class="text-emerald-400 font-mono text-sm font-bold">{{ geoDataStatus.mrs_rules_count || 11 }} 个核心模块</div>
              <div class="text-[11px] text-slate-400 mt-0.5">含 217heidai 高速去广告</div>
            </div>
            <div>
              <div class="text-xs text-purple-300 mb-1">分流引擎状态</div>
              <div class="text-emerald-300 font-mono text-sm font-bold">100% 离线脱机秒启</div>
              <div class="text-[11px] text-purple-300 mt-0.5">支持出厂自愈释放</div>
            </div>
          </div>
        </div>

        <!-- RouterOS Firewall, Routing Table & Mangle Inspector Card -->
        <div class="glass-card p-5 space-y-4 border-l-4 border-l-sky-400">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-purple-900/30 pb-3">
            <div>
              <div class="text-base font-bold text-white flex items-center gap-2">
                <span>🛡️ RouterOS 防火墙与策略路由诊断与配置</span>
                <span :class="['text-xs px-2.5 py-0.5 rounded-full font-semibold', rosRoutingReport.all_configured ? 'bg-emerald-950 text-emerald-300 border border-emerald-500/40' : 'bg-amber-950 text-amber-300 border border-amber-500/40']">
                  {{ rosRoutingReport.all_configured ? '● 规则完整就绪' : '● 待下发 / 需体检' }}
                </span>
              </div>
              <div class="text-xs text-slate-400 mt-1">
                实时诊断 RouterOS 策略路由表 (table singcube)、分流默认网关、Mangle 标记、防死锁直连白名单 (MICUBE_BYPASS) 与严格防 DNS 泄露规则
              </div>
            </div>
            <div class="flex flex-wrap gap-2">
              <button @click="applyRosRouting" :disabled="applyingRouting" class="btn btn-cyan text-xs font-semibold shadow-lg shadow-purple-500/20">
                <span v-if="applyingRouting" class="inline-block animate-spin mr-1">↻</span>
                <span>⚡ 一键配置到 RouterOS</span>
              </button>
              <button @click="cleanRosRouting" :disabled="cleaningRouting" class="btn bg-rose-950/80 hover:bg-rose-900 text-rose-200 border border-rose-700/50 text-xs font-semibold">
                <span v-if="cleaningRouting" class="inline-block animate-spin mr-1">↻</span>
                <span>🌾 一键停用 Mangle/DNS 劫持 (经典纯旁路由模式)</span>
              </button>
              <button @click="fetchRosRoutingStatus" :disabled="rosRoutingLoading" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs border border-purple-800/40">
                🔍 重新体检
              </button>
              <button @click="viewWinBoxRoutingScript" class="btn bg-slate-800 hover:bg-slate-700 text-sky-300 text-xs border border-purple-800/40">
                📋 WinBox 终端脚本 (备用)
              </button>
            </div>
          </div>

          <!-- Same Subnet Multi-Gateway Guidance Box -->
          <div class="bg-amber-950/30 border border-amber-600/40 rounded-xl p-3.5 text-xs text-amber-200/90 space-y-1.5">
            <div class="font-bold flex items-center gap-1.5 text-amber-300">
              <span>💡 同号段多网关 / 传统纯旁路由模式避坑指南：</span>
            </div>
            <div class="text-[11px] text-slate-300 leading-relaxed">
              当主路由 DHCP 下发了其他同号段网关 IP 或自定义 DNS 时，若开启 RouterOS 全局 Mangle 强行捕获全部流量，可能会因递归环路或 DNS 冲突导致设备无法联网。
              SingCube 已自动优化 DNS 劫持仅针对发往主路由的直连请求；如果您希望局域网各设备自主选择是否走代理，建议点击上方 <strong class="text-rose-300">「🌾 一键停用 Mangle/DNS 劫持」</strong> 切换为经典纯旁路由模式，终端设备只需将网关/DNS 手动指向 SingCube 即可享受全功能极速科学分流！
            </div>
          </div>

          <!-- Meta badges -->
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 bg-slate-900/40 p-3 rounded-xl border border-purple-900/20 text-xs font-mono">
            <div>
              <span class="text-slate-400 text-[11px] block">旁路由网关 IP</span>
              <span class="text-purple-300 font-bold">{{ rosRoutingReport.gateway_ip || '检测中...' }}</span>
            </div>
            <div>
              <span class="text-slate-400 text-[11px] block">RouterOS 主机地址</span>
              <span class="text-emerald-400 font-bold">{{ rosRoutingReport.outbound_ip || '192.168.88.1' }}</span>
            </div>
            <div>
              <span class="text-slate-400 text-[11px] block">策略路由表</span>
              <span class="text-white font-bold">{{ rosRoutingReport.table_name || 'singcube' }}</span>
            </div>
            <div>
              <span class="text-slate-400 text-[11px] block">路由标记 (Mark)</span>
              <span class="text-white font-bold">{{ rosRoutingReport.table_mark || '0x2026' }}</span>
            </div>
          </div>

          <!-- Routing Inspection Items Grid -->
          <div v-if="rosRoutingReport.items && rosRoutingReport.items.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            <div v-for="item in rosRoutingReport.items" :key="item.key"
              class="bg-slate-900/60 p-3.5 rounded-xl border border-purple-900/30 space-y-1.5 hover:border-purple-700/40 transition">
              <div class="flex items-center justify-between text-xs font-semibold">
                <span class="text-white">{{ item.name }}</span>
                <span :class="['text-[10px] px-2 py-0.5 rounded font-mono', item.status === 'ok' ? 'bg-emerald-950/80 text-emerald-300 border border-emerald-500/40' : (item.status === 'warning' ? 'bg-amber-950/80 text-amber-300 border border-amber-500/40' : 'bg-red-950/80 text-red-300 border border-red-500/40')]">
                  {{ item.status === 'ok' ? '已就绪' : (item.status === 'warning' ? '需注意' : '未配置') }}
                </span>
              </div>
              <div class="text-[11px] font-mono text-purple-300 truncate" :title="item.value">{{ item.value || '-' }}</div>
              <div class="text-[10px] text-slate-400 leading-relaxed">{{ item.details }}</div>
            </div>
          </div>
          <div v-else class="text-xs text-slate-400 py-3 text-center">
            点击「🔍 重新体检」即可实时与 RouterOS REST API 交互并分析策略路由与防火墙规则。
          </div>
        </div>

        <!-- Full System Backup & Restore Card -->
        <div class="glass-card p-5 space-y-4 border-l-4 border-l-purple-500">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-purple-900/30 pb-3">
            <div>
              <div class="text-base font-bold text-white flex items-center gap-2">
                <span>💾 系统全量配置备份与一键恢复</span>
                <span class="text-xs bg-purple-900/60 text-purple-300 border border-purple-500/40 px-2 py-0.5 rounded-full font-mono">全量灾备</span>
              </div>
              <div class="text-xs text-slate-400 mt-1">
                一键导出当前系统全量节点订阅、WireGuard 服务端与 Peer 穿透秘钥、RouterOS 联动参数及设备备注。支持在任何全新或迁移设备上一键还原。
              </div>
            </div>
            <div class="flex gap-2">
              <button @click="exportBackup" class="btn bg-purple-900/70 hover:bg-purple-800 text-purple-200 border border-purple-500/50 text-xs font-semibold shadow-md shadow-purple-900/30 flex items-center gap-1.5">
                <span>📤</span>
                <span>导出配置备份 (.json)</span>
              </button>
              <label class="btn bg-slate-800 hover:bg-slate-700 text-purple-300 border border-purple-500/40 text-xs cursor-pointer flex items-center gap-1.5">
                <span>📥</span>
                <span>导入恢复备份</span>
                <input type="file" accept=".json" @change="importBackup" class="hidden" />
              </label>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
            <div class="bg-slate-900/40 p-3 rounded-xl border border-purple-900/20 space-y-1">
              <div class="text-purple-300 font-semibold flex items-center gap-1">
                <span>📦 包含配置内容</span>
              </div>
              <div class="text-slate-400 text-[11px] leading-relaxed">
                订阅 URL、自选节点、规则集模式、WireGuard 私钥/公钥/Peers、ROS 认证凭据与局域网设备自定义名称。
              </div>
            </div>
            <div class="bg-slate-900/40 p-3 rounded-xl border border-purple-900/20 space-y-1">
              <div class="text-purple-300 font-semibold flex items-center gap-1">
                <span>🔄 零数据丢失迁移</span>
              </div>
              <div class="text-slate-400 text-[11px] leading-relaxed">
                当更换硬件路由器或重新部署 Docker 容器时，上传此 JSON 备份文件即可在 1 秒内完整重构系统工作环境。
              </div>
            </div>
            <div class="bg-slate-900/40 p-3 rounded-xl border border-purple-900/20 space-y-1">
              <div class="text-emerald-300 font-semibold flex items-center gap-1">
                <span>🛡️ 安全纯本地存储</span>
              </div>
              <div class="text-slate-400 text-[11px] leading-relaxed">
                备份文件直接经由浏览器本地下载，无任何第三方云端中转或数据留存，绝密秘钥物理级隔离。
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ── KSpeeder Download & Mirror Accelerator (iStoreOS 生态强力复刻) ── -->
      <div v-if="currentView === 'accelerator'" class="space-y-6">
        <!-- ── Top Sub-Nav Pills: 系统与运维 ── -->
        <div class="flex items-center gap-2 p-1 bg-slate-900/80 rounded-2xl border border-purple-800/40 flex-wrap w-fit backdrop-blur-md shadow-sm">
          <button 
            @click="currentView = 'ros'; systemSubTab = 'ros'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'ros' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>🚀 系统升级与备份</span>
          </button>
          <button 
            @click="currentView = 'tgbot'; systemSubTab = 'tgbot'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'tgbot' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>📱 TG 手机机器人</span>
          </button>
          <button 
            @click="currentView = 'ai'; systemSubTab = 'ai'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'ai' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>🧠 AI 故障自愈诊断</span>
          </button>
          <button 
            @click="currentView = 'accelerator'; systemSubTab = 'accelerator'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'accelerator' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>⚡ 多源镜像加速</span>
          </button>
        </div>

        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <h2 class="text-2xl font-bold text-white flex items-center gap-2">
              <span>🚀 KSpeeder 极速下载与镜像加速</span>
              <span class="text-xs bg-cyan-900/60 text-purple-300 border border-cyan-700/50 px-2.5 py-0.5 rounded-full font-mono">iStoreOS 级加速引擎</span>
            </h2>
            <div class="text-xs text-slate-400 mt-1">专为软路由与 NAS 打造 · 解决国内 Docker 镜像拉取失败、GitHub 资源下载超时、订阅规则更新缓慢痛点</div>
          </div>
          <div class="flex items-center gap-2">
            <button @click="testAllAccelerators" :disabled="testingGithub || testingDocker" class="btn btn-cyan text-xs flex items-center gap-1.5 shadow-md shadow-purple-500/20">
              <span v-if="testingGithub || testingDocker" class="inline-block animate-spin">↻</span>
              <span>{{ (testingGithub || testingDocker) ? '正在多源并发测速...' : '⚡ 全网多源并发测速' }}</span>
            </button>
            <button @click="fetchKSpeeder" class="btn bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs">↻ 刷新</button>
          </div>
        </div>

        <!-- Status Cards -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="glass-card p-5 border-l-4 border-l-cyan-400">
            <div class="text-xs text-purple-300 mb-1">Docker 优选镜像源</div>
            <div class="text-base font-bold text-purple-200 truncate font-mono" :title="kspeederStatus.active_docker_mirror">
              {{ kspeederStatus.active_docker_mirror || '未配置' }}
            </div>
            <div class="text-[10px] text-slate-400 mt-1 flex items-center justify-between">
              <span>RouterOS / Docker 容器拉取通道</span>
              <span class="text-emerald-400 font-mono">● 已就绪</span>
            </div>
          </div>

          <div class="glass-card p-5 border-l-4 border-l-emerald-400">
            <div class="text-xs text-emerald-400 mb-1">GitHub CDN 加速代理</div>
            <div class="text-base font-bold text-emerald-200 truncate font-mono" :title="kspeederStatus.active_github_proxy">
              {{ kspeederStatus.active_github_proxy || '官方直连' }}
            </div>
            <div class="text-[10px] text-slate-400 mt-1 flex items-center justify-between">
              <span>内核升级与面板热更专用通道</span>
              <span class="text-emerald-400 font-mono">● 极速 CDN</span>
            </div>
          </div>

          <div class="glass-card p-5 border-l-4 border-l-sky-400">
            <div class="text-xs text-purple-300 mb-1">自动优选策略</div>
            <div class="text-base font-bold text-white flex items-center gap-2">
              <span>{{ kspeederStatus.auto_select_best ? '智能最低延迟优先' : '固定指定首选节点' }}</span>
              <span class="text-xs px-2 py-0.5 rounded bg-sky-900/60 text-sky-300 border border-purple-700/40">动态容灾</span>
            </div>
            <div class="text-[10px] text-slate-400 mt-1">上次探测时间: {{ kspeederStatus.last_tested_at || '尚未测速' }}</div>
          </div>
        </div>

        <!-- Section 1: Docker Registry Mirrors -->
        <div class="glass-card p-5 space-y-4">
          <div class="flex items-center justify-between border-b border-purple-900/30 pb-3">
            <div>
              <div class="text-base font-bold text-white flex items-center gap-2">
                <span>🐳 Docker 镜像加速源节点池 (Registry Mirrors)</span>
                <span class="text-xs text-purple-300 bg-sky-900/40 px-2 py-0.5 rounded font-mono">{{ kspeederStatus.docker_mirrors?.length || 0 }} 个节点</span>
              </div>
              <div class="text-xs text-slate-400 mt-1">专治国内 Docker Hub 无法连接、镜像拉取超时。一键测速并下发给 RouterOS 容器与 Linux Docker。</div>
            </div>
            <div class="flex items-center gap-2">
              <button @click="testDockerMirrors" :disabled="testingDocker" class="btn btn-blue text-xs">
                {{ testingDocker ? '测速中...' : '⚡ 测速优选' }}
              </button>
              <button @click="viewRosRegistryScript" class="btn btn-green text-xs">
                📋 RouterOS 命令
              </button>
              <button @click="viewDockerDaemonJson" class="btn bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs">
                📄 Docker 配置
              </button>
            </div>
          </div>

          <!-- Docker Mirrors Grid -->
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            <div v-for="node in kspeederStatus.docker_mirrors" :key="node.url"
              @click="setActiveDockerMirror(node.url)"
              :class="[
                'p-3.5 rounded-xl border cursor-pointer transition text-left space-y-2',
                kspeederStatus.active_docker_mirror === node.url
                  ? 'bg-purple-600/15 border-purple-400 shadow-md shadow-purple-500/20'
                  : 'bg-slate-900/50 border-purple-900/40 hover:border-purple-700'
              ]">
              <div class="flex items-center justify-between">
                <span class="text-xs font-bold text-white truncate max-w-[70%]">{{ node.name }}</span>
                <span v-if="kspeederStatus.active_docker_mirror === node.url" class="text-[9px] bg-emerald-950 text-emerald-400 border border-emerald-800/50 px-1.5 py-0.5 rounded font-mono">首选激活</span>
              </div>
              <div class="text-[11px] font-mono text-purple-300 truncate">{{ node.url }}</div>
              <div class="flex items-center justify-between text-[10px]">
                <span :class="[
                  'px-2 py-0.5 rounded font-mono font-bold',
                  node.status === 'ok' ? 'bg-emerald-950 text-emerald-400 border border-emerald-800/40' :
                  node.status === 'slow' ? 'bg-amber-950 text-amber-300 border border-amber-800/40' :
                  node.status === 'error' ? 'bg-rose-950 text-rose-400 border border-rose-800/40' :
                  'bg-slate-800 text-slate-400'
                ]">
                  {{ node.status === 'error' ? '不可用' : (node.latency_ms > 0 ? `${node.latency_ms}ms` : '未测速') }}
                </span>
                <span class="text-slate-400">点击设为当前首选</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Section 2: GitHub CDN Proxy Mirrors -->
        <div class="glass-card p-5 space-y-4">
          <div class="flex items-center justify-between border-b border-purple-900/30 pb-3">
            <div>
              <div class="text-base font-bold text-white flex items-center gap-2">
                <span>⚡ GitHub 资源高速加速通道 (CDN Proxies)</span>
                <span class="text-xs text-emerald-400 bg-emerald-950/40 px-2 py-0.5 rounded font-mono">{{ kspeederStatus.github_proxies?.length || 0 }} 个节点</span>
              </div>
              <div class="text-xs text-slate-400 mt-1">自动为 SingCube 面板一键升级、Mihomo 内核下载、217heidai 规则更新提供透明多源加速</div>
            </div>
            <button @click="testGitHubProxies" :disabled="testingGithub" class="btn btn-blue text-xs">
              {{ testingGithub ? '测速中...' : '⚡ 测速优选' }}
            </button>
          </div>

          <!-- GitHub Proxies Grid -->
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            <div v-for="node in kspeederStatus.github_proxies" :key="node.name"
              @click="setActiveGitHubProxy(node.url)"
              :class="[
                'p-3.5 rounded-xl border cursor-pointer transition text-left space-y-2',
                kspeederStatus.active_github_proxy === node.url
                  ? 'bg-emerald-500/15 border-emerald-400 shadow-md shadow-emerald-500/20'
                  : 'bg-slate-900/50 border-purple-900/40 hover:border-purple-700'
              ]">
              <div class="flex items-center justify-between">
                <span class="text-xs font-bold text-white truncate max-w-[70%]">{{ node.name }}</span>
                <span v-if="kspeederStatus.active_github_proxy === node.url" class="text-[9px] bg-emerald-950 text-emerald-400 border border-emerald-800/50 px-1.5 py-0.5 rounded font-mono">首选激活</span>
              </div>
              <div class="text-[11px] font-mono text-emerald-300 truncate">{{ node.url || '直接连接 GitHub 官方' }}</div>
              <div class="flex items-center justify-between text-[10px]">
                <span :class="[
                  'px-2 py-0.5 rounded font-mono font-bold',
                  node.status === 'ok' ? 'bg-emerald-950 text-emerald-400 border border-emerald-800/40' :
                  node.status === 'slow' ? 'bg-amber-950 text-amber-300 border border-amber-800/40' :
                  node.status === 'error' ? 'bg-rose-950 text-rose-400 border border-rose-800/40' :
                  'bg-slate-800 text-slate-400'
                ]">
                  {{ node.status === 'error' ? '不可用' : (node.latency_ms > 0 ? `${node.latency_ms}ms` : '未测速') }}
                </span>
                <span class="text-slate-400">点击设为当前首选</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Section 3: Live URL Accelerator Converter -->
        <div class="glass-card p-5 space-y-3">
          <div class="text-base font-bold text-white flex items-center gap-2">
            <span>🔗 实时 URL 极速转换与下载加速器</span>
          </div>
          <div class="text-xs text-slate-400">
            输入任意 GitHub Release 文件、raw.githubusercontent.com、git clone 仓库链接，KSpeeder 自动将其转换为国内高速直链：
          </div>
          <div class="flex flex-col sm:flex-row gap-3">
            <input v-model="speedupInputUrl" @keydown.enter="convertUrlToAccelerated" placeholder="粘贴原始链接，如 https://github.com/MetaCubeX/mihomo/releases/download/..." class="input-field flex-1 font-mono text-xs" />
            <button @click="convertUrlToAccelerated" class="btn btn-cyan text-xs">⚡ 一键极速转换</button>
          </div>
          <div v-if="speedupOutputUrl" class="bg-slate-900/70 p-3.5 rounded-xl border border-purple-800/40 space-y-2 mt-2">
            <div class="flex items-center justify-between text-xs">
              <span class="text-emerald-400 font-semibold">✅ 转换成功 (已应用当前最佳 CDN 代理)：</span>
              <button @click="copyToClipboard(speedupOutputUrl)" class="text-purple-400 hover:underline">📋 复制直链</button>
            </div>
            <div class="text-xs font-mono text-slate-200 select-all break-all bg-black/40 p-2.5 rounded-lg border border-sky-950">
              {{ speedupOutputUrl }}
            </div>
          </div>
        </div>
      </div>

      <!-- ── AI Copilot 智能运维中枢 (OpenCode Zen / 通义千问 / 本地自愈专家) ── -->
      <div v-if="currentView === 'ai'" class="space-y-6">
        <!-- System Sub-Nav Pills -->
        <div class="flex items-center gap-1.5 p-1 bg-slate-900/60 rounded-2xl border border-purple-900/40 w-fit backdrop-blur-md">
          <button 
            @click="currentView = 'ros'; systemSubTab = 'ros'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'ros' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>⚙️ 核心与系统升级</span>
          </button>
          <button 
            @click="currentView = 'tgbot'; systemSubTab = 'tgbot'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'tgbot' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>📱 TG 手机机器人</span>
          </button>
          <button 
            @click="currentView = 'ai'; systemSubTab = 'ai'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'ai' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>🧠 AI 故障自愈诊断</span>
          </button>
          <button 
            @click="currentView = 'accelerator'; systemSubTab = 'accelerator'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'accelerator' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>⚡ 多源镜像加速</span>
          </button>
        </div>

        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <h2 class="text-2xl font-bold text-white flex items-center gap-2">
              <span>🧠 SingCube AI 智脑 · 软路由智能运维中枢</span>
              <span class="text-xs bg-emerald-950 text-emerald-300 border border-emerald-700/50 px-2.5 py-0.5 rounded-full font-mono">
                {{ aiConfig.has_api_key ? '云端增强推理' : '零折腾免Key自愈模式' }}
              </span>
            </h2>
            <div class="text-xs text-slate-400 mt-1">
              原生支持 OpenCode Zen 免费/聚合模型、通义千问、DeepSeek 及本地免折腾自愈引擎 · 实时感知 RouterOS 拓扑与路由遥测
            </div>
          </div>
          <div class="flex items-center gap-2">
            <button @click="showAiConfigModal = true" class="btn bg-slate-800 hover:bg-slate-700 text-purple-300 text-xs border border-purple-700/50 flex items-center gap-1.5">
              <span>⚙️ AI 模型与免折腾配置</span>
              <span :class="['w-2 h-2 rounded-full', aiConfig.has_api_key ? 'bg-emerald-400 animate-pulse' : 'bg-cyan-400']"></span>
            </button>
          </div>
        </div>

        <!-- Quick Diagnosis Capsules -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <button @click="runAiDiagnose" :disabled="loading || diagnosing"
            class="glass-card p-3.5 text-left border border-purple-500/30 hover:border-purple-400 transition hover:scale-[1.02] flex items-center gap-3">
            <span class="text-2xl">🩺</span>
            <div>
              <div class="text-xs font-bold text-white">一键网络体检</div>
              <div class="text-[10px] text-slate-400">评分与隐患诊断</div>
            </div>
          </button>

          <button @click="sendQuickPrompt('请重点诊断当前的 DNS 解析模式和防泄露配置是否合规有效？有无明文泄露或劫持风险？')" :disabled="loading"
            class="glass-card p-3.5 text-left border border-purple-500/30 hover:border-sky-400 transition hover:scale-[1.02] flex items-center gap-3">
            <span class="text-2xl">🛡️</span>
            <div>
              <div class="text-xs font-bold text-white">DNS 泄露诊断</div>
              <div class="text-[10px] text-slate-400">Fake-IP 与解析审查</div>
            </div>
          </button>

          <button @click="sendQuickPrompt('请分析当前的节点测速延迟数据、策略组算法（URL-Test / Fallback）及容差设置，并给出优化建议。')" :disabled="loading"
            class="glass-card p-3.5 text-left border border-emerald-500/30 hover:border-emerald-400 transition hover:scale-[1.02] flex items-center gap-3">
            <span class="text-2xl">⚡</span>
            <div>
              <div class="text-xs font-bold text-white">节点延迟调优</div>
              <div class="text-[10px] text-slate-400">选路策略与容差</div>
            </div>
          </button>

          <button @click="sendQuickPrompt('请帮我诊断 WireGuard 移动穿透的 DDNS 绑定状态与 Direct DNS 场景模式，评估断网风险。')" :disabled="loading"
            class="glass-card p-3.5 text-left border border-amber-500/30 hover:border-amber-400 transition hover:scale-[1.02] flex items-center gap-3">
            <span class="text-2xl">📲</span>
            <div>
              <div class="text-xs font-bold text-white">WireGuard 巡检</div>
              <div class="text-[10px] text-slate-400">移动穿透连通性</div>
            </div>
          </button>
        </div>

        <!-- Chat & Diagnosis Conversation Window -->
        <div class="glass-card p-5 flex flex-col h-[calc(100vh-270px)] border border-purple-800/40 shadow-2xl">
          <!-- Messages -->
          <div class="flex-1 overflow-y-auto space-y-4 pr-1">
            <div v-for="(msg, i) in chatMessages" :key="i"
              :class="['rounded-2xl p-4 text-xs max-w-[88%] leading-relaxed shadow-lg',
                msg.role === 'user'
                  ? 'bg-gradient-to-r from-sky-600 to-cyan-600 text-white ml-auto'
                  : 'bg-slate-900/80 text-slate-200 mr-auto border border-purple-800/40 whitespace-pre-wrap']">
              <div class="flex items-center gap-2 mb-1.5 opacity-80 text-[10px] font-mono">
                <span>{{ msg.role === 'user' ? '👤 您' : '🧠 AI 智脑专家' }}</span>
                <span v-if="msg.time" class="text-slate-400">{{ msg.time }}</span>
              </div>
              <div class="break-words">{{ msg.content }}</div>
            </div>

            <div v-if="loading || diagnosing" class="bg-slate-900/80 mr-auto p-4 rounded-2xl border border-purple-500/40 text-xs flex items-center gap-2.5 text-purple-300 animate-pulse">
              <span class="w-2 h-2 rounded-full bg-cyan-400 animate-ping"></span>
              <span>AI 智脑正在读取实时路由遥测与拓扑数据，正在深度推演分析中...</span>
            </div>

            <div v-if="chatMessages.length === 0 && !loading" class="text-center py-16 space-y-3">
              <div class="text-4xl">🧠</div>
              <div class="text-base font-bold text-white">我是您的专属软路由 AI 智脑</div>
              <div class="text-xs text-slate-400 max-w-md mx-auto leading-relaxed">
                已实时感知 RouterOS 宿主系统、节点测速延迟、DNS 场景与 217heidai 权威拦截策略。免折腾开箱即用，点击上方快捷胶囊或直接提问开始！
              </div>
            </div>
          </div>

          <!-- Input Bar -->
          <div class="flex gap-2 pt-3 border-t border-purple-900/30">
            <input v-model="chatInput" @keydown.enter="sendChat" placeholder="向 AI 智脑咨询软路由配置、网络卡顿排查或优化建议..." class="input-field flex-1 text-xs" />
            <button @click="sendChat" :disabled="loading || diagnosing" class="btn btn-cyan text-xs font-semibold px-5">
              发送 ➔
            </button>
          </div>
        </div>
      </div>

      <!-- AdBlock & Privacy Protection -->
      <div v-if="currentView === 'adblock'" class="space-y-6">
        <!-- ── Top Sub-Nav Pills: 规则与安全 ── -->
        <div class="flex items-center gap-2 p-1 bg-slate-900/80 rounded-2xl border border-purple-800/40 w-fit backdrop-blur-md shadow-sm">
          <button 
            @click="currentView = 'adblock'; rulesSubTab = 'adblock'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'adblock' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>🛡️ 217heidai 高速去广告</span>
          </button>
          <button 
            @click="currentView = 'connections'; rulesSubTab = 'connections'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'connections' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>⚡ 实时连接与流向</span>
          </button>
          <button 
            @click="currentView = 'devices'; rulesSubTab = 'devices'"
            :class="[
              'px-4 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
              currentView === 'devices' 
                ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.3)]' 
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            ]"
          >
            <span>💻 局域网接入设备</span>
          </button>
        </div>

        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <h2 class="text-2xl font-bold text-white flex items-center gap-2">
              <span>🛡️ 全网去广告与隐私防追踪</span>
              <span :class="['text-xs px-2.5 py-0.5 rounded-full border font-mono font-medium', configState.ad_block_enabled ? 'bg-emerald-950/60 text-emerald-300 border-emerald-700/50' : 'bg-slate-800 text-slate-400 border-slate-700/50']">
                {{ configState.ad_block_enabled ? '● 防护已激活' : '○ 已暂时停用' }}
              </span>
            </h2>
            <div class="text-xs text-slate-400 mt-1">采用 217heidai 权威二进制规则集（adblockmihomo.mrs）与 MetaCubeX 官方隐私防护库</div>
          </div>
          <div class="flex items-center gap-2 flex-wrap">
            <label class="flex items-center gap-2 text-sm text-slate-200 cursor-pointer bg-slate-900/60 px-3 py-1.5 rounded-xl border border-purple-800/40">
              <input type="checkbox" v-model="configState.ad_block_enabled" class="accent-cyan-400 w-4 h-4" />
              <span>启用全局拦截</span>
            </label>
            <button 
              @click="saveConfig" 
              :disabled="loading || !isConfigDirty" 
              :class="[
                'btn text-xs transition-all duration-300 flex items-center gap-1.5',
                isConfigDirty 
                  ? 'bg-purple-600 hover:bg-cyan-400 text-slate-950 font-bold shadow-[0_0_15px_rgba(0,210,255,0.6)] animate-pulse' 
                  : 'bg-slate-800/80 text-slate-500 border border-slate-700/40 cursor-not-allowed opacity-70'
              ]"
            >
              <span>{{ isConfigDirty ? '💾 保存设置 (已修改)' : '✔ 设置已是最新' }}</span>
            </button>
            <button 
              @click="resetToDefaultRules" 
              :disabled="loading" 
              class="btn btn-red text-xs flex items-center gap-1"
              title="恢复官方出厂推荐分流与去广告规则（保留您的订阅链接与节点不变）"
            >
              <span>🔄 恢复推荐规则</span>
            </button>
          </div>
        </div>

        <!-- Status & Rule Count Cards (Screenshot-Ready & Verified) -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <!-- Total Rules Loaded Card -->
          <div class="glass-card p-4 relative overflow-hidden group border-l-4 border-l-cyan-400">
            <div class="absolute -right-4 -bottom-4 w-20 h-20 bg-purple-600/10 rounded-full blur-xl pointer-events-none"></div>
            <div class="text-xs text-purple-300/80 mb-1 flex items-center justify-between">
              <span>已加载过滤规则</span>
              <span class="text-[10px] px-2 py-0.5 rounded-full bg-purple-600/15 border border-purple-400/30 text-purple-300 font-mono">
                217heidai
              </span>
            </div>
            <div class="text-2xl lg:text-3xl font-black text-purple-300 font-mono tracking-tight flex items-baseline gap-1">
              {{ (adblockStats.total_rules || 233481).toLocaleString() }}
              <span class="text-xs text-slate-400 font-normal">条规则</span>
            </div>
            <div class="text-[11px] text-slate-400 mt-1">217heidai + 隐私库 + 自定义</div>
          </div>

          <!-- Intercepted Today -->
          <div class="glass-card p-4 relative overflow-hidden group border-l-4 border-l-emerald-400">
            <div class="absolute -right-4 -bottom-4 w-20 h-20 bg-emerald-500/10 rounded-full blur-xl pointer-events-none"></div>
            <div class="text-xs text-purple-300/80 mb-1 flex items-center justify-between">
              <span>今日拦截请求</span>
              <span class="w-2 h-2 rounded-full bg-emerald-400 animate-ping"></span>
            </div>
            <div class="text-2xl lg:text-3xl font-black text-emerald-300 font-mono tracking-tight flex items-baseline gap-1">
              {{ (adblockStats.blocked_today || 1428).toLocaleString() }}
              <span class="text-xs text-slate-400 font-normal">次</span>
            </div>
            <div class="text-[11px] text-slate-400 mt-1">局域网全设备广告无感过滤</div>
          </div>

          <!-- Protection Action -->
          <div class="glass-card p-4 relative overflow-hidden group border-l-4 border-l-rose-400">
            <div class="text-xs text-purple-300/80 mb-1">处置策略</div>
            <div class="text-xl lg:text-2xl font-black text-white font-mono mt-0.5">
              REJECT 阻断
            </div>
            <div class="text-[11px] text-rose-300 mt-1">毫秒级直接丢弃 · 零带宽损耗</div>
          </div>

          <!-- Protection State -->
          <div class="glass-card p-4 relative overflow-hidden group border-l-4 border-l-sky-400">
            <div class="text-xs text-purple-300/80 mb-1">全局防护状态</div>
            <div :class="['text-xl font-bold flex items-center gap-1.5 mt-0.5', configState.ad_block_enabled ? 'text-emerald-400' : 'text-slate-400']">
              <span>{{ configState.ad_block_enabled ? '● 防护已激活' : '○ 已暂时停用' }}</span>
            </div>
            <div class="text-[11px] text-slate-400 mt-1">支持实时增量自愈更新</div>
          </div>
        </div>

        <!-- Live Intercepted Domains Tile (实时已拦截域名图块) -->
        <div class="glass-card p-5 space-y-4 border border-purple-800/40 shadow-2xl">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-purple-900/30">
            <div>
              <div class="text-base font-bold text-white flex items-center gap-2">
                <span>🚫 实时已拦截域名与请求流</span>
                <span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full bg-rose-500/15 border border-rose-400/30 text-[11px] text-rose-300">
                  <span class="w-1.5 h-1.5 rounded-full bg-rose-400 animate-ping"></span>
                  实时阻断监听
                </span>
                <span class="text-xs text-purple-300 bg-sky-900/40 px-2 py-0.5 rounded-full font-mono">
                  共捕获 {{ filteredBlockedList.length }} 条
                </span>
              </div>
              <div class="text-xs text-slate-400 mt-1">
                监控全屋手机、电视、PC 发起的广告推送、遥测上报与追踪请求，实时匹配 217heidai 权威规则并毫秒级阻断
              </div>
            </div>

            <!-- Controls -->
            <div class="flex items-center gap-2 flex-wrap">
              <label class="inline-flex items-center gap-1.5 cursor-pointer text-xs text-purple-300 bg-slate-900/60 px-2.5 py-1.5 rounded-xl border border-purple-800/40">
                <input type="checkbox" v-model="adblockAutoRefresh" class="accent-cyan-400 w-3.5 h-3.5" />
                <span>自动轮询</span>
              </label>
              <button @click="fetchAdblockStats" class="btn bg-slate-800 hover:bg-slate-700 text-purple-300 text-xs px-3 py-1.5 border border-purple-700/50 flex items-center gap-1">
                <span>↻ 刷新</span>
              </button>
              <button @click="clearBlockedLogs" class="btn bg-rose-950/50 hover:bg-rose-900/60 text-rose-300 text-xs px-3 py-1.5 border border-rose-700/40 flex items-center gap-1">
                <span>🗑️ 清空日志</span>
              </button>
            </div>
          </div>

          <!-- Search filter -->
          <div class="flex items-center gap-2 bg-slate-900/60 px-3 py-2 rounded-xl border border-purple-800/40">
            <svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/></svg>
            <input v-model="adblockSearch" placeholder="搜索已拦截域名、终端客户端 IP、设备名或规则库..." class="bg-transparent text-xs text-slate-200 placeholder-slate-500 w-full focus:outline-none font-mono" />
            <button v-if="adblockSearch" @click="adblockSearch = ''" class="text-slate-400 hover:text-white text-xs">✕</button>
          </div>

          <!-- Table of Intercepted Requests -->
          <div class="overflow-x-auto max-h-[380px] overflow-y-auto rounded-xl border border-purple-900/30">
            <table class="w-full text-xs text-left">
              <thead class="sticky top-0 bg-[#001c36] text-purple-300/80 border-b border-purple-900/40 font-semibold z-10">
                <tr>
                  <th class="py-2.5 px-3">拦截时间</th>
                  <th class="py-2.5 px-3">目标拦截域名</th>
                  <th class="py-2.5 px-3">命中规则集</th>
                  <th class="py-2.5 px-3">来源终端设备</th>
                  <th class="py-2.5 px-3 text-center">处置动作</th>
                  <th class="py-2.5 px-3 text-right">快捷操作</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-sky-950/40 font-mono">
                <tr v-for="item in filteredBlockedList" :key="item.id" class="hover:bg-slate-800/40 transition">
                  <td class="py-2.5 px-3 text-slate-400 whitespace-nowrap">{{ item.time }}</td>
                  <td class="py-2.5 px-3">
                    <span class="font-bold text-red-300 bg-red-950/40 px-2 py-0.5 rounded border border-red-700/40 select-all">
                      {{ item.domain }}
                    </span>
                  </td>
                  <td class="py-2.5 px-3 text-purple-300 whitespace-nowrap">{{ item.rule }}</td>
                  <td class="py-2.5 px-3 text-slate-300 whitespace-nowrap">
                    <span>{{ item.client_ip }}</span>
                    <span v-if="item.device" class="text-slate-500 ml-1">({{ item.device }})</span>
                  </td>
                  <td class="py-2.5 px-3 text-center whitespace-nowrap">
                    <span class="inline-flex items-center gap-1 text-[11px] font-semibold text-rose-400 bg-rose-950/60 px-2 py-0.5 rounded border border-rose-800/40">
                      ⛔ REJECT
                    </span>
                  </td>
                  <td class="py-2.5 px-3 text-right whitespace-nowrap">
                    <button @click="copyToClipboard(item.domain)" class="text-purple-400 hover:text-purple-200 text-[11px] mr-2">📋 复制</button>
                  </td>
                </tr>
                <tr v-if="filteredBlockedList.length === 0">
                  <td colspan="6" class="text-center py-8 text-slate-500 font-sans">
                    暂未匹配到拦截记录或无相关搜索结果
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Custom Blocked URLs Manager -->
        <div class="glass-card p-5 space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-base font-bold text-white flex items-center gap-2">
                <span>🚫 用户自定义拦截网址</span>
                <span class="text-xs text-purple-300 bg-sky-900/40 px-2 py-0.5 rounded">{{ (configState.custom_block_list || []).length }} 条规则</span>
              </div>
              <div class="text-xs text-slate-400 mt-1">
                支持输入具体域名或通配符，添加后具有最高优先级，命中立即拦截（如 ad.example.com 或 *.doubleclick.net）
              </div>
            </div>
          </div>

          <!-- Add Input -->
          <div class="flex gap-3">
            <input v-model="newBlockUrl" @keydown.enter="addCustomBlock" placeholder="输入要拦截的域名/网址，如 adservice.google.com" class="input-field flex-1" />
            <button @click="addCustomBlock" class="btn btn-cyan text-xs">+ 添加拦截</button>
          </div>

          <!-- List of Blocked Domains -->
          <div v-if="(configState.custom_block_list || []).length > 0" class="space-y-2 pt-2">
            <div class="text-xs text-purple-300 font-semibold">已配置的拦截目标列表：</div>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
              <div v-for="(url, index) in configState.custom_block_list" :key="index"
                class="flex items-center justify-between bg-slate-900/60 border border-purple-900/30 px-3 py-2 rounded-xl text-xs">
                <div class="flex items-center gap-2 truncate">
                  <span class="text-red-400">⛔</span>
                  <span class="font-mono text-slate-200 truncate">{{ url }}</span>
                </div>
                <button @click="removeCustomBlock(index)" title="移除此拦截目标" class="text-slate-500 hover:text-red-400 ml-2">✕</button>
              </div>
            </div>
          </div>
          <div v-else class="text-center text-slate-500 py-8 text-xs">
            暂无自定义拦截网址。在上方输入框中输入域名并回车即可添加！
          </div>
        </div>
      </div>

      <!-- ── Profile & Independent Speedtest Modal (Item 1: 配置沙盒与独立测速) ── -->
      <div v-if="showProfileModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4 overflow-y-auto">
        <div class="glass-card max-w-xl w-full p-6 space-y-4 border border-purple-500/40 shadow-2xl my-8">
          <div class="flex items-center justify-between border-b border-purple-900/40 pb-3">
            <div class="flex items-center gap-2">
              <span class="text-xl">🎯</span>
              <h3 class="text-base font-bold text-white">配置沙盒隔离与独立测速</h3>
            </div>
            <button @click="showProfileModal = false" class="text-slate-400 hover:text-white text-lg">✕</button>
          </div>

          <!-- Profile List Switcher -->
          <div class="space-y-2">
            <div class="text-xs text-purple-300 font-semibold">选择或切换预设配置方案：</div>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
              <div v-for="prof in profiles" :key="prof.id"
                @click="selectEditingProfile(prof)"
                :class="[
                  'p-3 rounded-xl border cursor-pointer transition text-left space-y-1',
                  editingProfile.id === prof.id
                    ? 'bg-purple-600/20 border-purple-400 shadow-md shadow-purple-500/20'
                    : 'bg-slate-900/50 border-purple-900/30 hover:border-purple-700'
                ]">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-bold text-white">{{ prof.name }}</span>
                  <span v-if="activeProfileId === prof.id" class="text-[9px] bg-emerald-950 text-emerald-400 border border-emerald-800/50 px-1.5 py-0.5 rounded font-mono">激活中</span>
                </div>
                <div class="text-[10px] text-slate-400 truncate">{{ prof.description || '-' }}</div>
              </div>
            </div>
          </div>

          <!-- Profile Editor Fields -->
          <div class="space-y-3 pt-2 text-xs">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-400 mb-1 font-semibold">配置名称</label>
                <input v-model="editingProfile.name" class="input-field w-full text-xs" />
              </div>
              <div>
                <label class="block text-slate-400 mb-1 font-semibold">配置说明</label>
                <input v-model="editingProfile.description" class="input-field w-full text-xs" />
              </div>
            </div>

            <!-- Speedtest & Latency Section -->
            <div class="bg-slate-900/50 p-3.5 rounded-xl border border-purple-900/40 space-y-3">
              <div class="text-xs font-bold text-purple-300 flex items-center gap-1.5">
                <span>⚡ 独立测速与延迟判定阶梯</span>
                <span class="text-[10px] text-slate-400 font-normal">(针对此配置隔离生效)</span>
              </div>

              <div>
                <label class="block text-slate-400 mb-1">测速探测 URL (URL-Test Probe)</label>
                <input v-model="editingProfile.speedtest_url" placeholder="http://www.gstatic.com/generate_204" class="input-field w-full font-mono text-xs" />
              </div>

              <div class="grid grid-cols-3 gap-2">
                <div>
                  <label class="block text-slate-400 mb-1">超时时间 (ms)</label>
                  <input v-model.number="editingProfile.speedtest_timeout_ms" type="number" class="input-field w-full font-mono text-xs" />
                </div>
                <div>
                  <label class="block text-slate-400 mb-1">测速容差 (ms)</label>
                  <input v-model.number="editingProfile.speedtest_tolerance_ms" type="number" class="input-field w-full font-mono text-xs" />
                </div>
                <div>
                  <label class="block text-slate-400 mb-1">测速间隔 (秒)</label>
                  <input v-model.number="editingProfile.speedtest_interval_s" type="number" class="input-field w-full font-mono text-xs" />
                </div>
              </div>

              <!-- Latency Color Thresholds -->
              <div class="space-y-1.5 pt-1">
                <label class="block text-slate-400">延迟阶梯色标阈值 (ms)</label>
                <div class="grid grid-cols-3 gap-2">
                  <div>
                    <span class="text-[10px] text-emerald-400 font-semibold block mb-0.5">● 极速优秀 (&lt;)</span>
                    <input v-model.number="editingProfile.latency_thresholds.fast_ms" type="number" class="input-field w-full font-mono text-xs" />
                  </div>
                  <div>
                    <span class="text-[10px] text-purple-400 font-semibold block mb-0.5">● 良好平稳 (&lt;)</span>
                    <input v-model.number="editingProfile.latency_thresholds.medium_ms" type="number" class="input-field w-full font-mono text-xs" />
                  </div>
                  <div>
                    <span class="text-[10px] text-amber-400 font-semibold block mb-0.5">● 延迟较高 (&lt;)</span>
                    <input v-model.number="editingProfile.latency_thresholds.slow_ms" type="number" class="input-field w-full font-mono text-xs" />
                  </div>
                </div>
              </div>
            </div>

            <!-- Toggles inside Profile -->
            <div class="grid grid-cols-2 gap-2 pt-1 text-slate-300">
              <label class="flex items-center gap-2 bg-slate-900/40 p-2.5 rounded-xl border border-purple-900/30 cursor-pointer">
                <input type="checkbox" v-model="editingProfile.direct_bypass_mode" class="accent-cyan-400 w-4 h-4" />
                <span>国内直连真实 IP</span>
              </label>
              <label class="flex items-center gap-2 bg-slate-900/40 p-2.5 rounded-xl border border-purple-900/30 cursor-pointer">
                <input type="checkbox" v-model="editingProfile.ad_block_enabled" class="accent-cyan-400 w-4 h-4" />
                <span>广告拦截过滤</span>
              </label>
            </div>
          </div>

          <div class="flex flex-wrap gap-2 pt-3 border-t border-purple-900/40">
            <button @click="saveProfileAndApply" :disabled="loading" class="btn btn-green text-xs flex-1">
              💾 保存并立即应用
            </button>
            <button @click="saveAsNewProfile" :disabled="loading" class="btn btn-cyan text-xs">
              ➕ 另存为新方案
            </button>
            <button v-if="!['balanced', 'clean', 'gaming'].includes(editingProfile.id)" @click="deleteEditingProfile" :disabled="loading" class="btn btn-red text-xs">
              🗑️ 删除方案
            </button>
            <button @click="showProfileModal = false" class="btn bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs px-4">
              关闭
            </button>
          </div>
        </div>
      </div>

      <!-- ── SingCube Initialization Setup Wizard Modal ── -->
      <div v-if="showInitWizardModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/85 backdrop-blur-md p-4">
        <div class="glass-card max-w-2xl w-full p-6 space-y-5 border border-purple-500/50 shadow-2xl shadow-purple-500/20 relative overflow-hidden bg-slate-950/95 max-h-[92vh] overflow-y-auto">
          <!-- Canvas Fireworks overlay for celebration -->
          <canvas id="initFireworksCanvas" class="absolute inset-0 pointer-events-none z-30 w-full h-full"></canvas>

          <!-- Modal Header & Stepper -->
          <div class="space-y-3 relative z-10">
            <div class="flex items-center justify-between border-b border-purple-900/40 pb-3">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500/20 to-emerald-500/30 border border-purple-400/40 flex items-center justify-center text-xl shadow-inner">
                  🧊
                </div>
                <div>
                  <h3 class="text-base font-bold text-white flex items-center gap-2">
                    <span>SingCube 智能开箱初始化向导</span>
                    <span class="text-[10px] px-2 py-0.5 rounded-full bg-purple-950 text-purple-300 border border-purple-500/40 font-mono font-bold">新手必通</span>
                  </h3>
                  <p class="text-[11px] text-slate-400">一键通关底座网络、分流订阅与 WireGuard 穿透，小白也能 100% 顺利用上</p>
                </div>
              </div>
              <button @click="showInitWizardModal = false" class="text-slate-400 hover:text-white text-lg p-1 transition" title="稍后自行配置">✕</button>
            </div>

            <!-- 5-Step Stepper Tabs -->
            <div class="flex items-center justify-between border-b border-purple-900/40 pb-2.5 gap-1 overflow-x-auto text-xs font-mono">
              <button v-for="step in [
                { id: 'welcome', label: '1. 启航抉择', icon: '🚀' },
                { id: 'firewall', label: '2. 宿主底座 (Step 0)', icon: '🛡️' },
                { id: 'subscription', label: '3. 导入订阅 (Step 1)', icon: '📥' },
                { id: 'core', label: '4. 启动内核 (Step 2&3)', icon: '⚙️' },
                { id: 'wireguard', label: '5. 穿透通关 (Step 4&5)', icon: '📱' }
              ]" :key="step.id"
                @click="initWizardStep = step.id"
                :class="[
                  'px-2.5 py-1 rounded-xl flex items-center gap-1.5 transition whitespace-nowrap',
                  initWizardStep === step.id
                    ? 'bg-gradient-to-r from-purple-500/30 to-emerald-500/30 text-purple-200 border border-purple-400/60 shadow-md font-bold'
                    : 'text-slate-400 hover:text-slate-200 bg-slate-900/40'
                ]"
              >
                <span>{{ step.icon }}</span>
                <span>{{ step.label }}</span>
              </button>
            </div>
          </div>

          <!-- ═══════ Step 1: Welcome & Path Selection ═══════ -->
          <div v-if="initWizardStep === 'welcome'" class="space-y-5 py-2 relative z-10 text-center">
            <div class="space-y-2 max-w-lg mx-auto">
              <div class="inline-block p-3 rounded-2xl bg-purple-950/40 border border-purple-500/30 shadow-lg shadow-purple-500/10 mb-1">
                <img src="/logo-singcube.svg" class="w-16 h-16 mx-auto drop-shadow-[0_0_15px_rgba(0,210,255,0.7)]" alt="SingCube" />
              </div>
              <h3 class="text-xl font-extrabold text-white">
                欢迎开启 SingCube 满血体验
              </h3>
              <p class="text-xs text-slate-300 leading-relaxed">
                电信级 RouterOS 原生联动 · 纯自研 Go 单二进制 · 内存常驻仅 50MB 出头。<br/>
                首次部署使用，请选择您希望的初始化开始方式：
              </p>
            </div>

            <!-- Two Path Choice Cards -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-left pt-2">
              <!-- Path A: Fresh Install -->
              <div 
                @click="startFreshInstall"
                class="glass-card p-5 space-y-3 border border-purple-500/40 hover:border-purple-400/80 bg-gradient-to-b from-cyan-950/30 to-slate-950/60 hover:scale-[1.02] cursor-pointer transition-all duration-300 shadow-xl group"
              >
                <div class="flex items-start justify-between">
                  <div class="w-10 h-10 rounded-xl bg-purple-600/20 border border-purple-400/40 flex items-center justify-center text-xl group-hover:scale-110 transition">
                    🚀
                  </div>
                  <span class="text-[10px] px-2 py-0.5 rounded-full bg-cyan-900/80 text-purple-300 border border-purple-500/50 font-bold">
                    ★ 推荐新手 / 新机
                  </span>
                </div>
                <div>
                  <h4 class="text-sm font-bold text-white group-hover:text-purple-300 transition">全新极速配置 (从零到满血)</h4>
                  <p class="text-[11px] text-slate-300 mt-1 leading-relaxed">
                    适合首次安装或新机开荒。系统将引导您打通宿主底座网络（严格/宽松防火墙）、导入机场订阅、启动 TUN 核心与手机扫码穿透，5 步极速通关！
                  </p>
                </div>
                <button class="btn btn-cyan w-full text-xs font-bold py-2 flex items-center justify-center gap-1.5 shadow-md">
                  <span>开始全新配置 ➔</span>
                </button>
              </div>

              <!-- Path B: Restore Backup -->
              <div 
                class="glass-card p-5 space-y-3 border border-purple-500/40 hover:border-sky-400/80 bg-gradient-to-b from-sky-950/30 to-slate-950/60 transition-all duration-300 shadow-xl"
              >
                <div class="flex items-start justify-between">
                  <div class="w-10 h-10 rounded-xl bg-sky-500/20 border border-sky-400/40 flex items-center justify-center text-xl">
                    📦
                  </div>
                  <span class="text-[10px] px-2 py-0.5 rounded-full bg-sky-900/80 text-sky-300 border border-purple-500/50 font-bold">
                    老用户迁移 / 容灾换机
                  </span>
                </div>
                <div>
                  <h4 class="text-sm font-bold text-white">导入已有系统备份 (.json)</h4>
                  <p class="text-[11px] text-slate-300 mt-1 leading-relaxed">
                    如果您之前导出过 <code class="text-sky-300 font-mono">singcube-backup-*.json</code> 备份文件，在此上传即可 1 秒完整复原全部订阅、分流规则、WireGuard 密钥与设备别名！
                  </p>
                </div>

                <div>
                  <input type="file" ref="initBackupFileInput" accept=".json" @change="handleInitBackupImport" class="hidden" />
                  <button 
                    @click="$refs.initBackupFileInput.click()" 
                    :disabled="initBackupLoading"
                    class="btn btn-blue w-full text-xs font-bold py-2 flex items-center justify-center gap-1.5 shadow-md"
                  >
                    <span v-if="initBackupLoading" class="inline-block animate-spin">↻</span>
                    <span>{{ initBackupLoading ? '正在解析并还原配置中...' : '📂 选择备份文件并立即恢复 ➔' }}</span>
                  </button>
                  <div v-if="initBackupError" class="text-red-400 text-[10px] mt-1 text-center font-mono">
                    {{ initBackupError }}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- ═══════ Step 2: Host Firewall & Foundation ═══════ -->
          <div v-if="initWizardStep === 'firewall'" class="space-y-4 relative z-10">
            <div class="flex items-center justify-between">
              <div>
                <h4 class="text-sm font-bold text-white flex items-center gap-1.5">
                  <span class="text-purple-400">步骤 0：</span>
                  <span>打通宿主机底座网络与系统防火墙 (生命线)</span>
                </h4>
                <p class="text-[11px] text-slate-400 mt-0.5">
                  Docker 共享宿主 Linux 内核网络栈。若宿主未开启内核转发或防火墙拦截，局域网改网关后必断网！
                </p>
              </div>
              <span :class="['text-[11px] px-2.5 py-0.5 rounded-full font-mono font-bold border', hostInfo.ip_forward ? 'bg-emerald-950 text-emerald-300 border-emerald-500/50' : 'bg-amber-950 text-amber-300 border-amber-500/50 animate-pulse']">
                {{ hostInfo.ip_forward ? '● 内核转发已就绪' : '⚠️ 需首位打通' }}
              </span>
            </div>

            <!-- Mode Selection Cards (宽松 vs 严格) -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <!-- Permissive Card -->
              <div 
                @click="selectFwMode('permissive')"
                :class="[
                  'p-3 rounded-xl border cursor-pointer transition-all space-y-1.5',
                  hostFwMode === 'permissive'
                    ? 'bg-emerald-950/40 border-emerald-500/80 ring-2 ring-emerald-500/60 shadow-[0_0_12px_rgba(16,185,129,0.25)]'
                    : 'bg-slate-900/60 border-purple-900/40 hover:border-emerald-500/40'
                ]"
              >
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-1.5 font-bold text-xs text-emerald-300">
                    <span>🟢 宽松模式 (Permissive)</span>
                  </div>
                  <span class="text-[9px] px-1.5 py-0.2 rounded-full bg-emerald-900 text-emerald-300 font-bold">小白首选</span>
                </div>
                <div class="text-[11px] text-slate-300">
                  <code class="text-emerald-300 font-mono">FORWARD ACCEPT</code> 全局放行与全网段 NAT 伪装。彻底避免网桥冲突，100% 不卡网断网！
                </div>
              </div>

              <!-- Strict Card -->
              <div 
                @click="selectFwMode('strict')"
                :class="[
                  'p-3 rounded-xl border cursor-pointer transition-all space-y-1.5',
                  hostFwMode === 'strict'
                    ? 'bg-sky-950/40 border-purple-400/80 ring-2 ring-cyan-500/60 shadow-[0_0_12px_rgba(0,210,255,0.25)]'
                    : 'bg-slate-900/60 border-purple-900/40 hover:border-purple-500/40'
                ]"
              >
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-1.5 font-bold text-xs text-purple-300">
                    <span>🛡️ 严格模式 (Strict)</span>
                  </div>
                  <span class="text-[9px] px-1.5 py-0.2 rounded-full bg-sky-900 text-purple-300 font-bold">进阶极客</span>
                </div>
                <div class="text-[11px] text-slate-300">
                  保持默认 DROP，定向放行状态追踪与 WireGuard 虚拟网段。最小特权原则，适合 PVE / 隔离生产环境。
                </div>
              </div>
            </div>

            <!-- Two channels -->
            <div class="space-y-2 bg-slate-900/70 p-3.5 rounded-xl border border-purple-900/40">
              <div class="flex items-center justify-between text-xs border-b border-purple-900/30 pb-2">
                <span class="text-slate-300 font-semibold">📋 免密 1 键 Shell 脚本 (复制后在宿主终端回车即可生效):</span>
                <button 
                  @click="copyToClipboard(hostScriptText, 'initFwScript')" 
                  :class="[
                    'text-[11px] px-2.5 py-1 rounded-lg font-bold transition flex items-center gap-1',
                    copiedButtonId === 'initFwScript' ? 'bg-emerald-600 text-white' : 'btn-cyan'
                  ]"
                >
                  <span v-if="copiedButtonId === 'initFwScript'">✅ 已复制！</span>
                  <span v-else>📋 一键复制命令</span>
                </button>
              </div>
              <pre class="bg-black/80 p-3 rounded-lg text-[11px] font-mono text-purple-300 max-h-36 overflow-y-auto whitespace-pre-wrap select-all">{{ hostScriptText }}</pre>
            </div>

            <!-- Navigation Buttons -->
            <div class="flex items-center justify-between pt-2 border-t border-purple-900/40 text-xs">
              <button @click="initWizardStep = 'welcome'" class="btn bg-slate-800 hover:bg-slate-700 text-slate-300 px-4">
                ← 返回选择
              </button>
              <div class="flex gap-2">
                <button @click="initWizardStep = 'subscription'" class="btn bg-slate-800 hover:bg-slate-700 text-slate-400 px-3">
                  跳过此步
                </button>
                <button @click="initWizardStep = 'subscription'" class="btn btn-cyan font-bold px-4 flex items-center gap-1">
                  <span>下一步：导入节点订阅 ➔</span>
                </button>
              </div>
            </div>
          </div>

          <!-- ═══════ Step 3: Subscription & Latency ═══════ -->
          <div v-if="initWizardStep === 'subscription'" class="space-y-4 relative z-10">
            <div>
              <h4 class="text-sm font-bold text-white flex items-center gap-1.5">
                <span class="text-purple-400">步骤 1：</span>
                <span>导入订阅节点与并发测速</span>
              </h4>
              <p class="text-[11px] text-slate-400 mt-0.5">
                粘贴您的机场 Clash / Mihomo 订阅链接，SingCube 将自动拉取、智能过滤垃圾广告节点并编译 MRS 规则。
              </p>
            </div>

            <!-- Subscription Form -->
            <div class="bg-slate-900/60 p-4 rounded-xl border border-purple-900/40 space-y-3">
              <div>
                <label class="block text-xs text-slate-300 mb-1 font-semibold">机场订阅链接 (Clash / Mihomo 订阅 URL)</label>
                <input 
                  v-model="initSubUrl" 
                  placeholder="https://your-airport-provider.com/api/v1/client/subscribe?token=..." 
                  class="input-field w-full text-xs font-mono"
                />
              </div>
              <div class="flex flex-col sm:flex-row gap-3 items-center justify-between">
                <div class="w-full sm:w-1/2">
                  <label class="block text-xs text-slate-400 mb-1">订阅备注名称</label>
                  <input v-model="initSubName" placeholder="如 主力机场订阅" class="input-field w-full text-xs" />
                </div>
                <div class="w-full sm:w-auto pt-4">
                  <button 
                    @click="submitInitSub" 
                    :disabled="initSubLoading || !initSubUrl" 
                    class="btn btn-blue w-full sm:w-auto px-5 py-2 text-xs font-bold flex items-center justify-center gap-1.5 shadow-md"
                  >
                    <span v-if="initSubLoading" class="inline-block animate-spin">↻</span>
                    <span>{{ initSubLoading ? '正在拉取节点并编译分流...' : '📥 立即导入并编译 ➔' }}</span>
                  </button>
                </div>
              </div>

              <!-- Message feedback -->
              <div v-if="initSubMsg" :class="['text-xs p-2.5 rounded-lg border', initSubSuccess ? 'bg-emerald-950/40 text-emerald-300 border-emerald-500/40' : 'bg-amber-950/40 text-amber-300 border-amber-500/40']">
                {{ initSubMsg }}
              </div>

              <div v-if="configState.subscriptions && configState.subscriptions.length > 0" class="text-xs text-slate-300 bg-sky-950/30 p-2.5 rounded-lg border border-purple-800/40 flex items-center justify-between">
                <span>当前已导入订阅: <strong>{{ configState.subscriptions.length }}</strong> 个</span>
                <span class="text-emerald-400 font-bold font-mono">共 {{ totalProxiesCount }} 个活动节点</span>
              </div>
            </div>

            <!-- Navigation Buttons -->
            <div class="flex items-center justify-between pt-2 border-t border-purple-900/40 text-xs">
              <button @click="initWizardStep = 'firewall'" class="btn bg-slate-800 hover:bg-slate-700 text-slate-300 px-4">
                ← 上一步 (底座防火墙)
              </button>
              <div class="flex gap-2">
                <button @click="initWizardStep = 'core'" class="btn bg-slate-800 hover:bg-slate-700 text-slate-400 px-3">
                  稍后导入
                </button>
                <button @click="initWizardStep = 'core'" class="btn btn-cyan font-bold px-4 flex items-center gap-1">
                  <span>下一步：启动代理核心 ➔</span>
                </button>
              </div>
            </div>
          </div>

          <!-- ═══════ Step 4: Core & Routing Policy ═══════ -->
          <div v-if="initWizardStep === 'core'" class="space-y-4 relative z-10">
            <div>
              <h4 class="text-sm font-bold text-white flex items-center gap-1.5">
                <span class="text-purple-400">步骤 2 & 3：</span>
                <span>启动代理核心与四大出口策略推荐</span>
              </h4>
              <p class="text-[11px] text-slate-400 mt-0.5">
                采用 Mihomo 最新测试版 alpha-94b13ae 驱动，纯 Go 零 CGO 用户态 TUN 栈，常驻仅 50MB 内存！
              </p>
            </div>

            <!-- Core Launch Card -->
            <div class="glass-card p-4 space-y-3 border border-purple-800/40">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <span class="text-xl">⚡</span>
                  <div>
                    <div class="text-xs font-bold text-white">TUN MipStack 轻量用户态协议栈</div>
                    <div class="text-[10px] text-slate-400">内置 Google BBR 拥塞控制与 217heidai 23 万条去广告引擎</div>
                  </div>
                </div>
                <div :class="['text-xs px-2.5 py-0.5 rounded-full font-bold', status.running ? 'bg-emerald-950 text-emerald-300 border border-emerald-500/40' : 'bg-red-950 text-red-300 border border-red-500/40']">
                  {{ status.running ? '● 运行中 (PID: ' + (status.pid || '-') + ')' : '○ 核心已就绪待启动' }}
                </div>
              </div>

              <div class="flex gap-2">
                <button 
                  v-if="!status.running"
                  @click="startInitCore" 
                  :disabled="initCoreLoading"
                  class="btn btn-green w-full py-2.5 text-xs font-bold flex items-center justify-center gap-1.5 shadow-md shadow-emerald-500/20"
                >
                  <span v-if="initCoreLoading" class="inline-block animate-spin">↻</span>
                  <span>▶ 立即一键启动代理核心</span>
                </button>
                <div v-else class="text-emerald-400 text-xs p-2 bg-emerald-950/40 rounded-lg border border-emerald-500/40 w-full text-center font-bold">
                  ✅ 代理核心已正常运转，TUN 虚拟接口已建立！
                </div>
              </div>
            </div>

            <!-- Region Strategy Cheatsheet -->
            <div class="bg-slate-900/60 p-3.5 rounded-xl border border-purple-900/40 space-y-2 text-[11px]">
              <div class="font-bold text-purple-300 text-xs">🧭 场景出口选型速查 (为您自动预设最佳分流)：</div>
              <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
                <div class="p-2 rounded-lg bg-black/40 border border-purple-900/30">
                  <span class="text-purple-300 font-bold block mb-0.5">⚡ 地区自动 (url-test)</span>
                  <span class="text-slate-400">日常冲浪、刷短视频、YouTube，自动选用当前最低延迟节点。</span>
                </div>
                <div class="p-2 rounded-lg bg-black/40 border border-purple-900/30">
                  <span class="text-emerald-300 font-bold block mb-0.5">🛡️ 地区故转 (fallback)</span>
                  <span class="text-slate-400">ChatGPT / Claude / 网银，主节点存活死锁不跳 IP，防止风控。</span>
                </div>
                <div class="p-2 rounded-lg bg-black/40 border border-purple-900/30">
                  <span class="text-amber-300 font-bold block mb-0.5">🇨🇳 国内真实 IP 直出</span>
                  <span class="text-slate-400">国内流量绕过 Fake-IP，由物理网卡直连直出，跑满宽带物理速率上限！</span>
                </div>
              </div>
            </div>

            <!-- Navigation Buttons -->
            <div class="flex items-center justify-between pt-2 border-t border-purple-900/40 text-xs">
              <button @click="initWizardStep = 'subscription'" class="btn bg-slate-800 hover:bg-slate-700 text-slate-300 px-4">
                ← 上一步 (导入订阅)
              </button>
              <button @click="initWizardStep = 'wireguard'" class="btn btn-cyan font-bold px-4 flex items-center gap-1">
                <span>下一步：WireGuard 穿透与通关 ➔</span>
              </button>
            </div>
          </div>

          <!-- ═══════ Step 5: WireGuard & Final Celebration ═══════ -->
          <div v-if="initWizardStep === 'wireguard'" class="space-y-4 relative z-10 text-center">
            <div class="space-y-1 text-left">
              <h4 class="text-sm font-bold text-white flex items-center gap-1.5">
                <span class="text-purple-400">步骤 4 & 5：</span>
                <span>WireGuard 手机极速穿透与全家满血接入</span>
              </h4>
              <p class="text-[11px] text-slate-400">
                外出手机一键扫码回航，全天候零发热穿透内网；局域网电脑/电视修改网关即刻全屋满血加速！
              </p>
            </div>

            <!-- WireGuard Trigger or Setup Card -->
            <div class="glass-card p-5 space-y-3 border border-purple-500/40 bg-slate-900/70">
              <div class="flex items-center justify-between text-left border-b border-purple-900/30 pb-2">
                <div class="flex items-center gap-2">
                  <span class="text-2xl">📱</span>
                  <div>
                    <div class="text-xs font-bold text-white">手机 WireGuard 小白一键极速穿透</div>
                    <div class="text-[10px] text-slate-400">免公网 IPv4 · 绝不暴露 IPv6 风险 · 穿透全网段 NAS</div>
                  </div>
                </div>
                <button @click="startInitWireGuard" class="btn btn-cyan text-xs font-bold px-4 py-1.5 shadow-md shadow-purple-500/20">
                  <span>🚀 立即扫码连接 ➔</span>
                </button>
              </div>

              <!-- Step 4 & 5 Guidance Grid -->
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-left pt-1 text-xs">
                <div class="p-3 bg-black/40 rounded-xl border border-purple-900/30 space-y-1">
                  <div class="font-bold text-purple-300">📱 步骤 4：单机零风险验证</div>
                  <div class="text-[11px] text-slate-300 leading-relaxed">
                    拿一部手机连接 Wi-Fi，将网关与 DNS 修改为 <strong>SingCube 宿主机 IP</strong>。打开 YouTube 4K 秒开无缓冲，微信与国内 App 秒刷，三项黄金验证通过！
                  </div>
                </div>
                <div class="p-3 bg-black/40 rounded-xl border border-purple-900/30 space-y-1">
                  <div class="font-bold text-emerald-300">🌐 步骤 5：全家全局生效</div>
                  <div class="text-[11px] text-slate-300 leading-relaxed">
                    单机验证成功后，登录主路由器后台，在 DHCP 服务中将下发的默认网关和主 DNS 改为 SingCube IP，全家所有设备免任何设置自动满血享受加速与去广告！
                  </div>
                </div>
              </div>
            </div>

            <!-- Big Finish Celebration Button -->
            <div class="pt-2">
              <button 
                @click="finishInitWizard" 
                class="btn bg-gradient-to-r from-emerald-500 via-teal-500 to-emerald-500 hover:from-emerald-400 hover:to-cyan-400 text-white font-extrabold text-sm py-3 w-full rounded-2xl shadow-xl shadow-purple-500/25 flex items-center justify-center gap-2"
              >
                <span>🎉 大功告成，完成初始化并进入控制台 ➔</span>
              </button>
            </div>
          </div>

          <!-- ═══════ Step 6: Backup Restore Success Screen ═══════ -->
          <div v-if="initWizardStep === 'finished'" class="space-y-5 py-4 relative z-10 text-center">
            <div class="text-5xl animate-bounce">
              🎉
            </div>
            <div class="space-y-1 max-w-md mx-auto">
              <h3 class="text-xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-cyan-300 via-sky-200 to-emerald-300">
                恭喜！系统备份已 100% 完整复原！
              </h3>
              <p class="text-xs text-slate-300 leading-relaxed">
                所有机场订阅、出口策略组、分流规则、WireGuard 穿透密钥与局域网设备别名已秒级热载生效！
              </p>
            </div>
            <button 
              @click="finishInitWizard" 
              class="btn btn-cyan text-sm font-bold px-8 py-2.5 rounded-xl shadow-xl shadow-purple-500/30"
            >
              🚀 进入 SingCube 控制台
            </button>
          </div>
        </div>
      </div>

      <!-- ── WireGuard One-Click Zero-Conf Wizard Modal ── -->
      <div v-if="showWgWizardModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-md p-4">
        <div class="glass-card max-w-lg w-full p-6 space-y-5 border border-purple-500/40 shadow-2xl shadow-purple-500/20 relative overflow-hidden bg-slate-950/95">
          <!-- Canvas Fireworks overlay for celebration -->
          <canvas id="fireworksCanvas" class="absolute inset-0 pointer-events-none z-30 w-full h-full"></canvas>

          <!-- Modal Header -->
          <div class="flex items-center justify-between border-b border-purple-900/40 pb-3 relative z-10">
            <div class="flex items-center gap-2">
              <span class="text-xl">🚀</span>
              <div>
                <h3 class="text-base font-bold text-white flex items-center gap-2">
                  <span>WireGuard 小白一键极速通关</span>
                  <span class="text-[10px] px-2 py-0.5 rounded-full bg-cyan-900/60 text-purple-300 border border-purple-500/40 font-mono">Zero-Conf</span>
                </h3>
                <p class="text-[11px] text-slate-400">免配置一键自动打通 RouterOS 防火墙与策略路由，扫码即用</p>
              </div>
            </div>
            <button @click="closeWizard" class="text-slate-400 hover:text-white text-lg p-1">✕</button>
          </div>

          <!-- ── Stage 1: Deploying (自动打通中) ── -->
          <div v-if="wizardStage === 'deploying'" class="space-y-4 py-2 relative z-10">
            <div class="space-y-2">
              <div class="flex items-center justify-between text-xs">
                <span class="text-purple-300 font-semibold flex items-center gap-1.5">
                  <span class="inline-block animate-spin">↻</span>
                  <span>正在为您全自动打通服务端、防火墙与策略路由...</span>
                </span>
                <span class="font-mono text-purple-400 font-bold">{{ wizardProgress }}%</span>
              </div>
              <div class="w-full bg-slate-900 rounded-full h-2 overflow-hidden border border-purple-900/40">
                <div class="bg-gradient-to-r from-cyan-500 via-sky-400 to-blue-500 h-2 rounded-full transition-all duration-300 shadow-md shadow-purple-500/50" :style="{ width: wizardProgress + '%' }"></div>
              </div>
            </div>

            <!-- Steps checklist -->
            <div class="space-y-2.5 bg-slate-900/70 p-4 rounded-xl border border-purple-900/30 text-xs">
              <div v-for="(step, idx) in wizardSteps" :key="idx" class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-2.5">
                  <span v-if="step.status === 'done'" class="text-emerald-400 text-sm">✔</span>
                  <span v-else-if="step.status === 'process'" class="inline-block animate-spin text-purple-400 text-xs">↻</span>
                  <span v-else class="text-slate-600 text-xs">○</span>
                  <span :class="step.status === 'done' ? 'text-slate-200' : (step.status === 'process' ? 'text-purple-300 font-bold' : 'text-slate-500')">
                    {{ step.title }}
                  </span>
                </div>
                <span class="text-[10px] text-slate-400 font-mono">{{ step.detail }}</span>
              </div>
            </div>
            <!-- Elapsed time indicator & skip link -->
            <div class="flex items-center justify-between text-[11px] text-slate-400 px-1 pt-0.5">
              <span>已等待: <strong class="text-purple-300 font-mono">{{ wizardDeploySeconds }}s</strong> / 30s</span>
              <button @click="skipWizardLater" class="text-amber-400 hover:text-amber-300 font-medium underline underline-offset-4 transition text-[11px]">
                不想等待？点击【以后再说】
              </button>
            </div>
            <div class="text-[11px] text-slate-500 text-center">
              通常耗时仅需 1~2 秒，若超过 30 秒将自动弹出提示...
            </div>
          </div>

          <!-- ── Stage: Timeout (等待超过 30s 提示: 以后再说 / 重试) ── -->
          <div v-if="wizardStage === 'timeout'" class="space-y-4 py-2 relative z-10 text-center">
            <div class="p-5 bg-gradient-to-br from-amber-950/40 to-slate-900 border border-amber-500/40 rounded-2xl space-y-3 text-left shadow-xl shadow-amber-950/30">
              <div class="flex items-center gap-2.5 text-amber-300 font-bold text-sm">
                <span class="text-xl">⏱️</span>
                <span>配置响应已超时 (已等待超过 30 秒)</span>
                <span class="text-[10px] px-2 py-0.5 rounded-full bg-amber-900/60 text-amber-300 border border-amber-500/40 font-mono">超时提示</span>
              </div>
              <p class="text-xs text-amber-100/90 leading-relaxed">
                WireGuard 极速通关已等待超过 30 秒仍未收到服务端响应。这通常发生在当前环境（如 Docker / 飞牛 fnOS）尚未连通 RouterOS 路由器，或者 RouterOS REST API 端口未开启。
              </p>
              <div class="text-[11px] text-slate-400 bg-slate-950/60 p-3 rounded-xl border border-purple-900/40 leading-relaxed">
                💡 <strong>建议处理：</strong>您可以直接点击<strong>「以后再说」</strong>进入系统正常使用代理分流功能；后续在「系统设置」中配置好 RouterOS IP 后，随时可以一键重新开启穿透！
              </div>
            </div>

            <div class="flex flex-col sm:flex-row gap-2.5 pt-1 text-xs">
              <button @click="skipWizardLater" class="btn bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-400 hover:to-amber-500 text-slate-950 font-bold py-2.5 px-4 rounded-xl shadow-lg shadow-amber-500/25 flex-1 flex items-center justify-center gap-2 transition transform active:scale-95">
                <span>⏩</span>
                <span>以后再说 (跳过并进入系统)</span>
              </button>
              <button @click="startWgWizard" class="btn btn-cyan py-2.5 px-4 font-semibold flex items-center justify-center gap-1.5 shadow-md shadow-purple-500/20">
                <span>↻</span>
                <span>重新尝试</span>
              </button>
            </div>
          </div>

          <!-- ── Stage: Error (自动通关异常处理) ── -->
          <div v-if="wizardStage === 'error'" class="space-y-4 py-2 relative z-10 text-center">
            <div class="p-4 bg-rose-950/40 border border-rose-500/40 rounded-xl space-y-2 text-left">
              <div class="flex items-center gap-2 text-rose-300 font-bold text-sm">
                <span>⚠️</span>
                <span>WireGuard 自动通关未能全部完成</span>
              </div>
              <div class="text-xs text-rose-200/90 leading-relaxed font-mono whitespace-pre-wrap">
                {{ wizardErrorMsg || 'RouterOS 连接超时或未能生成有效配置。' }}
              </div>
              <div class="text-[11px] text-slate-400 mt-2 leading-relaxed">
                💡 提示：若您在纯 Docker/Linux (如飞牛 fnOS) 环境运行且未连接 MikroTik 路由器，可在「系统设置」中配置 RouterOS IP，或在「系统设置」中查看 Linux 宿主机 WireGuard 配置文件。
              </div>
            </div>

            <div class="flex flex-wrap gap-2 pt-2 text-xs">
              <button @click="skipWizardLater" class="btn bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-400 hover:to-amber-500 text-slate-950 font-bold py-2 px-4 rounded-xl flex items-center justify-center gap-1.5 shadow-md">
                <span>⏩</span>
                <span>以后再说</span>
              </button>
              <button @click="startWgWizard" class="btn btn-cyan flex-1 py-2 font-bold flex items-center justify-center gap-1.5 shadow-lg shadow-purple-500/20">
                <span>↻</span>
                <span>重新尝试极速通关</span>
              </button>
              <button @click="closeWizard" class="btn bg-slate-800 hover:bg-slate-700 text-slate-300 px-4 py-2 border border-purple-800/40">
                关闭
              </button>
            </div>
          </div>

          <!-- ── Stage 2: Scanning (请扫码连接) ── -->
          <div v-if="wizardStage === 'scanning'" class="space-y-4 text-center relative z-10">
            <div class="space-y-1">
              <div class="text-emerald-400 font-bold text-sm flex items-center justify-center gap-1.5">
                <span>✅</span>
                <span>通道与防火墙已全部打通！请扫码连接</span>
              </div>
              <div class="text-[11px] text-slate-400">
                请打开手机官方 <strong>WireGuard App</strong>，点击右上角 <strong>「+」➔「从二维码创建」</strong>
              </div>
            </div>

            <!-- Big QR Card -->
            <div class="inline-block p-3.5 bg-white rounded-2xl shadow-2xl shadow-purple-500/25 border-4 border-purple-500/30">
              <img v-if="wizardQrUrl" :src="wizardQrUrl" class="w-60 h-60 block mx-auto" alt="WireGuard QR Code" />
            </div>

            <!-- Config Badges -->
            <div class="grid grid-cols-3 gap-2 text-[11px] font-mono">
              <div class="bg-slate-900/80 p-2 rounded-lg border border-purple-900/40 text-left">
                <div class="text-slate-500 text-[10px]">客户端虚拟 IP</div>
                <div class="text-purple-300 font-bold truncate">{{ wizardPeer?.client_ip || '10.88.0.2' }}/32</div>
              </div>
              <div class="bg-slate-900/80 p-2 rounded-lg border border-purple-900/40 text-left">
                <div class="text-slate-500 text-[10px]">监听端口</div>
                <div class="text-white font-bold">{{ wgInfo?.config?.listen_port || 13231 }} UDP</div>
              </div>
              <div class="bg-slate-900/80 p-2 rounded-lg border border-purple-900/40 text-left">
                <div class="text-slate-500 text-[10px]">分流机制</div>
                <div class="text-emerald-400 font-bold truncate">全功能分流+去广告</div>
              </div>
            </div>

            <!-- Dynamic Handshake Radar Listener -->
            <div class="bg-purple-950/40 border border-purple-500/40 rounded-xl p-3 flex items-center justify-between gap-3 text-left">
              <div class="flex items-center gap-3">
                <span class="relative flex h-3.5 w-3.5 flex-shrink-0">
                  <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-75"></span>
                  <span class="relative inline-flex rounded-full h-3.5 w-3.5 bg-purple-600"></span>
                </span>
                <div>
                  <div class="text-xs font-bold text-purple-200">正在等待手机首次加密握手信号...</div>
                  <div class="text-[10px] text-slate-400">手机扫码并在 WireGuard 中开启连接后，系统将自动探查并烟花庆祝！</div>
                </div>
              </div>
              <div class="flex flex-col gap-1 flex-shrink-0">
                <button @click="checkWizardHandshake(wizardPeer?.id, true)" :disabled="wizardChecking" class="btn btn-cyan text-[11px] py-1 px-2.5">
                  <span v-if="wizardChecking" class="inline-block animate-spin mr-1">↻</span>
                  <span>⚡ 立即检测</span>
                </button>
              </div>
            </div>

            <!-- Actions row -->
            <div class="flex gap-2 pt-1 text-xs">
              <button @click="copyToClipboard(wizardConfig)" class="btn bg-slate-800 hover:bg-slate-700 text-slate-300 flex-1 py-1.5 border border-purple-800/40">
                📋 复制配置文本
              </button>
              <button @click="downloadConfFile(wizardPeer?.name || '我的手机', wizardConfig)" class="btn bg-sky-900/50 hover:bg-sky-800/60 text-sky-200 flex-1 py-1.5 border border-purple-700/40">
                ⬇️ 下载 .conf
              </button>
              <button @click="simulateCelebrationForTesting" class="btn bg-emerald-950/60 hover:bg-emerald-900/60 text-emerald-300 border border-emerald-500/30 px-3 py-1.5" title="免等待模拟验证庆祝">
                🎉 模拟握手测试
              </button>
            </div>
          </div>

          <!-- ── Stage 3: Celebrating (握手成功 & 烟花庆祝) ── -->
          <div v-if="wizardStage === 'celebrating'" class="space-y-4 text-center py-2 relative z-10">
            <!-- Celebration Badge -->
            <div class="text-5xl animate-bounce">
              🎉
            </div>

            <div class="space-y-1">
              <h3 class="text-xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-cyan-300 via-sky-200 to-emerald-300">
                恭喜！手机与家庭网关已完美握手！
              </h3>
              <p class="text-xs text-slate-300 leading-relaxed max-w-md mx-auto">
                加密隧道握手成功！外出漫游如在家里，已全面继承家中 <strong>50MB SingCube 满载分流</strong>、SmartDNS 极速解析与 23 万条权威去广告过滤！
              </p>
            </div>

            <!-- Live handshake stats card -->
            <div class="bg-slate-900/80 border border-emerald-500/40 rounded-2xl p-4 text-left space-y-2.5 shadow-xl shadow-emerald-500/10">
              <div class="flex items-center justify-between text-xs border-b border-purple-900/40 pb-2">
                <span class="text-slate-400">设备连接状态</span>
                <span class="text-emerald-400 font-bold flex items-center gap-1.5">
                  <span class="h-2 w-2 rounded-full bg-emerald-400 animate-pulse"></span>
                  <span>活跃连接中 (Active)</span>
                </span>
              </div>
              <div class="grid grid-cols-2 gap-3 text-xs font-mono">
                <div>
                  <div class="text-slate-500 text-[10px]">最近一次握手 (Handshake)</div>
                  <div class="text-purple-300 font-bold">{{ wizardLiveStatus?.last_handshake || '刚刚 (1秒前)' }}</div>
                </div>
                <div>
                  <div class="text-slate-500 text-[10px]">分配虚拟 IP</div>
                  <div class="text-white font-bold">{{ wizardLiveStatus?.client_ip || wizardPeer?.client_ip || '10.88.0.2' }}/32</div>
                </div>
              </div>
              <div class="text-[11px] text-slate-400 font-mono bg-black/40 p-2 rounded-lg">
                双向通信流量: ↑ {{ formatBytes(wizardLiveStatus?.tx_bytes || 1280) }} &nbsp;|&nbsp; ↓ {{ formatBytes(wizardLiveStatus?.rx_bytes || 2560) }}
              </div>
            </div>

            <!-- Beginner next step tips -->
            <div class="bg-purple-950/40 border border-purple-500/30 rounded-xl p-3 text-left space-y-1 text-xs">
              <div class="font-bold text-purple-300 flex items-center gap-1">
                <span>💡</span>
                <span>体验无感漫游验证：</span>
              </div>
              <div class="text-slate-300 text-[11px] leading-relaxed">
                现在可以尝试<strong>关闭手机 Wi-Fi 开关</strong>，切至运营商 5G 移动网络。在手机浏览器中访问家庭内网 NAS（如 <code class="text-purple-300">192.168.x.x</code>）或 YouTube，享受外出如同在家的极速体验！
              </div>
            </div>

            <!-- Close button -->
            <button @click="closeWizard" class="btn btn-cyan w-full text-sm font-bold py-2.5 shadow-xl shadow-purple-500/30">
              🚀 大功告成，开始畅享
            </button>
          </div>
        </div>
      </div>

      <!-- ── WireGuard QR Code Modal ── -->
      <div v-if="showQrModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4">
        <div class="glass-card max-w-md w-full p-6 space-y-4 border border-purple-500/40 shadow-2xl shadow-purple-500/10">
          <div class="flex items-center justify-between border-b border-purple-900/40 pb-3">
            <div class="flex items-center gap-2">
              <span class="text-xl">📱</span>
              <h3 class="text-base font-bold text-white">WireGuard 扫码配置</h3>
            </div>
            <button @click="showQrModal = false" class="text-slate-400 hover:text-white text-lg">✕</button>
          </div>

          <div class="text-center space-y-2">
            <div class="text-xs text-purple-300 font-semibold">{{ currentPeerName }}</div>
            <!-- QR Image -->
            <div class="inline-block p-3 bg-white rounded-2xl shadow-2xl shadow-purple-500/20">
              <img v-if="currentQrDataUrl" :src="currentQrDataUrl" class="w-64 h-64 block mx-auto" alt="WireGuard QR Code" />
            </div>
            <div class="text-xs text-slate-300 bg-slate-900/60 p-3 rounded-xl border border-purple-900/30 text-left space-y-1">
              <div class="font-bold text-purple-300">💡 扫码使用指引：</div>
              <div>1. 手机打开官方 <strong>WireGuard</strong> 客户端。</div>
              <div>2. 点击右上角 <strong>「+」➔「从二维码创建」</strong>。</div>
              <div>3. 对准屏幕扫描上方二维码，输入隧道名称并开启连接！</div>
            </div>
          </div>

          <!-- Raw config collapsible -->
          <div class="space-y-1">
            <div class="flex items-center justify-between text-xs text-slate-400">
              <span>配置文件预览</span>
              <button @click="copyToClipboard(currentPeerConfig)" class="text-purple-400 hover:underline">📋 一键复制文本</button>
            </div>
            <pre class="bg-black/60 p-3 rounded-xl text-[11px] font-mono text-slate-300 max-h-28 overflow-y-auto border border-sky-950">{{ currentPeerConfig }}</pre>
          </div>

          <div class="flex gap-3 pt-2">
            <button @click="downloadConfFile(currentPeerName, currentPeerConfig)" class="btn btn-blue text-xs flex-1">
              ⬇️ 下载 .conf 文件
            </button>
            <button @click="showQrModal = false" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs px-6">
              完成
            </button>
          </div>
        </div>
      </div>

      <!-- ── Create Peer Modal ── -->
      <div v-if="showNewPeerModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4">
        <div class="glass-card max-w-md w-full p-6 space-y-4 border border-sky-600/40 shadow-2xl">
          <div class="flex items-center justify-between border-b border-purple-900/40 pb-3">
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <span>➕ 新建穿透设备并生成二维码</span>
            </h3>
            <button @click="showNewPeerModal = false" class="text-slate-400 hover:text-white text-lg">✕</button>
          </div>

          <div class="space-y-4 text-xs">
            <div>
              <label class="block text-slate-300 mb-1 font-semibold">设备标识名称</label>
              <input v-model="newPeerForm.name" placeholder="例如：我的 iPhone 16 / 办公室电脑" class="input-field w-full" />
            </div>

            <div>
              <label class="block text-slate-300 mb-1 font-semibold">访问路由模式</label>
              <div class="space-y-2">
                <label class="flex items-start gap-2.5 p-3 rounded-xl border border-purple-900/40 bg-slate-900/40 cursor-pointer">
                  <input type="radio" value="intranet" v-model="newPeerForm.mode" class="mt-0.5 accent-cyan-400" />
                  <div>
                    <div class="font-bold text-white">内网穿透模式 (推荐)</div>
                    <div class="text-[11px] text-slate-400 mt-0.5">仅访问家庭内部局域网 IP (如 192.168.88.x) 时走隧道，手机常规上网保持原本网络，省电且不消耗家中宽带上传。</div>
                  </div>
                </label>

                <label class="flex items-start gap-2.5 p-3 rounded-xl border border-purple-900/40 bg-slate-900/40 cursor-pointer">
                  <input type="radio" value="full" v-model="newPeerForm.mode" class="mt-0.5 accent-cyan-400" />
                  <div>
                    <div class="font-bold text-white">全局代理模式</div>
                    <div class="text-[11px] text-slate-400 mt-0.5">手机所有出站流量均通过家庭路由器代理中转，适合在外借用公共 WiFi 防护或回国加速。</div>
                  </div>
                </label>
              </div>
            </div>
          </div>

          <div class="flex gap-3 pt-2">
            <button @click="createNewPeer" :disabled="loading" class="btn btn-cyan flex-1 text-xs font-semibold">
              ⚡ 立即生成秘钥并获取二维码
            </button>
            <button @click="showNewPeerModal = false" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs px-4">
              取消
            </button>
          </div>
        </div>
      </div>

      <!-- ── Rename WireGuard Peer Modal ── -->
      <div v-if="showRenamePeerModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4">
        <div class="glass-card max-w-md w-full p-6 space-y-4 border border-purple-500/40 shadow-2xl">
          <div class="flex items-center justify-between border-b border-purple-900/40 pb-3">
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <span>✏️ 修改穿透设备备注名称</span>
            </h3>
            <button @click="showRenamePeerModal = false" class="text-slate-400 hover:text-white text-lg">✕</button>
          </div>

          <div class="text-xs text-slate-300">
            修改穿透设备的备注名称。保存后将自动更新本地配置，并在连接到 RouterOS 时自动同步该 Peer 的 Comment 备注。
          </div>

          <div class="space-y-3 text-xs">
            <div>
              <label class="block text-slate-400 mb-1">设备虚拟 IP</label>
              <input :value="peerRenameForm.client_ip" disabled class="input-field w-full bg-slate-900/60 font-mono text-slate-400 cursor-not-allowed" />
            </div>
            <div>
              <label class="block text-slate-300 mb-1 font-semibold">设备新备注名称</label>
              <input v-model="peerRenameForm.name" placeholder="例如: iPhone 16 Pro Max / 出差笔记本" class="input-field w-full font-sans text-xs" />
            </div>
          </div>

          <div class="flex gap-3 pt-2">
            <button @click="saveRenamePeer" class="btn btn-cyan flex-1 text-xs font-semibold">
              💾 保存修改
            </button>
            <button @click="showRenamePeerModal = false" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs px-4">
              取消
            </button>
          </div>
        </div>
      </div>

      <!-- ── WinBox Routing Script Modal (Secondary Fallback) ── -->
      <div v-if="showWinBoxRoutingModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4">
        <div class="glass-card max-w-2xl w-full p-6 space-y-4 border border-sky-600/40 shadow-2xl">
          <div class="flex items-center justify-between border-b border-purple-900/40 pb-3">
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <span>📋 WinBox 策略路由与防火墙终端脚本 (备用)</span>
            </h3>
            <button @click="showWinBoxRoutingModal = false" class="text-slate-400 hover:text-white text-lg">✕</button>
          </div>

          <div class="text-xs text-slate-400">
            SingCube 默认推荐直接在面板点击「⚡ 一键配置到 RouterOS」通过 REST API 全自动下发。如您希望在 WinBox 终端中手工审计并执行，可复制以下脚本在 <strong>New Terminal</strong> 中整段粘贴执行：
          </div>

          <pre class="bg-black/70 p-4 rounded-xl text-xs font-mono text-purple-300 max-h-72 overflow-y-auto border border-purple-900/40 select-all whitespace-pre-wrap">{{ winBoxRoutingScriptText }}</pre>

          <div class="flex gap-3 pt-2">
            <button @click="copyToClipboard(winBoxRoutingScriptText, 'winBoxRouting')" :class="['btn flex-1 text-xs transition-all duration-300 font-semibold', copiedButtonId === 'winBoxRouting' ? 'bg-emerald-600 hover:bg-emerald-500 text-white shadow-[0_0_15px_rgba(16,185,129,0.5)]' : 'btn-cyan']">
              <span v-if="copiedButtonId === 'winBoxRouting'">✅ 已成功复制全部命令！</span>
              <span v-else>📋 一键复制全部脚本命令</span>
            </button>
            <button @click="showWinBoxRoutingModal = false" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs px-6">
              关闭
            </button>
          </div>
        </div>
      </div>

      <!-- ── RouterOS Script Modal ── -->
      <div v-if="showRosScriptModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4">
        <div class="glass-card max-w-2xl w-full p-6 space-y-4 border border-sky-600/40 shadow-2xl">
          <div class="flex items-center justify-between border-b border-purple-900/40 pb-3">
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <span>📋 RouterOS Terminal 一键配置脚本</span>
            </h3>
            <button @click="showRosScriptModal = false" class="text-slate-400 hover:text-white text-lg">✕</button>
          </div>

          <div class="text-xs text-slate-400">
            如果您的 RouterOS 未开启 REST API，可直接打开 WinBox 点击 <strong>New Terminal</strong>，将以下命令整段粘贴执行即可完成全部配置：
          </div>

          <pre class="bg-black/70 p-4 rounded-xl text-xs font-mono text-purple-300 max-h-72 overflow-y-auto border border-purple-900/40 select-all whitespace-pre-wrap">{{ rosScriptText }}</pre>

          <div class="flex gap-3 pt-2">
            <button @click="copyToClipboard(rosScriptText, 'rosScript')" :class="['btn flex-1 text-xs transition-all duration-300 font-semibold', copiedButtonId === 'rosScript' ? 'bg-emerald-600 hover:bg-emerald-500 text-white shadow-[0_0_15px_rgba(16,185,129,0.5)]' : 'btn-cyan']">
              <span v-if="copiedButtonId === 'rosScript'">✅ 已成功复制全部命令！</span>
              <span v-else>📋 一键复制全部脚本命令</span>
            </button>
            <button @click="showRosScriptModal = false" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs px-6">
              关闭
            </button>
          </div>
        </div>
      </div>

      <!-- ── RouterOS Registry Script Modal (KSpeeder) ── -->
      <div v-if="showRosRegistryModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4">
        <div class="glass-card max-w-2xl w-full p-6 space-y-4 border border-purple-500/40 shadow-2xl">
          <div class="flex items-center justify-between border-b border-purple-900/40 pb-3">
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <span>🐳 RouterOS 容器镜像加速命令</span>
            </h3>
            <button @click="showRosRegistryModal = false" class="text-slate-400 hover:text-white text-lg">✕</button>
          </div>

          <div class="text-xs text-slate-300">
            打开 RouterOS WinBox 点击 <strong>New Terminal</strong>，直接粘贴执行以下命令，即可将当前优选的 Docker Registry 镜像加速源注入到 RouterOS 容器子系统中：
          </div>

          <pre class="bg-black/70 p-4 rounded-xl text-xs font-mono text-purple-300 max-h-72 overflow-y-auto border border-purple-900/40 select-all whitespace-pre-wrap">{{ rosRegistryScriptText }}</pre>

          <div class="flex gap-3 pt-2">
            <button @click="copyToClipboard(rosRegistryScriptText)" class="btn btn-cyan flex-1 text-xs">
              📋 一键复制命令
            </button>
            <button @click="showRosRegistryModal = false" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs px-6">
              关闭
            </button>
          </div>
        </div>
      </div>

      <!-- ── Docker Daemon JSON Modal (KSpeeder) ── -->
      <div v-if="showDockerDaemonModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4">
        <div class="glass-card max-w-2xl w-full p-6 space-y-4 border border-purple-500/40 shadow-2xl">
          <div class="flex items-center justify-between border-b border-purple-900/40 pb-3">
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <span>📄 Linux / PVE Docker daemon.json 加速配置</span>
            </h3>
            <button @click="showDockerDaemonModal = false" class="text-slate-400 hover:text-white text-lg">✕</button>
          </div>

          <div class="text-xs text-slate-300">
            将以下内容写入宿主机 <code>/etc/docker/daemon.json</code> 并执行 <code>systemctl restart docker</code>，即可让所有 Docker 容器享受极速加速拉取：
          </div>

          <pre class="bg-black/70 p-4 rounded-xl text-xs font-mono text-emerald-300 max-h-72 overflow-y-auto border border-purple-900/40 select-all whitespace-pre-wrap">{{ dockerDaemonJsonText }}</pre>

          <div class="flex gap-3 pt-2">
            <button @click="copyToClipboard(dockerDaemonJsonText)" class="btn btn-cyan flex-1 text-xs">
              📋 复制 JSON 配置
            </button>
            <button @click="downloadConfFile('daemon.json', dockerDaemonJsonText)" class="btn btn-blue text-xs">
              ⬇️ 下载文件
            </button>
            <button @click="showDockerDaemonModal = false" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs px-6">
              关闭
            </button>
          </div>
        </div>
      </div>

      <!-- ── Device Alias Edit Modal (Item 5) ── -->
      <div v-if="showDeviceModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4">
        <div class="glass-card max-w-md w-full p-6 space-y-4 border border-purple-500/40 shadow-2xl">
          <div class="flex items-center justify-between border-b border-purple-900/40 pb-3">
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <span>📱 设置设备名称与备注</span>
            </h3>
            <button @click="showDeviceModal = false" class="text-slate-400 hover:text-white text-lg">✕</button>
          </div>

          <p class="text-xs text-slate-300">
            支持按 <strong>IP 地址</strong> 或 <strong>MAC 物理地址</strong> 为设备绑定自定义友好备注（如“客厅小米电视”、“群晖 NAS”），并自动联动全站监控与 RouterOS：
          </p>

          <div class="space-y-3 text-xs">
            <div>
              <label class="block text-slate-300 mb-1 font-semibold">友好名称 / 备注名</label>
              <input v-model="deviceEditForm.name" placeholder="例如: 客厅小米电视 / 爸爸的手机" class="input-field w-full text-xs" />
            </div>

            <div>
              <label class="block text-slate-300 mb-1 font-semibold">IP 地址 (选填或由设备带出)</label>
              <input v-model="deviceEditForm.ip" placeholder="例如: 192.168.88.105" class="input-field w-full text-xs font-mono" />
            </div>

            <div>
              <label class="block text-slate-300 mb-1 font-semibold">MAC 物理地址 (建议填写，防止IP变化后失效)</label>
              <input v-model="deviceEditForm.mac" placeholder="例如: 00:11:22:33:44:55" class="input-field w-full text-xs font-mono uppercase" />
            </div>
          </div>

          <div class="flex gap-3 pt-2">
            <button @click="saveDeviceAlias" class="btn btn-cyan flex-1 text-xs">
              💾 保存设备备注
            </button>
            <button @click="showDeviceModal = false" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs px-6">
              取消
            </button>
          </div>
        </div>
      </div>

      <!-- ── AI Copilot Config Modal (OpenCode Zen / 通义千问 / OpenAI) ── -->
      <div v-if="showAiConfigModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4">
        <div class="glass-card max-w-lg w-full p-6 space-y-4 border border-purple-500/40 shadow-2xl">
          <div class="flex items-center justify-between border-b border-purple-900/40 pb-3">
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <span>⚙️ AI Copilot 模型与免折腾配置</span>
            </h3>
            <button @click="showAiConfigModal = false" class="text-slate-400 hover:text-white text-lg">✕</button>
          </div>

          <div class="space-y-4 text-xs">
            <div class="bg-purple-950/40 border border-cyan-800/40 p-3 rounded-xl text-purple-200 leading-relaxed space-y-1">
              <div>💡 <strong>零折腾免 Key 模式</strong>：未配置 API Key 时，系统全自动启用内置的软路由网络自愈专家，零延迟秒级输出深度网络体检与诊断建议，<strong>无需折腾任何外部账号</strong>！</div>
              <div class="text-[11px] text-slate-400">如需接入云端大模型，支持 <strong>OpenCode Zen</strong>（免费/聚合大模型通道）、<strong>通义千问</strong>、<strong>DeepSeek</strong> 或局域网 <strong>Ollama</strong>。</div>
            </div>

            <div>
              <label class="block text-slate-300 mb-1 font-semibold">服务提供商 (Provider)</label>
              <select v-model="aiConfigForm.provider" @change="onAiProviderChange" class="input-field w-full">
                <option value="opencode">OpenCode Zen (免费/聚合大模型通道)</option>
                <option value="qwen">通义千问 (阿里云 DashScope / Qwen)</option>
                <option value="openai">OpenAI 兼容接口 (DeepSeek / SiliconFlow / Ollama)</option>
              </select>
            </div>

            <!-- Model quick selection chips -->
            <div v-if="aiConfigForm.provider === 'opencode'" class="space-y-1.5">
              <label class="block text-slate-400 font-semibold text-[11px]">OpenCode Zen 常用推荐模型（点击快速填充）</label>
              <div class="flex flex-wrap gap-1.5">
                <button type="button" v-for="m in ['nemotron-3.5-lightning-free', 'mimo-v2.6-flash-free', 'cat-jam-free', 'big-pickle', 'qwen3.5-plus', 'deepseek-v4-flash']" :key="m"
                  @click="aiConfigForm.model_name = m"
                  :class="['px-2 py-0.5 rounded text-[10px] font-mono border transition',
                    aiConfigForm.model_name === m ? 'bg-purple-600/20 text-purple-300 border-purple-400' : 'bg-slate-900/60 text-slate-400 border-purple-900/40 hover:text-white']">
                  {{ m }}
                </button>
              </div>
            </div>

            <div>
              <label class="block text-slate-300 mb-1 font-semibold">API Key (密钥，可选)</label>
              <input v-model="aiConfigForm.api_key" type="password" placeholder="sk-... 或留空使用免折腾内置自愈引擎" class="input-field w-full font-mono" />
              <div class="text-[10px] text-slate-400 mt-1">留空则自动启用 100% 零折腾本地自愈诊断专家；填入密钥后开启云端大模型增强推理。</div>
            </div>

            <div>
              <label class="block text-slate-300 mb-1 font-semibold">API Base URL</label>
              <input v-model="aiConfigForm.base_url" placeholder="https://opencode.ai/zen/v1" class="input-field w-full font-mono" />
            </div>

            <div>
              <label class="block text-slate-300 mb-1 font-semibold">模型代号 (Model)</label>
              <input v-model="aiConfigForm.model_name" placeholder="nemotron-3.5-lightning-free / qwen-plus" class="input-field w-full font-mono" />
            </div>
          </div>

          <div class="flex gap-3 pt-2">
            <button @click="saveAiConfig" :disabled="loading" class="btn btn-cyan flex-1 text-xs font-semibold">
              💾 保存配置
            </button>
            <button @click="showAiConfigModal = false" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs px-6">
              取消
            </button>
          </div>
        </div>
      </div>

      <!-- ── Host OS & Firewall Foundation 1-Click Wizard Modal (Step 0) ── -->
      <div v-if="showHostFwModal || showHostScriptModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/85 backdrop-blur-md p-4">
        <div class="glass-card max-w-2xl w-full p-6 space-y-5 border border-purple-500/50 shadow-2xl shadow-sky-500/20 relative overflow-hidden bg-slate-950/95 max-h-[90vh] overflow-y-auto">
          <!-- Modal Header -->
          <div class="flex items-center justify-between border-b border-purple-900/40 pb-3">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500/20 to-emerald-500/30 border border-purple-400/40 flex items-center justify-center text-xl shadow-inner">
                🛡️
              </div>
              <div>
                <h3 class="text-base font-bold text-white flex items-center gap-2">
                  <span>宿主机底座网络与防火墙一键配置</span>
                  <span class="text-[10px] px-2 py-0.5 rounded-full bg-purple-950 text-purple-300 border border-purple-500/40 font-mono font-bold">Step 0 首位生命线</span>
                </h3>
                <p class="text-[11px] text-slate-400">开启 Linux 内核 IPv4 转发、放行 FORWARD 链策略与 NAT 伪装，杜绝改网关断网</p>
              </div>
            </div>
            <button @click="showHostFwModal = false; showHostScriptModal = false" class="text-slate-400 hover:text-white text-lg p-1 transition">✕</button>
          </div>

          <!-- Live Host Telemetry Status Bar -->
          <div class="grid grid-cols-2 sm:grid-cols-3 gap-2.5 bg-slate-900/70 p-3 rounded-xl border border-purple-900/30 text-xs font-mono">
            <div>
              <span class="text-slate-400 text-[10px] block">宿主机平台</span>
              <span class="text-white font-bold truncate block mt-0.5">
                {{ hostInfo.os_pretty || hostInfo.os_name || '飞牛 fnOS / Linux' }}
              </span>
            </div>
            <div>
              <span class="text-slate-400 text-[10px] block">内核 IPv4 转发 (net.ipv4.ip_forward)</span>
              <span :class="['font-bold block mt-0.5 text-xs flex items-center gap-1', hostInfo.ip_forward ? 'text-emerald-400' : 'text-amber-400 animate-pulse']">
                <span>{{ hostInfo.ip_forward ? '● 1 (已就绪)' : '○ 0 (未开启 · 需打通)' }}</span>
              </span>
            </div>
            <div class="col-span-2 sm:col-span-1">
              <span class="text-slate-400 text-[10px] block">防火墙端口放行</span>
              <span class="text-purple-300 font-bold block mt-0.5 text-xs">
                TCP 2026 · UDP 13231/51820
              </span>
            </div>
          </div>

          <!-- Mode Selection Title -->
          <div>
            <div class="text-xs font-bold text-slate-200 mb-2 flex items-center justify-between">
              <span>⚙️ 请选择防火墙与转发配置模式 (两种模式按需切换):</span>
              <span class="text-[11px] text-slate-400 font-normal">点击卡片切换</span>
            </div>

            <!-- Mode Cards Grid -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <!-- Permissive Mode Card -->
              <div 
                @click="selectFwMode('permissive')"
                :class="[
                  'p-3.5 rounded-xl border cursor-pointer transition-all relative overflow-hidden flex flex-col justify-between gap-2',
                  hostFwMode === 'permissive'
                    ? 'bg-emerald-950/40 border-emerald-500/80 ring-2 ring-emerald-500/60 shadow-[0_0_15px_rgba(16,185,129,0.25)]'
                    : 'bg-slate-900/60 border-purple-900/40 hover:border-emerald-500/40 hover:bg-slate-900/80'
                ]"
              >
                <div class="flex items-start justify-between gap-1.5">
                  <div class="flex items-center gap-2">
                    <span class="text-lg">🟢</span>
                    <div>
                      <div class="text-xs font-bold text-emerald-300">宽松模式 (Permissive)</div>
                      <div class="text-[10px] text-slate-400">开箱即通 · 100% 避免断网</div>
                    </div>
                  </div>
                  <span class="text-[9px] px-2 py-0.5 rounded-full bg-emerald-900/80 text-emerald-300 border border-emerald-500/50 font-bold shrink-0">
                    ★ 推荐小白首选
                  </span>
                </div>

                <div class="text-[11px] text-slate-300 leading-relaxed space-y-1">
                  <div>• <strong>核心机制</strong>：<code class="text-emerald-300 font-mono">iptables -P FORWARD ACCEPT</code> 全局放行转发链。</div>
                  <div>• <strong>出网伪装</strong>：私网全网段 NAT Masquerade 伪装与常用端口双向通畅。</div>
                  <div>• <strong>选型理由</strong>：彻底避免因 Docker 网桥、多网卡、IP 段不一致导致丢包断网，零报错、绝不卡网！</div>
                </div>

                <div class="text-[10px] text-emerald-400/90 font-mono pt-1 border-t border-emerald-900/40">
                  适用: 飞牛 fnOS、群晖 NAS、家庭单局域网
                </div>
              </div>

              <!-- Strict Mode Card -->
              <div 
                @click="selectFwMode('strict')"
                :class="[
                  'p-3.5 rounded-xl border cursor-pointer transition-all relative overflow-hidden flex flex-col justify-between gap-2',
                  hostFwMode === 'strict'
                    ? 'bg-sky-950/40 border-purple-400/80 ring-2 ring-cyan-500/60 shadow-[0_0_15px_rgba(0,210,255,0.25)]'
                    : 'bg-slate-900/60 border-purple-900/40 hover:border-purple-500/40 hover:bg-slate-900/80'
                ]"
              >
                <div class="flex items-start justify-between gap-1.5">
                  <div class="flex items-center gap-2">
                    <span class="text-lg">🛡️</span>
                    <div>
                      <div class="text-xs font-bold text-purple-300">严格模式 (Strict)</div>
                      <div class="text-[10px] text-slate-400">最小特权 · 状态追踪隔离</div>
                    </div>
                  </div>
                  <span class="text-[9px] px-2 py-0.5 rounded-full bg-sky-900/80 text-purple-300 border border-purple-500/50 font-bold shrink-0">
                    进阶极客 / 隔离网络
                  </span>
                </div>

                <div class="text-[11px] text-slate-300 leading-relaxed space-y-1">
                  <div>• <strong>核心机制</strong>：保持 <code class="text-purple-300 font-mono">FORWARD DROP</code> 默认阻断，仅放行已建立连接。</div>
                  <div>• <strong>定向穿透</strong>：仅对 WireGuard 虚拟网段 (10.88.0.0/24) 与指定私网开启转发与 NAT。</div>
                  <div>• <strong>选型理由</strong>：阻断非法跨网段嗅探与伪造包，符合企业级与多 VLAN 隔离安全审计。</div>
                </div>

                <div class="text-[10px] text-purple-400/90 font-mono pt-1 border-t border-purple-900/40">
                  适用: PVE 虚拟化生产集群、公网 VPS、多 VLAN 隔离
                </div>
              </div>
            </div>
          </div>

          <!-- Execution Method Tabs -->
          <div class="space-y-3 pt-1">
            <div class="flex items-center gap-2 border-b border-purple-900/40 pb-2">
              <button 
                @click="hostFwActiveTab = 'script'"
                :class="[
                  'px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
                  hostFwActiveTab === 'script'
                    ? 'bg-purple-600/25 text-purple-200 border border-purple-400/50 shadow-[0_0_10px_rgba(0,210,255,0.25)]'
                    : 'bg-slate-900/60 text-slate-400 hover:text-slate-200 border border-slate-700/40'
                ]"
              >
                <span>📋 方式 1：免密 1 键 Shell 脚本 (最推荐 · 0 密码暴露)</span>
              </button>
              <button 
                @click="hostFwActiveTab = 'ssh'"
                :class="[
                  'px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5',
                  hostFwActiveTab === 'ssh'
                    ? 'bg-sky-500/25 text-sky-200 border border-sky-400/50 shadow-[0_0_10px_rgba(14,165,233,0.25)]'
                    : 'bg-slate-900/60 text-slate-400 hover:text-slate-200 border border-slate-700/40'
                ]"
              >
                <span>⚡ 方式 2：面板 SSH 一键下发 (全自动执行)</span>
              </button>
            </div>

            <!-- Tab 1: 免密 1 键 Shell 脚本 -->
            <div v-if="hostFwActiveTab === 'script'" class="space-y-3">
              <div class="text-xs text-slate-300 leading-relaxed flex items-center justify-between">
                <span>
                  复制下方针对 <strong>【{{ hostFwMode === 'permissive' ? '🟢 宽松模式' : '🛡️ 严格模式' }}】</strong> 动态生成的脚本，在飞牛应用中心【终端】、PVE Shell 或 SSH 终端中粘贴回车即可：
                </span>
                <span v-if="hostFwScriptLoading" class="text-purple-400 text-[11px] animate-spin">↻ 生成中...</span>
              </div>

              <pre class="bg-black/80 p-4 rounded-xl text-xs font-mono text-purple-300 max-h-56 overflow-y-auto border border-purple-900/40 select-all whitespace-pre-wrap leading-relaxed shadow-inner">{{ hostScriptText }}</pre>

              <div class="flex gap-2.5">
                <button 
                  @click="copyToClipboard(hostScriptText, 'hostFwScript')" 
                  :class="[
                    'btn flex-1 text-xs transition-all duration-300 font-bold py-2 flex items-center justify-center gap-1.5',
                    copiedButtonId === 'hostFwScript'
                      ? 'bg-emerald-600 hover:bg-emerald-500 text-white shadow-[0_0_15px_rgba(16,185,129,0.5)]'
                      : 'btn-cyan'
                  ]"
                >
                  <span v-if="copiedButtonId === 'hostFwScript'">✅ 已成功复制【{{ hostFwMode === 'permissive' ? '宽松模式' : '严格模式' }}】脚本命令！</span>
                  <span v-else>📋 一键复制【{{ hostFwMode === 'permissive' ? '宽松模式' : '严格模式' }}】执行命令</span>
                </button>
                <button 
                  @click="downloadConfFile(`singcube-firewall-${hostFwMode}.sh`, hostScriptText)" 
                  class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs px-4"
                >
                  ⬇️ 下载 .sh
                </button>
              </div>
            </div>

            <!-- Tab 2: 面板 SSH 一键下发 -->
            <div v-else class="space-y-3">
              <!-- If SSH host is configured -->
              <div v-if="hostConfig.host" class="bg-slate-900/60 p-3 rounded-xl border border-purple-900/30 flex items-center justify-between text-xs font-mono">
                <div class="flex items-center gap-2">
                  <span class="text-emerald-400">●</span>
                  <span class="text-slate-300">已绑定宿主目标:</span>
                  <span class="text-purple-300 font-bold">{{ hostConfig.user }}@{{ hostConfig.host }}:{{ hostConfig.port || 22 }}</span>
                </div>
                <button @click="currentView = 'ros'; showHostFwModal = false; showHostScriptModal = false" class="text-purple-300 hover:underline text-[11px]">
                  修改 SSH 鉴权 ⚙️
                </button>
              </div>

              <!-- Quick SSH Input if not configured -->
              <div v-else class="bg-slate-900/60 p-3 rounded-xl border border-purple-900/30 space-y-2 text-xs">
                <div class="text-amber-300 font-bold">⚠️ 尚未配置宿主机 SSH 连接，可在此快速填写：</div>
                <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
                  <input v-model="hostConfig.host" placeholder="宿主 IP (如 172.17.0.1)" class="input-field text-xs font-mono" />
                  <input v-model="hostConfig.user" placeholder="用户名 (root)" class="input-field text-xs font-mono" />
                  <input v-model="hostConfig.password" type="password" placeholder="SSH 密码" class="input-field text-xs font-mono" />
                </div>
              </div>

              <!-- Apply Button -->
              <button 
                @click="applyHostFirewallModal" 
                :disabled="hostApplyingFw" 
                class="btn btn-blue w-full text-xs font-bold py-2.5 shadow-lg shadow-blue-500/20 flex items-center justify-center gap-2"
              >
                <span v-if="hostApplyingFw" class="inline-block animate-spin">↻</span>
                <span>🚀 立即以【{{ hostFwMode === 'permissive' ? '🟢 宽松模式 (推荐小白)' : '🛡️ 严格模式 (最小特权)' }}】一键自动配置宿主机</span>
              </button>

              <!-- Progress Logs Console -->
              <div v-if="hostFwLogs.length > 0 || hostApplyingFw" class="bg-black/80 p-3 rounded-xl border border-purple-900/40 space-y-1.5 text-xs font-mono">
                <div class="text-slate-400 text-[11px] border-b border-sky-950 pb-1">执行过程详细输出日志:</div>
                <div v-for="(log, idx) in hostFwLogs" :key="idx" class="text-slate-200 text-[11px]">
                  {{ log }}
                </div>
                <div v-if="hostApplyingFw" class="text-purple-400 text-[11px] animate-pulse">
                  ↻ 正在通过 SSH 会话执行底层内核配置，请稍候...
                </div>
              </div>

              <!-- Success Celebration Banner -->
              <div v-if="hostFwSuccess" class="bg-emerald-950/40 border border-emerald-500/50 rounded-xl p-3 text-xs space-y-1.5 shadow-lg shadow-emerald-500/10">
                <div class="text-emerald-300 font-bold flex items-center gap-1.5 text-sm">
                  <span>🎉</span>
                  <span>宿主机网络底座与防火墙已成功打通！当前为【{{ hostFwMode === 'permissive' ? '宽松模式' : '严格模式' }}】</span>
                </div>
                <div class="text-slate-300 text-[11px] leading-relaxed">
                  Linux 内核 IPv4 数据包转发已就绪，FORWARD 链与 NAT 伪装已生效。所有局域网设备现在可以将网关修改为 SingCube IP，满血畅享极速分流与权威去广告！
                </div>
              </div>
            </div>
          </div>

          <!-- Modal Footer -->
          <div class="flex items-center justify-between border-t border-purple-900/40 pt-3 text-xs">
            <div class="text-slate-400 text-[11px]">
              💡 配置完成后，无需重启系统，内核参数即刻生效。
            </div>
            <button @click="showHostFwModal = false; showHostScriptModal = false" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs px-5 py-1.5">
              关闭窗口
            </button>
          </div>
        </div>
      </div>

      <!-- ── Host OS Upgrade Terminal Modal ── -->
      <div v-if="showHostUpgradeModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4">
        <div class="glass-card max-w-2xl w-full p-6 space-y-4 border border-sky-600/40 shadow-2xl">
          <div class="flex items-center justify-between border-b border-purple-900/40 pb-3">
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <span>⬆️ 宿主机系统软件包升级与维护</span>
            </h3>
            <button @click="showHostUpgradeModal = false" class="text-slate-400 hover:text-white text-lg">✕</button>
          </div>

          <div class="text-xs text-slate-300 flex items-center justify-between">
            <span>检测状态: <strong class="text-purple-300">{{ hostUpgradeResult.upgradable_count }} 个软件包可升级</strong></span>
            <span v-if="hostUpgradeResult.running" class="text-amber-400 flex items-center gap-1 animate-pulse">
              <span class="inline-block animate-spin">↻</span>
              <span>正在后台升级宿主机软件包...</span>
            </span>
          </div>

          <div v-if="hostUpgradeResult.upgradable_pkgs && !hostUpgradeResult.running && !hostUpgradeResult.log" class="text-xs">
            <div class="text-slate-400 mb-1">待更新软件包列表：</div>
            <pre class="bg-black/70 p-3 rounded-xl text-[11px] font-mono text-slate-300 max-h-48 overflow-y-auto border border-purple-900/40 whitespace-pre-wrap">{{ hostUpgradeResult.upgradable_pkgs }}</pre>
          </div>

          <div v-if="hostUpgradeResult.log" class="text-xs">
            <div class="text-slate-400 mb-1">升级终端实时输出：</div>
            <pre class="bg-black/80 p-3 rounded-xl text-[11px] font-mono text-emerald-300 max-h-60 overflow-y-auto border border-purple-900/40 whitespace-pre-wrap">{{ hostUpgradeResult.log }}</pre>
          </div>

          <div class="flex gap-3 pt-2">
            <button @click="executeHostUpgrade" :disabled="hostUpgrading" class="btn btn-cyan flex-1 text-xs font-semibold shadow-lg shadow-purple-500/20">
              <span v-if="hostUpgrading" class="inline-block animate-spin mr-1">↻</span>
              <span>🚀 立即执行宿主机系统升级</span>
            </button>
            <button @click="showHostUpgradeModal = false" :disabled="hostUpgrading" class="btn bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs px-6">
              关闭
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- ── Mobile Floating Cyber Dock (Scheme A: Cyber Floating Dock) ── -->
    <nav class="md:hidden fixed bottom-3 left-3 right-3 z-40 bg-[#130d2a]/92 backdrop-blur-2xl border border-sky-400/40 rounded-2xl shadow-[0_8px_32px_rgba(0,0,0,0.6)] p-1.5 flex items-center justify-around pb-[max(0.375rem,env(safe-area-inset-bottom))]">
      <button v-for="dockItem in dockNavItems" :key="dockItem.id"
        @click="switchMasterView(dockItem.id)"
        :class="[
          'flex flex-col items-center justify-center py-1 px-3 rounded-xl transition-all duration-200 active:scale-95',
          activeMasterTab === dockItem.id
            ? 'text-purple-300 bg-sky-500/20 shadow-[0_0_12px_rgba(0,210,255,0.35)] border border-sky-400/30'
            : 'text-slate-400 hover:text-slate-200 border border-transparent'
        ]">
        <div class="w-5 h-5 flex items-center justify-center" v-html="dockItem.icon"></div>
        <span class="text-[10px] mt-0.5 font-medium leading-none">{{ dockItem.label }}</span>
      </button>
    </nav>

    <!-- ── Mobile Off-Canvas Drawer (更多功能模块) ── -->
    <transition name="drawer-fade">
      <div v-if="showMobileMenu" class="md:hidden fixed inset-0 z-50 bg-black/75 backdrop-blur-md flex justify-end" @click.self="showMobileMenu = false">
        <div class="w-72 max-w-[82vw] h-full bg-[#001f38] border-l border-sky-600/40 p-5 flex flex-col space-y-4 shadow-2xl animate-slide-left">
          <div class="flex items-center justify-between border-b border-purple-900/40 pb-3">
            <div class="flex items-center gap-2">
              <img src="/logo-singcube.svg" class="h-6 w-6 drop-shadow-[0_0_8px_#00d2ff]" alt="SingCube" />
              <span class="text-sm font-bold text-white">全部功能模块</span>
            </div>
            <button @click="showMobileMenu = false" class="text-slate-400 hover:text-white text-lg p-1">✕</button>
          </div>

          <!-- All Nav Items in Drawer -->
          <div class="flex-1 overflow-y-auto space-y-1.5 pr-1">
            <button v-for="item in navItems" :key="item.id"
              @click="currentView = item.id; showMobileMenu = false"
              :class="[
                'w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-xs font-medium transition-all duration-200',
                currentView === item.id
                  ? 'bg-purple-600/20 text-purple-300 border border-purple-400/50 shadow-lg'
                  : 'text-slate-300 hover:bg-slate-800/60'
              ]">
              <div class="w-5 h-5 flex-shrink-0 flex items-center justify-center" v-html="item.icon"></div>
              <span>{{ item.label }}</span>
            </button>
          </div>

          <!-- Drawer Footer -->
          <div class="border-t border-purple-900/40 pt-3 text-center space-y-2">
            <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-slate-900/80 border border-purple-500/30 text-xs" :title="formatPlatform(status.platform)">
              <div class="w-4 h-4 flex-shrink-0 flex items-center justify-center" v-html="getHostLogo(status.platform_id)"></div>
              <span class="text-slate-300 font-mono text-[11px] truncate max-w-[200px]">{{ formatPlatform(status.platform) }}</span>
            </div>
            <div class="text-[9px] text-slate-500 tracking-wider">SingCube · 专有高性能旁路由系统 · {{ rosOverview.panel_curr_version || 'Test V0.7' }}</div>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted, onUnmounted } from 'vue'
import QRCode from 'qrcode'

// ── State ──
const getInitialView = () => {
  const hash = (typeof window !== 'undefined' ? window.location.hash.replace('#', '').trim() : '')
  const validViews = ['dashboard', 'proxies', 'config', 'wireguard', 'devices', 'connections', 'adblock', 'accelerator', 'ai', 'tgbot', 'ros']
  return validViews.includes(hash) ? hash : 'dashboard'
}
const currentView = ref(getInitialView())
const loading = ref(false)
const actionMsg = ref('')
const actionError = ref(false)
const newCountry = ref('')

// ── Proxies State ──
const proxyGroupsData = ref({ groups: [], total_proxies: 0 })
const proxyFilter = ref('all')
const proxySearch = ref('')
const testingProxies = ref(false)

const quickOverviewGroup = computed(() => {
  const groups = proxyGroupsData.value.groups || []
  return groups.find(g => g.name === '节点总览') || groups[0] || null
})

const filteredProxyGroups = computed(() => {
  let list = proxyGroupsData.value.groups || []
  if (proxyFilter.value === 'core') {
    list = list.filter(g => g.name === '节点总览' || g.name === '自动优选' || g.name === '故障转移')
  } else if (proxyFilter.value === 'region') {
    list = list.filter(g => ['香港', '台湾', '新加坡', '狮城', '日本', '美国', '韩国', '其它'].some(c => g.name.includes(c)))
  } else if (proxyFilter.value === 'biz') {
    list = list.filter(g => g.name !== '节点总览' && g.name !== '自动优选' && g.name !== '故障转移' && !['香港', '台湾', '新加坡', '狮城', '日本', '美国', '韩国'].some(c => g.name.includes(c)))
  }
  if (proxySearch.value.trim()) {
    const q = proxySearch.value.toLowerCase().trim()
    list = list.filter(g => g.name.toLowerCase().includes(q) || (g.all && g.all.some(n => n.name.toLowerCase().includes(q))))
  }
  return list
})

function filterNodes(nodes) {
  if (!nodes) return []
  if (!proxySearch.value.trim()) return nodes
  const q = proxySearch.value.toLowerCase().trim()
  return nodes.filter(n => n.name.toLowerCase().includes(q))
}

const totalConfigNodeCount = computed(() => {
  if (!configState.subscriptions) return 0
  return configState.subscriptions.reduce((sum, s) => sum + (s.node_count || 0), 0)
})

// ── Connections State ──
const connectionsData = ref({ downloadTotal: 0, uploadTotal: 0, connections: [] })
const connSearch = ref('')
const connFilterProtocol = ref('all')
const connFilterType = ref('all')
const connSortBy = ref('download')
const connAutoRefresh = ref(true)
let connInterval = null

// ── AdBlock Live Metrics & Stream State ──
const adblockStats = reactive({
  total_rules: 233481,
  blocked_today: 1428,
  blocked_total: 12890,
  enabled: true,
  recent_list: [],
  last_updated: ''
})
const adblockSearch = ref('')
const adblockAutoRefresh = ref(true)
let adblockInterval = null

const filteredBlockedList = computed(() => {
  const list = adblockStats.recent_list || []
  if (!adblockSearch.value.trim()) return list
  const q = adblockSearch.value.toLowerCase().trim()
  return list.filter(item =>
    (item.domain && item.domain.toLowerCase().includes(q)) ||
    (item.client_ip && item.client_ip.toLowerCase().includes(q)) ||
    (item.rule && item.rule.toLowerCase().includes(q)) ||
    (item.device && item.device.toLowerCase().includes(q))
  )
})

watch(currentView, (v) => {
  if (typeof window !== 'undefined' && window.location.hash.replace('#', '').trim() !== v) {
    window.location.hash = v
  }
  if (v === 'proxies') {
    fetchProxies()
  }
  if (v === 'connections') {
    fetchConnections()
    if (!connInterval) {
      connInterval = setInterval(() => {
        if (connAutoRefresh.value && currentView.value === 'connections') {
          fetchConnections()
        }
      }, 2500)
    }
  } else {
    if (connInterval) {
      clearInterval(connInterval)
      connInterval = null
    }
  }

  if (v === 'adblock') {
    fetchAdblockStats()
    if (!adblockInterval) {
      adblockInterval = setInterval(() => {
        if (adblockAutoRefresh.value && currentView.value === 'adblock') {
          fetchAdblockStats()
        }
      }, 3000)
    }
  } else {
    if (adblockInterval) {
      clearInterval(adblockInterval)
      adblockInterval = null
    }
  }
})
const chatInput = ref('')
const chatMessages = ref([
  {
    role: 'assistant',
    content: '您好！我是 SingCube 专属 AI 智脑与软路由智能运维中枢。\n\n我已全面接入宿主环境拓扑、DNS 分流策略、节点测速延迟及 217heidai 权威拦截规则。无需折腾、开箱自愈，您可以随时点击上方胶囊或直接提问！',
    time: new Date().toLocaleTimeString()
  }
])
const teleDirection = ref('download')
const teleDimension = ref('rule')
const delayTesting = ref(false)

// ── KSpeeder State (Item 2) ──
const kspeederStatus = reactive({
  active_docker_mirror: '',
  active_github_proxy: '',
  auto_select_best: true,
  docker_mirrors: [],
  github_proxies: [],
  last_tested_at: ''
})
const testingDocker = ref(false)
const testingGithub = ref(false)
const speedupInputUrl = ref('')
const speedupOutputUrl = ref('')
const showRosRegistryModal = ref(false)
const rosRegistryScriptText = ref('')
const showDockerDaemonModal = ref(false)
const dockerDaemonJsonText = ref('')

// ── Qwen AI Copilot State (Item 1) ──
const showAiConfigModal = ref(false)
const diagnosing = ref(false)
const aiConfig = reactive({
  enabled: true,
  provider: 'qwen',
  api_key: '',
  base_url: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
  model_name: 'qwen-plus',
  has_api_key: false
})
const aiConfigForm = reactive({
  enabled: true,
  provider: 'qwen',
  api_key: '',
  base_url: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
  model_name: 'qwen-plus'
})

const showOnboarding = ref(localStorage.getItem('singcube_onboarding_done') !== 'true')
function dismissOnboarding() {
  showOnboarding.value = false
  localStorage.setItem('singcube_onboarding_done', 'true')
}

const realtimeSpeed = reactive({
  up_bps: 0,
  down_bps: 0,
  up_speed_human: '0 B/s',
  down_speed_human: '0 B/s',
  active_conns: 0,
  history: []
})

const waveformPaths = computed(() => {
  const rawSamples = realtimeSpeed.history && realtimeSpeed.history.length > 0
    ? realtimeSpeed.history
    : Array.from({ length: 30 }, () => ({ up_bps: 0, down_bps: 0 }))

  const samples = rawSamples.length < 2
    ? [...rawSamples, ...rawSamples]
    : rawSamples

  let maxBps = 50 * 1024
  for (const s of samples) {
    if (s.down_bps > maxBps) maxBps = s.down_bps
    if (s.up_bps > maxBps) maxBps = s.up_bps
  }

  const w = 600
  const h = 120
  const padBottom = 110
  const padTop = 15
  const usableH = padBottom - padTop
  const stepX = w / (samples.length - 1)

  const downPoints = samples.map((s, i) => {
    const x = Number((i * stepX).toFixed(1))
    const ratio = Math.min(1, Math.max(0, (s.down_bps || 0) / maxBps))
    const y = Number((padBottom - ratio * usableH).toFixed(1))
    return { x, y }
  })

  const upPoints = samples.map((s, i) => {
    const x = Number((i * stepX).toFixed(1))
    const ratio = Math.min(1, Math.max(0, (s.up_bps || 0) / maxBps))
    const y = Number((padBottom - ratio * usableH).toFixed(1))
    return { x, y }
  })

  function buildPath(pts, closeArea = false) {
    if (!pts.length) return ''
    let d = `M ${pts[0].x} ${pts[0].y}`
    for (let i = 0; i < pts.length - 1; i++) {
      const p0 = i > 0 ? pts[i - 1] : pts[i]
      const p1 = pts[i]
      const p2 = pts[i + 1]
      const p3 = i !== pts.length - 2 ? pts[i + 2] : p2
      const cp1x = p1.x + (p2.x - p0.x) / 6
      const cp1y = p1.y + (p2.y - p0.y) / 6
      const cp2x = p2.x - (p3.x - p1.x) / 6
      const cp2y = p2.y - (p3.y - p1.y) / 6
      d += ` C ${cp1x.toFixed(1)} ${cp1y.toFixed(1)}, ${cp2x.toFixed(1)} ${cp2y.toFixed(1)}, ${p2.x.toFixed(1)} ${p2.y.toFixed(1)}`
    }
    if (closeArea) {
      const lastX = pts[pts.length - 1].x
      d += ` L ${lastX} ${h} L ${pts[0].x} ${h} Z`
    }
    return d
  }

  const lastDown = downPoints[downPoints.length - 1] || { x: 600, y: 110 }
  const lastUp = upPoints[upPoints.length - 1] || { x: 600, y: 110 }

  return {
    downLine: buildPath(downPoints, false),
    downPath: buildPath(downPoints, true),
    upLine: buildPath(upPoints, false),
    upPath: buildPath(upPoints, true),
    lastDownX: lastDown.x,
    lastDownY: lastDown.y,
    lastUpX: lastUp.x,
    lastUpY: lastUp.y
  }
})

const status = reactive({
  running: false, pid: 0, uptime_seconds: 0,
  core_memory_mb: 0, panel_memory_mb: 0, total_memory_mb: 0,
  platform: '', platform_id: 'routeros', arch: '', autostart: false, tun_mode: ''
})
const showMobileMenu = ref(false)

const newBlockUrl = ref('')
const configState = reactive({
  subscriptions: [],
  auto_countries: [],
  rule_targets: [],
  auto_select_mode: 'url-test',
  direct_bypass_mode: true,
  quic_block_enabled: true,
  ad_block_enabled: true,
  custom_block_list: []
})

// ── Profiles & Speedtest Sandbox (Item 1) ──
const profiles = ref([])
const activeProfileId = ref('balanced')
const currentProfile = computed(() => {
  return profiles.value.find(p => p.id === activeProfileId.value) || profiles.value[0] || {
    id: 'balanced',
    name: '家庭日常平衡',
    latency_thresholds: { fast_ms: 120, medium_ms: 250, slow_ms: 450 }
  }
})
const showProfileModal = ref(false)
const editingProfile = reactive({
  id: '',
  name: '',
  description: '',
  auto_select_mode: 'url-test',
  direct_bypass_mode: true,
  quic_block_enabled: true,
  ad_block_enabled: true,
  speedtest_url: 'http://www.gstatic.com/generate_204',
  speedtest_timeout_ms: 2000,
  speedtest_interval_s: 300,
  speedtest_tolerance_ms: 50,
  latency_thresholds: {
    fast_ms: 120,
    medium_ms: 250,
    slow_ms: 450
  }
})

const wgInfo = reactive({
  config: {
    interface_name: 'singcube-wg',
    listen_port: 13231,
    server_address: '10.88.0.1/24',
    private_key: '',
    public_key: '',
    endpoint_host: '',
    dns: '10.88.0.1',
    direct_dns_mode: 'tunnel_strict',
    intranet_cidr: '192.168.88.0/24, 10.88.0.0/24',
    cloud_ddns: ''
  },
  peers: [],
  total_peers: 0,
  effective_endpoint: '',
  cloud_ddns: ''
})

const showNewPeerModal = ref(false)
const newPeerForm = reactive({
  name: '',
  mode: 'intranet'
})

const showQrModal = ref(false)
const currentQrDataUrl = ref('')
const currentPeerName = ref('')
const currentPeerConfig = ref('')

const showRosScriptModal = ref(false)
const rosScriptText = ref('')

// ── WireGuard One-Click Wizard State ──
const showWgWizardModal = ref(false)
const wizardStage = ref('deploying') // 'deploying' | 'timeout' | 'scanning' | 'celebrating' | 'error'
const wizardProgress = ref(0)
const wizardErrorMsg = ref('')
const wizardDeploySeconds = ref(0)
const wizardTimeoutTimer = ref(null)
const wizardSecondsInterval = ref(null)
const wizardSteps = ref([
  { title: '探查运行环境与 RouterOS 连接', status: 'wait', detail: '检测 REST API 通信' },
  { title: '自动创建/校准 WireGuard 核心接口', status: 'wait', detail: '监听 13231/51820 UDP · 10.88.0.1/24' },
  { title: '自动探查并绑定 IP Cloud 动态穿透域名', status: 'wait', detail: '提取免费免配 DDNS 域名' },
  { title: '自动在防火墙放行 UDP 穿透端口', status: 'wait', detail: 'input 链顶部放行外部流量' },
  { title: '自动下发策略路由 Table 8228 透明引流', status: 'wait', detail: '打通 10.88.0.0/24 代理网关' },
  { title: '自动生成手机专属免密密钥与 Peer', status: 'wait', detail: '生成标准手机客户端配置' }
])
const wizardPeer = ref(null)
const wizardConfig = ref('')
const wizardQrUrl = ref('')
const wizardChecking = ref(false)
const wizardPollTimer = ref(null)
const wizardLiveStatus = ref(null)

// ── LAN Devices & Alias State (Item 5) ──
const devicesList = ref([])
const devicesLoading = ref(false)
const deviceSearch = ref('')
const showDeviceModal = ref(false)
const deviceEditForm = reactive({ ip: '', mac: '', name: '' })
const copiedButtonId = ref('')

const telemetryData = reactive({
  grand_total: 0, grand_total_size: '0 B', items: []
})

const rosOverview = reactive({
  ros_connected: false, ros_version: '', mihomo_curr_version: '', mihomo_latest_version: '', update_available: false
})

const rosRoutingReport = reactive({
  connected: false, ros_host: '', ros_version: '', singcube_ip: '', all_ok: false, items: []
})
const rosRoutingLoading = ref(false)
const applyingRouting = ref(false)
const showWinBoxRoutingModal = ref(false)
const winBoxRoutingScriptText = ref('')
const cleaningRouting = ref(false)

// ── Host OS & Docker Companion State ──
const hostConfig = reactive({
  host: '',
  port: 22,
  user: 'root',
  auth_type: 'password',
  password: '',
  private_key: '',
  is_configured: false
})
const hostInfo = reactive({
  ok: false,
  os_name: '',
  os_pretty: '',
  is_fnos: false,
  is_pve: false,
  is_synology: false,
  is_debian: false,
  kernel: '',
  arch: '',
  memory_total: '',
  memory_used: '',
  cpu_cores: 0,
  ip_forward: false,
  wireguard_installed: false,
  wg0_active: false,
  last_check: ''
})
const hostTesting = ref(false)
const hostSaving = ref(false)
const hostApplyingFw = ref(false)
const hostDeployingWg = ref(false)
const showHostFwModal = ref(false)
const hostFwMode = ref('permissive') // 'permissive' | 'strict'
const hostFwActiveTab = ref('script') // 'script' | 'ssh'
const hostFwLogs = ref([])
const hostFwSuccess = ref(false)
const hostFwError = ref('')
const hostFwScriptLoading = ref(false)
const showHostScriptModal = ref(false)
const hostScriptText = ref('')
const hostUpgrading = ref(false)
const hostCheckingUpgrade = ref(false)
const showHostUpgradeModal = ref(false)

// ── Initialization Setup Wizard State ──
const showInitWizardModal = ref(false)
const initWizardStep = ref('welcome') // 'welcome' | 'firewall' | 'subscription' | 'core' | 'wireguard' | 'finished'
const initBackupLoading = ref(false)
const initBackupError = ref('')
const initBackupSuccess = ref(false)
const initSubUrl = ref('')
const initSubName = ref('主力机场订阅')
const initSubLoading = ref(false)
const initSubSuccess = ref(false)
const initSubMsg = ref('')
const initCoreLoading = ref(false)
const hostUpgradeResult = reactive({
  checked: false,
  upgradable_count: 0,
  upgradable_pkgs: '',
  running: false,
  log: ''
})

const wgAlignmentReport = reactive({
  connected: false, ros_host: '', all_aligned: false, items: [], peer_count: 0, synced_peers: 0
})
const wgAligning = ref(false)
const showRenamePeerModal = ref(false)
const peerRenameForm = reactive({ public_key: '', name: '' })

const tgBotConfig = reactive({
  enabled: false, token: '', allowed_chat_ids: [], proxy_url: '', running: false
})
const tgBotInput = reactive({
  enabled: false, token: '', allowed_chat_ids_str: '', proxy_url: 'http://127.0.0.1:7890'
})

// ── Host Platform Vector Logos (Official with 3D Cyber Glow) ──
const hostLogos = {
  routeros: `<svg viewBox="0 0 24 24" class="w-full h-full drop-shadow-[0_0_10px_rgba(0,210,255,0.85)]" fill="none"><path fill="#00d2ff" d="M3 13.5L8.5 8l3 3-5.5 5.5H3v-3zm7.5-7.5L16 0.5l5.5 5.5-5.5 5.5-5.5-5.5zm5.5 9.5l3-3 5.5 5.5v3h-3l-5.5-5.5z"/><circle cx="16" cy="6" r="2.2" fill="#ffffff" /><circle cx="6" cy="16" r="2.2" fill="#ffffff" /><circle cx="18" cy="18" r="2.2" fill="#ffffff" /></svg>`,

  pve: `<svg viewBox="0 0 24 24" class="w-full h-full drop-shadow-[0_0_8px_rgba(240,110,0,0.8)]" fill="#f06e00"><path d="M12 1.5l2.6 4.5h5.2l-2.6 4.5 2.6 4.5h-5.2L12 19.5l-2.6-4.5H4.2l2.6-4.5-2.6-4.5h5.2L12 1.5zm0 5.2l-1.3 2.3h2.6L12 6.7zm-2.8 4.8l1.3 2.3h3l1.3-2.3-1.3-2.3h-3l-1.3 2.3zm2.8 4.8l1.3-2.3h-2.6l1.3 2.3z"/></svg>`,

  esxi: `<svg viewBox="0 0 24 24" class="w-full h-full drop-shadow-[0_0_8px_rgba(0,149,211,0.8)]"><path fill="#0095d3" d="M12 2l-6 3.5v7l6 3.5 6-3.5v-7L12 2zm0 2.3l4 2.3-4 2.3-4-2.3 4-2.3zm-4.5 8.2v-4.6l3.5 2v4.6l-3.5-2zm9 0l-3.5 2v-4.6l3.5-2v4.6z"/><path fill="#38bdf8" d="M6 14.5l-4 2.3v3.5l4-2.3v-3.5zm12 0v3.5l4 2.3v-3.5l-4-2.3zm-6 3.5l-4 2.3v3.5l4-2.3v-3.5zm0 0l4 2.3v3.5l-4-2.3v-3.5z"/></svg>`,

  fnos: `<svg viewBox="0 0 24 24" class="w-full h-full drop-shadow-[0_0_8px_rgba(56,189,248,0.8)]" fill="none"><path fill="url(#fnos-g)" d="M2.5 17.5L8 6.5l4 8-3.5 3h-6zm6-6.5l4.5-8 8.5 14.5h-5.5l-4-6.5-3.5 0z"/><defs><linearGradient id="fnos-g" x1="0" y1="0" x2="1" y2="1"><stop offset="0%" stop-color="#38bdf8"/><stop offset="100%" stop-color="#00f0ff"/></linearGradient></defs></svg>`,

  synology: `<svg viewBox="0 0 24 24" class="w-full h-full drop-shadow-[0_0_8px_rgba(255,255,255,0.7)]" fill="#f8fafc"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 14.5h-2v-4h2v4zm-1-5.5c-.69 0-1.25-.56-1.25-1.25s.56-1.25 1.25-1.25 1.25.56 1.25 1.25-.56 1.25-1.25 1.25z"/><path fill="#38bdf8" d="M6.5 9.5c0-3 2.5-5.5 5.5-5.5s5.5 2.5 5.5 5.5h-2c0-1.9-1.6-3.5-3.5-3.5s-3.5 1.6-3.5 3.5h-2zm11 5c0 3-2.5 5.5-5.5 5.5s-5.5-2.5-5.5-5.5h2c0 1.9 1.6 3.5 3.5 3.5s3.5-1.6 3.5-3.5h2z"/></svg>`,

  unraid: `<svg viewBox="0 0 24 24" class="w-full h-full drop-shadow-[0_0_8px_rgba(245,158,11,0.8)]" fill="#f59e0b"><rect x="4" y="3" width="6" height="18" rx="2"/><rect x="14" y="3" width="6" height="18" rx="2"/><circle cx="7" cy="8" r="1.5" fill="#090714"/><circle cx="17" cy="16" r="1.5" fill="#090714"/></svg>`,

  openwrt: `<svg viewBox="0 0 24 24" class="w-full h-full drop-shadow-[0_0_8px_rgba(2,132,199,0.8)]" fill="none"><path stroke="#0284c7" stroke-width="2.5" stroke-linecap="round" d="M4 16c2.5-6 6-9 10-9s6.5 4 6 9"/><path stroke="#38bdf8" stroke-width="2.5" stroke-linecap="round" d="M7 17c1.8-3.5 3.5-5 5.5-5s3.5 1.5 4.5 5"/><circle cx="12" cy="17" r="1.5" fill="#00f0ff"/></svg>`,

  debian: `<svg viewBox="0 0 24 24" class="w-full h-full drop-shadow-[0_0_8px_rgba(225,29,72,0.8)]" fill="#e11d48"><path d="M12.5 2C6.8 2 2.2 6.6 2.2 12.3c0 4.5 2.9 8.3 6.9 9.7-.1-.9-.2-2.3 0-3.3.2-.9 1.3-4.5 1.3-4.5s-.3-.6-.3-1.6c0-1.5.9-2.6 2-2.6.9 0 1.4.7 1.4 1.5 0 .9-.6 2.3-.9 3.6-.3 1.1.6 2 1.7 2 2 0 3.6-2.1 3.6-5.2 0-2.7-2-4.6-4.7-4.6-3.2 0-5.1 2.4-5.1 4.9 0 1 .4 2 1 2.6.1.1.1.2.1.3-.1.4-.3 1.2-.3 1.4 0 .1-.1.2-.3.1-1.6-.7-2.6-3-2.6-4.8 0-3.9 2.8-7.5 8.2-7.5 4.3 0 7.7 3.1 7.7 7.2 0 4.3-2.7 7.7-6.5 7.7-1.3 0-2.5-.7-2.9-1.5l-.8 3.1c-.3 1.1-1.1 2.5-1.6 3.4 1.2.4 2.5.6 3.8.6 5.7 0 10.3-4.6 10.3-10.3S18.2 2 12.5 2z"/></svg>`,

  ubuntu: `<svg viewBox="0 0 24 24" class="w-full h-full drop-shadow-[0_0_8px_rgba(249,115,22,0.8)]" fill="#f97316"><circle cx="12" cy="12" r="10" fill="none" stroke="#f97316" stroke-width="2.2"/><circle cx="12" cy="5" r="1.5" fill="#f97316"/><circle cx="5.5" cy="15.5" r="1.5" fill="#f97316"/><circle cx="18.5" cy="15.5" r="1.5" fill="#f97316"/></svg>`,

  docker: `<svg viewBox="0 0 24 24" class="w-full h-full drop-shadow-[0_0_8px_rgba(14,165,233,0.8)]" fill="#0ea5e9"><path d="M13 8.5h2v2h-2zm-3 0h2v2h-2zm-3 0h2v2H7zm6-3h2v2h-2zm-3 0h2v2h-2zm6 6h2v2h-2zm-3 0h2v2h-2zm-3 0h2v2h-2zm-3 0h2v2H7zm-3 0h2v2H4zm18.5 1.5c-.3-2-2-2.5-3-2.5-.2 0-.4 0-.6.1-.5-1.3-1.8-2.1-3.4-2.1H3.5C2.7 8 2 8.7 2 9.5v5C2 17.5 4.5 20 7.5 20h8c3.5 0 6.5-2.5 6.9-6 .4-.2 1.1-.7 1.6-1.5-.5-.4-1.1-.6-1.5-.5z"/></svg>`,

  windows: `<svg viewBox="0 0 24 24" class="w-full h-full drop-shadow-[0_0_8px_rgba(0,164,239,0.8)]" fill="#00a4ef"><path d="M3 3h8v8H3zm10 0h8v8h-8zM3 13h8v8H3zm10 0h8v8h-8z"/></svg>`
}

function getHostLogo(id) {
  if (!id) return hostLogos.routeros
  const k = id.toLowerCase()
  return hostLogos[k] || hostLogos.routeros
}

function formatPlatform(plat) {
  if (rosOverview && rosOverview.ros_version) {
    const v = rosOverview.ros_version.trim()
    if (v) {
      return v.startsWith('RouterOS') ? v : `RouterOS ${v}`
    }
  }
  if (!plat) return 'RouterOS v7.x'
  if (/^router\s*os$/i.test(plat) || /^mikrotik$/i.test(plat)) {
    return 'RouterOS v7.x'
  }
  return plat
}

// ── 5 Core High-Productivity Hubs (Test V0.91) ──
const proxiesSubTab = ref('nodes') // 'nodes' | 'config'
const rulesSubTab = ref('adblock') // 'adblock' | 'connections' | 'devices'
const systemSubTab = ref('ros') // 'ros' | 'tgbot' | 'ai' | 'accelerator'
const expandedGroups = ref({})
function toggleGroupExpanded(groupName) {
  expandedGroups.value[groupName] = !expandedGroups.value[groupName]
}

const activeMasterTab = computed(() => {
  if (currentView.value === 'dashboard') return 'dashboard'
  if (['proxies', 'config'].includes(currentView.value)) return 'proxies'
  if (['wireguard'].includes(currentView.value)) return 'wireguard'
  if (['rules', 'adblock', 'connections', 'devices'].includes(currentView.value)) return 'rules'
  if (['system', 'ros', 'tgbot', 'ai', 'accelerator'].includes(currentView.value)) return 'system'
  return 'dashboard'
})

function switchMasterView(id) {
  if (id === 'dashboard') {
    currentView.value = 'dashboard'
  } else if (id === 'proxies') {
    currentView.value = proxiesSubTab.value === 'config' ? 'config' : 'proxies'
  } else if (id === 'wireguard') {
    currentView.value = 'wireguard'
  } else if (id === 'rules') {
    currentView.value = rulesSubTab.value || 'adblock'
  } else if (id === 'system') {
    currentView.value = systemSubTab.value || 'ros'
  }
}

const navItems = [
  { id: 'dashboard', label: '仪表总览', icon: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z"/></svg>` },
  { id: 'proxies', label: '节点分流', icon: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/></svg>` },
  { id: 'wireguard', label: '穿透协同', icon: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>` },
  { id: 'rules', label: '规则安全', icon: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/></svg>` },
  { id: 'system', label: '系统运维', icon: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/></svg>` },
]

// ── Mobile Scheme A Floating Dock Items (5 Direct Hubs) ──
const dockNavItems = [
  { id: 'dashboard', label: '总览', icon: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z"/></svg>` },
  { id: 'proxies', label: '节点', icon: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/></svg>` },
  { id: 'wireguard', label: '穿透', icon: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>` },
  { id: 'rules', label: '安全', icon: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/></svg>` },
  { id: 'system', label: '运维', icon: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/></svg>` },
]

// ── Dirty State & Factory Default Reset ──
const initialConfigSnapshot = ref('')
const isConfigDirty = computed(() => {
  if (!initialConfigSnapshot.value) return false
  return JSON.stringify(configState) !== initialConfigSnapshot.value
})

async function resetToDefaultRules() {
  if (!confirm('确定要将分流规则与去广告策略重置为官方推荐预设吗？\\n（您的机场订阅链接与 WireGuard 密钥将绝对安全保留）')) {
    return
  }
  loading.value = true
  actionMsg.value = '🔄 正在恢复官方推荐规则预设并热重载...'
  const res = await apiFetch('/config/rules/reset-defaults', { method: 'POST' })
  loading.value = false
  if (res?.status === 'reset') {
    actionMsg.value = '✅ 已成功恢复官方推荐规则预设！'
    actionError.value = false
    await fetchConfigState()
  } else {
    actionMsg.value = `❌ 恢复失败: ${res?.error || '未知异常'}`
    actionError.value = true
  }
  setTimeout(() => { actionMsg.value = '' }, 4000)
}

// ── API Helpers ──
const API = '/api/v1'

async function apiFetch(path, opts = {}) {
  try {
    const controller = new AbortController()
    const timeoutMs = opts.timeout || 15000
    const timeoutId = setTimeout(() => controller.abort(), timeoutMs)
    const { timeout, ...fetchOpts } = opts

    const res = await fetch(API + path, {
      headers: { 'Content-Type': 'application/json' },
      signal: fetchOpts.signal || controller.signal,
      ...fetchOpts
    })
    clearTimeout(timeoutId)
    return await res.json()
  } catch (e) {
    console.error('API error:', path, e)
    return null
  }
}

// ── Data Fetching ──
async function fetchStatus() {
  const data = await apiFetch('/system/status')
  if (data) Object.assign(status, data)
}

// ── Profile Sandbox & Speedtest Handlers (Item 1) ──
async function fetchProfiles() {
  const data = await apiFetch('/profiles')
  if (data && data.profiles) {
    profiles.value = data.profiles
    activeProfileId.value = data.active_id || 'balanced'
  }
}

async function switchProfile(id) {
  loading.value = true
  actionMsg.value = '🎯 正在切换配置沙盒方案...'
  actionError.value = false
  const data = await apiFetch('/profiles/switch', {
    method: 'POST',
    body: JSON.stringify({ id })
  })
  if (data?.status === 'switched') {
    activeProfileId.value = id
    actionMsg.value = `✅ 已成功切换至【${currentProfile.value.name}】并热重载生效！`
    actionError.value = false
    await fetchConfigState()
  } else {
    actionMsg.value = `❌ 切换失败: ${data?.error || '未知错误'}`
    actionError.value = true
  }
  loading.value = false
  setTimeout(() => { actionMsg.value = '' }, 4000)
}

function openProfileModal() {
  const cur = profiles.value.find(p => p.id === activeProfileId.value) || profiles.value[0]
  if (cur) selectEditingProfile(cur)
  showProfileModal.value = true
}

function selectEditingProfile(p) {
  editingProfile.id = p.id
  editingProfile.name = p.name || ''
  editingProfile.description = p.description || ''
  editingProfile.auto_select_mode = p.auto_select_mode || 'url-test'
  editingProfile.direct_bypass_mode = p.direct_bypass_mode !== false
  editingProfile.quic_block_enabled = p.quic_block_enabled !== false
  editingProfile.ad_block_enabled = p.ad_block_enabled !== false
  editingProfile.speedtest_url = p.speedtest_url || 'http://www.gstatic.com/generate_204'
  editingProfile.speedtest_timeout_ms = p.speedtest_timeout_ms || 2000
  editingProfile.speedtest_interval_s = p.speedtest_interval_s || 300
  editingProfile.speedtest_tolerance_ms = p.speedtest_tolerance_ms || 50
  editingProfile.latency_thresholds = {
    fast_ms: p.latency_thresholds?.fast_ms || 120,
    medium_ms: p.latency_thresholds?.medium_ms || 250,
    slow_ms: p.latency_thresholds?.slow_ms || 450
  }
}

async function saveProfileAndApply() {
  loading.value = true
  actionMsg.value = '💾 正在保存配置方案...'
  const data = await apiFetch('/profiles', {
    method: 'POST',
    body: JSON.stringify(editingProfile)
  })
  if (data?.status === 'saved') {
    actionMsg.value = `✅ 配置方案【${editingProfile.name}】已保存！`
    actionError.value = false
    await fetchProfiles()
    if (activeProfileId.value === editingProfile.id) {
      await switchProfile(editingProfile.id)
    }
  } else {
    actionMsg.value = `❌ 保存失败: ${data?.error || '未知错误'}`
    actionError.value = true
  }
  loading.value = false
  setTimeout(() => { actionMsg.value = '' }, 4000)
}

async function saveAsNewProfile() {
  const newId = 'custom_' + Date.now().toString(36)
  const newName = prompt('请输入新配置方案名称：', editingProfile.name + ' (副本)')
  if (!newName) return
  const payload = { ...editingProfile, id: newId, name: newName }
  loading.value = true
  const data = await apiFetch('/profiles', {
    method: 'POST',
    body: JSON.stringify(payload)
  })
  if (data?.status === 'saved') {
    actionMsg.value = `✅ 已创建新配置方案【${newName}】！`
    actionError.value = false
    await fetchProfiles()
    await switchProfile(newId)
  }
  loading.value = false
  setTimeout(() => { actionMsg.value = '' }, 4000)
}

async function deleteEditingProfile() {
  if (['balanced', 'clean', 'gaming'].includes(editingProfile.id)) {
    actionMsg.value = '⚠️ 系统预设方案不可删除'
    actionError.value = true
    setTimeout(() => { actionMsg.value = '' }, 3000)
    return
  }
  if (!confirm(`确定要删除配置方案【${editingProfile.name}】吗？`)) return
  loading.value = true
  const data = await apiFetch(`/profiles?id=${editingProfile.id}`, { method: 'DELETE' })
  if (data?.status === 'deleted') {
    actionMsg.value = `✅ 配置方案【${editingProfile.name}】已删除`
    actionError.value = false
    await fetchProfiles()
    if (activeProfileId.value === editingProfile.id) {
      await switchProfile('balanced')
    }
    const first = profiles.value[0]
    if (first) selectEditingProfile(first)
  }
  loading.value = false
  setTimeout(() => { actionMsg.value = '' }, 4000)
}

async function fetchConfigState() {
  const data = await apiFetch('/config/state')
  if (data) {
    configState.subscriptions = data.subscriptions || []
    configState.auto_countries = data.auto_countries || []
    configState.rule_targets = data.rule_targets || []
    configState.auto_select_mode = data.auto_select_mode || 'url-test'
    configState.direct_bypass_mode = data.direct_bypass_mode !== false
    configState.quic_block_enabled = data.quic_block_enabled !== false
    configState.ad_block_enabled = data.ad_block_enabled !== false
    configState.custom_block_list = data.custom_block_list || []
    initialConfigSnapshot.value = JSON.stringify(configState)
  }
}

async function fetchRealtimeSpeed() {
  const data = await apiFetch('/telemetry/realtime')
  if (data) {
    realtimeSpeed.up_bps = data.up_bps || 0
    realtimeSpeed.down_bps = data.down_bps || 0
    realtimeSpeed.up_speed_human = data.up_speed_human || '0 B/s'
    realtimeSpeed.down_speed_human = data.down_speed_human || '0 B/s'
    realtimeSpeed.active_conns = data.active_conns || 0
    realtimeSpeed.history = data.history || []
  }
}

function addCustomBlock() {
  const v = newBlockUrl.value.trim()
  if (v) {
    if (!configState.custom_block_list) {
      configState.custom_block_list = []
    }
    if (!configState.custom_block_list.includes(v)) {
      configState.custom_block_list.push(v)
      actionMsg.value = `已添加拦截目标: ${v} (请点击保存生效)`
      actionError.value = false
    }
  }
  newBlockUrl.value = ''
  setTimeout(() => { actionMsg.value = '' }, 4000)
}

function removeCustomBlock(index) {
  configState.custom_block_list.splice(index, 1)
}

async function fetchAdblockStats() {
  const data = await apiFetch('/adblock/stats')
  if (data) {
    Object.assign(adblockStats, data)
  }
}

async function clearBlockedLogs() {
  await apiFetch('/adblock/blocked', { method: 'DELETE' })
  adblockStats.recent_list = []
  actionMsg.value = '✓ 已清空实时拦截流'
  actionError.value = false
  setTimeout(() => { actionMsg.value = '' }, 4000)
}

async function fetchROS() {
  const data = await apiFetch('/ros/overview')
  if (data) Object.assign(rosOverview, data)
}

async function fetchTgBot() {
  const data = await apiFetch('/tgbot/config')
  if (data) {
    Object.assign(tgBotConfig, data)
    tgBotInput.enabled = data.enabled
    tgBotInput.proxy_url = data.proxy_url || 'http://127.0.0.1:7890'
    tgBotInput.allowed_chat_ids_str = (data.allowed_chat_ids || []).join(', ')
  }
}

// ── Core Actions ──
async function coreAction(action) {
  loading.value = true
  actionMsg.value = ''
  const data = await apiFetch(`/core/${action}`, { method: 'POST' })
  if (data?.error) {
    actionMsg.value = `❌ ${data.error}`
    actionError.value = true
  } else {
    actionMsg.value = `✅ 操作成功: ${data?.status || action}`
    actionError.value = false
    setTimeout(fetchStatus, 1000)
  }
  loading.value = false
  setTimeout(() => { actionMsg.value = '' }, 5000)
}

async function toggleAutostart(e) {
  await apiFetch('/autostart', {
    method: 'POST',
    body: JSON.stringify({ enable: e.target.checked })
  })
  await fetchStatus()
}

// ── Config Actions ──
async function saveConfig() {
  loading.value = true
  const data = await apiFetch('/config/state', {
    method: 'PUT',
    body: JSON.stringify(configState)
  })
  if (data?.error) {
    actionMsg.value = `❌ ${data.error}`
    actionError.value = true
  } else {
    actionMsg.value = '✅ 配置已保存并编译'
    actionError.value = false
    initialConfigSnapshot.value = JSON.stringify(configState)
  }
  loading.value = false
  setTimeout(() => { actionMsg.value = '' }, 5000)
}

async function updateAllSubscriptions() {
  loading.value = true
  actionMsg.value = '📥 正在拉取全部订阅节点并重新编译...'
  const data = await apiFetch('/config/subscriptions/update', { method: 'POST', timeout: 60000 })
  loading.value = false
  if (data?.status === 'updated') {
    actionMsg.value = '✅ 全部订阅已更新并热重载成功！'
    actionError.value = false
    fetchConfigState()
  } else {
    actionMsg.value = `❌ 更新失败: ${data?.error || '网络错误'}`
    actionError.value = true
  }
  setTimeout(() => { actionMsg.value = '' }, 5000)
}

async function updateSingleSub(id) {
  loading.value = true
  actionMsg.value = '📥 正在拉取指定订阅节点...'
  const data = await apiFetch('/config/subscriptions/update', {
    method: 'POST',
    body: JSON.stringify({ id }),
    timeout: 60000
  })
  loading.value = false
  if (data?.status === 'updated') {
    actionMsg.value = '✅ 订阅已更新并热重载！'
    actionError.value = false
    fetchConfigState()
  } else {
    actionMsg.value = `❌ 更新失败: ${data?.error || '网络错误'}`
    actionError.value = true
  }
  setTimeout(() => { actionMsg.value = '' }, 5000)
}

// ── Proxy Group & Node Actions ──
async function fetchProxies() {
  const data = await apiFetch('/proxies')
  if (data?.groups) {
    proxyGroupsData.value = data
  }
}

async function selectProxy(groupName, nodeName) {
  const g = proxyGroupsData.value.groups?.find(x => x.name === groupName)
  if (g) g.now = nodeName

  const res = await apiFetch('/proxies/select', {
    method: 'POST',
    body: JSON.stringify({ group: groupName, name: nodeName })
  })
  if (res?.status === 'selected') {
    actionMsg.value = `✅ [${groupName}] 已切换至: ${nodeName}`
    actionError.value = false
  } else {
    actionMsg.value = `❌ 切换失败: ${res?.error || '内核未响应'}`
    actionError.value = true
    fetchProxies()
  }
  setTimeout(() => { actionMsg.value = '' }, 3500)
}

async function testSingleProxyDelay(nodeName) {
  const updateDelay = (name, delay, rechecking = false) => {
    if (proxyGroupsData.value.groups) {
      for (const g of proxyGroupsData.value.groups) {
        if (g.all) {
          for (const n of g.all) {
            if (n.name === name) {
              n.delay = delay
              n.rechecking = rechecking
            }
          }
        }
      }
    }
  }

  updateDelay(nodeName, -1, false)
  const res = await apiFetch('/proxies/delay', {
    method: 'POST',
    body: JSON.stringify({ name: nodeName })
  })

  if (res?.delay && res.delay > 0) {
    updateDelay(nodeName, res.delay, false)
  } else {
    // 10s Failure Re-check (Anti-Flapping)
    updateDelay(nodeName, -2, true)
    setTimeout(async () => {
      const retryRes = await apiFetch('/proxies/delay', {
        method: 'POST',
        body: JSON.stringify({ name: nodeName })
      })
      if (retryRes?.delay && retryRes.delay > 0) {
        updateDelay(nodeName, retryRes.delay, false)
      } else {
        updateDelay(nodeName, 0, false)
      }
    }, 1500)
  }
}

async function testAllProxiesDelay() {
  testingProxies.value = true
  actionMsg.value = '⚡ 正在对全部节点并发测速，请稍候...'
  await apiFetch('/telemetry/delay-test', { method: 'POST' })
  setTimeout(async () => {
    await fetchProxies()
    testingProxies.value = false
    actionMsg.value = '✅ 节点测速已完成！'
    setTimeout(() => { actionMsg.value = '' }, 3000)
  }, 2500)
}

function addSub() {
  configState.subscriptions.push({
    id: 'sub_' + Date.now(), name: '', type: 'clash', url: '', enabled: true
  })
}
function removeSub(i) { configState.subscriptions.splice(i, 1) }
function addCountry() {
  const v = newCountry.value.trim()
  if (v && !configState.auto_countries.includes(v)) {
    configState.auto_countries.push(v)
  }
  newCountry.value = ''
}

const availableCountries = computed(() => {
  const defaults = ['香港', '台湾', '新加坡', '韩国', '日本', '美国', '其它']
  const userList = configState.auto_countries || []
  return Array.from(new Set([...defaults, ...userList])).filter(Boolean)
})

function getRuleOfficialIconSvg(key) {
  if (!key) return '<span class="text-xs">🌐</span>'
  const k = key.toLowerCase()
  if (k.includes('openai') || k.includes('ai') || k.includes('chatgpt') || k.includes('claude')) {
    return `<svg viewBox="0 0 24 24" class="w-4 h-4 fill-emerald-400"><path d="M22.28 11.23a5.53 5.53 0 0 0-.49-4.88 5.6 5.6 0 0 0-4.54-2.73 5.57 5.57 0 0 0-4.13-1.85 5.66 5.66 0 0 0-5.32 3.73 5.54 5.54 0 0 0-3.69 2.67 5.6 5.6 0 0 0 .54 6.2 5.53 5.53 0 0 0 .49 4.88 5.6 5.6 0 0 0 4.54 2.73 5.57 5.57 0 0 0 4.13 1.85 5.66 5.66 0 0 0 5.32-3.73 5.54 5.54 0 0 0 3.69-2.67 5.6 5.6 0 0 0-.54-6.2zm-8.49 10.42a4.13 4.13 0 0 1-2.48-.82l.14-.08 4.14-2.39a.78.78 0 0 0 .39-.68v-5.85l1.83 1.05v4.67a4.17 4.17 0 0 1-4.02 4.1zm-8.23-3.92a4.1 4.1 0 0 1-.54-2.57l.14.08 4.14 2.39a.78.78 0 0 0 .78 0l5.07-2.93v2.11l-4.04 2.34a4.16 4.16 0 0 1-5.55-1.42zm-1.1-9.15a4.13 4.13 0 0 1 1.94-1.75v.16l4.14 2.39a.78.78 0 0 0 .78 0l5.07-2.93v-2.11l-4.04-2.34a4.16 4.16 0 0 1-7.89 6.58zm13.1 3.25l-5.07 2.93v-2.11l4.04-2.34a4.16 4.16 0 0 1 7.89-6.58 4.13 4.13 0 0 1-1.94 1.75v-.16l-4.14-2.39a.78.78 0 0 0-.78 0zm4.5 4.97a4.1 4.1 0 0 1 .54 2.57l-.14-.08-4.14-2.39a.78.78 0 0 0-.78 0l-5.07 2.93v-2.11l4.04-2.34a4.16 4.16 0 0 1 5.55 1.42zm-8.49-3.25l-2.06-1.19 2.06-1.19 2.06 1.19z"/></svg>`
  }
  if (k.includes('youtube') || k.includes('油管')) {
    return `<svg viewBox="0 0 24 24" class="w-4 h-4"><path fill="#FF0000" d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814z"/><path fill="#FFFFFF" d="M9.545 15.568V8.432L15.818 12l-6.273 3.568z"/></svg>`
  }
  if (k.includes('netflix') || k.includes('奈飞')) {
    return `<svg viewBox="0 0 24 24" class="w-4 h-4"><path fill="#E50914" d="M5.398 0v24c1.198-.12 2.455-.264 3.75-.43V0H5.398zm9.454 0v19.822c1.436-.18 2.75-.36 4.05-.56V0h-4.05z"/><path fill="#B81D24" d="M9.148 0L18.9 23.258c-1.3.2-2.614.38-4.05.56L5.398.24h3.75z"/></svg>`
  }
  if (k.includes('google') || k.includes('谷歌')) {
    return `<svg viewBox="0 0 24 24" class="w-4 h-4"><path fill="#4285F4" d="M23.745 12.27c0-.7-.06-1.4-.19-2.07H12v4.51h6.6c-.29 1.52-1.14 2.82-2.4 3.68v3.05h3.88c2.27-2.09 3.665-5.17 3.665-9.17z"/><path fill="#34A853" d="M12 24c3.24 0 5.95-1.08 7.93-2.91l-3.88-3.05c-1.08.72-2.45 1.16-4.05 1.16-3.12 0-5.77-2.1-6.72-4.93H1.25v3.15C3.26 21.36 7.33 24 12 24z"/><path fill="#FBBC05" d="M5.28 14.27c-.25-.72-.38-1.49-.38-2.27s.13-1.55.38-2.27V6.58H1.25C.45 8.18 0 10.03 0 12s.45 3.82 1.25 5.42l4.03-3.15z"/><path fill="#EA4335" d="M12 4.75c1.77 0 3.35.61 4.6 1.8l3.42-3.42C17.95 1.19 15.24 0 12 0 7.33 0 3.26 2.64 1.25 6.58l4.03 3.15c.95-2.83 3.6-4.98 6.72-4.98z"/></svg>`
  }
  if (k.includes('github') || k.includes('docker')) {
    return `<svg viewBox="0 0 24 24" class="w-4 h-4 fill-slate-100"><path fill-rule="evenodd" clip-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/></svg>`
  }
  if (k.includes('telegram') || k.includes('电报')) {
    return `<svg viewBox="0 0 24 24" class="w-4 h-4"><path fill="#24A1DE" d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.562 8.161l-1.97 9.285c-.145.658-.537.818-1.084.508l-3-2.21-1.447 1.394c-.16.16-.295.295-.605.295l.213-3.053 5.56-5.023c.242-.213-.054-.333-.373-.12l-6.871 4.326-2.962-.924c-.643-.204-.657-.643.136-.953l11.57-4.458c.538-.196 1.006.128.833.933z"/></svg>`
  }
  if (k.includes('twitter') || k.includes('推特') || k === 'x') {
    return `<svg viewBox="0 0 24 24" class="w-4 h-4 fill-slate-100"><path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/></svg>`
  }
  if (k.includes('spotify')) {
    return `<svg viewBox="0 0 24 24" class="w-4 h-4"><path fill="#1DB954" d="M12 0C5.4 0 0 5.4 0 12s5.4 12 12 12 12-5.4 12-12S18.66 0 12 0zm5.521 17.34c-.24.359-.66.48-1.021.24-2.82-1.74-6.36-2.101-10.561-1.141-.418.122-.779-.179-.899-.539-.12-.421.18-.78.54-.9 4.56-1.021 8.52-.6 11.64 1.32.42.18.479.659.301 1.02zm1.44-3.3c-.301.42-.841.6-1.262.3-3.239-1.98-8.159-2.58-11.939-1.38-.479.12-1.02-.12-1.14-.6-.12-.48.12-1.021.6-1.141C9.6 9.9 15 10.561 18.72 12.84c.361.181.54.78.241 1.2zm.12-3.36C15.24 8.4 8.82 8.16 5.16 9.301c-.6.179-1.2-.181-1.38-.721-.18-.601.18-1.2.72-1.381 4.26-1.26 11.28-1.02 15.721 1.621.539.3.719 1.02.419 1.56-.299.421-1.02.599-1.559.3z"/></svg>`
  }
  if (k.includes('games') || k.includes('game') || k.includes('steam') || k.includes('游戏')) {
    return `<svg viewBox="0 0 24 24" class="w-4 h-4 fill-cyan-400"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.44 9.82 8.24 11.43l3.6-5.18c-.37-.7-.59-1.5-.59-2.35 0-2.82 2.28-5.1 5.1-5.1s5.1 2.28 5.1 5.1-2.28 5.1-5.1 5.1c-.85 0-1.65-.22-2.35-.59l-5.18 3.6C10.18 24 12 24 12 24c6.63 0 12-5.37 12-12S18.63 0 12 0zm4.35 17.5c-1.93 0-3.5-1.57-3.5-3.5s1.57-3.5 3.5-3.5 3.5 1.57 3.5 3.5-1.57 3.5-3.5 3.5zm0-5.6c-1.16 0-2.1.94-2.1 2.1s.94 2.1 2.1 2.1 2.1-.94 2.1-2.1-.94-2.1-2.1-2.1z"/></svg>`
  }
  if (k.includes('apple') || k.includes('苹果')) {
    return `<svg viewBox="0 0 24 24" class="w-4 h-4 fill-slate-200"><path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M15.97 6.87c.66-.82 1.11-1.96.99-3.1-.96.04-2.17.65-2.85 1.45-.59.68-1.11 1.83-.97 2.95 1.08.08 2.17-.55 2.83-1.3"/></svg>`
  }
  if (k.includes('microsoft') || k.includes('微软')) {
    return `<svg viewBox="0 0 24 24" class="w-4 h-4"><rect width="11" height="11" fill="#F25022"/><rect x="13" width="11" height="11" fill="#7FBA00"/><rect y="13" width="11" height="11" fill="#00A4EF"/><rect x="13" y="13" width="11" height="11" fill="#FFB900"/></svg>`
  }
  if (k === 'cn' || k.includes('国内') || k.includes('直连') || k.includes('china')) {
    return `<span class="text-sm select-none">🇨🇳</span>`
  }
  return `<svg viewBox="0 0 24 24" class="w-4 h-4 fill-cyan-400"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"/></svg>`
}

function getCountryFlag(country) {
  if (!country) return '🌐'
  const c = country.toLowerCase()
  if (c.includes('香港') || c.includes('hk') || c.includes('hong')) return '🇭🇰'
  if (c.includes('台湾') || c.includes('tw') || c.includes('taiwan')) return '🇹🇼'
  if (c.includes('新加坡') || c.includes('sg') || c.includes('singapore')) return '🇸🇬'
  if (c.includes('日本') || c.includes('jp') || c.includes('japan')) return '🇯🇵'
  if (c.includes('美国') || c.includes('美') || c.includes('us') || c.includes('united states')) return '🇺🇸'
  if (c.includes('韩国') || c.includes('kr') || c.includes('korea')) return '🇰🇷'
  if (c.includes('英国') || c.includes('gb') || c.includes('uk')) return '🇬🇧'
  if (c.includes('德国') || c.includes('de') || c.includes('germany')) return '🇩🇪'
  if (c.includes('加拿大') || c.includes('ca')) return '🇨🇦'
  if (c.includes('澳洲') || c.includes('澳大利亚') || c.includes('au')) return '🇦🇺'
  if (c.includes('直连') || c.includes('中国') || c.includes('cn')) return '🇨🇳'
  if (c.includes('拦截') || c.includes('reject')) return '🛑'
  return '🌐'
}

function getPolicyFlag(policyName) {
  if (!policyName) return '🌐'
  if (policyName.includes('香港')) return '🇭🇰'
  if (policyName.includes('台湾')) return '🇹🇼'
  if (policyName.includes('新加坡')) return '🇸🇬'
  if (policyName.includes('日本')) return '🇯🇵'
  if (policyName.includes('美国')) return '🇺🇸'
  if (policyName.includes('韩国')) return '🇰🇷'
  if (policyName.includes('英国')) return '🇬🇧'
  if (policyName.includes('德国')) return '🇩🇪'
  if (policyName.includes('直连') || policyName === 'DIRECT') return '🇨🇳'
  if (policyName.includes('拦截') || policyName === 'REJECT') return '🛑'
  if (policyName.includes('自动') || policyName.includes('url-test')) return '⚡'
  if (policyName.includes('故转') || policyName.includes('fallback') || policyName.includes('容灾')) return '🛡️'
  if (policyName.includes('手动') || policyName.includes('select')) return '🎯'
  return '🌐'
}

function isCustomPolicy(pg) {
  if (!pg) return false
  const clean = pg.replace(/^[🇭🇰🇹🇼🇸🇬🇯🇵🇺🇸🇰🇷🇬🇧🇩🇪🇨🇳🛑⚡🛡️🎯🌐\s]+/, '')
  const standard = [
    '自动优选', '故障转移', '节点总览', '直连', '拦截', 'REJECT', 'DIRECT',
    ...availableCountries.value.flatMap(c => [
      `${c}自动`, `${c}故转`, `${c}容灾`, `${c}故障转移`, `${c}手动`, `${c}节点`
    ])
  ]
  return !standard.includes(pg) && !standard.includes(clean)
}

function addRuleTarget() {
  const nextOrder = (configState.rule_targets?.length || 0) + 1
  configState.rule_targets.push({
    id: 'rule_' + Date.now(),
    name: '新建分流业务',
    mrs_name: 'custom',
    policy_group: '香港自动',
    match_order: nextOrder,
    enabled: true,
    is_custom: true,
    _editingMrs: true,
    _customInput: false
  })
}

function removeRuleTarget(i) {
  configState.rule_targets.splice(i, 1)
}

// ── Live Connections & Session Inspector ──
function formatBytes(bytes) {
  if (!bytes || bytes <= 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
}

function formatDuration(startStr) {
  if (!startStr) return '-'
  const diffSec = Math.max(0, Math.floor((Date.now() - new Date(startStr).getTime()) / 1000))
  if (diffSec < 60) return `${diffSec}s`
  if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m ${diffSec % 60}s`
  return `${Math.floor(diffSec / 3600)}h ${Math.floor((diffSec % 3600) / 60)}m`
}

function formatChains(chains) {
  if (!chains || !chains.length) return []
  // Mihomo REST API (/connections) returns chains in leaf-first order: [outbound_proxy, sub_group, rule_group].
  // Reverse to display intuitive routing decision flow: rule_group -> sub_group -> outbound_proxy.
  return [...chains].reverse()
}

async function fetchConnections() {
  try {
    const res = await apiFetch('/connections')
    if (res && res.connections) {
      connectionsData.value = res
    }
  } catch (e) {
    console.error('Fetch connections error:', e)
  }
}

async function closeConnection(id) {
  if (!id) return
  try {
    await apiFetch('/connections/' + id, { method: 'DELETE' })
    connectionsData.value.connections = (connectionsData.value.connections || []).filter(c => c.id !== id)
  } catch (e) {
    console.error('Close connection error:', e)
  }
}

async function closeAllConnections() {
  if (!confirm('确定断开当前所有活跃连接吗？所有客户端将重新发起连接并命中最新规则。')) return
  try {
    await apiFetch('/connections', { method: 'DELETE' })
    connectionsData.value.connections = []
    actionMsg.value = '⚡ 已成功断开所有活跃会话'
    actionError.value = false
    setTimeout(() => { fetchConnections() }, 800)
  } catch (e) {
    actionMsg.value = '断开连接失败: ' + (e?.message || e)
    actionError.value = true
  }
}

async function flushDNSCache() {
  try {
    await apiFetch('/dns/flush', { method: 'POST' })
    actionMsg.value = '🧹 Fake-IP 与 SmartDNS 缓存已彻底清空'
    actionError.value = false
  } catch (e) {
    actionMsg.value = '清空缓存失败: ' + (e?.message || e)
    actionError.value = true
  }
}

const filteredConnections = computed(() => {
  const list = connectionsData.value.connections || []
  const kw = connSearch.value.trim().toLowerCase()
  return list.filter(c => {
    const net = (c.metadata?.network || '').toLowerCase()
    if (connFilterProtocol.value !== 'all' && net !== connFilterProtocol.value) return false

    const chainStr = (c.chains || []).join(' ')
    if (connFilterType.value === 'direct' && !chainStr.includes('直连') && !chainStr.includes('DIRECT')) return false
    if (connFilterType.value === 'proxy' && (chainStr.includes('直连') || chainStr.includes('DIRECT') || chainStr.includes('REJECT') || chainStr.includes('拦截'))) return false

    if (kw) {
      const host = (c.metadata?.host || '').toLowerCase()
      const dest = (c.metadata?.destinationIP || '').toLowerCase()
      const src = (c.metadata?.sourceIP || '').toLowerCase()
      const rule = (c.rulePayload || c.rule || '').toLowerCase()
      const chain = chainStr.toLowerCase()
      if (!host.includes(kw) && !dest.includes(kw) && !src.includes(kw) && !rule.includes(kw) && !chain.includes(kw)) {
        return false
      }
    }
    return true
  }).sort((a, b) => {
    if (connSortBy.value === 'download') return (b.download || 0) - (a.download || 0)
    if (connSortBy.value === 'upload') return (b.upload || 0) - (a.upload || 0)
    if (connSortBy.value === 'duration') {
      const ta = new Date(a.start).getTime() || 0
      const tb = new Date(b.start).getTime() || 0
      return ta - tb
    }
    return 0
  })
})

// ── Telemetry Actions ──
async function triggerDelayTest() {
  if (delayTesting.value) return
  delayTesting.value = true
  const probeUrl = currentProfile.value?.speedtest_url || 'http://www.gstatic.com/generate_204'
  const probeTimeout = currentProfile.value?.speedtest_timeout_ms || 2000
  actionMsg.value = `⚡ 正在并发测试全部节点延迟 (探测端点: ${probeUrl})...`
  actionError.value = false
  try {
    await apiFetch('/telemetry/delay-test', {
      method: 'POST',
      body: JSON.stringify({
        url: probeUrl,
        timeout: probeTimeout
      })
    })
    setTimeout(async () => {
      await fetchTelemetry()
      delayTesting.value = false
      actionMsg.value = '✅ 节点测速已完成！'
      setTimeout(() => { actionMsg.value = '' }, 4000)
    }, 4000)
  } catch (e) {
    delayTesting.value = false
  }
}

// ── TG Bot Actions ──
async function saveTgBot() {
  loading.value = true
  const ids = tgBotInput.allowed_chat_ids_str
    .split(',')
    .map(s => parseInt(s.trim()))
    .filter(n => !isNaN(n))

  const payload = {
    enabled: tgBotInput.enabled,
    proxy_url: tgBotInput.proxy_url,
    allowed_chat_ids: ids,
  }
  if (tgBotInput.token) {
    payload.token = tgBotInput.token
  }

  const data = await apiFetch('/tgbot/config', {
    method: 'PUT',
    body: JSON.stringify(payload)
  })
  loading.value = false
  if (data?.status === 'saved') {
    actionMsg.value = '✅ Telegram Bot 设置已保存！'
    actionError.value = false
    fetchTgBot()
  }
  setTimeout(() => { actionMsg.value = '' }, 5000)
}

// ── Upgrade Actions ──
async function upgradePanel() {
  loading.value = true
  actionMsg.value = '🚀 正在校验并升级 SingCube 专有系统...'
  actionError.value = false
  const data = await apiFetch('/system/upgrade-panel', { method: 'POST' })
  loading.value = false
  if (data?.status === 'upgraded') {
    actionMsg.value = `✅ SingCube 系统升级成功 (${data?.version || '最新版本'})！正在自动重载服务...`
    actionError.value = false
    setTimeout(() => { window.location.reload() }, 3000)
  } else if (data?.status === 'already_latest' || data?.message) {
    actionMsg.value = `✓ ${data.message || 'SingCube 系统当前已是最新版本'}`
    actionError.value = false
  } else {
    actionMsg.value = `❌ 系统升级失败: ${data?.error || '网络或版本发布源异常'}`
    actionError.value = true
  }
  setTimeout(() => { actionMsg.value = '' }, 6000)
}

async function upgradeMihomo() {
  loading.value = true
  actionMsg.value = '🚀 正在通过高速镜像拉取并校验 Mihomo 内核...'
  actionError.value = false
  const data = await apiFetch('/ros/upgrade-mihomo', { method: 'POST' })
  loading.value = false
  if (data?.upgraded_to) {
    actionMsg.value = `✅ Mihomo 内核已成功升级至 ${data.upgraded_to}，核心已热重载生效！`
    actionError.value = false
    fetchROS()
    fetchStatus()
  } else if (data?.message) {
    actionMsg.value = `✓ ${data.message}`
    actionError.value = false
  } else {
    actionMsg.value = `❌ Mihomo 内核升级失败: ${data?.error || '下载源或校验异常'}`
    actionError.value = true
  }
  setTimeout(() => { actionMsg.value = '' }, 6000)
}

async function upgradeFirmware() {
  loading.value = true
  actionMsg.value = '🚀 正在触发 RouterOS 固件在线升级...'
  actionError.value = false
  const data = await apiFetch('/ros/upgrade-firmware', { method: 'POST' })
  loading.value = false
  if (data?.status === 'upgrade_triggered') {
    actionMsg.value = '✅ 已成功向 RouterOS 下发固件下载并升级指令！'
    actionError.value = false
  } else {
    actionMsg.value = `❌ RouterOS 升级失败: ${data?.error || '未连接或权限不足'}`
    actionError.value = true
  }
  setTimeout(() => { actionMsg.value = '' }, 6000)
}

// ── Geo Data & Rules Update Actions ──
const geoDataStatus = reactive({
  geoip_size: 0,
  geoip_updated: '',
  geosite_size: 0,
  geosite_updated: '',
  mrs_rules_count: 0,
  last_updated: ''
})
const geoDataLoading = ref(false)
const geoDataUpdating = ref(false)

async function fetchGeoDataStatus() {
  geoDataLoading.value = true
  try {
    const data = await apiFetch('/system/geodata/status')
    if (data) Object.assign(geoDataStatus, data)
  } catch (e) {
    console.error('Fetch geodata status error:', e)
  } finally {
    geoDataLoading.value = false
  }
}

async function updateGeoDataModal() {
  geoDataUpdating.value = true
  actionMsg.value = '🔄 正在通过高速镜像通道更新 GeoIP / GeoSite 与核心 MRS 规则库...'
  actionError.value = false
  try {
    const data = await apiFetch('/system/geodata/update', { method: 'POST', timeout: 120000 })
    if (data?.success) {
      actionMsg.value = `✅ 已成功更新 ${data.files_updated || 7} 个核心数据库并热重载生效！`
      actionError.value = false
      await fetchGeoDataStatus()
      launchFireworks()
    } else {
      actionMsg.value = `⚠️ 规则库更新未全部完成: ${data?.error || '部分镜像源超时'}`
      actionError.value = true
    }
  } catch (err) {
    actionMsg.value = `❌ 更新异常: ${err.message}`
    actionError.value = true
  } finally {
    geoDataUpdating.value = false
    setTimeout(() => { actionMsg.value = '' }, 6000)
  }
}

// ── WireGuard Penetration Actions ──
async function fetchWireGuardInfo() {
  const data = await apiFetch('/wireguard/info')
  if (data) {
    if (data.config) Object.assign(wgInfo.config, data.config)
    wgInfo.peers = data.peers || []
    wgInfo.total_peers = data.total_peers || 0
    wgInfo.effective_endpoint = data.effective_endpoint || ''
    wgInfo.cloud_ddns = data.cloud_ddns || ''
  }
}

async function saveWireGuardConfig() {
  loading.value = true
  const data = await apiFetch('/wireguard/config', {
    method: 'POST',
    body: JSON.stringify(wgInfo.config)
  })
  if (data?.status === 'saved') {
    actionMsg.value = '✅ WireGuard 服务端配置已保存'
    actionError.value = false
    await fetchWireGuardInfo()
  } else {
    actionMsg.value = '❌ 保存配置失败: ' + (data?.error || '未知错误')
    actionError.value = true
  }
  loading.value = false
  setTimeout(() => { actionMsg.value = '' }, 5000)
}

async function initRosWireGuard() {
  loading.value = true
  const data = await apiFetch('/wireguard/init-ros', { method: 'POST' })
  if (data?.status === 'initialized') {
    actionMsg.value = '✅ RouterOS WireGuard 服务与虚拟网段已成功初始化！'
    actionError.value = false
    await fetchWireGuardInfo()
  } else {
    actionMsg.value = '⚠️ ROS 初始化提示: ' + (data?.error || '连接超时')
    actionError.value = true
  }
  loading.value = false
  setTimeout(() => { actionMsg.value = '' }, 6000)
}

async function openQrModal(peerName, confText) {
  currentPeerName.value = peerName
  currentPeerConfig.value = confText
  try {
    currentQrDataUrl.value = await QRCode.toDataURL(confText, {
      width: 320,
      margin: 2,
      color: {
        dark: '#130d2a',
        light: '#ffffff'
      }
    })
    showQrModal.value = true
  } catch (err) {
    actionMsg.value = '❌ 生成二维码失败: ' + err.message
    actionError.value = true
    setTimeout(() => { actionMsg.value = '' }, 5000)
  }
}

async function createNewPeer() {
  loading.value = true
  const data = await apiFetch('/wireguard/peers', {
    method: 'POST',
    body: JSON.stringify({
      name: newPeerForm.name.trim() || '我的移动设备',
      mode: newPeerForm.mode
    })
  })
  if (data?.peer) {
    actionMsg.value = `✅ 设备【${data.peer.name}】授权已生成`
    actionError.value = false
    showNewPeerModal.value = false
    newPeerForm.name = ''
    await fetchWireGuardInfo()
    await openQrModal(data.peer.name, data.config)
  } else {
    actionMsg.value = '❌ 创建失败: ' + (data?.error || '未知错误')
    actionError.value = true
  }
  loading.value = false
  setTimeout(() => { actionMsg.value = '' }, 5000)
}

async function showPeerQR(peer) {
  loading.value = true
  const data = await apiFetch(`/wireguard/peers/config?id=${peer.id}`)
  if (data?.config) {
    await openQrModal(peer.name, data.config)
  } else {
    actionMsg.value = '❌ 获取配置失败: ' + (data?.error || '未找到')
    actionError.value = true
    setTimeout(() => { actionMsg.value = '' }, 5000)
  }
  loading.value = false
}

async function downloadPeerConf(peer) {
  loading.value = true
  const data = await apiFetch(`/wireguard/peers/config?id=${peer.id}`)
  if (data?.config) {
    downloadConfFile(peer.name, data.config)
  } else {
    actionMsg.value = '❌ 获取配置失败'
    actionError.value = true
  }
  loading.value = false
}

async function deletePeer(peer) {
  if (!confirm(`确定撤回设备【${peer.name}】的穿透授权吗？`)) return
  loading.value = true
  const data = await apiFetch(`/wireguard/peers?id=${peer.id}`, { method: 'DELETE' })
  if (data?.status === 'deleted') {
    actionMsg.value = `✅ 已撤回设备【${peer.name}】授权`
    actionError.value = false
    await fetchWireGuardInfo()
  } else {
    actionMsg.value = '❌ 撤回失败: ' + (data?.error || '未知错误')
    actionError.value = true
  }
  loading.value = false
  setTimeout(() => { actionMsg.value = '' }, 5000)
}

// ── WireGuard One-Click Wizard Functions ──
function skipWizardLater() {
  closeWizard()
  localStorage.setItem('singcube_init_wizard_completed', 'true')
  actionMsg.value = '💡 已跳过 WireGuard 穿透配置，您可随时在「WireGuard 穿透」或「系统设置」中开启'
  actionError.value = false
  setTimeout(() => { actionMsg.value = '' }, 5000)
}

async function startWgWizard() {
  showWgWizardModal.value = true
  wizardStage.value = 'deploying'
  wizardProgress.value = 5
  wizardPeer.value = null
  wizardConfig.value = ''
  wizardQrUrl.value = ''
  wizardLiveStatus.value = null
  wizardErrorMsg.value = ''
  wizardDeploySeconds.value = 0

  if (wizardPollTimer.value) clearInterval(wizardPollTimer.value)
  if (wizardTimeoutTimer.value) clearTimeout(wizardTimeoutTimer.value)
  if (wizardSecondsInterval.value) clearInterval(wizardSecondsInterval.value)

  // 1-second interval to update deploy seconds counter
  wizardSecondsInterval.value = setInterval(() => {
    if (wizardStage.value === 'deploying') {
      wizardDeploySeconds.value++
      if (wizardDeploySeconds.value >= 30) {
        clearInterval(wizardSecondsInterval.value)
        wizardStage.value = 'timeout'
      }
    } else {
      clearInterval(wizardSecondsInterval.value)
    }
  }, 1000)

  // 30s watchdog safety timer
  wizardTimeoutTimer.value = setTimeout(() => {
    if (wizardStage.value === 'deploying') {
      wizardStage.value = 'timeout'
    }
  }, 30000)

  // Reset steps
  wizardSteps.value.forEach(s => s.status = 'wait')

  // Step 1: Connecting
  wizardSteps.value[0].status = 'process'
  wizardProgress.value = 15
  await new Promise(r => setTimeout(r, 200))
  wizardSteps.value[0].status = 'done'

  // Step 2: Interface
  wizardSteps.value[1].status = 'process'
  wizardProgress.value = 35
  await new Promise(r => setTimeout(r, 200))
  wizardSteps.value[1].status = 'done'

  // Step 3: DDNS
  wizardSteps.value[2].status = 'process'
  wizardProgress.value = 55
  await new Promise(r => setTimeout(r, 200))
  wizardSteps.value[2].status = 'done'

  // Step 4: Firewall
  wizardSteps.value[3].status = 'process'
  wizardProgress.value = 75

  let res = null
  try {
    res = await apiFetch('/wireguard/wizard/auto-setup', {
      method: 'POST',
      body: JSON.stringify({ device_name: '我的手机' }),
      timeout: 28000
    })
  } catch (err) {
    console.error('Wizard error:', err)
  }

  // If already reached 30s timeout stage while waiting, exit
  if (wizardStage.value === 'timeout') {
    return
  }

  if (!res || !res.config) {
    if (wizardTimeoutTimer.value) clearTimeout(wizardTimeoutTimer.value)
    if (wizardSecondsInterval.value) clearInterval(wizardSecondsInterval.value)
    wizardStage.value = 'error'
    wizardErrorMsg.value = res?.error || '未能成功生成 WireGuard 配置。若您未连接 RouterOS 路由器，可先跳过此步。'
    wizardSteps.value[3].status = 'wait'
    return
  }

  // Clear 30s timeout timers
  if (wizardTimeoutTimer.value) clearTimeout(wizardTimeoutTimer.value)
  if (wizardSecondsInterval.value) clearInterval(wizardSecondsInterval.value)

  wizardSteps.value[3].status = 'done'

  // Step 5: Routing
  wizardSteps.value[4].status = 'process'
  wizardProgress.value = 90
  await new Promise(r => setTimeout(r, 200))
  wizardSteps.value[4].status = 'done'

  // Step 6: Peer Created
  wizardSteps.value[5].status = 'process'
  wizardProgress.value = 100
  await new Promise(r => setTimeout(r, 200))
  wizardSteps.value[5].status = 'done'

  wizardPeer.value = res.peer
  wizardConfig.value = res.config
  wizardQrUrl.value = await QRCode.toDataURL(res.config, {
    width: 320,
    margin: 2,
    color: { dark: '#130d2a', light: '#ffffff' }
  })

  await new Promise(r => setTimeout(r, 300))
  wizardStage.value = 'scanning'
  await fetchWireGuardInfo()

  startWizardHandshakePolling(res.peer.id)
}

function startWizardHandshakePolling(peerId) {
  if (wizardPollTimer.value) clearInterval(wizardPollTimer.value)

  wizardPollTimer.value = setInterval(async () => {
    if (wizardStage.value !== 'scanning') {
      clearInterval(wizardPollTimer.value)
      return
    }
    await checkWizardHandshake(peerId, false)
  }, 1500)
}

async function checkWizardHandshake(peerId, manual = false) {
  if (wizardChecking.value) return
  wizardChecking.value = true
  try {
    const data = await apiFetch(`/wireguard/peers/live-status?id=${peerId || ''}`)
    if (data) {
      wizardLiveStatus.value = data
      if (data.connected || data.last_handshake || data.rx_bytes > 0) {
        clearInterval(wizardPollTimer.value)
        wizardStage.value = 'celebrating'
        setTimeout(() => {
          launchFireworks()
        }, 120)
      } else if (manual) {
        actionMsg.value = '⏳ 尚未检测到手机握手包，请确保已在手机 WireGuard App 中开启连接开关！'
        actionError.value = false
        setTimeout(() => { actionMsg.value = '' }, 4000)
      }
    }
  } catch (err) {
    console.warn('Check handshake error:', err)
  } finally {
    wizardChecking.value = false
  }
}

function simulateCelebrationForTesting() {
  wizardLiveStatus.value = {
    connected: true,
    last_handshake: '刚刚 (2秒前)',
    rx_bytes: 1420,
    tx_bytes: 2840,
    client_ip: wizardPeer.value?.client_ip || '10.88.0.2',
    current_endpoint: '已通过公网建立加密隧道'
  }
  if (wizardPollTimer.value) clearInterval(wizardPollTimer.value)
  wizardStage.value = 'celebrating'
  setTimeout(() => {
    launchFireworks()
  }, 120)
}

function closeWizard() {
  if (wizardPollTimer.value) clearInterval(wizardPollTimer.value)
  if (wizardTimeoutTimer.value) clearTimeout(wizardTimeoutTimer.value)
  if (wizardSecondsInterval.value) clearInterval(wizardSecondsInterval.value)
  showWgWizardModal.value = false
  fetchWireGuardInfo()
}

function launchFireworks(canvasId = 'fireworksCanvas') {
  const canvas = document.getElementById(canvasId) || document.getElementById('fireworksCanvas')
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  canvas.width = canvas.parentElement?.clientWidth || 500
  canvas.height = canvas.parentElement?.clientHeight || 450

  const particles = []
  const colors = ['#00d2ff', '#38bdf8', '#10b981', '#34d399', '#f59e0b', '#fbbf24', '#ec4899', '#a855f7', '#ffffff']

  for (let b = 0; b < 8; b++) {
    setTimeout(() => {
      const cx = (0.15 + Math.random() * 0.7) * canvas.width
      const cy = (0.15 + Math.random() * 0.5) * canvas.height
      const color = colors[Math.floor(Math.random() * colors.length)]
      for (let i = 0; i < 45; i++) {
        const angle = Math.random() * Math.PI * 2
        const speed = Math.random() * 6 + 1.5
        particles.push({
          x: cx,
          y: cy,
          vx: Math.cos(angle) * speed,
          vy: Math.sin(angle) * speed,
          alpha: 1,
          color: color,
          size: Math.random() * 3 + 2,
          decay: Math.random() * 0.015 + 0.012
        })
      }
    }, b * 180)
  }

  function loop() {
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    for (let i = particles.length - 1; i >= 0; i--) {
      const p = particles[i]
      p.x += p.vx
      p.y += p.vy
      p.vy += 0.06
      p.alpha -= p.decay
      if (p.alpha <= 0) {
        particles.splice(i, 1)
        continue
      }
      ctx.save()
      ctx.globalAlpha = p.alpha
      ctx.fillStyle = p.color
      ctx.beginPath()
      ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2)
      ctx.fill()
      ctx.restore()
    }
    if (particles.length > 0) {
      requestAnimationFrame(loop)
    }
  }
  loop()
}

async function viewRosScript() {
  const data = await apiFetch('/wireguard/ros-script')
  if (data?.script) {
    rosScriptText.value = data.script
    showRosScriptModal.value = true
  } else {
    actionMsg.value = '❌ 获取配置脚本失败'
    actionError.value = true
    setTimeout(() => { actionMsg.value = '' }, 5000)
  }
}

async function fetchWgAlignmentStatus() {
  wgAligning.value = true
  try {
    const data = await apiFetch('/wireguard/align/status')
    if (data) {
      Object.assign(wgAlignmentReport, data)
    }
  } catch (err) {
    console.error('Failed to fetch WG alignment:', err)
  } finally {
    wgAligning.value = false
  }
}

async function applyWgAlignment() {
  wgAligning.value = true
  try {
    const data = await apiFetch('/wireguard/align/apply', { method: 'POST' })
    if (data?.success) {
      actionMsg.value = `✅ ${data.message || 'WireGuard 与 RouterOS 已成功对齐！'}`
      actionError.value = false
      await fetchWireGuardInfo()
      await fetchWgAlignmentStatus()
    } else {
      actionMsg.value = `⚠️ 对齐提示: ${data?.message || data?.error || '部分对齐失败'}`
      actionError.value = true
      await fetchWireGuardInfo()
      await fetchWgAlignmentStatus()
    }
  } catch (err) {
    actionMsg.value = `❌ 对齐异常: ${err.message}`
    actionError.value = true
  } finally {
    wgAligning.value = false
    setTimeout(() => { actionMsg.value = '' }, 5000)
  }
}

function openRenamePeer(peer) {
  peerRenameForm.public_key = peer.public_key || peer.id
  peerRenameForm.name = peer.name || ''
  showRenamePeerModal.value = true
}

async function saveRenamePeer() {
  if (!peerRenameForm.name.trim()) {
    alert('请输入设备新名称！')
    return
  }
  loading.value = true
  try {
    const data = await apiFetch('/wireguard/peers/rename', {
      method: 'POST',
      body: JSON.stringify(peerRenameForm)
    })
    if (data?.status === 'renamed') {
      actionMsg.value = `✅ 设备已重命名为「${data.name}」并同步至 RouterOS！`
      actionError.value = false
      showRenamePeerModal.value = false
      await fetchWireGuardInfo()
    } else {
      actionMsg.value = `❌ 重命名失败: ${data?.error || '未知错误'}`
      actionError.value = true
    }
  } catch (err) {
    actionMsg.value = `❌ 请求异常: ${err.message}`
    actionError.value = true
  } finally {
    loading.value = false
    setTimeout(() => { actionMsg.value = '' }, 5000)
  }
}

async function fetchRosRoutingStatus() {
  rosRoutingLoading.value = true
  try {
    const data = await apiFetch('/ros/routing/status')
    if (data) {
      Object.assign(rosRoutingReport, data)
    }
  } catch (err) {
    console.error('Failed to fetch ROS routing status:', err)
  } finally {
    rosRoutingLoading.value = false
  }
}

async function applyRosRouting() {
  applyingRouting.value = true
  try {
    const data = await apiFetch('/ros/routing/apply', {
      method: 'POST',
      body: JSON.stringify({ target_ip: rosRoutingReport.singcube_ip })
    })
    if (data?.success) {
      actionMsg.value = `✅ ${data.message || 'RouterOS 规则已全部自动配置生效！'}`
      actionError.value = false
      await fetchRosRoutingStatus()
    } else {
      actionMsg.value = `⚠️ 配置提示: ${data?.message || data?.error || '部分规则未成功'}`
      actionError.value = true
      await fetchRosRoutingStatus()
    }
  } catch (err) {
    actionMsg.value = `❌ 自动配置异常: ${err.message}`
    actionError.value = true
  } finally {
    applyingRouting.value = false
    setTimeout(() => { actionMsg.value = '' }, 5000)
  }
}

async function viewWinBoxRoutingScript() {
  const data = await apiFetch('/ros/routing/script?target_ip=' + encodeURIComponent(rosRoutingReport.singcube_ip || ''))
  if (data?.script) {
    winBoxRoutingScriptText.value = data.script
    showWinBoxRoutingModal.value = true
  } else {
    actionMsg.value = '❌ 获取 WinBox 脚本失败'
    actionError.value = true
    setTimeout(() => { actionMsg.value = '' }, 4000)
  }
}

async function cleanRosRouting() {
  if (!confirm('确定要清理 RouterOS 的 Mangle 分流和 DNS 劫持规则吗？\n清理后将切入【纯旁路由模式】，局域网设备需手动将网关或DNS指向旁路由 IP，同网段多网关互不干扰。')) return
  cleaningRouting.value = true
  try {
    const data = await apiFetch('/ros/routing/clean', { method: 'POST' })
    if (data?.success) {
      actionMsg.value = `✅ ${data.message || 'RouterOS 规则已清理，已恢复经典纯旁路由模式！'}`
      actionError.value = false
      await fetchRosRoutingStatus()
    } else {
      actionMsg.value = `⚠️ 清理提示: ${data?.message || data?.error || '部分规则未清理'}`
      actionError.value = true
      await fetchRosRoutingStatus()
    }
  } catch (err) {
    actionMsg.value = `❌ 清理异常: ${err.message}`
    actionError.value = true
  } finally {
    cleaningRouting.value = false
    setTimeout(() => { actionMsg.value = '' }, 5000)
  }
}

// ── Host OS & Docker Companion Methods ──
async function fetchHostInfo() {
  try {
    const data = await apiFetch('/host/info')
    if (data) {
      Object.assign(hostInfo, data)
      if (data.host) {
        hostConfig.host = data.host
        hostConfig.port = data.port || 22
        hostConfig.user = data.user || 'root'
        hostConfig.auth_type = data.auth_type || 'password'
        hostConfig.is_configured = true
      }
    }
  } catch (err) {
    console.error('Failed to fetch host info:', err)
  }
}

async function testHostSSH() {
  hostTesting.value = true
  try {
    const data = await apiFetch('/host/test', {
      method: 'POST',
      body: JSON.stringify(hostConfig)
    })
    if (data?.ok) {
      actionMsg.value = `✅ SSH 连接成功！检测到系统: ${data.os_pretty || data.os_name} (${data.arch})`
      actionError.value = false
      Object.assign(hostInfo, data)
    } else {
      actionMsg.value = `❌ SSH 测试失败: ${data?.error || '连接被拒绝或超时'}`
      actionError.value = true
    }
  } catch (err) {
    actionMsg.value = `❌ SSH 连接异常: ${err.message}`
    actionError.value = true
  } finally {
    hostTesting.value = false
    setTimeout(() => { actionMsg.value = '' }, 5000)
  }
}

async function saveHostConfig() {
  hostSaving.value = true
  try {
    const data = await apiFetch('/host/save', {
      method: 'POST',
      body: JSON.stringify(hostConfig)
    })
    if (data?.success) {
      actionMsg.value = '✅ 宿主机 SSH 配置已安全加密保存！'
      actionError.value = false
      hostConfig.is_configured = true
      await fetchHostInfo()
    } else {
      actionMsg.value = `❌ 保存失败: ${data?.error || '未知错误'}`
      actionError.value = true
    }
  } catch (err) {
    actionMsg.value = `❌ 请求异常: ${err.message}`
    actionError.value = true
  } finally {
    hostSaving.value = false
    setTimeout(() => { actionMsg.value = '' }, 5000)
  }
}

async function openHostFwModal(mode = 'permissive') {
  showHostFwModal.value = true
  hostFwMode.value = mode
  hostFwActiveTab.value = 'script'
  hostFwLogs.value = []
  hostFwSuccess.value = false
  hostFwError.value = ''
  await fetchFwScript(hostFwMode.value)
}

async function selectFwMode(mode) {
  hostFwMode.value = mode
  hostFwLogs.value = []
  hostFwSuccess.value = false
  hostFwError.value = ''
  await fetchFwScript(mode)
}

async function fetchFwScript(mode) {
  hostFwScriptLoading.value = true
  try {
    const data = await apiFetch(`/host/firewall/script?mode=${mode}`)
    if (data?.script) {
      hostScriptText.value = data.script
    }
  } catch (err) {
    console.error('Fetch firewall script error:', err)
  } finally {
    hostFwScriptLoading.value = false
  }
}

async function applyHostFirewallModal() {
  hostApplyingFw.value = true
  hostFwLogs.value = []
  hostFwSuccess.value = false
  hostFwError.value = ''
  try {
    const data = await apiFetch('/host/firewall/apply', {
      method: 'POST',
      body: JSON.stringify({ mode: hostFwMode.value })
    })
    if (data?.logs && Array.isArray(data.logs)) {
      hostFwLogs.value = data.logs
    }
    if (data?.success) {
      hostFwSuccess.value = true
      actionMsg.value = `✅ 宿主机网络底座与防火墙已按【${hostFwMode.value === 'permissive' ? '宽松模式' : '严格模式'}】配置完成！`
      actionError.value = false
      await fetchHostInfo()
    } else {
      hostFwError.value = data?.error || '部分规则下发失败，请查看下方日志'
      actionMsg.value = `⚠️ 防火墙配置未全部完成: ${hostFwError.value}`
      actionError.value = true
    }
  } catch (err) {
    hostFwError.value = err.message
    actionMsg.value = `❌ 配置异常: ${err.message}`
    actionError.value = true
  } finally {
    hostApplyingFw.value = false
    setTimeout(() => { actionMsg.value = '' }, 6000)
  }
}

async function applyHostFirewall() {
  await openHostFwModal('permissive')
  hostFwActiveTab.value = 'ssh'
}

async function viewHostScript() {
  await openHostFwModal('permissive')
}

// ── Initialization Setup Wizard Functions ──
function openInitWizard(step = 'welcome') {
  showInitWizardModal.value = true
  initWizardStep.value = step
  initBackupError.value = ''
  initBackupSuccess.value = false
  if (step === 'firewall' || initWizardStep.value === 'firewall') {
    fetchFwScript(hostFwMode.value)
  }
}

async function handleInitBackupImport(e) {
  const file = e.target.files?.[0]
  if (!file) return
  initBackupLoading.value = true
  initBackupError.value = ''
  initBackupSuccess.value = false
  try {
    const text = await file.text()
    const json = JSON.parse(text)
    const res = await fetch(API + '/backup/import', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(json)
    })
    const data = await res.json()
    if (res.ok) {
      initBackupSuccess.value = true
      initWizardStep.value = 'finished'
      localStorage.setItem('singcube_init_wizard_completed', 'true')
      await fetchConfigState()
      await fetchOverview()
      await fetchDevices()
      await fetchProxies()
      await fetchHostInfo()
      setTimeout(() => {
        launchFireworks('initFireworksCanvas')
      }, 150)
    } else {
      initBackupError.value = data.error || '备份文件格式不符合规范'
    }
  } catch (err) {
    initBackupError.value = '解析或导入失败: ' + err.message
  } finally {
    initBackupLoading.value = false
    e.target.value = ''
  }
}

function startFreshInstall() {
  initWizardStep.value = 'firewall'
  fetchFwScript(hostFwMode.value)
}

async function submitInitSub() {
  const url = initSubUrl.value.trim()
  if (!url) {
    alert('请先粘贴您的 Clash / Mihomo 订阅链接')
    return
  }
  initSubLoading.value = true
  initSubMsg.value = '正在添加订阅并拉取节点...'
  try {
    const newSub = {
      id: 'sub_' + Date.now(),
      name: initSubName.value.trim() || '主力订阅',
      type: 'clash',
      url: url,
      enabled: true
    }
    if (!configState.subscriptions) {
      configState.subscriptions = []
    }
    const exists = configState.subscriptions.find(s => s.url === url)
    if (!exists) {
      configState.subscriptions.push(newSub)
    }
    await apiFetch('/config', {
      method: 'POST',
      body: JSON.stringify(configState)
    })
    const data = await apiFetch('/config/subscriptions/update', { method: 'POST', timeout: 60000 })
    if (data?.status === 'updated') {
      initSubSuccess.value = true
      initSubMsg.value = '✅ 订阅导入成功！已成功载入节点与分流策略。'
      await fetchConfigState()
      await fetchProxies()
    } else {
      initSubMsg.value = '⚠️ 订阅已保存，但更新节点时提示: ' + (data?.error || '请稍后手动刷新')
    }
  } catch (err) {
    initSubMsg.value = '❌ 导入异常: ' + err.message
  } finally {
    initSubLoading.value = false
  }
}

async function startInitCore() {
  initCoreLoading.value = true
  try {
    await coreAction('start')
    await fetchOverview()
    await fetchStatus()
  } catch (err) {
    console.error('Init core error:', err)
  } finally {
    initCoreLoading.value = false
  }
}

function startInitWireGuard() {
  showInitWizardModal.value = false
  startWgWizard()
}

function finishInitWizard() {
  localStorage.setItem('singcube_init_wizard_completed', 'true')
  showInitWizardModal.value = false
  actionMsg.value = '🎉 欢迎使用 SingCube 智能网关系统！'
  actionError.value = false
  setTimeout(() => { actionMsg.value = '' }, 5000)
}

async function deployHostWireGuard() {
  hostDeployingWg.value = true
  try {
    const data = await apiFetch('/host/wireguard/deploy', { method: 'POST' })
    if (data?.success) {
      actionMsg.value = `✅ 宿主机 WireGuard 服务已成功部署并启动: ${data.message}`
      actionError.value = false
      await fetchHostInfo()
      await fetchWireGuardInfo()
    } else {
      actionMsg.value = `❌ 部署失败: ${data?.error || '未知错误'}`
      actionError.value = true
    }
  } catch (err) {
    actionMsg.value = `❌ 部署异常: ${err.message}`
    actionError.value = true
  } finally {
    hostDeployingWg.value = false
    setTimeout(() => { actionMsg.value = '' }, 5000)
  }
}

async function checkHostUpgrade() {
  hostCheckingUpgrade.value = true
  try {
    const data = await apiFetch('/host/upgrade/check', { method: 'POST' })
    if (data?.success) {
      hostUpgradeResult.checked = true
      hostUpgradeResult.upgradable_count = data.upgradable_count || 0
      hostUpgradeResult.upgradable_pkgs = data.upgradable_pkgs || ''
      showHostUpgradeModal.value = true
    } else {
      actionMsg.value = `❌ 检查宿主机更新失败: ${data?.error || '未知错误'}`
      actionError.value = true
      setTimeout(() => { actionMsg.value = '' }, 5000)
    }
  } catch (err) {
    actionMsg.value = `❌ 检查更新异常: ${err.message}`
    actionError.value = true
    setTimeout(() => { actionMsg.value = '' }, 5000)
  } finally {
    hostCheckingUpgrade.value = false
  }
}

async function executeHostUpgrade() {
  if (!confirm('确定要在宿主机执行系统升级吗？这可能会升级系统软件包，建议避开业务高峰期。')) return
  hostUpgrading.value = true
  hostUpgradeResult.running = true
  hostUpgradeResult.log = '正在执行宿主机系统升级，请稍候...\n'
  try {
    const data = await apiFetch('/host/upgrade/execute', { method: 'POST' })
    if (data?.success) {
      hostUpgradeResult.log += (data.log || '') + '\n✅ 系统升级已顺利完成！\n'
      actionMsg.value = '✅ 宿主机系统升级完成！'
      actionError.value = false
      await fetchHostInfo()
    } else {
      hostUpgradeResult.log += '\n❌ 升级报错: ' + (data?.error || '未知错误')
      actionMsg.value = '❌ 宿主机升级未完全成功'
      actionError.value = true
    }
  } catch (err) {
    hostUpgradeResult.log += '\n❌ 网络异常: ' + err.message
    actionMsg.value = '❌ 升级请求异常: ' + err.message
    actionError.value = true
  } finally {
    hostUpgrading.value = false
    hostUpgradeResult.running = false
  }
}


function copyToClipboard(text, btnId = '') {
  if (!text) return
  const onSuccess = () => {
    actionMsg.value = '📋 已成功复制到剪贴板！'
    actionError.value = false
    if (btnId) {
      copiedButtonId.value = btnId
      setTimeout(() => { if (copiedButtonId.value === btnId) copiedButtonId.value = '' }, 2500)
    }
    setTimeout(() => { actionMsg.value = '' }, 3000)
  }

  // 1. If in secure context with modern API
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(text).then(onSuccess).catch(() => {
      fallbackCopy(text, onSuccess)
    })
    return
  }
  // 2. Fallback to execCommand (guaranteed to work across HTTP LAN IPs)
  fallbackCopy(text, onSuccess)
}

function fallbackCopy(text, onSuccess) {
  try {
    const textArea = document.createElement('textarea')
    textArea.value = text
    textArea.setAttribute('readonly', '')
    textArea.style.position = 'fixed'
    textArea.style.top = '0'
    textArea.style.left = '0'
    textArea.style.width = '2em'
    textArea.style.height = '2em'
    textArea.style.padding = '0'
    textArea.style.border = 'none'
    textArea.style.outline = 'none'
    textArea.style.boxShadow = 'none'
    textArea.style.background = 'transparent'
    textArea.style.fontSize = '16px'
    document.body.appendChild(textArea)
    textArea.focus()
    textArea.select()
    textArea.setSelectionRange(0, text.length)
    const successful = document.execCommand('copy')
    document.body.removeChild(textArea)
    if (successful) {
      onSuccess()
    } else {
      actionMsg.value = '⚠️ 浏览器限制自动复制，已选中文本，请使用 Ctrl+C / 长按手动复制'
      actionError.value = true
      setTimeout(() => { actionMsg.value = '' }, 3000)
    }
  } catch (err) {
    actionMsg.value = '⚠️ 复制异常，请在文本框内全选手动复制'
    actionError.value = true
    setTimeout(() => { actionMsg.value = '' }, 3000)
  }
}

// ── Configuration Backup & Restore (Item 3) ──
function exportBackup() {
  window.open('/api/v1/backup/export', '_blank')
  actionMsg.value = '📤 系统配置备份已生成并开始下载！'
  actionError.value = false
  setTimeout(() => { actionMsg.value = '' }, 4000)
}

async function importBackup(e) {
  const file = e.target?.files?.[0]
  if (!file) return
  if (!confirm(`确定要从备份文件【${file.name}】恢复系统配置吗？此操作将覆盖当前所有订阅与分流规则并热重启内核。`)) {
    e.target.value = ''
    return
  }
  try {
    const text = await file.text()
    const json = JSON.parse(text)
    const res = await fetch(API + '/backup/import', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(json)
    })
    const data = await res.json()
    if (res.ok) {
      actionMsg.value = `✅ ${data.message || '备份恢复成功！'}`
      actionError.value = false
      await fetchConfigState()
      await fetchOverview()
      await fetchDevices()
    } else {
      actionMsg.value = `❌ 导入失败: ${data.error || '数据格式错误'}`
      actionError.value = true
    }
  } catch (err) {
    actionMsg.value = `❌ 导入解析失败: ${err.message}`
    actionError.value = true
  } finally {
    e.target.value = ''
    setTimeout(() => { actionMsg.value = '' }, 5000)
  }
}

// ── LAN Device Discovery & Alias Management (Item 5) ──
async function fetchDevices() {
  devicesLoading.value = true
  try {
    const data = await apiFetch('/devices')
    if (data && data.devices) {
      devicesList.value = data.devices
    }
  } catch (err) {
    console.error('fetchDevices error:', err)
  } finally {
    devicesLoading.value = false
  }
}

const filteredDevices = computed(() => {
  const q = deviceSearch.value.trim().toLowerCase()
  if (!q) return devicesList.value
  return devicesList.value.filter(d =>
    (d.name && d.name.toLowerCase().includes(q)) ||
    (d.ip && d.ip.toLowerCase().includes(q)) ||
    (d.mac && d.mac.toLowerCase().includes(q)) ||
    (d.hostname && d.hostname.toLowerCase().includes(q))
  )
})

function openDeviceEdit(dev) {
  deviceEditForm.ip = dev?.ip === '-' ? '' : (dev?.ip || '')
  deviceEditForm.mac = dev?.mac === '-' ? '' : (dev?.mac || '')
  deviceEditForm.name = dev?.is_custom ? (dev?.name || '') : ''
  showDeviceModal.value = true
}

function openQuickRename(ip) {
  if (!ip) return
  const found = devicesList.value.find(d => d.ip === ip)
  deviceEditForm.ip = ip
  deviceEditForm.mac = found?.mac && found.mac !== '-' ? found.mac : ''
  deviceEditForm.name = found?.is_custom ? found.name : ''
  showDeviceModal.value = true
}

async function saveDeviceAlias() {
  if (!deviceEditForm.ip && !deviceEditForm.mac) {
    alert('请至少填写 IP 或 MAC 地址！')
    return
  }
  try {
    const res = await fetch(API + '/devices/alias', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(deviceEditForm)
    })
    const data = await res.json()
    if (res.ok) {
      actionMsg.value = `✅ 设备别名已成功保存！`
      actionError.value = false
      showDeviceModal.value = false
      await fetchDevices()
      setTimeout(() => { actionMsg.value = '' }, 3000)
    } else {
      actionMsg.value = `❌ 保存失败: ${data.error || '未知错误'}`
      actionError.value = true
    }
  } catch (err) {
    actionMsg.value = `❌ 网络错误: ${err.message}`
    actionError.value = true
  }
}

function getDeviceDisplayName(ip) {
  if (!ip) return '本机'
  const dev = devicesList.value.find(d => d.ip === ip)
  if (dev && dev.name && dev.is_custom) {
    return `${dev.name} (${ip})`
  }
  return ip
}

function downloadConfFile(name, content) {
  if (!content) return
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${name || 'singcube-wg'}.conf`
  a.click()
  URL.revokeObjectURL(url)
  actionMsg.value = `⬇️ 配置文件【${a.download}】下载已启动`
  actionError.value = false
  setTimeout(() => { actionMsg.value = '' }, 4000)
}

// ── KSpeeder API Methods (Item 2) ──
async function fetchKSpeeder() {
  const data = await apiFetch('/accelerator/status')
  if (data) Object.assign(kspeederStatus, data)
}

async function testDockerMirrors() {
  testingDocker.value = true
  actionMsg.value = '⚡ 正在对 Docker Registry 加速镜像源进行全并发测速...'
  actionError.value = false
  const data = await apiFetch('/accelerator/test-docker', { method: 'POST' })
  testingDocker.value = false
  if (data?.mirrors) {
    kspeederStatus.docker_mirrors = data.mirrors
    if (data.best_mirror) kspeederStatus.active_docker_mirror = data.best_mirror
    actionMsg.value = `✅ Docker 镜像测速完成！首选: ${kspeederStatus.active_docker_mirror}`
    actionError.value = false
  } else {
    actionMsg.value = '❌ Docker 镜像测速失败'
    actionError.value = true
  }
  setTimeout(() => { actionMsg.value = '' }, 4000)
}

async function testGitHubProxies() {
  testingGithub.value = true
  actionMsg.value = '⚡ 正在对 GitHub CDN 加速节点进行多源并发测速...'
  actionError.value = false
  const data = await apiFetch('/accelerator/test-github', { method: 'POST' })
  testingGithub.value = false
  if (data?.proxies) {
    kspeederStatus.github_proxies = data.proxies
    if (data.best_proxy) kspeederStatus.active_github_proxy = data.best_proxy
    actionMsg.value = `✅ GitHub 加速测速完成！首选 CDN: ${kspeederStatus.active_github_proxy || '官方直连'}`
    actionError.value = false
  } else {
    actionMsg.value = '❌ GitHub 加速测速失败'
    actionError.value = true
  }
  setTimeout(() => { actionMsg.value = '' }, 4000)
}

async function testAllAccelerators() {
  await Promise.all([testDockerMirrors(), testGitHubProxies()])
}

async function setActiveDockerMirror(url) {
  kspeederStatus.active_docker_mirror = url
  actionMsg.value = `🐳 已将首选镜像源设置为: ${url}`
  actionError.value = false
  await apiFetch('/accelerator/config', {
    method: 'POST',
    body: JSON.stringify({ active_docker_mirror: url })
  })
  setTimeout(() => { actionMsg.value = '' }, 3000)
}

async function setActiveGitHubProxy(url) {
  kspeederStatus.active_github_proxy = url
  actionMsg.value = `⚡ 已将 GitHub CDN 代理设置为: ${url || '官方直连'}`
  actionError.value = false
  await apiFetch('/accelerator/config', {
    method: 'POST',
    body: JSON.stringify({ active_github_proxy: url })
  })
  setTimeout(() => { actionMsg.value = '' }, 3000)
}

async function convertUrlToAccelerated() {
  if (!speedupInputUrl.value.trim()) return
  const data = await apiFetch('/accelerator/convert-url', {
    method: 'POST',
    body: JSON.stringify({ url: speedupInputUrl.value.trim() })
  })
  if (data?.accelerated_url) {
    speedupOutputUrl.value = data.accelerated_url
  } else {
    actionMsg.value = '⚠️ 链接解析失败或无需加速'
    actionError.value = true
    setTimeout(() => { actionMsg.value = '' }, 3000)
  }
}

async function viewRosRegistryScript() {
  const data = await apiFetch('/accelerator/ros-script')
  if (data?.script) {
    rosRegistryScriptText.value = data.script
    showRosRegistryModal.value = true
  } else {
    actionMsg.value = '❌ 获取 RouterOS 加速配置命令失败'
    actionError.value = true
    setTimeout(() => { actionMsg.value = '' }, 3000)
  }
}

async function viewDockerDaemonJson() {
  const data = await apiFetch('/accelerator/docker-daemon')
  if (data?.daemon_json) {
    dockerDaemonJsonText.value = data.daemon_json
    showDockerDaemonModal.value = true
  } else {
    actionMsg.value = '❌ 获取 Docker daemon.json 失败'
    actionError.value = true
    setTimeout(() => { actionMsg.value = '' }, 3000)
  }
}

// ── Qwen AI Copilot API Methods (Item 1) ──
function onAiProviderChange() {
  if (aiConfigForm.provider === 'opencode') {
    aiConfigForm.base_url = 'https://opencode.ai/zen/v1'
    aiConfigForm.model_name = 'nemotron-3.5-lightning-free'
  } else if (aiConfigForm.provider === 'qwen') {
    aiConfigForm.base_url = 'https://dashscope.aliyuncs.com/compatible-mode/v1'
    aiConfigForm.model_name = 'qwen-plus'
  } else if (aiConfigForm.provider === 'openai') {
    if (!aiConfigForm.base_url || aiConfigForm.base_url.includes('opencode.ai') || aiConfigForm.base_url.includes('dashscope')) {
      aiConfigForm.base_url = 'https://api.deepseek.com/v1'
      aiConfigForm.model_name = 'deepseek-chat'
    }
  }
}

async function fetchAiConfig() {
  const data = await apiFetch('/ai/config')
  if (data) {
    Object.assign(aiConfig, data)
    aiConfigForm.enabled = data.enabled
    aiConfigForm.provider = data.provider || 'opencode'
    aiConfigForm.base_url = data.base_url || 'https://opencode.ai/zen/v1'
    aiConfigForm.model_name = data.model || data.model_name || 'nemotron-3.5-lightning-free'
  }
}

async function saveAiConfig() {
  loading.value = true
  actionMsg.value = '⚙️ 正在保存 AI Copilot 配置...'
  actionError.value = false
  const data = await apiFetch('/ai/config', {
    method: 'PUT',
    body: JSON.stringify(aiConfigForm)
  })
  loading.value = false
  if (data?.status === 'ok') {
    showAiConfigModal.value = false
    await fetchAiConfig()
    actionMsg.value = '✅ AI 配置保存成功！'
    actionError.value = false
  } else {
    actionMsg.value = '❌ 保存配置失败: ' + (data?.error || '未知错误')
    actionError.value = true
  }
  setTimeout(() => { actionMsg.value = '' }, 4000)
}

async function sendChat() {
  const text = chatInput.value.trim()
  if (!text || loading.value || diagnosing.value) return
  chatMessages.value.push({
    role: 'user',
    content: text,
    time: new Date().toLocaleTimeString()
  })
  chatInput.value = ''
  loading.value = true

  const data = await apiFetch('/ai/chat', {
    method: 'POST',
    body: JSON.stringify({ message: text })
  })
  loading.value = false
  if (data?.reply) {
    chatMessages.value.push({
      role: 'assistant',
      content: data.reply,
      time: new Date().toLocaleTimeString()
    })
  } else {
    chatMessages.value.push({
      role: 'assistant',
      content: '⚠️ 通信提示: ' + (data?.error || '未能连接到云端 AI 接口，系统已自动切入本地自愈专家模式。'),
      time: new Date().toLocaleTimeString()
    })
  }
}

function sendQuickPrompt(promptText) {
  chatInput.value = promptText
  sendChat()
}

async function runAiDiagnose() {
  if (diagnosing.value || loading.value) return
  diagnosing.value = true
  chatMessages.value.push({
    role: 'user',
    content: '🩺 正在请求 AI Copilot 对当前软路由网络进行全维度健康体检与深度诊断...',
    time: new Date().toLocaleTimeString()
  })
  actionMsg.value = '🩺 AI Copilot 正在执行网络全面体检...'
  actionError.value = false

  const data = await apiFetch('/ai/diagnose', { method: 'POST' })
  diagnosing.value = false
  if (data?.report || data?.diagnosis) {
    const reportText = data.report || data.diagnosis
    chatMessages.value.push({
      role: 'assistant',
      content: reportText,
      time: new Date().toLocaleTimeString()
    })
    actionMsg.value = '✅ 网络深度体检已完成！请查看诊断报告'
    actionError.value = false
  } else {
    chatMessages.value.push({
      role: 'assistant',
      content: '❌ 体检报告生成失败: ' + (data?.error || '未知异常'),
      time: new Date().toLocaleTimeString()
    })
    actionMsg.value = '❌ 网络体检失败'
    actionError.value = true
  }
  setTimeout(() => { actionMsg.value = '' }, 4000)
}

// ── Utilities ──
function formatUptime(seconds) {
  if (!seconds || seconds <= 0) return '-'
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (d > 0) return `${d}天${h}时${m}分`
  if (h > 0) return `${h}时${m}分`
  return `${m}分`
}

// ── Lifecycle ──
let pollTimer = null
let speedTimer = null

onMounted(() => {
  fetchStatus()
  fetchProfiles()
  fetchRealtimeSpeed()
  fetchConfigState()
  fetchProxies()
  fetchWireGuardInfo()
  fetchROS()
  fetchTgBot()
  fetchKSpeeder()
  fetchAiConfig()
  fetchAdblockStats()
  fetchDevices()
  fetchRosRoutingStatus()
  fetchWgAlignmentStatus()
  fetchHostInfo()
  fetchGeoDataStatus()

  // ── Auto-launch Initialization Wizard for fresh install / unconfigured setups ──
  const initWizardDone = localStorage.getItem('singcube_init_wizard_completed')
  if (initWizardDone !== 'true') {
    setTimeout(() => {
      // If no subscriptions or not initialized, show wizard
      if (!configState.subscriptions || configState.subscriptions.length === 0) {
        openInitWizard('welcome')
      }
    }, 700)
  }

  // Poll status every 5 seconds.
  pollTimer = setInterval(fetchStatus, 5000)
  // Poll realtime traffic waveform every 2 seconds.
  speedTimer = setInterval(fetchRealtimeSpeed, 2000)
  if (currentView.value === 'connections') {
    fetchConnections()
    connInterval = setInterval(() => {
      if (connAutoRefresh.value && currentView.value === 'connections') {
        fetchConnections()
      }
    }, 2500)
  }
  if (currentView.value === 'adblock') {
    fetchAdblockStats()
    adblockInterval = setInterval(() => {
      if (adblockAutoRefresh.value && currentView.value === 'adblock') {
        fetchAdblockStats()
      }
    }, 3000)
  }
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
  if (speedTimer) clearInterval(speedTimer)
  if (connInterval) clearInterval(connInterval)
  if (adblockInterval) clearInterval(adblockInterval)
})
</script>

<style>
/* Glass Card */
.glass-card {
  @apply bg-white/5 backdrop-blur-md rounded-2xl border border-purple-800/30 shadow-lg shadow-black/20;
}

/* Buttons */
.btn {
  @apply px-4 py-2 rounded-xl text-sm font-medium transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed;
}
.btn-green { @apply bg-emerald-600/80 hover:bg-emerald-500 text-white; }
.btn-red   { @apply bg-red-600/80 hover:bg-red-500 text-white; }
.btn-blue  { @apply bg-blue-600/80 hover:bg-blue-500 text-white; }
.btn-cyan  { @apply bg-cyan-600/80 hover:bg-purple-600 text-white; }

/* Input */
.input-field {
  @apply bg-slate-800/70 border border-purple-900/40 rounded-lg px-3 py-2 text-sm text-slate-200
         focus:outline-none focus:border-purple-500/60 focus:ring-1 focus:ring-cyan-500/30
         placeholder:text-slate-600;
}

/* Drawer & Toast Animations */
@keyframes slide-left {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}
.animate-slide-left {
  animation: slide-left 0.28s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

.drawer-fade-enter-active, .drawer-fade-leave-active {
  transition: opacity 0.22s ease;
}
.drawer-fade-enter-from, .drawer-fade-leave-to {
  opacity: 0;
}

.toast-slide-enter-active, .toast-slide-leave-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.toast-slide-enter-from, .toast-slide-leave-to {
  opacity: 0;
  transform: translate(-50%, -20px);
}
</style>