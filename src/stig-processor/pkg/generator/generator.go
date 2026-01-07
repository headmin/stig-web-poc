package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/stig-processor/pkg/parser"
	"github.com/stig-processor/pkg/types"
)

// FleetPolicyGenerator handles generating Fleet policies from STIG rules
type FleetPolicyGenerator struct {
	options   *types.ProcessingOptions
	regParser *parser.RegistryParser
	stats     *types.ProcessingStatistics
}

// NewFleetPolicyGenerator creates a new Fleet policy generator
func NewFleetPolicyGenerator(options *types.ProcessingOptions) *FleetPolicyGenerator {
	return &FleetPolicyGenerator{
		options:   options,
		regParser: parser.NewRegistryParser(options.Verbose),
	}
}

// GeneratePolicy creates a Fleet policy from a STIG rule and registry checks
func (g *FleetPolicyGenerator) GeneratePolicy(group *types.STIGGroup, regChecks []*types.RegistryCheck) (*types.FleetPolicy, error) {
	// Validate inputs
	if group == nil {
		return nil, fmt.Errorf("group cannot be nil")
	}
	if len(regChecks) == 0 {
		return nil, fmt.Errorf("registry checks cannot be empty")
	}

	// Generate osquery SQL
	query := g.regParser.GenerateOsquerySQL(regChecks)

	// Create policy name (sanitized)
	policyName := g.sanitizePolicyName(fmt.Sprintf("stig-%s-%s", group.GroupID, group.RuleVersion))

	// Determine criticality based on severity
	critical := strings.EqualFold(group.RuleSeverity, string(types.SeverityHigh))

	// Create labels for better organization
	labels := map[string]string{
		"stig.group_id":     group.GroupID,
		"stig.rule_version": group.RuleVersion,
		"stig.severity":     strings.ToLower(group.RuleSeverity),
		"stig.rule_id":      group.RuleID,
		"compliance.type":   "stig",
		"compliance.source": "disa",
	}

	// Create annotations with additional metadata
	annotations := map[string]string{
		"stig.rule_weight":    group.RuleWeight,
		"stig.rule_ident":     group.RuleIdent,
		"stig.check_system":   group.RuleCheckSystem,
		"stig.fix_id":         group.RuleFixID,
		"generated.timestamp": time.Now().UTC().Format(time.RFC3339),
		"generated.tool":      "stig-processor",
	}

	// Add primary registry check info to annotations (use first check)
	if len(regChecks) > 0 {
		primaryCheck := regChecks[0]
		annotations["registry.hive"] = primaryCheck.Hive
		annotations["registry.path"] = primaryCheck.Path
		annotations["registry.value_name"] = primaryCheck.ValueName
		annotations["registry.comparison"] = primaryCheck.Comparison

		if primaryCheck.ValueType != "" {
			annotations["registry.value_type"] = primaryCheck.ValueType
		}
		if primaryCheck.Value != "" {
			annotations["registry.expected_value"] = primaryCheck.Value
		}

		// If multiple checks, note that in annotations
		if len(regChecks) > 1 {
			annotations["registry.multiple_checks"] = fmt.Sprintf("%d", len(regChecks))
		}
	}

	// Build comprehensive description
	description := g.buildPolicyDescription(group, regChecks)

	policy := &types.FleetPolicy{
		APIVersion: types.FleetAPIVersion,
		Kind:       types.FleetKindPolicy,
		Metadata: types.PolicyMeta{
			Name:        policyName,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: types.PolicySpec{
			Name:        fmt.Sprintf("STIG %s: %s", group.GroupID, group.RuleTitle),
			Query:       query,
			Description: description,
			Resolution:  g.buildResolutionText(group),
			Platform:    types.PlatformWindows,
			Critical:    critical,
		},
	}

	// Validate the generated policy
	if err := g.validatePolicy(policy); err != nil {
		return nil, fmt.Errorf("generated policy failed validation: %w", err)
	}

	return policy, nil
}

// buildPolicyDescription creates a comprehensive description for the policy
func (g *FleetPolicyGenerator) buildPolicyDescription(group *types.STIGGroup, regChecks []*types.RegistryCheck) string {
	var desc strings.Builder

	// Header with basic information
	desc.WriteString(fmt.Sprintf("STIG Rule %s (Severity: %s)\n\n", group.GroupID, group.RuleSeverity))

	// Rule title and vulnerability discussion
	if group.RuleVulnDiscussion != "" {
		desc.WriteString("Vulnerability Discussion:\n")
		desc.WriteString(g.formatTextBlock(group.RuleVulnDiscussion))
		desc.WriteString("\n\n")
	}

	// Check content
	desc.WriteString("Check Content:\n")
	desc.WriteString(g.formatTextBlock(group.RuleCheckContent))
	desc.WriteString("\n\n")

	// Registry check details
	if len(regChecks) == 1 {
		desc.WriteString("Registry Check Details:\n")
		regCheck := regChecks[0]
		desc.WriteString(fmt.Sprintf("- Hive: %s\n", regCheck.Hive))
		desc.WriteString(fmt.Sprintf("- Path: \\%s\\\n", regCheck.Path))
		desc.WriteString(fmt.Sprintf("- Value Name: %s\n", regCheck.ValueName))
		if regCheck.ValueType != "" {
			desc.WriteString(fmt.Sprintf("- Value Type: %s\n", regCheck.ValueType))
		}
		if regCheck.Value != "" {
			desc.WriteString(fmt.Sprintf("- Expected Value: %s\n", regCheck.Value))
		}
		if regCheck.Comparison != "equals" {
			desc.WriteString(fmt.Sprintf("- Comparison: %s\n", regCheck.Comparison))
		}
	} else {
		desc.WriteString(fmt.Sprintf("Registry Checks (%d total):\n", len(regChecks)))
		for i, regCheck := range regChecks {
			desc.WriteString(fmt.Sprintf("Check %d:\n", i+1))
			desc.WriteString(fmt.Sprintf("  - Hive: %s\n", regCheck.Hive))
			desc.WriteString(fmt.Sprintf("  - Path: \\%s\\\n", regCheck.Path))
			desc.WriteString(fmt.Sprintf("  - Value Name: %s\n", regCheck.ValueName))
			if regCheck.ValueType != "" {
				desc.WriteString(fmt.Sprintf("  - Value Type: %s\n", regCheck.ValueType))
			}
			if regCheck.Value != "" {
				desc.WriteString(fmt.Sprintf("  - Expected Value: %s\n", regCheck.Value))
			}
			if regCheck.Comparison != "equals" {
				desc.WriteString(fmt.Sprintf("  - Comparison: %s\n", regCheck.Comparison))
			}
			desc.WriteString("\n")
		}
	}

	// Additional metadata
	if group.RuleIdent != "" {
		desc.WriteString(fmt.Sprintf("\nCCI: %s\n", group.RuleIdent))
	}

	// Mitigation information if available
	if group.RuleMitigations != "" {
		desc.WriteString("\nMitigations:\n")
		desc.WriteString(g.formatTextBlock(group.RuleMitigations))
	}

	return desc.String()
}

// buildResolutionText creates resolution instructions
func (g *FleetPolicyGenerator) buildResolutionText(group *types.STIGGroup) string {
	if group.RuleFixText != "" {
		return g.formatTextBlock(group.RuleFixText)
	}
	return "Refer to STIG documentation for remediation steps."
}

// formatTextBlock formats text for better readability
func (g *FleetPolicyGenerator) formatTextBlock(text string) string {
	// Clean up the text
	text = strings.TrimSpace(text)

	// Replace multiple consecutive whitespace with single space
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	// Add proper line breaks for readability
	text = strings.ReplaceAll(text, ". ", ".\n")

	return text
}

// sanitizePolicyName creates a valid policy name from the input
func (g *FleetPolicyGenerator) sanitizePolicyName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace invalid characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	name = reg.ReplaceAllString(name, "-")

	// Remove multiple consecutive hyphens
	reg2 := regexp.MustCompile(`-+`)
	name = reg2.ReplaceAllString(name, "-")

	// Trim hyphens from start/end
	name = strings.Trim(name, "-")

	// Ensure name is not empty and not too long
	if name == "" {
		name = "stig-policy"
	}
	if len(name) > 253 { // Kubernetes name limit
		name = name[:253]
		name = strings.TrimSuffix(name, "-")
	}

	return name
}

