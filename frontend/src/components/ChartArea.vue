<script setup>
import { computed } from 'vue'
import VueApexCharts from 'vue3-apexcharts'

const props = defineProps({
  target: { type: Object, default: null },
  metrics: { type: Array, default: () => [] },
  scale: { type: String, default: 'Raw Data' },
  chartMode: { type: String, default: 'stacked' }, // 'stacked' or 'line'
  loading: { type: Boolean, default: false },
})

const emit = defineEmits(['update:chartMode'])

const isServerType = computed(() => props.target?.type === 'server')

// Format timestamps for X-axis categories
const categories = computed(() => {
  return props.metrics.map((m) => {
    const d = new Date(m.timestamp)
    if (props.scale === 'Daily Avg') {
      return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
    }
    return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  })
})

// Build Series based on Target Type
const series = computed(() => {
  if (props.metrics.length === 0) return []

  if (isServerType.value) {
    // Server Linux: CPU, RAM, Disk, Network, and Response Time (SSH Latency)
    return [
      {
        name: 'CPU Usage (%)',
        data: props.metrics.map((m) => Number((m.cpu_pct || 0).toFixed(1))),
        color: '#f43f5e', // Rose
      },
      {
        name: 'RAM Usage (%)',
        data: props.metrics.map((m) => Number((m.ram_pct || 0).toFixed(1))),
        color: '#06b6d4', // Cyan
      },
      {
        name: 'Disk Space (%)',
        data: props.metrics.map((m) => Number((m.disk_pct || 0).toFixed(1))),
        color: '#8b5cf6', // Violet
      },
      {
        name: 'Network Speed (MB/s)',
        data: props.metrics.map((m) => Number((m.network_speed || 0).toFixed(2))),
        color: '#10b981', // Emerald
      },
      {
        name: 'Response Time (ms)',
        data: props.metrics.map((m) => Number((m.latency_ms || 0).toFixed(1))),
        color: '#f59e0b', // Amber / Yellow
      },
    ]
  }

  // Website / API / DB: Latency Breakdown Stacked Area
  if (props.chartMode === 'stacked') {
    return [
      {
        name: 'DNS Lookup (ms)',
        data: props.metrics.map((m) => {
          const v = m.details?.dns_lookup_ms ?? (m.latency_ms > 0 ? m.latency_ms * 0.1 : 0)
          return Number(v.toFixed(1))
        }),
        color: '#8b5cf6', // Layer 1 (Purple)
      },
      {
        name: 'TCP Connect (ms)',
        data: props.metrics.map((m) => {
          const v = m.details?.tcp_connect_ms ?? (m.latency_ms > 0 ? m.latency_ms * 0.15 : 0)
          return Number(v.toFixed(1))
        }),
        color: '#3b82f6', // Layer 2 (Blue)
      },
      {
        name: 'TLS Handshake (ms)',
        data: props.metrics.map((m) => {
          const v = m.details?.tls_handshake_ms ?? (m.latency_ms > 0 ? m.latency_ms * 0.25 : 0)
          return Number(v.toFixed(1))
        }),
        color: '#06b6d4', // Layer 3 (Cyan)
      },
      {
        name: 'TTFB (ms)',
        data: props.metrics.map((m) => {
          const v = m.details?.ttfb_ms ?? (m.latency_ms > 0 ? m.latency_ms * 0.35 : 0)
          return Number(v.toFixed(1))
        }),
        color: '#14b8a6', // Layer 4 (Teal)
      },
      {
        name: 'Content Download (ms)',
        data: props.metrics.map((m) => {
          const v = m.details?.content_download_ms ?? (m.latency_ms > 0 ? m.latency_ms * 0.15 : 0)
          return Number(v.toFixed(1))
        }),
        color: '#10b981', // Layer 5 (Green)
      },
    ]
  }

  // Single line latency
  return [
    {
      name: 'Total Latency (ms)',
      data: props.metrics.map((m) => Number(m.latency_ms.toFixed(1))),
      color: '#06b6d4',
    },
    {
      name: 'Max Latency (ms)',
      data: props.metrics.map((m) => Number((m.max_latency_ms || m.latency_ms).toFixed(1))),
      color: '#f59e0b',
    },
  ]
})

