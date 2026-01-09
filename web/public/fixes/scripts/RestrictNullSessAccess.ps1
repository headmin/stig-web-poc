$RegistryPath = "HKLM:\System\CurrentControlSet\Services\LanManServer\Parameters"
$ValueName = "RestrictNullSessAccess"
$ValueData = 1

# Create the registry path if it doesn't exist
if (-not (Test-Path $RegistryPath)) {
    New-Item -Path $RegistryPath -Force | Out-Null
}

# Set the registry value
Set-ItemProperty -Path $RegistryPath -Name $ValueName -Value $ValueData -Type DWord

# Verify the change
Get-ItemProperty -Path $RegistryPath -Name $ValueName