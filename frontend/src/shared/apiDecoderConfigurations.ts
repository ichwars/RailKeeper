import type { components } from "../generated/api";

export type DecoderInstance = components["schemas"]["DecoderInstance"];
export type DecoderInstallation = components["schemas"]["DecoderInstallation"];
export type DecoderCVWorkspace = components["schemas"]["DecoderCVWorkspace"];
export type DecoderConfigurationContext = components["schemas"]["DecoderConfigurationContext"];
export type DecoderCVSnapshotSummary = components["schemas"]["DecoderCVSnapshotSummary"];
export type DecoderCVSnapshot = components["schemas"]["DecoderCVSnapshot"];
export type DecoderRecordedCVValue = components["schemas"]["DecoderRecordedCVValue"];

type APIRequest = <T>(path: string, init?: RequestInit) => Promise<T>;

export function createDecoderConfigurationsAPI(request: APIRequest) {
  return {
    decoderConfigurationContext: (vehicleId: string) =>
      request<DecoderConfigurationContext>(`/vehicles/${encodeURIComponent(vehicleId)}/decoder-configurations`),
    decoderCVSnapshots: (vehicleId?: string) =>
      request<DecoderCVSnapshotSummary[]>(`/decoder-configurations/snapshots${vehicleId === undefined ? "" : `?vehicleId=${encodeURIComponent(vehicleId)}`}`),
    decoderCVSnapshot: (id: string) => request<DecoderCVSnapshot>(`/decoder-configurations/snapshots/${encodeURIComponent(id)}`),
    decoderSnapshotSourceUrl: (snapshotId: string, sourceId: string) =>
      `/api/v1/decoder-configurations/snapshots/${encodeURIComponent(snapshotId)}/sources/${encodeURIComponent(sourceId)}/download`
  };
}
