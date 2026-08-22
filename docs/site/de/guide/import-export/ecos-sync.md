---
title: ECoS-Lokabgleich
description: Ausgewählte ESU-ECoS-Lokdaten lesen, prüfen, importieren und ausdrücklich schreiben.
audience: user
status: stable
reviewedVersion: 0.1.20
lastReviewed: 2026-08-16
---

# ECoS-Lokabgleich

Der ECoS-Arbeitsbereich liest Lokstammdaten, statische Funktionsbeschreibungen und bereitgestellte
CV-Werte aus einer aktiven ESU ECoS. Jede Lok bleibt in einer Arbeitsliste, bis ein Admin sie im
Fahrzeugeditor öffnet und speichert oder bewusst überspringt.

Alle ECoS-API-Routen sind ausschließlich für Admin verfügbar. Die Digitalzentrale muss vor der
Verwendung unter **Einstellungen > Digitalzentralen** eingerichtet und geprüft werden. Host und
Port sind im Import-Arbeitsbereich schreibgeschützt.

## Unterstützte und nicht unterstützte Digitalzentralen

| Aktiver Anbieter | Verhalten unter Import/Export in v0.1.20 |
| --- | --- |
| ESU ECoS | Leseablauf, Fahrzeugübergabe, CV- und statische Funktionsvorschläge sowie geprüfter Schreib-Sync sind verfügbar. |
| Z21, Intellibox 3 oder Märklin CS3 | RailKeeper zeigt, dass der Importpfad vorbereitet, aber noch nicht umgesetzt ist. |
| Kein aktiver Anbieter | Die Seite zeigt einen Verweis zu den Digitalzentralen-Einstellungen. |

RailKeeper überwacht in diesem Ablauf weder Geschwindigkeit noch Richtung, aktive
Funktionszustände, Lokbilder oder Anlagenobjekte. Laufzeitattribute wie Geschwindigkeit, Richtung
und Funktionsstatus werden aus der Prüfung ausgeschlossen. Nur statische Funktionsbeschreibungen
und CV-Werte, die das ECoS-Lokobjekt bereitstellt, kommen infrage.

## Lok-Arbeitsliste einlesen

Mit aktivierter ECoS und gespeichertem Host:

1. **Import/Export** öffnen.
2. Live-Monitor-Zustand sowie gespeicherten Host und Port prüfen.
3. **Daten holen** wählen.
4. Auf Lokanzahl und Arbeitsliste warten.
5. ECoS-Name, Objekt-ID, Adresse, erkannten RailKeeper-Treffer, Decoderhinweis, CV-Anzahl und
   aktuellen Arbeitslistenstatus prüfen.

RailKeeper fragt den ECoS-Live-Zustand alle 15 Sekunden ab. Ist der Live-Monitor bereits verbunden,
holt die Seite einmal automatisch Daten, sofern weder Prüfergebnis noch Importsitzung aktiv sind.

Ein erfolgreicher Abruf erzeugt eine Arbeitssitzung im Sitzungsspeicher des Browser-Tabs. Nach der
Rückkehr aus dem Fahrzeugeditor werden Liste und Zustände **offen**, **in Bearbeitung**,
**gespeichert** oder **übersprungen** wiederhergestellt. Das Schließen des Browser-Tabs beendet
diese browserseitige Speicherung.

## So wird ein bestehendes Fahrzeug vorgeschlagen

RailKeeper prüft Kandidaten in dieser Reihenfolge:

1. vorhandene externe ECoS-Zuordnung mit derselben ECoS-Objekt-ID;
2. gleiche Decoder-Adresse unter **Digital / Decoder-Nr.**;
3. normalisierter Vergleich mit Fahrzeugbezeichnung oder Fahrzeugnummer.

Der Namensvergleich ignoriert Groß- und Kleinschreibung, Umlautschreibweise, Satzzeichen sowie ein
führendes `BR` oder `V` vor einer Zahl. Bei längeren Namen wird auch ein Enthaltensein akzeptiert.
Diese Treffer sind Vorschläge, kein Identitätsnachweis. Inventarnummer und Fahrzeug müssen vor
Update oder ECoS-Schreibvorgang geprüft werden.

## CV- und Funktionsvorschläge prüfen

Arbeitsliste und CV-Vorschau zeigen ausschließlich Werte, welche die ECoS geliefert hat.
Standard-CV-Definitionen ergänzen nach Möglichkeit Bedeutung, Kategorie und Interpretation. CV8
kann den Decoder-Hersteller aus RailKeeper-Stammdaten erkennen. Eine fehlende oder unbekannte CV8
bleibt sichtbar ungeklärt. Fehlende CVs blockieren den Import nicht.

Statische ECoS-Funktionsbeschreibungen werden zu F-Tasten-Vorschlägen und nach Möglichkeit
RailKeeper-Symbolen zugeordnet. Der aktive Ein-/Aus-Zustand wird nie importiert. Funktions- und
CV-Vorschau müssen vor dem Speichern im Fahrzeugeditor geprüft werden.

