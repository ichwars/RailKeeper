import type {
  DataTransferCSVMappingInput, DataTransferExportResult, DataTransferIssueResolution, DataTransferJob,
  DataTransferJobDetails, DataTransferJobFilter, DataTransferJobInput, DataTransferPreview,
  DataTransferProfile, DataTransferProfileInput, DataTransferSummary
} from "../features/importExport/dataTransferModel";
import type { APIRequest } from "./apiTransport";

export function createDataTransferAPI(request: APIRequest) {
  return {
    dataTransferSummary: () => request<DataTransferSummary>("/data-transfer/summary"),
    dataTransferProfiles: () => request<DataTransferProfile[]>("/data-transfer/profiles"),
    createDataTransferProfile: (input: DataTransferProfileInput) =>
      request<DataTransferProfile>("/data-transfer/profiles", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    updateDataTransferProfile: (id: string, input: DataTransferProfileInput) =>
      request<DataTransferProfile>(`/data-transfer/profiles/${encodeURIComponent(id)}`, {
        method: "PUT",
        body: JSON.stringify(input)
      }),
    disableDataTransferProfile: (id: string) =>
      request<void>(`/data-transfer/profiles/${encodeURIComponent(id)}`, { method: "DELETE" }),
    dataTransferJobs: (filter: DataTransferJobFilter = {}) => {
      const query = new URLSearchParams();
      if (filter.profileId) query.set("profileId", filter.profileId);
      if (filter.direction) query.set("direction", filter.direction);
      for (const state of filter.states || []) query.append("states", state);
      if (filter.limit !== undefined) query.set("limit", String(filter.limit));
      const suffix = query.size > 0 ? `?${query.toString()}` : "";
      return request<DataTransferJob[]>(`/data-transfer/jobs${suffix}`);
    },
    dataTransferJob: (id: string) =>
      request<DataTransferJobDetails>(`/data-transfer/jobs/${encodeURIComponent(id)}`),
    deleteDataTransferJob: (id: string) =>
      request<void>(`/data-transfer/jobs/${encodeURIComponent(id)}`, { method: "DELETE" }),
    retryDataTransferJob: (id: string) =>
      request<DataTransferJob>(`/data-transfer/jobs/${encodeURIComponent(id)}/retry`, { method: "POST" }),
    createDataTransferExportJob: (input: DataTransferJobInput) =>
      request<DataTransferJob>("/data-transfer/jobs/export", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    executeDataTransferExportJob: (id: string) =>
      request<DataTransferExportResult>(`/data-transfer/jobs/${encodeURIComponent(id)}/execute`, {
        method: "POST"
      }),
    createDataTransferImportJob: (input: DataTransferJobInput) =>
      request<DataTransferJob>("/data-transfer/jobs/import", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    uploadDataTransferImport: (id: string, file: File, mapping?: DataTransferCSVMappingInput) => {
      const form = new FormData();
      if (mapping) form.append("mapping", JSON.stringify(mapping));
      form.append("file", file);
      return request<DataTransferPreview>(`/data-transfer/jobs/${encodeURIComponent(id)}/upload`, {
        method: "POST",
        body: form
      }, { timeoutMs: 120000 });
    },
    resolveDataTransferIssue: (jobId: string, issueId: string, resolution: DataTransferIssueResolution) =>
      request<DataTransferJob>(
        `/data-transfer/jobs/${encodeURIComponent(jobId)}/issues/${encodeURIComponent(issueId)}`,
        { method: "PUT", body: JSON.stringify({ resolution }) }
      ),
    confirmDataTransferImport: (id: string, expectedRevision: number) =>
      request<DataTransferJob>(`/data-transfer/jobs/${encodeURIComponent(id)}/confirm`, {
        method: "POST",
        body: JSON.stringify({ confirm: true, expectedRevision })
      }),
    cancelDataTransferImport: (id: string) =>
      request<DataTransferJob>(`/data-transfer/jobs/${encodeURIComponent(id)}/cancel`, { method: "POST" }),
    dataTransferArtifactDownloadUrl: (id: string) =>
      `/api/v1/data-transfer/artifacts/${encodeURIComponent(id)}/download`,
    deleteDataTransferArtifact: (id: string) =>
      request<void>(`/data-transfer/artifacts/${encodeURIComponent(id)}`, { method: "DELETE" }),
    openDataTransferArtifactFolder: () =>
      request<void>("/data-transfer/artifacts/open-folder", { method: "POST" })
  };
}
