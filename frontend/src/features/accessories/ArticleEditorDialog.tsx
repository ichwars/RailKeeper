import { useEffect, useRef, useState, type KeyboardEvent } from "react";
import { X } from "lucide-react";

import type {
  AccessoryArticle,
  AccessoryArticleType,
  AccessoryDuplicateCandidate,
  MasterDataEntry
} from "../../shared/api";
import { api } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { ArticleCoreTab } from "./ArticleCoreTab";
import { ArticlePurchaseDocumentsTab } from "./ArticlePurchaseDocumentsTab";
import { ArticleStockTab } from "./ArticleStockTab";
import { ArticleSubjectTab } from "./ArticleSubjectTab";
import { ArticleUsageHistoryTab } from "./ArticleUsageHistoryTab";
import type {
  ArticleEditorFieldErrors,
  ArticleEditorForm,
  ArticleEditorMode,
  ArticleEditorTab,
  ArticleEditorTabErrors
} from "./articleEditorModel";
import type { ArticleEditorPermissions, ArticleEditorResources } from "./useArticleEditorController";
import { AccessoryConfirmDialog } from "./AccessoryConfirmDialog";
import {
  compatibleAttributesForType,
  compatibleNumberDraftsForType,
  type CustomArticleSubjectFieldDefinition
} from "./articleTypeFields";
import { articleTypeLabel } from "./articleTypes";

export type ArticleEditorDialogProps = {
  mode: ArticleEditorMode;
  form: ArticleEditorForm;
  article: AccessoryArticle | null;
  activeTab: ArticleEditorTab;
  hasUsageHistory: boolean;
  saving: boolean;
  loading: boolean;
  error: string;
  resourceError: string;
  fieldErrors: ArticleEditorFieldErrors;
  tabErrors: ArticleEditorTabErrors;
  subjectFieldErrors: Record<string, string>;
  customFields: readonly CustomArticleSubjectFieldDefinition[];
  customFieldsLoading: boolean;
  customFieldsError: string;
  articleTypeEntries: MasterDataEntry[];
  articleTypeEntriesLoading: boolean;
  articleTypeEntriesError: string;
  subtypeEntries: MasterDataEntry[];
  subtypeEntriesLoading: boolean;
  subtypeEntriesError: string;
  duplicateCandidates: AccessoryDuplicateCandidate[];
  closeConfirmationOpen: boolean;
  permissions: ArticleEditorPermissions;
  resources: ArticleEditorResources;
  resourcesStale: boolean;
  returnFocusTo?: HTMLElement | null;
  onChange: (patch: Partial<ArticleEditorForm>) => void;
  onTabChange: (tab: ArticleEditorTab) => void;
  onSubmit: () => void | Promise<void>;
  onRequestClose: () => void;
  onConfirmClose: () => void;
  onCancelClose: () => void;
  onConfirmDuplicate: () => void | Promise<void>;
  onCancelDuplicate: () => void;
  onResourcesChanged: () => Promise<void>;
  onRetryResources: () => Promise<void>;
  onRetryCustomFields: () => Promise<void>;
  onRetryArticleTypeEntries: () => Promise<void>;
  onRetrySubtypeEntries: () => Promise<void>;
  onSubdraftDirty: (scope: string, dirty: boolean) => void;
};

const focusableSelector = [
  "button:not([disabled])",
  "input:not([disabled]):not([type='hidden'])",
  "select:not([disabled])",
  "textarea:not([disabled])",
  "a[href]",
  "details > summary:first-of-type",
  "[contenteditable='true']",
  "[tabindex]:not([tabindex='-1'])"
].join(",");

function focusableElements(container: HTMLElement | null): HTMLElement[] {
  if (!container) return [];
  return Array.from(container.querySelectorAll<HTMLElement>(focusableSelector)).filter((element) => {
    if (element.tabIndex < 0 || element.closest("[hidden], [aria-hidden='true'], [inert]")) return false;
    const style = window.getComputedStyle(element);
    return style.display !== "none" && style.visibility !== "hidden";
  });
}

