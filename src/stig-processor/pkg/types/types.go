package types

import "time"

// STIG data structures represent the input JSON format from DISA STIG files
type STIGBenchmark struct {
	ID          int         `json:"id"`
	BenchmarkID string      `json:"benchmarkId"`
	Slug        string      `json:"slug"`
	Status      string      `json:"status"`
	StatusDate  string      `json:"statusDate"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Version     string      `json:"version"`
	CreatedAt   string      `json:"createdAt"`
	UpdatedAt   string      `json:"updatedAt"`
	Groups      []STIGGroup `json:"groups"`
}

type STIGGroup struct {
	ID                 int    `json:"id"`
	BenchmarkID        int    `json:"benchmarkId"`
	GroupID            string `json:"groupId"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	RuleID             string `json:"ruleId"`
	RuleWeight         string `json:"ruleWeight"`
	RuleSeverity       string `json:"ruleSeverity"`
	RuleVersion        string `json:"ruleVersion"`
	RuleTitle          string `json:"ruleTitle"`
	RuleVulnDiscussion string `json:"ruleVulnDiscussion"`
	RuleFalsePositives string `json:"ruleFalsePositives"`
	RuleFalseNegatives string `json:"ruleFalseNegatives"`
	RuleDocumentable   string `json:"ruleDocumentable"`
	RuleMitigations    string `json:"ruleMitigations"`
	RuleIdent          string `json:"ruleIdent"`
	RuleFixText        string `json:"ruleFixText"`
	RuleFixID          string `json:"ruleFixId"`
	RuleCheckSystem    string `json:"ruleCheckSystem"`
	RuleCheckContent   string `json:"ruleCheckContent"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}

// Fleet policy structures represent the output YAML format for Fleet
type FleetPolicy struct {
	APIVersion string     `yaml:"apiVersion" json:"apiVersion"`
	Kind       string     `yaml:"kind" json:"kind"`
	Metadata   PolicyMeta `yaml:"metadata" json:"metadata"`
	Spec       PolicySpec `yaml:"spec" json:"spec"`
}

type PolicyMeta struct {
	Name        string            `yaml:"name" json:"name"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

type PolicySpec struct {
	Name        string `yaml:"name" json:"name"`
	Query       string `yaml:"query" json:"query"`
	Description string `yaml:"description" json:"description"`
	Resolution  string `yaml:"resolution" json:"resolution"`
	Platform    string `yaml:"platform" json:"platform"`
	Critical    bool   `yaml:"critical" json:"critical"`
}

// Registry check structure represents parsed Windows registry information
type RegistryCheck struct {
	Hive       string
	Path       string
	ValueName  string
	ValueType  string
	Value      string
	Comparison string // "equals", "greater_equal", "less_equal", "not_exists", "must_exist"
}

// Processing configuration and results
type ProcessingOptions struct {
	InputFile string
	OutputDir string
	Format    string
	Severity  string
	Verbose   bool
	DryRun    bool
	Pretty    bool
	Timeout   time.Duration
}

type ProcessingResult struct {
	Total        int
	Automatable  int
	ManualReview int
	Policies     []FleetPolicy
	Errors       []ProcessingError
	Duration     time.Duration
}

type ProcessingError struct {
	GroupID   string
	RuleID    string
	Message   string
	Type      ErrorType
	Timestamp time.Time
}

type ErrorType string

const (
	ErrorTypeParsingFailed    ErrorType = "parsing_failed"
	ErrorTypeValidationFailed ErrorType = "validation_failed"
	ErrorTypeFileWriteFailed  ErrorType = "file_write_failed"
	ErrorTypeUnknown          ErrorType = "unknown"
)

// Statistics and summary structures
type ProcessingStatistics struct {
	Title                string
	Version              string
	TotalRules           int
	RegistryRules        int
	GroupPolicyRules     int
	ManualRules          int
	SeverityDistribution map[string]int
	ProcessingTime       time.Duration
}

type ProcessingSummary struct {
	TotalRules        int                 `yaml:"total_rules" json:"total_rules"`
	Automatable       int                 `yaml:"automatable" json:"automatable"`
	ManualReview      int                 `yaml:"manual_review" json:"manual_review"`
	PoliciesGenerated int                 `yaml:"policies_generated" json:"policies_generated"`
	ProcessingTime    string              `yaml:"processing_time" json:"processing_time"`
	Timestamp         string              `yaml:"timestamp" json:"timestamp"`
	Policies          []PolicySummaryItem `yaml:"policies" json:"policies"`
	Errors            []ProcessingError   `yaml:"errors,omitempty" json:"errors,omitempty"`
}

type PolicySummaryItem struct {
	Name        string `yaml:"name" json:"name"`
	Title       string `yaml:"title" json:"title"`
	Platform    string `yaml:"platform" json:"platform"`
	Critical    bool   `yaml:"critical" json:"critical"`
	Severity    string `yaml:"severity" json:"severity"`
	GroupID     string `yaml:"group_id" json:"group_id"`
	RuleVersion string `yaml:"rule_version" json:"rule_version"`
}

// Validation structures
type ValidationResult struct {
	Valid  bool
	Count  int
	Errors []ValidationError
}

type ValidationError struct {
	FilePath string
	LineNum  int
	Message  string
	Type     ValidationErrorType
}

type ValidationErrorType string

const (
	ValidationErrorYAMLSyntax  ValidationErrorType = "yaml_syntax"
	ValidationErrorJSONSyntax  ValidationErrorType = "json_syntax"
	ValidationErrorFleetSchema ValidationErrorType = "fleet_schema"
	ValidationErrorSQLSyntax   ValidationErrorType = "sql_syntax"
)

// Severity levels enumeration
type SeverityLevel string

const (
	SeverityLow    SeverityLevel = "low"
	SeverityMedium SeverityLevel = "medium"
	SeverityHigh   SeverityLevel = "high"
)

// Valid severity levels for validation
var ValidSeverityLevels = []SeverityLevel{
	SeverityLow,
	SeverityMedium,
	SeverityHigh,
}

// Registry hive constants for Windows registry paths
const (
	HKeyLocalMachine  = "HKEY_LOCAL_MACHINE"
	HKeyCurrentUser   = "HKEY_CURRENT_USER"
	HKeyUsers         = "HKEY_USERS"
	HKeyClassesRoot   = "HKEY_CLASSES_ROOT"
	HKeyCurrentConfig = "HKEY_CURRENT_CONFIG"
)

// Registry value types for Windows registry values
const (
	RegSZ       = "REG_SZ"
	RegExpandSZ = "REG_EXPAND_SZ"
	RegBinary   = "REG_BINARY"
	RegDWord    = "REG_DWORD"
	RegQWord    = "REG_QWORD"
	RegMultiSZ  = "REG_MULTI_SZ"
)

// Fleet policy constants
const (
	FleetAPIVersion = "v1"
	FleetKindPolicy = "policy"
	PlatformWindows = "windows"
	PlatformLinux   = "linux"
	PlatformMacOS   = "darwin"
)

// Processing events for logging and monitoring
type ProcessingEvent struct {
	Type      EventType `json:"type"`
	Message   string    `json:"message"`
	GroupID   string    `json:"group_id,omitempty"`
	RuleID    string    `json:"rule_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Details   any       `json:"details,omitempty"`
}

type EventType string

const (
	EventTypeInfo    EventType = "info"
	EventTypeWarning EventType = "warning"
	EventTypeError   EventType = "error"
	EventTypeSuccess EventType = "success"
	EventTypeDebug   EventType = "debug"
)

// Configuration defaults
const (
	DefaultOutputFormat   = "yaml"
	DefaultOutputDir      = "output"
	DefaultInputFile      = "microsoft-windows-11-security-technical-implementation-guide.json"
	DefaultTimeout        = 5 * time.Minute
	DefaultMaxFileSize    = 100 * 1024 * 1024 // 100MB
	DefaultMaxPolicyCount = 1000
)
