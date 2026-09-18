<script setup>
import { ref } from 'vue'
import { useMonitorStore } from '../stores/monitor'

const props = defineProps({
  isOpen: { type: Boolean, default: false },
})

const emit = defineEmits(['close', 'openCreate', 'openEdit'])

const monitorStore = useMonitorStore()

const targetToDelete = ref(null)
const purgeHistory = ref(false)
const isDeleting = ref(false)

function promptDelete(target) {
  targetToDelete.value = target
  purgeHistory.value = false
}

async function confirmDelete() {
  if (!targetToDelete.value) return
  isDeleting.value = true
  try {
    await monitorStore.deleteTarget(targetToDelete.value.id, purgeHistory.value)
    targetToDelete.value = null
  } catch (err) {
    alert(err.message)
  } finally {
    isDeleting.value = false
  }
}
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-dark-950/80 backdrop-blur-md">
    <div class="glass-panel w-full max-w-4xl rounded-2xl p-6 border border-slate-700/80 shadow-2xl overflow-hidden max-h-[90vh] flex flex-col">
      <!-- Modal Header -->
      <div class="flex items-center justify-between border-b border-slate-800 pb-4 mb-4">
        <div>
          <h2 class="text-lg font-bold text-slate-100 flex items-center gap-2">
            <svg class="w-5 h-5 text-cyan-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 20h9"></path>
              <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path>
            </svg>
            Kelola Target Monitoring
          </h2>
          <p class="text-xs text-slate-400 font-mono mt-0.5">
            Daftar server, web endpoint, API dan database yang dipantau
          </p>
        </div>

        <div class="flex items-center gap-2">
          <button
            @click="emit('openCreate')"
            class="px-3 py-1.5 rounded-xl bg-cyan-500/20 hover:bg-cyan-500/30 text-cyan-300 border border-cyan-500/30 text-xs font-semibold transition"
          >
            + Tambah Target
          </button>
          <button @click="emit('close')" class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Targets Table -->
      <div class="overflow-x-auto flex-1">
        <table class="w-full text-left text-xs">
          <thead>
            <tr class="border-b border-slate-800 text-slate-400 font-mono uppercase text-[11px]">
              <th class="py-3 px-3">Nama Target</th>
              <th class="py-3 px-3">Tipe</th>
              <th class="py-3 px-3">Host / Endpoint</th>
              <th class="py-3 px-3">Interval</th>
              <th class="py-3 px-3">Status</th>
              <th class="py-3 px-3 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 font-mono">
            <tr v-if="monitorStore.targets.length === 0">
              <td colspan="6" class="py-8 text-center text-slate-500">
                Belum ada target monitoring. Klik "+ Tambah Target" untuk memulai.
              </td>
            </tr>
            <tr
              v-for="t in monitorStore.targets"
              :key="t.id"
              class="hover:bg-slate-800/40 transition group"
            >
              <td class="py-3 px-3 font-semibold text-slate-200">
                {{ t.name }}
              </td>
              <td class="py-3 px-3">
                <span class="px-2 py-0.5 rounded bg-slate-800 text-slate-300 text-[10px] uppercase border border-slate-700">
                  {{ t.type }}
                </span>
              </td>
              <td class="py-3 px-3 text-slate-400 max-w-[200px] truncate">
                {{ t.host }}<span v-if="t.port">:{{ t.port }}</span>
              </td>
              <td class="py-3 px-3 text-slate-400">
                {{ t.polling_interval }}s
              </td>
              <td class="py-3 px-3">
                <span
                  class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-bold"
                  :class="{
                    'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20': t.status === 'ONLINE',
                    'bg-rose-500/10 text-rose-400 border border-rose-500/20': t.status === 'OFFLINE',
                    'bg-amber-500/10 text-amber-400 border border-amber-500/20': t.status === 'PENDING' || !t.status,
                  }"
                >
                  <span class="w-1.5 h-1.5 rounded-full" :class="{
                    'bg-emerald-500': t.status === 'ONLINE',
                    'bg-rose-500': t.status === 'OFFLINE',
                    'bg-amber-500': t.status === 'PENDING' || !t.status,
                  }"></span>
                  {{ t.status || 'PENDING' }}
                </span>
              </td>
              <td class="py-3 px-3 text-right">
                <div class="flex items-center justify-end gap-2">
                  <button
                    @click="emit('openEdit', t)"
                    class="p-1 rounded bg-slate-800 hover:bg-cyan-500/20 text-slate-300 hover:text-cyan-300 transition"
                    title="Edit Target"
                  >
                    <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                      <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                    </svg>
                  </button>
                  <button
                    @click="promptDelete(t)"
                    class="p-1 rounded bg-slate-800 hover:bg-rose-500/20 text-slate-300 hover:text-rose-400 transition"
                    title="Hapus Target"
                  >
                    <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="3 6 5 6 21 6"></polyline>
                      <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                    </svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Delete Confirmation Sub-Modal -->
      <div v-if="targetToDelete" class="mt-4 p-4 rounded-xl bg-rose-500/10 border border-rose-500/30">
        <div class="text-sm font-bold text-rose-300">
          Konfirmasi Hapus Target: "{{ targetToDelete.name }}"?
        </div>
        <p class="text-xs text-slate-300 mt-1">
          Aksi ini akan menghentikan background collector untuk target ini.
        </p>

        <label class="flex items-center gap-2 mt-2 text-xs font-mono text-slate-300 cursor-pointer">
          <input type="checkbox" v-model="purgeHistory" class="rounded bg-slate-900 border-slate-700 text-rose-500 focus:ring-0" />
          Hapus seluruh data riwayat & metrics historis (Raw, Hourly, Daily)
        </label>

        <div class="flex items-center justify-end gap-2 mt-3">
          <button
            @click="targetToDelete = null"
            class="px-3 py-1.5 rounded-lg bg-slate-800 text-slate-300 text-xs font-medium"
          >
            Batal
          </button>
          <button
            @click="confirmDelete"
            :disabled="isDeleting"
            class="px-3.5 py-1.5 rounded-lg bg-rose-600 hover:bg-rose-500 text-white text-xs font-bold transition disabled:opacity-50"
          >
            {{ isDeleting ? 'Menghapus...' : 'Ya, Hapus Target' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
