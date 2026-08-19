---
title: Ersteinrichtung und Anmeldung
description: Ersten Administrator anlegen, sicher anmelden und den Zugang zu RailKeeper wiederherstellen.
audience: user
status: stable
reviewedVersion: 0.1.19.2
lastReviewed: 2026-08-16
---

# Ersteinrichtung und Anmeldung

Dieses Kapitel erklärt das erste Administratorkonto, die normale Anmeldung und Anmeldung mit
Zwei-Faktor-Code, die Abmeldung sowie die Passwort-Wiederherstellung. Es beschreibt den stabilen
RailKeeper-Stand v0.1.19.2.

## Voraussetzungen

RailKeeper muss bereits installiert und im Browser erreichbar sein. Das Formular zur
Ersteinrichtung erscheint nur, solange die RailKeeper-Datenbank kein Benutzerkonto enthält.
RailKeeper besitzt keinen vorgegebenen Benutzernamen und kein Standardpasswort.

Erscheint stattdessen die Anmeldung, wurde die Ersteinrichtung bereits abgeschlossen. Melde dich
mit einem bestehenden Konto an oder wende dich an die Person, die diese Instanz betreibt.

## Ersten Administrator anlegen

Für die einmalige Ersteinrichtung ist kein vorhandenes Konto und keine Rolle erforderlich.

1. Öffne RailKeeper im Browser.
2. Fülle alle Felder unter **Ersteinrichtung** aus.

   | Feld | Anforderung | Zweck |
   | --- | --- | --- |
   | Benutzername | Erforderlich, nach Entfernen äußerer Leerzeichen mindestens 3 Zeichen | Identifiziert das Konto bei der Anmeldung |
   | E-Mail-Adresse | Erforderlich, gültige E-Mail-Adresse | Empfängt Wiederherstellungslinks, wenn der E-Mail-Versand eingerichtet ist |
   | Passwort | Erforderlich, mindestens 12 Zeichen | Schützt das erste Konto |
   | Passwort wiederholen | Erforderlich und identisch mit **Passwort** | Verhindert einen unbemerkten Tippfehler |

3. Wähle **Admin erstellen**.
4. Verwende nach erfolgreicher Ersteinrichtung das normale Anmeldeformular mit den neuen
   Zugangsdaten.

Das erste Konto erhält die Rollen Admin, Editor und Viewer. Die Ersteinrichtung kann keinen zweiten
ersten Administrator anlegen. Zum Schutz vor wiederholten Versuchen akzeptiert RailKeeper höchstens
fünf Einrichtungsversuche pro Client-Adresse innerhalb von zehn Minuten.

## Anmelden

1. Gib **Benutzername** und **Passwort** ein.
2. Wähle **Anmelden**.
3. Ist für das Konto die Zwei-Faktor-Authentifizierung aktiviert, zeigt RailKeeper jetzt das Feld
   **Zwei-Faktor-Code** an. Gib den aktuellen Code aus der eingerichteten Authenticator-App ein und
   wähle erneut **Anmelden**.

Eine erfolgreiche Anmeldung erzeugt eine serverseitige Sitzung mit einer Laufzeit von bis zu zwölf
Stunden. Benutzer mit den Rollen Admin, Editor, Viewer oder Planner starten in der **Übersicht**.
Ein Konto, das ausschließlich die Rolle Messe besitzt, startet unter **Ausstellung**.

RailKeeper akzeptiert höchstens zehn Anmeldeanfragen von einer Client-Adresse innerhalb von fünf
Minuten. Warte vor dem nächsten Versuch, statt das Formular wiederholt abzusenden.

## Abmelden

Verwende das Abmeldesymbol in der Fußzeile der Seitenleiste. RailKeeper widerruft die aktuelle
serverseitige Sitzung und kehrt zum Anmeldeformular zurück. Das bloße Schließen des Browser-Tabs
meldet die Sitzung nicht ausdrücklich ab.

## Vergessenes Passwort wiederherstellen

Die Passwort-Wiederherstellung verwendet die für das Konto gespeicherte E-Mail-Adresse.