## Ein Fahrzeug anlegen oder aktualisieren

Bei einer nicht zugeordneten Lok **Fahrzeug erfassen**, bei einem Vorschlag **Treffer
aktualisieren** wählen. Die allgemeine CSV/XML/JSON-Prüfung wird nicht verwendet. RailKeeper öffnet
direkt den Fahrzeugeditor und übergibt:

- ECoS-Objekt-ID, Name, Adresse, Protokoll und Profil als Quellkontext;
- Name, Kategorie, Digitalzustand, Decoder-Adresse, Decoder-Typ und Beschreibung als
  Fahrzeugfeld-Vorschläge;
- eine externe ECoS-Zuordnung;
- statische Funktionsvorschläge;
- CV-Wert-Vorschläge.

Bei einem neuen Fahrzeug lautet die voreingestellte Kategorie `Lokomotive`. Hersteller,
Bezeichnung, Spurweite, Kategorie und Gattung bleiben Pflichtfelder. Nicht von der ECoS gelieferte
Felder werden hervorgehoben und müssen vor dem Speichern geklärt werden.

Das Speichern führt mehrere Schreibvorgänge nacheinander aus: zuerst Fahrzeuggrunddaten, dann
ECoS-Zuordnung, CV-Werte und konfigurierte Funktionen. Erst nach dem vollständigen Erfolg wird die
Arbeitslistenzeile **gespeichert** und RailKeeper kehrt zur ECoS-Sitzung zurück. Scheitert ein
späterer Schritt, kann das Grundfahrzeug bereits angelegt oder aktualisiert sein. Vor einem erneuten
Versuch muss es geprüft werden, um Duplikate oder wiederholte Werte zu vermeiden.

**Überspringen** verwenden, wenn eine Lok unverändert bleiben soll. **Nächste offene Lok** wählt
den nächsten offenen Eintrag, danach einen noch in Bearbeitung markierten Eintrag, wenn kein offener
Eintrag mehr vorhanden ist.

## Geprüfte RailKeeper-Daten in die ECoS schreiben

Der Schreib-Sync ist nur bei einer Lok mit erkanntem RailKeeper-Treffer verfügbar:

1. Prüfen, ob das vorgeschlagene Fahrzeug stimmt.
2. **Sync prüfen** wählen. Diese Vorschau schreibt nichts in die Digitalzentrale.
3. Bis zu drei angezeigte Unterschiede zwischen aktuellem ECoS-Wert und gewünschtem
   RailKeeper-Wert prüfen.
4. Sind Änderungen vorhanden, **In ECoS schreiben** wählen.
5. Die Browser-Rückfrage für diese Lok bestätigen.

Der stabile Schreibumfang ist bewusst eng: nur Name, Adresse und Protokoll. Leere Sollwerte werden
nicht geschrieben. RailKeeper sendet die kombinierte ECoS-Änderung erst nach ausdrücklicher
Bestätigung, lädt danach das RailKeeper-Fahrzeug neu und markiert dessen externe Zuordnung nach
Möglichkeit als synchronisiert.

CVs, Funktionsdefinitionen, Geschwindigkeit, Richtung, aktive Funktionen, Bilder und Anlagenobjekte
werden nicht in die Digitalzentrale geschrieben.

## Fehlerbehebung

| Symptom | Prüfen |
| --- | --- |
| ECoS-Arbeitsbereich fehlt oder ist deaktiviert | Prüfen, ob ein Admin ECoS aktiviert und unter Digitalzentralen einen nicht leeren Host gespeichert hat. |
| Ein anderer Anbieter ist aktiv | v0.1.20 setzt diesen Importpfad nur für ECoS um. |
| Datenabruf scheitert | Netzwerkerreichbarkeit, gespeicherten Host und Port, ECoS-Verfügbarkeit und Admin-Rolle prüfen. |
| Ein falsches Fahrzeug wird vorgeschlagen | Nicht aktualisieren oder synchronisieren. Zuerst externe Zuordnung oder Decoder-/Namensdaten im Fahrzeugablauf korrigieren. |
| Eine CV oder Funktion fehlt | Das ECoS-Lokobjekt hat sie nicht bereitgestellt oder RailKeeper hat sie nicht als statischen unterstützten Wert erkannt. Fehlende CVs blockieren das Speichern nicht. |
| Arbeitsliste kehrt zurück, aber Zeile ist nicht gespeichert | Die vollständige Folge aus Fahrzeug, Zuordnung, CVs und Funktionen wurde nicht beendet. Fahrzeug und Fehler vor erneutem Versuch prüfen. |
| **In ECoS schreiben** ist nicht sichtbar | Zuerst **Sync prüfen** ausführen. Die Schreibschaltfläche erscheint nur bei einem Treffer mit nicht leerem, noch nicht angewendetem Änderungsplan. |

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert RailKeeper **v0.1.20** und wurde zuletzt am 16.08.2026 geprüft.
