import { describe, expect, it } from "vitest";
import { cvValueFixture } from "../../test/fixtures/vehicles";
import { cvDraftConflicts, cvDraftInput, cvDraftRequest, cvDraftValues, maxManualCVChanges,
  rebaseCVChange, stageCVDeletion, stageCVInput, type CVDraftChange } from "./cvDraft";

describe("manual CV draft", () => {
  const stored = [cvValueFixture({ id: "base", cvNumber: 257, decoderProfile: "A" }),
    cvValueFixture({ id: "zero", cvNumber: 257, decoderProfile: "A", indexCV31: 0, indexCV32: 0 }),
    cvValueFixture({ id: "other-profile", cvNumber: 257, decoderProfile: "B" })];

  it("stages exact profile/index identities and preserves unspecified metadata", () => {
    const result = stageCVInput([], stored, "vehicle-1", { cvNumber: 257, value: 0, decoderProfile: " A ", indexCV31: 0, indexCV32: 0 }, null, "new");
    expect(result.error).toBeUndefined();
    const changes = result.changes || [];
    expect(cvDraftRequest(changes)).toMatchObject([{ id: "zero", expectedRevision: "c".repeat(64), value: {
      value: 0, description: "Adresse", decoderProfile: "A", indexCV31: 0, indexCV32: 0
    } }]);
    const working = cvDraftValues(stored, changes, "vehicle-1");
    expect(working.find(value => value.id === "base")).toEqual(stored[0]);
    expect(working.find(value => value.id === "other-profile")).toEqual(stored[2]);
    expect(stored[1].value).toBe(3);
  });

  it("rejects profile/index moves onto another identity without changing the draft", () => {
    expect(stageCVInput([], stored, "vehicle-1", { ...cvDraftInput(stored[0]), indexCV31: 0, indexCV32: 0 }, "base", "new"))
      .toEqual({ error: "occupied" });
    expect(stageCVInput([], stored, "vehicle-1", { ...cvDraftInput(stored[0]), decoderProfile: "B" }, "base", "new"))
      .toEqual({ error: "occupied" });
    expect(stageCVInput([], stored, "vehicle-1", { cvNumber: 257, value: 7, indexCV31: 0 }, null, "new"))
      .toEqual({ error: "invalid" });
  });

  it("keeps the original revision and source evidence when a pending row is edited again", () => {
    const original = cvValueFixture({ sourceLine: 12, sourceExportedAt: "2026-10-17T17:41:59", sourceFileId: "owned-file" });
    const first = stageCVInput([], [original], "vehicle-1", { ...cvDraftInput(original), value: 7 }, original.id, "new").changes || [];
    const elsewhere = { ...original, revision: "d".repeat(64), value: 9, description: "Newer source note" };
    const second = stageCVInput(first, [elsewhere], "vehicle-1", { ...cvDraftInput(original), value: 8 }, original.id, "new").changes || [];
    expect(cvDraftRequest(second)[0]).toMatchObject({ expectedRevision: original.revision, value: {
      sourceLine: 12, sourceExportedAt: original.sourceExportedAt, sourceFileId: "owned-file", value: 8
    } });
    expect(cvDraftConflicts(second, [elsewhere])).toHaveLength(1);
    const resolved = rebaseCVChange(second, original.id, elsewhere);
    expect(cvDraftConflicts(resolved, [elsewhere])).toEqual([]);
    expect(cvDraftRequest(resolved)[0]).toMatchObject({ expectedRevision: elsewhere.revision,
      value: { value: 8, description: "Newer source note" } });
  });

  it("stages deletions, removes unsaved rows locally, and never sends draft response fields", () => {
    const original = stored[0];
    const deleted = stageCVDeletion([], original) || [];
    expect(cvDraftRequest(deleted)).toEqual([{ id: original.id, expectedRevision: original.revision, delete: true }]);
    expect(cvDraftValues(stored, deleted, "vehicle-1")).toHaveLength(2);
    const created = stageCVInput([], [], "vehicle-1", { cvNumber: 1, value: 0 }, null, "local-new").changes || [];
    const pending = cvDraftValues([], created, "vehicle-1")[0];
    expect(stageCVDeletion(created, pending)).toEqual([]);
    expect(cvDraftRequest(created)[0]).not.toHaveProperty("id");
    expect(cvDraftRequest(created)[0].value).not.toHaveProperty("revision");
    expect(cvDraftRequest(created)[0].value).not.toHaveProperty("history");
  });

  it("retains a remotely deleted pending row for conflict review and rejects implicit recreation", () => {
    const changed = stageCVInput([], [stored[0]], "vehicle-1", { ...cvDraftInput(stored[0]), value: 7 }, "base", "new").changes || [];
    expect(cvDraftValues([], changed, "vehicle-1")).toMatchObject([{ id: "base", value: 7 }]);
    expect(cvDraftConflicts(changed, [])).toHaveLength(1);
    expect(stageCVInput([], [], "vehicle-1", { cvNumber: 257, value: 7 }, "base", "new")).toEqual({ error: "revision" });
  });

  it("detects a newly occupied target even when the edited row's own revision stayed unchanged", () => {
    const changes = stageCVInput([], [stored[0]], "vehicle-1", { ...cvDraftInput(stored[0]), decoderProfile: "B" }, "base", "new").changes || [];
    expect(cvDraftConflicts(changes, [stored[0]])).toEqual([]);
    expect(cvDraftConflicts(changes, [stored[0], stored[2]])).toHaveLength(1);
  });

  it("orders explicit identity moves before dependent rows and keeps cycles as visible conflicts", () => {
    const a = cvValueFixture({ id: "a", decoderProfile: "A" }), b = cvValueFixture({ id: "b", decoderProfile: "B" });
    const changes: CVDraftChange[] = [{ id: a.id, original: a, value: { ...cvDraftInput(a), decoderProfile: "B" } },
      { id: b.id, original: b, value: { ...cvDraftInput(b), decoderProfile: "C" } }];
    expect(cvDraftConflicts(changes, [a, b])).toEqual([]);
    expect(cvDraftRequest(changes).map(change => change.id)).toEqual(["b", "a"]);
    const cyclic = [changes[0], { ...changes[1], value: { ...cvDraftInput(b), decoderProfile: "A" } }];
    expect(cvDraftConflicts(cyclic, [a, b])).toHaveLength(2);
    expect(() => cvDraftRequest(cyclic)).toThrow("cyclic identity moves");
  });

  it("limits pending changes while allowing corrections and discarding individual entries", () => {
    const full: CVDraftChange[] = Array.from({ length: maxManualCVChanges }, (_, index) => ({ id: `local-${index}`,
      original: null, value: { cvNumber: 1, value: 1, decoderProfile: `P${index}` } }));
    expect(stageCVInput(full, [], "vehicle-1", { cvNumber: 1, value: 1, decoderProfile: "another" }, null, "extra"))
      .toEqual({ error: "limit" });
    expect(stageCVInput(full, [], "vehicle-1", { cvNumber: 1, value: 2, decoderProfile: "P0" }, "local-0", "extra").changes)
      .toHaveLength(maxManualCVChanges);
    expect(stageCVDeletion(full, cvDraftValues([], full, "vehicle-1")[0])).toHaveLength(maxManualCVChanges - 1);
  });
});
