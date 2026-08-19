---
title: Exhibition workspace
description: Prepare, maintain, review, and print exhibition lists safely.
audience: user
status: stable
reviewedVersion: 0.1.19.2
lastReviewed: 2026-08-16
---

# Exhibition workspace

The **Exhibition list** workspace keeps the information needed for an exhibition or running day in
one operational view. Each list has a date, an open or locked state, and its own locomotive
entries. Operators can maintain an open list without entering the general vehicle inventory.

This chapter documents stable RailKeeper v0.1.19.2. It does not explain user administration,
master-data administration, backup operation, or the layout workspace.

## Access rights and isolation

RailKeeper exposes the workspace to two roles with different purposes:

| Role | Stable access |
| --- | --- |
| Messe | Opens the isolated exhibition workspace. It can read and print lists and maintain entries while a list is open. |
| Admin | Opens the workspace and additionally creates, edits, locks, unlocks, and deletes lists. Admin can also delete entries from an open list. |

Editor, Viewer, and Planner do not open the exhibition workspace on their own. When one of these
roles is combined with Messe, the entry form can also load general inventory master-data choices.
A pure Messe account remains isolated from those choices, but it can still use the exhibition
fields provided by the dialog. Function symbols remain available to Messe.

The server checks the same distinction. Admin is accepted for every protected exhibition request;
Messe is accepted for reading lists and for creating or editing entries. An unavailable control is
therefore not the only protection against an unauthorized write.

## Understand lists and entries

The page contains two related tables:

| Area | Purpose |
| --- | --- |
| Lists | Shows **Designation**, **Date**, number of **Entries**, **Status**, and available actions. |
| Entries | Shows the image, owner and operating days, locomotive data, control data, configured function keys, and actions for the selected list. |

Selecting a list row loads its entries in the adjacent table. After the initial list request,
RailKeeper selects the first returned list when no selection exists. Stable storage order is newest
date first and then designation; the browser also starts with **Date** descending.

Select a sortable list heading to sort by designation, date, entry count, or status. Select the
active heading again to reverse it. Entry sorting starts with **Owner** ascending. Its detailed
columns and rules are covered in [Entries and printing](./entries-and-printing).

## Follow the exhibition workflow

Use this sequence for predictable operation:

1. An Admin creates a list with a designation and date.
2. Messe or Admin selects the open list.
3. Operators add entries and complete owner, locomotive, operating, image, and function data.
4. Address conflicts are resolved before an entry is saved.
5. Operators use **View** or **Print list** to review the current entries.
6. An Admin locks the list when entry maintenance should stop.
7. The locked list remains available for viewing and printing. An Admin can unlock it for a
   deliberate correction.

List actions and entry actions are separate server writes. Closing a dialog without submitting it
does not save its form. After creating, editing, or deleting a list, RailKeeper reloads the list
table; locking or unlocking updates the affected row directly. After a successful entry write, it
reloads the selected list's entries and updates the displayed count.

## Work with an open or locked list

The **Status** column shows **open** or **locked**.

| Operation | Open list | Locked list |
| --- | --- | --- |
| Select and read | Yes | Yes |
| View complete list | Yes | Yes |
| Print | Yes | Yes |
| Create or edit an entry | Yes | No |
| Delete an entry as Admin | Yes | No |
| Edit list fields as Admin | Yes | Yes |
| Lock or unlock as Admin | Lock | Unlock |
| Delete list as Admin | Yes | Yes |

The entry add and edit controls are disabled for a locked list. The server also rejects entry
changes if the list became locked after the page loaded. Unlock the list before correcting an
entry; do not interpret a disabled control as missing data.

Every list provides **View** and **Print** actions. Admin additionally sees **Edit**, **Lock** or
**Unlock**, and **Delete**. The entry panel also provides **Print list** and, for an open selection,
the add-entry control.

## Read loading, empty, and error states

| State | Meaning and next action |
| --- | --- |
| **Loading...** | The list request is still running. Wait before choosing a list. |
| **No exhibition list created yet.** | No list exists. An Admin must create the first list. |
| **Please select a list.** / **No list selected.** | Select a row in the list table before working with entries. |
| **No entries in this list yet.** | The selected list is empty. Add an entry while it is open. |
| Message above the tables | A list, entry, symbol, or master-data request failed. Read the message before continuing. |

Some follow-up requests do not clear previously displayed data before reporting an error. If a
list or entry refresh fails, reload the workspace and verify the selected list, its status, entry
count, and stored entries before retrying a write. This avoids repeating an operation that may
already have succeeded before the refresh failed.

## Continue with a focused workflow

- [Lists and locking](./lists-and-locking) explains Admin preparation, lock effects, and deletion.
- [Entries and printing](./entries-and-printing) explains every entry field, address conflicts,
  images, function keys, viewing, and report printing.

## Related pages

- [User Guide overview](/guide/)
- [Lists and locking](./lists-and-locking)
- [Entries and printing](./entries-and-printing)
- [Vehicle inventory and core records](/guide/vehicles/)

## Documented RailKeeper version

This page describes stable RailKeeper v0.1.19.2. Development behavior on `main` may differ and is
not part of this user workflow.
