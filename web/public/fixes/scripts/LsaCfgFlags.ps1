if (-not (Test-Path "HKLM:\SOFTWARE\Policies\Microsoft\Windows\DeviceGuard")) {
    New-Item -Path "HKLM:\SOFTWARE\Policies\Microsoft\Windows\DeviceGuard" -Force | Out-Null
}
New-ItemProperty -Path "HKLM:\SOFTWARE\Policies\Microsoft\Windows\DeviceGuard" -Name "LsaCfgFlags" -Value 1 -Type DWord -Force
