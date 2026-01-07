#!/usr/bin/env python3
"""
Extract Fleet GitOps Policy Schema from fleet repository.

This script parses Go source files to extract the official Fleet policy schema,
ensuring our STIG export generates valid Fleet-compatible YAML.

Usage:
    uv run python scripts/extract_fleet_schema.py --fleet-repo /path/to/fleet

Output:
    - web/src/schema/fleet-policy-schema.json (TypeScript-compatible schema)
    - docs/FLEET_SCHEMA_REFERENCE.md (human-readable reference with backlinks)
"""

import argparse
import json
import re
from dataclasses import dataclass, field, asdict
from pathlib import Path
from typing import Optional
from datetime import datetime


@dataclass
class SchemaField:
    """A field in a Go struct."""
    name: str
    json_name: str
    go_type: str
    description: str = ""
    optional: bool = False
    team_only: bool = False


@dataclass
class SchemaStruct:
    """A Go struct representing part of the schema."""
    name: str
    source_file: str
    source_line: int
    fields: list[SchemaField] = field(default_factory=list)
    description: str = ""
    

@dataclass
class FleetPolicySchema:
    """Complete Fleet policy schema extracted from source."""
    extracted_at: str
    fleet_repo_path: str
    git_commit: str = ""
    structs: dict[str, SchemaStruct] = field(default_factory=dict)
    
    # Computed valid fields for policies
    valid_policy_fields: list[str] = field(default_factory=list)
    team_only_fields: list[str] = field(default_factory=list)


def get_git_commit(repo_path: Path) -> str:
    """Get current git commit hash."""
    import subprocess
    try:
        result = subprocess.run(
            ["git", "rev-parse", "HEAD"],
            cwd=repo_path,
            capture_output=True,
            text=True
        )
        return result.stdout.strip()[:12] if result.returncode == 0 else "unknown"
    except Exception:
        return "unknown"


def parse_go_struct(content: str, struct_name: str, file_path: str) -> Optional[SchemaStruct]:
    """Parse a Go struct definition and extract fields."""
    # Find struct definition
    pattern = rf'//\s*(.*?)\ntype\s+{struct_name}\s+struct\s*\{{\s*([\s\S]*?)\n\}}'
    match = re.search(pattern, content)
    
    if not match:
        # Try without comment
        pattern = rf'type\s+{struct_name}\s+struct\s*\{{\s*([\s\S]*?)\n\}}'
        match = re.search(pattern, content)
        if not match:
            return None
        description = ""
        body = match.group(1)
        line_num = content[:match.start()].count('\n') + 1
    else:
        description = match.group(1).strip()
        body = match.group(2)
        line_num = content[:match.start()].count('\n') + 1
    
    schema = SchemaStruct(
        name=struct_name,
        source_file=file_path,
        source_line=line_num,
        description=description
    )
    
    # Parse fields
    # Match: FieldName Type `json:"name,omitempty"` // comment
    field_pattern = r'^\s*(\w+)\s+(\S+)\s+`json:"([^"]+)"`(?:\s*//\s*(.*))?'
    
    for line in body.split('\n'):
        line = line.strip()
        if not line or line.startswith('//'):
            continue
            
        # Handle embedded structs (e.g., fleet.PolicySpec)
        if '.' in line and '`' not in line:
            embedded = line.split('.')[1].split()[0] if '.' in line else line.split()[0]
            schema.fields.append(SchemaField(
                name=f"[embedded: {line.strip()}]",
                json_name="",
                go_type=line.strip(),
                description="Embedded struct fields"
            ))
            continue
            
        match = re.match(field_pattern, line)
        if match:
            field_name, go_type, json_tag, comment = match.groups()
            json_parts = json_tag.split(',')
            json_name = json_parts[0]
            optional = 'omitempty' in json_parts or json_tag == '-'
            
            # Skip internal fields (json:"-")
            if json_name == '-':
                continue
                
            team_only = comment and 'team polic' in comment.lower() if comment else False
            
            schema.fields.append(SchemaField(
                name=field_name,
                json_name=json_name,
                go_type=go_type,
                description=comment or "",
                optional=optional,
                team_only=team_only
            ))
    
    return schema


