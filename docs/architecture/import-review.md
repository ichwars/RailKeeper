# Import comparison and selection

`ImportReviewContext` names the actual source, target and effect of a review. `ImportReviewStatus`
uses one vocabulary: new, different, identical, unclear and excluded. Status text remains visible
alongside icons. Comparison, selection and readiness are separate concepts; an unselected different
value remains different, and a ready action does not prove two source records identical.

The existing specialized parsers, supported formats, validation, selection keys and write paths remain
in their owning features. The shared components perform no requests and hold no adoption state.

| Review | Mapping and retained scope |
| --- | --- |
| CV files | Existing `new`, `changed` and `same` comparison is retained. Invalid/duplicate identities are excluded. Structured reason and changed-field keys localize explanations; legacy messages remain available to old local consumers. Filename, selected count, profile and nullable index identity remain visible. Original-file upload and function adoption stay separate. |
| Article search | Existing empty/equal/different comparison is retained. Conflicting external fields are unclear; protected fields are excluded. Source URL, score, trace and specialized field/image selectors remain. Vehicle and accessory drafts identify their respective targets. Only the later record save persists the draft. |
| CSV/JSON transfer | Source-field adoption is new for create/copy, different for replace, excluded for invalid/skipped input or reuse of an existing target, and unclear for unresolved review. Existing readiness severity and all conflict decisions remain alongside it. Excluded source fields do not disable the existing reuse/link action and do not claim the records identical. The job names its source and local inventory target. |
| Command-station comparison | Existing equal/deviation/new classifications map directly; missing or conflicting observations remain unclear. The context explicitly describes a read comparison. Existing hardware previews, grants, expiry checks, supported capabilities and explicit write consent remain separate. |

These components do not implement request idempotency, change transaction boundaries or authorize
device writes. Reviewed import jobs, CV identity/source checks and device-specific confirmation remain
the authority for writes. Article search keeps its existing synchronous loading state.

## Confirmed CV imports

`useVehicleCVImportController` owns file comparison, selection and the outcome of one CV import.
It sends the revisions observed in the preview, including an empty revision for an identity observed
absent. A revision supplied by a file never authorizes a write. The existing import transaction checks
every selected identity and commits values and histories together. Unselected rows are not submitted.
The 10,000-row import limit remains separate from the smaller manual editing draft.

A conflict or uncertain response retains the selection and blocks another write until the user
compares again. That comparison only reads current values. It retains originally selected differing
rows and deselects identical results, including values committed before a lost response. A confirmed
save clears its selection before refreshing the vehicle, so a refresh failure cannot repeat the save.
Other open vehicle, function and manual CV drafts survive both comparison and import.

The HTTP business-data budget remains 8 MiB. A bounded allowance for the existing optional comparison
revisions prevents their metadata from reducing that budget. This changes neither accepted file
formats nor file-upload limits. Legacy clients may still omit revisions; the reviewed frontend always
supplies them. No additional persistence or backup format is introduced.

## Import job completion

CSV/JSON confirmation retains its existing job, source hash, revision, target fingerprints and single
transaction. The same job cannot apply twice, including concurrent confirmations. An intentional new
import creates a distinct job through the existing workflow.

After a conflicting or uncertain completion response, `useTransferImportConfirmation` reads that job.
A completed job is shown as completed and cannot be submitted, cancelled or uploaded again from the
dialog. An unavailable status blocks mutations and offers a read-only status check. A still open job
requires a fresh upload and full review, preserving the existing stale-preview policy. Recovery never
creates a new job or resends the confirmation. History remains the authority for its actual result.

## Command-station vehicle adoption

Creating a vehicle from an ECoS or CS3 draft submits its selected source mapping in the same request.
The existing unique provider/object identity is bound in the core vehicle transaction. An already
assigned source rejects the complete creation, including function changes, audit and inventory-number
allocation. A retry after a lost creation reply therefore cannot create an unlinked duplicate. The
dialog retains its draft, reloads inventory on a mapping conflict and requests review of the existing
assignment. Ordinary vehicle creation remains independent when no mapping is supplied.

Mapping metadata can still be refreshed for the same vehicle. The owner condition is checked inside
the SQL conflict branch, and the returned mapping comes from that statement. A competing assignment
cannot silently move the source to another vehicle. Existing explicit same-vehicle rebind behavior
and the separate station CV adoption remain. No device write, new persistence or backup format is
introduced; mappings already participate in application backups.

## Article suggestion scope

Vehicle and accessory search controllers bind pending responses to their current editor target and
request generation. Closing, changing targets or starting another search invalidates older replies;
they cannot finish the current loading state or populate another draft. Search remains synchronous
with its existing single loading state and specialized eligibility rules.

Field confirmation pins the draft snapshot used to initialise its selection or observed at an explicit
checkbox selection. If selected target fields change, the complete adoption is rejected, those fields are
deselected and the dialog asks for review against their current values. No fields or images are
applied until another confirmation. Each result keeps its actual selection index even when URLs
repeat, and one confirmation is consumed once. Saving the record remains a separate guarded action.

Field checkboxes have explicit names. The bounded comparison table wraps complete values; on narrow
screens it scrolls separately with a visible hint and keyboard access. Its primary adoption action
uses theme tokens with text contrast of at least 4.5:1 and mobile selection targets of at least 44 px.
