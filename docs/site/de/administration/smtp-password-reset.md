---
title: SMTP und Passwort-Reset einrichten
description: SMTP und Passwort-Reset einrichten.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# SMTP und Passwort-Reset einrichten

**Ziel:** Passwort-Rücksetzlinks über den eigenen Maildienst zustellen. Einrichtung und Testversand benötigen **Admin**.
Du brauchst die SMTP-Angaben deines Mailanbieters und eine RailKeeper-Adresse, die der Empfänger erreichen kann.

## 1. Felder ausfüllen

Öffne **Einstellungen > Authentifizierung > SMTP für Passwort-Reset**.

<GuideFigure name="account-smtp-fields" alt="SMTP-Formular mit reservierten Beispieladressen, Verschlüsselungswahl und Testempfänger." caption="Echte Oberfläche, absichtlich nicht zustellfähige example.invalid-Adressen. Es wurde keine Test-Mail versandt; die Abbildung ist keine Mailserver-Abnahme." />

| Feld | Was du einträgst |
| --- | --- |
| Öffentliche RailKeeper-URL | Die vollständige Adresse, über die Benutzer diese Installation erreichen |
| SMTP-Host und Port | Vorgaben deines Mailanbieters; nicht die Adresse einer Digitalzentrale |
| Verschlüsselung | Zum Anbieter passende Einstellung, etwa STARTTLS oder implizites TLS |
| Absender | Für den Maildienst zugelassene Absenderadresse |
| Benutzername und Passwort | Eigene SMTP-Zugangsdaten beziehungsweise ein vorgesehenes App-Passwort |
| Test-Empfänger | Ein von dir kontrolliertes Postfach für den anschließenden Test |

Die Bildwerte sind keine funktionierende Konfiguration. Eine URL mit localhost im Rücksetzlink
würde beim Empfänger dessen eigenen Rechner bezeichnen. Für Zugriff im Netzwerk müssen URL,
HTTPS, Proxy und sichere Cookies zusammenpassen.

## 2. Speichern und gezielt testen

Aktiviere **SMTP aktivieren**, prüfe alle Angaben und wähle **Speichern**.
Wähle danach bewusst **Test-Mail senden**. Diese Aktion verschickt eine echte E-Mail an den
angegebenen Empfänger. Kontrolliere Eingang und gegebenenfalls Spamordner.

Ein leeres Passwortfeld erhält ein bereits gespeichertes Passwort. Der gesonderte Schalter zum
Löschen entfernt es ausdrücklich. Das vorhandene Passwort wird nicht im Klartext zurückgegeben.

## 3. Den Rücksetzablauf prüfen

1. Prüfe, dass das vorgesehene lokale Konto die richtige E-Mail-Adresse besitzt.
2. Öffne eine getrennte Browseranmeldung und **Passwort vergessen?**.
3. Fordere einen Rücksetzlink an und prüfe im eigenen Postfach, dass er auf die richtige Installation zeigt.
4. Öffne den aktuellen Link, wähle ein eigenes neues Passwort und prüfe danach die Anmeldung.

Die Anforderungsbestätigung verrät absichtlich nicht, ob ein Konto zu der Adresse existiert.
Nur die neueste offene Anforderung gilt; Links laufen nach 30 Minuten ab und sind nur einmal verwendbar.
Ein Passwort-Reset hebt eine aktive Zwei-Faktor-Pflicht nicht auf.

## 4. Fehler eingrenzen

| Beobachtung | Prüfen |
| --- | --- |
| Verbindung fehlgeschlagen | SMTP-Host, Port und Erreichbarkeit vom RailKeeper-Server |
| Anmeldung abgelehnt | SMTP-Konto, Passwort und Vorgaben des Anbieters |
| Kein Eingang trotz Bestätigung | Empfänger, Spamordner, Anbieterprotokolle und konkrete Testmeldung |
| Link zeigt falsche Installation | Öffentliche URL, Hostname und Proxy-Konfiguration |
| Link abgelaufen oder bereits benutzt | Eine neue Anforderung stellen und nur den neuesten Link verwenden |

Die Formularzustände sind in der lokalen Demo geprüft. Ein echter SMTP-Versand und ein realer
E-Mail-Rücksetzlink wurden für diese Dokumentation nicht getestet.
[Anmeldehilfe](../guide/getting-started/) · [Installationskonfiguration](deployment-configuration).
