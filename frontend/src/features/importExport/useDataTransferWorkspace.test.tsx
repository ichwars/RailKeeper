import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import type {
  DataTransferJob,
  DataTransferJobDetails,
  DataTransferProfile,
  DataTransferSummary
} from "./dataTransferModel";
import { useDataTransferWorkspace } from "./useDataTransferWorkspace";

const summary: DataTransferSummary = {
  openJobs: 3,
  selectedRecords: 42,
  lastExportAt: "2026-08-20T08:00:00Z",
  artifactCount: 2,
  artifactBytes: 2048,
  openFolderAvailable: true,
  artifactDirectory: "C:\\RailKeeper\\exports"
};

const vehicleProfile = profileFixture({ id: "profile-vehicles", areas: ["vehicles"] });
const exhibitionProfile = profileFixture({ id: "profile-exhibition", areas: ["exhibitionLists"] });
const completedJob = jobFixture({ id: "job-completed", state: "completed" });
const reviewJob = jobFixture({ id: "job-review", state: "review_required" });

describe("useDataTransferWorkspace", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(api, "dataTransferSummary").mockResolvedValue(summary);
    vi.spyOn(api, "dataTransferProfiles").mockResolvedValue([vehicleProfile, exhibitionProfile]);
    vi.spyOn(api, "dataTransferJobs").mockResolvedValue([completedJob, reviewJob]);
    vi.spyOn(api, "dataTransferJob").mockImplementation(async (id) => detailsFixture(
      id === reviewJob.id ? reviewJob : completedJob
    ));
  });

  it("loads dashboard data and selects the first open job", async () => {
    const { result } = renderHook(() => useDataTransferWorkspace(["Editor"]));

    await waitFor(() => expect(result.current.loading).toBe(false));
    await waitFor(() => expect(result.current.selectedJobDetails?.job.id).toBe("job-review"));

    expect(result.current.summary.openJobs).toBe(3);
    expect(result.current.selectedJob?.state).toBe("review_required");
    expect(api.dataTransferSummary).toHaveBeenCalledOnce();
    expect(api.dataTransferProfiles).toHaveBeenCalledOnce();
    expect(api.dataTransferJobs).toHaveBeenCalledWith({ limit: 100, states: [] });
  });

  it("limits a Messe workspace to exhibition lists and hides administrative actions", async () => {
    const hiddenVehicleJob = jobFixture({ id: "job-vehicle", areas: ["vehicles"] });
    const exhibitionJob = jobFixture({ id: "job-exhibition", areas: ["exhibitionLists"] });
    vi.mocked(api.dataTransferJobs).mockResolvedValue([hiddenVehicleJob, exhibitionJob]);

    const { result } = renderHook(() => useDataTransferWorkspace(["Messe", "Viewer"]));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.availableAreas).toEqual(["exhibitionLists"]);
    expect(result.current.profiles).toEqual([exhibitionProfile]);
    expect(result.current.jobs).toEqual([exhibitionJob]);
    expect(result.current.capabilities).toMatchObject({
      canImport: true,
      canExport: true,
      canCreateProfiles: false,
      canUpdateProfiles: false,
      canDisableProfiles: false,
      canDeleteArtifacts: false,
      canOpenFolder: false
    });
  });

  it("separates Editor profile editing from Admin-only profile disabling", async () => {
    const { result, rerender } = renderHook(
      ({ roles }: { roles: string[] }) => useDataTransferWorkspace(roles),
      { initialProps: { roles: ["Editor"] } }
    );
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.capabilities).toMatchObject({
      canCreateProfiles: true,
      canUpdateProfiles: true,
      canDisableProfiles: false
    });

    rerender({ roles: ["Admin"] });
    await waitFor(() => expect(result.current.capabilities.canDisableProfiles).toBe(true));
    expect(result.current.capabilities).toMatchObject({
      canCreateProfiles: true,
      canUpdateProfiles: true,
      canDisableProfiles: true
    });
  });

  it("applies filters and loads a selected job detail", async () => {
    const { result } = renderHook(() => useDataTransferWorkspace(["Admin"]));
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => result.current.setFilters({ direction: "export", states: ["failed"], limit: 25 }));
    await waitFor(() => expect(api.dataTransferJobs).toHaveBeenLastCalledWith({
      direction: "export",
      states: ["failed"],
      limit: 25
    }));

    act(() => result.current.selectJob(completedJob.id));
    await waitFor(() => expect(result.current.selectedJobDetails?.job.id).toBe(completedJob.id));
  });

  it("runs named mutations through the transfer API and refreshes workspace state", async () => {
    const created = jobFixture({ id: "job-created", state: "draft", direction: "export" });
    vi.spyOn(api, "createDataTransferExportJob").mockResolvedValue(created);
    vi.spyOn(api, "executeDataTransferExportJob").mockResolvedValue({
      job: { ...created, state: "completed" },
      artifact: {
        id: "artifact-1",
        jobId: created.id,
        relativePath: "exports/export.json",
        displayName: "export.json",
        mimeType: "application/json",
        sizeBytes: 128,
        sha256: "abc",
        deletedAt: "",
        createdAt: "2026-08-20T10:00:00Z"
      },
      openFolderAvailable: true
    });
    vi.spyOn(api, "openDataTransferArtifactFolder").mockResolvedValue(undefined);

    const { result } = renderHook(() => useDataTransferWorkspace(["Admin"]));
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(() => result.current.createExportJob(vehicleProfile.id));
    expect(api.createDataTransferExportJob).toHaveBeenCalledWith({ profileId: vehicleProfile.id });
    expect(result.current.selectedJobId).toBe(created.id);

    await act(() => result.current.executeExportJob(created.id));
    expect(api.executeDataTransferExportJob).toHaveBeenCalledWith(created.id);

    await act(() => result.current.openArtifactFolder());
    expect(api.openDataTransferArtifactFolder).toHaveBeenCalledOnce();
    expect(result.current.artifactDownloadUrl("artifact-1")).toBe(
      "/api/v1/data-transfer/artifacts/artifact-1/download"
    );
  });

  it("keeps the newest filtered dashboard response when an older request finishes last", async () => {
    const oldSummary = deferred<DataTransferSummary>();
    const oldProfiles = deferred<DataTransferProfile[]>();
    const oldJobs = deferred<DataTransferJob[]>();
    const filteredJob = jobFixture({ id: "job-filtered", direction: "export", state: "failed" });
    const filteredSummary = { ...summary, openJobs: 0 };
    vi.mocked(api.dataTransferSummary)
      .mockReturnValueOnce(oldSummary.promise)
      .mockResolvedValue(filteredSummary);
    vi.mocked(api.dataTransferProfiles)
      .mockReturnValueOnce(oldProfiles.promise)
      .mockResolvedValue([exhibitionProfile]);
    vi.mocked(api.dataTransferJobs)
      .mockReturnValueOnce(oldJobs.promise)
      .mockResolvedValue([filteredJob]);

    const { result } = renderHook(() => useDataTransferWorkspace(["Admin"]));
    act(() => result.current.setFilters({ direction: "export", states: ["failed"], limit: 25 }));
    await waitFor(() => expect(api.dataTransferJobs).toHaveBeenCalledTimes(2));
    await waitFor(() => expect(result.current.selectedJob?.id).toBe(filteredJob.id));

    await act(async () => {
      oldSummary.resolve(summary);
      oldProfiles.resolve([vehicleProfile]);
      oldJobs.resolve([reviewJob]);
      await oldJobs.promise;
    });

    expect(result.current.summary.openJobs).toBe(0);
    expect(result.current.jobs).toEqual([filteredJob]);
    expect(result.current.selectedJob?.id).toBe(filteredJob.id);
  });

  it("ignores stale job details after a newer selection", async () => {
    const oldDetails = deferred<DataTransferJobDetails>();
    vi.mocked(api.dataTransferJobs).mockResolvedValue([reviewJob, completedJob]);
    vi.mocked(api.dataTransferJob).mockImplementation((id) =>
      id === reviewJob.id ? oldDetails.promise : Promise.resolve(detailsFixture(completedJob))
    );

    const { result } = renderHook(() => useDataTransferWorkspace(["Admin"]));
    await waitFor(() => expect(api.dataTransferJob).toHaveBeenCalledWith(reviewJob.id));
    act(() => result.current.selectJob(completedJob.id));
    await waitFor(() => expect(result.current.selectedJobDetails?.job.id).toBe(completedJob.id));

    await act(async () => oldDetails.resolve(detailsFixture(reviewJob)));

    expect(result.current.selectedJobId).toBe(completedJob.id);
    expect(result.current.selectedJobDetails?.job.id).toBe(completedJob.id);
  });

  it("does not repopulate job details after selection is cleared", async () => {
    const oldDetails = deferred<DataTransferJobDetails>();
    vi.mocked(api.dataTransferJobs).mockResolvedValue([reviewJob]);
    vi.mocked(api.dataTransferJob).mockReturnValue(oldDetails.promise);

    const { result } = renderHook(() => useDataTransferWorkspace(["Admin"]));
    await waitFor(() => expect(api.dataTransferJob).toHaveBeenCalledWith(reviewJob.id));
    act(() => result.current.selectJob(null));
    await act(async () => oldDetails.resolve(detailsFixture(reviewJob)));

    expect(result.current.selectedJobId).toBeNull();
    expect(result.current.selectedJobDetails).toBeNull();
  });

  it("does not continue a pending dashboard load after unmount", async () => {
    const pendingSummary = deferred<DataTransferSummary>();
    const pendingProfiles = deferred<DataTransferProfile[]>();
    const pendingJobs = deferred<DataTransferJob[]>();
    vi.mocked(api.dataTransferSummary).mockReturnValue(pendingSummary.promise);
    vi.mocked(api.dataTransferProfiles).mockReturnValue(pendingProfiles.promise);
    vi.mocked(api.dataTransferJobs).mockReturnValue(pendingJobs.promise);

    const { unmount } = renderHook(() => useDataTransferWorkspace(["Admin"]));
    await waitFor(() => expect(api.dataTransferJobs).toHaveBeenCalledOnce());
    unmount();
    await act(async () => {
      pendingSummary.resolve(summary);
      pendingProfiles.resolve([vehicleProfile]);
      pendingJobs.resolve([reviewJob]);
      await pendingJobs.promise;
    });

    expect(api.dataTransferJob).not.toHaveBeenCalled();
  });

  it("keeps dashboard data visible when the selected job detail disappears", async () => {
    vi.mocked(api.dataTransferJob).mockRejectedValue(new Error("Transfer job not found"));

    const { result } = renderHook(() => useDataTransferWorkspace(["Editor"]));
    await waitFor(() => expect(result.current.loading).toBe(false));
    await waitFor(() => expect(result.current.detailLoading).toBe(false));

    expect(result.current.summary).toEqual(summary);
    expect(result.current.profiles).toEqual([vehicleProfile, exhibitionProfile]);
    expect(result.current.jobs).toEqual([completedJob, reviewJob]);
    expect(result.current.selectedJob?.id).toBe(reviewJob.id);
    expect(result.current.selectedJobDetails).toBeNull();
    expect(result.current.error).toBe("");
    expect(result.current.detailError).toBe("Transfer job not found");
  });
});

