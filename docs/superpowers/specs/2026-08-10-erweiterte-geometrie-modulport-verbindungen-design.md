# Erweiterte Geometrie, Paket B: Modulport-Verbindungen

**Datum:** 2026-08-10

**Status:** Aus der abgestimmten Etappe 4 abgeleitet, lokale Umsetzung ohne Veröffentlichung

## Ziel

Paket B verbindet Modulports innerhalb einer Aufbaukonfiguration. RailKeeper leitet aus den lokalen
Portkoordinaten und den gespeicherten Positionen der Einheiten globale Portpositionen ab, erkennt
exakte Verbindungen und zeigt offene oder inkompatible Übergänge. Planner und Admin können eine
Einheit gezielt an den nächsten kompatiblen Port ausrichten.

## Berechnung

Die Berechnung bleibt rein abgeleitet. Es entsteht keine Verbindungstabelle. Ein Port wird durch die
Position und Rotation seiner Einheit in das globale Koordinatensystem der Aufbaukonfiguration
transformiert. Zwei Ports bilden eine Verbindung, wenn:

- sie verschiedenen Einheiten angehören,
- Art und normalisierte Schnittstellenkennung identisch sind,
- ihr Abstand höchstens 0,25 mm beträgt,
- ihre Richtungen mit höchstens 0,5 Grad Abweichung gegeneinander zeigen,
- keiner der Ports bereits einer anderen Verbindung zugeordnet ist.

Offene aktive Ports erzeugen einen Hinweis. Ports, die höchstens 25 mm auseinanderliegen und
geometrisch innerhalb von 10 Grad gegeneinander zeigen, aber fachlich nicht kompatibel sind, erzeugen
einen Inkompatibilitätshinweis. Archivierte Ports werden weder verbunden noch bewertet.

## Kontrolliertes Einrasten

Die Einrastvorschau sucht für eine ausgewählte Einheit den nächsten kompatiblen Zielport einer
anderen Einheit. Innerhalb von 25 mm und 10 Grad berechnet sie die exakte Zielposition und Rotation.
Bei gleichem Abstand entscheidet die stabile Reihenfolge aus Ziel-Einheit und Ziel-Port.

Die Vorschau verändert keine Daten. Die Oberfläche übernimmt die vorgeschlagene Pose in das
Aufbauformular und kennzeichnet dies als noch nicht gespeichert. Erst der vorhandene Befehl
`Aufbau speichern` schreibt die Konfiguration mit optimistischer Versionierung. Eine automatische
Modulreihenfolge oder unaufgeforderte Bewegung findet nicht statt.

## Anwendung und API

Ein fokussiertes Repository lädt eine Aufbaukonfiguration mit ihren Einheiten und aktiven Ports. Die
fachliche Transformation, Analyse und Einrastberechnung liegt als reine Funktion in der Domäne.

- `GET /api/v1/layout-configurations/{id}/port-analysis` liefert Verbindungen und Hinweise.
- `POST /api/v1/layout-configurations/{id}/unit-snap-preview` liefert eine unverbindliche Pose.
- Lesen bleibt für Viewer, Editor, Planner und Admin erlaubt; Messe bleibt ausgeschlossen.
- Die Vorschau ist für Planner und Admin freigegeben und bleibt CSRF-geschützt.

Da beide Ergebnisse abgeleitet sind und keine neue Tabelle entsteht, bleibt Backup-Format 7
unverändert.

## Oberfläche

Der Bereich `Aufbauten` zeigt für die gewählte Konfiguration eine kompakte Anschlussprüfung mit
Summen für Verbindungen, offene Ports und inkompatible Übergänge. Darunter stehen nachvollziehbare
Einheiten- und Portnamen. Im Bearbeitungsformular erhält jede zugeordnete Einheit mit aktiven Ports
die Aktion `An Ports ausrichten`. Erfolg übernimmt X, Y und Rotation in das Formular; ein fehlender
Treffer lässt die Werte unverändert und zeigt einen neutralen Hinweis.

Alle Auswahlfelder und Aktionen verwenden vorhandene RailKeeper-Komponenten und Tokens. Deutsch und
Englisch werden gemeinsam gepflegt. Viewer sehen die Prüfung, aber keine Ausrichtungsaktion.

## Abgrenzung

Nicht Teil dieses Pakets sind Flexgleise, Höhenprofile, Kollisionsprüfung von Modulkonturen,
persistierte Kabelverbindungen, automatische Modulreihenfolge und digitale Steuerbefehle.

## Abnahme

- Transformationen, Toleranzgrenzen, stabile Auswahl und Inkompatibilitäten sind durch Domänentests
  abgedeckt.
- API-Rollen und CSRF-Verhalten sind geprüft; OpenAPI entspricht den Laufzeittypen.
- Die Oberfläche zeigt Analysezustände und übernimmt eine Einrastvorschau ohne vorzeitiges Speichern.
- Backend-Gesamttests, Frontend-Gesamttests, Produktionsbuild und lokale Browserabnahme sind grün.
