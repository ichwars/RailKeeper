import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { api, ApiError } from "../../shared/api";
import { translate, useI18n } from "../../shared/i18n";
import {
  type DigitalCenterCompareFilter,
  type DigitalCenterProvider,
  type DigitalCenterReadSession,
  type DigitalCenterSessionMessage,
  type DigitalCenterSummary,
  type DigitalCenterWorkItem,
  type DigitalCenterWorkItemFilter,
  type DigitalCenterWorkspaceDialog,
  type DigitalCenterWorkspaceErrors,
  type DigitalCenterWorkspaceTab,
  type DigitalCenterWriteConfirmation,
  type DigitalCenterWriteField,
  type DigitalCenterWritePreview,
  type ECoSLiveStatus,
  emptyDigitalCenterWorkItemPage
} from "./digitalCenterModel";

type DigitalCenterWorkspaceOptions = {
  pollIntervalMs?: number;
};

type LoadingState = {
  workspace: boolean;
  live: boolean;
  read: boolean;
  worklist: boolean;
  detail: boolean;
  write: boolean;
};

const emptyErrors: DigitalCenterWorkspaceErrors = {
  workspace: "",
  live: "",
  read: "",
  worklist: "",
  detail: "",
  messages: "",
  write: ""
};

const unfilteredSessionTotal: DigitalCenterWorkItemFilter = {
  query: "",
  compareStatus: "all",
  page: 1,
  pageSize: 1
};

