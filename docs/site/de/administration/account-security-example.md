---
title: Konten und Zwei-Faktor-Anmeldung einrichten
description: Konten und Zwei-Faktor-Anmeldung einrichten.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Konten und Zwei-Faktor-Anmeldung einrichten

**Ziel:** Ein Konto mit passenden Rechten anlegen und seine eigene Zwei-Faktor-Anmeldung prüfen.
Benutzerverwaltung benötigt **Admin**. Persönliches Passwort und Zwei-Faktor-Einstellung gehören zum jeweils angemeldeten Konto.

## 1. Ein eigenes Testkonto anlegen

Öffne **Einstellungen > Authentifizierung > Benutzerverwaltung > Neuer Benutzer**.
Trage einen eindeutigen Namen, eine passende E-Mail-Adresse und ein eigenes Passwort mit mindestens
zwölf Zeichen ein. Wiederhole das Passwort und wähle nur die benötigten Rollen.

<GuideFigure name="account-create-viewer" alt="Benutzerformular mit ausschließlich ausgewählter Viewer-Rolle." caption="Tatsächliche Anlage eines getrennten Demo-Kontos. Die Beispieladresse endet auf example.invalid; Passwörter werden nicht veröffentlicht." />

Für die Übung wähle **Viewer** und **Speichern**. Öffne danach eine getrennte Browsersitzung und
melde dieses Konto an. Es soll Bestand lesen können und keine Admin-Sicherung herunterladen dürfen.
Genau diese beiden Grenzen wurden in der Demo geprüft.

| Aufgabe | Geeignete Rolle |
| --- | --- |
| Bestand ansehen | Viewer |
| Fahrzeuge, Zubehör und CVs pflegen | Editor, bei Bedarf zusätzlich Viewer |
| Vorhandene Veranstaltungen bedienen | Messe |
| Veranstaltungen mitplanen | Editor zusammen mit Messe |
| Benutzer, Sicherungen und Zentralen-Schreibabläufe verwalten | Admin |

Ein sichtbarer Einstellungsreiter allein verleiht keine Berechtigung. Der Server prüft jede Aktion.
Eine frühere Planer-Beschreibung gehört nicht zum aktuellen Vier-Rollen-Modell.

## 2. Zwei-Faktor-Anmeldung am eigenen Konto vorbereiten

Melde dich mit dem vorgesehenen Konto an und öffne **Einstellungen > Authentifizierung**.
Betätige **Zwei-Faktor-Auth vorbereiten**. Scanne den QR-Code mit deiner Authenticator-App
oder trage das Secret dort ein. Bewahre die Wiederherstellung deiner Authenticator-App geschützt auf.

<GuideFigure name="account-two-factor-setup" alt="Zwei-Faktor-Einrichtung mit verdecktem QR-Code und Secret sowie leerem Bestätigungscode." caption="Die echte Einrichtung am Demo-Konto. QR-Code und Secret sind schon bei der Aufnahme verdeckt; es werden keine verwendbaren Anmeldeschlüssel veröffentlicht." />

Trage den aktuellen sechsstelligen Code in **Bestätigungscode** ein und wähle **Aktivieren**.
Erst der gültige Code aktiviert die Zwei-Faktor-Pflicht; das Öffnen der Vorbereitung allein genügt nicht.

## 3. Eine frische Anmeldung prüfen

Lasse die vorhandene Sitzung zunächst geöffnet. Öffne ein privates Browserfenster, gib Benutzername
und Passwort ein und wähle **Anmelden**. Nun wird der **Zwei-Faktor-Code** verlangt.

<GuideFigure name="account-two-factor-login" alt="Neue Anmeldung, die zusätzlich den Zwei-Faktor-Code verlangt." caption="Tatsächlich geprüfte Anmeldung: Passwort allein reicht nach Aktivierung nicht; ein gültiger Authenticator-Code ermöglicht die Anmeldung." />

Gib den aktuellen Code ein und kontrolliere, dass das richtige Konto mit denselben Rollen geöffnet wird.
Bei abgewiesenen Codes prüfe Uhrzeit und Kontoeintrag in der Authenticator-App.
Veröffentliche weder Code, QR-Code, Secret noch Sitzungsdaten in einem Hilfethema.

## 4. Deaktivieren und Wiederherstellung verstehen

Zum Deaktivieren brauchst du dein aktuelles Passwort und bei aktiver Zwei-Faktor-Anmeldung einen
gültigen Code. Die Aktion widerruft andere Sitzungen des Kontos.
Ein Passwort-Reset hebt die Zwei-Faktor-Pflicht nicht automatisch auf.
RailKeeper stellt in diesem Ablauf keine eigene Liste einmaliger Wiederherstellungscodes bereit.

Stelle vor Aktivierung sicher, dass du deine Authenticator-Daten wiederherstellen kannst.
Ein Anwendungssicherungs-Restore enthält keine Konten und stellt daher auch dieses Secret nicht wieder her.
Eine vollständige private Serversicherung ist ein anderer Wiederherstellungsumfang.

## 5. Sitzungen kontrollieren

Als Admin prüfe unter **Authentifizierung > Sitzungen**, welche Anmeldungen noch aktiv sind, und
widerrufe eine nicht mehr benötigte Sitzung gezielt. Prüfe Rollenänderungen mit einer frischen Anmeldung.
[Anmeldung und Passwort vergessen](../guide/getting-started/) · [SMTP einrichten](smtp-password-reset).
