<script setup lang="ts">
import { computed } from 'vue'

import { useBenchmark } from '@/composables/useBenchmark'
import { useSelection } from '@/composables/useSelection'
import type { Rule } from '@/types/benchmark'

const emit = defineEmits<{
  select: [rule: Rule]
}>()

const { filteredRules, selectedCategory, allRules, selectedCategoryId } = useBenchmark()
const { isSelected, toggleSelection } = useSelection()

// Use all rules when no category is selected
const displayRules = computed(() => {
  if (selectedCategoryId.value === null) {
    return allRules.value
  }
  return filteredRules.value
})

function onRuleClick(rule: Rule) {
  emit('select', rule)
}

function onCheckboxChange(event: Event, rule: Rule) {
  event.stopPropagation()
  toggleSelection(rule)
}

function getSeverityClass(severity: string): string {
  return `badge badge-${severity}`
}

</script>

<template>
  <div class="rule-list">
    <!-- Header -->
    <div class="list-header">
      <span v-if="selectedCategory">
        {{ selectedCategory.name }} ({{ displayRules.length }} rules)
      </span>
      <span v-else>
        All Rules ({{ displayRules.length }})
      </span>
    </div>

    <!-- Empty state -->
    <div v-if="displayRules.length === 0" class="empty-state">
      <h3>No rules found</h3>
      <p>Try adjusting your filters or search query</p>
    </div>

    <!-- Rule items -->
    <div
      v-for="(rule, index) in displayRules"
      :key="`${rule.ruleId || rule.id || index}-${index}`"
      :class="['rule-item', isSelected(rule) ? 'selected' : '']"
      @click="onRuleClick(rule)"
    >
      <input
        type="checkbox"
        :checked="isSelected(rule)"
        @click.stop
        @change="onCheckboxChange($event, rule)"
      />

      <div class="rule-content">
        <div class="rule-header">
          <span class="rule-id">{{ rule.ruleId || rule.id }}</span>
          <span :class="getSeverityClass(rule.severity)">{{ rule.severity }}</span>
          <span v-if="rule.automatable" class="badge badge-auto">Auto</span>
          <span v-else class="badge badge-manual">Manual</span>
          <span v-if="rule.fix" class="badge badge-fix">Fix</span>
        </div>
        <div class="rule-title">{{ rule.title }}</div>
        <div class="rule-meta">
          <span v-if="rule.tags && rule.tags.length > 0">
            {{ rule.tags.slice(0, 3).join(', ') }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.list-header {
  padding: var(--spacing-md);
  font-weight: 600;
  color: var(--color-text-secondary);
  border-bottom: 1px solid var(--color-border);
  background: var(--color-background);
  position: sticky;
  top: 0;
  z-index: 10;
}
</style>
