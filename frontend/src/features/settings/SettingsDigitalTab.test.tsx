import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type DigitalCenterSettings, type ECoSLiveStatus } from "../../shared/api";
import { SettingsDigitalTab } from "./SettingsDigitalTab";

const settings = (ecosEnabled = false): DigitalCenterSettings => ({
  provider: "ecos",
  ecos: { enabled: ecosEnabled, host: "192.168.2.151", port: "15471" },
  z21: { enabled: false, host: "", port: "21105" },
  intellibox3: { enabled: false, host: "", port: "21105" },
  cs3: { enabled: false, host: "", port: "80" }
});

const liveStatus: ECoSLiveStatus = {
  provider: "ecos",
  connected: false,
  state: "stopped",
  blocksReceived: 0,
  repliesReceived: 0,
  eventsReceived: 0,
  pulseSamples: [],
  recentEvents: [],
  diagnosis: { connectionState: "stopped", passive: true },
  message: "ECoS-Live-Verbindung inaktiv."
};

describe("SettingsDigitalTab commissioning workflow", () => {
  beforeEach(() => {
    vi.spyOn(api, "digitalSettings").mockResolvedValue(settings());
    vi.spyOn(api, "updateDigitalSettings").mockImplementation(async (input) => input);
    vi.spyOn(api, "getECoSLiveStatus").mockResolvedValue(liveStatus);
    vi.spyOn(api, "testECoSConnection").mockResolvedValue({
      connected: true,
      host: "192.168.2.151",
      port: 15471,
      applicationVersion: "4.3.1",
      message: "ECoS-Verbindung erfolgreich getestet."
    });
  });

  it("requires a successful test before activating the adapter", async () => {
    const user = userEvent.setup();
    render(
      <SettingsDigitalTab
        canManageUsers
        formatDateTime={(value) => value}
        username="admin"
      />
    );

    expect(await screen.findByText("192.168.2.151:15471")).toBeInTheDocument();
    const activate = screen.getByRole("button", { name: "Adapter aktivieren" });
    expect(activate).toBeDisabled();

    await user.click(screen.getByRole("button", { name: "Verbindung testen" }));
    expect(await screen.findByText("Verbindung erfolgreich geprüft")).toBeInTheDocument();
    expect(activate).toBeEnabled();

    await user.click(activate);
    await waitFor(() => {
      expect(api.updateDigitalSettings).toHaveBeenLastCalledWith(
        expect.objectContaining({ ecos: expect.objectContaining({ enabled: true }) })
      );
    });
    expect(await screen.findByText("Adapter und Importbereich wurden aktiviert.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Live-Monitor starten" })).toBeEnabled();
  });

  it("makes the diagnostic and ECoS message tabs interactive", async () => {
    const user = userEvent.setup();
    render(
      <SettingsDigitalTab
        canManageUsers
        formatDateTime={(value) => value}
        username="admin"
      />
    );

    await screen.findByText("192.168.2.151:15471");
    await user.click(screen.getByRole("tab", { name: "Letzte Diagnose" }));
    expect(screen.getByRole("tabpanel")).toHaveTextContent("Dieser Adapter bietet derzeit nur den Verbindungstest.");

    await user.click(screen.getByRole("tab", { name: "ECoS-Meldungen" }));
    expect(screen.getByRole("tabpanel")).toHaveTextContent("ECoS-Live-Verbindung inaktiv.");
    expect(screen.getByRole("button", { name: "Status aktualisieren" })).toBeEnabled();
  });

  it("removes a configured connection only after confirmation", async () => {
    const user = userEvent.setup();
    vi.mocked(api.digitalSettings).mockResolvedValue(settings(true));
    const confirm = vi.spyOn(window, "confirm").mockReturnValue(true);

    render(
      <SettingsDigitalTab
        canManageUsers
        formatDateTime={(value) => value}
        username="admin"
      />
    );

    await screen.findByText("192.168.2.151:15471");
    await user.click(screen.getByRole("button", { name: "Verbindung entfernen" }));

    expect(confirm).toHaveBeenCalledWith(expect.stringContaining("ESU ECoS"));
    await waitFor(() => {
      expect(api.updateDigitalSettings).toHaveBeenLastCalledWith(
        expect.objectContaining({ ecos: { enabled: false, host: "", port: "15471" } })
      );
    });
    expect(await screen.findByText("Adapter-Verbindung wurde entfernt.")).toBeInTheDocument();
  });

  it("runs CS3 read-only diagnostics and exposes the import safety boundary", async () => {
    const user = userEvent.setup();
    vi.mocked(api.digitalSettings).mockResolvedValue({
      ...settings(),
      provider: "cs3",
      cs3: { enabled: true, host: "192.168.2.46", port: "80" }
    });
    vi.spyOn(api, "probeCS3Connection").mockResolvedValue({
      provider: "cs3",
      connected: true,
      host: "192.168.2.46",
      port: 80,
      message: "CS3-Diagnose abgeschlossen.",
      fields: { readOnly: "true", locomotiveCount: "1" },
      commands: [{
        name: "CS3_LOCOMOTIVE_API",
        description: "Read-only Loklisten-API",
        request: "GET /app/api/locos",
        commandHex: "",
        ok: true,
        fields: { readOnly: "true", locomotiveCount: "1" }
      }]
    });

    render(
      <SettingsDigitalTab
        canManageUsers
        formatDateTime={(value) => value}
        username="admin"
      />
    );

    expect(await screen.findByText("192.168.2.46:80")).toBeInTheDocument();
    expect(screen.getByText("Liest Name, Adresse und Protokoll nur lesend in den Vergleichsbereich ein.")).toBeInTheDocument();
    await user.click(screen.getByRole("tab", { name: "Letzte Diagnose" }));
    await user.click(screen.getByRole("button", { name: "Diagnosedaten lesen" }));

    await waitFor(() => {
      expect(api.probeCS3Connection).toHaveBeenCalledWith({ host: "192.168.2.46", port: 80 });
    });
    expect(await screen.findByText("GET /app/api/locos")).toBeInTheDocument();
    expect(screen.getAllByText("true").length).toBeGreaterThan(0);
  });
});
