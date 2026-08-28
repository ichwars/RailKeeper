---
title: Reservierungen, Einbauten und Verwendung
description: Zubehör reservieren, Einbauten erfassen und die Verwendungshistorie verstehen.
audience: user
status: stable
reviewedVersion: 0.1.20.3
lastReviewed: 2026-08-16
---

# Reservierungen, Einbauten und Verwendung

Reservierungen ordnen verfügbaren Zubehörbestand einer künftigen Verwendung zu. Einbauten halten
fest, dass der Bestand das Lager physisch verlassen hat und verwendet wird. RailKeeper hält beide
Schritte, Einzelstück-Lebenszyklus, Bestandsbewegungen, Zustände und Verwendungshistorie in einer
lokalen Transaktion konsistent.

## Rollen und Zuordnungsziele

Administratoren und Bearbeiter können reservieren, stornieren, einbauen, Zustände ändern und
Einbauten ausbauen. Ein Planer kann im ansonsten schreibgeschützten Artikeldialog Reservierungen
anlegen und stornieren, aber weder einbauen, ausbauen, Zustände ändern noch allgemeinen Bestand
verwalten. Betrachter besitzen Leserecht. Messe hat keinen allgemeinen Zubehörzugriff.

Jede Reservierung und jeder Einbau benötigt genau ein Ziel:

| Ziel | Auswahl im stabilen Dialog |
| --- | --- |
| Fahrzeug | Inventarnummer und Fahrzeugname. |
| Anlage | Eine nicht archivierte Anlage. |
| Anlageneinheit | Eine nicht archivierte Anlageneinheit. |

Der Dialog verwendet vorhandene Ziele, legt aber keine Anlagen an und bearbeitet sie nicht. Dieses
Kapitel dokumentiert weder das Anlegen noch das Bearbeiten von Anlagen.

Fünf optionale Felder begleiten beide Abläufe: **Platzierung**, **Digitaladresse**,
**Decoderausgang**, **Anschluss** und **Verdrahtungshinweise**. Sie beschreiben, wo und wie das
Zubehör vorgesehen beziehungsweise tatsächlich angeschlossen ist. **Notizen** sind davon getrennte
private Arbeitsnotizen.

## Zuordnungsübersicht lesen

Übersicht und Zuordnungsdienst leiten abhängig von der Bestandsstrategie diese Mengen ab:

| Wert | Stabile Bedeutung |
| --- | --- |
| Eigenbestand | Weiterhin eigene physische Einheiten: eingelagerte plus aktiv eingebaute Mengeneinheiten oder alle Einzelstücke. Hybrid kombiniert Mengenbestand, Einzelstücke und aktive Mengeneinbauten, ohne individualisierte Stücke doppelt zu zählen. |
| Eingelagert | Mengenbestand plus, soweit unterstützt, Einzelstücke an einem Lagerort mit Lebenszyklus Eingelagert oder Reserviert. |
| Reserviert | Menge aller aktiven Reservierungen, einschließlich eins je reserviertem Einzelstück. |
| Eingebaut | Menge aller noch nicht ausgebauten Einbauten. |
| Verfügbar | `max(Eingelagert - Reserviert, 0)`. |
| Fehlend | Betrag, um den Reserviert größer als Eingelagert ist. Normale Befehle verhindern dies, wiederhergestellte oder historische Daten können es zeigen. |

Bei Mengenbestand senkt ein Einbau Eingelagert, bleibt aber bis zum Ausbau im Eigenbestand. Ein
Einzelstück in Wartung oder Ausgemustert bleibt Eigentum, zählt aber nicht als Eingelagert oder
Verfügbar.

## Bestand reservieren

1. Wählen Sie Fahrzeug, Anlage oder Anlageneinheit und das genaue Ziel.
2. Erfassen Sie optional die fünf technischen Platzierungsfelder.
3. Wählen Sie bei einem Hybridartikel **Menge** oder **Einzelstück** als Quelle.
4. Wählen Sie einen aktiven Lagerort. Bei Einzelverwaltung wählen Sie ein dort eingelagertes
   Einzelstück, die Menge ist fest eins. Andernfalls geben Sie eine positive ganze Menge ein.
