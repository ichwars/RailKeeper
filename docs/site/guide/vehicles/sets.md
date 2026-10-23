---
title: Create and edit a coach set
description: Create and edit a coach set.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Create and edit a coach set

**Goal:** Maintain shared product information once while finding each coach separately.
Use Admin or Editor and synthetic exercise data.

## 1. Create a two-coach set

1. Open **Vehicle inventory > New vehicle** and choose **Set** entry.
2. Name it **Demo-Wagenset**, manufacturer **Roco**, article number **DEMO-SET**, gauge **H0**.
3. Choose a suitable category and subtype, for example passenger coach and express coach.
4. Add **Demo-Wagen A** and **Demo-Wagen B**, with different vehicle numbers.
5. Review shared details and save.

<GuideFigure name="vehicle-set" alt="Set creation with shared fields and individual members." caption="The set draft groups product information; members retain their individual vehicle details. Synthetic demo." />

**Expected:** One set inventory number and two independent vehicle inventory numbers.
The active set scheme supplies the set number; members use their matching vehicle schemes.

## 2. Find members again

Search inventory for **Demo-Wagenset** and expand its row using the arrow.

<GuideFigure name="set-members" alt="Expanded demo coach set with its two individually managed members." caption="The actual two-coach set in inventory. Searching by set name keeps its members together." />

Open an individual coach to inspect its vehicle number or technical details.
On narrower views, scroll the action table horizontally if needed.

## 3. Change shared information

Select **Edit** on the set row rather than a member row.
Change the shared article number from **DEMO-SET** to **DEMO-SET-2**.

<GuideFigure name="set-shared-edit" alt="Set editor with its shared article number changed before saving." caption="The shared record is being edited. This does not replace the two coaches' individual vehicle numbers." />

Select **Save**, reopen the set and check **DEMO-SET-2**, two members and their distinct vehicle
numbers. This shared change was actually saved in the demo.

| Information | Edit at |
| --- | --- |
| Shared article number, EAN, purchase and packaging | Set |
| Vehicle number or individual technical detail | Individual vehicle |
| Set presentation image | Set dialog Upload |
| Individual coach images | That vehicle's Uploads |

## 4. Transfer related data

CSV/JSON transfer can carry set structure, but contains no media and not all vehicle subrecords.
For a complete compatible move, use an application backup.
[Export limits](../import-export/exports) · [Practice a restore](../../administration/restore-exercise).
