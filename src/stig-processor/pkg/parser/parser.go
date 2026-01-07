package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/stig-processor/pkg/types"
)

// STIGParser handles parsing of STIG JSON files and extracting registry checks
type STIGParser struct {
	verbose bool
}

// NewSTIGParser creates a new STIG parser instance
func NewSTIGParser(verbose bool) *STIGParser {
	return &STIGParser{
		verbose: verbose,
	}
}

// ParseSTIGFile loads and parses a STIG JSON file
func (p *STIGParser) ParseSTIGFile(filePath string) (*types.STIGBenchmark, error) {
	// Check if file exists and get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to access file %s: %w", filePath, err)
	}

	// Check file size (prevent loading extremely large files)
	if fileInfo.Size() > types.DefaultMaxFileSize {
		return nil, fmt.Errorf("file %s is too large (%d bytes), maximum allowed is %d bytes",
			filePath, fileInfo.Size(), types.DefaultMaxFileSize)
	}

	if p.verbose {
		fmt.Printf("Loading STIG file: %s (%d bytes)\n", filePath, fileInfo.Size())
	}

	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse JSON
	var stig types.STIGBenchmark
	if err := json.Unmarshal(data, &stig); err != nil {
		return nil, fmt.Errorf("failed to parse STIG JSON: %w", err)
	}

	if p.verbose {
		fmt.Printf("Parsed STIG: %s v%s with %d rules\n", stig.Title, stig.Version, len(stig.Groups))
	}

	return &stig, nil
}

// RegistryParser handles parsing Windows registry check patterns
type RegistryParser struct {
	verbose bool
	// Compiled regex patterns for better performance
	hiveRegex      *regexp.Regexp
	pathRegex      *regexp.Regexp
	valueNameRegex *regexp.Regexp
	valueTypeRegex *regexp.Regexp
	valueRegex     *regexp.Regexp
}

// NewRegistryParser creates a new registry parser with compiled regex patterns
func NewRegistryParser(verbose bool) *RegistryParser {
	return &RegistryParser{
		verbose:        verbose,
		hiveRegex:      regexp.MustCompile(`Registry Hive:\s*(HKEY_[A-Z_]+)`),
		pathRegex:      regexp.MustCompile(`Registry Path:\s*\\(.+?)\\?\s*(?:\n|$)`),
		valueNameRegex: regexp.MustCompile(`Value Name:\s*(.+?)\s*(?:\n|$)`),
		valueTypeRegex: regexp.MustCompile(`(?:Value Type|Type):\s*(REG_[A-Z_]+)`),
		valueRegex:     regexp.MustCompile(`Value:\s*(.+?)(?:\s*(?:\n|$))`),
	}
}

// ParseRegistryCheck extracts registry check information from STIG rule check content
func (rp *RegistryParser) ParseRegistryCheck(checkContent string) ([]*types.RegistryCheck, bool) {
	// Quick check - if it doesn't mention registry, it's not a registry check
	if !strings.Contains(checkContent, "Registry Hive:") {
		return nil, false
	}

	var checks []*types.RegistryCheck

	// Find all hives in the content (there might be multiple checks)
	hiveMatches := rp.hiveRegex.FindAllStringSubmatch(checkContent, -1)
	hivePositions := rp.hiveRegex.FindAllStringSubmatchIndex(checkContent, -1)

	if len(hiveMatches) == 0 {
		return nil, false
	}

	for i, hiveMatch := range hiveMatches {
		// Get the section for this registry check
		start := hivePositions[i][0]
		end := len(checkContent)
		if i+1 < len(hivePositions) {
			end = hivePositions[i+1][0]
		}
		section := checkContent[start:end]

		hive := strings.TrimSpace(hiveMatch[1])

		// Extract registry path
		pathMatch := rp.pathRegex.FindStringSubmatch(section)
		if pathMatch == nil {
			if rp.verbose {
				fmt.Printf("  No valid registry path found for hive %s\n", hive)
			}
			continue
		}

		// Extract value name
		nameMatch := rp.valueNameRegex.FindStringSubmatch(section)
		if nameMatch == nil {
			if rp.verbose {
				fmt.Printf("  No valid value name found for hive %s\n", hive)
			}
			continue
		}

		// Clean up the path
		path := strings.TrimSpace(pathMatch[1])
		path = strings.ReplaceAll(path, "\\\\", "\\")
		path = strings.TrimSuffix(path, "\\")

		valueName := strings.TrimSpace(nameMatch[1])

		// Extract value type (optional)
		valueType := "REG_DWORD" // default
		if typeMatch := rp.valueTypeRegex.FindStringSubmatch(section); typeMatch != nil {
			valueType = strings.TrimSpace(typeMatch[1])
		}

		// Extract expected value (optional)
		expectedValue := ""
		if valueMatch := rp.valueRegex.FindStringSubmatch(section); valueMatch != nil {
			expectedValue = rp.cleanRegistryValue(strings.TrimSpace(valueMatch[1]))
		}

		// Determine comparison operator
		comparison := rp.determineComparison(section, expectedValue)

		// Validate registry hive
		if !rp.isValidRegistryHive(hive) {
			if rp.verbose {
				fmt.Printf("  Invalid registry hive: %s\n", hive)
			}
			continue
		}

		regCheck := &types.RegistryCheck{
			Hive:       hive,
			Path:       path,
			ValueName:  valueName,
			ValueType:  valueType,
			Value:      expectedValue,
			Comparison: comparison,
		}

		checks = append(checks, regCheck)

		if rp.verbose {
			fmt.Printf("  Found registry check: %s\\%s\\%s = %s (%s)\n",
				hive, path, valueName, expectedValue, comparison)
		}
	}

	return checks, len(checks) > 0
}

