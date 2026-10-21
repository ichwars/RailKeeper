# Roco 10882: interpretation quality

The exact Roco 10882 profile contains 35 located definition groups covering 62 unindexed addresses,
expanded from 17. Its manufacturer reference has 22 rows, expanded from 20. Total catalogue:
693 groups (690 functional and three ESU reservations), 284 reference sets and 29,221 rows for
532 of 599 models. The other profiles, references and inventory remain unchanged.

## Evidence and scope

Primary source: [Roco 10882/10883/10884 manual, 8010884920 VI/2012](https://www.roco.cc/documents/pdf/Anleitung_10882.pdf),
pages 2–7, downloaded and tables 4–7 visually inspected on 2026-10-20. PDF SHA256:
`a5fe8303c15db927b125ef4dfd15d0ffa7e4f3f404850537971b0cfe669b9883`.

`rules.roco10882` enables one fixed evaluator for the exact 10882 PluX16 model, family Roco10882,
with technical manufacturer identifier 145 (ZIMO). Go, TypeScript and OpenAPI share the flag;
foreign model algorithms and ABC/constant-distance/custom-curve metadata are rejected. Roco's
retail brand identifier 161 never substitutes for technical CV8. CV8 alone does not identify 10882;
conflicting ZIMO-family CV evidence disables this model interpretation. Other Roco models,
10883/10884, MX types and current MN decoders do not inherit the rule. Firmware bytes 7/65 stay
separate identifiers, without an invented firmware or hardware generation.

The settings directory and model details display the documentation profile and make it searchable,
without changing the inventory identity link. The definition table for this model and76420 does
not merge undocumented foundation registers back into their explicitly bounded definitions.

## Reviewed facts and source conflicts

- CV33–46 use shifted masks: CV33–37 have bits0–7 for front/rear lights and AUX1–6;
  CV38–40 have bits0–5 for AUX2–7; CV41–46 have bits0–2 for AUX5–7. The 10882 physically
  offers front/rear light and AUX1/AUX2 only (page3). Unavailable outputs and set bits in grey
  table cells get explicit warnings. Shared defaults at CV37 onward reference missing outputs
  or contradict the grey fields, so they do not become model-specific factory values. No
  consumer name, symbol, key mode or inversion is inferred. Technical mapping never modifies
  manual vehicle functions.
- CV124 bit5 permits braking recognition with analog operation disabled by CV29 bit2=0.
  Page7 incorrectly labels analog permission as bit3 while still giving value4. The detailed
  page5 table and page6 prose establish bit2, so RailCom bit3 is never substituted. CV112 bit6
  selects the DC or AC/Märklin mode using the manual's terminology. The UI distinguishes
  stored type from activation and qualifies this terminology as Roco's documentation. Neither
  ABC nor constant braking distance or consist parameters are established by this source.
- CV28 bits0/1 select broadcast and CV-data transmission; CV29 bit3 also permits RailCom.
  These are configured options, not proof of an active connection or RailComPlus. CV60 zero
  means no dimming; other values remain raw. CV114 bits0–3 exempt the four available outputs.
  Higher bits concern unavailable outputs. CV152 belongs exclusively to10883 and is removed
  from this model's manufacturer reference. Page2 records a tip to increase CV51 by8
  for directional DELTA lighting, without a complete register layout. CV51 stays raw; no
  current enable state, bit mask or factory default is inferred.
- CV155/156 values1–12 identify shunting and delay-disable keys. The manual does not explain
  zero unambiguously, so zero is not guessed to mean F0 or an active assignment. No current
  ZIMO MN interpretation is borrowed. CV3/4 have no verified seconds conversion in this edition.
- CV29 selects the explicitly documented three-point CV2/5/6 characteristic or stored CV67–94
  table. The graph shows only the selected valid points, without falling back to an inactive
  or unknown curve. CV2/5 use1–255 and CV6 uses1–252. No modern ZIMO special zero/one values,
  CV66/95 trims or smoothing are borrowed. The table registers are explicitly referenced by
  CV29, but their values/defaults are not printed; the profile's0–255 limits are CV-byte limits,
  not a claim that this manufacturer table prints every range and default.

The raw reference corrects extracted factory-column values that had mistakenly repeated the
range column. Existing documentary defaults (3,4,6,2,252,85, etc.) are preserved as reference
values, never inserted into vehicle readings. Every definition has a specific source location.
Missing/invalid/out-of-range/indexed values remain unknown. Raw bytes, index identity, source
lines, export dates and manual fields are unchanged. Exact previous released profile/reference
fingerprints allow startup and restore upgrades while protecting operator metadata, custom
origins, names, ordering, activation and source URLs. Schema71 and backup24 are unchanged.