5. Ergänzen Sie optional eine Notiz, wählen Sie **Reservierung anlegen** und bestätigen Sie.

Die bestätigte Transaktion prüft Artikel, Ziel, aktiven Lagerort, Strategie und verfügbare Quelle.
Eine Mengenreservierung erzeugt eine aktive Reservierung und senkt nur Verfügbar. Sie ändert weder
die physische Menge noch das Bestandsjournal. Eine Einzelstückreservierung ändert zusätzlich den
Lebenszyklus von Eingelagert zu Reserviert, behält aber den Lagerort. Dasselbe Einzelstück kann
nicht erneut aktiv reserviert werden.

Ist die verfügbare Menge zu klein, gehört das Einzelstück zu einem anderen Artikel oder Lagerort,
ist es nicht Eingelagert oder ist ein Ziel ungültig, scheitert die ganze Reservierung ohne
Teilzustand.

## Reservierung stornieren

Nur eine aktive Reservierung bietet **Stornieren** an. Der Befehl öffnet eine Bestätigung. Die
Bestätigung setzt den Status auf Storniert und lädt die zugehörigen Ressourcen neu. Mengenbestand
bleibt physisch unverändert und wird wieder verfügbar. Ein reserviertes Einzelstück wechselt am
bestehenden Lagerort von Reserviert zu Eingelagert.

Eine erfüllte oder bereits stornierte Reservierung ist unveränderlich. Erfüllt bedeutet, dass ein
Einbau sie verbraucht hat. Bauen Sie diesen Einbau aus, statt die Reservierung wieder zu öffnen.

## Einbau erfassen

Ein Administrator oder Bearbeiter öffnet den Artikel zum Bearbeiten und wählt entweder eine aktive
Reservierung oder **Ohne Reservierung**.

Mit Reservierung sind Ziel, Quelllagerort, gegebenenfalls Einzelstück und Menge an die Reservierung
gebunden. Ihre Platzierungsdaten werden in das Einbauformular übernommen. Ein nicht leerer
Einbauwert kann einen geerbten Platzierungswert ersetzen, ein leerer Wert erhält den
Reservierungswert. Zustand und Einbaunotizen bleiben wählbar.

Ohne Reservierung wählen Sie Ziel, optionale Platzierungsdaten, Quelle, positive ganze Menge oder
ein eingelagertes Einzelstück, Zustand und optionale Notizen. Hybridartikel erlauben erneut Menge
oder Einzelstück. Der direkte Einbau eines Einzelstücks wird abgelehnt, wenn es aktiv reserviert ist.

**Einbau erfassen** öffnet eine Bestätigung. Die erfolgreiche Transaktion wirkt so:

| Quelle | Physischer Bestand oder Einzelstück | Reservierung | Einbau und Journal |
| --- | --- | --- | --- |
| Menge | Zieht die Menge am Quelllagerort ab. Ein direkter Einbau schützt alle aktiven Reservierungen dieses Ortes. | Gewählte aktive Reservierung wird Erfüllt. | Erzeugt einen aktiven Einbau und eine Bewegung Einbau mit negativer Menge. |
| Einzelstück | Setzt das gewählte eingelagerte oder reservierte Einzelstück auf Eingebaut, entfernt den Lagerort und setzt den gewählten Zustand. | Gewählte aktive Reservierung wird Erfüllt. | Erzeugt einen aktiven Einbau. Es entsteht keine Mengenbewegung. |

Gültige Einbauzustände sind Einsatzbereit, Wartung fällig, Defekt und Unbekannt. Die manuelle
Oberfläche beginnt mit Einsatzbereit. Änderungen an Bestand, Einzelstück, Reservierung, Einbau,
Bewegung und Audit sind atomar. Scheitert eine Prüfung, bleibt nichts davon erhalten.

## Einbauzustand ändern

