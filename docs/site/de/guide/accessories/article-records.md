---
title: Artikelstammdaten und Fachangaben
description: Zubehörartikel anlegen, ihre Identität pflegen und technische Angaben prüfen.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Artikelstammdaten und Fachangaben

Ein Zubehörartikel ist der gemeinsame Produktdatensatz für Bestand, Einkäufe, Dokumente,
Reservierungen und Einbauten. Legen Sie ihn einmal für einen Hersteller- und Katalogartikel an und
erfassen Sie anschließend Mengen oder Einzelstücke dazu. Dieses Kapitel beschreibt die
Artikelidentität und die artabhängigen Fachangaben in RailKeeper v0.1.17.6.

## Zugriffsrechte und Speichern

Administratoren und Bearbeiter können Artikel anlegen und bearbeiten. Betrachter und Planer können
sie ansehen, die Artikelfelder bleiben jedoch schreibgeschützt. Ein Planer kann im ansonsten
schreibgeschützten Dialog Reservierungen anlegen oder stornieren. Daraus folgt kein Recht, den
Artikel selbst zu ändern. Die Rolle Messe hat keinen allgemeinen Zugriff auf Zubehör.

Der Dialog trennt **Artikel**, **Bestand**, **Einkäufe & Dokumente**, den artikelspezifischen Reiter
und bei vorhandenen Daten den **Nutzungsverlauf**. Die zentrale Aktion **Speichern** schreibt das
Artikelformular einschließlich Artikelart, Fachangaben, Bestandsstrategie und Mindestbestand.
Bestandsbewegungen, Einkäufe, Dokumente, Reservierungen und Einbauten besitzen eigene Aktionen und
können bereits gespeichert sein, während andere Artikelfelder noch ungespeichert sind. Schließen
Sie einen Entwurf ab oder verwerfen Sie ihn bewusst, bevor Sie eine solche Sofortaktion beginnen.

Beim Schließen mit geänderten Artikelfeldern, vorgemerkten Suchbildern oder einem unfertigen
Unterformular erscheint eine Verwerfbestätigung. Im Anlegemodus lassen sich Bestand und zugehörige
Ressourcen erst speichern, nachdem der Artikelstamm einmal erfolgreich gespeichert wurde.

## Artikel anlegen

1. Öffnen Sie **Zubehör** und wählen Sie **Artikel anlegen**.
2. Wählen Sie unter **Artikel** Hersteller, Artikelart und Unterart und geben Sie den Namen ein.
3. Ergänzen Sie Katalogidentität, Spurweiten, Maßstab, Verpackungseinheit, Bestandseinheit und
   optionale Referenzdaten.
4. Füllen Sie bekannte Fachangaben im Reiter der gewählten Artikelart aus.
5. Wählen Sie unter **Bestand** Bestandsstrategie und Mindestbestand. Die Folgen beschreibt
   [Bestand, Einkäufe und Dokumente](./stock-purchases-documents) im Detail.
6. Wählen Sie **Speichern** und bearbeiten Sie markierte Reiter oder die Dublettenwarnung.

Pflichtfelder sind Hersteller, Name, Unterart, Bestandseinheit und eine positive ganzzahlige
Verpackungseinheit. Der Mindestbestand muss eine ganze Zahl ab null sein. Technische Zahlen müssen
den angezeigten Bereich und die Schrittweite einhalten. Als Dezimaltrennzeichen wird auch ein Komma
akzeptiert.

RailKeeper vergibt die Artikel-Inventarnummer erst in der erfolgreichen Anlegetransaktion. Dafür
wird das aktive Inventarnummernschema **Artikel** verwendet, dessen Standardformat Werte wie
`RK-ART-000001` erzeugt. Ein fehlgeschlagener Anlageversuch verbraucht die vorgemerkte Nummer nicht.
Das Feld lässt sich im Artikeldialog weder eingeben noch ändern. Fehlt ein aktives Schema, bleibt
das Anlegen gesperrt, bis ein Administrator die Inventarnummernkonfiguration repariert.

Bei einem neuen Datensatz wählt RailKeeper die erste aktive konfigurierte Artikelart. Ist keine
aktive Artikelart verfügbar, bleibt das Anlegen gesperrt. Anfangswerte sind Verpackungseinheit 1,
Bestandseinheit **Stück**, Mindestbestand 0, Bestandsstrategie **Mengenbestand** und
Herstellerstatus **Unbekannt**.

