import { type SetStateAction, useCallback, useEffect, useMemo, useRef, useState } from "react";

import { api, ApiError } from "../../shared/api";
import {
  type DataTransferArea,
  type DataTransferExportResult,
  type DataTransferIssueResolution,
  type DataTransferJob,
  type DataTransferJobDetails,
  type DataTransferJobFilter,
  type DataTransferPreview,
  type DataTransferProfile,
  type DataTransferProfileInput,
  type DataTransferSummary,
  emptyDataTransferSummary
} from "./dataTransferModel";

const allAreas: DataTransferArea[] = ["vehicles", "accessories", "exhibitionLists"];
const openJobStates = new Set<DataTransferJob["state"]>([
  "draft",
  "reading",
  "review_required",
  "ready",
  "running"
]);

export type DataTransferDialog =
  | { kind: "import"; profileId?: string; jobId?: string }
  | { kind: "export"; profileId?: string; jobId?: string }
  | { kind: "profile"; profileId?: string };

export type DataTransferCapabilities = {
  canImport: boolean;
  canExport: boolean;
  canRetryImport: boolean;
  canRetryExport: boolean;
  canCreateProfiles: boolean;
  canUpdateProfiles: boolean;
  canDisableProfiles: boolean;
  canDeleteArtifacts: boolean;
  canOpenFolder: boolean;
};

