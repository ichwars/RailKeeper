import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api, type VehicleSet } from "../../shared/api";
import { VehicleSetEditorDialog } from "./VehicleSetEditorDialog";

const setFixture: VehicleSet = {
	id: "set-1", inventoryNumber: "RK-SET-000001", name: "Rheingold", manufacturer: "Roco",
	articleNumber: "45923", gauge: "H0", category: "Wagen", gattung: "Reisezugwagen", members: [],
	createdAt: "2026-08-01T00:00:00Z", updatedAt: "2026-08-01T00:00:00Z"
};

describe("VehicleSetEditorDialog", () => {
	afterEach(() => vi.restoreAllMocks());

	it("edits shared data without exposing inventory numbers", async () => {
		const user = userEvent.setup();
		vi.spyOn(api, "vehicleSet").mockResolvedValue(setFixture);
		const update = vi.spyOn(api, "updateVehicleSet").mockResolvedValue({ ...setFixture, name: "Rheingold neu" });
		const onUpdated = vi.fn();
		render(<VehicleSetEditorDialog setId="set-1" onClose={vi.fn()} onUpdated={onUpdated} />);

		const name = await screen.findByLabelText(/bezeichnung/i);
		await user.clear(name);
		await user.type(name, "Rheingold neu");
		expect(screen.queryByDisplayValue("RK-SET-000001")).not.toBeInTheDocument();
		await user.click(screen.getByRole("button", { name: /speichern/i }));
		expect(update).toHaveBeenCalledWith("set-1", expect.objectContaining({ name: "Rheingold neu" }));
		expect(onUpdated).toHaveBeenCalledWith(expect.objectContaining({ name: "Rheingold neu" }));
	});
});
