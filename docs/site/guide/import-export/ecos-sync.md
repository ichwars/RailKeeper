---
title: Import ECoS locomotives and cached CVs
description: Import ECoS locomotives and cached CVs.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Import ECoS locomotives and cached CVs

**Goal:** Compare ECoS locomotive data with your inventory and import cached CV values. This requires Admin and a [configured ECoS connection](../../administration/digital-centers).


<GuideFigure name="central-workspace" alt="Command stations is a separate workspace; this demo has no station connected." caption="Command stations is a separate workspace; this demo has no station connected. Synthetic demo data, not factory settings. Click for full size." />

## 1. Read the list and select the correct locomotive

1. Open **Command stations** in the sidebar.
2. Check the active ECoS and select **Read data**.
3. Select the locomotive in the worklist.
4. Compare name, address, protocol and ECoS object ID with the actual vehicle.

A saved external assignment is useful evidence. Address and name alone do not prove identity. Review suggested RailKeeper matches before importing. The comparison worklist is stored on the server and can be reopened later.

## 2. Read all cached CVs

Open **Stored ECoS CV values** and start **Read all stored CVs**. RailKeeper queries CV1 through CV1024 in small batches and displays progress. You can stop and resume. This does not read the decoder again on a programming track: it retrieves only the ECoS's stored CV cache.

Missing values remain unknown rather than becoming zero. Creating or assigning a vehicle includes the complete retrieval. Once retrieval is complete, **Add CV values** is available for an already assigned vehicle.

## 3. Review the import

Check counts, values and the target inventory. Existing values are not silently replaced; differences remain traceable. Indexed CVs and CVs without an index are separate records. ECoS remains identifiable as the source.

CV8 may establish the manufacturer. Model-specific rules also require a matching model identity, such as the vehicle's saved decoder type. The library cannot reconstruct missing index pages: ambiguous values retain their unresolved identity. [Understand CVs](../vehicles/decoder-cv) and [decoder library](../vehicles/decoder-library).

## 4. Create or assign a vehicle

Choose the appropriate worklist action. For a new vehicle, complete required fields the station cannot supply, such as manufacturer, gauge and type. Review an existing-vehicle assignment particularly carefully. Save only after checking it.

Vehicle creation with external assignment has several steps. If a later action fails, inspect the saved vehicle and job before creating another one to avoid duplicates.

## What can be written back?

Reading and importing into RailKeeper does not change decoder CVs. The separate write workflow previews locomotive-data changes and requires confirmation. Image transfer and assignment have their [own preview](../vehicles/locomotive-images). CV import does not program a decoder or replace a LokProgrammer project.

## When something is missing

| Observation | Next step |
| --- | --- |
| Website works, protocol connection fails | Check TCP port 15471 from the RailKeeper server; the website uses a different port |
| Fewer CVs than in LokProgrammer | ECoS supplies only its cache; import the complete TXT export in the vehicle if needed |
| A value has no explanation | Check model, firmware and index context; missing evidence stays unresolved |
| Wrong vehicle match | Stop the import and correct the assignment |
| Incomplete retrieval | Check progress and errors, then resume deliberately |

The [CS3 read workflow](cs3-import) supplies different data. Z21 and Intellibox 3 do not yet offer an equivalent locomotive/CV import.
