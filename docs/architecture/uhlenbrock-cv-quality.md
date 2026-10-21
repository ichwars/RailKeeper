# Uhlenbrock IntelliDrive 76420: interpretation quality

The exact76420 profile contains43 located definition groups, covering70 unindexed addresses.
The separate manufacturer reference contains34 rows, grouping33–42 and67–94. The catalogue
now contains675 groups (672 functional and three ESU reservations),284 reference sets and29,219
reference rows for the same532/599 models. Inventory and other profiles/references are unchanged.

## Evidence and conflicts

Primary source: [German instructions, 11.10 ru](https://www.uhlenbrock.de/de_DE/service/download/handbook/de/archiv/Bes76420.pdf),
pp.2,4–6. PDF SHA256: `afd8ec41bf923ed7bf9a34b9900cdeb289ca44cf456ab48b730bd6cd985c46c3`.
Both original PDFs were downloaded; German table pages5/6 were visually inspected. The
[English comparison copy](https://www.uhlenbrock.de/de_DE/service/download/handbook/en/Bes76420e.pdf)
is retained as source0, but source1 (German) supplies the located definitions. Retrieval:2026-10-20.

The German original resolves four English inconsistencies: CV17 high byte192–231 (not199–231),
CV55 I constant (factory10), CV56 control reference (factory32; not regulation frequency), and
CV98=0 without output time limitation. The CV7 software byte does not establish a full firmware
version or compatible newer variants. References remain marked partial because the bounded
manual evidence does not claim all firmware states.

## Scope and evaluation

`rules.uhlenbrock76420` enables only manufacturer85, family IntelliDrive, exact model IntelliDrive76420.
The Go validator rejects mixed foreign rules, model/family/manufacturer changes and generic ABC,
constant-distance or one-bit analog rules. Go, TypeScript and OpenAPI share the flag. CV8=85
alone does not identify this model. The reviewed rules never transfer to76425,76320 or other
IntelliDrive models. Non-DCC readings, indexed values (including0/0), manufacturer conflicts,
missing or invalid values do not activate guessed semantics. Undocumented DCC-parent registers
are not inherited for a verified76420.

- CV29 bits0/1/2/4/5 select direction, speed steps, analog permission, curve and address.
  Bits3/6/7 are unused: no RailCom flag is inferred. CV17/18 are added; the address summary
  honors the documented long-address range128–9999. CV19 bit7 reverses consist direction,
  independently of function-key inversion. CV21/22 permissions are not established.
- CV49 bit0 has reversed enable polarity for load regulation; CV53–58 remain tuning parameters.
  CV49 bits3/4 independently block Motorola/DCC; both set explicitly blocks speed commands,
  while programming is still possible. Bit1 selects SUSI sound or LISSY Mini, without asserting
  attached hardware. CV60 accepts only0/9; CV61 is a configured threshold, not a measured temperature.
- DC analog requires CV29 bit2 and CV51=2/3; AC requires the same permission and CV51=1/3.
  Märklin braking requires CV29 bit2 and CV49 bit7. CV49 bit2 selects braking toCV52 rather
  than zero. Neither ABC nor a constant braking distance is established by these settings.
- CV33–42 use one uniform six-bit mask: front/rear light output, A1, A2, shunting and delay
  control. F0 has forward/reverse rows; F1–F8 have separate masks. No shifted Kühn positions,
  fitted-load names, symbol, latching/momentary key mode or inversion is inferred. Technical
  mapping is visible without creating or modifying vehicle functions. CV49 bit6 swaps light
  connections; this is displayed independently of key inversion and travel direction.
- CV98 selects timed A1/A2 outputs, CV99 is100ms increments. The UI distinguishes stored duration
  from the selected outputs; zero duration gets no undocumented special meaning. A timed output
  is not a momentary function key. A1/A2 dimming requires CV49 bit5 and usesCV50.
- CV2/5/6 use1–63 and require CV2 < CV6 < CV5. No unproven complete interpolation or overall
  acceleration-time formula is plotted. Only the selected CV67–94 table is drawn, with stored
  points0–255. CV47/48 are raw correction parameters0–63; no factor/percentage is guessed.
  CV65/66 are Motorola programming helpers, not DCC bank selectors or speed trims.

Original readings, index identity, export provenance and manual vehicle/function fields are
unchanged. Exact previous bundled profile/reference fingerprints allow startup and restore
upgrades; operator metadata edits, custom origins, labels, ordering, activation and source URLs
remain protected. Schema71 and backup24 are unchanged. Tests cover scope, unknown states,
byte/allowed-value bounds, source conflicts, conditional interpretation, mappings and preservation.
