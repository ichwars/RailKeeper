---
title: Entries and printing
description: Record exhibition locomotives, resolve address conflicts, and print the operating list.
audience: user
status: stable
reviewedVersion: 0.1.20
lastReviewed: 2026-08-16
---

# Entries and printing

An exhibition entry describes one locomotive as it is needed during the selected event. The entry
combines owner and model identity with operating days, digital or analog control data, one image,
function keys, and notes.

## Entry rights and persistence

Messe and Admin can create and edit entries while the selected list is open. Only Admin sees the
entry **Delete** action. A locked list rejects all three mutations, including an Admin deletion.

The entry dialog is one draft with three tabs: **General**, **Vehicle image**, and **Function keys**.
Changes on any tab remain browser form state until **Save** succeeds. Saving sends the complete
entry, closes the dialog, and reloads the selected list's entries. **Cancel** or the close button
discards the current dialog changes without a separate warning.

Saved entries, inline image data, and function settings are local RailKeeper application data and
belong to the app backup. Unsaved form changes and print options do not.

## Create or edit an entry

Select an open list and use the add-entry control in the entry panel. **Create entry** opens on the
**General** tab. Complete the two required fields:

| Field | Rule |
| --- | --- |
| Owner | Required. Leading and trailing whitespace is removed on the server. |
| Locomotive name | Required. Leading and trailing whitespace is removed on the server. |

Users who can read inventory may select an optional **Inventory vehicle**. RailKeeper copies its
model and decoder fields into the exhibition draft and stores the vehicle link. The event fields
remain independent and never modify vehicle master data. Pure Messe accounts can use **Guest
vehicle / manual entry** but cannot browse general inventory.

Use **Edit** on an existing row to open **Edit entry** with its current values.

Select **Save** only after checking all three tabs. The button is disabled while saving or while a
the request is running. A list that became locked after the dialog opened is still rejected by the
server.

## Complete general and control data

The **General** tab contains three groups.

### Basic data

| Field | Input and meaning |
| --- | --- |
| Owner | Required free text for the person or club responsible for the locomotive. |
| Locomotive name | Required free text shown as the main model identity. |
| Manufacturer | Selection from active manufacturer master data when the account may read it. |
| Class | Optional free text. |
| Type | Selection from active vehicle-type master data when available. |
| Epoch | Selection from active epoch master data when available. |
| Railway company | Selection from active railway-company master data when available. |

Admin can load these general master-data choices. A Messe account combined with Viewer, Editor, or
Planner can also load them. Pure Messe receives only **No selection** in those four selectors. The
selectors do not accept arbitrary text. When a stored choice cannot be loaded, the control can show
**No selection** even though the entry draft still carries its previous value. Do not deliberately
choose **No selection** and save if that stored classification must be retained; use Admin or a
Messe account with an inventory-capable role to verify it first.

### Control

| Field | Input and meaning |
| --- | --- |
| Decoder type | Optional free text. |
| Adapter / interface | No selection, NEM 651, NEM 652, PluX16, PluX22, MTC21, Next18, 8-polig, or 21-polig. |
| DCC address | Optional free text checked against other loaded entries in this list. |
| SX address | Optional free text checked separately against other loaded entries. |
| Analog | Switch indicating that analog operation is available. |

The table and report combine DCC and SX into **Address**, show analog as Yes or No, and list decoder
type and interface separately. Empty values appear as a dash.

**Class** is saved with the entry but is not shown in the stable table, read-only list view, or
printed report. Reopen **Edit entry** when that value must be checked.

### No. / lettering / features

The **Notes** text area stores free-form operating information. The server removes leading and
trailing whitespace from all text fields when saving but preserves their internal content.

## Select exhibition days

The day selector offers **All days** and **Day 1** through **Day 4**.

- Select **All days** to replace any individual selection.
- Selecting one numbered day while **All days** is active changes the entry to that day only.
- Select more numbered days to build a multi-day scope.
- No numbered day or all four numbered days normalize to **All days** when saved.
- Stored numbered days are normalized to the order Day 1, Day 2, Day 3, Day 4.

The selected scope appears below the owner in tables and reports. Printing does not filter entries
to one day; it prints every entry in the report and displays each entry's scope.

## Resolve operational conflicts

The server evaluates every entry in the selected list and reports missing required data, the same
inventory vehicle used more than once, and overlapping digital addresses. An address conflict
requires the same interface and address on overlapping event days. Analog or unavailable entries
do not produce an address conflict.

