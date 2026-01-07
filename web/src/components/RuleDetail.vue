<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Rule } from '@/types/benchmark'
import CodePreview from '@/components/CodePreview.vue'

const props = defineProps<{
  rule: Rule | null
}>()

const activeTab = ref<'description' | 'check' | 'fix' | 'query'>('description')

// Reset to description tab when rule changes
watch(() => props.rule, () => {
  activeTab.value = 'description'
})

function getSeverityClass(severity: string): string {
  return `badge badge-${severity}`
}
</script>

<template>
  <div class="rule-detail">
    <!-- Empty state -->
    <div v-if="!rule" class="empty-state" style="height: 100%;">
      <h3>Select a rule</h3>
      <p>Click on a rule to view its details</p>
    </div>

    <!-- Rule details -->
    <template v-else>
      <!-- Header -->
      <div class="detail-header">
        <div class="rule-id">{{ rule.ruleId || rule.id }}</div>
        <h2>{{ rule.title }}</h2>
        <div class="meta">
          <span :class="getSeverityClass(rule.severity)">{{ rule.severity }}</span>
          <span v-if="rule.automatable" class="badge badge-auto">Auto</span>
          <span v-else class="badge badge-manual">Manual</span>
          <span v-if="rule.fix" class="badge badge-fix">Fix</span>
        </div>
      </div>

      <!-- Tabs -->
      <div class="detail-tabs">
        <button
          :class="['detail-tab', activeTab === 'description' ? 'active' : '']"
          @click="activeTab = 'description'"
        >
          Description
        </button>
        <button
          :class="['detail-tab', activeTab === 'check' ? 'active' : '']"
          @click="activeTab = 'check'"
        >
          Check
        </button>
        <button
          :class="['detail-tab', activeTab === 'fix' ? 'active' : '']"
          @click="activeTab = 'fix'"
        >
          Fix
        </button>
        <button
          v-if="rule.query"
          :class="['detail-tab', activeTab === 'query' ? 'active' : '']"
          @click="activeTab = 'query'"
        >
          Query
        </button>
      </div>

      <!-- Tab content -->
      <div class="detail-content">
        <!-- Description tab -->
        <div v-if="activeTab === 'description'">
          <div class="detail-section">
            <h3>Vulnerability Discussion</h3>
            <p>{{ rule.description || 'No description available.' }}</p>
          </div>

          <div v-if="rule.registryChecks && rule.registryChecks.length > 0" class="detail-section">
            <h3>Registry Checks</h3>
            <div
              v-for="(check, idx) in rule.registryChecks"
              :key="idx"
              class="registry-check"
            >
              <div class="reg-row"><span class="reg-label">Hive:</span> {{ check.hive }}</div>
              <div class="reg-row"><span class="reg-label">Path:</span> {{ check.path }}</div>
              <div class="reg-row"><span class="reg-label">Value:</span> {{ check.valueName }}</div>
              <div v-if="check.expectedValue" class="reg-row"><span class="reg-label">Expected:</span> {{ check.expectedValue }}</div>
              <div class="reg-row"><span class="reg-label">Comparison:</span> {{ check.comparison }}</div>
            </div>
          </div>

          <div v-if="rule.tags && rule.tags.length > 0" class="detail-section">
            <h3>Tags</h3>
            <div class="tags">
              <span v-for="tag in rule.tags" :key="tag" class="tag">{{ tag }}</span>
            </div>
          </div>
        </div>

        <!-- Check tab -->
        <div v-if="activeTab === 'check'">
          <div class="detail-section">
            <h3>Check Content</h3>
            <p v-if="rule.checkContent">{{ rule.checkContent }}</p>
            <p v-else class="text-muted">No check content available. See Query tab for osquery SQL.</p>
          </div>
        </div>

        <!-- Fix tab -->
        <div v-if="activeTab === 'fix'">
          <div class="detail-section">
            <h3>Resolution</h3>
            <p>{{ rule.fixText || 'No fix text available.' }}</p>
          </div>

          <div v-if="rule.fix" class="detail-section">
            <h3>Fix Script</h3>
            <CodePreview
              :code="rule.fix.content"
              :language="rule.fix.type === 'ps1' ? 'powershell' : 'xml'"
              :filename="rule.fix.filename"
            />
          </div>
        </div>

        <!-- Query tab -->
        <div v-if="activeTab === 'query' && rule.query">
          <div class="detail-section">
            <h3>osquery SQL</h3>
            <CodePreview
              :code="rule.query"
              language="sql"
              filename="query.sql"
            />
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.rule-detail {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.registry-check {
  background: var(--color-background);
  padding: var(--spacing-sm);
  border-radius: var(--radius-sm);
  margin-bottom: var(--spacing-xs);
  font-size: 11px;
  font-family: var(--font-mono);
}

.reg-row {
  margin-bottom: 2px;
}

.reg-row:last-child {
  margin-bottom: 0;
}

.reg-label {
  color: var(--color-text-muted);
  font-weight: 500;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
}

.tag {
  padding: 1px 6px;
  background: var(--color-background);
  border-radius: var(--radius-xs);
  font-size: 10px;
  color: var(--color-text-secondary);
}

.text-muted {
  color: var(--color-text-muted);
  font-style: italic;
}
</style>
