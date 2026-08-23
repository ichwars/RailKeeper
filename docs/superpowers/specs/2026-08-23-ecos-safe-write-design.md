# Sicheres Zurückschreiben von Lokdaten in die ECoS

## Deutsch

### Ziel

RailKeeper soll geprüfte Lokstammdaten sicher in eine ESU ECoS zurückschreiben. Die erste
produktive Ausbaustufe bleibt bewusst auf Name, Decoder-Adresse und Protokoll begrenzt. Jeder
Schreibvorgang benötigt eine frische Vorschau, eine ausdrückliche Bestätigung und eine erfolgreiche
Kontroll-Lesung.

Der vorhandene Schreibpfad wird gehärtet und an einer realen ECoS abgenommen. Er wird nicht durch
einen zweiten parallelen Ablauf ersetzt.

### Ausgangslage

RailKeeper besitzt bereits:

- einen Admin-geschützten ECoS-Lese- und Vergleichsablauf;
- eine nicht schreibende Vorschau für Name, Decoder-Adresse und Protokoll;
- eine benutzergebundene, einmal verwendbare und zehn Minuten gültige Schreibfreigabe;
- einen kombinierten ECoS-Schreibbefehl;
- eine gezielte Kontroll-Lesung und Audit-Einträge;
- eine Zuordnung zwischen ECoS-Objekt und RailKeeper-Fahrzeug.

Die vorhandene Implementierung ist automatisiert getestet, koordiniert den Live-Monitor und andere
ECoS-Operationen aber noch nicht als einen exklusiven Ablauf. Außerdem fehlt eine blockierende
Prüfung, ob eine gewünschte Decoder-Adresse bereits von einem anderen ECoS-Objekt verwendet wird.

### Umfang

Version 1 schreibt ausschließlich:

- Lokname;
- Decoder-Adresse;
- Protokoll.

CV-Werte, Funktionsdefinitionen, aktive Funktionen, Geschwindigkeit, Richtung, Bilder,
Anlagenobjekte und neue ECoS-Lokobjekte bleiben außerhalb dieses Vorhabens.

### Abgleich mit Issue #132

Dieses Design konkretisiert die in Issue #132 bereits beschriebene Härtung des vorhandenen
Schreibpfads. Es erweitert den dort freigegebenen Schreibumfang nicht.

- `WriteLocomotives` bleibt nur für den vollständig geschützten ECoS-Ablauf aktiv.
- `WriteCVs` bleibt deaktiviert.
- Die frische Loklistenabfrage für Adresskonflikte verwendet ausschließlich freigegebene
  Lokstammdaten. Laufzeit- und Anlagenzustände werden weder gelesen noch geschrieben.
- Anonymisierte Realgeräteantworten ergänzen nach der Geräteabnahme das Parser- und
  Regressionstestkorpus.
- Hardwaremodell, Firmwarestand und festgestellte Abweichungen werden in der ECoS-
  Kompatibilitätsmatrix dokumentiert.

### Architektur

Der `DigitalCenterWorkspaceService` bleibt für Berechtigungen, Sitzungen, Vergleich,
Konfliktprüfung, Vorschau, Freigaben, Audit und Ergebnisaufbereitung verantwortlich. Der
`ECoSService` bleibt die Grenze zur Gerätekommunikation.

Eine exklusive ECoS-Operationssperre serialisiert folgende Vorgänge:

- Daten lesen;
- Live-Monitor starten oder stoppen;
- Schreibvorgang mit Kontroll-Lesung.

Der koordinierte Schreibvorgang merkt sich den ursprünglichen Live-Zustand. War der Live-Monitor
aktiv, wird er vor jeder schreibrelevanten Vorprüfung kontrolliert beendet. Nach Abschluss oder
Abbruch wird er wieder gestartet. Ein fehlgeschlagener Neustart verändert ein bereits verifiziertes
Schreibergebnis nicht, sondern erzeugt eine getrennte Warnung und Diagnosemeldung.

Parallele Schreibbestätigungen werden serialisiert. Eine Freigabe bleibt trotzdem einmalig und wird
vor der ersten möglichen Geräteänderung atomar verbraucht.

### Vorschau

Die Vorschau ist vollständig lesend und verwendet die in den Einstellungen hinterlegte ECoS.
Clientseitig übermittelte Hosts oder Ports werden nicht akzeptiert.

