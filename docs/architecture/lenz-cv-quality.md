# Lenz STANDARD+ V2: interpretation quality

The reviewed profile has 77 located records (104 unindexed addresses), expanded from 64 records.
It requires CV8=99 and the exact model alias or the existing documented CV7 identifiers77/79.
CV7 is not a numerical firmware version. No rules are transferred to STANDARD, STANDARD+,
SILVER or GOLD models. Unknown firmware stays unknown.

Primary evidence is [STANDARD+ V2 10231-02, edition5, November2020](https://www.lenz-elektronik.de/media/be/06/d9/1734004977/b_10231-02.pdf?ts=1734004977),
pp.19–26 and sections5.2–5.13. Supplemental evidence is the manufacturer's
[Plus decoder handbook90008, edition4, October2017](https://www.lenz-elektronik.de/media/f4/c5/da/1733844005/b_90008_d.pdf),
especially pp.32–51 and63–67, linked from the current manufacturer FAQ. Stored source hashes
are `3f1e6f449e916f714b5ea931a6d4058f339717e273a722cee1ab65cf6cf8f1ba` and
`4dc59a9ea624533d8358138562bdcf3ac59dcfb19d2137404674964737aa6363` respectively.
The older shared handbook has hardware restrictions that predate STANDARD+ V2; the model
manual takes precedence. Extraction and PDF renderings remain research artifacts, not source files.

## Located meanings and summaries

- CV28 bits0/1 permit RailCom channels1/2. The summary requires CV29 bit3 and at least one
  channel. A disabled master permission establishes disabled RailCom; missing channels with the
  master enabled remain unknown. These are not RailComPlus registration or function-key data.
- CV29 documents 14/27 versus28/128 steps, DC permission, internal smoothing versus the selected
  CV67–94 table, direction and address selection. Long locomotive addresses are bounded to1–9999.
  CV19 is documented only as0–127 (Lenz command stations1–99); no bit7 consist direction is
  inferred. This profile does not establish CV21/22 function permissions.
- CV30 has stored lamp-short, overtemperature and motor/track-short flags. Zero does not prove a
  successful live hardware test. Reading these flags does not clear them.
- CV33–47 and129–144 identify outputs A–D for F0–F28. F0 and F1 use separate forward/reverse
  CVs; F2–F28 do not acquire an invented direction condition. Each mapping CV has its own located
  default, including CV36=8 and CV39=16. Tabled output bit4 remains explicitly unconfirmed.
  CV57/58/59/61/64 assign F1–F8 to dimming, shunting, disabling delays and light effects.
  They are assignments, not evidence of latching/momentary mode or key inversion. The technical
  Control disclosure and CV output view show these mappings without modifying vehicle functions.
- CV49 selects F0–F28 for the motor output in function-decoder mode;0 means F0. CV50 bit4 enables
  that mode and CV120 dims it. The UI never changes the vehicle type or a device mode automatically.
  CV50 bits0–3 select control set0–5; sets4/5 use CV113/114. Other bits configure the back-EMF
  divider, disable load regulation, or select approximately19Hz versus23kHz drive. CV9 is separate.
- CV51 bit1 enables ABC. Constant braking with ABC requires bits0 and1; bit7 independently
  configures constant braking on transition to speed step0. CV52 alone never enables braking.
  Bit2 removes ABC direction dependence. Bits3/4 select alternative shuttle modes; both set
  leaves the mode unknown. Bit5 permits polarity-independent DC stops only when CV29 bit2
  disables DC operation. Bit6 is unused. These are configured modes, not a live braking state:
  functions assigned by CV58/59 can suppress constant braking during operation.
- CV52 remains a calibrated distance parameter, never metres. CV53 is an ABC slow-speed
  parameter, never km/h. CV54 remains a raw shuttle-dwell parameter because of conflicting
  time statements. No exact elapsed time or duration is inferred from it.
- CV60 and62 use decimal units/tens for output effects, not independent bits. Invalid digit
  combinations remain unknown. The shared handbook pp.50–51 disambiguates CV63 frequency:
  `f=1/[0.03*(1+CV63)]`Hz. The value is a configured frequency, not an output measurement.
- CV112 and147–149 use16ms increments. Summaries convert only stored coupler timing values.
  CV145/146 describe direction-specific coupler-output selection; no fitted coupler is inferred.

## Smoothed characteristic

`rules.lenzStandardV2` enables one bounded client evaluator, accepted by Go validation only for
manufacturer99 and exact family/model STANDARD+ V2, with the documented ABC rule. Competing
ZIMO, motor-characteristic, ESU mapping and scalar brake rules are rejected. Go, TypeScript and
OpenAPI share this flag; executable formulas cannot be supplied through metadata.

The speed view retains raw CV2/6/5 support parameters, including zero, without borrowing ZIMO's
special values. With CV29 bit4 cleared, it explains the internal smoothing and draws no invented
curve. Only an explicitly selected CV67–94 table is plotted. Missing/invalid CV29 or a selected
table with no stored points never falls back to an inactive characteristic. Partial tables show only
their stored points; no factory values fill gaps. This changes only the reviewed Lenz profile.

## Conflicts and limits

- The model has four confirmed outputs A–D. The leaflet nevertheless lists CV118 for output E,
  references E=16 for couplers and supplies CV39=16 without an E mapping column. CV118 remains
  outside functional coverage; bit4 is warned, never labelled as a fitted output E.
- CV126/127 appear in the model table as SUSI pointer/transport registers. This is insufficient
  to prove a usable SUSI interface. They remain qualified documentation, not index selectors,
  hardware capabilities, or an authorization to query/write external modules.
- CV29 factory byte14 disagrees with its per-bit RailCom default. The complete byte remains a
  source reference; no per-bit factory claim or missing vehicle value is manufactured.
- CV54 documents0–255 and1–256seconds while the shared text also describes a4second default
  and0 as immediate departure. No exact conversion is implemented.
- The model CV149 example conflicts with its table default30. The Plus handbook p.65 confirms
  the16ms increment and30 as its example, but no document default populates a vehicle reading.
- No unsupported conversion for CV3/4, no measured speeds/distances, and no Lenz key-mode or
  inversion takeover are added. Existing ESU F0/F1/F3 evidence remains separate.

Existing CV values, profile groups, index identity, source lines, export times and vehicle/function
drafts are never modified by viewing. All77 records round-trip within the existing128KiB profile
limit. Schema71 and backup24 stay unchanged. Startup/restore upgrades only the exact previous
bundled metadata; edited/custom profiles and operator labels, source URL, ordering and activation
remain protected. Other profiles, inventory and all284 manufacturer-reference sets are unchanged.
