import { render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { vehicleFixture } from "../../test/fixtures/vehicles";
import { VehicleInventoryTable } from "./VehicleInventoryTable";

describe("VehicleInventoryTable", () => {
  it("renders selected columns in order with localized values", () => {
    const vehicle = vehicleFixture({
      series: "BR 218",
      digital: true,
      purchaseDate: "2026-08-15"
    });
    render(
      <VehicleInventoryTable
        vehicles={[vehicle]}
        columns={["series", "digital", "purchaseDate"]}
        allVisibleSelected={false}
        selectedVehicleIDs={new Set()}
        sort={{ key: "series", direction: "asc" }}
        onToggleSort={vi.fn()}
        onToggleSelection={vi.fn()}
        onToggleAllVisibleSelection={vi.fn()}
        onOpenDetail={vi.fn()}
        onToggleExhibition={vi.fn()}
        renderQuickMenu={() => null}
      />
    );

    const headers = screen.getAllByRole("columnheader").map((cell) => cell.textContent?.trim());
    expect(headers).toEqual(["", "Baureihe", "Digital", "Datum", "Aktionen"]);
    expect(screen.getByText("BR 218")).toBeInTheDocument();
    expect(screen.getByText("ja")).toBeInTheDocument();
    expect(screen.getByText("15.08.2026")).toBeInTheDocument();
    expect(screen.queryByText(vehicle.inventoryNumber)).not.toBeInTheDocument();
  });

  it("keeps selection and exhibition controls operational", () => {
    const vehicle = vehicleFixture({ exhibition: true });
    render(
      <VehicleInventoryTable
        vehicles={[vehicle]}
        columns={["inventoryNumber", "exhibition"]}
        allVisibleSelected={true}
        selectedVehicleIDs={new Set([vehicle.id])}
        sort={{ key: "inventoryNumber", direction: "asc" }}
        onToggleSort={vi.fn()}
        onToggleSelection={vi.fn()}
        onToggleAllVisibleSelection={vi.fn()}
        onOpenDetail={vi.fn()}
        onToggleExhibition={vi.fn()}
        renderQuickMenu={() => null}
      />
    );

    const row = screen.getAllByRole("row")[1];
    expect(within(row).getAllByRole("checkbox")).toHaveLength(2);
  });
});
