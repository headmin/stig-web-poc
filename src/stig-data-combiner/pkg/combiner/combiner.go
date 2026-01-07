package combiner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/stig-data-combiner/pkg/schema"
)

// STIGGroup represents a rule from the DISA STIG JSON
type STIGGroup struct {
	GroupID            string `json:"groupId"`            // V-253281
	RuleID             string `json:"ruleId"`             // SV-253281r991589_rule
	RuleVersion        string `json:"ruleVersion"`        // WN11-00-000135
	RuleTitle          string `json:"ruleTitle"`          // A host-based firewall must be...
	RuleSeverity       string `json:"ruleSeverity"`       // high, medium, low
	RuleVulnDiscussion string `json:"ruleVulnDiscussion"` // Description
	RuleFixText        string `json:"ruleFixText"`        // Resolution
	RuleCheckContent   string `json:"ruleCheckContent"`   // Check instructions
	RuleIdent          string `json:"ruleIdent"`          // CCI-000366
}

// STIGData represents the DISA STIG JSON structure
type STIGData struct {
	Title   string      `json:"title"`
	Version string      `json:"version"`
	Groups  []STIGGroup `json:"groups"`
}

// WinSTIGPolicy represents a policy from win-stig/stig-policy-queries.yml
type WinSTIGPolicy struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Spec       WinSTIGPolicySpec `yaml:"spec"`
}

// WinSTIGPolicySpec contains the policy specification
type WinSTIGPolicySpec struct {
	Name         string `yaml:"name"`
	Platform     string `yaml:"platform"`
	Platforms    string `yaml:"platforms"`
	Description  string `yaml:"description"`
	Resolution   string `yaml:"resolution"`
	Query        string `yaml:"query"`
	Purpose      string `yaml:"purpose"`
	Tags         string `yaml:"tags"`
	Contributors string `yaml:"contributors"`
	Fix          string `yaml:"fix,omitempty"`
}

// Combiner merges STIG rules with fix files
type Combiner struct {
	stigPath    string // Path to STIG JSON file
	winSTIGPath string // Path to win-stig repository
	verbose     bool
}

// NewCombiner creates a new Combiner instance
func NewCombiner(stigPath, winSTIGPath string, verbose bool) *Combiner {
	return &Combiner{
		stigPath:    stigPath,
		winSTIGPath: winSTIGPath,
		verbose:     verbose,
	}
}

// Combine reads all sources and produces unified BenchmarkData
func (c *Combiner) Combine() (*schema.BenchmarkData, error) {
	// Read STIG JSON (primary source with proper IDs)
	stigData, err := c.readSTIGJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to read STIG JSON: %w", err)
	}

	if c.verbose {
		fmt.Printf("Read %d rules from STIG JSON\n", len(stigData.Groups))
	}

	// Read win-stig policies (has osquery SQL and fix mappings)
	policies, err := c.readWinSTIGPolicies()
	if err != nil {
		return nil, fmt.Errorf("failed to read win-stig policies: %w", err)
	}

	if c.verbose {
		fmt.Printf("Read %d policies from win-stig\n", len(policies))
	}

	// Read fix files
	fixes, err := c.readFixFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to read fix files: %w", err)
	}

	if c.verbose {
		fmt.Printf("Read %d fix files\n", len(fixes))
	}

	// Build policy lookup by title (normalized)
	policyByTitle := c.buildPolicyIndex(policies)

	// Convert STIG groups to rules, enriching with win-stig data
	rules := c.convertToRules(stigData.Groups, policyByTitle, fixes)

	// Categorize rules
	categories := c.categorizeRules(rules)

	// Build final data
	data := &schema.BenchmarkData{
		Meta: schema.Meta{
			Framework:   "STIG",
			Title:       stigData.Title,
			Version:     stigData.Version,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		},
		Categories: categories,
	}

	return data, nil
}

// readSTIGJSON reads the DISA STIG JSON file
func (c *Combiner) readSTIGJSON() (*STIGData, error) {
	// Look for STIG JSON in common locations
	searchPaths := []string{
		c.stigPath,
		filepath.Join(filepath.Dir(c.winSTIGPath), "microsoft-windows-11-security-technical-implementation-guide.json"),
		"microsoft-windows-11-security-technical-implementation-guide.json",
	}

	var data []byte
	var err error
	var foundPath string

	for _, path := range searchPaths {
		if path == "" {
			continue
		}
		data, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}

	if data == nil {
		return nil, fmt.Errorf("STIG JSON not found in any expected location")
	}

	if c.verbose {
		fmt.Printf("Using STIG JSON from: %s\n", foundPath)
	}

	var stigData STIGData
	if err := json.Unmarshal(data, &stigData); err != nil {
		return nil, fmt.Errorf("failed to parse STIG JSON: %w", err)
	}

	return &stigData, nil
}

