# STIG POC

Proof of concept for STIG Benchmark Web Builder.

## Tools Overview

### Go Applications
- **stig-processor**: Main CLI tool that converts STIG JSON to Fleet policies (YAML) with advanced multi-query registry parsing
- **stig-data-combiner**: Processes STIG data for the web interface (creates benchmark-data.json)

### GitHub Actions
1. **[1] Fetch STIG**: Downloads latest DISA STIG JSON files
2. **[2] Extract Fleet Schema**: Gets Fleet policy schema for web validation
3. **[3] Generate STIG Artifacts**: Creates web data files from STIG JSON
4. **[4] Release Go Binaries**: Builds/releases the Go applications
5. **[5] Deploy Web**: Publishes the Vue.js STIG browser to GitHub Pages

## Architecture

### Where We Are Now
Instead of the original Marimo notebook that required manual Python execution, we now have:
- **Interactive web interface** (Vue.js) for browsing/filtering STIG rules
- **Automated CLI tools** (Go) for generating Fleet policies
- **Complete CI/CD pipeline** for automated processing and deployment

### STIG Data Handling
No splitting needed - we load the complete STIG JSON (258 rules) and the web interface handles filtering/categorization by severity, automation status, and rule type in real-time. The Go tools process the entire JSON and generate individual Fleet policy files automatically.

## TEST DEPLYOMENT

Visit https://headmin.github.io/stig-web-poc/
