---
title: Import ECoS locomotives and cached CVs
description: Import ECoS locomotives and cached CVs.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-29
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

## 2. Review saved locomotive data together

**Read data** collects locomotive metadata, saved function keys and symbols, and complete CV
exports including index pages in the same comparison. Locally configured backup inspection
checks and supplements these sources internally. No separate file download or import choice
is required. No decoder is read again and no programming track is needed.

Open the locomotive in the worklist to review saved values, provenance and disagreements.
Missing entries remain unknown. Saved data and profile defaults are not fresh decoder
measurements. An unavailable internal source remains visible as a limitation.

## 3. Review the import

Conflicting sources retain their differing values for review and are excluded from automatic
adoption. Unverified memory bytes are not known CV readings. For linked vehicles, **Add CV values**
adds only missing complete register addresses from the saved snapshot. Existing values and manual
entries remain intact.

Use **Show supplied values and provenance** to check counts, values and the **CV31 / CV32** column. CV257 on page 16/3 differs from CV257 on page 16/8. Existing readings at the same full address retain their value and provenance, including across profiles. Repeating the import creates no duplicates. ECoS remains identifiable as the source.

The library also considers the saved ECoS decoder profile. An exact, unique match to an active
model profile can enable its documented CV and function rules even when CV8 is absent. The display
explicitly identifies the ECoS profile as its source. Conflicts with the manual decoder type or
supplied manufacturer/model CVs block this match. Generic, unknown or ambiguous profiles do not
receive guessed model rules.

Manual decoder types, CV readings and function assignments remain unchanged. Missing CV7/CV8,
firmware and index pages are never invented. Function names and symbols are added explicitly to
the draft. The profile is saved ECoS context, not a new decoder measurement. After clearing the
comparison worklist or restoring JSON without that workspace, read ECoS again to restore this
context. [Understand CVs](../vehicles/decoder-cv) and [decoder library](../vehicles/decoder-library).

## 4. Create or assign a vehicle

Previously read CVs and function keys are included in the same vehicle draft. Known symbols
and supported meanings are checked against the decoder library. Unknown codes and disagreements
remain visible. Review the normal vehicle and function fields and save the vehicle; no additional
import dialog is required.

Choose the appropriate worklist action. For a new vehicle, complete required fields the station cannot supply, such as manufacturer, gauge and type. Review an existing-vehicle assignment particularly carefully. Save only after checking it.

Vehicle creation with external assignment has several steps. If a later action fails, inspect the saved vehicle and job before creating another one to avoid duplicates.

## Function keys in the vehicle draft

For a new locomotive, the ordinary vehicle draft already includes function keys collected in
the same read. Review names, symbols and key modes in the normal function fields. Unknown
codes retain their provenance. Source/library disagreements are not resolved automatically,
and existing manual assignments remain intact.

<details>
<summary>Review existing vehicles separately when needed</summary>

1. Save the ECoS locomotive’s mapping to the vehicle under **Command stations**.
2. Open **Edit vehicle > Control > Import function keys from ECoS**. Reading requires Admin access.
3. Choose **Read from ECoS**. Keys F0–F31 are checked; hidden and unavailable keys create no entries.
   No locomotive function is switched.
4. Compare ECoS, existing RailKeeper fields and the **Decoder library** side by side.
   The library uses the selected CV profile, matching saved ECoS decoder profiles,
   supported decoder identity and available index pages.
   Missing index pages are never guessed. Unknown symbol codes remain visible.
5. Select individual rows or choose **Select empty fields**. Conflicting ECoS and decoder symbols
   or key modes block the row. Review those fields individually in the function editor.
6. Choose **Apply selection to draft** and review the functions. Manual names, symbols,
   saved key modes and personal notes are preserved. Source evidence records ECoS, key,
   symbol code and read time; any CV evidence used remains visible.
7. Save the vehicle together or discard the draft with **Cancel**.

</details>

Model rules also cover **LokPilot 5 nano DCC**, **LokPilot micro V4.0 / DCC** and
**LokPilot Standard**, each using its reviewed CV and mapping tables. Missing registers
are not treated as zero; incomplete mapping rows keep their function meaning unknown.
The reviewed older Standard mapping table covers keys through F18. ECoS keys F19–F31
can still be imported using the available station information.

Symbols use documented ECoS protocol codes, including 10 for shunting and 11 for momentum.
Unknown codes remain visible for manual assignment. The library update preserves local
symbol settings and existing vehicle assignments.

<GuideFigure name="ecos-functions" alt="Function preview showing protected manual fields and a decoder-library conflict." caption="Actual RailKeeper components with synthetic example data and a simulated ECoS reply. Selected empty fields are added to the draft; no function is switched on the station." />

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
