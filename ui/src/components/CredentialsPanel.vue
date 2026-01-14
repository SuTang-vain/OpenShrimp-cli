<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Credential {
  model: string
  key: string
  source: string
  provider: string
  set: boolean
  env_var?: string
}

const credentials = ref<{ credentials: Credential[]; count: number } | null>(null)
const loading = ref(true)
const showAddModal = ref(false)
const selectedModel = ref('')
const selectedKey = ref('')
const credentialValue = ref('')
const envVarName = ref('')
const providerName = ref('')
const saving = ref(false)

const commonKeys = [
  { name: 'ANTHROPIC_API_KEY', provider: 'anthropic' },
  { name: 'MINIMAX_API_KEY', provider: 'minimax' },
  { name: 'ZHIPU_API_KEY', provider: 'zhipu' },
  { name: 'GOOGLE_API_KEY', provider: 'google' },
  { name: 'OPENAI_API_KEY', provider: 'openai' }
]

const fetchCredentials = async () => {
  try {
    const res = await fetch('/api/credentials')
    if (res.ok) {
      credentials.value = await res.json()
    }
  } catch (e) {
    console.error('Failed to fetch credentials:', e)
  } finally {
    loading.value = false
  }
}

const addCredential = async () => {
  if (!selectedModel.value || !selectedKey.value) return

  saving.value = true
  try {
    const res = await fetch('/api/credentials', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        model: selectedModel.value,
        key: selectedKey.value,
        value: credentialValue.value || undefined,
        env_var: envVarName.value || undefined,
        provider: providerName.value
      })
    })

    if (res.ok) {
      showAddModal.value = false
      resetForm()
      await fetchCredentials()
    }
  } catch (e) {
    console.error('Failed to save credential:', e)
  } finally {
    saving.value = false
  }
}

const deleteCredential = async (model: string, key: string) => {
  if (!confirm(`Delete credential ${model}:${key}?`)) return

  try {
    const res = await fetch(`/api/credentials/${model}/${key}`, {
      method: 'DELETE'
    })

    if (res.ok) {
      await fetchCredentials()
    }
  } catch (e) {
    console.error('Failed to delete credential:', e)
  }
}

const resetForm = () => {
  selectedModel.value = ''
  selectedKey.value = ''
  credentialValue.value = ''
  envVarName.value = ''
  providerName.value = ''
}

const selectKey = (key: string, provider: string) => {
  selectedKey.value = key
  providerName.value = provider
}

onMounted(() => {
  fetchCredentials()
})
</script>

<template>
  <div class="card">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold">API Credentials</h2>
      <button @click="showAddModal = true" class="btn btn-primary btn-sm">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
        </svg>
      </button>
    </div>

    <div v-if="loading" class="space-y-3 animate-pulse">
      <div v-for="i in 2" :key="i" class="h-16 bg-slate-700 rounded-lg"></div>
    </div>

    <div v-else-if="credentials && credentials.count > 0" class="space-y-3">
      <div
        v-for="cred in credentials.credentials"
        :key="cred.model + cred.key"
        class="bg-slate-900 rounded-lg p-4"
      >
        <div class="flex items-center justify-between">
          <div>
            <div class="flex items-center gap-2">
              <span class="font-medium">{{ cred.model }}</span>
              <span class="text-xs bg-slate-700 px-2 py-0.5 rounded text-slate-300">
                {{ cred.key }}
              </span>
            </div>
            <div class="text-sm text-slate-400 mt-1">
              <span :class="cred.set ? 'text-green-400' : 'text-yellow-400'">
                {{ cred.set ? 'Configured' : 'Not set' }}
              </span>
              <span v-if="cred.source" class="ml-2">({{ cred.source }})</span>
              <span v-if="cred.env_var" class="ml-2">Env: {{ cred.env_var }}</span>
            </div>
          </div>

          <button
            @click="deleteCredential(cred.model, cred.key)"
            class="text-red-400 hover:text-red-300 p-1"
            title="Delete credential"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <div v-else class="text-center py-8 text-slate-400">
      <svg class="w-12 h-12 mx-auto mb-3 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
      </svg>
      <p class="text-sm">No credentials configured</p>
      <p class="text-xs mt-1">Click + to add API credentials</p>
    </div>
  </div>

  <!-- Add Credential Modal -->
  <div v-if="showAddModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
    <div class="bg-slate-800 rounded-lg p-6 w-full max-w-md mx-4">
      <h3 class="text-lg font-semibold mb-4">Add API Credential</h3>

      <div class="space-y-4">
        <div>
          <label class="block text-sm text-slate-400 mb-1">Model Name</label>
          <input
            v-model="selectedModel"
            type="text"
            placeholder="e.g., claude-sonnet-4"
            class="input w-full"
          />
        </div>

        <div>
          <label class="block text-sm text-slate-400 mb-1">API Key Name</label>
          <input
            v-model="selectedKey"
            type="text"
            placeholder="e.g., ANTHROPIC_API_KEY"
            class="input w-full"
          />
        </div>

        <div>
          <label class="block text-sm text-slate-400 mb-2">Common Keys</label>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="key in commonKeys"
              :key="key.name"
              @click="selectKey(key.name, key.provider)"
              class="text-xs px-2 py-1 bg-slate-700 hover:bg-slate-600 rounded transition-colors"
              :class="{ 'ring-2 ring-green-500': selectedKey === key.name }"
            >
              {{ key.name }}
            </button>
          </div>
        </div>

        <div>
          <label class="block text-sm text-slate-400 mb-1">Provider (optional)</label>
          <input
            v-model="providerName"
            type="text"
            placeholder="e.g., anthropic"
            class="input w-full"
          />
        </div>

        <div class="border-t border-slate-700 pt-4">
          <p class="text-sm text-slate-400 mb-2">Store credential as:</p>

          <div class="space-y-2">
            <label class="flex items-center gap-2">
              <input type="radio" v-model="envVarName" value="" class="radio" />
              <span class="text-sm">Direct value (stored in keychain)</span>
            </label>
            <input
              v-model="credentialValue"
              type="password"
              placeholder="Enter API key value..."
              class="input w-full ml-6"
              :disabled="envVarName !== ''"
            />

            <label class="flex items-center gap-2">
              <input type="radio" v-model="envVarName" value="custom" class="radio" />
              <span class="text-sm">Environment variable</span>
            </label>
            <input
              v-model="envVarName"
              type="text"
              placeholder="Enter env var name (e.g., ANTHROPIC_API_KEY)"
              class="input w-full ml-6"
              :disabled="envVarName === ''"
            />
          </div>
        </div>
      </div>

      <div class="flex gap-3 mt-6">
        <button @click="showAddModal = false" class="btn btn-secondary flex-1">Cancel</button>
        <button
          @click="addCredential"
          :disabled="!selectedModel || !selectedKey || saving"
          class="btn btn-primary flex-1"
        >
          {{ saving ? 'Saving...' : 'Save' }}
        </button>
      </div>
    </div>
  </div>
</template>
