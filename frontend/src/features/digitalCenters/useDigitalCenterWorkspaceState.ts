import { useCallback, useMemo, useRef, useState } from "react";
import { ApiError } from "../../shared/api";
import { translate, useI18n } from "../../shared/i18n";
import {
  type DigitalCenterCompareFilter,
  type DigitalCenterProvider,
  type DigitalCenterReadSession,
  type DigitalCenterSessionMessage,
  type DigitalCenterSummary,
  type DigitalCenterWorkItem,
  type DigitalCenterWorkspaceDialog,
  type DigitalCenterWorkspaceErrors,
  type DigitalCenterWorkspaceTab,
  type DigitalCenterWriteConfirmation,
  type DigitalCenterWritePreview,
  type ECoSLiveStatus,
  emptyDigitalCenterWorkItemPage
} from "./digitalCenterModel";
import {
  digitalCenterAddressConflictMessage, digitalCenterErrorTranslationKey, errorMessage
} from "./digitalCenterWorkspacePolicy";
import type { DigitalCenterWorkspaceOptions } from "./digitalCenterWorkspacePolicy";

type LoadingState = { workspace: boolean; live: boolean; read: boolean; worklist: boolean; detail: boolean; write: boolean; };

export const emptyErrors: DigitalCenterWorkspaceErrors = {
  workspace: "",
  live: "",
  read: "",
  worklist: "",
  detail: "",
  messages: "",
  write: ""
};

type useDigitalCenterWorkspaceStateProps = { options: DigitalCenterWorkspaceOptions; };

export function useDigitalCenterWorkspaceState({ options }: useDigitalCenterWorkspaceStateProps) {
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

  const clampWorklistPageRef = useRef(false);

  const completedReadRevisionRef = useRef(0);

  const requestsRef = useRef({ workspace: 0, live: 0, read: 0, worklist: 0, detail: 0, messages: 0, write: 0 });

  const [centers, setCenters] = useState<DigitalCenterSummary[]>([]);

  const [selectedProvider, setSelectedProvider] = useState<DigitalCenterProvider | null>(null);

  const [liveStatus, setLiveStatus] = useState<ECoSLiveStatus | null>(null);

  const [readSession, setReadSession] = useState<DigitalCenterReadSession | null>(null);

  const [workItems, setWorkItems] = useState(emptyDigitalCenterWorkItemPage);

  const [worklistRevision, setWorklistRevision] = useState(0);

  const [sessionTotal, setSessionTotal] = useState(0);

  const [messages, setMessages] = useState<DigitalCenterSessionMessage[]>([]);

  const [messagesRevision, setMessagesRevision] = useState(0);

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

  const reviewingSnapshot = Boolean(readSession && readSession.provider === selectedProvider &&
    !readSession.capabilities.readLocomotives);

  const actions = useMemo(() => {
    const capabilities = selectedCenter?.capabilities;
    const active = Boolean(selectedCenter?.active) && !reviewingSnapshot;
    return {
      canTestConnection: active && Boolean(capabilities?.testConnection),
      canRead: active && Boolean(capabilities?.readLocomotives),
      canMonitor: active && Boolean(capabilities?.liveMonitor),
      canWrite: active && Boolean(capabilities?.writeLocomotives),
      canWriteCVs: active && Boolean(capabilities?.writeCVs),
      canDiagnose: active && Boolean(capabilities?.diagnose)
    };
  }, [reviewingSnapshot, selectedCenter]);

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

  return {
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
  };
}

export type DigitalCenterWorkspaceState = ReturnType<typeof useDigitalCenterWorkspaceState>;
