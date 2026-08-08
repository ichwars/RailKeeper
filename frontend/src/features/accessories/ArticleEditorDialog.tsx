import { useEffect, useRef, type KeyboardEvent } from "react";
import { X } from "lucide-react";

import type { AccessoryArticle, AccessoryDuplicateCandidate } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { ArticleCoreTab } from "./ArticleCoreTab";
import { ArticlePurchaseDocumentsTab } from "./ArticlePurchaseDocumentsTab";
import { ArticleStockTab } from "./ArticleStockTab";
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

export type ArticleEditorDialogProps = {
  mode: ArticleEditorMode;
  form: ArticleEditorForm;
  article: AccessoryArticle | null;
  activeTab: ArticleEditorTab;
  hasUsageHistory: boolean;
  saving: boolean;
  loading: boolean;
  error: string;
  fieldErrors: ArticleEditorFieldErrors;
  tabErrors: ArticleEditorTabErrors;
  duplicateCandidates: AccessoryDuplicateCandidate[];
  closeConfirmationOpen: boolean;
  permissions: ArticleEditorPermissions;
  resources: ArticleEditorResources;
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
  onSubdraftDirty: (scope: string, dirty: boolean) => void;
};

const focusableSelector = [
  "button:not([disabled])",
  "input:not([disabled])",
  "textarea:not([disabled])",
  "a[href]",
  "[tabindex]:not([tabindex='-1'])"
].join(",");

export function ArticleEditorDialog(props: ArticleEditorDialogProps) {
  const { t } = useI18n();
  const layerRef = useRef<HTMLDivElement | null>(null);
  const confirmationPending = props.closeConfirmationOpen || props.duplicateCandidates.length > 0;
  const readOnly = props.mode === "view" || !props.permissions.canEdit || props.saving || confirmationPending;
  const plannerReservationMode = props.mode === "view" && !props.permissions.canEdit && props.permissions.canReserve;
  const title = props.mode === "create"
    ? t("accessories.editor.create")
    : props.mode === "edit" ? t("accessories.editor.edit") : t("accessories.editor.view");
  const tabs: Array<{ key: ArticleEditorTab; label: string; subject?: boolean }> = [
    { key: "article", label: t("accessories.editor.tabs.article") },
    { key: "stock", label: t("accessories.editor.tabs.stock") },
    { key: "purchaseDocuments", label: t("accessories.editor.tabs.purchaseDocuments") },
    {
      key: "subject",
      label: t("accessories.editor.tabs.subject", { type: t(`accessories.articleType.${props.form.articleType}`) }),
      subject: true
    },
    ...(props.hasUsageHistory
      ? [{ key: "usageHistory" as const, label: t("accessories.editor.tabs.usageHistory") }]
      : [])
  ];

  useEffect(() => {
    if (props.loading) return;
    const initial = layerRef.current?.querySelector<HTMLElement>(
      props.mode === "view" ? "[data-article-dialog-close]" : "[data-article-initial-focus]"
    );
    initial?.focus();
  }, [props.loading, props.mode]);

  useEffect(() => () => props.returnFocusTo?.focus(), [props.returnFocusTo]);

  useEffect(() => {
    if (!tabs.some((tab) => tab.key === props.activeTab)) props.onTabChange("article");
  }, [props.activeTab, props.hasUsageHistory]);

  const trapFocus = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === "Escape") {
      event.preventDefault();
      props.onRequestClose();
      return;
    }
    if (event.key !== "Tab" || props.closeConfirmationOpen || props.duplicateCandidates.length > 0) return;
    const focusable = Array.from(layerRef.current?.querySelectorAll<HTMLElement>(focusableSelector) || [])
      .filter((element) => !element.closest("[aria-hidden='true']"));
    if (focusable.length === 0) return;
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    const target = event.target as HTMLElement;
    if (!event.shiftKey && target.hasAttribute("data-article-dialog-close")) {
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
        <button type="button" className="icon-button" data-article-dialog-close onClick={props.onRequestClose}
          aria-label={t("accessories.editor.close")} title={t("accessories.editor.close")}>
          <X size={18} aria-hidden="true" />
        </button>
      </header>
      <nav className="modal-tabs article-editor-tabs" role="tablist" aria-label={t("accessories.editor.tabs.label")}>
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
            <ArticleCoreTab form={props.form} article={props.article} errors={props.fieldErrors}
              disabled={readOnly} onChange={props.onChange} />
          </div>
          <div hidden={props.activeTab !== "stock"} aria-hidden={props.activeTab !== "stock"}>
            <ArticleStockTab article={props.article} form={props.form} errors={props.fieldErrors}
              resources={props.resources} disabled={readOnly || !props.permissions.canManageStock}
              canReserve={!props.saving && !confirmationPending && props.permissions.canReserve &&
                (props.mode !== "view" || plannerReservationMode)}
              canInstall={!props.saving && !confirmationPending && props.permissions.canInstall && props.mode !== "view"}
              onChange={props.onChange} onChanged={props.onResourcesChanged}
              onDirtyChange={props.onSubdraftDirty} />
          </div>
          <div hidden={props.activeTab !== "purchaseDocuments"} aria-hidden={props.activeTab !== "purchaseDocuments"}>
            <ArticlePurchaseDocumentsTab article={props.article} resources={props.resources}
              disabled={readOnly || !props.permissions.canManageStock} onChanged={props.onResourcesChanged}
              onDirtyChange={(dirty) => props.onSubdraftDirty("purchaseDocuments", dirty)} />
          </div>
          <div hidden={props.activeTab !== "subject"} aria-hidden={props.activeTab !== "subject"}>
            <section className="article-editor-tab article-subject-seam" data-testid="article-subject-tab"
              aria-label={t("accessories.editor.tabs.subject", {
                type: t(`accessories.articleType.${props.form.articleType}`)
              })}>
              <p className="article-editor-hint">{t("accessories.editor.subjectTask12")}</p>
            </section>
          </div>
          {props.hasUsageHistory && props.article ? <div hidden={props.activeTab !== "usageHistory"}
            aria-hidden={props.activeTab !== "usageHistory"}>
            <ArticleUsageHistoryTab article={props.article} resources={props.resources} />
          </div> : null}
        </> : null}
      </div>
      <footer className="modal-actions article-editor-actions">
        {props.error ? <p className="form-message" role="alert">{props.error}</p> : null}
        <button type="button" className="secondary-button" onClick={props.onRequestClose}>
          {readOnly ? t("common.close") : t("common.cancel")}
        </button>
        {!readOnly ? <button type="button" className="primary-button" disabled={props.saving || props.loading}
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
  </div>;
}
