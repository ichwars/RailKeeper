---
title: Vehicle sets
description: Vehicle sets.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Vehicle sets

Several coaches supplied in one box belong together, but still have their own vehicle numbers and details. Use a **set** for this.


<GuideFigure name="vehicle-set" alt="A set has shared product details and individually tracked members." caption="A set has shared product details and individually tracked members. Synthetic demo data, not factory settings. Click for full size." />

## Create a set

1. Open **Vehicle inventory** and **New vehicle**.
2. Choose the **set** creation mode.
3. Enter shared information: name, manufacturer, article number, and gauge.
4. Add at least two members. Give each coach its own name and vehicle number.
5. Check shared purchase and packaging information, then save the set.

**Result:** The set and its members remain linked. Individual vehicles can still be found and edited separately. RailKeeper assigns separate inventory numbers to the set and its vehicles.

## Change shared or individual information?

| Information | Where to edit it |
| --- | --- |
| shared article number, EAN, purchase, or packaging information | edit the set |
| a coach's vehicle number or technical differences | edit that vehicle |
| shared picture or set representation | the set's upload section |

The set view lists its members. Before saving, check whether you are editing the shared record or one coach. A vehicle image and a shared set image can serve different purposes.

## Transfer data

For a complete move, use an [application backup](../../administration/backup-restore). A CSV/JSON transfer can retain set membership, but it does not include media or all vehicle child records. [Export limits](../import-export/exports).
