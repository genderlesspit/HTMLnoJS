# dependencies.ps1 - Install Go if it doesn't exist

$ErrorActionPreference = "Stop"

function Test-GoInstalled {
    try {
        $version = go version 2>$null
        if ($version -match "go version go(\d+\.\d+\.\d+)") {
            Write-Host "Go is already installed: $version" -ForegroundColor Green
            return $true
        }
    }
    catch {
        return $false
    }
    return $false
}

function Install-Go {
    Write-Host "Go not found. Installing Go..." -ForegroundColor Yellow

    # Get latest Go version from GitHub API
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/golang/go/tags" -UseBasicParsing
        $latestVersion = ($response | Where-Object { $_.name -match "^go\d+\.\d+\.\d+$" } | Select-Object -First 1).name
        $version = $latestVersion -replace "go", ""
    }
    catch {
        # Fallback to a known stable version
        $version = "1.21.5"
        Write-Warning "Could not fetch latest version, using Go $version"
    }

    $downloadUrl = "https://go.dev/dl/go$version.windows-amd64.msi"
    $tempFile = "$env:TEMP\go$version.windows-amd64.msi"

    Write-Host "Downloading Go $version..." -ForegroundColor Blue

    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing

        Write-Host "Installing Go $version..." -ForegroundColor Blue
        Start-Process -FilePath $tempFile -ArgumentList "/quiet" -Wait

        # Add Go to PATH if not already there
        $goPath = "C:\Program Files\Go\bin"
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")

        if ($currentPath -notlike "*$goPath*") {
            Write-Host "Adding Go to system PATH..." -ForegroundColor Blue
            [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$goPath", "Machine")
            $env:PATH += ";$goPath"
        }

        # Cleanup
        Remove-Item $tempFile -ErrorAction SilentlyContinue

        Write-Host "Go installation completed!" -ForegroundColor Green
        Write-Host "Please restart your terminal or run: refreshenv" -ForegroundColor Yellow

    }
    catch {
        Write-Error "Failed to install Go: $_"
        exit 1
    }
}

# Main execution
Write-Host "Checking Go installation..." -ForegroundColor Cyan

if (-not (Test-GoInstalled)) {
    Install-Go
} else {
    Write-Host "All dependencies satisfied!" -ForegroundColor Green
}

# Verify installation
Write-Host "`nVerifying installation..." -ForegroundColor Cyan
try {
    $version = go version
    Write-Host "Success: $version" -ForegroundColor Green
}
catch {
    Write-Warning "Go may not be in PATH yet. Try restarting your terminal."
}