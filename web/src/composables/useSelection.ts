import { ref, computed } from 'vue'
import type { Rule } from '@/types/benchmark'

// Set of selected rule IDs (using ruleId as key)
const selectedRuleIds = ref<Set<string>>(new Set())

export function useSelection() {
  // Check if a rule is selected
  const isSelected = (rule: Rule): boolean => {
    return selectedRuleIds.value.has(rule.ruleId || rule.id)
  }

  // Toggle selection
  const toggleSelection = (rule: Rule) => {
    const key = rule.ruleId || rule.id
    if (selectedRuleIds.value.has(key)) {
      selectedRuleIds.value.delete(key)
    } else {
      selectedRuleIds.value.add(key)
    }
    // Trigger reactivity
    selectedRuleIds.value = new Set(selectedRuleIds.value)
  }

  // Select a rule
  const selectRule = (rule: Rule) => {
    const key = rule.ruleId || rule.id
    selectedRuleIds.value.add(key)
    selectedRuleIds.value = new Set(selectedRuleIds.value)
  }

  // Deselect a rule
  const deselectRule = (rule: Rule) => {
    const key = rule.ruleId || rule.id
    selectedRuleIds.value.delete(key)
    selectedRuleIds.value = new Set(selectedRuleIds.value)
  }

  // Select multiple rules
  const selectRules = (rules: Rule[]) => {
    rules.forEach(rule => {
      const key = rule.ruleId || rule.id
      selectedRuleIds.value.add(key)
    })
    selectedRuleIds.value = new Set(selectedRuleIds.value)
  }

  // Deselect multiple rules
  const deselectRules = (rules: Rule[]) => {
    rules.forEach(rule => {
      const key = rule.ruleId || rule.id
      selectedRuleIds.value.delete(key)
    })
    selectedRuleIds.value = new Set(selectedRuleIds.value)
  }

  // Clear all selections
  const clearSelection = () => {
    selectedRuleIds.value = new Set()
  }

  // Select all from a list
  const selectAll = (rules: Rule[]) => {
    selectRules(rules)
  }

  // Deselect all from a list
  const deselectAll = (rules: Rule[]) => {
    deselectRules(rules)
  }

  // Count of selected rules
  const selectedCount = computed(() => selectedRuleIds.value.size)

  // Get all selected rule IDs
  const getSelectedIds = (): string[] => {
    return Array.from(selectedRuleIds.value)
  }

  // Check if all rules in a list are selected
  const areAllSelected = (rules: Rule[]): boolean => {
    if (rules.length === 0) return false
    return rules.every(rule => isSelected(rule))
  }

  // Check if some (but not all) rules in a list are selected
  const areSomeSelected = (rules: Rule[]): boolean => {
    if (rules.length === 0) return false
    const selectedInList = rules.filter(rule => isSelected(rule))
    return selectedInList.length > 0 && selectedInList.length < rules.length
  }

  return {
    selectedRuleIds,
    selectedCount,
    isSelected,
    toggleSelection,
    selectRule,
    deselectRule,
    selectRules,
    deselectRules,
    selectAll,
    deselectAll,
    clearSelection,
    getSelectedIds,
    areAllSelected,
    areSomeSelected
  }
}
