---
title: Versions and help
description: Versions and help.
audience: reference
status: stable
reviewedVersion: 0.2.0
lastReviewed: 2026-11-10
---

# Versions and help

## Which version am I using?

RailKeeper shows its version in the interface and under **Settings > General > Updates**. A release is a published version. `main` is the current development branch and may already include additional features. Screenshots come from isolated demos; each page records its own editorial review date. Older installed releases may differ. Identify a development build by its commit or package provenance as well as its displayed version.

The **Build** line shows a short commit and a separate **Release**, **Development build**,
**Local changes** or **Unknown build origin** state. Hover over the commit for its full value.
Include version and build when reporting an issue. `railkeeper version` (Windows:
`RailKeeper.exe version`) prints the same identity plus compiler/platform without opening
application data. An old server without build metadata keeps its existing version display.

A page's visible **editorial review** date records content review. **Last updated** at the bottom is its Git modification date. Neither alone proves testing with real hardware. The [coverage overview](coverage) explains the evidence.

Handbook pictures show **v0.2.0**. Capture version and editorial review scope are recorded separately.
The complete [German changelog](https://github.com/ichwars/RailKeeper/blob/main/CHANGELOG.de.md#020---2026-11-09)
and [English changelog](https://github.com/ichwars/RailKeeper/blob/main/CHANGELOG.md#020---2026-11-09)
describe this release's changes.

The [v0.2.0 release evidence](https://github.com/ichwars/RailKeeper/blob/main/docs/releases/v0.2.0.md#go-test-and-coverage-evidence)
links the successful Go test run and permanently retained coverage reports.

## Feedback from the v0.2.x trial

Use the short [experience report template](https://github.com/ichwars/RailKeeper/issues/new?template=experience_report.md).
Include version/build, installation method, optional command station and trial period. What worked,
what was unclear and which feature was difficult to find? Fill in either language section.
The [initial club feedback review](https://github.com/ichwars/RailKeeper/issues/376)
is planned for 23 November to 7 December 2026. Keep new features separate during this period;
verified 0.2.x fixes and related documentation go through PRs.
A review without actual feedback does not constitute practical acceptance.

## Where should I ask?

| Need | Destination |
| --- | --- |
| How do I use a feature? | [Handbook](../guide/) and [Q&A](https://github.com/ichwars/RailKeeper/discussions/categories/q-a) |
| An idea or desired workflow | [Ideas](https://github.com/ichwars/RailKeeper/discussions/categories/ideas) |
| Reproducible bug | [New issue](https://github.com/ichwars/RailKeeper/issues/new/choose) |
| Experience from the trial | [Experience report](https://github.com/ichwars/RailKeeper/issues/new?template=experience_report.md) |
| Vulnerability or private data | [Private security report](https://github.com/ichwars/RailKeeper/security/advisories/new) |
| Packages and official release notes | [GitHub Releases](https://github.com/ichwars/RailKeeper/releases) |
| Current project entry | [Discussion welcome](https://github.com/ichwars/RailKeeper/discussions/1) |

## Ask a useful question

Include version, Windows/Docker, language, steps, expected and actual results. For a command station, include model and firmware. A cropped screenshot can illustrate the affected area. Remove private collection data, passwords, session tokens and internal addresses you do not want published. Do not attach a complete backup to a public issue.

Mark a helpful Q&A reply as the answer when it solves the problem, so other readers can find it.
