function Get-RailKeeperBuildIdentity {
  param([string]$Repo, [string]$Version)

  $result = [pscustomobject]@{ Commit = "unknown"; Kind = "development"; Modified = "unknown" }
  if (-not (Get-Command git -ErrorAction SilentlyContinue)) { return $result }
  $revision = & git -C $Repo rev-parse --verify HEAD 2>$null
  if ($LASTEXITCODE -ne 0 -or $revision -notmatch '^[0-9a-f]{40}([0-9a-f]{24})?$') { return $result }
  $result.Commit = $revision.Trim()
  $changes = @(& git -C $Repo status --porcelain --untracked-files=normal)
  if ($LASTEXITCODE -ne 0) { return $result }
  $result.Modified = if ($changes.Count -gt 0) { "true" } else { "false" }
  $tags = @(& git -C $Repo tag --points-at HEAD)
  if ($LASTEXITCODE -eq 0 -and $result.Modified -eq "false" -and $tags -contains ("v" + $Version.TrimStart('v'))) {
    $result.Kind = "release"
  }
  return $result
}
