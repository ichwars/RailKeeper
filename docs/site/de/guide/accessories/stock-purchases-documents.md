---
title: Bestand, Käufe und Dokumente
description: Zubehörmengen, Einzelstücke, Käufe, Bilder und Dokumente verwalten.
audience: user
status: stable
reviewedVersion: 0.1.20.3
lastReviewed: 2026-08-16
---

# Bestand, Käufe und Dokumente

RailKeeper kann austauschbare Zubehöreinheiten als Menge zählen, physische Stücke einzeln verwalten
oder Einheiten bei Bedarf aus dem Mengenbestand in die Einzelverwaltung überführen. Käufe und
Dokumente gehören zum selben Artikel, besitzen aber eigene Sofortaktionen zum Speichern. Dieses
Kapitel beschreibt das stabile Verhalten von RailKeeper v0.1.20.3.

Administratoren und Bearbeiter können diese Ressourcen verwalten. Betrachter und Planer können sie
ansehen. Das Reservierungsrecht eines Planers umfasst jedoch keine Änderungen an Bestand, Käufen,
Einzelstücken oder Dokumenten. Speichern Sie den Artikelstamm einmal, bevor Sie zugehörige
Ressourcen anlegen.

## Bestandsstrategie wählen

Die Strategie wird zusammen mit dem Artikel über **Änderungen speichern** gesichert. Sie
bestimmt, welche Bestandsbefehle und Datensätze verfügbar sind:

| Strategie | Mengenbestand | Einzelstücke | In Bestand gebuchter Kauf |
| --- | --- | --- | --- |
| **Mengenbestand** | Mengen je Lagerort korrigieren und umbuchen. | Nicht verfügbar. | Addiert die Kaufmenge am gewählten Lagerort und schreibt eine Bewegung Kauf. |
| **Einzelverwaltung** | Keine Mengenkorrektur oder Umbuchung. | Ein Datensatz je physischem Stück. | Erzeugt je gekaufter Einheit ein eingelagertes Einzelstück mit Zustand Unbekannt. Es entsteht keine Mengenbewegung. |
| **Menge, später einzeln** | Mengen korrigieren und umbuchen. | Jeweils eine verfügbare Einheit in ein Einzelstück überführen. | Erhöht den Mengenbestand wie Mengenbestand. Gekaufte Einheiten können später individualisiert werden. |

Verwenden Sie **Mengenbestand** für austauschbares Material, **Einzelverwaltung**, wenn jede Einheit
eine eigene Identität oder einen eigenen Zustand benötigt, und die Hybridstrategie, wenn Packungen
zunächst als Menge beginnen und einzelne Stücke später einen Lebenszyklus erhalten sollen.

Ein Strategiewechsel kann durch vorhandenen Bestand und Zuordnungen blockiert sein. Mengenwerte,
Mengenreservierungen oder Mengeneinbauten verhindern den Wechsel zur reinen Einzelverwaltung.
Vorhandene Einzelstücke und deren Reservierungs-, Einbau- oder Zustandsverlauf verhindern das
Entfernen der Einzelverwaltung. Ein abgelehnter Wechsel schreibt nichts.

Der **Mindestbestand** ist ein nicht negativer ganzzahliger Planungswert. Er wird mit dem Artikel
gespeichert und legt selbst weder Bestand noch Reservierung oder Bestellung an.

## Bestand und Verfügbarkeit lesen

Der Reiter **Bestand** zeigt die gezählte Menge für jeden referenzierten Lagerort. Die Summe ergibt
den Mengenbestand. Einzelstücke stehen in einer separaten Tabelle und werden nicht zu dieser Summe
addiert.

Der Mengenbestand ist die physisch vorhandene Menge. Die **verfügbare** Menge ist kleiner, wenn
aktive Mengenreservierungen bestehen. RailKeeper verhindert deshalb, dass eine Korrektur,
Umbuchung oder Individualisierung einen Lagerort unter seine aktiv reservierte Menge senkt.
Eingebaute Mengeneinheiten wurden bereits aus dem eingelagerten Bestand entfernt.

