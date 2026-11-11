import { useDigitalCenterLive } from "./useDigitalCenterLive";
import { unfilteredSessionTotal } from "./digitalCenterWorkspacePolicy";
import type { DigitalCenterWorkspaceOptions } from "./digitalCenterWorkspacePolicy";
export type { DigitalCenterWorkspaceOptions } from "./digitalCenterWorkspacePolicy";
import { useCallback, useEffect, useMemo } from "react";
import { observeAction } from "../../shared/asyncAction";
import { useDebouncedSearch } from "../../shared/useDebouncedSearch";
import { normalizeLiveStatus } from "./digitalCenterWorkspacePolicy";
import { emptyErrors, useDigitalCenterWorkspaceState } from "./useDigitalCenterWorkspaceState";
import { useDigitalCenterWorkspaceWrites } from "./useDigitalCenterWorkspaceWrites";

import { api } from "../../shared/api";
import {
  type DigitalCenterCompareFilter,
  type DigitalCenterProvider,
  type DigitalCenterWorkItem,
  type DigitalCenterWorkItemFilter,
  type DigitalCenterWorkspaceDialog,
  emptyDigitalCenterWorkItemPage
} from "./digitalCenterModel";

export function useDigitalCentersWorkspace(options: DigitalCenterWorkspaceOptions = {}) {
  const {
  workspaceText, localizedError, pollIntervalMs, mountedRef, selectedProviderRef, readSessionIDRef,
  selectedItemIDRef, clampWorklistPageRef, completedReadRevisionRef, requestsRef, centers,
  setCenters, selectedProvider, setSelectedProvider, liveStatus, setLiveStatus, readSession,
  setReadSession, workItems, setWorkItems, worklistRevision, setWorklistRevision, sessionTotal,
  setSessionTotal, messages, setMessages, messagesRevision, setMessagesRevision, selectedItemId,
  setSelectedItemID, selectedItem, setSelectedItem, search, setSearchState, compareStatus,
  setCompareStatusState, page, setPageState, pageSize, setPageSizeState, tab, setTab, dialog,
  setDialog, writePreview, setWritePreview, writeConfirmation, setWriteConfirmation, errors,
  setErrors, loading, setLoading, selectedCenter, reviewingSnapshot, actions, setError,
  setLoadingArea, resetWriteDialogState, suspendSessionActions, clearSessionDependents,
  } = useDigitalCenterWorkspaceState({ options });

  const loadWorkspace = useCallback(async () => {
    const requestID = ++requestsRef.current.workspace;
    const readRequestID = requestsRef.current.read;
    const readRevision = completedReadRevisionRef.current;
    setError("workspace", "");
    setLoadingArea("workspace", true);
    try {
      const response = await api.digitalCenterWorkspace();
      if (!mountedRef.current || requestID !== requestsRef.current.workspace) return;
      setCenters(current => response.centers.map(center => {
        const newer = completedReadRevisionRef.current !== readRevision
          ? current.find(previous => previous.provider === center.provider &&
            previous.host === center.host && previous.port === center.port)?.lastReadSession
          : undefined;
        return newer ? { ...center, lastReadSession: newer } : center;
      }));
      const retained = selectedProviderRef.current &&
        response.centers.some((center) => center.provider === selectedProviderRef.current)
        ? selectedProviderRef.current
        : null;
      const next = retained ?? response.centers.find((center) => center.selected)?.provider ??
        response.centers.find((center) => center.active)?.provider ?? response.centers[0]?.provider ?? null;
      selectedProviderRef.current = next;
      setSelectedProvider(next);
      const saved = response.centers.find(center => center.provider === next)?.lastReadSession;
      if (saved && !readSessionIDRef.current && requestsRef.current.read === readRequestID) {
        readSessionIDRef.current = saved.id;
        setReadSession(saved);
      }
    } catch (loadError) {
      if (mountedRef.current && requestID === requestsRef.current.workspace) {
        setError("workspace", localizedError(loadError));
      }
    } finally {
      if (mountedRef.current && requestID === requestsRef.current.workspace) {
        setLoadingArea("workspace", false);
      }
    }
  }, [completedReadRevisionRef, localizedError, mountedRef, readSessionIDRef, requestsRef, selectedProviderRef, setCenters, setError, setLoadingArea, setReadSession, setSelectedProvider]);
  const {
    loadLiveStatus,
    runLiveMutation,
  } = useDigitalCenterLive({
    requestsRef,
    setError,
    setLoadingArea,
    mountedRef,
    selectedProviderRef,
    setLiveStatus,
    localizedError,
    workspaceText,
    selectedProvider,
    actions,
    liveStatus,
    pollIntervalMs,
    readSession,
  });

  useEffect(() => {
    const pending = requestsRef.current;
    mountedRef.current = true;
    observeAction(loadWorkspace());
    return () => {
      mountedRef.current = false;
      readSessionIDRef.current = null;
      pending.workspace += 1;
      pending.live += 1;
      pending.read += 1;
      pending.worklist += 1;
      pending.detail += 1;
      pending.messages += 1;
      pending.write += 1;
    };
  }, [loadWorkspace, mountedRef, readSessionIDRef, requestsRef]);

  const searchQuery = useDebouncedSearch(search);
  const searchPending = search !== searchQuery;
  const filter = useMemo<DigitalCenterWorkItemFilter>(() => ({
    query: searchQuery,
    compareStatus,
    page,
    pageSize
  }), [compareStatus, page, pageSize, searchQuery]);

  useEffect(() => {
    const sessionID = readSession?.id;
    if (!sessionID || searchPending) return;
    const clampPage = clampWorklistPageRef.current;
    clampWorklistPageRef.current = false;
    const requestID = ++requestsRef.current.worklist;
    setError("worklist", "");
    setLoadingArea("worklist", true);
    const loadPage = async () => {
      const result = await api.digitalCenterWorkItems(sessionID, filter);
      if (result.sessionTotal !== undefined || (!filter.query.trim() && filter.compareStatus === "all")) {
        return result;
      }
      const aggregate = await api.digitalCenterWorkItems(sessionID, unfilteredSessionTotal);
      return { ...result, sessionTotal: aggregate.total };
    };
    void loadPage()
      .then((result) => {
        if (mountedRef.current && requestID === requestsRef.current.worklist &&
          readSessionIDRef.current === sessionID) {
          const lastPage = Math.max(1, result.totalPages);
          if (clampPage && filter.page > lastPage) {
            setPageState(lastPage);
            setSessionTotal(result.sessionTotal ?? result.total);
            return;
          }
          setWorkItems(result);
          setSessionTotal(result.sessionTotal ?? result.total);
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
  }, [clampWorklistPageRef, filter, localizedError, mountedRef, readSession?.id, readSessionIDRef, requestsRef,
    searchPending, setError, setLoadingArea, setPageState, setSessionTotal, setWorkItems, worklistRevision]);

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
  }, [localizedError, messagesRevision, mountedRef, readSession?.id, readSessionIDRef, requestsRef, setError, setMessages]);

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
  }, [localizedError, mountedRef, readSession?.id, readSessionIDRef, requestsRef, selectedItemIDRef, selectedItemId, setError, setLoadingArea, setSelectedItem]);

  const selectCenter = useCallback((provider: DigitalCenterProvider) => {
    if (selectedProviderRef.current === provider) return;
    selectedProviderRef.current = provider;
    requestsRef.current.live += 1;
    requestsRef.current.read += 1;
    requestsRef.current.worklist += 1;
    requestsRef.current.detail += 1;
    requestsRef.current.messages += 1;
    requestsRef.current.write += 1;
    const saved = centers.find(center => center.provider === provider)?.lastReadSession ?? null;
    readSessionIDRef.current = saved?.id ?? null;
    selectedItemIDRef.current = null;
    setSelectedProvider(provider);
    setLiveStatus(null);
    setReadSession(saved);
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
  }, [centers, readSessionIDRef, requestsRef, selectedItemIDRef, selectedProviderRef, setDialog, setErrors, setLiveStatus, setLoading, setMessages, setReadSession, setSelectedItem, setSelectedItemID, setSelectedProvider, setSessionTotal, setWorkItems, setWriteConfirmation, setWritePreview]);

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
        if (session.state !== "ready") {
          setError("read", workspaceText("digitalCenters.error.readTitle"));
          if (readSessionIDRef.current) {
            const messageRequestID = ++requestsRef.current.messages;
            setError("messages", "");
            try {
              const result = await api.digitalCenterSessionMessages(session.id);
              if (mountedRef.current && requestID === requestsRef.current.read &&
                messageRequestID === requestsRef.current.messages &&
                selectedProviderRef.current === provider) setMessages(result.messages);
            } catch (messageError) {
              if (mountedRef.current && requestID === requestsRef.current.read &&
                messageRequestID === requestsRef.current.messages &&
                selectedProviderRef.current === provider) setError("messages", localizedError(messageError));
            }
            return session;
          }
        }
        clearSessionDependents(true);
        readSessionIDRef.current = session.id;
        setReadSession(session);
        if (session.state === "ready") {
          completedReadRevisionRef.current += 1;
          setCenters(current => current.map(center => center.provider === provider
            ? { ...center, lastReadSession: session } : center));
        }
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
  }, [actions.canRead, clearSessionDependents, completedReadRevisionRef, localizedError, mountedRef, readSessionIDRef, requestsRef, selectedProviderRef, setCenters, setError, setLoadingArea, setMessages, setReadSession, suspendSessionActions, workspaceText]);

  const loadSession = useCallback(async (sessionID: string) => {
    const requestID = ++requestsRef.current.read;
    suspendSessionActions();
    setError("read", "");
    setLoadingArea("read", true);
    try {
      const session = await api.digitalCenterReadSession(sessionID);
      if (!mountedRef.current || requestID !== requestsRef.current.read) return session;
      clearSessionDependents();
      requestsRef.current.live += 1;
      selectedProviderRef.current = session.provider;
      readSessionIDRef.current = session.id;
      setSelectedProvider(session.provider);
      setReadSession(session);
      setLiveStatus(null);
      setPageState(1);
      setSearchState("");
      setCompareStatusState("all");
      return session;
    } catch (loadError) {
      if (mountedRef.current && requestID === requestsRef.current.read) {
        setError("read", localizedError(loadError));
      }
      throw loadError;
    } finally {
      if (mountedRef.current && requestID === requestsRef.current.read) {
        setLoadingArea("read", false);
      }
    }
  }, [clearSessionDependents, localizedError, mountedRef, readSessionIDRef, requestsRef, selectedProviderRef, setCompareStatusState, setError, setLiveStatus, setLoadingArea, setPageState, setReadSession, setSearchState, setSelectedProvider, suspendSessionActions]);

  const selectItem = useCallback((itemID: string) => {
    requestsRef.current.detail += 1;
    selectedItemIDRef.current = itemID;
    setSelectedItemID(itemID);
    setSelectedItem(null);
    resetWriteDialogState();
  }, [requestsRef, resetWriteDialogState, selectedItemIDRef, setSelectedItem, setSelectedItemID]);

  const closeDetail = useCallback(() => {
    requestsRef.current.detail += 1;
    selectedItemIDRef.current = null;
    setSelectedItemID(null);
    setSelectedItem(null);
    setDialog(null);
    resetWriteDialogState();
  }, [requestsRef, resetWriteDialogState, selectedItemIDRef, setDialog, setSelectedItem, setSelectedItemID]);
  const {
  previewWrite, confirmWrite,
  } = useDigitalCenterWorkspaceWrites({
  readSession,
  selectedItemIDRef,
  actions,
  workspaceText,
  setError,
  requestsRef,
  setLoadingArea,
  mountedRef,
  readSessionIDRef,
  setWritePreview,
  setWriteConfirmation,
  localizedError,
  writePreview,
  setMessagesRevision,
  setSelectedItem,
  clampWorklistPageRef,
  setWorklistRevision,
  selectedProviderRef,
  loadLiveStatus,
  setReadSession,
  clearSessionDependents
  });

  const setSearch = useCallback((value: string) => {
    requestsRef.current.worklist += 1;
    setSearchState(value);
  }, [requestsRef, setSearchState]);
  const setCompareStatus = useCallback((value: DigitalCenterCompareFilter) => {
    requestsRef.current.worklist += 1;
    setCompareStatusState(value);
  }, [requestsRef, setCompareStatusState]);
  const setPage = useCallback((value: number) => {
    requestsRef.current.worklist += 1;
    setPageState(Math.max(1, value));
  }, [requestsRef, setPageState]);
  const setPageSize = useCallback((value: number) => {
    requestsRef.current.worklist += 1;
    setPageSizeState(Math.min(100, Math.max(1, value)));
  }, [requestsRef, setPageSizeState]);

  return {
    centers,
    selectedProvider,
    selectedCenter,
    selectCenter,
    liveStatus,
    connectionState: liveStatus?.state ?? "stopped",
    readSession,
    readData,
    loadSession,
    reviewingSnapshot,
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
    acceptCVRead: (item: DigitalCenterWorkItem) => {
      if (!mountedRef.current || item.sessionId !== readSessionIDRef.current || item.id !== selectedItemIDRef.current) return;
      setSelectedItem(item);
      setWorkItems((current) => ({ ...current,
        items: current.items.map((entry) => entry.id === item.id ? item : entry) }));
    },
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
