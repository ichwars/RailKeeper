import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { vehicleFixture } from "../../test/fixtures/vehicles";
import { VehicleSetInventoryRow } from "./VehicleSetInventoryRow";
import type { VehicleInventorySetGroup } from "./vehicleSetGroups";

const groupFixture = (): VehicleInventorySetGroup => {
	const set = {
		id: "set-1", inventoryNumber: "RK-SET-000001", name: "Rheingold", manufacturer: "Roco",
		articleNumber: "45923", gauge: "H0", epoch: "III", acquisitionType: "Neu",
		purchaseDate: "2026-08-01", purchasePrice: "179.90", condition: "Neuwertig",
		memberCount: 2, position: 1
	};
	return {
		kind: "set", id: set.id, set, visibleMemberCount: 2, totalMemberCount: 2,
		members: [
			vehicleFixture({ id: "member-1", images: [], vehicleSet: set }),
			vehicleFixture({ id: "member-2", images: [], vehicleSet: { ...set, position: 2 } })
		]
	};
};

describe("VehicleSetInventoryRow", () => {
	it("renders aligned cells and tri-state selection for visible members", async () => {
		const user = userEvent.setup();
		const onToggleSelection = vi.fn();
		const { container } = render(
			<table><tbody><VehicleSetInventoryRow
				group={groupFixture()}
				columns={["type", "inventoryNumber", "manufacturer", "name"]}
				collapsed={false}
				selectedVehicleIDs={new Set(["member-1"])}
				onToggleCollapsed={vi.fn()}
				onToggleSelection={onToggleSelection}
				onOpen={vi.fn()}
				onEdit={vi.fn()}
				onDuplicate={vi.fn()}
			/></tbody></table>
		);

		const checkbox = screen.getByRole("checkbox", { name: /RK-SET-000001/ });
		expect(checkbox).toHaveProperty("indeterminate", true);
		await user.click(checkbox);
		expect(onToggleSelection).toHaveBeenCalledWith(["member-1", "member-2"]);
		expect(container.querySelector(".vehicle-set-inventory-row td[colspan]")).toBeNull();
		expect(screen.getByText("RK-SET-000001")).toBeInTheDocument();
		expect(container.querySelector(
			".vehicle-set-type-cell .inventory-thumb, .vehicle-set-type-cell .image-placeholder"
		)).toBeInTheDocument();
		expect(screen.getByLabelText("Keine Vorschau")).toBeVisible();
		expect(screen.queryByText("Keine Vorschau")).not.toBeInTheDocument();
		expect(screen.queryByText("Set")).not.toBeInTheDocument();
	});
});
