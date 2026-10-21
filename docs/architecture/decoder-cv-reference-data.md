# Manufacturer CV reference data

RailKeeper stores curated manufacturer facts as local `decoder_profile` master-data metadata.
They are separate from `vehicle_cv_values`, vehicle functions and uploaded original documents.
The shared `backend/internal/decoderlibrary/catalog.json` is the bundled source for Go and
TypeScript. Startup and restore persist missing profiles and upgrade only known, unedited
bundled revisions. No new SQLite table or schema migration is needed for optional documentation
fields. Existing operator labels, activation, ordering and local metadata remain protected.

## Reviewed scope

The catalogue contains 2,091 records: 2,088 functional definitions and three ESU reservation records,
including 18 DCC foundation records. A record may describe an explicitly bounded address range; this count does
not establish that every byte in that range has complete semantics. The eight product brands
are represented by these selected documentation profiles, not by all 599 catalogue entries:

| Documentation profile | Definition records | Manufacturer source |
| --- | ---: | --- |
| ESU LokPilot 5 micro DCC | 220 | [LokPilot 5 manual, edition 8](https://www.esu.eu/download/betriebsanleitungen/digitaldecoder/?tx_esudownloads_pi1%5BdownloadItem%5D=803bc0046a0244b6416c82c6abfd385e), 217 functional definitions and three explicit reservations; additional firmware evidence retained |
| D&H DH18A / DH05C / DH10C / DH16A / DH12A / DH14B | 120 each | [CV table, firmware 3.15](https://doehler-haass.de/cms/pages/haeufige-fragen/decoder-cv-tabelle.php), model-limited rows with hardware qualifications |
| D&H DH21A | 121 | Same DCC table, pp.24–27 hardware variants/conflicts; CV144 partially explained (bits1–5), bit0 unresolved |
| D&H DH21B / DH22B | 126 each | Same DCC table, pp.28–31/36–37; dedicated AUX7/8/11/12 assignment and AUX7/8 conditions |
| D&H DH22A / DHSP10A | 120 each | Same DCC table, hardware pp.32–35/40–41; no B-only output rules |
| D&H DH24A | 127 | Same DCC table, pp.38–39, including AUX10 assignment |
| ZIMO MN180N18 | 68 | [MS/MN/FS manual, SW5.29 (2026-07-01)](https://www.zimo.at/web2010/documents/MS-MN-Decoder.pdf), selected non-sound registers |
| Lenz STANDARD+ V2 | 77 | [10231-02 manual, edition 5](https://www.lenz-elektronik.de/media/be/06/d9/1734004977/b_10231-02.pdf?ts=1734004977), pp.19–26 |
| Kühn N45 | 59 | [N45 manual, revision b](https://www.kuehn-digital.de/files/N45_de_b.pdf), sections 4.3/4.4 |
| Uhlenbrock IntelliDrive 76420 | 43 | [German 76420 manual, 11.10 ru](https://www.uhlenbrock.de/de_DE/service/download/handbook/de/archiv/Bes76420.pdf), pp.2,4–6; [bounded evaluation](uhlenbrock-cv-quality.md) |
| Roco 10882 | 35 | [10882/10883/10884 manual, 2012](https://www.roco.cc/documents/pdf/Anleitung_10882.pdf), only 10882-applicable rows and model-limited shared masks; [bounded evaluation](roco-cv-quality.md) |
| PIKO 56500 PluX22 | 111 | [Detailed 56500 manual](https://www.piko-shop.de/is.php?id=28620), selected base and exact indexed effect registers |

The original ESU indexed definitions and F0/F1/F3 evidence remain in place. Located references
and documented defaults remain optional for older records. Full sound mapping, every firmware revision and all hardware variants are
outside this bounded set. PDF text extraction produces research candidates; only curated facts
enter the bundled catalogue. Original PDFs and extraction scratch files are not committed.

## Catalogue-wide manufacturer references

The separate `decoder_reference` master-data type accounts for every catalogue model without
promoting extracted manual tables into executable rules. The current collection contains 310
reference sets and 30,410 CV associations for 580 of 599 models. Another 19 models have no CV
rows: 15 need matching documentation, while four are auxiliary modules or Motorola receivers
without their own DCC CV table. The five legacy Lenz models previously limited to CV7/CV8 now
provide 217 reference rows from their original manuals, with exact edition and source locations.
Two additional original references cover LE030 (June 1997) and LE0511A/LE0511D (third edition
07/03) with 31 rows. They retain the different CV67–81 and CV67–94 speed-table boundaries,
printed source pages and indexed-text availability restrictions. Unclear byte defaults and
register numbers are not inferred. The old direct LE0511 PDF URL was unavailable; no PDF hash
is claimed for its indexed original text or the LE030 web reading.
Four Roco-branded articles (685404/685504/686101/686201) have a separate, shared MX622 reference:
the [original Fleischmann 2022 catalogue](https://www.fleischmann.de/documents/pdf/2/20220627_FLM_Zubehoer_DE.pdf)
explicitly equates these articles to wired MX622, MX622N and MX622R on pp.16–17. The
[ZIMO MX manual](https://www.zimo.at/web2010/documents/MX-KleineDecoder.pdf), dated 19 August 2024
on p.2, lists this family on p.3. Its p.69 supplies 40 curated reference rows covering 83 addresses.
The catalogue establishes model identity; each CV row instead points to the ZIMO CV table.
No OEM factory values, sound CVs, unchecked firmware extensions or automatic identity rules are
added. CV17/18 remain a combined address without an invented per-byte range. CV6/CV61 have no
exhaustive range because their table entries are conditional or incomplete. Both downloaded
original documents retain individual SHA-256 hashes and their retrieval dates.
The [complete review of the remaining 26 exact models](decoder-cv-remaining-models.md)
accounts for every model. Five ESU variants receive five shared firmware rows, while six PIKO XP
projects receive 28 shared rows. These 33 additions rely on explicit family-wide release notes,
with firmware conditions in each heading; full variant/project tables still remain missing.
PIKO CV48 retains the changed 1.16/1.18 threshold, and CV30 bit4 retains the FAQ/release-history
version conflict. No project defaults, unbounded `ff` ranges, XP-G-only parameters or executable
rules are inferred. The other 15 exact models remain without CV rows, with individual evidence
and rejection reasons recorded. Stable internal `cv-gap-*` keys are retained when a formerly empty
group gains partial firmware evidence; its title and coverage status reflect the actual scope.
Even references with rows are marked `partial`: headings alone do not establish every bit, hardware
variant, index page or firmware restriction.
See [coverage and open models](decoder-cv-reference-coverage.md).

Each reference stores model IDs, bilingual title/limitations, source documents, status and
bounded factual rows. A row carries CV/range, source-language technical heading, literal optional
range/factory text, source index, one-based PDF page or original printed page in a web reading
(zero for nonpaged evidence), applicability (`manual` or `model`) and an explicit index scope (`base`, `indexed`, `unspecified`). Unknown
pages never become base CVs. Multiregister defaults are not split into guessed byte values.
Conflicting repeated base-register headings are excluded. Official Lenz editor XML facts preserve
the exact ZIP entry and archive/file hashes; GUI Value/DefaultValue fields are not hardware
factory evidence. Read-only reference data never enables device writes or automatic model matching.

Startup/restore seed missing references with the same operator-edit protection as profiles.
Master-data transfer and backup validation include references; each reference is bounded to
512 KiB, 1,024 rows and ten sources. No database/schema or backup format version change is needed.
`GET /decoder-library` returns compact `referenceCoverage` summaries, while authenticated
`GET /decoder-library/references/{modelID}` returns active rows only for the selected model.
The frontend does not bundle the large reference corpus. Settings load it on the selected model's CV tab; vehicle CV lists/details load only the exactly associated model on demand.
The functional definitions and 44 identity rules remain separate from these manufacturer references.

The master-data cache and per-type list sort decoded entries in Go by active state,
sort order and label. Sorting multi-megabyte metadata blobs inside SQLite would spill
to temporary files, which are unavailable in the production read-only container.
This preserves ordering without a new writable mount or a SQLite schema change.

## Model-by-model reference qualification

Every one of the599 inventory models is accounted for by a manufacturer reference or an explicit
missing-document entry. The599-model regression audit checks exact source/index association,
conflicting headings, inactive references and preservation of original facts. Reference rows were
already stored; this step makes unambiguous headings useful in vehicle CV lists and detail drawers.
It does not claim that all599 models have executable interpretation or hardware-tested firmware.

| Manufacturer | Models | With reference rows | Documentation gaps | Reviewed functional profiles |
| --- | ---: | ---: | ---: | ---: |
| ESU | 58 | 58 | 0 | 1 |
| ZIMO | 177 | 177 | 0 | 1 |
| Doehler & Haass | 41 | 41 | 0 | 12 |
| Lenz | 49 | 35 | 14 | 1 |
| Kühn | 7 | 6 | 1 | 1 |
| Uhlenbrock | 52 | 49 | 3 | 1 |
| Roco | 30 | 29 | 1 | 1 |
| PIKO | 185 | 185 | 0 | 1 |

Vehicle headings require an exact technical model identity or a unique model/article label, without
manufacturer contradictions. Family names and shared article numbers do not select a model.
A manufacturer reference can only supplement an otherwise unexplained CV, retains `known=false`,
and never adds a rule to the import resolver, a bit interpretation, inferred state or function setting.
The manual/model applicability, firmware limitations, literal range/default text, document/page and
revision stay visible. A documented default never fills a reading. Existing manual descriptions
have precedence in the value list and remain unmodified in the record.

Unknown index scopes cannot match; base requires both indices absent, while indexed rows require
both exact indices. Explicit unindexed registers above256 remain unindexed. Conflicting overlapping
headings/defaults/qualifications are omitted rather than resolved by source order. Loading, failure,
retry and missing-document states remain explicit. Late model responses and library refreshes cannot
reuse a stale model association. Only the current profile group receives hints. API/DB/backup formats
remain unchanged; no network retrieval of manuals occurs while viewing a vehicle.

## Structured interpretation metadata

- `reference.sourceIndex` addresses the containing profile's `sources` array;
  `reference.location` identifies the table row, page or section.
- `factoryDefault` is optional and must lie within the definition's bounds and finite value set,
  if present. Variable, conflicting or unspecified defaults are omitted. It never fills a missing
  vehicle CV or proves the state of a physical decoder.
- `access` is `read-only`, `read-write` or `command`. A reset register may return an identifier
  when read while accepting a reset command when written; the accompanying help preserves that
  distinction. None of these fields enables device writes.
- `unit` is bilingual documentation. Compound or model-dependent scales remain qualified in
  help text. No executable conversion expressions are accepted.
- `allowedValues` records finite sets such as PIKO CV96 `{1,6}` and Uhlenbrock CV60 `{0,9}`;
  an interval alone must not imply all intermediate values are documented.
- A source may carry `retrievedAt` and `sha256`. A document hash records the reviewed bytes,
  not ongoing online availability. No background fetch is introduced.

All fields round-trip through the authenticated API, master-data transfer and backups. The
existing per-profile 128 KiB and inventory 512 KiB limits remain unchanged. Raw CV bytes,
index identities, export provenance and manual function settings remain independent.
The UI uses names/help and bit explanations. The ESU detail view below also displays documentary
units, defaults and located evidence; these fields remain read-only documentation.

The ESU quality review adds three explicit reservation records (six addresses). All 220 ESU records
now have located evidence. The CV detail drawer displays the selected evidence, documented unit
and factory reference separately from the read value. Reserved addresses have no functional
semantics and do not increase interpretation coverage. See the
[ESU quality boundaries and reference-export accounting](esu-cv-quality.md).

The [DH18A quality review](dh-cv-quality.md) documents120 located records for DH18A and120 each for DH05C/DH10C/DH16A/DH12A/DH14B, with explicit hardware-generation limits.
The catalogue therefore contains 2,091 records (2,088 functional definitions plus three ESU reservation
records). `rules.motorCurveCV` identifies a documented model-specific scalar characteristic CV;
the frontend displays that register instead of inferring DCC curve points. D&H brake-ramp rules
use CV154/155. Output masks, conditions and timers never establish function-key mode or inversion.

## Contradictions and OEM boundaries

- Roco's 10882 manual explicitly names ZIMO and CV8=145. Its branded catalogue entry remains
  `brandOnly` and receives a `documentationProfileId`, not an identity/profile link based on
  Roco's registered brand identifier161. Name-based interpretation still requires technical
  CV8=145. The reference does not automatically identify another Roco decoder.
- The German Uhlenbrock76420 manual (11.10 ru) resolves English errors: CV55 is the I constant,
  CV56 the control reference, CV17 high byte192–231 and CV98=0 means no timed outputs.
  CV29 bit3 is unused, so no RailCom flag is borrowed from the DCC foundation or76425.
- PIKO56500 repeats CV101 with different meanings. Neither meaning is guessed. Effect CV267
  on page18/0 is distinct from the same address on page1/1 or without an index.
- Kühn's summary repeats CV49 for outputs C/D, but the detailed section explicitly identifies
  CV51/52. Those detailed assignments are recorded; function-mapping masks are not generalized
  across key groups.
- Lenz uses one-based bit numbers with zero-based numbers in parentheses. Internal bit fields
  use the latter. The contradictory numerical example for CV149 is qualified; no derived factory
  time is recorded from it.
- ESU's table singles out LokPilot5 DCC for a different CV3/4 scale. No unverified conversion
  for micro DCC is introduced. Firmware-specific CV11 evidence remains separate.

Regression tests cover documented values, non-contiguous sets, source validation, index isolation,
OEM conflicts, lossless metadata storage/backup and guarded upgrades from the deployed catalogue.

The [MN180N18 quality review](zimo-cv-quality.md) expands the ZIMO profile from 21 to 68
located records. Its source scope is firmware 5.29, with future SW6.00, sound-only, MX-only and
large-scale functions excluded. CV250=127 still identifies MN180 only, without proving Next18.

The [STANDARD+ V2 quality review](lenz-cv-quality.md) expands its profile from64 to77 located
records. CV51 braking, CV28/29 RailCom, output masks, motor settings and smoothed speed
parameters are evaluated only for that reviewed model. Unconfirmed output E and contradictory
dwell timing remain qualified; no function-key mode or inversion is inferred.


The [Kühn N45 quality review](kuehn-cv-quality.md) expands54 to59 located groups and
adds CV55/60–63 to its manufacturer reference. The bounded evaluator handles shifted output
masks, E/F logic/SUSI gating, decimal effects, motor and braking parameters without key-flag
takeover. Other Kühn models do not inherit N45 rules.

PIKO56500 interpretation is limited to the exact PluX22 model: [reviewed facts and source conflicts](piko-cv-quality.md). Its 111 groups cover218 CV addresses, including all90 output parameters on index18/0. Other PIKO profiles do not inherit this evaluator.

DH21A adds121 records, including one partially explained CV144. Counts describe located registers,
not complete bit coverage or identified hardware generation. See the [D&H quality review](dh-cv-quality.md).