// validatePolicy performs validation on a generated Fleet policy
func (g *FleetPolicyGenerator) validatePolicy(policy *types.FleetPolicy) error {
	if policy == nil {
		return fmt.Errorf("policy cannot be nil")
	}

	// Validate required fields
	if policy.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if policy.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if policy.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if policy.Spec.Name == "" {
		return fmt.Errorf("spec.name is required")
	}
	if policy.Spec.Query == "" {
		return fmt.Errorf("spec.query is required")
	}

	// Validate policy name format
	nameRegex := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	if !nameRegex.MatchString(policy.Metadata.Name) {
		return fmt.Errorf("invalid policy name format: %s", policy.Metadata.Name)
	}

	// Validate SQL query (basic checks)
	if err := g.validateOsquerySQL(policy.Spec.Query); err != nil {
		return fmt.Errorf("invalid osquery SQL: %w", err)
	}

	return nil
}

// validateOsquerySQL performs basic validation on osquery SQL
func (g *FleetPolicyGenerator) validateOsquerySQL(query string) error {
	query = strings.TrimSpace(strings.ToLower(query))

	// Must start with SELECT
	if !strings.HasPrefix(query, "select") {
		return fmt.Errorf("query must start with SELECT")
	}

	// Must contain FROM registry
	if !strings.Contains(query, "from registry") {
		return fmt.Errorf("query must select from registry table")
	}

	// Must contain WHERE clause
	if !strings.Contains(query, "where") {
		return fmt.Errorf("query must contain WHERE clause")
	}

	// Check for basic SQL injection patterns (but allow trailing semicolons)
	dangerousPatterns := []string{
		"--", "/*", "*/", "xp_", "sp_", "drop", "delete", "update", "insert",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(query, pattern) {
			return fmt.Errorf("query contains potentially dangerous pattern: %s", pattern)
		}
	}

	// Check for dangerous semicolons (not at the end)
	if strings.Contains(query, ";") && !strings.HasSuffix(strings.TrimSpace(query), ";") {
		return fmt.Errorf("query contains potentially dangerous pattern: ; (not at end)")
	}

	return nil
}

