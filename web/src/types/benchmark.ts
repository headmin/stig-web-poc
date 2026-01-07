// Unified benchmark data schema - matches Go schema-builder output

export interface BenchmarkData {
  meta: Meta
  categories: Category[]
}

export interface Meta {
  framework: string
  title: string
  version: string
  generatedAt: string
}

export interface Category {
  id: string
  name: string
  description?: string
  rules: Rule[]
}

export interface Rule {
  id: string
  ruleId: string
  title: string
  severity: 'high' | 'medium' | 'low'
  description: string
  checkContent: string
  fixText: string
  automatable: boolean
  query?: string
  fix?: Fix
  registryChecks?: RegistryCheck[]
  cci?: string
  weight?: string
  tags: string[]
}

export interface Fix {
  filename: string
  type: 'xml' | 'ps1'
  content: string
}

export interface RegistryCheck {
  hive: string
  path: string
  valueName: string
  valueType?: string
  expectedValue?: string
  comparison: 'equals' | 'greater_equal' | 'less_equal' | 'not_exists' | 'must_exist'
}

// Severity helpers
export const severityOrder: Record<string, number> = {
  high: 0,
  medium: 1,
  low: 2
}

export const severityColors: Record<string, string> = {
  high: '#dc3545',
  medium: '#ffc107',
  low: '#28a745'
}
