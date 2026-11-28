---
title: Code signing policy for Windows
description: Draft policy for approving, signing and verifying RailKeeper for Windows.
audience: reference
status: development
reviewedVersion: main
lastReviewed: 2026-11-28
---

# Code signing policy for Windows

> **Draft.** The signing rules below are proposed for introduction. RailKeeper currently
> has no active signing through the SignPath Foundation.

## Which files will be signed?

The proposed signing scope is `RailKeeper.exe` and `Start RailKeeper.exe` in the
official Windows x64 Standalone ZIP. The **Windows Standalone Package** workflow
builds both from a verified release tag in the public `ichwars/RailKeeper` repository
on GitHub-hosted runners. Package validation checks the product name and version
against the package and release tag.
[Source code](https://github.com/ichwars/RailKeeper) ·
[Windows build workflow](https://github.com/ichwars/RailKeeper/blob/main/.github/workflows/windows-standalone.yml)

## Who reviews and approves?

| Responsibility | Assignment in this draft |
| --- | --- |
| Author with commit rights | `@ichwars` is the maintainer named in `CODEOWNERS`. |
| Review | `@ichwars` is the proposed reviewer. Contributions from people without commit rights require review by an authorized team member before acceptance. |
| Manual signing approval | `@ichwars` is the proposed approver. Each signing request requires approval for the specific release build. |

Before activation, the project confirms these roles and verifies multi-factor authentication
for the accounts involved on GitHub and with the signing service.
[Repository code owners](https://github.com/ichwars/RailKeeper/blob/main/.github/CODEOWNERS) ·
[SignPath Foundation conditions](https://signpath.org/terms)

## From build to download

1. The approver checks the tag, source commit, successful build run and corresponding
   workflow artifact, then manually approves the signing request.
2. Both executables receive timestamped Authenticode signatures after the final resource
   change. The signing configuration enforces the product name `RailKeeper` and a common
   product version. The ZIP, its checksum and build attestation follow.
3. Package verification checks the version and expected publisher. The corresponding release
   notes name the publisher shown in the certificate and link to the available evidence.
   [GitHub Releases](https://github.com/ichwars/RailKeeper/releases)

After acceptance and activation, the downloads will identify [SignPath.io](https://about.signpath.io/)
as the free signing service and the [SignPath Foundation](https://signpath.org/)
as the certificate holder.
[SignPath integration with GitHub](https://docs.signpath.io/trusted-build-systems/github)

## Verify a signed download

For a release explicitly marked as signed, first compare the ZIP's SHA-256 value with
`checksums.txt` for that release. Extract the ZIP and use PowerShell to check both
signatures. Expect `Valid` and the publisher named in the release notes for each file.

```powershell
Get-AuthenticodeSignature -LiteralPath '.\RailKeeper.exe', '.\Start RailKeeper.exe' |
  Select-Object Path, Status, @{Name='Publisher'; Expression={$_.SignerCertificate.Subject}}
```

If either check fails, leave the file unopened and use the private reporting channel.
[Security reporting](https://github.com/ichwars/RailKeeper/security/policy)

## Privacy information for external connections

- **Updates and downloads.** A manual update check retrieves public release data from
  GitHub or the configured update source. It sends no collection data. The browser
  downloads the selected ZIP directly from GitHub.
  [Updates and operation](../administration/system-operations) ·
  [GitHub privacy statement](https://docs.github.com/en/site-policy/privacy-policies/github-general-privacy-statement)
- **Article search.** When you start a search, RailKeeper processes the search input,
  such as a manufacturer, article number or EAN. Depending on the selected sources,
  search terms derived from that input go to DuckDuckGo, and RailKeeper fetches
  external provider, result or document pages. You can turn article search off in the
  current browser. [Article search and settings](../guide/vehicles/search-and-spares) ·
  [DuckDuckGo privacy policy](https://duckduckgo.com/privacy)
- **Voluntary issue reports.** When the form is configured, submitting a report creates
  a GitHub issue. Review the preview before submitting it. You can turn off diagnostics,
  and screenshots are optional. The reporting guide explains the remaining fields and which
  data is visible. [Issue reports and submitted data](../guide/settings/bug-reporting) ·
  [GitHub privacy statement](https://docs.github.com/en/site-policy/privacy-policies/github-general-privacy-statement)