// WritePolicy writes a policy to a file in the specified format
func (g *FleetPolicyGenerator) WritePolicy(policy *types.FleetPolicy, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var data []byte
	var err error
	var filename string

	switch g.options.Format {
	case "json":
		if g.options.Pretty {
			data, err = json.MarshalIndent(policy, "", "  ")
		} else {
			data, err = json.Marshal(policy)
		}
		filename = fmt.Sprintf("%s.json", policy.Metadata.Name)
	default: // yaml
		data, err = yaml.Marshal(policy)
		filename = fmt.Sprintf("%s.yaml", policy.Metadata.Name)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	filepath := filepath.Join(outputDir, filename)
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write policy file %s: %w", filepath, err)
	}

	if g.options.Verbose {
		fmt.Printf("Written policy: %s\n", filepath)
	}

	return nil
}

// WriteSummary writes a processing summary file
func (g *FleetPolicyGenerator) WriteSummary(result *types.ProcessingResult, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create summary data
	summary := &types.ProcessingSummary{
		TotalRules:        result.Total,
		Automatable:       result.Automatable,
		ManualReview:      result.ManualReview,
		PoliciesGenerated: len(result.Policies),
		ProcessingTime:    result.Duration.String(),
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
		Policies:          make([]types.PolicySummaryItem, 0, len(result.Policies)),
		Errors:            result.Errors,
	}

	// Sort policies by name for consistent output
	sort.Slice(result.Policies, func(i, j int) bool {
		return result.Policies[i].Metadata.Name < result.Policies[j].Metadata.Name
	})

	// Add policy summary items
	for _, policy := range result.Policies {
		item := types.PolicySummaryItem{
			Name:     policy.Metadata.Name,
			Title:    policy.Spec.Name,
			Platform: policy.Spec.Platform,
			Critical: policy.Spec.Critical,
		}

		// Extract metadata from labels/annotations
		if severity, exists := policy.Metadata.Labels["stig.severity"]; exists {
			item.Severity = severity
		}
		if groupID, exists := policy.Metadata.Labels["stig.group_id"]; exists {
			item.GroupID = groupID
		}
		if ruleVersion, exists := policy.Metadata.Labels["stig.rule_version"]; exists {
			item.RuleVersion = ruleVersion
		}

		summary.Policies = append(summary.Policies, item)
	}

	// Marshal and write summary
	var data []byte
	var err error
	var filename string

	switch g.options.Format {
	case "json":
		if g.options.Pretty {
			data, err = json.MarshalIndent(summary, "", "  ")
		} else {
			data, err = json.Marshal(summary)
		}
		filename = "stig-summary.json"
	default: // yaml
		data, err = yaml.Marshal(summary)
		filename = "stig-summary.yaml"
	}

	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	filepath := filepath.Join(outputDir, filename)
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write summary file %s: %w", filepath, err)
	}

	if g.options.Verbose {
		fmt.Printf("Written summary: %s\n", filepath)
	}

	return nil
}

