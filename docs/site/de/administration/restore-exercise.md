---
title: Eine Wiederherstellung sicher üben
description: Eine Wiederherstellung sicher üben.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Eine Wiederherstellung sicher üben

**Ziel:** Eine Sicherung wirklich zurückspielen und ihren Inhalt kontrollieren. Du brauchst
**Admin** und eine **getrennte Testinstallation mit eigenem Datenordner oder Docker-Volume**.
Die Wiederherstellung ersetzt deren Bestand und Medien. Verwende diese Übung nicht auf deinem echten Bestand.

## 1. Einen erkennbaren Übungsbestand erstellen

1. Starte die getrennte Installation und lege ihren eigenen Admin an.
2. Erstelle die Lok aus der [F0-/F1-/F3-Übung](../guide/vehicles/functions-walkthrough).
   Übernimm und speichere ihre drei Funktionen und 144 CV-Werte.
3. Erstelle ein [Set mit zwei Wagen](../guide/vehicles/sets), sodass du insgesamt drei Fahrzeuge hast.
4. Lade das künstliche Lokbild bei der Lok hoch und lege eine offene Wartung an.

<GuideDownload file="locomotive-demo.png">Künstliches Lokbild für die Wiederherstellungsübung</GuideDownload>

Diese Dateien enthalten keine privaten Sammlungs- oder Anmeldedaten. Die Übungssicherung erzeugst
du selbst in RailKeeper. Eine echte Sicherung gehört nicht in ein öffentliches Issue oder Repository.

## 2. Sichern und den Kontrollstand notieren

Öffne **Einstellungen > Backup > Backup erstellen** und lade die JSON-Datei herunter.
Bewahre sie außerhalb des Test-Datenordners auf. Notiere:

| Kontrollpunkt dieser Übung | Erwarteter Sicherungsinhalt |
| --- | --- |
| Fahrzeuge | 3: Lok und zwei Set-Mitglieder |
| Set | 1 mit zwei Mitgliedern |
| CVs der Lok | 144, mit Originalzeilen und Exportdatum |
| Funktionen der Lok | F0, F1, F3, jeweils Dauer und Invertiert Nein, mit Herkunft |
| Medien und Wartung | ein Lokbild und ein Wartungseintrag |
| Anmeldung | gehört zur lokalen Installation und wird nicht aus der Datei übernommen |

Stammdaten und weitere Übungsdaten erhöhen die angezeigte Gesamtzahl der Datensätze.
Diese Zahl ist nicht gleich der Zahl deiner Fahrzeuge.

Lege **nach** der Sicherung ein zusätzliches Fahrzeug namens **NACH DER SICHERUNG** an.
Es dient als sichtbarer Nachweis, dass tatsächlich der frühere Stand wiederhergestellt wird.

## 3. Datei prüfen, ohne wiederherzustellen

Wähle die gespeicherte JSON unter **Backup wiederherstellen**.

<GuideFigure name="restore-preflight" alt="Kompatible Übungssicherung mit Version, Tabellen, Datensätzen und Dateien vor der Wiederherstellung." caption="Echte Prüfung einer selbst exportierten Demo-Sicherung. Die Dateiauswahl ändert noch keine Bestandsdaten; die Bestätigung fehlt noch." />

Prüfe Kompatibilität, Dateianzahl, Fehler und Warnungen. Beim geprüften Stand lautet das
Sicherungsformat Version **24**. Die Dateiauswahl allein spielt nichts zurück.

## 4. Wiederherstellung bewusst bestätigen

Gib **WIEDERHERSTELLEN** ein. Wähle **Backup einspielen** und lies die abschließende Rückfrage.

<GuideFigure name="restore-confirmation" alt="Letzte Rückfrage vor dem Einspielen der kompatiblen Übungssicherung." caption="Erst diese Bestätigung löst die Wiederherstellung aus. Der Test nutzt ausschließlich die getrennte Demo." />

Bestätige nur auf der vorgesehenen Testinstallation und warte auf das Ergebnis.
Wenn die Antwort ausbleibt, prüfe erst den tatsächlichen Bestand und die Protokolle, bevor du wiederholst.

## 5. Den wiederhergestellten Stand prüfen

<GuideFigure name="restore-result" alt="Fahrzeugbestand nach dem tatsächlichen Wiederherstellen der Demo-Sicherung." caption="Die Lok und beide Set-Mitglieder sind vorhanden. Das erst nach der Sicherung angelegte Kontrollfahrzeug ist verschwunden." />

- Melde dich bei Bedarf wieder mit dem lokalen Testkonto an.
- Das Fahrzeug **NACH DER SICHERUNG** muss fehlen; Lok und Set müssen vorhanden sein.
- Prüfe die 144 CVs einschließlich Indexseiten und Herkunft.
- Öffne F0/F1/F3 und ihre gespeicherten Nachweise.
- Öffne das Bild in voller Größe und prüfe die Wartung.
- Erstelle anschließend eine neue Sicherung und prüfe auch diese Datei.

Diese Schritte wurden mit echten Export-/Prüf-/Restore-Aktionen der isolierten Demo in beiden
Sprachen durchgeführt. Sie belegen keinen vollständigen Server-Wiederanlauf und keinen Windows- oder Docker-Migrationstest.

## Eine inkompatible Datei erkennen

<GuideFigure name="restore-incompatible" alt="Abgewiesene Übungssicherung mit nicht unterstützter Versionsnummer und gesperrtem Einspielen." caption="Negativtest: Eine Kopie mit absichtlich ungültiger Versionsnummer wird vor dem Restore abgewiesen. Das Original bleibt unverändert." />

Bei einer nicht unterstützten Version bleibt **Backup einspielen** gesperrt. Beschaffe eine passende
RailKeeper-Version oder die richtige Sicherung. Ändere bei echten Sicherungen nicht die Versionsnummer,
um die Prüfung zu umgehen. Beschädigte oder unvollständige Dateien benötigen ebenfalls Klärung.

[Welche Sicherung brauche ich?](backup-restore) · [Vor und nach Updates prüfen](system-operations)

Die Anzeige **Dateien: 0** schließt Bilder nicht aus: Im aktuellen Speicher liegen Medien auch als Datenbankinhalte vor und werden mit den Tabellen gesichert. Prüfe nach dem Restore deshalb das tatsächlich geöffnete Bild.