// ApexCharts Options
const chartOptions = computed(() => {
  const isStacked = props.chartMode === 'stacked' && !isServerType.value

  return {
    chart: {
      type: 'area',
      stacked: isStacked,
      height: 340,
      fontFamily: 'Inter, sans-serif',
      toolbar: { show: false },
      background: 'transparent',
      animations: {
        enabled: true,
        easing: 'easeinout',
        speed: 500,
      },
    },
    theme: { mode: 'dark' },
    stroke: {
      curve: 'smooth',
      width: isServerType.value ? [2, 2, 2, 2, 2.5] : 2,
    },
    fill: {
      type: 'gradient',
      gradient: {
        shade: 'dark',
        type: 'vertical',
        shadeIntensity: 0.5,
        gradientToColors: undefined,
        inverseColors: true,
        opacityFrom: isStacked ? 0.65 : 0.35,
        opacityTo: 0.05,
        stops: [0, 90, 100],
      },
    },
    dataLabels: { enabled: false },
    grid: {
      borderColor: '#1e293b',
      strokeDashArray: 4,
      xaxis: { lines: { show: true } },
      yaxis: { lines: { show: true } },
    },
    xaxis: {
      categories: categories.value,
      labels: {
        style: { colors: '#64748b', fontSize: '11px', fontFamily: 'JetBrains Mono, monospace' },
        rotate: 0,
        maxHeight: 40,
      },
      axisBorder: { color: '#334155' },
      axisTicks: { color: '#334155' },
    },
    yaxis: isServerType.value
      ? [
          {
            seriesName: 'CPU Usage (%)',
            title: {
              text: 'Resource Usage (%)',
              style: { color: '#64748b', fontSize: '11px', fontFamily: 'Inter, sans-serif' },
            },
            min: 0,
            max: 100,
            labels: {
              style: { colors: '#64748b', fontSize: '11px', fontFamily: 'JetBrains Mono, monospace' },
              formatter: (val) => `${val.toFixed(0)}%`,
            },
          },
          {
            seriesName: 'CPU Usage (%)',
            show: false,
          },
          {
            seriesName: 'CPU Usage (%)',
            show: false,
          },
          {
            seriesName: 'CPU Usage (%)',
            show: false,
          },
          {
            opposite: true,
            seriesName: 'Response Time (ms)',
            title: {
              text: 'Response Time (ms)',
              style: { color: '#f59e0b', fontSize: '11px', fontFamily: 'Inter, sans-serif' },
            },
            labels: {
              style: { colors: '#f59e0b', fontSize: '11px', fontFamily: 'JetBrains Mono, monospace' },
              formatter: (val) => `${val.toFixed(0)} ms`,
            },
          },
        ]
      : {
          labels: {
            style: { colors: '#64748b', fontSize: '11px', fontFamily: 'JetBrains Mono, monospace' },
            formatter: (val) => `${val.toFixed(0)}ms`,
          },
        },
    legend: {
      position: 'top',
      horizontalAlign: 'right',
      labels: { colors: '#94a3b8' },
      fontSize: '12px',
      markers: { radius: 3 },
    },
    tooltip: {
      theme: 'dark',
      x: { show: true },
      y: {
        formatter: (val, opts) => {
          if (val === undefined || val === null) return ''
          if (isServerType.value) {
            const seriesName = opts?.w?.globals?.seriesNames?.[opts?.seriesIndex] || ''
            if (seriesName.includes('Response Time') || seriesName.includes('Latency')) {
              return `${val} ms`
            }
            if (seriesName.includes('Network')) {
              return `${val} MB/s`
            }
            return `${val} %`
          }
          return `${val} ms`
        },
      },
    },
  }
})
</script>

<template>
  <div class="glass-panel rounded-2xl p-5 relative overflow-hidden">
    <!-- Header: Title, Scale Indicator, and View Toggles -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
      <div>
        <div class="flex items-center gap-2">
          <h3 class="text-sm font-bold tracking-wider uppercase text-slate-200">
            {{ isServerType ? 'System Resource Metrics' : 'Latency Breakdown Analytics' }}
          </h3>
          <span class="px-2 py-0.5 text-[10px] font-mono rounded-full bg-slate-800 text-cyan-400 border border-slate-700">
            Scale: {{ scale }}
          </span>
        </div>
        <p class="text-xs text-slate-400 mt-0.5">
          {{ isServerType ? 'Pergerakan beban CPU, RAM, Disk, Jaringan, dan Response Time (Latency ms)' : 'Visualisasi 5-layer HTTP/TLS roundtrip time' }}
        </p>
      </div>

      <!-- Mode Toggle: Stacked Area vs Smooth Line Chart -->
      <div class="flex items-center gap-2" v-if="!isServerType">
        <div class="inline-flex rounded-lg bg-slate-900 border border-slate-700/80 p-0.5 text-xs font-mono">
          <button
            @click="emit('update:chartMode', 'stacked')"
            class="px-2.5 py-1 rounded-md transition-all font-medium"
            :class="chartMode === 'stacked' ? 'bg-cyan-500/20 text-cyan-300 shadow-sm border border-cyan-500/30' : 'text-slate-400 hover:text-slate-200'"
          >
            Stacked Area
          </button>
          <button
            @click="emit('update:chartMode', 'line')"
            class="px-2.5 py-1 rounded-md transition-all font-medium"
            :class="chartMode === 'line' ? 'bg-cyan-500/20 text-cyan-300 shadow-sm border border-cyan-500/30' : 'text-slate-400 hover:text-slate-200'"
          >
            Line Chart
          </button>
        </div>
      </div>
    </div>

    <!-- Chart Container -->
    <div class="w-full min-h-[340px] relative">
      <div v-if="loading" class="absolute inset-0 bg-dark-950/50 backdrop-blur-sm z-10 flex items-center justify-center rounded-xl">
        <div class="flex items-center gap-3 text-cyan-400 font-mono text-xs">
          <svg class="animate-spin h-5 w-5" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          Memuat metrik analitik...
        </div>
      </div>

      <div v-if="metrics.length === 0 && !loading" class="h-[340px] flex flex-col items-center justify-center text-slate-500">
        <svg class="w-12 h-12 text-slate-600 mb-2" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M3 3v18h18"></path>
          <path d="m19 9-5 5-4-4-3 3"></path>
        </svg>
        <span class="text-sm font-mono">Belum ada metrik terkumpul untuk target ini</span>
        <span class="text-xs text-slate-600 mt-1">Background collector sedang melakukan polling...</span>
      </div>

      <div v-else>
        <VueApexCharts
          type="area"
          height="340"
          :options="chartOptions"
          :series="series"
        />
      </div>
    </div>
  </div>
</template>
