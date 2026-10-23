---
title: Darstellung
description: Designmodus wählen und helle sowie dunkle Varianten getrennt konfigurieren.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Darstellung

Öffne **Einstellungen > Darstellung**. Wähle **System**, **Hell** oder **Dunkel** und prüfe die Lesbarkeit in der gewünschten Variante. Du kannst die Varianten für hell und dunkel getrennt einstellen.

<GuideFigure name="appearance" alt="Heller und dunkler Modus lassen sich getrennt anpassen." caption="Heller und dunkler Modus lassen sich getrennt anpassen. Künstliche Demodaten, keine Werkseinstellungen. Anklicken für die volle Größe." />

Unter **Einstellungen > Darstellung** werden Farbmodus sowie helle und dunkle Variante von
RailKeeper gesteuert. Jede Änderung wirkt sofort. Eine getrennte Speichern-Schaltfläche gibt es
nicht.

## Designmodus

| Modus | Verhalten |
| --- | --- |
| **System** | Folgt der Hell-/Dunkel-Vorgabe des Betriebssystems oder Browsers. |
| **Hell** | Verwendet dauerhaft die konfigurierte helle Variante. |
| **Dunkel** | Verwendet dauerhaft die konfigurierte dunkle Variante. |

Die Design-Schaltfläche am unteren Ende der Seitenleiste stellt dieselbe Moduswahl als Schnellzugriff
bereit. Jeder Klick wechselt von Dunkel zu Hell, von Hell zu System und von System zu Dunkel.

Ein reines Messe-Konto kann diesen Schnellzugriff verwenden, obwohl es die Einstellungen nicht
öffnen darf. Die Auswahl wirkt in diesem Browser, die Messe-Rolle darf sie aber nicht in das
geschützte Profil schreiben.

## Helle und dunkle Varianten konfigurieren

RailKeeper speichert getrennte Einstellungen für beide Farbmodi. Eine Änderung der hellen Variante
verändert die dunkle Variante nicht und umgekehrt.

| Option | Helle Variante | Dunkle Variante | Wirkung |
| --- | --- | --- | --- |
| Hintergrund | Neutral, Warm, Kühl | Neutral, Warm, Kühl, OLED Schwarz | Ändert Grund-, Panel- und Eingabeflächen. |
| Akzent | Grün, Blau, Gold | Grün, Blau, Gold | Ändert interaktive und hervorgehobene Farben. |
| Stil | Klassisch, Kompakt, Kontrast | Klassisch, Kompakt, Kontrast | Wählt Standarddarstellung, engere Abstände zwischen Einstellungskarten oder stärkere Trennlinien und Sekundärtexte. |

Bei ausgewähltem **System** verwendet RailKeeper automatisch die konfigurierte helle Variante, wenn
das System helle Farben anfordert, und die konfigurierte dunkle Variante bei dunkler Vorgabe. Beide
Varianten bleiben deshalb sinnvoll, auch wenn weder Hell noch Dunkel erzwungen wird.

**OLED Schwarz** steht nur für den dunklen Hintergrund bereit und verwendet eine schwarze
Grundfläche sowie sehr dunkle Panels. **Kompakt** verkleinert derzeit hauptsächlich die Abstände
zwischen Einstellungskarten; es ist keine globale Tabellendichte. **Kontrast** verstärkt Trennlinien
und zurückgenommene Texte, ersetzt aber nicht Browser-Zoom oder Bedienungshilfen des Betriebssystems.

## Speicherung und Benutzerbezug

Der gewählte Modus und alle sechs Variantenoptionen werden im Browser sowie im Profil des aktuellen
Benutzers gespeichert. RailKeeper stellt den Modus bereits in der Anwendungshülle wieder her. Beim
Öffnen der Einstellungen werden zusätzlich die detaillierten hellen und dunklen Optionen geladen
und angewendet.

Der Profilzugriff im Hintergrund zeigt keine Erfolgsmeldung. Ist der Server vorübergehend nicht
erreichbar, kann der aktuelle Browser die lokale Auswahl weiter anzeigen, während ein anderer
Browser noch den älteren Profilwert besitzt. Nach wiederhergestellter Verbindung Einstellungen neu
laden und die Auswahl bei Bedarf erneut setzen.

Anwendungssicherung und Wiederherstellung schließen Benutzer-Profileinstellungen absichtlich aus.
Darstellungsvorgaben gehören daher nicht zur Anwendungssicherung.

## Fehlerbehebung

| Symptom | Prüfen |
| --- | --- |
| Systemmodus wechselt unerwartet | Farbmodus des Betriebssystems oder Browsers prüfen. Der Systemmodus folgt diesem Signal. |
| Hintergrund oder Akzent scheinen wirkungslos | Prüfen, ob die Anwendung gerade die bearbeitete helle oder dunkle Variante verwendet. |
| Ein anderer Browser zeigt ein früheres Design | Einstellungen mit demselben Benutzer öffnen und neu laden. Nach einem Verbindungsfehler den Modus oder die Variante bei Bedarf erneut wählen. |
| Kompakt verdichtet die Bestandstabellen nicht | Derzeit verkleinert der Stil hauptsächlich Abstände zwischen Einstellungskarten. |
| Kontrast reicht weiterhin nicht aus | Kontrast-Stil mit Browser-Zoom und Bedienungshilfen des Betriebssystems kombinieren. |
