# v0.2.0 stabilization and acceptance

## Release boundary

v0.2.0 consolidates the modernization after v0.1.22. Changes are grouped into dependency
integration and stabilization, with a single stable release after qualification. Feature
work pauses during the initial feedback period; regressions receive focused fixes with a
clear reproduction and preservation check. No rollout is inferred from a green pull request:
the running version, build identity and platform artifacts must identify the merged source.

The initial technical review covered all twenty original assessment items and N1–N7.
Architecture boundaries, no SQL in the application coordination layer, complete coverage
scope, platform builds, authentication, imports, decoder/CV and valuation preservation
remain mandatory. Executable file size and compiler equality now have negative CI fixtures.
API contract tests exercise successful requests, roles/CSRF, malformed inputs, unavailable
storage, legacy path confinement, native printer payloads and unchanged original bytes.

## Local rollout

1. Confirm successful CI and native AMD64/ARM64/Windows qualification for the exact source.
2. Retain a protected full installation backup, including local configuration and authentication,
   outside Git. Keep the independent collection export and its compatibility report.
3. Start the new binary/image against an offline clone first. Compare original typed database
   cells, file bytes and local authentication before opening the updated instance.
4. At cutover, take a fresh stopped-instance backup and compare the original records and files
   again. Verify health, setup state, protected routes, version and proxy access.
5. Retain the previous software and cutover backup for recovery. After any future schema
   change, restore a matching full backup rather than running an older binary on migrated data.

## Feedback period

Use a 14-day observation period after the verified rollout. Record the actual start/end,
installation version, scenarios tried and any reproduction. A calendar entry or automated
fixture does not count as a completed human pilot. A club/physical-hardware acceptance
needs real participants and actual read/compare/confirmed-adoption results. No messages
are sent to testers automatically.

- Open existing vehicles and accessory articles, including images, documents, stock, storage,
  maintenance, purchases and set membership. Compare a known record before/after reload.
- Enter a valuation, correct it and compare history, effective dates and separate currencies.
  Check that unknown historical values remain explicitly unknown.
- Open all decoder configurations, CV working groups, indexed CVs, histories and original
  source files. Cancel an edit, retry a rejected save and download an original file.
- Read a command-station roster, leave the view and return. Confirm the stored result stays
  until a new read; compare assignments before adopting any change. Device writes require
  their usual explicit confirmation.
- Try German/English and light/dark on the normal desktop and a narrow display. Check search,
  footer labels, focus, calendars and permanently reachable report/settings/sign-out actions.
- Export a collection backup and validate it. Exercise restore on an isolated copy and verify
  that target-local users remain unchanged. Never use a live collection for a destructive test.

## Product follow-ups

Mobile offline capture and synchronization between sites require a separately accepted
conflict, identity, privacy and recovery design. They are not incomplete implementations
hidden behind v0.2.0 controls. The current local-first deployment and its existing workflows
remain the release scope.