RailKeeper prüft:

1. Admin-Rolle und aktive ECoS-Konfiguration;
2. fertige, höchstens zehn Minuten alte Lesesitzung;
3. eindeutige Zuordnung zwischen Arbeitslisteneintrag, ECoS-Objekt und RailKeeper-Fahrzeug;
4. konfliktfreien Arbeitslistenzustand;
5. unterstützte und nicht leere Sollwerte;
6. tatsächliche Abweichungen der ausgewählten Felder;
7. falls die Decoder-Adresse geschrieben werden soll: keine Verwendung der gewünschten Adresse
   durch ein anderes ECoS-Lokobjekt.

Ein Adresskonflikt blockiert die Vorschau. Die Oberfläche nennt Name und ECoS-Objekt-ID der anderen
Lok, soweit diese Daten sicher gelesen wurden.

Die Freigabe bindet Sitzung, Arbeitslisteneintrag, Anbieter, ECoS-Objekt-ID, Richtung, Felder,
Ausgangswerte, Sollwerte und Benutzer kryptografisch aneinander. Persistiert wird nur der Hash des
öffentlichen Tokens.

### Bestätigung und Schreiben

Nach der ausdrücklichen Bestätigung läuft der Vorgang unter der exklusiven Operationssperre:

1. Freigabe atomar verbrauchen.
2. Ursprünglichen Live-Zustand erfassen.
3. Live-Monitor kontrolliert stoppen, falls er aktiv ist.
4. ECoS-Lokliste und Zielobjekt frisch lesen.
5. Zuordnung, Ausgangswerte, Sollwerte und bei freigegebener Adressänderung die Adressbelegung
   erneut prüfen.
6. Vorschau-Hash neu berechnen und mit der Freigabe vergleichen.
7. Genau einen kombinierten ECoS-Befehl für die noch abweichenden freigegebenen Felder senden.
8. Antwortblock vollständig prüfen.
9. Zielobjekt gezielt zurücklesen.
10. Alle geschriebenen Felder normalisiert und feldweise vergleichen.
11. Arbeitslisteneintrag, externe Zuordnung und Audit-Ergebnis aktualisieren.
12. Live-Monitor wieder starten, falls er zuvor aktiv war.

Jede Abweichung vor Schritt 7 beendet den Ablauf ohne Geräteänderung und verlangt eine neue
Vorschau. RailKeeper wiederholt einen unklaren oder fehlgeschlagenen Schreibbefehl niemals
automatisch.

### Adresskonflikte

Soll die Decoder-Adresse geschrieben werden, wird sie gegen alle frisch gelesenen
ECoS-Lokobjekte geprüft. Das Zielobjekt selbst wird ausgenommen. Verwendet mindestens ein anderes
Objekt dieselbe Adresse, wird der Vorgang blockiert. Reine Namens- oder Protokolländerungen werden
nicht durch einen bereits vorhandenen, unveränderten Adresskonflikt blockiert.

Eine bewusste Übersteuerung ist in Version 1 nicht vorgesehen. Der Benutzer muss zuerst den
Adresskonflikt an der richtigen Stelle auflösen und anschließend neu lesen.

### Verifikation und Zustände

Ein Schreibvorgang kann folgende fachliche Ergebnisse haben:

- **Geschrieben und verifiziert:** ECoS meldet alle freigegebenen Sollwerte zurück.
- **Ohne Änderung abgebrochen:** Vorschau ist veraltet, die Zuordnung änderte sich oder ein
  Adresskonflikt entstand.
- **Schreibprüfung fehlgeschlagen:** ECoS beantwortete den Schreibbefehl, meldet danach aber
  abweichende Werte.
- **Schreibzustand unklar:** Schreibantwort oder Kontroll-Lesung fehlt. RailKeeper darf weder Erfolg
  behaupten noch automatisch erneut schreiben.
- **Schreiben fehlgeschlagen:** ECoS lehnt den Befehl eindeutig ab, bevor ein Erfolg bestätigt
  werden kann.

Nach erfolgreicher Verifikation werden die zurückgelesenen ECoS-Werte in den Arbeitslisteneintrag
übernommen, der Vergleich neu berechnet und die externe Zuordnung als synchronisiert markiert. Bei
einem unklaren oder abweichenden Ergebnis bleibt die Zuordnung unsynchronisiert und die tatsächlich
gelesene Abweichung sichtbar.

