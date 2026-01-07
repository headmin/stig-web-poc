# STIG Processor

A containerized Go application that processes DISA STIG (Security Technical Implementation Guide) rules and converts them into Fleet/osquery policy queries for automated compliance checking.

## 🚀 Quick Start

```bash
# Build and run
cd src/stig-processor
make quick-start

# Or using Docker
make docker-build docker-run
```

## 📁 Project Structure

```
src/stig-processor/
├── cmd/                          # Main application entry point
│   └── main.go                   # CLI interface with commands
├── pkg/                          # Public packages (importable)
│   ├── types/                    # Data structures and types
│   │   └── types.go              # STIG, Fleet, and processing types
│   ├── parser/                   # STIG and registry parsing logic
│   │   └── parser.go             # Parse STIG JSON and registry patterns
│   └── generator/                # Fleet policy generation
│       └── generator.go          # Generate YAML/JSON policies
├── internal/                     # Private application packages
│   └── processor/                # Main processing orchestration
│       └── processor.go          # Workflow coordination and validation
├── go.mod                        # Go module definition
├── go.sum                        # Dependency checksums
├── Dockerfile                    # Multi-stage container build
├── Makefile                      # Build automation and shortcuts
└── README.md                     # This file
```

## ✨ Features

### Core Capabilities
- ✅ **STIG JSON Parsing**: Processes DISA STIG JSON files (428+ rules)
- ✅ **Registry Pattern Matching**: Extracts Windows registry checks automatically
- ✅ **Fleet Policy Generation**: Creates GitOps-ready YAML/JSON policies
- ✅ **osquery SQL Generation**: Converts registry checks to SQL queries
- ✅ **Severity Filtering**: Filter by low/medium/high severity levels
- ✅ **Comprehensive Validation**: Syntax and schema validation
- ✅ **Statistics & Analytics**: Detailed processing and automation metrics

### Advanced Features
- 🔧 **Structured Logging**: Verbose mode with detailed processing info
- 🐋 **Container-First Design**: Optimized for containerized environments
- 🚀 **Performance Optimized**: Processes hundreds of rules in milliseconds
- 🛡️ **Security Focused**: Non-root containers, vulnerability scanning
- 📊 **Rich Metadata**: Labels, annotations, and comprehensive descriptions

## 🛠 Installation & Usage

### Prerequisites

- Go 1.21+
- Make
- Docker (optional)

### Local Development

```bash
# Set up development environment
make dev-setup

# Build the binary
make build-local

# Run with sample data
./stig-processor -input ../../microsoft-windows-11-security-technical-implementation-guide.json -output ./output -verbose

# Show help
./stig-processor -help
```

### Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-input` | Input STIG JSON file path | `microsoft-windows-11-security-technical-implementation-guide.json` |
| `-output` | Output directory for Fleet policies | `output` |
| `-format` | Output format (`yaml` or `json`) | `yaml` |
| `-severity` | Filter by severity (`low`, `medium`, `high`) | _(all)_ |
| `-verbose` | Enable verbose logging | `false` |
| `-dry-run` | Dry run without writing files | `false` |
| `-pretty` | Pretty print JSON output | `false` |
| `-timeout` | Processing timeout | `5m` |
| `-stats` | Show STIG statistics only | `false` |
| `-validate` | Validate existing policies only | `false` |
| `-version` | Show version information | `false` |

### Examples

```bash
# Basic processing
./stig-processor -input stig.json -output policies/ -verbose

# Filter high-severity rules only
./stig-processor -severity high -format json -verbose

# Show statistics without generating policies
./stig-processor -stats -input stig.json

# Validate existing policies
./stig-processor -validate -output policies/

# Dry run to see what would be generated
./stig-processor -dry-run -verbose
```

## 🐋 Docker Usage

### Build Container

```bash
# Build with version info
make docker-build VERSION=v1.0.0

# Or using Docker directly
docker build -t stig-processor \
  --build-arg VERSION=v1.0.0 \
  --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  .
```

### Run Container

