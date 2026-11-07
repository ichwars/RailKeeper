# Regression baseline

The baseline protects existing RailKeeper data and workflows during incremental changes. It uses only
new synthetic installations. It never accepts a production URL, an existing data directory, or a private backup.
Go and frontend tests remain required; this baseline adds process, restore and browser checks.

On the managed `codex-dev` host, start a complete test command with `codex-test-tmp run --`.
Temporary copies and browser profiles use its `TMPDIR` on the data disk and are removed afterwards.
Retain reports and other required evidence in the project's ignored `output/` directory before cleanup.
Other hosts use their own temporary-directory configuration; this launcher is not a RailKeeper dependency.

## Feature matrix

`tools/regression/matrix.py` inventories all registered HTTP methods, paths and access declarations.
Its original reviewed digest froze 200 routes at commit `2cdfd4430ffb28d0d726d2de6cb67bc43b329946`.
The current reviewed digest covers 222 routes, including the additive manual CV draft, catalog, decoder
configuration and manual valuation endpoints. Original methods, paths and access declarations remain protected.
An intentional contract change requires reviewing the matrix and updating the digest. The script verifies
that evidence source files exist. This source inventory does not establish behavioral coverage of every route.

| Area | UI and workflow | Persistence and existing evidence |
| --- | --- | --- |
| Setup/accounts | First run, login, roles, TOTP, sessions | Local authentication; API smoke and auth role tests |
| System/profile | Settings, storage, updates, SMTP, audit, printers | Installation settings; storage/version tests |
| Bug reports | Text, optional diagnostics, screenshot/contact | Local submission state; route/delivery tests with simulated receivers |
| Vehicles/sets | Inventory, individual copies, sets, numbering, reports, QR | Vehicle/set identities and references; set/report tests |
| Media | Images, attachments, preview/download, maintenance links | Blob/file contents and metadata; vehicle/backup tests |
| Workshop | Maintenance, condition, spare parts | Maintenance and spare-part histories; vehicle tests |
| Functions/CVs | Function draft, flags, index pages, decoder library, speed curve | Functions/CV provenance/history; reference evaluator and draft tests |
| Decoder inventory/archive | Individual units, installation episodes, explicit working targets, recorded partial states | Frozen hardware/source evidence; binding conflicts and legacy adoption tests |
| Accessories | Catalogue, quantity/individual stock, documents, reservation/installation | Stock, assets, allocations, histories; API lifecycle tests |
| Exhibition | Entries, conflicts, locking, print | Lists, entries, exceptions; workspace tests |
| Transfer/search | Formats, profiles, preview, conflict review, jobs; article search | Profiles and transfer data; import/export/search tests |
| Master data | Entries, relations, activation, symbols, import/export | Master-data tables; API and seed tests |
| Overview | Inventory, maintenance, valuation | Derived views; valuation/frontend tests |
| Command stations | Read/compare, grants, image transfer, backup import | Mappings/CVs plus local workspace state; simulated protocol tests |
| Backup/restore | Export, compatibility preflight, confirmed restore | Explicit app/installation boundary; backup/API tests |
| Legacy archive | Dormant planning records remain restorable | Retirement and legacy-backup compatibility tests |

The generated matrix lists every route's actual access declaration and the owning test paths. Admin,
Editor, Viewer and Messe boundaries continue to be checked server-side. Public health/setup/login routes
remain distinct from authenticated routes.

```sh
python3 tools/regression/matrix.py
python3 -m unittest discover -s tools/regression -p 'test_*.py'
```

The [quality gates](quality-gates.md) define the fixed frontend/backend measurement scopes,
compiler/minimum-version policy and recurring race/fuzz checks. Those automated results
remain separate from the preservation and platform qualification below.

## Synthetic reference collection and process checks

The reference contains two separate vehicles with the same model article, a two-member set with individual
notes, an image linked to maintenance, an attachment, a decoder source file, functions with nullable flags,
CV values with history and source timestamps, separate profiles, and indexed/unindexed CV257 including page 0/0.
It also contains hierarchical storage, quantity stock, an individual asset, reservation/installation records,
an accessory document, an exhibition assignment, a command-station mapping and a completed transfer export
with its stored file. A batch CV import exercises the real import path as well as individual CV edits.
No physical station is configured or contacted. Test passwords are random and remain in process memory.

Build the existing application and install the pinned browser test tool:

```sh
cd backend
go build -o ../output/regression/railkeeper ./cmd/railkeeper
cd ../frontend
npm ci
npm run build
cd ../tools/regression
npm ci
npx playwright install chromium
cd ../..
python3 tools/regression/run_baseline.py \
  --binary output/regression/railkeeper --browser --output output/regression-run
```

On a minimal Linux system, Chromium also needs its system libraries and font configuration. Install those
through the documented Playwright host setup, or use an already prepared test environment. Windows uses
`python`, `npm.cmd` and a `.exe` binary. Reports, screenshots, databases and synthetic exports remain under
ignored `output/`; they are not committed or added to the distribution package. The output directory must
be new, so reruns cannot overwrite an earlier reference or an existing installation.

For an upgrade comparison, the source binary prepares the reference and verifies its original export
and restore behavior. Both browser suites run on the current target binary/image after the fresh app
restore, using its own account. This also exercises new write contracts on the matching frontend and
backend versions while preserving the complete old-to-new database, API and file comparisons.

The Windows standalone workflow also runs focused native Go media tests through its package builder
before packaging. These cover
shared legacy paths (including case aliases on Windows), rejected-upload cleanup, committed image
revision receipts, CV source-link guards, retained historical provenance and canceled-request cleanup.

The runner exercises:

