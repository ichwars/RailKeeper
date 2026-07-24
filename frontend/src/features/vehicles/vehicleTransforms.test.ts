import { describe, expect, it } from "vitest";

import type { VehicleAttachment } from "../../shared/api";
import {
  functionFixture,
  imageFixture,
  vehicleFixture
} from "../../test/fixtures/vehicles";
import {
  attachmentsToEditState,
  functionsToEditState,
  previewImageUrl,
  primaryImage,
  vehicleExhibitionEligible,
  vehicleImagesToPending,
  vehicleToExhibitionEntry,
  vehicleToForm
} from "./vehicleTransforms";

describe("vehicle transforms", () => {
  it("maps a persisted vehicle into an editable form", () => {
    const form = vehicleToForm(vehicleFixture({ articleSourceUrl: undefined, ean: undefined }));
    expect(form).toMatchObject({
      inventoryNumber: "RK-LOK-000001",
      manufacturer: "ESU",
      name: "BR 106",
      articleSourceUrl: "",
      ean: "",
      digital: true
    });
  });

  it("chooses the primary image and thumbnail preview", () => {
    const images = [
      imageFixture({ id: "image-1", isPrimary: false, url: "/first.jpg" }),
      imageFixture({ id: "image-2", isPrimary: true, url: "/primary.jpg", thumbnailUrl: "/thumb.jpg" })
    ];
    expect(primaryImage(images)?.url).toBe("/primary.jpg");
    expect(previewImageUrl(primaryImage(images))).toBe("/thumb.jpg");
    expect(vehicleImagesToPending(vehicleFixture({ images }))[1]).toMatchObject({
      id: "image-2",
      persisted: true,
      isPrimary: true
    });
  });

  it("creates attachment and function edit state", () => {
    const attachment: VehicleAttachment = {
      id: "attachment-1",
      vehicleId: "vehicle-1",
      fileName: "manual.pdf",
      originalName: "manual.pdf",
      description: "Anleitung",
      category: "Dokumentation",
      sizeBytes: 42,
      createdAt: "2026-07-23T08:00:00Z",
      updatedAt: "2026-07-23T08:00:00Z"
    };
    expect(attachmentsToEditState([attachment])[attachment.id]).toEqual({
      description: "Anleitung",
      category: "Dokumentation",
      maintenanceId: ""
    });
    expect(functionsToEditState([functionFixture()]).F0).toMatchObject({ name: "Licht", persisted: true });
  });

  it("checks exhibition eligibility and builds an entry", () => {
    expect(vehicleExhibitionEligible(vehicleFixture())).toBe(true);
    expect(vehicleExhibitionEligible(vehicleFixture({ digitalDecoderNumber: "" }))).toBe(false);

    const entry = vehicleToExhibitionEntry(vehicleFixture({ vehicleNumber: "106 001", additionalInfo: "DR" }), "Daniel");
    expect(entry).toMatchObject({
      vehicleId: "vehicle-1",
      owner: "Daniel",
      locomotiveName: "BR 106",
      decoderNumber: "1001",
      notes: "106 001 · DR"
    });
    expect(JSON.parse(entry.functionKeys || "[]")).toEqual([
      { key: "F0", name: "Licht", type: "licht", symbolKey: "" }
    ]);
  });
});
