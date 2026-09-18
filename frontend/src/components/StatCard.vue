<script setup>
import { computed } from 'vue'

const props = defineProps({
  stats: { type: Object, default: null },
  target: { type: Object, default: null },
})

const isOnline = computed(() => props.stats?.status === 'ONLINE')

// Format latency
const latencyDisplay = computed(() => {
  if (!props.stats || props.stats.current_latency_ms == null) return '--'
  const lat = props.stats.current_latency_ms
  if (lat < 1 && lat > 0) return lat.toFixed(2) + ' ms'
  return Math.round(lat) + ' ms'
})

// Latency threshold color
const latencyColor = computed(() => {
  const lat = props.stats?.current_latency_ms || 0
  if (lat <= 0) return 'text-slate-400'
  if (lat < 200) return 'text-emerald-400'
  if (lat < 600) return 'text-amber-400'
  return 'text-rose-400'
})

// CPU & RAM threshold colors
function getResourceColor(pct) {
  if (pct >= 95) return 'text-rose-400 bg-rose-500'
  if (pct >= 80) return 'text-amber-400 bg-amber-500'
  return 'text-emerald-400 bg-emerald-500'
}

// Sparkline SVG polyline points
const sparklinePoints = computed(() => {
  if (!props.stats?.sparkline || props.stats.sparkline.length < 2) return ''
  const data = props.stats.sparkline
  const width = 120
  const height = 28
  const max = Math.max(...data, 1)
  const min = Math.min(...data, 0)
  const range = max - min || 1

  return data
    .map((val, idx) => {
      const x = (idx / (data.length - 1)) * width
      const y = height - ((val - min) / range) * (height - 6) - 3
      return `${x.toFixed(1)},${y.toFixed(1)}`
    })
    .join(' ')
})
</script>