## Referenz der allgemeinen Felder

| Feld | Stabiles Verhalten |
| --- | --- |
| Inventarnummer | Automatisch aus dem Schema **Artikel** vergeben und schreibgeschützt. |
| Hersteller | Erforderliche aktive Stammdatenauswahl. Ein gespeicherter inaktiver oder historischer Wert bleibt sichtbar, kann aber nicht neu gewählt werden. |
| Artikelnummer | Optionale Hersteller- oder Katalognummer. Wenn ausgefüllt, aktiviert sie zusammen mit dem Hersteller die Dublettenprüfung. |
| Name | Erforderlicher Artikelname. |
| EAN / GTIN | Optionaler Barcodewert und eigenständiges Kriterium der Artikelsuche. |
| Herstellerstatus | Angekündigt, Verfügbar, Ausgelaufen oder Unbekannt. |
| Artikelart | Gleis, Signal, Decoder, Elektrik / Steuerung, Gebäude / Ausstattung, Landschaftsverbrauchsmaterial, Beleuchtung oder Sonstiges, abhängig von der aktiven Konfiguration. |
| Unterart | Erforderliche konfigurierte Unterart der gewählten Artikelart. Eine unveränderte historische Unterart bleibt lesbar und kann beibehalten werden. |
| Spurweiten | Keine, eine oder mehrere aktive Spurweiten aus den Stammdaten. Vorhandene inaktive Werte bleiben sichtbar. |
| Maßstab | Optionaler Text. Solange automatisch verwaltet, folgt er der ersten gewählten aktiven Spurweite mit hinterlegtem Maßstab. |
| Verpackungseinheit | Erforderliche positive ganze Zahl, wie viele Bestandseinheiten eine Packung enthält. |
| Bestandseinheit | Erforderlicher aktiver Stammdatenwert, zum Beispiel Stück. Vorhandene inaktive Werte bleiben sichtbar. |
| Beschreibung | Optionaler Freitext. |
| Hersteller-URL | Optionale Herstellerreferenz. Die stabile Zubehör-Trefferauswahl füllt sie nicht. |
| Produkt-URL | Optionale Produktreferenz. Suchtreffer können sie nur aus einer HTTP- oder HTTPS-Quelle übernehmen. |
| Alternative Artikelnummern | Optionale, durch Komma oder Zeilenumbruch getrennte Werte. Leere Einträge werden entfernt. |
| Schlagwörter | Optionale, durch Komma oder Zeilenumbruch getrennte Begriffe. Vorschläge kombinieren Name, Hersteller, Artikelart und Unterart ohne Dubletten unabhängig von Groß- und Kleinschreibung. |
| Kompatibilität | Optionaler Freitext für unterstützte Systeme oder Produkte. |
| Interne Notizen | Optionale lokale Notizen. |

Bei normalen Textfeldern entfernt RailKeeper führende und nachgestellte Leerzeichen. Eine manuelle
Änderung an **Maßstab** oder **Schlagwörter** beendet den jeweiligen Vorschlagsautomatismus für die
aktuelle Bearbeitung. Die Automatik arbeitet beim Anlegen und Bearbeiten, nie in der Leseansicht.

Hersteller, Artikelarten, Unterarten, Spurweiten, Bestandseinheiten und eigene Felder stammen aus
den Stammdaten. RailKeeper erhält einen unveränderten inaktiven Wert eines bestehenden Artikels,
erlaubt aber weder sein neues Hinzufügen noch seine Änderung. So bleiben historische Datensätze
bearbeitbar, ohne ihre Klassifikation stillschweigend zu ersetzen.

## Artikelart und Fachangaben wählen

Der letzte Hauptreiter trägt den Namen der gewählten Artikelart. Seine Felder sind optional, sofern
die lokale Stammdatenkonfiguration nichts anderes verlangt. Jeder eingetragene Wert wird jedoch
geprüft. Ganzzahlige Felder verwenden die Schrittweite 1. Andere nicht negative Zahlen haben keine
feste Schrittweite, sofern nachfolgend nichts anderes angegeben ist.

