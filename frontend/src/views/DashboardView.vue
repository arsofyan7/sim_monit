<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useMonitorStore } from '../stores/monitor'
import Navbar from '../components/Navbar.vue'
import StatCard from '../components/StatCard.vue'
import ChartArea from '../components/ChartArea.vue'
import ServiceMatrix from '../components/ServiceMatrix.vue'
import TargetModal from '../components/TargetModal.vue'
import ManageTargetsModal from '../components/ManageTargetsModal.vue'

const monitorStore = useMonitorStore()

const isCreateModalOpen = ref(false)
const isManageModalOpen = ref(false)
const targetToEdit = ref(null)

// Custom Date Range Picker Modal state
const isCustomDateModalOpen = ref(false)
const customFrom = ref('')
const customTo = ref('')

let pollTimer = null

onMounted(async () => {
  await monitorStore.refreshAll()

  // Setup auto-refresh polling every 8 seconds
  pollTimer = setInterval(() => {
    if (monitorStore.selectedTargetId) {
      monitorStore.fetchTargetStats(monitorStore.selectedTargetId)
      monitorStore.fetchMetrics(monitorStore.selectedTargetId)
    }
  }, 8000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

watch(
  () => monitorStore.selectedTargetId,
  (newId) => {
    if (newId) {
      monitorStore.fetchTargetStats(newId)
      monitorStore.fetchMetrics(newId)
    }
  }
)

function openCreateModal() {
  targetToEdit.value = null
  isCreateModalOpen.value = true
}

function openEditModal(target) {
  targetToEdit.value = target
  isManageModalOpen.value = false
  isCreateModalOpen.value = true
}

function openManageModal() {
  isManageModalOpen.value = true
}

function applyCustomDateRange() {
  if (customFrom.value) {
    monitorStore.customDateFrom = customFrom.value
    monitorStore.customDateTo = customTo.value
    monitorStore.setTimeRange('custom')
    isCustomDateModalOpen.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-dark-950 flex flex-col selection:bg-cyan-500/30 selection:text-cyan-200">
    <!-- Top Navigation Bar -->
    <Navbar
      :onOpenCreateModal="openCreateModal"
      :onOpenManageModal="openManageModal"
    />

    <main class="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6">
      <!-- STATE 1: Empty State (Belum ada target) -->
      <div
        v-if="monitorStore.targets.length === 0 && !monitorStore.loading"
        class="glass-panel rounded-3xl p-12 text-center max-w-2xl mx-auto my-12 border border-slate-800"
      >
        <div class="w-16 h-16 rounded-2xl bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 mx-auto flex items-center justify-center mb-4">
          <svg class="w-8 h-8" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <rect width="20" height="8" x="2" y="2" rx="2" ry="2"></rect>
            <rect width="20" height="8" x="2" y="14" rx="2" ry="2"></rect>
            <line x1="6" x2="6.01" y1="6" y2="6"></line>
            <line x1="6" x2="6.01" y1="18" y2="18"></line>
          </svg>
        </div>

        <h2 class="text-xl font-bold text-white mb-2">
          Belum ada server atau service yang dipantau
        </h2>
        <p class="text-xs text-slate-400 font-mono max-w-md mx-auto mb-6">
          Tambahkan server Linux (SSH), website, REST API endpoint, atau database Anda untuk mulai memantau metrik performa secara agentless.
        </p>

        <button
          @click="openCreateModal"
          class="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-gradient-to-r from-cyan-500 to-emerald-500 hover:from-cyan-400 hover:to-emerald-400 text-dark-950 font-bold text-xs tracking-wider transition shadow-lg shadow-cyan-500/25"
        >
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
          + Tambah Target Monitoring
        </button>
      </div>

      <!-- STATE 2: Active Monitoring State -->
      <div v-else class="space-y-6">
        <!-- Control Header: Title & Time-Range Pill Selector -->
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <div class="flex items-center gap-3">
              <h1 class="text-xl font-black tracking-tight text-white flex items-center gap-2">
                {{ monitorStore.selectedTarget?.name || 'Overview' }}
              </h1>
              <span class="text-xs font-mono px-2 py-0.5 rounded-full bg-slate-800 text-slate-300 border border-slate-700 uppercase">
                {{ monitorStore.selectedTarget?.type }}
              </span>
            </div>
            <p class="text-xs text-slate-400 font-mono mt-0.5 truncate">
              Endpoint: {{ monitorStore.selectedTarget?.host }}
              <span v-if="monitorStore.selectedTarget?.port">:{{ monitorStore.selectedTarget?.port }}</span>
              • Polling: {{ monitorStore.selectedTarget?.polling_interval }}s
            </p>
          </div>

          <!-- Time-Range Pills Filter & Custom Picker -->
          <div class="flex items-center gap-2">
            <div class="inline-flex rounded-xl bg-slate-900 border border-slate-800 p-1 text-xs font-mono">
              <button
                v-for="range in ['1h', '24h', '7d', 'mtd', 'ytd']"
                :key="range"
                @click="monitorStore.setTimeRange(range)"
                class="px-3 py-1 rounded-lg uppercase transition-all font-semibold"
                :class="monitorStore.selectedTimeRange === range ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/30 shadow-sm' : 'text-slate-400 hover:text-slate-200'"
              >
                {{ range }}
              </button>

              <button
                @click="isCustomDateModalOpen = true"
                class="px-3 py-1 rounded-lg uppercase transition-all font-semibold flex items-center gap-1"
                :class="monitorStore.selectedTimeRange === 'custom' ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/30' : 'text-slate-400 hover:text-slate-200'"
              >
                <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                  <line x1="16" y1="2" x2="16" y2="6"></line>
                  <line x1="8" y1="2" x2="8" y2="6"></line>
                  <line x1="3" y1="10" x2="21" y2="10"></line>
                </svg>
                Custom
              </button>
            </div>
          </div>
        </div>

        <!-- Section 2: Top Stat Cards Bar -->
        <StatCard
          :stats="monitorStore.stats"
          :target="monitorStore.selectedTarget"
        />

        <!-- Section 3: Main Analytics Chart Area -->
        <ChartArea
          :target="monitorStore.selectedTarget"
          :metrics="monitorStore.metricsData"
          :scale="monitorStore.metricsScale"
          :loading="monitorStore.loading"
          v-model:chartMode="monitorStore.chartMode"
        />

        <!-- Section 4: Service & Port Health Matrix -->
        <ServiceMatrix
          :portMatrix="monitorStore.stats?.port_matrix || monitorStore.stats?.portMatrix || []"
        />
      </div>
    </main>

    <!-- Custom Date Range Picker Modal -->
    <div v-if="isCustomDateModalOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-dark-950/80 backdrop-blur-md">
      <div class="glass-panel w-full max-w-sm rounded-2xl p-5 border border-slate-800 shadow-2xl">
        <h3 class="text-sm font-bold text-slate-100 mb-1">Pilih Rentang Waktu Custom</h3>
        <p class="text-xs text-slate-400 font-mono mb-4">Tentukan tanggal mulai dan selesai analitik</p>

        <div class="space-y-3">
          <div>
            <label class="block text-xs font-mono text-slate-400 mb-1">Dari Tanggal (From)</label>
            <input
              v-model="customFrom"
              type="date"
              class="w-full px-3 py-2 rounded-xl bg-slate-900 border border-slate-700 text-xs font-mono text-slate-200 focus:border-cyan-500 focus:outline-none"
            />
          </div>
          <div>
            <label class="block text-xs font-mono text-slate-400 mb-1">Sampai Tanggal (To)</label>
            <input
              v-model="customTo"
              type="date"
              class="w-full px-3 py-2 rounded-xl bg-slate-900 border border-slate-700 text-xs font-mono text-slate-200 focus:border-cyan-500 focus:outline-none"
            />
          </div>
        </div>

        <div class="flex items-center justify-end gap-2 mt-5">
          <button
            @click="isCustomDateModalOpen = false"
            class="px-3 py-1.5 rounded-lg bg-slate-800 text-slate-300 text-xs font-medium"
          >
            Batal
          </button>
          <button
            @click="applyCustomDateRange"
            :disabled="!customFrom"
            class="px-4 py-1.5 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-dark-950 font-bold text-xs disabled:opacity-50"
          >
            Terapkan Filter
          </button>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <TargetModal
      :isOpen="isCreateModalOpen"
      :targetToEdit="targetToEdit"
      @close="isCreateModalOpen = false"
      @saved="monitorStore.refreshAll"
    />

    <ManageTargetsModal
      :isOpen="isManageModalOpen"
      @close="isManageModalOpen = false"
      @openCreate="isManageModalOpen = false; openCreateModal()"
      @openEdit="openEditModal"
    />
  </div>
</template>
