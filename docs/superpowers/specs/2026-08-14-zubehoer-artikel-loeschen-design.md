# Zubehörartikel sicher löschen

## Ziel

Administratoren können einen vollständig unbenutzten Zubehörartikel endgültig löschen. Das Löschen
entfernt keine Bestands-, Kauf- oder Nutzungshistorie und bleibt deshalb für jeden Artikel mit
fachlichen Bezügen gesperrt. Archivieren und Wiederherstellen bleiben unverändert für Administratoren
und Editoren verfügbar.

## Fachliche Regeln

Ein Artikel darf nur gelöscht werden, wenn alle folgenden Bedingungen erfüllt sind:

- Der gesamte Bestand ist null.
- Es existieren keine individuellen Artikelobjekte.
- Es existieren keine Bestandsbewegungen oder Käufe.
- Es existieren keine Reservierungen oder Einbauten, unabhängig von deren aktuellem Status.
- Es existiert keine Nutzungshistorie.
- Der Artikel ist mit keiner technischen Anlagenposition verknüpft.

Leere Bestandszeilen, benutzerdefinierte Attribute, Bilder und sonstige Dokumente gelten nicht als
Nutzung. Sie werden beim erlaubten Löschen mit entfernt. Nicht mehr referenzierte Datei-Blobs werden
über die bestehende Blob-Bereinigung gelöscht. Ein vorheriges Archivieren ist nicht erforderlich.

## API und Berechtigung

Die API erhält `DELETE /api/v1/accessory-products/{id}`. Die Route verwendet die bestehende
Admin-Berechtigung und bleibt serverseitig gegen Editor, Viewer, Planner und Messe gesperrt.

Der Anwendungsdienst führt Prüfung und Löschen als eine atomare Operation aus. Damit kann sich der
Zustand nicht zwischen einer Vorabprüfung und dem Löschen ändern. Die Infrastruktur löscht zuerst
erlaubte abhängige Metadaten und leere Bestandszeilen, anschließend den Artikel. Dokument-Blob-IDs
werden zurückgegeben und nach dem erfolgreichen Datenbankabschluss mit der bestehenden
Referenzbereinigung geprüft.

Ergebnisse:

- `204 No Content` nach erfolgreichem Löschen.
- `404 Not Found`, wenn der Artikel nicht existiert.
- `409 Conflict` mit dem Problemcode `accessory_delete_blocked`, wenn eine Schutzbedingung greift.
- Die bestehende Authentifizierungs- und CSRF-Behandlung bleibt unverändert aktiv.

Der öffentliche Vertrag in `openapi/railkeeper.yaml` und der Frontend-API-Adapter werden synchron
aktualisiert. Das erfolgreiche Löschen wird mit Akteur und Artikel-ID im Audit-Log erfasst.

## Oberfläche

Das gemeinsame Drei-Punkte-Menü erhält unter Archivieren/Wiederherstellen den Eintrag
„Artikel löschen“. Er erscheint ausschließlich für Administratoren und damit automatisch in Tabelle,
Kachelansicht und kompakter Mobilansicht.

Die Aktion öffnet den bestehenden `AccessoryConfirmDialog` als gefährliche Bestätigung. Titel und
Text nennen Inventarnummer und Bezeichnung eindeutig. Erst die Bestätigung sendet den Löschbefehl.
Während der Anfrage bleibt die Bestätigung gesperrt.

Nach Erfolg schließt der Dialog und die Übersicht lädt ihre Daten neu. Bei `409 Conflict` bleibt der
Dialog geöffnet und zeigt eine verständliche Meldung, dass Bestand oder historische Bezüge das
Löschen verhindern. Andere API-Fehler werden über die bestehende Fehlerdarstellung ausgegeben.

## Komponenten und Datenfluss

- `ArticleActions` zeigt die Admin-Aktion und meldet den ausgewählten Artikel an `AccessoriesView`.
- `AccessoriesView` verwaltet genau einen ausstehenden Löschvorgang und rendert den
  Bestätigungsdialog außerhalb der drei Präsentationsvarianten.
- `useArticleOverview` führt den API-Aufruf aus und lädt nach Erfolg Liste und Kennzahlen neu.
- Der API-Handler delegiert ausschließlich an den Anwendungsdienst; Löschregeln verbleiben nicht im
  Handler oder Frontend.
- Repository und Blob-Dienst bewahren Datenbank- und Dateispeicher-Konsistenz.

## Fehler- und Sicherheitsverhalten

- Die UI-Sichtbarkeit ersetzt keine serverseitige Rollenprüfung.
- Fremdschlüssel bleiben als letzte Schutzschicht bestehen.
- Ein fehlgeschlagener Löschversuch verändert weder Artikel noch Dokumente oder Blobs.
- Datei-Blobs werden nur entfernt, wenn keine Referenz aus Fahrzeugen oder anderen Artikeln besteht.
- Es gibt keine Kaskade über Bestands-, Kauf-, Reservierungs-, Einbau- oder Nutzungshistorie.

## Tests und Abnahme

Backend-Tests belegen:

- Admin-Zugriff und Ablehnung aller nicht berechtigten Rollen.
- Erfolgreiches Löschen eines unbenutzten Artikels einschließlich Metadaten und Blob-Bereinigung.
- Konflikte für Bestand, Einzelobjekte, Bewegungen, Käufe, Reservierungen, Einbauten und
  Anlagenverknüpfungen.
- Unveränderte Daten nach jedem abgelehnten Versuch.
- `404`, Audit-Eintrag und OpenAPI-Routenvertrag.

Frontend-Tests belegen:

- Sichtbarkeit ausschließlich für Admins in den gemeinsamen Aktionen.
- Bestätigung vor dem API-Aufruf.
- Neuladen der Übersicht nach Erfolg.
- Erhalt des Artikels und sichtbare Fehlermeldung nach einem Konflikt.
- Deutsche und englische Beschriftungen.

Abschließend laufen `go test ./...`, der vollständige Frontend-Test, `npm.cmd run build` und eine
Browserprüfung in Tabelle, Kachelansicht und kompakter Mobilansicht.
