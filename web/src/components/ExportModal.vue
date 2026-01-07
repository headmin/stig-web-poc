<script setup lang="ts">
import { ref, computed } from 'vue'
import { useSelection } from '@/composables/useSelection'
import { useBenchmark } from '@/composables/useBenchmark'
import { useExport, type ExportFormat } from '@/composables/useExport'

const emit = defineEmits<{
  close: []
}>()

const { selectedCount, getSelectedIds } = useSelection()
const { allRules } = useBenchmark()
const { exportToZip, isExporting } = useExport()

const profileName = ref('stig-windows11-policies')
const includeFixes = ref(true)
const splitBySeverity = ref(false)
const exportFormat = ref<ExportFormat>('gitops')

const selectedRules = computed(() => {
  const ids = getSelectedIds()
  return allRules.value.filter(r => ids.includes(r.ruleId || r.id))
})

const rulesWithFixes = computed(() => {
  return selectedRules.value.filter(r => r.fix).length
})

const outputPreview = computed(() => {
  const name = profileName.value || 'export'

  if (exportFormat.value === 'gitops') {
    let tree = `${name}/
├── default.yml
├── lib/
│   └── windows/`

    if (includeFixes.value && rulesWithFixes.value > 0) {
      // With fixes: show full structure
      tree += `
│       ├── policies/`
      if (splitBySeverity.value) {
        tree += `
│       │   ├── stig-high.policies.yml
│       │   ├── stig-medium.policies.yml
│       │   └── stig-low.policies.yml`
      } else {
        tree += `
│       │   └── stig.policies.yml`
      }
      tree += `
│       ├── scripts/
│       │   └── *.ps1
│       └── configuration-profiles/
│           └── *.xml`
    } else {
      // Without fixes: just policies
      tree += `
│       └── policies/`
      if (splitBySeverity.value) {
        tree += `
│           ├── stig-high.policies.yml
│           ├── stig-medium.policies.yml
│           └── stig-low.policies.yml`
      } else {
        tree += `
│           └── stig.policies.yml`
      }
    }

    tree += `
└── README.md`

    return tree
  } else {
    let tree = `${name}/
├── policies/`

    if (splitBySeverity.value) {
      tree += `
│   ├── high.yml
│   ├── medium.yml
│   └── low.yml`
    } else {
      tree += `
│   └── policies.yml`
    }

    if (includeFixes.value && rulesWithFixes.value > 0) {
      tree += `
├── fixes/
│   ├── xml/
│   └── ps1/`
    }

    tree += `
└── README.md`

    return tree
  }
})

async function handleExport() {
  await exportToZip({
    rules: selectedRules.value,
    profileName: profileName.value,
    includeFixes: includeFixes.value,
    splitBySeverity: splitBySeverity.value,
    format: exportFormat.value
  })
  emit('close')
}

function handleClose() {
  emit('close')
}
</script>

<template>
  <div class="modal-overlay" @click.self="handleClose">
    <div class="modal">
      <div class="modal-header">
        <h2>Export Configuration</h2>
        <button class="modal-close" @click="handleClose">&times;</button>
      </div>

      <div class="modal-body">
        <div class="export-summary">
          <div class="summary-item">
            <span class="label">Selected Rules</span>
            <span class="value">{{ selectedCount }}</span>
          </div>
          <div class="summary-item">
            <span class="label">With Fixes</span>
            <span class="value">{{ rulesWithFixes }}</span>
          </div>
        </div>

        <div class="form-group">
          <label>Export Format</label>
          <div class="format-options">
            <label class="format-option" :class="{ active: exportFormat === 'gitops' }">
              <input type="radio" v-model="exportFormat" value="gitops" />
              <div class="format-content">
                <span class="format-name">Fleet GitOps</span>
                <span class="format-desc">Simple list format for fleetctl gitops</span>
              </div>
            </label>
            <label class="format-option" :class="{ active: exportFormat === 'legacy' }">
              <input type="radio" v-model="exportFormat" value="legacy" />
              <div class="format-content">
                <span class="format-name">Fileset</span>
                <span class="format-desc">apiVersion/kind structure</span>
              </div>
            </label>
          </div>
        </div>

        <div class="form-group">
          <label for="profileName">Profile Name</label>
          <input
            id="profileName"
            type="text"
            v-model="profileName"
            placeholder="e.g., stig-windows11-policies"
          />
        </div>

        <div class="form-group">
          <label class="checkbox-label">
            <input type="checkbox" v-model="includeFixes" />
            Include fix files (XML/PowerShell)
          </label>
        </div>

        <div class="form-group">
          <label class="checkbox-label">
            <input type="checkbox" v-model="splitBySeverity" />
            Split policies by severity level
          </label>
        </div>

        <div class="output-preview">
          <h4>Output Structure</h4>
          <pre class="tree">{{ outputPreview }}</pre>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" @click="handleClose">Cancel</button>
        <button
          class="btn btn-primary"
          @click="handleExport"
          :disabled="isExporting || selectedCount === 0"
        >
          {{ isExporting ? 'Exporting...' : 'Download ZIP' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.export-summary {
  display: flex;
  gap: var(--spacing-lg);
  padding: var(--spacing-md);
  background: var(--color-background);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-md);
}

.summary-item {
  display: flex;
  flex-direction: column;
}

.summary-item .label {
  font-size: 11px;
  color: var(--color-text-muted);
}

.summary-item .value {
  font-size: 20px;
  font-weight: 600;
  color: var(--color-primary);
}

.format-options {
  display: flex;
  gap: var(--spacing-sm);
  margin-top: var(--spacing-xs);
}

.format-option {
  flex: 1;
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.15s;
}

.format-option:hover {
  border-color: var(--color-primary-light);
}

.format-option.active {
  border-color: var(--color-primary);
  background: rgba(106, 103, 206, 0.05);
}

.format-option input[type="radio"] {
  margin-top: 2px;
  accent-color: var(--color-primary);
}

.format-content {
  display: flex;
  flex-direction: column;
}

.format-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text);
}

.format-desc {
  font-size: 10px;
  color: var(--color-text-muted);
}

.output-preview {
  margin-top: var(--spacing-md);
  padding: var(--spacing-sm);
  background: var(--color-background);
  border-radius: var(--radius-sm);
}

.output-preview h4 {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--color-text-muted);
  margin-bottom: var(--spacing-xs);
}

.output-preview .tree {
  font-family: var(--font-mono);
  font-size: 10px;
  line-height: 1.4;
  color: var(--color-text-secondary);
  margin: 0;
  white-space: pre;
}
</style>
