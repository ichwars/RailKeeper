import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { vehicleFixture } from "../../test/fixtures/vehicles";
import { VehicleSetInventoryMobileCard } from "./VehicleSetInventoryMobileCard";
import type { VehicleInventorySetGroup } from "./vehicleSetGroups";

const group = (): VehicleInventorySetGroup => {
	const set = {
		id: "set-1", inventoryNumber: "RK-SET-000001", name: "Rheingold", manufacturer: "Roco",
		articleNumber: "45923", gauge: "H0", memberCount: 4, position: 1
	};
	return {
		kind: "set", id: set.id, set, visibleMemberCount: 2, totalMemberCount: 4,
		members: [vehicleFixture({ id: "member-1", inventoryNumber: "RK-WAG-000001", vehicleSet: set }),
			vehicleFixture({ id: "member-2", inventoryNumber: "RK-WAG-000002", vehicleSet: { ...set, position: 2 } })]
	};
};

describe("VehicleSetInventoryMobileCard", () => {
	it("shows canonical set data, count and all available actions", async () => {
		const user = userEvent.setup();
		const onToggleExpanded = vi.fn();
		const onOpen = vi.fn();
		const onEdit = vi.fn();
		const onDuplicate = vi.fn();
		render(<VehicleSetInventoryMobileCard group={group()} expanded={false}
			onToggleExpanded={onToggleExpanded} onOpen={onOpen} onEdit={onEdit} onDuplicate={onDuplicate} />);

		expect(screen.getByText("RK-SET-000001")).toBeVisible();
		expect(screen.getByText("Rheingold")).toBeVisible();
		expect(screen.getByText(/2 von 4/)).toBeVisible();
		await user.click(screen.getByRole("button", { name: /Set aufklappen/ }));
		await user.click(screen.getByRole("button", { name: "Anzeigen" }));
		await user.click(screen.getByRole("button", { name: "Bearbeiten" }));
		await user.click(screen.getByRole("button", { name: "Duplizieren" }));
		expect(onToggleExpanded).toHaveBeenCalled();
		expect(onOpen).toHaveBeenCalledWith("set-1");
		expect(onEdit).toHaveBeenCalledWith("set-1");
		expect(onDuplicate).toHaveBeenCalledWith("set-1");
	});
});
