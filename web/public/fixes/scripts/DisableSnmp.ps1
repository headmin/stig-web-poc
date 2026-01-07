Disable-WindowsOptionalFeature -Online -FeatureName "SNMP" -NoRestart

Get-WindowsOptionalFeature -Online -FeatureName "SNMP"