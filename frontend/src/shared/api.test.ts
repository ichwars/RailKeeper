import { afterEach, describe, expect, it, vi } from "vitest";

import { api } from "./api";

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
