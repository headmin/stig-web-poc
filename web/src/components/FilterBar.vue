<script setup lang="ts">
import { useBenchmark } from '@/composables/useBenchmark'
import { useSelection } from '@/composables/useSelection'

const {
  searchQuery,
  setSearchQuery,
  severityFilter,
  toggleSeverity,
  automatableFilter,
  setAutomatableFilter,
  hasFixFilter,
  toggleHasFixFilter,
  filteredRules,
  allRules,
  selectedCategoryId
} = useBenchmark()

const { selectAll, deselectAll, areAllSelected } = useSelection()

function onSearchInput(event: Event) {
  const target = event.target as HTMLInputElement
  setSearchQuery(target.value)
}

// Get the currently displayed rules
function getDisplayedRules() {
  return selectedCategoryId.value === null ? allRules.value : filteredRules.value
}

function onSelectAll() {
  const rules = getDisplayedRules()
  if (areAllSelected(rules)) {
    deselectAll(rules)
  } else {
    selectAll(rules)
  }
}
</script>

<template>
  <div class="filter-bar">
    <!-- Search -->
    <input
      type="text"
      placeholder="Search rules..."
      :value="searchQuery"
      @input="onSearchInput"
    />

    <!-- Severity filters -->
    <div class="filter-group">
      <button
        :class="['severity-btn', 'high', severityFilter.includes('high') ? 'active' : '']"
        @click="toggleSeverity('high')"
        title="High severity"
      >
        High
      </button>
      <button
        :class="['severity-btn', 'medium', severityFilter.includes('medium') ? 'active' : '']"
        @click="toggleSeverity('medium')"
        title="Medium severity"
      >
        Med
      </button>
      <button
        :class="['severity-btn', 'low', severityFilter.includes('low') ? 'active' : '']"
        @click="toggleSeverity('low')"
        title="Low severity"
      >
        Low
      </button>
    </div>

    <div class="divider"></div>

    <!-- Automatable filter -->
    <div class="filter-group">
      <select
        class="filter-select"
        :value="automatableFilter"
        @change="setAutomatableFilter(($event.target as HTMLSelectElement).value as any)"
      >
        <option value="all">All Types</option>
        <option value="automatable">Automatable</option>
        <option value="manual">Manual</option>
      </select>
    </div>

    <!-- Has Fix filter -->
    <button
      :class="['btn', 'btn-sm', hasFixFilter ? 'btn-primary' : 'btn-secondary']"
      @click="toggleHasFixFilter"
    >
      Has Fix
    </button>

    <div class="flex-1"></div>

    <!-- Select all -->
    <button class="btn btn-sm btn-secondary" @click="onSelectAll">
      {{ areAllSelected(getDisplayedRules()) ? 'Deselect All' : 'Select All' }}
    </button>
  </div>
</template>

<style scoped>
.divider {
  width: 1px;
  height: 16px;
  background: var(--color-border);
}

.flex-1 {
  flex: 1;
}
</style>
