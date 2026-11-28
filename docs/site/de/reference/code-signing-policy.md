---
title: Signaturrichtlinie für Windows
description: Entwurf zur Freigabe, Signierung und Prüfung von RailKeeper für Windows.
audience: reference
status: development
reviewedVersion: main
lastReviewed: 2026-11-28
---

# Signaturrichtlinie für Windows

> **Entwurf.** Die beschriebenen Signaturregeln sind zur Einführung vorgesehen.
> RailKeeper hat derzeit keine aktive Signierung über die SignPath Foundation.

## Welche Dateien werden signiert?

Für die Signierung sind `RailKeeper.exe` und `Start RailKeeper.exe` im offiziellen
Windows-x64-Standalone-ZIP vorgesehen. Der Workflow **Windows Standalone Package** baut
beide aus dem geprüften Release-Tag des öffentlichen Repositories `ichwars/RailKeeper`
auf GitHub-gehosteten Runnern. Die Paketprüfung gleicht Produktname und Produktversion
mit dem Paket und dem Release-Tag ab.
[Quellcode](https://github.com/ichwars/RailKeeper) ·
[Windows-Build-Workflow](https://github.com/ichwars/RailKeeper/blob/main/.github/workflows/windows-standalone.yml)

## Wer prüft und gibt frei?

| Aufgabe | Zuständigkeit im Entwurf |
| --- | --- |
| Autor mit Commit-Recht | `@ichwars` ist der in `CODEOWNERS` benannte Maintainer. |
| Review | `@ichwars` ist als Reviewer vorgesehen. Beiträge von Personen ohne Commit-Recht benötigen vor der Übernahme ein Review durch ein zuständiges Teammitglied. |
| Manuelle Signaturfreigabe | `@ichwars` ist als Freigabeperson vorgesehen. Jede Signaturanfrage erfordert eine Freigabe für den konkreten Release-Build. |

Vor der Aktivierung werden die Rollen bestätigt und Mehrfaktor-Authentifizierung für
die beteiligten Konten auf GitHub und beim Signaturdienst geprüft.
[Zuständigkeiten im Repository](https://github.com/ichwars/RailKeeper/blob/main/.github/CODEOWNERS) ·
[Bedingungen der SignPath Foundation](https://signpath.org/terms)

## Ablauf vom Build zum Download

1. Die Freigabeperson prüft Tag, Quell-Commit, erfolgreichen Build-Lauf und das dazugehörige
   Workflow-Artefakt und genehmigt die Signaturanfrage manuell.
2. Beide EXE-Dateien erhalten nach der letzten Ressourcenänderung Authenticode-Signaturen
   mit Zeitstempel. Die Signaturkonfiguration prüft den Produktnamen `RailKeeper` und
   dieselbe Produktversion für beide Dateien. Danach entstehen das ZIP, seine Prüfsumme
   und die Build-Attestierung.
3. Die Paketprüfung gleicht Version und erwarteten Herausgeber ab. Die zugehörigen Release-Notizen
   nennen den Herausgeber aus dem Zertifikat und verlinken die verfügbaren Nachweise.
   [GitHub-Releases](https://github.com/ichwars/RailKeeper/releases)

Nach Aufnahme und Aktivierung nennen die Downloads [SignPath.io](https://about.signpath.io/)
als kostenlosen Signaturdienst und die [SignPath Foundation](https://signpath.org/)
als Zertifikatsinhaber.
[GitHub-Integration von SignPath](https://docs.signpath.io/trusted-build-systems/github)

## Einen signierten Download prüfen

Bei Releases, die ausdrücklich als signiert gekennzeichnet sind, vergleiche zuerst den
SHA-256-Wert des ZIPs mit `checksums.txt` des zugehörigen Releases. Entpacke das ZIP und
prüfe in PowerShell beide Signaturen. Erwartet werden jeweils der Status `Valid` und
der in den Release-Notizen genannte Herausgeber.

```powershell
Get-AuthenticodeSignature -LiteralPath '.\RailKeeper.exe', '.\Start RailKeeper.exe' |
  Select-Object Path, Status, @{Name='Herausgeber'; Expression={$_.SignerCertificate.Subject}}
```

Bei einer Abweichung lass die Datei geschlossen und nutze den privaten Meldeweg.
[Sicherheitsmeldungen](https://github.com/ichwars/RailKeeper/security/policy)

## Datenschutzhinweise zu externen Verbindungen

- **Updates und Downloads.** Eine manuelle Updatesuche ruft öffentliche Release-Daten
  von GitHub oder von der konfigurierten Updatequelle ab. Sie sendet keine Sammlungsdaten.
  Der Browser lädt das ausgewählte ZIP direkt von GitHub.
  [Updates und Betrieb](../administration/system-operations) ·
  [Datenschutzerklärung von GitHub](https://docs.github.com/en/site-policy/privacy-policies/github-general-privacy-statement)
- **Artikelsuche.** Bei einer gestarteten Suche verarbeitet RailKeeper die Suchangaben,
  etwa Hersteller, Artikelnummer oder EAN. Je nach gewählten Quellen gehen daraus
  abgeleitete Suchbegriffe an DuckDuckGo, und RailKeeper ruft externe Anbieter-, Ergebnis-
  oder Dokumentseiten ab. Die Artikelsuche kann im jeweiligen Browser deaktiviert werden.
  [Artikelsuche und Einstellungen](../guide/vehicles/search-and-spares) ·
  [Datenschutzerklärung von DuckDuckGo](https://duckduckgo.com/privacy)
- **Freiwillige Fehlerberichte.** Ist das Formular eingerichtet, werden abgesendete Berichte
  als GitHub-Issue angelegt. Prüfe die Vorschau vor dem Absenden. Diagnosedaten lassen sich
  abwählen und Screenshots sind freiwillig. Für weitere Felder und die Sichtbarkeit der Daten
  gilt die Anleitung zum Meldeformular.
  [Fehlerberichte und übermittelte Daten](../guide/settings/bug-reporting) ·
  [Datenschutzerklärung von GitHub](https://docs.github.com/en/site-policy/privacy-policies/github-general-privacy-statement)
