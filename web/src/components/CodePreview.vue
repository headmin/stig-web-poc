<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  code: string
  language: string
  filename?: string
}>()

const copied = ref(false)

function copyCode() {
  navigator.clipboard.writeText(props.code)
  copied.value = true
  setTimeout(() => {
    copied.value = false
  }, 2000)
}
</script>

<template>
  <div class="code-block">
    <div class="code-header">
      <span>{{ filename || language }}</span>
      <button class="copy-btn" @click="copyCode">
        {{ copied ? 'Copied!' : 'Copy' }}
      </button>
    </div>
    <div class="code-content">
      <pre>{{ code }}</pre>
    </div>
  </div>
</template>

<style scoped>
.copy-btn {
  padding: 2px 8px;
  font-size: 11px;
  background: #444;
  border: none;
  border-radius: var(--radius-sm);
  color: #ccc;
  cursor: pointer;
  transition: all 0.2s;
}

.copy-btn:hover {
  background: #555;
  color: white;
}
</style>