### Oberfläche

Der vorhandene Lok-Vergleichsdialog bleibt der einzige Einstieg:

1. Unterschiede anzeigen;
2. **Schreibvorschau erstellen**;
3. Richtung `RailKeeper → ECoS`, Felder, Ausgangs- und Sollwerte anzeigen;
4. ausdrückliche Bestätigungs-Checkbox aktivieren;
5. **In ECoS schreiben**;
6. Fortschritt, Ergebnis und Kontroll-Lesung anzeigen.

Während Pause, Vorprüfung, Schreiben, Kontroll-Lesung und Neustart sind weitere Schreibaktionen
gesperrt. Doppelklicks erzeugen keinen zweiten Vorgang.

Die Oberfläche unterscheidet klar:

- roten blockierenden Adress- oder Vorschaukonflikt;
- roten beziehungsweise neutralen unklaren Schreibzustand mit Hinweis, nicht erneut zu klicken;
- grünen verifizierten Erfolg;
- gelbe Warnung, wenn nur der Neustart des Live-Monitors scheitert.

Ein Neustartfehler bietet die vorhandene Aktion zum manuellen Start des Live-Monitors an. Normale
Fehlermeldungen enthalten keine Rohbefehle, Verbindungsgeheimnisse oder internen Netzwerkdetails.

### API und Datenvertrag

Die vorhandenen Vorschau- und Bestätigungsendpunkte bleiben bestehen. Ihre Antworten werden nur
erweitert, wenn die Oberfläche strukturierte Angaben für folgende Zustände benötigt:

- kollidierendes ECoS-Objekt;
- tatsächlich verifizierte Werte;
- unklarer Schreibzustand;
- Ergebnis des Live-Monitor-Neustarts.

Backend, `frontend/src/shared/api.ts` und `openapi/railkeeper.yaml` werden gemeinsam geändert. Alle
Schreibendpunkte bleiben Admin- und CSRF-geschützt. Host und Port stammen ausschließlich aus den
serverseitigen Einstellungen.

### Audit und Diagnose

Jeder bestätigte Versuch erhält einen Audit-Eintrag mit Benutzer, Fahrzeug, Anbieter,
ECoS-Objekt-ID, freigegebenen Feldern und fachlichem Ergebnis. Tokens, Rohbefehle und mögliche
Verbindungsgeheimnisse werden nicht protokolliert.

Ein fehlgeschlagener Live-Neustart wird zusätzlich als Arbeitsbereichsmeldung und in der
Live-Diagnose erfasst. Ein verifizierter Gerätewert bleibt dabei unverändert erfolgreich.

### Tests

Automatisierte Tests decken ab:

- simulierte ECoS-TCP-Kommunikation für kombinierten `set`-Befehl und anschließendes `query`;
- anonymisierte Gerätefixtures für erfolgreiche, unvollständige, fehlerhafte und unterbrochene
  Antworten;
- Normalisierung und exakten Vergleich von Name, Adresse und Protokoll;
- blockierte doppelte Decoder-Adressen;
- veraltete, verbrauchte, benutzerfremde und manipulierte Freigaben;
- geänderte Ausgangswerte zwischen Vorschau und Bestätigung;
- Serialisierung paralleler Bestätigungen;
- Stopp und Wiederanlauf eines zuvor aktiven Live-Monitors;
- vollständigen Abbruch, wenn der Live-Monitor nicht sauber pausiert werden kann;
- eindeutige Ablehnung, Verifikationsabweichung und unklaren Schreibzustand;
- keine automatische Wiederholung eines Schreibbefehls;
- Fortbestand eines verifizierten Ergebnisses bei fehlgeschlagenem Live-Neustart;
- Arbeitslisten-, Mapping- und Audit-Aktualisierung;
- API-Vertrag, Rollen- und CSRF-Schutz;
- deutsche und englische UI-Texte, Ladezustände, Konflikte und Mehrfachklickschutz.

### Reale ECoS-Abnahme

Die produktive Freigabe erfolgt mit einer bewusst ausgewählten Testlok:

1. Ursprungswerte dokumentieren.
2. Nur den Namen ändern, schreiben und zurücklesen.
3. Eine nachweislich freie Decoder-Adresse schreiben und zurücklesen.
4. Eine Protokolländerung separat schreiben und zurücklesen.
5. Den automatischen Pause- und Wiederanlauf des Live-Monitors prüfen.
6. Einen absichtlichen Adresskonflikt als blockierten Negativtest prüfen.
7. Ursprungswerte über denselben geprüften Ablauf wiederherstellen.

Die Abnahme dokumentiert ECoS-Modell, Firmware-Version, getestete Lok-Objekt-ID und Ergebnis. Nach
Entfernung von Hostnamen, IP-Adressen, individuellen Loknamen und anderen lokalen Kennungen werden
geeignete Antwortblöcke als Regressionstest-Fixtures übernommen. Das Ergebnis aktualisiert außerdem
die ECoS-Kompatibilitätsmatrix. Bis dahin bleibt die Funktion als noch nicht realgeräteverifiziert
gekennzeichnet.

---

## English

### Goal

RailKeeper shall safely write reviewed locomotive master data back to an ESU ECoS. The first
production scope is deliberately limited to name, decoder address, and protocol. Every write
requires a fresh preview, explicit confirmation, and a successful verification read.

The existing write path will be hardened and accepted against a real ECoS. It will not be replaced
by a second parallel workflow.

### Current state

RailKeeper already provides:

- an Admin-protected ECoS read and comparison workflow;
- a non-mutating preview for name, decoder address, and protocol;
- an actor-bound, single-use write grant that expires after ten minutes;
- one combined ECoS write command;
- a targeted verification read and audit records;
- a mapping between an ECoS object and a RailKeeper vehicle.

The current implementation is covered by automated tests, but it does not yet coordinate the live
monitor and other ECoS operations as one exclusive operation. It also lacks a blocking check for a
desired decoder address already used by another ECoS object.

### Scope

Version 1 writes only:

- locomotive name;
- decoder address;
- protocol.

CV values, function definitions, active functions, speed, direction, images, layout objects, and
new ECoS locomotive objects are out of scope.

### Alignment with issue #132

This design details the hardening of the existing write path already described in issue #132. It
does not expand the write scope approved there.

- `WriteLocomotives` remains enabled only for the fully protected ECoS workflow.
- `WriteCVs` remains disabled.
- The fresh locomotive-list query used for address conflicts reads only approved locomotive master
  data. Runtime and layout state are neither read nor written.
- Anonymized real-device replies extend the parser and regression fixture corpus after device
  acceptance.
- Hardware model, firmware version, and observed deviations are recorded in the ECoS compatibility
  matrix.

### Architecture

`DigitalCenterWorkspaceService` remains responsible for authorization, sessions, comparison,
conflicts, previews, grants, auditing, and result presentation. `ECoSService` remains the device
communication boundary.

An exclusive ECoS operation gate serializes:

- data reads;
- live monitor start and stop;
- a write followed by its verification read.

The coordinated write remembers the original live state. If the live monitor was running, it is
stopped cleanly before any write-related preflight. It is restarted after completion or abort. A
failed restart does not change an already verified write result; it creates a separate warning and
diagnostic message.

Concurrent write confirmations are serialized. Each grant is still single-use and is consumed
atomically before the first possible device mutation.

### Preview

The preview is read-only and uses the ECoS configured in server settings. Client-supplied hosts or
ports are not accepted.

RailKeeper verifies:

1. Admin role and active ECoS configuration;
2. a completed read session no older than ten minutes;
3. an unambiguous relationship between work item, ECoS object, and RailKeeper vehicle;
4. a conflict-free work item;
5. supported and non-empty desired values;
6. actual differences for the selected fields;
7. when the decoder address is selected for writing, no other ECoS locomotive object uses the
   desired address.

An address collision blocks the preview. The UI identifies the other locomotive by name and ECoS
object ID when those values were read safely.

The grant cryptographically binds session, work item, provider, ECoS object ID, direction, fields,
current values, desired values, and actor. Only the public token hash is persisted.

### Confirmation and write

After explicit confirmation, the operation runs under the exclusive gate:

1. Consume the grant atomically.
2. Capture the original live state.
3. Stop the live monitor cleanly if it was running.
4. Read the ECoS locomotive list and target object again.
5. Recheck mapping, current values, desired values, and address ownership when an address change
   was approved.
