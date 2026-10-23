---
title: Organize exhibitions and operating days
description: Organize exhibitions and operating days.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Organize exhibitions and operating days

Open **Exhibition** and choose an event on the left. Dates and status appear at the top;
manage participants and vehicles on the right. [The illustrated exercise](practice) covers a complete workflow.

<GuideFigure name="exhibition-open" alt="Current event workspace with metrics, day filters and vehicles." caption="Actual UI with synthetic data. A remaining check is shown honestly rather than presented as ready." />

## Access rights and isolation

| Account | Current scope |
| --- | --- |
| Admin | Workspace, event planning, status changes and deletion |
| Messe | Use existing events; maintain entries during open planning; review and print |
| Editor together with Messe | Additionally manage events and status |
| Viewer together with Messe | The Messe workflows above; Viewer alone cannot open this workspace |

**Editor** alone has no Exhibition navigation access. Add **Messe** when the same person also
needs this task. Deletion remains Admin-only.
Operating entries do not change vehicle master data. A vehicle's **Exhibition** switch does not create an entry.

## Understand lists and entries

An event has a name, start, end, location and organisational details.
Its entries describe owner, locomotive, station, address, operating days and availability.
Optionally assign an inventory vehicle or enter a guest vehicle manually.

Metrics count entries, owners, conflicts and ready vehicles. **Readiness** separates address checks,
function keys and images. Ready does not mean RailKeeper has physically driven or controlled a locomotive.

## Follow the exhibition workflow

1. Create a **New event** and check its dates.
2. Add entries under **Participants and vehicles**.
3. Select **All days** or a specific day; filter using **All**, **Ready**, **Check** and search.
4. **Check conflicts** and resolve missing details or overlapping addresses.
5. Lock the list, print its report and later advance event status.

[Events and locking](lists-and-locking) · [Entries and reports](entries-and-printing).

## Work with an open or locked list

| Status | Workflow meaning |
| --- | --- |
| Draft | Open planning before proceeding with event operation |
| Open | Maintain and review entries |
| Locked | Planning frozen; entries cannot be edited, printing remains available |
| Running | Event started; entries remain protected |
| Completed | Operation finished; planning can be reopened |
| Archived | Hidden by default; find through **Show all events** or the action menu |

Entry editing is available only in **Draft** or **Open** without a lock.
**More actions** offers transitions appropriate to the current status. A list lock does not protect hardware.

## Read loading, empty, and error states

Wait for the chosen event to load. Add an entry to an empty event.
For a stale revision, reload and compare before saving.
For **Missing data**, open the entry and conflict display; for **Unavailable**, check participation.
A missing image or function keys can still produce **Check** status.

## Continue with a focused workflow

- [Illustrated operating-day exercise](practice)
- [Planning, status and conflict-aware locking](lists-and-locking)
- [Entries, images, functions and printing](entries-and-printing)

## Related pages

[Vehicle inventory](../vehicles/) · [Accounts and roles](../../administration/users-sessions-security).
