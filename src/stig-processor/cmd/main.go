package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stig-processor/internal/processor"
	"github.com/stig-processor/pkg/types"
)

var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	// Parse command line flags
	var (
		inputFile   = flag.String("input", types.DefaultInputFile, "Input STIG JSON file")
		outputDir   = flag.String("output", types.DefaultOutputDir, "Output directory for Fleet policies")
		format      = flag.String("format", types.DefaultOutputFormat, "Output format (yaml, json)")
		severity    = flag.String("severity", "", "Filter by severity (low, medium, high)")
		verbose     = flag.Bool("verbose", false, "Enable verbose logging")
		dryRun      = flag.Bool("dry-run", false, "Dry run - don't write files")
		pretty      = flag.Bool("pretty", false, "Pretty print JSON output")
		timeout     = flag.Duration("timeout", types.DefaultTimeout, "Processing timeout")
		showVersion = flag.Bool("version", false, "Show version information")
		showStats   = flag.Bool("stats", false, "Show STIG statistics only")
		validate    = flag.Bool("validate", false, "Validate existing policies only")
		help        = flag.Bool("help", false, "Show help message")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "STIG Processor - Convert DISA STIG rules to Fleet/osquery policies\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -input stig.json -output policies/\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -severity high -format json -verbose\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -stats -input stig.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -validate -output policies/\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -dry-run -verbose\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
	}

	flag.Parse()

	// Handle special flags
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nReceived interrupt signal, shutting down...")
		cancel()
	}()

	// Create processing options
	options := &types.ProcessingOptions{
		InputFile: *inputFile,
		OutputDir: *outputDir,
		Format:    *format,
		Severity:  *severity,
		Verbose:   *verbose,
		DryRun:    *dryRun,
		Pretty:    *pretty,
		Timeout:   *timeout,
	}

	// Create processor
	proc := processor.NewSTIGProcessor(options)

	// Handle different operation modes
	if *validate {
		if err := runValidation(proc); err != nil {
			log.Fatalf("Validation failed: %v", err)
		}
		return
	}

	if *showStats {
		if err := runStatistics(proc); err != nil {
			log.Fatalf("Statistics failed: %v", err)
		}
		return
	}

	// Run main processing
	if err := runProcessing(ctx, proc, options); err != nil {
		log.Fatalf("Processing failed: %v", err)
	}
}

// printVersion displays version information
func printVersion() {
	fmt.Printf("STIG Processor %s\n", version)
	fmt.Printf("Build time: %s\n", buildTime)
	fmt.Printf("Git commit: %s\n", gitCommit)
}

// runProcessing executes the main STIG processing workflow
func runProcessing(ctx context.Context, proc *processor.STIGProcessor, options *types.ProcessingOptions) error {
	if options.Verbose {
		fmt.Printf("🔍 STIG Processor v%s\n", version)
		fmt.Printf("Processing DISA STIG rules for Fleet/osquery compliance...\n\n")
	}

	// Start processing
	start := time.Now()
	result, err := proc.ProcessWithContext(ctx)
	if err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}

	// Display results
	printProcessingResults(result, options, time.Since(start))

	// Handle any non-critical errors
	if len(result.Errors) > 0 && options.Verbose {
		fmt.Printf("\n⚠️  Warnings and non-critical errors:\n")
		for _, err := range result.Errors {
			fmt.Printf("  • %s: %s\n", err.Type, err.Message)
			if err.GroupID != "" {
				fmt.Printf("    Rule: %s\n", err.GroupID)
			}
		}
	}

	return nil
}

// runValidation validates existing Fleet policy files
func runValidation(proc *processor.STIGProcessor) error {
	fmt.Printf("🔍 Validating Fleet Policies\n\n")

	validation, err := proc.ValidatePolicies()
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if validation.Valid {
		fmt.Printf("✅ All policies are valid!\n")
		fmt.Printf("Validated %d policy files\n", validation.Count)
	} else {
		fmt.Printf("❌ Validation failed!\n\n")
		for _, validationErr := range validation.Errors {
			fmt.Printf("  • %s: %s\n", validationErr.FilePath, validationErr.Message)
		}
		return fmt.Errorf("found %d validation errors", len(validation.Errors))
	}

	return nil
}

// runStatistics displays STIG file statistics
func runStatistics(proc *processor.STIGProcessor) error {
	fmt.Printf("📊 STIG Statistics\n\n")

	stats, err := proc.GetStatistics()
	if err != nil {
		return fmt.Errorf("statistics error: %w", err)
	}

	printStatistics(stats)
	return nil
}