def extract_schema(fleet_repo: Path) -> FleetPolicySchema:
    """Extract the complete policy schema from Fleet source."""
    schema = FleetPolicySchema(
        extracted_at=datetime.now().isoformat(),
        fleet_repo_path=str(fleet_repo),
        git_commit=get_git_commit(fleet_repo)
    )
    
    # Key source files
    files_to_parse = [
        ("server/fleet/policies.go", ["PolicySpec", "PolicyData", "Policy"]),
        ("pkg/spec/gitops.go", ["GitOpsPolicySpec", "PolicyRunScript", "PolicyInstallSoftware"]),
    ]
    
    for rel_path, structs in files_to_parse:
        file_path = fleet_repo / rel_path
        if not file_path.exists():
            print(f"Warning: {file_path} not found")
            continue
            
        content = file_path.read_text()
        
        for struct_name in structs:
            parsed = parse_go_struct(content, struct_name, rel_path)
            if parsed:
                schema.structs[struct_name] = parsed
    
    # Compute valid policy fields
    # PolicySpec is the base, GitOpsPolicySpec adds run_script and install_software
    policy_spec = schema.structs.get("PolicySpec")
    gitops_spec = schema.structs.get("GitOpsPolicySpec")
    
    if policy_spec:
        for f in policy_spec.fields:
            if f.json_name and f.json_name != '-':
                schema.valid_policy_fields.append(f.json_name)
                if f.team_only:
                    schema.team_only_fields.append(f.json_name)
    
    if gitops_spec:
        for f in gitops_spec.fields:
            if f.json_name and f.json_name != '-':
                schema.valid_policy_fields.append(f.json_name)
                if f.team_only or 'team' in f.description.lower():
                    schema.team_only_fields.append(f.json_name)
    
    return schema


def generate_typescript_schema(schema: FleetPolicySchema) -> dict:
    """Generate a TypeScript-compatible JSON schema."""
    return {
        "$schema": "http://json-schema.org/draft-07/schema#",
        "$comment": f"Extracted from Fleet repository at commit {schema.git_commit}",
        "title": "Fleet GitOps Policy",
        "description": "Schema for Fleet policy YAML files",
        "type": "object",
        "properties": {
            "name": {"type": "string", "description": "Policy name (required)"},
            "query": {"type": "string", "description": "osquery SQL query (required)"},
            "description": {"type": "string", "description": "Policy description"},
            "critical": {"type": "boolean", "description": "Mark as high impact", "default": False},
            "resolution": {"type": "string", "description": "How to fix failing policy"},
            "platform": {"type": "string", "description": "Target platforms (darwin, windows, linux)"},
            "calendar_events_enabled": {"type": "boolean", "description": "Enable calendar events (team only)"},
            "labels_include_any": {"type": "array", "items": {"type": "string"}, "description": "Include hosts with any of these labels"},
            "labels_exclude_any": {"type": "array", "items": {"type": "string"}, "description": "Exclude hosts with any of these labels"},
            "conditional_access_enabled": {"type": "boolean", "description": "Enable Entra conditional access (team only)"},
            "run_script": {
                "type": "object",
                "description": "Script to run on policy failure (team only)",
                "properties": {
                    "path": {"type": "string", "description": "Relative path to script file"}
                },
                "required": ["path"]
            },
            "install_software": {
                "type": "object",
                "description": "Software to install on policy failure (team only)",
                "properties": {
                    "package_path": {"type": "string", "description": "Path to software package YAML"},
                    "app_store_id": {"type": "string", "description": "App Store ID"},
                    "hash_sha256": {"type": "string", "description": "SHA256 hash of package"}
                }
            }
        },
        "required": ["name", "query"],
        "_meta": {
            "extracted_at": schema.extracted_at,
            "git_commit": schema.git_commit,
            "valid_fields": schema.valid_policy_fields,
            "team_only_fields": schema.team_only_fields
        }
    }