// cleanRegistryValue cleans and normalizes registry values
func (rp *RegistryParser) cleanRegistryValue(value string) string {
	// Remove common prefixes and suffixes
	value = strings.TrimSpace(value)

	// Handle hexadecimal values (0x prefix)
	if strings.HasPrefix(value, "0x") {
		value = strings.TrimPrefix(value, "0x")
	}

	// Remove parenthetical explanations like "0x00000001 (1)"
	if parenIndex := strings.Index(value, "("); parenIndex != -1 {
		beforeParen := strings.TrimSpace(value[:parenIndex])
		if beforeParen != "" {
			value = beforeParen
		}
	}

	// Handle "or greater" type specifications
	if strings.Contains(strings.ToLower(value), "or greater") {
		parts := strings.Fields(value)
		if len(parts) > 0 {
			value = parts[0] // Take just the numeric part
		}
	}

	// Handle "or less" type specifications
	if strings.Contains(strings.ToLower(value), "or less") {
		parts := strings.Fields(value)
		if len(parts) > 0 {
			value = parts[0] // Take just the numeric part
		}
	}

	return strings.TrimSpace(value)
}

// isValidRegistryHive checks if the registry hive is one of the known valid hives
func (rp *RegistryParser) isValidRegistryHive(hive string) bool {
	validHives := []string{
		types.HKeyLocalMachine,
		types.HKeyCurrentUser,
		types.HKeyUsers,
		types.HKeyClassesRoot,
		types.HKeyCurrentConfig,
	}

	for _, validHive := range validHives {
		if hive == validHive {
			return true
		}
	}
	return false
}

