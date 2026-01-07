Disable-WindowsOptionalFeature -Online -FeatureName "TFTP" -NoRestart

Get-WindowsOptionalFeature -Online -FeatureName "TFTP"