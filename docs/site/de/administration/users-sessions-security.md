---
title: Konten und Zugriffsrechte
description: Konten und Zugriffsrechte.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Konten und Zugriffsrechte

**Ziel:** Jeder Person nur die passenden Möglichkeiten geben. Konten verwaltet ein **Admin** unter **Einstellungen > Authentifizierung**.


<GuideFigure name="users" alt="Authentifizierung verbindet persönliche Sicherheit mit den Admin-Werkzeugen." caption="Authentifizierung verbindet persönliche Sicherheit mit den Admin-Werkzeugen. Künstliche Demodaten, keine Werkseinstellungen. Anklicken für die volle Größe." />

## Ein Konto anlegen

1. Öffne die Benutzerverwaltung und wähle die Aktion zum Anlegen eines Benutzers.
2. Trage Benutzername, E-Mail und ein eigenes Passwort ein.
3. Wähle nur die benötigten Rollen.
4. Speichere das Konto und prüfe die Anmeldung getrennt vom eigenen Administratorkonto.

| Rolle | Wofür sie gedacht ist |
| --- | --- |
| Admin | Installation, Konten, Sicherungen und geschützte Zentralenaktionen verwalten |
| Editor | Bestandsdaten anlegen und bearbeiten |
| Viewer | Bestand ansehen und freigegebene Ausgaben verwenden |
| Messe | den zugelassenen Ausstellungsarbeitsbereich bedienen |

Mehrere Rollen können kombiniert werden. Die Rollen eines Kontos bestimmen, welche Menüs und Aktionen erlaubt sind; ein sichtbarer Link umgeht keine Serverprüfung.

## Eigenes Konto schützen

Verwende die persönlichen Passwort- und Zwei-Faktor-Einstellungen im Bereich **Authentifizierung**. Richte eine Authenticator-App ein und bewahre erforderliche Wiederherstellungsangaben geschützt auf. [Anmeldung und Passwort vergessen](../guide/getting-started/).

## Bei verlorenen oder fremden Sitzungen

Ein Admin prüft die Sitzungsübersicht und widerruft nicht mehr benötigte Sitzungen. Ein Passwortwechsel oder administrativer Reset kann bestehende Sitzungen ungültig machen. Prüfe nach einer Rechteänderung mit einer neuen Anmeldung, was das betroffene Konto tatsächlich öffnen darf.

Benutzer und Anmeldedaten werden nicht in Anwendungssicherungen übernommen. Für den vollständigen Wiederanlauf einer Installation benötigt der Betreiber zusätzlich eine private Serversicherung.

[Bebilderte Konto- und Zwei-Faktor-Übung](account-security-example) · [SMTP und Passwort-Reset](smtp-password-reset).
