import { describe, expect, it } from "vitest";

import { createBugReportID, areaForPathname, validateBugReportForm } from "./bugReportModel";

describe("bug report model", () => {
  it.each([
    ["/overview", "overview"],
    ["/vehicles/vehicle-id", "vehicles"],
    ["/accessories", "accessories"],
    ["/exhibition", "exhibition"],
    ["/import-export", "importExport"],
    ["/digital-centers", "digitalCenters"],
    ["/settings", "settings"],
    ["/unknown", "other"]
  ])("maps %s to %s", (pathname, expected) => {
    expect(areaForPathname(pathname)).toBe(expected);
  });

  it("requires the description and public screenshot consent", () => {
    const screenshot = new File(["image"], "screen.png", { type: "image/png" });
    expect(validateBugReportForm({
      area: "vehicles",
      description: "zu kurz",
      expectedBehavior: "",
      reproductionSteps: "",
      email: "bad address",
      screenshot,
      screenshotConsent: false
    })).toEqual({
      description: "description_length",
      email: "email_invalid",
      screenshotConsent: "screenshot_consent_required"
    });
  });
});

it("creates a cryptographically random v4 report ID on a plain HTTP origin", () => {
  const source = { getRandomValues: crypto.getRandomValues.bind(crypto) };
  const first = createBugReportID(source);
  expect(first).toMatch(/^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$/);
  expect(createBugReportID(source)).not.toBe(first);
});
