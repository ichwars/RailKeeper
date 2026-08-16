---
title: Fahrzeugwartung und Zustand
description: Fahrzeugwartungen erfassen, planen, abschließen, prüfen und sicher entfernen.
audience: user
status: stable
reviewedVersion: 0.1.18
lastReviewed: 2026-08-16
---

# Fahrzeugwartung und Zustand

Der Tab **Wartung** erfasst Arbeiten, Reparaturen, Umbauten, Zustand, Termine und Kosten an einem
Fahrzeug. Admin, Editor, Viewer und Planner können gespeicherte Einträge ansehen. Nur Admin und
Editor dürfen sie anlegen, ändern, abschließen oder löschen. Der Server erzwingt diese Grenze.

## Tab Wartung öffnen

Öffne ein Fahrzeug im **Fahrzeugbestand**, wähle **Bearbeiten** und dann **Wartung**. Das Fahrzeug
muss bereits einmal gespeichert worden sein, weil Wartungen seine gespeicherte Fahrzeug-ID
benötigen. Ein neues ungespeichertes Fahrzeug zeigt **Wartungseinträge können nach dem ersten
Speichern hinzugefügt werden.** Speichere den Grunddatensatz, bevor du einen Eintrag anlegst.

## Wartungsübersicht lesen

Oben im Tab stehen drei Zähler:

- **Fällig** zählt Einträge mit einem Fälligkeitsdatum bis einschließlich des heutigen lokalen
  Datums, deren Status nicht `erledigt` ist.
- **Geplant/offen** zählt jeden Eintrag, dessen Status nicht `erledigt` ist, einschließlich der
  bereits als fällig gezählten Einträge.
- **Erledigt** zählt Einträge mit dem Status `erledigt`.

**Fällig** ist daher eine Teilmenge von **Geplant/offen**, keine getrennte Statussumme. Ein Eintrag
mit Status `geplant` und einem Datum bis heute wird als fällig angezeigt und gezählt, obwohl sein
gespeicherter Status nicht `faellig` lautet.

RailKeeper zeigt offene Einträge vor erledigten. Innerhalb beider Gruppen stehen Einträge mit Datum
vor Einträgen ohne Datum, frühere Fälligkeitsdaten zuerst und bei gleichen Werten die zuletzt
angelegten Einträge zuerst.

## Wartungseintrag hinzufügen

Jede Wartungs-Schreibaktion, **Eintrag hinzufügen**, **Eintrag speichern**, **Erledigt** oder
**Wartung löschen**, wirkt sofort und wartet nicht auf **Änderungen speichern** am Fahrzeug. Nach
Erfolg lädt RailKeeper das vollständige Fahrzeug neu, ersetzt das Grunddatenformular und lädt den
funktionsbezogenen Editorzustand neu. Dadurch können ungespeicherte Grunddaten, Bildmetadaten,
Beilagenänderungen, ein teilweise ausgefülltes Wartungsformular und ausstehende Änderungen anderer
Tabs verloren gehen. Speichere oder verwirf bewusst alle ausstehenden Fahrzeugänderungen vor jeder
Wartungs-Schreibaktion.

Trage die Wartungsdaten ein und wähle **Eintrag hinzufügen**. Nach Erfolg setzt RailKeeper das
Formular auf Art `Wartung` und Status `geplant` zurück.

## Felder, Werte und Validierung

- **Art:** Erforderlich. Gespeicherte Werte sind `Wartung`, `Reparatur`, `Umbau`, `Superung`,
  `Reinigung`, `Schmierung`, `Decoder-Einbau` und `Ersatzteiltausch`. Standard: `Wartung`.
- **Status:** Erforderlich. Gespeicherte Werte sind `geplant`, `faellig` und `erledigt`; angezeigt
  als geplant, fällig und erledigt. Die Schreibweise `fällig` wird beim Speichern zu `faellig`
  normalisiert. Standard: `geplant`.
- **Zustand:** Optional. Gespeicherte Werte sind `neuwertig`, `sehr gut`, `gut`, `gebraucht` und
  `reparaturbedürftig`.
