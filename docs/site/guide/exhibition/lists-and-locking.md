---
title: Lists and locking
description: Create exhibition lists, control their lock state, and remove them safely.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Lists and locking

An exhibition list groups the locomotive entries for one named exhibition or running date. Admin
prepares and controls lists; Messe concentrates on operating entries inside an open list.

## Roles for operation and administration

| Action | Messe | Admin |
| --- | --- | --- |
| List, select, view, and print lists | Yes | Yes |
| Create or edit a list | No | Yes |
| Lock or unlock a list | No | Yes |
| Delete a list | No | Yes |
| Maintain entries in an open list | Yes | Yes |
| Delete an entry from an open list | No | Yes |

Admin receives direct access to the exhibition workspace. It does not need an additional Messe
role. Editor, Viewer, and Planner neither expose the workspace nor grant list management by
themselves.

## Create a list

Select **New list** and complete the **Create exhibition list** dialog:

| Field | Rule |
| --- | --- |
| Designation | Required. RailKeeper removes leading and trailing whitespace when saving. |
| Date | Required. A new dialog starts with the browser-generated UTC calendar date. Verify it near local midnight. |

Select **Save**. A successful request creates an open list, selects it, closes the dialog, and
reloads the list table. The create operation is an immediate server write; the dialog values are
only local form state until it succeeds.

If either required field is empty, the browser prevents normal submission. The server also rejects
blank values after trimming them. In stable v0.1.17.6 the server checks that the date value is not
empty; the date control supplies the normal calendar value.

## Select, sort, and inspect lists

Select any row to make it the active list. The entry table then loads that list's entries. The
selected row remains highlighted.

The list table starts with **Date** descending. Select these headings to sort in the browser:

- **Designation**;
- **Date**;
- **Entries**;
- **Status**.

Selecting a different heading starts ascending. Selecting the active ascending heading changes it
to descending; selecting it once more returns to ascending.

The **View** action loads the list's current entries and opens a read-only table. It is available to
Messe and Admin for open and locked lists. **Print** follows the same read path and opens the report
options described in [Entries and printing](./entries-and-printing).

## Edit designation or date

Admin selects **Edit** on the required row. **Edit exhibition list** opens with the stored
designation and date. Change either field and select **Save**.

Editing a list does not change its entries or lock state. It is allowed while the list is open or
locked. After a successful write, RailKeeper keeps the saved list selected and reloads all lists.
If saving fails, the dialog remains open and the message above the workspace reports the error.
Correct the values or reload before trying again.

## Lock or unlock a list

Select **Lock** when no further entry changes should be accepted. RailKeeper immediately writes the
new state and replaces **open** with **locked** in the table. There is no confirmation dialog.

A lock has this exact boundary:

- selection, reading, **View**, and **Print** continue to work;
- list designation and date can still be edited by Admin;
- a list can still be deleted by Admin;
- entry creation, entry editing, and entry deletion are rejected;
- both disabled controls and server validation enforce the entry restriction.

Select **Unlock** to restore entry maintenance. Unlocking is also immediate and has no confirmation.
If a lock request fails or the displayed state is uncertain, reload the workspace and read the
**Status** column before changing an entry.

## Delete a list

Admin selects **Delete** and confirms:

> Really delete exhibition list "Dortmund 2026"?

Deletion is permanent. Removing a list also removes all of its exhibition entries through the
database relationship. It does not delete general vehicle records. Export a current validated app
backup before deleting a populated list when recovery may be required.

After a successful delete, RailKeeper clears the selection when it referred to that list and
reloads the remaining lists. Cancel the confirmation to leave the list unchanged.

## Resolve list errors

| Situation | Stable result and recovery |
| --- | --- |
| Empty designation or date | Save is rejected. Complete both required fields. |
| List no longer exists | The server returns not found. Reload the workspace and select an existing list. |
| Missing Admin authority | Create, edit, lock, unlock, and delete requests are forbidden. Use an authorized Admin account. |
| Save request fails | The dialog stays open and a message appears. Verify the values and stored state before retrying. |
| Lock or unlock appears unchanged | Reload before entry maintenance; the server state is authoritative. |
| Delete result is uncertain | Reload and verify whether the list still exists before selecting **Delete** again. |
| Entry operation reports a locked list | Reload the status. Only Admin can deliberately unlock it. |

Successful create, edit, lock, unlock, and delete operations are recorded in RailKeeper's audit
log. Audit-log operation belongs to administration and is not documented on this user page.

## Related pages

- [Exhibition workspace](./)
- [Entries and printing](./entries-and-printing)
- [User Guide overview](/guide/)

## Documented RailKeeper version

This page describes stable RailKeeper v0.1.17.6. Development behavior on `main` may differ and is
not part of this user workflow.
