import { useEffect, useState, type ReactNode } from "react";

import type { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { VehicleInventoryMobileCard } from "./VehicleInventoryMobileCard";
import { VehicleSetInventoryMobileCard } from "./VehicleSetInventoryMobileCard";
import type { VehicleTableColumn } from "./vehicleTableColumns";
import { groupVehicleInventory } from "./vehicleSetGroups";
import type { SortDirection, SortKey } from "./vehicleViewModel";

type VehicleInventoryMobileListProps = {
  vehicles: Vehicle[];
  columns: readonly VehicleTableColumn[];
  sort: { key: SortKey; direction: SortDirection };
  onOpenDetail: (vehicle: Vehicle) => void;
  onOpenEdit: (vehicle: Vehicle) => void;
  onOpenSet?: (setID: string) => void;
  onEditSet?: (setID: string) => void;
  onDuplicateSet?: (setID: string) => void;
  renderQuickMenu: (vehicle: Vehicle) => ReactNode;
};

export function VehicleInventoryMobileList({
  vehicles,
  columns,
  sort,
  onOpenDetail,
  onOpenEdit,
  onOpenSet = () => undefined,
  onEditSet,
  onDuplicateSet,
  renderQuickMenu
}: VehicleInventoryMobileListProps) {
  const { t } = useI18n();
  const [expandedVehicleIDs, setExpandedVehicleIDs] = useState<Set<string>>(() => new Set());
  const [expandedSetIDs, setExpandedSetIDs] = useState<Set<string>>(() => new Set());
  const groupedVehicles = groupVehicleInventory(vehicles, sort);

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
        <section className="vehicle-mobile-set-tree" key={group.id}>
          <VehicleSetInventoryMobileCard
            group={group}
            expanded={expandedSetIDs.has(group.id)}
            onToggleExpanded={() => setExpandedSetIDs((current) => {
              const next = new Set(current);
              if (next.has(group.id)) next.delete(group.id);
              else next.add(group.id);
              return next;
            })}
            onOpen={onOpenSet}
            onEdit={onEditSet}
            onDuplicate={onDuplicateSet}
          />
          {expandedSetIDs.has(group.id) && (
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
      ))}
    </div>
  );
}