export function useDigitalCentersWorkspace(options: DigitalCenterWorkspaceOptions = {}) {
  const { language } = useI18n();
  const languageRef = useRef(language);
  languageRef.current = language;
  const workspaceText = useCallback(
    (key: string, values?: Record<string, string | number>) => translate(languageRef.current, key, values),
    []
  );
  const localizedError = useCallback((error: unknown) => {
	if (error instanceof ApiError && error.code === "digital_center_address_conflict") {
	  return digitalCenterAddressConflictMessage(error, workspaceText);
	}
    const translationKey = error instanceof ApiError ? digitalCenterErrorTranslationKey(error.code) : null;
    return translationKey
      ? workspaceText(translationKey)
      : errorMessage(error, workspaceText("digitalCenters.error.requestFailed"));
  }, [workspaceText]);
  const pollIntervalMs = options.pollIntervalMs ?? 1000;
  const mountedRef = useRef(true);
  const selectedProviderRef = useRef<DigitalCenterProvider | null>(null);
  const readSessionIDRef = useRef<string | null>(null);
  const selectedItemIDRef = useRef<string | null>(null);
  const requestsRef = useRef({ workspace: 0, live: 0, read: 0, worklist: 0, detail: 0, messages: 0, write: 0 });

  const [centers, setCenters] = useState<DigitalCenterSummary[]>([]);
  const [selectedProvider, setSelectedProvider] = useState<DigitalCenterProvider | null>(null);
  const [liveStatus, setLiveStatus] = useState<ECoSLiveStatus | null>(null);
  const [readSession, setReadSession] = useState<DigitalCenterReadSession | null>(null);
  const [workItems, setWorkItems] = useState(emptyDigitalCenterWorkItemPage);
  const [sessionTotal, setSessionTotal] = useState(0);
  const [messages, setMessages] = useState<DigitalCenterSessionMessage[]>([]);
  const [selectedItemId, setSelectedItemID] = useState<string | null>(null);
  const [selectedItem, setSelectedItem] = useState<DigitalCenterWorkItem | null>(null);
  const [search, setSearchState] = useState("");
  const [compareStatus, setCompareStatusState] = useState<DigitalCenterCompareFilter>("all");
  const [page, setPageState] = useState(1);
  const [pageSize, setPageSizeState] = useState(10);
  const [tab, setTab] = useState<DigitalCenterWorkspaceTab>("live");
  const [dialog, setDialog] = useState<DigitalCenterWorkspaceDialog | null>(null);
  const [writePreview, setWritePreview] = useState<DigitalCenterWritePreview | null>(null);
  const [writeConfirmation, setWriteConfirmation] = useState<DigitalCenterWriteConfirmation | null>(null);
  const [errors, setErrors] = useState<DigitalCenterWorkspaceErrors>(emptyErrors);
  const [loading, setLoading] = useState<LoadingState>({
    workspace: true,
    live: false,
    read: false,
    worklist: false,
    detail: false,
    write: false
  });

  const selectedCenter = useMemo(
    () => centers.find((center) => center.provider === selectedProvider) ?? null,
    [centers, selectedProvider]
  );
  const actions = useMemo(() => {
    const capabilities = selectedCenter?.capabilities;
    const active = Boolean(selectedCenter?.active);
    return {
      canTestConnection: active && Boolean(capabilities?.testConnection),
      canRead: active && Boolean(capabilities?.readLocomotives),
      canMonitor: active && Boolean(capabilities?.liveMonitor),
      canWrite: active && Boolean(capabilities?.writeLocomotives),
      canWriteCVs: active && Boolean(capabilities?.writeCVs),
      canDiagnose: active && Boolean(capabilities?.diagnose)
    };
  }, [selectedCenter]);

  const setError = useCallback((area: keyof DigitalCenterWorkspaceErrors, message: string) => {
    setErrors((current) => ({ ...current, [area]: message }));
  }, []);
  const setLoadingArea = useCallback((area: keyof LoadingState, value: boolean) => {
    setLoading((current) => ({ ...current, [area]: value }));
  }, []);
  const resetWriteDialogState = useCallback(() => {
    requestsRef.current.write += 1;
    setWritePreview(null);
    setWriteConfirmation(null);
    setLoading((current) => ({ ...current, write: false }));
    setErrors((current) => ({ ...current, write: "" }));
  }, []);
  const suspendSessionActions = useCallback(() => {
    requestsRef.current.detail += 1;
    requestsRef.current.write += 1;
    selectedItemIDRef.current = null;
    setSelectedItemID(null);
    setSelectedItem(null);
    setDialog(null);
    setWritePreview(null);
    setWriteConfirmation(null);
    setLoading((current) => ({ ...current, detail: false, write: false }));
  }, []);
  const clearSessionDependents = useCallback((preserveWorklist = false) => {
    requestsRef.current.worklist += 1;
    requestsRef.current.detail += 1;
    requestsRef.current.messages += 1;
    requestsRef.current.write += 1;
    selectedItemIDRef.current = null;
    if (!preserveWorklist) {
      setWorkItems(emptyDigitalCenterWorkItemPage);
      setSessionTotal(0);
    }
    setMessages([]);
    setSelectedItemID(null);
    setSelectedItem(null);
    setDialog(null);
    setWritePreview(null);
    setWriteConfirmation(null);
    setLoading((current) => ({ ...current, worklist: false, detail: false, write: false }));
  }, []);

  const loadWorkspace = useCallback(async () => {
    const requestID = ++requestsRef.current.workspace;
    setError("workspace", "");
    setLoadingArea("workspace", true);
    try {
      const response = await api.digitalCenterWorkspace();
      if (!mountedRef.current || requestID !== requestsRef.current.workspace) return;
      setCenters(response.centers);
      const retained = selectedProviderRef.current &&
        response.centers.some((center) => center.provider === selectedProviderRef.current)
        ? selectedProviderRef.current
        : null;
      const next = retained ?? response.centers.find((center) => center.selected)?.provider ??
        response.centers.find((center) => center.active)?.provider ?? response.centers[0]?.provider ?? null;
      selectedProviderRef.current = next;
      setSelectedProvider(next);
    } catch (loadError) {
      if (mountedRef.current && requestID === requestsRef.current.workspace) {
        setError("workspace", localizedError(loadError));
      }
    } finally {
      if (mountedRef.current && requestID === requestsRef.current.workspace) {
        setLoadingArea("workspace", false);
      }
    }
  }, [localizedError, setError, setLoadingArea]);

  const loadLiveStatus = useCallback(async (provider: DigitalCenterProvider) => {
    const requestID = ++requestsRef.current.live;
    setError("live", "");
    setLoadingArea("live", true);
    try {
      const status = normalizeLiveStatus(await api.digitalCenterLiveStatus(provider));
      if (mountedRef.current && requestID === requestsRef.current.live &&
        selectedProviderRef.current === provider) {
        setLiveStatus(status);
      }
      return status;
    } catch (loadError) {
      if (mountedRef.current && requestID === requestsRef.current.live &&
        selectedProviderRef.current === provider) {
        const message = localizedError(loadError);
        setLiveStatus((current) => current?.provider === provider && current.state === "running"
          ? {
            ...current,
            connected: false,
            state: "interrupted",
            pulseSamples: [],
            recentEvents: [],
            diagnosis: { ...current.diagnosis, connectionState: "interrupted", lastError: message },
            error: message,
            message: workspaceText("digitalCenters.error.liveInterrupted")
          }
          : current);
        setError("live", message);
      }
      throw loadError;
    } finally {
      if (mountedRef.current && requestID === requestsRef.current.live) {
        setLoadingArea("live", false);
      }
    }
  }, [localizedError, setError, setLoadingArea, workspaceText]);

  useEffect(() => {
    mountedRef.current = true;
    void loadWorkspace();
    return () => {
      mountedRef.current = false;
      readSessionIDRef.current = null;
      requestsRef.current.workspace += 1;
      requestsRef.current.live += 1;
      requestsRef.current.read += 1;
      requestsRef.current.worklist += 1;
      requestsRef.current.detail += 1;
      requestsRef.current.messages += 1;
      requestsRef.current.write += 1;
    };
  }, [loadWorkspace]);

  useEffect(() => {
    if (!selectedProvider || !actions.canMonitor) {
      requestsRef.current.live += 1;
      setLiveStatus(null);
      setLoadingArea("live", false);
      return;
    }
    void loadLiveStatus(selectedProvider).catch(() => undefined);
  }, [actions.canMonitor, loadLiveStatus, selectedProvider, setLoadingArea]);

  useEffect(() => {
    if (!selectedProvider || !actions.canMonitor || liveStatus?.state !== "running" ||
      liveStatus.provider !== selectedProvider) return;
    const timer = window.setTimeout(() => {
      void loadLiveStatus(selectedProvider).catch(() => undefined);
    }, Math.max(1, pollIntervalMs));
    return () => window.clearTimeout(timer);
  }, [actions.canMonitor, liveStatus, loadLiveStatus, pollIntervalMs, selectedProvider]);

  const filter = useMemo<DigitalCenterWorkItemFilter>(() => ({
    query: search,
    compareStatus,
    page,
    pageSize
  }), [compareStatus, page, pageSize, search]);

  useEffect(() => {
    const sessionID = readSession?.id;
    if (!sessionID) return;
    const requestID = ++requestsRef.current.worklist;
    setError("worklist", "");
    setLoadingArea("worklist", true);
    const aggregateRequest = filter.query.trim() || filter.compareStatus !== "all"
      ? api.digitalCenterWorkItems(sessionID, unfilteredSessionTotal)
      : Promise.resolve(null);
    void Promise.all([api.digitalCenterWorkItems(sessionID, filter), aggregateRequest])
      .then(([result, aggregate]) => {
        if (mountedRef.current && requestID === requestsRef.current.worklist &&
          readSessionIDRef.current === sessionID) {
          setWorkItems(result);
          setSessionTotal(aggregate?.total ?? result.total);
        }
      })
      .catch((loadError: unknown) => {
        if (mountedRef.current && requestID === requestsRef.current.worklist &&
          readSessionIDRef.current === sessionID) {
          setError("worklist", localizedError(loadError));
        }
      })
      .finally(() => {
        if (mountedRef.current && requestID === requestsRef.current.worklist &&
          readSessionIDRef.current === sessionID) {
          setLoadingArea("worklist", false);
        }
      });
  }, [filter, localizedError, readSession?.id, setError, setLoadingArea]);

  useEffect(() => {
    const sessionID = readSession?.id;
    if (!sessionID) return;
    const requestID = ++requestsRef.current.messages;
    setError("messages", "");
    void api.digitalCenterSessionMessages(sessionID)
      .then((result) => {
        if (mountedRef.current && requestID === requestsRef.current.messages &&
          readSessionIDRef.current === sessionID) {
          setMessages(result.messages);
        }
      })
      .catch((loadError: unknown) => {
        if (mountedRef.current && requestID === requestsRef.current.messages &&
          readSessionIDRef.current === sessionID) {
          setError("messages", localizedError(loadError));
        }
      });
  }, [localizedError, readSession?.id, setError]);

  useEffect(() => {
    const sessionID = readSession?.id;
    const itemID = selectedItemId;
    if (!sessionID || !itemID) return;
    const requestID = ++requestsRef.current.detail;
    setError("detail", "");
    setLoadingArea("detail", true);
    void api.digitalCenterWorkItem(sessionID, itemID)
      .then((result) => {
        if (mountedRef.current && requestID === requestsRef.current.detail &&
          readSessionIDRef.current === sessionID &&
          selectedItemIDRef.current === itemID) {
          setSelectedItem(result);
        }
      })
      .catch((loadError: unknown) => {
        if (mountedRef.current && requestID === requestsRef.current.detail &&
          readSessionIDRef.current === sessionID && selectedItemIDRef.current === itemID) {
          setError("detail", localizedError(loadError));
        }
      })
      .finally(() => {
        if (mountedRef.current && requestID === requestsRef.current.detail &&
          readSessionIDRef.current === sessionID && selectedItemIDRef.current === itemID) {
          setLoadingArea("detail", false);
        }
      });
  }, [localizedError, readSession?.id, selectedItemId, setError, setLoadingArea]);

  const selectCenter = useCallback((provider: DigitalCenterProvider) => {
    if (selectedProviderRef.current === provider) return;
    selectedProviderRef.current = provider;
    requestsRef.current.live += 1;
    requestsRef.current.read += 1;
    requestsRef.current.worklist += 1;
    requestsRef.current.detail += 1;
    requestsRef.current.messages += 1;
    requestsRef.current.write += 1;
    readSessionIDRef.current = null;
    selectedItemIDRef.current = null;
    setSelectedProvider(provider);
    setLiveStatus(null);
    setReadSession(null);
    setWorkItems(emptyDigitalCenterWorkItemPage);
    setSessionTotal(0);
    setMessages([]);
    setSelectedItemID(null);
    setSelectedItem(null);
    setDialog(null);
    setWritePreview(null);
    setWriteConfirmation(null);
    setLoading((current) => ({
      workspace: current.workspace,
      live: false,
      read: false,
      worklist: false,
      detail: false,
      write: false
    }));
    setErrors((current) => ({ ...emptyErrors, workspace: current.workspace }));
  }, []);

  const readData = useCallback(async () => {
    const provider = selectedProviderRef.current;
    if (!provider || !actions.canRead) {
      const error = new Error(workspaceText("digitalCenters.error.readUnsupported"));
      setError("read", error.message);
      throw error;
    }
    const requestID = ++requestsRef.current.read;
    suspendSessionActions();
    setError("read", "");
    setError("write", "");
    setLoadingArea("read", true);
    try {
      const session = await api.startDigitalCenterReadSession(provider);
      if (mountedRef.current && requestID === requestsRef.current.read &&
        selectedProviderRef.current === provider) {
        clearSessionDependents(true);
        readSessionIDRef.current = session.id;
        setReadSession(session);
      }
      return session;
    } catch (readError) {
      if (mountedRef.current && requestID === requestsRef.current.read) {
        setError("read", localizedError(readError));
      }
      throw readError;
    } finally {
      if (mountedRef.current && requestID === requestsRef.current.read) {
        setLoadingArea("read", false);
      }
    }
  }, [actions.canRead, clearSessionDependents, localizedError, setError, setLoadingArea, suspendSessionActions,
    workspaceText]);

  const runLiveMutation = useCallback(async (operation: "start" | "stop") => {
    const provider = selectedProviderRef.current;
    if (!provider || !actions.canMonitor) {
      const error = new Error(workspaceText("digitalCenters.error.monitorUnsupported"));
      setError("live", error.message);
      throw error;
    }
    const requestID = ++requestsRef.current.live;
    setError("live", "");
    setLoadingArea("live", true);
    try {
      const sessionID = readSession?.id;
      const status = normalizeLiveStatus(operation === "start"
        ? await api.startDigitalCenterLive(provider, sessionID)
        : await api.stopDigitalCenterLive(provider, sessionID));
      if (mountedRef.current && requestID === requestsRef.current.live &&
        selectedProviderRef.current === provider) {
        setLiveStatus(status);
      }
      return status;
    } catch (liveError) {
      if (mountedRef.current && requestID === requestsRef.current.live) {
        setError("live", localizedError(liveError));
      }
      throw liveError;
    } finally {
      if (mountedRef.current && requestID === requestsRef.current.live) {
        setLoadingArea("live", false);
      }
    }
  }, [actions.canMonitor, localizedError, readSession?.id, setError, setLoadingArea, workspaceText]);

  const selectItem = useCallback((itemID: string) => {
    requestsRef.current.detail += 1;
    selectedItemIDRef.current = itemID;
    setSelectedItemID(itemID);
    setSelectedItem(null);
    resetWriteDialogState();
  }, [resetWriteDialogState]);

  const closeDetail = useCallback(() => {
    requestsRef.current.detail += 1;
    selectedItemIDRef.current = null;
    setSelectedItemID(null);
    setSelectedItem(null);
    setDialog(null);
    resetWriteDialogState();
  }, [resetWriteDialogState]);

  const previewWrite = useCallback(async (fields: DigitalCenterWriteField[]) => {
    const sessionID = readSession?.id;
    const itemID = selectedItemIDRef.current;
    if (!actions.canWrite || !sessionID || !itemID) {
      const error = new Error(workspaceText("digitalCenters.error.previewUnavailable"));
      setError("write", error.message);
      throw error;
    }
    const requestID = ++requestsRef.current.write;
    setError("write", "");
    setLoadingArea("write", true);
    try {
      const preview = await api.previewDigitalCenterWrite(sessionID, itemID, { fields });
      if (mountedRef.current && requestID === requestsRef.current.write &&
        readSessionIDRef.current === sessionID &&
        selectedItemIDRef.current === itemID) {
        setWritePreview(preview);
        setWriteConfirmation(null);
      }
      return preview;
    } catch (writeError) {
      if (mountedRef.current && requestID === requestsRef.current.write &&
        readSessionIDRef.current === sessionID && selectedItemIDRef.current === itemID) {
        setError("write", localizedError(writeError));
      }
      throw writeError;
    } finally {
      if (mountedRef.current && requestID === requestsRef.current.write &&
        readSessionIDRef.current === sessionID && selectedItemIDRef.current === itemID) {
        setLoadingArea("write", false);
      }
    }
  }, [actions.canWrite, localizedError, readSession?.id, setError, setLoadingArea, workspaceText]);

  const confirmWrite = useCallback(async () => {
    const sessionID = readSession?.id;
    const itemID = selectedItemIDRef.current;
    const preview = writePreview;
    if (!actions.canWrite || !sessionID || !itemID || !preview || preview.itemId !== itemID) {
      const error = new Error(workspaceText("digitalCenters.error.previewRequired"));
      setError("write", error.message);
      throw error;
    }
    if (!validWriteGrant(preview)) {
      const error = new Error(workspaceText("digitalCenters.error.grantExpired"));
      setWritePreview(null);
      setWriteConfirmation(null);
      setError("write", error.message);
      throw error;
    }
    const requestID = ++requestsRef.current.write;
    setError("write", "");
    setLoadingArea("write", true);
    try {
      const confirmation = await api.confirmDigitalCenterWrite(sessionID, itemID, {
        token: preview.token,
        confirm: true,
        fields: preview.fields
      });
      if (mountedRef.current && requestID === requestsRef.current.write &&
        readSessionIDRef.current === sessionID &&
        selectedItemIDRef.current === itemID) {
        setWriteConfirmation(confirmation);
        if (confirmation.workItem) {
          setSelectedItem(confirmation.workItem);
          setWorkItems((current) => ({
            ...current,
            items: current.items.map((item) => item.id === confirmation.workItem?.id
              ? confirmation.workItem
              : item)
          }));
        }
      }
      if (confirmation.liveMonitor.wasRunning &&
        selectedProviderRef.current === confirmation.provider) {
        await loadLiveStatus(confirmation.provider).catch(() => undefined);
      }
      return confirmation;
    } catch (writeError) {
      if (mountedRef.current && requestID === requestsRef.current.write &&
        readSessionIDRef.current === sessionID && selectedItemIDRef.current === itemID) {
        if (writeError instanceof ApiError && writeError.code === "digital_center_address_conflict") {
          setWritePreview(null);
          setWriteConfirmation(null);
          setError("write", digitalCenterAddressConflictMessage(writeError, workspaceText));
        } else if (writeError instanceof ApiError && writeError.status === 409) {
          readSessionIDRef.current = null;
          setReadSession(null);
          clearSessionDependents();
          setError("write", workspaceText("digitalCenters.error.grantConflict"));
        } else {
          setError("write", localizedError(writeError));
        }
      }
      throw writeError;
    } finally {
      if (mountedRef.current && requestID === requestsRef.current.write &&
        readSessionIDRef.current === sessionID && selectedItemIDRef.current === itemID) {
        setLoadingArea("write", false);
      }
    }
  }, [actions.canWrite, clearSessionDependents, loadLiveStatus, localizedError, readSession?.id, setError,
    setLoadingArea, workspaceText, writePreview]);

  const setSearch = useCallback((value: string) => {
    requestsRef.current.worklist += 1;
    setSearchState(value);
  }, []);
  const setCompareStatus = useCallback((value: DigitalCenterCompareFilter) => {
    requestsRef.current.worklist += 1;
    setCompareStatusState(value);
  }, []);
  const setPage = useCallback((value: number) => {
    requestsRef.current.worklist += 1;
    setPageState(Math.max(1, value));
  }, []);
  const setPageSize = useCallback((value: number) => {
    requestsRef.current.worklist += 1;
    setPageSizeState(Math.min(100, Math.max(1, value)));
  }, []);

  return {
    centers,
    selectedProvider,
    selectedCenter,
    selectCenter,
    liveStatus,
    connectionState: liveStatus?.state ?? "stopped",
    readSession,
    readData,
    workItems,
    sessionTotal,
    messages,
    search,
    setSearch,
    compareStatus,
    setCompareStatus,
    page,
    setPage,
    pageSize,
    setPageSize,
    selectedItemId,
    selectedItem,
    selectItem,
    closeDetail,
    tab,
    setTab,
    dialog,
    openDialog: (kind: DigitalCenterWorkspaceDialog["kind"], itemId = selectedItemIDRef.current) => {
      if (itemId) setDialog({ kind, itemId });
    },
    closeDialog: () => setDialog(null),
    writePreview,
    writeConfirmation,
    previewWrite,
    confirmWrite,
    startLive: () => runLiveMutation("start"),
    stopLive: () => runLiveMutation("stop"),
    actions,
    loading,
    errors,
    refresh: loadWorkspace
  };
}

function normalizeLiveStatus(status: ECoSLiveStatus): ECoSLiveStatus {
  if (Array.isArray(status.pulseSamples) && Array.isArray(status.recentEvents)) return status;
  return {
    ...status,
    pulseSamples: status.pulseSamples ?? [],
    recentEvents: status.recentEvents ?? []
  };
}

function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

function digitalCenterErrorTranslationKey(code: string) {
  return code === "digital_center_conflict_unresolved"
    ? "digitalCenters.error.conflictUnresolved"
    : null;
}

function digitalCenterAddressConflictMessage(
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

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function validWriteGrant(preview: DigitalCenterWritePreview) {
  const expiresAt = Date.parse(preview.expiresAt);
  return preview.token.trim().length > 0 && Number.isFinite(expiresAt) && expiresAt > Date.now();
}
