# gita-cli Windows Installer
# Usage: irm https://raw.githubusercontent.com/ACS-lessgo/gita-cli/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo    = "ACS-lessgo/gita-cli"
$binary  = "gita-windows-amd64.exe"
$destDir = "$env:USERPROFILE\.local\bin"
$destExe = "$destDir\gita.exe"
$url     = "https://github.com/$repo/releases/latest/download/$binary"

Write-Host ""
Write-Host "Downloading gita-cli for windows/amd64..." -ForegroundColor Cyan
Write-Host "URL: $url"

# ── Create install dir ────────────────────────────────────────────────────────
New-Item -ItemType Directory -Force -Path $destDir | Out-Null

# ── Download ──────────────────────────────────────────────────────────────────
try {
    Invoke-WebRequest -Uri $url -OutFile $destExe -UseBasicParsing
} catch {
    Write-Host ""
    Write-Host "ERROR: Download failed." -ForegroundColor Red
    Write-Host "Please download manually from:"
    Write-Host "https://github.com/$repo/releases/latest" -ForegroundColor Yellow
    exit 1
}

# ── Add to user PATH if not already present ───────────────────────────────────
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($currentPath -notlike "*$destDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$destDir", "User")
    Write-Host ""
    Write-Host "Added $destDir to your PATH." -ForegroundColor Green
    Write-Host "NOTE: Open a new terminal for PATH to take effect." -ForegroundColor Yellow
} else {
    Write-Host "$destDir already in PATH." -ForegroundColor Green
}

Write-Host ""
Write-Host "✓ gita installed → $destExe" -ForegroundColor Green
Write-Host ""
Write-Host "Open a new terminal and run:  gita" -ForegroundColor Cyan
Write-Host ""