Wählen Sie bei einem aktiven Einbau Einsatzbereit, Wartung fällig, Defekt oder Unbekannt und danach
**Zustand speichern**. Auch bei scheinbar unveränderter Auswahl erscheint eine Bestätigung.

Die Bestätigung aktualisiert den Einbau, ergänzt den Zustandsverlauf um vorherigen und neuen Wert
und übernimmt den Zustand bei einem eingebauten Einzelstück. Bei Mengenbestand existiert kein
Einzelstück zum Aktualisieren. Ein ausgebauter Einbau lässt sich nicht mehr ändern. Nach Erfolg
werden die Ressourcen neu geladen.

## Einbau ausbauen

Wählen Sie **Ausbauen** an einem aktiven Einbau, bestimmen Sie den Verbleib, ergänzen Sie optionale
Ausbaunotizen und bestätigen Sie den eigenen Ausbaudialog.

| Verbleib | Ergebnis bei Menge | Ergebnis beim Einzelstück |
| --- | --- | --- |
| Eingelagert | Erfordert einen aktiven Zielort, führt die vollständige Menge zurück und schreibt eine positive Bewegung Ausbau. | Lebenszyklus Eingelagert, gewählter Lagerort, Zustand bleibt erhalten. |
| In Wartung | Führt keine Menge in den Bestand zurück. | Lebenszyklus Wartung, kein Lagerort, Zustand bleibt erhalten. |
| Defekt | Führt keine Menge in den Bestand zurück. | Lebenszyklus Wartung, kein Lagerort, Zustand wird Defekt. |
| Ausgemustert | Führt keine Menge in den Bestand zurück. | Lebenszyklus Ausgemustert, kein Lagerort, Zustand bleibt erhalten. |

Der Ausbau schließt den Einbau mit Bearbeiter, Zeit, Verbleib und getrennten Ausbaunotizen. Er
löscht ihn nicht. Ein geschlossener Einbau kann weder erneut ausgebaut noch im Zustand geändert
werden. Bestands- oder Einzelstückänderung, Abschluss, gegebenenfalls Bewegung und Audit bilden
eine Transaktion.

## Verwendungshistorie lesen

Der Reiter **Verwendungshistorie** erscheint, sobald Reservierungs-, Einbau- oder Verlaufsdaten
vorliegen. Sein oberer Bereich wiederholt nur aktuelle aktive Reservierungen und noch nicht
ausgebaute Einbauten schreibgeschützt.

Danach zeigt die Historie Ereignisse vom ältesten zum neuesten mit Datum und Uhrzeit, Ereignisart,
Menge, Ziel und Zustand oder Verbleib:

| Ereignis | Entsteht bei |
| --- | --- |
| Reservierung | Anlegen der Reservierung. Ihr Datensatz trägt später Aktiv, Erfüllt oder Storniert. Die Stornierung ist keine eigene Ereigniszeile. |
| Einbau | Anlegen des Einbaus. |
| Zustandsänderung | Bestätigter Zustand eines aktiven Einbaus, vorheriger und neuer Zustand bleiben erhalten. |
| Ausbau | Schließen des Einbaus mit seinem endgültigen Verbleib. |

Einbau- und Ausbaunotizen sowie detaillierte Platzierungsfelder bleiben gespeichert, werden in der
kompakten stabilen Historientabelle jedoch nicht angezeigt. Verwenden Sie den aktuellen
Zuordnungsdatensatz und Sicherungen, wenn diese Angaben betrieblich wichtig sind.

## Bestand und Lebenszyklus konsistent halten

Verwenden Sie den Ablauf, der das reale Ereignis beschreibt. Ersetzen Sie einen Einbau nicht durch
eine negative Bestandskorrektur und einen Ausbau nicht durch eine positive. Solche Abkürzungen
ließen Ziel, Einbaudatensatz, Einzelstück-Lebenszyklus und Verwendungshistorie aus.

