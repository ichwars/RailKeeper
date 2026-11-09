import type {
  DigitalCenterMessagesResponse, DigitalCenterProvider, DigitalCenterReadSession, DigitalCenterWorkItem,
  DigitalCenterWorkItemFilter, DigitalCenterWorkItemPage, DigitalCenterWorkspaceSummary,
  DigitalCenterWriteConfirmInput, DigitalCenterWriteConfirmation, DigitalCenterWritePreview,
  DigitalCenterWritePreviewInput, ECoSBackupImportResult, ECoSLiveStatus
} from "../features/digitalCenters/digitalCenterModel";
import type {
  DigitalCenterConnectionInput, DigitalCenterConnectionResult, DigitalCenterProbeResult, ECoSConnectionInput,
  ECoSConnectionResult, ECoSLocomotiveSummary, ECoSLocomotiveSyncInput, ECoSLocomotiveSyncResult,
  ECoSRawProbe
} from "./apiDigitalCenterModels";
import type { ECoSFunctionSnapshot } from "./apiInventoryModels";
import type { APIRequest } from "./apiTransport";
import type { ECoSImageOptions, ECoSImagePreview, ECoSImageResult } from "./ecosImageModel";

function digitalCenterSessionSuffix(sessionID?: string) {
  const normalized = sessionID?.trim();
  return normalized ? `?sessionId=${encodeURIComponent(normalized)}` : "";
}

