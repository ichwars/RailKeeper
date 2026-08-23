---
title: Settings
description: Understand personal preferences, appearance, and the boundaries between user and administration settings.
audience: user
status: stable
reviewedVersion: 0.1.20.1
lastReviewed: 2026-08-16
---

# Settings

The **Settings** workspace combines personal preferences with several administrative tools. This
chapter covers the settings that shape an individual user's navigation and appearance. Master data,
command stations, application backup, storage, updates, users, sessions, and authentication are
separate topics even though RailKeeper presents them in the same workspace.

This chapter documents stable RailKeeper v0.1.20.1.

## Access rights

Admin, Editor, Viewer, and Planner can open **Settings**. A pure Messe account remains in the
isolated exhibition workspace and cannot open this page. Personal profile settings use the
authenticated profile API and are isolated by user.

| Action | Admin | Editor | Viewer | Planner |
| --- | --- | --- | --- | --- |
| Open **General** and **Appearance** | Yes | Yes | Yes | Yes |
| Change personal profile preferences | Yes | Yes | Yes | Yes |
| Read named system printers | Yes | No | No | No |
| Run Admin-only operations | Yes | No | No | No |

The visible page is not the permission boundary. RailKeeper checks protected system and data
operations on the server. Reordering or hiding a navigation entry never grants or removes access.

## Choose the correct settings area

| Tab or area | Purpose | Documentation owner |
| --- | --- | --- |
| **General** | Language, start page, date and time preferences, printing preference, and sidebar order | [Personal preferences](/guide/settings/personal-preferences) |
| **Appearance** | System, light, or dark mode plus separate color and style variants | [Appearance](/guide/settings/appearance) |
| **General > Inventory numbers** | Number schemes for inventory categories | [Master-data administration](/administration/master-data) |
| **General > Article search** | Enable article search and choose source groups | [Article search](/guide/vehicles/search-and-spares) |
| **General > Updates** | Release checks and Windows package download | Releases and support |
| **General > Storage** | Storage usage and optimization | System operations |
| **Data** | General and accessory master data, lifecycle, and transfer | [Master-data administration](/administration/master-data) |
| **Command stations** | ECoS, Z21, Intellibox 3, and CS3 configuration | Digital-center administration |
| **Import/Export** | Backup, restore, and master-data transfer inside Settings | Backup and [master-data administration](/administration/master-data) |
| **Authentication** | Own password and two-factor setup; Admin user, session, audit, and SMTP tools | Users, sessions, and security |

The main [Import and export workspace](/guide/import-export/) is a different page. It exchanges
vehicle records and performs the ECoS locomotive workflow. The **Import/Export** tab inside
**Settings** contains administrative backup and master-data transfer instead.

## How personal preferences are stored

Most preferences apply immediately in the browser, are written to browser storage, and are also
sent to the current user's profile. The application shell restores theme mode and sidebar
preferences. Opening Settings restores the other synchronized preferences.

There are important exceptions and limits:

- Interface language is browser-local in v0.1.20.1 and is not written to the server profile.
- The collapsed or expanded state of the main sidebar is browser-local. Sidebar order and hidden
  entries are profile settings.
- Profile saving runs in the background. The page does not show a success or failure message for
  these small preference writes. If the server write fails, the local browser value can still
  remain active.
- Application backups deliberately exclude user accounts and `user_settings`. Do not treat a
  backup as an export of personal preferences.

## Safe working sequence

1. Open **Settings > General** and choose the language and stable start page you want.
2. Arrange the sidebar, but keep frequently used workspaces visible.
3. Open **Appearance** and choose **System**, **Light**, or **Dark**.
4. Adjust the light and dark variants independently when required.
5. Reload RailKeeper once when verifying that a preference was restored.
6. When moving to another browser, sign in with the same user. Set the interface language again
   because language is not profile-synchronized in v0.1.20.1.

## Documented RailKeeper version

This chapter documents stable RailKeeper **v0.1.20.1** and was last reviewed on 2026-08-16.