// GenerateOsquerySQL converts multiple registry checks into osquery SQL - MATCHES PYTHON LOGIC
func (rp *RegistryParser) GenerateOsquerySQL(regChecks []*types.RegistryCheck) string {
	if len(regChecks) == 0 {
		return ""
	}

	var conditions []string

	for _, check := range regChecks {
		// Build full path like Python version (includes value name in path)
		fullPath := fmt.Sprintf("%s\\%s\\%s", check.Hive, check.Path, check.ValueName)

		switch check.Comparison {
		case "not_exists":
			// Check that the key doesn't exist
			conditions = append(conditions, fmt.Sprintf("NOT EXISTS (SELECT 1 FROM registry WHERE path = '%s')", fullPath))
		case "must_exist":
			conditions = append(conditions, fmt.Sprintf("path = '%s'", fullPath))
		case "greater_equal":
			conditions = append(conditions, fmt.Sprintf("(path = '%s' AND CAST(data AS INTEGER) >= %s)", fullPath, check.Value))
		case "less_equal":
			conditions = append(conditions, fmt.Sprintf("(path = '%s' AND CAST(data AS INTEGER) <= %s)", fullPath, check.Value))
		default:
			// Handle different value types like Python version
			if check.ValueType == types.RegSZ || check.ValueType == types.RegExpandSZ {
				// String type - check if value exists and is not empty
				if len(check.Value) > 50 {
					// Long string - just check it exists and is not empty
					conditions = append(conditions, fmt.Sprintf("(path = '%s' AND data != '' AND LENGTH(data) > 0)", fullPath))
				} else if check.Value != "" {
					// Short string - exact match, escape single quotes
					escapedValue := strings.ReplaceAll(check.Value, "'", "''")
					conditions = append(conditions, fmt.Sprintf("(path = '%s' AND data = '%s')", fullPath, escapedValue))
				} else {
					// No expected value - just check exists
					conditions = append(conditions, fmt.Sprintf("(path = '%s' AND data IS NOT NULL)", fullPath))
				}
			} else if check.ValueType == types.RegMultiSZ {
				// Multi-string - check exists and not empty
				conditions = append(conditions, fmt.Sprintf("(path = '%s' AND data != '' AND LENGTH(data) > 0)", fullPath))
			} else {
				// Default equals comparison (REG_DWORD, REG_QWORD, etc.)
				if rp.isNumericValue(check.Value, check.ValueType) {
					conditions = append(conditions, fmt.Sprintf("(path = '%s' AND data = '%s')", fullPath, check.Value))
				} else {
					// Escape single quotes for safety
					escapedValue := strings.ReplaceAll(check.Value, "'", "''")
					conditions = append(conditions, fmt.Sprintf("(path = '%s' AND data = '%s')", fullPath, escapedValue))
				}
			}
		}
	}

	// Build final query like Python version
	if len(conditions) == 1 {
		if strings.Contains(conditions[0], "NOT EXISTS") {
			return fmt.Sprintf("SELECT 1 WHERE %s;", conditions[0])
		}
		return fmt.Sprintf("SELECT 1 FROM registry WHERE %s;", conditions[0])
	}

	// Multiple checks - need to verify all pass
	return fmt.Sprintf("SELECT 1 FROM registry WHERE %s;", strings.Join(conditions, " AND "))
}

// isNumericValue determines if a registry value should be treated as numeric
func (rp *RegistryParser) isNumericValue(value, valueType string) bool {
	// REG_DWORD and REG_QWORD are always numeric
	if valueType == types.RegDWord || valueType == types.RegQWord {
		return true
	}

	// Try to parse as integer
	_, err := strconv.ParseInt(value, 0, 64)
	return err == nil
}

// parseNumericValue parses a string value as a numeric value
func (rp *RegistryParser) parseNumericValue(value string) (int64, error) {
	// Handle hexadecimal values
	if strings.HasPrefix(strings.ToLower(value), "0x") {
		return strconv.ParseInt(value, 0, 64)
	}

	// Try decimal first
	if num, err := strconv.ParseInt(value, 10, 64); err == nil {
		return num, nil
	}

	// Try hexadecimal without 0x prefix
	if num, err := strconv.ParseInt(value, 16, 64); err == nil {
		return num, nil
	}

	return 0, fmt.Errorf("cannot parse '%s' as numeric value", value)
}

// Statistics provides statistics about STIG parsing
type Statistics struct {
	parser *STIGParser
}

// NewStatistics creates a new statistics analyzer
func NewStatistics(parser *STIGParser) *Statistics {
	return &Statistics{parser: parser}
}

// AnalyzeSTIG provides comprehensive statistics about a STIG benchmark
func (s *Statistics) AnalyzeSTIG(stig *types.STIGBenchmark) *types.ProcessingStatistics {
	start := time.Now()

	stats := &types.ProcessingStatistics{
		Title:                stig.Title,
		Version:              stig.Version,
		TotalRules:           len(stig.Groups),
		SeverityDistribution: make(map[string]int),
	}

	regParser := NewRegistryParser(false) // Don't need verbose for stats

	for _, group := range stig.Groups {
		// Count by severity
		severity := strings.ToLower(group.RuleSeverity)
		stats.SeverityDistribution[severity]++

		// Categorize rule type
		if _, isRegistry := regParser.ParseRegistryCheck(group.RuleCheckContent); isRegistry {
			stats.RegistryRules++
		} else if s.isGroupPolicyRule(group.RuleCheckContent) {
			stats.GroupPolicyRules++
		} else {
			stats.ManualRules++
		}
	}

	stats.ProcessingTime = time.Since(start)
	return stats
}

