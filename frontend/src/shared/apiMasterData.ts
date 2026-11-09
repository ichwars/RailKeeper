import type {
  MasterDataEntry, MasterDataImportResult, MasterDataInput, MasterDataRelation
} from "./apiMasterDataModels";
import type { APIRequest } from "./apiTransport";
import type { DecoderLibrary } from "./decoderLibraryModel";

export function createMasterDataAPI(request: APIRequest) {
  return {
    masterData: (type: string, activeOnly = false) =>
      request<MasterDataEntry[]>(
        `/master-data/${encodeURIComponent(type)}${activeOnly ? "?active=true" : ""}`
      ),
    decoderLibrary: (includeInactive = false) => request<DecoderLibrary>(`/decoder-library${includeInactive ? "?includeInactive=true" : ""}`),
    decoderReferences: (modelId: string) => request<import("./decoderLibraryModel").DecoderCVReference[]>(`/decoder-library/references/${encodeURIComponent(modelId)}`),
    masterDataAll: (activeOnly = false) =>
      request<Record<string, MasterDataEntry[]>>(
        `/master-data-all${activeOnly ? "?active=true" : ""}`,
        {},
        { retries: 1, timeoutMs: 30000 }
      ),
    managedMasterData: (type: string) =>
      request<MasterDataEntry[]>(
        `/master-data/${encodeURIComponent(type)}?management=true`
      ),
    managedMasterDataAll: () =>
      request<Record<string, MasterDataEntry[]>>(
        "/master-data-all?management=true",
        {},
        { retries: 1, timeoutMs: 30000 }
      ),
    createMasterData: (type: string, input: MasterDataInput) =>
      request<MasterDataEntry>(`/master-data/${encodeURIComponent(type)}`, {
        method: "POST",
        body: JSON.stringify(input)
      }),
    updateMasterData: (type: string, key: string, input: MasterDataInput) =>
      request<MasterDataEntry>(`/master-data/${encodeURIComponent(type)}/${encodeURIComponent(key)}`, {
        method: "PUT",
        body: JSON.stringify(input)
      }),
    setMasterDataActive: (type: string, key: string, active: boolean) =>
      request<MasterDataEntry>(
        `/master-data/${encodeURIComponent(type)}/${encodeURIComponent(key)}/active`,
        {
          method: "PATCH",
          body: JSON.stringify({ active })
        }
      ),
    setMasterDataActiveMany: (type: string, keys: string[], active: false) =>
      request<MasterDataEntry[]>(`/master-data/${encodeURIComponent(type)}/active`, {
        method: "PATCH",
        body: JSON.stringify({ keys, active })
      }),
    deleteMasterData: (type: string, key: string) =>
      request<void>(`/master-data/${encodeURIComponent(type)}/${encodeURIComponent(key)}`, {
        method: "DELETE"
      }),
    masterDataRelations: (parentType: string, childType: string) =>
      request<MasterDataRelation[]>(
        `/master-data-relations?parentType=${encodeURIComponent(parentType)}&childType=${encodeURIComponent(childType)}`
      ),
    masterDataExportUrl: () => "/api/v1/master-data/export",
    importMasterData: (file: File) => {
      const form = new FormData();
      form.append("file", file);
      return request<MasterDataImportResult>(
        "/master-data/import",
        {
          method: "POST",
          body: form
        },
        { timeoutMs: 120000 }
      );
    }
  };
}
