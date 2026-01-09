Set-SmbClientConfiguration -RequireSecuritySignature $true -Confirm:$false

# Verify the change
Get-SmbClientConfiguration | Select-Object RequireSecuritySignature