6. Recompute the preview hash and compare it with the grant.
7. Send exactly one combined ECoS command for the approved fields that still differ.
8. Validate the complete reply block.
9. Read the target object again.
10. Normalize and compare every written field.
11. Update work item, external mapping, and audit result.
12. Restart the live monitor if it was previously running.

Any mismatch before step 7 aborts without a device mutation and requires a new preview. RailKeeper
never retries an ambiguous or failed write command automatically.

### Address conflicts

When the decoder address is selected for writing, it is checked against all freshly read ECoS
locomotive objects. The target object itself is excluded. If any other object uses the same
address, the operation is blocked. A name-only or protocol-only change is not blocked by an
existing unchanged address collision.

Version 1 does not offer an override. The user must resolve the collision at the correct source and
perform a fresh read.

### Verification and states

A write can produce these domain results:

- **Written and verified:** the ECoS reports all approved desired values.
- **Aborted without change:** the preview became stale, the mapping changed, or an address collision
  appeared.
- **Write verification failed:** the ECoS answered the write but reports different values afterward.
- **Write state unknown:** the write reply or verification read is unavailable. RailKeeper must not
  claim success or retry automatically.
- **Write failed:** the ECoS clearly rejected the command before success could be established.

After successful verification, the returned ECoS values update the work item, comparison state, and
external mapping. For unknown or mismatching outcomes, the mapping remains unsynchronized and any
values that were actually read remain visible.

### User interface

The existing locomotive comparison dialog remains the only entry point:

1. Show differences.
2. Select **Create write preview**.
3. Show direction `RailKeeper → ECoS`, fields, current values, and desired values.
4. Enable the explicit confirmation checkbox.
5. Select **Write to ECoS**.
6. Show progress, result, and verification read.

Further write actions are disabled while pausing, checking, writing, verifying, and restarting.
Repeated clicks cannot create a second operation.

The UI clearly distinguishes:

- a red blocking address or stale-preview conflict;
- a red or neutral unknown write state with an instruction not to retry immediately;
- a green verified result;
- an amber warning when only the live monitor restart failed.

A restart warning offers the existing manual live-monitor start action. Normal errors do not expose
raw commands, connection secrets, or internal network details.

### API and data contract

The existing preview and confirmation endpoints remain. Their responses are extended only where
the UI needs structured data for:

- the colliding ECoS object;
- the values actually verified;
- an unknown write state;
- the live monitor restart outcome.

Backend, `frontend/src/shared/api.ts`, and `openapi/railkeeper.yaml` change together. All write
endpoints remain Admin-only and CSRF-protected. Host and port come exclusively from server-side
settings.

### Audit and diagnostics

Every confirmed attempt creates an audit record containing actor, vehicle, provider, ECoS object
ID, approved fields, and domain result. Tokens, raw commands, and possible connection secrets are
not logged.

A failed live restart is also recorded as a workspace message and live diagnostic. An already
verified device result remains successful.

### Tests

Automated coverage includes:

- simulated ECoS TCP communication for one combined `set` and the following `query`;
- anonymized device fixtures for successful, incomplete, malformed, and interrupted replies;
- normalization and exact comparison of name, address, and protocol;
- blocked duplicate decoder addresses;
- stale, consumed, actor-mismatched, and tampered grants;
- changed current values between preview and confirmation;
- serialization of concurrent confirmations;
- stop and restart of a previously running live monitor;
- complete abort when the live monitor cannot be paused cleanly;
- explicit rejection, verification mismatch, and unknown write state;
- no automatic retry of a write command;
- preserved verified success when live restart fails;
- work item, mapping, and audit updates;
- API contract, role, and CSRF protection;
- German and English UI copy, loading states, conflicts, and repeat-click prevention.

### Real ECoS acceptance

Production acceptance uses a deliberately selected test locomotive:

1. Record the original values.
2. Change only the name, write it, and read it back.
3. Write a demonstrably unused decoder address and read it back.
4. Write a protocol change separately and read it back.
5. Verify automatic live-monitor pause and restart.
6. Verify an intentional address collision as a blocked negative test.
7. Restore the original values through the same reviewed workflow.

The acceptance record includes ECoS model, firmware version, tested locomotive object ID, and
result. Suitable reply blocks become regression fixtures after removing host names, IP addresses,
individual locomotive names, and other local identifiers. The result also updates the ECoS
compatibility matrix. Until then, the feature remains marked as not yet verified against real
hardware.
