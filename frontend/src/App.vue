<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from './stores/auth'

const authStore = useAuthStore()

onMounted(() => {
  authStore.fetchCurrentUser()
})
</script>

<template>
  <router-view />
</template>
