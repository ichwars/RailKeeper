# RailKeeper Abschlussnotiz 2026-06-04

## Erledigt

- Punkt 1: `VehiclesView.tsx` fachlich weiter aufgeteilt; Inventarsteuerung liegt in einem eigenen Controller-Hook.
- Punkt 2: OCR-Unterstützung für gescannte PDF-Ersatzteillisten ergänzt.
- Punkt 3: PDF-Parser mit zusätzlichen Hersteller-Fixtures und robusteren Erkennungen gehärtet.
- Punkt 4: Ersatzteil-Websuche stärker begrenzt; Piko/Roco-Suchen laufen über die bekannten Herstellerquellen und übernehmen Preis/Link.
- Punkt 5: Schreib-Sync zur Digitalzentrale ausgebaut und abgesichert.
- Punkt 6: Z21- und CS3-Adapter ergänzt.
- Punkt 7: Auth/2FA im Backend verschärft, inklusive TOTP-Flows, Session-/Audit-Anbindung und Tests.
- Punkt 9: CRLF/LF über `.gitattributes` vereinheitlicht.
- Punkt 10: Mehrere UI-Anpassungen umgesetzt: systemweite Icon-Behandlung, Tabellen-Ausrichtung, Upload-Aktionen, PDF-Anzeige, Ersatzteilaktionen, Kartenbilder, QR-Code ohne Logo, Menü-Verlinkungen, native Datumseingabe ersetzt, Messelistenbilder vergrößert und Eintrag-Button als Icon.
- Übersicht/Bestand: Listenwert-Berechnung korrigiert, Handlungsbedarf filtert Bestand/Karten, Ausstellungsschalter werden nach Messelisten-Datum plus einem Tag automatisch zurückgesetzt.
- Fahrzeugdaten: Quellenzeile zweispaltig mit Digitalzentralen-ID, sofern vorhanden.

## Abschlussprüfung

- Backend: `go test ./...` erfolgreich.
- Frontend: `npm.cmd run build` erfolgreich.
- Whitespace: `git diff --check` ohne Befund.
- Encoding/Umlaute: keine kaputten Umlaute gefunden; der einzige Treffer ist die absichtliche `repairMojibake`-Erkennung im Ersatzteil-Suchcode.
- Debug-/Altlasten: keine `TODO`, `FIXME`, `debugger` oder `console.log`-Reste im geprüften Bereich gefunden.
- Historische UI-Notiz im Benutzerbereich entfernt.
- Harte deutsche Meldungen in den Einstellungen auf i18n-Keys umgestellt.

## Noch offen

- Native Browser-Dialoge (`window.confirm`/`window.alert`) sind noch vorhanden. Die Texte sind übersetzt, aber die Dialoge sollten später durch einen gemeinsamen App-Dialog ersetzt werden, damit wirklich keine nativen Browser-Komponenten mehr erscheinen.
- ECoS/Z21/CS3 sollten noch gegen echte Hardware bzw. stabile Testgeräte geprüft werden; die Adapter- und API-Pfade sind vorbereitet.
- OCR/PDF-Parser und Ersatzteilquellen sollten bei neuen Hersteller-PDFs weiter mit Fixtures ergänzt werden.
- Externe Verfügbarkeitsstatus-Abfragen bleiben abhängig von Piko/Roco-Seitenstruktur und sollten nach Herstelleränderungen kontrolliert werden.
- Die finale Browser-Sichtprüfung der letzten Messelistenänderung konnte wegen einer abgebrochenen internen Browser-Verbindung nicht abgeschlossen werden; Build und statische Checks sind grün.
