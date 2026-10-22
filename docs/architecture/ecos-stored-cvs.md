# Stored ECoS CV values

The comparison workspace preserves validated CV values from the read adapter. A selected ECoS
locomotive offers **Read all stored CVs**, covering CV1 through CV1024. Creating or assigning
a vehicle, and adding CVs to an already linked vehicle, automatically finishes this full scan
before adoption. The separate read button allows reviewing it beforehand. The reader requests
32 consecutive numbers per authenticated, CSRF-protected call, retaining confirmed progress in
the existing work-item snapshot. Missing numbers never terminate the scan. Stopping prevents
the next batch; a failed batch does not advance the cursor and can be resumed. Restarting a
completed scan replaces its workspace snapshot, not vehicle data. Retrying a committed cursor after a lost response
returns the confirmed workspace snapshot without repeating device reads.

Each request uses `get(id, cv[nr])`, documented in ESU's *Netzwerkspezifikation für das PC-Interface*,
third edition, June 2011, section 7.11, page 8. It reads the locomotive object's stored memory,
without requesting control or invoking decoder programming. The header, object ID, CV number,
byte value and successful end marker must agree. Negative replies supply no values; malformed,
truncated and mismatched exchanges fail the batch. A confirmed zero byte is a valid value.
Replies and batch sizes are bounded. Context cancellation closes the connection.

The configured target must still match the saved session. The endpoint accepts no target,
client-supplied CV values or arbitrary range size. Comparison operations share the existing
ECoS operation lock.

## Adoption and provenance

The create draft previews all supplied values. Saving a new vehicle or explicitly assigning an
existing one imports the server-owned snapshot after linking the ECoS object. Already linked
vehicles can use **Add CV values** directly. The local CV transaction verifies the external
mapping and adds missing unindexed CV numbers only. Existing unindexed readings under any
profile retain their value, description, original source line, export timestamp and history.
Repeating the import adds no duplicate values. No CV import writes to the ECoS.

Imported readings use profile/category `ECoS`. Their description records the ECoS host, object,
CV number and UTC read time and labels the value as cached memory. Export dates, source lines,
decoder freshness and index pages are not invented. Indexed vehicle readings remain distinct;
higher memory CVs are unindexed with an explicit unknown-page explanation in the UI. No schema
or backup format change is required.

## Evidence and limits

The legacy ESU specification documents neither a complete CV-list operation nor a CV31/CV32
page identity for these replies. This scan covers the numbered DCC CV memory range, not all
possible decoder index pages. It cannot replace an indexed LokProgrammer export.

A read-only check on the locally connected ECoS firmware 4.3.3 returned `profile["Empty"]` with
an obsolete-attribute warning and status 22 for the sampled CV requests, including with a view
registration. That establishes no CV-memory availability for that locomotive. The UI reports
zero supplied values clearly rather than treating the network connection as broken or filling
the vehicle with zero defaults. The application never switches profiles to populate the cache.

Source: [ESU specification, public mirror](https://github.com/cbries/railessentials/blob/6ae221d51865a6ce2e32ed97f9c991e64f7b5b8a/ecoslibNet48/Documentation/ecos_pc_interface3.pdf).
