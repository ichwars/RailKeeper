# ECoS function keys and decoder-library review

Admins can read the ECoS locomotive linked to a saved vehicle from its Control tab. The read
uses only the configured, enabled ECoS and the persisted external object ID. Missing or ambiguous
mappings and changed address/protocol identity block the request. Identity is checked before and
after reading; a station-configuration change also invalidates the snapshot. No client target,
control acquisition, function switch or decoder programming command is accepted.

`GET /vehicles/{id}/ecos-functions` reads F0–F31 using `get(id, funcexists[nr])` under a view
registration. ESU's [PC interface specification, edition 3, section 7.11, p.8](https://www.esu.eu/uploads/tx_esudownloads/Netzwerkspezifikation_fuer_PC-Interface_ESU_KG_DE_Auflage_3.pdf)
documents the descriptor, hidden/unavailable sentinels and optional `moment` marker. These are
station key settings, not physical decoder-output evidence. All 32 exact command replies, object
IDs, key numbers, codes and successful end markers must validate. Individual lines, blocks,
interleaved events and the overall request are bounded. Any rejected or incomplete reply discards
the complete snapshot. Closing the request cancels the socket. The response is not HTTP-cached.
Nonnegative descriptors are retained as opaque signed 32-bit values, rather than restricted to
the current symbol catalog. A real read-only check returned code 7431 on firmware 4.3.3.
An unknown code neither aborts the other keys nor acquires a guessed meaning or masked icon ID.
Only exact, active compatibility metadata supplies a symbol; the original code remains in provenance.

The existing decoder-library provider and selected CV profile supply `decoderAnalysis` and
`functionCVSuggestion`; no separate interpretation catalog or inferred index pages are introduced.
Only a uniquely active numeric compatibility-symbol match supplies a station symbol. Verified
RailComPlus descriptors and model-specific CV mapping may add evidence. Numeric symbol or verified
key-mode disagreements between decoder and station block automatic supplementation, including
unknown station symbols. Library loading/failure blocks the action; an unidentified decoder is
shown explicitly and supplies no model-specific interpretation.

A comparison lists current RailKeeper fields, station data, library evidence and proposed empty
fields. Selection starts empty. Confirming the selection changes only the function draft; saving
uses the existing atomic vehicle/function transaction and per-function expected revisions.
Nonempty names/symbols, saved or explicitly chosen modes, existing flags, deleted rows and all
personal note bytes remain protected. UTF-8 provenance limits block rather than truncate a row.
Source notes retain the key, numeric descriptor, mode, read time, station host and any CV evidence.
Direction dependence and inversion are never derived from the station descriptor.

The reviewed profiles include LokPilot 5 nano DCC, LokPilot micro V4.0 (including DCC)
and LokPilot Standard. Nano membership and V5 addresses are scoped to the pinned JMRI
definition; its output defaults and proprietary pages are not copied from micro DCC.
V4 uses its own 32-row mapping on index 16/2–3, with physical outputs at row offset 9
and logical outputs at offset 11. Standard uses its documented base registers and
directional mapping, including shunting bit 6 and momentum bit 7. Its reviewed 2015
mapping ends at F18; later F19/F20 and additional mapping banks stay uninterpreted.
Sparse station exports do not establish zero values for omitted conditions or outputs.
Missing V4 condition bytes or Nano control bytes prevent naming a mapping.

Artwork keys are stable identifiers, not ECoS wire codes. The bundled compatibility
codes follow ESU PC interface specification edition 3, §8.2: 10 shunting, 11 momentum,
32 coupling, 260 interior lighting, 1607 notch up. The 94 drawings remain available;
only 33 documented numeric matches enable automatic proposals. Undocumented codes
(including 6 and the opaque descriptor 7431) retain their raw values without a guessed symbol.
The generator emits a metadata fingerprint table. Startup changes only numeric compatibility
metadata matching the original bundled fingerprint, preserving edited metadata, labels,
activation, ordering, sources and vehicle assignments. Existing migrations are unchanged.

No migration, persisted read snapshot, dependency or backup-format extension is required. Protocol
shape was checked read-only against firmware 4.3.3; the complete UI import/save path still requires
its own hardware acceptance. Synthetic UI checks do not establish other firmware compatibility.
