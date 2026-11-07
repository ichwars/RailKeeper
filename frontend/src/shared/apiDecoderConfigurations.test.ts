import { afterEach, describe, expect, it, vi } from "vitest";
import { api, type DecoderCVSnapshot } from "./api";

describe("private decoder archive API", () => {
  afterEach(() => vi.unstubAllGlobals());

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
});
