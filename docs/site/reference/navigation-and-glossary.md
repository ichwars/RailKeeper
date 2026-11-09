---
title: Navigation and glossary
description: Navigation and glossary.
audience: reference
status: development
reviewedVersion: main
lastReviewed: 2026-11-09
---

# Navigation and glossary

## Where do I go?

| You want to … | Open in RailKeeper … | Guide |
| --- | --- | --- |
| maintain a locomotive or wagon | Vehicle inventory | [Your first vehicle](../guide/quick-start) |
| record a set with individual vehicle numbers | Vehicle inventory, new set | [Vehicle sets](../guide/vehicles/sets) |
| maintain symbols and function keys | Edit vehicle > Control | [Decoders and CVs](../guide/vehicles/decoder-cv) |
| attach a CV TXT export or decoder file | Edit vehicle > CV > Files & import | [CV import](../guide/vehicles/decoder-cv#files-for-a-complete-import) |
| look up a decoder | Settings > Data > Decoder library | [Library](../guide/vehicles/decoder-library) |
| compare data with a command station | Command stations | [Connect a station](../administration/digital-centers) |
| exchange inventory files | Import/Export | [Data transfer](../guide/import-export/) |
| prevent data loss | Settings > Backup | [Backups](../administration/backup-restore) |

## Shared controls

Enter numbers directly. For fields with a fixed step, the keyboard arrow keys and the
**Increase value** and **Decrease value** buttons change the value within its bounds.
Fields accepting arbitrary decimal values use direct entry. An empty field remains distinct
from an entered zero. Choose files with the labelled button, which also supports Tab and Enter.
Table checkboxes can display a partially selected group.

## Important terms

| Term | Plain explanation |
| --- | --- |
| Inventory number | Your unique identifier for an individual item, separate from the manufacturer's article number |
| Article number | Manufacturer's product identifier; multiple items may share it |
| Vehicle number | Running number printed on the vehicle, for example 218 123-4 |
| Digital address | Address used by the command station to contact the decoder |
| Decoder profile | Documented rules for a specific model; the manufacturer alone cannot identify every CV |
| CV | A numbered decoder setting; its raw value and its explanation are separate information |
| Index page | Additional addressing through CV31/CV32; the same CV number can have different values and meanings on different pages |
| Provenance | Evidence of a value's source, such as export file, original line or station |
| Switching mode | Latching stays on/off; momentary is intended as a push button |
| Inversion | Reversal of the described function, independent of travel direction |
| Transfer profile | Reusable settings for an import or export |
| Application backup | Restorable application data and media, excluding local user accounts and session keys |
| Server | The computer running RailKeeper; for remote access it may differ from your browser computer |

## Keyboard and pictures

Tab moves between controls; Enter activates buttons. Escape closes many dialogs; check unsaved-change notices before leaving. Click handbook pictures to open them at full size. Use the handbook search to find a term or task.

## Dates and unknown event times

Date fields accept `DD.MM.YYYY` and `YYYY-MM-DD`. Calendar labels follow the chosen German
or English language. Opening the calendar leaves the current value intact; **Today** chooses
today only after you activate it. **Clear date** or **Clear** removes an optional date.
An empty installation/removal date continues to mean unknown. Invalid or incomplete typed
dates return to the previous value when leaving the field; check the displayed date before saving.
