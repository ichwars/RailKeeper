import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { setLanguage } from "../../shared/i18n";
import { createVehicleFilterDefinitions } from "./vehicleInventoryRenderers";
import { translate } from "../../shared/i18n";
import { VehicleInventoryFilters, type VehicleInventoryFilterProps } from "./VehicleInventoryFilters";

function fixture(overrides: Partial<VehicleInventoryFilterProps> = {}): VehicleInventoryFilterProps {
  return {
    inventoryFilter: "all", maintenanceFilter: "all", qualityFilter: "none",
    manufacturerFilter: "", categoryFilter: "", gattungFilter: "", railwayCompanyFilter: "",
    epochFilter: "", adapterFilter: "", exhibitionReadyFilter: false,
    ...createVehicleFilterDefinitions({ vehicleCount: 2, counts: { all: 2, digital: 1, analog: 1,
      withImages: 1, withoutImages: 1, maintenanceDue: 1, withoutMaintenance: 1 },
      t: (key, values) => translate("de", key, values) }),
    inventoryFilterOptions: { manufacturers: ["Roco"], categories: ["Lokomotive"], gattungen: ["Diesellok"],
      railwayCompanies: ["DB"], epochs: ["IV"], adapters: ["PluX22"] },
    resultCount: 2, hasActiveInventoryFilters: false,
    onInventoryFilterChange: vi.fn(), onMaintenanceFilterChange: vi.fn(), onQualityFilterChange: vi.fn(),
    onManufacturerFilterChange: vi.fn(), onCategoryFilterChange: vi.fn(), onGattungFilterChange: vi.fn(),
    onRailwayCompanyFilterChange: vi.fn(), onEpochFilterChange: vi.fn(), onAdapterFilterChange: vi.fn(),
    onExhibitionReadyFilterChange: vi.fn(), onResetFilters: vi.fn(), ...overrides
  };
}

describe("visible inventory filters", () => {
  beforeEach(() => setLanguage("de"));

  it("keeps quick filters visibly named and forwards independent selections", async () => {
    const user = userEvent.setup();
    const props = fixture();
    render(<VehicleInventoryFilters {...props} />);
    const quick = screen.getByRole("group", { name: "Schnellfilter" });
    expect(within(quick).getByRole("button", { name: "Digital" })).toHaveTextContent("Digital");
    await user.click(within(quick).getByRole("button", { name: "Digital" }));
    expect(props.onInventoryFilterChange).toHaveBeenCalledWith("digital");
    expect(props.onMaintenanceFilterChange).not.toHaveBeenCalled();
  });

  it("retains and individually clears active additional filters while collapsed", async () => {
    const user = userEvent.setup();
    const props = fixture({ epochFilter: "IV", railwayCompanyFilter: "DB", hasActiveInventoryFilters: true });
    render(<VehicleInventoryFilters {...props} />);
    expect(screen.queryByRole("group", { name: "Weitere Filter" })).not.toBeInTheDocument();
    const remove = screen.getByRole("button", { name: "Filter entfernen: Epoche: IV" });
    await user.click(remove);
    expect(props.onEpochFilterChange).toHaveBeenCalledExactlyOnceWith("");
    expect(props.onRailwayCompanyFilterChange).not.toHaveBeenCalled();
    expect(props.onResetFilters).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Weitere Filter (2 aktiv)" })).toHaveFocus();
  });

  it("expands and collapses additional controls without resetting their values", async () => {
    const user = userEvent.setup();
    const props = fixture({ adapterFilter: "PluX22", qualityFilter: "missingEan", hasActiveInventoryFilters: true });
    render(<VehicleInventoryFilters {...props} />);
    const more = screen.getByRole("button", { name: "Weitere Filter (2 aktiv)" });
    await user.click(more);
    expect(more).toHaveAttribute("aria-expanded", "true");
    expect(screen.getByRole("button", { name: "Adapter" })).toHaveTextContent("PluX22");
    await user.click(more);
    expect(more).toHaveAttribute("aria-expanded", "false");
    expect(props.onAdapterFilterChange).not.toHaveBeenCalled();
    expect(props.onQualityFilterChange).not.toHaveBeenCalled();
    expect(screen.getByRole("group", { name: "Aktive Filter" })).toHaveTextContent("PluX22");
  });

  it("clears the dependent class when changing category and preserves other filters", async () => {
    const user = userEvent.setup();
    const props = fixture({ categoryFilter: "Lokomotive", gattungFilter: "Diesellok", epochFilter: "IV", hasActiveInventoryFilters: true });
    render(<VehicleInventoryFilters {...props} />);
    await user.click(screen.getByRole("button", { name: "Filter entfernen: Kategorie: Lokomotive" }));
    expect(props.onCategoryFilterChange).toHaveBeenCalledExactlyOnceWith("");
    expect(props.onGattungFilterChange).toHaveBeenCalledExactlyOnceWith("");
    expect(props.onEpochFilterChange).not.toHaveBeenCalled();
  });

  it("selects data checks and exhibition readiness through the existing typed callbacks", async () => {
    const user = userEvent.setup();
    const props = fixture();
    render(<VehicleInventoryFilters {...props} />);
    await user.click(screen.getByRole("button", { name: "Weitere Filter" }));
    await user.click(screen.getByRole("button", { name: "Datenprüfung" }));
    await user.click(screen.getByRole("option", { name: "Ohne EAN" }));
    expect(props.onQualityFilterChange).toHaveBeenCalledWith("missingEan");
    await user.click(screen.getByRole("button", { name: "Messe tauglich" }));
    expect(props.onExhibitionReadyFilterChange).toHaveBeenCalledWith(true);
  });

  it("retains a selected value when it disappears from the current option list", async () => {
    const user = userEvent.setup();
    const props = fixture({ adapterFilter: "Legacy connector", hasActiveInventoryFilters: true,
      inventoryFilterOptions: { manufacturers: [], categories: [], gattungen: [], railwayCompanies: [], epochs: [], adapters: [] } });
    render(<VehicleInventoryFilters {...props} />);
    await user.click(screen.getByRole("button", { name: "Weitere Filter (1 aktiv)" }));
    expect(screen.getByRole("button", { name: "Adapter" })).toHaveTextContent("Legacy connector");
    expect(props.onAdapterFilterChange).not.toHaveBeenCalled();
  });
});
