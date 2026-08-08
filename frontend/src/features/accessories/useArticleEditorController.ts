import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import {
  api,
  type AccessoryArticle,
  type AccessoryAsset,
  type AccessoryDocument,
  type AccessoryDuplicateCandidate,
  type AccessoryInstallation,
  type AccessoryPurchase,
  type AccessoryReservation,
  type AccessoryStockMovement,
  type AccessoryStockSummary,
  type AccessoryUsageHistory,
  type Layout,
  type LayoutUnit,
  type StorageLocation,
  type Vehicle
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import {
  articleEditorWriteInput,
  articleToEditorForm,
  emptyArticleEditorForm,
  isArticleEditorDirty,
  validateArticleEditorForm,
  type ArticleEditorFieldErrors,
  type ArticleEditorForm,
  type ArticleEditorMode,
  type ArticleEditorTab,
  type ArticleEditorTabErrors
} from "./articleEditorModel";
import { fetchArticleEditorResourcePatch } from "./articleEditorResources";

export type ArticleEditorPermissions = {
  canEdit: boolean;
  canManageStock: boolean;
  canReserve: boolean;
  canInstall: boolean;
};

export type ArticleEditorResources = {
  locations: StorageLocation[];
  stock: AccessoryStockSummary | null;
  movements: AccessoryStockMovement[];
  assets: AccessoryAsset[];
  purchases: AccessoryPurchase[];
  documents: AccessoryDocument[];
  reservations: AccessoryReservation[];
  installations: AccessoryInstallation[];
  usageHistory: AccessoryUsageHistory | null;
  vehicles: Vehicle[];
  layouts: Layout[];
  units: LayoutUnit[];
};

const emptyResources = (): ArticleEditorResources => ({
  locations: [],
  stock: null,
  movements: [],
  assets: [],
  purchases: [],
  documents: [],
  reservations: [],
  installations: [],
  usageHistory: null,
  vehicles: [],
  layouts: [],
  units: []
});

function errorMessage(reason: unknown): string {
  return reason instanceof Error ? reason.message : "Die Artikelaktion konnte nicht abgeschlossen werden.";
}

export function useArticleEditorController({
  roles,
  onSaved
}: {
  roles: string[];
  onSaved?: () => void | Promise<void>;
}) {
  const { t } = useI18n();
  const [mode, setMode] = useState<ArticleEditorMode>("create");
  const [isOpen, setIsOpen] = useState(false);
  const [article, setArticle] = useState<AccessoryArticle | null>(null);
  const [form, setForm] = useState<ArticleEditorForm>(emptyArticleEditorForm);
  const [initialForm, setInitialForm] = useState<ArticleEditorForm>(emptyArticleEditorForm);
  const [activeTab, setActiveTab] = useState<ArticleEditorTab>("article");
  const [fieldErrors, setFieldErrors] = useState<ArticleEditorFieldErrors>({});
  const [tabErrors, setTabErrors] = useState<ArticleEditorTabErrors>({});
  const [duplicateCandidates, setDuplicateCandidates] = useState<AccessoryDuplicateCandidate[]>([]);
  const [closeConfirmationOpen, setCloseConfirmationOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [hasUsageHistory, setHasUsageHistory] = useState(false);
  const [returnFocusTo, setReturnFocusTo] = useState<HTMLElement | null>(null);
  const [resources, setResources] = useState<ArticleEditorResources>(emptyResources);
  const [resourcesStale, setResourcesStale] = useState(false);
  const [detailReady, setDetailReady] = useState(false);
  const [duplicateDraft, setDuplicateDraft] = useState<ArticleEditorForm | null>(null);
  const [subdraftDirty, setSubdraftDirtyState] = useState<Record<string, boolean>>({});
  const [sessionKey, setSessionKey] = useState(0);
  const generationRef = useRef(0);
  const resourceRequestRef = useRef(0);
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      generationRef.current += 1;
      resourceRequestRef.current += 1;
    };
  }, []);

  const permissions = useMemo<ArticleEditorPermissions>(() => {
    const canEdit = roles.includes("Admin") || roles.includes("Editor");
    return {
      canEdit,
      canManageStock: canEdit,
      canReserve: canEdit || roles.includes("Planner"),
      canInstall: canEdit
    };
  }, [roles]);

  const resetTransientState = () => {
    setActiveTab("article");
    setFieldErrors({});
    setTabErrors({});
    setDuplicateCandidates([]);
    setDuplicateDraft(null);
    setSubdraftDirtyState({});
    setCloseConfirmationOpen(false);
    setError("");
  };

  const isCurrent = useCallback((generation: number) =>
    mountedRef.current && generationRef.current === generation, []);

  const loadResources = useCallback(async (articleId: string, generation: number, rejectOnFailure: boolean) => {
    const request = ++resourceRequestRef.current;
    const result = await fetchArticleEditorResourcePatch(articleId);
    if (!isCurrent(generation) || resourceRequestRef.current !== request) return;
    setResources((current) => ({ ...current, ...result.patch }));
    const hasUsage = (result.patch.reservations?.length || 0) > 0 ||
      (result.patch.installations?.length || 0) > 0 || (result.patch.usageHistory?.events.length || 0) > 0;
    if (hasUsage) setHasUsageHistory(true);
    if (result.errors.length > 0) {
      setResourcesStale(true);
      setError(result.errors[0].message);
      if (rejectOnFailure) throw result.errors[0];
      return;
    }
    setResourcesStale(false);
    if (rejectOnFailure) setError("");
  }, [isCurrent]);

  const openCreate = () => {
    const generation = ++generationRef.current;
    resourceRequestRef.current += 1;
    setSessionKey(generation);
    const next = emptyArticleEditorForm();
    setMode("create");
    setArticle(null);
    setForm(next);
    setInitialForm(next);
    setResources(emptyResources());
    setResourcesStale(false);
    setDetailReady(true);
    setLoading(false);
    setHasUsageHistory(false);
    setReturnFocusTo(document.activeElement instanceof HTMLElement ? document.activeElement : null);
    resetTransientState();
    setIsOpen(true);
    void api.storageLocations().then((locations) => {
      if (isCurrent(generation)) setResources((current) => ({ ...current, locations }));
    }).catch((reason) => {
      if (isCurrent(generation)) setError(errorMessage(reason));
    });
  };

  const openArticle = (id: string, nextMode: Exclude<ArticleEditorMode, "create">, usageSignal: boolean) => {
    const generation = ++generationRef.current;
    resourceRequestRef.current += 1;
    setSessionKey(generation);
    const empty = emptyArticleEditorForm();
    setMode(nextMode);
    setArticle(null);
    setForm(empty);
    setInitialForm(empty);
    setDetailReady(false);
    setHasUsageHistory(usageSignal);
    setReturnFocusTo(document.activeElement instanceof HTMLElement ? document.activeElement : null);
    resetTransientState();
    setResources(emptyResources());
    setResourcesStale(false);
    setIsOpen(true);
    setLoading(true);
    void api.accessoryArticle(id).then((loaded) => {
      if (!isCurrent(generation)) return;
      const next = articleToEditorForm(loaded);
      setArticle(loaded);
      setForm(next);
      setInitialForm(next);
      setDetailReady(true);
      void loadResources(id, generation, false);
    }).catch((reason) => {
      if (isCurrent(generation)) setError(errorMessage(reason));
    }).finally(() => {
      if (isCurrent(generation)) setLoading(false);
    });
  };

  const changeForm = (patch: Partial<ArticleEditorForm>) => {
    if (saving || duplicateCandidates.length > 0) return;
    const nextForm = { ...form, ...patch };
    setForm(nextForm);
    setFieldErrors((current) => {
      const next = { ...current };
      Object.keys(patch).forEach((key) => delete next[key as keyof ArticleEditorForm]);
      return next;
    });
    const validation = validateArticleEditorForm(nextForm);
    setTabErrors((current) => {
      const next = { ...current };
      (Object.keys(current) as ArticleEditorTab[]).forEach((tab) => {
        if (!validation.tabErrors[tab]) delete next[tab];
      });
      return next;
    });
  };

  const closeNow = () => {
    generationRef.current += 1;
    resourceRequestRef.current += 1;
    setIsOpen(false);
    setCloseConfirmationOpen(false);
    setDuplicateCandidates([]);
    setDuplicateDraft(null);
    setArticle(null);
    setDetailReady(false);
    setLoading(false);
    setResources(emptyResources());
    setResourcesStale(false);
  };

  const requestClose = () => {
    if (mode !== "view" && (isArticleEditorDirty(form, initialForm) || Object.values(subdraftDirty).some(Boolean))) {
      setCloseConfirmationOpen(true);
      return;
    }
    closeNow();
  };

  const save = async (draft: ArticleEditorForm = form) => {
    if (!permissions.canEdit || mode === "view") return;
    if (mode === "edit" && (!article || !detailReady)) {
      setError(t("accessories.editor.detailRequired"));
      return;
    }
    setSaving(true);
    setError("");
    try {
      const input = articleEditorWriteInput(draft);
      const saved = mode === "edit" && article
        ? await api.updateAccessoryArticle(article.id, input)
        : await api.createAccessoryArticle(input);
      setArticle(saved);
      setInitialForm(draft);
      await onSaved?.();
      closeNow();
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setSaving(false);
    }
  };

  const submit = async () => {
    if (mode === "edit" && (!article || !detailReady)) {
      setError(t("accessories.editor.detailRequired"));
      return;
    }
    const validation = validateArticleEditorForm(form, {
      required: t("accessories.editor.validation.required"),
      positive: t("accessories.editor.validation.positive"),
      nonnegative: t("accessories.editor.validation.nonnegative")
    });
    setFieldErrors(validation.fieldErrors);
    setTabErrors(validation.tabErrors);
    const firstInvalidTab = (["article", "stock", "purchaseDocuments", "subject"] as const)
      .find((tab) => validation.tabErrors[tab]);
    if (firstInvalidTab) {
      setActiveTab(firstInvalidTab);
      return;
    }
    if (form.articleNumber.trim()) {
      const generation = generationRef.current;
      const checkedDraft = structuredClone(form);
      setSaving(true);
      setError("");
      try {
        const result = await api.checkAccessoryArticleDuplicates({
          manufacturer: checkedDraft.manufacturer.trim(),
          articleNumber: checkedDraft.articleNumber.trim(),
          excludeId: mode === "edit" ? article?.id : undefined
        });
        if (!isCurrent(generation)) return;
        if (result.candidates.length > 0) {
          setDuplicateDraft(checkedDraft);
          setDuplicateCandidates(result.candidates);
          return;
        }
      } catch (reason) {
        if (isCurrent(generation)) setError(errorMessage(reason));
        return;
      } finally {
        if (isCurrent(generation)) setSaving(false);
      }
    }
    await save();
  };

  const confirmDuplicateSave = async () => {
    const checkedDraft = duplicateDraft;
    if (!checkedDraft) return;
    setDuplicateCandidates([]);
    setDuplicateDraft(null);
    await save(checkedDraft);
  };

  const refreshResources = async () => {
    if (!article) return;
    await loadResources(article.id, generationRef.current, true);
  };

  const retryResources = async () => {
    await refreshResources();
  };

  const setSubdraftDirty = (scope: string, dirty: boolean) => {
    setSubdraftDirtyState((current) => current[scope] === dirty ? current : { ...current, [scope]: dirty });
  };

  return {
    mode,
    isOpen,
    article,
    form,
    activeTab,
    fieldErrors,
    tabErrors,
    duplicateCandidates,
    closeConfirmationOpen,
    loading,
    saving,
    error,
    hasUsageHistory,
    returnFocusTo,
    resources,
    resourcesStale,
    sessionKey,
    permissions,
    isFormReadOnly: mode === "view" || !permissions.canEdit,
    openCreate,
    openArticle,
    changeForm,
    setActiveTab,
    submit,
    requestClose,
    confirmClose: closeNow,
    cancelClose: () => setCloseConfirmationOpen(false),
    confirmDuplicateSave,
    cancelDuplicateSave: () => { setDuplicateCandidates([]); setDuplicateDraft(null); },
    refreshResources,
    retryResources,
    setSubdraftDirty
  };
}