// isGroupPolicyRule determines if a rule is related to Group Policy
func (s *Statistics) isGroupPolicyRule(checkContent string) bool {
	groupPolicyIndicators := []string{
		"Group Policy",
		"gpedit.msc",
		"Local Group Policy Editor",
		"Computer Configuration >> Administrative Templates",
		"User Configuration >> Administrative Templates",
		"gpresult",
		"Administrative Templates",
	}

	checkLower := strings.ToLower(checkContent)
	for _, indicator := range groupPolicyIndicators {
		if strings.Contains(checkLower, strings.ToLower(indicator)) {
			return true
		}
	}

	return false
}

// ValidateRegistryChecks performs additional validation on parsed registry checks
func (rp *RegistryParser) ValidateRegistryChecks(regChecks []*types.RegistryCheck) []types.ValidationError {
	var errors []types.ValidationError

	for i, regCheck := range regChecks {
		// Validate hive
		if !rp.isValidRegistryHive(regCheck.Hive) {
			errors = append(errors, types.ValidationError{
				Message: fmt.Sprintf("Check %d: Invalid registry hive: %s", i+1, regCheck.Hive),
				Type:    types.ValidationErrorFleetSchema,
			})
		}

		// Validate path format
		if strings.Contains(regCheck.Path, "..") {
			errors = append(errors, types.ValidationError{
				Message: fmt.Sprintf("Check %d: Registry path contains invalid '..' sequence", i+1),
				Type:    types.ValidationErrorFleetSchema,
			})
		}

		// Validate value name (cannot be empty)
		if strings.TrimSpace(regCheck.ValueName) == "" {
			errors = append(errors, types.ValidationError{
				Message: fmt.Sprintf("Check %d: Registry value name cannot be empty", i+1),
				Type:    types.ValidationErrorFleetSchema,
			})
		}

		// Validate value type if provided
		if regCheck.ValueType != "" && !rp.isValidValueType(regCheck.ValueType) {
			errors = append(errors, types.ValidationError{
				Message: fmt.Sprintf("Check %d: Invalid registry value type: %s", i+1, regCheck.ValueType),
				Type:    types.ValidationErrorFleetSchema,
			})
		}

		// Validate comparison operator
		validComparisons := []string{"equals", "greater_equal", "less_equal", "not_exists", "must_exist"}
		validComparison := false
		for _, valid := range validComparisons {
			if regCheck.Comparison == valid {
				validComparison = true
				break
			}
		}
		if !validComparison {
			errors = append(errors, types.ValidationError{
				Message: fmt.Sprintf("Check %d: Invalid comparison operator: %s", i+1, regCheck.Comparison),
				Type:    types.ValidationErrorFleetSchema,
			})
		}
	}

	return errors
}

// isValidValueType checks if the registry value type is valid
func (rp *RegistryParser) isValidValueType(valueType string) bool {
	validTypes := []string{
		types.RegSZ,
		types.RegExpandSZ,
		types.RegBinary,
		types.RegDWord,
		types.RegQWord,
		types.RegMultiSZ,
	}

	for _, validType := range validTypes {
		if valueType == validType {
			return true
		}
	}
	return false
}

// determineComparison determines the comparison operator from the check content
func (rp *RegistryParser) determineComparison(section, expectedValue string) string {
	sectionLower := strings.ToLower(section)

	// Check for comparison operators in the check content
	if strings.Contains(sectionLower, "or greater") || strings.Contains(section, ">=") {
		return "greater_equal"
	}
	if strings.Contains(sectionLower, "or fewer") || strings.Contains(sectionLower, "or less") || strings.Contains(section, "<=") {
		return "less_equal"
	}
	// Be more specific about "must not exist" - look for direct statements
	if strings.Contains(sectionLower, "must not exist") || strings.Contains(sectionLower, "should not exist") {
		return "not_exists"
	}
	// For "must_exist", look for more specific patterns that aren't just error messages
	if (strings.Contains(sectionLower, "must exist") || strings.Contains(sectionLower, "should exist")) &&
		!strings.Contains(sectionLower, "does not exist") {
		return "must_exist"
	}

	// Default to equals comparison
	return "equals"
}
