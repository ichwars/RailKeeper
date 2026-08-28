---
title: Read CS3 locomotives without writing
description: Safely read Märklin CS3 locomotives and compare them in the Digital centers workspace.
audience: user
status: stable
reviewedVersion: 0.1.20.2
lastReviewed: 2026-08-28
---

# Read CS3 locomotives without writing

RailKeeper reads the locomotive roster of an active Märklin CS3 or CS3 Plus over HTTP into the
existing **Digital centers** workspace. The workflow creates a persistent comparison preview. It
does not write to the CS3 and does not control a locomotive.

All CS3 routes are Admin-only. Host and port come from the server configuration under
**Settings > Digital command stations**. Values sent to the read action cannot override the stored
target. RailKeeper accepts only private IP addresses on the local network, rejects loopback,
link-local, and public targets, and pins resolved host names to the validated IP address for the
complete request.

## Supported CS3 API generations

| CS3 firmware generation | Read-only endpoint | Behavior |
| --- | --- | --- |
| 2.6 and later | `/app/api/locos` | Always checked and used first. |
| before 2.6 | `/app/api/loks` | Used only when the current endpoint returns an explicit HTTP 404. |

Other errors do not trigger a silent fallback. Authentication, redirects, HTML pages, and unknown
responses do not count as a successful CS3 connection.

Märklin describes the CS3 and CS3 Plus as local command stations with a locomotive database but
does not publish a stable contract for these web-app endpoints. The endpoint forms and field names
were derived from publicly documented TrainControl compatibility data. RailKeeper fixtures
`cs3_loks_pre_2_6_anonymized.json` and `cs3_locos_2_6_anonymized.json` reproduce those response
shapes with anonymized values. They contain no private layout or locomotive data. The RailKeeper
project has not directly verified these fixtures against physical hardware.

## Data read and deliberately omitted

RailKeeper accepts only these fields per locomotive:

- `uid` as the stable external CS3 ID;
- `name`, or the decoded `internname` as a fallback;
- `address` as the decoder address;
- `dectyp` as a normalized protocol such as MFX, Motorola, or DCC.

RailKeeper ignores speed, direction, active functions, icons, CVs, and layout objects. It starts no
live monitor and sends no control or write command. A decoder address is only a comparison
attribute. A name alone never creates an automatic match.

## Read and compare safely

1. Configure the CS3 host and HTTP port under **Settings > Digital command stations**.
2. Run **Test connection**. Success requires a compatible and valid JSON locomotive roster.
3. Under **Latest diagnostics**, optionally choose **Read diagnostics**. RailKeeper shows API path,
   generation, HTTP status, content type, count, and the `readOnly` marker.
4. Activate the adapter and open the **Digital centers** workspace.
5. Choose **Read data**. RailKeeper creates a new read session and comparison work list.
6. Review new, missing, deviating, and ambiguous entries. Multiple address candidates remain a
   visible conflict.

The request is limited to HTTP GET, an 8 MiB response, and 5,000 locomotives. Redirects are
rejected. Only JSON content types and fully validated UIDs, names, addresses, and protocols reach
the preview.

## Troubleshooting

| Diagnosis | Check |
| --- | --- |
| Network or timeout error | Check CS3 IP, port, local network, and reachability from the RailKeeper server. |
| Target rejected | Configure a private CS3 address on the local network. Public, loopback, and link-local addresses are not allowed. |
| Authentication error | Check CS3 web-app access protection. RailKeeper does not bypass authentication. |
| Redirect rejected | Configure the direct local CS3 host. RailKeeper does not follow redirects. |
| HTML or non-JSON response | Check that host and port reach the CS3 API rather than a login or proxy page. |
| No supported roster API | Check firmware and web-app availability. Both known endpoints returned 404. |
| Invalid locomotive data | A locomotive UID, name, decoder address, or protocol is outside the safe limits. The complete response is rejected. |
| Locomotive appears as a conflict | Several RailKeeper vehicles match address and protocol. Review the assignment manually. |

## Documented RailKeeper version

This chapter documents RailKeeper **v0.1.20.2** and was last reviewed on 2026-08-28.
