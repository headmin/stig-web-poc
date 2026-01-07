package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// STIG data structures
type STIGBenchmark struct {
	ID          int         `json:"id"`
	BenchmarkID string      `json:"benchmarkId"`
	Slug        string      `json:"slug"`
	Status      string      `json:"status"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Version     string      `json:"version"`
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
}

// Fleet policy structures
type FleetPolicy struct {
	APIVersion string     `yaml:"apiVersion"`
	Kind       string     `yaml:"kind"`
	Metadata   Metadata   `yaml:"metadata"`
	Spec       PolicySpec `yaml:"spec"`
}

type Metadata struct {
	Name string `yaml:"name"`
}

type PolicySpec struct {
	Name        string `yaml:"name"`
	Query       string `yaml:"query"`
	Description string `yaml:"description"`
	Resolution  string `yaml:"resolution"`
	Platform    string `yaml:"platform"`
	Critical    bool   `yaml:"critical"`
}

// Registry check structure
type RegistryCheck struct {
	Hive      string
	Path      string
	ValueName string
	ValueType string
	Value     string
}

// Processor result
type ProcessingResult struct {
	Automatable  int
	ManualReview int
	Total        int
	Policies     []FleetPolicy
}

func main() {
	var (
		inputFile = flag.String("input", "microsoft-windows-11-security-technical-implementation-guide.json", "Input STIG JSON file")
		outputDir = flag.String("output", "output", "Output directory for Fleet policies")
		format    = flag.String("format", "yaml", "Output format (yaml, json)")
		severity  = flag.String("severity", "", "Filter by severity (low, medium, high)")
		verbose   = flag.Bool("verbose", false, "Enable verbose logging")
		dryRun    = flag.Bool("dry-run", false, "Dry run - don't write files")
	)
	flag.Parse()

	processor := &STIGProcessor{
		InputFile: *inputFile,
		OutputDir: *outputDir,
		Format:    *format,
		Severity:  *severity,
		Verbose:   *verbose,
		DryRun:    *dryRun,
	}

	result, err := processor.Process()
	if err != nil {
		log.Fatalf("Error processing STIG: %v", err)
	}

	fmt.Printf("STIG Processing Complete!\n")
	fmt.Printf("Total rules: %d\n", result.Total)
	fmt.Printf("Automatable: %d\n", result.Automatable)
	fmt.Printf("Manual review: %d\n", result.ManualReview)
	fmt.Printf("Fleet policies generated: %d\n", len(result.Policies))

	if !*dryRun {
		fmt.Printf("Output written to: %s\n", *outputDir)
	}
}

type STIGProcessor struct {
	InputFile string
	OutputDir string
	Format    string
	Severity  string
	Verbose   bool
	DryRun    bool
}

func (p *STIGProcessor) Process() (*ProcessingResult, error) {
	// Load STIG data
	stig, err := p.loadSTIGData()
	if err != nil {
		return nil, fmt.Errorf("failed to load STIG data: %v", err)
	}

	result := &ProcessingResult{
		Total:    len(stig.Groups),
		Policies: make([]FleetPolicy, 0),
	}

	// Create output directory
	if !p.DryRun {
		if err := os.MkdirAll(p.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %v", err)
		}
	}

	// Process each rule
	for _, group := range stig.Groups {
		// Filter by severity if specified
		if p.Severity != "" && !strings.EqualFold(group.RuleSeverity, p.Severity) {
			continue
		}

		regCheck, automatable := p.parseRegistryCheck(group.RuleCheckContent)
		if automatable {
			result.Automatable++
			policy := p.createFleetPolicy(group, regCheck)
			result.Policies = append(result.Policies, policy)

			if p.Verbose {
				fmt.Printf("[AUTOMATABLE] %s: %s\n", group.GroupID, group.RuleTitle)
			}

			// Write individual policy file
			if !p.DryRun {
				if err := p.writePolicyFile(policy); err != nil {
					log.Printf("Warning: failed to write policy %s: %v", policy.Metadata.Name, err)
				}
			}
		} else {
			result.ManualReview++
			if p.Verbose {
				fmt.Printf("[MANUAL] %s: %s\n", group.GroupID, group.RuleTitle)
			}
		}
	}

	// Write summary file
	if !p.DryRun {
		if err := p.writeSummaryFile(result); err != nil {
			log.Printf("Warning: failed to write summary: %v", err)
		}
	}

	return result, nil
}

func (p *STIGProcessor) loadSTIGData() (*STIGBenchmark, error) {
	data, err := os.ReadFile(p.InputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var stig STIGBenchmark
	if err := json.Unmarshal(data, &stig); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &stig, nil
}

func (p *STIGProcessor) parseRegistryCheck(checkContent string) (*RegistryCheck, bool) {
	// Regex patterns for registry check components
	hiveRegex := regexp.MustCompile(`Registry Hive:\s*(HKEY_[A-Z_]+)`)
	pathRegex := regexp.MustCompile(`Registry Path:\s*\\(.+?)\\?\s*\n`)
	valueNameRegex := regexp.MustCompile(`Value Name:\s*(.+?)\s*\n`)
	valueTypeRegex := regexp.MustCompile(`(?:Value Type|Type):\s*(REG_[A-Z]+)`)
	valueRegex := regexp.MustCompile(`Value:\s*(.+?)(?:\s*\n|$)`)

	// Check if this looks like a registry check
	if !strings.Contains(checkContent, "Registry Hive:") {
		return nil, false
	}

	regCheck := &RegistryCheck{}

	// Extract hive
	if match := hiveRegex.FindStringSubmatch(checkContent); match != nil {
		regCheck.Hive = match[1]
	} else {
		return nil, false
	}

	// Extract path
	if match := pathRegex.FindStringSubmatch(checkContent); match != nil {
		regCheck.Path = strings.TrimSpace(match[1])
	} else {
		return nil, false
	}

	// Extract value name
	if match := valueNameRegex.FindStringSubmatch(checkContent); match != nil {
		regCheck.ValueName = strings.TrimSpace(match[1])
	} else {
		return nil, false
	}

	// Extract value type (optional)
	if match := valueTypeRegex.FindStringSubmatch(checkContent); match != nil {
		regCheck.ValueType = match[1]
	}

	// Extract value (optional)
	if match := valueRegex.FindStringSubmatch(checkContent); match != nil {
		value := strings.TrimSpace(match[1])
		// Clean up common value formats
		value = strings.ReplaceAll(value, "0x", "")
		value = strings.TrimSpace(strings.Split(value, "(")[0]) // Remove parenthetical explanations
		regCheck.Value = value
	}

	// Must have at least hive, path, and value name to be automatable
	return regCheck, regCheck.Hive != "" && regCheck.Path != "" && regCheck.ValueName != ""
}

func (p *STIGProcessor) createFleetPolicy(group STIGGroup, regCheck *RegistryCheck) FleetPolicy {
	// Generate osquery SQL
	query := p.generateOsquerySQL(regCheck)

	// Create policy name (sanitized)
	policyName := p.sanitizePolicyName(fmt.Sprintf("stig-%s-%s", group.GroupID, group.RuleVersion))

	// Determine criticality based on severity
	critical := strings.EqualFold(group.RuleSeverity, "high")

	return FleetPolicy{
		APIVersion: "v1",
		Kind:       "policy",
		Metadata: Metadata{
			Name: policyName,
		},
		Spec: PolicySpec{
			Name:        fmt.Sprintf("STIG %s: %s", group.GroupID, group.RuleTitle),
			Query:       query,
			Description: fmt.Sprintf("STIG Rule %s (Severity: %s)\n\nCheck Content: %s\n\nFix: %s", group.GroupID, group.RuleSeverity, group.RuleCheckContent, group.RuleFixText),
			Resolution:  group.RuleFixText,
			Platform:    "windows",
			Critical:    critical,
		},
	}
}

func (p *STIGProcessor) generateOsquerySQL(regCheck *RegistryCheck) string {
	// Convert HKEY to osquery format
	var hive string
	switch regCheck.Hive {
	case "HKEY_LOCAL_MACHINE":
		hive = "HKEY_LOCAL_MACHINE"
	case "HKEY_CURRENT_USER":
		hive = "HKEY_CURRENT_USER"
	case "HKEY_USERS":
		hive = "HKEY_USERS"
	default:
		hive = regCheck.Hive
	}

	// Build registry path
	fullPath := hive + "\\" + regCheck.Path

	// Base query
	query := fmt.Sprintf("SELECT name, type, data FROM registry WHERE path = '%s' AND name = '%s'",
		strings.ReplaceAll(fullPath, "\\", "\\\\"), regCheck.ValueName)

	// Add value check if we have an expected value
	if regCheck.Value != "" {
		// Try to parse as number first
		if val, err := strconv.Atoi(regCheck.Value); err == nil {
			query += fmt.Sprintf(" AND CAST(data AS INTEGER) = %d", val)
		} else {
			query += fmt.Sprintf(" AND data = '%s'", regCheck.Value)
		}
	}

	return query
}

func (p *STIGProcessor) sanitizePolicyName(name string) string {
	// Replace invalid characters with hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-]`)
	sanitized := reg.ReplaceAllString(strings.ToLower(name), "-")

	// Remove multiple consecutive hyphens
	reg2 := regexp.MustCompile(`-+`)
	sanitized = reg2.ReplaceAllString(sanitized, "-")

	// Trim hyphens from start/end
	sanitized = strings.Trim(sanitized, "-")

	return sanitized
}