// BatchGenerate generates multiple policies from a list of STIG groups
func (g *FleetPolicyGenerator) BatchGenerate(groups []types.STIGGroup) *types.ProcessingResult {
	start := time.Now()

	result := &types.ProcessingResult{
		Total:    len(groups),
		Policies: make([]types.FleetPolicy, 0),
		Errors:   make([]types.ProcessingError, 0),
	}

	for _, group := range groups {
		// Filter by severity if specified
		if g.options.Severity != "" && !strings.EqualFold(group.RuleSeverity, g.options.Severity) {
			continue
		}

		// Try to parse as registry check
		regChecks, automatable := g.regParser.ParseRegistryCheck(group.RuleCheckContent)
		if automatable {
			result.Automatable++

			// Generate policy
			policy, err := g.GeneratePolicy(&group, regChecks)
			if err != nil {
				result.Errors = append(result.Errors, types.ProcessingError{
					GroupID:   group.GroupID,
					RuleID:    group.RuleID,
					Message:   err.Error(),
					Type:      types.ErrorTypeValidationFailed,
					Timestamp: time.Now(),
				})
				continue
			}

			result.Policies = append(result.Policies, *policy)

			if g.options.Verbose {
				fmt.Printf("[AUTOMATABLE] %s: %s\n", group.GroupID, group.RuleTitle)
			}

			// Write individual policy file if not dry run
			if !g.options.DryRun {
				if err := g.WritePolicy(policy, g.options.OutputDir); err != nil {
					result.Errors = append(result.Errors, types.ProcessingError{
						GroupID:   group.GroupID,
						RuleID:    group.RuleID,
						Message:   fmt.Sprintf("failed to write policy: %v", err),
						Type:      types.ErrorTypeFileWriteFailed,
						Timestamp: time.Now(),
					})
				}
			}
		} else {
			result.ManualReview++
			if g.options.Verbose {
				fmt.Printf("[MANUAL] %s: %s\n", group.GroupID, group.RuleTitle)
			}
		}
	}

	result.Duration = time.Since(start)

	// Write summary file if not dry run
	if !g.options.DryRun {
		if err := g.WriteSummary(result, g.options.OutputDir); err != nil {
			result.Errors = append(result.Errors, types.ProcessingError{
				Message:   fmt.Sprintf("failed to write summary: %v", err),
				Type:      types.ErrorTypeFileWriteFailed,
				Timestamp: time.Now(),
			})
		}
	}

	return result
}
