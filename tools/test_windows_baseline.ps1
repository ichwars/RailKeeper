param(
  [Parameter(Mandatory = $true)]
  [string]$PackageDir
)

$ErrorActionPreference = "Stop"
$repo = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot ".."))
$package = [IO.Path]::GetFullPath($PackageDir)
$regression = Join-Path $repo "tools\regression"
$testRoot = Join-Path $repo ("output\windows-baseline-" + [Guid]::NewGuid().ToString("N"))
$previousSource = Join-Path $testRoot "previous-source"
$previousBinary = Join-Path $testRoot "railkeeper-v0.1.22.exe"
$archive = Join-Path $testRoot "previous-source.tar"
$baselineCommit = "debfad6198299052f07220883c1476581be691cc"

if ($env:OS -ne "Windows_NT") {
  throw "Native Windows baseline requires Windows."
}
if (-not (Test-Path -LiteralPath (Join-Path $package "RailKeeper.exe"))) {
  throw "Packaged RailKeeper.exe is missing."
}
New-Item -ItemType Directory -Path $previousSource -Force | Out-Null

# Fetch only the pinned public baseline source. The current checkout is never reset.
& git -C $repo fetch --no-tags --depth=1 origin $baselineCommit
if ($LASTEXITCODE -ne 0) { throw "Could not fetch the pinned baseline source." }
& git -C $repo archive --format=tar --output=$archive $baselineCommit backend
if ($LASTEXITCODE -ne 0) { throw "Could not archive the pinned baseline source." }
& tar -xf $archive -C $previousSource
if ($LASTEXITCODE -ne 0) { throw "Could not extract the pinned baseline source." }

$previousCache = $env:GOCACHE
Push-Location (Join-Path $previousSource "backend")
try {
  $env:GOCACHE = Join-Path $repo ".cache\go-build"
  New-Item -ItemType Directory -Path $env:GOCACHE -Force | Out-Null
  & go build -buildvcs=false -o $previousBinary ./cmd/railkeeper
  if ($LASTEXITCODE -ne 0) { throw "Previous baseline binary build failed." }
} finally {
  $env:GOCACHE = $previousCache
  Pop-Location
}

& npm.cmd --prefix $regression ci
if ($LASTEXITCODE -ne 0) { throw "Browser test dependency installation failed." }
& node (Join-Path $regression "node_modules\playwright\cli.js") install chromium
if ($LASTEXITCODE -ne 0) { throw "Chromium installation failed." }
& python (Join-Path $regression "matrix.py")
if ($LASTEXITCODE -ne 0) { throw "Feature matrix check failed." }
& python -m unittest discover -s $regression -p "test_*.py"
if ($LASTEXITCODE -ne 0) { throw "Preservation comparison tests failed." }

& python (Join-Path $regression "run_baseline.py") `
  --binary (Join-Path $package "RailKeeper.exe") `
  --backend $package --static-dir (Join-Path $package "web") `
  --previous-binary $previousBinary --previous-backend (Join-Path $previousSource "backend") `
  --catalog-backfill --catalog-authority --catalog-writes --decoder-model --decoder-backfill --browser --output (Join-Path $testRoot "run")
if ($LASTEXITCODE -ne 0) { throw "Native Windows upgrade/restore/browser baseline failed." }
Write-Host "Native Windows baseline passed. Report: $(Join-Path $testRoot 'run\report.json')"
