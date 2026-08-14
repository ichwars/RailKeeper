# Entwurf: Bitte um Prüfung der begrenzten ECoS-Anbindung in RailKeeper

> **Status: Entwurf, nicht versendet.**
>
> Dieses Schreiben darf erst nach einer neuen ausdrücklichen Freigabe versendet werden.

An die<br>
ESU electronic solutions ulm GmbH & Co. KG<br>
Edisonallee 29<br>
89231 Neu-Ulm

## Betreff

Bitte um Prüfung einer begrenzten ECoS-Anbindung und der verwendeten Funktionstastensymbole im
freien Projekt RailKeeper

Sehr geehrte Damen und Herren,

ich entwickle RailKeeper, eine kostenlose, lokal betriebene und quelloffene Anwendung zur
Verwaltung von Modelleisenbahn-Fahrzeugen. Das Projekt ist unter
<https://github.com/ichwars/RailKeeper> öffentlich einsehbar und wird unter der Lizenz
GNU Affero General Public License Version 3, ausschließlich Version 3 (`AGPL-3.0-only`),
veröffentlicht.

RailKeeper ist ein unabhängiges Projekt und steht in keiner Verbindung zu ESU. Eine Zertifizierung,
Freigabe oder offizielle Kompatibilität durch ESU wird nicht behauptet. ECoS ist eine Marke der
ESU electronic solutions ulm GmbH & Co. KG.

Ich möchte Sie frühzeitig und transparent um eine Prüfung des begrenzten Funktionsumfangs bitten.
Die ECoS-Anbindung dient ausschließlich dazu, ausgewählte Lokstammdaten, statische
Funktionstastenbeschreibungen und CV-Werte in die lokale Fahrzeugsammlung zu übernehmen. Vor einer
Änderung zeigt RailKeeper eine Vorschau und verlangt eine ausdrückliche Bestätigung.

## Verwendeter Funktionsumfang

RailKeeper liest derzeit ausschließlich folgende Informationen:

- ECoS-Systemobjekt 1: `info` und `status` zur Verbindungserkennung
- Lokomotiv-Objektmanager 10: `addr`, `name` und `protocol` zur Auflistung
- einzelne Lokomotivobjekte: `profile`, `protocol`, `name`, `addr` und `funcdesc`
- zusätzliche statische Decoderinformationen: `cv`, `cvs`, `cvlist` und `functionmapping`
- gezielte CV-Abfragen: CV 1 bis 8 sowie CV 7, 8, 17, 18 und 29

Nach einer Vorschau und ausdrücklichen Bestätigung kann RailKeeper bei einem ausgewählten
Lokomotivobjekt ausschließlich `name`, `addr` und `protocol` schreiben. Weitere Schreib- oder
Steuerbefehle sind nicht vorgesehen.

Bewusst nicht verwendet werden:

- aktuelle Geschwindigkeit oder Fahrstufe
- aktuelle Fahrtrichtung
- aktive Funktionszustände oder `funcset`
- Fahr-, Funktions-, STOP- oder GO-Befehle
- Schaltartikel beziehungsweise Magnetartikel und deren Objektmanager
- Fahrwege
- S88- beziehungsweise Rückmeldeobjekte
- Booster und deren Objektmanager
- sonstige ECoS-Objektmanager
- Lokbilder, Bildreferenzen oder Bildfelder aus der ECoS

## Funktionstastensymbole

RailKeeper enthält lokal eingebettete SVG-Funktionstastensymbole und Zuordnungen. Die
Repository-Metadaten nennen als Herkunft beziehungsweise Arbeitsgrundlage:

- `Funktionstasten_SVG_Variante_1_172_Symbole.zip`
- `ESU_Funktionssymbole_V1_Variante2_Feinlinien_aktiv_inaktiv_SVG.zip`
- `50200_ECoS_Uebersicht_Funktionstastensymbole_ESU_KG_DE-EN_Auflage-3-1.pdf`

Die Symbole werden nicht automatisch aus einer ECoS ausgelesen. Ich möchte ausdrücklich klären,
ob ESU Einwände gegen die Aufnahme und Weitergabe dieser Symbole oder Zuordnungen im öffentlichen
RailKeeper-Repository hat und ob dafür besondere Kennzeichnungen, Bedingungen oder eine gesonderte
Erlaubnis erforderlich sind. Die AGPL-Lizenz des Projekts soll keine Rechte an Marken, Grafiken,
Dokumentationen oder sonstigen Rechten Dritter beanspruchen oder übertragen.

## Bitte um Rückmeldung

Bitte teilen Sie mir möglichst konkret mit,

- ob ESU gegen den beschriebenen lesenden oder schreibenden Funktionsumfang Einwände hat,
- ob einzelne Abfragen, Felder oder Schreibvorgänge entfernt oder angepasst werden sollen,
- ob die Funktionstastensymbole oder Zuordnungen entfernt, ersetzt oder gesondert lizenziert werden
  müssen,
- welche Marken-, Urheberrechts- oder sonstigen Hinweise ESU für erforderlich hält.

Sollte ESU Einwände haben, bin ich bereit, die betroffenen Funktionen, Daten oder Grafiken
anzupassen oder aus RailKeeper zu entfernen. Eine konkrete Benennung der beanstandeten Bestandteile
würde mir eine schnelle und zielgerichtete Umsetzung ermöglichen.

Vielen Dank für Ihre Prüfung.

Mit freundlichen Grüßen

Daniel Roth<br>
GitHub: <https://github.com/ichwars><br>
Projekt: <https://github.com/ichwars/RailKeeper>
