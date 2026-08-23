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