export function ArticleEditorDialog(props: ArticleEditorDialogProps) {
  const { t } = useI18n();
  const layerRef = useRef<HTMLDivElement | null>(null);
  const viewInitialFocusRef = useRef<HTMLButtonElement | null>(null);
  const tabListRef = useRef<HTMLElement | null>(null);
  const [pendingArticleType, setPendingArticleType] = useState<AccessoryArticleType | null>(null);
  const confirmationPending = props.closeConfirmationOpen || props.duplicateCandidates.length > 0 ||
    pendingArticleType !== null;
  const readOnly = props.mode === "view" || !props.permissions.canEdit || props.saving || confirmationPending;
  const plannerReservationMode = props.mode === "view" && !props.permissions.canEdit && props.permissions.canReserve;
  const title = props.mode === "create"
    ? t("accessories.editor.create")
    : props.mode === "edit" ? t("accessories.editor.edit") : t("accessories.editor.view");
  const configuredArticleTypeLabel = articleTypeLabel(props.form.articleType, props.articleTypeEntries, t);
  const primaryImageDocument = props.resources.documents.find((document) =>
    document.category === "image" && document.isPrimary);
  const displayedArticle = props.article ? {
    ...props.article,
    primaryImageUrl: primaryImageDocument
      ? api.accessoryDocumentDownloadPath(props.article.id, primaryImageDocument.id)
      : props.resources.documentsLoaded ? undefined : props.article.primaryImageUrl
  } : null;
  const createTypeConfigurationUnavailable = props.mode === "create" &&
    (props.articleTypeEntriesLoading || Boolean(props.articleTypeEntriesError));
  const tabs: Array<{ key: ArticleEditorTab; label: string; subject?: boolean }> = [
    { key: "article", label: t("accessories.editor.tabs.article") },
    { key: "stock", label: t("accessories.editor.tabs.stock") },
    { key: "purchaseDocuments", label: t("accessories.editor.tabs.purchaseDocuments") },
    {
      key: "subject",
      label: t("accessories.editor.tabs.subject", { type: configuredArticleTypeLabel }),
      subject: true
    },
    ...(props.hasUsageHistory
      ? [{ key: "usageHistory" as const, label: t("accessories.editor.tabs.usageHistory") }]
      : [])
  ];

  const changeCore = (patch: Partial<ArticleEditorForm>) => {
    const nextType = patch.articleType;
    if (!nextType || nextType === props.form.articleType) {
      props.onChange(patch);
      return;
    }
    const compatibleAttributes = compatibleAttributesForType(nextType, props.form.attributes, props.customFields);
    const compatibleDrafts = compatibleNumberDraftsForType(
      nextType, props.form.attributeNumberDrafts, props.customFields
    );
    const draftCount = Object.values(props.form.attributeNumberDrafts).filter((draft) => draft.trim() !== "").length;
    const compatibleDraftCount = Object.values(compatibleDrafts).filter((draft) => draft.trim() !== "").length;
    const discardsValues = props.form.subtype.trim() !== "" ||
      compatibleAttributes.length !== props.form.attributes.length ||
      compatibleDraftCount !== draftCount;
    if (discardsValues) {
      setPendingArticleType(nextType);
      return;
    }
    props.onChange({ ...patch, subtype: "", attributes: compatibleAttributes,
      attributeNumberDrafts: compatibleDrafts });
  };

  const confirmArticleTypeChange = () => {
    if (!pendingArticleType) return;
    props.onChange({
      articleType: pendingArticleType,
      subtype: "",
      attributes: compatibleAttributesForType(pendingArticleType, props.form.attributes, props.customFields),
      attributeNumberDrafts: compatibleNumberDraftsForType(
        pendingArticleType,
        props.form.attributeNumberDrafts,
        props.customFields
      )
    });
    setPendingArticleType(null);
  };

  useEffect(() => {
    if (props.loading) return;
    const initial = props.mode === "view"
      ? viewInitialFocusRef.current
      : layerRef.current?.querySelector<HTMLElement>("[data-article-initial-focus]");
    initial?.focus();
  }, [props.loading, props.mode]);

  useEffect(() => () => props.returnFocusTo?.focus(), [props.returnFocusTo]);

  useEffect(() => {
    if (!tabs.some((tab) => tab.key === props.activeTab)) props.onTabChange("article");
  }, [props.activeTab, props.hasUsageHistory]);

  useEffect(() => {
    const activeTab = tabListRef.current?.querySelector<HTMLElement>("[role='tab'][aria-selected='true']");
    if (typeof activeTab?.scrollIntoView === "function") {
      activeTab.scrollIntoView({ block: "nearest", inline: "nearest" });
    }
  }, [props.activeTab]);

  const trapFocus = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === "Escape") {
      event.preventDefault();
      props.onRequestClose();
      return;
    }
    if (event.key !== "Tab" || confirmationPending) return;
    const focusable = focusableElements(layerRef.current);
    if (focusable.length === 0) return;
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    const target = event.target as HTMLElement;
    if (props.mode !== "view" && !event.shiftKey && target.hasAttribute("data-article-dialog-close")) {
      event.preventDefault();
      layerRef.current?.querySelector<HTMLElement>("[data-article-initial-focus]")?.focus();
      return;
    }
    if ((!event.shiftKey && target === last) ||
        (event.shiftKey && target === first)) {
      event.preventDefault();
      (event.shiftKey ? last : first).focus();
    }
  };

  const tabLabel = (tab: typeof tabs[number]) => {
    const hasError = Boolean(props.tabErrors[tab.key]);
    return hasError ? `${tab.label}, ${t("accessories.editor.tabError")}` : tab.label;
  };

  return <div ref={layerRef} className="modal-layer article-editor-layer" role="dialog" aria-modal="true"
    aria-label={title} onKeyDown={trapFocus}>
    <section className="vehicle-modal article-editor-dialog">
      <header className="modal-head">
        <div><h2>{title}</h2>{props.article ? <p>{props.article.manufacturer} · {props.article.articleNumber}</p> : null}</div>
        <button ref={viewInitialFocusRef} type="button" className="icon-button" data-article-dialog-close
          onClick={props.onRequestClose}
          aria-label={t("accessories.editor.close")} title={t("accessories.editor.close")}>
          <X size={18} aria-hidden="true" />
        </button>
      </header>
      <nav ref={tabListRef} className="modal-tabs article-editor-tabs" role="tablist"
        aria-label={t("accessories.editor.tabs.label")}>
        {tabs.map((tab) => <button key={tab.key} type="button" role="tab" data-tab-kind={tab.subject ? "subject" : "fixed"}
          aria-selected={props.activeTab === tab.key} aria-label={tabLabel(tab)}
          disabled={props.saving || confirmationPending}
          className={`${props.activeTab === tab.key ? "active" : ""} ${props.tabErrors[tab.key] ? "has-error" : ""}`.trim()}
          onClick={() => props.onTabChange(tab.key)}>
          <span>{tab.label}</span>{props.tabErrors[tab.key] ? <span className="article-tab-error" aria-hidden="true">!</span> : null}
        </button>)}
      </nav>
      <div className="modal-body article-editor-body">
        {props.loading ? <p className="empty-state">{t("accessories.editor.loading")}</p> : null}
        {!props.loading ? <>
          <div hidden={props.activeTab !== "article"} aria-hidden={props.activeTab !== "article"}>
            <ArticleCoreTab form={props.form} article={displayedArticle} errors={props.fieldErrors}
              disabled={readOnly} articleTypeDisabled={props.customFieldsLoading ||
                props.articleTypeEntriesLoading || Boolean(props.articleTypeEntriesError)}
              typeDependentDisabled={createTypeConfigurationUnavailable}
              otherArticleTypeDisabled={Boolean(props.customFieldsError)}
              articleTypeEntries={props.articleTypeEntries}
              subtypeEntries={props.subtypeEntries}
              subtypeEntriesLoading={props.subtypeEntriesLoading}
              subtypeEntriesError={props.subtypeEntriesError}
              onChange={changeCore} />
          </div>
          <div hidden={props.activeTab !== "stock"} aria-hidden={props.activeTab !== "stock"}>
            <ArticleStockTab article={props.article} form={props.form} errors={props.fieldErrors}
              resources={props.resources}
              disabled={readOnly || !props.permissions.canManageStock || props.resourcesStale}
              canReserve={!props.saving && !confirmationPending && props.permissions.canReserve &&
                !props.resourcesStale && (props.mode !== "view" || plannerReservationMode)}
              canInstall={!props.saving && !confirmationPending && !props.resourcesStale &&
                props.permissions.canInstall && props.mode !== "view"}
              onChange={props.onChange} onChanged={props.onResourcesChanged}
              onDirtyChange={props.onSubdraftDirty} />
          </div>
          <div hidden={props.activeTab !== "purchaseDocuments"} aria-hidden={props.activeTab !== "purchaseDocuments"}>
            <ArticlePurchaseDocumentsTab article={props.article} resources={props.resources}
              disabled={readOnly || !props.permissions.canManageStock || props.resourcesStale}
              onChanged={props.onResourcesChanged}
              onDirtyChange={(dirty) => props.onSubdraftDirty("purchaseDocuments", dirty)} />
          </div>
          <div hidden={props.activeTab !== "subject"} aria-hidden={props.activeTab !== "subject"}>
            <ArticleSubjectTab form={props.form} disabled={readOnly || createTypeConfigurationUnavailable}
              error={props.fieldErrors.attributes}
              active={props.activeTab === "subject"} customFields={props.customFields}
              articleTypeEntries={props.articleTypeEntries}
              loading={props.customFieldsLoading} loadError={props.customFieldsError}
              subjectFieldErrors={props.subjectFieldErrors} onChange={props.onChange} />
          </div>
          {props.hasUsageHistory && props.article ? <div hidden={props.activeTab !== "usageHistory"}
            aria-hidden={props.activeTab !== "usageHistory"}>
            <ArticleUsageHistoryTab article={props.article} resources={props.resources} />
          </div> : null}
        </> : null}
      </div>
      <footer className="modal-actions article-editor-actions">
        {props.resourceError ? <div className="article-editor-resource-error">
          <p className="form-message" role="alert">{props.resourceError}</p>
          {props.resourcesStale ? <button type="button" className="secondary-button"
            onClick={() => void props.onRetryResources().catch(() => undefined)}>
            {t("accessories.editor.retryResources")}</button> : null}
        </div> : null}
        {props.customFieldsError ? <div className="article-editor-resource-error">
          <p className="form-message" role="alert">{props.customFieldsError}</p>
          <button type="button" className="secondary-button" disabled={props.customFieldsLoading}
            onClick={() => void props.onRetryCustomFields().catch(() => undefined)}>
            {t("accessories.editor.retryCustomFields")}</button>
        </div> : null}
        {props.articleTypeEntriesError ? <div className="article-editor-resource-error">
          <p className="form-message" role="alert">{props.articleTypeEntriesError}</p>
          <button type="button" className="secondary-button" disabled={props.articleTypeEntriesLoading}
            onClick={() => void props.onRetryArticleTypeEntries().catch(() => undefined)}>
            {t("accessories.editor.articleTypes.retry")}</button>
        </div> : null}
        {props.subtypeEntriesError ? <div className="article-editor-resource-error">
          <p className="form-message" role="alert">{props.subtypeEntriesError}</p>
          <button type="button" className="secondary-button" disabled={props.subtypeEntriesLoading}
            onClick={() => void props.onRetrySubtypeEntries().catch(() => undefined)}>
            {t("accessories.editor.subtypes.retry")}</button>
        </div> : null}
        {props.error && props.error !== props.resourceError
          ? <p className="form-message" role="alert">{props.error}</p> : null}
        <button type="button" className="secondary-button" onClick={props.onRequestClose}>
          {readOnly ? t("common.close") : t("common.cancel")}
        </button>
        {!readOnly ? <button type="button" className="primary-button"
          disabled={props.saving || props.loading || (props.form.articleType === "other" &&
            (props.customFieldsLoading || Boolean(props.customFieldsError))) ||
            ((props.mode === "create" || props.article?.articleType !== props.form.articleType) &&
              (props.articleTypeEntriesLoading || Boolean(props.articleTypeEntriesError)))}
          onClick={() => void props.onSubmit()}>
          {props.saving ? t("common.saving") : props.mode === "create"
            ? t("accessories.editor.createAction") : t("accessories.editor.saveAction")}
        </button> : null}
      </footer>
    </section>
    <AccessoryConfirmDialog action={props.closeConfirmationOpen ? {
      title: t("accessories.editor.dirty.title"),
      body: t("accessories.editor.dirty.body"),
      cancelLabel: t("accessories.editor.dirty.keep"),
      confirmLabel: t("accessories.editor.dirty.discard"),
      dangerous: true,
      run: props.onConfirmClose
    } : null} onClose={props.onCancelClose} />
    <AccessoryConfirmDialog action={props.duplicateCandidates.length > 0 ? {
      title: t("accessories.editor.duplicate.title"),
      body: <><p>{t("accessories.editor.duplicate.body")}</p><ul>{props.duplicateCandidates.map((candidate) =>
        <li key={candidate.id}><strong>{candidate.manufacturer} {candidate.articleNumber}</strong> · {candidate.name}</li>
      )}</ul></>,
      confirmLabel: t("accessories.editor.duplicate.confirm"),
      run: props.onConfirmDuplicate
    } : null} onClose={props.onCancelDuplicate} />
    <AccessoryConfirmDialog action={pendingArticleType ? {
      title: t("accessories.editor.typeChange.title"),
      body: t("accessories.editor.typeChange.body"),
      cancelLabel: t("accessories.editor.typeChange.keep"),
      confirmLabel: t("accessories.editor.typeChange.discard"),
      dangerous: true,
      run: confirmArticleTypeChange
    } : null} onClose={() => setPendingArticleType(null)} />
  </div>;
}
