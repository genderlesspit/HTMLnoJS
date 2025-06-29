param (
    [string]$Project     = ".",
    [int]   $Port        = 3000,
    [int]   $FastAPIPort = 3001
)

$ErrorActionPreference = "Stop"

Write-Host "Checking Go installation..." -ForegroundColor Cyan

function Test-GoInstalled {
    try {
        $version = go version 2>$null
        if ($version -match "go version go(\d+\.\d+\.\d+)") {
            Write-Host "Go is already installed: $version" -ForegroundColor Green
            return $true
        }
    } catch {
        return $false
    }
    return $false
}

function Install-Go {
    Write-Host "Installing Go..." -ForegroundColor Yellow

    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/golang/go/tags" -UseBasicParsing
        $latestVersion = ($response | Where-Object { $_.name -match "^go\d+\.\d+\.\d+$" } | Select-Object -First 1).name
        $version = $latestVersion -replace "go", ""
    } catch {
        $version = "1.21.5"
        Write-Warning "Could not fetch latest version, falling back to Go $version"
    }

    $downloadUrl = "https://go.dev/dl/go$version.windows-amd64.msi"
    $tempFile = "$env:TEMP\go$version.windows-amd64.msi"

    try {
        Write-Host "Downloading Go $version..." -ForegroundColor Blue
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing

        Write-Host "Installing MSI..." -ForegroundColor Blue
        Start-Process -FilePath $tempFile -ArgumentList "/quiet" -Wait

        $goPath = "C:\Program Files\Go\bin"
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")

        if ($currentPath -notlike "*$goPath*") {
            Write-Host "Updating PATH..." -ForegroundColor Blue
            [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$goPath", "Machine")
            $env:PATH += ";$goPath"
        }

        Remove-Item $tempFile -ErrorAction SilentlyContinue
        Write-Host "Go installed successfully" -ForegroundColor Green
    } catch {
        Write-Error "Failed to install Go: $_"
        exit 1
    }
}

if (-not (Test-GoInstalled)) {
    Install-Go
}

# Re-check after install
try {
    $version = go version
    Write-Host "Go is ready: $version" -ForegroundColor Green
} catch {
    Write-Error "Go is not available in PATH. Try restarting your terminal."
    exit 1
}

# Find main.go - look in go-server directory relative to script location
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$parentDir = Split-Path -Parent $scriptDir
$goServerDir = Join-Path $parentDir "go-server"
$mainGo = Join-Path $goServerDir "main.go"

# Fallback to project directory if go-server structure not found
if (-not (Test-Path $mainGo)) {
    $mainGo = Join-Path $Project "main.go"
    $goServerDir = $Project
}

if (-not (Test-Path $mainGo)) {
    Write-Host "main.go not found at: $mainGo" -ForegroundColor Red
    Write-Host "Tried locations:" -ForegroundColor Yellow
    Write-Host "  - $goServerDir/main.go" -ForegroundColor Yellow
    Write-Host "  - $Project/main.go" -ForegroundColor Yellow
    exit 1
}

Write-Host "Found main.go at: $mainGo" -ForegroundColor Green

Push-Location $goServerDir
try {
    Write-Host "Running: go run main.go -directory `"$Project`" -port $Port -fastapi-port $FastAPIPort" -ForegroundColor Yellow
    go run main.go `
        -directory "$Project" `
        -port $Port `
        -fastapi-port $FastAPIPort
} catch {
    Write-Error "Failed to start Go server: $_"
    exit 1
} finally {
    Pop-Location
}