// readWinSTIGPolicies reads the stig-policy-queries.yml file
func (c *Combiner) readWinSTIGPolicies() ([]WinSTIGPolicy, error) {
	path := filepath.Join(c.winSTIGPath, "stig-policy-queries.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	// The file contains multiple YAML documents separated by ---
	var policies []WinSTIGPolicy
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))

	for {
		var policy WinSTIGPolicy
		err := decoder.Decode(&policy)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			// Skip invalid documents
			continue
		}
		if policy.Kind == "policy" {
			policies = append(policies, policy)
		}
	}

	return policies, nil
}

// readFixFiles reads all fix files from win-stig/fix/
func (c *Combiner) readFixFiles() (map[string]*schema.Fix, error) {
	fixDir := filepath.Join(c.winSTIGPath, "fix")
	fixes := make(map[string]*schema.Fix)

	entries, err := os.ReadDir(fixDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read fix directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		ext := strings.ToLower(filepath.Ext(filename))

		var fixType string
		switch ext {
		case ".xml":
			fixType = schema.FixTypeXML
		case ".ps1":
			fixType = schema.FixTypePowerShell
		default:
			continue // Skip unknown file types
		}

		content, err := os.ReadFile(filepath.Join(fixDir, filename))
		if err != nil {
			if c.verbose {
				fmt.Printf("Warning: failed to read fix file %s: %v\n", filename, err)
			}
			continue
		}

		fixes[filename] = &schema.Fix{
			Filename: filename,
			Type:     fixType,
			Content:  string(content),
		}
	}

	return fixes, nil
}

// normalizeTitle normalizes a title for matching
func normalizeTitle(title string) string {
	// Remove common prefixes
	title = strings.TrimPrefix(title, "STIG - ")

	// Lowercase and trim
	title = strings.ToLower(strings.TrimSpace(title))

	// Remove punctuation for fuzzy matching
	title = strings.ReplaceAll(title, ".", "")
	title = strings.ReplaceAll(title, ",", "")
	title = strings.ReplaceAll(title, "'", "")
	title = strings.ReplaceAll(title, "\"", "")

	return title
}

// buildPolicyIndex creates a lookup map from normalized title to policy
func (c *Combiner) buildPolicyIndex(policies []WinSTIGPolicy) map[string]*WinSTIGPolicy {
	index := make(map[string]*WinSTIGPolicy)

	for i := range policies {
		policy := &policies[i]
		normalizedName := normalizeTitle(policy.Spec.Name)
		index[normalizedName] = policy
	}

	return index
}

// convertToRules converts STIG groups to unified Rule format, enriched with win-stig data
func (c *Combiner) convertToRules(groups []STIGGroup, policyByTitle map[string]*WinSTIGPolicy, fixes map[string]*schema.Fix) []schema.Rule {
	var rules []schema.Rule

	matched := 0
	unmatched := 0

	for _, group := range groups {
		// Normalize title for matching
		normalizedTitle := normalizeTitle(group.RuleTitle)

		// Find matching win-stig policy
		policy := policyByTitle[normalizedTitle]

		// Determine if automatable and get query
		automatable := false
		query := ""
		var fix *schema.Fix

		if policy != nil {
			matched++
			// Check if it's a real query (not just a manual check placeholder)
			query = policy.Spec.Query
			automatable = !strings.Contains(query, "Manual check required") &&
				!strings.Contains(query, "SELECT 0 WHERE 1=0")

			// Link fix file if specified
			if policy.Spec.Fix != "" {
				if f, exists := fixes[policy.Spec.Fix]; exists {
					fix = f
				}
			}
		} else {
			unmatched++
		}

		// Parse registry checks from check content
		registryChecks := parseRegistryChecks(group.RuleCheckContent)

		// Build title with STIG ID prefix
		title := fmt.Sprintf("%s - %s", group.RuleVersion, group.RuleTitle)

		rule := schema.Rule{
			ID:             group.GroupID,
			RuleID:         group.RuleVersion,
			Title:          title,
			Severity:       group.RuleSeverity,
			Description:    group.RuleVulnDiscussion,
			CheckContent:   group.RuleCheckContent,
			FixText:        group.RuleFixText,
			Automatable:    automatable,
			Query:          query,
			RegistryChecks: registryChecks,
			Fix:            fix,
			CCI:            group.RuleIdent,
			Tags:           []string{"STIG", "Windows11", group.RuleSeverity},
		}

		rules = append(rules, rule)
	}

	if c.verbose {
		fmt.Printf("Matched %d rules with win-stig policies, %d unmatched\n", matched, unmatched)
	}

	return rules
}

