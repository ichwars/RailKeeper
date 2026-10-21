# Doehler & Haass: interpretation quality

DH18A, DH05C, DH10C, DH16A, DH12A and DH14B each contain120 located definition records;
DH21A has121 including partially explained CV144. These include direct
manufacturer address/consist definitions. Model-specific interpretation requires CV8=97 with
a matching profile name or the reviewed CV261 type code (180,52,102,160,120,141,200 respectively).
The documented scope starts at firmware3.15. Unknown firmware remains unknown; descriptive
CV7 examples do not introduce a new firmware detector.

Sources are the manufacturer's [DCC CV table](https://doehler-haass.de/cms/pages/haeufige-fragen/decoder-cv-tabelle.php)
and [vehicle/function decoder manual, January 2026](https://doehler-haass.de/cms/media/pdf/DH05C_DH10C_DH16A.pdf).
Reviewed byte hashes and retrieval dates are retained. Definitions reference the DCC table on
pp.65–80 and the relevant appendices, never the SX2 parameter table starting on p.86. DH18A
hardware restrictions follow both generations on pp.22–23: AUX3–6 are logic-level outputs,
and SUSI pin use is conditional. CV261 does not identify the hardware generation.

## Reviewed meanings

- CV13/14 describe analog/MM1 function states; CV28 describes RailCom channels and dynamic use.
  CV29 bit 2 remains analog permission, not dynamic channel selection. The source gives CV29
  factory alternatives6/14, so no single default is guessed.
- CV9 distinguishes frequency selection from low-frequency operation. CV27 leaves unused bits 2/3
  without a braking meaning, and records a finite set of documented combinations.
- CV33–47/64 are output assignment masks. Bits6/7 change from dimming/shunting to AUX5/AUX6
  when CV137 bit 4 enables extended mapping. CV35 also programs CV47, CV36 also CV64;
  these programming side effects are documentation only, never hardware writes by RailKeeper.
- CV60 bits 2/3 jointly select shuttle operation. They are not independent on/off properties.
  CV63 has different scales for ordinary start delay, reversing dwell and restart dwell;
  no single unconditional unit is attached.
- CV113–116/125–128 assign keys that switch off an output despite its on-assignment. These
  masks do not establish global function-key inversion. CV117–120/129–130 output timers use
  100 ms increments, with zero disabling the timer; they do not establish momentary key mode.
- CV137 bit 0 is explicitly enabled by default for DH18A. Its model-specific factory reference
  is1, instead of the generic table's 0. Bits0/4/5 jointly control SUSI/logic-output use.
  Bit2 inverts SUSI direction, not an individual function key.
- CV145–152/157 encode additive condition groups, not independent binary bits: direction 0/1/2,
  movement 0/3/6, F0 condition 0/9/18 and shunting behavior 0/27/54/81/108/135. CV148 is located
  in the online table because the PDF's corresponding row is omitted.
- CV153/156 describe initial output assignment and dimming masks. CV401–412 select a remapped
  key number (0 disabled,1–28=F1–F28,29=F0); they are unindexed CVs, even above 256.
- CV461/463/465/467 describe output dimming subject to model restrictions. They do not prove
  independent dimming or fitted loads. CV135/136 calibrate RailCom speed feedback; a stored
  calibration value is not a RailKeeper speed measurement.

## Manufacturer motor characteristic and braking

The profile declares `rules.motorCurveCV=48`. This optional rule must point to an individually
documented, unindexed motor definition in that profile. It does not contain an executable
formula or inferred curve points. The explained motor section links CV48. The speed-curve view
shows the stored parameter and a manufacturer's-characteristic explanation instead of plotting
CV2/5/6 or CV67–94 as a DCC characteristic. CV29 remains visible as a raw register. Indexed
copies never replace base curve data. A missing CV48 remains unknown, with no factory fallback.
For the verified custom-characteristic profile, base CV6 and CV67–94 do not inherit generic DCC
curve-point meanings. Explicit model definitions retain priority; raw readings are preserved.

Brake-ramp configuration uses CV154/155, not ESU's CV254/255. CV154>0 enables the configured
ramp; CV155=0 reuses CV154 for reverse travel. A positive CV155 alone does not enable the ramp.
Time from maximum speed is raw value/8 seconds (p.78, appendix 5 p.111). Actual stopping distance
depends on calibration and braking conditions; the configured-state badge does not measure it.

## Deliberately open or excluded

- CV131–133: the main table and online table disagree with appendix 1's register references.
  These are not promoted to functional definitions in this stage.
- CV144: the dynamic-channel description and the web table's cross-link are inconsistent. Earlier
  profiles leave it undocumented. DH21A now explains only corroborated bits1–5; bit0 stays raw,
  and no operational GPIO inference is introduced.
- CV98/99/101–103 and 158/159 are restricted to other hardware. They do not become DH18A rules.
- CV260 remains read-only in the profile. A PDF reset aside is not promoted to a write command.
- Output conditions, timer values and SUSI direction never become function-key inversion,
  momentary mode or manual function edits. No D&H counterpart to the verified ESU checkbox
  export tests is claimed.

All raw readings, index identities and import provenance remain independent. Startup/restore
upgrades only exact known previous bundled metadata and preserves local edits and operator
settings. Each profile round-trips within the existing 128 KiB limit; schema 71 and backup 24 remain
unchanged. Other decoder profiles and the 284 manufacturer reference sets remain separate from this
focused semantic review. Current aggregate counts are maintained in the reference-data overview.


## Exact DH05C and DH10C extension

The shared manufacturer manual is the identical reviewed January2026 PDF (SHA256
`e919dfc7bc42a6f1135fd4698caa6d979c48df84b24db5cf1e9b12f17c6d9601`). The online CV-table snapshot
checked2026-10-20 has SHA256 `a520a0aa046075e6d33ae1bccb5324ebf17771f52b4a1996f44c92b9f1159478`.
Hardware pages14–17 and the combined CV137 table on page105 were visually inspected.

DH05C uses CV8=97/CV261=52; DH10C uses CV8=97/CV261=102. These are exact model identifiers,
not hardware-generation or fitted-output detection. Both profiles retain the documented firmware
scope3.15; a missing firmware remains unknown and an explicitly older firmware disables model rules.
Current source tables cover two generations of each model. Generation1 shares ZCLK/ZDAT between
SUSI and logic AUX3/AUX4. Generation2 provides separate AUX3/AUX4, with ZCLK/ZDAT shared between
SUSI and logic AUX5/AUX6. DH10C generation2 amplifies AUX3/AUX4 while its generation1 does not;
DH05C's documented additional outputs remain logic-level. A type code alone establishes none of
these pin or load limits. Conditional output definitions explicitly preserve this uncertainty.

CV137's factory reference is0 for the compact models, while the special DH18A default remains1.
No unconditional output-map or SUSI pin-state algorithm, key mode or inversion is introduced.
CV33–47/64 remain documented register masks; remapping, output-off masks, timers and additive
conditions retain their existing meaning without modifying manual function records.

The five DCC address/consist registers17,18,19,21,22 now have direct page67 definitions in all
three profiles. The table prints0–255 for each address byte; the evaluator separately checks the
DCC high-byte encoding and1–9999 address limit. CV19 values0/128 are inactive; bit7 reverses the
consist direction. CV21/22 are address-command permissions, not fitted function names. CV27/28
source locations are corrected to page68 and CV14 to page67. Existing vehicle data is untouched.

`rules.modelDefinitionsOnly` prevents unreviewed DCC foundation meanings from being merged into
a successfully matched compatible profile. The same declaration governs Go resolution, TypeScript
analysis, settings definition tables and CV searches. Unknown bytes remain raw and unsupported
values do not become interpreted states. When a model is inactive, conflicting or explicitly outside
the reviewed firmware scope, generic fallback follows the existing resolution rules.

Inventory revision2026-10-20.4 adds the two profile links, generation notes and the documented
historical DH05C-1 ribbon-cable/NEM651 variant (manual p.14, generation1). It does not claim
current availability. Existing order/interface mappings remain unchanged; the total across model records is747,
previously746. Model names, type rules, manufacturers, families and manufacturer reference rows
are unchanged. Exact old inventory revision2026-10-20.3 and DH18A revision2026-10-20.2 are trusted
upgrade fingerprints; locally edited metadata and operator settings remain protected. Current profile totals are listed below. The284 reference sets/29,224 rows and
532/599 documented models remain unchanged. No real DH05C/DH10C export has been supplied yet.


## Exact DH16A extension

The DH16A profile has120 DCC definitions, firmware scope3.15 and CV8=97/CV261=160. Its four
existing order/interface variants (-0,-2,-3,-4) remain unchanged. Inventory revision2026-10-20.5
adds only the DH16A profile link and a model-specific hardware note. Exact previous inventory
revision2026-10-20.4 is trusted for upgrades; locally edited metadata and operator settings remain
protected. No runtime algorithm, schema or backup-format change is introduced.

Primary sources are the shared [January2026 manual](https://doehler-haass.de/cms/media/pdf/DH05C_DH10C_DH16A.pdf),
pages20–21,65–80,105,112; the [manufacturer CV table](https://doehler-haass.de/cms/pages/haeufige-fragen/decoder-cv-tabelle.php);
and the [current DH16A product page](https://doehler-haass.de/cms/pages/produkte/fahrzeugdecoder/dh16a.php).
The existing same-day PDF was reused after checking the official online sources; its hash remains
`e919dfc7bc42a6f1135fd4698caa6d979c48df84b24db5cf1e9b12f17c6d9601`. Both hardware pages were
visually inspected. The current product page describes generation2, while the manual also documents
generation1. Neither CV261 nor the unchanged order numbers distinguish these generations.

Unlike DH10C generation1, DH16A has amplified AUX3/AUX4 in both documented generations;
AUX5/AUX6 are logic-level and share ZCLK/ZDAT with SUSI. Thus compact-model text that restricts
AUX5/AUX6 to generation2 must not be inherited. Mapping, initial/dimming masks, timer, condition
and dimming explanations record the DH16A-specific facts. CV137 bits0/4/5 must still be considered
together according to appendix3; default0 is retained, not the DH18A-specific1. The common table
warns that not every output on every decoder is independently dimmable; no new independent-dimming
algorithm or consumer identification is claimed. Unindexed CV401 onward are not index-bank data.

Imports without a decoder name can select this exact profile from the two type registers. Missing,
indexed, conflicting or inactive evidence never supplies DH16A-specific meanings. Raw readings,
source lines, export dates, manual functions, mode and inversion are preserved. No real DH16A CV
export has been supplied for hardware comparison yet. Manufacturer reference files are unchanged.


## Exact historical DH12A extension

The historical DH12A has120 located DCC definitions from firmware3.15. Its existing PluX12
connection and DH12A order number are unchanged. The manufacturer explicitly marks it unavailable
on the [DH12A product page](https://doehler-haass.de/cms/pages/produkte/fahrzeugdecoder/dh12a.php).
Inventory revision2026-10-20.6 links the exact profile, adds CV261=120 model evidence and records
its hardware limits. Manufacturer identity must also match CV8=97. A type byte never establishes
firmware, fitted consumers or actual wiring. Indexed copies do not supply base model evidence.

The January2026 manual hardware page18 and appendix3 on page105 were visually checked against
the existing reviewed same-day PDF. The official online manual, CV table and product page were
checked; the reviewed PDF hash remains
`e919dfc7bc42a6f1135fd4698caa6d979c48df84b24db5cf1e9b12f17c6d9601`.
The source rules remain limited to the DCC table on pages65–80, not the SX2 parameter table.

DH12A AUX3–6 are logic-level outputs. AUX3/AUX4 have separate connections, while ZCLK/ZDAT
are shared between SUSI and additional logic outputs. CV137 bits0/4/5 must be read together using
appendix3; assigning ZCLK/ZDAT does not change the other pads or connector outputs. Default0 is
retained. DH16A's amplified AUX3/AUX4 and compact-model generation restrictions do not apply.
Output mapping masks, timers, conditions, initial/dimming masks and dimming descriptions retain
these limits. No output-load or independent-dimming detector is introduced. Key mode and key
inversion remain unknown; manual functions and import provenance are preserved.

CV48 and the CV154/155 brake-ramp rules use this model's directly located definitions. Unreviewed
or conflicting CVs remain excluded as listed above, without borrowing foundation curve points.
Exact previous inventory revision2026-10-20.5 is trusted for bundled upgrades; edited metadata and
operator settings remain protected. Existing profiles, variants, manufacturer reference files,
schema71 and backup24 are unchanged. No real DH12A export has been supplied for hardware comparison.


## Exact DH14B mTc14 extension

DH14B has120 located DCC register definitions from firmware3.15, with the exact model identity
CV8=97/CV261=141. The existing mTc14 foil connector and DH14B order number remain unchanged.
Inventory revision2026-10-20.7 adds only its profile link, identification rule and hardware note;
exact previous inventory revision2026-10-20.6 is trusted for bundled upgrades. Local metadata and
operator settings are preserved. Other profiles, reference files, schema71 and backup24 are unchanged.

Primary evidence is the [January2026 manual](https://doehler-haass.de/cms/media/pdf/DH05C_DH10C_DH16A.pdf),
particularly hardware page19, the DCC table on pages65–80 and appendix3 on page105, together with
the [DCC CV table](https://doehler-haass.de/cms/pages/haeufige-fragen/decoder-cv-tabelle.php).
The reviewed same-day PDF was reused after checking the official online sources; its SHA256
remains `e919dfc7bc42a6f1135fd4698caa6d979c48df84b24db5cf1e9b12f17c6d9601`.
Hardware page19 and the appendix3 selection table were visually checked.

DH14B has amplified LV/LR/AUX1/AUX2 and logic-level AUX3/AUX4 on ZCLK/ZDAT, shared with SUSI.
Unlike DH12A, no separate AUX3/AUX4 connections are documented. Appendix3 expressly provides an
alternative AUX5/AUX6 logical assignment on those same pins. Thus the eight AUX5/AUX6-specific
register definitions remain documented, subject to the shared-pin condition; they never establish
additional physical connections, concurrent AUX3–6 use, attached consumers or independent dimming.
Mapping, output-off masks, timers, conditions, initial/dimming masks and dimming explanations carry
this qualification. Register counts are not hardware output counts.

CV137 bits0/4/5 must be considered together: their contribution0/32 means SUSI without extended
mapping,16/48 SUSI with extended mapping,1/17 logic AUX3/AUX4 without/with extended mapping,
49 logic AUX5/AUX6 with extended mapping,33 SUSI disabled without extended mapping. Other bits
are separate. The bit0 labels therefore defer to the joint selection rather than asserting SUSI
from one bit alone. Factory reference0 does not supply a missing reading. Bit2 only inverts SUSI
direction; masks and timers do not establish function-key mode or inversion.

Historical status is retained. The [manufacturer delivery list](https://doehler-haass.de/cms/pages/vertrieb/liefertermine.php)
(stand30August2026) marks DH14B as an outgoing product with remaining stock; the cached product
page says unavailable, while direct retrieval returned503 during review. No unconditional current
availability statement is imported. The manual's VS/buffer connection restriction is hardware
documentation, not a RailKeeper wiring instruction or hardware-write capability.

Known/missing/indexed/conflicting readings retain the existing conservative rules; CV48 is a
parameter, not invented characteristic points. Manual vehicle/functions, source lines and export
timestamps remain unchanged. No real DH14B CV export has been supplied for hardware comparison.


## Exact DH21A generation and connection review

DH21A has121 located DCC register definitions from firmware3.15. CV8=97/CV261=200 identifies
the exact model; CV261=201 belongs to DH21B. Variant labels -0/-2/-3/-4/-5 select the DH21A profile
without identifying generation, actual wiring or fitted consumers. The existing five article/interface
variants remain intact. Inventory2026-10-21.1 links the profile, model evidence and qualified hardware
notes. Exact previous inventory2026-10-20.7 is trusted for bundled upgrades; local metadata and
operator settings remain protected. Existing profiles, references, schema71 and backup24 are unchanged.

The [January2026 manual](https://doehler-haass.de/cms/media/pdf/DH05C_DH10C_DH16A.pdf),
[DCC CV table](https://doehler-haass.de/cms/pages/haeufige-fragen/decoder-cv-tabelle.php) and
[DH21A generation2 product page](https://doehler-haass.de/cms/pages/produkte/fahrzeugdecoder/dh21a.php)
were downloaded again on2026-10-21. PDF SHA256 remains
`e919dfc7bc42a6f1135fd4698caa6d979c48df84b24db5cf1e9b12f17c6d9601`;
CV-table HTML SHA256 is`170b5b191845fd445daf2a9fbbee8a2e040142d644e9363d58359b9423637be4`;
product HTML SHA256 is`9945e3ae840ae28d1e0e778c80d1df3dafa938e41675dbad370f5aed20ec1f63`.
The HTML byte hash changed since the previous review, but extracted CV-table text is identical.
Hardware pages24–27 were visually inspected. DCC definitions refer to pages65–80 and appendix3,
not SX2 parameters. Sources retain retrieval dates and byte hashes in profile metadata.

AUX3/AUX4 are amplified on DH21A-4 in both documented generations and generation2 -0/-2/-3;
DH21A-5 uses logic-level outputs. Therefore order code-4 alone does not identify generation, but
the two manual hardware tables agree about its AUX3/AUX4 amplification. The former broad note
that it cannot establish any output characteristics has been narrowed accordingly.

AUX5/AUX6 have a genuine source conflict: generation2 table p.26 lists1A for -0/-2/-3/-4, while
the p.27 pin legend and product page call ZCLK/ZDAT alternatives unamplified. No typo correction
or hidden hardware explanation is guessed. The affected output-level claim stays unresolved in
mapping masks, timers, output-off conditions, initial/dimming masks and dimming descriptions.
CV137 records variant distinctions and this conflict, requires joint bits0/4/5 selection and keeps
its factory reference0 separate from a reading. SUSI direction inversion does not imply key inversion.

CV144 is partially documented: PDF p.76 and the online table agree about range0–63, factory0 and
bits1–5 (immediate start, analog lighting, GPIO section output/input and short-term storage). Only
those bits receive meanings. Bit0 remains unexplained because the dynamic-channel wording and
web reference to CV29 bit2 do not agree; CV29 bit2 remains analog permission. GPIO/AUX8 is shown
on generation2 pp.26–27, not on generation1 pp.24–25. No active GPIO-brake state, generation,
AUX8 key mapping or wiring is inferred. CV98/99/158/159 remain excluded despite the GPIO/AUX8
hardware note; their table scope names other models. Record counts include this partial CV144,
not full interpretation of every bit. Earlier profile definitions remain unchanged.

The hardware documentation includes AC-analog capability for both generations. CV12 explains
MM/internal direction settings without declaring the connected supply or current analog operation.
CV48 and CV154/155 remain the documented characteristic/brake-ramp rules. Raw values, index
identities, import provenance and manual functions are preserved. A real DH21A export has not
been supplied for hardware comparison.


## DH21B, DH22A, DH22B, DH24A and DHSP10A

The shared January2026 manufacturer manual, hardware pp.28–41, DCC pp.65–80 and appendix3 p.105,
provides bounded exact profiles for these five additional models. Hardware tables were visually
checked. The PDF hash remains `e919dfc7bc42a6f1135fd4698caa6d979c48df84b24db5cf1e9b12f17c6d9601`;
the online CV table checked21October has SHA256
`170b5b191845fd445daf2a9fbbee8a2e040142d644e9363d58359b9423637be4`.

| Model | CV261 (CV8=97 required) | Definitions | Hardware qualification |
| --- | ---: | ---: | --- |
| DH21B | 201 | 126 | -4 AUX3/4 amplified, -5 logic; AUX5–8 logic; SUSI shares AUX11/12 |
| DH22A | 202 | 120 | Both generations AUX3/4 amplified; AUX5/6 logic and shared with SUSI |
| DH22B | 203 | 126 | AUX3–7 amplified, AUX8 logic; SUSI shares AUX11/12 |
| DH24A | 204 | 127 | E24; AUX3/4 logic, AUX5–8 amplified; SUSI shares AUX11/12; AUX10 logic |
| DHSP10A | 110 | 120 | AUX3/4 amplified; AUX5/6 logic and shared with SUSI; buffer is hardware |

Firmware scope remains3.15. No firmware/generation/output load is identified from a type code.
CV137 bit5 is explicitly irrelevant to DH21B/DH22B/DH24A and omitted from their bit labels;
the generic appendix AUX3/4-versus-AUX5/6 selector cannot replace these hardware assignments.
CV98/99/102/103 key-number assignment (0 disabled,1–28=F1–F28,29=F0) applies to those three
models. CV101 AUX10 is included only for DH24A, whose hardware table actually offers this output.
CV158/159 carry additive AUX7/8 conditions, located on pp.78/79 respectively. Neither output
assignment nor condition registers establish key mode or inversion. CV144 is deliberately not
copied from the partial DH21A evidence. The existing model-specific motor/brake rules retain their
unknown/invalid/indexed and manual-data protections.

Inventory2026-10-21.2 changes only these five profile links/notes/identification fields. Existing
variants and sources in other profiles remain intact. Exact inventory2026-10-21.1 is a trusted
bundled-upgrade fingerprint; operator-edited metadata remains protected. The library contains20
profiles including DCC foundation,2,091 definition records (2,088 functional and3 reserved), and19
catalogue models with reviewed functional profiles. Manufacturer-reference coverage remains532/599;
those documentary references are qualified separately and do not enable automatic interpretation.
No real export or hardware comparison for these five models has been provided.
