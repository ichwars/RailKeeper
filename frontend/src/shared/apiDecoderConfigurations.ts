import type { components } from "../generated/api";

export type DecoderInstance = components["schemas"]["DecoderInstance"];
export type DecoderInstallation = components["schemas"]["DecoderInstallation"];
export type DecoderCVWorkspace = components["schemas"]["DecoderCVWorkspace"];
export type DecoderConfigurationContext = components["schemas"]["DecoderConfigurationContext"];
export type DecoderCVSnapshotSummary = components["schemas"]["DecoderCVSnapshotSummary"];
export type DecoderCVSnapshot = components["schemas"]["DecoderCVSnapshot"];
export type DecoderRecordedCVValue = components["schemas"]["DecoderRecordedCVValue"];
export type DecoderInstanceInput = components["schemas"]["DecoderInstanceInput"];
export type DecoderInstallationInput = components["schemas"]["DecoderInstallationInput"];
export type DecoderRemovalInput = components["schemas"]["DecoderRemovalInput"];
export type DecoderInstallationDatesInput = components["schemas"]["DecoderInstallationDatesInput"];
export type DecoderCVWorkspaceInput = components["schemas"]["DecoderCVWorkspaceInput"];
export type DecoderCVCheckpointInput = components["schemas"]["DecoderCVCheckpointInput"];
export type DecoderCVAssignmentInput = components["schemas"]["DecoderCVAssignmentInput"];

type APIRequest = <T>(path: string, init?: RequestInit) => Promise<T>;

export function createDecoderConfigurationsAPI(request: APIRequest) {
  return {
    decoderConfigurationContext: (vehicleId: string) =>
      request<DecoderConfigurationContext>(`/vehicles/${encodeURIComponent(vehicleId)}/decoder-configurations`),
    decoderCVSnapshots: (vehicleId?: string) =>
      request<DecoderCVSnapshotSummary[]>(`/decoder-configurations/snapshots${vehicleId === undefined ? "" : `?vehicleId=${encodeURIComponent(vehicleId)}`}`),
    decoderCVSnapshot: (id: string) => request<DecoderCVSnapshot>(`/decoder-configurations/snapshots/${encodeURIComponent(id)}`),
    decoderSnapshotSourceUrl: (snapshotId: string, sourceId: string) =>
      `/api/v1/decoder-configurations/snapshots/${encodeURIComponent(snapshotId)}/sources/${encodeURIComponent(sourceId)}/download`,
    createDecoderInstance: (vehicleId: string, input: DecoderInstanceInput) =>
      request<DecoderConfigurationContext>(`/vehicles/${encodeURIComponent(vehicleId)}/decoder-configurations/instances`, {
        method: "POST", body: JSON.stringify(input)
      }),
    updateDecoderInstance: (vehicleId: string, input: DecoderInstanceInput) =>
      request<DecoderConfigurationContext>(`/vehicles/${encodeURIComponent(vehicleId)}/decoder-configurations/instances/${encodeURIComponent(input.id)}`, {
        method: "PUT", body: JSON.stringify(input)
      }),
    installDecoder: (vehicleId: string, input: DecoderInstallationInput) =>
      request<DecoderConfigurationContext>(`/vehicles/${encodeURIComponent(vehicleId)}/decoder-configurations/installations`, {
        method: "POST", body: JSON.stringify(input)
      }),
    removeDecoder: (vehicleId: string, installationId: string, input: DecoderRemovalInput) =>
      request<DecoderConfigurationContext>(`/vehicles/${encodeURIComponent(vehicleId)}/decoder-configurations/installations/${encodeURIComponent(installationId)}/removal`, {
        method: "POST", body: JSON.stringify(input)
      }),
    correctDecoderInstallationDates: (vehicleId: string, installationId: string, input: DecoderInstallationDatesInput) =>
      request<DecoderConfigurationContext>(`/vehicles/${encodeURIComponent(vehicleId)}/decoder-configurations/installations/${encodeURIComponent(installationId)}/dates`, {
        method: "PUT", body: JSON.stringify(input)
      }),
    createDecoderCVWorkspace: (vehicleId: string, input: DecoderCVWorkspaceInput) =>
      request<DecoderConfigurationContext>(`/vehicles/${encodeURIComponent(vehicleId)}/decoder-configurations/workspaces`, {
        method: "POST", body: JSON.stringify(input)
      }),
    assignDecoderCVWorkspace: (vehicleId: string, workspaceId: string, input: DecoderCVAssignmentInput) =>
      request<DecoderConfigurationContext>(`/vehicles/${encodeURIComponent(vehicleId)}/decoder-configurations/workspaces/${encodeURIComponent(workspaceId)}/assignment`, {
        method: "PUT", body: JSON.stringify(input)
      }),
    captureDecoderCVCheckpoint: (vehicleId: string, workspaceId: string, input: DecoderCVCheckpointInput) =>
      request<DecoderConfigurationContext>(`/vehicles/${encodeURIComponent(vehicleId)}/decoder-configurations/workspaces/${encodeURIComponent(workspaceId)}/checkpoints`, {
        method: "POST", body: JSON.stringify(input)
      })
  };
}
