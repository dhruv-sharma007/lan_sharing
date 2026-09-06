$ErrorActionPreference = "Continue"

Write-Host "Uninstalling LanShare..."

# Stop the running process if it exists
$process = Get-Process lanshare -ErrorAction SilentlyContinue
if ($process) {
    Write-Host "Stopping LanShare process..."
    Stop-Process -Name lanshare -Force
}

# Remove binary directory
$installDir = Join-Path $env:LOCALAPPDATA "LanShare"
if (Test-Path $installDir) {
    Write-Host "Removing installation directory at $installDir..."
    Remove-Item -Path $installDir -Recurse -Force
}

# Remove startup shortcut
$startupDir = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\Startup"
$shortcutPath = Join-Path $startupDir "LanShare.lnk"
if (Test-Path $shortcutPath) {
    Write-Host "Removing startup shortcut at $shortcutPath..."
    Remove-Item -Path $shortcutPath -Force
}

Write-Host "LanShare has been completely uninstalled!"