// categorizeRules groups rules into categories based on DISA STIG rule ID prefix
func (c *Combiner) categorizeRules(rules []schema.Rule) []schema.Category {
	// DISA STIG categories based on rule ID prefix
	prefixCategories := map[string]string{
		"WN11-00": "General Requirements",
		"WN11-AC": "Account Policies",
		"WN11-AU": "Audit Policy",
		"WN11-CC": "Computer Configuration",
		"WN11-PK": "Public Key Policies",
		"WN11-RG": "Registry",
		"WN11-SO": "Security Options",
		"WN11-UR": "User Rights Assignment",
	}

	// Define order
	categoryOrder := []string{"WN11-00", "WN11-AC", "WN11-AU", "WN11-CC", "WN11-PK", "WN11-RG", "WN11-SO", "WN11-UR"}

	categoryMap := make(map[string]*schema.Category)
	for _, prefix := range categoryOrder {
		categoryMap[prefix] = &schema.Category{
			ID:    prefix,
			Name:  prefixCategories[prefix],
			Rules: []schema.Rule{},
		}
	}

	for _, rule := range rules {
		// Extract prefix from ruleId (e.g., "WN11-AU-000005" -> "WN11-AU")
		prefix := "WN11-00" // default
		if len(rule.RuleID) >= 7 {
			p := rule.RuleID[:7]
			if _, ok := prefixCategories[p]; ok {
				prefix = p
			}
		}

		categoryMap[prefix].Rules = append(categoryMap[prefix].Rules, rule)
	}

	// Build ordered list, only include non-empty categories
	var categories []schema.Category
	for _, prefix := range categoryOrder {
		if cat, ok := categoryMap[prefix]; ok && len(cat.Rules) > 0 {
			categories = append(categories, *cat)
		}
	}

	return categories
}

// parseRegistryChecks extracts registry check info from STIG check content
func parseRegistryChecks(checkContent string) []schema.RegistryCheck {
	var checks []schema.RegistryCheck

	// Patterns for registry checks in STIG check content
	hivePattern := regexp.MustCompile(`Registry Hive:\s*(HKEY_[A-Z_]+)`)
	pathPattern := regexp.MustCompile(`Registry Path:\s*\\?([^\n]+)`)
	valueNamePattern := regexp.MustCompile(`Value Name:\s*([^\n]+)`)
	valueTypePattern := regexp.MustCompile(`(?:Value Type|Type):\s*(REG_[A-Z_]+)`)
	valuePattern := regexp.MustCompile(`Value:\s*([^\n]+)`)

	// Find all hives in the content
	hiveMatches := hivePattern.FindAllStringSubmatchIndex(checkContent, -1)

	for i, hiveMatch := range hiveMatches {
		if len(hiveMatch) < 4 {
			continue
		}

		// Get the section for this registry check
		start := hiveMatch[0]
		end := len(checkContent)
		if i+1 < len(hiveMatches) {
			end = hiveMatches[i+1][0]
		}
		section := checkContent[start:end]

		hive := checkContent[hiveMatch[2]:hiveMatch[3]]

		pathMatch := pathPattern.FindStringSubmatch(section)
		nameMatch := valueNamePattern.FindStringSubmatch(section)
		typeMatch := valueTypePattern.FindStringSubmatch(section)
		valueMatch := valuePattern.FindStringSubmatch(section)

		if len(pathMatch) >= 2 && len(nameMatch) >= 2 {
			path := strings.TrimSpace(pathMatch[1])
			path = strings.TrimSuffix(path, "\\")
			valueName := strings.TrimSpace(nameMatch[1])
			valueType := "REG_DWORD"
			if len(typeMatch) >= 2 {
				valueType = strings.TrimSpace(typeMatch[1])
			}
			expectedValue := ""
			if len(valueMatch) >= 2 {
				expectedValue = strings.TrimSpace(valueMatch[1])
			}

			// Parse the expected value and comparison
			comparison := "equals"
			cleanValue := expectedValue

			// Handle hex values like 0x00000001 (1)
			hexPattern := regexp.MustCompile(`0x[0-9a-fA-F]+\s*\((\d+)\)`)
			if hexMatch := hexPattern.FindStringSubmatch(expectedValue); len(hexMatch) >= 2 {
				cleanValue = hexMatch[1]
			}

			// Detect comparison type
			sectionLower := strings.ToLower(section)
			if strings.Contains(sectionLower, "or greater") || strings.Contains(section, ">=") {
				comparison = "greater_equal"
			} else if strings.Contains(sectionLower, "or fewer") || strings.Contains(sectionLower, "or less") {
				comparison = "less_equal"
			} else if strings.Contains(sectionLower, "must not exist") {
				comparison = "not_exists"
			} else if strings.Contains(sectionLower, "does not exist") && strings.Contains(sectionLower, "finding") {
				comparison = "must_exist"
			}

			checks = append(checks, schema.RegistryCheck{
				Hive:          hive,
				Path:          path,
				ValueName:     valueName,
				ValueType:     valueType,
				ExpectedValue: cleanValue,
				Comparison:    comparison,
			})
		}
	}

	return checks
}
