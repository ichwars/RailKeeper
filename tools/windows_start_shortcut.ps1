function New-RailKeeperStartShortcut {
  param([string]$PackageDir)

  $path = Join-Path $PackageDir "Start RailKeeper.lnk"
  $executable = Get-Item -LiteralPath (Join-Path $PackageDir "RailKeeper.exe")
  $stream = [IO.File]::Open($path, [IO.FileMode]::CreateNew, [IO.FileAccess]::Write)
  $writer = [IO.BinaryWriter]::new($stream)
  try {
    # MS-SHLLINK header. Store only a relative target and Unicode arguments.
    $writer.Write([uint32]0x4c)
    $writer.Write(([Guid]"00021401-0000-0000-c000-000000000046").ToByteArray())
    $writer.Write([uint32]0xa8) # HasRelativePath | HasArguments | IsUnicode
    $writer.Write([uint32]0x80) # FILE_ATTRIBUTE_NORMAL
    $writer.Write([byte[]]::new(24)) # Target timestamps left unset
    $writer.Write([uint32]$executable.Length)
    $writer.Write([int32]0) # Default icon index
    $writer.Write([uint32]7) # SW_SHOWMINNOACTIVE
    $writer.Write([byte[]]::new(12)) # No hotkey, reserved fields

    foreach ($value in @(".\RailKeeper.exe", "--standalone")) {
      $writer.Write([uint16]$value.Length)
      $writer.Write([Text.Encoding]::Unicode.GetBytes($value))
    }
    $writer.Write([uint32]0) # ExtraData terminator
  } finally {
    $writer.Dispose()
  }
}
