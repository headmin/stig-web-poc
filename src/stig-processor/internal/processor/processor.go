package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/stig-processor/pkg/generator"
	"github.com/stig-processor/pkg/parser"
	"github.com/stig-processor/pkg/types"
)

// STIGProcessor orchestrates the entire STIG processing workflow
type STIGProcessor struct {
	options   *types.ProcessingOptions
	parser    *parser.STIGParser
	generator *generator.FleetPolicyGenerator
	stats     *parser.Statistics
}

// NewSTIGProcessor creates a new STIG processor with the given options
func NewSTIGProcessor(options *types.ProcessingOptions) *STIGProcessor {
	if options == nil {
		options = &types.ProcessingOptions{
			InputFile: types.DefaultInputFile,
			OutputDir: types.DefaultOutputDir,
			Format:    types.DefaultOutputFormat,
			Timeout:   types.DefaultTimeout,
		}
	}

	// Set defaults for missing values
	if options.InputFile == "" {
		options.InputFile = types.DefaultInputFile
	}
	if options.OutputDir == "" {
		options.OutputDir = types.DefaultOutputDir
	}
	if options.Format == "" {
		options.Format = types.DefaultOutputFormat
	}
	if options.Timeout == 0 {
		options.Timeout = types.DefaultTimeout
	}

	stigParser := parser.NewSTIGParser(options.Verbose)

	return &STIGProcessor{
		options:   options,
		parser:    stigParser,
		generator: generator.NewFleetPolicyGenerator(options),
		stats:     parser.NewStatistics(stigParser),
	}
}

// Process executes the complete STIG processing workflow
func (sp *STIGProcessor) Process() (*types.ProcessingResult, error) {
	return sp.ProcessWithContext(context.Background())
}

// ProcessWithContext executes the STIG processing workflow with context for cancellation
func (sp *STIGProcessor) ProcessWithContext(ctx context.Context) (*types.ProcessingResult, error) {
	start := time.Now()

	if sp.options.Verbose {
		fmt.Printf("Starting STIG processing with options:\n")
		fmt.Printf("  Input: %s\n", sp.options.InputFile)
		fmt.Printf("  Output: %s\n", sp.options.OutputDir)
		fmt.Printf("  Format: %s\n", sp.options.Format)
		fmt.Printf("  Severity filter: %s\n", sp.options.Severity)
		fmt.Printf("  Dry run: %t\n", sp.options.DryRun)
		fmt.Printf("\n")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, sp.options.Timeout)
	defer cancel()

	// Phase 1: Validate inputs
	if err := sp.validateInputs(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Phase 2: Parse STIG file
	stig, err := sp.parseSTIGFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse STIG file: %w", err)
	}

	// Phase 3: Filter rules if severity is specified
	groups := sp.filterGroups(stig.Groups)

	// Phase 4: Generate Fleet policies
	result := sp.generatePolicies(ctx, groups)
	result.Duration = time.Since(start)

	// Phase 5: Post-process and finalize
	if err := sp.finalizeProcessing(result); err != nil {
		return result, fmt.Errorf("finalization failed: %w", err)
	}

	return result, nil
}

