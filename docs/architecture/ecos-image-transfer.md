# ECoS image upload and assignment

The existing image preparation dialog offers an additional Admin-only flow for saved vehicles:
read the configured ECoS roster, select a locomotive, preview its current and proposed image,
acknowledge the target, then explicitly confirm upload and assignment. A saved ECoS external
mapping preselects the object; addresses alone never imply a match. Download remains independent.
Vehicle edits, local original images, CV values and external mappings are not modified by a transfer.

## Protocol evidence

- The original ESU *Netzwerkspezifikation für das PC-Interface*, third edition, June 2011,
  section 7.11, page 8, documents `get(id, locodesc)` and
  `set(id, locodesc[imagetype, imageindex])`. The original is available through the
  [ESU download directory](https://www.esu.eu/en/downloads/instruction-manuals/digital-systems/).
  The reviewed public copy is pinned at
  [RailEssentials revision 6ae221d](https://github.com/cbries/railessentials/blob/6ae221d51865a6ce2e32ed97f9c991e64f7b5b8a/ecoslibNet48/Documentation/ecos_pc_interface3.pdf).
- Section 6.6 documents error 25 (NERROR_NOCONTROL) when another controller owns a
  locomotive, including a locomotive selected on an ECoS driving display. A separately
  authorized single hardware probe confirmed `request(id, control)` returning
  `<END 25 (couldn't get control)>` on firmware 4.3.3. No assignment command followed.
  RailKeeper reports `control_busy` and asks the operator to deselect the locomotive on
  both driving displays and other control applications. It never adds `force`.
- Read-only checks on ECoS firmware 4.3.3 confirmed descriptors for both internal images
  (`LOCO_TYPE_DIESEL,IMAGE_TYPE_INT,118`) and a user image
  (`LOCO_TYPE_MISC,IMAGE_TYPE_USER,0`). Deprecated `symbol` is not used.
- The same firmware's web interface exposes the user inventory at
  `/fcgi/ecos-locoimages?type=user`, the free-slot form at `/fcgi/ecos-uploadimage`,
  and multipart POST `/fcgi/ecos-douploadimage`. Fields are `desc`, `type` (`s/d/e/m`),
  `uploaded_file`, `lang`, `id`, `start`, `items`, `submit`. Image reads use
  `/loco/image?type=user|internal&index=N` (the inventory calls internal images `system`).
  The adapter validates this form and complete
  50-slot pages before writing; changed/login pages fail closed. Only advertised pagination
  is read, bounded to 500 image slots. These web endpoints are observed firmware behavior,
  not a vendor guarantee for every firmware version.
- The multipart request follows the observed form order (`desc`, `type`, file, then hidden
  fields and submit), with fixed filename `railkeeper.bmp`, `Content-Type: image/bmp` and a
  known request length. A generic binary file type and unordered fields are avoided for
  compatibility with the embedded upload handler. This is browser-form compatibility,
  not evidence that a particular firmware rejects every other valid multipart request.
- The existing [ESU format evidence](locomotive-image-export.md) supplies 190 × 40 RGB24 BMP.
  The exact output is 22,934 bytes, with a 40-byte Windows DIB header and 572-byte aligned rows.

Sources and read-only device behavior were checked on 22 October 2026. No manufacturer PDF,
vendor image or firmware is bundled. The user previously confirmed manual export/import.
The user confirmed direct upload and automatic assignment on firmware 4.3.3. The subsequent
pre-upload control reservation and duplicate detection have simulated-device coverage;
their revised order still needs hardware acceptance. Older firmware may require a restart before a new image
can be assigned. RailKeeper never restarts the station automatically.

## Boundaries and failure behavior

- Three vehicle-scoped endpoints provide options, multipart preview and JSON confirmation.
  All require Admin; both POSTs require existing CSRF protection. Only the configured active
  ECoS is used. Requests cannot choose a host, port, URL or different image at confirmation.
- HTTP uses port 80, no environment proxy, no redirects, bounded bodies and timeouts. Hostname
  resolution must contain only private LAN addresses; one address is pinned for HTTP and TCP.
  Changed settings or resolution invalidate confirmation. Loopback, link-local, public addresses
  and URL syntax are rejected. A web password/login page is unsupported and fails closed.
- A cryptographically random, five-minute, single-use grant binds actor, saved vehicle, source
  bytes/hash, target name/address/protocol/descriptor, old image, free slot and configured center.
  The store holds at most 64 grants, one per actor; a new preview invalidates that actor's
  preceding preview. Restart discards all grants. Neither grants nor image buffers enter backups.
- A locomotive can retain a `USER` image reference after that image was deleted on the
  station. A complete inventory reporting that referenced slot free is represented as
  `currentImageMissing: true` with an empty `currentImage`. Preview shows a specific
  warning and still allows a reviewed replacement. Confirmation and the control-held
  pre-upload check verify that the slot remains empty; refilling it invalidates preview.
  A 404, unreadable image or failed inventory alone never proves absence and never
  permits this fallback. This stale reference was observed on firmware 4.3.3 while the
  web interface and upload form remained reachable.
- API errors distinguish image content/read failures (`ecos_image_read`), incomplete
  inventory checks (`ecos_image_inventory`), unsupported/unreadable upload forms
  (`ecos_image_form`) and connection/interface checks (`ecos_image_web`). UI text identifies
  the failed preparation step instead of treating every image failure as an unreachable
  HTTP service. None of these preview errors start an image upload or assignment.
- The existing workspace operation mutex serializes transfers and workspace operations. Before
  the first write, RailKeeper consumes the grant, pauses passive monitoring, rechecks identity,
  descriptor, current image and free slot, and records audit intent. An unavailable audit store
  prevents writing. Monitoring resumes with its interruption callback; restart failure is shown.
- New uploads use only a free image slot. The form and slot are checked again immediately before POST.
  Existing images are never intentionally overwritten or deleted. The ECoS has no atomic slot
  reservation, so another operator must not edit image slots during transfer; a race after the
  final check cannot be eliminated by this client. There is no forced locomotive control.
- Preview reads every advertised inventory page (at most ten pages/500 slots) and compares
  occupied images pixel by pixel before choosing a free slot. An identical image is reused,
  including one beyond the first free slot or uploaded before a browser/application restart.
  Existing descriptions are retained. An unreadable image or incomplete inventory blocks
  preparation; it must not silently permit another upload. Reads are sequential, bounded by
  the preview deadline and stop once a verified match is found. This requires no mapping DB
  or source bytes in persistent storage. Confirmation rechecks the selected image/free slot.
- After explicit confirmation and audit intent, RailKeeper requests locomotive control
  **before** uploading. The same TCP connection stays open through the HTTP upload,
  image verification, `locodesc` assignment and descriptor readback. Control is released
  on every exit, and the socket is closed at the operation deadline. A failed control
  request (`control_busy` for error 25, otherwise `upload_blocked`) sends no image POST
  and no assignment. Target identity/descriptor are rechecked after acquisition.
  An already verified, correctly assigned image needs no control request or upload.
- Uploaded slots must appear occupied in the refreshed inventory, and images are read back and
  compared pixel by pixel. Exact RGB pixels or an exact
  supported RGB332 palette conversion are accepted; there is no general similarity threshold.
  Firmware 4.3.3 was also verified with integer Floyd–Steinberg diffusion: top-down,
  left-to-right, nearest palette colour with lower-level tie breaking, clamping after an
  arithmetic shift of accumulated error by four bits, and weights 7/3/5/1. Exact
  RGB(182,182,182) background pixels discard incoming error and do not propagate error.
  This conversion requires the observed 8×8×4 palette (red/green 0,36,73,109,146,182,219,255;
  blue 0,73,182,255). All 7,600 pixels of two stored device images matched the original BMP,
  whose hash matched the upload audit record. Synthetic reference vectors cover rounding,
  background reset, image boundaries and rejection of changed or displaced images; private
  vehicle images are not bundled as fixtures. This verifies stored image content, not a
  completed locomotive assignment. Unknown conversions remain unverified.
  New browser exports pre-quantize to the observed palette using this same integer
  diffusion, displaying the exported pixels in the preview. Palette-exact source pixels
  are idempotent under that conversion. Verification still requires an exact device
  readback; the next physical upload remains an acceptance check, not a presumed success.
  Unrecognized conversions fail verification and prevent assignment. Only verified images reach
  the separate `locodesc` assignment, under the control reservation acquired before upload,
  without `force`, followed by descriptor readback and release. No speed, direction,
  functions, address or CV is written.
- Upload and assignment cannot be atomic. A later network interruption, forced takeover by
  another controller or failed image readback can still leave an uploaded, unassigned image.
  No automatic deletion is safe: another locomotive may already use it. A fresh preview
  reuses a matching stored image instead of allocating another slot.
  `upload_unknown`, `image_not_stored`, `image_unverified` and
  `assignment_unknown` preserve the chosen slot and object ID, explain possible partial effects
  and direct the user to inspect the ECoS. No POST is retried, no rollback/delete or reboot is
  attempted. A lost browser reply is also explicitly uncertain. An already uploaded image can be
  assigned manually on the ECoS. A verified image also offers an explicit assignment-only
  retry: multipart preview includes `existingSlot` (0–499), checks its occupied state and
  pixels against the supplied BMP, and issues a fresh grant with `existingImage: true`.
  Confirmation rechecks image content, current target and all normal grant constraints,
  audits intent, acquires control, skips the upload entirely, then attempts assignment once without force.
  Failure or cancellation of that preview retains the known slot in the open dialog.
  Unverified images and lost upload replies do not offer this retry. Changed image bytes
  require a fresh preparation; changing target always requires a new target-bound preview.
  `image_not_stored` specifically means a complete refreshed inventory still reports the slot
  as free. An unreadable or changed inventory remains `image_unverified`; it never proves the
  image absent. Neither case allows assignment or an automatic repeat of the upload.
- Audit intent and result contain actor, vehicle, object ID, slot, source SHA-256 and whether
  the stored image was reused, never token,
  image bytes or credentials. A result audit failure is reported without hiding verified effects.

For deployment, the application container needs outbound TCP to the configured ECoS on both
15471 (or the configured protocol port) and 80. Adding port 80 to a restricted host firewall is
a separate operator action; the application does not alter firewall policy.
