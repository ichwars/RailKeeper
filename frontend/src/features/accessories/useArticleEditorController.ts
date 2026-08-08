import { useCallback, useMemo, useState } from "react";

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
    setCloseConfirmationOpen(false);
    setError("");
  };

  const loadResources = useCallback(async (articleId: string) => {
    const shared = await Promise.allSettled([api.storageLocations(), api.vehicles(), api.layouts()]);
    const locations = shared[0].status === "fulfilled" ? shared[0].value : [];
    const vehicles = shared[1].status === "fulfilled" ? shared[1].value : [];
    const layouts = shared[2].status === "fulfilled" ? shared[2].value : [];
    const unitsResult = await Promise.allSettled(layouts.map((layout) => api.layoutUnits(layout.id)));
    const units = unitsResult.flatMap((result) => result.status === "fulfilled" ? result.value : []);
    const related = await Promise.allSettled([
      api.accessoryStock(articleId),
      api.accessoryStockMovements(articleId),
      api.accessoryAssets(articleId),
      api.accessoryPurchases(articleId),
      api.accessoryDocuments(articleId),
      api.accessoryReservations(articleId),
      api.accessoryInstallations(articleId),
      api.accessoryUsageHistory(articleId)
    ]);
    const value = <T,>(index: number, fallback: T): T =>
      related[index]?.status === "fulfilled" ? related[index].value as T : fallback;
    const next: ArticleEditorResources = {
      locations,
      vehicles,
      layouts,
      units,
      stock: value(0, null),
      movements: value(1, []),
      assets: value(2, []),
      purchases: value(3, []),
      documents: value(4, []),
      reservations: value(5, []),
      installations: value(6, []),
      usageHistory: value(7, null)
    };
    setResources(next);
    if (next.reservations.length > 0 || next.installations.length > 0 ||
        (next.usageHistory?.events.length || 0) > 0) setHasUsageHistory(true);
  }, []);

  const openCreate = () => {
    const next = emptyArticleEditorForm();
    setMode("create");
    setArticle(null);
    setForm(next);
    setInitialForm(next);
    setResources(emptyResources());
    setHasUsageHistory(false);
    setReturnFocusTo(document.activeElement instanceof HTMLElement ? document.activeElement : null);
    resetTransientState();
    setIsOpen(true);
    void Promise.allSettled([api.storageLocations()]).then(([locations]) => {
      if (locations.status === "fulfilled") setResources((current) => ({ ...current, locations: locations.value }));
    });
  };

  const openArticle = (id: string, nextMode: Exclude<ArticleEditorMode, "create">, usageSignal: boolean) => {
    setMode(nextMode);
    setArticle(null);
    setHasUsageHistory(usageSignal);
    setReturnFocusTo(document.activeElement instanceof HTMLElement ? document.activeElement : null);
    resetTransientState();
    setResources(emptyResources());
    setIsOpen(true);
    setLoading(true);
    void api.accessoryArticle(id).then((loaded) => {
      const next = articleToEditorForm(loaded);
      setArticle(loaded);
      setForm(next);
      setInitialForm(next);
      void loadResources(id);
    }).catch((reason) => setError(errorMessage(reason))).finally(() => setLoading(false));
  };

  const changeForm = (patch: Partial<ArticleEditorForm>) => {
    setForm((current) => ({ ...current, ...patch }));
    setFieldErrors((current) => {
      const next = { ...current };
      Object.keys(patch).forEach((key) => delete next[key as keyof ArticleEditorForm]);
      return next;
    });
  };

  const closeNow = () => {
    setIsOpen(false);
    setCloseConfirmationOpen(false);
    setDuplicateCandidates([]);
  };

  const requestClose = () => {
    if (mode !== "view" && isArticleEditorDirty(form, initialForm)) {
      setCloseConfirmationOpen(true);
      return;
    }
    closeNow();
  };

  const save = async () => {
    if (!permissions.canEdit || mode === "view") return;
    setSaving(true);
    setError("");
    try {
      const input = articleEditorWriteInput(form);
      const saved = mode === "edit" && article
        ? await api.updateAccessoryArticle(article.id, input)
        : await api.createAccessoryArticle(input);
      setArticle(saved);
      setInitialForm(form);
      await onSaved?.();
      closeNow();
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setSaving(false);
    }
  };

  const submit = async () => {
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
      setSaving(true);
      setError("");
      try {
        const result = await api.checkAccessoryArticleDuplicates({
          manufacturer: form.manufacturer.trim(),
          articleNumber: form.articleNumber.trim(),
          excludeId: mode === "edit" ? article?.id : undefined
        });
        if (result.candidates.length > 0) {
          setDuplicateCandidates(result.candidates);
          return;
        }
      } catch (reason) {
        setError(errorMessage(reason));
        return;
      } finally {
        setSaving(false);
      }
    }
    await save();
  };

  const confirmDuplicateSave = async () => {
    setDuplicateCandidates([]);
    await save();
  };

  const refreshResources = async () => {
    if (!article) return;
    await loadResources(article.id);
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
    cancelDuplicateSave: () => setDuplicateCandidates([]),
    refreshResources
  };
}
