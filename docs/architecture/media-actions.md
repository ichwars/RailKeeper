# Vehicle and media revision guards

Vehicle detail responses expose a computed `revision` for core fields and set context, and
`imagesRevision` for the complete image collection. Image and attachment records expose their own
`revision`. These tokens hash typed stored content, including metadata and relevant private file
references, rather than relying on the second-resolution modification clock. Tokens reveal no file
paths or blob identifiers. They are computed without schema or backup-format changes.

## Compatible write requests

The vehicle PUT accepts optional `expectedRevision` and `expectedImagesRevision`. The latter is checked
only when `images` is present: omission preserves images, while an explicit empty array requests their
removal. The server reads and checks the current core and image snapshots in the same transaction as
core fields, image metadata, function changes, inventory-number history and audit. A mismatch returns
409 (`vehicle_conflict` or `vehicle_image_conflict`) and commits none of those changes. Core data and
its response token come from one transaction snapshot; separately loaded images carry their own
matching collection token.

Attachment PUT accepts optional `expectedRevision`. Independent image and attachment DELETE accept
an optional `expectedRevision` query parameter. Current metadata, revision checks, maintenance
ownership checks and mutation share a transaction. A stale token returns 409 (`attachment_conflict`
or `image_conflict`); malformed tokens return 400. A maintenance-linked image remains protected.
Requests omitting the optional guards retain the legacy contract. Roles, CSRF and file limits remain
server-side requirements. The UI must retain the original token with a dirty draft and offer an
explicit comparison before rebasing it; a refresh alone must not silently authorize stale fields.

## Editor comparison

The current editor supplies both observed vehicle tokens on its vehicle save and retains pending core,
function, image and attachment fields when a conflict response refreshes the saved record. Core fields
use the existing three-way comparison. Image titles, source and maintenance links and attachment notes,
categories and links are compared separately; competing primary selections and image orders are also
reviewed. Untouched fields and new uploads merge without erasing pending fields. Comparisons follow
current typing and do not display private hashes or maintenance IDs.

Conflicting image fields block the vehicle save until a user explicitly keeps the draft or adopts the
stored fields. Attachment conflicts block only that attachment's separate save. Pending fields for a
record deleted elsewhere remain visible until explicitly discarded; they never recreate the removed
media record. An exhibition flag action first reads the current complete vehicle, sends its core
revision and omits images. Missing comparison tokens block the editor write and prompt a reload.

## File completion and cleanup

Uploads store new payloads under fresh blob identifiers before creating their metadata reference.
Existing file references are replaced only inside a successful metadata transaction. Failed remote
image localization/adoption may leave newly allocated blobs unreferenced; create/update clean those
up after rollback or commit using a bounded cleanup context that survives request cancellation.
Every blob cleanup uses one conditional DELETE checking all image/thumbnail, attachment, CV-file,
accessory-document and dedicated-set-image references. Referenced payloads remain stored.

Legacy filesystem payloads are removed after their metadata deletion succeeds, and only after a
global check of image, thumbnail, attachment and CV-file paths. The check also recognizes normalized
path aliases. A failed reference check leaves the file intact. All filesystem paths still pass the
configured data-directory confinement check. Filesystem removal is a separate, best-effort operation;
the SQL transaction does not make it atomic. Current uploads use database blobs and do not create a
filesystem staging area.

Media actions remain independent from the vehicle/CV drafts. A future shared media draft would first
need owned temporary staging with explicit size/lifetime limits and recovery. These revision guards
do not introduce that staging, batch-upload atomicity, or request-retry idempotency.

## Verification

Transaction tests cover unchanged clocks, stale core/image/attachment drafts, omission versus explicit
image clearing, matching response tokens, full rollback and cleanup that preserves shared blobs.
HTTP tests exercise conflicts, malformed guards, CSRF and shared legacy file preservation through the
last reference. Cross-version API comparison ignores only these new computed fields on their exact
record types; raw database, backup and filesystem comparisons remain complete.

The target-browser regression uses a separate authenticated writer to cause real core, image and
attachment conflicts, verifies rejected requests leave the complete aggregate unchanged, resolves the
reviewed fields, saves with current revisions and compares every unrelated field and original file byte.
Both languages, four viewport widths, both themes, keyboard focus and mobile target sizes are covered.
