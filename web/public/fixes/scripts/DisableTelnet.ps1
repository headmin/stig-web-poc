Disable-WindowsOptionalFeature -Online -FeatureName "TelnetClient" -NoRestart

Get-WindowsOptionalFeature -Online -FeatureName "TelnetClient"