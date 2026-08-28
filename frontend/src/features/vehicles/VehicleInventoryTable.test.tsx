import { fireEvent, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { vehicleFixture } from "../../test/fixtures/vehicles";
import { VehicleInventoryTable } from "./VehicleInventoryTable";
import { vehicleTableColumnWidthDefinitions } from "./vehicleTableColumns";

describe("VehicleInventoryTable", () => {
	it("renders sets collapsed by default and aligns children after expansion", async () => {
		const user = userEvent.setup();
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
		expect(document.querySelectorAll(".vehicle-set-child-row")).toHaveLength(0);
		await user.click(screen.getByRole("button", { name: /Set aufklappen/ }));
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
		expect(within(setRow.querySelector(".vehicle-column-category") as HTMLElement)
			.getByText("Verschieden")).toBeVisible();
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
    expect(screen.getByRole("table")).toHaveStyle("--vehicle-data-column-count: 3");
  });

  it("applies configured widths and exposes handles only for configurable columns", () => {
    const onCommitColumnWidth = vi.fn();
    render(
      <VehicleInventoryTable
        vehicles={[vehicleFixture()]}
        columns={["inventoryNumber", "manufacturer"]}
        columnWidths={{ manufacturer: 196 }}
        allVisibleSelected={false}
        selectedVehicleIDs={new Set()}
        sort={{ key: "inventoryNumber", direction: "asc" }}
        onToggleSort={vi.fn()}
        onToggleSelection={vi.fn()}
        onToggleAllVisibleSelection={vi.fn()}
        onOpenDetail={vi.fn()}
        onToggleExhibition={vi.fn()}
        onPreviewColumnWidth={vi.fn()}
        onCommitColumnWidth={onCommitColumnWidth}
        renderQuickMenu={() => null}
      />
    );

    const expectedWidth = 64 + 122 +
      vehicleTableColumnWidthDefinitions.inventoryNumber.defaultWidth + 196;
    const table = screen.getByRole("table");
    expect(table).toHaveStyle(
      `--vehicle-table-min-width: ${expectedWidth}px`
    );
    expect(table.querySelector('col[data-column="inventoryNumber"]')).toHaveStyle(
      `width: ${vehicleTableColumnWidthDefinitions.inventoryNumber.defaultWidth}px`
    );
    expect(table.querySelector('col[data-column="manufacturer"]')).toHaveStyle("width: 196px");
    expect(table.querySelector("col.table-fill-cell")).toHaveStyle(
      `width: max(0px, calc(100% - ${expectedWidth}px))`
    );
    expect(table.querySelector("col.select-cell")).toHaveStyle({
      width: "64px",
      minWidth: "64px",
      maxWidth: "64px"
    });
    expect(table.querySelector("col.actions-cell")).toHaveStyle({
      width: "122px",
      minWidth: "122px",
      maxWidth: "122px"
    });
    expect(screen.getAllByRole("separator")).toHaveLength(2);
    fireEvent.keyDown(screen.getByRole("separator", {
      name: "Breite von Hersteller ändern"
    }), { key: "ArrowRight" });
    expect(onCommitColumnWidth).toHaveBeenCalledWith("manufacturer", 204);
  });

  it("hides column resizers while profile widths are loading", () => {
    render(
      <VehicleInventoryTable
        vehicles={[vehicleFixture()]}
        columns={["inventoryNumber", "manufacturer"]}
        columnWidthsLoading
        allVisibleSelected={false}
        selectedVehicleIDs={new Set()}
        sort={{ key: "inventoryNumber", direction: "asc" }}
        onToggleSort={vi.fn()}
        onToggleSelection={vi.fn()}
        onToggleAllVisibleSelection={vi.fn()}
        onOpenDetail={vi.fn()}
        onToggleExhibition={vi.fn()}
        onPreviewColumnWidth={vi.fn()}
        onCommitColumnWidth={vi.fn()}
        renderQuickMenu={() => null}
      />
    );

    expect(screen.queryAllByRole("separator")).toHaveLength(0);
  });

  it("keeps the exhibition control operational for eligible vehicles", async () => {
    const user = userEvent.setup();
    const vehicle = vehicleFixture({ exhibition: false });
    const onToggleExhibition = vi.fn();
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
        onToggleExhibition={onToggleExhibition}
        renderQuickMenu={() => null}
      />
    );

    const row = screen.getAllByRole("row")[1];
    expect(within(row).getAllByRole("checkbox")).toHaveLength(2);
    const exhibitionControl = within(row).getByRole("checkbox", { name: "Ausstellungsstatus ändern" });
    expect(exhibitionControl).toBeEnabled();
    expect(exhibitionControl).not.toBeChecked();

    await user.click(exhibitionControl);
    expect(onToggleExhibition).toHaveBeenCalledWith(vehicle, true);
  });

  it("keeps the exhibition control operational for already active vehicles", async () => {
    const user = userEvent.setup();
    const vehicle = vehicleFixture({
      exhibition: true,
      digital: false,
      digitalDecoderNumber: ""
    });
    const onToggleExhibition = vi.fn();
    render(
      <VehicleInventoryTable
        vehicles={[vehicle]}
        columns={["exhibition"]}
        allVisibleSelected={false}
        selectedVehicleIDs={new Set()}
        sort={{ key: "inventoryNumber", direction: "asc" }}
        onToggleSort={vi.fn()}
        onToggleSelection={vi.fn()}
        onToggleAllVisibleSelection={vi.fn()}
        onOpenDetail={vi.fn()}
        onToggleExhibition={onToggleExhibition}
        renderQuickMenu={() => null}
      />
    );

    const exhibitionControl = screen.getByRole("checkbox", { name: "Ausstellungsstatus ändern" });
    expect(exhibitionControl).toBeEnabled();
    expect(exhibitionControl).toBeChecked();

    await user.click(exhibitionControl);
    expect(onToggleExhibition).toHaveBeenCalledWith(vehicle, false);
  });

  it("explains an unavailable exhibition status without rendering a disabled control", () => {
    const vehicle = vehicleFixture({
      exhibition: false,
      digital: false,
      digitalDecoderNumber: ""
    });
    render(
      <VehicleInventoryTable
        vehicles={[vehicle]}
        columns={["exhibition"]}
        allVisibleSelected={false}
        selectedVehicleIDs={new Set()}
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
    expect(
      within(row).queryByRole("checkbox", { name: "Ausstellungsstatus ändern" })
    ).not.toBeInTheDocument();
    expect(within(row).getByText("Digital ja + Decoder-Nr. erforderlich")).toBeVisible();
    const status = within(row).getByLabelText(
      "Ausstellung ist erst aktiv, wenn Digital ja und eine Decoder-Nr. gepflegt ist."
    );
    expect(status.querySelector("svg")).toBeInTheDocument();
  });
});
