import type { VehicleValuation, VehicleValuationInput } from "../../shared/apiValuations";
import { newRandomIdentity } from "../../shared/randomIdentity";

export type ValuationDraft = Required<VehicleValuationInput>;
export const valuationChangedEvent = "railkeeper-valuations-changed";
export const valuationHistoryPageSize = 100;

export function valuationLocalDate(date = new Date()): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

export function canonicalValuationAmount(value: string): string | null {
  const match = /^([0-9]{1,12})(?:[.,]([0-9]{1,6}))?$/.exec(value.trim());
  if (!match) return null;
  const major = BigInt(match[1]).toString();
  const fraction = (match[2] || "").replace(/0+$/, "").padEnd(2, "0");
  return `${major}.${fraction}`;
}

export function newValuationDraft(record?: VehicleValuation, withdrawn = false): ValuationDraft {
  return {
    id: newRandomIdentity(),
    amount: record?.amount || "", currency: record?.currency || "EUR",
    valuedOn: record?.valuedOn || valuationLocalDate(),
    source: withdrawn ? "" : record?.source || "", notes: record?.notes || "",
    replacesId: record?.id || "", withdrawn
  };
}

export function sameValuationRequest(record: VehicleValuation, draft: ValuationDraft) {
  return record.id === draft.id && record.amount === canonicalValuationAmount(draft.amount) &&
    record.currency === draft.currency.trim().toUpperCase() && record.valuedOn === draft.valuedOn &&
    record.source === draft.source.trim() && record.notes === draft.notes.trim() &&
    record.replacesId === draft.replacesId && record.withdrawn === draft.withdrawn;
}
