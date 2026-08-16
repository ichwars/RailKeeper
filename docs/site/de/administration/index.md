---
title: Installation und Administration
description: RailKeeper installieren, konfigurieren, absichern und betreiben.
audience: admin
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Installation und Administration

Dieser Bereich behandelt Windows Portable, Docker, Laufzeitkonfiguration, Benutzer und Rollen,
SMTP, Sicherung und Wiederherstellung, Updates, TLS, Uploads, OCR, Drucker,
Betriebsprüfungen und konservative Fehlerbehebung.

Die Administrationsanleitungen beschreiben die stabile Laufzeit v0.1.17.6 und erhalten das lokale,
selbst gehostete Sicherheitsmodell von RailKeeper.

## Lebenszyklus der Stammdaten

RailKeeper unterscheidet zwischen mitgelieferten Einträgen und eigenen Einträgen, die ein
Administrator oder Editor angelegt hat. Mitgelieferte Einträge lassen sich bearbeiten und
deaktivieren, aber nicht endgültig löschen. Ein ungenutzter eigener Eintrag kann endgültig gelöscht
werden. Sobald ein eigener Eintrag verwendet wird, lässt er sich nur noch deaktivieren.

Deaktivierte Einträge werden für neue Fahrzeuge, Zubehörartikel oder andere Datensätze nicht mehr
angeboten. Bestehende Datensätze behalten ihren gespeicherten Wert und zeigen ihn weiterhin als
inaktiv an. Eine Deaktivierung macht historische Bestandsdaten daher nicht ungültig.

Bearbeitungen und Deaktivierungen bleiben nach einem Neustart, einem Abgleich der mitgelieferten
Stammdaten, einem Update, dem Export und Import von Stammdaten sowie einer Sicherung und
Wiederherstellung der Anwendung erhalten. Vor umfangreichen Stammdatenänderungen wird dennoch eine
aktuelle RailKeeper-Sicherung empfohlen.
