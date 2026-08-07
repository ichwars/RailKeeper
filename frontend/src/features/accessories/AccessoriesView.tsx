import { useCallback, useEffect, useRef, useState } from "react";
import { PackageOpen, RefreshCw } from "lucide-react";

import {
  api,
  type AccessoryAllocationSummary,
  type AccessoryAsset,
  type AccessoryProduct,
  type AccessoryStockSummary,
  type StorageLocation
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AccessoryProductsPanel } from "./AccessoryProductsPanel";
import { AccessoryLocationsPanel } from "./AccessoryLocationsPanel";
import { AccessoryStockPanel } from "./AccessoryStockPanel";

type AccessoryTab = "products" | "locations" | "stock" | "assets";

export function AccessoriesView({ roles }: { roles: string[] }) {
  const [products, setProducts] = useState<AccessoryProduct[]>([]);
  const [locations, setLocations] = useState<StorageLocation[]>([]);
  const [selectedID, setSelectedID] = useState("");
  const [productQuery, setProductQuery] = useState("");
  const [stock, setStock] = useState<AccessoryStockSummary | null>(null);
  const [assets, setAssets] = useState<AccessoryAsset[]>([]);
  const [summary, setSummary] = useState<AccessoryAllocationSummary | null>(null);
  const [tab, setTab] = useState<AccessoryTab>("products");
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const detailRequest = useRef(0);
  const { t } = useI18n();
  const genericError = t("accessories.error.generic");
  const canEdit = roles.includes("Admin") || roles.includes("Editor");
  const selected = products.find((product) => product.id === selectedID) || null;

  const searchProducts = useCallback(async (query: string) => {
    setMessage("");
    try {
      const next = await api.accessoryProducts(query);
      setProducts(next);
      setSelectedID((current) => next.some((product) => product.id === current) ? current : next[0]?.id || "");
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : genericError);
    }
  }, [genericError]);

  const loadLocations = useCallback(async () => {
    try {
      setLocations(await api.storageLocations());
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : genericError);
    }
  }, [genericError]);

  useEffect(() => {
    let active = true;
    setLoading(true);
    Promise.all([api.accessoryProducts(""), api.storageLocations()])
      .then(([nextProducts, nextLocations]) => {
        if (!active) return;
        setProducts(nextProducts);
        setLocations(nextLocations);
        setSelectedID(nextProducts[0]?.id || "");
      })
      .catch((reason: Error) => active && setMessage(reason.message))
      .finally(() => active && setLoading(false));
    return () => { active = false; };
  }, []);

  const loadSelected = useCallback(async () => {
    const request = ++detailRequest.current;
    if (!selected) {
      setStock(null); setAssets([]); setSummary(null); return;
    }
    setMessage("");
    try {
      const [nextStock, nextAssets, nextSummary] = await Promise.all([
        api.accessoryStock(selected.id),
        selected.trackingMode === "individual" ? api.accessoryAssets(selected.id) : Promise.resolve([]),
        api.accessoryAllocationSummary(selected.id)
      ]);
      if (request !== detailRequest.current) return;
      setStock(nextStock); setAssets(nextAssets); setSummary(nextSummary);
    } catch (reason) {
      if (request !== detailRequest.current) return;
      setMessage(reason instanceof Error ? reason.message : genericError);
    }
  }, [genericError, selected]);

  useEffect(() => {
    void loadSelected();
    return () => { detailRequest.current += 1; };
  }, [loadSelected]);

  if (loading) return <section className="panel"><p>{t("accessories.loading")}</p></section>;

  return <>
    <section className="inventory-head accessory-head">
      <div><p className="eyebrow">{t("accessories.eyebrow")}</p><h1>{t("accessories.title")}</h1><p>{t("accessories.subtitle")}</p></div>
      <button type="button" className="icon-button" onClick={() => {
        setProductQuery(""); void searchProducts(""); void loadSelected();
      }}
        aria-label={t("common.refresh")} title={t("common.refresh")}><RefreshCw size={16} /></button>
    </section>
    {message ? <p className="form-message">{message}</p> : null}
    {summary ? <section className="accessory-summary" aria-label={t("accessories.summary.title")}>
      {(["owned", "stored", "reserved", "installed", "available", "missing"] as const).map((key) =>
        <div key={key}><span>{t(`accessories.summary.${key}`)}</span><strong>{summary[key]}</strong></div>)}
    </section> : null}
    <div className="accessory-tabs" role="tablist" aria-label={t("accessories.tabs.label")}>
      {(["products", "locations", "stock", "assets"] as const).map((item) => <button key={item} type="button"
        role="tab" aria-selected={tab === item} className={tab === item ? "active" : ""} onClick={() => setTab(item)}>
        {item === "products" ? <PackageOpen size={15} /> : null}{t(`accessories.tabs.${item}`)}
      </button>)}
    </div>
    {tab === "products" ? <AccessoryProductsPanel products={products} selectedID={selectedID} query={productQuery}
      canEdit={canEdit} onSelect={setSelectedID} onQueryChange={setProductQuery} onSearch={searchProducts}
      onSaved={async (product) => {
        setProductQuery(""); await searchProducts(""); setSelectedID(product.id);
      }} /> : tab === "locations" ? <AccessoryLocationsPanel locations={locations} canEdit={canEdit}
        onChanged={loadLocations} /> : <AccessoryStockPanel mode={tab} product={selected} stock={stock} assets={assets}
      locations={locations} canEdit={canEdit} onChanged={loadSelected} />}
  </>;
}