```bash
# Using Makefile
make docker-run

# Or using Docker directly
docker run --rm \
  -v $(pwd)/../../:/app/input \
  -v $(pwd)/output:/app/output \
  stig-processor \
  -input /app/input/microsoft-windows-11-security-technical-implementation-guide.json \
  -output /app/output \
  -severity high \
  -verbose
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `INPUT_FILE` | Default input file path | `/app/input/microsoft-windows-11-security-technical-implementation-guide.json` |
| `OUTPUT_DIR` | Default output directory | `/app/output` |
| `FORMAT` | Default output format | `yaml` |
| `VERBOSE` | Enable verbose logging | `false` |
| `DRY_RUN` | Enable dry run mode | `false` |
| `TIMEOUT` | Processing timeout | `5m` |

## 📊 Processing Results

### Sample Statistics

From Windows 11 STIG v2 (258 rules):

```
📊 STIG Statistics

📋 File Information:
  Title: Microsoft Windows 11 Security Technical Implementation Guide
  Version: 2
  Total Rules: 258

🔍 Rule Categories:
  Registry Checks: 122 (47.3%)  ← Automatable
  Group Policy: 46 (17.8%)      ← Manual Review
  Manual Review: 90 (34.9%)     ← Manual Review

📊 Severity Distribution:
  🔴 high: 28 (10.9%)
  🟡 medium: 213 (82.6%)
  🟢 low: 17 (6.6%)

🤖 Automation Potential:
  Automatable: 122 rules (47.3%)
  Manual effort saved: ~244 hours
```

### Generated Fleet Policies

Example policy output:

```yaml
apiVersion: v1
kind: policy
metadata:
  name: stig-v-253370-wn11-cc-000075
  labels:
    compliance.source: disa
    compliance.type: stig
    stig.group_id: V-253370
    stig.severity: high
  annotations:
    generated.timestamp: "2026-01-06T16:33:25Z"
    generated.tool: stig-processor
    registry.hive: HKEY_LOCAL_MACHINE
    registry.path: SOFTWARE\Policies\Microsoft\Windows\DeviceGuard
    registry.value_name: LsaCfgFlags
    registry.expected_value: "1"
spec:
  name: "STIG V-253370: Credential Guard must be running on Windows 11 domain-joined systems"
  query: "SELECT name, type, data FROM registry WHERE path = 'HKEY_LOCAL_MACHINE\\SOFTWARE\\Policies\\Microsoft\\Windows\\DeviceGuard' AND name = 'LsaCfgFlags' AND CAST(data AS INTEGER) = 1"
  description: "STIG Rule V-253370 (Severity: high)\n\nCredential Guard uses virtualization-based security..."
  resolution: "Configure the policy value for Computer Configuration >> Administrative Templates..."
  platform: windows
  critical: true
```

### Processing Summary

Generated summary file:

```yaml
total_rules: 28
automatable: 15
manual_review: 13
policies_generated: 15
processing_time: 13.350167ms
timestamp: "2026-01-06T16:33:25Z"
policies:
  - name: stig-v-253370-wn11-cc-000075
    title: "STIG V-253370: Credential Guard must be running..."
    platform: windows
    critical: true
    severity: high
    group_id: V-253370
    rule_version: WN11-CC-000075
```

## 🏗 Architecture

### Package Structure

- **`cmd/`**: CLI interface and main application entry point
- **`pkg/types/`**: Shared data structures for STIG and Fleet policies
- **`pkg/parser/`**: STIG JSON parsing and registry pattern extraction
- **`pkg/generator/`**: Fleet policy generation and formatting
- **`internal/processor/`**: Main workflow orchestration (internal only)

### Processing Pipeline

```
1. Input Validation
   ├── File existence
   ├── Format validation
   └── Parameter validation

2. STIG Parsing
   ├── JSON parsing
   ├── Structure validation
   └── Rule extraction

3. Registry Pattern Matching
   ├── Hive extraction (HKEY_*)
   ├── Path normalization
   ├── Value name/type parsing
   └── Expected value handling

4. Fleet Policy Generation
   ├── osquery SQL generation
   ├── Metadata enrichment
   ├── Label/annotation creation
   └── Validation

5. Output Generation
   ├── YAML/JSON formatting
   ├── File writing
   └── Summary creation
```

### Registry Check Processing

The processor automatically converts Windows registry patterns:

**Input (STIG Check Content)**:
```
Registry Hive: HKEY_LOCAL_MACHINE
Registry Path: \SOFTWARE\Policies\Microsoft\Windows\DeviceGuard\
Value Name: LsaCfgFlags
Value Type: REG_DWORD
Value: 0x00000001 (1)
```

**Output (osquery SQL)**:
```sql
SELECT name, type, data FROM registry 
WHERE path = 'HKEY_LOCAL_MACHINE\\SOFTWARE\\Policies\\Microsoft\\Windows\\DeviceGuard' 
AND name = 'LsaCfgFlags' 
AND CAST(data AS INTEGER) = 1
```

## 🧪 Development

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run benchmarks
make test-benchmark

# Integration tests
make test-integration
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Vet code
make vet

# Run all quality checks
make quality-check
```

