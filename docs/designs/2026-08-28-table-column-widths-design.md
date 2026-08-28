# Anpassbare Spaltenbreiten für Fahrzeuge und Zubehör

Status: beschlossen

Issue: #141

Datum: 28.08.2026

## Ziel

Die Desktop-Tabellen für Fahrzeugbestand und Zubehör erhalten direkt erkennbare Ziehbereiche im
Tabellenkopf. Sichtbarkeit, Reihenfolge und Breite bilden pro Benutzer und Tabelle ein gemeinsames
Layout, das über die vorhandenen Profileinstellungen gespeichert wird. Auswahl- und Aktionsspalten
bleiben fest. Die mobilen Karten- und Kompaktansichten ändern sich nicht.

## Ausgangslage

- Der Fahrzeugbestand speichert sichtbare Spalten und Reihenfolge bereits unter
  `railkeeper.vehicles.tableColumns` im Benutzerprofil.
- Die Zubehörverwaltung speichert sichtbare Spalten bisher nur im lokalen Browser unter
  `railkeeper.accessories.tableColumns` und verwendet eine feste Reihenfolge.
- Beide Tabellen verwenden feste CSS-Regeln für Mindestbreiten. Eine gemeinsame Validierung oder
  zugängliche Größenänderung existiert noch nicht.
- Der vorhandene Profilendpunkt speichert partielle String-Werte pro Benutzer und benötigt keine
  Backend- oder Datenbankänderung.

## Layoutmodell und Kompatibilität

Beide bestehenden Einstellungsschlüssel bleiben erhalten. Ihr Wert wird zu einem versionierten
JSON-Objekt erweitert:

```json
{
  "version": 1,
  "columns": ["inventoryNumber", "manufacturer", "name"],
  "widths": {
    "manufacturer": 184,
    "name": 286
  }
}
```

Nur vom Standard abweichende Breiten müssen gespeichert werden. Dadurch bleiben die Werte kompakt
und neu eingeführte Spalten verwenden automatisch ihre aktuellen Standardbreiten.

Die Parser akzeptieren weiterhin die bisherigen JSON-Arrays. Fahrzeugprofile werden dadurch ohne
Migration lesbar. Fehlt beim Zubehör der Profilwert, wird der bisherige `localStorage`-Wert als
Startwert gelesen und einmalig in das Benutzerprofil geschrieben. Ein fehlerhafter oder unbekannter
Wert fällt auf das jeweilige Standardlayout zurück.

## Gemeinsame Grundlage

Ein gemeinsames Frontend-Modul übernimmt:

- das versionierte Layoutformat,
- die Validierung bekannter, eindeutiger Spalten,
- das Normalisieren endlicher Ganzzahlbreiten,
- das Begrenzen auf die je Spalte definierten Mindest- und Maximalwerte,
- das Beibehalten von Breiten ausgeblendeter Spalten,
- die Berechnung der Tabellenmindestbreite,
- das Laden und geordnete Speichern über den Profilendpunkt.

Ein gemeinsamer Resize-Handle übernimmt Pointer- und Tastaturbedienung. Domänenspezifische Module
definieren nur Spalten, Standardbreiten, Grenzen und ihre Regeln für mindestens eine sichtbare
Identitätsspalte.

## Bedienung

Jede veränderbare Desktop-Spalte zeigt am rechten Rand des Tabellenkopfs einen schmalen, über die
volle Kopfhöhe erreichbaren Ziehbereich. Der Bereich wird bei Hover und Fokus deutlicher und nutzt
den Cursor `col-resize`.

Der Handle ist zusätzlich per Tastatur erreichbar:

- Pfeil links und rechts verkleinern oder vergrößern in 8-Pixel-Schritten.
- Umschalt plus Pfeiltaste verwendet 32-Pixel-Schritte.
- Pos1 setzt die Mindestbreite, Ende die Maximalbreite.
- Ein Doppelklick stellt die Standardbreite dieser Spalte wieder her.

Der Handle verwendet die Separator-Semantik mit aktuellem, minimalem und maximalem Wert. Während
des Ziehens wird die Darstellung sofort aktualisiert. Erst am Ende des Ziehens wird das vollständige
Layout gespeichert, damit keine Folge unnötiger Profilrequests entsteht.

## Tabellenverhalten

- Die Auswahlspalte und die Aktionsspalte behalten ihre bestehenden festen Breiten und erhalten
  keinen Handle.
- Die konfigurierte Gesamtbreite wird als Mindestbreite der Tabelle verwendet.
- Ist der Tabellencontainer breiter, füllt die Tabelle weiterhin die verfügbare Fläche.
- Ist die konfigurierte Gesamtbreite größer, scrollt ausschließlich der vorhandene Tabellenwrapper
  horizontal. Es entsteht kein horizontaler Scrollbalken für die ganze Seite.
- Zellen behalten die vorhandene Kürzung und zugängliche Volltexte.
- Ausgeblendete Spalten verlieren ihre gespeicherte Breite nicht.

## Fahrzeugbestand

Der vorhandene Spaltenauswahldialog bleibt bestehen. Sein Zurücksetzen stellt sichtbare Spalten,
Reihenfolge und alle Breiten auf die Fahrzeugstandards zurück. Die bestehende Profilpersistenz wird
auf das gemeinsame Layoutmodell umgestellt.

## Zubehörverwaltung

Die Zubehörverwaltung erhält einen eigenen Profil-Hook auf derselben Grundlage. Der vorhandene
Spaltenauswahldialog zeigt die sichtbaren Spalten in ihrer aktuellen Reihenfolge mit Hoch- und
Runter-Aktionen sowie die übrigen verfügbaren Spalten. Zurücksetzen stellt Sichtbarkeit,
Standardreihenfolge und Standardbreiten wieder her.

## Fehlerbehandlung

- Ein Ladefehler zeigt die bestehende, lokalisierte Meldungsfläche und verwendet das Standardlayout.
- Ein Speicherfehler lässt die lokale Darstellung unverändert und meldet den Fehler.
- Ungültige Breiten, unbekannte Spalten, Duplikate und ungültige JSON-Werte werden verworfen oder
  normalisiert, bevor sie die Tabelle erreichen.
- Mehrere schnelle Änderungen werden in Reihenfolge gespeichert. Ein fehlgeschlagener Request
  blockiert spätere Änderungen nicht.

## Tests und Prüfung

- Reine Modelltests decken Legacy-Arrays, versionierte Werte, unbekannte Spalten, Duplikate,
  ungültige Breiten, Grenzen, ausgeblendete Breiten und Zurücksetzen ab.
- Hook-Tests decken Profil-Laden, geordnete Speicherung, Fehlermeldungen und die Zubehörmigration ab.
- Komponententests prüfen Pointer- und Tastaturänderung, feste Auswahl-/Aktionsspalten,
  Tabellenmindestbreite, Reihenfolge und Reset.
- Visuelle Prüfung erfolgt bei 2580 px, 1440 px und 820 px, in Hell und Dunkel, mit langen deutschen
  Überschriften, sichtbarem Tastaturfokus und ohne globalen horizontalen Überlauf.

## Nicht im Umfang

- Größenänderung in mobilen Karten- oder Kompaktansichten
- serverseitige Tabellenlayouts außerhalb der bestehenden Benutzerprofile
- automatische Breitenmessung aus allen Zeilen
- veränderbare Auswahl- oder Aktionsspalten