Das Bestandsjournal ist absteigend chronologisch sortiert und enthält diese stabilen Bewegungsarten:

| Bewegung | Menge und Quelle |
| --- | --- |
| Kauf | Positive Kaufmenge am Ziellagerort, mit dem Kauf verknüpft. Nur Mengen- und Hybridstrategie. |
| Korrektur | Vorzeichenbehaftete manuelle Änderung an einem Lagerort. |
| Umbuchung Ausgang | Negative Menge an der Quelle, mit derselben Umbuchung wie der Eingang verknüpft. |
| Umbuchung Eingang | Positive Menge am Ziel. |
| Individualisierung | `-1` an der Quelle, mit dem neu erzeugten Einzelstück verknüpft. |
| Einbau | Negative eingebaute Menge an der Quelle bei einem Mengeneinbau. |
| Ausbau | Positive zurückgeführte Menge am Ziel, wenn ein Mengeneinbau wieder eingelagert wird. |

Das Journal zeigt Datum, Bewegungsart, vorzeichenbehaftete Menge und Notiz. Eine Umbuchung schreibt
zwei Zeilen in einer Transaktion. Das Anlegen oder Stornieren einer Reservierung erzeugt keine
Bestandsbewegung, weil sich die Verfügbarkeit, nicht die eigene Menge ändert.

## Mengenbestand korrigieren

Verwenden Sie **Bestand korrigieren** für eine physische Korrektur, die weder Kauf noch Umbuchung
ist:

1. Wählen Sie einen aktiven Lagerort.
2. Geben Sie eine ganze Zahl ungleich null ein. Ein positiver Wert fügt Einheiten hinzu, ein
   negativer entfernt sie.
3. Wählen Sie **Buchen** und prüfen Sie die Bestätigung mit vorzeichenbehafteter Menge und
   Artikelname.
4. Bestätigen Sie, um neue Lagermenge, eine Journalzeile Korrektur und das Audit-Ereignis in einer
   Transaktion zu schreiben.

Die Aktion gibt es nur für Mengen- und Hybridartikel. RailKeeper lehnt inaktive Lagerorte, null,
nicht ganzzahlige Werte oder eine Absenkung unter aktive Reservierungen ab. Ein Abbruch der
Bestätigung schreibt nichts. Nach Erfolg werden alle Artikelressourcen neu geladen.

Verwenden Sie Käufe für Wareneingänge, zu denen Händler oder Rechnung erfasst werden sollen. Eine
positive manuelle Korrektur erzeugt keinen Kaufdatensatz. Der stabile Korrekturbefehl besitzt kein
Notizfeld.

## Bestand zwischen Lagerorten umbuchen

Wählen Sie verschiedene aktive Quell- und Ziellagerorte, geben Sie eine positive ganze Menge und
optional eine Notiz ein. **Umbuchen** öffnet eine Bestätigung. Die Bestätigung führt atomar aus:

- Die Menge wird an der Quelle abgezogen, ohne reservierte Einheiten anzutasten.
- Die Menge wird am Ziel addiert.
- Verknüpfte Journalzeilen Umbuchung Ausgang und Umbuchung Eingang erhalten dieselbe optionale
  Notiz.
- Die Artikelressourcen werden neu geladen.

Die Umbuchung ändert keinen Lagerort, wenn die Quelle nicht genügend verfügbare Menge besitzt,
Quelle und Ziel gleich sind, ein Lagerort oder ein übergeordneter Lagerort archiviert ist oder ein
Pflichtwert fehlt. Der Befehl ist nur bei Mengen- und Hybridartikeln vorhanden.

## Einzelstücke verwalten

Einzel- und Hybridartikel zeigen die Einzelstücktabelle. Bei einem Einzelartikel fügt
**Einzelstück anlegen** einen physischen Datensatz hinzu, ohne den Mengenbestand zu ändern. Bei
einem Hybridartikel verbraucht **Individualisieren** genau eine verfügbare Mengeneinheit am
gewählten Lagerort und erzeugt das Einzelstück in derselben Transaktion. Zusätzlich entsteht die
Bewegung Individualisierung mit `-1`.

