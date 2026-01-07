package schema

// BenchmarkData is the unified schema for benchmark data consumed by the web UI
type BenchmarkData struct {
	Meta       Meta       `json:"meta"`
	Categories []Category `json:"categories"`
}

// Meta contains metadata about the benchmark
type Meta struct {
	Framework   string `json:"framework"`   // e.g., "STIG", "CIS"
	Title       string `json:"title"`       // e.g., "Windows 11 Security Technical Implementation Guide"
	Version     string `json:"version"`     // e.g., "v2r2"
	GeneratedAt string `json:"generatedAt"` // ISO timestamp
}

// Category groups related rules together
type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Rules       []Rule `json:"rules"`
}

// Rule represents a single benchmark rule/control
type Rule struct {
	ID       string `json:"id"`     // e.g., "V-253380"
	RuleID   string `json:"ruleId"` // e.g., "WN11-00-000001"
	Title    string `json:"title"`
	Severity string `json:"severity"` // "high", "medium", "low"

	// Content
	Description  string `json:"description"`  // Vulnerability discussion
	CheckContent string `json:"checkContent"` // How to verify compliance
	FixText      string `json:"fixText"`      // Manual remediation instructions

	// Automation
	Automatable bool   `json:"automatable"`
	Query       string `json:"query,omitempty"` // osquery SQL (if automatable)

	// Linked fix file
	Fix *Fix `json:"fix,omitempty"`

	// Registry details (if applicable)
	RegistryChecks []RegistryCheck `json:"registryChecks,omitempty"`

	// Metadata
	CCI    string   `json:"cci,omitempty"`
	Weight string   `json:"weight,omitempty"`
	Tags   []string `json:"tags"`
}

// Fix represents a remediation script/config file
type Fix struct {
	Filename string `json:"filename"` // e.g., "SolicitedRemoteAssistance.xml"
	Type     string `json:"type"`     // "xml" or "ps1"
	Content  string `json:"content"`  // Embedded file content
}

// RegistryCheck represents a Windows registry check
type RegistryCheck struct {
	Hive          string `json:"hive"`
	Path          string `json:"path"`
	ValueName     string `json:"valueName"`
	ValueType     string `json:"valueType,omitempty"`
	ExpectedValue string `json:"expectedValue,omitempty"`
	Comparison    string `json:"comparison"` // "equals", "greater_equal", "less_equal", "not_exists", "must_exist"
}

// Severity levels
const (
	SeverityHigh   = "high"
	SeverityMedium = "medium"
	SeverityLow    = "low"
)

// Fix types
const (
	FixTypeXML        = "xml"
	FixTypePowerShell = "ps1"
)
