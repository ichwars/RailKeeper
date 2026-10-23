---
title: Bestand exportieren
description: Bestand exportieren.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Bestand exportieren

**Ziel:** Eine überprüfbare Bestandsdatei für Tabellen oder eine andere RailKeeper-Installation erzeugen. Admin, Editor und Viewer dürfen exportieren. Reine Messe-Konten sind auf Ausstellungslisten beschränkt.


<GuideFigure name="import-export" alt="Exportprofile bestimmen Datenbereich und Dateiformat." caption="Exportprofile bestimmen Datenbereich und Dateiformat. Künstliche Demodaten, keine Werkseinstellungen. Anklicken für die volle Größe." />

## Einen Export erstellen

1. Öffne **Import/Export > Profil anlegen** und wähle die Richtung **Export**.
2. Wähle den benötigten Bereich und das Format. Ein CSV-Profil enthält genau **Fahrzeuge** oder **Zubehör**. Ausstellungslisten und mehrere Bereiche benötigen **JSON**.
3. Wähle **Neuer Export** und das Profil; alternativ starte über das aktive Profil.
4. Prüfe Bereiche und Format und wähle **Export ausführen**.
5. Lade das Ergebnis über **Datei herunterladen** herunter.

Der Auftrag und die erzeugte Datei bleiben im Transferverlauf nachvollziehbar. Ein Admin kann Dateien später löschen. Der Speicherort im Arbeitsbereich ist ein Ort auf dem RailKeeper-Server, nicht automatisch ein Ordner auf deinem Browser-PC.

## Welches Format passt?

| Zweck | Format |
| --- | --- |
| Fahrzeug- oder Zubehörfelder im Tabellenprogramm prüfen | CSV |
| Mehrere Bereiche mit Beziehungen in eine andere Installation übertragen | RailKeeper-JSON |
| Fahrzeugsets mit individuellen Mitgliedern übertragen | CSV mit Set-Spalten oder RailKeeper-JSON |
| Lesbare Liste drucken oder als PDF speichern | Report im Fahrzeugbestand beziehungsweise in der Ausstellung |
| Bestand samt Medien wiederherstellen | [Anwendungssicherung](../../administration/backup-restore) |

Eine Suche oder Auswahl im Fahrzeugbestand wird nicht automatisch zur Auswahl eines Transferauftrags. Prüfe den Umfang im Exportdialog.

## Grenzen beachten

Ein Exportpaket ist keine vollständige Sicherung. Fahrzeugmedien, CV-Datensätze, Funktionen, Wartungen und lokale Anmeldedaten gehören nicht in den Fahrzeugtransfer. Die aktuelle JSON-Paketstruktur wird serverseitig versioniert und beim [Import](file-import) geprüft; ändere Kennungen, Paketversionen oder Prüfsummen nicht von Hand.

Behandle Exportdateien als private Bestandsdaten. Preise, Lagerorte und Notizen können darin enthalten sein.
