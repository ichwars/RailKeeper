param(
  [string]$Version = "",
  [string]$SourceRoot = (Join-Path $PSScriptRoot "..")
)

$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "windows_package_validation.ps1")
. (Join-Path $PSScriptRoot "windows_build_identity.ps1")
. (Join-Path $PSScriptRoot "windows_package_licenses.ps1")
. (Join-Path $PSScriptRoot "windows_start_shortcut.ps1")

$repo = Resolve-Path $SourceRoot
$frontend = Join-Path $repo "frontend"
$backend = Join-Path $repo "backend"
$packageRoot = Join-Path $repo "dist\windows-standalone"
$packageDir = Join-Path $packageRoot "RailKeeper-Windows-Standalone"
$brandIconSource = Join-Path $repo "frontend\public\brand\railkeeper-mark.png"

$main = Get-Content (Join-Path $backend "cmd\railkeeper\main.go") -Raw
if ($main -notmatch '(?m)^\s*version\s*=\s*"([0-9]+(\.[0-9]+){2,3}(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?)"') {
  throw "Could not read RailKeeper version from backend/cmd/railkeeper/main.go"
}
$sourceVersion = $Matches[1]
if ($env:GITHUB_ACTIONS -eq "true" -and $env:GITHUB_REF_TYPE -eq "tag" -and
    $env:GITHUB_REF_NAME -cne "v$sourceVersion") {
  throw "Windows release tag $env:GITHUB_REF_NAME differs from runtime version $sourceVersion."
}
if ($Version -and $Version -cne $sourceVersion) {
  throw "Requested package version $Version differs from runtime version $sourceVersion."
}
$Version = $sourceVersion
$npm = if ($IsWindows) { "npm.cmd" } else { "npm" }

$railKeeperBuild = Get-RailKeeperBuildIdentity -Repo $repo -Version $Version

$zipPath = Join-Path $packageRoot ("RailKeeper-windows-x64-v{0}.zip" -f $Version)

$resolvedPackageRoot = [IO.Path]::GetFullPath($packageRoot)
$resolvedPackageDir = [IO.Path]::GetFullPath($packageDir)
$resolvedZipPath = [IO.Path]::GetFullPath($zipPath)
if (-not $resolvedPackageDir.StartsWith(
    $resolvedPackageRoot + [IO.Path]::DirectorySeparatorChar,
    [StringComparison]::OrdinalIgnoreCase)) {
  throw "Package directory escaped package root: $resolvedPackageDir"
}
if (-not $resolvedZipPath.StartsWith(
    $resolvedPackageRoot + [IO.Path]::DirectorySeparatorChar,
    [StringComparison]::OrdinalIgnoreCase)) {
  throw "Package archive escaped package root: $resolvedZipPath"
}

if (Test-Path -LiteralPath $resolvedPackageDir) {
  Remove-Item -LiteralPath $packageDir -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $packageDir | Out-Null

Push-Location $frontend
try {
  & $npm run build
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
foreach ($license in @("LICENSE.md", "THIRD_PARTY_NOTICES.md")) {
  Copy-Item -LiteralPath (Join-Path $repo $license) -Destination $packageDir
}

$resourceTemp = [IO.Directory]::CreateTempSubdirectory("railkeeper-resources-").FullName

$previousGoos = $env:GOOS
$previousGoarch = $env:GOARCH
$previousCgo = $env:CGO_ENABLED
$previousGocache = $env:GOCACHE
Push-Location $backend
try {
  $unversionedExe = Join-Path $resourceTemp "RailKeeper.exe"
  $exePath = Join-Path $packageDir "RailKeeper.exe"
  $env:GOOS = "windows"
  $env:GOARCH = "amd64"
  $env:GOCACHE = Join-Path $repo ".cache\go-build"
  New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null
  if ($env:GITHUB_ACTIONS -eq "true" -and $env:OS -eq "Windows_NT") {
    & go test ./... -count=1 -timeout=20m
    if ($LASTEXITCODE -ne 0) {
      throw "Native Windows backend qualification failed with exit code $LASTEXITCODE"
    }
  }
  $env:CGO_ENABLED = "0"
  $buildFlags = "-s -w -X railkeeper/backend/internal/buildinfo.commit=$($railKeeperBuild.Commit) -X railkeeper/backend/internal/buildinfo.kind=$($railKeeperBuild.Kind) -X railkeeper/backend/internal/buildinfo.modified=$($railKeeperBuild.Modified)"
  & go build -mod=readonly -buildvcs=false -trimpath -ldflags $buildFlags -o $unversionedExe ./cmd/railkeeper
  if ($LASTEXITCODE -ne 0) {
    throw "Windows backend build failed with exit code $LASTEXITCODE"
  }
  Write-RailKeeperPackageLicenses -Frontend $frontend -Npm $npm -PackageDir $packageDir

  $env:GOOS = $null
  $env:GOARCH = $null
  & go -C (Join-Path $PSScriptRoot "windows-resources") run -mod=readonly . $Version $brandIconSource $unversionedExe $exePath
  if ($LASTEXITCODE -ne 0) { throw "Could not write Windows icon and version resources." }
} finally {
  $env:GOOS = $previousGoos
  $env:GOARCH = $previousGoarch
  $env:CGO_ENABLED = $previousCgo
  $env:GOCACHE = $previousGocache
  Pop-Location
  Remove-Item -LiteralPath $resourceTemp -Recurse -Force
}

New-RailKeeperStartShortcut -PackageDir $packageDir
Assert-RailKeeperPackageDirectory -PackageDir $packageDir

if (Test-Path -LiteralPath $zipPath) {
  Remove-Item -LiteralPath $zipPath -Force
}
Compress-Archive -Path (Join-Path $packageDir "*") -DestinationPath $zipPath -Force
Assert-RailKeeperPackageArchive -ZipPath $zipPath -PackageDir $packageDir

if ($IsWindows) {
  $extractionRoot = [IO.Directory]::CreateTempSubdirectory("railkeeper-zip-").FullName
  $extractedPackage = Join-Path $extractionRoot "Moved RailKeeper package"
  try {
    Expand-Archive -LiteralPath $zipPath -DestinationPath $extractedPackage
    Assert-RailKeeperExecutableVersion -ExePath (Join-Path $extractedPackage "RailKeeper.exe") -Version $Version
    if ($env:GITHUB_ACTIONS -eq "true") {
      & python -u (Join-Path $PSScriptRoot "test_windows_launcher.py") $extractedPackage | Write-Host
      if ($LASTEXITCODE -ne 0) { throw "Native Windows launcher qualification failed." }
      & (Join-Path $PSScriptRoot "test_windows_baseline.ps1") -PackageDir $extractedPackage -SourceRoot $repo
    }
  } finally {
    Remove-Item -LiteralPath $extractionRoot -Recurse -Force
  }
}

Write-Host "Windows Standalone package created:"
Write-Host $zipPath
