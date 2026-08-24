$ErrorActionPreference = "Stop"

$repoOwner = "dhruv-sharma007"
$repoName = "lan_sharing"

Write-Host "Installing LanShare..."

# Detect OS and Arch
$os = "windows"
$arch = "amd64" # We only support amd64 for Windows currently

$binaryName = "lanshare-$os-$arch.exe"
$releasesUrl = "https://api.github.com/repos/$repoOwner/$repoName/releases/latest"

Write-Host "Fetching latest release..."
$release = Invoke-RestMethod -Uri $releasesUrl -Headers @{ "Accept" = "application/vnd.github.v3+json" }

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

$installDir = Join-Path $env:LOCALAPPDATA "LanShare"
$installPath = Join-Path $installDir "lanshare.exe"

if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
}

Write-Host "Downloading $binaryName..."
Invoke-WebRequest -Uri $downloadUrl -OutFile $installPath

Write-Host "Configuring startup shortcut..."
$startupDir = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\Startup"
$shortcutPath = Join-Path $startupDir "LanShare.lnk"

$wshShell = New-Object -ComObject WScript.Shell
$shortcut = $wshShell.CreateShortcut($shortcutPath)
$shortcut.TargetPath = $installPath
$shortcut.WorkingDirectory = $installDir
$shortcut.WindowStyle = 7 # Minimized
$shortcut.Save()

Write-Host "Starting LanShare..."
Start-Process -FilePath $installPath -WindowStyle Hidden

Write-Host "LanShare installed successfully!"