// printProcessingResults displays the processing results in a formatted way
func printProcessingResults(result *types.ProcessingResult, options *types.ProcessingOptions, duration time.Duration) {
	fmt.Printf("✅ STIG Processing Complete!\n\n")

	// Main statistics
	fmt.Printf("📊 Processing Summary:\n")
	fmt.Printf("  Total rules processed: %d\n", result.Total)
	fmt.Printf("  Automatable rules: %d (%.1f%%)\n",
		result.Automatable,
		float64(result.Automatable)/float64(result.Total)*100)
	fmt.Printf("  Manual review required: %d (%.1f%%)\n",
		result.ManualReview,
		float64(result.ManualReview)/float64(result.Total)*100)
	fmt.Printf("  Fleet policies generated: %d\n", len(result.Policies))
	fmt.Printf("  Processing time: %v\n", duration)

	if len(result.Errors) > 0 {
		fmt.Printf("  Errors encountered: %d\n", len(result.Errors))
	}

	// Output location info
	if !options.DryRun {
		fmt.Printf("\n📁 Output:\n")
		fmt.Printf("  Location: %s\n", options.OutputDir)
		fmt.Printf("  Format: %s\n", options.Format)
		if options.Severity != "" {
			fmt.Printf("  Severity filter: %s\n", options.Severity)
		}
	} else {
		fmt.Printf("\n🚫 Dry run mode - no files written\n")
	}

	// Breakdown by severity if verbose and we have policies
	if options.Verbose && len(result.Policies) > 0 {
		printSeverityBreakdown(result.Policies)
	}

	// Success indicators
	if result.Automatable > 0 {
		automationRate := float64(result.Automatable) / float64(result.Total) * 100
		if automationRate >= 50 {
			fmt.Printf("\n🎉 Good automation coverage: %.1f%% of rules can be automated!\n", automationRate)
		} else if automationRate >= 25 {
			fmt.Printf("\n👍 Moderate automation coverage: %.1f%% of rules can be automated\n", automationRate)
		} else {
			fmt.Printf("\n⚠️  Low automation coverage: %.1f%% of rules can be automated\n", automationRate)
		}
	}
}

// printSeverityBreakdown displays a breakdown of policies by severity
func printSeverityBreakdown(policies []types.FleetPolicy) {
	fmt.Printf("\n📊 Breakdown by Severity:\n")

	severityCount := make(map[string]int)
	criticalCount := 0

	for _, policy := range policies {
		if severity, exists := policy.Metadata.Labels["stig.severity"]; exists {
			severityCount[severity]++
		} else {
			severityCount["unknown"]++
		}

		if policy.Spec.Critical {
			criticalCount++
		}
	}

	// Print in order: high, medium, low, unknown
	severityOrder := []string{"high", "medium", "low", "unknown"}
	for _, severity := range severityOrder {
		if count, exists := severityCount[severity]; exists && count > 0 {
			var emoji string
			switch severity {
			case "high":
				emoji = "🔴"
			case "medium":
				emoji = "🟡"
			case "low":
				emoji = "🟢"
			default:
				emoji = "⚪"
			}
			fmt.Printf("  %s %s: %d\n", emoji, severity, count)
		}
	}

	if criticalCount > 0 {
		fmt.Printf("  ⚡ Critical policies: %d\n", criticalCount)
	}
}

// printStatistics displays detailed STIG statistics
func printStatistics(stats *types.ProcessingStatistics) {
	fmt.Printf("📋 File Information:\n")
	fmt.Printf("  Title: %s\n", stats.Title)
	fmt.Printf("  Version: %s\n", stats.Version)
	fmt.Printf("  Total Rules: %d\n", stats.TotalRules)
	fmt.Printf("  Processing Time: %v\n", stats.ProcessingTime)

	fmt.Printf("\n🔍 Rule Categories:\n")
	fmt.Printf("  Registry Checks: %d (%.1f%%)\n",
		stats.RegistryRules,
		float64(stats.RegistryRules)/float64(stats.TotalRules)*100)
	fmt.Printf("  Group Policy: %d (%.1f%%)\n",
		stats.GroupPolicyRules,
		float64(stats.GroupPolicyRules)/float64(stats.TotalRules)*100)
	fmt.Printf("  Manual Review: %d (%.1f%%)\n",
		stats.ManualRules,
		float64(stats.ManualRules)/float64(stats.TotalRules)*100)

	if len(stats.SeverityDistribution) > 0 {
		fmt.Printf("\n📊 Severity Distribution:\n")
		severityOrder := []string{"high", "medium", "low"}
		for _, severity := range severityOrder {
			if count, exists := stats.SeverityDistribution[severity]; exists {
				var emoji string
				switch severity {
				case "high":
					emoji = "🔴"
				case "medium":
					emoji = "🟡"
				case "low":
					emoji = "🟢"
				}
				fmt.Printf("  %s %s: %d (%.1f%%)\n",
					emoji,
					severity,
					count,
					float64(count)/float64(stats.TotalRules)*100)
			}
		}

		// Show any other severities
		for severity, count := range stats.SeverityDistribution {
			found := false
			for _, ordered := range severityOrder {
				if severity == ordered {
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("  ⚪ %s: %d (%.1f%%)\n",
					severity,
					count,
					float64(count)/float64(stats.TotalRules)*100)
			}
		}
	}

	// Automation potential
	automatable := stats.RegistryRules
	fmt.Printf("\n🤖 Automation Potential:\n")
	fmt.Printf("  Automatable: %d rules (%.1f%%)\n",
		automatable,
		float64(automatable)/float64(stats.TotalRules)*100)
	fmt.Printf("  Manual effort saved: ~%d hours\n", automatable*2) // Estimate 2 hours per automated rule
}