| Feld des Einzelstücks | Stabile Regel |
| --- | --- |
| Inventarnummer | Optionale Stückidentität. Wenn angegeben, muss sie ohne Beachtung der Groß- und Kleinschreibung unter allen Zubehör-Einzelstücken eindeutig sein. Sie ist von der Artikel-Inventarnummer getrennt. |
| Seriennummer | Optionale Herstellerseriennummer. |
| Zustand | Einsatzbereit, Wartung fällig, Defekt oder Unbekannt. Bei manueller Erfassung anfangs Einsatzbereit. |
| Lebenszyklus | Eingelagert, Wartung oder Ausgemustert. Anfangswert Eingelagert. Reserviert und Eingebaut werden durch Zuordnungsbefehle gesteuert und sind hier nicht wählbar. |
| Lagerort | Im stabilen Formular erforderlicher aktiver Lagerort. Beim Hybridartikel stammt das Einzelstück aus genau diesem Lagerort. |
| Kaufdatum | Optionales Kalenderdatum, gespeichert im Format `JJJJ-MM-TT`. |
| Kaufpreis | Optionale nicht negative Zahl in Schritten von 0,01. |
| Garantie bis | Optionales Kalenderdatum. |
| Notizen | Optionaler lokaler Text. |

**Einzelstück speichern** öffnet immer eine Bestätigung. Bei einem Hybridartikel bricht eine zu
kleine verfügbare Menge die gesamte Individualisierung einschließlich des neuen Einzelstücks ab.
Leere Inventar- und Seriennummern sind erlaubt, eine doppelte nicht leere Inventarnummer nicht.

Über das Bearbeitungssymbol laden Sie ein vorhandenes Einzelstück in dasselbe Formular. Auch das
Speichern der Bearbeitung verlangt eine Bestätigung und ändert nur dieses Einzelstück. Ein
Einzelstück mit Lebenszyklus Reserviert oder Eingebaut besitzt in dieser Tabelle keine
Bearbeitungsaktion. Ändern Sie es über die zugehörige Reservierung oder den Einbau, damit
Lebenszyklus, Lagerort und Verlauf konsistent bleiben.

## Käufe erfassen

Die Kaufliste ist eine stabile, absteigend chronologische Historie. RailKeeper v0.1.20.3 bietet das
Hinzufügen, aber keine Aktion zum Bearbeiten oder Löschen eines Kaufs.

| Kauffeld | Stabile Regel |
| --- | --- |
| Kaufdatum | Erforderliches gültiges Datum, anfangs der heutige Tag. |
| Händler | Optionaler Text. |
| Menge | Erforderliche positive ganze Zahl. |
| Stückpreis | Optionaler Wert. Ein Zahlenwert erzeugt die angezeigte Vorschau Menge mal Stückpreis. |
| Währung | Das stabile Formular verwendet EUR. Der gespeicherte API-Wert ist, sofern vorhanden, ein dreistelliger Großbuchstabencode. |
| Rechnungsnummer | Optionaler Text. |
| Garantie bis | Optionales gültiges Datum. |
| In Bestand buchen | Ohne Auswahl entsteht nur der Kaufdatensatz. Bei Auswahl ist ein aktives Ziel erforderlich und der Bestand entsteht in derselben Transaktion. |
| Lagerort | Nur für In Bestand buchen erforderlich. |
| Notizen | Optionaler Text, der bei Mengenbestand in die Bewegung Kauf übernommen wird. |

**Kauf buchen** schreibt sofort und ohne Bestätigung. Die Wirkung hängt von der Artikelstrategie ab:

| Strategie und Option | Atomares Ergebnis |
| --- | --- |
| Beliebige Strategie, In Bestand buchen aus | Ein Kaufdatensatz, keine Änderung an Bestand, Einzelstücken oder Bewegungen. |
| Mengenbestand, In Bestand buchen an | Ein Kauf, die vollständige Menge am Lagerort und eine verknüpfte Bewegung Kauf. |
| Hybrid, In Bestand buchen an | Wie Mengenbestand. Gewählte Einheiten können später individualisiert werden. |
| Einzelverwaltung, In Bestand buchen an | Ein Kauf und je Einheit ein eingelagertes Einzelstück. Jedes übernimmt Kaufdatum, Stückpreis, Garantie, Lagerort und Kaufverknüpfung, der Zustand beginnt mit Unbekannt. Es entsteht keine Mengenbewegung. |

