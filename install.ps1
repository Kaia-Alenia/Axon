$ErrorActionPreference = "Stop"

Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║         Axon Server - Global Installation Script          ║" -ForegroundColor Cyan
Write-Host "║                  Wireless Input Bridge                     ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

$Repo = "Kaia-Alenia/Axon"
$ReleaseUrl = "https://github.com/$Repo/releases/latest/download"


$Arch = $env:PROCESSOR_ARCHITECTURE.ToLower()
$ArchName = ""

if ($Arch -eq "amd64" -or $Arch -eq "x86_64") {
    $ArchName = "amd64"
} elseif ($Arch -eq "arm64") {
    $ArchName = "arm64"
} elseif ($Arch -eq "x86") {
    $ArchName = "386"
} else {
    Write-Host "[ERROR] Unsupported architecture: $Arch" -ForegroundColor Red
    exit 1
}

Write-Host "Detected platform: windows-$ArchName" -ForegroundColor Cyan

$BinaryName = "axon-windows-${ArchName}.exe"
$DownloadUrl = "${ReleaseUrl}/${BinaryName}"
$InstallDir = "$env:ProgramFiles\Axon"
$InstallPath = "$InstallDir\axon.exe"

Write-Host "[DOWNLOAD] Downloading from: $DownloadUrl" -ForegroundColor Cyan

$TempFile = [System.IO.Path]::GetTempFileName()
Try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempFile -UseBasicParsing
} Catch {
    Write-Host "[ERROR] Failed to download binary" -ForegroundColor Red
    exit 1
}

Write-Host "[SUCCESS] Downloaded successfully" -ForegroundColor Green
Write-Host "[INSTALL] Installing to: $InstallPath" -ForegroundColor Cyan

Get-Process -Name "axon" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue

if (Test-Path -Path $InstallDir) {
    Remove-Item -Path "$InstallDir\*" -Force -ErrorAction SilentlyContinue
} else {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

Copy-Item -Path $TempFile -Destination $InstallPath -Force
Remove-Item -Path $TempFile -Force

Remove-NetFirewallRule -DisplayName "Axon Server" -ErrorAction SilentlyContinue
New-NetFirewallRule -DisplayName "Axon Server" -Direction Inbound -Program $InstallPath -Action Allow -Profile Any -ErrorAction SilentlyContinue | Out-Null


$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notmatch [regex]::Escape($InstallDir)) {
    $NewPath = $UserPath + ";$InstallDir"
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    Write-Host "[WARNING] Added $InstallDir to PATH. You may need to restart your terminal." -ForegroundColor Yellow
}

Write-Host "[SUCCESS] Installation successful!" -ForegroundColor Green
Write-Host ""
Write-Host "[DONE] Installation complete!" -ForegroundColor Green
Write-Host "Run 'axon --help' to see available options" -ForegroundColor Cyan
Write-Host "Run 'axon --version' to verify the installation" -ForegroundColor Cyan
