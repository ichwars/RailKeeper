---
title: Look up a decoder
description: Look up a decoder.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Look up a decoder

The decoder library explains what RailKeeper knows about a particular decoder model. It is a reference and does not change vehicle readings.


<GuideFigure name="decoder-library" alt="Filter by manufacturer, family or model; details show evidence and limitations." caption="Filter by manufacturer, family or model; details show evidence and limitations. Synthetic demo data, not factory settings. Click for full size." />

## Find a model

1. Open **Settings > Data > General master data > Decoder library**.
2. Choose a manufacturer on the left, or search by model, article number, manufacturer ID, or CV.
3. Open the matching model and check interfaces and article numbers.
4. Read the **coverage** information. Being listed in the catalog does not mean every CV can be interpreted.
5. Open **CV definitions**, **Identification**, or **Sources**, depending on your question.

## What do the labels mean?

| Label | Meaning |
| --- | --- |
| DCC basics | documented meanings of general base registers |
| CV rules verified | the profile has reviewed interpretation rules |
| manufacturer reference available | a documented table or note can be consulted; it is not automatically an interpretation rule |
| partially captured or no matching documentation | the remaining limitation is stated explicitly |

A similar model name or identical connector does not establish the same CV layout. Choose the exact model and check the source's firmware scope.

## What happens to saved CVs?

RailKeeper interprets existing readings when displaying them. New explanations normally do not require another import. The library does not replace missing readings or unknown index pages. Conflicts remain visible.

For ECoS memory values, the vehicle's recorded decoder type and matching manufacturer ID can establish the profile. The source remains ECoS. [Understand CV values](decoder-cv).