| Aktion | Mengenverfügbarkeit | Einzelstück-Lebenszyklus | Dauerhafter Verlauf |
| --- | --- | --- | --- |
| Reservieren | Verfügbar sinkt, Eingelagert unverändert. | Eingelagert zu Reserviert. | Reservierungsdatensatz und -ereignis. |
| Stornieren | Verfügbar wird frei, Eingelagert unverändert. | Reserviert zu Eingelagert. | Reservierung bleibt Storniert. |
| Einbauen | Eingelagert sinkt, Eingebaut steigt. | Eingelagert/Reserviert zu Eingebaut, Lagerort entfällt. | Einbauereignis und gegebenenfalls Mengenbewegung. |
| Zustand ändern | Keine Mengenänderung. | Zustand des eingebauten Stücks folgt dem Einbau. | Ereignis Zustandsänderung. |
| Eingelagert ausbauen | Eingelagert steigt, Eingebaut sinkt. | Eingebaut zu Eingelagert am Zielort. | Ausbauereignis und gegebenenfalls Mengenbewegung. |
| Anders ausbauen | Eingebaut sinkt ohne Rückkehr in Mengenbestand. | Eingebaut zu Wartung oder Ausgemustert, Defekt setzt zusätzlich den Zustand. | Ausbauereignis. |

Eine Änderung der Artikel-Bestandsstrategie ist gesperrt, solange bestehende Mengen- oder
Einzelstückabhängigkeiten danach nicht mehr darstellbar wären. Klären Sie aktive und historische
Abhängigkeiten bewusst, statt eine Klassifikationsänderung zu erzwingen.

## Zuordnungsfehler beheben

| Situation | Nächster Schritt |
| --- | --- |
| Speichern ist deaktiviert | Wählen Sie genau ein Ziel und einen aktiven Lagerort, bei Einzelverwaltung zusätzlich ein eingelagertes Einzelstück. |
| Bestand reicht nicht | Prüfen Sie aktive Reservierungen an der Quelle und reduzieren Sie die Menge oder wählen Sie einen anderen Lagerort. |
| Einzelstückkonflikt | Prüfen Sie Artikel, Lagerort, Lebenszyklus Eingelagert/Reserviert und eine andere aktive Reservierung oder einen Einbau. |
| Reservierung lässt sich nicht stornieren | Nur aktive Reservierungen sind stornierbar. Eine erfüllte gehört zu einem Einbau. |
| Einbau aus Reservierung kollidiert | Ändern Sie Ziel, Lagerort, Einzelstück und Menge nicht. Wählen Sie die aktuelle aktive Reservierung erneut. |
| Verbleib Eingelagert ist gesperrt oder fehlerhaft | Wählen Sie einen aktiven Zielort mit ebenfalls aktiver übergeordneter Kette. Andere Verbleibe dürfen keinen Ort enthalten. |
| Zustand oder Ausbau kollidiert | Laden Sie neu. Eine andere Aktion kann den Einbau bereits geändert oder geschlossen haben. |
| Schreiben möglicherweise erfolgreich, Neuladen fehlgeschlagen | Wiederholen Sie nichts. Verwenden Sie **Erneut versuchen** und prüfen Sie Status, Einbau, Einzelstück, Journal und Historie. |
| Berechtigung fehlt | Planer dürfen nur reservieren und stornieren. Einbau, Zustand und Ausbau erfordern Administrator oder Bearbeiter. |

Ein fehlgeschlagenes Neuladen nach dem Schreiben markiert den Editor als veraltet und sperrt
weitere Zuordnungs- und Bestandsaktionen, bis ein vollständiger Wiederholungsversuch gelingt. So
verwendet kein zweiter Befehl einen alten Verfügbarkeits- oder Lebenszyklusstand.

## Verwandte Seiten

- [Überblick zum Benutzerhandbuch](/de/guide/)
- [Zubehörübersicht](./)
- [Artikelstammdaten und Fachangaben](./article-records)
- [Bestand, Käufe und Dokumente](./stock-purchases-documents)
- [Fahrzeugbestand und Stammdaten](../vehicles/)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert RailKeeper **v0.1.20.3** und wurde zuletzt am 16.08.2026 geprüft.
