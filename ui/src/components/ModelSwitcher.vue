<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Model {
  name: string
  full_name: string
  provider: string
  endpoint: string
  model_id: string
  default: boolean
}

const models = ref<{ models: Model[]; current: string } | null>(null)
const loading = ref(true)
const switching = ref(false)

const fetchModels = async () => {
  try {
    const res = await fetch('/api/models')
    if (res.ok) {
      models.value = await res.json()
    }
  } catch (e) {
    console.error('Failed to fetch models:', e)
  } finally {
    loading.value = false
  }
}

const switchModel = async (name: string) => {
  switching.value = true
  try {
    const res = await fetch('/api/switch', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ model: name })
    })
    if (res.ok) {
      await fetchModels()
    }
  } catch (e) {
    console.error('Failed to switch model:', e)
  } finally {
    switching.value = false
  }
}

const getProviderColor = (provider: string) => {
  const colors: Record<string, string> = {
    anthropic: 'text-orange-400',
    minimax: 'text-blue-400',
    zhipu: 'text-purple-400'
  }
  return colors[provider] || 'text-slate-400'
}

onMounted(() => {
  fetchModels()
})
</script>

<template>
  <div class="card">
    <h2 class="text-lg font-semibold mb-4">Model Switcher</h2>

    <div v-if="loading" class="space-y-3 animate-pulse">
      <div v-for="i in 3" :key="i" class="h-16 bg-slate-700 rounded-lg"></div>
    </div>

    <div v-else-if="models" class="space-y-3">
      <div
        v-for="model in models.models"
        :key="model.name"
        class="bg-slate-900 rounded-lg p-4 cursor-pointer transition-all hover:bg-slate-700"
        :class="{ 'ring-2 ring-green-500': models.current === model.name }"
        @click="switchModel(model.name)"
      >
        <div class="flex items-center justify-between">
          <div>
            <div class="flex items-center gap-2">
              <span class="font-medium">{{ model.full_name }}</span>
              <span v-if="model.default" class="status-badge status-healthy">Default</span>
            </div>
            <div class="text-sm text-slate-400 mt-1">
              <span :class="getProviderColor(model.provider)">{{ model.provider }}</span>
            </div>
          </div>

          <div v-if="models.current === model.name" class="text-green-400">
            <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
            </svg>
          </div>
        </div>

        <div v-if="switching" class="mt-2 text-sm text-green-400">
          Switching...
        </div>
      </div>
    </div>
  </div>
</template>
