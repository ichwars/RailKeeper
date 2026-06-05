param(
  [string]$Version = ""
)

$ErrorActionPreference = "Stop"

$repo = Resolve-Path (Join-Path $PSScriptRoot "..")
$frontend = Join-Path $repo "frontend"
$backend = Join-Path $repo "backend"
$packageRoot = Join-Path $repo "dist\windows-portable"
$packageDir = Join-Path $packageRoot "RailKeeper-Portable"

if ([string]::IsNullOrWhiteSpace($Version)) {
  $main = Get-Content (Join-Path $backend "cmd\railkeeper\main.go") -Raw
  if ($main -notmatch 'version\s+=\s+"([^"]+)"') {
    throw "Could not read RailKeeper version from backend/cmd/railkeeper/main.go"
  }
  $Version = $Matches[1]
}

$zipPath = Join-Path $packageRoot ("RailKeeper-windows-x64-v{0}.zip" -f $Version)

if (Test-Path $packageDir) {
  Remove-Item -LiteralPath $packageDir -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $packageDir | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $packageDir "data") | Out-Null

Push-Location $frontend
try {
  & npm.cmd run build
  if ($LASTEXITCODE -ne 0) {
    throw "Frontend build failed with exit code $LASTEXITCODE"
  }
} finally {
  Pop-Location
}

Copy-Item -Path (Join-Path $frontend "dist") -Destination (Join-Path $packageDir "web") -Recurse
Copy-Item -Path (Join-Path $backend "migrations") -Destination (Join-Path $packageDir "migrations") -Recurse
Copy-Item -Path (Join-Path $backend "seeds") -Destination (Join-Path $packageDir "seeds") -Recurse
Copy-Item -Path (Join-Path $repo "deploy\windows\start-railkeeper.bat") -Destination (Join-Path $packageDir "start-railkeeper.bat")
Copy-Item -Path (Join-Path $repo "deploy\windows\README-Windows.txt") -Destination (Join-Path $packageDir "README-Windows.txt")

Push-Location $backend
try {
  $previousGoos = $env:GOOS
  $previousGoarch = $env:GOARCH
  $previousGocache = $env:GOCACHE
  $env:GOOS = "windows"
  $env:GOARCH = "amd64"
  $env:GOCACHE = Join-Path $repo ".cache\go-build"
  New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null
  & go build -trimpath -ldflags "-s -w" -o (Join-Path $packageDir "RailKeeper.exe") ./cmd/railkeeper
  if ($LASTEXITCODE -ne 0) {
    throw "Windows backend build failed with exit code $LASTEXITCODE"
  }
} finally {
  $env:GOOS = $previousGoos
  $env:GOARCH = $previousGoarch
  $env:GOCACHE = $previousGocache
  Pop-Location
}

if (Test-Path $zipPath) {
  Remove-Item -LiteralPath $zipPath -Force
}
Compress-Archive -Path (Join-Path $packageDir "*") -DestinationPath $zipPath -Force

Write-Host "Portable package created:"
Write-Host $zipPath
