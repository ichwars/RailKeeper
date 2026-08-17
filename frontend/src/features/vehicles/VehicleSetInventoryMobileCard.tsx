import { ChevronDown, ChevronUp, Copy, Eye, Layers3, Pencil } from "lucide-react";

import { useI18n } from "../../shared/i18n";
import type { VehicleInventorySetGroup } from "./vehicleSetGroups";

type VehicleSetInventoryMobileCardProps = {
  group: VehicleInventorySetGroup;
  expanded: boolean;
  onToggleExpanded: () => void;
  onOpen: (setID: string) => void;
  onEdit?: (setID: string) => void;
  onDuplicate?: (setID: string) => void;
};

export function VehicleSetInventoryMobileCard({
  group,
  expanded,
  onToggleExpanded,
  onOpen,
  onEdit,
  onDuplicate
}: VehicleSetInventoryMobileCardProps) {
  const { t } = useI18n();
  const acquisitionSummary = [
    group.set.acquisitionType,
    group.set.purchaseDate,
    group.set.purchasePrice,
    group.set.condition
  ].filter(Boolean).join(" · ");

  return (
    <article className="vehicle-mobile-set-card">
      <button
        type="button"
        className="vehicle-mobile-set-toggle"
        onClick={onToggleExpanded}
        aria-expanded={expanded}
        aria-label={expanded ? t("vehicles.set.collapse") : t("vehicles.set.expand")}
      >
        <span className="vehicle-type-badge set"><Layers3 size={15} />{t("vehicles.set.type")}</span>
        <span className="vehicle-mobile-set-identity">
          <strong>{group.set.inventoryNumber}</strong>
          <span>{group.set.name}</span>
          <small>{[group.set.manufacturer, group.set.articleNumber].filter(Boolean).join(" · ")}</small>
        </span>
        {expanded ? <ChevronUp size={18} /> : <ChevronDown size={18} />}
      </button>

      <div className="vehicle-mobile-set-meta">
        <span>{t("vehicles.set.visibleOfTotal", {
          visible: group.visibleMemberCount,
          total: group.totalMemberCount
        })}</span>
        {acquisitionSummary && <small>{acquisitionSummary}</small>}
      </div>

      <div className="vehicle-mobile-set-actions">
        <button type="button" className="icon-button" onClick={() => onOpen(group.id)}
          aria-label={t("vehicles.view")}><Eye size={17} /></button>
        {onEdit && <button type="button" className="icon-button" onClick={() => onEdit(group.id)}
          aria-label={t("common.edit")}><Pencil size={17} /></button>}
        {onDuplicate && <button type="button" className="icon-button" onClick={() => onDuplicate(group.id)}
          aria-label={t("common.duplicate")}><Copy size={17} /></button>}
      </div>
    </article>
  );
}
