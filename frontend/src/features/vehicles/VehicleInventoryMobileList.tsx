import { Fragment, useEffect, useState, type ReactNode } from "react";
import { ChevronDown, ChevronUp, Layers3 } from "lucide-react";

import type { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { VehicleInventoryMobileCard } from "./VehicleInventoryMobileCard";
import type { VehicleTableColumn } from "./vehicleTableColumns";
import { groupVehicleInventory } from "./vehicleSetGroups";

type VehicleInventoryMobileListProps = {
  vehicles: Vehicle[];
  columns: readonly VehicleTableColumn[];
  onOpenDetail: (vehicle: Vehicle) => void;
  onOpenEdit: (vehicle: Vehicle) => void;
  renderQuickMenu: (vehicle: Vehicle) => ReactNode;
};

export function VehicleInventoryMobileList({
  vehicles,
  columns,
  onOpenDetail,
  onOpenEdit,
  renderQuickMenu
}: VehicleInventoryMobileListProps) {
  const { t } = useI18n();
  const [expandedVehicleIDs, setExpandedVehicleIDs] = useState<Set<string>>(() => new Set());
  const [collapsedSetIDs, setCollapsedSetIDs] = useState<Set<string>>(() => new Set());
  const groupedVehicles = groupVehicleInventory(vehicles);

  useEffect(() => {
    const visibleIDs = new Set(vehicles.map((vehicle) => vehicle.id));
    setExpandedVehicleIDs((current) => {
      const next = new Set([...current].filter((vehicleID) => visibleIDs.has(vehicleID)));
      return next.size === current.size ? current : next;
    });
  }, [vehicles]);

  const toggleExpanded = (vehicleID: string) => {
    setExpandedVehicleIDs((current) => {
      const next = new Set(current);
      if (next.has(vehicleID)) next.delete(vehicleID);
      else next.add(vehicleID);
      return next;
    });
  };

  return (
    <div className="inventory-mobile-list" aria-label={t("vehicles.mobileList")}>
      {groupedVehicles.map((group) => group.kind === "single" ? (
        <VehicleInventoryMobileCard
          key={group.vehicle.id}
          vehicle={group.vehicle}
          columns={columns}
          expanded={expandedVehicleIDs.has(group.vehicle.id)}
          onToggleExpanded={() => toggleExpanded(group.vehicle.id)}
          onOpenDetail={onOpenDetail}
          onOpenEdit={onOpenEdit}
          renderQuickMenu={renderQuickMenu}
        />
      ) : (
        <Fragment key={group.id}>
          <section className="vehicle-mobile-set">
            <button
              type="button"
              className="vehicle-mobile-set-head"
              onClick={() => setCollapsedSetIDs((current) => {
                const next = new Set(current);
                if (next.has(group.id)) next.delete(group.id);
                else next.add(group.id);
                return next;
              })}
              aria-expanded={!collapsedSetIDs.has(group.id)}
            >
              <Layers3 size={18} />
              <span><strong>{group.name}</strong><small>{t("vehicles.set.memberCount", { count: group.members.length })}</small></span>
              {collapsedSetIDs.has(group.id) ? <ChevronDown size={18} /> : <ChevronUp size={18} />}
            </button>
            {!collapsedSetIDs.has(group.id) && (
              <div className="vehicle-mobile-set-members">
                {group.members.map((vehicle) => (
                  <VehicleInventoryMobileCard
                    key={vehicle.id}
                    vehicle={vehicle}
                    columns={columns}
                    expanded={expandedVehicleIDs.has(vehicle.id)}
                    onToggleExpanded={() => toggleExpanded(vehicle.id)}
                    onOpenDetail={onOpenDetail}
                    onOpenEdit={onOpenEdit}
                    renderQuickMenu={renderQuickMenu}
                  />
                ))}
              </div>
            )}
          </section>
        </Fragment>
      ))}
    </div>
  );
}
