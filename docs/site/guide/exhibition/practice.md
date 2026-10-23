---
title: Review, lock and print an operating day
description: Review, lock and print an operating day.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Review, lock and print an operating day

**Goal:** Review a saved event for completeness and generate the intended report.
Use **Admin** and synthetic data. This workflow sends no writes to a command station.

## 1. Prepare an event

Choose **Exhibition > New event**, creating **Demo operating day** with start and end.
Select it on the left and add owner **Demo club**, locomotive **Demo F0 F1 F3**, address **3** and
function keys. Initially leave **Command station / interface** empty to explore validation.

<GuideFigure name="exhibition-open" alt="Current event with a deliberately incomplete demo entry." caption="The actual check identifies missing data. An existing inventory locomotive does not replace a complete operating entry." />

## 2. Identify and correct the conflict

Select **Lock list**. The missing station/interface prevents silent locking.

<GuideFigure name="exhibition-lock-conflicts" alt="Confirmation when trying to lock the incomplete list." caption="The demo cancels and corrects the entry first. An exception would need a reason and remain visible." />

Select **Cancel**, edit the entry and fill **Command station / interface: Demo station**.

<GuideFigure name="exhibition-entry" alt="Entry dialog after filling the demo station." caption="Control context is separate from adapter or decoder connector. Saving changes only the operating entry." />

Save and check conflicts again. A missing image can still produce **Check**.

## 3. Lock planning

Select **Lock list** again. Once required details are complete, locking saves directly.

<GuideFigure name="exhibition-locked" alt="Saved event after successful locking." caption="Entries are protected. Printing and status-appropriate actions remain available." />

## 4. Inspect the complete report

Select **Print** in the event header. In the browser print workflow, check event, locomotive, owner,
address and function keys. The report includes all event entries independently of table filters.

<GuideFigure name="exhibition-print-report" alt="Actually generated report of the reviewed event." caption="Actual report HTML rendered separately in a browser. This is not an OS print dialog or printer-driver acceptance." />

Only then choose a printer or PDF destination. Reopen planning with **Unlock list** if necessary.
[Status and roles](lists-and-locking) · [Entry and print details](entries-and-printing).
