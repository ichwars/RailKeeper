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
- The existing [ESU format evidence](locomotive-image-export.md) supplies 190 × 40 RGB24 BMP.
  The exact output is 22,934 bytes, with a 40-byte Windows DIB header and 572-byte aligned rows.

Sources and read-only device behavior were checked on 22 October 2026. No manufacturer PDF,
vendor image or firmware is bundled. The user previously confirmed manual export/import.
The new direct upload/assignment has automated simulated-device coverage; a confirmed physical
write remains a separate acceptance step. Older firmware may require a restart before a new image
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
- The existing workspace operation mutex serializes transfers and workspace operations. Before
  the first write, RailKeeper consumes the grant, pauses passive monitoring, rechecks identity,
  descriptor, current image and free slot, and records audit intent. An unavailable audit store
  prevents writing. Monitoring resumes with its interruption callback; restart failure is shown.
- Only a free image slot is used. The form and slot are checked again immediately before POST.
  Existing images are never intentionally overwritten or deleted. The ECoS has no atomic slot
  reservation, so another operator must not edit image slots during transfer; a race after the
  final check cannot be eliminated by this client. There is no forced locomotive control.
- Uploaded slots must appear occupied in the refreshed inventory, and images are read back and
  compared pixel by pixel. Exact RGB pixels or an exact
  supported RGB332 palette conversion are accepted; there is no general similarity threshold.
  Unrecognized conversions fail verification and prevent assignment. Only verified images reach
  the separate `locodesc` assignment, under `request(id, control)` without `force`, followed by
  release and a descriptor readback. No speed, direction, functions, address or CV is written.
- Upload and assignment cannot be atomic. `upload_unknown`, `image_unverified` and
  `assignment_unknown` preserve the chosen slot and object ID, explain possible partial effects
  and direct the user to inspect the ECoS. No POST is retried, no rollback/delete or reboot is
  attempted. A lost browser reply is also explicitly uncertain. An already uploaded image can be
  assigned manually on the ECoS; the first version does not retry assignment independently.
- Audit intent and result contain actor, vehicle, object ID, slot and source SHA-256, never token,
  image bytes or credentials. A result audit failure is reported without hiding verified effects.

For deployment, the application container needs outbound TCP to the configured ECoS on both
15471 (or the configured protocol port) and 80. Adding port 80 to a restricted host firewall is
a separate operator action; the application does not alter firewall policy.
