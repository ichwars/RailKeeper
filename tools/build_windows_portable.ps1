param(
  [string]$Version = ""
)

$ErrorActionPreference = "Stop"

$repo = Resolve-Path (Join-Path $PSScriptRoot "..")
$frontend = Join-Path $repo "frontend"
$backend = Join-Path $repo "backend"
$packageRoot = Join-Path $repo "dist\windows-portable"
$packageDir = Join-Path $packageRoot "RailKeeper-Portable"
$brandIconSource = Join-Path $repo "frontend\public\brand\railkeeper-mark.png"

function Write-UInt16LE {
  param(
    [System.IO.BinaryWriter]$Writer,
    [int]$Value
  )
  $Writer.Write([UInt16]$Value)
}

function Write-UInt32LE {
  param(
    [System.IO.BinaryWriter]$Writer,
    [long]$Value
  )
  $Writer.Write([UInt32]$Value)
}

function New-IconImageSet {
  param(
    [string]$SourcePng
  )

  if (-not (Test-Path $SourcePng)) {
    throw "RailKeeper icon source not found: $SourcePng"
  }

  Add-Type -AssemblyName System.Drawing
  $source = [System.Drawing.Image]::FromFile($SourcePng)
  $sizes = @(256, 64, 48, 32, 16)
  $images = @()

  try {
    foreach ($size in $sizes) {
      $bitmap = New-Object System.Drawing.Bitmap $size, $size
      $graphics = [System.Drawing.Graphics]::FromImage($bitmap)
      try {
        $graphics.Clear([System.Drawing.Color]::Transparent)
        $graphics.CompositingQuality = [System.Drawing.Drawing2D.CompositingQuality]::HighQuality
        $graphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
        $graphics.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
        $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality

        $scale = [Math]::Min($size / $source.Width, $size / $source.Height)
        $drawWidth = [int][Math]::Round($source.Width * $scale)
        $drawHeight = [int][Math]::Round($source.Height * $scale)
        $x = [int][Math]::Floor(($size - $drawWidth) / 2)
        $y = [int][Math]::Floor(($size - $drawHeight) / 2)
        $graphics.DrawImage($source, $x, $y, $drawWidth, $drawHeight)

        $stream = New-Object System.IO.MemoryStream
        try {
          $bitmap.Save($stream, [System.Drawing.Imaging.ImageFormat]::Png)
          $images += [pscustomobject]@{
            Width = $size
            Height = $size
            Bytes = $stream.ToArray()
          }
        } finally {
          $stream.Dispose()
        }
      } finally {
        $graphics.Dispose()
        $bitmap.Dispose()
      }
    }
  } finally {
    $source.Dispose()
  }

  return $images
}

function New-GroupIconResource {
  param(
    [array]$Images
  )

  $stream = New-Object System.IO.MemoryStream
  $writer = New-Object System.IO.BinaryWriter $stream
  try {
    Write-UInt16LE $writer 0
    Write-UInt16LE $writer 1
    Write-UInt16LE $writer $Images.Count

    for ($index = 0; $index -lt $Images.Count; $index++) {
      $image = $Images[$index]
      $widthByte = if ($image.Width -ge 256) { 0 } else { $image.Width }
      $heightByte = if ($image.Height -ge 256) { 0 } else { $image.Height }
      $writer.Write([byte]$widthByte)
      $writer.Write([byte]$heightByte)
      $writer.Write([byte]0)
      $writer.Write([byte]0)
      Write-UInt16LE $writer 1
      Write-UInt16LE $writer 32
      Write-UInt32LE $writer $image.Bytes.Length
      Write-UInt16LE $writer ($index + 1)
    }

    return $stream.ToArray()
  } finally {
    $writer.Dispose()
    $stream.Dispose()
  }
}

function Set-RailKeeperExecutableIcon {
  param(
    [string]$ExePath,
    [string]$SourcePng
  )

  if (-not ("RailKeeperResourceUpdater" -as [type])) {
    Add-Type @"
using System;
using System.Runtime.InteropServices;

public static class RailKeeperResourceUpdater {
  [DllImport("kernel32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
  public static extern IntPtr BeginUpdateResource(string pFileName, bool bDeleteExistingResources);

  [DllImport("kernel32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
  public static extern bool UpdateResource(IntPtr hUpdate, IntPtr lpType, IntPtr lpName, ushort wLanguage, byte[] lpData, uint cbData);

  [DllImport("kernel32.dll", SetLastError = true)]
  public static extern bool EndUpdateResource(IntPtr hUpdate, bool fDiscard);

  public static IntPtr MakeIntResource(int value) {
    return (IntPtr)value;
  }
}
"@
  }

  $images = New-IconImageSet -SourcePng $SourcePng
  $groupIcon = New-GroupIconResource -Images $images
  $rtIcon = [RailKeeperResourceUpdater]::MakeIntResource(3)
  $rtGroupIcon = [RailKeeperResourceUpdater]::MakeIntResource(14)
  $languageNeutral = [UInt16]0
  $handle = [RailKeeperResourceUpdater]::BeginUpdateResource($ExePath, $false)
  if ($handle -eq [IntPtr]::Zero) {
    $errorCode = [Runtime.InteropServices.Marshal]::GetLastWin32Error()
    throw "Could not open RailKeeper.exe for icon resources. Win32 error: $errorCode"
  }

  $discard = $true
  try {
    for ($index = 0; $index -lt $images.Count; $index++) {
      $resourceID = $index + 1
      $bytes = [byte[]]$images[$index].Bytes
      $ok = [RailKeeperResourceUpdater]::UpdateResource($handle, $rtIcon, [RailKeeperResourceUpdater]::MakeIntResource($resourceID), $languageNeutral, $bytes, [UInt32]$bytes.Length)
      if (-not $ok) {
        $errorCode = [Runtime.InteropServices.Marshal]::GetLastWin32Error()
        throw "Could not write icon image resource $resourceID. Win32 error: $errorCode"
      }
    }

    $ok = [RailKeeperResourceUpdater]::UpdateResource($handle, $rtGroupIcon, [RailKeeperResourceUpdater]::MakeIntResource(1), $languageNeutral, [byte[]]$groupIcon, [UInt32]$groupIcon.Length)
    if (-not $ok) {
      $errorCode = [Runtime.InteropServices.Marshal]::GetLastWin32Error()
      throw "Could not write icon group resource. Win32 error: $errorCode"
    }

    $discard = $false
  } finally {
    $ok = [RailKeeperResourceUpdater]::EndUpdateResource($handle, $discard)
    if (-not $ok -and -not $discard) {
      $errorCode = [Runtime.InteropServices.Marshal]::GetLastWin32Error()
      throw "Could not finalize RailKeeper.exe icon resources. Win32 error: $errorCode"
    }
  }
}

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
  $exePath = Join-Path $packageDir "RailKeeper.exe"
  $env:GOOS = "windows"
  $env:GOARCH = "amd64"
  $env:GOCACHE = Join-Path $repo ".cache\go-build"
  New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null
  & go build -trimpath -ldflags "-s -w" -o $exePath ./cmd/railkeeper
  if ($LASTEXITCODE -ne 0) {
    throw "Windows backend build failed with exit code $LASTEXITCODE"
  }
  Set-RailKeeperExecutableIcon -ExePath $exePath -SourcePng $brandIconSource
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
