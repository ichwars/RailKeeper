import type { ValuationDraft } from "./valuationModel";

// A tab-local journal retains the request identity when the dialog or document
// closes before its response is known. It never submits or retries a request.
function key(owner: string, vehicleID: string) {
  return `railkeeper.pendingValuation:${JSON.stringify([owner, vehicleID])}`;
}

function decode(raw: string | null): ValuationDraft | null {
  if (!raw || raw.length > 32768) return null;
  try {
    const value: unknown = JSON.parse(raw);
    if (!value || typeof value !== "object" || Array.isArray(value)) return null;
    const item = value as Record<string, unknown>;
    const fields = ["id", "amount", "currency", "valuedOn", "source", "notes", "replacesId"] as const;
    if (Object.keys(item).length !== fields.length + 1 || typeof item.withdrawn !== "boolean" ||
      fields.some(field => typeof item[field] !== "string") || !/^[a-f0-9]{32}$/.test(String(item.id))) return null;
    return {
      id: String(item.id), amount: String(item.amount), currency: String(item.currency),
      valuedOn: String(item.valuedOn), source: String(item.source), notes: String(item.notes),
      replacesId: String(item.replacesId), withdrawn: item.withdrawn
    };
  } catch {
    return null;
  }
}

export function pendingValuation(owner: string, vehicleID: string): ValuationDraft | null {
  try { return decode(sessionStorage.getItem(key(owner, vehicleID))); }
  catch { return null; }
}

export function rememberPendingValuation(owner: string, vehicleID: string, draft: ValuationDraft): boolean {
  try {
    const current = pendingValuation(owner, vehicleID);
    if (current && current.id !== draft.id) return false;
    const raw = JSON.stringify(draft);
    if (!decode(raw)) return false;
    sessionStorage.setItem(key(owner, vehicleID), raw);
    return sessionStorage.getItem(key(owner, vehicleID)) === raw;
  } catch {
    return false;
  }
}

export function clearPendingValuation(owner: string, vehicleID: string, id: string) {
  try {
    if (pendingValuation(owner, vehicleID)?.id === id) sessionStorage.removeItem(key(owner, vehicleID));
  } catch {
    // Keeping an acknowledged operation merely requires another read comparison.
  }
}