Kauf und optionale Bestandsbuchung laufen in einer Datenbanktransaktion. Scheitert ein Teil, werden
weder Kauf noch Bestand oder Einzelstücke gespeichert. Nach Erfolg wird das Formular zurückgesetzt
und RailKeeper lädt die zugehörigen Ressourcen neu.

## Bilder und Dokumente verwalten

Der Dokumentbereich zeigt Originaldateiname und Kategorie. Jeder Benutzer mit Leserecht für
Zubehör kann ein Dokument herunterladen. Administratoren und Bearbeiter können hochladen, das
Hauptbild wählen und löschen.

| Aktion | Bestätigung und Wirkung |
| --- | --- |
| Hochladen | Keine Bestätigung. Datei, Kategorie und optionale Beschreibung wählen. Das erste hochgeladene Bild wird zum Hauptbild, wenn noch keines vorhanden ist. Anschließend werden die Ressourcen neu geladen. |
| Als Hauptbild festlegen | Keine Bestätigung. Bei einem Bild verfügbar, das nicht Hauptbild ist. Entfernt in einer Transaktion die bisherige Hauptmarkierung, setzt dieses Bild und lädt neu. |
| Herunterladen | Kein Schreiben und keine Bestätigung. Bilder können direkt angezeigt, andere Typen als Anhang heruntergeladen werden. |
| Löschen | Destruktive Bestätigung. Entfernt die Dokumentmetadaten, löscht danach die gespeicherte Datei nur ohne weitere Referenz und lädt neu. Beim Löschen des Hauptbilds wird kein anderes Bild automatisch nachgerückt. |

Die Kategorien sind Rechnung, Lieferschein, Anleitung, Datenblatt, Grundriss, Bild und Sonstiges.
Nur ein Bild kann Hauptbild sein.

Akzeptierte Dateiendungen sind `.pdf`, `.jpg`, `.jpeg`, `.png`, `.webp`, `.zip`, `.txt`, `.csv`,
`.json` und `.xml`. Der erkannte Inhalt muss zur Endung passen. JSON, CSV und XML werden zusätzlich
strukturell geprüft. Leere Dateien, unsichere Namen, ausführbare oder skriptartige Typen, HTML,
abweichende MIME-Typen und Dateien oberhalb der konfigurierten Anhangsgrenze werden abgelehnt. Die
Standardgrenze beträgt 25 MB über `RAILKEEPER_MAX_ATTACHMENT_MB`. Hochgeladene Dateien bleiben im
lokalen Datenmodell von RailKeeper. Der Download erfordert eine autorisierte Anfrage an die
Anwendung.

Bilder aus der Artikeldatensuche verwenden einen separaten URL-Import. RailKeeper akzeptiert nur
öffentliche HTTP- oder HTTPS-Adressen, lehnt localhost sowie private oder interne Adressen vor der
Verbindung und bei Weiterleitungen ab, folgt höchstens fünf Weiterleitungen, verwendet eine
Anfragegrenze von zehn Sekunden und akzeptiert nur erkannten JPEG-, PNG- oder WebP-Inhalt innerhalb
derselben Größenbegrenzung. Der Idempotenzschlüssel verhindert, dass dasselbe vorgemerkte Bild für
den Artikel doppelt gespeichert wird.

Zubehörartikel, Mengenbestände, Bewegungen, Einzelstücke, Käufe, Dokumentmetadaten und gespeicherte
Dateien sind in der Anwendungssicherung enthalten. Halten Sie vor destruktiver Dokument- oder
Bestandsbereinigung eine aktuelle geprüfte Sicherung bereit.

## Daten bei Sofortaktionen schützen