1. Wähle im Anmeldeformular **Passwort vergessen?**.
2. Gib die E-Mail-Adresse des Kontos ein.
3. Wähle **Reset anfordern**.
4. Öffne den neuesten empfangenen Wiederherstellungslink.
5. Gib ein neues Passwort mit mindestens 12 Zeichen ein und wiederhole es.
6. Setze das Passwort, kehre zur Anmeldung zurück und verwende die neuen Zugangsdaten.

Die im Anmeldeformular angezeigte Bestätigung und die HTTP-Antwort sind für bekannte und unbekannte
Adressen absichtlich identisch. Der Reset-Ablauf legt dadurch nicht offen, ob ein Konto existiert.

Bei eingerichtetem SMTP-Versand sendet RailKeeper den Link per E-Mail. Ohne SMTP werden Reset-Token
weder an den Browser zurückgegeben noch ins Serverprotokoll geschrieben. Wende dich an einen
Administrator oder Betreiber, wenn keine Nachricht ankommt.

Nur die neueste offene Reset-Anfrage bleibt gültig. Ein Wiederherstellungslink läuft nach 30
Minuten ab und kann einmal verwendet werden. Der erfolgreiche Reset widerruft alle bestehenden
Sitzungen dieses Benutzers, auch in anderen Browsern.

RailKeeper begrenzt Reset-Anfragen auf fünf pro Client-Adresse innerhalb von zehn Minuten und
Reset-Bestätigungen auf zehn innerhalb von zehn Minuten.

## Fehlerbehebung

| Problem | Ursache und Maßnahme |
| --- | --- |
| Das Formular zur Ersteinrichtung erscheint nicht | Es existiert bereits mindestens ein Benutzer. Verwende die Anmeldung oder wende dich an den Betreiber. |
| Die Ersteinrichtung wird abgelehnt | Prüfe die Mindestlänge von 3 Zeichen für den Benutzernamen, das E-Mail-Format, die Mindestlänge von 12 Zeichen für das Passwort und beide Passworteingaben. Warte nach wiederholten Versuchen das zehnminütige Begrenzungsfenster ab. |
| Die Anmeldung meldet ungültige Zugangsdaten | Prüfe Benutzername und Passwort. RailKeeper verwendet für ein unbekanntes Konto und ein falsches Passwort absichtlich dieselbe Meldung. |
| RailKeeper verlangt einen Zwei-Faktor-Code | Für das Konto ist Zwei-Faktor-Authentifizierung aktiviert. Gib den aktuellen Code aus der Authenticator-App ein. Ein falscher oder abgelaufener Code wird als ungültige Zugangsdaten gemeldet. |
| Es kommt keine E-Mail zur Passwort-Wiederherstellung an | Prüfe Adresse und Spam-Ordner und wende dich danach an einen Administrator. SMTP ist möglicherweise nicht eingerichtet, und die im Formular angezeigte Bestätigung bestätigt nicht, dass das Konto existiert. |
| Ein Reset-Link ist ungültig oder abgelaufen | Fordere einen neuen Link an. Frühere Links werden durch eine neuere Anfrage, nach 30 Minuten oder nach der ersten Verwendung ungültig. |
| RailKeeper meldet zu viele Versuche | Sende das Formular nicht weiter ab und warte das zutreffende fünf- oder zehnminütige Begrenzungsfenster ab. |

## Sicherheitshinweise

- Verwende ein nur für RailKeeper genutztes, nicht wiederverwendetes Passwort und teile keine
  Administrator-Zugangsdaten.
- Verwende HTTPS und sichere Cookies, wenn die Instanz über ein Netzwerk erreichbar ist. Die
  Betriebsanforderungen stehen unter [Installation und Administration](/de/administration/).
- RailKeeper verwendet für bekannte und unbekannte Adressen dieselbe Reset-Antwort. Überwache
  trotzdem wiederholte Reset-Anfragen und halte den Netzwerkzugriff so eng wie möglich.
- Bitte einen Administrator, unerwartete Anmelde- oder Wiederherstellungsaktivität zu prüfen, statt
  weiterhin Zugangsdaten zu erraten.

## Verwandte Seiten

- [Überblick des Benutzerhandbuchs](/de/guide/)
- [Installation und Administration](/de/administration/)

## Dokumentierter RailKeeper-Stand

Diese Seite dokumentiert den stabilen RailKeeper-Stand **v0.1.19.2** und wurde zuletzt am
2026-08-16 mit der Anwendung abgeglichen.
