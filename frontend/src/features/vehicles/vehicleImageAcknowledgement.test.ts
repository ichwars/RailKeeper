import { describe, expect, it } from "vitest";
import { imageFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { acknowledgeVehicleImage } from "./vehicleImageAcknowledgement";
import { refreshMediaDraft } from "./mediaDraftConflicts";
import { vehicleImagesToPending } from "./vehicleTransforms";

describe("acknowledged image baselines", () => {
  it("adopts newer changes to an acknowledged upload while keeping an older image's dirty title", () => {
    const before = vehicleFixture();
    const local = vehicleImagesToPending(before).map(image => ({ ...image, title: "Own older draft" }));
    const stored = imageFixture({ id: "uploaded", isPrimary: false, title: "Upload", sortOrder: 1 });
    const ack = acknowledgeVehicleImage(local, before, stored);
    const incoming = vehicleFixture({ images: [...(before.images || []), { ...stored, title: "Newer saved title" }] });
    const refreshed = refreshMediaDraft(ack.images, { edits: {}, conflicts: [] },
      { ...before, images: ack.baselineImages }, incoming);
    expect(refreshed.images.map(image => image.title)).toEqual(["Own older draft", "Newer saved title"]);
    expect(refreshed.state.conflicts).toEqual([]);
  });

  it("reflects a confirmed primary change without inventing an unsaved old primary or image order", () => {
    const before = vehicleFixture();
    const stored = imageFixture({ id: "new-primary", isPrimary: true, sortOrder: 1 });
    const ack = acknowledgeVehicleImage(vehicleImagesToPending(before), before, stored);
    expect(ack.images.map(image => image.id)).toEqual(["new-primary", "image-1"]);
    expect(ack.images.filter(image => image.isPrimary).map(image => image.id)).toEqual(["new-primary"]);
  });

  it("retains a different unsaved remote primary when another import is automatically made primary", () => {
    const before = vehicleFixture({ images: [] });
    const pending = [
      { id: "first", url: "https://example.test/first.png", source: "", title: "First", isPrimary: false },
      { id: "chosen", url: "https://example.test/second.png", source: "", title: "Chosen", isPrimary: true }
    ];
    const stored = imageFixture({ id: "stored-first", isPrimary: true });
    const ack = acknowledgeVehicleImage(pending, before, stored, "first");
    expect(ack.images.some(image => image.id === "first")).toBe(false);
    expect(ack.images.filter(image => image.isPrimary).map(image => image.id)).toEqual(["chosen"]);
    expect(ack.baselineImages[0].isPrimary).toBe(true);
  });
});
