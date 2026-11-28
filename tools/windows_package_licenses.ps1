function Get-RailKeeperComponentNotices {
  param([string]$Component, [string]$Directory)

  $notices = @(Get-ChildItem -LiteralPath $Directory -File | Where-Object {
    $_.Name -match '^(licen[cs]e|copying|copyright|notice)([.-].*)?$'
  } | Sort-Object Name)
  if (-not $notices) {
    throw "No license text found for $Component. Review the package before redistributing it."
  }
  foreach ($notice in $notices) {
    $text = [IO.File]::ReadAllText($notice.FullName).TrimEnd()
    if (-not $text) { throw "Empty license notice for $Component in $($notice.Name)." }
    "===== $Component / $($notice.Name) =====`n`n$text"
  }
}

function Write-RailKeeperPackageLicenses {
  param([string]$Frontend, [string]$Npm, [string]$PackageDir)

  $goRoot = & go env GOROOT
  if ($LASTEXITCODE -ne 0) { throw "Could not locate the Go standard library." }
  $goVersion = & go env GOVERSION
  if ($LASTEXITCODE -ne 0) { throw "Could not identify the Go standard library version." }
  $modules = & go list -mod=readonly -deps -f '{{with .Module}}{{if not .Main}}{{.Path}} {{.Version}}|{{.Dir}}{{end}}{{end}}' ./cmd/railkeeper ./cmd/railkeeper-launcher
  if ($LASTEXITCODE -ne 0) { throw "Could not identify Go modules for the Windows executables." }

  Push-Location $Frontend
  try {
    $npmDirectories = @(& $Npm ls --omit=dev --all --parseable)
    if ($LASTEXITCODE -ne 0) { throw "Could not identify frontend production packages." }
  } finally {
    Pop-Location
  }
  # The bundle may also contain runtime helpers emitted by the build tools.
  $npmDirectories += Join-Path $Frontend "node_modules/vite"
  $npmDirectories += Join-Path $Frontend "node_modules/rolldown"

  $sections = @(
    Get-RailKeeperComponentNotices -Component "Go standard library $goVersion" -Directory $goRoot
    foreach ($module in ($modules | Where-Object { $_ } | Sort-Object -Unique)) {
      $name, $directory = $module -split '\|', 2
      Get-RailKeeperComponentNotices -Component $name -Directory $directory
      if ($name -like 'github.com/wailsapp/wails/webview2 *') {
        Get-RailKeeperComponentNotices -Component "$name (WebView2 loader)" -Directory (Join-Path $directory "webviewloader")
      }
    }
    foreach ($directory in ($npmDirectories | Where-Object { $_ -and $_ -ne $Frontend } | Sort-Object -Unique)) {
      $manifest = Get-Content -LiteralPath (Join-Path $directory "package.json") -Raw | ConvertFrom-Json
      Get-RailKeeperComponentNotices -Component "npm $($manifest.name) $($manifest.version)" -Directory $directory
    }
  )
  $header = "Third-party licenses for the RailKeeper Windows Standalone package.`n" +
    "See web/fonts/LICENSE-Selawik.txt for the bundled Selawik font."
  ($header + "`n`n" + (($sections | Sort-Object -Unique -CaseSensitive) -join "`n`n")) |
    Set-Content -LiteralPath (Join-Path $PackageDir "THIRD_PARTY_LICENSES.txt") -Encoding utf8
}
