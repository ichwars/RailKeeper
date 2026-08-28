# Documentation governance

## Purpose

The repository documentation contains durable information for users, operators, contributors, and
maintainers. Task-bound working notes do not belong in the published documentation tree.

## Content kept in `docs`

- the German and English user guides under `docs/site`
- architecture and security documentation
- operational runbooks and the product roadmap
- release notes
- documentation source assets, validation scripts, and required manifests
- screenshots that are referenced by maintained documentation

## Content kept outside `docs`

- temporary implementation plans and feature designs
- session handovers and progress notes
- generated build output, caches, and dependency directories

Task-bound plans and designs are stored outside the Git repository under
`C:\Users\droth\Documents\Codex\RailKeeper\work-notes`. Session handovers remain under
`C:\Users\droth\Documents\Codex\RailKeeper\session-handovers`.

The repository ignores the retired paths `docs/designs`, `docs/plans`, and `docs/superpowers` to
prevent accidental reintroduction. Stable decisions that remain relevant after implementation must
be rewritten as focused architecture, security, operator, or user documentation.

## Cleanup scope

The cleanup removes all tracked files from the three retired working-document directories. It keeps
the documentation site, release notes, screenshots, architecture documents, required JSON
manifests, package metadata, and validation scripts. The documentation test and production build
must pass after the cleanup.