- **Fällig am:** Optionales gültiges Kalenderdatum im Format `YYYY-MM-DD`. Es steuert die
  datumsbasierte Fälligkeitsberechnung.
- **Durchgeführt am:** Optionales gültiges Kalenderdatum im Format `YYYY-MM-DD`. Beim Speichern
  eines erledigten Eintrags ohne dieses Datum wird das heutige lokale Datum eingesetzt.
- **Kosten:** Optionaler nicht negativer Dezimalbetrag. Komma und Punkt werden als Dezimaltrenner
  akzeptiert. Äußere und innere Leerzeichen sowie ein nachgestelltes Eurozeichen werden vor der
  Prüfung entfernt. Gültige Zahlen zeigt die Oberfläche als EUR im deutschen Zahlenformat an.
- **Notizen:** Optional. Der Server entfernt äußere Leerzeichen und akzeptiert höchstens 4.000
  Zeichen.

Der Server weist unbekannte Arten, Status und Zustände, ungültige Daten, negative oder nicht
numerische Kosten sowie zu lange Notizen zurück.

## Fälligkeit und Abschluss verstehen

Der gespeicherte Status `faellig` und die datumsbasierte Fälligkeitsmarkierung hängen zusammen,
sind aber unabhängig. Die Auswahl des Status fällig ergänzt kein Fälligkeitsdatum. Umgekehrt wird
ein offener Eintrag mit einem Fälligkeitsdatum bis heute hervorgehoben und gezählt, auch wenn sein
Status weiter `geplant` lautet.

Ein Eintrag zählt erst dann nicht mehr zu **Fällig** und **Geplant/offen**, wenn sein Status
`erledigt` ist. Das Abschlussdatum beschreibt den Abschluss, markiert den Eintrag allein aber nicht
als erledigt.

## Eintrag bearbeiten oder abbrechen

Wähle **Wartung bearbeiten** an einer Karte, um ihre Werte in das Formular zu übernehmen. Die
Hauptaktion lautet nun **Eintrag speichern**, zusätzlich erscheint **Abbrechen**. Speichern ändert
den Eintrag sofort und lädt das vollständige Fahrzeug neu. **Abbrechen** setzt das Wartungsformular
zurück, ohne den gespeicherten Eintrag zu ändern.

## Eintrag als erledigt markieren

Wähle **Erledigt** an einer offenen Karte, um sie sofort abzuschließen. RailKeeper fragt nicht
zusätzlich nach einer Bestätigung. Art, Zustand, Fälligkeitsdatum, Kosten, Notizen und ein bereits
vorhandenes Abschlussdatum bleiben erhalten. Fehlt das Abschlussdatum, verwendet RailKeeper das
heutige lokale Datum.

## Verknüpfte Medien prüfen

Sobald mindestens eine Verknüpfung besteht, zeigt die Wartungskarte unter **Verknüpfte Medien**
getrennte Zähler für Bilder und Beilagen. Bildverknüpfungen werden im Tab **Uploads** verwaltet. Die
Auswahl **Keine Wartung** an einem Bild bleibt eine ausstehende Metadatenänderung, bis du
**Änderungen speichern** am Fahrzeug verwendest.

RailKeeper v0.1.18 kann Beilagen anzeigen, die bereits eine Wartungsreferenz enthalten. Die
normale Beilagenzeile bietet jedoch keine Funktion zum Zuweisen oder Ändern dieser Referenz. Nutze
keinen erfundenen Zuordnungsablauf, den die Oberfläche nicht bereitstellt.

## Wartung sicher löschen

**Wartung löschen** entfernt den Eintrag sofort ohne Bestätigungsdialog. Der Server blockiert das
Löschen bei verknüpften Medien nicht und löscht oder trennt diese Medien ebenfalls nicht. Ihre
gespeicherte Wartungs-ID kann deshalb auf einen gelöschten Eintrag zeigen.

Vor dem Löschen eines Eintrags:

1. Speichere oder verwirf bewusst jede andere ausstehende Fahrzeugänderung.
2. Prüfe die Zähler für verknüpfte Bilder und Beilagen.
3. Wähle im Tab **Uploads** für jedes verknüpfte Bild **Keine Wartung** und nutze
   **Änderungen speichern**.
4. Muss eine verknüpfte Beilage ihre Zuordnung behalten, behalte auch den Wartungseintrag.
   RailKeeper v0.1.18 besitzt keinen Editor für Beilagenverknüpfungen. Entferne eine Beilage erst
   nach Sicherungs- und Inhaltsprüfung, wenn die Datei tatsächlich nicht mehr benötigt wird.
5. Kehre zu **Wartung** zurück, prüfe, dass **Verknüpfte Medien** nicht mehr erscheint, und lösche
   erst danach den Eintrag.

Das Löschen der Wartung löscht keine Bild- oder Beilagendatei. Eine veraltete, nicht leere
Wartungs-ID kann das spätere Löschen eines Bildes blockieren, bis die Verknüpfung entfernt und
gespeichert wurde.

## Rollen-, Speicher- und Sicherungsgrenzen

Admin, Editor, Viewer und Planner können Wartungen lesen. Serverseitiges Anlegen, Ändern,
Abschließen und Löschen erfordert Admin oder Editor. Deaktivierte Bedienelemente erklären den
aktuellen Modus, ersetzen aber nicht die serverseitige Berechtigungsprüfung.

Wartungszeilen und Fahrzeugmedien sind lokale RailKeeper-Anwendungsdaten und gehören zum Umfang der
Anwendungssicherung. Erstelle vor umfangreichen Aufräumarbeiten eine aktuelle Sicherung und lasse
sie von einem Admin validieren. Das veröffentlichte Vorgehen für Medien steht unter
[Fahrzeugbilder und Beilagen](/de/guide/vehicles/media). Export, Validierung und Wiederherstellung
gehören zur Administration und werden hier nicht wiederholt.

## Leere und fehlerhafte Zustände

- **Fahrzeug ist nicht gespeichert:** Speichere den Grunddatensatz, bevor du eine Wartung
  hinzufügst.
- **Kein Eintrag vorhanden:** Nutze **Eintrag hinzufügen**, nachdem du das richtige Fahrzeug
  geprüft hast.
- **Eingabe wird abgelehnt:** Prüfe Art, Status, Zustand, Daten, nicht negative Kosten und
  Notizlänge.
- **Schreibaktion schlägt fehl:** Lasse das Formular offen, lies den Fehler, prüfe Sitzung und
  Verbindung und versuche es erneut.
- **Übersicht wirkt veraltet:** Stelle sicher, dass keine ungespeicherten Änderungen bestehen, und
  lade das Fahrzeug neu.
- **Fällig-Zähler passt nicht zum Status:** Fälligkeit wird aus Datum und Abschluss berechnet, nicht
  nur aus dem gespeicherten Status.
- **Verknüpfte Medien werden angezeigt:** Löse die Verknüpfungen vor dem Löschen und nutze die
  Medienseite für Bildmetadaten.
- **Eintrag wurde ohne Nachfrage gelöscht:** Das Löschen wirkt sofort. Wiederherstellung benötigt
  eine geeignete Sicherung.

Eine fehlgeschlagene Aktion macht eine frühere erfolgreiche, unabhängige Wartungs- oder
Medienaktion nicht rückgängig.

## Verwandte Seiten

- [Übersicht des Benutzerhandbuchs](/de/guide/)
- [Übersicht, Kennzahlen und Datenqualität](/de/guide/overview/)
- [Fahrzeugbestand und Grunddaten](/de/guide/vehicles/)
- [Fahrzeugbilder und Beilagen](/de/guide/vehicles/media)
- [Decoder, Funktionen und CV-Daten](/de/guide/vehicles/decoder-cv)
- [Artikelsuche, Web-Dokumente und Ersatzteile](/de/guide/vehicles/search-and-spares)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert die stabile RailKeeper-Version **v0.1.18** und wurde zuletzt am
16.08.2026 geprüft.
