import { useCallback, useEffect, useRef, useState } from "react";
import { Pencil, Plus, Wrench } from "lucide-react";

import {
  ApiError,
  api,
  type AccessoryArticleListItem,
  type LayoutTechnicalPosition,
  type LayoutTechnicalPositionInput,
  type LayoutUnit
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { LayoutTechnicalPositionDialog } from "./LayoutTechnicalPositionDialog";

export function LayoutTechnicalPositionsPanel({ units, canPlan }: {
  units: LayoutUnit[];
  canPlan: boolean;
}) {
  const activeUnits = units.filter((unit) => !unit.archived);
  const [unitID, setUnitID] = useState(activeUnits[0]?.id || "");
  const [positions, setPositions] = useState<LayoutTechnicalPosition[]>([]);
  const [products, setProducts] = useState<AccessoryArticleListItem[]>([]);
  const [dialogPosition, setDialogPosition] = useState<LayoutTechnicalPosition | null | undefined>(undefined);
  const [returnFocusTo, setReturnFocusTo] = useState<HTMLElement | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [dialogMessage, setDialogMessage] = useState("");
  const [conflict, setConflict] = useState(false);
  const createButtonRef = useRef<HTMLButtonElement | null>(null);
  const { t } = useI18n();
  const selectedUnit = activeUnits.find((unit) => unit.id === unitID);

  useEffect(() => {
    if (activeUnits.some((unit) => unit.id === unitID)) return;
    setUnitID(activeUnits[0]?.id || "");
  }, [activeUnits, unitID]);

  const loadPositions = useCallback(async () => {
    if (!unitID) {
      setPositions([]);
      return [];
    }
    const next = await api.layoutTechnicalPositions(unitID);
    setPositions(next);
    return next;
  }, [unitID]);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setMessage("");
    loadPositions()
      .catch((reason: Error) => active && setMessage(reason.message))
      .finally(() => active && setLoading(false));
    return () => { active = false; };
  }, [loadPositions]);

  useEffect(() => {
    let active = true;
    api.accessoryArticles({ sort: "article", direction: "asc" })
      .then((result) => active && setProducts(result.items))
      .catch((reason: Error) => active && setMessage(reason.message));
    return () => { active = false; };
  }, []);

  const openCreate = () => {
    setReturnFocusTo(createButtonRef.current);
    setDialogMessage("");
    setConflict(false);
    setDialogPosition(null);
  };

  const openEdit = (position: LayoutTechnicalPosition, trigger: HTMLElement) => {
    setReturnFocusTo(trigger);
    setDialogMessage("");
    setConflict(false);
    setDialogPosition(position);
  };

  const save = async (input: LayoutTechnicalPositionInput, expectedVersion?: number) => {
    if (!selectedUnit) return;
    setSaving(true);
    setDialogMessage("");
    setConflict(false);
    try {
      if (dialogPosition?.id && expectedVersion) {
        await api.updateLayoutTechnicalPosition(dialogPosition.id, { ...input, expectedVersion });
      } else {
        await api.createLayoutTechnicalPosition(selectedUnit.id, input);
      }
      await loadPositions();
      setDialogPosition(undefined);
    } catch (reason) {
      if (reason instanceof ApiError && reason.status === 409) {
        setConflict(true);
        setDialogMessage(t("layouts.technology.conflict"));
      } else {
        setDialogMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
      }
    } finally {
      setSaving(false);
    }
  };

  const reloadConflict = async () => {
    if (!dialogPosition?.id) return;
    try {
      const next = await loadPositions();
      const current = next.find((position) => position.id === dialogPosition.id);
      if (!current) {
        setDialogMessage(t("layouts.technology.missingAfterReload"));
        return;
      }
      setConflict(false);
      setDialogMessage(t("layouts.technology.reloaded"));
      return current.version;
    } catch (reason) {
      setDialogMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    }
  };

  const productName = (productID?: string) => {
    const product = products.find((item) => item.id === productID);
    if (!product) return productID ? t("layouts.technology.unknownArticle") : "-";
    return [product.inventoryNumber, product.manufacturer, product.articleNumber, product.name]
      .filter(Boolean).join(" · ");
  };

  return <section className="panel layout-technology-panel">
    <div className="layout-panel-head">
      <div className="panel-title"><Wrench size={17} /><h3>{t("layouts.technology.title")}</h3></div>
      {canPlan && selectedUnit ? <button ref={createButtonRef} type="button"
        className="primary-button compact-action" onClick={openCreate}>
        <Plus size={16} />{t("layouts.technology.create")}
      </button> : null}
    </div>
    {activeUnits.length === 0 ? <p className="layout-empty">{t("layouts.technology.noUnits")}</p> : <>
      <div className="layout-technology-toolbar">
        <label className="app-field"><span className="app-field-label">{t("layouts.plans.unit")}</span>
          <AppSelect value={unitID} aria-label={t("layouts.plans.unit")}
            onChange={(event) => setUnitID(event.target.value)}>
            {activeUnits.map((unit) => <option key={unit.id} value={unit.id}>{unit.name}</option>)}
          </AppSelect>
        </label>
        <p>{t("layouts.technology.summary", { count: positions.length })}</p>
      </div>
      {message ? <p className="form-message" role="alert">{message}</p> : null}
      {loading ? <p>{t("layouts.technology.loading")}</p> : positions.length === 0
        ? <p className="layout-empty">{t("layouts.technology.empty")}</p>
        : <div className="table-wrap"><table className="layout-table layout-technology-table">
          <thead><tr><th>{t("layouts.field.name")}</th><th>{t("layouts.technology.kind")}</th>
            <th>{t("layouts.technology.coordinates")}</th><th>{t("layouts.technology.rotation")}</th>
            <th>{t("layouts.technology.article")}</th><th>{t("layouts.field.status")}</th>
            {canPlan ? <th><span className="sr-only">{t("layouts.plans.actions")}</span></th> : null}</tr></thead>
          <tbody>{positions.map((position) => <tr key={position.id}>
            <td><strong>{position.label}</strong>{position.description ? <small>{position.description}</small> : null}</td>
            <td>{t(`layouts.positionKind.${position.kind}`)}</td>
            <td>{position.positionXMm} / {position.positionYMm} mm</td>
            <td>{position.rotationDegrees}°</td><td>{productName(position.productId)}</td>
            <td><span className={position.archived ? "status-pill archived" : "status-pill"}>
              {t(position.archived ? "layouts.status.archived" : "layouts.status.active")}</span></td>
            {canPlan ? <td><button type="button" className="icon-button"
              aria-label={t("layouts.technology.editLabel", { name: position.label })}
              title={t("layouts.technology.editLabel", { name: position.label })}
              onClick={(event) => openEdit(position, event.currentTarget)}><Pencil size={15} /></button></td> : null}
          </tr>)}</tbody>
        </table></div>}
    </>}
    {dialogPosition !== undefined && selectedUnit ? <LayoutTechnicalPositionDialog unit={selectedUnit}
      position={dialogPosition || undefined} products={products} saving={saving} message={dialogMessage}
      conflict={conflict} returnFocusTo={returnFocusTo} onSubmit={save} onReloadConflict={reloadConflict}
      onClose={() => { if (!saving) setDialogPosition(undefined); }} /> : null}
  </section>;
}
