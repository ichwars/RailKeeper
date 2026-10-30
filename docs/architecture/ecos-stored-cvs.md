# Stored ECoS CV values

**Read data** creates one complete normal comparison of saved roster metadata, function keys
and indexed CV data. Vehicle adoption reuses that snapshot. The selected locomotive's
complete database export comes from `/fcgi/ecos-infolokcv?lang=en&target=xml&oid=…`. This is the Excel XML
download linked by the ECoS web interface. It includes explicitly indexed addresses that a scan
of CV1–CV1024 cannot enumerate. For example, `4099:257` retains CV257 on page CV31=16/CV32=3.
The combined unsigned 16-bit page is split into its high and low bytes. Page zero stays distinct
from a missing page. Current selector CVs never supply a missing page for another cache entry.

The reader uses the existing pinned private-LAN HTTP transport without proxies or redirects and
the same resolved address for TCP identity checks before and after the export. Name, object,
address, protocol and profile must remain stable. The workbook namespace, single worksheet name,
table dimensions, complete rows, addresses, byte ranges, decimal/hex agreement and unique full
register addresses are validated. HTML, truncated XML, DTDs, sparse/formula cells and ambiguous
registers fail the entire snapshot. Limits are 4 MiB and 8,192 entries; more than 1,024 entries
are valid. Supplied zero bytes are retained.

An authenticated, CSRF-protected call uses `source: "database"` and `startCV: 1`. Its completed
work-item snapshot has `cvRead.source: "database"`, `nextCV: 1025` and `unavailable: 0`. This cursor
is a completion marker, not an entry count. A retry returns the committed database snapshot;
`restart: true` explicitly replaces it. A failed replacement retains the previous snapshot.
Stopping a pending UI read prevents adoption. It does not start another request or change vehicle
data. The configured station must still match the session and remain active after reading.
Comparison operations retain the shared ECoS operation lock.

The normal reader uses the documented roster and metadata operations plus saved `funcexists`
key descriptors. It never runs the diagnostic TCP CV probes, takes control, programs a decoder
or changes index registers. The roster and each locomotive identity are checked before/after
collection, and the configured host/port must remain unchanged. Read context is bounded to
90 seconds with a 95-second response deadline only on the relevant Admin read operations.

If private backup configuration is available, the same pinned HTTP transport reads
`/cgi/save?lang=en` once per complete roster. The verified backup is an internal source, not
another user import. XML and text configuration values are combined by full register address.
Differing values are recorded in `cvSources.conflicts` and omitted from importable `cvs`.
They never silently replace an existing value. A conflict-bearing otherwise equal work item
is marked for review; existing new/mapping-conflict statuses retain their meaning. Missing or
failed internal source access produces a visible limitation rather than fabricated data.

The v151 `loks.bin` layout additionally supports bounded metadata decoding for normal and
MULTI records. The stored byte area has no verified validity/freshness marker and is zero-filled
by initialization. Nonzero CV1–CV255 candidates therefore remain unverified. Candidates that
disagree with a text/export value quarantine that address and retain both values for review;
raw-only candidates are stored separately as `unverifiedValues`, never as known CVs. Zero bytes
are not converted into known zero readings. Unsupported binary layouts leave text parsing intact
and mark binary metadata unavailable. Function masks/mode slots are useful evidence, but the
complete F0–F31 snapshot comes from the saved key-descriptor interface, without guessing icon codes.

`center.functionSnapshot` and completed `cvRead` are saved in the normal work item. Creating a
vehicle uses that snapshot directly in ordinary CV/function draft fields. Active symbol mappings
and decoder-library evidence prepare supported suggestions; unknown codes remain source notes,
and manual draft values are never overwritten. A separate function import dialog is not required
for a newly adopted locomotive. Existing manually assigned functions remain unchanged.

On source conflicts the raw profile name is `sourceProfile` for display only and `profile` is
empty. This prevents quarantined identity CVs from becoming a misleading missing-CV/profile-name
fallback. The latest conflicting normal snapshot also suppresses older qualifying vehicle profile
context. Exact register identities and existing manual readings remain protected.

## Legacy numbered reader

Clients omitting `source` retain the 32-number batch API over CV1–CV1024 with confirmed progress
and retry protection. They must explicitly restart to replace an indexed database snapshot.
Each request uses `get(id, cv[nr])`, documented in ESU's *Netzwerkspezifikation für das PC-Interface*,
third edition, June 2011, section 7.11. This reads locomotive-object memory. The header, object,
CV number, byte and successful end marker must agree. Unsupported replies supply no value;
malformed exchanges fail the batch without advancing its cursor. Index pages remain unknown.

## Adoption and provenance

