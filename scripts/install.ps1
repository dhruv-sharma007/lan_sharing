$ErrorActionPreference = "Stop"

$repoOwner = "dhruv-sharma007"
$repoName = "lan_sharing"

$appName = "LanShare"
$taskName = "LanShare"

Write-Host "Installing LanShare..."

# --------------------------------------------------
# Detect architecture
# --------------------------------------------------

$os = "windows"

if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq "X64") {
    $arch = "amd64"
}
elseif ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq "Arm64") {
    $arch = "arm64"
}
else {
    Write-Error "Unsupported Windows architecture."
    exit 1
}

$binaryName = "lanshare-$os-$arch.exe"

Write-Host "Detected: $os/$arch"

# --------------------------------------------------
# Find latest GitHub release
# --------------------------------------------------

$releasesUrl = "https://api.github.com/repos/$repoOwner/$repoName/releases/latest"

Write-Host "Fetching latest release..."

$release = Invoke-RestMethod `
    -Uri $releasesUrl `
    -Headers @{
        "Accept" = "application/vnd.github+json"
        "User-Agent" = "LanShare-Installer"
    }

$downloadUrl = $null

foreach ($asset in $release.assets) {
    if ($asset.name -eq $binaryName) {
        $downloadUrl = $asset.browser_download_url
        break
    }
}

if ($null -eq $downloadUrl) {
    Write-Error "Could not find asset $binaryName in the latest release."
    exit 1
}

# --------------------------------------------------
# Installation directory
# --------------------------------------------------

$installDir = Join-Path $env:LOCALAPPDATA "LanShare"
$installPath = Join-Path $installDir "lanshare.exe"

if (-not (Test-Path $installDir)) {
    New-Item `
        -ItemType Directory `
        -Force `
        -Path $installDir | Out-Null
}

# --------------------------------------------------
# Stop currently running LanShare
# --------------------------------------------------

Write-Host "Stopping existing LanShare instance..."

Get-Process -Name "lanshare" -ErrorAction SilentlyContinue |
    Stop-Process -Force -ErrorAction SilentlyContinue

# --------------------------------------------------
# Download
# --------------------------------------------------

$tempPath = Join-Path $env:TEMP "lanshare-download.exe"

Write-Host "Downloading $binaryName..."

Invoke-WebRequest `
    -Uri $downloadUrl `
    -OutFile $tempPath

# Replace installed binary
Move-Item `
    -Path $tempPath `
    -Destination $installPath `
    -Force

Write-Host "Installed:"
Write-Host "  $installPath"

# --------------------------------------------------
# Remove old Startup shortcut
# --------------------------------------------------

$startupDir = Join-Path `
    $env:APPDATA `
    "Microsoft\Windows\Start Menu\Programs\Startup"

$oldShortcut = Join-Path $startupDir "LanShare.lnk"

if (Test-Path $oldShortcut) {
    Write-Host "Removing old Startup shortcut..."
    Remove-Item $oldShortcut -Force
}

# --------------------------------------------------
# Create Scheduled Task
# --------------------------------------------------

Write-Host "Configuring LanShare background task..."

# Remove previous task if it exists
$existingTask = Get-ScheduledTask `
    -TaskName $taskName `
    -ErrorAction SilentlyContinue

if ($null -ne $existingTask) {
    Unregister-ScheduledTask `
        -TaskName $taskName `
        -Confirm:$false
}

$currentUser = [System.Security.Principal.WindowsIdentity]::GetCurrent().Name

# Run LanShare
$action = New-ScheduledTaskAction `
    -Execute $installPath `
    -WorkingDirectory $installDir

# Start when this user logs in
$trigger = New-ScheduledTaskTrigger `
    -AtLogOn `
    -User $currentUser

# Run as current logged-in user
$principal = New-ScheduledTaskPrincipal `
    -UserId $currentUser `
    -LogonType Interactive `
    -RunLevel Limited

# Service-like restart behavior
$settings = New-ScheduledTaskSettingsSet `
    -RestartCount 10 `
    -RestartInterval (New-TimeSpan -Minutes 1) `
    -StartWhenAvailable `
    -ExecutionTimeLimit ([TimeSpan]::Zero) `
    -MultipleInstances IgnoreNew

$task = New-ScheduledTask `
    -Action $action `
    -Trigger $trigger `
    -Principal $principal `
    -Settings $settings `
    -Description "LanShare LAN file sharing background service"

Register-ScheduledTask `
    -TaskName $taskName `
    -InputObject $task `
    -Force | Out-Null

# --------------------------------------------------
# Start immediately
# --------------------------------------------------

Write-Host "Starting LanShare..."

Start-ScheduledTask -TaskName $taskName

# --------------------------------------------------
# Done
# --------------------------------------------------

Write-Host ""
Write-Host "--------------------------------------"
Write-Host "LanShare installed successfully"
Write-Host "--------------------------------------"
Write-Host ""
Write-Host "Binary:"
Write-Host "  $installPath"
Write-Host ""
Write-Host "Scheduled Task:"
Write-Host "  $taskName"
Write-Host ""
Write-Host "Useful commands:"
Write-Host ""
Write-Host '  Get-ScheduledTask -TaskName "LanShare"'
Write-Host '  Start-ScheduledTask -TaskName "LanShare"'
Write-Host '  Stop-ScheduledTask -TaskName "LanShare"'
Write-Host '  Unregister-ScheduledTask -TaskName "LanShare"'