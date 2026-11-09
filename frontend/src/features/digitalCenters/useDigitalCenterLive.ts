import { useCallback, useEffect } from "react";
import { api } from "../../shared/api";
import { normalizeLiveStatus } from "./digitalCenterWorkspacePolicy";
import type { DigitalCenterProvider } from "./digitalCenterModel";
import type { DigitalCenterWorkspaceState } from "./useDigitalCenterWorkspaceState";

type useDigitalCenterLiveProps = {
  requestsRef: DigitalCenterWorkspaceState["requestsRef"];
  setError: DigitalCenterWorkspaceState["setError"];
  setLoadingArea: DigitalCenterWorkspaceState["setLoadingArea"];
  mountedRef: DigitalCenterWorkspaceState["mountedRef"];
  selectedProviderRef: DigitalCenterWorkspaceState["selectedProviderRef"];
  setLiveStatus: DigitalCenterWorkspaceState["setLiveStatus"];
  localizedError: DigitalCenterWorkspaceState["localizedError"];
  workspaceText: DigitalCenterWorkspaceState["workspaceText"];
  selectedProvider: DigitalCenterWorkspaceState["selectedProvider"];
  actions: DigitalCenterWorkspaceState["actions"];
  liveStatus: DigitalCenterWorkspaceState["liveStatus"];
  pollIntervalMs: DigitalCenterWorkspaceState["pollIntervalMs"];
  readSession: DigitalCenterWorkspaceState["readSession"];
};

export function useDigitalCenterLive({
  requestsRef, setError, setLoadingArea, mountedRef, selectedProviderRef, setLiveStatus, localizedError, workspaceText, selectedProvider, actions, liveStatus, pollIntervalMs, readSession
}: useDigitalCenterLiveProps) {
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
  }, [localizedError, mountedRef, requestsRef, selectedProviderRef, setError, setLiveStatus, setLoadingArea, workspaceText]);

  useEffect(() => {
    if (!selectedProvider || !actions.canMonitor) {
      requestsRef.current.live += 1;
      setLiveStatus(null);
      setLoadingArea("live", false);
      return;
    }
    void loadLiveStatus(selectedProvider).catch(() => undefined);
  }, [actions.canMonitor, loadLiveStatus, requestsRef, selectedProvider, setLiveStatus, setLoadingArea]);

  useEffect(() => {
    if (!selectedProvider || !actions.canMonitor || liveStatus?.state !== "running" ||
      liveStatus.provider !== selectedProvider) return;
    const timer = window.setTimeout(() => {
      void loadLiveStatus(selectedProvider).catch(() => undefined);
    }, Math.max(1, pollIntervalMs));
    return () => window.clearTimeout(timer);
  }, [actions.canMonitor, liveStatus, loadLiveStatus, pollIntervalMs, selectedProvider]);

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
  }, [actions.canMonitor, localizedError, mountedRef, readSession?.id, requestsRef, selectedProviderRef, setError, setLiveStatus, setLoadingArea, workspaceText]);

  return {
    loadLiveStatus,
    runLiveMutation,
  };
}
