New-Item -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Netbt\Parameters" -Force -ErrorAction SilentlyContinue | Out-Null
New-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Netbt\Parameters" -Name "NoNameReleaseOnDemand" -Value 1 -Type DWord -Force
