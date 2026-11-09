import type { DigitalCenterWorkItemFilter } from "./digitalCenterModel";
import { ApiError } from "../../shared/api";
import {
  type DigitalCenterWritePreview,
  type ECoSLiveStatus
} from "./digitalCenterModel";

export function normalizeLiveStatus(status: ECoSLiveStatus): ECoSLiveStatus {
  if (Array.isArray(status.pulseSamples) && Array.isArray(status.recentEvents)) return status;
  return {
    ...status,
    pulseSamples: status.pulseSamples ?? [],
    recentEvents: status.recentEvents ?? []
  };
}

export function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

export function digitalCenterErrorTranslationKey(code: string) {
  return code === "digital_center_conflict_unresolved"
    ? "digitalCenters.error.conflictUnresolved"
    : null;
}

export function digitalCenterAddressConflictMessage(
  error: ApiError,
  text: (key: string, values?: Record<string, string | number>) => string
) {
  const details = isRecord(error.details) ? error.details : null;
  const name = typeof details?.name === "string" ? details.name.trim() : "";
  const address = typeof details?.decoderAddress === "number" ? details.decoderAddress : null;
  return name && address !== null
    ? text("digitalCenters.error.addressConflict", { name, address })
    : text("digitalCenters.error.addressConflictGeneric");
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

export function validWriteGrant(preview: DigitalCenterWritePreview) {
  const expiresAt = Date.parse(preview.expiresAt);
  return preview.token.trim().length > 0 && Number.isFinite(expiresAt) && expiresAt > Date.now();
}

export type DigitalCenterWorkspaceOptions = {
  pollIntervalMs?: number;
};

export const unfilteredSessionTotal: DigitalCenterWorkItemFilter = {
  query: "",
  compareStatus: "all",
  page: 1,
  pageSize: 1
};
