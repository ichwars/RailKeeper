import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api, type VehicleSet } from "../../shared/api";
import { VehicleSetEditorDialog } from "./VehicleSetEditorDialog";

const setFixture: VehicleSet = {
	id: "set-1", inventoryNumber: "RK-SET-000001", name: "Rheingold", manufacturer: "Roco",
	articleNumber: "45923", gauge: "H0", category: "Wagen", gattung: "Reisezugwagen", members: [],
	createdAt: "2026-08-01T00:00:00Z", updatedAt: "2026-08-01T00:00:00Z", mainImageMode: "automatic"
};

const memberImage = {
	id: "image-1", vehicleId: "member-1", url: "/image-1", thumbnailUrl: "/image-1-thumb",
	isPrimary: true, sortOrder: 0, createdAt: "2026-08-01T00:00:00Z"
};

describe("VehicleSetEditorDialog", () => {
	afterEach(() => vi.restoreAllMocks());

	it("separates general fields and image management into tabs", async () => {
		const user = userEvent.setup();
		vi.spyOn(api, "vehicleSet").mockResolvedValue(setFixture);
		render(<VehicleSetEditorDialog setId="set-1" onClose={vi.fn()} onUpdated={vi.fn()} />);

		const general = await screen.findByRole("tab", { name: "Allgemein" });
		const upload = screen.getByRole("tab", { name: "Upload" });
		expect(general).toHaveAttribute("aria-selected", "true");
		expect(screen.getByLabelText(/bezeichnung/i)).toBeVisible();
		expect(screen.queryByRole("heading", { name: "Hauptbild" })).not.toBeInTheDocument();

		await user.click(upload);
		expect(upload).toHaveAttribute("aria-selected", "true");
		expect(screen.getByRole("heading", { name: "Hauptbild" })).toBeVisible();
		expect(screen.queryByLabelText(/bezeichnung/i)).not.toBeInTheDocument();
		expect(screen.queryByRole("button", { name: "Speichern" })).not.toBeInTheDocument();
		const footer = screen.getByRole("dialog").querySelector("footer");
		expect(footer).not.toBeNull();
		expect(within(footer!).getByRole("button", { name: "Schließen" })).toBeVisible();
	});

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

	it("selects member images, uploads a dedicated image and restores automatic selection", async () => {
		const user = userEvent.setup();
		const fixture = {
			...setFixture,
			mainImageMode: "automatic" as const,
			members: [{
				id: "member-1", inventoryNumber: "RK-WAG-1", name: "Speisewagen", manufacturer: "Roco",
				gauge: "H0", digital: false, dtDecoder: false, exhibitionReady: false, exhibition: false,
				abcBrakes: false, couplingSame: false, driveEnabled: false, headlightsEnabled: false,
				lightingEnabled: false, soundGeneratorEnabled: false, smokeGeneratorEnabled: false,
				qrCodeEnabled: false, images: [memberImage], createdAt: "2026-08-01T00:00:00Z",
				updatedAt: "2026-08-01T00:00:00Z"
			}]
		};
		vi.spyOn(api, "vehicleSet").mockResolvedValue(fixture);
		const choose = vi.spyOn(api, "setVehicleSetMainImage").mockResolvedValue({
			...fixture, mainImageMode: "member", selectedMemberImageId: "image-1"
		});
		const upload = vi.spyOn(api, "uploadVehicleSetImage").mockResolvedValue({
			...fixture, mainImageMode: "dedicated"
		});
		render(<VehicleSetEditorDialog setId="set-1" onClose={vi.fn()} onUpdated={vi.fn()} />);

		await user.click(await screen.findByRole("tab", { name: "Upload" }));
		expect(screen.getByRole("heading", { name: /Hauptbild/ })).toBeVisible();
		await user.click(screen.getByRole("button", { name: /Als Hauptbild verwenden/ }));
		expect(choose).toHaveBeenCalledWith("set-1", { mode: "member", memberImageId: "image-1" });

		const file = new File(["image"], "set.png", { type: "image/png" });
		await user.upload(screen.getByLabelText(/Eigenes Setbild hochladen/), file);
		expect(upload).toHaveBeenCalledWith("set-1", file);

		await user.click(screen.getByRole("button", { name: /Automatische Auswahl verwenden/ }));
		expect(choose).toHaveBeenCalledWith("set-1", { mode: "automatic" });
	});
});
