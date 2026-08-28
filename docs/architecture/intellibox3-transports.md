# Intellibox 3 transport and validation boundary

## Decision

RailKeeper models the Uhlenbrock Intellibox 3 as one provider with two independent transports:

| Transport | Server status | Enabled capabilities | Network behavior |
| --- | --- | --- | --- |
| `z21_udp` | `available` | Connection test and read-only diagnostics | Sends only `LAN_GET_SERIAL_NUMBER`, `LAN_GET_CODE`, and `LAN_GET_HWINFO` to the configured UDP port. |
| `loconet_tcp` | `planned` | None | No port, session, framing, reconnect, import, monitoring, or write implementation exists. RailKeeper sends no LocoNet TCP request. |

Provider-level capabilities remain the union of implemented transport capabilities. Locomotive import,
live monitoring, locomotive writes, and CV writes remain disabled for Intellibox 3 in the server and UI.

## Z21 protocol boundary

The UDP adapter follows the framing constraints documented by the Z21 LAN protocol specification 1.13:

- the primary UDP port is 21105;
- a UDP payload is limited to 1472 bytes;
- one UDP datagram may contain multiple length-prefixed datasets;
- asynchronous datasets may arrive between request and response;
- `LAN_GET_SERIAL_NUMBER`, `LAN_GET_CODE`, and `LAN_GET_HWINFO` responses must match the requested
  header and their documented payload sizes of 4, 1, and 8 bytes.

RailKeeper parses every complete dataset in a bounded datagram, skips unrelated asynchronous datasets,
and accepts only the matching response. Oversized, truncated, malformed, wrong-length, or repeatedly
unrelated replies produce an explicit diagnostic instead of a successful connection result.

`backend/internal/application/testdata/z21_protocol_v1_13.json` is a protocol-shape fixture derived from
the official Z21 specification. It is deliberately marked `hardwareCapture: false` and must not be cited
as evidence for an Intellibox 3 firmware version.

## Real hardware validation record

Status on 2026-08-28: **not executed**. No Intellibox 3 was accessible from the development or CI
environment. Therefore this change does not claim that a real device or firmware version has passed.

| Required evidence | Current value |
| --- | --- |
| Hardware model and article number | Intellibox 3 is the intended target; no test unit was available. |
| Firmware version | Not recorded because no device was available. |
| Network mode and configured port | Not recorded on hardware. The software default is Z21-compatible UDP port 21105. |
| Connection and three diagnostic replies | Covered by protocol fixtures and loopback tests only. |
| Anonymized device reply fixtures | None. Add only after capture from an identified hardware and firmware combination. |
| Device-specific protocol deviations | Unknown. Do not infer deviations from generic Z21 fixtures. |

Before real-device validation can be marked complete, record the hardware model, firmware version,
network mode, configured port, result of each of the four operations, and anonymized raw reply datasets.
Any deviation must become a named regression fixture with the expected safe diagnostic. Captures must not
contain local IP addresses, credentials, personal data, or unrelated broadcast traffic.

## LocoNet over TCP boundary

Uhlenbrock documents that Intellibox 3 provides a LocoNet-over-TCP server and supports the Z21 network
protocol. The public product information does not provide enough transport detail to safely implement the
LocoNet TCP port, framing, session behavior, or locomotive data model. Those values must not be guessed.
Implementation remains blocked until authoritative protocol documentation or reproducible device evidence
defines these boundaries. A future adapter must remain separate from Z21 UDP and start read-only with the
existing comparison and conflict workflow.

## Primary sources

- [Uhlenbrock Intellibox 3 product information](https://www.uhlenbrock.de/de_DE/produkte/digizen/I000C67D-001.htm!ArcEntryInfo=0004.0.I000C67D)
- [Uhlenbrock Intellibox 3 quick guide](https://www.uhlenbrock.de/de_DE/service/download/handbook/de/Intellibox%203%20Kurzanleitung.pdf)
- [Z21 LAN protocol specification 1.13](https://www.z21.eu/media/Kwc_Basic_DownloadTag_Component/root-en-main_47-1652-959-downloadTag-download/default/d559b9cf/1646977702/z21-lan-protokoll-en.pdf)
