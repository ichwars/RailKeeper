# ESU LokPilot 5 micro DCC: interpretation quality

The reviewed profile contains 217 functional definition records and three reservation records
covering six base addresses. Every record has a source index and a location within that source.
Twenty-nine records carry a documented unit; 57 carry a documented factory reference. These
are documentation facts, never replacements for missing readings or evidence of fitted hardware.
The profile remains below the existing 128 KiB storage/import limit.

## Manufacturer evidence

The principal source is ESU's [LokPilot 5 family manual, eighth edition, January 2023,
from firmware 5.3.128](https://www.esu.eu/download/betriebsanleitungen/digitaldecoder/?tx_esudownloads_pi1%5BdownloadItem%5D=803bc0046a0244b6416c82c6abfd385e).
The reviewed download hash remains in the profile. Newer registers retain their separate
[LokProgrammer release-note evidence](https://www.esu.eu/download/software/lokprogrammer/)
and minimum firmware. CV23/24 refer explicitly to NMRA S-9.2.2 rather than claiming an ESU
table entry. CV245 retains the separate ESU coupler manual as its source.

- Base CV57/58 (p.84), CV141–143 (p.86) and CV250 (p.87) are explicitly reserved.
  LokSound descriptions for these addresses must not become LokPilot functions.
- CV124 bit0 concerns direction retention across operating-mode changes (pp.49/85).
  It does not invert a function key. Hardware-specific SUSI, H0 and Fx options remain qualified.
- Control CVs N/O (pp.62–65) describe assignments to a mapping row. Their bits do not report
  current states. A set shunting-mode bit therefore does not establish that shunting is active.
- CV102/111 use 16 ms steps, CV112 uses 0.065536 s steps, CV113 uses 0.032768 s steps,
  CV116–119 use 0.1 ms steps, CV132 uses 0.065 s steps, CV247/248 use 0.016 s steps,
  and CV252 uses 65 ms steps. Documentation keeps these different scales separate.
- Trims retain the zero-value exception. ABC shuttle waiting time retains its zero-disable
  condition. Internal speed steps are not kilometres per hour; CV254/255 are not centimetres.
- No unverified micro-DCC time conversion is introduced for CV3/4. CV11/114/115/131/199/245
  still do not acquire an invented scale. Output parameters remain dependent on output mode.

## Reference-export accounting

The 1,367-value, 15-index-page reference export resolves as follows. This measures the existence
of bounded definitions, not complete bit semantics for every byte. Reservations do not increase
`knownValues`; RailComPlus continues to have a separate, qualified interpretation.

| Scope | Functional definition | Explicitly reserved | No general definition |
| --- | ---: | ---: | ---: |
| Base | 112 | 6 | 14 |
| 1/0 | 0 | 0 | 33 |
| 1/1 | 0 | 0 | 64 |
| 1/2 | 0 | 0 | 20 |
| 16/0 | 69 | 0 | 1 |
| 16/3–6, each | 143 | 0 | 1 |
| 16/7 | 71 | 0 | 1 |
| 16/8 | 80 | 0 | 0 |
| 16/9–11, each | 79 | 0 | 1 |
| 16/12 | 39 | 0 | 1 |
| 16/13 | 40 | 0 | 0 |
| **Total** | **1,220** | **6** | **141** |

The 141 remaining addresses comprise:

- Base: CV100,109,110,191–195,200–205. No model/firmware-safe manufacturer meaning is established.
- 1/0: CV257–288 and 321. The existing RailComPlus reader recognises the terminated name within
  CV261–288; headers, padding and CV321 are not assigned an invented meaning.
- 1/1: CV257–320. Existing function descriptors are decoded separately, with the F0/F1/F3
  checkbox evidence and manual-setting protection unchanged. They are not new manufacturer-table rules.
- 1/2: CV257–276. No reliable interpretation is established.
- 16/0: CV272. The manufacturer's output table names CV273 at the disputed position; an
  expected sequential address alone does not establish CV272.
- 16/3–7: CV263, and 16/9–12: CV263. The mapping table contains conflicting first-row G/Q
  addresses. The prior conservative exclusions remain in place.

## Storage, import and presentation

`DecoderDefinition.reserved` is optional. A reservation requires a source and cannot specify bits,
unit, access, finite functional values or a factory value. Backend and frontend exclude reservations
from the functional interpretation count. Raw bytes, exact index pairs, original lines, export
timestamps and manually saved function flags are independent and remain unchanged.

The CV detail drawer shows the selected definition's exact source/location and applicable minimum
firmware. Units and factory references are separate from the read value. Reserved registers use the
neutral unresolved notice, and the library does not present their raw-byte bounds as functional
settings. Startup/restore upgrades only the exact recognised previous bundle; local edits and
operator activation/ordering are protected by the existing seed policy.

The catalogue-wide source/index/schema validation covers all 284 reference sets. This focused
semantic review applies to the ESU micro-DCC profile. It does not establish full bit/firmware
coverage for the other 598 catalogue models or close the 67 documented model gaps.