The create draft previews all supplied values. Creating or assigning a vehicle imports the
server-owned snapshot after linking the ECoS object. Already linked vehicles can use **Add CV
values** directly. The transaction verifies the external mapping and adds missing full register
addresses only. Existing readings at the same CV number and exact optional CV31/CV32 pair, under
any profile, retain their value, description, source line, export timestamp and history.
Different pages of the same CV remain separate. Repeating an import adds no duplicates.

Imported readings use profile/category `ECoS`. Their description records the host, object, CV,
read time and known index pair, and labels the value as cached memory. Explicit database pages
survive both the new-vehicle draft and existing-vehicle import. Missing pages, source export
dates, original lines and decoder freshness are not invented. Unknown higher entries retain
the existing unknown-page explanation. No migration or backup-format change is required.

## Library interpretation

The source profile `ECoS` remains separate from the decoder model. For that source only, the
vehicle's recorded decoder type can select a library profile when CV8 agrees and the existing
model-identification checks report no conflict. An exact, unambiguous ECoS decoder-profile name
can also select an active model profile from the library, including when CV8 is absent. The
identity inventory must be available, and manual decoder labels, supplied manufacturer readings
and model-identification CVs must not contradict that selection. Generic NMRA, unknown, inactive,
ambiguous and family-only profile names do not enable model rules. A neighbouring model's rules
are never substituted. Firmware-dependent rules retain their existing verification requirements.

The optional read-only vehicle field `ecosDecoderProfile` derives its name, object, host and read
time from the latest saved comparison for the vehicle's unique ECoS mapping and currently
configured host/port. Address/protocol conflicts or a newer missing/invalid snapshot suppress it.
An initially incomplete roster snapshot qualifies after a verified complete database CV export.
This lookup contacts no device and does not change `decoder_type` or CV rows. Comparison previews
use their own saved source profile. If comparison workspaces are cleared or absent after a JSON
restore, read the ECoS database again to restore this context.

CV-based identification continues to work without a vehicle type. Readings from other source
profiles are never merged to supply missing identity or values. The UI distinguishes ECoS profile
context, the vehicle type plus CV8, and decoder identity CVs. Missing CV8/firmware remain missing;
the source profile never becomes a measured manufacturer reading. CV-derived function names and
symbols are explicit draft suggestions; existing manual assignments remain intact.

Definitions match exact register addresses. Explicitly unindexed model definitions are usable even
above CV256; the number alone does not make a register indexed. An ECoS value with no page must
not match an indexed definition, including when the catalog currently documents only one page.
If the selected profile documents indexed and unindexed definitions for the same number, a
page-less ECoS reading remains unresolved. Current CV31/CV32, factory defaults and matching byte
values cannot supply the missing identity of older cache entries. Unknown higher entries appear
as **Unknown index page** in the raw table, page filter and detail drawer, without bit meanings
or a guessed table hint. The drawer exposes cache provenance separately from manufacturer sources.
An explicitly unindexed, model-specific manufacturer table row may remain a documentation hint
when CV8 is available and neither the selected profile nor the loaded references document an
indexed or unspecified alternative for that number. It never becomes an executable analysis rule.

Interpretation is computed on display, including for existing imported ECoS entries. It does not
rewrite values, profile labels, index fields, provenance or history, and performs no device writes.

## Evidence and limits

The legacy ESU specification does not describe a complete CV-list operation or page identity for
the numbered reply format. The database reader was independently checked on firmware 4.3.3 on
27 October 2026: Generic NMRA profiles supplied 35 base entries, while a LokPilot 5 micro DCC
profile supplied 1,457 entries including 1,374 explicitly indexed entries on 13 pages. These are
observed database contents, not guaranteed counts or fresh measurements. A separate indexed TCP
cache request, `get(id, cv[4099:257])`, returned the same value as the XML row.

Some model profiles omit CV7/CV8 even with extensive indexed data. An ECoS profile name is source
context, not an invented manufacturer reading or verified decoder identity. Cached values can
include profile defaults or disagree with the roster address. Missing values stay missing. The
application never switches profiles, sets CV31/CV32, requests control or invokes decoder
programming to populate the cache. Direct decoder reading and programming-track/RailCom operations
remain outside this workflow.

Targeted indexed cache reads do not necessarily recover registers omitted by the export. On the
same firmware, a LokPilot V4.0 micro DCC snapshot lacked condition byte I and action byte N in all
32 mapping rows (64 registers); a LokPilot 5 nano DCC snapshot lacked condition byte J and action
bytes L/P in all 72 rows (216 registers). Every missing address returned a successful TCP reply
without a value, while known control addresses returned their stored bytes. Fresh HTML and XML
tables agreed with each other and the earlier imported snapshots. This establishes missing data
on those read interfaces, not zero bytes or a limitation of the physical decoder. Missing fields
must not be filled from defaults to enable otherwise incomplete mapping suggestions.

