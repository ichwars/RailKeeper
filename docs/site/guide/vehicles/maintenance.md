---
title: Vehicle maintenance and condition
description: Record, schedule, complete, review, and safely remove vehicle maintenance entries.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Vehicle maintenance and condition

The **Maintenance** tab records work, repairs, conversions, condition, dates, and costs for one
vehicle. Admin, Editor, Viewer, and Planner can inspect stored entries. Only Admin and Editor can
create, change, complete, or delete them, and the server enforces that boundary.

## Open the Maintenance tab

Open a vehicle from **Vehicle inventory**, choose **Edit**, then select **Maintenance**. The vehicle
must have been saved once because maintenance needs its stored vehicle ID. A new unsaved vehicle
shows **Maintenance entries can be added after the first save.** Save the core record before adding
an entry.

## Read the maintenance summary

The top of the tab shows three counters:

- **Due** counts entries with a due date on or before the current local date whose status is not
  `erledigt`.
- **Planned/open** counts every entry whose status is not `erledigt`, including entries already
  counted as due.
- **Done** counts entries whose status is `erledigt`.

**Due** is therefore a subset of **Planned/open**, not a separate status total. A `geplant` entry
whose date is today or earlier is displayed and counted as due even though its stored status is not
`faellig`.

RailKeeper lists open entries before done entries. Within each group, dated entries come before
undated entries, earlier due dates come first, and equal entries use newest creation first.

## Add a maintenance entry

Every maintenance write, **Add entry**, **Save entry**, **Done**, or **Delete maintenance**, is an
immediate request that does not wait for the vehicle's **Save changes** action. After success,
RailKeeper reloads the complete vehicle, replaces the core form, and reloads the feature-specific
editor state. This can discard unsaved core fields, image metadata, attachment edits, a partially
entered maintenance form, and pending changes on other tabs. Save or intentionally discard all
pending vehicle changes before any maintenance write.

Enter the maintenance data and choose **Add entry**. After success, RailKeeper resets the form to
type `Wartung` and status `geplant`.

## Fields, values, and validation

- **Type:** Required. The stored German values are `Wartung`, `Reparatur`, `Umbau`, `Superung`,
  `Reinigung`, `Schmierung`, `Decoder-Einbau`, and `Ersatzteiltausch`. The English labels are
  Maintenance, Repair, Conversion, Detail upgrade, Cleaning, Lubrication, Decoder installation,
  and Spare part replacement. Default: `Wartung`.
- **Status:** Required. Stored values are `geplant`, `faellig`, and `erledigt`; shown in English as
  planned, due, and done. The German spelling `fällig` is normalized to `faellig`. Default:
  `geplant`.
- **Condition:** Optional. Stored values are `neuwertig`, `sehr gut`, `gut`, `gebraucht`, and
  `reparaturbedürftig`. These values are displayed unchanged in the English interface and also
  remain German in stored data.
- **Due on:** Optional valid calendar date in `YYYY-MM-DD` form. It controls date-based due
  calculation.
- **Completed on:** Optional valid calendar date in `YYYY-MM-DD` form. Saving a done entry without
  this date supplies today's local date.
- **Cost:** Optional non-negative decimal amount. Comma and point decimal separators are accepted.
  Surrounding spaces, internal spaces, and one trailing euro sign are removed before validation.
  Accepted numbers are displayed as EUR using the German locale.
- **Notes:** Optional. The server trims the value and accepts at most 4,000 characters.

The server rejects unknown types, statuses, and conditions, invalid dates, negative or non-numeric
costs, and notes above the limit.

## Understand due dates and completion

The stored status `faellig` and date-based due highlighting are related but independent. Selecting
the due status does not add a due date. Conversely, an open entry with a due date on or before today
is highlighted and counted as due even when its stored status remains `geplant`.

An entry stops contributing to **Due** and **Planned/open** only when its status becomes `erledigt`.
The completion date is descriptive; it does not by itself mark the entry done.

## Edit or cancel an entry

Choose **Edit maintenance** on a card to copy its values into the form. The main action becomes
**Save entry**, and **Cancel** appears. Saving updates the entry immediately and reloads the complete
vehicle. **Cancel** resets the maintenance form without changing the stored entry.

## Mark an entry done

Choose **Done** on an open card to complete it immediately. RailKeeper does not ask for separate
confirmation. It preserves the type, condition, due date, cost, notes, and any existing completion
date. If the completion date is empty, RailKeeper uses the current local date.

## Check linked media

When at least one link exists, the maintenance card shows image and attachment counts below
**Linked media**. Manage image links in **Uploads**. Selecting **No maintenance** for an image is
pending metadata until you use the vehicle's **Save changes** action.

Stable v0.1.17.6 can display attachments that already contain a maintenance reference, but the
normal attachment row cannot assign or change that reference. Do not invent a reassignment flow
that the interface does not provide.

## Delete maintenance safely

**Delete maintenance** removes the entry immediately without a confirmation dialog. The backend
does not block deletion for linked media and does not delete or detach those media records. Their
stored maintenance ID can therefore point to a deleted entry.

Before deleting an entry:

1. Save or intentionally discard every other pending vehicle change.
2. Check its linked image and attachment counts.
3. In **Uploads**, select **No maintenance** for every linked image and use **Save changes**.
4. If a linked attachment must retain the association, keep the maintenance entry. Stable
   v0.1.17.6 has no attachment-link editor. Remove an attachment only after backup and content
   review when the file is genuinely no longer needed.
5. Return to **Maintenance**, confirm that **Linked media** no longer appears, then delete the entry.

Deleting maintenance does not delete an image or attachment file. A stale, non-empty maintenance
ID can later prevent image deletion until the link is cleared and saved.

## Roles, storage, and backup boundaries

Admin, Editor, Viewer, and Planner can read maintenance. Server-side create, update, complete, and
delete operations require Admin or Editor. Disabled controls help explain the current mode but do
not replace server authorization.

Maintenance rows and vehicle uploads are local RailKeeper application data and belong to the
application backup scope. Before substantial cleanup, create a current backup and have an Admin
validate it. Use [Vehicle images and attachments](/guide/vehicles/media) for the published media
safety workflow. Backup export, validation, and restore operation belong to administration and are
not repeated here.

## Empty and error states

- **Vehicle is not saved:** Save the core record before adding maintenance.
- **No entry exists:** Use **Add entry** after checking that the correct vehicle is open.
- **Input is rejected:** Check type, status, condition, dates, non-negative cost, and note length.
- **Write action fails:** Keep the form open, read the error, check the session and connection,
  then retry.
- **Summary looks outdated:** Confirm that no unsaved changes remain, then reload the vehicle.
- **Due count differs from status:** Due is calculated from date and completion, not only the
  stored status.
- **Linked media is shown:** Resolve the links before deletion and use the media page for image
  metadata.
- **Entry was deleted without a prompt:** Deletion is immediate. Recovery requires a suitable
  backup.

A failed action does not undo an earlier successful independent maintenance or media action.

## Related pages

- [User Guide overview](/guide/)
- [Overview, metrics, and data quality](/guide/overview/)
- [Vehicle inventory and core records](/guide/vehicles/)
- [Vehicle images and attachments](/guide/vehicles/media)
- [Decoder, functions, and CV data](/guide/vehicles/decoder-cv)
- [Article search, web documents, and spare parts](/guide/vehicles/search-and-spares)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.17.6** and was last reviewed on 2026-08-16.