| Artikelart | Stabile Fachfelder |
| --- | --- |
| Gleis | Gleissystem; Länge und Radius in mm; Winkel von 0 bis 360° in 0,1°-Schritten; Richtung Links, Rechts oder Symmetrisch; Herzstückwinkel von 0 bis 180° in 0,1°-Schritten; Schwellenart; Profilhöhe in mm; Bettung; nicht negative ganze Anzahl Anschlüsse; Digitaltauglich. |
| Signal | Vorbild; Epochen I bis VI; Signalbegriffe Halt, Fahrt, Langsamfahrt und Rangierfahrt; nicht negative ganze Anzahl LEDs; Bauhöhe in mm; nicht negative Betriebsspannung AC und DC; Montageart Aufbau, Unterflur, Mast oder Wand; Antrieb Manuell, Motor, Servo oder Magnetantrieb; Integrierter Decoder; Steuermodul. |
| Decoder | Schnittstelle Kabel, NEM 651, NEM 652, PluX16, PluX22, 21MTC oder Next18; Protokolle DCC, Motorola, Selectrix, mfx und RailCom; nicht negative ganze Funktions- und Servoausgänge; nicht negativer Motor-, Ausgangs- und Summenstrom in mA; RailCom; SUSI; Abmessungen; Firmware. |
| Elektrik / Steuerung | Nicht negative Eingangs- und Ausgangsspannung, Strom in A und Leistung in W; nicht negative ganze Anzahl Kanäle; Protokolle DCC, Motorola, Selectrix, LocoNet, CAN und S88; Anschlussarten Schraubklemme, Steckverbinder, RJ45, Busanschluss und Kabel; Schutzfunktionen Kurzschluss, Überlast, Übertemperatur und Verpolung; kompatible Artikelarten Gleis, Signal, Decoder und Beleuchtung. |
| Gebäude / Ausstattung | Epochen I bis VI; Abmessungen; Grundfläche; Material; Bauform Bausatz, Fertigmodell oder Teilbausatz; nicht negative ganze Teilezahl; Schwierigkeitsgrad Einfach, Mittel oder Anspruchsvoll; Beleuchtungsmöglichkeiten Innenbeleuchtung, Bahnsteigbeleuchtung, Straßenbeleuchtung und Effektbeleuchtung; Grundriss vorhanden. |
| Landschaftsverbrauchsmaterial | Material; Farbe; Jahreszeit; nicht negativer Inhalt; Inhaltseinheit Stück, Packung, Meter, Gramm oder Milliliter; Faser- oder Korngröße; Reichweite; geeignete Maßstäbe Z, N, TT, H0, 0, 1 und G; Sicherheitshinweise. |
| Beleuchtung | Lichtfarbe; nicht negative Farbtemperatur in K, Spannung in V und Strom in mA; Stromart Wechselstrom, Gleichstrom oder Wechsel- oder Gleichstrom; nicht negative ganze Anzahl LEDs; Dimmbar; Abmessungen; Montageart Aufbau, Unterflur, Mast oder Wand. |
| Sonstiges | Aktive konfigurierte eigene Felder. Unterstützt werden Text, Zahl, Ja/Nein, Datum, Einfachauswahl und Mehrfachauswahl. Auswahlfelder erscheinen nur mit konfigurierten Auswahlwerten. |

Ja/Nein-Felder unterscheiden ein ausdrücklich gewähltes Ja oder Nein von einem nicht gespeicherten
Wert. Mehrfachauswahlen lehnen unbekannte oder doppelte Auswahlwerte ab. Negative Zahlen, Werte
außerhalb eines genannten Bereichs und Werte entgegen einer genannten Schrittweite verhindern das
Speichern und markieren den artikelspezifischen Reiter.

Ein Wechsel der Artikelart leert die Unterart. Würde der Wechsel die Unterart, einen Fachwert oder
eine unfertige Zahl verwerfen, fragt RailKeeper nach einer Bestätigung. Kompatible Felder mit
gleichem Schlüssel und Datentyp bleiben erhalten. Unvereinbare Felder werden erst nach Bestätigung
entfernt. Brechen Sie die Bestätigung ab, um die aktuelle Artikelart und ihre Werte zu behalten.

Für **Sonstiges** werden die Felder durch aktive `accessory_custom_field`-Stammdaten definiert. Ein
historischer eigener Wert, dessen Definition inaktiv geworden oder verschwunden ist, bleibt nur
unverändert erhalten. Er kann weder neu hinzugefügt noch bearbeitet werden.