Saving an entry refreshes the workspace and its conflict count. Open **Check conflicts** to inspect
the affected records. Correct the entry whenever possible. If the overlap is deliberate, enter a
reason and save a documented exception. The conflict remains visible as an exception. Locking a
list with unresolved conflicts requires explicit confirmation and a reason for every exception.

## Add or remove the image

The **Vehicle image** tab stores one embedded image per entry.

| Action | Stable behavior |
| --- | --- |
| **Choose image** | Accepts PNG, JPEG, or WebP up to 10 MB and reads it into the entry draft as inline data. |
| Replace image | Choosing another supported file replaces the draft image. |
| **Remove image** | Clears the draft source. It is permanent only after the entry save succeeds. |

If the browser cannot read a selected local image, the draft source is not replaced. Choose a
supported readable file and verify the preview before selecting **Save**.

## Configure function keys F0 through F31

The **Function keys** tab contains F0 through F31. Each row can hold:

- an optional function name;
- an optional function symbol;
- one type: `standard`, `licht`, `sound`, `kupplung`, `rauch`, or `sonderfunktion`.

For a new entry, F0 starts as **Fahrlicht**, type `licht`, with the light symbol. Other keys start
unconfigured with type `standard`. A row becomes configured when it has a name, a symbol, or a type
different from its default. The summary counts configured, `sound`, and `licht` rows.

Selecting a symbol applies the symbol label as the function name when the picker supplies one.
Review the name after choosing a symbol. On **Save**, RailKeeper stores only configured functions as
structured data. The table, detail view, and report show only those functions.

Stable v0.1.20 can also read earlier plain-text values separated by commas, semicolons, or lines
when each item starts with a key such as `F1 Sound`. Opening and saving such an entry writes the
currently configured structured representation.

## Read, sort, and delete entries

The entry table starts with **Owner** ascending and offers these sortable headings:

- **Owner**;
- **Locomotive name**;
- **Control**, sorted by DCC address;
- **Function keys**, sorted by the stored function representation.

The image column is not sortable. The locomotive cell combines locomotive name, railway company,
manufacturer, type, epoch, and notes when present. The control cell shows address, analog state,
decoder, and interface.

Admin can select **Delete** on an entry in an open list and confirm:

> Really delete entry "BR 218"?

The delete is immediate and permanent, then RailKeeper reloads the entries and count. It does not
delete a general vehicle. Messe does not receive the delete control. A locked list rejects the
server request even for Admin.

## View and print the report

**View** opens a read-only table for one list. It shows image, owner and days, locomotive data,
control data, and configured function keys. **Print** from that dialog continues to **Print report**.

You can also use **Print** on a list row or **Print list** above the selected entry table. If the
action belongs to another list, RailKeeper loads that list's entries first. The selected-list action
uses the entries currently displayed.

In **Print report**, keep **Print with images** selected or clear it to omit the image column. The
generated report uses A4 landscape and contains:

- RailKeeper mark, list designation, date, and entry count;
- optional image;
- owner and day scope;
- locomotive identity and available manufacturer, type, and epoch;
- addresses, analog state, decoder, and interface;
- configured function keys and symbols;
- notes.

Select **Print** to open the browser's print dialog. Paper selection, margins, scaling, printer
availability, cancellation, and the final output remain browser and operating-system decisions.
An empty list produces a report row reading **No entries.**

## Resolve entry and print errors

| Situation | Stable result and recovery |
| --- | --- |
| Missing owner or locomotive name | Browser or server rejects the save. Complete both required fields. |
| Conflict count increases after saving | Open **Check conflicts**, correct the records, or document a justified exception. |
| List became locked | Save or delete is rejected. Reload the status; Admin must unlock before a correction. |
| Master-data selectors contain only **No selection** | The account cannot read general inventory master data. Avoid clearing stored choices. |
| Image has no preview | Choose a readable PNG, JPEG, or WebP up to 10 MB before saving. |
| Save succeeds but reload fails | Reopen the workspace and inspect the stored entry before retrying. |
| Detail or report does not open | Reload the list and entries, then retry the read-only action. |
| Browser print dialog closes without output | Reopen **Print report**; no RailKeeper data was changed. |

## Related pages

- [Exhibition workspace](./)
- [Lists and locking](./lists-and-locking)
- [User Guide overview](/guide/)
- [Vehicle inventory and core records](/guide/vehicles/)

## Documented RailKeeper version

This page describes stable RailKeeper v0.1.20. Development behavior on `main` may differ and is
not part of this user workflow.