describe("data transfer API", () => {
  const fetchMock = vi.fn().mockResolvedValue({
    ok: true,
    status: 200,
    statusText: "OK",
    json: async () => ({})
  });

  beforeEach(() => {
    vi.restoreAllMocks();
    fetchMock.mockClear();
    vi.stubGlobal("fetch", fetchMock);
    document.cookie = "rk_csrf=transfer-test-token; path=/";
  });

  afterEach(() => vi.unstubAllGlobals());

  it("uses the documented transfer routes, methods, filters, and multipart upload", async () => {
    const profileInput = {
      name: "Fahrzeuge",
      direction: "export" as const,
      format: "railkeeper-json" as const,
      areas: ["vehicles" as const]
    };
    const file = new File(["payload"], "transfer.json", { type: "application/json" });

    await api.dataTransferSummary();
    await api.dataTransferProfiles();
    await api.createDataTransferProfile(profileInput);
    await api.updateDataTransferProfile("profile/1", profileInput);
    await api.disableDataTransferProfile("profile/1");
    await api.dataTransferJobs({ profileId: "profile/1", direction: "import", states: ["ready", "failed"], limit: 25 });
    await api.dataTransferJob("job/1");
    await api.retryDataTransferJob("job/1");
    await api.createDataTransferExportJob({ profileId: "profile/1" });
    await api.executeDataTransferExportJob("job/1");
    await api.createDataTransferImportJob({ profileId: "profile/1" });
    await api.uploadDataTransferImport("job/1", file);
    await api.resolveDataTransferIssue("job/1", "issue/1", "use_existing");
    await api.confirmDataTransferImport("job/1");
    await api.cancelDataTransferImport("job/1");
    await api.deleteDataTransferArtifact("artifact/1");
    await api.openDataTransferArtifactFolder();

    expect(fetchMock.mock.calls.map(([url, init]) => [url, init.method ?? "GET"])).toEqual([
      ["/api/v1/data-transfer/summary", "GET"],
      ["/api/v1/data-transfer/profiles", "GET"],
      ["/api/v1/data-transfer/profiles", "POST"],
      ["/api/v1/data-transfer/profiles/profile%2F1", "PUT"],
      ["/api/v1/data-transfer/profiles/profile%2F1", "DELETE"],
      ["/api/v1/data-transfer/jobs?profileId=profile%2F1&direction=import&states=ready&states=failed&limit=25", "GET"],
      ["/api/v1/data-transfer/jobs/job%2F1", "GET"],
      ["/api/v1/data-transfer/jobs/job%2F1/retry", "POST"],
      ["/api/v1/data-transfer/jobs/export", "POST"],
      ["/api/v1/data-transfer/jobs/job%2F1/execute", "POST"],
      ["/api/v1/data-transfer/jobs/import", "POST"],
      ["/api/v1/data-transfer/jobs/job%2F1/upload", "POST"],
      ["/api/v1/data-transfer/jobs/job%2F1/issues/issue%2F1", "PUT"],
      ["/api/v1/data-transfer/jobs/job%2F1/confirm", "POST"],
      ["/api/v1/data-transfer/jobs/job%2F1/cancel", "POST"],
      ["/api/v1/data-transfer/artifacts/artifact%2F1", "DELETE"],
      ["/api/v1/data-transfer/artifacts/open-folder", "POST"]
    ]);
    const uploadInit = fetchMock.mock.calls[11]?.[1];
    expect(uploadInit.body).toBeInstanceOf(FormData);
    expect(uploadInit.headers["Content-Type"]).toBeUndefined();
    expect(uploadInit.headers["X-CSRF-Token"]).toBe("transfer-test-token");
    expect(api.dataTransferArtifactDownloadUrl("artifact/1")).toBe(
      "/api/v1/data-transfer/artifacts/artifact%2F1/download"
    );
  });
});