## Barcode und Artikeldatensuche verwenden

Der Reiter **Artikel** verwendet denselben geprüften Trefferablauf wie die Fahrzeugsuche, jedoch
mit zubehörspezifischen Eingaben und Zielfeldern. Die normale Suche ist verfügbar, wenn entweder:

- eine EAN vorhanden ist oder
- Hersteller und erste gewählte Spurweite zusammen mit entweder Artikelnummer oder Name vorhanden
  sind.

RailKeeper übermittelt Hersteller, Artikelnummer, Name, erste Spurweite, aktuelle allgemeine Felder
und die aktuellen Fachangaben an die konfigurierten Suchquellen. Ist die Artikelsuche in den
Einstellungen deaktiviert, endet die Aktion ohne Änderung am Artikel.

Der Trefferdialog kann Hersteller, Artikelnummer, Name, EAN, Maßstab, Beschreibung, Produkt-URL und
Spurweite übernehmen. Hersteller oder Spurweite sind nur auswählbar, wenn sie einem aktiven
bekannten Stammdatenwert entsprechen. Eine gewählte Spurweite wird den vorhandenen hinzugefügt.
Ungültige allgemeine Werte und Produkt-URLs außerhalb von HTTP oder HTTPS werden ignoriert. In
v0.1.17.6 bietet die Trefferauswahl technische Ergebnisfelder nur für **Gleis** an. Andere
Artikelarten können die allgemeinen Treffergruppen, aber keine Fachangaben übernehmen.

Nur Felder mit aktuell leerem Formularwert sind in der Prüfung vorausgewählt. Vorhandene Werte sind
nicht zum Ersetzen markiert. Bilder werden nie automatisch gewählt. Wählen Sie nur vertrauenswürdige
Werte und übernehmen Sie diese in den lokalen Entwurf. Gespeichert wird erst nach erfolgreichem
Speichern des Artikels.

**Barcode** öffnet den Scanner mit der aktuellen EAN. Ein nicht leerer gescannter oder manuell
eingegebener Wert ersetzt die lokale EAN und startet sofort eine reine EAN-Suche. Kamerafreigabe und
manuelle Eingabe beschreibt [Artikelsuche, Web-Dokumente und Ersatzteile](../vehicles/search-and-spares).

Gewählte Suchbilder werden vorgemerkt. Nach dem Speichern des Artikelstamms importiert RailKeeper
sie nacheinander. Das erste erfolgreiche Bild wird nur dann zum Hauptbild, wenn der Zubehörartikel
noch kein Hauptbild besitzt. Ein fehlerhaftes Bild setzt weder den Artikel noch frühere Bilder
zurück. Spätere vorgemerkte Bilder werden trotzdem versucht. Der Dialog bleibt mit dem ersten
Importfehler geöffnet und behält fehlgeschlagene Bilder als Vormerkung. Schließen Sie ihn erst,
nachdem Sie über Wiederholen oder Verwerfen entschieden haben.

## Dublettenkandidaten prüfen

Ist die Artikelnummer nicht leer, prüft RailKeeper vor dem Speichern auf eine exakte Kombination
aus Hersteller und Artikelnummer. Führende und nachgestellte Leerzeichen sowie Groß- und
Kleinschreibung werden dabei ignoriert. Beim Bearbeiten ist der aktuelle Artikel ausgeschlossen.
Name, EAN, Spurweite und
Unterart ändern den Treffer nicht.

Bei Kandidaten zeigt RailKeeper diese vor dem Schreiben an. Mit **Abbrechen** kehren Sie zum Entwurf
zurück. **Trotzdem speichern** legt die festgehaltenen Werte an oder aktualisiert sie. Die Warnung
erlaubt beabsichtigte Varianten. Sie führt keine Datensätze zusammen und überträgt weder Bestand,
Dokumente noch Nutzungen eines anderen Artikels.

## Bearbeiten ohne Datenverlust

Öffnen Sie **Artikel bearbeiten**, ändern Sie den Entwurf und verwenden Sie die zentrale Aktion
**Speichern**. Die Validierung wechselt zum ersten betroffenen Reiter in dieser Reihenfolge:
Artikel, Bestand, Einkäufe & Dokumente, danach der artikelspezifische Reiter. Rote Markierungen an
Reitern kennzeichnen noch ungültige Werte.