func (p *STIGProcessor) writePolicyFile(policy FleetPolicy) error {
	filename := fmt.Sprintf("%s.yaml", policy.Metadata.Name)
	filepath := filepath.Join(p.OutputDir, filename)

	var data []byte
	var err error

	switch p.Format {
	case "json":
		data, err = json.MarshalIndent(policy, "", "  ")
		filepath = strings.ReplaceAll(filepath, ".yaml", ".json")
	default: // yaml
		data, err = yaml.Marshal(policy)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal policy: %v", err)
	}

	return os.WriteFile(filepath, data, 0644)
}

func (p *STIGProcessor) writeSummaryFile(result *ProcessingResult) error {
	summary := map[string]interface{}{
		"total_rules":        result.Total,
		"automatable":        result.Automatable,
		"manual_review":      result.ManualReview,
		"policies_generated": len(result.Policies),
		"policies":           make([]map[string]string, 0, len(result.Policies)),
	}

	// Sort policies by name for consistent output
	sort.Slice(result.Policies, func(i, j int) bool {
		return result.Policies[i].Metadata.Name < result.Policies[j].Metadata.Name
	})

	for _, policy := range result.Policies {
		summary["policies"] = append(summary["policies"].([]map[string]string), map[string]string{
			"name":     policy.Metadata.Name,
			"title":    policy.Spec.Name,
			"platform": policy.Spec.Platform,
			"critical": fmt.Sprintf("%t", policy.Spec.Critical),
		})
	}

	var data []byte
	var err error
	var filename string

	switch p.Format {
	case "json":
		data, err = json.MarshalIndent(summary, "", "  ")
		filename = "stig-summary.json"
	default: // yaml
		data, err = yaml.Marshal(summary)
		filename = "stig-summary.yaml"
	}

	if err != nil {
		return fmt.Errorf("failed to marshal summary: %v", err)
	}

	filepath := filepath.Join(p.OutputDir, filename)
	return os.WriteFile(filepath, data, 0644)
}
