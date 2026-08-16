---
title: Benutzerhandbuch
description: Alle stabilen Arbeitsabläufe in RailKeeper kennenlernen.
audience: user
status: stable
reviewedVersion: 0.1.18
lastReviewed: 2026-08-16
---

# Benutzerhandbuch

Dieses Handbuch erklärt den vollständigen veröffentlichten RailKeeper-Arbeitsablauf für Sammler,
Vereine und kleine Werkstätten. Es umfasst Ersteinrichtung, Navigation, Fahrzeuge, Zubehör,
Decoderdaten, Wartung, Dokumente, Berichte, Ausstellungen sowie kontrollierten Import und Export.

Die Inhalte dieses Bereichs beschreiben RailKeeper v0.1.18. Der Anlagen-Arbeitsbereich befindet
sich noch in Entwicklung und wird deshalb nicht als stabile Benutzerfunktion dargestellt.

## Hier beginnen

[Ersteinrichtung und Anmeldung](/de/guide/getting-started/) erklärt, wie der erste Administrator
angelegt wird, die An- und Abmeldung funktioniert, eine Anmeldung mit Zwei-Faktor-Code abgeschlossen
und ein vergessenes Passwort wiederhergestellt wird.

Fahre mit [Übersicht, Kennzahlen und Datenqualität](/de/guide/overview/) fort, um das Dashboard zu
verstehen, Bestandslücken als gefilterte Fahrzeuglisten zu öffnen und die Kacheln anzuordnen.

[Fahrzeugbestand und Grunddaten](/de/guide/vehicles/) erklärt Suche und Filter, das Anlegen und
Bearbeiten von Grunddaten, Reports und QR-Etiketten sowie das sichere Löschen.

[Fahrzeugbilder und Beilagen](/de/guide/vehicles/media) behandelt lokale Bilder, Haupt- und
Alternativbilder, allgemeine Beilagen, Anzeigen und Herunterladen sowie das sichere Entfernen.

[Fahrzeugwartung und Zustand](/de/guide/vehicles/maintenance) erklärt Wartungseinträge,
Fälligkeiten, Abschluss, Kosten, verknüpfte Medien und sicheres Löschen.

[Decoder, Funktionen und CV-Daten](/de/guide/vehicles/decoder-cv) behandelt Zuordnungen F0-F31,
die schreibgeschützte Fahrkurve, CV-Werte und Austausch, Decoder-Dateivorschauen und sicheres
Speichern.

[Artikelsuche, Web-Dokumente und Ersatzteile](/de/guide/vehicles/search-and-spares) erklärt, wie du
externe Vorschläge prüfst, ausgewählte Artikeldaten und Bilder speicherst, Dokumente importierst und
Ersatzteile pflegst, ohne teilweise Schreibvorgänge zu übersehen.

[Zubehör](/de/guide/accessories/) beschreibt den getrennten Zubehör-Artikelbestand, technische
Produktdaten, Mengen- und Einzelbestand, Käufe, Dokumente, Reservierungen, Einbauten und
Verwendungshistorie.

[Messearbeitsbereich](/de/guide/exhibition/) erklärt, wie du Listen vorbereitest, Betriebseinträge
pflegst, Sperren verwendest, DCC- und SX-Adresskonflikte löst und den Veranstaltungsreport druckst.

## Windows Standalone aktualisieren

Unter **Einstellungen > Allgemein > Updates** nach einer neuen Version suchen. Erkennt RailKeeper
das passende vertrauenswürdige Windows-ZIP, enthält die Download-Schaltfläche dessen Version. Die
Schaltfläche startet ausschließlich den Browser-Download. RailKeeper installiert oder ersetzt keine
Dateien.

Eine Anwendungssicherung erstellen, RailKeeper beenden und das heruntergeladene ZIP in einen neuen
Programmordner entpacken. Diese Kopie starten und den aktiven Datenpfad sowie den Bestand prüfen,
bevor der vorherige Programmordner entfernt wird. Der sichere Standardpfad
`%LOCALAPPDATA%\RailKeeper\data` bleibt getrennt und unverändert. Fehlt die passende
Download-Schaltfläche, die verlinkte GitHub-Release-Seite verwenden. Docker-Installationen werden
weiterhin mit `docker compose pull` und anschließend `docker compose up -d` aktualisiert.
