if (-not (Test-Path "HKLM:\SOFTWARE\Policies\Microsoft\Windows\Network Connections")) {
    New-Item -Path "HKLM:\SOFTWARE\Policies\Microsoft\Windows\Network Connections" -Force | Out-Null
}
New-ItemProperty -Path "HKLM:\SOFTWARE\Policies\Microsoft\Windows\Network Connections" -Name "NC_ShowSharedAccessUI" -Value 0 -Type DWord -Force
