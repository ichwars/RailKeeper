---
title: ECoS locomotive sync
description: Read, review, import, and explicitly write selected ESU ECoS locomotive data.
audience: user
status: stable
reviewedVersion: 0.1.20.2
lastReviewed: 2026-08-16
---

# ECoS locomotive sync

The ECoS workspace reads locomotive master data, static function descriptions, and exposed CV
values from an active ESU ECoS. Every locomotive remains in a work list until an Admin opens it in
the vehicle editor, saves it, or deliberately skips it.

All ECoS API routes are Admin-only. Configure and test the command station under **Settings >
Digital command stations** before using this page. Host and port are read-only in the import
workspace.

## Supported and unsupported command stations

| Active provider | Import/Export behavior in v0.1.20.2 |
| --- | --- |
| ESU ECoS | Read workflow, vehicle handoff, CV and static-function suggestions, and reviewed write sync are available. |
| Z21, Intellibox 3, or Märklin CS3 | RailKeeper shows that the import path is prepared but not implemented. |
| No active provider | The page shows a link to the Digital command-station settings. |

RailKeeper does not monitor speed, direction, active function states, locomotive images, or layout
objects in this workflow. Runtime attributes such as speed, direction, and function state are
excluded from the review. Only static function descriptions and CV values exposed by the ECoS
locomotive object are candidates.

## Read the locomotive work list

With an enabled ECoS and stored host:

1. Open **Import/Export**.
2. Check the live-monitor state and stored host and port.
3. Choose **Fetch data**.
4. Wait for the locomotive count and work list.
5. Review the ECoS name, object ID, address, detected RailKeeper match, decoder hint, CV count, and
   current work-list status.

RailKeeper polls the ECoS live status every 15 seconds. When the live monitor is already connected,
the page automatically fetches once if no probe or import session is active.

A successful fetch creates a work session in the browser tab's session storage. Returning from the
vehicle editor restores the list and its **open**, **editing**, **saved**, or **skipped** states.
Closing the browser tab ends that browser-side persistence.

## How an existing vehicle is suggested

RailKeeper checks candidates in this order:

1. an existing external ECoS mapping with the same ECoS object ID;
2. the same decoder address in **Digital / decoder no.**;
3. a normalized match against vehicle designation or vehicle number.

Name comparison ignores case, umlaut spelling, punctuation, and a leading `BR` or `V` before a
number. For longer names it also accepts containment. These matches are suggestions, not identity
proof. Verify the inventory number and vehicle before updating or writing to the ECoS.

## Review CV and function suggestions

The work list and CV preview show only values returned by the ECoS. Standard CV definitions add a
meaning, category, and interpretation where possible. CV8 can identify the decoder manufacturer
from RailKeeper master data; an absent or unknown CV8 remains visibly unresolved. Missing CVs do
not block the import.

Static ECoS function descriptions become F-key suggestions and are matched to RailKeeper symbols
where possible. Active on/off state is never imported. Review the function and CV previews in the
vehicle editor before saving.

## Create or update one vehicle

Choose **Create vehicle** for an unmatched locomotive or **Update match** for a suggestion. The
generic CSV/XML/JSON review is not used. RailKeeper opens the vehicle editor directly and carries:

- ECoS object ID, name, address, protocol, and profile as source context;
- name, category, digital state, decoder address, decoder type, and description as vehicle-field
  suggestions;
- an external ECoS mapping;
- static function suggestions;
- CV-value suggestions.

For a new vehicle, Category defaults to the stored German value `Lokomotive`, even in the English
interface. Manufacturer, Designation, Gauge, Category, and Subtype remain required. Fields that the
ECoS could not supply are highlighted and must be resolved before saving.

Saving runs several writes in sequence: vehicle core data first, then the ECoS mapping, CV values,
and configured functions. Only after the full sequence succeeds does the work-list row become
**saved** and RailKeeper return to the ECoS session. If a later step fails, the core vehicle may
already exist or be updated. Inspect it before retrying to avoid duplicates or repeated values.

Use **Skip** when a locomotive should remain unchanged. **Next open loco** selects the next open
entry, then an entry still marked editing when no open one remains.

## Write reviewed RailKeeper data to the ECoS

Write sync is available only for a locomotive with a detected RailKeeper match:

1. Verify that the suggested vehicle is correct.
2. Choose **Check sync**. This is a dry run and does not write to the command station.
3. Review up to three displayed differences between the current ECoS value and the desired
   RailKeeper value.
4. If changes exist, choose **Write to ECoS**.
5. Confirm the browser prompt for that locomotive.

The stable write scope is deliberately narrow: Name, Address, and Protocol only. Empty desired
values are not written. RailKeeper sends the combined ECoS change only after explicit confirmation,
then refreshes the RailKeeper vehicle and marks its external mapping as synchronized when possible.

Writing does not send CVs, function definitions, speed, direction, active functions, images, or
layout objects to the command station.

## Troubleshooting

| Symptom | Check |
| --- | --- |
| The ECoS workspace is absent or disabled | Confirm an Admin enabled ECoS and stored a non-empty host under Digital command stations. |
| Another provider is active | v0.1.20.2 only implements this import path for ECoS. |
| Fetching fails | Check network reachability, stored host and port, ECoS availability, and Admin role. |
| A wrong vehicle is suggested | Do not update or sync it. Correct the external mapping or decoder/name data through the vehicle workflow first. |
| A CV or function is missing | The ECoS locomotive object did not expose it, or RailKeeper did not recognize it as a static supported value. Missing CVs do not block saving. |
| The work list returned after saving but the row is not saved | The complete vehicle, mapping, CV, and function sequence did not finish. Inspect the vehicle and error before retrying. |
| **Write to ECoS** is not visible | Run **Check sync** first. The write button appears only for a match with a non-empty, unapplied change plan. |

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.20.2** and was last reviewed on 2026-08-16.
