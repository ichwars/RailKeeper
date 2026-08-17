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

  it("reassigns article images when their member is removed", () => {
    const initial = {
      ...createVehicleCreateWizardState(emptyVehicle),
      kind: "set" as const,
      members: [
        emptyVehicleSetMemberDraft(),
        emptyVehicleSetMemberDraft(),
        { form: { ...emptyVehicle, name: "Packwagen" }, touched: true }
      ],
      articleImageOwners: { "https://example.test/packwagen.jpg": 2 }
    };

    const requested = vehicleCreateWizardReducer(initial, { type: "set-member-count", count: 2 });
    const confirmed = vehicleCreateWizardReducer(requested, { type: "confirm-member-reduction" });

    expect(confirmed.articleImageOwners).toEqual({ "https://example.test/packwagen.jpg": 1 });
  });

  it("clamps set members to the backend limit", () => {
    const initial = { ...createVehicleCreateWizardState(emptyVehicle), kind: "set" as const };
    const resized = vehicleCreateWizardReducer(initial, { type: "set-member-count", count: 101 });
    expect(resized.members).toHaveLength(100);

    const full = vehicleCreateWizardReducer(resized, { type: "add-member" });
    expect(full.members).toHaveLength(100);
  });

  it("pads a one-member set duplicate to the backend minimum", () => {
    const state = createVehicleCreateWizardState(emptyVehicle, {
      kind: "set",
      shared: { ...emptyVehicle, name: "Restset" },
      members: [{ ...emptyVehicle, name: "Letzter Wagen" }]
    });

    expect(state.members).toHaveLength(2);
    expect(state.members[0].form.name).toBe("Letzter Wagen");
    expect(state.members[1].touched).toBe(false);
  });

  it("moves the active tab back to the set when its member is removed", () => {
    const initial = {
      ...createVehicleCreateWizardState(emptyVehicle),
      kind: "set" as const,
      members: [emptyVehicleSetMemberDraft(), emptyVehicleSetMemberDraft(), emptyVehicleSetMemberDraft()],
      activeDetailsTab: "member:2" as const
    };

    const resized = vehicleCreateWizardReducer(initial, { type: "set-member-count", count: 2 });

    expect(resized.members).toHaveLength(2);
    expect(resized.activeDetailsTab).toBe("set");
  });

  it("tracks exactly which member fields override imported shared data", () => {
    const initial = { ...createVehicleCreateWizardState(emptyVehicle), kind: "set" as const };
    const updated = vehicleCreateWizardReducer(initial, {
      type: "update-member",
      index: 0,
      patch: { digital: false, lengthMm: "240" }
    });

    expect(updated.members[0].overriddenFields).toEqual(expect.arrayContaining(["digital", "lengthMm"]));
  });

  it("restores a valid draft and rejects an incompatible version", () => {
    const state = createVehicleCreateWizardState({ ...emptyVehicle, name: "TEE Roland" });
    expect(saveVehicleCreateDraft(state)).toEqual({ kind: "saved" });
    expect(loadVehicleCreateDraft()).toEqual(expect.objectContaining({
      kind: "loaded",
      state: expect.objectContaining({ shared: expect.objectContaining({ name: "TEE Roland" }) })
    }));

    localStorage.setItem(vehicleCreateDraftKey("local"), JSON.stringify({ version: 99, savedAt: "now", state: {} }));
    expect(loadVehicleCreateDraft()).toEqual({ kind: "invalid" });
  });

  it("normalizes legacy drafts that cannot restore article search state", () => {
    const state = {
      ...createVehicleCreateWizardState({ ...emptyVehicle, name: "Legacy" }),
      step: "article" as const,
      articleStage: "review" as const,
      selectedResultIndex: 0
    };
    localStorage.setItem(vehicleCreateDraftKey("local"), JSON.stringify({
      version: 1,
      savedAt: new Date().toISOString(),
      state
    }));

    expect(loadVehicleCreateDraft()).toEqual(expect.objectContaining({
      kind: "loaded",
      state: expect.objectContaining({ articleStage: "input", selectedResultIndex: null }),
      articleSearch: null
    }));
  });

  it("persists article search selections so reviewed images can be restored", () => {
    const state = {
      ...createVehicleCreateWizardState({ ...emptyVehicle, name: "TEE Roland" }),
      step: "details" as const,
      articleStage: "review" as const,
      selectedResultIndex: 0,
      articleImportApplied: true
    };
    const articleSearch = {
      response: {
        query: "Roco 6280002",
        results: [{
          source: "manufacturer", title: "Roco 6280002", url: "https://roco.cc/6280002",
          snippet: "Set", score: 10, fields: {},
          images: [{ url: "https://roco.cc/set.jpg", title: "Set", source: "manufacturer" }]
        }]
      },
      selectedFields: {},
      selectedImages: { "image-key": true }
    };

    expect(saveVehicleCreateDraft(state, "local", articleSearch)).toEqual({ kind: "saved" });
    expect(loadVehicleCreateDraft()).toEqual(expect.objectContaining({
      kind: "loaded",
      state: expect.objectContaining({ articleImportApplied: true }),
      articleSearch
    }));
  });

  it("scopes drafts to the signed-in user", () => {
    const state = createVehicleCreateWizardState({ ...emptyVehicle, name: "Privater Entwurf" });

    expect(saveVehicleCreateDraft(state, "alice")).toEqual({ kind: "saved" });
    expect(loadVehicleCreateDraft("bob")).toEqual({ kind: "empty" });
    expect(loadVehicleCreateDraft("alice")).toEqual(expect.objectContaining({ kind: "loaded" }));
  });

  it("rejects malformed state and keeps storage failures non-blocking", () => {
    localStorage.setItem(vehicleCreateDraftKey("local"), JSON.stringify({
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
    expect(loadVehicleCreateDraft("local", storage)).toEqual({ kind: "error" });
    expect(saveVehicleCreateDraft(createVehicleCreateWizardState(emptyVehicle), "local", null, storage))
      .toEqual({ kind: "error" });
    expect(clearVehicleCreateDraft("local", storage)).toEqual({ kind: "error" });
  });

  it("keeps draft operations non-blocking when browser storage access itself is denied", () => {
    const descriptor = Object.getOwnPropertyDescriptor(globalThis, "localStorage");
    Object.defineProperty(globalThis, "localStorage", {
      configurable: true,
      get: () => { throw new DOMException("denied", "SecurityError"); }
    });
    try {
      expect(loadVehicleCreateDraft()).toEqual({ kind: "error" });
      expect(saveVehicleCreateDraft(createVehicleCreateWizardState(emptyVehicle))).toEqual({ kind: "error" });
      expect(clearVehicleCreateDraft()).toEqual({ kind: "error" });
    } finally {
      if (descriptor) Object.defineProperty(globalThis, "localStorage", descriptor);
      else delete (globalThis as { localStorage?: Storage }).localStorage;
    }
  });
});
