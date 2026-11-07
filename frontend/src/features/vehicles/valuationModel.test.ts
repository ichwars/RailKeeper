import { describe, expect, it } from "vitest";
import { canonicalValuationAmount, newValuationDraft, sameValuationRequest, valuationLocalDate } from "./valuationModel";
import type { VehicleValuation } from "../../shared/apiValuations";

const record: VehicleValuation = {
  id: "1".repeat(32), vehicleId: "a", amount: "1234.123456", currency: "EUR",
  valuedOn: "2026-01-01", source: "Manual", notes: "Original", replacesId: "", withdrawn: false,
  recordedAt: "2026-11-07T12:00:00.000000000Z", replacedById: ""
};

describe("exact manual valuation requests", () => {
  it("defaults to the user's local calendar date including near midnight", () => {
    expect(valuationLocalDate(new Date(2026, 10, 7, 0, 15))).toBe("2026-11-07");
    expect(valuationLocalDate(new Date(2026, 0, 2, 23, 59))).toBe("2026-01-02");
  });
  it.each([["0", "0.00"], ["0001,2", "1.20"], ["1234.123456", "1234.123456"], [" 1,000000 ", "1.00"],
  ["999999999999.999999", "999999999999.999999"]])("normalizes %s without floating point", (input, expected) => {
    expect(canonicalValuationAmount(input)).toBe(expected);
  });
  it.each(["-1", "+1", "1e3", "1.1234567", "1,000.00", "١٢", "1000000000000", ".2", "2.", ""])("rejects %s", input => {
    expect(canonicalValuationAmount(input)).toBeNull();
  });
  it("keeps correction context and uses identities supported on plain HTTP LAN", () => {
    const draft = newValuationDraft(record);
    expect(draft.id).toMatch(/^[0-9a-f]{32}$/);
    expect(draft.id).not.toBe(record.id);
    expect(draft.replacesId).toBe(record.id);
    expect(draft.amount).toBe(record.amount);
    expect(draft.source).toBe(record.source);
    const withdrawn = newValuationDraft(record, true);
    expect(withdrawn.source).toBe("");
    expect(withdrawn.withdrawn).toBe(true);
    expect(record.replacesId).toBe("");
  });
  it("acknowledges only the same normalized request", () => {
    const draft = {      
...newValuationDraft(), id: record.id, amount: "001234,123456", currency: " eur ",
      valuedOn: record.valuedOn, source: " Manual ", notes: " Original "    
};
    expect(sameValuationRequest(record, draft)).toBe(true);
    for (const override of [{ id: "2".repeat(32) }, { amount: "0" }, { currency: "CHF" }, { source: "Other" }, { notes: "Changed" },
    { valuedOn: "2026-02-01" }, { replacesId: "other" }, { withdrawn: true }]) {
      expect(sameValuationRequest(record, { ...draft, ...override })).toBe(false);
    }
  });
});
