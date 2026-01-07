Disable-WindowsOptionalFeature -Online -FeatureName "SimpleTCP" -Remove -NoRestart

Get-WindowsOptionalFeature -Online -FeatureName "SimpleTCP"