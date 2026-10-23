# ECoS compatibility diagnostics

## Purpose

The ECoS connection test is both a transport check and the safety gate for activating the adapter.
Receiving any TCP data is therefore insufficient. RailKeeper requires a successful ECoS end marker and
the complete set of connection-identification fields before activation is allowed.

## Required response

The read-only command remains `get(1, info, status)`. A usable response must contain:

- `status`;
- `ProtocolVersion`;
- `ApplicationVersion`;
- `HardwareVersion`;
- `<END 0 (OK)>`.

No runtime control command is sent by this check.

## Status model

| Status | Meaning | Adapter activation |
| --- | --- | --- |
| `unreachable` | No complete transport exchange could be made. | Blocked |
| `rejected` | ECoS returned a non-zero end status. | Blocked |
| `partial` | The end marker or at least one required field is missing. | Blocked |
| `unverified` | The response is complete, but this application firmware has no recorded real-device acceptance in RailKeeper. | Allowed with a visible warning |
| `verified` | The response is complete and the application firmware matches a recorded real-device acceptance. | Allowed |

The current verified entry is application firmware **4.3.3**, accepted with real ESU ECoS hardware during
PR #143. That acceptance did not add an anonymized raw response to the repository. RailKeeper therefore
does not claim a hardware fixture or broaden the matrix beyond the recorded firmware version.

Complete responses from other versions remain usable because their required protocol shape is present,
but the UI labels them as not yet device-verified. A new version becomes `verified` only after a documented
real-device run and, where available, an anonymized regression fixture.

## Failure handling

Missing fields are returned as a bounded string array and shown in the commissioning workflow. Rejected
end status and a missing end marker have separate classifications and messages. Raw response lines retain
the existing diagnostic behavior, while server-side `connected` stays false for rejected and partial
responses so the frontend cannot activate the adapter from an incomplete test.

## Lokliste mit Mehrfachtraktionen / Roster containing consists

`queryObjects(10, addr, name, protocol)` kann neben einzelnen Fahrzeugen Verbünde mit
`protocol[MULTI]` und ohne `addr` liefern. Diese ECoS-Objekte werden im Rohabruf erhalten,
aber nicht mit Lok-/CV-Detailabfragen geprüft und nicht als Fahrzeuge in die Arbeitsliste übernommen.
Die Abschlussmeldung nennt die Anzahl der ausgelassenen Verbünde. Ungültige IDs, doppelte IDs,
ungültige Namen und fehlerhafte einzelne Lokdatensätze werden weiterhin abgewiesen.

The ECoS roster may contain `protocol[MULTI]` consists without an `addr` attribute. Raw diagnostics
retain their roster summaries; locomotive-only detail/CV probes and workspace vehicle comparison
exclude them. The completion message reports the excluded count. This exception applies only to
ECoS consists, and does not relax object identity, name, or individual locomotive validation.

The distinction between locomotive objects and consists is described in ESU's
[PC interface specification, edition 3, section 7.3](https://www.esu.eu/uploads/tx_esudownloads/Netzwerkspezifikation_fuer_PC-Interface_ESU_KG_DE_Auflage_3.pdf).
The exact `MULTI` response shape was observed on application firmware **4.3.3**, protocol **0.5**;
an anonymized mixed-roster regression test covers it. This does not add general consist management.
