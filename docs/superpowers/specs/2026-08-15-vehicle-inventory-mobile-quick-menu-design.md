# Fahrzeugbestand: mobile Karten und sichtbares Kurzmenü

## Ziel

Die mobile Bestandsansicht wird als kompakte, aufklappbare Fahrzeugliste gestaltet. Sie soll die
für den schnellen Überblick wichtigsten Daten zuerst zeigen und weitere ausgewählte Spalten erst
bei Bedarf öffnen. Gleichzeitig wird das Kurzmenü der Desktop-Tabelle so korrigiert, dass es nicht
mehr an der Zeilengrenze abgeschnitten wird.

Die Änderung ergänzt die Umsetzung der Issues #80 und #81. Suche, Filter, gespeicherte
Spaltenauswahl, Spaltenreihenfolge und fachliche Daten bleiben unverändert.

## Gewählte mobile Darstellung

Von drei verglichenen Varianten wird Variante B, die aufklappbare Karte, umgesetzt. Jede Karte ist
beim Laden eingeklappt und kann unabhängig von anderen Karten geöffnet oder geschlossen werden.
Der Zustand gilt nur für die aktuelle Ansicht und wird weder im Browser noch serverseitig
gespeichert.

Der eingeklappte Kopf zeigt in kompakter Form:

- das Fahrzeugbild,
- Inventarnummer und Bezeichnung,
- Hersteller und Artikelnummer,
- Spurweite und Epoche.

Die Angaben werden nur gezeigt, wenn die zugehörige Spalte in der gespeicherten Auswahl sichtbar
ist. Die bestehende Sicherheitsregel bleibt erhalten: Wenn keine Datenspalte mehr sichtbar wäre,
wird die Inventarnummer automatisch eingeblendet. Ist die Bildspalte ausgeblendet, schließt der
Textbereich die freie Fläche. Ist sie sichtbar, aber kein Bild vorhanden, erscheint die vorhandene
neutrale Bilddarstellung.

Beim Aufklappen erscheinen alle weiteren sichtbaren Spalten in der gespeicherten Reihenfolge.
Bereits im Kopf dargestellte Werte werden im Detailbereich nicht wiederholt. Lange Werte brechen
innerhalb der Karte um und erzeugen keinen horizontalen Dokumentüberlauf.

## Bedienung

Die gesamte Kopfzeile erhält eine eindeutige Schaltfläche zum Ein- und Ausklappen. Beschriftung,
`aria-expanded` und Fokusdarstellung machen den Zustand auch ohne Farbe verständlich. Ein Klick auf
die Aufklappsteuerung öffnet nicht versehentlich die Fahrzeugdetailseite.

Anzeigen, Bearbeiten und Kurzmenü bleiben als eigener kompakter Aktionsbereich erreichbar. Die
Aktionen besitzen zugängliche Namen und genügend Abstand für Touch-Bedienung. Das Kurzmenü liegt
über nachfolgenden Karten und wird nicht durch deren Begrenzung abgeschnitten.

Die bestehende Spaltenverwaltung steuert weiterhin Desktop-Tabelle und mobile Liste gemeinsam.
Auswahl, Reihenfolge, automatische Inventarnummer und "Auf Standard zurücksetzen" erhalten mobil
keine abweichende Speicherung oder zweite Konfiguration.

## Komponenten und Datenfluss

`VehicleInventoryMobileList` bleibt für die mobile Liste verantwortlich. Die Komponente erhält
weiterhin die normalisierte Spaltendefinition, Fahrzeugdaten und vorhandenen Aktionshandler. Ein
lokaler Satz geöffneter Fahrzeug-IDs verwaltet ausschließlich den Aufklappzustand.

Eine kleine, rein darstellende Kartenkomponente trennt Kopf, zusätzliche Felder und Aktionen. Sie
verwendet dieselben Spaltendefinitionen und Formatierer wie die Desktop-Tabelle. Dadurch bleiben
Beschriftungen, Ja/Nein-Werte und Feldreihenfolge konsistent, ohne eine zweite fachliche
Feldzuordnung einzuführen.

Es entstehen keine neuen API-Endpunkte, Datenbankfelder oder Profileinstellungen. Filter-, Such-
und Spaltenänderungen liefern wie bisher die sichtbare Ergebnisliste; nicht mehr vorhandene
Fahrzeug-IDs werden aus dem lokalen Aufklappzustand entfernt.

## Desktop-Kurzmenü

Das Desktop-Kurzmenü wird derzeit durch `overflow: hidden` an der Tabellenzelle mit den Aktionen
auf die Höhe der Fahrzeugzeile beschnitten. Die Tabellenstile werden so eingegrenzt, dass die
Ellipsis-Regel weiterhin nur Datenzellen kürzt, während Aktionszelle und Menü sichtbaren Überlauf
und eine passende Stapelreihenfolge erhalten.

Die Korrektur ändert weder Inhalt noch Berechtigungen des Menüs. Das Menü bleibt am auslösenden
Button ausgerichtet, liegt über den folgenden Zeilen und bleibt innerhalb des sichtbaren
Tabellenbereichs bedienbar.

## Fehler- und Randfälle

- Fahrzeuge ohne Bild verwenden die vorhandene neutrale Bilddarstellung, sofern die Bildspalte
  sichtbar ist.
- Leere Werte werden mit der vorhandenen neutralen Darstellung ausgegeben, nicht als leere
  unbeschriftete Zeile.
- Nach Filter-, Such- oder Seitenwechsel bleiben keine verwaisten geöffneten Karten zurück.
- Sehr lange deutsche und englische Werte brechen innerhalb der Karte um.
- Ein geöffnetes Kurzmenü wird beim bestehenden Außenklick, mit Escape oder nach einer Aktion
  geschlossen.
- Viewer-, Editor- und Admin-Berechtigungen bleiben unverändert; nicht erlaubte Aktionen werden
  weiterhin nicht angeboten.

## Tests und lokale Abnahme

Frontend-Tests decken mindestens folgende Fälle ab:

- mobile Karten sind anfänglich eingeklappt,
- Bild und kompakte Kopfdaten erscheinen entsprechend der sichtbaren Spalten,
- Auf- und Zuklappen wirkt nur auf die gewählte Karte,
- zusätzliche Felder erscheinen ohne Duplikate in gespeicherter Reihenfolge,
- ausgeblendete Felder und Bildbereiche erzeugen keine Lücken,
- Aktionsschaltflächen lösen nicht versehentlich das Aufklappen aus,
- das Desktop-Kurzmenü wird nicht durch die Aktionszelle abgeschnitten.

Danach werden die betroffenen Frontend-Tests und `npm.cmd run build` ausgeführt. Die lokale
Sichtprüfung umfasst Mobil- und Desktopbreite, Hell- und Dunkelmodus, lange deutsche Texte,
Fahrzeuge mit und ohne Bild sowie das Kurzmenü in oberen und unteren Tabellenzeilen. Eine
Veröffentlichung oder ein Push erfolgt erst nach der gemeinsamen lokalen Abnahme.
