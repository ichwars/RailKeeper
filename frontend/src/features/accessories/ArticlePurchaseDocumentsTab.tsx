import { useEffect, useState, type FormEvent } from "react";
import { Download, FileText, Image as ImageIcon, ShoppingCart, Trash2 } from "lucide-react";

import {
  api,
  type AccessoryArticle,
  type AccessoryDocumentCategory,
  type AccessoryPurchaseInput
} from "../../shared/api";
import { formatDate, useI18n } from "../../shared/i18n";
import { activeStorageLocations, storageLocationPath } from "../../shared/storageLocations";
import { AppDateInput } from "../../shared/ui/AppDateInput";
import { AppFilePicker } from "../../shared/ui/AppFilePicker";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppTextInput } from "../../shared/ui/AppTextInput";
import type { ArticleEditorResources } from "./useArticleEditorController";
import { AccessoryConfirmDialog, type AccessoryPendingAction } from "./AccessoryConfirmDialog";
import { articlePurchaseWriteInput } from "./articleEditorModel";

const today = () => {
  const value = new Date();
  const year = value.getFullYear();
  const month = String(value.getMonth() + 1).padStart(2, "0");
  const day = String(value.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
};
const emptyPurchase = (): AccessoryPurchaseInput => ({ purchasedAt: today(), quantity: 1, currency: "EUR" });
const documentCategories: AccessoryDocumentCategory[] = [
  "invoice", "delivery_note", "manual", "data_sheet", "floor_plan", "image", "other"
];

export function ArticlePurchaseDocumentsTab({ article, resources, disabled, onChanged, onDirtyChange }: {
  article: AccessoryArticle | null;
  resources: ArticleEditorResources;
  disabled: boolean;
  onChanged: () => Promise<void>;
  onDirtyChange: (dirty: boolean) => void;
}) {
  const { t, language } = useI18n();
  const [purchase, setPurchase] = useState<AccessoryPurchaseInput>(emptyPurchase);
  const [purchaseQuantity, setPurchaseQuantity] = useState("1");
  const [file, setFile] = useState<File | null>(null);
  const [category, setCategory] = useState<AccessoryDocumentCategory>("other");
  const [description, setDescription] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [action, setAction] = useState<AccessoryPendingAction | null>(null);
  const locations = activeStorageLocations(resources.locations);
  const locationId = locations.some((location) => location.id === purchase.storageLocationId)
    ? purchase.storageLocationId || "" : locations[0]?.id || "";
  const purchaseDirty = purchaseQuantity !== "1" || JSON.stringify(purchase) !== JSON.stringify(emptyPurchase());
  const documentDirty = Boolean(file || description || category !== "other");
  const hasPrimaryImage = Boolean(article?.primaryImageUrl || resources.documents.some((document) =>
    document.category === "image" && document.isPrimary));

  useEffect(() => onDirtyChange(purchaseDirty || documentDirty), [documentDirty, onDirtyChange, purchaseDirty]);

  const submitPurchase = async (event: FormEvent) => {
    event.preventDefault();
    if (!article) return;
    setBusy(true);
    setError("");
    try {
      await api.createAccessoryPurchase(article.id, articlePurchaseWriteInput(
        purchase, purchaseQuantity, locationId
      ));
      setPurchase(emptyPurchase());
      setPurchaseQuantity("1");
      await onChanged().catch(() => undefined);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("accessories.error.generic"));
    } finally {
      setBusy(false);
    }
  };

  const submitDocument = async (event: FormEvent) => {
    event.preventDefault();
    if (!article || !file) return;
    setBusy(true);
    setError("");
    try {
      await api.uploadAccessoryDocument(article.id, {
        file,
        category,
        description: description.trim() || undefined,
        ...(category === "image" ? { isPrimary: !hasPrimaryImage } : {})
      });
      setFile(null);
      setCategory("other");
      setDescription("");
      await onChanged().catch(() => undefined);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("accessories.error.generic"));
    } finally {
      setBusy(false);
    }
  };

  const removeDocument = (documentId: string, name: string) => {
    if (!article) return;
    setAction({
      title: t("accessories.editor.documents.deleteTitle"),
      body: t("accessories.editor.documents.deleteBody", { name }),
      confirmLabel: t("common.delete"),
      dangerous: true,
      run: () => api.deleteAccessoryDocument(article.id, documentId),
      afterSuccess: onChanged
    });
  };

  const makePrimaryImage = async (documentId: string, documentDescription?: string) => {
    if (!article) return;
    setBusy(true);
    setError("");
    try {
      await api.updateAccessoryDocument(article.id, documentId, {
        category: "image", description: documentDescription, isPrimary: true
      });
      await onChanged().catch(() => undefined);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("accessories.error.generic"));
    } finally {
      setBusy(false);
    }
  };

  if (!article) return <section className="article-editor-tab">
    <p className="article-editor-hint">{t("accessories.editor.saveBeforePurchase")}</p>
  </section>;

  return <section className="article-editor-tab article-purchase-documents-tab"
    aria-label={t("accessories.editor.tabs.purchaseDocuments")}>
    {error ? <p className="form-message" role="alert">{error}</p> : null}
    <section className="article-editor-section">
      <div className="panel-head"><ShoppingCart size={17} aria-hidden="true" /><h3>{t("accessories.editor.purchase.title")}</h3></div>
      <div className="article-editor-split">
        <div className="table-wrap"><table><thead><tr>
          <th>{t("accessories.field.purchaseDate")}</th><th>{t("accessories.editor.purchase.supplier")}</th>
          <th>{t("accessories.field.quantity")}</th><th>{t("accessories.editor.purchase.price")}</th>
        </tr></thead><tbody>{resources.purchases.map((item) => <tr key={item.id}>
          <td>{formatDate(item.purchasedAt, language)}</td><td>{item.supplier || "-"}</td>
          <td>{item.quantity}</td><td>{item.unitPrice ? `${item.unitPrice} ${item.currency || ""}` : "-"}</td>
        </tr>)}</tbody></table></div>
        {!disabled ? <form className="accessory-form" onSubmit={submitPurchase}>
          <h4>{t("accessories.editor.purchase.add")}</h4>
          <label className="app-field"><span className="app-field-label">{t("accessories.field.purchaseDate")}</span>
            <AppDateInput required value={purchase.purchasedAt}
              onChange={(event) => setPurchase((current) => ({ ...current, purchasedAt: event.target.value }))} />
          </label>
          <AppTextInput label={t("accessories.editor.purchase.supplier")} value={purchase.supplier || ""}
            onChange={(event) => setPurchase((current) => ({ ...current, supplier: event.target.value }))} />
          <AppNumberInput label={t("accessories.field.quantity")} required min="1" step="1" value={purchaseQuantity}
            onValueChange={setPurchaseQuantity} />
          <AppTextInput label={t("accessories.editor.purchase.unitPrice")} inputMode="decimal"
            value={purchase.unitPrice || ""}
            onChange={(event) => setPurchase((current) => ({ ...current, unitPrice: event.target.value }))} />
          <p className="article-purchase-total">{t("accessories.editor.purchase.totalPrice")}: <strong>{
            purchase.unitPrice && Number.isFinite(Number(purchase.unitPrice))
              ? (Number(purchase.unitPrice) * Number(purchaseQuantity || 0)).toFixed(2) : "-"
          } {purchase.currency || ""}</strong></p>
          <AppTextInput label={t("accessories.editor.purchase.invoiceNumber")} value={purchase.invoiceNumber || ""}
            onChange={(event) => setPurchase((current) => ({ ...current, invoiceNumber: event.target.value }))} />
          <label className="app-field"><span className="app-field-label">{t("accessories.field.warrantyUntil")}</span>
            <AppDateInput value={purchase.warrantyUntil || ""}
              onChange={(event) => setPurchase((current) => ({ ...current, warrantyUntil: event.target.value }))} />
          </label>
          <label className="article-checkbox"><input type="checkbox" checked={Boolean(purchase.bookToStock)}
            onChange={(event) => setPurchase((current) => ({ ...current, bookToStock: event.target.checked }))} />
            <span>{t("accessories.editor.purchase.bookToStock")}</span>
          </label>
          {purchase.bookToStock ? <label className="app-field">
            <span className="app-field-label">{t("accessories.field.location")}</span>
            <AppSelect value={locationId} aria-label={t("accessories.field.location")}
              onChange={(event) => setPurchase((current) => ({ ...current, storageLocationId: event.target.value }))}>
              {locations.map((location) => <option key={location.id} value={location.id}>
                {storageLocationPath(location, resources.locations)}</option>)}
            </AppSelect>
          </label> : null}
          <AppTextInput label={t("accessories.field.notes")} value={purchase.notes || ""}
            onChange={(event) => setPurchase((current) => ({ ...current, notes: event.target.value }))} />
          <button type="submit" className="primary-button" disabled={busy || Number(purchaseQuantity) <= 0 ||
            Boolean(purchase.bookToStock && !locationId)}>{t("accessories.editor.purchase.book")}</button>
        </form> : null}
      </div>
    </section>

    <section className="article-editor-section">
      <div className="panel-head"><FileText size={17} aria-hidden="true" /><h3>{t("accessories.editor.documents.title")}</h3></div>
      <div className="article-editor-split">
        <div className="article-document-list">{resources.documents.map((document) => <article key={document.id}>
          <div><strong>{document.originalName}</strong><small>{t(`accessories.editor.documentCategory.${document.category}`)}</small></div>
          <div className="article-row-actions">
            {!disabled && document.category === "image" && !document.isPrimary ? <button type="button"
              className="icon-button" disabled={busy}
              aria-label={t("accessories.editor.documents.makePrimaryNamed", { name: document.originalName })}
              title={t("accessories.editor.documents.makePrimaryNamed", { name: document.originalName })}
              onClick={() => void makePrimaryImage(document.id, document.description)}>
              <ImageIcon size={16} aria-hidden="true" />
            </button> : null}
            <a className="icon-button" href={api.accessoryDocumentDownloadPath(article.id, document.id)}
              aria-label={t("accessories.editor.documents.downloadNamed", { name: document.originalName })}
              title={t("accessories.editor.documents.downloadNamed", { name: document.originalName })}>
              <Download size={16} aria-hidden="true" />
            </a>
            {!disabled ? <button type="button" className="icon-button" disabled={busy}
              aria-label={t("accessories.editor.documents.deleteNamed", { name: document.originalName })}
              title={t("accessories.editor.documents.deleteNamed", { name: document.originalName })}
              onClick={() => removeDocument(document.id, document.originalName)}>
              <Trash2 size={16} aria-hidden="true" /></button> : null}
          </div>
        </article>)}</div>
        {!disabled ? <form className="accessory-form" onSubmit={submitDocument}>
          <h4>{t("accessories.editor.documents.add")}</h4>
          <AppFilePicker label={t("accessories.editor.documents.file")} file={file} required
            triggerLabel={t("accessories.editor.documents.choose")} clearLabel={t("accessories.editor.documents.clear")}
            emptyLabel={t("accessories.editor.documents.none")} onFileChange={setFile} />
          <label className="app-field"><span className="app-field-label">{t("accessories.editor.documents.category")}</span>
            <AppSelect value={category} aria-label={t("accessories.editor.documents.category")}
              onChange={(event) => setCategory(event.target.value as AccessoryDocumentCategory)}>
              {documentCategories.map((item) => <option key={item} value={item}>
                {t(`accessories.editor.documentCategory.${item}`)}</option>)}
            </AppSelect>
          </label>
          <AppTextInput label={t("accessories.field.description")} value={description}
            onChange={(event) => setDescription(event.target.value)} />
          <button type="submit" className="primary-button" disabled={!file || busy}>
            {t("accessories.editor.documents.upload")}
          </button>
        </form> : null}
      </div>
    </section>
    <AccessoryConfirmDialog action={action} onClose={() => setAction(null)} />
  </section>;
}
