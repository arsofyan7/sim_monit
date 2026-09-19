<script setup>
import { computed } from 'vue'

const props = defineProps({
  incidents: { type: Array, default: () => [] },
  target: { type: Object, default: null },
})

const hasOngoingIncident = computed(() => {
  return props.incidents.some((i) => !i.resolved_at)
})

function formatDuration(durationSec, startedAt, resolvedAt) {
  let sec = Number(durationSec) || 0
  if (!resolvedAt && startedAt) {
    // Ongoing incident: calculate duration from startedAt to now
    const start = new Date(startedAt).getTime()
    sec = Math.max(1, Math.floor((Date.now() - start) / 1000))
  }

  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = sec % 60

  if (h > 0) {
    return `${h} Jam ${m} Menit`
  }
  if (m > 0) {
    return `${m} Menit ${s > 0 ? s + ' Dtk' : ''}`
  }
  return `${s} Detik`
}

function formatDateTime(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function formatTimeOnly(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}
</script>

<template>
  <div class="glass-panel rounded-2xl p-5 border border-slate-800 shadow-xl flex flex-col h-full">
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-slate-800/80 pb-4 mb-4">
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-xl bg-gradient-to-tr from-amber-500/20 to-rose-500/20 border border-amber-500/30 flex items-center justify-center text-amber-400 shadow-sm">
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
            <line x1="12" y1="9" x2="12" y2="13"></line>
            <line x1="12" y1="17" x2="12.01" y2="17"></line>
          </svg>
        </div>
        <div>
          <div class="flex items-center gap-2">
            <h3 class="text-sm font-bold text-slate-100">Riwayat Insiden & Downtime</h3>
            <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-slate-800 text-slate-400 border border-slate-700">
              {{ incidents.length }} Insiden
            </span>
          </div>
          <p class="text-xs text-slate-400 font-mono mt-0.5">
            Log gangguan operasional target dan durasi pemulihan
          </p>
        </div>
      </div>

      <!-- Live Status Pill -->
      <div>
        <span
          v-if="hasOngoingIncident"
          class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-mono font-semibold bg-rose-500/20 text-rose-400 border border-rose-500/40 animate-pulse shadow-sm shadow-rose-500/20"
        >
          <span class="w-2 h-2 rounded-full bg-rose-500"></span>
          Insiden Aktif
        </span>
        <span
          v-else
          class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-mono font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
        >
          <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
          Semua Normal
        </span>
      </div>
    </div>

    <!-- Body -->
    <div class="flex-1 flex flex-col justify-center">
      <!-- Empty State: No Incidents Recorded -->
      <div
        v-if="incidents.length === 0"
        class="py-10 flex flex-col items-center justify-center text-center text-slate-400 space-y-2"
      >
        <div class="w-12 h-12 rounded-2xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center justify-center mb-1">
          <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
            <polyline points="9 12 11 14 15 10"></polyline>
          </svg>
        </div>
        <p class="text-sm font-semibold text-slate-200">Tidak Ada Insiden Tercatat</p>
        <p class="text-xs font-mono text-slate-500 max-w-xs">
          Target beroperasi 100% normal. Belum ada catatan downtime yang terdeteksi.
        </p>
      </div>

      <!-- Incident List -->
      <div v-else class="space-y-2.5 max-h-[300px] overflow-y-auto pr-1">
        <div
          v-for="inc in incidents"
          :key="inc.id"
          class="p-3.5 rounded-xl border transition-all text-xs font-mono"
          :class="!inc.resolved_at ? 'bg-rose-950/20 border-rose-500/40 shadow-sm shadow-rose-500/10' : 'bg-slate-900/60 border-slate-800/80 hover:border-slate-700'"
        >
          <!-- Top Row: Status & Duration Badge -->
          <div class="flex items-center justify-between gap-2 mb-1.5">
            <div class="flex items-center gap-2">
              <span
                v-if="!inc.resolved_at"
                class="px-2 py-0.5 rounded text-[10px] uppercase font-bold bg-rose-500/20 text-rose-300 border border-rose-500/40 flex items-center gap-1.5"
              >
                <span class="w-1.5 h-1.5 rounded-full bg-rose-400 animate-ping"></span>
                SEDANG DOWN
              </span>
              <span
                v-else
                class="px-2 py-0.5 rounded text-[10px] uppercase font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
              >
                RESOLVED
              </span>

              <span class="text-slate-300 font-semibold">
                Downtime: {{ formatDuration(inc.duration_seconds, inc.started_at, inc.resolved_at) }}
              </span>
            </div>

            <!-- Cause Badge -->
            <span v-if="inc.cause" class="text-[10px] text-slate-400 truncate max-w-[200px]" :title="inc.cause">
              {{ inc.cause }}
            </span>
          </div>

          <!-- Time Range Details -->
          <div class="text-[11px] text-slate-400 flex items-center gap-1.5 mt-1">
            <svg class="w-3.5 h-3.5 text-slate-500 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"></circle>
              <polyline points="12 6 12 12 16 14"></polyline>
            </svg>
            <span v-if="!inc.resolved_at">
              Dimulai sejak: <strong class="text-slate-200">{{ formatDateTime(inc.started_at) }}</strong> (masih berlangsung)
            </span>
            <span v-else>
              Rentang: <strong class="text-slate-200">{{ formatDateTime(inc.started_at) }}</strong> — <strong class="text-slate-200">{{ formatTimeOnly(inc.resolved_at) }} WIB</strong>
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
