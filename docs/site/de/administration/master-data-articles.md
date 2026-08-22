---
title: Artikelstammdaten und Lagerorte
description: Bestandseinheiten, Artikelklassifizierung, eigene Felder und hierarchische Lagerorte konfigurieren.
audience: admin
status: stable
reviewedVersion: 0.1.20
lastReviewed: 2026-08-16
---

# Artikelstammdaten und Lagerorte

Unter **Einstellungen > Daten > Artikeldaten** stehen die kontrollierten Werte für Zubehörartikel.
Die Seite trennt Bestandseinheiten, Artikelklassifizierung, eigene Felder und Lagerorte. Admin und
Editor können sie ändern. Viewer und Planner sehen ausdrücklich eine Nur-Lese-Ansicht.

## Bestandseinheiten

Bestandseinheiten beschreiben, was Artikelmengen zählen, beispielsweise Stück, Packungen, Meter
oder Gramm. Jeder Eintrag besitzt einen unveränderlichen Schlüssel, eine bearbeitbare Bezeichnung,
einen Aktivstatus und eine Herkunft.

Eine neue Einheit nur anlegen, wenn die vorhandene Liste den Bestand nicht ausdrücken kann. Eine
Deaktivierung entfernt sie aus neuen Artikelauswahlen. Bereits zugeordnete Artikel behalten den
historischen Wert. Eine eigene Einheit kann nur endgültig gelöscht werden, solange kein Artikel
ihren Schlüssel verwendet.

## Artikelarten und Unterarten

RailKeeper liefert acht geschützte Artikelart-Schlüssel aus:

- Gleis
- Signal
- Decoder
- Elektrik / Steuerung
- Gebäude / Ausstattung
- Landschaftsverbrauchsmaterial
- Beleuchtung
- Sonstiges

Bezeichnungen und Aktivstatus lassen sich verwalten, aber v0.1.20 erlaubt weder neue noch gelöschte
Artikelart-Schlüssel. Dadurch bleiben die zugehörigen Verträge für Fachangaben geschützt. Eine
Deaktivierung entfernt die Art aus neuen Artikelauswahlen, ohne vorhandene Artikel umzuschreiben.

Unterarten verfeinern eine dieser Artikelarten. Admin und Editor können eigene Unterarten anlegen,
bearbeiten, deaktivieren, reaktivieren und im ungenutzten Zustand löschen. Ihr Schlüssel muss mit
dem Schlüssel der übergeordneten Artikelart und einem Doppelpunkt beginnen, beispielsweise
`track:turnout`. An diesem Präfix erkennt RailKeeper die zugehörige Artikelart.

Eine geänderte oder deaktivierte Klassifizierung migriert Artikeldatensätze niemals automatisch.
Die betroffenen Artikel prüfen und bewusst entscheiden, ob ihre historische Klassifizierung bleibt
oder einzeln bearbeitet wird.

## Kontrollierte eigene Felder

Eigene Felder erscheinen nur bei Artikeln der Art **Sonstiges**. Jede Definition besitzt Schlüssel,
Bezeichnung, Aktivstatus und einen dieser Datentypen:

| Datentyp | Konfiguration und gespeicherter Wert |
| --- | --- |
| Text | Freitext |
| Zahl | Zahlenwert und optionale Einheit |
| Ja / Nein | Ausdrücklicher boolescher Wert |
| Datum | Datumswert |
| Einfachauswahl | Genau ein konfigurierter Auswahlwert |
| Mehrfachauswahl | Kein, ein oder mehrere konfigurierte Auswahlwerte |

Einfach- und Mehrfachauswahl benötigen eine nicht leere, kommagetrennte Optionsliste. Die Oberfläche
entfernt leere Optionen, der Server weist doppelte Optionen ab. Den Datentyp sorgfältig wählen: Sobald
ein eigenes Feld in gespeicherten Attributen eines **Sonstiges**-Artikels verwendet wird, kann sein
Schlüssel nicht mehr gelöscht werden. Historische Werte mit inaktiver Definition bleiben lesbar,
lassen sich aber weder neu hinzufügen noch ändern.

## Lagerorte

Lagerorte bilden eine Hierarchie außerhalb der kontrollierten Stammdatentabelle. Jeder Ort besitzt
einen erforderlichen Namen, optionalen übergeordneten Ort, optionale Beschreibung und Archivstatus.
Die Liste zeigt den vollständigen Pfad, beispielsweise `Lager / Schrank A / Schublade 3`.

Namen müssen innerhalb derselben Ebene ohne Beachtung der Groß- und Kleinschreibung eindeutig sein.
Ein Lagerort kann weder sein eigener übergeordneter Ort sein noch unter einen seiner Nachfolger
verschoben werden. Die Elternauswahl blendet diese ungültigen Möglichkeiten aus.

In v0.1.20 gibt es keine endgültige Löschaktion. Einen nicht mehr verwendeten Ort archivieren und
bei Bedarf reaktivieren. Bestand und Einzelstücke können einen Ort nur verwenden, wenn dieser Ort
und alle übergeordneten Orte aktiv sind. Das Archivieren eines Elternorts sperrt daher seinen
gesamten Zweig für neue Bestandsvorgänge, ohne gespeicherte Ortsverweise zu löschen.

Eine sichere Hierarchie wird so aufgebaut:

1. Übergeordneten Raum oder Lagerbereich anlegen.
2. Schränke oder Regale mit diesem Elternort anlegen.
3. Fächer oder Schubladen unter dem passenden Zweig ergänzen.
4. Vor der Bestandszuordnung den vollständig angezeigten Pfad prüfen.
5. Veraltete Endpunkte zuerst archivieren, danach ihre leeren Elternorte.

Lagerorte sind in Anwendungssicherung und Wiederherstellung enthalten, aber nicht im eigenständigen
Stammdaten-JSON.

## Lebenszyklus und Konflikte

Die Aktionsspalte folgt den vom Server berechneten Möglichkeiten. Mitgelieferte Einträge zeigen nie
das endgültige Löschen. Verwendete eigene Einträge verlieren diese Aktion ebenfalls. Deaktivieren
und Löschen erfordern eine Bestätigung, Reaktivieren schreibt sofort.

Meldet das Speichern einen Konflikt, auf einen doppelten gleichgeordneten Lagerort, eine geschützte
Artikelart, einen verwendeten eigenen Wert oder eine ungültige Optionsliste prüfen. Ein
abgewiesener Vorgang lässt die gespeicherte Konfiguration unverändert.

## Verwandte Seiten

- [Stammdaten-Administration](./master-data)
- [Allgemeine Stammdaten](./master-data-general)
- [Artikelstammdaten und Fachangaben](/de/guide/accessories/article-records)
- [Bestand, Käufe und Dokumente](/de/guide/accessories/stock-purchases-documents)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert RailKeeper **v0.1.20** und wurde zuletzt am 16.08.2026 geprüft.
