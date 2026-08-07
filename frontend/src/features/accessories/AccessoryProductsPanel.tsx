import { useState } from "react";
import type { FormEvent } from "react";
import { PackagePlus, Search } from "lucide-react";

import { api, type AccessoryProduct, type AccessoryProductInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";

const emptyProduct: AccessoryProductInput = {
  manufacturer: "",
  articleNumber: "",
  name: "",
  category: "",
  trackingMode: "quantity",
  description: ""
};

export function AccessoryProductsPanel({ products, selectedID, query, canEdit, onSelect, onQueryChange, onSearch,
  onSaved }: {
  products: AccessoryProduct[];
  selectedID: string;
  query: string;
  canEdit: boolean;
  onSelect: (id: string) => void;
  onQueryChange: (query: string) => void;
  onSearch: (query: string) => Promise<void>;
  onSaved: (product: AccessoryProduct) => Promise<void>;
}) {
  const [form, setForm] = useState<AccessoryProductInput>(emptyProduct);
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [editingID, setEditingID] = useState("");
  const { t } = useI18n();

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    setBusy(true);
    setMessage("");
    try {
      const saved = editingID
        ? await api.updateAccessoryProduct(editingID, form)
        : await api.createAccessoryProduct(form);
      setForm(emptyProduct);
      setEditingID("");
      await onSaved(saved);
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("accessories.error.generic"));
    } finally {
      setBusy(false);
    }
  };

  const startEditing = () => {
    const selected = products.find((product) => product.id === selectedID);
    if (!selected) return;
    setEditingID(selected.id);
    setForm({
      manufacturer: selected.manufacturer,
      articleNumber: selected.articleNumber || "",
      name: selected.name,
      category: selected.category,
      trackingMode: selected.trackingMode,
      description: selected.description || ""
    });
  };

  const startNew = () => {
    setEditingID("");
    setForm(emptyProduct);
    setMessage("");
  };

  return (
    <div className="accessory-work-grid">
      <section className="panel accessory-list-panel">
        <div className="panel-head form-head">
          <div><Search size={17} aria-hidden="true" /><h2>{t("accessories.products.title")}</h2></div>
        </div>
        <form className="accessory-search" onSubmit={(event) => { event.preventDefault(); void onSearch(query); }}>
          <label>
            <span className="sr-only">{t("accessories.search.label")}</span>
            <input aria-label={t("accessories.search.label")} value={query}
              onChange={(event) => onQueryChange(event.target.value)} placeholder={t("accessories.search.placeholder")} />
          </label>
          <button type="submit" className="secondary-button">{t("common.search")}</button>
        </form>
        {products.length === 0 ? <p className="accessory-empty">{t("accessories.products.empty")}</p> : (
          <div className="table-wrap">
            <table className="accessory-table">
              <thead><tr><th>{t("accessories.field.product")}</th><th>{t("accessories.field.category")}</th><th>{t("accessories.field.tracking")}</th></tr></thead>
              <tbody>{products.map((product) => (
                <tr key={product.id} className={selectedID === product.id ? "selected-row" : ""}>
                  <td><button type="button" className="inventory-name-link" onClick={() => onSelect(product.id)}>
                    <strong>{product.name}</strong><small>{product.manufacturer} {product.articleNumber}</small>
                  </button></td>
                  <td>{product.category}</td>
                  <td>{t(`accessories.tracking.${product.trackingMode}`)}</td>
                </tr>
              ))}</tbody>
            </table>
          </div>
        )}
      </section>

      {canEdit ? (
        <section className="panel accessory-form-panel">
          <div className="panel-head"><PackagePlus size={17} aria-hidden="true" />
            <h2>{t(editingID ? "accessories.products.editTitle" : "accessories.products.create")}</h2>
          </div>
          <form className="accessory-form" onSubmit={submit}>
            <label>{t("accessories.field.manufacturer")}<input required value={form.manufacturer}
              onChange={(event) => setForm((current) => ({ ...current, manufacturer: event.target.value }))} /></label>
            <label>{t("accessories.field.articleNumber")}<input value={form.articleNumber}
              onChange={(event) => setForm((current) => ({ ...current, articleNumber: event.target.value }))} /></label>
            <label>{t("accessories.field.name")}<input required value={form.name}
              onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} /></label>
            <label>{t("accessories.field.category")}<input required value={form.category}
              onChange={(event) => setForm((current) => ({ ...current, category: event.target.value }))} /></label>
            <label>{t("accessories.field.tracking")}<AppSelect value={form.trackingMode}
              onChange={(event) => setForm((current) => ({ ...current, trackingMode: event.target.value as AccessoryProductInput["trackingMode"] }))}>
              <option value="quantity">{t("accessories.tracking.quantity")}</option>
              <option value="individual">{t("accessories.tracking.individual")}</option>
            </AppSelect></label>
            <label>{t("accessories.field.description")}<textarea value={form.description}
              onChange={(event) => setForm((current) => ({ ...current, description: event.target.value }))} /></label>
            {message ? <p className="form-message">{message}</p> : null}
            <div className="accessory-form-actions">
              {editingID ? <button type="button" className="secondary-button" onClick={startNew}>
                {t("accessories.products.new")}
              </button> : <button type="button" className="secondary-button" onClick={startEditing}
                disabled={!selectedID}>{t("accessories.products.edit")}</button>}
              <button type="submit" className="primary-button" disabled={busy}>{t("accessories.products.save")}</button>
            </div>
          </form>
        </section>
      ) : null}
    </div>
  );
}
