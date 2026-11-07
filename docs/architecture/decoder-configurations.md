# Decoder installations and recorded CV working states

Migration `0076_decoder_configuration_model.sql` adds five tables without changing existing
vehicle, CV, source-file or history rows. This schema/backup package prepares later adoption and
commands; it does not yet add an inventory workflow or infer a decoder from a profile name.

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

The five tables are initially empty. Before schema migration, retain an app export, a consistent
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
The engine is prepared separately from its startup/restore activation; normal CV writes and
the existing application workflows remain unchanged in this preparation package.
