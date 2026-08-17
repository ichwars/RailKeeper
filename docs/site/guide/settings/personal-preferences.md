---
title: Personal preferences
description: Configure language, start page, date and time choices, printing, and sidebar order.
audience: user
status: stable
reviewedVersion: 0.1.19.1
lastReviewed: 2026-08-16
---

# Personal preferences

Open **Settings > General** for the personal preference controls. Changes take effect immediately;
there is no separate Save button.

## Language

Choose **Deutsch** or **English**. RailKeeper immediately changes the interface text and the
document language marker used by the browser.

The language choice is stored only in the current browser in stable v0.1.19.1. It does not follow the
user profile to another browser or device. A browser without a stored choice starts in German.

## Default view

The default view is used when RailKeeper starts from its root address without a more specific path.
Available choices are Overview, Vehicle inventory, Accessories, Layout, Exhibition, Import/Export,
and Settings.

Keep these boundaries in mind:

- A direct link such as `/vehicles` or `/settings` takes precedence over the default.
- Signing in initially routes the account to its first permitted workspace. The default does not
  bypass that role-based decision.
- A stored destination that the current role cannot open is replaced by the first allowed view.
- **Layout** is selectable in v0.1.19.1 but remains an unpublished development workspace and is not
  shown in the normal sidebar. Choose another start page for the stable user workflow.

## Date and time format

The date preference offers **System default**, a German day-month-year format, and ISO
year-month-day. The time preference offers **System default**, 24-hour, and 12-hour time.

RailKeeper stores and synchronizes both choices. Stable v0.1.19.1 does not yet connect these settings
to all date and time output across the application. Many current views continue to format values
from the selected interface language or their own fixed formatter. Treat the controls as prepared
preferences, not as a guaranteed global override.

## Default printer

The selector contains:

- **System dialog / default printer**;
- a named printer when an Admin account can read configured or operating-system printers;
- **Ask every time**;
- **Save as PDF**.

When RailKeeper can discover a system default while **System dialog / default printer** is active,
it stores the detected named printer as the preference. Printer discovery itself is Admin-only.
Other roles can still retain and change the general print preference.

Stable v0.1.19.1 stores this value but does not route all print actions to the chosen destination.
Vehicle inventory printing, exhibition reports, and Import/Export printing still open the browser
print dialog. The browser and operating system remain responsible for choosing the real printer or
**Save as PDF** destination.

## Sidebar order and visibility

The **Sidebar Order** list controls the main navigation for the current user:

1. Use the up and down buttons to move an entry.
2. Use the eye button to hide or show an entry.
3. Select **Reset** to restore the default order and show all entries.

**Settings** cannot be hidden, so the configuration page remains reachable. Hiding an entry removes
only its sidebar link. The underlying page can still be opened directly when the account has the
required role. Entries unavailable to the current role remain filtered out regardless of their
saved position.

The list can include **Layout** even though stable v0.1.19.1 does not publish that workspace in the
normal sidebar. Ordering or showing it here does not turn it into a published user feature.

Sidebar order and hidden entries are stored separately for each username and synchronized through
that user's profile. The small arrow at the bottom of the sidebar only collapses or expands the
sidebar in the current browser; that collapsed state is not part of these profile preferences.

## Troubleshooting

| Symptom | Check |
| --- | --- |
| A preference changed locally but not on another browser | Reload Settings while signed in as the same user. The background profile write may have failed even though the local value remained active. |
| The interface language differs on another device | Set it again there. Language is browser-local in v0.1.19.1. |
| The selected start page is not opened after sign-in | Role-based login routing and direct URLs take precedence. Open the root address to test the stored default. |
| A hidden page can still be opened | Hiding changes navigation only, not authorization or routing. |
| Named printers do not appear | System-printer discovery requires Admin and depends on the server operating system or printer configuration. |
| The selected printer or date format has no effect | These preferences are stored but are not yet used globally by every v0.1.19.1 workflow. |

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.19.1** and was last reviewed on 2026-08-16.
