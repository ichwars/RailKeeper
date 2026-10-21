# ZIMO MN180N18: interpretation quality

The MN180N18 profile contains 68 definition records, expanded from 21. All reference located
rows in the official [MS/MN/FS manual](https://www.zimo.at/web2010/documents/MS-MN-Decoder.pdf),
SW5.29, edition listed as 2026-07-01. The reviewed PDF was retrieved on 2026-10-20; SHA-256:
`2e40762b6f4dcd5d761fe1507d70c8b997657b2fbe98d480ce63993f01f5ec43`.
The former source date 2026-07-24 was not supported by the title page and has been corrected.

## Scope and identity

The profile's reviewed firmware scope now starts at 5.29 rather than 4.229. This identifies the
source scope, not the firmware version at which each feature was introduced. Known earlier
firmware falls back to DCC foundation meanings; unknown firmware remains unconfirmed, with the
existing source-scope notice. CV7/65/839 describe main/sub/patch identifiers without introducing
automatic numeric ZIMO firmware detection. CV248/249/842 describe the separate bootloader.

CV8=145 and CV250=127 still identify the MN180 base type, not the N18 interface. Model interpretation
requires the matching MN180N18 profile name. Unindexed registers above 256 never acquire CV31/32
pages. Read-only identity bytes, patch registers and exported measurements have no guessed
factory defaults. The catalogue inventory and manufacturer-reference corpus remain unchanged.

## Reviewed distinctions

- CV2 is an internal speed step, not a voltage measurement. CV3/4 specify total ramp time at
  0.9 seconds per increment; progression follows the selected speed characteristic.
- CV5=0/1 each mean 255; CV6=1 selects the smoothed default characteristic, with the middle
  speed approximately one quarter of maximum. CV6=0 is undocumented. The speed view explains the effective maximum and middle parameter separately from the
  original bytes. It does not draw an invented three-point characteristic. A selected
  CV67–94 table can still be plotted from its stored points; missing CV29 does not select a curve.
- CV9 uses decimal digits for sampling and measurement gap; it is not a frequency in Hz.
  Its zero-description is internally contradictory, so zero is not normalized to 55.
  CV56 excludes exact tens 10–90 and maps 0 to 55. CV147/148/149 mean integral/differential/proportional;
  modifying CV56 also updates these parameters, but not vice versa.
- CV57=0 selects relative reference; 100–255 represent tenths of a volt. Direction trims 66/95
  preserve special value 0 without interpreting it as zero motor speed. The 28-point factory
  table is curved; no missing point is synthesized.
- CV12 DC permission also requires CV29 bit 2. A shared-table mfx bit does not prove mfx support
  for MN180N18. CV28 high-current bit 6 belongs to large-scale hardware; DCC-A bit 7 is not an
  ESU RailComPlus symbol-layout guarantee.
- CV15/16 are access-lock keys, not addresses. CV19 bit 7 reverses consist travel direction,
  not function keys. Long ZIMO consists require both CV19 and CV20. The summary masks both
  bit7 flags before applying the documented decimal extension; missing or undocumented bytes
  and addresses above10239 remain unknown. CV22 bit 2 inconsistently names F0 in both tables, so no manufacturer
  bit explanation is assigned; bit 6 is auto-consist and bit 7 groups F13–F27.
- CV33–46 are output assignments. Higher-function bit positions change with CV61; unconditional
  output-bit names are deliberately omitted. These masks and CV157's MAN override key do not
  establish latching/momentary operation or global inversion. No ZIMO counterpart to the ESU
  checkbox-export evidence is claimed.
- CV100 is an exported PoM measurement with signed direction encoding, not a live view or the
  threshold. CV134 is the configured asymmetry threshold. CV193 enables ABC shuttle operation
  and gives reversing dwell; CV194=255 is an indefinite intermediate stop, not 255 seconds.
- CV140 is a finite mode selector, not a bitmask. CV141 is a calibrated distance parameter;
  CV830–833 only apply at CV141=0, as directional high/low-byte pairs at 16 times the resolution.
  The configured-state summary uses the finite CV140 mode, independently of parameter
  completeness. Each direction remains unknown until its own required bytes exist. No exact
  physical distance or live braking state is inferred.
- CV153 controls time without DCC reception until braking starts, not time until standstill.

## Deliberately unresolved

The manual explicitly includes future SW6.00 content. CV58/123/151 and sound/large-scale/MX-only
features are not promoted. HLU has contradictory status: CV27 calls it active by default while
the dedicated chapter is marked SW6.00. Its activation bit receives no affirmative bit explanation;
related parameters retain conditional wording. CV27's range cell conflicts with its own bit
combinations, so no finite mask-derived allowed set is invented. Ambiguous or absent defaults
remain absent. Only clearly located factual rows are included; the 68 records do not establish
complete firmware, hardware or function-key coverage.

All stored readings, invalid values, index identities and import provenance remain separate from
reference meanings. Startup/restore upgrade only exact recognized bundled metadata; operator
edits and settings remain protected. Existing schema 71 and backup 24 are unchanged.

## Bounded summary algorithm

`rules.zimoMN=true` selects the reviewed MN180N18 calculation, not an executable expression.
Validation requires ZIMO manufacturer145, MN family, modelMN180N18 and source scope at least5.29;
it rejects a simultaneous scalar motor characteristic, generic constant-brake rule or mapping
layout. This is separate from base-type identification and remains gated by an active matched
profile and the existing firmware compatibility check. The known bundled profile revision is
upgraded with the same local-edit protection as other decoder metadata.

The address calculation is `(CV20 &127) ×100 + (CV19 &127)`; CV19 bit7 supplies travel direction,
CV20 bit7 controls RailCom. CV20's documented value set0–102/128–130 and the exclusion of CV19=128
are respected. An absent CV20 is never replaced with a factory zero. CV22 bit7 expands permissions
F13–F27; bit6 is reported as auto-consist, not a function. With ambiguous bit2 set, known permissions
are still displayed, with an explicit unresolved-permission note instead of a guessed F9.

DC analog requires both CV29 bit2 and CV12 bit0. CV12=0 is a special unstable setting in the
source and therefore does not establish a reliable all-formats-off state. Brake mode0 is off;
1/2/3/11/12/13 are configured on, with separate trigger/onset explanations. Positive CV141 applies
to both directions. Only CV141=0 selects the independent CV830/831 and CV832/833 byte pairs;
no missing direction is copied from the other. Parameters are labelled in CV141 or1/16-CV141
increments, never metres. These read-only summaries do not fill manual vehicle fields or alter
raw CVs/provenance. No schema or backup version change is required.