### Debugging

```bash
# Build with debug symbols
make build-debug

# Run with debugger (requires delve)
make debug-run

# Or manually
dlv exec ./stig-processor-debug -- -input stig.json -verbose
```

### Development Workflow

```bash
# 1. Set up environment
make dev-setup

# 2. Run in development mode
make run-dev

# 3. Make changes and test
make test

# 4. Validate output
make validate-output

# 5. Build final binary
make build
```

## 🚀 Deployment

### GitHub Actions Integration

The processor is designed to run in CI/CD pipelines. See the main project's `.github/workflows/stig-processor.yml` for a complete GitHub Actions setup.

### Production Deployment

```bash
# 1. Build production container
make docker-build VERSION=v1.0.0

# 2. Push to registry
make docker-push

# 3. Deploy to production
docker run --rm \
  -v /production/stig-data:/app/input \
  -v /production/fleet-policies:/app/output \
  ghcr.io/your-org/stig-processor:v1.0.0 \
  -severity high -format yaml -verbose
```

### Fleet Integration

```bash
# 1. Generate policies
./stig-processor -input stig.json -output fleet-policies/

# 2. Apply to Fleet
fleetctl apply -f fleet-policies/

# 3. Monitor compliance
fleetctl get policies
```

## 📈 Performance

### Benchmarks

| Metric | Value |
|--------|-------|
| **Processing Time** | < 1 second for 400+ rules |
| **Memory Usage** | ~25MB peak |
| **Container Size** | ~15MB (multi-stage build) |
| **Cold Start** | < 100ms |

### Optimization Features

- Pre-compiled regex patterns for performance
- Streaming JSON parsing for large files
- Concurrent policy generation (where applicable)
- Minimal memory allocations
- Efficient container layering

## 🔒 Security

### Container Security

- Non-root user execution
- Minimal Alpine Linux base image
- No sensitive data in logs
- Vulnerability scanning with Trivy
- Read-only file system support

### Data Protection

- STIG data processed in memory only
- Generated policies contain no credentials
- Temporary files automatically cleaned
- No network connections required

## 🐛 Troubleshooting

### Common Issues

**No policies generated**:
```bash
# Check input file format
./stig-processor -stats -input your-file.json

# Enable verbose logging
./stig-processor -input your-file.json -verbose -dry-run
```

**Invalid YAML/JSON output**:
```bash
# Validate generated policies
./stig-processor -validate -output ./policies/

# Check for parsing errors
./stig-processor -input stig.json -verbose | grep -i error
```

**Container permission errors**:
```bash
# Ensure proper volume mounts
docker run --rm \
  -v $(pwd):/app/input:ro \
  -v $(pwd)/output:/app/output \
  stig-processor

# Check file permissions
ls -la microsoft-windows-11-security-technical-implementation-guide.json
```

### Debug Mode

```bash
# Maximum verbosity
./stig-processor -input stig.json -verbose -dry-run

# Show specific rule processing
./stig-processor -input stig.json -verbose 2>&1 | grep "V-253370"

# Container debugging
docker run --rm -it --entrypoint /bin/sh stig-processor
```

## 📝 Contributing

### Code Structure Guidelines

1. **Separation of Concerns**: 
   - `cmd/` for CLI only
   - `pkg/` for reusable components
   - `internal/` for application-specific logic

2. **Error Handling**:
   - Use typed errors from `pkg/types`
   - Provide context in error messages
   - Log errors at appropriate levels

3. **Testing**:
   - Unit tests for all packages
   - Integration tests for workflows
   - Benchmarks for performance-critical code

4. **Documentation**:
   - GoDoc comments for all public functions
   - README updates for new features
   - Examples in code comments

### Submitting Changes

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Run quality checks: `make quality-check`
5. Test with real STIG data
6. Submit pull request

## 📄 License

This project is intended for educational and compliance automation purposes. Ensure generated policies align with your organization's security requirements before deployment.

---

**Version**: Latest  
**Go Version**: 1.21+  
**Container Registry**: `ghcr.io/your-org/stig-processor`  
**Documentation**: See `IMPLEMENTATION_COMPARISON.md` for detailed comparisons with other implementations.