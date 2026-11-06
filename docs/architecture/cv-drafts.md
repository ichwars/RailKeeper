# Atomic CV changes and compatibility

CV identity remains **vehicle + exact trimmed decoder profile + CV number + nullable CV31/CV32 pair**.
An unindexed value and page 0/0 are different identities. Protocol, description, category and source
metadata describe that identity; they do not select a different CV. Both index fields are required
together for CVs 257–512.

`POST /vehicles/{id}/cv-values/changes` accepts 1–500 explicit changes, up to 2 MiB. Each existing ID
requires the revision originally read by its editor. New entries omit the ID and supply an empty
revision, requiring their identity to be absent. A deletion names an existing ID/revision and supplies
no value. Duplicate IDs/target keys, unknown fields, missing numeric values, invalid index pairs and
foreign source files are rejected. This endpoint stores local inventory records and does not program
hardware.

The application owns one SQLite transaction for the complete selection. Every row's current revision,
identity collision and source-file ownership are checked in that transaction. A later conflict,
validation error or SQL failure rolls back all preceding row and history changes. Moving an ID to
another profile/index identity never silently replaces its occupant. A separate explicit deletion is
required if that occupant is to be removed.

`vehicledata` executes CV row and history writes with the caller's transaction. The shared application
writer also serves legacy create/update, reviewed imports and stored ECoS CV adoption. Original IDs,
creation times and unrelated profiles/index pages remain intact. Value changes append the existing
history format. Omitted source line/export timestamp preserves existing provenance; an unknown source
timestamp stays unknown. Source export wall-clock time is never inferred from the save timestamp.

## Revisions and older callers

Stored CV responses add a computed SHA-256 `revision` over all persisted row fields, including profile,
index pair, metadata and provenance. It is not based solely on second-resolution clocks. History is
still returned separately. Revisions are calculated consistently for single reads and vehicle/list
responses and are not stored in the database or backup.

Existing create/update/import inputs may supply `expectedRevision`; omission retains their request
contract. An empty revision requires a missing identity, while a stored revision must match. Legacy
delete additionally accepts the optional `expectedRevision` query parameter. Collisions and stale
revisions return HTTP 409; a manual batch applies no subset. All routes retain Editor/Admin and CSRF
checks, and successful batch completion creates one audit event.

A repeated create-only batch cannot produce a second CV with the same identity. This is conflict
protection, not a durable request receipt: a lost successful response can require reloading and review
before a caller knows that its first request committed.

## Manual editor

The raw CV view stages additions, edits and deletions with **Stage in CV draft**. Its review compares
stored and pending identities, values and metadata; **Save CV draft** is separate from the vehicle's
bottom save. Discarding a draft or closing the editor does not submit it. A new-number form targeting
an existing identity preserves unspecified metadata; clearing a field deliberately uses that row's edit
action. The limit is 500 pending changes, with individual discard and a bounded review table.

A conflict retains all pending input and reloads current data without claiming a successful save.
**Use this CV as the new draft baseline** is an explicit decision: a three-way comparison keeps locally
changed fields while adopting newer fields which the draft did not edit. Occupied target identities
remain visible and require editing the pending identity or discarding that change. An externally deleted
row remains inspectable as its original draft baseline and is never implicitly recreated. A lost
response may leave values already stored; loading/reviewing the current state makes this visible.

Confirming a batch preserves further form input which was not staged. Independent vehicle refreshes
retain the CV draft; the CV commit retains pending vehicle/functions. Interpreted decoder values,
function suggestions and speed curves continue to use stored CVs until the CV draft commits. A
successful batch has a status result, while actual errors retain the editor's error presentation.

## Preservation evidence

Application and HTTP tests cover index page zero vs unindexed data, two profiles, source provenance,
original IDs, histories, backup/restore, stale revisions, identity moves, source ownership, a failure
in a later SQL write, explicit numeric zero, roles/CSRF and legacy calls with/without revision guards.
The regression matrix retains all original 200 routes/access declarations and adds only this endpoint.
Cross-version API comparison excludes only the computed CV revision; persisted tables/files remain
fully compared.
