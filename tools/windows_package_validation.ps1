$ErrorActionPreference = "Stop"

$script:ForbiddenRailKeeperDirectories = @(
  "data",
  "uploads",
  "attachments",
  "thumbnails",
  "backups"
)

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
}

function Assert-RailKeeperPackageArchive {
  param([string]$ZipPath)

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
    foreach ($entry in $archive.Entries) {
      Assert-SafeRailKeeperPackageEntry -RelativePath $entry.FullName
    }
  } finally {
    $archive.Dispose()
  }
}
