---
title: Reservations, installations, and usage
description: Reserve accessory stock, record installations, and understand usage history.
audience: user
status: stable
reviewedVersion: 0.1.19
lastReviewed: 2026-08-16
---

# Reservations, installations, and usage

Reservations assign available accessory stock to a future use. Installations record that the stock
has physically left storage and is in use. RailKeeper keeps both steps, the item lifecycle, stock
movements, conditions, and usage history consistent in one local transaction.

## Roles and allocation targets

Admin and Editor users can reserve, cancel, install, update conditions, and remove installations.
A Planner can create and cancel reservations in the otherwise read-only article dialog, but cannot
install, remove, change condition, or manage general stock. Viewer is read-only. Messe has no
general accessories access.

Every reservation and installation requires exactly one target:

| Target | Selection shown by the stable dialog |
| --- | --- |
| Vehicle | Inventory number and vehicle name. |
| Layout | One non-archived layout. |
| Layout unit | One non-archived layout unit. |

The dialog consumes existing targets but does not create or edit layouts. This chapter does not
document layout creation or editing.

Five optional placement fields can accompany either workflow: **Placement**, **Digital address**,
**Decoder output**, **Connection**, and **Wiring notes**. They describe where and how the accessory
is intended to be or is actually connected. **Notes** is a separate private workflow note.

## Read the allocation summary

The overview and allocation service derive these quantities from the selected inventory strategy:

| Value | Stable meaning |
| --- | --- |
| Owned | Physical units still owned: stored plus actively installed counted units, or all individual items. Hybrid combines counted stock, individual items, and active counted installations without double counting individualized units. |
| Stored | Counted stock plus, where supported, items at a storage location with lifecycle Stored or Reserved. |
| Reserved | Quantity of all active reservations, including one per reserved item. |
| Installed | Quantity of all installations not yet removed. |
| Available | `max(Stored - Reserved, 0)`. |
| Missing | Any amount by which Reserved exceeds Stored. Normal commands prevent this, but restored or legacy data can expose it. |

For counted stock, an installation reduces Stored but stays in Owned until removal. For an
individual item, maintenance and retired items remain owned but are not Stored or Available.

## Reserve stock

1. Select vehicle, layout, or layout unit and the exact target.
2. Optionally enter the five technical placement fields.
3. For a hybrid article, choose **Quantity** or **Individual item** as the source.
4. Select an active storage location. For individual tracking, select a Stored item at that same
   location; its quantity is fixed to 1. Otherwise enter a positive whole quantity.
5. Add an optional note, select **Create reservation**, and confirm.

The confirmed transaction validates the article, target, active location, strategy, and available
source. A counted reservation creates an Active reservation and reduces Available only. It neither
changes physical stock nor writes a stock movement. An item reservation additionally changes that
item from Stored to Reserved while keeping its storage location. The same item cannot have another
active reservation.

If available counted quantity is too small, the item belongs to another article or location, the
item is not Stored, or any target is invalid, the entire reservation fails without a partial state.

## Cancel a reservation

Only an Active reservation offers **Cancel**. The command opens a confirmation. Confirming changes
its status to Cancelled and reloads the related resources. Counted stock remains physically
unchanged and becomes available again. A reserved item returns from Reserved to Stored at its
existing location.

A Fulfilled or already Cancelled reservation is immutable and cannot be cancelled. Fulfilled means
that an installation consumed it; remove that installation instead of trying to reopen the
reservation.

## Record an installation

Admin or Editor opens the article in edit mode and chooses either an Active reservation or
**Without reservation**.

With a reservation, target, source location, item where applicable, and quantity are fixed to the
reservation. Its placement data is copied into the installation form. A non-empty installation
value can replace an inherited placement value; an empty value retains the reservation value.
Condition and installation notes remain selectable.

Without a reservation, select target, optional placement data, source, positive whole quantity or
one Stored item, condition, and optional notes. Hybrid articles again allow the quantity or item
path. Direct item installation is rejected if the item has an active reservation.

**Record installation** opens a confirmation. The successful transaction has these effects:

| Source | Physical stock or item | Reservation | Installation and journal |
| --- | --- | --- | --- |
| Counted quantity | Subtracts the quantity from the source location. Direct installation protects every active reservation at that location. | Selected Active reservation becomes Fulfilled. | Creates an active installation and an Installation movement with the negative quantity. |
| Individual item | Changes the selected Stored or Reserved item to Installed, clears its storage location, and sets the selected condition. | Selected Active reservation becomes Fulfilled. | Creates an active installation. No quantity movement is written. |

The valid installation conditions are Ready, Maintenance due, Defective, and Unknown. Manual UI
entry starts with Ready. All stock, item, reservation, installation, movement, and audit changes are
atomic. If one check fails, none is kept.

