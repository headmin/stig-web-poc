import { ref, computed } from "vue";
import type { BenchmarkData, Category, Rule } from "@/types/benchmark";
import benchmarkData from "@/data/benchmark-data.json";

// Cast the imported data to our type
const data = benchmarkData as BenchmarkData;

// Reactive state
const selectedCategoryId = ref<string | null>(null);
const searchQuery = ref("");
const severityFilter = ref<string[]>(["high", "medium", "low"]);
const automatableFilter = ref<"all" | "automatable" | "manual">("all");
const hasFixFilter = ref<boolean>(false);

export function useBenchmark() {
  // All categories
  const categories = computed(() => data.categories);

  // Metadata
  const meta = computed(() => data.meta);

  // Selected category
  const selectedCategory = computed(() => {
    if (!selectedCategoryId.value) return null;
    return (
      categories.value.find((c) => c.id === selectedCategoryId.value) ?? null
    );
  });

  // Filtered rules for selected category
  const filteredRules = computed(() => {
    if (!selectedCategory.value) return [];

    let rules = selectedCategory.value.rules;

    // Apply search filter
    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase();
      rules = rules.filter(
        (rule) =>
          rule.title.toLowerCase().includes(query) ||
          rule.ruleId.toLowerCase().includes(query) ||
          rule.id.toLowerCase().includes(query) ||
          rule.description.toLowerCase().includes(query),
      );
    }

    // Apply severity filter
    rules = rules.filter((rule) =>
      severityFilter.value.includes(rule.severity),
    );

    // Apply automatable filter
    if (automatableFilter.value === "automatable") {
      rules = rules.filter((rule) => rule.automatable);
    } else if (automatableFilter.value === "manual") {
      rules = rules.filter((rule) => !rule.automatable);
    }

    // Apply has fix filter
    if (hasFixFilter.value) {
      rules = rules.filter((rule) => rule.fix);
    }

    return rules;
  });

  // All rules (flattened) with filters applied
  const allRules = computed(() => {
    let rules = categories.value.flatMap((c) => c.rules);

    // Apply search filter
    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase();
      rules = rules.filter(
        (rule) =>
          rule.title.toLowerCase().includes(query) ||
          rule.ruleId.toLowerCase().includes(query) ||
          rule.id.toLowerCase().includes(query) ||
          rule.description.toLowerCase().includes(query),
      );
    }

    // Apply severity filter
    rules = rules.filter((rule) =>
      severityFilter.value.includes(rule.severity),
    );

    // Apply automatable filter
    if (automatableFilter.value === "automatable") {
      rules = rules.filter((rule) => rule.automatable);
    } else if (automatableFilter.value === "manual") {
      rules = rules.filter((rule) => !rule.automatable);
    }

    // Apply has fix filter
    if (hasFixFilter.value) {
      rules = rules.filter((rule) => rule.fix);
    }

    return rules;
  });

  // Global search across all categories
  const globalSearchResults = computed(() => {
    if (!searchQuery.value || searchQuery.value.length < 2) return [];

    const query = searchQuery.value.toLowerCase();
    return allRules.value
      .filter(
        (rule) =>
          rule.title.toLowerCase().includes(query) ||
          rule.ruleId.toLowerCase().includes(query) ||
          rule.id.toLowerCase().includes(query),
      )
      .slice(0, 20); // Limit results
  });

  // Statistics
  const stats = computed(() => {
    const all = allRules.value;
    return {
      total: all.length,
      automatable: all.filter((r) => r.automatable).length,
      manual: all.filter((r) => !r.automatable).length,
      withFixes: all.filter((r) => r.fix).length,
      high: all.filter((r) => r.severity === "high").length,
      medium: all.filter((r) => r.severity === "medium").length,
      low: all.filter((r) => r.severity === "low").length,
    };
  });

  // Category stats
  const categoryStats = computed(() => {
    return categories.value.map((cat) => ({
      id: cat.id,
      name: cat.name,
      total: cat.rules.length,
      automatable: cat.rules.filter((r) => r.automatable).length,
      high: cat.rules.filter((r) => r.severity === "high").length,
      medium: cat.rules.filter((r) => r.severity === "medium").length,
      low: cat.rules.filter((r) => r.severity === "low").length,
    }));
  });

  // Actions
  function selectCategory(categoryId: string | null) {
    selectedCategoryId.value = categoryId;
  }

  function setSearchQuery(query: string) {
    searchQuery.value = query;
  }

  function toggleSeverity(severity: string) {
    const idx = severityFilter.value.indexOf(severity);
    if (idx === -1) {
      severityFilter.value.push(severity);
    } else {
      severityFilter.value.splice(idx, 1);
    }
  }

  function setAutomatableFilter(filter: "all" | "automatable" | "manual") {
    automatableFilter.value = filter;
  }

  function toggleHasFixFilter() {
    hasFixFilter.value = !hasFixFilter.value;
  }

  function getRuleById(ruleId: string): Rule | undefined {
    return allRules.value.find((r) => r.ruleId === ruleId || r.id === ruleId);
  }

  function getCategoryForRule(rule: Rule): Category | undefined {
    return categories.value.find((c) =>
      c.rules.some((r) => r.ruleId === rule.ruleId),
    );
  }

  return {
    // State
    categories,
    meta,
    selectedCategoryId,
    selectedCategory,
    searchQuery,
    severityFilter,
    automatableFilter,
    hasFixFilter,

    // Computed
    filteredRules,
    allRules,
    globalSearchResults,
    stats,
    categoryStats,

    // Actions
    selectCategory,
    setSearchQuery,
    toggleSeverity,
    setAutomatableFilter,
    toggleHasFixFilter,
    getRuleById,
    getCategoryForRule,
  };
}
