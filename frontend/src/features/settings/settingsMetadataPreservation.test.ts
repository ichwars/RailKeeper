import { describe, expect, it } from "vitest";
import type { MasterDataEntry } from "../../shared/api";
import {
  applyVisibleMetadata, cv8BinaryText, cv8CountryText, cv8DecimalText, cv8DecimalValue, cv8HexText,
  cv8NameText, domainFromUrl, emptyForm, entryToForm, externalLink, manufacturerAliasesText,
  manufacturerNeedsWebsiteReview, manufacturerSearchDomains, metadataList, metadataNumberString,
  metadataString, normalizeCV8Binary, normalizeCV8Decimal, normalizeCV8Hex
} from "./settingsModel";

function record(metadata: Record<string, unknown> = {}, extra: Partial<MasterDataEntry> = {}): MasterDataEntry {
  return { id: "cv8-original", type: "cv8_manufacturer", key: "unknown", label: "151 - ESU", active: true,
    sortOrder: 5, metadata, createdAt: "2026-11-08T12:00:00Z", updatedAt: "2026-11-08T12:00:00Z", ...extra };
}

describe("master-data metadata and CV8 preservation", () => {
  it("round-trips every manufacturer byte including zero through decimal, binary and hex forms", () => {
    for (let value = 0; value <= 255; value++) {
      const decimal = String(value), binary = value.toString(2).padStart(8, "0"), hex = `0x${value.toString(16).toUpperCase().padStart(2, "0")}`;
      expect(normalizeCV8Decimal(` ${decimal} `)).toBe(decimal);
      expect(normalizeCV8Binary("", decimal)).toBe(binary);
      expect(normalizeCV8Hex("", decimal)).toBe(hex);
      const original = Object.freeze({ opaque: { keep: "old manufacturer metadata" }, decimal: 151 });
      const saved = applyVisibleMetadata("cv8_manufacturer", original, { ...emptyForm, cvDecimal: decimal, cvCountry: " de " });
      expect(saved).toMatchObject({ decimal: value, cvDecimal: decimal, cv8: value, binary, cvBinary: binary, hex, cvHex: hex, country: "DE", cvCountry: "DE", opaque: original.opaque });
      expect(original.decimal).toBe(151);
      const form = entryToForm(record(saved));
      expect(form.cvDecimal).toBe(decimal); expect(form.cvBinary).toBe(binary); expect(form.cvHex).toBe(hex);
      expect(form.label).toBe("ESU");
    }
  });

  it("retains legacy CV8 metadata precedence, fallback identities and formatted names", () => {
    expect(cv8DecimalValue(record({ decimal: 0, cvDecimal: "151" }))).toBe(0);
    expect(cv8DecimalValue(record({ decimal: 256, cvDecimal: " 151 " }))).toBe(151);
    expect(cv8DecimalValue(record({ decimal: -1, cvDecimal: {}, cv8: 145 }))).toBe(145);
    expect(cv8DecimalValue(record({ manufacturerId: "18" }))).toBe(18);
    expect(cv8DecimalValue(record({}, { key: "legacy-0017", label: "Stored label" }))).toBe(17);
    expect(cv8DecimalValue(record({}, { key: "invalid-256", label: " 151 - ESU " }))).toBe(151);
    expect(cv8DecimalValue(record({}, { key: "unknown", label: "Unknown maker" }))).toBeUndefined();
    expect(cv8NameText(record({}, { label: "17 - " }))).toBe("17 -");
    expect(cv8DecimalText(record({}, { label: "Unknown maker" }))).toBe("");
    expect(cv8BinaryText(record({ binary: "0b1001 0111", decimal: 1 }))).toBe("10010111");
    expect(cv8HexText(record({ cvHex: "97", decimal: 1 }))).toBe("0x97");
    expect(cv8CountryText(record({ cvCountry: " DE " }))).toBe("DE");
    expect(cv8BinaryText(record({}, { label: "Unknown maker" }))).toBe("");
    expect(cv8HexText(record({}, { label: "Unknown maker" }))).toBe("");
  });

  it.each(["", "invalid", "-1", "256"])("does not invent a byte when the decimal input is %s", value => {
    expect(normalizeCV8Decimal(value)).toBe("");
    expect(normalizeCV8Hex("invalid", value)).toBe("");
    expect(normalizeCV8Binary("invalid", value)).toBe("");
  });

  it("respects explicit textual byte forms and removes only cleared visible CV8 fields", () => {
    expect(normalizeCV8Hex("0xa", "151")).toBe("0x0A");
    expect(normalizeCV8Binary("0B101", "151")).toBe("00000101");
    const original = { decimal: 151, cvDecimal: "151", cv8: 151, binary: "10010111", cvBinary: "10010111", hex: "0x97", cvHex: "0x97", country: "DE", cvCountry: "DE", unrelated: { keep: true } };
    const next = applyVisibleMetadata("cv8_manufacturer", original, { ...emptyForm });
    expect(next).toEqual({ unrelated: original.unrelated });
    expect(original.decimal).toBe(151);
  });

  it("preserves opaque manufacturer metadata while applying explicit aliases and domains", () => {
    const original = { opaque: { keep: true }, aliases: ["Legacy"], nominalScales: ["TT"], website: "https://old.test", searchDomains: ["old.test"] };
    const next = applyVisibleMetadata("manufacturer", original, { ...emptyForm, website: " https://www.new.test/catalog ", aliasesText: "Name A; Name B\nName C", nominalScalesText: "TT, H0" });
    expect(next).toEqual({ opaque: original.opaque, website: "https://www.new.test/catalog", searchDomains: ["new.test"], aliases: ["Name A", "Name B", "Name C"], nominalScales: ["TT", "H0"] });
    expect(applyVisibleMetadata("manufacturer", original, { ...emptyForm })).toEqual({ opaque: original.opaque });
    expect(applyVisibleMetadata("manufacturer", {}, { ...emptyForm, website: "https://new.test", searchDomainsText: "explicit.test" }).searchDomains).toEqual(["explicit.test"]);
    expect(original.aliases).toEqual(["Legacy"]);
  });

  it("keeps raw array aliases and distinguishes actual manufacturer sites from reviewable wiki sources", () => {
    expect(manufacturerAliasesText(record({ aliases: [" Legacy ", "", null, 17, "Name"] }))).toBe(" Legacy , Name");
    expect(metadataList(record({ aliases: "Name A; Name B\nName C" }), "aliases")).toEqual(["Name A", "Name B", "Name C"]);
    expect(metadataList(record({ aliases: 17 }), "aliases")).toEqual([]);
    expect(metadataNumberString(record({ value: 0 }), "value")).toBe("0");
    expect(metadataNumberString(record({ value: " 007 " }), "value")).toBe("007");
    expect(metadataNumberString(record({ value: Infinity }), "value")).toBe("");
    expect(metadataString(record({ value: 17 }), "value")).toBe("");
    expect(manufacturerSearchDomains(record({ website: "https://www.maker.test/path" }))).toEqual(["maker.test"]);
    expect(manufacturerSearchDomains(record({ searchDomains: ["specific.test"], website: "https://maker.test" }))).toEqual(["specific.test"]);
    expect(manufacturerSearchDomains(record({ website: "" }))).toEqual([]);
    expect(domainFromUrl("www.maker.test/path")).toBe("maker.test");
    expect(domainFromUrl("http://[invalid")).toBe("");
    expect(manufacturerNeedsWebsiteReview(record({}, { sourceUrl: "https://wikipedia.org/wiki/Model" }))).toBe(true);
    expect(manufacturerNeedsWebsiteReview(record({ website: "https://maker.test" }, { sourceUrl: "https://wikipedia.org/wiki/Model" }))).toBe(false);
    expect(externalLink(record({ website: " https://maker.test " }, { sourceUrl: "https://source.test" }))).toMatchObject({ href: "https://maker.test" });
    expect(externalLink(record({}, { sourceUrl: "https://source.test" }))).toMatchObject({ href: "https://source.test" });
    expect(externalLink(record())).toBeNull();
  });
});
