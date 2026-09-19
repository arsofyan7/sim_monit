<script setup>
import { ref, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useMonitorStore } from '../stores/monitor'

const props = defineProps({
  isOpen: { type: Boolean, default: false },
})

const emit = defineEmits(['close'])

const authStore = useAuthStore()
const monitorStore = useMonitorStore()

const isLoading = ref(false)
const isSaving = ref(false)
const isTesting = ref(false)
const showToken = ref(false)
const saveSuccess = ref(false)
const testResult = ref(null) // { success: boolean, message: string }
const activeTestTargetId = ref('global') // 'global' or target.id

const form = ref({
  telegram_enabled: false,
  telegram_bot_token: '',
  telegram_mode: 'all', // 'all' or 'per_target'
  telegram_chat_id: '',
  target_mappings: {},
})

// Initialize target mappings helper
function ensureTargetMappings() {
  const mappings = { ...(form.value.target_mappings || {}) }
  for (const t of monitorStore.targets) {
    if (!mappings[t.id]) {
      mappings[t.id] = {
        enabled: true,
        chat_id: form.value.telegram_chat_id || '',
      }
    }
  }
  form.value.target_mappings = mappings
}

async function fetchSettings() {
  if (!authStore.token) return
  isLoading.value = true
  testResult.value = null
  saveSuccess.value = false

  try {
    const res = await fetch('/api/notifications/telegram', {
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    if (res.ok) {
      const data = await res.json()
      form.value = {
        telegram_enabled: data.telegram_enabled ?? false,
        telegram_bot_token: data.telegram_bot_token || '',
        telegram_mode: data.telegram_mode || 'all',
        telegram_chat_id: data.telegram_chat_id || '',
        target_mappings: data.target_mappings || {},
      }
      ensureTargetMappings()
    }
  } catch (err) {
    console.error('Failed to fetch telegram settings:', err)
  } finally {
    isLoading.value = false
  }
}

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      fetchSettings()
    }
  }
)

async function handleSave() {
  if (!authStore.token) return
  isSaving.value = true
  saveSuccess.value = false
  testResult.value = null

  try {
    const res = await fetch('/api/notifications/telegram', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify(form.value),
    })

    const data = await res.json()
    if (!res.ok) {
      throw new Error(data.error || 'Gagal menyimpan pengaturan')
    }

    saveSuccess.value = true
    setTimeout(() => {
      saveSuccess.value = false
    }, 3000)
  } catch (err) {
    alert(err.message)
  } finally {
    isSaving.value = false
  }
}

