<script setup>
const props = defineProps({
  portMatrix: { type: Array, default: () => [] },
})

function getStatusBadge(status) {
  switch (status) {
    case 'CONNECTED':
    case 'LISTENING':
    case 'RESOLVED':
      return {
        bg: 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400',
        dot: 'bg-emerald-500',
        text: status,
      }
    case 'CLOSED':
    case 'UNREACHABLE':
      return {
        bg: 'bg-rose-500/10 border-rose-500/20 text-rose-400',
        dot: 'bg-rose-500',
        text: status,
      }
    default:
      return {
        bg: 'bg-amber-500/10 border-amber-500/20 text-amber-400',
        dot: 'bg-amber-500',
        text: status || 'CHECKING',
      }
  }
}

function formatLatency(item) {
  const val = item?.latency_ms ?? item?.latency
  if (val != null && !isNaN(val) && val > 0) {
    return val.toFixed(1) + ' ms'
  }
  return '--'
}
</script>

<template>
  <div class="glass-panel rounded-2xl p-5 border border-slate-800 shadow-xl flex flex-col h-full">
    <div class="flex items-center justify-between border-b border-slate-800/80 pb-4 mb-4">
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-xl bg-gradient-to-tr from-cyan-500/20 to-indigo-500/20 border border-cyan-500/30 flex items-center justify-center text-cyan-400 shadow-sm">
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect>
            <rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect>
            <line x1="6" y1="6" x2="6.01" y2="6"></line>
            <line x1="6" y1="18" x2="6.01" y2="18"></line>
          </svg>
        </div>
        <div>
          <div class="flex items-center gap-2">
            <h3 class="text-sm font-bold text-slate-100">Service & Port Health</h3>
            <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-slate-800 text-slate-400 border border-slate-700">
              {{ portMatrix.length }} Ports
            </span>
          </div>
          <p class="text-xs text-slate-400 font-mono mt-0.5">
            Pengecekan live port & service internal via socket probe agentless
          </p>
        </div>
      </div>
    </div>

    <div class="flex-1 flex flex-col justify-center">
      <div v-if="portMatrix.length === 0" class="py-10 text-center text-slate-500 font-mono text-xs">
        Tidak ada port matrix yang terdeteksi untuk target ini.
      </div>

      <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-2.5 max-h-[300px] overflow-y-auto pr-1">
        <div
          v-for="(item, idx) in portMatrix"
          :key="idx"
          class="bg-slate-900/90 border border-slate-800 rounded-xl p-3 hover:border-slate-700 transition flex flex-col justify-between group min-w-0 shadow-sm"
        >
          <div class="flex items-start justify-between gap-2 min-w-0">
            <div class="min-w-0 flex-1">
              <div class="text-xs font-semibold text-slate-200 group-hover:text-cyan-300 transition truncate" :title="item.name">
                {{ item.name }}
              </div>
              <div class="text-[11px] font-mono text-slate-400 mt-0.5 truncate">
                Port: {{ item.port }}
              </div>
            </div>

            <!-- Status badge: shrink-0 and whitespace-nowrap to prevent overflowing outside card -->
            <div
              class="flex items-center gap-1.5 px-2 py-0.5 rounded-full border text-[10px] font-mono font-semibold shrink-0 whitespace-nowrap"
              :class="getStatusBadge(item.status).bg"
            >
              <span class="w-1.5 h-1.5 rounded-full shrink-0" :class="getStatusBadge(item.status).dot"></span>
              <span>{{ getStatusBadge(item.status).text }}</span>
            </div>
          </div>

          <div class="mt-2.5 pt-2 border-t border-slate-800/80 flex items-center justify-between text-[11px] font-mono text-slate-400">
            <span class="truncate">Latency:</span>
            <span class="text-slate-300 font-medium shrink-0 ml-1">
              {{ formatLatency(item) }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
