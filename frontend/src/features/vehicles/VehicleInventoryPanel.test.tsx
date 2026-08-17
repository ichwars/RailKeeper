import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { vehicleFixture } from "../../test/fixtures/vehicles";
import { VehicleInventoryPanel } from "./VehicleInventoryPanel";

describe("VehicleInventoryPanel card view", () => {
  it("preserves set hierarchy and set actions", async () => {
    const user = userEvent.setup();
    const set = {
      id: "set-1",
      inventoryNumber: "RK-SET-000001",
      name: "Rheingold",
      manufacturer: "Roco",
      articleNumber: "45923",
      gauge: "H0",
      memberCount: 2,
      position: 1
    };
    const vehicles = [
      vehicleFixture({ id: "member-1", vehicleSet: set }),
      vehicleFixture({ id: "member-2", vehicleSet: { ...set, position: 2 } })
    ];
    const onOpenSet = vi.fn();
    const onEditSet = vi.fn();
    const onDuplicateSet = vi.fn();

    render(<VehicleInventoryPanel
      vehicles={vehicles} sortedVehicles={vehicles} loading={false} message="" query=""
      columns={["type", "inventoryNumber", "manufacturer", "articleNumber", "name"]}
      columnsLoading={false} sort={{ key: "inventoryNumber", direction: "asc" }} inventoryView="cards"
      inventoryFilter="all" maintenanceFilter="all" qualityFilter="none"
      manufacturerFilter="" categoryFilter="" gattungFilter="" railwayCompanyFilter=""
      epochFilter="" adapterFilter="" exhibitionReadyFilter={false}
      inventorySummary={{ categories: 1, digital: 0, analog: 2, withImages: 0 }}
      maintenanceReminderSummary={{ due: 0, upcoming: 0 }} nextMaintenanceReminder={null}
      inventoryFilters={[]} maintenanceFilters={[]}
      inventoryFilterOptions={{ manufacturers: [], categories: [], gattungen: [], railwayCompanies: [], epochs: [], adapters: [] }}
      hasActiveInventoryFilters={false} allVisibleSelected={false} selectedVehicleIDs={new Set()}
      onCreate={vi.fn()} onReload={vi.fn()} onOpenReport={vi.fn()} onQueryChange={vi.fn()}
      onToggleColumn={vi.fn()} onMoveColumn={vi.fn()} onResetColumns={vi.fn()} onToggleSort={vi.fn()}
      onInventoryViewChange={vi.fn()} onInventoryFilterChange={vi.fn()} onMaintenanceFilterChange={vi.fn()}
      onQualityFilterChange={vi.fn()} onManufacturerFilterChange={vi.fn()} onCategoryFilterChange={vi.fn()}
      onGattungFilterChange={vi.fn()} onRailwayCompanyFilterChange={vi.fn()} onEpochFilterChange={vi.fn()}
      onAdapterFilterChange={vi.fn()} onExhibitionReadyFilterChange={vi.fn()} onResetFilters={vi.fn()}
      onOpenDetail={vi.fn()} onOpenEdit={vi.fn()} onDelete={vi.fn()} onToggleSelection={vi.fn()}
      onToggleSetSelection={vi.fn()} onToggleAllVisibleSelection={vi.fn()} onToggleExhibition={vi.fn()}
      onOpenSet={onOpenSet} onEditSet={onEditSet} onDuplicateSet={onDuplicateSet} renderQuickMenu={() => null}
    />);

    expect(screen.getAllByText("RK-SET-000001").length).toBeGreaterThan(0);
    await user.click(screen.getByRole("button", { name: /Set anzeigen/ }));
    await user.click(screen.getByRole("button", { name: /Set Bearbeiten/ }));
    await user.click(screen.getByRole("button", { name: /Set Duplizieren/ }));
    expect(onOpenSet).toHaveBeenCalledWith("set-1");
    expect(onEditSet).toHaveBeenCalledWith("set-1");
    expect(onDuplicateSet).toHaveBeenCalledWith("set-1");
  });
});
