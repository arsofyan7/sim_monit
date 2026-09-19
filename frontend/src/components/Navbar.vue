<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useMonitorStore } from '../stores/monitor'
import { useThemeStore } from '../stores/theme'
import { useRouter } from 'vue-router'

const props = defineProps({
  onOpenCreateModal: { type: Function, required: true },
  onOpenManageModal: { type: Function, required: true },
  onOpenTelegramModal: { type: Function, default: () => {} },
})

const authStore = useAuthStore()
const monitorStore = useMonitorStore()
const themeStore = useThemeStore()
const router = useRouter()

const dropdownOpen = ref(false)
const dropdownRef = ref(null)

function toggleDropdown() {
  dropdownOpen.value = !dropdownOpen.value
}

function handleSelectTarget(id) {
  monitorStore.selectTarget(id)
  dropdownOpen.value = false
}

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

function handleClickOutside(e) {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target)) {
    dropdownOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <header class="sticky top-0 z-40 w-full border-b border-slate-800/80 bg-dark-950/80 backdrop-blur-xl">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between gap-4">
      <!-- Logo & Brand -->
      <div class="flex items-center gap-3">
        <div class="relative flex items-center justify-center w-10 h-10 rounded-xl bg-gradient-to-tr from-cyan-500/20 via-emerald-500/20 to-indigo-500/20 border border-cyan-500/30 shadow-lg shadow-cyan-500/10">
          <svg class="w-5 h-5 text-cyan-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"></path>
          </svg>
          <span class="absolute -top-0.5 -right-0.5 flex h-2.5 w-2.5">
            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
            <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500"></span>
          </span>
        </div>
        <div>
          <div class="flex items-center gap-2">
            <span class="text-lg font-extrabold tracking-wider bg-gradient-to-r from-cyan-400 via-emerald-400 to-teal-300 bg-clip-text text-transparent">
              SIM_MONIT
            </span>
            <span class="px-1.5 py-0.5 text-[10px] font-mono font-semibold uppercase tracking-wider rounded bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
              Agentless
            </span>
          </div>
          <p class="text-[11px] text-slate-400 font-mono tracking-tight hidden sm:block">
            Infrastructure & Service Observer
          </p>
        </div>
      </div>

      <!-- Center: Target Selector Dropdown -->
      <div class="relative" ref="dropdownRef" v-if="monitorStore.targets.length > 0">
        <button
          @click="toggleDropdown"
          class="flex items-center gap-2.5 px-3.5 py-1.5 rounded-lg bg-slate-900/90 border border-slate-700/80 hover:border-cyan-500/50 transition-all text-sm font-medium shadow-inner"
        >
          <!-- Status Indicator Dot -->
          <span class="relative flex h-2 w-2">
            <span
              v-if="monitorStore.selectedTarget?.status === 'ONLINE'"
              class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"
            ></span>
            <span
              class="relative inline-flex rounded-full h-2 w-2"
              :class="{
                'bg-emerald-500': monitorStore.selectedTarget?.status === 'ONLINE',
                'bg-rose-500': monitorStore.selectedTarget?.status === 'OFFLINE',
                'bg-amber-500': monitorStore.selectedTarget?.status === 'PENDING' || !monitorStore.selectedTarget?.status,
              }"
            ></span>
          </span>

          <!-- Type Icon / Name -->
          <span class="text-slate-300 font-semibold max-w-[140px] sm:max-w-[200px] truncate">
            {{ monitorStore.selectedTarget?.name || 'Pilih Target' }}
          </span>

          <span class="text-[10px] uppercase font-mono px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 border border-slate-700">
            {{ monitorStore.selectedTarget?.type }}
          </span>

          <!-- Chevron -->
          <svg
            class="w-4 h-4 text-slate-400 transition-transform duration-200"
            :class="{ 'rotate-180': dropdownOpen }"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        <!-- Dropdown Menu -->
        <div
          v-if="dropdownOpen"
          class="absolute left-0 mt-2 w-72 rounded-xl bg-slate-900 border border-slate-700/80 shadow-2xl shadow-black/80 py-2 z-50 animate-in fade-in slide-in-from-top-2 duration-150 backdrop-blur-xl"
        >
          <div class="px-3 py-1.5 text-[11px] font-mono uppercase tracking-wider text-slate-400 border-b border-slate-800">
            Daftar Target Monitoring ({{ monitorStore.targets.length }})
          </div>

          <div class="max-h-60 overflow-y-auto py-1">
            <button
              v-for="t in monitorStore.targets"
              :key="t.id"
              @click="handleSelectTarget(t.id)"
              class="w-full flex items-center justify-between px-3 py-2 text-left text-sm hover:bg-slate-800/80 transition-colors"
              :class="{ 'bg-cyan-500/10 text-cyan-300': monitorStore.selectedTargetId === t.id }"
            >
              <div class="flex items-center gap-2 min-w-0">
                <span
                  class="w-2 h-2 rounded-full shrink-0"
                  :class="{
                    'bg-emerald-500 shadow-emerald-500/50 shadow-sm': t.status === 'ONLINE',
                    'bg-rose-500 shadow-rose-500/50 shadow-sm': t.status === 'OFFLINE',
                    'bg-amber-500': t.status === 'PENDING' || !t.status,
                  }"
                ></span>
                <span class="truncate font-medium text-slate-200">{{ t.name }}</span>
              </div>
              <span class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 uppercase">
                {{ t.type }}
              </span>
            </button>
          </div>

          <div class="border-t border-slate-800 pt-1.5 mt-1 px-2 flex gap-1.5">
            <button
              @click="props.onOpenCreateModal(); dropdownOpen = false;"
              class="flex-1 text-center text-xs py-1.5 px-2 rounded bg-cyan-600/20 hover:bg-cyan-600/30 text-cyan-300 font-medium transition"
            >
              + Tambah Target
            </button>
            <button
              @click="props.onOpenManageModal(); dropdownOpen = false;"
              class="flex-1 text-center text-xs py-1.5 px-2 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 font-medium transition"
            >
              Kelola Semua
            </button>
          </div>
        </div>
      </div>

      <!-- Right Actions: Telegram Alert + Quick Manage + User Profile & Logout -->
      <div class="flex items-center gap-2 sm:gap-3">
        <button
          @click="props.onOpenTelegramModal"
          title="Pengaturan Notifikasi Telegram"
          class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg bg-sky-500/10 hover:bg-sky-500/20 text-sky-400 hover:text-sky-300 border border-sky-500/30 transition shadow-sm"
        >
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69a.2.2 0 00-.05-.18c-.06-.05-.14-.03-.21-.02-.09.02-1.49.95-4.22 2.79-.4.27-.76.41-1.08.4-.36-.01-1.04-.2-1.55-.37-.63-.2-1.12-.31-1.08-.66.02-.18.27-.36.74-.55 2.92-1.27 4.86-2.11 5.83-2.51 2.78-1.16 3.35-1.36 3.73-1.36.08 0 .27.02.39.12.1.08.13.19.14.27-.01.06.01.24 0 .37z"/>
          </svg>
          <span class="hidden sm:inline font-semibold">Telegram Alert</span>
        </button>

        <button
          @click="props.onOpenManageModal"
          class="hidden md:flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg bg-slate-800/80 hover:bg-slate-700 text-slate-300 hover:text-white border border-slate-700 transition"
        >
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 20h9"></path>
            <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path>
          </svg>
          Kelola Target
        </button>

        <!-- Theme Toggle (Dark / Light Mode) -->
        <button
          @click="themeStore.toggleTheme"
          :title="themeStore.isDark ? 'Ganti ke Light Mode' : 'Ganti ke Dark Mode'"
          class="p-2 rounded-lg bg-slate-800/80 hover:bg-slate-700 text-amber-400 dark:text-amber-300 border border-slate-700 transition flex items-center justify-center shadow-sm"
        >
          <!-- Sun icon when in dark mode (click to switch to light mode) -->
          <svg v-if="themeStore.isDark" class="w-4 h-4 text-amber-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="5"></circle>
            <line x1="12" y1="1" x2="12" y2="3"></line>
            <line x1="12" y1="21" x2="12" y2="23"></line>
            <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
            <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
            <line x1="1" y1="12" x2="3" y2="12"></line>
            <line x1="21" y1="12" x2="23" y2="12"></line>
            <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
            <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
          </svg>
          <!-- Moon icon when in light mode (click to switch to dark mode) -->
          <svg v-else class="w-4 h-4 text-indigo-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
          </svg>
        </button>

        <!-- User Profile Pill -->
        <div class="flex items-center gap-2.5 pl-2 border-l border-slate-800">
          <div class="w-8 h-8 rounded-lg bg-gradient-to-tr from-cyan-600 to-indigo-600 flex items-center justify-center font-bold text-xs text-white shadow-sm shadow-cyan-500/20">
            {{ authStore.user?.name ? authStore.user.name.charAt(0).toUpperCase() : 'U' }}
          </div>
          <div class="hidden sm:block text-left leading-tight">
            <div class="text-xs font-semibold text-slate-200 truncate max-w-[100px]">
              {{ authStore.user?.name || 'User' }}
            </div>
            <div class="text-[10px] text-slate-400 font-mono truncate max-w-[100px]">
              {{ authStore.user?.email || 'admin@sim.monit' }}
            </div>
          </div>
          <button
            @click="handleLogout"
            title="Keluar / Logout"
            class="p-1.5 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 transition"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
              <polyline points="16 17 21 12 16 7"></polyline>
              <line x1="21" y1="12" x2="9" y2="12"></line>
            </svg>
          </button>
        </div>
      </div>
    </div>
  </header>
</template>
