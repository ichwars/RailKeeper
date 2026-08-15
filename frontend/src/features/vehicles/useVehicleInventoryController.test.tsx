import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";

import { analogVehicleFixture, maintenanceFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { inventoryViewSettingKey } from "./vehicleViewModel";
import {
  defaultVehicleTableColumns,
  type VehicleTableColumn
} from "./vehicleTableColumns";
import { useVehicleInventoryController } from "./useVehicleInventoryController";

describe("useVehicleInventoryController", () => {
  beforeEach(() => {
    window.history.replaceState(null, "", "/vehicles");
  });

  it("filters inventory by operational and quality criteria", () => {
    const digital = vehicleFixture({ ean: "", digitalDecoderNumber: "" });
    const analog = analogVehicleFixture({ manufacturer: "Roco", category: "Lokomotive", gattung: "Dampflok" });
    const { result } = renderHook(() => useVehicleInventoryController([digital, analog]));

    act(() => result.current.setInventoryFilter("analog"));
    expect(result.current.filteredVehicles.map((vehicle) => vehicle.id)).toEqual([analog.id]);

    act(() => result.current.setInventoryFilter("all"));
    act(() => result.current.setQualityFilter("digitalMissingDecoder"));
    expect(result.current.filteredVehicles.map((vehicle) => vehicle.id)).toEqual([digital.id]);

    act(() => result.current.setQualityFilter("none"));
    act(() => result.current.setManufacturerFilter("Roco"));
    expect(result.current.filteredVehicles.map((vehicle) => vehicle.id)).toEqual([analog.id]);
  });

  it("sorts and selects visible vehicles", () => {
    const later = vehicleFixture({ id: "later", inventoryNumber: "RK-LOK-000010" });
    const earlier = analogVehicleFixture({ id: "earlier", inventoryNumber: "RK-LOK-000002" });
    const { result } = renderHook(() => useVehicleInventoryController([later, earlier]));

    expect(result.current.sortedVehicles.map((vehicle) => vehicle.id)).toEqual([earlier.id, later.id]);
    act(() => result.current.toggleSort("inventoryNumber"));
    expect(result.current.sortedVehicles.map((vehicle) => vehicle.id)).toEqual([later.id, earlier.id]);

    act(() => result.current.toggleAllVisibleSelection());
    expect(result.current.allVisibleSelected).toBe(true);
    expect(result.current.selectedVisibleVehicles).toHaveLength(2);
  });

  it("applies and clears URL gap presets", () => {
    window.history.replaceState(null, "", "/vehicles?gap=no-main-image");
    const withoutImage = analogVehicleFixture();
    const withImage = vehicleFixture();
    const { result } = renderHook(() => useVehicleInventoryController([withImage, withoutImage]));

    expect(result.current.inventoryFilter).toBe("withoutImages");
    expect(result.current.filteredVehicles.map((vehicle) => vehicle.id)).toEqual([withoutImage.id]);

    act(() => result.current.resetInventoryFilters());
    expect(window.location.pathname).toBe("/vehicles");
    expect(window.location.search).toBe("");
  });

  it("tracks due maintenance and persists the view mode", () => {
    const due = vehicleFixture({ maintenance: [maintenanceFixture({ dueDate: "2020-01-01" })] });
    const { result } = renderHook(() => useVehicleInventoryController([due, analogVehicleFixture()]));

    expect(result.current.inventoryFilterCounts.maintenanceDue).toBe(1);
    expect(result.current.maintenanceReminderSummary.due).toBe(1);

    act(() => result.current.setInventoryViewMode("cards"));
    expect(result.current.inventoryView).toBe("cards");
    expect(window.localStorage.getItem(inventoryViewSettingKey)).toBe("cards");
  });

  it("combines railway company, epoch, and adapter filters", () => {
    const db = vehicleFixture({
      id: "db",
      railwayCompany: "DB",
      epoch: "IV",
      adapter: "PluX22"
    });
    const dr = analogVehicleFixture({
      id: "dr",
      railwayCompany: "DR",
      epoch: "III",
      adapter: "NEM 652"
    });
    const { result } = renderHook(() =>
      useVehicleInventoryController([db, dr], defaultVehicleTableColumns)
    );

    act(() => result.current.setRailwayCompanyFilter("DB"));
    act(() => result.current.setEpochFilter("IV"));
    act(() => result.current.setAdapterFilter("PluX22"));

    expect(result.current.filteredVehicles.map((vehicle) => vehicle.id)).toEqual(["db"]);
    expect(result.current.inventoryFilterOptions).toMatchObject({
      railwayCompanies: ["DB", "DR"],
      epochs: ["III", "IV"],
      adapters: ["NEM 652", "PluX22"]
    });

    act(() => result.current.resetInventoryFilters());
    expect(result.current.hasActiveInventoryFilters).toBe(false);
  });

  it("falls back when the active sort column is hidden", async () => {
    const { result, rerender } = renderHook(
      ({ columns }: { columns: VehicleTableColumn[] }) =>
        useVehicleInventoryController(
          [vehicleFixture(), analogVehicleFixture()],
          columns
        ),
      { initialProps: { columns: ["inventoryNumber", "series"] } }
    );

    act(() => result.current.toggleSort("series"));
    rerender({ columns: ["inventoryNumber"] });

    await waitFor(() => {
      expect(result.current.sort).toEqual({ key: "inventoryNumber", direction: "asc" });
    });
  });
});
