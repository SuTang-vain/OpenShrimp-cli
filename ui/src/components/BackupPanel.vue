<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Backup {
  name: string
  path: string
  size: number
  files: number
  timestamp: string
  tool_count: number
}

const backups = ref<Backup[]>([])
// Use a ref for the backup name input
const newBackupName = ref('')
const loading = ref(true)
const creating = ref(false)
const restoring = ref<string | null>(null)
const deleting = ref<string | null>(null)

const fetchBackups = async () => {
  try {
    const res = await fetch('/api/backups')
    if (res.ok) {
      const data = await res.json()
      backups.value = data.backups || []
    }
  } catch (e) {
    console.error('Failed to fetch backups:', e)
  } finally {
    loading.value = false
  }
}

const createBackup = async () => {
  creating.value = true
  try {
    const res = await fetch('/api/backups', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: newBackupName.value || undefined })
    })
    if (res.ok) {
      newBackupName.value = ''
      await fetchBackups()
    }
  } catch (e) {
    console.error('Failed to create backup:', e)
  } finally {
    creating.value = false
  }
}

const restoreBackup = async (name: string) => {
  if (!confirm(`Restore backup ${name}?`)) return
  restoring.value = name
  try {
    const res = await fetch(`/api/backups/${name}/restore`, { method: 'POST' })
    if (res.ok) {
      await fetchBackups()
    }
  } catch (e) {
    console.error('Failed to restore backup:', e)
  } finally {
    restoring.value = null
  }
}

const deleteBackup = async (name: string) => {
  if (!confirm(`Delete backup ${name}?`)) return
  deleting.value = name
  try {
    const res = await fetch(`/api/backups/${name}`, { method: 'DELETE' })
    if (res.ok) {
      await fetchBackups()
    }
  } catch (e) {
    console.error('Failed to delete backup:', e)
  } finally {
    deleting.value = null
  }
}

const formatSize = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleString()
}

onMounted(() => {
  fetchBackups()
})
</script>

<template>
  <div class="card">
    <h2 class="text-lg font-semibold mb-4">Backups</h2>

    <!-- Create Backup -->
    <div class="flex gap-2 mb-4">
      <input
        v-model="newBackupName"
        type="text"
        placeholder="Backup name (optional)"
        class="input flex-1"
        @keyup.enter="createBackup"
      />
      <button
        @click="createBackup"
        :disabled="creating"
        class="btn btn-primary"
      >
        {{ creating ? '...' : '+' }}
      </button>
    </div>

    <div v-if="loading" class="space-y-3 animate-pulse">
      <div v-for="i in 2" :key="i" class="h-16 bg-slate-700 rounded-lg"></div>
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="backup in backups"
        :key="backup.name"
        class="bg-slate-900 rounded-lg p-4"
      >
        <div class="flex items-center justify-between mb-2">
          <span class="font-medium">{{ backup.name }}</span>
          <span class="text-sm text-slate-400">{{ formatSize(backup.size) }}</span>
        </div>

        <div class="text-sm text-slate-400 mb-3">
          {{ backup.files }} files · {{ formatDate(backup.timestamp) }}
        </div>

        <div class="flex gap-2">
          <button
            @click="restoreBackup(backup.name)"
            :disabled="restoring === backup.name"
            class="btn btn-secondary text-sm flex-1"
          >
            {{ restoring === backup.name ? '...' : 'Restore' }}
          </button>
          <button
            @click="deleteBackup(backup.name)"
            :disabled="deleting === backup.name"
            class="btn btn-danger text-sm"
          >
            ✕
          </button>
        </div>
      </div>

      <div v-if="backups.length === 0" class="text-center text-slate-400 py-4">
        No backups yet
      </div>
    </div>
  </div>
</template>
