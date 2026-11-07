# Decoder installations and recorded CV working states

Migration `0076_decoder_configuration_model.sql` adds five tables without changing existing
vehicle, CV, source-file or history rows. Migration `0077_decoder_legacy_workspaces.sql` seeds
exact legacy working groups. Startup then records their available evidence after file-blob migration.
No physical decoder is inferred from a profile name; authoring workflows follow separately.

## Authority and identity

`vehicle_cv_values` remains the sole mutable current CV representation, with its existing IDs,
profile/index identity, source provenance and value history. Additional working groups associate
one vehicle and its exact stored profile text. Legacy whitespace is not automatically normalized
or merged. New input continues to use the existing cleaning rules. A profile is a working label,
not evidence of physical hardware identity. Distinct working labels can represent separate
decoders in the same vehicle without changing the existing vehicle/profile API semantics.

| Table | Responsibility |
| --- | --- |
| `decoder_instances` | Explicitly identified physical decoder, including local label and optional manufacturer/model/serial details. |
| `decoder_installations` | One installation episode, its vehicle reference and separately known or unknown installation/removal dates. |
| `decoder_cv_workspaces` | Current vehicle/profile working group and an optional explicitly confirmed installation. |
| `decoder_cv_snapshots` | Immutable recorded partial working state, original CV IDs/history/provenance and frozen hardware description. |
| `decoder_cv_snapshot_sources` | Original source metadata and retained blob content, or explicit unavailable content. |

A vehicle can have multiple installed decoders. A decoder has at most one current installation.
Unknown dates remain SQL `NULL`; the recording timestamp is independent. Deleting a vehicle
retains its historical ID, name and inventory number on installation and snapshot records.
An active installation then becomes `unresolved`, without inventing a removal date. Current
working groups may disappear with the vehicle; recorded snapshots and their sources remain.

## Meaning of a snapshot

Every snapshot has scope `partial_working_state`. It records the available working values at its
recording time, not a complete measurement or a single source export. Individual CV source lines,
source export timestamps, index pages and prior value changes retain their existing meaning.
Unknown integer values from historical storage are retained; a snapshot does not turn them into
valid programming commands. Unindexed values and index page `0/0` stay distinct.

Payload and source rows reject updates. Normal inventory deletion does not cascade into historical
snapshots. There is no ordinary snapshot deletion command. A deliberate complete application
restore replaces them together with the rest of the restored collection.

Physical identity is optional and must be confirmed separately. An unassigned snapshot stores an
empty hardware object. An assigned snapshot captures decoder/installation labels and dates as they
were recorded, so later model corrections or removal do not rewrite the old description.

Original file metadata and content are separate. Frozen source references keep blobs alive after
the original CV-file record is deleted. An absent original can be described as `unavailable`;
its source ID is retained without fabricating bytes or a source date. All reads remain private.

## Backup, migration and rollback

Backup format 27 includes all five tables and retains the format-26 catalog rules. Preflight
rejects missing required tables, unknown model columns, invalid relationships, duplicate active
installations, invalid dates, malformed snapshot evidence and missing frozen blobs. It validates
historical vehicle references independently of currently existing inventory. Format numbers must
not be lowered to bypass these checks. Supported formats 1 through 26 remain readable.

Schema 76 introduces empty tables; schema 77 and initialization adopt existing CV/file groups.
Before schema migration, retain an app export, a consistent
full installation and its matching previous binary. Rehearse on an isolated copy and compare every
existing table/file plus the automatic pre-migration SQLite snapshot. Once new decoder data is
used, an older program or export format cannot represent it losslessly. Use a compatible repair
version or restore the verified complete old installation before any new input has been accepted.

Backend tests cover multiple decoders, nullable dates, raw profiles, CV IDs/history, source content
retention, immutable hardware evidence, inventory deletion, full restore and rejected malformed
backups. Native Windows exercises the same model cases before the existing end-to-end reference.

## Transactional adoption engine

The prepared capture engine reads exact vehicle/profile groups and records their current values,
individual provenance and real history in one transaction. It never matches hardware by model or
profile label. Repeated legacy adoption keeps existing snapshot identities and timestamps.
Transport revision tokens are excluded from historical evidence.

Source content is reused from existing blobs or copied from a confined source directory inside
the same transaction. An insertion failure rolls back source blobs, snapshots and new groups
together. Missing originals can reuse unambiguous archived bytes, including identical content
under different blob identities. Contradictory or unreadable originals remain unresolved.
Startup activates this engine before exposing the application. Restore completes missing legacy
groups and their evidence inside its existing transaction, before replacing uploads and committing.
Its source root is the newly staged backup, never the old installation. Missing originals stay
unavailable even if the old installation contains a file with the same name. Existing captured
snapshots and explicitly provided hardware assignments retain their identities and timestamps.

Local filesystem sources are confined to the upload directory using `os.Root`. Directory symlinks
cannot escape that boundary. Legacy Windows separators are normalized only when reading; stored
path text remains unchanged. Active filesystem references outside uploads fail backup preflight
before destructive restore. Canonical blob references do not reread obsolete filesystem paths.

Normal CV writes use the existing authoritative working rows. Manual create/update/delete,
reviewed batches, file/station imports and source-file create/delete maintain the affected
working groups in the same transaction. A profile move touches both exact stored groups without
changing their identity, adoption origin or confirmed installation. Newly introduced groups have
origin `manual` and no inferred hardware. Empty groups remain available for deliberate later work.

### Exact scoped CV writes

