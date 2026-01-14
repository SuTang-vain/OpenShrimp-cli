<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface TaskResult {
  success: boolean
  message: string
  start_time: string
  end_time: string
  duration: number
}

interface Task {
  id: string
  type: string
  schedule: string
  enabled: boolean
  tool: string
  last_run?: string
  last_result?: TaskResult
  created_at: string
}

const tasks = ref<Task[]>([])
const loading = ref(true)
const showForm = ref(false)

const newTask = ref({
  id: '',
  type: 'cleanup',
  schedule: '0 0 * * *',
  tool: '',
  enabled: true
})

const fetchTasks = async () => {
  try {
    const res = await fetch('/api/scheduler')
    if (res.ok) {
      const data = await res.json()
      tasks.value = data.tasks || []
    }
  } catch (e) {
    console.error('Failed to fetch tasks:', e)
  } finally {
    loading.value = false
  }
}

const createTask = async () => {
  try {
    const res = await fetch('/api/scheduler', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(newTask.value)
    })
    if (res.ok) {
      showForm.value = false
      newTask.value = { id: '', type: 'cleanup', schedule: '0 0 * * *', tool: '', enabled: true }
      await fetchTasks()
    }
  } catch (e) {
    console.error('Failed to create task:', e)
  }
}

const toggleTask = async (task: Task) => {
  try {
    await fetch(`/api/scheduler/${task.id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled: !task.enabled, schedule: task.schedule })
    })
    await fetchTasks()
  } catch (e) {
    console.error('Failed to update task:', e)
  }
}

const deleteTask = async (id: string) => {
  if (!confirm(`Delete task ${id}?`)) return
  try {
    const res = await fetch(`/api/scheduler/${id}`, { method: 'DELETE' })
    if (res.ok) {
      await fetchTasks()
    }
  } catch (e) {
    console.error('Failed to delete task:', e)
  }
}

const runTask = async (id: string) => {
  try {
    await fetch(`/api/scheduler/${id}/run`, { method: 'POST' })
  } catch (e) {
    console.error('Failed to run task:', e)
  }
}

const getTypeColor = (type: string) => {
  return type === 'cleanup' ? 'text-blue-400' : 'text-green-400'
}

const formatDuration = (ms: number) => {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(1)}s`
}

const cronDescriptions: Record<string, string> = {
  '0 0 * * *': 'Every day at midnight',
  '0 0 * * 0': 'Every Sunday at midnight',
  '0 0 1 * *': 'First day of each month',
  '0 0 * * 1': 'Every Monday at midnight',
  '*/30 * * * *': 'Every 30 minutes',
  '0 */6 * * *': 'Every 6 hours'
}

onMounted(() => {
  fetchTasks()
})
</script>

<template>
  <div class="card">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold">Scheduled Tasks</h2>
      <button @click="showForm = !showForm" class="btn btn-primary text-sm">
        {{ showForm ? 'Cancel' : '+ Add Task' }}
      </button>
    </div>

    <!-- Add Task Form -->
    <div v-if="showForm" class="bg-slate-900 rounded-lg p-4 mb-4">
      <div class="grid grid-cols-2 gap-3 mb-3">
        <div>
          <label class="block text-sm text-slate-400 mb-1">Task ID</label>
          <input v-model="newTask.id" type="text" placeholder="daily-cleanup" class="input w-full" />
        </div>
        <div>
          <label class="block text-sm text-slate-400 mb-1">Type</label>
          <select v-model="newTask.type" class="input w-full">
            <option value="cleanup">Cleanup</option>
            <option value="backup">Backup</option>
          </select>
        </div>
      </div>

      <div class="mb-3">
        <label class="block text-sm text-slate-400 mb-1">Schedule (cron)</label>
        <input v-model="newTask.schedule" type="text" placeholder="0 0 * * *" class="input w-full" />
        <p class="text-xs text-slate-500 mt-1">{{ cronDescriptions[newTask.schedule] || 'Custom schedule' }}</p>
      </div>

      <div v-if="newTask.type === 'cleanup'" class="mb-3">
        <label class="block text-sm text-slate-400 mb-1">Tool (optional)</label>
        <input v-model="newTask.tool" type="text" placeholder="claude (all tools if empty)" class="input w-full" />
      </div>

      <div class="flex gap-2">
        <label class="flex items-center gap-2">
          <input v-model="newTask.enabled" type="checkbox" class="w-4 h-4" />
          <span class="text-sm">Enabled</span>
        </label>
      </div>

      <button @click="createTask" class="btn btn-primary mt-3 w-full">Create Task</button>
    </div>

    <!-- Task List -->
    <div v-if="loading" class="space-y-3 animate-pulse">
      <div v-for="i in 2" :key="i" class="h-20 bg-slate-700 rounded-lg"></div>
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="task in tasks"
        :key="task.id"
        class="bg-slate-900 rounded-lg p-4"
        :class="{ 'opacity-50': !task.enabled }"
      >
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center gap-3">
            <span class="font-medium">{{ task.id }}</span>
            <span :class="['text-xs', getTypeColor(task.type)]">{{ task.type }}</span>
          </div>
          <div class="flex items-center gap-2">
            <button
              @click="runTask(task.id)"
              class="text-xs text-green-400 hover:text-green-300"
            >
              Run
            </button>
            <button
              @click="deleteTask(task.id)"
              class="text-xs text-red-400 hover:text-red-300"
            >
              Delete
            </button>
          </div>
        </div>

        <div class="text-sm text-slate-400 mb-2">
          <span class="font-mono bg-slate-800 px-2 py-0.5 rounded">{{ task.schedule }}</span>
        </div>

        <div class="flex items-center gap-4 text-xs text-slate-500">
          <label class="flex items-center gap-1">
            <input
              type="checkbox"
              :checked="task.enabled"
              @change="toggleTask(task)"
              class="w-3 h-3"
            />
            {{ task.enabled ? 'Enabled' : 'Disabled' }}
          </label>
          <span v-if="task.last_run">
            Last run: {{ new Date(task.last_run).toLocaleString() }}
          </span>
        </div>

        <div v-if="task.last_result" class="mt-2 text-xs">
          <span :class="task.last_result.success ? 'text-green-400' : 'text-red-400'">
            {{ task.last_result.success ? '✓' : '✗' }} {{ task.last_result.message }}
          </span>
          <span class="text-slate-500 ml-2">
            ({{ formatDuration(task.last_result.duration) }})
          </span>
        </div>
      </div>

      <div v-if="tasks.length === 0" class="text-center text-slate-400 py-4">
        No scheduled tasks
      </div>
    </div>
  </div>
</template>
