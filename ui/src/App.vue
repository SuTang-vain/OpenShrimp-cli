<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import Dashboard from './components/Dashboard.vue'
import ToolList from './components/ToolList.vue'
import ModelSwitcher from './components/ModelSwitcher.vue'
import BackupPanel from './components/BackupPanel.vue'
import LinkManager from './components/LinkManager.vue'
import SchedulerPanel from './components/SchedulerPanel.vue'
import CredentialsPanel from './components/CredentialsPanel.vue'

const connected = ref(false)
const reconnecting = ref(false)

const checkConnection = async () => {
  try {
    const res = await fetch('/api/health')
    connected.value = res.ok
  } catch {
    connected.value = false
  }
}

const handleReconnect = async () => {
  reconnecting.value = true
  await checkConnection()
  setTimeout(() => {
    reconnecting.value = false
  }, 2000)
}

let interval: number

onMounted(() => {
  checkConnection()
  interval = setInterval(checkConnection, 5000) as unknown as number
})

onUnmounted(() => {
  clearInterval(interval)
})
</script>

<template>
  <div class="min-h-screen bg-slate-900 text-white">
    <!-- Header -->
    <header class="border-b border-slate-700 bg-slate-800/50 backdrop-blur-sm sticky top-0 z-10">
      <div class="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 bg-green-500 rounded-lg flex items-center justify-center">
            <span class="text-xl font-bold">S</span>
          </div>
          <div>
            <h1 class="text-xl font-bold">OpenShrimp</h1>
            <p class="text-xs text-slate-400">AI Tools Manager</p>
          </div>
        </div>
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2">
            <div
              class="w-2 h-2 rounded-full"
              :class="connected ? 'bg-green-500' : 'bg-red-500'"
            ></div>
            <span class="text-sm text-slate-400">
              {{ connected ? 'Connected' : 'Disconnected' }}
            </span>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-6 py-8">
      <!-- Reconnect Banner -->
      <div
        v-if="!connected && !reconnecting"
        class="bg-yellow-500/10 border border-yellow-500/30 rounded-lg p-4 mb-6 flex items-center justify-between"
      >
        <span class="text-yellow-400">Daemon not running. Start with: <code class="bg-slate-800 px-2 py-1 rounded">ai-mgr daemon</code></span>
        <button @click="handleReconnect" class="btn btn-secondary">
          Reconnect
        </button>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Left Column -->
        <div class="lg:col-span-2 space-y-6">
          <Dashboard />
          <ToolList />
        </div>

        <!-- Right Column -->
        <div class="space-y-6">
          <ModelSwitcher />
          <CredentialsPanel />
          <SchedulerPanel />
          <BackupPanel />
          <LinkManager />
        </div>
      </div>
    </main>
  </div>
</template>