Clients may opt into `workspaceId` and `expectedWorkspaceRevision` when editing/importing CVs
or uploading their sources. The binding revision identifies the exact vehicle/group/profile and
its physical installation/state. It stays stable across ordinary CV edits, so multiple rows can
commit together. Assignment, removal or a changed group identity invalidate old confirmations.
Direct scoped writes/imports also require the row's `expectedRevision`; manual batches use the
existing enclosing change revision.

The server resolves the group inside the write transaction. The submitted profile must match its
trimmed label, while the stored profile retains its exact original text. Legacy `"A"` and `" A "`
therefore remain separate even in one batch/import. Mixed scoped/unscoped aliases to the same
actual CV identity are rejected before any writes. Confirmations are not mutated during import.
Requests without the optional scope keep their existing trimmed-profile semantics.

Scoped source uploads retain the selected group instead of inferring another profile from a
filename. Rejected uploads clean up newly unreferenced blobs. Import comparison overhead includes
both row and binding revisions, while the original 8 MiB business-data budget remains unchanged.

Ordinary writes do not create hidden full snapshots. Existing immutable evidence remains unchanged;
recording a checkpoint and confirming an installation are separate commands. Restoring an existing
workspace preserves its metadata and snapshots; older backups still receive legacy adoption.

## Explicit hardware commands

The application layer provides a consistent read context and explicit physical-instance,
installation, removal and date-correction commands, exposed through the private HTTP API.
Matching model names never merge instances. A second decoder in one vehicle is allowed; one
physical decoder cannot have two active installations. Removal keeps the installation episode,
its vehicle identity and prior notes. A later installation creates a new episode.

Unknown dates remain nullable even when recording or removing an installation today. Date
corrections validate real calendar dates and chronological order without modifying earlier
snapshot hardware descriptions. Existing imported instance identities remain editable; new
instances and installation episodes require distinct client-generated identities.

Commands compare a current context revision inside their transaction. That revision includes
physical-instance and installation records plus exact working CVs, histories and source-file
metadata, including changes made without advancing a timestamp. Failed comparisons and duplicate
create confirmations make no changes. Hardware changes and their audit record commit together.

## Deliberate checkpoints and assignment

Creating a named working group adds no CVs or hardware. Recording a checkpoint captures the exact
selected group, its individual CV provenance/history, and its available original sources in one
transaction. New checkpoint identities belong to the confirmation request; replaying an old
confirmation or reusing an existing identity is rejected without duplicates.

Assigning or clearing hardware is a separate command. Before changing the current association,
it records the previous partial working state with its previous hardware description. A new
association must refer to an active installation in the same vehicle. It neither copies nor
retags CVs and does not change existing snapshots. Creating a separate named group is appropriate
when starting work for a replacement decoder rather than reinterpreting old working values.

The read layer exposes archived summaries, complete checkpoint contents and private original-source
downloads independently of current vehicle/workspace records. Unavailable originals return an
explicit unavailable result; a path or a different live file is never substituted. Source metadata
and compressed original bytes are read together from the database. Archive readers expose no
internal storage paths or blob identifiers. Checkpoint/association changes and their audit records
commit together, including any newly captured blobs.

## Private read API

The vehicle decoder-context route and the global snapshot archive require an inventory role
(Viewer, Editor or Admin). Messe has no access. Snapshot details and retained-source downloads
use their historical identities directly and remain readable after vehicle deletion. Downloads
serve only the retained blob of the requested snapshot/source pair as an attachment. Missing
evidence returns `410 decoder_source_unavailable`; an unknown pair returns `404`.

OpenAPI and generated transport types distinguish recorded CV values from writable CV input.
Historical unknown integers, exact profile text, source timestamps and index-page absence are
preserved without presenting those records as valid programming commands. Read responses contain
no internal storage paths or blob identifiers. Decoder writes use separate explicit commands.

## Confirmed write API

Instance, installation, date, workspace, association and checkpoint routes require Editor or
Admin, the existing CSRF protection and a current context revision. Every documented edit field
is required, including empty optional text and explicit null dates/associations. Unknown fields,
trailing JSON, invalid null text and requests exceeding 64 KiB are rejected before application
commands run. Instance edits retain their ID and must match the path ID.

Clients retain new instance/episode/group/checkpoint identities with the confirmation request.
A lost reply must not trigger an automatic new identity or a modified replay. HTTP 409 requires
reloading the saved context and reviewing the intended change. Checkpoints contain only saved
working rows; open browser drafts are excluded. These inventory commands do not write to devices.

## Working-group comparisons for original rows

CV and source-file DELETE requests accept optional `workspaceId` and
`expectedWorkspaceRevision` query parameters together with the existing `expectedRevision`.
Both comparisons are required when selecting a working group. The exact stored profile must
match that group; a trimmed alias is insufficient. Missing pairs, repeated query fields,
foreign groups and invalid comparisons are rejected. A changed assignment or installation
state yields HTTP 409 before deletion, even if the CV or file revision itself stayed unchanged.

For a manual CV batch, the outer scope fields compare the original row, for both deletion and
value updates. New rows cannot have an original-group comparison. Value fields inside `value`
select the destination group. A scoped profile move also requires the original-group comparison.
Direct PUT requests carry that comparison in the same three query parameters; the query row
revision must equal the JSON row revision. This checks both hardware bindings when moving a row.
Source comparisons are internal application fields, not import-file data or writable JSON fields.

Every comparison and mutation uses the existing transaction, so a later conflict rolls back
earlier edits in the same batch. Legacy callers
that omit the optional scope retain their contract. Confirmed deletion preserves recorded
partial states and any original blobs retained by those states.
