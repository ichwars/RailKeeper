import { afterEach, describe, expect, it, vi } from "vitest";

import { collectBugReportDiagnostics, collectBugReportDiagnosticsWithClientHints } from "./bugReportDiagnostics";
import type { ClientErrorBuffer } from "./clientErrorBuffer";

describe("bug report diagnostics", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("collects the allowlisted presentation and navigation fields", () => {
    Object.defineProperty(window, "innerWidth", { configurable: true, value: 1280 });
    Object.defineProperty(window, "innerHeight", { configurable: true, value: 720 });
    window.history.replaceState({}, "", "/vehicles/vehicle-id?secret=1#fragment");
    vi.stubGlobal("navigator", { userAgent: "Mozilla/5.0 (Windows NT 10.0) Firefox/128.0" });
    const buffer = { snapshot: () => [] } as unknown as ClientErrorBuffer;

    const result = collectBugReportDiagnostics({
      area: "vehicles", roles: ["Viewer"], language: "de", buffer,
    });

    expect(result).toMatchObject({
      browserFamily: "Firefox",
      browserMajor: "128",
      osFamily: "Windows",
      viewportWidth: 1280,
      viewportHeight: 720,
      pathname: "/vehicles",
      area: "vehicles",
      roles: ["Viewer"]
    });
    expect(result).not.toHaveProperty("userAgent");
    expect(result).not.toHaveProperty("username");
  });

  it("prefers bounded user-agent client hints without exposing the raw user agent", async () => {
    vi.stubGlobal("navigator", {
      userAgent: "secret raw user agent",
      userAgentData: {
        brands: [{ brand: "Chromium", version: "140" }],
        platform: "Windows",
        getHighEntropyValues: vi.fn().mockResolvedValue({
          fullVersionList: [{ brand: "Chromium", version: "140.1.2.3" }],
          platformVersion: "15.0.0"
        })
      }
    });
    const buffer = { snapshot: () => [] } as unknown as ClientErrorBuffer;

    const result = await collectBugReportDiagnosticsWithClientHints({
      area: "settings", roles: ["Admin"], language: "de", buffer
    });

    expect(result).toMatchObject({
      browserFamily: "Chrome", browserMajor: "140", osFamily: "Windows", osMajor: "15"
    });
    expect(JSON.stringify(result)).not.toContain("secret raw user agent");
  });
});
