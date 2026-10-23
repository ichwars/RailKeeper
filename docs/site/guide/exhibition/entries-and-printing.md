---
title: Review and print operating entries
description: Review and print operating entries.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Review and print operating entries

**Goal:** Record a locomotive, owner, control information and operating days, then find it in the report.

<GuideFigure name="exhibition-entry" alt="Current entry dialog with General, Vehicle image and Function keys." caption="The actual exercise fills a missing station/interface. Operating entries do not change vehicle master data." />

## Entry rights and persistence

Admin, Messe and appropriate combined accounts can edit entries in unlocked **Draft** or **Open**
planning. Only Admin deletes. Dialog changes remain drafts until **Save**; **Cancel** discards them.
Server checks and revisions prevent unchecked overwriting.

## Create or edit an entry

Choose **Add entry** under **Participants and vehicles**, or the edit icon on a locomotive row.
The dialog contains **General**, **Vehicle image** and **Function keys**.
An optional inventory-vehicle picker is available to accounts with inventory read access.
Selection fills known information; particularly review address, station and operating days.

## Complete general and control data

### Basic data

Fill owner, locomotive name, manufacturer, series, category, epoch, railway company and availability.
A guest vehicle need not be linked to inventory. **Unavailable** records participation status rather than deleting it.

### Control

Decoder type and adapter describe the locomotive. **DCC address** and **SX address** are separate.
**Command station / interface** identifies the intended control context, not the decoder connector.
Digital entries require an appropriate address and this context. **Analogue** identifies analogue operation.

### No. / lettering / features

Use notes for operating information. They do not change individual vehicle master data.

## Select exhibition days

Choose **All days** or intended event days. Workspace day filters change the view rather than saved
participation. Review assignments after changing event dates.

## Resolve operational conflicts

**Check conflicts** identifies missing details, duplicate vehicles and overlapping addresses on the
same control context and days. Correct the affected entries.
Documented exceptions remain visible. Missing images or function keys can produce **Check** even
without address conflicts. **Ready** is a data check, not a test drive.

## Add or remove the image

Open **Vehicle image**, select a suitable file or remove the current choice.
**Save** persists the operating-entry change. Review the image before printing.
This action does not automatically replace an inventory vehicle's image.

## Configure function keys F0 through F31

Open **Function keys**, add a function and choose key, symbol and name. Remove unwanted rows.
The list saves together with the operating entry. It is separate from decoder CVs and the regular
vehicle editor's Control list.

## Read, sort, and delete entries

Use day filters, **All / Ready / Check** and **Search entries**. The current table has no former
per-column sort buttons. Deletion requires Admin and explicit confirmation; protected planning hides the action.

## Print the exhibition list

Select **Print** in the event header. The current workspace generates the report directly and opens
the browser print workflow; it has no preceding image-choice dialog.
Check the event name, intended entries, addresses and functions.
The report includes the whole event entry set rather than just the active table filter.

<GuideFigure name="exhibition-print-report" alt="Actually generated report of the demo event." caption="The generated report HTML was inspected in a browser. This does not establish printer-driver or native PDF-output acceptance." />

Choose printer or PDF destination and check paper format, page breaks and images.
Printing changes no operating entry. The saved default-printer preference is not automatically used
by every printing workflow.

## Resolve entry and print errors

Reload and compare stale data. For missing actions, check roles and status.
For an empty report, first check the selected event and its entries.
Open hidden browser dialogs; investigate physical printing issues at OS and printer level.

## Related pages

[Illustrated exercise](practice) · [Event status](lists-and-locking).