Bestandsbefehle, Speichern von Einzelstücken, Kaufanlage, Dokument-Upload, Hauptbildwechsel und
Dokumentlöschung speichern unabhängig von der Artikelaktion **Änderungen speichern**. Ihre
Bestätigungen schützen den jeweils genannten Befehl, nicht ungespeicherte Artikelfelder an anderer
Stelle im Dialog.

Nach jeder erfolgreichen Unteraktion fragt RailKeeper Bestand, Journal, Einzelstücke, Käufe,
Dokumente, Reservierungen, Einbauten, Verlauf, Lagerorte, Fahrzeuge und Anlagenressourcen erneut ab.
Diese Lesevorgänge können teilweise fehlschlagen, nachdem der Schreibvorgang bereits bestätigt ist.
Dann gilt:

1. Behandeln Sie den Befehl als möglicherweise erfolgreich. Senden Sie ihn nicht aus veralteten
   Daten erneut.
2. Lesen Sie den Ressourcenfehler im Editor. RailKeeper kennzeichnet die Ressourcen als veraltet
   und sperrt weitere Bestands-, Dokument-, Reservierungs- und Einbauaktionen.
3. Wählen Sie **Erneut versuchen**, bis ein vollständiges Neuladen gelingt.
4. Prüfen Sie Lagerzeile, Einzelstück, Kauf, Dokument und Journal, bevor Sie über einen weiteren
   Schreibvorgang entscheiden.

Dasselbe gilt nach einer Dokumentlöschung. Die Metadatenlöschung wird vor der Bereinigung einer
nicht mehr referenzierten Datei bestätigt. Ein Fehler bei der Bereinigung kann deshalb auf einen
bereits aus der Liste entfernten Eintrag folgen. Laden Sie neu, bevor Sie erneut löschen.

## Bestands- und Dokumentfehler beheben

| Situation | Nächster Schritt |
| --- | --- |
| Korrektur oder Umbuchung meldet zu wenig Bestand | Prüfen Sie aktive Reservierungen an der Quelle. Verwenden Sie nur die freie Menge. |
| Lagerort fehlt oder wird abgelehnt | Wählen Sie einen aktiven Lagerort, dessen übergeordnete Orte ebenfalls aktiv sind. Die Lagerortverwaltung liegt außerhalb dieses Artikeldialogs. |
| Individualisierung scheitert | Prüfen Sie Hybridstrategie und mindestens eine nicht reservierte Mengeneinheit am gewählten Lagerort. |
| Inventarnummer des Einzelstücks kollidiert | Verwenden Sie eine andere Nummer oder bearbeiten Sie das vorhandene Stück. Andere Groß- und Kleinschreibung erzeugt keine eigene Nummer. |
| Reserviertes oder eingebautes Stück ist nicht bearbeitbar | Beenden oder stornieren Sie seine Zuordnung über die Reservierungs- oder Einbausteuerung. |
| Kauf wird abgelehnt | Prüfen Sie Kaufdatum, positive ganze Menge, Garantiedatum und Ziel bei Bestandsbuchung. |
| Dateityp wird nicht unterstützt | Verwenden Sie eine zulässige Endung mit tatsächlich passendem Inhalt. Umbenennen reicht nicht. |
| Datei ist zu groß | Verkleinern Sie sie unter die konfigurierte Anhangsgrenze oder lassen Sie die Grenze durch den Betreiber prüfen. |
| URL-Bild wird abgelehnt | Verwenden Sie eine öffentliche HTTP(S)-Adresse zu JPEG, PNG oder WebP ohne Weiterleitung auf eine private Adresse. |
| Schreiben erfolgreich, Neuladen fehlgeschlagen | Wiederholen Sie den Befehl nicht. Laden Sie die Ressourcen erneut und prüfen Sie zuerst Journal oder Ressourcenliste. |

## Verwandte Seiten

- [Überblick zum Benutzerhandbuch](/de/guide/)
- [Zubehörübersicht](./)
- [Artikelstammdaten und Fachangaben](./article-records)
- [Reservierungen, Einbauten und Verwendung](./allocations-history)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert RailKeeper **v0.1.20.3** und wurde zuletzt am 16.08.2026 geprüft.
