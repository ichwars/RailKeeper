import type { VehicleCVValue, VehicleCVValueInput } from "../../shared/api";
import { buildCVImportPreview, type CVImportPreview } from "./cvImport";

export class CVImportGuardError extends Error {
  constructor(readonly code: "noSelection" | "missingRevision") { super(code); }
}

export function guardedCVImportValues(preview: CVImportPreview): VehicleCVValueInput[] {
  const rows = preview.rows.filter(row => row.selected && row.status !== "invalid");
  if (!rows.length) throw new CVImportGuardError("noSelection");
  return rows.map(row => {
    if (row.existing && !/^[0-9a-f]{64}$/.test(row.existing.revision || "")) {
      throw new CVImportGuardError("missingRevision");
    }
    // A source file cannot authorize a write. Pin the revision observed in this comparison.
    return { ...row.input, expectedRevision: row.existing?.revision || "" };
  });
}

export function recompareCVImport(preview: CVImportPreview, current: VehicleCVValue[]): CVImportPreview {
  const selected = new Set(preview.rows.filter(row => row.selected).map(row => row.id));
  const refreshed = buildCVImportPreview(preview.fileName, preview.rows.map(row => row.input), current);
  return { ...refreshed, vehicleId: preview.vehicleId, rows: refreshed.rows.map(row => ({ ...row,
    selected: selected.has(row.id) && row.status !== "same" && row.status !== "invalid"
  })) };
}