function profileFixture(overrides: Partial<DataTransferProfile> = {}): DataTransferProfile {
  return {
    id: "profile-1",
    name: "Fahrzeuge",
    direction: "import",
    format: "railkeeper-json",
    areas: ["vehicles"],
    options: {},
    enabled: true,
    createdByUserId: "user-1",
    createdAt: "2026-08-20T08:00:00Z",
    updatedAt: "2026-08-20T08:00:00Z",
    ...overrides
  };
}

function jobFixture(overrides: Partial<DataTransferJob> = {}): DataTransferJob {
  return {
    id: "job-1",
    profileId: "profile-1",
    profileName: "Fahrzeuge",
    direction: "import",
    format: "railkeeper-json",
    areas: ["vehicles"],
    options: {},
    state: "draft",
    stage: "created",
    sourceName: "",
    sourceSha256: "",
    packageVersion: 1,
    revision: 1,
    totalRecords: 0,
    readyRecords: 0,
    warningRecords: 0,
    errorRecords: 0,
    preview: {},
    createdByUserId: "user-1",
    confirmedByUserId: "",
    confirmedAt: "",
    completedAt: "",
    resultMessage: "",
    createdAt: "2026-08-20T08:00:00Z",
    updatedAt: "2026-08-20T08:00:00Z",
    ...overrides
  };
}

function detailsFixture(job: DataTransferJob): DataTransferJobDetails {
  return { job, issues: [], artifacts: [] };
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((promiseResolve, promiseReject) => {
    resolve = promiseResolve;
    reject = promiseReject;
  });
  return { promise, resolve, reject };
}
