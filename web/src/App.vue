<script setup lang="ts">
import { ref } from 'vue'
import { useBenchmark } from '@/composables/useBenchmark'
import { useSelection } from '@/composables/useSelection'
import CategoryNav from '@/components/CategoryNav.vue'
import FilterBar from '@/components/FilterBar.vue'
import RuleList from '@/components/RuleList.vue'
import RuleDetail from '@/components/RuleDetail.vue'
import ExportModal from '@/components/ExportModal.vue'
import type { Rule } from '@/types/benchmark'

const { meta, stats } = useBenchmark()
const { selectedCount } = useSelection()

const selectedRule = ref<Rule | null>(null)
const showExportModal = ref(false)

function onRuleSelect(rule: Rule) {
  selectedRule.value = rule
}

function onExport() {
  showExportModal.value = true
}
</script>

<template>
  <div class="app-container">
    <!-- Header -->
    <header class="app-header">
      <div class="header-left">
        
        <div>
          <h1>STIG Benchmark Builder</h1>
        </div>
      </div>
      <div class="header-right">
        <span class="header-stat">{{ stats.total }} rules</span>
        <span class="header-divider">|</span>
        <span class="header-stat">{{ stats.automatable }} automatable</span>
        <span class="header-divider">|</span>
        <span class="header-stat">{{ stats.withFixes }} with fixes</span>
        <button
          class="btn btn-primary"
          @click="onExport"
          :disabled="selectedCount === 0"
        >
          Export ({{ selectedCount }})
        </button>
      </div>
    </header>

    <!-- Version bar -->
    <div class="version-bar">
      <span>{{ meta.title }}</span>
      <span class="version-tag">{{ meta.version }}</span>
    </div>

    <!-- Main content -->
    <main class="app-main">
      <!-- Sidebar: Category navigation -->
      <aside class="sidebar">
        <CategoryNav />
      </aside>

      <!-- Content area -->
      <div class="content">
        <!-- Filter bar -->
        <FilterBar />

        <!-- Split view: Rule list and detail -->
        <div class="split-view">
          <!-- Rule list -->
          <div class="rule-list-panel">
            <RuleList @select="onRuleSelect" />
          </div>

          <!-- Detail panel -->
          <aside class="detail-panel">
            <RuleDetail :rule="selectedRule" />
          </aside>
        </div>
      </div>
    </main>

    <!-- Export modal -->
    <ExportModal
      v-if="showExportModal"
      @close="showExportModal = false"
    />
  </div>
</template>

<style scoped>
.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logo {
  width: 24px;
  height: 24px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-stat {
  font-size: 12px;
  color: var(--color-text-muted);
}

.header-divider {
  color: var(--color-border);
  font-size: 12px;
}

.version-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  background: var(--color-background);
  border-bottom: 1px solid var(--color-border);
  font-size: 11px;
  color: var(--color-text-secondary);
}

.version-tag {
  padding: 1px 6px;
  background: var(--color-primary);
  color: white;
  border-radius: 3px;
  font-weight: 600;
  font-size: 10px;
}

.split-view {
  display: flex;
  flex: 1;
  overflow: hidden;
}
</style>
