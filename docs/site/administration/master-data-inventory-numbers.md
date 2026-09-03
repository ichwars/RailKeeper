---
title: Inventory-number schemes
description: Configure automatic vehicle and accessory article inventory numbers without collisions.
audience: admin
status: stable
reviewedVersion: 0.1.20.4
lastReviewed: 2026-08-16
---

# Inventory-number schemes

Open **Settings > General > Inventory numbers** to control automatic identities assigned during
vehicle and accessory-article creation. A scheme contains a unique category, prefix, next number,
padding width, active state, and preview.

Stable v0.1.20.3 normally provides these active schemes:

| Category | Default prefix | Default example | Used for |
| --- | --- | --- | --- |
| Vehicle | `RK-FAH` | `RK-FAH-000001` | Vehicle fallback |
| Locomotive | `RK-LOK` | `RK-LOK-000001` | Categories containing `lok` |
| Wagon | `RK-WAG` | `RK-WAG-000001` | Categories containing `wagen` or `waggon` |
| Article | `RK-ART` | `RK-ART-000001` | Accessory article records |

The displayed labels are localized, but the stored categories above use their German values
`Fahrzeug`, `Lokomotive`, `Wagen`, and `Artikel`.

## Assignment rules

RailKeeper reserves the number inside the same database transaction that creates the record. A
successful reservation increments **Next number**. If the create transaction fails, its number is
not consumed.

For a vehicle, RailKeeper first looks for an active scheme whose category exactly matches the
selected vehicle category. It then tries the classified Locomotive or Wagon fallback and finally
the Vehicle scheme. An accessory article requires the active Article scheme and has no fallback.

The generated form is always `PREFIX-NUMBER`. RailKeeper uppercases the saved prefix, trims it, and
replaces spaces with hyphens. Padding must be from 1 through 12, and the next number must be at least
1. Padding is a minimum width, so a larger number is never truncated.

Before assignment, RailKeeper checks the relevant record table. If a candidate is already used, it
tries the next value and advances past the collision. It stops after 500 attempts. Vehicle and
article numbers are checked in their respective tables, not as one shared global namespace.

## Create or edit a scheme

1. Review the existing categories before adding one. Each category can have only one scheme.
2. Enter the category, prefix, next number, padding, and active state in the new row.
3. Check the preview and select **Create**.
4. For an existing row, edit its values in place and select **Save**.

There is no delete endpoint in v0.1.20.3. Deactivate an unused scheme instead. Editing or
deactivating a scheme never renumbers existing records.

Only Admin and Editor may create or save schemes. Viewer and Planner can read the table. The stable
UI can still show editable-looking fields to those read-only roles, but the server rejects the
write.

## Change the next number safely

Moving **Next number** forward intentionally leaves a gap. Moving it backward does not reuse an
existing number silently because collision checking skips occupied candidates, but it can require
many attempts and makes audits harder. Prefer a forward-only sequence.

Before changing prefixes or categories:

1. Export an application backup.
2. Record the current preview and next number.
3. Confirm that the prefix does not overlap another scheme in the same record domain.
4. Save the scheme and create one test record.
5. Verify the assigned number and the updated next value.

When article creation reports that no number can be allocated, reactivate or repair the Article
scheme. For vehicles, verify the exact category scheme and its Locomotive, Wagon, and Vehicle
fallbacks.

## Backup and restore behavior

Inventory-number schemes belong to the application database and are included in application backup
and restore. They are excluded from the master-data JSON transfer. For legacy article rows without
an inventory number, restore ensures that an Article scheme exists. It assigns missing numbers only
when that scheme is active; otherwise restore is rejected. Generated values do not duplicate other
restored article numbers. The scheme can remain numerically behind existing article numbers; later
article creation safely skips those collisions.

## Related pages

- [Master-data administration](./master-data)
- [Master-data transfer](./master-data-transfer)
- [Vehicle inventory](/guide/vehicles/)
- [Article records and technical data](/guide/accessories/article-records)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.20.3** and was last reviewed on 2026-08-16.
