<script setup>
import { ref, watch, reactive } from 'vue'
import { useMonitorStore } from '../stores/monitor'

const props = defineProps({
  isOpen: { type: Boolean, default: false },
  targetToEdit: { type: Object, default: null },
})

const emit = defineEmits(['close', 'saved'])

const monitorStore = useMonitorStore()

const isSubmitting = ref(false)
const isTesting = ref(false)
const testResult = ref(null)
const errorMessage = ref('')

const form = reactive({
  name: '',
  type: 'website', // server, website, api, database
  host: '',
  port: 443,
  polling_interval: 60,
  config: {
    username: 'root',
    password: '',
    private_key: '',
    url: '',
    method: 'GET',
    headers: {},
    expected_status_code: 200,
    db_engine: 'postgres',
    db_name: '',
  },
})

// Header key-value pairs for API type
const headerList = ref([{ key: '', value: '' }])

watch(
  () => props.targetToEdit,
  (val) => {
    if (val) {
      form.name = val.name
      form.type = val.type
      form.host = val.host
      form.port = val.port
      form.polling_interval = val.polling_interval
      form.config = {
        username: val.config?.username || 'root',
        password: val.config?.password || '',
        private_key: val.config?.private_key || '',
        url: val.config?.url || val.host,
        method: val.config?.method || 'GET',
        headers: val.config?.headers || {},
        expected_status_code: val.config?.expected_status_code || 200,
        db_engine: val.config?.db_engine || 'postgres',
        db_name: val.config?.db_name || '',
      }
      if (val.config?.headers) {
        headerList.value = Object.entries(val.config.headers).map(([k, v]) => ({ key: k, value: v }))
      } else {
        headerList.value = [{ key: '', value: '' }]
      }
    } else {
      resetForm()
    }
    testResult.value = null
    errorMessage.value = ''
  },
  { immediate: true }
)

function resetForm() {
  form.name = ''
  form.type = 'website'
  form.host = 'https://'
  form.port = 443
  form.polling_interval = 60
  form.config = {
    username: 'root',
    password: '',
    private_key: '',
    url: '',
    method: 'GET',
    headers: {},
    expected_status_code: 200,
    db_engine: 'postgres',
    db_name: '',
  }
  headerList.value = [{ key: '', value: '' }]
}

function handleTypeChange(newType) {
  form.type = newType
  testResult.value = null
  if (newType === 'website') {
    form.host = form.host && form.host.startsWith('http') ? form.host : 'https://google.com'
    form.port = 443
    form.config.expected_status_code = 200
  } else if (newType === 'api') {
    form.host = form.host && form.host.startsWith('http') ? form.host : 'https://api.github.com'
    form.port = 443
    form.config.method = 'GET'
    form.config.expected_status_code = 200
  } else if (newType === 'server') {
    form.host = '192.168.1.100'
    form.port = 22
    form.config.username = 'root'
  } else if (newType === 'database') {
    form.host = '127.0.0.1'
    form.port = 5432
    form.config.db_engine = 'postgres'
  }
}

function addHeader() {
  headerList.value.push({ key: '', value: '' })
}

function removeHeader(index) {
  headerList.value.splice(index, 1)
}

function buildHeadersObject() {
  const headers = {}
  for (const h of headerList.value) {
    if (h.key.trim()) {
      headers[h.key.trim()] = h.value.trim()
    }
  }
  return headers
}

async function handleTestConnection() {
  testResult.value = null
  errorMessage.value = ''
  isTesting.value = true

  const headersObj = buildHeadersObject()
  form.config.headers = headersObj

  try {
    const payload = {
      type: form.type,
      host: form.host,
      port: Number(form.port) || 0,
      config: { ...form.config },
    }
    const res = await monitorStore.testConnection(payload)
    testResult.value = res
  } catch (err) {
    testResult.value = {
      success: false,
      message: err.message,
      latency_ms: 0,
    }
  } finally {
    isTesting.value = false
  }
}

