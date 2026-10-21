# Kühn N45: interpretation quality

The N45 profile contains 59 located definitions (86 unindexed addresses), expanded from 54.
Evidence is the manufacturer's [N45 manual, revision b](https://www.kuehn-digital.de/files/N45_de_b.pdf),
sections 4.3/4.4, pp.10–20. The manufacturer-hosted web-readable text was reviewed on 2026-10-20.
The PDF endpoint could not be downloaded or rendered, so no PDF hash or visual PDF verification
is claimed. Conflicting source statements remain visible instead of being silently repaired.

## Bounded scope

`rules.kuehnN45` enables one fixed evaluator for manufacturer157, exact family/model N45, with
ABC rule CV27/mask3. Go, TypeScript and OpenAPI share this flag; conflicting manufacturer rules
are rejected. The manual explicitly covers N45, N45-P and N45-P8, which have exact aliases.
CV8 alone does not identify N45, and CV7/CV127 are not invented model/connector identifiers or a
combined firmware version. Neither the Next18 variant N45-18 nor other Kühn models inherit these
rules. N025, T125/T145, T65 and F060 retain their separately sourced manufacturer reference tables;
T62 retains its explicit documentation gap. For example, N025 CV58 assigns output B, whereas
N45 CV58 assigns shunting/delay keys. A manufacturer-wide CV58 rule would be incorrect.

## Located meanings

- CV55 and60–63 are added to both the interpretation profile and the N45 reference table.
  These registers use decimal tens/units, not bit flags. CV60/61/62 separately dim B/A, D/C and
  F/E. CV63 uses pull-in time and holding level; holding digit8 turns the output off after pull-in,
  while digit9 is undocumented. No fitted coupler is inferred. CV55 timing endpoints do not
  establish a complete timing formula; CV123 only provides the example9=4.5s. Both stay raw.
- CV33–46 assign F0–F12 to A–F. The bit positions shift at F4 and F9, and F0 has separate
  forward/reverse rows. E/F are logic-level only and are blocked by CV56 bit2 when SUSI is enabled.
  Missing CV56 leaves their availability unknown. CV47–52 describe output effects and directional
  suppression. These are neither key inversion nor latching/momentary mode. CV58 separately
  assigns F1–F4 to shunting (bits4–7) and disabling delays (bits0–3); CV59 assigns dimming keys.
  The Control technical disclosure and CV explanations show assignments without modifying functions.
- CV9 selects approximately15.6kHz,120Hz or32kHz. CV56 bits0/1 enable the back-EMF divider and
  load regulation; CV53/54 are parameters, not enable flags. CV95 zero or128 means reverse factor1.
  CV113 uses the documented random-interval lookup. CV13 bits0–7 select F1–F8 in analog mode,
  whereas CV14 bits2–5 select F9–F12. F0 is always enabled in analog mode.
- CV27 bits0/1 permit ABC polarities; bits4/5 configure DC braking. Setting both DC bits blocks
  analog operation even when CV29 bit2 permits it. Missing inputs remain unknown. No constant
  braking distance is inferred. CV29 bit3 documents RailCom permission; the manual does not
  establish individual CV28 channel meanings. No Lenz channel mapping is borrowed.

## Source conflicts and presentation

The summary repeats CV49 for outputs C/D; the detailed table on p.16 establishes CV51/52.
CV112 is summarized as0–31, but p.19 explicitly documents bit5/value32 to disable short-term
memory. The evaluator recognizes that documented bit and retains the contradiction in its help.
The separate reference keeps the printed overview range. Unassigned/reserved bits get no
invented meaning. Consist direction and CV21/22 permissions are not established by this manual;
consist bytes above127 remain unknown in the address summary, while raw bytes remain accessible.

CV5/CV6 zero/one are special internal-curve selectors. The speed view retains raw parameters and
does not plot a guessed three-point characteristic. It plots stored CV67–94 points only when
CV29 selects the table, with the documented limitation that128-step operation ignores it. CV29
does not distinguish28 from128, so the UI does not claim the table is currently in use.
Missing/invalid configuration and empty selected tables never fall back to an inactive curve.

Original values, profile groups, index pages, source lines, export dates, manual fields and function
takeover are unchanged. The exact previous bundled profile/reference metadata is fingerprinted
for startup/restore upgrade; edits/custom origins and operator labels/order/activation/source URLs
stay protected. Schema71 and backup24 remain unchanged. Other profiles, catalogue models and
manufacturer references are unchanged. There are now670 definition groups (667 functional and
three ESU reservations),284 reference sets and29,216 reference rows for the same532/599 models.