export function createDigitalCentersAPI(request: APIRequest) {
  return {
    ecosImageOptions: (vehicleId: string) => request<ECoSImageOptions>(
      `/vehicles/${encodeURIComponent(vehicleId)}/ecos-image/options`, {}, { timeoutMs: 35000 }),
    previewECoSImage: (vehicleId: string, objectId: number, description: string, image: Blob, existingSlot?: number) => {
      const body = new FormData();
      body.append("objectId", String(objectId)); body.append("description", description);
      body.append("image", image, "railkeeper.bmp");
      if (existingSlot !== undefined) body.append("existingSlot", String(existingSlot));
      return request<ECoSImagePreview>(`/vehicles/${encodeURIComponent(vehicleId)}/ecos-image/preview`,
        { method: "POST", body }, { timeoutMs: 80000 });
    },
    confirmECoSImage: (vehicleId: string, token: string) => request<ECoSImageResult>(
      `/vehicles/${encodeURIComponent(vehicleId)}/ecos-image/confirm`,
      { method: "POST", body: JSON.stringify({ token, confirm: true }) }, { timeoutMs: 105000 }),
    readVehicleECoSFunctions: (vehicleId: string, signal?: AbortSignal) => request<ECoSFunctionSnapshot>(
      `/vehicles/${encodeURIComponent(vehicleId)}/ecos-functions`, { signal }, { timeoutMs: 45000 }),
    testECoSConnection: (input: ECoSConnectionInput) =>
      request<ECoSConnectionResult>(
        "/ecos/test",
        {
          method: "POST",
          body: JSON.stringify(input)
        },
        { timeoutMs: 10000 }
      ),
    probeECoSLocomotiveRaw: (input: ECoSConnectionInput) =>
      request<ECoSRawProbe>(
        "/ecos/locomotives/raw",
        {
          method: "POST",
          body: JSON.stringify(input)
        },
        { timeoutMs: 120000 }
      ),
    countECoSLocomotives: (input: ECoSConnectionInput) =>
      request<ECoSLocomotiveSummary>(
        "/ecos/locomotives/count",
        {
          method: "POST",
          body: JSON.stringify(input)
        },
        { timeoutMs: 10000 }
      ),
    getECoSLiveStatus: () =>
      request<ECoSLiveStatus>(`/digital-centers/${encodeURIComponent("ecos")}/live/status`),
    digitalCenterWorkspace: () =>
      request<DigitalCenterWorkspaceSummary>("/digital-centers/workspace"),
    importECoSBackup: (file: File, options: { confirmStation: boolean; stationHost: string }) =>
      request<ECoSBackupImportResult>("/digital-centers/ecos/backup-import", {
        method: "POST",
        headers: {
          "Content-Type": "application/octet-stream",
          "X-RailKeeper-Confirm-Station": String(options.confirmStation),
          ...(options.confirmStation ? { "X-RailKeeper-Backup-Station": options.stationHost } : {})
        },
        body: file
      }, { timeoutMs: 120000 }),
    startDigitalCenterReadSession: (provider: DigitalCenterProvider) =>
      request<DigitalCenterReadSession>(
        `/digital-centers/${encodeURIComponent(provider)}/read-sessions`,
        { method: "POST", body: JSON.stringify({}) },
        { timeoutMs: 120000 }
      ),
    digitalCenterReadSession: (sessionID: string) =>
      request<DigitalCenterReadSession>(
        `/digital-centers/read-sessions/${encodeURIComponent(sessionID)}`
      ),
    digitalCenterWorkItems: (sessionID: string, filter: DigitalCenterWorkItemFilter) => {
      const query = new URLSearchParams();
      const search = filter.query.trim();
      if (search) query.set("q", search);
      if (filter.compareStatus !== "all") query.set("compareStatus", filter.compareStatus);
      query.set("page", String(filter.page));
      query.set("pageSize", String(filter.pageSize));
      return request<DigitalCenterWorkItemPage>(
        `/digital-centers/read-sessions/${encodeURIComponent(sessionID)}/items?${query.toString()}`
      );
    },
    digitalCenterWorkItem: (sessionID: string, itemID: string) =>
      request<DigitalCenterWorkItem>(
        `/digital-centers/read-sessions/${encodeURIComponent(sessionID)}/items/${encodeURIComponent(itemID)}`
      ),
    readDigitalCenterCVBatch: (sessionID: string, itemID: string, startCV: number,
      options: { restart?: boolean; source?: "database" } = {}) =>
      request<DigitalCenterWorkItem>(
        `/digital-centers/read-sessions/${encodeURIComponent(sessionID)}/items/${encodeURIComponent(itemID)}/cv-read`,
        { method: "POST", body: JSON.stringify({ startCV, restart: Boolean(options.restart), source: options.source }) },
        { timeoutMs: options.source === "database" ? 100000 : 30000 }
      ),
    importDigitalCenterCVs: (sessionID: string, itemID: string, vehicleId: string) =>
      request<{ imported: number; preserved: number }>(
        `/digital-centers/read-sessions/${encodeURIComponent(sessionID)}/items/${encodeURIComponent(itemID)}/cv-import`,
        { method: "POST", body: JSON.stringify({ vehicleId }) }
      ),
    digitalCenterSessionMessages: (sessionID: string) =>
      request<DigitalCenterMessagesResponse>(
        `/digital-centers/read-sessions/${encodeURIComponent(sessionID)}/messages`
      ),
    digitalCenterLiveStatus: (provider: DigitalCenterProvider) =>
      request<ECoSLiveStatus>(`/digital-centers/${encodeURIComponent(provider)}/live/status`),
    startDigitalCenterLive: (provider: DigitalCenterProvider, sessionID?: string) =>
      request<ECoSLiveStatus>(
        `/digital-centers/${encodeURIComponent(provider)}/live/start` + digitalCenterSessionSuffix(sessionID),
        { method: "POST", body: JSON.stringify({}) }
      ),
    stopDigitalCenterLive: (provider: DigitalCenterProvider, sessionID?: string) =>
      request<ECoSLiveStatus>(
        `/digital-centers/${encodeURIComponent(provider)}/live/stop` + digitalCenterSessionSuffix(sessionID),
        { method: "POST", body: JSON.stringify({}) }
      ),
    previewDigitalCenterWrite: (
      sessionID: string,
      itemID: string,
      input: DigitalCenterWritePreviewInput
    ) => request<DigitalCenterWritePreview>(
      `/digital-centers/read-sessions/${encodeURIComponent(sessionID)}/items/${encodeURIComponent(itemID)}/write-preview`,
      { method: "POST", body: JSON.stringify(input) },
      { timeoutMs: 30000 }
    ),
    confirmDigitalCenterWrite: (
      sessionID: string,
      itemID: string,
      input: DigitalCenterWriteConfirmInput
    ) => request<DigitalCenterWriteConfirmation>(
      `/digital-centers/read-sessions/${encodeURIComponent(sessionID)}/items/${encodeURIComponent(itemID)}/write-confirm`,
      { method: "POST", body: JSON.stringify(input) },
      { timeoutMs: 30000 }
    ),
    syncECoSLocomotive: (input: ECoSLocomotiveSyncInput) =>
      request<ECoSLocomotiveSyncResult>(
        "/digital-centers/ecos/locomotives/sync",
        {
          method: "POST",
          body: JSON.stringify(input)
        },
        { timeoutMs: 30000 }
      ),
    startECoSLive: (input: ECoSConnectionInput) =>
      request<ECoSLiveStatus>(
        `/digital-centers/${encodeURIComponent("ecos")}/live/start`,
        {
          method: "POST",
          body: JSON.stringify(input)
        },
        { timeoutMs: 10000 }
      ),
    stopECoSLive: () =>
      request<ECoSLiveStatus>(`/digital-centers/${encodeURIComponent("ecos")}/live/stop`, {
        method: "POST"
      }),
    testZ21Connection: (input: DigitalCenterConnectionInput) =>
      request<DigitalCenterConnectionResult>(
        "/digital-centers/z21/test",
        {
          method: "POST",
          body: JSON.stringify(input)
        },
        { timeoutMs: 10000 }
      ),
    probeZ21Connection: (input: DigitalCenterConnectionInput) =>
      request<DigitalCenterProbeResult>(
        "/digital-centers/z21/probe",
        {
          method: "POST",
          body: JSON.stringify(input)
        },
        { timeoutMs: 15000 }
      ),
    testIntellibox3Connection: (input: DigitalCenterConnectionInput) =>
      request<DigitalCenterConnectionResult>(
        "/digital-centers/intellibox3/test",
        {
          method: "POST",
          body: JSON.stringify(input)
        },
        { timeoutMs: 10000 }
      ),
    probeIntellibox3Connection: (input: DigitalCenterConnectionInput) =>
      request<DigitalCenterProbeResult>(
        "/digital-centers/intellibox3/probe",
        {
          method: "POST",
          body: JSON.stringify(input)
        },
        { timeoutMs: 15000 }
      ),
    testCS3Connection: (input: DigitalCenterConnectionInput) =>
      request<DigitalCenterConnectionResult>(
        "/digital-centers/cs3/test",
        {
          method: "POST",
          body: JSON.stringify(input)
        },
        { timeoutMs: 10000 }
      ),
    probeCS3Connection: (input: DigitalCenterConnectionInput) =>
      request<DigitalCenterProbeResult>(
        "/digital-centers/cs3/probe",
        {
          method: "POST",
          body: JSON.stringify(input)
        },
        { timeoutMs: 15000 }
      )
  };
}
