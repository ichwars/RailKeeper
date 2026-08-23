import { describe, expect, it } from "vitest";

import type { VehicleSet } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { vehicleSetDuplicatePrefill } from "./vehicleSetDuplicate";

const setFixture: VehicleSet = {
	id: "set-1",
	inventoryNumber: "RK-SET-000001",
	name: "Rheingold",
	manufacturer: "Roco",
	articleNumber: "45923",
	articleSourceUrl: "https://example.test/45923",
	gauge: "H0",
	epoch: "III",
	railwayCompany: "DB",
	category: "Wagen",
	gattung: "Reisezugwagen",
	description: "Setbeschreibung",
	ean: "1234567890123",
	productionPeriod: "2020",
	listPrice: "199.90",
	acquisitionType: "Neu",
	acquiredFrom: "Fachhändler",
	purchasePrice: "179.90",
	purchaseDate: "2026-08-01",
	storageLocation: "Vitrine",
	storageDetails: "Fach 2",
	condition: "Neuwertig",
	conditionDetails: "",
	packaging: "Originalverpackung",
	members: [
		vehicleFixture({
			id: "member-1",
			inventoryNumber: "RK-WAG-000001",
			name: "Speisewagen",
			digital: true,
			digitalDecoderNumber: "17",
			lengthMm: "282",
			images: [
				{ id: "local", vehicleId: "member-1", url: "/api/v1/local", sourceUrl: "https://example.test/localized.jpg", isPrimary: true, sortOrder: 0, createdAt: "now" },
				{ id: "remote", vehicleId: "member-1", url: "https://example.test/wagen.jpg", isPrimary: false, sortOrder: 1, createdAt: "now" }
			]
		}),
		vehicleFixture({ id: "member-2", inventoryNumber: "RK-WAG-000002", name: "Abteilwagen" })
	],
	createdAt: "2026-08-01T00:00:00Z",
	updatedAt: "2026-08-01T00:00:00Z",
	mainImageMode: "automatic"
};

describe("vehicleSetDuplicatePrefill", () => {
	it("clears set and member identities while retaining safe catalogue data", () => {
		const prefill = vehicleSetDuplicatePrefill(setFixture);

		expect(prefill.kind).toBe("set");
		expect(prefill.shared.inventoryNumber).toBe("");
		expect(prefill.shared.name).toBe("Rheingold");
		expect(prefill.members.every((member) => member.inventoryNumber === "")).toBe(true);
		expect(prefill.members.map((member) => member.name)).toEqual(["Speisewagen", "Abteilwagen"]);
		expect(prefill.shared).toMatchObject({ digital: false, digitalDecoderNumber: "", lengthMm: "" });
		expect(prefill.members[0]).toMatchObject({ digital: true, digitalDecoderNumber: "17", lengthMm: "282" });
		expect(prefill.members[0].images).toEqual([
			expect.objectContaining({ url: "https://example.test/localized.jpg" }),
			expect.objectContaining({ url: "https://example.test/wagen.jpg" })
		]);
	});
});
