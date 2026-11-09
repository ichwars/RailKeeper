import { useCallback } from "react";
import { ApiError, api } from "../../shared/api";
import {
  type DigitalCenterProvider,
  type DigitalCenterWriteField,
  type DigitalCenterWriteOperation,
  type ECoSLiveStatus
} from "./digitalCenterModel";
import { digitalCenterAddressConflictMessage, validWriteGrant } from "./digitalCenterWorkspacePolicy";
import type { DigitalCenterWorkspaceState } from "./useDigitalCenterWorkspaceState";



type useDigitalCenterWorkspaceWritesProps = {
  readSession: DigitalCenterWorkspaceState["readSession"];
  selectedItemIDRef: DigitalCenterWorkspaceState["selectedItemIDRef"];
  actions: DigitalCenterWorkspaceState["actions"];
  workspaceText: DigitalCenterWorkspaceState["workspaceText"];
  setError: DigitalCenterWorkspaceState["setError"];
  requestsRef: DigitalCenterWorkspaceState["requestsRef"];
  setLoadingArea: DigitalCenterWorkspaceState["setLoadingArea"];
  mountedRef: DigitalCenterWorkspaceState["mountedRef"];
  readSessionIDRef: DigitalCenterWorkspaceState["readSessionIDRef"];
  setWritePreview: DigitalCenterWorkspaceState["setWritePreview"];
  setWriteConfirmation: DigitalCenterWorkspaceState["setWriteConfirmation"];
  localizedError: DigitalCenterWorkspaceState["localizedError"];
  writePreview: DigitalCenterWorkspaceState["writePreview"];
  setMessagesRevision: DigitalCenterWorkspaceState["setMessagesRevision"];
  setSelectedItem: DigitalCenterWorkspaceState["setSelectedItem"];
  clampWorklistPageRef: DigitalCenterWorkspaceState["clampWorklistPageRef"];
  setWorklistRevision: DigitalCenterWorkspaceState["setWorklistRevision"];
  selectedProviderRef: DigitalCenterWorkspaceState["selectedProviderRef"];
  loadLiveStatus: (provider: DigitalCenterProvider) => Promise<ECoSLiveStatus>;
  setReadSession: DigitalCenterWorkspaceState["setReadSession"];
  clearSessionDependents: DigitalCenterWorkspaceState["clearSessionDependents"];
};

export function useDigitalCenterWorkspaceWrites({
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
}: useDigitalCenterWorkspaceWritesProps) {
  const previewWrite = useCallback(async (
    fields: DigitalCenterWriteField[], operation: DigitalCenterWriteOperation = "update"
  ) => {
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
      const preview = await api.previewDigitalCenterWrite(sessionID, itemID, { fields, operation });
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
  }, [actions.canWrite, localizedError, mountedRef, readSession?.id, readSessionIDRef, requestsRef, selectedItemIDRef, setError, setLoadingArea, setWriteConfirmation, setWritePreview, workspaceText]);

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
        operation: preview.operation,
        token: preview.token,
        confirm: true,
        fields: preview.fields
      });
      if (mountedRef.current && requestID === requestsRef.current.write &&
        readSessionIDRef.current === sessionID &&
        selectedItemIDRef.current === itemID) {
        setWriteConfirmation(confirmation);
        setMessagesRevision((current) => current + 1);
        if (confirmation.workItem) {
          setSelectedItem(confirmation.workItem);
          clampWorklistPageRef.current = true;
          setWorklistRevision((current) => current + 1);
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
        } else if (writeError instanceof ApiError && writeError.code === "ecos_live_pause_failed") {
          setWritePreview(null);
          setWriteConfirmation(null);
          setError("write", localizedError(writeError));
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
  }, [actions.canWrite, clampWorklistPageRef, clearSessionDependents, loadLiveStatus, localizedError, mountedRef, readSession?.id, readSessionIDRef, requestsRef, selectedItemIDRef, selectedProviderRef, setError, setLoadingArea, setMessagesRevision, setReadSession, setSelectedItem, setWorklistRevision, setWriteConfirmation, setWritePreview, workspaceText, writePreview]);

  return {
    previewWrite, confirmWrite,
  };
}
