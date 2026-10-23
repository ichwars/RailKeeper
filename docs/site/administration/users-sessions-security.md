---
title: Accounts and permissions
description: Accounts and permissions.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Accounts and permissions

**Goal:** Give each person the appropriate access. An **Admin** manages accounts under **Settings > Authentication**.


<GuideFigure name="users" alt="Authentication combines personal security with administrative tools." caption="Authentication combines personal security with administrative tools. Synthetic demo data, not factory settings. Click for full size." />

## Add an account

1. Open user management and choose the action for adding a user.
2. Enter username, email, and a separate password.
3. Select only the required roles.
4. Save the account and check its sign-in separately from your own administrator account.

| Role | Intended use |
| --- | --- |
| Admin | manage the installation, accounts, backups, and protected command-station operations |
| Editor | add and edit inventory data |
| Viewer | browse inventory and use permitted outputs |
| Messe | use the allowed exhibition workspace |

Roles can be combined. Account roles determine which menus and actions are allowed; a visible link does not bypass server authorization.

## Protect your account

Use the personal password and two-factor settings under **Authentication**. Set up an authenticator app and protect any required recovery information. [Sign-in and forgotten passwords](../guide/getting-started/).

## Lost or unfamiliar sessions

An Admin can inspect sessions and revoke those no longer needed. Password changes or administrative resets can invalidate existing sessions. After changing permissions, check access with a fresh sign-in to the affected account.

Application backups do not transfer users or sign-in credentials. Operators also need a private server backup for full installation recovery.

[Illustrated account and two-factor exercise](account-security-example) · [SMTP and password reset](smtp-password-reset).
