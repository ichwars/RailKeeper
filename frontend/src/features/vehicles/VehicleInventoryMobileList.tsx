import { useEffect, useState, type ReactNode } from "react";

import type { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { VehicleInventoryMobileCard } from "./VehicleInventoryMobileCard";
import type { VehicleTableColumn } from "./vehicleTableColumns";

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
      {vehicles.map((vehicle) => (
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
  );
}
