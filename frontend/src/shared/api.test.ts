import { afterEach, describe, expect, it, vi } from "vitest";

import { api, ApiError } from "./api";

describe("vehicle set image API", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("allows the same upload timeout as vehicle images", async () => {
    const timeout = vi.spyOn(window, "setTimeout");
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: () => Promise.resolve({ id: "set-1" })
    }));

    await api.uploadVehicleSetImage("set-1", new File(["image"], "set.png", { type: "image/png" }));

    expect(timeout).toHaveBeenCalledWith(expect.any(Function), 30000);
  });
});

describe("structured API problems", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("preserves safe problem details for conflict handling", async () => {
	vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
	  ok: false,
	  status: 409,
	  statusText: "Conflict",
	  json: () => Promise.resolve({
		error: "digital_center_address_conflict",
		message: "Address already used.",
		details: { objectId: 2002, name: "Other", decoderAddress: 18 }
	  })
	}));

	const error = await api.digitalCenterWorkspace().catch((caught: unknown) => caught);

	expect(error).toBeInstanceOf(ApiError);
	expect((error as ApiError).details).toEqual({ objectId: 2002, name: "Other", decoderAddress: 18 });
  });
});

describe("ECoS backup upload API", () => {
  afterEach(() => { vi.unstubAllGlobals(); document.cookie = "rk_csrf=; Max-Age=0; path=/"; });

  it("sends raw backup bytes with CSRF protection and the confirmed configured station", async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 201, json: () => Promise.resolve({}) });
    vi.stubGlobal("fetch", fetch);
    document.cookie = "rk_csrf=backup-csrf; path=/";
    const file = new File(["encrypted backup"], "bestand.eco");
    await api.importECoSBackup(file, { confirmStation: true, stationHost: "ecos.local" });
    expect(fetch).toHaveBeenCalledWith("/api/v1/digital-centers/ecos/backup-import", expect.objectContaining({
      method: "POST", credentials: "include", body: file,
      headers: expect.objectContaining({ "Content-Type": "application/octet-stream",
        "X-CSRF-Token": "backup-csrf", "X-RailKeeper-Confirm-Station": "true",
        "X-RailKeeper-Backup-Station": "ecos.local" })
    }));
    expect(fetch.mock.calls[0][1].headers).not.toHaveProperty("X-RailKeeper-Backup-Filename");
  });
});

describe("ECoS database read API", () => {
  afterEach(() => { vi.useRealTimers(); vi.unstubAllGlobals(); });

  it("waits for the database and internal backup beyond thirty seconds", async () => {
    vi.useFakeTimers();
    const snapshot = { id: "item-1" };
    vi.stubGlobal("fetch", vi.fn().mockImplementation((_url: string, init: RequestInit) =>
      new Promise((resolve, reject) => {
        init.signal?.addEventListener("abort", () => reject(new DOMException("Aborted", "AbortError")));
        window.setTimeout(() => resolve({ ok: true, status: 200, json: async () => snapshot }), 40000);
      })));

    const result = api.readDigitalCenterCVBatch("session-1", "item-1", 1, { source: "database" });
    const completed = expect(result).resolves.toEqual(snapshot);
    await vi.advanceTimersByTimeAsync(40000);
    await completed;
  });
});
