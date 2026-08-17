import { render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { vehicleFixture } from "../../test/fixtures/vehicles";
import { VehicleInventoryTable } from "./VehicleInventoryTable";

describe("VehicleInventoryTable", () => {
	it("renders a set as an aligned inventory row followed by child vehicles", () => {
		const set = {
			id: "set-1", inventoryNumber: "RK-SET-000001", name: "Rheingold", manufacturer: "Roco",
			articleNumber: "45923", gauge: "H0", memberCount: 2, position: 1
		};
		render(
			<VehicleInventoryTable
				vehicles={[
					vehicleFixture({
						id: "member-1", articleNumber: "45923", vehicleNumber: "50 80 11-35 001-2",
						railwayCompany: "DB", category: "Wagen", gattung: "Reisezugwagen", images: [], vehicleSet: set
					}),
					vehicleFixture({
						id: "member-2", articleNumber: "45923", vehicleNumber: "50 80 11-35 002-0",
						images: [], vehicleSet: { ...set, position: 2 }
					})
				]}
				columns={["type", "inventoryNumber", "manufacturer", "articleNumber", "name", "railwayCompany", "category"]}
				allVisibleSelected={false}
				selectedVehicleIDs={new Set()}
				sort={{ key: "inventoryNumber", direction: "asc" }}
				onToggleSort={vi.fn()}
				onToggleSelection={vi.fn()}
				onToggleSetSelection={vi.fn()}
				onToggleAllVisibleSelection={vi.fn()}
				onOpenDetail={vi.fn()}
				onOpenSet={vi.fn()}
				onToggleExhibition={vi.fn()}
				renderQuickMenu={() => null}
			/>
		);

		expect(screen.getByText("RK-SET-000001")).toBeInTheDocument();
		expect(document.querySelector(".vehicle-set-inventory-row td[colspan]")).toBeNull();
		expect(document.querySelectorAll(".vehicle-set-child-row")).toHaveLength(2);
		expect(document.querySelectorAll(".vehicle-set-child-row-last")).toHaveLength(1);
		expect(document.querySelectorAll(
			".vehicle-member-type-cell .inventory-thumb, .vehicle-member-type-cell .image-placeholder"
		)).toHaveLength(2);
		expect(screen.getAllByLabelText("Keine Vorschau")).toHaveLength(3);
		expect(screen.queryByText("Keine Vorschau")).not.toBeInTheDocument();
		expect(screen.queryByText("Fahrzeug")).not.toBeInTheDocument();
		const memberRows = document.querySelectorAll(".vehicle-set-child-row");
		expect(within(memberRows[0].querySelector(".vehicle-column-articleNumber") as HTMLElement)
			.getByText("45923")).toBeVisible();
		expect(within(memberRows[1].querySelector(".vehicle-column-articleNumber") as HTMLElement)
			.getByText("45923")).toBeVisible();
		expect(within(memberRows[0].querySelector(".vehicle-column-inventoryNumber") as HTMLElement)
			.getByText("50 80 11-35 001-2")).toBeVisible();
		const setRow = document.querySelector(".vehicle-set-inventory-row") as HTMLElement;
		expect(within(setRow.querySelector(".vehicle-column-railwayCompany") as HTMLElement).getByText("DB")).toBeVisible();
		expect(within(setRow.querySelector(".vehicle-column-category") as HTMLElement).getByText("Wagen")).toBeVisible();
	});

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