// validateInputs performs pre-flight validation of input parameters
func (sp *STIGProcessor) validateInputs() error {
	// Check if input file exists
	if _, err := os.Stat(sp.options.InputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", sp.options.InputFile)
	}

	// Validate output format
	if sp.options.Format != "yaml" && sp.options.Format != "json" {
		return fmt.Errorf("invalid output format: %s (must be 'yaml' or 'json')", sp.options.Format)
	}

	// Validate severity filter if provided
	if sp.options.Severity != "" {
		validSeverity := false
		for _, level := range types.ValidSeverityLevels {
			if strings.EqualFold(sp.options.Severity, string(level)) {
				validSeverity = true
				break
			}
		}
		if !validSeverity {
			return fmt.Errorf("invalid severity level: %s (must be one of: low, medium, high)", sp.options.Severity)
		}
	}

	// Create output directory if it doesn't exist and we're not in dry-run mode
	if !sp.options.DryRun {
		if err := os.MkdirAll(sp.options.OutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	return nil
}

// parseSTIGFile loads and parses the STIG JSON file
func (sp *STIGProcessor) parseSTIGFile(ctx context.Context) (*types.STIGBenchmark, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if sp.options.Verbose {
		fmt.Printf("Parsing STIG file: %s\n", sp.options.InputFile)
	}

	stig, err := sp.parser.ParseSTIGFile(sp.options.InputFile)
	if err != nil {
		return nil, err
	}

	if sp.options.Verbose {
		fmt.Printf("Successfully parsed STIG: %s v%s (%d rules)\n",
			stig.Title, stig.Version, len(stig.Groups))
	}

	return stig, nil
}

// filterGroups filters STIG groups based on severity if specified
func (sp *STIGProcessor) filterGroups(groups []types.STIGGroup) []types.STIGGroup {
	if sp.options.Severity == "" {
		return groups
	}

	filtered := make([]types.STIGGroup, 0)
	for _, group := range groups {
		if strings.EqualFold(group.RuleSeverity, sp.options.Severity) {
			filtered = append(filtered, group)
		}
	}

	if sp.options.Verbose {
		fmt.Printf("Filtered %d rules by severity '%s' (from %d total)\n",
			len(filtered), sp.options.Severity, len(groups))
	}

	return filtered
}

// generatePolicies generates Fleet policies from the filtered STIG groups
func (sp *STIGProcessor) generatePolicies(ctx context.Context, groups []types.STIGGroup) *types.ProcessingResult {
	if sp.options.Verbose {
		fmt.Printf("Generating Fleet policies from %d rules...\n", len(groups))
	}

	// Use the generator's batch processing
	result := sp.generator.BatchGenerate(groups)

	// Check for context cancellation after processing
	select {
	case <-ctx.Done():
		// Add cancellation error to result
		result.Errors = append(result.Errors, types.ProcessingError{
			Message:   "processing was cancelled",
			Type:      types.ErrorTypeUnknown,
			Timestamp: time.Now(),
		})
	default:
	}

	return result
}

// finalizeProcessing performs final cleanup and validation
func (sp *STIGProcessor) finalizeProcessing(result *types.ProcessingResult) error {
	if sp.options.Verbose {
		fmt.Printf("\nProcessing complete:\n")
		fmt.Printf("  Total rules processed: %d\n", result.Total)
		fmt.Printf("  Automatable rules: %d\n", result.Automatable)
		fmt.Printf("  Manual review required: %d\n", result.ManualReview)
		fmt.Printf("  Fleet policies generated: %d\n", len(result.Policies))
		fmt.Printf("  Processing time: %v\n", result.Duration)

		if len(result.Errors) > 0 {
			fmt.Printf("  Errors encountered: %d\n", len(result.Errors))
		}
	}

	// Validate that we generated some policies if we had automatable rules
	if result.Automatable > 0 && len(result.Policies) == 0 {
		return fmt.Errorf("expected to generate policies but none were created")
	}

	// Check for critical errors that should fail the process
	criticalErrors := sp.filterCriticalErrors(result.Errors)
	if len(criticalErrors) > 0 {
		return fmt.Errorf("processing failed due to %d critical errors", len(criticalErrors))
	}

	return nil
}

// filterCriticalErrors identifies errors that should cause the process to fail
func (sp *STIGProcessor) filterCriticalErrors(errors []types.ProcessingError) []types.ProcessingError {
	critical := make([]types.ProcessingError, 0)

	for _, err := range errors {
		// File write errors during non-dry-run are critical
		if err.Type == types.ErrorTypeFileWriteFailed && !sp.options.DryRun {
			critical = append(critical, err)
		}
		// Add other critical error conditions as needed
	}

	return critical
}

// GetStatistics analyzes the STIG file and returns processing statistics
func (sp *STIGProcessor) GetStatistics() (*types.ProcessingStatistics, error) {
	stig, err := sp.parser.ParseSTIGFile(sp.options.InputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse STIG file for statistics: %w", err)
	}

	return sp.stats.AnalyzeSTIG(stig), nil
}

// ValidatePolicies validates existing Fleet policy files in the output directory
func (sp *STIGProcessor) ValidatePolicies() (*types.ValidationResult, error) {
	return sp.validatePolicyFiles(sp.options.OutputDir)
}

// validatePolicyFiles validates all policy files in the given directory
func (sp *STIGProcessor) validatePolicyFiles(dir string) (*types.ValidationResult, error) {
	result := &types.ValidationResult{
		Valid:  true,
		Count:  0,
		Errors: make([]types.ValidationError, 0),
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return result, nil // No files to validate
	}

	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	// Validate each policy file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasPrefix(filename, "stig-") {
			continue // Not a STIG policy file
		}

		if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			if err := sp.validateYAMLFile(dir, filename); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, types.ValidationError{
					FilePath: filename,
					Message:  err.Error(),
					Type:     types.ValidationErrorYAMLSyntax,
				})
			}
			result.Count++
		} else if strings.HasSuffix(filename, ".json") {
			if err := sp.validateJSONFile(dir, filename); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, types.ValidationError{
					FilePath: filename,
					Message:  err.Error(),
					Type:     types.ValidationErrorJSONSyntax,
				})
			}
			result.Count++
		}
	}

	return result, nil
}

// validateYAMLFile validates a single YAML policy file
func (sp *STIGProcessor) validateYAMLFile(dir, filename string) error {
	filepath := fmt.Sprintf("%s/%s", dir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var policy types.FleetPolicy
	if err := yaml.Unmarshal(data, &policy); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	// Basic Fleet policy validation
	return sp.validateFleetPolicyStructure(&policy)
}

// validateJSONFile validates a single JSON policy file
func (sp *STIGProcessor) validateJSONFile(dir, filename string) error {
	filepath := fmt.Sprintf("%s/%s", dir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var policy types.FleetPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return fmt.Errorf("invalid JSON syntax: %w", err)
	}

	// Basic Fleet policy validation
	return sp.validateFleetPolicyStructure(&policy)
}

// validateFleetPolicyStructure validates the structure of a Fleet policy
func (sp *STIGProcessor) validateFleetPolicyStructure(policy *types.FleetPolicy) error {
	if policy.APIVersion != types.FleetAPIVersion {
		return fmt.Errorf("invalid apiVersion: expected %s, got %s", types.FleetAPIVersion, policy.APIVersion)
	}

	if policy.Kind != types.FleetKindPolicy {
		return fmt.Errorf("invalid kind: expected %s, got %s", types.FleetKindPolicy, policy.Kind)
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

	if policy.Spec.Platform == "" {
		return fmt.Errorf("spec.platform is required")
	}

	return nil
}

// ProcessingOptions returns the current processing options
func (sp *STIGProcessor) ProcessingOptions() *types.ProcessingOptions {
	return sp.options
}

// UpdateOptions updates the processing options
func (sp *STIGProcessor) UpdateOptions(options *types.ProcessingOptions) {
	if options != nil {
		sp.options = options
		sp.generator = generator.NewFleetPolicyGenerator(options)
	}
}
