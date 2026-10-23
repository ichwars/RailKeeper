---
title: Einstellungen
description: Persönliche Einstellungen, Darstellung und die Abgrenzung zur Administration verstehen.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Einstellungen

Öffne **Einstellungen**. Unter **Allgemein** findest du persönliche Vorgaben, unter **Darstellung** Farben und Designmodus. **Backup**, **Digitalzentralen** und **Authentifizierung** haben eigene Anleitungen und benötigen je nach Aktion Admin-Rechte.

<GuideFigure name="preferences" alt="Persönliche Vorgaben und administrative Werkzeuge sind getrennte Einstellungsbereiche." caption="Persönliche Vorgaben und administrative Werkzeuge sind getrennte Einstellungsbereiche. Künstliche Demodaten, keine Werkseinstellungen. Anklicken für die volle Größe." />

Der Arbeitsbereich **Einstellungen** verbindet persönliche Einstellungen mit mehreren
Administrationswerkzeugen. Dieses Kapitel behandelt die Einstellungen für Navigation und
Darstellung eines einzelnen Benutzers. Stammdaten, Digitalzentralen, Anwendungssicherung, Speicher,
Updates, Benutzer, Sitzungen und Authentifizierung sind eigene Themen, obwohl RailKeeper sie im
selben Arbeitsbereich anzeigt.

Dieses Kapitel dokumentiert RailKeeper.

## Zugriffsrechte

Admin, Editor, Viewer können **Einstellungen** öffnen. Ein reines Messe-Konto bleibt im
isolierten Messearbeitsbereich und kann diese Seite nicht öffnen. Persönliche Profileinstellungen
verwenden die angemeldete Profil-API und sind nach Benutzer getrennt.

| Aktion | Admin | Editor | Viewer |
| --- | --- | --- | --- |
| **Allgemein** und **Darstellung** öffnen | Ja | Ja | Ja |
| Persönliche Profileinstellungen ändern | Ja | Ja | Ja |
| Benannte Systemdrucker lesen | Ja | Nein | Nein |
| Ausschließlich für Admins erlaubte Vorgänge ausführen | Ja | Nein | Nein |

Die sichtbare Seite ist nicht die Berechtigungsgrenze. RailKeeper prüft geschützte System- und
Datenvorgänge auf dem Server. Das Verschieben oder Ausblenden eines Navigationseintrags erteilt oder
entfernt niemals Berechtigungen.

## Den richtigen Einstellungsbereich wählen

| Register oder Bereich | Zweck | Zuständiges Kapitel |
| --- | --- | --- |
| **Allgemein** | Sprache, Startseite, Datums- und Zeitvorgaben, Druckvorgabe und Seitenleisten-Reihenfolge | [Persönliche Einstellungen](/de/guide/settings/personal-preferences) |
| **Darstellung** | System-, Hell- oder Dunkelmodus sowie getrennte Farb- und Stilvarianten | [Darstellung](/de/guide/settings/appearance) |
| **Allgemein > Inventarnummern** | Nummernschemata für Bestandskategorien | [Stammdaten-Administration](/de/administration/master-data) |
| **Allgemein > Artikelsuche** | Artikelsuche aktivieren und Quellgruppen wählen | [Artikelsuche](/de/guide/vehicles/search-and-spares) |
| **Allgemein > Updates** | Versionsprüfung und Download des Windows-Pakets | Releases und Support |
| **Allgemein > Speicher** | Speichernutzung und Optimierung | Systembetrieb |
| **Daten** | Allgemeine und Zubehör-Stammdaten, Lebenszyklus und Transfer | [Stammdaten-Administration](/de/administration/master-data) |
| **Digitalzentralen** | ECoS-, Z21-, Intellibox-3- und CS3-Konfiguration | [Digitalzentralen-Administration](/de/administration/digital-centers) |
| **Backup** | Sicherung, Wiederherstellung und Stammdaten-Transfer innerhalb der Einstellungen | Sicherungs- und [Stammdaten-Administration](/de/administration/master-data) |
| **Authentifizierung** | Eigenes Passwort und Zwei-Faktor-Einrichtung; Admin-Werkzeuge für Benutzer, Sitzungen, Audit und SMTP | [Benutzer, Sitzungen und Sicherheit](/de/administration/users-sessions-security) |

Der allgemeine [Arbeitsbereich Import und Export](/de/guide/import-export/) ist eine andere Seite. Er
tauscht Bestandsdaten aus. Der ECoS-Lokablauf liegt unter **Digitalzentralen**. Das Register **Backup** in den **Einstellungen** enthält administrative Sicherung und Stammdaten-Transfer. [Sicherungsanleitung](/de/administration/backup-restore).

## Speicherung persönlicher Einstellungen

Die meisten Vorgaben wirken sofort im Browser, werden im Browserspeicher abgelegt und zusätzlich an
das Profil des aktuellen Benutzers gesendet. Die Anwendungshülle stellt Designmodus und
Seitenleistenvorgaben wieder her. Beim Öffnen der Einstellungen folgen die weiteren synchronisierten
Vorgaben.

Dabei gelten wichtige Ausnahmen und Grenzen:

- Die Oberflächensprache ist derzeit nur im Browser gespeichert und wird nicht in das
  Serverprofil geschrieben.
- Der ein- oder ausgeklappte Zustand der Hauptseitenleiste ist browserlokal. Reihenfolge und
  ausgeblendete Einträge sind Profileinstellungen.
- Das Speichern des Profils läuft im Hintergrund. Für diese kleinen Schreibvorgänge zeigt die Seite
  keine Erfolgs- oder Fehlermeldung. Scheitert der Serverzugriff, kann der lokale Browserwert
  trotzdem aktiv bleiben.
- Anwendungssicherungen schließen Benutzerkonten und `user_settings` absichtlich aus. Eine
  Sicherung ist daher kein Export persönlicher Einstellungen.

## Sicherer Arbeitsablauf

1. **Einstellungen > Allgemein** öffnen und gewünschte Sprache sowie stabile Startseite wählen.
2. Seitenleiste anordnen, häufig verwendete Arbeitsbereiche aber sichtbar lassen.
3. **Darstellung** öffnen und **System**, **Hell** oder **Dunkel** wählen.
4. Helle und dunkle Variante bei Bedarf getrennt anpassen.
5. RailKeeper einmal neu laden, um die Wiederherstellung einer Vorgabe zu prüfen.
6. Bei einem anderen Browser mit demselben Benutzer anmelden. Die Oberflächensprache erneut setzen,
   weil sie derzeit nicht über das Profil synchronisiert wird.
