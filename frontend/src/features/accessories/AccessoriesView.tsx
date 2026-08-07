import { useCallback, useEffect, useRef, useState } from "react";
import { PackageOpen, RefreshCw } from "lucide-react";

import {
  api,
  type AccessoryAllocationSummary,
  type AccessoryAsset,
  type AccessoryInstallation,
  type AccessoryProduct,
  type AccessoryReservation,
  type AccessoryStockSummary,
  type Layout,
  type LayoutUnit,
  type StorageLocation,
  type Vehicle
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AccessoryInstallationsPanel } from "./AccessoryInstallationsPanel";
import { AccessoryLocationsPanel } from "./AccessoryLocationsPanel";
import { AccessoryProductsPanel } from "./AccessoryProductsPanel";
import { AccessoryReservationsPanel } from "./AccessoryReservationsPanel";
import { AccessoryStockPanel } from "./AccessoryStockPanel";

type AccessoryTab = "products" | "locations" | "stock" | "assets" | "reservations" | "installations";

export function AccessoriesView({ roles }: { roles: string[] }) {
  const [products, setProducts] = useState<AccessoryProduct[]>([]);
  const [locations, setLocations] = useState<StorageLocation[]>([]);
  const [selectedID, setSelectedID] = useState("");
  const [productQuery, setProductQuery] = useState("");
  const [stock, setStock] = useState<AccessoryStockSummary | null>(null);
  const [assets, setAssets] = useState<AccessoryAsset[]>([]);
  const [summary, setSummary] = useState<AccessoryAllocationSummary | null>(null);
  const [reservations, setReservations] = useState<AccessoryReservation[]>([]);
  const [installations, setInstallations] = useState<AccessoryInstallation[]>([]);
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [layouts, setLayouts] = useState<Layout[]>([]);
  const [units, setUnits] = useState<LayoutUnit[]>([]);
  const [tab, setTab] = useState<AccessoryTab>("products");
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const detailRequest = useRef(0);
  const { t } = useI18n();
  const genericError = t("accessories.error.generic");
  const canEdit = roles.includes("Admin") || roles.includes("Editor");
  const canReserve = canEdit || roles.includes("Planner");
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
    const layoutTargets = api.layouts().then(async (nextLayouts) => ({
      layouts: nextLayouts,
      units: (await Promise.all(nextLayouts.map((layout) => api.layoutUnits(layout.id)))).flat()
    }));
    Promise.all([api.accessoryProducts(""), api.storageLocations(), api.vehicles(), layoutTargets])
      .then(([nextProducts, nextLocations, nextVehicles, nextTargets]) => {
        if (!active) return;
        setProducts(nextProducts);
        setLocations(nextLocations);
        setVehicles(nextVehicles);
        setLayouts(nextTargets.layouts);
        setUnits(nextTargets.units);
        setSelectedID(nextProducts[0]?.id || "");
      })
      .catch((reason: Error) => active && setMessage(reason.message))
      .finally(() => active && setLoading(false));
    return () => { active = false; };
  }, []);

  const loadSelected = useCallback(async () => {
    const request = ++detailRequest.current;
    if (!selected) {
      setStock(null); setAssets([]); setSummary(null); setReservations([]); setInstallations([]); return;
    }
    setMessage("");
    try {
      const [nextStock, nextAssets, nextSummary, nextReservations, nextInstallations] = await Promise.all([
        api.accessoryStock(selected.id),
        selected.trackingMode === "individual" ? api.accessoryAssets(selected.id) : Promise.resolve([]),
        api.accessoryAllocationSummary(selected.id),
        api.accessoryReservations(selected.id),
        api.accessoryInstallations(selected.id)
      ]);
      if (request !== detailRequest.current) return;
      setStock(nextStock); setAssets(nextAssets); setSummary(nextSummary);
      setReservations(nextReservations); setInstallations(nextInstallations);
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
      {(["products", "locations", "stock", "assets", "reservations", "installations"] as const).map((item) =>
        <button key={item} type="button"
        role="tab" aria-selected={tab === item} className={tab === item ? "active" : ""} onClick={() => setTab(item)}>
        {item === "products" ? <PackageOpen size={15} /> : null}{t(`accessories.tabs.${item}`)}
      </button>)}
    </div>
    {tab === "products" ? <AccessoryProductsPanel products={products} selectedID={selectedID} query={productQuery}
      canEdit={canEdit} onSelect={setSelectedID} onQueryChange={setProductQuery} onSearch={searchProducts}
      onSaved={async (product) => {
        setProductQuery(""); await searchProducts(""); setSelectedID(product.id);
      }} /> : tab === "locations" ? <AccessoryLocationsPanel locations={locations} canEdit={canEdit}
        onChanged={loadLocations} /> : tab === "stock" || tab === "assets"
        ? <AccessoryStockPanel key={selected?.id || "none"} mode={tab} product={selected}
        stock={stock} assets={assets} locations={locations} canEdit={canEdit} onChanged={loadSelected} />
        : tab === "reservations" ? <AccessoryReservationsPanel key={selected?.id || "none"} product={selected}
          reservations={reservations}
          assets={assets} locations={locations} vehicles={vehicles} layouts={layouts} units={units}
          canReserve={canReserve} onChanged={loadSelected} />
          : <AccessoryInstallationsPanel key={selected?.id || "none"} product={selected} reservations={reservations}
            installations={installations} assets={assets} locations={locations} vehicles={vehicles}
            layouts={layouts} units={units} canInstall={canEdit} onChanged={loadSelected} />}
  </>;
}