The device web interface also offers a binary `.eco` configuration backup through
`/cgi/save?lang=en`. The observed firmware 4.3.3 download is an OpenPGP message: a version-3
ElGamal encrypted session-key packet (tag 1), followed by a version-1 encrypted data packet with
integrity protection (tag 18). GnuPG and an independent bounded packet reader agree on its framing.
An official ESU example has an outer `ECHK` wrapper around an OpenPGP message for the same recipient.
The matching private key is required to recover the session key; recognizing these headers neither
decrypts CV values nor verifies the encrypted integrity check. The offline importer below reads
the saved text CV configurations; additional CVs or mappings in binary locomotive data remain
unverified. The earlier partial CS1 update did not include the private keyring. A full related
rescue image later supplied a matching key for local provisioning; a firmware package alone is
not a guarantee of decryption. The importer currently accepts the observed bare OpenPGP envelope,
without an outer `ECHK` wrapper. It is never an automatic fallback for device reads.
A device backup must never be restored or a decoder profile changed merely to inspect stored CVs.

Backup format references: [OpenPGP packet format](https://www.rfc-editor.org/rfc/rfc4880.html#section-5.1),
[official ESU example](https://www.esu.eu/fileadmin/img/support/Tips_Tricks/ECoS_new_functions/StartZiel/ECos2_FAQ_StartZiel.eco).

## Maintenance compatibility for offline backups

The old explicit upload endpoint remains for compatibility and diagnostics. It is deprecated
in OpenAPI and has no upload panel in the normal Digital Centres workflow. Regular users use
**Read data**, compare the resulting records and adopt missing vehicles in the same flow.

Admin can upload an encrypted `.eco` file through `POST /digital-centers/ecos/backup-import`
as raw `application/octet-stream`, with the normal session/CSRF protections. The server reads
at most 8 MiB encrypted data and verifies the complete OpenPGP integrity check before parsing
the gzip/tar or ZIP contents in memory. Limits include 32 MiB decoded data, 10,000 archive members,
64 locomotive configurations and 8,192 CVs per configuration. Unsafe paths, links, duplicate
addresses, malformed configurations and incomplete archives fail without creating a session.
No archive file is extracted, executed or restored. This uses the already bundled Go crypto
dependency and does not invoke a GnuPG process or install firmware.
The deprecated `golang.org/x/crypto/openpgp` package is confined to reading the station's
legacy format, with narrowly scoped staticcheck exceptions on those imports. It is not used
to create application backups or provide general encryption. Integrity verification, input
and expansion limits, private key confinement and Admin access remain mandatory.

An administrator provisions `private/ecos-backup/keyring.gpg` and `passphrase` beneath the private
data directory, outside `uploads`. They are read with a confined filesystem root and size limits;
missing or unusable configuration returns a generic unavailable error. Use private directory/file
permissions (0700/0600). The application does not distribute vendor keys, expose credentials in
its API or include this configuration in ordinary JSON application backups. A full private server
backup may include it and must retain the existing restricted backup permissions.

Import creates a separate comparison session atomically. Existing sessions and vehicle records
remain intact. Names and decoder-profile context can come from prior saved comparisons and are
identified as such, rather than newly recovered backup identities or live measurements. Unknown
objects retain a neutral object-number label. Without station confirmation there are no vehicle
links. The optional `X-RailKeeper-Confirm-Station: true` header requires a matching
`X-RailKeeper-Backup-Station` host and permits only unambiguous existing mappings with matching
saved object, address and protocol. These headers never provide a network target.

Offline snapshots use `cvRead.source: "backup"`, the import time, the assigned station host and
a completed cursor of 1025. Import time is not presented as the original export time. The backup's
serial and CV counts appear in the result summary. Explicit index bytes, including page zero, are kept.
The session has no device capabilities. Its review link includes `sessionId` and `review=backup`;
the UI loads saved session/items using GET requests and retains the link for reload. Device,
vehicle creation and assignment actions are unavailable in this review. The two CV read paths
also reject attempts to replace a backup snapshot before contacting any device.

For an already linked vehicle, an explicit **Add CV values** action imports the saved snapshot
directly without first reading the ECoS. Existing register values, descriptions and histories,
manual decoder choices and function assignments remain protected by the existing additive import.
The backup's host must agree with its session. Backup comparisons are excluded from the lookup
of live ECoS decoder-profile context, so they cannot displace an existing validated profile.

Legacy command source: [ESU specification](https://www.esu.eu/uploads/tx_esudownloads/Netzwerkspezifikation_fuer_PC-Interface_ESU_KG_DE_Auflage_3.pdf).
