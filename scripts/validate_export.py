#!/usr/bin/env python3
"""
Validate exported STIG policy YAML against Fleet schema.

Usage:
    uv run python scripts/validate_export.py path/to/policies.yml
"""

import argparse
import json
import sys
from pathlib import Path

import yaml


def load_schema() -> dict:
    """Load the Fleet policy schema."""
    schema_path = Path(__file__).parent.parent / "web/src/schema/fleet-policy-schema.json"
    if not schema_path.exists():
        print(f"Error: Schema not found at {schema_path}")
        print("Run: uv run python scripts/extract_fleet_schema.py")
        sys.exit(1)
    return json.loads(schema_path.read_text())


def validate_policy(policy: dict, schema: dict, policy_index: int) -> list[str]:
    """Validate a single policy against the schema."""
    errors = []
    valid_fields = set(schema["_meta"]["valid_fields"])
    
    # Check for unknown fields
    for key in policy.keys():
        if key not in valid_fields:
            errors.append(f"Policy {policy_index} '{policy.get('name', 'unknown')}': Invalid field '{key}' - not in Fleet schema")
    
    # Check required fields
    for required in schema.get("required", []):
        if required not in policy:
            errors.append(f"Policy {policy_index} '{policy.get('name', 'unknown')}': Missing required field '{required}'")
    
    # Validate run_script structure
    if "run_script" in policy:
        rs = policy["run_script"]
        if not isinstance(rs, dict):
            errors.append(f"Policy {policy_index}: run_script must be an object with 'path' field")
        elif "path" not in rs:
            errors.append(f"Policy {policy_index}: run_script missing required 'path' field")
    
    # Validate install_software structure
    if "install_software" in policy:
        isw = policy["install_software"]
        if not isinstance(isw, dict):
            errors.append(f"Policy {policy_index}: install_software must be an object")
        else:
            valid_isw_fields = {"package_path", "app_store_id", "hash_sha256"}
            for key in isw.keys():
                if key not in valid_isw_fields:
                    errors.append(f"Policy {policy_index}: install_software has invalid field '{key}'")
    
    return errors


def validate_yaml_file(path: Path) -> tuple[int, list[str]]:
    """Validate a YAML file containing policies."""
    schema = load_schema()
    
    content = path.read_text()
    
    # Handle multi-document YAML (---) or list format
    all_errors = []
    policy_count = 0
    
    try:
        docs = list(yaml.safe_load_all(content))
    except yaml.YAMLError as e:
        return 0, [f"YAML parse error: {e}"]
    
    for doc in docs:
        if doc is None:
            continue
        
        # Handle list of policies or single policy
        policies = doc if isinstance(doc, list) else [doc]
        
        for policy in policies:
            if not isinstance(policy, dict):
                continue
            policy_count += 1
            errors = validate_policy(policy, schema, policy_count)
            all_errors.extend(errors)
    
    return policy_count, all_errors


def main():
    parser = argparse.ArgumentParser(description="Validate Fleet policy YAML")
    parser.add_argument("file", type=Path, help="YAML file to validate")
    parser.add_argument("--quiet", "-q", action="store_true", help="Only show errors")
    
    args = parser.parse_args()
    
    if not args.file.exists():
        print(f"Error: File not found: {args.file}")
        return 1
    
    policy_count, errors = validate_yaml_file(args.file)
    
    if not args.quiet:
        print(f"Validated {policy_count} policies in {args.file}")
    
    if errors:
        print(f"\n❌ Found {len(errors)} validation errors:\n")
        for error in errors:
            print(f"  • {error}")
        return 1
    else:
        if not args.quiet:
            print("✅ All policies are valid Fleet schema")
        return 0


if __name__ == "__main__":
    sys.exit(main())
