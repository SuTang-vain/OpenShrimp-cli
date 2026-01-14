<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface LinkInfo {
  tool_name: string
  link_path: string
  target_path: string
  exists: boolean
  valid: boolean
  is_symlink: boolean
  error?: string
}

const links = ref<{ links: LinkInfo[] } | null>(null)
const loading = ref(true)
const creating = ref<string | null>(null)
const removing = ref<string | null>(null)

const fetchLinks = async () => {
  try {
    const res = await fetch('/api/links')
    if (res.ok) {
      links.value = await res.json()
    }
  } catch (e) {
    console.error('Failed to fetch links:', e)
  } finally {
    loading.value = false
  }
}

const createLink = async (toolName: string) => {
  creating.value = toolName
  try {
    const res = await fetch('/api/links', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ tool: toolName })
    })
    if (res.ok) {
      await fetchLinks()
    }
  } catch (e) {
    console.error('Failed to create link:', e)
  } finally {
    creating.value = null
  }
}

const removeLink = async (toolName: string) => {
  removing.value = toolName
  try {
    const res = await fetch(`/api/links/${toolName}`, { method: 'DELETE' })
    if (res.ok) {
      await fetchLinks()
    }
  } catch (e) {
    console.error('Failed to remove link:', e)
  } finally {
    removing.value = null
  }
}

const getStatusClass = (link: LinkInfo) => {
  if (!link.exists) return 'status-warning'
  if (!link.valid) return 'status-error'
  return 'status-healthy'
}

const getStatusText = (link: LinkInfo) => {
  if (!link.exists) return 'Missing'
  if (!link.valid) return link.error || 'Invalid'
  return 'Valid'
}

onMounted(() => {
  fetchLinks()
})
</script>

<template>
  <div class="card">
    <h2 class="text-lg font-semibold mb-4">Symlinks</h2>

    <div v-if="loading" class="space-y-3 animate-pulse">
      <div v-for="i in 3" :key="i" class="h-16 bg-slate-700 rounded-lg"></div>
    </div>

    <div v-else-if="links" class="space-y-3">
      <div
        v-for="link in links.links"
        :key="link.tool_name"
        class="bg-slate-900 rounded-lg p-4"
      >
        <div class="flex items-center justify-between mb-2">
          <span class="font-medium">{{ link.tool_name }}</span>
          <span :class="['status-badge', getStatusClass(link)]">
            {{ getStatusText(link) }}
          </span>
        </div>

        <div class="text-xs text-slate-500 mb-3">
          <div>Link: {{ link.link_path }}</div>
          <div>Target: {{ link.target_path }}</div>
        </div>

        <div class="flex gap-2">
          <button
            v-if="!link.exists || !link.valid"
            @click="createLink(link.tool_name)"
            :disabled="creating === link.tool_name"
            class="btn btn-primary text-sm flex-1"
          >
            {{ creating === link.tool_name ? '...' : 'Create' }}
          </button>
          <button
            v-else
            @click="removeLink(link.tool_name)"
            :disabled="removing === link.tool_name"
            class="btn btn-secondary text-sm flex-1"
          >
            {{ removing === link.tool_name ? '...' : 'Remove' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
