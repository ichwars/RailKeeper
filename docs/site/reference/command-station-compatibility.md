---
title: "Command stations: scope and evidence"
description: "Command stations: scope and evidence."
audience: reference
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Command stations: scope and evidence

This overview separates implemented functions from documented device acceptance.
It describes the reviewed **main development branch on 23 October 2026**; older releases may differ.

## Functions by connection

| Connection | Reading and comparison | Writing to the station | Locomotive images |
| --- | --- | --- | --- |
| ECoS | Locomotive data, static functions, comparison worklist; stored CV1–1024 cache | Confirmed locomotive-data writes; no CV programming | Preparation, download and confirmed upload/assignment |
| CS3 / CS3 Plus | Roster UID, name, address and protocol; no CVs | Unavailable | Prepare PNG and import manually on CS3 |
| Z21 / z21 | Test, diagnostics and bounded address comparison for known RailKeeper addresses | Unavailable | Prepare PNG for the Z21 app; no image storage in the station |
| Intellibox 3 | Test and diagnostics using Z21-compatible UDP | Unavailable | Prepare PNG for IB3 Configuration; no direct RailKeeper upload |

A readable reply or suitable image dimensions do not establish successful operation on every
model. Image export and direct upload are separate capabilities.

## Recorded device evidence

| Device / firmware | Evidence | Remaining boundary |
| --- | --- | --- |
| ECoS, application firmware **4.3.3** | Recorded locomotive-data write acceptance after PR #143; later user-confirmed image export, upload and automatic assignment | Does not establish a full model/hardware matrix; no anonymized raw-response fixture was retained for that earlier acceptance |
| ECoS 4.3.3, busy locomotive control | One authorized protocol probe confirmed error 25 when control was occupied | The revised pre-upload reservation, deduplication and pause/resume order has automated/simulated coverage; additional device acceptance remains open |
| Other ECoS firmware | Complete compatible identification replies can be used as **not device-verified** | No blanket acceptance of all firmware versions |
| CS3, pre-2.6 and 2.6+ API shapes | Parser/regression tests based on modeled compatible response forms | No direct complete RailKeeper hardware acceptance recorded |
| Z21 / z21 | Protocol-based UDP fixtures and automated tests | No recorded physical model/firmware acceptance of these read workflows |
| Intellibox 3 | Automated Z21-compatible transport tests | Physical Intellibox acceptance pending; LocoNet over TCP is planned, not implemented |

The image dialog's **ECoS 50200/50210/50220** labels identify the intended output format.
They do not establish three separate hardware acceptances.
New handbook images show an isolated demo. Simulated station replies are explicitly labelled
in the [image guide](../guide/vehicles/locomotive-images).

## What does connection status mean?

| ECoS result | Meaning and next step |
| --- | --- |
| Unreachable or rejected | Check address, server reachability, port and error |
| Partial | Required identity fields or end marker are missing; do not activate |
| Not device-verified | Complete reply, but no recorded acceptance for this firmware |
| Device-verified | Firmware matches the recorded acceptance; edge cases retain their own evidence limits |

Ping or web access does not replace the protocol test. ECoS needs TCP **15471** and additionally
HTTP **80** for images. CS3 uses its configured HTTP port; Z21 and Intellibox 3 default to UDP **21105**.

## Contribute another device acceptance

Record model, hardware revision, firmware, RailKeeper version, transport and exact workflow.
Record expected and actual results, including interruptions and conflicts.
Keep captures limited to necessary fields and remove personal information.
Device writes need explicit authorization; connection checks begin read-only.

## Sources and remaining work

- [ECoS compatibility](https://github.com/ichwars/RailKeeper/blob/main/docs/architecture/ecos-compatibility.md)
- [Image transfer and recorded protocol evidence](https://github.com/ichwars/RailKeeper/blob/main/docs/architecture/ecos-image-transfer.md)
- [CS3 reply forms and limits](../guide/import-export/cs3-import)
- [Z21 reading scope](https://github.com/ichwars/RailKeeper/blob/main/docs/architecture/z21-read-scope.md)
- [Intellibox transport boundaries](https://github.com/ichwars/RailKeeper/blob/main/docs/architecture/intellibox3-transports.md)
- Open integration work: [ECoS #132](https://github.com/ichwars/RailKeeper/issues/132),
  [Z21 #133](https://github.com/ichwars/RailKeeper/issues/133), [Intellibox 3 #134](https://github.com/ichwars/RailKeeper/issues/134).
