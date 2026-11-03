import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { BugReportCapability } from "../../shared/apiBugReports";
import { BugReportDialog } from "./BugReportDialog";
import { clientErrorBuffer, installClientErrorCapture } from "./clientErrorBuffer";
import { readDiagnosticPreference, saveDiagnosticPreference } from "./bugReportPreferences";

const capability: BugReportCapability = {
  enabled: true, fallbackUrl: "https://github.com/ichwars/RailKeeper/issues/new",
  screenshotAvailable: false, maxScreenshotBytes: 5 * 1024 * 1024,
  acceptedScreenshotTypes: [], includedDiagnosticCategories: [], excludedDiagnosticCategories: [],
  contactAvailable: false,
  serverDiagnostics: {
    railkeeperVersion: "test", deploymentType: "native", goVersion: "test", operatingSystem: "linux",
    architecture: "amd64", schemaVersion: 72, enabledDigitalProviderTypes: []
  }
};

describe("bug report diagnostic opt-out", () => {
  beforeEach(() => {
    localStorage.clear();
    saveDiagnosticPreference(true);
    clientErrorBuffer.clear();
  });

  it("clears errors, stops capture and remembers the choice", () => {
    const stop = installClientErrorCapture();
    window.dispatchEvent(new ErrorEvent("error", { error: new TypeError("private message") }));
    expect(clientErrorBuffer.snapshot()).toHaveLength(1);
    saveDiagnosticPreference(false);
    expect(clientErrorBuffer.snapshot()).toEqual([]);
    window.dispatchEvent(new ErrorEvent("error", { error: new Error("another private message") }));
    expect(clientErrorBuffer.snapshot()).toEqual([]);
    expect(readDiagnosticPreference()).toBe(false);
    stop();
  });

  it("does not load diagnostics or send them after a remembered opt-out", async () => {
    saveDiagnosticPreference(false);
    const user = userEvent.setup();
    const load = vi.fn();
    const submit = vi.fn().mockResolvedValue({ reportId: "test", issueNumber: 1,
      issueUrl: "https://github.com/ichwars/RailKeeper/issues/1", status: "complete", warning: "" });
    render(<BugReportDialog capability={capability} initialArea="vehicles" roles={["Viewer"]}
      onLoadDiagnostics={load} onClose={vi.fn()} onSubmit={submit} />);
    const toggle = screen.getByRole("checkbox", { name: /Technische Diagnosedaten|technical diagnostics/i });
    expect(toggle).not.toBeChecked();
    expect(load).not.toHaveBeenCalled();
    await user.type(screen.getByLabelText(/Was ist passiert|What happened/i), "Ein reproduzierbarer Fehler beim Speichern.");
    await user.click(screen.getByRole("button", { name: /Fehlerbericht senden|Send bug report/i }));
    await waitFor(() => expect(submit).toHaveBeenCalledOnce());
    expect(submit.mock.calls[0][0]).toMatchObject({ includeDiagnostics: false, clientDiagnostics: null });
  });

  it("can re-enable diagnostics and shows a recoverable collection failure", async () => {
    saveDiagnosticPreference(false);
    const user = userEvent.setup();
    const load = vi.fn().mockRejectedValue(new Error("offline"));
    render(<BugReportDialog capability={capability} roles={["Admin"]}
      onLoadDiagnostics={load} onClose={vi.fn()} onSubmit={vi.fn()} />);
    await user.click(screen.getByRole("checkbox", { name: /Technische Diagnosedaten|technical diagnostics/i }));
    expect(await screen.findByRole("alert")).toHaveTextContent(/Serverdiagnose|Server diagnostics/i);
    expect(load).toHaveBeenCalledOnce();
    expect(readDiagnosticPreference()).toBe(true);
  });
});
