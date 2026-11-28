$ErrorActionPreference = "Stop"

$script:ForbiddenRailKeeperDirectories = @(
  "data",
  "uploads",
  "attachments",
  "thumbnails",
  "backups"
)

$script:RequiredRailKeeperPackageFiles = @(
  "RailKeeper.exe",
  "Start RailKeeper.exe",
  "start-railkeeper.bat",
  "LICENSE.md",
  "THIRD_PARTY_NOTICES.md",
  "THIRD_PARTY_LICENSES.txt",
  "web/fonts/LICENSE-Selawik.txt"
)

function Assert-RailKeeperExecutableVersion {
  param([string]$ExePath, [string]$Version, [switch]$MetadataOnly)

  $info = [Diagnostics.FileVersionInfo]::GetVersionInfo($ExePath)
  $expectedNumber = @((($Version -split '-', 2)[0] -split '\.') | ForEach-Object { [int]$_ })
  if ($expectedNumber.Count -eq 3) { $expectedNumber += 0 }
  $fileNumber = @($info.FileMajorPart, $info.FileMinorPart, $info.FileBuildPart, $info.FilePrivatePart)
  $productNumber = @($info.ProductMajorPart, $info.ProductMinorPart, $info.ProductBuildPart, $info.ProductPrivatePart)
  if ($info.ProductName -cne "RailKeeper" -or $info.OriginalFilename -cne [IO.Path]::GetFileName($ExePath) -or
      $info.ProductVersion -cne $Version -or $info.FileVersion -cne $Version -or
      ($fileNumber -join '.') -ne ($expectedNumber -join '.') -or
      ($productNumber -join '.') -ne ($expectedNumber -join '.')) {
    throw "Windows version metadata does not match RailKeeper $Version."
  }

  if ($MetadataOnly) {
    Write-Host "Verified Windows executable $($info.OriginalFilename) ($($info.ProductName) $Version, PE $($fileNumber -join '.'))."
    return
  }

  $runtimeVersion = & $ExePath version
  if ($LASTEXITCODE -ne 0 -or ($runtimeVersion | ConvertFrom-Json).version -cne $Version) {
    throw "Packaged RailKeeper.exe version does not match $Version."
  }
  Write-Host "Verified Windows executable $($info.OriginalFilename) ($($info.ProductName) $Version, PE $($fileNumber -join '.')) against its CLI version."
}

function Assert-SafeRailKeeperPackageEntry {
  param([string]$RelativePath)

  $normalized = $RelativePath.Replace("\", "/").Trim("/")
  if ([string]::IsNullOrWhiteSpace($normalized)) {
    return
  }
  if ($RelativePath.StartsWith("/") -or $RelativePath.StartsWith("\") -or
      $normalized -match "^[A-Za-z]:") {
    throw "Unsafe absolute package entry: $RelativePath"
  }

  $segments = $normalized.Split("/", [StringSplitOptions]::RemoveEmptyEntries)
  if ($segments -contains "." -or $segments -contains "..") {
    throw "Unsafe traversing package entry: $RelativePath"
  }
  foreach ($segment in $segments) {
    if ($script:ForbiddenRailKeeperDirectories -contains $segment.ToLowerInvariant()) {
      throw "Forbidden package entry: $RelativePath"
    }
  }

  $fileName = $segments[-1].ToLowerInvariant()
  if ($fileName.EndsWith(".db") -or $fileName.EndsWith(".db-wal") -or
      $fileName.EndsWith(".db-shm")) {
    throw "Forbidden package entry: $RelativePath"
  }
}

function Assert-RailKeeperPackageDirectory {
  param([string]$PackageDir)

  if ([string]::IsNullOrWhiteSpace($PackageDir)) {
    throw "Package directory is required."
  }
  $rootItem = Get-Item -LiteralPath $PackageDir -ErrorAction Stop
  if (-not $rootItem.PSIsContainer) {
    throw "Package directory is not a directory: $PackageDir"
  }
  $rootPath = [IO.Path]::GetFullPath($rootItem.FullName)
  foreach ($item in Get-ChildItem -LiteralPath $rootPath -Recurse -Force) {
    $relativePath = [IO.Path]::GetRelativePath($rootPath, $item.FullName)
    Assert-SafeRailKeeperPackageEntry -RelativePath $relativePath
  }
  foreach ($name in $script:RequiredRailKeeperPackageFiles) {
    $file = Get-Item -LiteralPath (Join-Path $rootPath $name) -ErrorAction SilentlyContinue
    if (-not $file -or $file.PSIsContainer -or -not $file.Length) {
      throw "Missing or empty package file: $name"
    }
  }
}

function Assert-RailKeeperPackageArchive {
  param([string]$ZipPath, [string]$PackageDir)

  if ([string]::IsNullOrWhiteSpace($ZipPath)) {
    throw "Package archive is required."
  }
  $archiveItem = Get-Item -LiteralPath $ZipPath -ErrorAction Stop
  if ($archiveItem.PSIsContainer) {
    throw "Package archive is not a file: $ZipPath"
  }

  Add-Type -AssemblyName System.IO.Compression.FileSystem
  $archive = [IO.Compression.ZipFile]::OpenRead($archiveItem.FullName)
  try {
    $files = @{}
    foreach ($entry in $archive.Entries) {
      Assert-SafeRailKeeperPackageEntry -RelativePath $entry.FullName
      $name = $entry.FullName.Replace("\", "/")
      if ($files.ContainsKey($name)) { throw "Duplicate archive entry: $name" }
      $files.Add($name, $entry)
    }
    foreach ($name in $script:RequiredRailKeeperPackageFiles) {
      $entry = $files[$name]
      if (-not $entry -or -not $entry.Length) { throw "Missing or empty archive file: $name" }
      if ($PackageDir) {
        $stream = $entry.Open()
        try {
          $actualHash = [Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($stream))
        } finally {
          $stream.Dispose()
        }
        $expectedHash = (Get-FileHash -LiteralPath (Join-Path $PackageDir $name)).Hash
        if ($actualHash -ne $expectedHash) { throw "Archive file differs from qualified package: $name" }
      }
    }
  } finally {
    $archive.Dispose()
  }
}
