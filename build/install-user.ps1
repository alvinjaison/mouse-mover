# Mouse Mover Per-User Installer
# Does NOT require administrator privileges.
# Usage: powershell -ExecutionPolicy Bypass -File install-user.ps1

$ErrorActionPreference = "Stop"

$installDir = Join-Path $env:LOCALAPPDATA "MouseMover"
$exeName = "MouseMover.exe"
$startMenu = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs"
$shortcutName = "Mouse Mover.lnk"

# Find the exe (look in same dir as this script, or in dist/)
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$exeSrc = Join-Path $scriptDir $exeName
if (-not (Test-Path $exeSrc)) {
    $exeSrc = Join-Path $scriptDir "..\dist\$exeName"
}
if (-not (Test-Path $exeSrc)) {
    Write-Error "Cannot find $exeName. Place this script next to the exe or run from the project root."
    exit 1
}

# Create install directory.
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}

# Copy exe.
Copy-Item $exeSrc (Join-Path $installDir $exeName) -Force
Write-Host "Installed: $installDir\$exeName"

# Create Start Menu shortcut.
$ws = New-Object -ComObject WScript.Shell
$sc = $ws.CreateShortcut((Join-Path $startMenu $shortcutName))
$sc.TargetPath = Join-Path $installDir $exeName
$sc.WorkingDirectory = $installDir
$sc.Description = "Mouse Mover — Keep your session active"
$sc.Save()
Write-Host "Shortcut: $startMenu\$shortcutName"

Write-Host ""
Write-Host "Done! To uninstall:"
Write-Host "  1. Delete $installDir"
Write-Host "  2. Delete $startMenu\$shortcutName"