async function handleSubmit() {
  errorMessage.value = ''
  isSubmitting.value = true

  const headersObj = buildHeadersObject()
  form.config.headers = headersObj
  if (form.type === 'website' || form.type === 'api') {
    form.config.url = form.host
  }

  const payload = {
    name: form.name,
    type: form.type,
    host: form.host,
    port: Number(form.port) || 0,
    polling_interval: Number(form.polling_interval) || 60,
    config: form.config,
  }

  try {
    if (props.targetToEdit) {
      await monitorStore.updateTarget(props.targetToEdit.id, payload)
    } else {
      await monitorStore.createTarget(payload)
    }
    emit('saved')
    emit('close')
  } catch (err) {
    errorMessage.value = err.message
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-dark-950/80 backdrop-blur-md">
    <div class="glass-panel w-full max-w-2xl rounded-2xl p-6 border border-slate-700/80 shadow-2xl overflow-y-auto max-h-[90vh]">
      <!-- Header -->
      <div class="flex items-center justify-between border-b border-slate-800 pb-4 mb-5">
        <div>
          <h2 class="text-lg font-bold text-slate-100 flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-cyan-400"></span>
            {{ targetToEdit ? 'Edit Target Monitoring' : 'Tambah Target Monitoring Baru' }}
          </h2>
          <p class="text-xs text-slate-400 font-mono mt-0.5">
            Konfigurasi parameter agentless polling collector
          </p>
        </div>
        <button @click="emit('close')" class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <form @submit.prevent="handleSubmit" class="space-y-4">
        <!-- Error Alert -->
        <div v-if="errorMessage" class="p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs font-mono">
          {{ errorMessage }}
        </div>

        <!-- Target Name -->
        <div>
          <label class="block text-xs font-mono uppercase text-slate-400 mb-1">Nama Target</label>
          <input
            v-model="form.name"
            required
            type="text"
            placeholder="misal: Server DB Prod / Web Portal"
            class="w-full px-3.5 py-2 rounded-xl bg-slate-900 border border-slate-700 text-slate-100 text-sm focus:border-cyan-500 focus:outline-none"
          />
        </div>

        <!-- Target Type Selector (Pills) -->
        <div>
          <label class="block text-xs font-mono uppercase text-slate-400 mb-1.5">Tipe Target</label>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
            <button
              type="button"
              v-for="t in [
                { id: 'website', label: 'Website', sub: 'HTTP/HTTPS' },
                { id: 'api', label: 'API Endpoint', sub: 'REST / Headers' },
                { id: 'server', label: 'Server Linux', sub: 'SSH Agentless' },
                { id: 'database', label: 'Database', sub: 'TCP Probe' },
              ]"
              :key="t.id"
              @click="handleTypeChange(t.id)"
              class="p-2.5 rounded-xl border text-left transition-all"
              :class="form.type === t.id ? 'bg-cyan-500/15 border-cyan-500/50 text-cyan-300' : 'bg-slate-900/60 border-slate-800 text-slate-400 hover:border-slate-700'"
            >
              <div class="text-xs font-bold">{{ t.label }}</div>
              <div class="text-[10px] font-mono text-slate-500">{{ t.sub }}</div>
            </button>
          </div>
        </div>

        <!-- Host / IP & Port -->
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <div class="sm:col-span-2">
            <label class="block text-xs font-mono uppercase text-slate-400 mb-1">
              {{ form.type === 'website' || form.type === 'api' ? 'URL Target' : 'Host / IP Address' }}
            </label>
            <input
              v-model="form.host"
              required
              type="text"
              :placeholder="form.type === 'website' ? 'https://example.com' : '192.168.1.50'"
              class="w-full px-3.5 py-2 rounded-xl bg-slate-900 border border-slate-700 text-slate-100 text-sm focus:border-cyan-500 focus:outline-none font-mono"
            />
          </div>
          <div>
            <label class="block text-xs font-mono uppercase text-slate-400 mb-1">Port</label>
            <input
              v-model="form.port"
              type="number"
              class="w-full px-3.5 py-2 rounded-xl bg-slate-900 border border-slate-700 text-slate-100 text-sm focus:border-cyan-500 focus:outline-none font-mono"
            />
          </div>
        </div>

        <!-- Config Fields for Server (SSH) -->
        <div v-if="form.type === 'server'" class="p-3.5 rounded-xl bg-slate-900/50 border border-slate-800 space-y-3">
          <div class="text-xs font-mono text-cyan-400 font-semibold uppercase">Kredensial SSH (Agentless)</div>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label class="block text-xs text-slate-400 mb-1">Username SSH</label>
              <input
                v-model="form.config.username"
                type="text"
                placeholder="root"
                class="w-full px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-700 text-sm text-slate-200 font-mono focus:border-cyan-500 focus:outline-none"
              />
            </div>
            <div>
              <label class="block text-xs text-slate-400 mb-1">Password SSH</label>
              <input
                v-model="form.config.password"
                type="password"
                placeholder="••••••••"
                class="w-full px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-700 text-sm text-slate-200 font-mono focus:border-cyan-500 focus:outline-none"
              />
            </div>
          </div>
          <div>
            <label class="block text-xs text-slate-400 mb-1">SSH Private Key (Opsional)</label>
            <textarea
              v-model="form.config.private_key"
              rows="2"
              placeholder="-----BEGIN OPENSSH PRIVATE KEY-----..."
              class="w-full px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-700 text-xs text-slate-200 font-mono focus:border-cyan-500 focus:outline-none"
            ></textarea>
          </div>
        </div>

        <!-- Config Fields for Website / API -->
        <div v-if="form.type === 'website' || form.type === 'api'" class="p-3.5 rounded-xl bg-slate-900/50 border border-slate-800 space-y-3">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label class="block text-xs text-slate-400 mb-1">Expected Status Code</label>
              <input
                v-model="form.config.expected_status_code"
                type="number"
                placeholder="200"
                class="w-full px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-700 text-sm text-slate-200 font-mono focus:border-cyan-500 focus:outline-none"
              />
            </div>
            <div v-if="form.type === 'api'">
              <label class="block text-xs text-slate-400 mb-1">HTTP Method</label>
              <select
                v-model="form.config.method"
                class="w-full px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-700 text-sm text-slate-200 font-mono focus:border-cyan-500 focus:outline-none"
              >
                <option value="GET">GET</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
                <option value="HEAD">HEAD</option>
              </select>
            </div>
          </div>

          <!-- API Headers builder -->
          <div v-if="form.type === 'api'">
            <div class="flex items-center justify-between mb-1.5">
              <label class="text-xs text-slate-400">Request Headers</label>
              <button @click="addHeader" type="button" class="text-[11px] text-cyan-400 hover:underline">
                + Tambah Header
              </button>
            </div>
            <div v-for="(h, idx) in headerList" :key="idx" class="flex gap-2 mb-2">
              <input
                v-model="h.key"
                type="text"
                placeholder="Key (e.g. Authorization)"
                class="w-1/2 px-2.5 py-1 text-xs rounded-lg bg-slate-900 border border-slate-700 text-slate-200 font-mono"
              />
              <input
                v-model="h.value"
                type="text"
                placeholder="Value (e.g. Bearer token)"
                class="w-1/2 px-2.5 py-1 text-xs rounded-lg bg-slate-900 border border-slate-700 text-slate-200 font-mono"
              />
              <button @click="removeHeader(idx)" type="button" class="p-1 text-slate-500 hover:text-rose-400">
                &times;
              </button>
            </div>
          </div>
        </div>

        <!-- Polling Interval -->
        <div>
          <label class="block text-xs font-mono uppercase text-slate-400 mb-1.5">Polling Interval</label>
          <div class="grid grid-cols-4 gap-2">
            <button
              type="button"
              v-for="sec in [10, 30, 60, 300]"
              :key="sec"
              @click="form.polling_interval = sec"
              class="py-2 rounded-xl border text-center font-mono text-xs font-medium transition"
              :class="form.polling_interval === sec ? 'bg-cyan-500/20 border-cyan-500/50 text-cyan-300' : 'bg-slate-900/60 border-slate-800 text-slate-400 hover:border-slate-700'"
            >
              {{ sec < 60 ? `${sec} detik` : `${sec / 60} menit` }}
            </button>
          </div>
        </div>

        <!-- Test Connection Result Box -->
        <div v-if="testResult" class="p-3 rounded-xl border text-xs font-mono" :class="testResult.success ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400' : 'bg-rose-500/10 border-rose-500/30 text-rose-400'">
          <div class="font-bold flex items-center justify-between">
            <span>{{ testResult.success ? 'Test Berhasil!' : 'Test Gagal' }}</span>
            <span v-if="testResult.latency_ms > 0">{{ testResult.latency_ms.toFixed(1) }} ms</span>
          </div>
          <div class="text-[11px] mt-0.5 text-slate-300">{{ testResult.message }}</div>
        </div>

        <!-- Action Buttons -->
        <div class="flex items-center justify-between pt-4 border-t border-slate-800">
          <!-- Test Connection Button -->
          <button
            @click="handleTestConnection"
            type="button"
            :disabled="isTesting || !form.host"
            class="flex items-center gap-2 px-3.5 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-mono transition disabled:opacity-50"
          >
            <svg v-if="isTesting" class="animate-spin h-3.5 w-3.5 text-cyan-400" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <svg v-else class="w-3.5 h-3.5 text-cyan-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
            </svg>
            {{ isTesting ? 'Testing...' : 'Test Connection' }}
          </button>

          <div class="flex items-center gap-2">
            <button
              @click="emit('close')"
              type="button"
              class="px-4 py-2 rounded-xl bg-slate-800/80 hover:bg-slate-700 text-slate-300 text-xs font-medium transition"
            >
              Batal
            </button>
            <button
              type="submit"
              :disabled="isSubmitting"
              class="px-5 py-2 rounded-xl bg-gradient-to-r from-cyan-500 to-emerald-500 hover:from-cyan-400 hover:to-emerald-400 text-dark-950 font-bold text-xs transition shadow-lg shadow-cyan-500/20 disabled:opacity-50"
            >
              {{ isSubmitting ? 'Menyimpan...' : 'Simpan Target' }}
            </button>
          </div>
        </div>
      </form>
    </div>
  </div>
</template>
