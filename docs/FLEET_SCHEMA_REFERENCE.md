# Fleet GitOps Policy Schema Reference

> **Auto-generated** from Fleet source at commit [`6d9a29d4ce3e`](https://github.com/fleetdm/fleet/tree/6d9a29d4ce3e)
> 
> Extracted: 2026-01-07T13:59:27.635083

## Valid Policy Fields

These are the **only** fields that should be used in Fleet policy YAML files.
Using any other field names will result in invalid configuration.

### Core Fields (Global & Team Policies)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | ✅ | Policy name (must be unique) |
| `query` | string | ✅ | osquery SQL query |
| `description` | string | | Policy description |
| `critical` | boolean | | Mark as high impact (default: false) |
| `resolution` | string | | How to fix a failing policy |
| `platform` | string | | Target platforms: `darwin`, `windows`, `linux` |

### Team-Only Fields

These fields can **only** be used in team policy files, not global policies.

| Field | Type | Description |
|-------|------|-------------|
| `run_script` | object | Script to run on policy failure |
| `run_script.path` | string | Relative path to script file |
| `install_software` | object | Software to install on policy failure |
| `install_software.package_path` | string | Path to software package YAML |
| `install_software.app_store_id` | string | App Store ID for VPP apps |
| `install_software.hash_sha256` | string | SHA256 hash of package |
| `calendar_events_enabled` | boolean | Enable calendar events for failures |
| `labels_include_any` | string[] | Include hosts with any of these labels |
| `labels_exclude_any` | string[] | Exclude hosts with any of these labels |
| `conditional_access_enabled` | boolean | Enable Entra conditional access |

## Example Policy YAML

```yaml
# Basic policy (works globally or in team)
- name: Windows - Disk encryption enabled
  query: |
    SELECT 1 FROM bitlocker_info
    WHERE drive_letter = 'C:'
    AND protection_status = 1;
  critical: false
  description: Checks if BitLocker is enabled on C: drive
  resolution: Enable BitLocker via Windows Settings > Privacy & Security
  platform: windows

# Team policy with script remediation
- name: macOS - Nudge installed
  query: SELECT 1 FROM apps WHERE bundle_identifier = 'com.github.macadmins.Nudge';
  critical: true
  description: Ensures Nudge is installed for OS update enforcement
  resolution: Nudge will be automatically installed
  platform: darwin
  run_script:
    path: ../scripts/install-nudge.sh
```

## Source Files

- [`PolicySpec`](https://github.com/fleetdm/fleet/blob/6d9a29d4ce3e/server/fleet/policies.go#L354) - server/fleet/policies.go:354
- [`PolicyData`](https://github.com/fleetdm/fleet/blob/6d9a29d4ce3e/server/fleet/policies.go#L225) - server/fleet/policies.go:225
- [`Policy`](https://github.com/fleetdm/fleet/blob/6d9a29d4ce3e/server/fleet/policies.go#L276) - server/fleet/policies.go:276
- [`GitOpsPolicySpec`](https://github.com/fleetdm/fleet/blob/6d9a29d4ce3e/pkg/spec/gitops.go#L184) - pkg/spec/gitops.go:184
- [`PolicyRunScript`](https://github.com/fleetdm/fleet/blob/6d9a29d4ce3e/pkg/spec/gitops.go#L196) - pkg/spec/gitops.go:196
- [`PolicyInstallSoftware`](https://github.com/fleetdm/fleet/blob/6d9a29d4ce3e/pkg/spec/gitops.go#L200) - pkg/spec/gitops.go:200

## Invalid Fields (DO NOT USE)

The following field names are **NOT** part of the Fleet schema and should never be used:

| Invalid Field | Notes |
|---------------|-------|
| `fix_script` | ❌ Use `run_script.path` instead |
| `fix` | ❌ Not a valid field |
| `remediation` | ❌ Use `resolution` instead |
| `severity` | ❌ Use `critical: true` for high severity |
| `tags` | ❌ Not supported in GitOps format |

---

*This file is auto-generated. Do not edit manually.*
*Run `uv run python scripts/extract_fleet_schema.py` to update.*