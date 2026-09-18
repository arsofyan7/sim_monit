<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const router = useRouter()

const name = ref('')
const email = ref('')
const password = ref('')
const errorMessage = ref('')
const isLoading = ref(false)

async function handleRegister() {
  errorMessage.value = ''
  isLoading.value = true
  try {
    await authStore.register(name.value, email.value, password.value)
    router.push('/dashboard')
  } catch (err) {
    errorMessage.value = err.message
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 relative overflow-hidden bg-dark-950">
    <div class="absolute top-1/4 right-1/4 w-96 h-96 bg-cyan-500/10 rounded-full blur-3xl pointer-events-none"></div>
    <div class="absolute bottom-1/4 left-1/4 w-96 h-96 bg-emerald-500/10 rounded-full blur-3xl pointer-events-none"></div>

    <div class="w-full max-w-md relative z-10">
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-slate-900 border border-cyan-500/30 shadow-xl shadow-cyan-500/10 mb-4">
          <svg class="w-8 h-8 text-cyan-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"></path>
          </svg>
        </div>
        <h1 class="text-2xl font-black tracking-tight text-white">
          SIM_MONIT
        </h1>
        <p class="text-xs text-slate-400 font-mono mt-1">
          Agentless Infrastructure & Service Observer
        </p>
      </div>

      <div class="glass-panel rounded-2xl p-8 border border-slate-800 shadow-2xl">
        <h2 class="text-base font-bold text-slate-100 mb-1">Buat Akun Baru</h2>
        <p class="text-xs text-slate-400 mb-6">Mulai memantau infrastruktur Anda secara agentless</p>

        <div v-if="errorMessage" class="mb-4 p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs font-mono">
          {{ errorMessage }}
        </div>

        <form @submit.prevent="handleRegister" class="space-y-4">
          <div>
            <label class="block text-xs font-mono uppercase text-slate-400 mb-1">Nama Lengkap</label>
            <input
              v-model="name"
              type="text"
              required
              class="w-full px-3.5 py-2.5 rounded-xl bg-slate-900 border border-slate-700 text-slate-100 text-sm focus:border-cyan-500 focus:outline-none transition"
              placeholder="e.g. Aris Sofyan"
            />
          </div>

          <div>
            <label class="block text-xs font-mono uppercase text-slate-400 mb-1">Email</label>
            <input
              v-model="email"
              type="email"
              required
              class="w-full px-3.5 py-2.5 rounded-xl bg-slate-900 border border-slate-700 text-slate-100 text-sm focus:border-cyan-500 focus:outline-none transition"
              placeholder="user@siberhub.id"
            />
          </div>

          <div>
            <label class="block text-xs font-mono uppercase text-slate-400 mb-1">Password</label>
            <input
              v-model="password"
              type="password"
              required
              minlength="6"
              class="w-full px-3.5 py-2.5 rounded-xl bg-slate-900 border border-slate-700 text-slate-100 text-sm focus:border-cyan-500 focus:outline-none transition"
              placeholder="Minimal 6 karakter"
            />
          </div>

          <button
            type="submit"
            :disabled="isLoading"
            class="w-full py-2.5 px-4 rounded-xl bg-gradient-to-r from-cyan-500 to-emerald-500 hover:from-cyan-400 hover:to-emerald-400 text-dark-950 font-bold text-sm transition shadow-lg shadow-cyan-500/25 disabled:opacity-50 mt-2"
          >
            {{ isLoading ? 'Mendaftarkan...' : 'Daftar Sekarang' }}
          </button>
        </form>

        <div class="mt-6 pt-5 border-t border-slate-800 text-center text-xs text-slate-400">
          Sudah memiliki akun?
          <router-link to="/login" class="text-cyan-400 hover:text-cyan-300 font-semibold ml-1">
            Masuk di sini
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>