export function useDataTransferWorkspace(roles: string[] = []) {
  const isAdmin = roles.includes("Admin");
  const isEditor = roles.includes("Editor");
  const isMesse = roles.includes("Messe");
  const canRead = isAdmin || isEditor || isMesse || roles.includes("Viewer") || roles.includes("Planner");
  const messeOnly = isMesse && !isAdmin && !isEditor;
  const availableAreas = useMemo<DataTransferArea[]>(
    () => messeOnly ? ["exhibitionLists"] : [...allAreas],
    [messeOnly]
  );

  const [summary, setSummary] = useState<DataTransferSummary>(emptyDataTransferSummary);
  const [profiles, setProfiles] = useState<DataTransferProfile[]>([]);
  const [jobs, setJobs] = useState<DataTransferJob[]>([]);
  const [allJobs, setAllJobs] = useState<DataTransferJob[]>([]);
  const [selectedJobId, setSelectedJobId] = useState<string | null>(null);
  const selectedJobIdRef = useRef<string | null>(null);
  const mountedRef = useRef(true);
  const dashboardRequestRef = useRef(0);
  const detailRequestRef = useRef(0);
  const [selectedJobDetails, setSelectedJobDetails] = useState<DataTransferJobDetails | null>(null);
  const [filters, setFiltersState] = useState<DataTransferJobFilter>({ states: [], limit: 100 });
  const [detailRevision, setDetailRevision] = useState(0);
  const [dialog, setDialog] = useState<DataTransferDialog | null>(null);
  const [loading, setLoading] = useState(true);
  const [detailLoading, setDetailLoading] = useState(false);
  const [mutating, setMutating] = useState(false);
  const [error, setError] = useState("");
  const [detailError, setDetailError] = useState("");
  const [reuploadRequiredJobIds, setReuploadRequiredJobIds] = useState<Set<string>>(() => new Set());

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      dashboardRequestRef.current += 1;
      detailRequestRef.current += 1;
    };
  }, []);

  const visibleProfile = useCallback(
    (profile: DataTransferProfile) => !messeOnly || profile.areas.every((area) => area === "exhibitionLists"),
    [messeOnly]
  );
  const visibleJob = useCallback(
    (job: DataTransferJob) => !messeOnly || job.areas.every((area) => area === "exhibitionLists"),
    [messeOnly]
  );

  const loadWorkspace = useCallback(async (preferredJob?: DataTransferJob) => {
    const requestId = ++dashboardRequestRef.current;
    setError("");
    setLoading(true);
    try {
      const unfiltered = isUnfilteredJobFilter(filters);
      const loadedJobsRequest = api.dataTransferJobs(filters);
      const allJobsRequest = unfiltered
        ? loadedJobsRequest
        : api.dataTransferJobs({ states: [], limit: 100 });
      const [nextSummary, nextProfiles, loadedJobs, loadedAllJobs] = await Promise.all([
        api.dataTransferSummary(),
        api.dataTransferProfiles(),
        loadedJobsRequest,
        allJobsRequest
      ]);
      if (!mountedRef.current || requestId !== dashboardRequestRef.current) return;

      const visibleJobs = loadedJobs.filter(visibleJob);
      const visibleAllJobs = loadedAllJobs.filter(visibleJob);
      const nextJobs = preferredJob && visibleJob(preferredJob) && !visibleJobs.some((job) => job.id === preferredJob.id)
        ? [preferredJob, ...visibleJobs]
        : visibleJobs;
      const nextAllJobs = preferredJob && visibleJob(preferredJob) &&
        !visibleAllJobs.some((job) => job.id === preferredJob.id)
        ? [preferredJob, ...visibleAllJobs]
        : visibleAllJobs;
      const preferredId = preferredJob?.id;
      const retainedId = selectedJobIdRef.current && nextJobs.some((job) => job.id === selectedJobIdRef.current)
        ? selectedJobIdRef.current
        : null;
      const firstOpenJob = nextJobs.find((job) => openJobStates.has(job.state));
      const nextSelectedId = preferredId || retainedId || firstOpenJob?.id || nextJobs[0]?.id || null;

      setSummary(nextSummary);
      setProfiles(nextProfiles.filter(visibleProfile));
      setJobs(nextJobs);
      setAllJobs(nextAllJobs);
      selectedJobIdRef.current = nextSelectedId;
      setSelectedJobId(nextSelectedId);
      setDetailRevision((revision) => revision + 1);
    } catch (loadError) {
      if (mountedRef.current && requestId === dashboardRequestRef.current) {
        setError(errorMessage(loadError));
      }
    } finally {
      if (mountedRef.current && requestId === dashboardRequestRef.current) {
        setLoading(false);
      }
    }
  }, [filters, visibleJob, visibleProfile]);

  useEffect(() => {
    void loadWorkspace();
  }, [loadWorkspace]);

  useEffect(() => {
    const requestId = ++detailRequestRef.current;
    const id = selectedJobId;
    setSelectedJobDetails(null);
    setDetailError("");
    if (!id) {
      setDetailLoading(false);
      return;
    }

    setDetailLoading(true);
    void api.dataTransferJob(id)
      .then((details) => {
        if (mountedRef.current && requestId === detailRequestRef.current &&
          selectedJobIdRef.current === id && visibleJob(details.job)) {
          setSelectedJobDetails(details);
        }
      })
      .catch((loadError: unknown) => {
        if (mountedRef.current && requestId === detailRequestRef.current && selectedJobIdRef.current === id) {
          setDetailError(errorMessage(loadError));
        }
      })
      .finally(() => {
        if (mountedRef.current && requestId === detailRequestRef.current) {
          setDetailLoading(false);
        }
      });
  }, [detailRevision, selectedJobId, visibleJob]);

  const setFilters = useCallback((next: SetStateAction<DataTransferJobFilter>) => {
    dashboardRequestRef.current += 1;
    setFiltersState(next);
  }, []);

  const selectJob = useCallback((id: string | null) => {
    if (selectedJobIdRef.current === id) return;

    detailRequestRef.current += 1;
    selectedJobIdRef.current = id;
    setSelectedJobId(id);
    setSelectedJobDetails(null);
    setDetailError("");
    if (!id) setDetailLoading(false);
  }, []);

  const runMutation = useCallback(async <T,>(mutation: () => Promise<T>, jobFromResult?: (result: T) => DataTransferJob) => {
    setMutating(true);
    setError("");
    try {
      const result = await mutation();
      await loadWorkspace(jobFromResult?.(result));
      return result;
    } catch (mutationError) {
      if (mountedRef.current) setError(errorMessage(mutationError));
      throw mutationError;
    } finally {
      if (mountedRef.current) setMutating(false);
    }
  }, [loadWorkspace]);

  const createProfile = useCallback((input: DataTransferProfileInput) =>
    runMutation(() => api.createDataTransferProfile(input)), [runMutation]);
  const updateProfile = useCallback((id: string, input: DataTransferProfileInput) =>
    runMutation(() => api.updateDataTransferProfile(id, input)), [runMutation]);
  const disableProfile = useCallback((id: string) =>
    runMutation(() => api.disableDataTransferProfile(id)), [runMutation]);
  const createImportJob = useCallback((profileId: string) =>
    runMutation(() => api.createDataTransferImportJob({ profileId }), (job) => job), [runMutation]);
  const uploadImportFile = useCallback(async (jobId: string, file: File): Promise<DataTransferPreview> => {
    const preview = await runMutation(() => api.uploadDataTransferImport(jobId, file), (result) => result.job);
    setReuploadRequiredJobIds((current) => {
      if (!current.has(jobId)) return current;
      const next = new Set(current);
      next.delete(jobId);
      return next;
    });
    return preview;
  }, [runMutation]);
  const resolveIssue = useCallback((jobId: string, issueId: string, resolution: DataTransferIssueResolution) =>
    runMutation(() => api.resolveDataTransferIssue(jobId, issueId, resolution), (job) => job), [runMutation]);
  const confirmImport = useCallback(async (jobId: string) => {
    try {
      return await runMutation(() => api.confirmDataTransferImport(jobId), (job) => job);
    } catch (confirmError) {
      if (confirmError instanceof ApiError && confirmError.status === 409) {
        setReuploadRequiredJobIds((current) => new Set(current).add(jobId));
      }
      throw confirmError;
    }
  }, [runMutation]);
  const cancelImport = useCallback((jobId: string) =>
    runMutation(() => api.cancelDataTransferImport(jobId), (job) => job), [runMutation]);
  const createExportJob = useCallback((profileId: string) =>
    runMutation(() => api.createDataTransferExportJob({ profileId }), (job) => job), [runMutation]);
  const executeExportJob = useCallback((jobId: string): Promise<DataTransferExportResult> =>
    runMutation(() => api.executeDataTransferExportJob(jobId), (result) => result.job), [runMutation]);
  const retryJob = useCallback((jobId: string) =>
    runMutation(() => api.retryDataTransferJob(jobId), (job) => job), [runMutation]);
  const refreshJobDetails = useCallback(async (jobId: string) => {
    const details = await api.dataTransferJob(jobId);
    if (!visibleJob(details.job)) throw new Error("Der Auftrag ist für diese Rolle nicht verfügbar.");
    if (mountedRef.current) {
      setError("");
      setSelectedJobDetails(details);
      setJobs((current) => current.map((job) => job.id === details.job.id ? details.job : job));
      setAllJobs((current) => current.map((job) => job.id === details.job.id ? details.job : job));
    }
    return details;
  }, [visibleJob]);
  const deleteArtifact = useCallback((artifactId: string) =>
    runMutation(() => api.deleteDataTransferArtifact(artifactId)), [runMutation]);
  const openArtifactFolder = useCallback(() => {
    if (!isAdmin || !summary.openFolderAvailable) {
      return Promise.reject(new Error("Opening the export folder is unavailable."));
    }
    return runMutation(() => api.openDataTransferArtifactFolder());
  }, [isAdmin, runMutation, summary.openFolderAvailable]);

  const capabilities = useMemo<DataTransferCapabilities>(() => ({
    canImport: isAdmin || isEditor || isMesse,
    canExport: canRead,
    canRetryImport: isAdmin || isEditor || isMesse,
    canRetryExport: isAdmin || isEditor || isMesse,
    canCreateProfiles: isAdmin || isEditor,
    canUpdateProfiles: isAdmin || isEditor,
    canDisableProfiles: isAdmin,
    canDeleteArtifacts: isAdmin,
    canOpenFolder: isAdmin && summary.openFolderAvailable
  }), [canRead, isAdmin, isEditor, isMesse, summary.openFolderAvailable]);

  const selectedJob = jobs.find((job) => job.id === selectedJobId) || selectedJobDetails?.job || null;

  return {
    summary,
    profiles,
    jobs,
    allJobs,
    filters,
    setFilters,
    selectedJobId,
    selectedJob,
    selectedJobDetails,
    selectJob,
    dialog,
    openDialog: (kind: DataTransferDialog["kind"], profileId?: string, jobId?: string) =>
      setDialog({ kind, profileId, ...(kind === "profile" ? {} : { jobId }) }),
    closeDialog: () => setDialog(null),
    loading,
    detailLoading,
    mutating,
    error,
    detailError,
    availableAreas,
    capabilities,
    refresh: () => loadWorkspace(),
    createProfile,
    updateProfile,
    disableProfile,
    createImportJob,
    uploadImportFile,
    resolveIssue,
    confirmImport,
    importRequiresReupload: (jobId?: string) => Boolean(jobId && reuploadRequiredJobIds.has(jobId)),
    cancelImport,
    createExportJob,
    executeExportJob,
    retryJob,
    refreshJobDetails,
    artifactDownloadUrl: api.dataTransferArtifactDownloadUrl,
    deleteArtifact,
    openArtifactFolder
  };
}

function isUnfilteredJobFilter(filters: DataTransferJobFilter) {
  return !filters.profileId && !filters.direction && !filters.states?.length && (filters.limit ?? 100) === 100;
}

function errorMessage(error: unknown) {
  return error instanceof Error ? error.message : "Der Datentransfer konnte nicht verarbeitet werden.";
}
