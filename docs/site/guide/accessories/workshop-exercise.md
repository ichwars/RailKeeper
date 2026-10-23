---
title: Practice accessory usage and maintenance
description: Practice accessory usage and maintenance.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Practice accessory usage and maintenance

**Goal:** Follow how reservation, installation and removal change stock, then complete vehicle
maintenance. Use **Admin** or **Editor** and synthetic data only.

## 1. Start with four pieces

Create a quantity article **Demo lighting board**, unit Piece. Add an active master-data storage
location **Demo drawer** and book four pieces there as initial stock.
Use an existing test vehicle as the allocation target.
[Create an article](article-records) · [Book quantity stock](stock-purchases-documents).

The pictures show actual local demo transactions. Intermediate states were prepared through local
application APIs and checked in the actual UI. No command-station operation is involved.

## 2. Reserve one piece

Edit the article and open **Usage & history**. Choose vehicle, location and quantity **1**.
Optionally describe placement and connection. Select **Create reservation** and confirm.

<GuideFigure name="stock-reserved" alt="Active reservation of one piece for the demo vehicle." caption="Four pieces remain in storage, one is reserved and three are available." />

A reservation holds material for a vehicle. It does not establish a physical installation.

## 3. Install the reservation

Select the active reservation in the installation form. Review inherited target and source.
Choose condition, then **Record installation**, and confirm.

<GuideFigure name="stock-installed" alt="Active installation of the previously reserved accessory." caption="The reservation is fulfilled. Three pieces remain in storage and one is installed." />

## 4. Remove it back to storage

Select **Remove** at the active installation, choose **Stored** and the active storage location.
Review and confirm.

<GuideFigure name="stock-returned" alt="Usage history after removal back into stored stock." caption="History remains present. All four pieces are stored and available again." />

| State | Stored | Reserved | Installed | Available |
| --- | --- | --- | --- | --- |
| Initial | 4 | 0 | 0 | 4 |
| Reserved | 4 | 1 | 0 | 3 |
| Installed | 3 | 0 | 1 | 3 |
| Back in storage | 4 | 0 | 0 | 4 |

This table applies to the quantity article with removal disposition **Stored**.
Individual assets and removal to defective, maintenance or retired have different lifecycle effects.
[Details and troubleshooting](allocations-history).

## 5. Complete vehicle maintenance

Open **Edit > Maintenance** on the test vehicle. Add a cleaning record with status **planned**
and a past due date, then save that maintenance entry.

<GuideFigure name="maintenance-open" alt="Saved overdue cleaning record for the demo vehicle." caption="The record is saved and still open. Other maintenance records may also contribute to totals." />

Perform the work and select **Done** at this record.

<GuideFigure name="maintenance-done" alt="Maintenance after being actually marked done." caption="The entry now has done status and a completion date. This action saves immediately." />

Reopen the vehicle and check status and date. Maintenance actions save independently of the
vehicle's main save button and may reload it. Deliberately save or discard other drafts first.
[Maintenance reference](../vehicles/maintenance).
