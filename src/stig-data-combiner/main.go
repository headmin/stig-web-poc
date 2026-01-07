package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/stig-data-combiner/pkg/combiner"
)

func main() {
	// Command line flags
	stigPath := flag.String("stig", "", "Path to STIG JSON file (optional, will auto-detect)")
	winSTIGPath := flag.String("win-stig", "", "Path to win-stig repository (required)")
	outputPath := flag.String("output", "benchmark-data.json", "Output JSON file path")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	pretty := flag.Bool("pretty", true, "Pretty-print JSON output")

	flag.Parse()

	// Validate required flags
	if *winSTIGPath == "" {
		fmt.Fprintln(os.Stderr, "Error: -win-stig flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Verify win-stig path exists
	if _, err := os.Stat(*winSTIGPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: win-stig path does not exist: %s\n", *winSTIGPath)
		os.Exit(1)
	}

	// Verify required files exist
	policyFile := filepath.Join(*winSTIGPath, "stig-policy-queries.yml")
	if _, err := os.Stat(policyFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: stig-policy-queries.yml not found in %s\n", *winSTIGPath)
		os.Exit(1)
	}

	fixDir := filepath.Join(*winSTIGPath, "fix")
	if _, err := os.Stat(fixDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: fix directory not found in %s\n", *winSTIGPath)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Win-STIG path: %s\n", *winSTIGPath)
		fmt.Printf("Output path: %s\n", *outputPath)
		if *stigPath != "" {
			fmt.Printf("STIG JSON path: %s\n", *stigPath)
		}
	}

	// Create combiner and process
	c := combiner.NewCombiner(*stigPath, *winSTIGPath, *verbose)
	data, err := c.Combine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error combining data: %v\n", err)
		os.Exit(1)
	}

	// Marshal to JSON
	var jsonData []byte
	if *pretty {
		jsonData, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonData, err = json.Marshal(data)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Write output file
	if err := os.WriteFile(*outputPath, jsonData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	fmt.Printf("Generated benchmark data:\n")
	fmt.Printf("  Framework: %s\n", data.Meta.Framework)
	fmt.Printf("  Title: %s\n", data.Meta.Title)
	fmt.Printf("  Version: %s\n", data.Meta.Version)
	fmt.Printf("  Categories: %d\n", len(data.Categories))

	totalRules := 0
	automatableRules := 0
	rulesWithFixes := 0
	for _, cat := range data.Categories {
		totalRules += len(cat.Rules)
		for _, rule := range cat.Rules {
			if rule.Automatable {
				automatableRules++
			}
			if rule.Fix != nil {
				rulesWithFixes++
			}
		}
	}

	fmt.Printf("  Total rules: %d\n", totalRules)
	fmt.Printf("  Automatable: %d\n", automatableRules)
	fmt.Printf("  With fixes: %d\n", rulesWithFixes)
	fmt.Printf("\nOutput written to: %s\n", *outputPath)
}
