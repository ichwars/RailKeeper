$ErrorActionPreference = "Stop"

$validationScript = Join-Path $PSScriptRoot "windows_package_validation.ps1"
. $validationScript

function New-GoodPackageTree {
  param([string]$Path)

  New-Item -ItemType Directory -Force -Path $Path | Out-Null
  foreach ($directory in @("web", "migrations", "seeds")) {
    New-Item -ItemType Directory -Force -Path (Join-Path $Path $directory) | Out-Null
  }
  Set-Content -LiteralPath (Join-Path $Path "RailKeeper.exe") -Value "exe"
  Set-Content -LiteralPath (Join-Path $Path "web\index.html") -Value "web"
  Set-Content -LiteralPath (Join-Path $Path "migrations\0001.sql") -Value "migration"
  Set-Content -LiteralPath (Join-Path $Path "seeds\master_data.json") -Value "seed"
  Set-Content -LiteralPath (Join-Path $Path "start-railkeeper.bat") -Value "start"
  Set-Content -LiteralPath (Join-Path $Path "README-Windows.txt") -Value "readme"
}

function Assert-ValidationRejectsPath {
  param(
    [scriptblock]$Action,
    [string]$ExpectedPath
  )

  try {
    & $Action
  } catch {
    $message = $_.Exception.Message.Replace("/", "\")
    if (-not $message.Contains($ExpectedPath.Replace("/", "\"))) {
      throw "Validation failed without naming '$ExpectedPath': $message"
    }
    return
  }
  throw "Validation accepted forbidden package path: $ExpectedPath"
}

$temporaryParent = [IO.Path]::GetFullPath([IO.Path]::GetTempPath())
$temporaryRoot = Join-Path $temporaryParent ("railkeeper-package-validation-" + [Guid]::NewGuid().ToString("N"))
$resolvedRoot = [IO.Path]::GetFullPath($temporaryRoot)
if (-not $resolvedRoot.StartsWith($temporaryParent, [StringComparison]::OrdinalIgnoreCase)) {
  throw "Temporary test root escaped the system temporary directory: $resolvedRoot"
}

try {
  New-Item -ItemType Directory -Path $resolvedRoot | Out-Null
  $goodPackage = Join-Path $resolvedRoot "good\RailKeeper-Windows-Standalone"
  New-GoodPackageTree -Path $goodPackage
  Assert-RailKeeperPackageDirectory -PackageDir $goodPackage

  $goodZip = Join-Path $resolvedRoot "good.zip"
  Compress-Archive -Path (Join-Path $goodPackage "*") -DestinationPath $goodZip
  Assert-RailKeeperPackageArchive -ZipPath $goodZip

  $badCases = @(
    @{ RelativePath = "data\user.txt"; ExpectedPath = "data" },
    @{ RelativePath = "railkeeper.db"; ExpectedPath = "railkeeper.db" },
    @{ RelativePath = "railkeeper.db-wal"; ExpectedPath = "railkeeper.db-wal" },
    @{ RelativePath = "uploads\manual.pdf"; ExpectedPath = "uploads" },
    @{ RelativePath = "backups\backup.zip"; ExpectedPath = "backups" }
  )
  foreach ($badCase in $badCases) {
    $relativePath = $badCase.RelativePath
    $expectedPath = $badCase.ExpectedPath
    $caseID = [Convert]::ToHexString([Text.Encoding]::UTF8.GetBytes($relativePath))
    $badPackage = Join-Path $resolvedRoot ("bad-" + $caseID)
    Copy-Item -LiteralPath $goodPackage -Destination $badPackage -Recurse
    $forbiddenPath = Join-Path $badPackage $relativePath
    New-Item -ItemType Directory -Force -Path (Split-Path $forbiddenPath -Parent) | Out-Null
    Set-Content -LiteralPath $forbiddenPath -Value "forbidden"

    Assert-ValidationRejectsPath -ExpectedPath $expectedPath -Action {
      Assert-RailKeeperPackageDirectory -PackageDir $badPackage
    }

    $badZip = Join-Path $resolvedRoot ("bad-" + $caseID + ".zip")
    Compress-Archive -Path (Join-Path $badPackage "*") -DestinationPath $badZip
    Assert-ValidationRejectsPath -ExpectedPath $expectedPath -Action {
      Assert-RailKeeperPackageArchive -ZipPath $badZip
    }
  }

  Write-Host "Windows package validation regression passed."
} finally {
  if (Test-Path -LiteralPath $resolvedRoot) {
    $cleanupPath = [IO.Path]::GetFullPath($resolvedRoot)
    if ($cleanupPath -ne $resolvedRoot -or
        -not $cleanupPath.StartsWith($temporaryParent, [StringComparison]::OrdinalIgnoreCase)) {
      throw "Refusing to clean unverified test root: $cleanupPath"
    }
    Remove-Item -LiteralPath $cleanupPath -Recurse -Force
  }
}
