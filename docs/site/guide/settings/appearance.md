---
title: Appearance
description: Choose the theme mode and configure light and dark variants independently.
audience: user
status: stable
reviewedVersion: 0.1.19
lastReviewed: 2026-08-16
---

# Appearance

Open **Settings > Appearance** to control RailKeeper's color mode and its light and dark variants.
Every change applies immediately. There is no separate Save button.

## Theme mode

| Mode | Behavior |
| --- | --- |
| **System** | Follows the operating system or browser light/dark preference. |
| **Light** | Keeps RailKeeper in the configured light variant. |
| **Dark** | Keeps RailKeeper in the configured dark variant. |

The theme button in the sidebar footer provides the same mode choice as a shortcut. Each click
cycles from Dark to Light, from Light to System, and from System to Dark.

A pure Messe account can use this footer shortcut even though it cannot open Settings. Its choice
applies in that browser, but the Messe role cannot write the protected profile setting.

## Configure light and dark variants

RailKeeper stores separate settings for each color mode. Changing the light variant does not alter
the dark variant, and vice versa.

| Option | Light variant | Dark variant | Effect |
| --- | --- | --- | --- |
| Background | Neutral, Warm, Cool | Neutral, Warm, Cool, OLED Black | Changes the base, panel, and input surfaces. |
| Accent | Green, Blue, Gold | Green, Blue, Gold | Changes interactive and highlighted colors. |
| Style | Classic, Compact, Contrast | Classic, Compact, Contrast | Chooses the standard treatment, tighter Settings-card gaps, or stronger separators and secondary text. |

With **System** selected, RailKeeper automatically uses the configured light variant while the
system requests light colors and the configured dark variant while it requests dark colors. Both
variant cards therefore remain useful even when neither Light nor Dark is forced.

**OLED Black** is available only for the dark background. It uses a black base and very dark panel
surfaces. **Compact** in v0.1.19 primarily tightens spacing between Settings cards; it is not a
global table-density control. **Contrast** strengthens separators and muted text but does not
replace browser zoom or operating-system accessibility settings.

## Persistence and user scope

The selected mode and all six variant choices are written to browser storage and to the current
user's profile. RailKeeper restores the mode in the application shell. Opening Settings restores
and applies the detailed light and dark options as well.

The background profile write has no visible success message. If the server is temporarily
unavailable, the current browser can keep showing the local choice while another browser still has
the older profile value. Reload Settings after connectivity returns and select the choice again if
necessary.

Application backup and restore deliberately exclude user profile settings. Appearance choices are
therefore not part of an application backup.

## Troubleshooting

| Symptom | Check |
| --- | --- |
| System mode changes unexpectedly | Check the operating-system or browser color preference. System mode follows that signal. |
| A background or accent appears to have no effect | Confirm whether the application currently uses the light or dark variant you edited. |
| Another browser shows a previous theme | Open Settings with the same user and reload. If needed, choose the mode or variant again after server connectivity returns. |
| Compact does not make inventory tables denser | In v0.1.19 it mainly reduces gaps between Settings cards. |
| Contrast is still insufficient | Combine the Contrast style with browser zoom and operating-system accessibility controls. |

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.19** and was last reviewed on 2026-08-16.
