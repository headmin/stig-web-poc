<script setup lang="ts">
import { useBenchmark } from '@/composables/useBenchmark'
import { useSelection } from '@/composables/useSelection'

const { categories, selectedCategoryId, selectCategory, categoryStats } = useBenchmark()
const { selectedRuleIds } = useSelection()

function getSelectedCount(categoryId: string): number {
  const cat = categories.value.find(c => c.id === categoryId)
  if (!cat) return 0
  return cat.rules.filter(r => selectedRuleIds.value.has(r.ruleId || r.id)).length
}

function getCategoryClass(categoryId: string): string[] {
  const classes = ['category-header']
  if (selectedCategoryId.value === categoryId) {
    classes.push('active')
  }
  return classes
}

function getTotalSelected(): number {
  return categories.value.reduce((sum, cat) => {
    return sum + cat.rules.filter(r => selectedRuleIds.value.has(r.ruleId || r.id)).length
  }, 0)
}
</script>

<template>
  <nav class="category-nav">
    <div class="category-nav-header">Categories</div>

    <div
      v-for="cat in categoryStats"
      :key="cat.id"
      class="category-item"
    >
      <div
        :class="getCategoryClass(cat.id)"
        @click="selectCategory(cat.id)"
      >
        <span class="name">{{ cat.name }}</span>
        <span class="count">
          <template v-if="getSelectedCount(cat.id) > 0">
            {{ getSelectedCount(cat.id) }}/
          </template>
          {{ cat.total }}
        </span>
      </div>
    </div>

    <!-- Show all option -->
    <div class="category-divider"></div>
    <div class="category-item">
      <div
        :class="['category-header', selectedCategoryId === null ? 'active' : '']"
        @click="selectCategory(null)"
      >
        <span class="name">All Rules</span>
        <span class="count">
          <template v-if="getTotalSelected() > 0">
            {{ getTotalSelected() }}/
          </template>
          {{ categories.reduce((sum, c) => sum + c.rules.length, 0) }}
        </span>
      </div>
    </div>
  </nav>
</template>

<style scoped>
.category-divider {
  margin: 8px 6px;
  border-top: 1px solid var(--color-border-light);
}
</style>
