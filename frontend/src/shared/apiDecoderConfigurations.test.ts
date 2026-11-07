import { afterEach, describe, expect, it, vi } from "vitest";
import { api, ApiError, type DecoderCVSnapshot, type DecoderInstanceInput } from "./api";

describe("private decoder archive API", () => {
  afterEach(() => { vi.unstubAllGlobals(); document.cookie = "rk_csrf=; Max-Age=0; path=/"; });

  it("preserves exact opaque identities in current and historical reads", async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({}) });
    vi.stubGlobal("fetch", fetch);
    await api.decoderConfigurationContext("vehicle/1");
    await api.decoderCVSnapshots();
    await api.decoderCVSnapshots("deleted/1");
    await api.decoderCVSnapshot("snapshot/1");
    expect(fetch.mock.calls.map(([path]) => path)).toEqual([
      "/api/v1/vehicles/vehicle%2F1/decoder-configurations", "/api/v1/decoder-configurations/snapshots",
      "/api/v1/decoder-configurations/snapshots?vehicleId=deleted%2F1", "/api/v1/decoder-configurations/snapshots/snapshot%2F1"
    ]);
    expect(api.decoderSnapshotSourceUrl("snapshot/1", "source/1")).toBe(
      "/api/v1/decoder-configurations/snapshots/snapshot%2F1/sources/source%2F1/download");
  });

  it("keeps unknown values, raw labels and explicitly unavailable originals", async () => {
    const snapshot: DecoderCVSnapshot = {
      id: "snapshot", workspaceId: "group", vehicleReferenceId: "deleted", vehicleName: "Original vehicle",
      vehicleInventoryNumber: "INV-1", profile: " Raw ", installationId: null, kind: "checkpoint",
      scope: "partial_working_state", note: "Partial, not measured", recordedAt: "2026-11-07T03:00:00Z", valueCount: 1,
      hardware: null, payload: { values: [{ id: "cv", vehicleId: "deleted", cvNumber: 257, value: -1,
        indexCV31: 0, indexCV32: 0, decoderProfile: " Raw ", sourceFileId: "missing", sourceLine: 2,
        sourceExportedAt: "2026-10-20T08:30:00", createdAt: "old", updatedAt: "old" }] },
      sources: [{ sourceFileId: "missing", contentState: "unavailable", metadata: {
        id: "missing", vehicleId: "", fileName: "", originalName: "", sizeBytes: 0, createdAt: "", updatedAt: ""
      } }]
    };
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => snapshot });
    vi.stubGlobal("fetch", fetch);
    await expect(api.decoderCVSnapshot("snapshot")).resolves.toEqual(snapshot);
    expect(fetch).toHaveBeenCalledTimes(1);
  });

  it("keeps stable confirmation identities, explicit null dates and CSRF for every writer", async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({}) });
    vi.stubGlobal("fetch", fetch);
    document.cookie = "rk_csrf=decoder-csrf; path=/";
    const instance: DecoderInstanceInput = { id: "a".repeat(32), expectedRevision: "b".repeat(64),
      label: "Physical decoder", manufacturer: "", model: "", serialNumber: "", notes: "" };
    const installation = { id: "c".repeat(32), expectedRevision: instance.expectedRevision,
      decoderId: instance.id, installedAt: null, notes: "" };
    const workspace = { id: "d".repeat(32), expectedRevision: instance.expectedRevision, profile: "Working" };
    const assignment = { expectedRevision: instance.expectedRevision, installationId: null, snapshotId: "e".repeat(32) };
    const checkpoint = { id: "f".repeat(32), expectedRevision: instance.expectedRevision, note: "Partial" };
    const removal = { expectedRevision: instance.expectedRevision, removedAt: null, note: "" };
    const dates = { expectedRevision: instance.expectedRevision, installedAt: null, removedAt: null };
    const root = "/api/v1/vehicles/vehicle%2F1/decoder-configurations";
    const cases = [
      { path: `${root}/instances`, method: "POST", input: instance, run: () => api.createDecoderInstance("vehicle/1", instance) },
      { path: `${root}/instances/${instance.id}`, method: "PUT", input: instance, run: () => api.updateDecoderInstance("vehicle/1", instance) },
      { path: `${root}/installations`, method: "POST", input: installation, run: () => api.installDecoder("vehicle/1", installation) },
      { path: `${root}/installations/episode%2F1/removal`, method: "POST", input: removal, run: () => api.removeDecoder("vehicle/1", "episode/1", removal) },
      { path: `${root}/installations/episode%2F1/dates`, method: "PUT", input: dates, run: () => api.correctDecoderInstallationDates("vehicle/1", "episode/1", dates) },
      { path: `${root}/workspaces`, method: "POST", input: workspace, run: () => api.createDecoderCVWorkspace("vehicle/1", workspace) },
      { path: `${root}/workspaces/group%2F1/assignment`, method: "PUT", input: assignment, run: () => api.assignDecoderCVWorkspace("vehicle/1", "group/1", assignment) },
      { path: `${root}/workspaces/group%2F1/checkpoints`, method: "POST", input: checkpoint, run: () => api.captureDecoderCVCheckpoint("vehicle/1", "group/1", checkpoint) }
    ];
    for (const item of cases) {
      fetch.mockClear();
      await item.run(); await item.run();
      for (const [path, options] of fetch.mock.calls) {
        expect(path).toBe(item.path);
        expect(options.method).toBe(item.method);
        expect(options.headers["X-CSRF-Token"]).toBe("decoder-csrf");
        expect(JSON.parse(options.body)).toEqual(item.input);
      }
    }
  });

  it("exposes stale comparisons without automatically replaying a checkpoint", async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: false, status: 409, statusText: "Conflict",
      json: async () => ({ error: "decoder_conflict", message: "Reload and review." }) });
    vi.stubGlobal("fetch", fetch);
    const input = { id: "a".repeat(32), expectedRevision: "b".repeat(64), note: "Retained draft" };
    await expect(api.captureDecoderCVCheckpoint("vehicle", "group", input)).rejects.toMatchObject<Partial<ApiError>>({
      status: 409, code: "decoder_conflict"
    });
    expect(input.note).toBe("Retained draft");
    expect(fetch).toHaveBeenCalledTimes(1);
  });
});
