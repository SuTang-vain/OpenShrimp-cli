<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Tool {
  name: string
  path: string
  found: boolean
  enabled: boolean
  disk_usage?: string
  status: string
}

interface ScanResult {
  tools: Tool[]
  total: number
  enabled: number
}

const tools = ref<ScanResult | null>(null)
const loading = ref(true)
const cleaning = ref<string | null>(null)

const fetchTools = async () => {
  try {
    const res = await fetch('/api/tools')
    if (res.ok) {
      tools.value = await res.json()
    }
  } catch (e) {
    console.error('Failed to fetch tools:', e)
  } finally {
    loading.value = false
  }
}

const cleanupTool = async (name: string) => {
  cleaning.value = name
  try {
    const res = await fetch(`/api/tools/${name}/cleanup`, { method: 'POST' })
    if (res.ok) {
      await fetchTools()
    }
  } catch (e) {
    console.error('Failed to cleanup:', e)
  } finally {
    cleaning.value = null
  }
}

const getStatusClass = (found: boolean) => {
  if (!found) return 'status-error'
  return 'status-healthy'
}

const getStatusText = (found: boolean) => {
  if (!found) return 'Not Found'
  return 'Active'
}

onMounted(() => {
  fetchTools()
})
</script>

<template>
  <div class="card">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold">AI Tools</h2>
      <button @click="fetchTools" class="text-sm text-green-400 hover:text-green-300">
        Refresh
      </button>
    </div>

    <div v-if="loading" class="space-y-3 animate-pulse">
      <div v-for="i in 3" :key="i" class="h-16 bg-slate-700 rounded-lg"></div>
    </div>

    <div v-else-if="tools" class="space-y-3">
      <div
        v-for="tool in tools.tools"
        :key="tool.name"
        class="flex items-center justify-between bg-slate-900 rounded-lg p-4"
      >
        <div class="flex items-center gap-4">
          <div
            class="w-10 h-10 rounded-lg flex items-center justify-center"
            :class="tool.found ? 'bg-green-500/20' : 'bg-red-500/20'"
          >
            <span class="text-xl">{{ tool.found ? '✓' : '✗' }}</span>
          </div>
          <div>
            <div class="font-medium">{{ tool.name }}</div>
            <div class="text-sm text-slate-400">{{ tool.path }}</div>
          </div>
        </div>

        <div class="flex items-center gap-4">
          <span :class="['status-badge', getStatusClass(tool.found)]">
            {{ getStatusText(tool.found) }}
          </span>

          <button
            v-if="tool.found"
            @click="cleanupTool(tool.name)"
            :disabled="cleaning === tool.name"
            class="btn btn-secondary text-sm"
          >
            {{ cleaning === tool.name ? 'Cleaning...' : 'Cleanup' }}
          </button>
        </div>
      </div>

      <div v-if="tools.tools.length === 0" class="text-center text-slate-400 py-8">
        No tools found
      </div>
    </div>
  </div>
</template>