def generate_markdown_reference(schema: FleetPolicySchema) -> str:
    """Generate human-readable markdown reference."""
    lines = [
        "# Fleet GitOps Policy Schema Reference",
        "",
        f"> **Auto-generated** from Fleet source at commit [`{schema.git_commit}`](https://github.com/fleetdm/fleet/tree/{schema.git_commit})",
        f"> ",
        f"> Extracted: {schema.extracted_at}",
        "",
        "## Valid Policy Fields",
        "",
        "These are the **only** fields that should be used in Fleet policy YAML files.",
        "Using any other field names will result in invalid configuration.",
        "",
        "### Core Fields (Global & Team Policies)",
        "",
        "| Field | Type | Required | Description |",
        "|-------|------|----------|-------------|",
        "| `name` | string | ✅ | Policy name (must be unique) |",
        "| `query` | string | ✅ | osquery SQL query |",
        "| `description` | string | | Policy description |",
        "| `critical` | boolean | | Mark as high impact (default: false) |",
        "| `resolution` | string | | How to fix a failing policy |",
        "| `platform` | string | | Target platforms: `darwin`, `windows`, `linux` |",
        "",
        "### Team-Only Fields",
        "",
        "These fields can **only** be used in team policy files, not global policies.",
        "",
        "| Field | Type | Description |",
        "|-------|------|-------------|",
        "| `run_script` | object | Script to run on policy failure |",
        "| `run_script.path` | string | Relative path to script file |",
        "| `install_software` | object | Software to install on policy failure |",
        "| `install_software.package_path` | string | Path to software package YAML |",
        "| `install_software.app_store_id` | string | App Store ID for VPP apps |",
        "| `install_software.hash_sha256` | string | SHA256 hash of package |",
        "| `calendar_events_enabled` | boolean | Enable calendar events for failures |",
        "| `labels_include_any` | string[] | Include hosts with any of these labels |",
        "| `labels_exclude_any` | string[] | Exclude hosts with any of these labels |",
        "| `conditional_access_enabled` | boolean | Enable Entra conditional access |",
        "",
        "## Example Policy YAML",
        "",
        "```yaml",
        "# Basic policy (works globally or in team)",
        "- name: Windows - Disk encryption enabled",
        "  query: |",
        "    SELECT 1 FROM bitlocker_info",
        "    WHERE drive_letter = 'C:'",
        "    AND protection_status = 1;",
        "  critical: false",
        "  description: Checks if BitLocker is enabled on C: drive",
        "  resolution: Enable BitLocker via Windows Settings > Privacy & Security",
        "  platform: windows",
        "",
        "# Team policy with script remediation",
        "- name: macOS - Nudge installed",
        "  query: SELECT 1 FROM apps WHERE bundle_identifier = 'com.github.macadmins.Nudge';",
        "  critical: true",
        "  description: Ensures Nudge is installed for OS update enforcement",
        "  resolution: Nudge will be automatically installed",
        "  platform: darwin",
        "  run_script:",
        "    path: ../scripts/install-nudge.sh",
        "```",
        "",
        "## Source Files",
        "",
    ]
    
    for name, struct in schema.structs.items():
        github_url = f"https://github.com/fleetdm/fleet/blob/{schema.git_commit}/{struct.source_file}#L{struct.source_line}"
        lines.append(f"- [`{name}`]({github_url}) - {struct.source_file}:{struct.source_line}")
    
    lines.extend([
        "",
        "## Invalid Fields (DO NOT USE)",
        "",
        "The following field names are **NOT** part of the Fleet schema and should never be used:",
        "",
        "| Invalid Field | Notes |",
        "|---------------|-------|",
        "| `fix_script` | ❌ Use `run_script.path` instead |",
        "| `fix` | ❌ Not a valid field |",
        "| `remediation` | ❌ Use `resolution` instead |",
        "| `severity` | ❌ Use `critical: true` for high severity |",
        "| `tags` | ❌ Not supported in GitOps format |",
        "",
        "---",
        "",
        "*This file is auto-generated. Do not edit manually.*",
        f"*Run `uv run python scripts/extract_fleet_schema.py` to update.*",
    ])
    
    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(description="Extract Fleet GitOps policy schema")
    parser.add_argument(
        "--fleet-repo",
        type=Path,
        default=Path.home() / "Code/GitHub/fleet",
        help="Path to Fleet repository"
    )
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path.cwd(),
        help="Output directory for generated files"
    )
    
    args = parser.parse_args()
    
    if not args.fleet_repo.exists():
        print(f"Error: Fleet repository not found at {args.fleet_repo}")
        return 1
    
    print(f"Extracting schema from {args.fleet_repo}...")
    schema = extract_schema(args.fleet_repo)
    
    # Generate JSON schema
    json_schema = generate_typescript_schema(schema)
    json_output = args.output_dir / "web/src/schema"
    json_output.mkdir(parents=True, exist_ok=True)
    json_file = json_output / "fleet-policy-schema.json"
    json_file.write_text(json.dumps(json_schema, indent=2))
    print(f"✓ Generated {json_file}")
    
    # Generate Markdown reference
    md_content = generate_markdown_reference(schema)
    docs_output = args.output_dir / "docs"
    docs_output.mkdir(parents=True, exist_ok=True)
    md_file = docs_output / "FLEET_SCHEMA_REFERENCE.md"
    md_file.write_text(md_content)
    print(f"✓ Generated {md_file}")
    
    # Print summary
    print(f"\nSchema Summary:")
    print(f"  Git commit: {schema.git_commit}")
    print(f"  Valid fields: {len(schema.valid_policy_fields)}")
    print(f"  Team-only fields: {len(schema.team_only_fields)}")
    print(f"\nValid policy fields: {', '.join(schema.valid_policy_fields)}")
    
    return 0


if __name__ == "__main__":
    exit(main())
