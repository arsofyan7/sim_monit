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
  <div class="glass-panel rounded-2xl p-5">
    <div class="flex items-center justify-between mb-4">
      <div>
        <h3 class="text-sm font-bold tracking-wider uppercase text-slate-200">
          Service & Port Health Matrix
        </h3>
        <p class="text-xs text-slate-400 mt-0.5">
          Pengecekan live port & service internal via socket probe agentless
        </p>
      </div>
      <span class="text-xs font-mono text-slate-400">
        {{ portMatrix.length }} Services Monitored
      </span>
    </div>

    <div v-if="portMatrix.length === 0" class="py-8 text-center text-slate-500 font-mono text-xs">
      Tidak ada port matrix yang terdeteksi untuk target ini.
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
      <div
        v-for="(item, idx) in portMatrix"
        :key="idx"
        class="bg-slate-900/90 border border-slate-800 rounded-xl p-3.5 hover:border-slate-700 transition flex flex-col justify-between group"
      >
        <div class="flex items-start justify-between gap-2">
          <div>
            <div class="text-xs font-semibold text-slate-200 group-hover:text-cyan-300 transition">
              {{ item.name }}
            </div>
            <div class="text-[11px] font-mono text-slate-400 mt-0.5">
              Port: {{ item.port }}
            </div>
          </div>

          <!-- Status badge -->
          <div
            class="flex items-center gap-1.5 px-2 py-0.5 rounded-full border text-[10px] font-mono font-semibold"
            :class="getStatusBadge(item.status).bg"
          >
            <span class="w-1.5 h-1.5 rounded-full" :class="getStatusBadge(item.status).dot"></span>
            <span>{{ getStatusBadge(item.status).text }}</span>
          </div>
        </div>

        <div class="mt-3 pt-2.5 border-t border-slate-800/80 flex items-center justify-between text-[11px] font-mono text-slate-400">
          <span>Latency Probe:</span>
          <span class="text-slate-300 font-medium">
            {{ formatLatency(item) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
