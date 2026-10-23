---
title: Model and firmware notes
description: Model and firmware notes.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Model and firmware notes

<GuideFigure name="decoder-library" alt="Decoder library with model selection and sources." caption="Rules apply within their stated source and firmware scope." />

## Digital settings in the vehicle dialog

**Model → Digital** retains the editable digital address, separate DT double-heading address,
decoder type and manual ABC setting. The selected decoder profile supplies the active DCC address,
consist address/direction, enabled consist functions (CV21/CV22), ABC and constant braking distance,
with their source CVs. Missing values stay unknown; profiles and indexed pages are never merged.
**Use CV settings** copies the known address, recognized decoder type and ABC state into the form;
**Save changes** persists these fields. DT is never copied from CV19.

The **CV** tab separates **Explained**, **All CV values** and **Files & import**.
**Import CV list** handles the complete structured values; **Upload decoder project** stores the
original ESUX or other project file. The raw table supports profile, index and nonzero filters;
`CV29` searches for that exact CV number; meanings are also searchable. **Export CV values** is
available in this view. The shared decoder strip shows profile, active address, value count and
index pages. **Explained** aligns topic, description, reading and source; source links open a right-hand panel for exactly one CV. It shows decimal, hexadecimal and
binary values, index page, decoder profile and actually stored provenance. Bit explanations appear
only for base CV29. **Edit CV value** opens the same record, including its index page, in the editor.
CV imports preserve the physical source line (including headers and blank lines) and the export
timestamp from the TXT header, shown in the detail panel. The timestamp retains its source wall-clock
time without an assumed timezone. JSON exports and backups include these fields; edits or imports
without provenance preserve existing metadata. Older records without a source still show “Not recorded”;
the record timestamp is never used as the export date. Backup format 22 preserves provenance and older
backups remain readable. Additional consist/decoder settings are collapsed. ESU DC
analog permission is derived from CV29 and CV50 together. **Files & import** explains the two file types and lists imported records and
stored original files. Basic explanations and individual CV29 bits complement
the raw values. On reimport, the TXT decoder header is preserved as CV7 provenance and supplies
the full firmware version; CV7=255 alone is never treated as a firmware version.

For recognized ESU LokPilot 5 profiles (manufacturer 151), **Control** displays simple positive
key mappings from the indexed mapping pages, including direction, output and CV source. Suggestions fill missing details in the draft; **Save changes** saves function changes together. Missing fields and complex
conditions are not guessed. AUX3 output mode 31 is identified as PowerPack Control without claiming
that a PowerPack is installed; the actual consumers on AUX outputs require confirmation.

