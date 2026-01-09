if (-not (Test-Path "HKLM:\SYSTEM\CurrentControlSet\Services\LDAP")) {
    New-Item -Path "HKLM:\SYSTEM\CurrentControlSet\Services\LDAP" -Force | Out-Null
}
New-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\LDAP" -Name "LDAPClientIntegrity" -Value 1 -Type DWord -Force
