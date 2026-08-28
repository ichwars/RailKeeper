---
title: First setup and sign-in
description: Create the first administrator, sign in securely, and recover access to RailKeeper.
audience: user
status: stable
reviewedVersion: 0.1.20.3
lastReviewed: 2026-08-16
---

# First setup and sign-in

This chapter covers the first administrator account, normal and two-factor sign-in, sign-out, and
password recovery. It describes stable RailKeeper v0.1.20.3.

## Before you start

RailKeeper must already be installed and reachable in your browser. The first-setup form appears
only while the RailKeeper database contains no user account. RailKeeper has no default username or
password.

If a sign-in form appears instead, setup has already been completed. Sign in with an existing
account or contact the person who operates the instance.

## Create the first administrator

No existing account or role is required for the one-time setup.

1. Open RailKeeper in your browser.
2. Complete every field on the **First Setup** form.

   | Field | Requirement | Purpose |
   | --- | --- | --- |
   | Username | Required, at least 3 characters after surrounding spaces are removed | Identifies the account at sign-in |
   | Email address | Required, valid email address | Receives password-recovery links when email delivery is configured |
   | Password | Required, at least 12 characters | Protects the first account |
   | Repeat password | Required and identical to **Password** | Prevents an unnoticed typing error |

3. Select **Create admin**.
4. After RailKeeper confirms the setup, use the normal sign-in form with the new credentials.

The first account receives the Admin, Editor, and Viewer roles. Setup cannot create a second first
administrator. For protection against repeated attempts, RailKeeper accepts at most five setup
attempts per client address within ten minutes.

## Sign in

1. Enter your **Username** and **Password**.
2. Select **Sign in**.
3. If two-factor authentication is enabled for the account, RailKeeper now shows the
   **Two-factor code** field. Enter the current code from the configured authenticator and select
   **Sign in** again.

A successful sign-in creates a server-side session that lasts for up to 12 hours. Admin, Editor,
Viewer, and Planner users start on **Overview**. An account that has only the Messe role starts on
**Exhibition**.

RailKeeper accepts at most ten sign-in requests from one client address within five minutes. Wait
before trying again instead of repeatedly submitting the form.

## Sign out

Use the log-out icon in the sidebar footer. RailKeeper revokes the current server-side session and
returns to the sign-in form. Closing only the browser tab does not explicitly sign out the session.

## Recover a forgotten password

Password recovery depends on the email address stored for the account.

1. On the sign-in form, select **Forgot password?**.
2. Enter the account email address.
3. Select **Request reset**.
4. Open the newest reset link you receive.
5. Enter and repeat a new password with at least 12 characters.
6. Set the password, return to sign-in, and use the new credentials.

The confirmation shown in the sign-in form and the HTTP response is deliberately identical for
known and unknown addresses so the reset flow does not disclose whether an account exists.

When SMTP is configured, RailKeeper sends the link by email. Without SMTP, reset tokens are neither
returned to the browser nor written to server logs. Contact an administrator or operator if no
message arrives.

Only the newest open reset request remains valid. A reset link expires after 30 minutes and can be
used once. Completing the reset revokes every existing session for that user, including sessions in
other browsers.

RailKeeper limits reset requests to five per client address within ten minutes and reset
confirmations to ten within ten minutes.

## Troubleshooting

| Problem | Cause and action |
| --- | --- |
| The first-setup form does not appear | At least one user already exists. Use the sign-in form or contact the operator. |
| Setup is rejected | Check the 3-character username minimum, the email format, the 12-character password minimum, and both password entries. If attempts were repeated, wait for the ten-minute rate-limit window. |
| Sign-in reports invalid credentials | Check the username and password. RailKeeper deliberately uses the same error for an unknown account and a wrong password. |
| RailKeeper asks for a two-factor code | The account has two-factor authentication enabled. Enter the current code from its authenticator. A wrong or expired code is reported as invalid credentials. |
| No password-reset email arrives | Check the address and spam folder, then contact an administrator. SMTP may not be configured, and the confirmation displayed by the form does not confirm that the account exists. |
| A reset link is invalid or expired | Request a new link. Earlier links become invalid when a newer request is created, after 30 minutes, or after first use. |
| RailKeeper reports too many attempts | Stop submitting the form and wait for the applicable five-minute or ten-minute rate-limit window. |

## Security notes

- Use a unique password and do not share administrator credentials.
- Use HTTPS and secure cookies when the instance is reachable over a network. See
  [Installation and Administration](/administration/) for the operating requirements.
- RailKeeper returns the same reset response for known and unknown addresses. Continue to monitor
  repeated reset attempts and keep network access as narrow as practical.
- Ask an administrator to review unexpected sign-in or recovery activity rather than continuing to
  guess credentials.

## Related pages

- [User Guide overview](/guide/)
- [Installation and Administration](/administration/)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.20.3** and was last checked against the application on
2026-08-16.
