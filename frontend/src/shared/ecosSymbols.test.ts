import { expect, it } from "vitest";
import corrections from "../../../backend/internal/application/function_symbol_ecos_codes.json";
import type { MasterDataEntry } from "./api";
import { exactECoSSymbol } from "./ecosSymbols";

it("uses documented ECoS wire codes independently of stable artwork keys", () => {
  const symbols: MasterDataEntry[] = corrections.symbols.map(s => ({ id: s.key, type: "symbols", key: s.key,
    label: s.key, active: true, sortOrder: 0, createdAt: "", updatedAt: "",
    metadata: s.ecosCode == null ? {} : { ecosCode: s.ecosCode } }));
  expect(exactECoSSymbol(10, symbols)?.key).toBe("esu-f009-rangiergang");
  expect(exactECoSSymbol(11, symbols)?.key).toBe("esu-f010-abv");
  expect(exactECoSSymbol(32, symbols)?.key).toBe("coupling");
  expect(exactECoSSymbol(260, symbols)?.key).toBe("esu-f004-innenraumbeleuchtung");
  expect(exactECoSSymbol(1607, symbols)?.key).toBe("esu-f032-diesel-fahrstufe-auf");
  expect(exactECoSSymbol(6, symbols)).toBeUndefined();
  expect(exactECoSSymbol(7431, symbols)).toBeUndefined();
  const manual = { ...symbols[0], id: "club", key: "club", metadata: { ecosCode: 7431 } };
  // Unreviewed wide descriptors stay opaque even with a guessed compatibility entry.
  expect(exactECoSSymbol(7431, [...symbols, manual])).toBeUndefined();
  expect(exactECoSSymbol(10, [...symbols, { ...manual, metadata: { ecosCode: 10 } }])).toBeUndefined();
});