<template>
  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
    <!-- Card 1: System Status & Heartbeat -->
    <div class="glass-panel rounded-2xl p-4 transition-all glass-panel-hover relative overflow-hidden group">
      <div class="absolute -right-8 -top-8 w-24 h-24 bg-emerald-500/5 rounded-full blur-xl group-hover:bg-emerald-500/10 transition-colors"></div>
      
      <div class="flex items-center justify-between">
        <span class="text-xs font-mono font-medium text-slate-400 tracking-wider uppercase">
          Status Target
        </span>
        <!-- Heartbeat pulse -->
        <div class="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-slate-800/80 border border-slate-700/60">
          <span class="relative flex h-2 w-2">
            <span
              v-if="isOnline"
              class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"
            ></span>
            <span
              class="relative inline-flex rounded-full h-2 w-2"
              :class="isOnline ? 'bg-emerald-500' : 'bg-rose-500'"
            ></span>
          </span>
          <span class="text-[10px] font-mono font-bold" :class="isOnline ? 'text-emerald-400' : 'text-rose-400'">
            {{ isOnline ? 'ONLINE' : 'OFFLINE' }}
          </span>
        </div>
      </div>

      <div class="mt-3 flex items-baseline gap-2">
        <div class="text-2xl font-black tracking-tight" :class="isOnline ? 'text-emerald-400' : 'text-rose-400'">
          {{ isOnline ? 'OPERATIONAL' : 'INCIDENT' }}
        </div>
      </div>

      <div class="mt-2 flex items-center gap-1.5 text-xs text-slate-400 font-mono">
        <svg class="w-3.5 h-3.5 text-slate-500 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <polyline points="12 6 12 12 16 14"></polyline>
        </svg>
        <span class="truncate">{{ stats?.uptime_duration || 'Uptime: kalkulasi...' }}</span>
      </div>
    </div>

    <!-- Card 2: Average Latency & Sparkline -->
    <div class="glass-panel rounded-2xl p-4 transition-all glass-panel-hover relative overflow-hidden group">
      <div class="absolute -right-8 -top-8 w-24 h-24 bg-cyan-500/5 rounded-full blur-xl group-hover:bg-cyan-500/10 transition-colors"></div>

      <div class="flex items-center justify-between">
        <span class="text-xs font-mono font-medium text-slate-400 tracking-wider uppercase">
          Latency / Response
        </span>
        <span class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
          RT Ping
        </span>
      </div>

      <div class="mt-3 flex items-center justify-between">
        <div class="text-2xl font-black font-mono tracking-tight" :class="latencyColor">
          {{ latencyDisplay }}
        </div>

        <!-- Mini Sparkline SVG -->
        <div class="h-7 w-28 shrink-0 flex items-center" v-if="sparklinePoints">
          <svg class="w-full h-full overflow-visible" viewBox="0 0 120 28">
            <polyline
              fill="none"
              stroke="#06b6d4"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              :points="sparklinePoints"
            />
          </svg>
        </div>
      </div>

      <div class="mt-2 text-xs text-slate-400 font-mono flex items-center justify-between">
        <span>Trend 1 jam terakhir</span>
        <span class="text-[10px] text-cyan-400">Live</span>
      </div>
    </div>

    <!-- Card 3: Overall Uptime Rate -->
    <div class="glass-panel rounded-2xl p-4 transition-all glass-panel-hover relative overflow-hidden group">
      <div class="absolute -right-8 -top-8 w-24 h-24 bg-indigo-500/5 rounded-full blur-xl group-hover:bg-indigo-500/10 transition-colors"></div>

      <div class="flex items-center justify-between">
        <span class="text-xs font-mono font-medium text-slate-400 tracking-wider uppercase">
          Service Availability
        </span>
        <span class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
          SLA
        </span>
      </div>

      <div class="mt-3 flex items-baseline justify-between">
        <div class="text-2xl font-black font-mono tracking-tight text-white">
          {{ stats?.uptime_rate_pct != null ? stats.uptime_rate_pct.toFixed(2) + '%' : '100.00%' }}
        </div>
        <span class="text-[11px] font-mono text-slate-400" v-if="stats?.total_checks > 0">
          {{ stats.online_checks }}/{{ stats.total_checks }} UP
        </span>
      </div>

      <div class="mt-2 flex items-center gap-2">
        <div class="w-full bg-slate-800 rounded-full h-1.5 overflow-hidden">
          <div
            class="h-full rounded-full transition-all duration-500 bg-gradient-to-r from-emerald-500 to-teal-400"
            :style="{ width: `${stats?.uptime_rate_pct || 100}%` }"
          ></div>
        </div>
        <span class="text-[10px] font-mono text-slate-400 shrink-0 uppercase">
          {{ stats?.time_range || '24h' }}
        </span>
      </div>

      <div class="mt-1.5 text-[10px] font-mono text-slate-400 truncate">
        Calculated based on selected time-range
      </div>
    </div>

    <!-- Card 4: Resource Overview / Health Summary -->
    <div class="glass-panel rounded-2xl p-4 transition-all glass-panel-hover relative overflow-hidden group">
      <div class="flex items-center justify-between">
        <span class="text-xs font-mono font-medium text-slate-400 tracking-wider uppercase">
          Health / Diagnostics
        </span>
        <span class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 border border-slate-700 uppercase">
          {{ target?.type || 'Diagnostic' }}
        </span>
      </div>

      <!-- If target is Server (Linux SSH): CPU & RAM Mini Gauges -->
      <div v-if="target?.type === 'server'" class="mt-2.5 space-y-2">
        <div>
          <div class="flex justify-between text-xs font-mono mb-0.5">
            <span class="text-slate-400">CPU Usage</span>
            <span :class="getResourceColor(stats?.cpu_pct || 0).split(' ')[0]" class="font-bold">
              {{ (stats?.cpu_pct || 0).toFixed(1) }}%
            </span>
          </div>
          <div class="w-full bg-slate-800 rounded-full h-1.5 overflow-hidden">
            <div
              class="h-full rounded-full transition-all duration-300"
              :class="getResourceColor(stats?.cpu_pct || 0).split(' ')[1]"
              :style="{ width: `${Math.min(stats?.cpu_pct || 0, 100)}%` }"
            ></div>
          </div>
        </div>

        <div>
          <div class="flex justify-between text-xs font-mono mb-0.5">
            <span class="text-slate-400">RAM Usage</span>
            <span :class="getResourceColor(stats?.ram_pct || 0).split(' ')[0]" class="font-bold">
              {{ (stats?.ram_pct || 0).toFixed(1) }}%
            </span>
          </div>
          <div class="w-full bg-slate-800 rounded-full h-1.5 overflow-hidden">
            <div
              class="h-full rounded-full transition-all duration-300"
              :class="getResourceColor(stats?.ram_pct || 0).split(' ')[1]"
              :style="{ width: `${Math.min(stats?.ram_pct || 0, 100)}%` }"
            ></div>
          </div>
        </div>
      </div>

      <!-- If target is Website / API / DB -->
      <div v-else class="mt-3">
        <div class="flex items-center gap-2">
          <div
            class="px-2.5 py-1 rounded-lg font-mono font-bold text-sm"
            :class="isOnline ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'"
          >
            {{ stats?.http_status ? `${stats.http_status} ${isOnline ? 'OK' : 'ERROR'}` : (isOnline ? 'LISTENING' : 'CLOSED') }}
          </div>
        </div>
        <div class="mt-2 text-xs text-slate-400 font-mono truncate">
          Target: {{ target?.host || 'Endpoint' }}
        </div>
      </div>
    </div>
  </div>
</template>