Eine geänderte Bestandsstrategie ist nur erlaubt, wenn alle vorhandenen Bestands- und
Zuordnungsdaten darstellbar bleiben. RailKeeper blockiert insbesondere den Wechsel von einer
mengenfähigen Strategie zu reiner Einzelverwaltung, solange Mengenbestand, Mengenreservierungen
oder Mengeneinbauten bestehen. Ebenso wird das Entfernen der Einzelverwaltung blockiert, solange
Einzelstücke oder deren Reservierungen, Einbauten oder Zustandsverlauf vorhanden sind. Bereinigen
oder beenden Sie diese Abhängigkeiten bewusst und versuchen Sie es erneut. Eine abgelehnte
Aktualisierung lässt den gespeicherten Artikel unverändert.

Können erforderliche Stammdaten oder Definitionen eigener Felder nicht geladen werden, markiert der
Editor seine Ressourcen als veraltet und sperrt betroffene Schreibaktionen. Verwenden Sie
**Erneut versuchen**, ersetzen Sie angezeigte historische Werte nicht aus dem Gedächtnis und
speichern Sie erst nach erfolgreichem Laden. Ist bereits das Artikeldetail fehlgeschlagen, schließen
und öffnen Sie es neu, statt einen unvollständigen Ersatzstand zu bearbeiten.

## Artikel archivieren, wiederherstellen oder löschen

Administratoren und Bearbeiter können einen Artikel in der Übersicht archivieren oder
wiederherstellen. Beide Aktionen schreiben sofort und ohne Bestätigung. Archivierte Artikel
verschwinden aus der normalen Liste, bleiben über den Statusfilter **Archiviert** erreichbar und
können wiederhergestellt werden.

Nur ein Administrator kann einen Artikel endgültig löschen. RailKeeper verlangt dafür eine
Bestätigung. Ein Bestand ungleich null, Einzelstücke, Bestandsbewegungen, Einkäufe, Reservierungen,
Einbauten oder technische Anlagenpositionen mit Artikelverweis blockieren das Löschen.
Artikeldokumente allein blockieren es nicht. Ihre Metadaten werden entfernt, anschließend versucht
RailKeeper nicht mehr referenzierte gespeicherte Dateien zu entfernen. Das Löschen lässt sich nicht
rückgängig machen. Prüfen Sie vorher eine aktuelle Sicherung.

## Validierungs- und Ressourcenfehler beheben

| Situation | Nächster Schritt |
| --- | --- |
| Pflichtfeld oder Zahl ist ungültig | Öffnen Sie den markierten Reiter und korrigieren Sie jede Feldmeldung. |
| Es kann keine Inventarnummer vergeben werden | Lassen Sie das Inventarnummernschema **Artikel** durch einen Administrator aktivieren oder reparieren. |
| Artikelart, Unterart oder Stammdaten fehlen | Laden Sie die Ressourcen erneut. Nur aktive konfigurierte Werte können neu gewählt werden. |
| Dublettenkandidaten erscheinen | Vergleichen Sie Hersteller und Artikelnummer und brechen Sie ab oder speichern Sie die Variante bewusst. |
| Strategieänderung meldet einen Konflikt | Klären Sie unvereinbaren Bestand, Einzelstücke, Reservierungen, Einbauten oder Zustandsverlauf. |
| Suche kann nicht starten | Ergänzen Sie EAN oder die erforderliche Kombination aus Hersteller, Name, Artikelnummer und Spurweite und prüfen Sie die Einstellungen. |
| Suchbild kann nicht importiert werden | Der Artikel kann bereits gespeichert sein. Wiederholen oder verwerfen Sie nur die fehlgeschlagenen vorgemerkten Bilder. |
| Speichern ist verboten | Verwenden Sie ein Administrator- oder Bearbeiterkonto. Planungsrecht für Reservierungen ist kein Artikel-Bearbeitungsrecht. |

## Verwandte Seiten

- [Überblick zum Benutzerhandbuch](/de/guide/)
- [Zubehörübersicht](./)
- [Bestand, Einkäufe und Dokumente](./stock-purchases-documents)
- [Zuordnungen und Nutzungsverlauf](./allocations-history)
- [Artikelsuche, Web-Dokumente und Ersatzteile](../vehicles/search-and-spares)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert RailKeeper **v0.1.17.6** und wurde zuletzt am 16.08.2026 geprüft.