## Update installation condition

For an active installation, choose Ready, Maintenance due, Defective, or Unknown and select
**Save condition**. A confirmation appears even when the selection looks unchanged.

Confirming updates the installation, appends a condition-history row with previous and new value,
and updates the condition of an installed individual item. Counted stock has no item record to
update. A removed installation cannot be changed. Resources reload after success.

## Remove an installation

Select **Remove** on an active installation, choose the disposition, add optional removal notes,
and confirm the separate removal dialog.

| Disposition | Counted quantity result | Individual item result |
| --- | --- | --- |
| Stored | Requires an active destination, returns the full quantity, and writes a positive Removal movement. | Lifecycle Stored, selected storage location, condition preserved. |
| Maintenance | Does not return counted quantity to stock. | Lifecycle Maintenance, no storage location, condition preserved. |
| Defective | Does not return counted quantity to stock. | Lifecycle Maintenance, no storage location, condition set to Defective. |
| Retired | Does not return counted quantity to stock. | Lifecycle Retired, no storage location, condition preserved. |

Removal closes the installation with remover, time, disposition, and separate removal notes. It
does not delete the installation. A closed installation cannot be removed again or have its
condition changed. The stock/item change, closure, movement where applicable, and audit event are
one transaction.

## Read usage history

The **Usage & history** tab appears when the article has reservation, installation, or history data.
Its upper section repeats only current Active reservations and installations not yet removed in
read-only form.

The history table then shows events from oldest to newest with date and time, event type, quantity,
target, and condition or removal disposition:

| Event | Created when |
| --- | --- |
| Reservation | The reservation is created. Its record later carries Active, Fulfilled, or Cancelled status; cancellation is not a separate event row. |
| Installation | The installation is created. |
| Condition change | An active installation condition is confirmed, with previous and new condition retained. |
| Removal | The installation is closed, with its final disposition. |

Installation notes, removal notes, and detailed placement fields remain stored but the condensed
stable history table does not display them. Use the current allocation record and backups when
those details are operationally important.

## Keep stock and lifecycle state consistent

Use the workflow command that describes the real event. Do not compensate for an installation with
a manual negative stock adjustment, or for a removal with a positive adjustment. Those shortcuts
would omit the target, installation record, item lifecycle, and usage history.

| Action | Quantity availability | Item lifecycle | Durable history |
| --- | --- | --- | --- |
| Reserve | Decreases Available, Stored unchanged. | Stored to Reserved. | Reservation record/event. |
| Cancel | Restores Available, Stored unchanged. | Reserved to Stored. | Reservation remains Cancelled. |
| Install | Stored decreases and Installed increases. | Stored/Reserved to Installed; location cleared. | Installation event and quantity movement where applicable. |
| Change condition | No quantity change. | Installed item condition follows installation. | Condition-change event. |
| Remove to storage | Stored increases and Installed decreases. | Installed to Stored at target location. | Removal event and quantity movement where applicable. |
| Remove otherwise | Installed decreases without counted stock return. | Installed to Maintenance or Retired, Defective also sets condition. | Removal event. |

Changing the article's inventory strategy is blocked while existing quantity or item dependencies
would become unrepresentable. Resolve active and historical dependencies deliberately instead of
forcing a classification change.

## Resolve allocation errors

| Situation | Next step |
| --- | --- |
| Save is disabled | Select an exact target and active location; for individual tracking also select a Stored item. |
| Insufficient stock | Check active reservations at that source and reduce the quantity or choose another location. |
| Item conflict | Verify article, location, Stored/Reserved lifecycle, and whether another active reservation or installation uses it. |
| Reservation cannot be cancelled | Only Active reservations can be cancelled. A Fulfilled reservation belongs to an installation. |
| Installation from reservation conflicts | Do not change its target, location, item, or quantity. Reselect the current Active reservation and retry. |
| Stored removal is disabled or rejected | Select an active destination whose parent chain is active. Other dispositions must not include a location. |
| Condition or removal conflicts | Reload. Another action may already have closed or changed the installation. |
| Write may have succeeded but refresh failed | Do not repeat it. Use **Retry**, then verify reservation status, installation, item lifecycle, stock journal, and history. |
| Permission is denied | Planner can reserve and cancel only. Installation, condition, and removal require Admin or Editor. |

A failed post-write resource reload marks the editor stale and disables further allocation and
stock writes until a complete retry succeeds. This prevents a second command from using an old
availability or lifecycle state.

## Related pages

- [User Guide overview](/guide/)
- [Accessories overview](./)
- [Article records and technical data](./article-records)
- [Stock, purchases, and documents](./stock-purchases-documents)
- [Vehicle inventory and core records](../vehicles/)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.19** and was last reviewed on 2026-08-16.
