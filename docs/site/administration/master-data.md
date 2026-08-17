---
title: Master-data administration
description: Govern RailKeeper master data, inventory-number schemes, locations, and transfers safely.
audience: admin
status: stable
reviewedVersion: 0.1.19.1
lastReviewed: 2026-08-16
---

# Master-data administration

RailKeeper uses controlled master data wherever records must share names, classifications, symbols,
or numbering rules. Most of it is managed under **Settings > Data**. Inventory-number schemes are
under **Settings > General**, and the master-data JSON transfer is under
**Settings > Import/Export**.

This section documents stable RailKeeper v0.1.19.1. It separates four administrative workflows:

- [General master data](./master-data-general) for vehicles, manufacturers, CV8 identities, and
  function symbols.
- [Article master data and storage locations](./master-data-articles) for accessory catalogue and
  stock structure.
- [Inventory-number schemes](./master-data-inventory-numbers) for automatic vehicle and article
  identities.
- [Master-data transfer](./master-data-transfer) for complete JSON export and reconciliation.

## Access rights

| Action | Admin | Editor | Viewer | Planner | Messe |
| --- | --- | --- | --- | --- | --- |
| Open Settings and read master data | Yes | Yes | Yes | Yes | No |
| Create, edit, deactivate, or reactivate entries | Yes | Yes | No | No | No |
| Permanently delete an eligible custom entry | Yes | Yes | No | No | No |
| Manage storage locations | Yes | Yes | No | No | No |
| Manage inventory-number schemes | Yes | Yes | No | No | No |
| Export or import the master-data document | Yes | No | No | No | No |

Viewer and Planner inherit read access. Article management explicitly displays a read-only state.
Some controls in the general master-data and inventory-number panels can still remain visible to
those roles in v0.1.19.1, but the server rejects every write. Visible controls are never an access
grant.

A pure Messe account cannot open Settings. It can read the active function-symbol subset only
through the isolated exhibition workflow.

## Understand origin and lifecycle

Every controlled entry has an origin:

| Origin | Meaning | Permanent deletion |
| --- | --- | --- |
| **Bundled** | Shipped and reconciled by RailKeeper | Never allowed |
| **Custom** | Created locally or imported without a matching bundled identity | Allowed only while unused |

Both origins can be edited, deactivated, and reactivated. Deactivation removes an entry from new
selections. Existing records retain their stored value and continue to show it as inactive. It is
therefore the safe way to retire a value that remains part of inventory history.

Permanent deletion is deliberately narrower. RailKeeper offers it only for a custom entry whose
type has a known usage check and whose key or label is not referenced. Standard article-type keys
are additionally protected. Deleting an eligible entry also removes its master-data relations and
cannot be undone.

## Use a conservative change sequence

1. Create and validate a current application backup before a broad change.
2. Export the current master-data JSON when changing many classifications.
3. Prefer editing a spelling mistake over creating a near-duplicate value.
4. Deactivate an obsolete value first and verify affected vehicle, article, and exhibition views.
5. Delete only a confirmed unused custom entry.
6. Test creation of one affected vehicle or article before continuing with a batch of changes.

Inventory-number schemes and storage locations are not part of the master-data JSON export. Use an
application backup when those settings must be recoverable together with the inventory.

## Related pages

- [Settings overview and permissions](/guide/settings/)
- [Accessories overview](/guide/accessories/)
- [Vehicle inventory](/guide/vehicles/)
- [Installation and Administration](./)

## Documented RailKeeper version

This section documents stable RailKeeper **v0.1.19.1** and was last reviewed on 2026-08-16.
