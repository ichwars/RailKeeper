import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { BugReportCapability, BugReportSubmissionInput } from "../../shared/apiBugReports";
import { BugReportDialog } from "./BugReportDialog";

const capability: BugReportCapability = {
  enabled: true,
  fallbackUrl: "https://github.com/ichwars/RailKeeper/issues/new",
  maxScreenshotBytes: 5 * 1024 * 1024,
  acceptedScreenshotTypes: ["image/png", "image/jpeg", "image/webp"],
  includedDiagnosticCategories: ["railkeeper", "runtime", "browser", "roles", "clientErrors"],
  excludedDiagnosticCategories: ["inventoryData", "inventoryCounts", "connectedDevices", "credentials"],
  contactAvailable: true,
  serverDiagnostics: {
    railkeeperVersion: "0.1.20",
    deploymentType: "native",
    goVersion: "go1.26.0",
    operatingSystem: "windows",
    architecture: "amd64",
    schemaVersion: 67,
    enabledDigitalProviderTypes: []
  }
};

describe("BugReportDialog", () => {
  beforeEach(() => {
    vi.spyOn(crypto, "randomUUID").mockReturnValue("8b1d6340-0f66-472e-9954-58f1e9d88c88");
  });

  it("focuses and explains an invalid required description", async () => {
    const user = userEvent.setup();
    render(<BugReportDialog capability={capability} initialArea="vehicles" roles={["Viewer"]}
      onClose={vi.fn()} onSubmit={vi.fn()} />);

    await user.type(screen.getByLabelText(/Was ist passiert|What happened/i), "zu kurz");
    await user.click(screen.getByRole("button", { name: /Fehlerbericht senden|Send bug report/i }));

    const description = screen.getByLabelText(/Was ist passiert|What happened/i);
    expect(description).toHaveFocus();
    expect(screen.getByRole("alert")).toHaveTextContent(/mindestens 20|at least 20/i);
  });

  it("explains the area selection, keeps its menu inside the dialog and submits the chosen area", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockResolvedValue({
      reportId: "8b1d6340-0f66-472e-9954-58f1e9d88c88", issueNumber: 184,
      issueUrl: "https://github.com/ichwars/RailKeeper/issues/184", status: "complete", warning: ""
    });
    render(<BugReportDialog capability={capability} initialArea="overview" roles={["Viewer"]}
      onClose={vi.fn()} onSubmit={onSubmit} />);
    const dialog = screen.getByRole("dialog", { name: /Fehler melden|Report issue/i });
    const selector = screen.getByRole("button", { name: /Wo tritt der Fehler auf|Where does the problem occur/i });
    expect(selector).toHaveAccessibleDescription(/Sonstiges|Other/);
    expect(selector).toHaveAttribute("aria-required", "true");
    await user.click(selector);
    expect(dialog).toContainElement(screen.getByRole("listbox"));
    expect(screen.getAllByRole("option")).toHaveLength(8);
    await user.click(screen.getByRole("option", { name: /Import und Export|Import and export/i }));
    expect(selector).toHaveTextContent(/Import und Export|Import and export/i);
    expect(selector).toHaveFocus();
    await user.type(screen.getByLabelText(/Was ist passiert|What happened/i),
      "Der Import lässt sich nach der Dateiauswahl nicht starten.");
    await user.click(screen.getByRole("button", { name: /Fehlerbericht senden|Send bug report/i }));
    await waitFor(() => expect(onSubmit).toHaveBeenCalledOnce());
    expect(onSubmit.mock.calls[0][0]).toMatchObject({ area: "importExport", clientDiagnostics: { area: "importExport" } });
  });

  it("submits current diagnostics and shows the created issue", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockResolvedValue({
      reportId: "8b1d6340-0f66-472e-9954-58f1e9d88c88",
      issueNumber: 184,
      issueUrl: "https://github.com/ichwars/RailKeeper/issues/184",
      status: "complete",
      warning: ""
    });
    render(<BugReportDialog capability={capability} initialArea="vehicles" roles={["Viewer"]}
      onClose={vi.fn()} onSubmit={onSubmit} />);
    await user.type(screen.getByLabelText(/Was ist passiert|What happened/i),
      "Das Vorschaubild verschwindet nach dem Speichern.");
    await user.click(screen.getByRole("button", { name: /Fehlerbericht senden|Send bug report/i }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    const input = onSubmit.mock.calls[0][0] as BugReportSubmissionInput;
    expect(input.reportId).toBe("8b1d6340-0f66-472e-9954-58f1e9d88c88");
    expect(input.area).toBe("vehicles");
    expect(input.clientDiagnostics?.roles).toEqual(["Viewer"]);
    expect(screen.getByRole("link", { name: /#184/ })).toHaveAttribute("href", input.reportId &&
      "https://github.com/ichwars/RailKeeper/issues/184");
  });

  it("retries a partial report with the same report id", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn()
      .mockResolvedValueOnce({
        reportId: "8b1d6340-0f66-472e-9954-58f1e9d88c88", issueNumber: 184,
        issueUrl: "https://github.com/ichwars/RailKeeper/issues/184", status: "partial",
        warning: "Screenshot fehlt"
      })
      .mockResolvedValueOnce({
        reportId: "8b1d6340-0f66-472e-9954-58f1e9d88c88", issueNumber: 184,
        issueUrl: "https://github.com/ichwars/RailKeeper/issues/184", status: "complete", warning: ""
      });
    render(<BugReportDialog capability={capability} initialArea="vehicles" roles={["Viewer"]}
      onClose={vi.fn()} onSubmit={onSubmit} />);
    await user.type(screen.getByLabelText(/Was ist passiert|What happened/i),
      "Das Vorschaubild verschwindet nach dem Speichern.");
    await user.click(screen.getByRole("button", { name: /Fehlerbericht senden|Send bug report/i }));
    await screen.findByText("Screenshot fehlt");
    await user.click(screen.getByRole("button", { name: /Erneut versuchen|Retry/i }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(2));
    expect(onSubmit.mock.calls[0][0].reportId).toBe(onSubmit.mock.calls[1][0].reportId);
  });

  it("discloses the prepared values, explicit exclusions, and the security-reporting path", async () => {
    const user = userEvent.setup();
    render(<BugReportDialog capability={capability} initialArea="settings" roles={["Admin"]}
      onClose={vi.fn()} onSubmit={vi.fn()} />);
    await user.click(screen.getByText(/Welche Daten|Which data/i));
    expect(screen.getByText(/"railkeeperVersion": "0.1.20"/)).toBeInTheDocument();
    expect(screen.getByText(/"roles": \[/)).toBeInTheDocument();
    expect(screen.getByText(/"Admin"/)).toBeInTheDocument();
    expect(screen.getByText(/Bestandszahlen|inventory counts/i)).toBeInTheDocument();
    expect(screen.getByText(/verbundene Geräte|connected devices/i)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Sicherheitslücke privat melden|Report a vulnerability privately/i }))
      .toHaveAttribute("href", "https://github.com/ichwars/RailKeeper/security/advisories/new");
  });
});
