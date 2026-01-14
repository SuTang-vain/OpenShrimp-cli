<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

interface Stats {
  disk_usage: {
    total_bytes: number
    formatted: string
  }
  backups_count: number
  tools_count: number
}

const stats = ref<Stats | null>(null)
const loading = ref(true)
let ws: WebSocket | null = null

const fetchStats = async () => {
  try {
    const res = await fetch('/api/stats')
    if (res.ok) {
      stats.value = await res.json()
    }
  } catch (e) {
    console.error('Failed to fetch stats:', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchStats()

  // Connect to WebSocket for real-time updates
  ws = new WebSocket(`ws://${location.host}/ws`)
  ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    if (data.type === 'cleanup_complete' || data.type === 'backup_created') {
      fetchStats()
    }
  }
})

onUnmounted(() => {
  ws?.close()
})
</script>

<template>
  <div class="card">
    <h2 class="text-lg font-semibold mb-4">Overview</h2>

    <div v-if="loading" class="animate-pulse">
      <div class="h-20 bg-slate-700 rounded-lg"></div>
    </div>

    <div v-else class="grid grid-cols-3 gap-4">
      <div class="bg-slate-900 rounded-lg p-4">
        <div class="text-slate-400 text-sm">Disk Usage</div>
        <div class="text-2xl font-bold text-green-400 mt-1">
          {{ stats?.disk_usage?.formatted || 'N/A' }}
        </div>
      </div>

      <div class="bg-slate-900 rounded-lg p-4">
        <div class="text-slate-400 text-sm">Tools</div>
        <div class="text-2xl font-bold mt-1">{{ stats?.tools_count || 0 }}</div>
      </div>

      <div class="bg-slate-900 rounded-lg p-4">
        <div class="text-slate-400 text-sm">Backups</div>
        <div class="text-2xl font-bold mt-1">{{ stats?.backups_count || 0 }}</div>
      </div>
    </div>
  </div>
</template>