References: [ESU LokPilot 5, 8th edition](https://www.esu.eu/en/downloads/instruction-manuals/digital-decoders/?tx_esudownloads_pi1%5BdownloadItem%5D=0318b68e0cb850d9101947622d1041ea)
and [NMRA S-9.2.2](https://www.nmra.org/sites/default/files/s-9.2.2_2012_10.pdf).

For **ZIMO MN180N18**, 68 definition groups have located evidence, including decoder lock,
PID regulation, output assignment, ABC shuttle operation and stopping-distance registers.
CV5=0/1 each mean the highest internal step; CV6=1 selects a smoothed default characteristic.
CV140 is a mode selector (0/1/2/3/11/12/13), not a direction bitfield. Vehicle readings remain
unchanged. The manual also includes SW6.00 previews and contradictory rows; these do not
establish automatically confirmed functions, key modes or inversion. The consist address requires CV19 and CV20 together. RailCom and direction flags stay
separate; an absent CV20 keeps the address unknown. Grouped F13–F27 permissions, auto-consist
and the unresolved CV22 bit2 are distinguished.

The braking view shows CV140 mode and separate directional parameters: positive CV141 applies
to both directions; CV141=0 selects CV830/831 and CV832/833 byte pairs. A missing byte leaves
that direction unknown. Parameters are not measured metres. The speed view explains smoothed
ZIMO support parameters alongside unchanged raw bytes. A28-point table is plotted only when
selected by CV29. DC analog status uses CV12 and CV29 together.

For **Lenz STANDARD+ V2**,77 located definition groups are reviewed. Evaluation separates
ABC constant braking (CV51 bits0/1) from speed-step0 braking (bit7). CV52 remains a parameter
without invented metres. DC stops require CV51 bit5 and DC operation disabled in CV29.
Selecting both shuttle modes leaves the mode unknown.

Outputs A–D and their keys F0–F28 are shown with direction and source CVs. Bit4 remains
unconfirmed and never establishes output E. RailCom requires CV29 permission and enabled
channels in CV28. Motor settings, stored fault flags, decimal effects and coupler timing have
separate explanations. No key mode or inversion is taken over. The speed view shows raw
internal-smoothing parameters or an explicitly selected CV67–94 table. Existing stored readings
do not need to be imported again.


## Kühn N45: reviewed interpretation

The N45 profile has **59 definition groups** with manufacturer manual locations, including
CV55 and CV60–63. Interpretation covers motor frequency/load regulation, ABC/DC braking options,
decimal dimming/coupler parameters and F0–F12 output assignments. Output bit positions shift at
F4 and F9. E/F are logic-level and blocked by enabled SUSI; without CV56 their availability stays
unknown. No key mode, inversion or fitted loads are inferred, and vehicle functions remain unchanged.

The speed view retains CV5/CV6 special values as raw parameters without inventing an exact
characteristic. Only an explicitly selected stored table is plotted. The manual specifies14/28-step
use, not128; CV29 alone cannot distinguish28/128 steps.

Manual inconsistencies (CV49 versus CV51/52 and the CV112 range versus bit5) remain visible in
the help. The official web-readable text was reviewed; no PDF hash is claimed because the file
could not be downloaded. These rules cover N45/N45-P/N45-P8, not N45-18, N025, T62, T65,
T125/T145 or F060. Their own reference tables remain separate. Stored readings do not need reimport.

### Uhlenbrock IntelliDrive 76420

For the exact **76420** model with CV8=85, interpretation follows the German original manual.
**Address & operation** combines CV29 bit2 with CV51 for DC analog permission; CV29 bit3 is
unused and is not interpreted as RailCom. CV19 documents consist address and reversed consist
direction. Consist function-key permissions remain unknown without manufacturer evidence.

**Driving behaviour**, **Braking** and **Outputs & keys** explain load regulation CV49/53–58,
the conditions for Märklin DC braking, the braking target CV49/52, and uniform CV33–42 function
masks for F0–F8. CV98 selects timed outputs A1/A2 and CV99 sets their duration in100ms increments.
Output timing does not establish momentary key mode or inversion; manual functions are preserved.
The light wiring swap in CV49 bit6 is shown separately.

In **Speed curve**, CV2/5/6 remain raw parameters on their1–63 scale. The stored28-point table
is drawn only when selected by CV29 bit4. Speed corrections are CV47/48; CV66 is a Motorola
programming page on this model. Missing or invalid readings are never replaced with factory
settings. Each explanation links to source CVs and a located manufacturer reference; original
export dates and line numbers remain intact.

### Roco 10882 (2012 manual, ZIMO)

The reviewed profile requires technical ZIMO identifier **CV8=145** and the matching model name.
Roco brand identifier161 or CV8 alone is insufficient. Current ZIMO MN rules are not substituted;
software bytes CV7/65 remain separate identifiers.

**Braking** shows permission from CV29 bit2=0 and CV124 bit5=1, and the stored type from CV112
bit6 using the manual's terminology. The incorrect bit number in its summary is explained.
ABC, constant braking distance and consist parameters remain unknown without matching evidence.

**Outputs & keys** follows the shifted CV33–46 bit positions. The10882 only has front/rear lights
and AUX1/AUX2; assignments to other outputs or grey table cells receive explicit warnings.
CV60/114 dimming, the CV51 DELTA lighting note and documented CV155/156 key assignments are
explained. Zero in CV155/156 remains unknown. The display does not modify manual functions or
infer key mode/inversion.

**Speed curve** draws only the stored valid points selected by CV29. The source drawer separates
original line, export date and raw value from documentary factory settings and manufacturer
references. Existing readings are re-evaluated without another import.

### PIKO 56500 PluX22

An exact PIKO56500 match enables simple mapping (CV96=1 only), ABC, constant braking-distance
and output-parameter explanations for index18/0. CV141 distance alone does not activate braking:
CV145 selects triggers, ABC/DC additionally require CV27, and CV142 selects a separate test key.
Missing prerequisites remain unknown.

**Outputs & keys** shows both parameter sets A/B for front/rear light and A1–A7, separating
effects, PWM and automatic switch-off. Auto-off requires direction and speed flags. Smoke
effects40/41 do not use flags. Unavailable outputs are marked. Advanced mapping (CV96=6)
never falls back to simple assignments. Connected loads, symbols, key mode and inversion remain
unknown; manual functions are preserved. The source button opens the exact stored bank value,
including original line and export timestamp.

Manufacturer documents conflict at CV101 and analog thresholds CV102–104. Without verified
firmware context, their functions, units and defaults remain unresolved; CV101 is not assigned.
The library records the sources and conflicts. The speed graph only shows the characteristic
selected by CV29, without falling back to an inactive curve.

### D&H DH05C and DH10C

DH05C and DH10C have separate reviewed profiles from firmware3.15. CV8=97 plus unindexed CV261
identifies DH05C at52 and DH10C at102. An indexed value with the same number is not model evidence.
The analysis uses the D&H CV48 motor characteristic instead of a guessed DCC three-point curve,
and shows the CV154/155 brake-ramp configuration.

The type code does not identify hardware generation. AUX3–6 and SUSI explanations therefore
retain the manual's hardware limits. CV137 has factory reference0 here, compared with1 on DH18A.
Output masks, switch-off timers and SUSI direction inversion establish neither key mode nor key
inversion. Manual function data is preserved. Address/consist registers have direct manufacturer
references; raw values and provenance remain unchanged. Arbitrary CV7 bytes never establish
an invented firmware version.

### D&H DH16A

DH16A has its own reviewed profile with120 CV definitions from firmware3.15. CV8=97 plus
unindexed CV261=160 identifies the model even without an exported profile name. Hardware generation
remains unknown. Stored readings supply address, consist, ABC and brake-ramp explanations;
CV48 remains a characteristic parameter without invented curve points.

The manual specifies amplified AUX3/AUX4 in both generations. AUX5/AUX6 are logic-level outputs
sharing pins with SUSI, so check CV137 and actual wiring. The documented CV137 factory reference
is0; DH18A's special1 does not apply. Output conditions and SUSI direction inversion establish
no function-key inversion. Manual data, original source line and export timestamp remain unchanged.


### D&H DH12A (historical PluX12 decoder)

The discontinued DH12A has its own reviewed profile with120 CV definitions from firmware3.15.
CV8=97 plus unindexed CV261=120 identifies the model even without an exported profile name.
Unknown firmware is not guessed. Stored readings supply address, consist, ABC and brake-ramp
explanations; CV48 remains a characteristic parameter without invented curve points.

DH12A AUX3–6 are logic-level outputs. AUX3/AUX4 have separate connections; ZCLK/ZDAT alternatively
provide SUSI or logic outputs. Observe CV137 bits0/4/5 together and check actual wiring. The
documented factory reference0 is not an imported vehicle reading. Output timers, conditions and
SUSI direction inversion establish neither function-key mode nor inversion. Manual data,
original source line and export timestamp remain unchanged.


### D&H DH14B (mTc14)

DH14B has its own reviewed profile with120 CV definitions from firmware3.15. CV8=97 plus
unindexed CV261=141 establishes the model. Stored readings supply address, consist, ABC and
brake-ramp explanations; CV48 remains a parameter without invented characteristic points.

Logic-level AUX3/AUX4 share ZCLK/ZDAT with SUSI. AUX5/AUX6 registers describe alternative logical
assignments on those same pins, not additional connections. CV137 bits0/4/5 must be considered
together; bit0 alone cannot establish SUSI state. The library lists the documented combinations
and records hardware limits beside the output CVs. These do not establish key mode, key inversion
or attached loads. Manual data, original source line and export timestamp remain unchanged.


### D&H DH21A (generations and connection variants)

CV8=97 plus unindexed CV261=200 identifies DH21A. Its reviewed profile contains121 register
definitions from firmware3.15, including partially explained CV144. DH21A-0/-2/-3/-4/-5 labels
select the same profile without establishing hardware generation.

AUX3/AUX4 are amplified on -4 in both generations and logic-level on -5. Some AUX5/AUX6 level
specifications in the manual conflict; these levels explicitly remain unresolved. CV137 and the
output-CV explanations record the variants and limits. No fitted consumers or function-key mode
are inferred.

CV144 explains corroborated bits1–5 only; bit0 remains open. GPIO/AUX8 is documented for generation2
only. Stored GPIO configuration does not establish brake wiring. Documented AC analog capability
does not identify the currently connected supply. Source lines, export timestamps and manual
data remain unchanged.


### Additional D&H models and manufacturer references

DH21B, DH22A, DH22B, DH24A and DHSP10A have separate reviewed profiles from firmware3.15.
CV8=97 and unindexed CV261=201/202/203/204/110 identify the corresponding model. Connection
variant, generation, firmware and actual attached loads remain unknown. SUSI shares AUX11/12
on DH21B/DH22B/DH24A; the generic CV137 bit5 selector does not apply. CV101 is documented only
for DH24A with its actual AUX10. Output conditions and key-number assignments do not establish
function-key mode or inversion.

For other exactly named or identified models, the value list and CV details show additional
**manufacturer references** from stored CV tables. These headings are explicitly documentary;
model/firmware qualifications and located sources remain visible in the drawer. Factory references
are separate from actual readings. Hints do not activate automatic interpretation or modify manual
entries. Reviewed explanations and operator descriptions retain priority. Unknown index pages,
conflicting associations and missing documents remain unresolved. No repeat CV import is needed.
