import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api, type PlanTrackObject, type TrackMaterialStatus } from "../../shared/api";
import { TrackPlanReservationDialog } from "./TrackPlanReservationDialog";

const material: TrackMaterialStatus = {
  geometryId: "g1", manufacturer: "Tillig", articleNumber: "83101", name: "Gleisstück G1",
  requiredQuantity: 1, productIds: ["product-1"], inventoryNumbers: ["RK-ART-000001"],
  physicalQuantity: 3, reservedQuantity: 0, availableQuantity: 3, missingQuantity: 0
};
const object = {
  id: "object-1", lineageId: "lineage-1", revisionId: "revision-1", geometryId: "g1",
  geometry: { id: "g1", libraryId: "tillig-v1", articleNumber: "83101", name: "Gleisstück G1",
    kind: "straight", lengthMm: 166, sourceUrl: "https://example.test", status: "verified",
    createdAt: "2026-08-09T10:00:00Z", geometry: { schemaVersion: 1, ports: [], routes: [] } },
  positionXMm: 0, positionYMm: 0, rotationDegrees: 0,
  elevationStartMm: 0, elevationEndMm: 0, version: 2,
  createdAt: "2026-08-09T10:00:00Z", updatedAt: "2026-08-09T10:00:00Z"
} satisfies PlanTrackObject;

describe("TrackPlanReservationDialog", () => {
  afterEach(() => vi.restoreAllMocks());

  it("requires explicit confirmation and reserves the selected plan object", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "accessoryStock").mockResolvedValue({
      productId: "product-1", trackingMode: "quantity", totalQuantity: 3,
      locations: [{ locationId: "location-1", locationName: "Gleislager", quantity: 3,
        updatedAt: "2026-08-09T10:00:00Z" }]
    });
    vi.spyOn(api, "accessoryAssets").mockResolvedValue([]);
    vi.spyOn(api, "storageLocations").mockResolvedValue([]);
    const reserve = vi.spyOn(api, "reserveTrackPlanMaterials").mockResolvedValue({
      revisionId: "revision-1", reservations: [], materials: [material]
    });
    const reserved = vi.fn();
    render(<TrackPlanReservationDialog revisionId="revision-1" object={object} material={material}
      onClose={vi.fn()} onReserved={reserved} />);

    const dialog = screen.getByRole("dialog", { name: "Gleismaterial reservieren" });
    await screen.findByText("Gleislager · 3 Stück");
    expect(screen.getByRole("button", { name: "Reservieren" })).toBeDisabled();
    await user.click(screen.getByRole("checkbox", { name: /verbindlich.*reservieren/i }));
    await user.click(screen.getByRole("button", { name: "Reservieren" }));

    await waitFor(() => expect(reserve).toHaveBeenCalledWith("revision-1", {
      confirmed: true,
      items: [{ trackObjectId: "object-1", productId: "product-1", locationId: "location-1",
        expectedObjectVersion: 2 }]
    }));
    expect(reserved).toHaveBeenCalled();
    expect(dialog).toBeInTheDocument();
  });
});
