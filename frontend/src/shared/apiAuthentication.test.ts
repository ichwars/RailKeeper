import { afterEach, describe, expect, it, vi } from "vitest";
import { api, ApiError, captureSessionGuard } from "./api";
import { setCSRFToken } from "./apiTransport";

describe("session revalidation ownership", () => {
  afterEach(() => { vi.unstubAllGlobals(); setCSRFToken(""); });

  it.each([401, 403, 500])("invalidates queued work only for confirmed unauthorized session status %s", async status => {
    setCSRFToken("original-owner");
    const isCurrent = captureSessionGuard();
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: false, status, statusText: "Rejected",
      json: async () => ({ error: status === 401 ? "unauthorized" : "session_failed" })
    }));
    await expect(api.session()).rejects.toBeInstanceOf(ApiError);
    expect(isCurrent()).toBe(status !== 401);
  });

  it("refreshes the CSRF owner before an explicit logout retry", async () => {
    setCSRFToken("original-owner");
    const previous = captureSessionGuard();
    const current = { username: "operator", roles: ["Viewer"], csrfToken: "renewed-owner", twoFactorEnabled: false };
    const fetch = vi.fn()
      .mockResolvedValueOnce({ ok: true, status: 200, json: async () => current })
      .mockResolvedValueOnce({ ok: true, status: 204 });
    vi.stubGlobal("fetch", fetch);
    expect(await api.session()).toEqual(current);
    expect(previous()).toBe(false);
    await api.logout();
    expect(fetch).toHaveBeenLastCalledWith("/api/v1/auth/logout", expect.objectContaining({
      method: "POST", headers: expect.objectContaining({ "X-CSRF-Token": "renewed-owner" })
    }));
  });
});
