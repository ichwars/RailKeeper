import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api, type VehicleSet } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { VehicleSetSummaryDialog } from "./VehicleSetSummaryDialog";

const setFixture = (): VehicleSet => ({
	id: "set-1", inventoryNumber: "RK-SET-000001", name: "Rheingold", manufacturer: "Roco",
	articleNumber: "45923", gauge: "H0", category: "Wagen", gattung: "Reisezugwagen",
	members: [vehicleFixture({ id: "member-2", name: "Wagen 2" }), vehicleFixture({ id: "member-1", name: "Wagen 1" })],
	createdAt: "2026-08-01T00:00:00Z", updatedAt: "2026-08-01T00:00:00Z"
});

describe("VehicleSetSummaryDialog", () => {
	afterEach(() => vi.restoreAllMocks());

	it("loads and shows the canonical set to viewers without edit actions", async () => {
		vi.spyOn(api, "vehicleSet").mockResolvedValue(setFixture());
		render(<VehicleSetSummaryDialog setId="set-1" canEdit={false} onClose={vi.fn()}
			onUpdated={vi.fn()} onDuplicate={vi.fn()} />);

		expect(screen.getByText(/laden/i)).toBeInTheDocument();
		expect(await screen.findByText("RK-SET-000001")).toBeInTheDocument();
		expect(screen.getByText("Wagen 1")).toBeInTheDocument();
		expect(screen.queryByRole("button", { name: /bearbeiten/i })).not.toBeInTheDocument();
	});

	it("shows load errors and exposes edit and duplicate to editors", async () => {
		const user = userEvent.setup();
		vi.spyOn(api, "vehicleSet").mockResolvedValue(setFixture());
		const onEdit = vi.fn();
		const onDuplicate = vi.fn();
		render(<VehicleSetSummaryDialog setId="set-1" canEdit onClose={vi.fn()}
			onUpdated={vi.fn()} onEdit={onEdit} onDuplicate={onDuplicate} />);
		await screen.findByText("RK-SET-000001");
		await user.click(screen.getByRole("button", { name: /bearbeiten/i }));
		await user.click(screen.getByRole("button", { name: /duplizieren/i }));
		expect(onEdit).toHaveBeenCalledWith("set-1");
		await waitFor(() => expect(onDuplicate).toHaveBeenCalled());
	});
});
