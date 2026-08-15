import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { vehicleFixture } from "../../test/fixtures/vehicles";
import { VehicleInventoryMobileList } from "./VehicleInventoryMobileList";

describe("VehicleInventoryMobileList", () => {
  it("renders the same selected fields in order without hidden defaults", () => {
    const vehicle = vehicleFixture({
      series: "BR 218",
      digital: true,
      purchaseDate: "2026-08-15"
    });
    render(
      <VehicleInventoryMobileList
        vehicles={[vehicle]}
        columns={["series", "digital", "purchaseDate"]}
        onOpenDetail={vi.fn()}
        onOpenEdit={vi.fn()}
        onDelete={vi.fn()}
        renderQuickMenu={() => null}
      />
    );

    expect(screen.getAllByRole("term").map((term) => term.textContent)).toEqual([
      "Baureihe",
      "Digital",
      "Datum"
    ]);
    expect(screen.getAllByRole("definition").map((value) => value.textContent)).toEqual([
      "BR 218",
      "ja",
      "15.08.2026"
    ]);
    expect(screen.queryByText(vehicle.inventoryNumber)).not.toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Bearbeiten" })).toHaveLength(1);
    expect(screen.getAllByRole("button", { name: "Löschen" })).toHaveLength(1);
  });
});