1. Real API creation of the reference; authentication, all four roles and CSRF boundaries.
2. Optional actual browser login, all seven main workspaces, edit/cancel/save/reopen in DE/EN, and
   inventory rendering at 2580/1440/820/390 px in light/dark themes without page-wide overflow.
   A full export before/after editing also verifies that unrelated data is untouched. Only edit clocks
   of the edited vehicle and its submitted images may change; image provenance remains exact.
3. App export, compatibility preflight and byte-exact image/attachment downloads.
4. Rejection of a future backup without changing application data.
5. In-place app restore that removes a post-backup marker and restores the complete saved collection.
6. App restore into a second installation while retaining its own account, roles and settings.
   With `--browser`, the target binary/image also exercises a confirmed upload while core/function/image/
   attachment/CV-form drafts are open. It verifies that refresh preserves the drafts, cancel discards them,
   and only the independently confirmed upload reaches storage, in both languages and themes.
   Manual CV browser cases then stage/cancel without a write, exercise an actual second-session edit,
   verify complete rollback on a later stale row, explicitly resolve it and confirm the whole selection.
   Unselected CVs, provenance, history and the unsaved vehicle/function drafts remain checked; CV controls
   are also checked for mobile bounds, target size, keyboard focus and a pending request.
   Decoder cases distinguish identical models, unknown physical dates and immutable recorded states.
   Working-target cases confirm the default explicitly, preserve staged/open/import/file targets after
   another selection, reject a concurrent hardware change and compare again without an automatic write.
   Original file bytes survive a rejected upload and its separately confirmed retry. Deletion uses
   the file's original group, independently of the new-input default.
7. Stopping the source, copying its complete data directory and the test launcher's external upload-limit
   configuration, verifying all tables/files, restoring into another directory, restarting with that
   configuration, logging in with the original account and rechecking collection/valuation/artifact downloads.
8. On Windows, standalone startup using an isolated `LOCALAPPDATA/RailKeeper/data`, never the real user's default.

The exported application tables, profiles and files are compared by full content, retaining IDs and timestamps.
For a fresh installation, two existing restore policies are applied explicitly to the expected result:
the target's bundled workshop symbols remain authoritative, and a profile creator absent locally is cleared.
All other fields must match. These rules do not import users. A full installation copy preserves those users.

The cross-version API observations omit only the new computed `revision` field on stored CV records,
which older binaries do not emit. All persisted CV fields, IDs, index pages, sources and histories remain
compared. Transaction/HTTP tests independently verify revision conflicts. Raw table/file comparisons do
not omit any CV data and the backup format does not store computed revisions.

The read-only SQLite manifest hashes sorted full rows and column metadata, checks integrity and foreign keys,
and distinguishes NULL, numeric zero, text and blobs. File contents use SHA-256. Reordering rows is harmless;
equal counts with changed data fail. DB files and WAL/SHM are compared as logical SQLite contents rather than
physical page bytes. All external files are compared. The source is stopped before the full installation copy.
This is a cold restore exercise, not a tool for live production backups.

## Upgrade and platform evidence

Without `--previous-binary`, the process check is a restart/restore test. An actual upgrade additionally requires
a different previous application and its matching migration/seed directory:

The [complete upgrade command](modernization-rollout.md#einen-eindeutigen-kandidaten-qualifizieren)
builds the previous application separately and retains its matching migrations for the whole test run.
It places temporary sources under `TMPDIR` and writes lasting evidence into the project.

The baseline release source is pinned to `debfad6198299052f07220883c1476581be691cc` (v0.1.22).
The previous binary is built from that revision separately; the current checkout is not reset.
The optional `--expected-commit FULL_SHA` requires the actual running process to report that
clean source identity, compiler and platform through its local version endpoint. Reports retain
that identity independently of update-feed status. Native Windows packaging also checks the
resource-patched executable's `version` command and its compiler against `.go-version`.
The valuation process test abruptly kills only its own isolated child process, restarts the
same synthetic data, validates the existing session, and compares every historical valuation
and application row. Controlled SQLite FULL and staged-file errors separately prove restore
rollback retains application rows, original upload bytes and local authentication.

The report distinguishes upgrade from restart and records binary/image identities.
Application code, migrations and seeds of both versions must be available with that exercise.
The application JSON backup intentionally excludes job/workspace operational history and installation settings;
the complete data-directory copy checks those records too. A real installation additionally needs its external
configuration and matching software/image as described in the [server restore guide](site/administration/server-restore.md).

An existing image may be exercised with `--image IMAGE` on a **rootless local Docker daemon**. The runner
creates only randomly named test containers and binds only its new data directories. Container root maps to
the invoking host user so the test can compare files; this is separate from the normal unprivileged-image
startup smoke test. Each container is removed on success or failure. Existing containers/volumes are untouched.
Native ARM64 behavior still requires a native ARM64 runner; emulation is not equivalent evidence.

The Windows packaging CI invokes `tools/test_windows_baseline.ps1` after building and checking the package.
It runs the real packaged EXE, the synthetic restore suite and Chromium on Windows, with a separately built
v0.1.22 source binary for the upgrade exercise. Test output stays outside the package. Local packaging keeps
its previous prerequisites; this extended CI check requires the runner's Python and browser test dependencies.

Preserve each report with its tested revision and environment. A passing source inventory or browser screenshot
does not prove physical hardware, SMTP delivery, camera/scanner access, native printing or human usability.
Existing device/firmware acceptance limits remain separate. A loss of data, media, provenance, role separation
or an existing workflow blocks the affected modernization step.

The [modernization rollout guide](modernization-rollout.md) defines the human pilot tasks,
candidate identification, full-installation recovery and the boundary before a production rollout.
