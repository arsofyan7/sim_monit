import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  // Read saved theme or default to 'dark'
  const savedTheme = localStorage.getItem('sim_monit_theme') || 'dark'
  const currentTheme = ref(savedTheme)
  const isDark = ref(savedTheme === 'dark')

  function applyTheme(theme) {
    currentTheme.value = theme
    isDark.value = theme === 'dark'
    localStorage.setItem('sim_monit_theme', theme)

    const root = document.documentElement
    if (theme === 'dark') {
      root.classList.add('dark')
      root.classList.remove('light')
    } else {
      root.classList.remove('dark')
      root.classList.add('light')
    }
  }

  function toggleTheme() {
    applyTheme(isDark.value ? 'light' : 'dark')
  }

  // Initial apply
  applyTheme(currentTheme.value)

  return {
    currentTheme,
    isDark,
    toggleTheme,
    applyTheme,
  }
})
