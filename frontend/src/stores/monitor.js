import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from './auth'

export const useMonitorStore = defineStore('monitor', () => {
  const authStore = useAuthStore()

  const targets = ref([])
  const selectedTargetId = ref(null)
  const selectedTimeRange = ref('1h') // '1h', '24h', '7d', 'mtd', 'ytd', 'custom'
  const customDateFrom = ref('')
  const customDateTo = ref('')
  const chartMode = ref('stacked') // 'stacked' or 'line'

  const metricsData = ref([])
  const metricsScale = ref('Raw Data')
  const stats = ref(null)
  const incidents = ref([])
  const loading = ref(false)
  const error = ref(null)

  const selectedTarget = computed(() => {
    return targets.value.find((t) => t.id === selectedTargetId.value) || null
  })

  async function fetchTargets() {
    if (!authStore.token) return
    try {
      const res = await fetch('/api/targets', {
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      if (!res.ok) throw new Error('Gagal mengambil daftar target')
      const data = await res.json()
      targets.value = data

      // Auto-select first target if none selected or if previous target was deleted
      if (data.length > 0) {
        if (!selectedTargetId.value || !data.some((t) => t.id === selectedTargetId.value)) {
          selectedTargetId.value = data[0].id
        }
      } else {
        selectedTargetId.value = null
      }
    } catch (err) {
      error.value = err.message
    }
  }

  async function fetchTargetStats(targetId = selectedTargetId.value) {
    if (!targetId || !authStore.token) return
    try {
      let url = `/api/targets/${targetId}/stats?range=${selectedTimeRange.value}`
      if (selectedTimeRange.value === 'custom' && customDateFrom.value) {
        url += `&from=${encodeURIComponent(customDateFrom.value)}`
        if (customDateTo.value) {
          url += `&to=${encodeURIComponent(customDateTo.value)}`
        }
      }

      const res = await fetch(url, {
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      if (res.ok) {
        stats.value = await res.json()
      }
    } catch (err) {
      console.error('Error fetching stats:', err)
    }
  }

  async function fetchMetrics(targetId = selectedTargetId.value, silent = false) {
    if (!targetId || !authStore.token) return
    if (!silent) loading.value = true
    try {
      let url = `/api/targets/${targetId}/metrics?range=${selectedTimeRange.value}`
      if (selectedTimeRange.value === 'custom' && customDateFrom.value) {
        url += `&from=${encodeURIComponent(customDateFrom.value)}`
        if (customDateTo.value) {
          url += `&to=${encodeURIComponent(customDateTo.value)}`
        }
      }

      const res = await fetch(url, {
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      if (res.ok) {
        const result = await res.json()
        metricsData.value = result.data || []
        metricsScale.value = result.scale || 'Raw Data'
      }
    } catch (err) {
      error.value = err.message
    } finally {
      if (!silent) loading.value = false
    }
  }

  async function fetchTargetIncidents(targetId = selectedTargetId.value) {
    if (!targetId || !authStore.token) return
    try {
      const res = await fetch(`/api/targets/${targetId}/incidents`, {
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      if (res.ok) {
        incidents.value = await res.json()
      }
    } catch (err) {
      console.error('Error fetching incidents:', err)
    }
  }

  async function refreshAll(silent = false) {
    await fetchTargets()
    if (selectedTargetId.value) {
      await Promise.all([
        fetchTargetStats(selectedTargetId.value),
        fetchMetrics(selectedTargetId.value, silent),
        fetchTargetIncidents(selectedTargetId.value),
      ])
    }
  }

  async function createTarget(payload) {
    const res = await fetch('/api/targets', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify(payload),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || 'Gagal membuat target')

    await fetchTargets()
    selectedTargetId.value = data.id
    await refreshAll()
    return data
  }

  async function updateTarget(id, payload) {
    const res = await fetch(`/api/targets/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify(payload),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || 'Gagal memperbarui target')

    await fetchTargets()
    await refreshAll()
    return data
  }

  async function deleteTarget(id, purgeHistory = false) {
    const res = await fetch(`/api/targets/${id}?purge_history=${purgeHistory}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || 'Gagal menghapus target')

    await fetchTargets()
    if (targets.value.length > 0) {
      selectedTargetId.value = targets.value[0].id
      await refreshAll()
    } else {
      selectedTargetId.value = null
      stats.value = null
      metricsData.value = []
    }
    return data
  }

  async function testConnection(payload) {
    const res = await fetch('/api/targets/test-connection', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify(payload),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || 'Test koneksi gagal')
    return data
  }

  function setTimeRange(range) {
    selectedTimeRange.value = range
    if (selectedTargetId.value) {
      fetchMetrics(selectedTargetId.value)
      fetchTargetStats(selectedTargetId.value)
    }
  }

  function selectTarget(id) {
    selectedTargetId.value = id
    refreshAll()
  }

  return {
    targets,
    selectedTargetId,
    selectedTarget,
    selectedTimeRange,
    customDateFrom,
    customDateTo,
    chartMode,
    metricsData,
    metricsScale,
    stats,
    incidents,
    loading,
    error,
    fetchTargets,
    fetchTargetStats,
    fetchMetrics,
    fetchTargetIncidents,
    refreshAll,
    createTarget,
    updateTarget,
    deleteTarget,
    testConnection,
    setTimeRange,
    selectTarget,
  }
})
