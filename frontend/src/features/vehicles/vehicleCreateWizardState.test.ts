import { beforeEach, describe, expect, it, vi } from "vitest";

import { emptyVehicle } from "./vehicleViewModel";
import {
  clearVehicleCreateDraft,
  createVehicleCreateWizardState,
  emptyVehicleSetMemberDraft,
  loadVehicleCreateDraft,
  saveVehicleCreateDraft,
  vehicleCreateDraftKey,
  vehicleCreateWizardReducer
} from "./vehicleCreateWizardState";

describe("vehicleCreateWizardState", () => {
  beforeEach(() => localStorage.clear());

  it("moves between named steps and explicit article stages", () => {
    const initial = createVehicleCreateWizardState(emptyVehicle);
    const article = vehicleCreateWizardReducer(initial, { type: "go-to-step", step: "article" });
    const results = vehicleCreateWizardReducer(article, { type: "set-article-stage", stage: "results" });
    const review = vehicleCreateWizardReducer(results, { type: "select-article-result", index: 2 });

    expect(article.step).toBe("article");
    expect(results.articleStage).toBe("results");
    expect(review).toMatchObject({ articleStage: "review", selectedResultIndex: 2 });
  });

  it("preserves member forms until a reduction is explicitly confirmed", () => {
    const initial = {
      ...createVehicleCreateWizardState(emptyVehicle),
      kind: "set" as const,
      members: [
        emptyVehicleSetMemberDraft(),
        { form: { ...emptyVehicle, name: "Speisewagen" }, touched: true },
        emptyVehicleSetMemberDraft()
      ]
    };

    const requested = vehicleCreateWizardReducer(initial, { type: "set-member-count", count: 2 });
    expect(requested.members).toHaveLength(3);
    expect(requested.pendingMemberReduction).toEqual({ requestedCount: 2, populatedIndexes: [1] });

    const confirmed = vehicleCreateWizardReducer(requested, { type: "confirm-member-reduction" });
    expect(confirmed.members).toHaveLength(2);
    expect(confirmed.members[1].form.name).toBe("Speisewagen");
  });

  it("restores a valid version-1 draft and rejects an incompatible version", () => {
    const state = createVehicleCreateWizardState({ ...emptyVehicle, name: "TEE Roland" });
    expect(saveVehicleCreateDraft(state)).toEqual({ kind: "saved" });
    expect(loadVehicleCreateDraft()).toEqual(expect.objectContaining({
      kind: "loaded",
      state: expect.objectContaining({ shared: expect.objectContaining({ name: "TEE Roland" }) })
    }));

    localStorage.setItem(vehicleCreateDraftKey, JSON.stringify({ version: 99, savedAt: "now", state: {} }));
    expect(loadVehicleCreateDraft()).toEqual({ kind: "invalid" });
  });

  it("rejects malformed state and keeps storage failures non-blocking", () => {
    localStorage.setItem(vehicleCreateDraftKey, JSON.stringify({
      version: 1,
      savedAt: new Date().toISOString(),
      state: { kind: "set", step: "basics", members: "invalid" }
    }));
    expect(loadVehicleCreateDraft()).toEqual({ kind: "invalid" });

    const storage = {
      getItem: vi.fn(() => { throw new Error("unavailable"); }),
      setItem: vi.fn(() => { throw new Error("unavailable"); }),
      removeItem: vi.fn(() => { throw new Error("unavailable"); })
    };
    expect(loadVehicleCreateDraft(storage)).toEqual({ kind: "error" });
    expect(saveVehicleCreateDraft(createVehicleCreateWizardState(emptyVehicle), storage)).toEqual({ kind: "error" });
    expect(clearVehicleCreateDraft(storage)).toEqual({ kind: "error" });
  });
});
