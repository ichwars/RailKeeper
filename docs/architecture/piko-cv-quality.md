# PIKO56500: interpretation quality

The exact PIKO56500 PluX22 profile has111 located definition groups, covering218 CV addresses
(previously58 groups). Ten stride-based groups cover all90 indexed output parameters. The
manufacturer reference retains201 PDF rows and adds3 distinct firmware-qualified FAQ facts:
204 rows. Catalogue total:746 groups (743 functional,3 ESU reservations),284 reference sets,
29,224 rows for532 of599 models. Other profiles and inventory identities are unchanged.

## Primary sources and limits

- [PIKO56500 detailed manual, PDF28620](https://www.piko-shop.de/is.php?id=28620),12 pages,
  downloaded2026-10-20; pages4,5,8,9,11 visually reviewed. SHA256:
  `9d3c587b4f44b94f2c77821152ae20a3b267d99342dcd633c8c00e027ebf8d3a`.
- [PIKO digital FAQ](https://www.piko-shop.de/de/fragen-zum-digitalsystem.html), checked2026-10-20,
  particularly2025-01-28 (firmware1.16.0/CV7=16) and2026-01-29 (1.19.0/CV7=19).

`rules.piko56500` is a fixed, model-limited Go/TypeScript/OpenAPI flag, requiring manufacturer162,
family SmartDecoder XP5.1 and the exact PluX22 model. It rejects mixed model algorithms and
unreviewed constant-brake/custom-curve metadata. CV8 alone is insufficient. Other XP hardware,
older4.1 decoders and sound variants never inherit these rules. Unreviewed DCC foundation
registers are not merged into this model's definitions, including the contradictory CV101.

## Reviewed interpretation

- Base values require complete index-free identity, integer/range validation and, for sparse
  masks, documented allowed bits. Long addresses are128–9999. CV12 operating permissions and
  CV29 analog permission are combined for DC/AC. RailCom permission (CV29 bit3), transmit
  configuration and RailComPlus permission (CV28 bit7) remain distinct. None is a live connection
  or an ESU RailComPlus name/icon/key-mode decoder. CV15/16 compare stored lock key/identifier;
  there is no actual programming-lock test or device write.
- Constant-distance braking uses CV145 trigger bits0–3; ABC/DC also require the corresponding
  CV27 permission. CV142 selects a separate test key (0 off;1–68 F1–F68). CV141 stores cm and
  cannot establish activation. Missing prerequisites and undocumented CV145 bits4/5 remain
  unknown unless a separate documented trigger establishes activation. CV48's approximate
  threshold is1.6V+0.1V per step; CV50 remains an internal speed parameter,51/52 use100ms steps.
- Simple mapping applies only with CV96=1;6 is advanced mapping and never falls back to the
  simple table. CV33–37 start at front light,38–42 at A2,43–46 at A5. The56500 has front/rear
  light and A1–A7 only; higher assignments are marked unavailable. No attached consumer,
  symbol, key mode or inversion is inferred. The technical mapping cannot alter manual keys.
  Logical-key registers154–157 accept0–68 and255 off; the printed maximum68 conflicts with
  factory255, which is explicitly documented as disabled. CV157's69–127 remain unresolved.
- Exact index18/0:9 outputs,10 CVs per output, two5-CV sets A/B (effect,PWM,flags,param1,param2).
  PWM uses0–64 without invented percentage conversion. Auto-off requires at least one
  direction flag0/1 AND one speed flag2/3/4. Timers180–188 use0.5s steps. Fade flags6/7 are
  distinct from function key inversion. Smoke40/41 ignore flags and repurpose parameters;
  missing/unknown effects do not acquire guessed flag semantics. Effect255 does not acquire
  a guessed active auto-off state. Other banks and unindexed bytes cannot substitute for18/0.
  The source drawer opens an actually present parameter even in an incomplete set.
- CV29 selects the3-point or28-point speed characteristic; no inactive or unknown curve is
  plotted as fallback. Trims66/95 remain raw documented bytes without an invented formula.

## Explicit conflicts and exclusions

The manual prints CV101 twice with incompatible meanings. Its CV102/103 analog assignments
conflict with the FAQ from firmware1.16.0, which assigns102 sound start,103 motor start,104
maximum speed. Original PDF rows are preserved with qualifiers and FAQ rows remain separate.
The evaluator deliberately does not infer a complete firmware from arbitrary CV7 bytes, nor
choose a threshold role/default/unit without verified firmware context. All three profile
entries explain the conflict rather than claiming an unconditional role.

CV170 has a0–64 prose range versus0–255 table range: no brightness conversion. CV182 has a
0–63 table range versus general0–255 prose: higher values remain raw, without a verified time.
Some A4/A5 parameter rows are mislabelled A2; the explicitly described10-CV block sequence
on page4 determines their scope. CV47's conditional SUSI/servo text is ambiguous and is not
expanded into executable mode logic. Sound-project and undocumented advanced-mapping rules
remain outside this evaluator. Raw references remain explicitly partial coverage.

Profile/reference revision2026-10-20.2 upgrades only exact previously bundled fingerprints.
Edited/custom metadata and operator labels, activation, ordering and source URLs are preserved.
Raw vehicle bytes, index pages, source lines, export dates and manual fields remain unchanged.
Schema71 and backup24 do not change. No real56500 CV export has yet been supplied for a hardware
comparison; automated and browser validation uses artificial fixtures.
