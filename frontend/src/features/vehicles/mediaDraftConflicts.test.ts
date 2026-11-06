import { describe, expect, it } from "vitest";
import type { VehicleAttachment } from "../../shared/api";
import { imageFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { attachmentsToEditState, vehicleImagesToPending } from "./vehicleTransforms";
import { activeMediaConflicts, refreshMediaDraft, resolveMediaImage } from "./mediaDraftConflicts";

const attachment: VehicleAttachment = {
  id: "a", vehicleId: "vehicle-1", fileName: "manual.txt", originalName: "manual.txt", sizeBytes: 4,
  description: "Original", category: "manual", createdAt: "original", updatedAt: "original", revision: "a".repeat(64)
};

describe("media draft conflicts", () => {
  it("shows current typing in conflict comparisons and releases fields matching stored values", () => {
    const before = vehicleFixture({ images: [imageFixture({ title: "Original" })] });
    const images = vehicleImagesToPending(before).map(image => ({ ...image, title: "Own" }));
    const incoming = vehicleFixture({ images: [imageFixture({ title: "Newer" })] });
    const result = refreshMediaDraft(images, { edits: {}, conflicts: [] }, before, incoming);
    const typing = result.images.map(image => ({ ...image, title: "Own revised again" }));
    expect(activeMediaConflicts(result.state, typing, incoming)[0].fields[0])
      .toEqual({ field: "title", local: "Own revised again", saved: "Newer" });
    expect(activeMediaConflicts(result.state, typing.map(image => ({ ...image, title: "Newer" })), incoming)).toEqual([]);
  });

  it("merges untouched image metadata and new uploads while retaining a dirty title", () => {
    const before = vehicleFixture({ images: [imageFixture({ title: "Original" })] });
    const images = vehicleImagesToPending(before).map(image => ({ ...image, title: "Own" }));
    const incoming = vehicleFixture({ images: [imageFixture({ title: "Original", maintenanceId: "new-link" }),
      imageFixture({ id: "new-upload", title: "Upload", isPrimary: false })] });
    const result = refreshMediaDraft(images, { edits: {}, conflicts: [] }, before, incoming);
    expect(result.state.conflicts).toEqual([]);
    expect(result.images).toEqual([expect.objectContaining({ title: "Own", maintenanceId: "new-link" }),
      expect.objectContaining({ id: "new-upload" })]);
  });

  it("retains a field conflict across refreshes and uses saved values only for the reviewed fields", () => {
    const before = vehicleFixture({ images: [imageFixture({ title: "Original" })] });
    const images = vehicleImagesToPending(before).map(image => ({ ...image, title: "Own", maintenanceId: "own-link" }));
    const incoming = vehicleFixture({ images: [imageFixture({ title: "Newer" })] });
    const first = refreshMediaDraft(images, { edits: {}, conflicts: [] }, before, incoming);
    const again = refreshMediaDraft(first.images, first.state, incoming, incoming);
    expect(again.state.conflicts).toHaveLength(1);
    expect(again.state.conflicts[0].fields).toEqual([{ field: "title", local: "Own", saved: "Newer" }]);
    expect(resolveMediaImage(again.images, again.state.conflicts[0], incoming, "saved")[0])
      .toMatchObject({ title: "Newer", maintenanceId: "own-link" });
    expect(resolveMediaImage(again.images, again.state.conflicts[0], incoming, "draft")[0].title).toBe("Own");
  });

  it("keeps a deleted image's pending metadata visible until explicitly discarded", () => {
    const before = vehicleFixture();
    const images = vehicleImagesToPending(before).map(image => ({ ...image, title: "Keep for review" }));
    const incoming = vehicleFixture({ images: [] });
    const first = refreshMediaDraft(images, { edits: {}, conflicts: [] }, before, incoming);
    const again = refreshMediaDraft(first.images, first.state, incoming, incoming);
    expect(again.images).toHaveLength(1);
    expect(again.state.conflicts[0]).toMatchObject({ kind: "image", deleted: true,
      fields: [expect.objectContaining({ field: "title", local: "Keep for review" })] });
    expect(resolveMediaImage(again.images, again.state.conflicts[0], incoming, "saved")).toEqual([]);
  });

  it("detects competing primary-image selections even when they affect different image rows", () => {
    const before = vehicleFixture({ images: [imageFixture({ title: "A" }),
      imageFixture({ id: "b", title: "B", isPrimary: false }), imageFixture({ id: "c", title: "C", isPrimary: false })] });
    const images = vehicleImagesToPending(before).map(image => ({ ...image, isPrimary: image.id === "b" }));
    const incoming = { ...before, images: before.images?.map(image => ({ ...image, isPrimary: image.id === "c" })) };
    const result = refreshMediaDraft(images, { edits: {}, conflicts: [] }, before, incoming);
    expect(result.state.conflicts.map(conflict => conflict.id)).toEqual(["@primary"]);
    const chosen = resolveMediaImage(result.images, result.state.conflicts[0], incoming, "saved");
    expect(chosen.filter(image => image.isPrimary).map(image => image.id)).toEqual(["c"]);
  });

  it("detects competing image orders and preserves unsaved remote images when using the saved order", () => {
    const before = vehicleFixture({ images: [imageFixture(), imageFixture({ id: "b", isPrimary: false }),
      imageFixture({ id: "c", isPrimary: false })] });
    const [a, b, c] = vehicleImagesToPending(before);
    const images = [b, a, c, { id: "remote", url: "https://example.test/pending.png", source: "", title: "Pending" }];
    const incoming = { ...before, images: [before.images![0], before.images![2], before.images![1]] };
    const result = refreshMediaDraft(images, { edits: {}, conflicts: [] }, before, incoming);
    const conflict = result.state.conflicts.find(value => value.id === "@order")!;
    expect(conflict).toBeDefined();
    expect(resolveMediaImage(result.images, conflict, incoming, "saved").map(image => image.id))
      .toEqual(["image-1", "c", "b", "remote"]);
  });

  it("merges unrelated attachment metadata while preserving a conflicting edit and its review", () => {
    const before = vehicleFixture({ attachments: [attachment] });
    const state = { edits: { ...attachmentsToEditState([attachment]), a: {
      ...attachmentsToEditState([attachment]).a, description: "Own"
    } }, conflicts: [] };
    const incoming = vehicleFixture({ attachments: [{ ...attachment, description: "Newer", category: "invoice", revision: "b".repeat(64) }] });
    const result = refreshMediaDraft(vehicleImagesToPending(before), state, before, incoming);
    expect(result.state.edits.a).toMatchObject({ description: "Own", category: "invoice", expectedRevision: "b".repeat(64) });
    expect(result.state.conflicts).toEqual([expect.objectContaining({ kind: "attachment", id: "a",
      fields: [{ field: "description", local: "Own", saved: "Newer" }] })]);
    const repeated = refreshMediaDraft(result.images, result.state, incoming, incoming);
    expect(repeated.state.conflicts).toEqual(result.state.conflicts);
    const removed = refreshMediaDraft(result.images, result.state, incoming, vehicleFixture({ attachments: [] }));
    expect(removed.state.conflicts[0]).toMatchObject({ kind: "attachment", deleted: true });
    expect(removed.state.edits.a.description).toBe("Own");
  });
});
