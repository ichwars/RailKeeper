import { describe, expect, it, vi } from "vitest";

import { createBugReportAPI, type BugReportClientDiagnostics } from "./apiBugReports";

const diagnostics: BugReportClientDiagnostics = {
  browserFamily: "Firefox",
  browserMajor: "128",
  osFamily: "Windows",
  osMajor: "",
  viewportWidth: 1440,
  viewportHeight: 900,
  language: "de",
  timeZone: "Europe/Berlin",
  theme: "dark",
  pathname: "/vehicles",
  area: "vehicles",
  roles: ["Viewer"],
  errors: []
};

describe("bug report API", () => {
  it("loads the authenticated capability", async () => {
    const request = vi.fn().mockResolvedValue({ enabled: false });
    const api = createBugReportAPI(request);

    await api.bugReportConfig();

    expect(request).toHaveBeenCalledWith("/bug-reports/config");
  });

  it("submits exact multipart fields and keeps the provided report id", async () => {
    const request = vi.fn().mockResolvedValue({ status: "complete" });
    const api = createBugReportAPI(request);
    const screenshot = new File(["image"], "screenshot.png", { type: "image/png" });

    await api.submitBugReport({
      reportId: "8b1d6340-0f66-472e-9954-58f1e9d88c88",
      area: "vehicles",
      description: "Das Vorschaubild verschwindet nach dem Speichern.",
      expectedBehavior: "Das Bild bleibt sichtbar.",
      reproductionSteps: "Fahrzeug öffnen und speichern.",
      email: "maintainer@example.de",
      clientDiagnostics: diagnostics,
      screenshotConsent: true,
      screenshot
    });

    const [path, init] = request.mock.calls[0] as [string, RequestInit];
    expect(path).toBe("/bug-reports");
    expect(init.method).toBe("POST");
    expect(init.body).toBeInstanceOf(FormData);
    const form = init.body as FormData;
    expect(form.get("reportId")).toBe("8b1d6340-0f66-472e-9954-58f1e9d88c88");
    expect(form.get("clientDiagnostics")).toBe(JSON.stringify(diagnostics));
    expect(form.getAll("screenshot")).toEqual([screenshot]);
    expect(form.get("screenshotConsent")).toBe("true");
  });
  it("omits even supplied diagnostics when inclusion is disabled", async () => {
    const request = vi.fn().mockResolvedValue({ status: "complete" });
    const api = createBugReportAPI(request);
    await api.submitBugReport({
      reportId: "8b1d6340-0f66-472e-9954-58f1e9d88c88", area: "vehicles",
      description: "Ein Fehler mit ausreichender Beschreibung.", expectedBehavior: "", reproductionSteps: "",
      email: "", includeDiagnostics: false, clientDiagnostics: diagnostics, screenshotConsent: false, screenshot: null
    });
    const form = request.mock.calls[0][1].body as FormData;
    expect(form.get("includeDiagnostics")).toBe("false");
    expect(form.has("clientDiagnostics")).toBe(false);
  });

});
