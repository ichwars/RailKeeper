# Vehicle save boundaries

The vehicle editor contains both a shared draft and independently saved records. Its save guidance
names the current tab's boundary in German and English. **Cancel** discards pending edits and does not
undo a completed independent action. View mode remains read-only; all writes retain server-side roles
and CSRF checks.

| Path | When it writes | Affected records and cancellation |
| --- | --- | --- |
| Model/control fields | Vehicle save | Core vehicle fields are saved together; cancel discards pending fields. |
| Function edits, deletion, JSON import | Vehicle save | `functionChanges` shares the core transaction. Original revisions detect conflicting function changes; failed validation/conflicts retain the draft. |
| CV/library and command-station function suggestions | Vehicle save | Protected draft fields and source evidence remain in `functionDraft`; suggestions are reviewable before saving. |
| Decoder-file function preview | Vehicle save | The explicit preview action supplements missing fields in the same function draft. Assigned fields, intentional edits/deletions, nullable flags, modes and original revisions stay protected. Contradictory file suggestions are not silently selected. Source filenames remain separate from editable own notes. |
| Image titles, order, primary image, maintenance link | Vehicle save | Existing image metadata shares the core transaction. Pending remote images import afterwards. Each successful import becomes a persisted draft item immediately. |
| Local image/attachment upload | File selection/drop | These are independent uploads. A successful upload remains after cancel. The result is shown after refreshing the vehicle. |
| Persisted image deletion | Its delete action | Independent deletion; maintenance-linked images remain protected. Removing an unsaved remote image only changes the draft. |
| Attachment metadata | The attachment's save action | Independent metadata change. Typing into its fields alone does not save, and the vehicle's bottom save does not submit it. |
| Attachment deletion | Explicit delete confirmation | Independent deletion; canceling the confirmation changes nothing. |
| Maintenance and spare parts | Their own save/delete/import actions | Independent workshop records. In-progress workshop forms survive other aggregate refreshes. |
| Manual CV values | Save CV draft | Adding, editing and removing values stages up to 500 explicit changes. The separate CV save commits them atomically with revision/source/identity guards; cancel or discard drops the pending changes. Unsubmitted form input and the vehicle/function draft survive the CV save. |
| CV value/file imports | Reviewed adoption/upload action | Confirmed imports/files are independent records. File-preview preparation and selection alone do not write CV values. |
| Web-document adoption | Confirmed document import | Independent attachments, separate from the vehicle draft. |
| Local command-station adoption | After vehicle save | Mapping and CV adoption follow the acknowledged core save. These operations change RailKeeper records, not physical devices. |
| Hardware actions | Separate device-specific confirmation | Existing capability, grant and device confirmation boundaries remain in place. Closing the vehicle editor does not reverse hardware actions. |

## Refreshing an open editor

`useVehicleEditorController` distinguishes initialization, an independent refresh and acknowledgment of
a core save. A refresh compares the current draft, its previous saved baseline and the new response.
Untouched core fields adopt the new value; locally edited fields retain their draft. If both sides changed
the same field differently, the editor shows both values and requires a choice before its next core save.
The conflict persists across later refreshes until resolved. Responses from closed/superseded editors
are ignored.

`useVehicleFunctionsController.loadDetail` retains dirty same-vehicle edits with their original revisions.
The media controller reconciles existing image metadata and attachment drafts against refreshed records.
It does not reset pending remote images, CV form input or workshop forms. An acknowledgment copies only
independent side records missing from a core-write response; omitted optional core fields still clear their
old values. The final saved detail is authoritative for submitted image metadata.

## Failure and retry boundaries

The core save is acknowledged before remote image and local mapping/CV adoption. An ancillary error
therefore leaves the editor on the already created vehicle. Successful remote image imports replace
temporary draft IDs immediately, so retrying after a later image failed includes those stored IDs rather
than importing the same image again. A browser response lost before acknowledgment remains a separate
retry concern; acknowledgment alone is not server-side idempotency.

Core-field reconciliation detects changes observed during refresh. The API additionally supports
optional [content revision guards](media-actions.md) for core fields, image metadata and independent
image/attachment operations. Clients retain the original token with a dirty draft and must explicitly
resolve conflicts before rebasing it. Function revision conflicts remain enforced in the shared server
transaction. File operations and separately confirmed actions are not made atomic by the core SQL
transaction.

## Regression evidence

Controller tests cover late responses, typing during refresh, persistent conflicts, cleared optional
fields, media metadata and partial image-import retries. `tools/regression/editor-boundaries.mjs` runs
against the **target** binary/image after a synthetic cross-installation restore. In both languages it
keeps core, function, image, attachment and CV-form drafts open while uploading an attachment, checks their
retention, cancels, then verifies through the real API that only the confirmed upload was stored. It also
reopens the editor and inspects the save guidance at all supported widths in both themes.
`cv-draft.mjs` additionally tests manual staging/cancel, an actual concurrent writer, rollback of an
earlier valid row when a later row conflicts, explicit rebase, saved history/provenance, busy controls
and retained vehicle/function drafts. CV interpretation continues to use stored values until commit.