async function handleTestNotification() {
  if (!form.value.telegram_bot_token) {
    testResult.value = {
      success: false,
      message: 'Harap masukkan Telegram Bot Token terlebih dahulu.',
    }
    return
  }

  let chatIDToTest = form.value.telegram_chat_id
  if (form.value.telegram_mode === 'per_target') {
    if (activeTestTargetId.value !== 'global' && form.value.target_mappings[activeTestTargetId.value]) {
      chatIDToTest = form.value.target_mappings[activeTestTargetId.value].chat_id
    } else {
      // Find first target with non-empty chat_id
      const firstTarget = monitorStore.targets.find(
        (t) => form.value.target_mappings[t.id]?.chat_id
      )
      if (firstTarget) {
        chatIDToTest = form.value.target_mappings[firstTarget.id].chat_id
      }
    }
  }

  if (!chatIDToTest) {
    testResult.value = {
      success: false,
      message: 'Harap tentukan Chat ID penerima sebelum melakukan tes.',
    }
    return
  }

  isTesting.value = true
  testResult.value = null

  try {
    const res = await fetch('/api/notifications/telegram/test', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify({
        bot_token: form.value.telegram_bot_token,
        chat_id: chatIDToTest,
      }),
    })

    const data = await res.json()
    if (!res.ok) {
      testResult.value = {
        success: false,
        message: data.error || 'Pengujian gagal dikirim ke Telegram.',
      }
    } else {
      testResult.value = {
        success: true,
        message: `Pesan tes berhasil dikirim ke Chat ID: ${chatIDToTest}! Silakan cek Telegram Anda.`,
      }
    }
  } catch (err) {
    testResult.value = {
      success: false,
      message: 'Gagal menghubungi server: ' + err.message,
    }
  } finally {
    isTesting.value = false
  }
}
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-dark-950/80 backdrop-blur-md">
    <div class="glass-panel w-full max-w-2xl rounded-2xl p-6 border border-slate-700/80 shadow-2xl overflow-hidden max-h-[90vh] flex flex-col">
      <!-- Modal Header -->
      <div class="flex items-center justify-between border-b border-slate-800 pb-4 mb-4">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-tr from-sky-500/20 to-blue-600/20 border border-sky-500/30 flex items-center justify-center text-sky-400">
            <!-- Telegram Paper Plane Icon -->
            <svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69a.2.2 0 00-.05-.18c-.06-.05-.14-.03-.21-.02-.09.02-1.49.95-4.22 2.79-.4.27-.76.41-1.08.4-.36-.01-1.04-.2-1.55-.37-.63-.2-1.12-.31-1.08-.66.02-.18.27-.36.74-.55 2.92-1.27 4.86-2.11 5.83-2.51 2.78-1.16 3.35-1.36 3.73-1.36.08 0 .27.02.39.12.1.08.13.19.14.27-.01.06.01.24 0 .37z"/>
            </svg>
          </div>
          <div>
            <h2 class="text-lg font-bold text-slate-100 flex items-center gap-2">
              Pengaturan Notifikasi Telegram
            </h2>
            <p class="text-xs text-slate-400 font-mono mt-0.5">
              Kirim peringatan instan saat target OFFLINE dan kembali ONLINE
            </p>
          </div>
        </div>

        <button
          @click="emit('close')"
          class="p-1.5 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-slate-800 transition"
        >
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>

      <!-- Loading State -->
      <div v-if="isLoading" class="py-12 flex flex-col items-center justify-center text-slate-400 gap-2">
        <svg class="animate-spin h-6 w-6 text-sky-400" viewBox="0 0 24 24" fill="none">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <span class="text-xs font-mono">Memuat konfigurasi...</span>
      </div>

      <!-- Modal Body -->
      <div v-else class="flex-1 overflow-y-auto space-y-5 pr-1 text-slate-200 text-sm">
        <!-- 1. Master Toggle -->
        <div class="p-4 rounded-xl bg-slate-900/80 border border-slate-800 flex items-center justify-between">
          <div class="space-y-0.5">
            <span class="text-sm font-semibold text-slate-100 flex items-center gap-2">
              Aktifkan Alert Telegram
              <span
                class="px-2 py-0.5 text-[10px] font-mono rounded uppercase"
                :class="form.telegram_enabled ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30' : 'bg-slate-800 text-slate-400'"
              >
                {{ form.telegram_enabled ? 'Aktif' : 'Non-Aktif' }}
              </span>
            </span>
            <p class="text-xs text-slate-400 font-mono">
              Nyalakan untuk mengirim alert otomatis saat terjadi transisi status target
            </p>
          </div>

          <!-- Switch Slider -->
          <label class="relative inline-flex items-center cursor-pointer">
            <input
              type="checkbox"
              v-model="form.telegram_enabled"
              class="sr-only peer"
            />
            <div class="w-11 h-6 bg-slate-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-sky-500"></div>
          </label>
        </div>

        <!-- 2. Bot Token Input -->
        <div class="space-y-1.5">
          <div class="flex items-center justify-between">
            <label class="block text-xs font-mono uppercase tracking-wider text-slate-300 font-semibold">
              Telegram Bot Token
            </label>
            <a
              href="https://t.me/BotFather"
              target="_blank"
              rel="noopener noreferrer"
              class="text-[11px] text-sky-400 hover:text-sky-300 underline font-mono flex items-center gap-1"
            >
              Buat bot di @BotFather
              <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path>
                <polyline points="15 3 21 3 21 9"></polyline>
                <line x1="10" y1="14" x2="21" y2="3"></line>
              </svg>
            </a>
          </div>

          <div class="relative">
            <input
              :type="showToken ? 'text' : 'password'"
              v-model="form.telegram_bot_token"
              placeholder="123456789:ABCDefghIJKlmNoPQRstuVWXyz..."
              class="w-full px-3.5 py-2.5 pr-10 rounded-xl bg-slate-900 border border-slate-700/80 focus:border-sky-500 focus:ring-1 focus:ring-sky-500 outline-none font-mono text-xs text-slate-200 placeholder-slate-600 transition"
            />
            <button
              type="button"
              @click="showToken = !showToken"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-200 text-xs font-mono"
            >
              {{ showToken ? 'Hide' : 'Show' }}
            </button>
          </div>
          <p class="text-[11px] text-slate-500 font-mono">
            * Token bot disimpan secara aman di database lokal SIM_MONIT.
          </p>
        </div>

        <!-- 3. Routing Mode Selection -->
        <div class="space-y-2">
          <label class="block text-xs font-mono uppercase tracking-wider text-slate-300 font-semibold">
            Pilihan Distribusi Chat ID
          </label>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <label
              class="flex items-start gap-3 p-3.5 rounded-xl border cursor-pointer transition-all"
              :class="form.telegram_mode === 'all' ? 'bg-sky-500/10 border-sky-500/40 text-sky-200 shadow-sm' : 'bg-slate-900/60 border-slate-800 text-slate-400 hover:border-slate-700'"
            >
              <input
                type="radio"
                value="all"
                v-model="form.telegram_mode"
                class="mt-0.5 text-sky-500 focus:ring-sky-500"
              />
              <div class="text-xs">
                <div class="font-bold text-slate-200 mb-0.5">Semua Target (Global Chat ID)</div>
                <div class="text-slate-400 font-mono text-[11px] leading-relaxed">
                  Satu Chat ID yang sama untuk seluruh notifikasi target monitoring.
                </div>
              </div>
            </label>

            <label
              class="flex items-start gap-3 p-3.5 rounded-xl border cursor-pointer transition-all"
              :class="form.telegram_mode === 'per_target' ? 'bg-sky-500/10 border-sky-500/40 text-sky-200 shadow-sm' : 'bg-slate-900/60 border-slate-800 text-slate-400 hover:border-slate-700'"
            >
              <input
                type="radio"
                value="per_target"
                v-model="form.telegram_mode"
                class="mt-0.5 text-sky-500 focus:ring-sky-500"
              />
              <div class="text-xs">
                <div class="font-bold text-slate-200 mb-0.5">Tentukan Per-Target</div>
                <div class="text-slate-400 font-mono text-[11px] leading-relaxed">
                  Konfigurasikan Chat ID spesifik untuk masing-masing target (grup/channel berbeda).
                </div>
              </div>
            </label>
          </div>
        </div>

        <!-- Mode A: Global Chat ID Input -->
        <div v-if="form.telegram_mode === 'all'" class="space-y-1.5 p-4 rounded-xl bg-slate-900/60 border border-slate-800">
          <label class="block text-xs font-mono uppercase tracking-wider text-slate-300 font-semibold">
            Global Telegram Chat ID
          </label>
          <input
            type="text"
            v-model="form.telegram_chat_id"
            placeholder="Contoh: -1001234567890 (Grup) atau 12345678 (Akun)"
            class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-700/80 focus:border-sky-500 focus:ring-1 focus:ring-sky-500 outline-none font-mono text-xs text-slate-200 placeholder-slate-600 transition"
          />
          <p class="text-[11px] text-slate-400 font-mono leading-relaxed">
            Tips: Pastikan bot telah dimasukkan ke dalam grup/channel target dan diberi izin kirim pesan. Chat ID akun pribadi bisa dicek via bot <code class="text-sky-300">@userinfobot</code>.
          </p>
        </div>

        <!-- Mode B: Per-Target Routing Table -->
        <div v-else class="space-y-3 p-4 rounded-xl bg-slate-900/60 border border-slate-800">
          <div class="flex items-center justify-between">
            <label class="block text-xs font-mono uppercase tracking-wider text-slate-300 font-semibold">
              Daftar Target & Chat ID Spesifik
            </label>
            <span class="text-[11px] font-mono text-slate-400">
              Total: {{ monitorStore.targets.length }} Target
            </span>
          </div>

          <div v-if="monitorStore.targets.length === 0" class="text-center py-6 text-slate-500 font-mono text-xs">
            Belum ada target yang dibuat di SIM_MONIT.
          </div>

          <div v-else class="space-y-2.5 max-h-60 overflow-y-auto pr-1">
            <div
              v-for="target in monitorStore.targets"
              :key="target.id"
              class="p-3 rounded-xl bg-slate-950/80 border border-slate-800/90 flex flex-col sm:flex-row sm:items-center justify-between gap-3"
            >
              <div class="flex items-center gap-3 min-w-[180px]">
                <input
                  type="checkbox"
                  v-model="form.target_mappings[target.id].enabled"
                  class="rounded bg-slate-900 border-slate-700 text-sky-500 focus:ring-sky-500 w-4 h-4 cursor-pointer"
                />
                <div class="truncate">
                  <div class="font-medium text-slate-200 text-xs truncate">{{ target.name }}</div>
                  <div class="text-[10px] font-mono text-slate-400 truncate">
                    {{ target.host }}<span v-if="target.port">:{{ target.port }}</span> • <span class="uppercase text-sky-400">{{ target.type }}</span>
                  </div>
                </div>
              </div>

              <!-- Chat ID Input for Target -->
              <div class="flex-1 max-w-sm">
                <input
                  type="text"
                  v-model="form.target_mappings[target.id].chat_id"
                  :disabled="!form.target_mappings[target.id].enabled"
                  placeholder="Chat ID (misal: -100123456789)"
                  class="w-full px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-700/80 focus:border-sky-500 outline-none font-mono text-xs text-slate-200 placeholder-slate-600 disabled:opacity-40 disabled:cursor-not-allowed transition"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Test Notification Alert Box -->
        <div v-if="testResult" class="p-3.5 rounded-xl border text-xs font-mono leading-relaxed" :class="testResult.success ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300' : 'bg-rose-500/10 border-rose-500/30 text-rose-300'">
          <div class="flex items-start gap-2">
            <span v-if="testResult.success">✓</span>
            <span v-else>✕</span>
            <span>{{ testResult.message }}</span>
          </div>
        </div>

        <!-- Save Success Alert Box -->
        <div v-if="saveSuccess" class="p-3.5 rounded-xl border bg-emerald-500/10 border-emerald-500/30 text-emerald-300 text-xs font-mono flex items-center gap-2">
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="20 6 9 17 4 12"></polyline>
          </svg>
          Pengaturan notifikasi Telegram berhasil disimpan!
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="border-t border-slate-800 pt-4 mt-4 flex flex-col sm:flex-row items-center justify-between gap-3">
        <!-- Left: Test Button -->
        <button
          type="button"
          @click="handleTestNotification"
          :disabled="isTesting || !form.telegram_bot_token"
          class="w-full sm:w-auto px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-xs font-semibold flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed transition"
        >
          <svg v-if="isTesting" class="animate-spin h-3.5 w-3.5 text-slate-300" viewBox="0 0 24 24" fill="none">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <svg v-else class="w-3.5 h-3.5 text-sky-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="22" y1="2" x2="11" y2="13"></line>
            <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
          </svg>
          {{ isTesting ? 'Mengirim pesan tes...' : 'Test Kirim Pesan' }}
        </button>

        <!-- Right: Cancel & Save Buttons -->
        <div class="w-full sm:w-auto flex items-center gap-2.5 justify-end">
          <button
            type="button"
            @click="emit('close')"
            class="px-4 py-2 rounded-xl bg-slate-800/80 hover:bg-slate-800 text-slate-400 hover:text-slate-200 text-xs font-semibold transition"
          >
            Tutup
          </button>
          <button
            type="button"
            @click="handleSave"
            :disabled="isSaving"
            class="px-5 py-2 rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 hover:from-sky-400 hover:to-blue-500 text-white font-semibold text-xs shadow-lg shadow-sky-500/20 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 transition"
          >
            <svg v-if="isSaving" class="animate-spin h-3.5 w-3.5 text-white" viewBox="0 0 24 24" fill="none">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            {{ isSaving ? 'Menyimpan...' : 'Simpan Pengaturan' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
