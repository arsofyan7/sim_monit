import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('sim_monit_token') || '')
  const user = ref(JSON.parse(localStorage.getItem('sim_monit_user') || 'null'))
  const isAuthenticated = ref(!!token.value)

  async function login(email, password) {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    const data = await res.json()
    if (!res.ok) {
      throw new Error(data.error || 'Login gagal')
    }

    token.value = data.token
    user.value = data.user
    isAuthenticated.value = true
    localStorage.setItem('sim_monit_token', data.token)
    localStorage.setItem('sim_monit_user', JSON.stringify(data.user))
    return data
  }

  async function register(name, email, password) {
    const res = await fetch('/api/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, email, password }),
    })
    const data = await res.json()
    if (!res.ok) {
      throw new Error(data.error || 'Registrasi gagal')
    }

    token.value = data.token
    user.value = data.user
    isAuthenticated.value = true
    localStorage.setItem('sim_monit_token', data.token)
    localStorage.setItem('sim_monit_user', JSON.stringify(data.user))
    return data
  }

  async function fetchCurrentUser() {
    if (!token.value) return null
    try {
      const res = await fetch('/api/auth/me', {
        headers: { Authorization: `Bearer ${token.value}` },
      })
      if (res.ok) {
        const u = await res.json()
        user.value = u
        localStorage.setItem('sim_monit_user', JSON.stringify(u))
        return u
      } else {
        logout()
      }
    } catch {
      // Offline / network issue
    }
    return null
  }

  function logout() {
    token.value = ''
    user.value = null
    isAuthenticated.value = false
    localStorage.removeItem('sim_monit_token')
    localStorage.removeItem('sim_monit_user')
    fetch('/api/auth/logout', { method: 'POST' }).catch(() => {})
  }

  return {
    token,
    user,
    isAuthenticated,
    login,
    register,
    fetchCurrentUser,
    logout,
  }
})
