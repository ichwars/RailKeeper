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
