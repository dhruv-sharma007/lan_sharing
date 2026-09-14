$ErrorActionPreference = "Continue"

$appName = "LanShare"
$taskName = "LanShare"

Write-Host "Uninstalling $appName..."

# Remove the Scheduled Task created by the current installer before
# terminating the executable, so task restart settings cannot relaunch it.
$existingTask = Get-ScheduledTask `
    -TaskName $taskName `
    -ErrorAction SilentlyContinue

if ($null -ne $existingTask) {
    Write-Host "Stopping and removing Scheduled Task $taskName..."

    Stop-ScheduledTask `
        -TaskName $taskName `
        -ErrorAction SilentlyContinue

    Unregister-ScheduledTask `
        -TaskName $taskName `
        -Confirm:$false `
        -ErrorAction SilentlyContinue
}

# Stop a process left over from a legacy installation or failed task teardown.
$process = Get-Process `
    -Name "lanshare" `
    -ErrorAction SilentlyContinue

if ($null -ne $process) {
    Write-Host "Stopping LanShare process..."
    Stop-Process `
        -Name "lanshare" `
        -Force `
        -ErrorAction SilentlyContinue
}

# Remove the legacy Startup shortcut used by earlier installers.
$startupDir = Join-Path `
    $env:APPDATA `
    "Microsoft\Windows\Start Menu\Programs\Startup"

$shortcutPath = Join-Path $startupDir "LanShare.lnk"

if (Test-Path $shortcutPath) {
    Write-Host "Removing legacy Startup shortcut at $shortcutPath..."
    Remove-Item `
        -Path $shortcutPath `
        -Force
}

$installDir = Join-Path $env:LOCALAPPDATA "LanShare"

if (Test-Path $installDir) {
    Write-Host "Removing installation directory at $installDir..."
    Remove-Item `
        -Path $installDir `
        -Recurse `
        -Force
}

Write-Host "$appName has been completely uninstalled!"
