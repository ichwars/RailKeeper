---
title: Events, status and locking
description: Events, status and locking.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Events, status and locking

**Goal:** Prepare an event and deliberately control whether its planning can be edited.

<GuideFigure name="exhibition-locked" alt="Event workspace showing a locked demo list." caption="The actual status change is saved. Printing and status-appropriate actions remain visible." />

## Roles for operation and administration

Admin and accounts combining **Editor plus Messe** manage planning and status.
Messe maintains existing entries during open planning. Event and entry deletion remain Admin-only.
[Access overview](./#access-rights-and-isolation).

## Create a list

Select **New event**, fill name, start and end, then optional location, description and internal
organisation. **Save** creates and selects it. Check dates before assigning entry days.
**Cancel** discards the draft.

## Select, sort, and inspect lists

Choose an event on the left. **Show all events** includes archived events.
The current UI uses event buttons rather than the former sortable list table.
Dates and status appear in the event header.

## Edit designation or date

Choose **More actions > Edit event** while planning is editable.
Saving checks the expected revision. An intervening change may reject your update; reload and compare.
Date changes can affect existing operating-day assignments, so review entries afterwards.

## Lock or unlock a list

Select **Lock list** in the event header. Without conflicts, the status change saves directly.
Existing conflicts open a confirmation dialog.

<GuideFigure name="exhibition-lock-conflicts" alt="Conflict-aware locking dialog requiring an explicit reason." caption="Negative test: conflicts do not lead to silent locking. The exercise first cancels and fills the missing station/interface." />

Prefer **Cancel**, correct the entry and check again. A deliberate exception requires a reason and
**Lock anyway**. Problems remain visible; an exception does not make the entry fully ready.
**Unlock list** reopens planning. **More actions** offers start, completion, reopening and archiving
where permitted by the current status.

## Delete a list

As Admin, choose **More actions > Delete** and read the confirmation for that event.
Its operating entries belong to it. Preserve required data beforehand.
Archiving is offered separately and suits later reference.

## Resolve list errors

| Message or state | Next step |
| --- | --- |
| Missing required details | Check name and valid dates |
| Stale revision | Reload and compare your changes with current data |
| Conflicts during locking | Review and correct, or deliberately record a justified exception |
| Missing action | Check roles and current status |
| Apparently missing event | Show archived events and check selection |

## Related pages

[Entries and printing](entries-and-printing) · [Illustrated exercise](practice).
