import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { api, ApiError } from "../../shared/api";
import type {
  DigitalCenterCapabilities,
  DigitalCenterReadSession,
  DigitalCenterSessionMessage,
  DigitalCenterSummary,
  DigitalCenterWorkItem,
  DigitalCenterWriteConfirmation,
  DigitalCenterWritePreview,
  ECoSLiveStatus
} from "./digitalCenterModel";
import { DigitalCentersView } from "./DigitalCentersView";

describe("Digital Centers operational journeys", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    window.history.replaceState(null, "", "/digital-centers");
    vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockImplementation(() => null);
    mockWorkspaceAPI();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("uses application-owned selection controls", async () => {
    const { container } = render(<DigitalCentersView roles={["Admin"]} />);

    await screen.findByRole("heading", { name: "Digitalzentralen" });
    expect(container.ownerDocument.querySelector("select")).not.toBeInTheDocument();
  });

  it("creates a read session without changing Settings or calling a write API", async () => {
    const user = userEvent.setup();
    render(<DigitalCentersView roles={["Admin"]} />);

    await user.click(await screen.findByRole("button", { name: "Daten lesen" }));
    expect(await screen.findByRole("cell", { name: "BR 218" })).toBeInTheDocument();

    expect(api.startDigitalCenterReadSession).toHaveBeenCalledWith("ecos");
    expect(api.updateDigitalSettings).not.toHaveBeenCalled();
    expect(api.previewDigitalCenterWrite).not.toHaveBeenCalled();
    expect(api.confirmDigitalCenterWrite).not.toHaveBeenCalled();
    for (const label of ["OK", "Abweichung", "Fehlt in Zentrale", "Neu", "Konflikt"]) {
      expect(screen.getAllByRole("cell", { name: label }).length).toBeGreaterThan(0);
    }
  });

  it("preserves search, quick and advanced filters, page, and page size after detail closes", async () => {
    const user = userEvent.setup();
    render(<DigitalCentersView roles={["Admin"]} />);
    await user.click(await screen.findByRole("button", { name: "Daten lesen" }));
    await screen.findByRole("cell", { name: "BR 218" });

    await user.type(screen.getByRole("searchbox", { name: "Lok suchen" }), "BR");
    await user.click(screen.getByRole("button", { name: "Abweichung filtern" }));
    await user.click(screen.getByRole("button", { name: "Weitere Filter" }));
    await user.click(screen.getByRole("button", { name: "Konflikt filtern" }));
    await user.click(screen.getByRole("button", { name: "Zeilen pro Seite" }));
    await user.click(screen.getByRole("option", { name: "25" }));
    await user.click(screen.getByRole("button", { name: "Nächste Seite" }));
    await user.click(screen.getByRole("button", { name: "BR 218 vergleichen" }));
    await screen.findByRole("dialog", { name: "Lok-Vergleich BR 218" });
    await user.click(screen.getByRole("button", { name: "Vergleich schließen" }));

    expect(screen.getByRole("searchbox", { name: "Lok suchen" })).toHaveValue("BR");
    expect(screen.getByRole("button", { name: "Konflikt filtern" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "Zeilen pro Seite" })).toHaveTextContent("25");
    expect(screen.getByText("26–50 von 60")).toBeInTheDocument();
  });

  it("disables unsupported operations and explains each capability boundary", async () => {
    vi.mocked(api.digitalCenterWorkspace).mockResolvedValue({
      centers: [centerFixture({ capabilities: capabilitiesFixture(false) })]
    });
    const user = userEvent.setup();
    render(<DigitalCentersView roles={["Admin"]} />);
    await screen.findByRole("button", { name: "ESU ECoS0 LoksAktiv" });

    expect(screen.getByRole("button", { name: "Live-Monitor nicht verfügbar" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Daten lesen" })).toBeDisabled();
    expect(screen.getByText("Lesen wird von dieser Digitalzentrale nicht unterstützt.")).toBeInTheDocument();
    expect(screen.getByText("Live-Monitoring wird von dieser Digitalzentrale nicht unterstützt."))
      .toBeInTheDocument();
    expect(screen.getByText("Diese Digitalzentrale unterstützt keine Schreibbefehle.")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Digitalzentrale konfigurieren" }));
    expect(window.location.pathname + window.location.search).toBe("/settings?tab=digital");
  });

  it("starts and stops live monitoring without continuing to poll a stopped monitor", async () => {
    const user = userEvent.setup();
    render(<DigitalCentersView roles={["Admin"]} />);

    await user.click(await screen.findByRole("button", { name: "Live-Monitor starten" }));
    expect(await screen.findByRole("button", { name: "Live-Monitor stoppen" })).toBeEnabled();
    expect(screen.getByText("Letzte Ereignisse")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Live-Monitor stoppen" }));
    expect(await screen.findByRole("button", { name: "Live-Monitor starten" })).toBeEnabled();

    const callsAtStop = vi.mocked(api.digitalCenterLiveStatus).mock.calls.length;
    await new Promise((resolve) => setTimeout(resolve, 1100));
    expect(api.digitalCenterLiveStatus).toHaveBeenCalledTimes(callsAtStop);
  });

  it("marks connection loss as interrupted and removes stale live samples", async () => {
    vi.mocked(api.digitalCenterLiveStatus)
      .mockResolvedValueOnce(liveStatusFixture({ state: "running", connected: true }))
      .mockRejectedValueOnce(new Error("Verbindung verloren"));
    render(<DigitalCentersView roles={["Admin"]} />);

    expect(await screen.findByText("Letzte Ereignisse")).toBeInTheDocument();
    expect(await screen.findByText("Verbindung unterbrochen", {}, { timeout: 2000 })).toBeInTheDocument();
    expect(screen.queryByText("Letzte Ereignisse")).not.toBeInTheDocument();
    expect(screen.queryByRole("img", { name: "Live-Puls der Digitalzentrale" })).not.toBeInTheDocument();
  });

  it("shows safe diagnosis fields and actionable station messages", async () => {
    const user = userEvent.setup();
    render(<DigitalCentersView roles={["Admin"]} />);
    await user.click(await screen.findByRole("button", { name: "Daten lesen" }));
    await waitFor(() => expect(api.digitalCenterSessionMessages).toHaveBeenCalled());

    await user.click(screen.getByRole("tab", { name: "Diagnose" }));
    expect(screen.getByText("Unterstützte Funktionen")).toBeInTheDocument();
    expect(screen.getByText("Lesen: unterstützt")).toBeInTheDocument();
    expect(screen.getByText("Schreiben: unterstützt")).toBeInTheDocument();
    expect(screen.getAllByText("center.local:15471").length).toBeGreaterThan(0);
    expect(screen.getByText("Latenz").nextElementSibling).toHaveTextContent("Nicht gemessen");
    expect(screen.getByText("Protokollfehler").nextElementSibling).toHaveTextContent("1");
    expect(screen.getByText("Letzte erfolgreiche Kommunikation")).toBeInTheDocument();
    expect(screen.queryByText(/request\(|reply\(|authorization|password/i)).not.toBeInTheDocument();

    await user.click(screen.getByRole("tab", { name: "Meldungen" }));
    const message = screen.getByRole("article", { name: "Warnung von ESU ECoS" });
    expect(within(message).getByText("Eine Lokmeldung konnte nicht gelesen werden.")).toBeInTheDocument();
    expect(within(message).getByText("Verbindung prüfen und Daten erneut lesen.")).toBeInTheDocument();
    expect(within(message).getByText(/\d{2}:\d{2}:\d{2}/)).toBeInTheDocument();
  });

  it("previews field changes and requires an explicit confirmation before verified success", async () => {
    const user = userEvent.setup();
    render(<DigitalCentersView roles={["Admin"]} />);
    await openComparison(user);

    await user.click(screen.getByRole("button", { name: "Schreibvorschau erstellen" }));
    expect(await screen.findByText("RailKeeper → Digitalzentrale")).toBeInTheDocument();
    const previewRegion = screen.getByRole("region", { name: "Schreibvorschau" });
    expect(within(previewRegion).getByRole("cell", { name: "Alte Lok" })).toBeInTheDocument();
    expect(within(previewRegion).getByRole("cell", { name: "BR 218" })).toBeInTheDocument();
    const consent = screen.getByRole("checkbox", {
      name: "Ich bestätige, dass die angezeigten Werte in die Digitalzentrale geschrieben werden."
    });
    expect(consent.closest("label")).toHaveClass("app-checkbox", "digital-write-confirmation");
    const confirm = screen.getByRole("button", { name: "In die Digitalzentrale schreiben" });
    expect(confirm).toBeDisabled();
    await user.click(consent);
    expect(confirm).toBeEnabled();
    await user.click(confirm);

    expect(await screen.findByText("Schreiben verifiziert")).toBeInTheDocument();
    expect(api.previewDigitalCenterWrite).toHaveBeenCalledWith("session-1", "item-ok", {
      fields: ["name", "address"],
      operation: "update"
    });
    expect(api.confirmDigitalCenterWrite).toHaveBeenCalledWith("session-1", "item-ok", {
      token: "public-grant",
      confirm: true,
      fields: ["name", "address"],
      operation: "update"
    });
  });

  it("keeps focus in the dialog after verified success and restores the trigger on Escape", async () => {
    const user = userEvent.setup();
    render(<DigitalCentersView roles={["Admin"]} />);
    await openComparison(user);
    const trigger = screen.getByRole("button", { name: "BR 218 vergleichen" });
    await user.click(screen.getByRole("button", { name: "Schreibvorschau erstellen" }));
    await user.click(await screen.findByRole("checkbox"));
    await user.click(screen.getByRole("button", { name: "In die Digitalzentrale schreiben" }));

    await screen.findByText("Schreiben verifiziert");
    expect(screen.getByRole("button", { name: "Vergleich schließen" })).toHaveFocus();
    await user.keyboard("{Escape}");

    expect(screen.queryByRole("dialog", { name: /Lok-Vergleich/ })).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it("keeps the station locomotive count independent from work-list filters", async () => {
    vi.mocked(api.digitalCenterWorkItems).mockImplementation(async (_sessionID, filter) => ({
      items: filter.query ? workItems.slice(0, 4) : workItems,
      page: filter.page,
      pageSize: filter.pageSize,
      total: filter.query ? 4 : 42,
      totalPages: filter.query ? 1 : 5
    }));
    const user = userEvent.setup();
    render(<DigitalCentersView roles={["Admin"]} />);
    await user.click(await screen.findByRole("button", { name: "Daten lesen" }));
    expect(await screen.findByRole("button", { name: "ESU ECoS42 LoksAktiv" })).toBeInTheDocument();

    await user.type(screen.getByRole("searchbox", { name: "Lok suchen" }), "V 60 1059");
    await waitFor(() => expect(api.digitalCenterWorkItems).toHaveBeenLastCalledWith(
      readSession.id,
      expect.objectContaining({ query: "V 60 1059" })
    ));

    expect(screen.getByRole("button", { name: "ESU ECoS42 LoksAktiv" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "ESU ECoS4 LoksAktiv" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Daten lesen" }));
    await waitFor(() => expect(api.startDigitalCenterReadSession).toHaveBeenCalledTimes(2));
    expect(screen.getByRole("button", { name: "ESU ECoS42 LoksAktiv" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "ESU ECoS4 LoksAktiv" })).not.toBeInTheDocument();
  });

  it("invalidates a conflicting grant and requires a fresh read and preview", async () => {
    vi.mocked(api.confirmDigitalCenterWrite).mockRejectedValueOnce(
      new ApiError("Freigabe nicht mehr gültig", "digital_center_write_conflict", 409)
    );
    const user = userEvent.setup();
    render(<DigitalCentersView roles={["Admin"]} />);
    await openComparison(user);
    await user.click(screen.getByRole("button", { name: "Schreibvorschau erstellen" }));
    await screen.findByRole("checkbox", {
      name: "Ich bestätige, dass die angezeigten Werte in die Digitalzentrale geschrieben werden."
    });
    await user.click(screen.getByRole("checkbox"));
    await user.click(screen.getByRole("button", { name: "In die Digitalzentrale schreiben" }));

    expect(await screen.findByRole("alert", { name: "Schreibkonflikt" }))
      .toHaveTextContent("Daten erneut lesen und eine neue Schreibvorschau erstellen");
    expect(screen.queryByRole("dialog", { name: /Lok-Vergleich/ })).not.toBeInTheDocument();
    expect(screen.getByText("Noch keine Lokdaten gelesen")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Daten lesen" }));
    expect(api.startDigitalCenterReadSession).toHaveBeenCalledTimes(2);
  });

  it("shows an unknown write result and a failed live restart without claiming success", async () => {
	vi.mocked(api.confirmDigitalCenterWrite).mockResolvedValueOnce(confirmationFixture({
	  applied: false,
	  verified: false,
	  result: "unknown",
	  message: "Der Schreibstatus ist unbekannt.",
	  liveMonitor: { wasRunning: true, restarted: false }
	}));
	const user = userEvent.setup();
	render(<DigitalCentersView roles={["Admin"]} />);
	await openComparison(user);
	await user.click(screen.getByRole("button", { name: "Schreibvorschau erstellen" }));
	await user.click(await screen.findByRole("checkbox"));
	await user.click(screen.getByRole("button", { name: "In die Digitalzentrale schreiben" }));

	expect(await screen.findByText("Schreibstatus unbekannt")).toBeInTheDocument();
	expect(screen.getByText(/Live-Monitoring konnte nicht automatisch neu gestartet/)).toBeInTheDocument();
	expect(screen.queryByText("Schreiben verifiziert")).not.toBeInTheDocument();
  });

  it("keeps confirmation disabled when the preview grant is expired", async () => {
    vi.mocked(api.previewDigitalCenterWrite).mockResolvedValueOnce({
      ...writePreview,
      expiresAt: "2020-01-01T00:00:00Z"
    });
    const user = userEvent.setup();
    render(<DigitalCentersView roles={["Admin"]} />);
    await openComparison(user);
    await user.click(screen.getByRole("button", { name: "Schreibvorschau erstellen" }));
    await user.click(await screen.findByRole("checkbox"));

    expect(screen.getByRole("button", { name: "In die Digitalzentrale schreiben" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "In die Digitalzentrale schreiben" }))
      .toHaveAttribute("title", "Die Schreibfreigabe ist nicht mehr gültig.");
    expect(api.confirmDigitalCenterWrite).not.toHaveBeenCalled();
  });

  it("presents a verification mismatch as a failure, never as success", async () => {
    vi.mocked(api.confirmDigitalCenterWrite).mockResolvedValueOnce(confirmationFixture({
      applied: true,
      verified: false,
      result: "verification_failed",
      message: "Die Zentrale meldet weiterhin den alten Namen."
    }));
    const user = userEvent.setup();
    render(<DigitalCentersView roles={["Admin"]} />);
    await openComparison(user);
    await user.click(screen.getByRole("button", { name: "Schreibvorschau erstellen" }));
    await user.click(await screen.findByRole("checkbox"));
    await user.click(screen.getByRole("button", { name: "In die Digitalzentrale schreiben" }));

    expect(await screen.findByText("Schreibprüfung fehlgeschlagen")).toBeInTheDocument();
    expect(screen.getByText("Die Zentrale meldet weiterhin den alten Namen.")).toBeInTheDocument();
    expect(screen.queryByText("Schreiben verifiziert")).not.toBeInTheDocument();
  });
});

async function openComparison(user: ReturnType<typeof userEvent.setup>) {
  await user.click(await screen.findByRole("button", { name: "Daten lesen" }));
  await user.click(await screen.findByRole("button", { name: "BR 218 vergleichen" }));
  await screen.findByRole("dialog", { name: "Lok-Vergleich BR 218" });
}

function mockWorkspaceAPI() {
  vi.spyOn(api, "digitalCenterWorkspace").mockResolvedValue({ centers: [centerFixture()] });
  vi.spyOn(api, "digitalCenterLiveStatus").mockResolvedValue(liveStatusFixture());
  vi.spyOn(api, "startDigitalCenterReadSession").mockResolvedValue(readSession);
  vi.spyOn(api, "digitalCenterWorkItems").mockImplementation(async (_sessionID, filter) => ({
    items: workItems,
    page: filter.page,
    pageSize: filter.pageSize,
    total: 60,
    totalPages: 3
  }));
  vi.spyOn(api, "digitalCenterWorkItem").mockImplementation(async (_sessionID, itemID) =>
    workItems.find((item) => item.id === itemID) ?? workItems[0]
  );
  vi.spyOn(api, "digitalCenterSessionMessages").mockResolvedValue({ messages: sessionMessages });
  vi.spyOn(api, "startDigitalCenterLive").mockResolvedValue(
    liveStatusFixture({ state: "running", connected: true })
  );
  vi.spyOn(api, "stopDigitalCenterLive").mockResolvedValue(liveStatusFixture());
  vi.spyOn(api, "previewDigitalCenterWrite").mockResolvedValue(writePreview);
  vi.spyOn(api, "confirmDigitalCenterWrite").mockResolvedValue(confirmationFixture());
  vi.spyOn(api, "updateDigitalSettings");
}

const allCapabilities = capabilitiesFixture(true);
const readSession: DigitalCenterReadSession = {
  id: "session-1",
  provider: "ecos",
  state: "ready",
  host: "center.local",
  port: 15471,
  capabilities: allCapabilities,
  readStartedAt: "2026-08-21T10:00:00Z",
  readCompletedAt: "2026-08-21T10:00:01Z",
  createdByUserId: "admin-1",
  createdAt: "2026-08-21T10:00:00Z",
  updatedAt: "2026-08-21T10:00:01Z"
};

const workItems: DigitalCenterWorkItem[] = [
  workItemFixture({ id: "item-ok", name: "BR 218", compareStatus: "ok" }),
  workItemFixture({ id: "item-deviation", name: "ICE 3", compareStatus: "deviation" }),
  workItemFixture({ id: "item-missing", name: "BR 103", compareStatus: "missing" }),
  workItemFixture({ id: "item-new", name: "BR 50", compareStatus: "new" }),
  workItemFixture({ id: "item-conflict", name: "BR 232", compareStatus: "conflict" })
];

const sessionMessages: DigitalCenterSessionMessage[] = [{
  id: "message-1",
  sessionId: readSession.id,
  severity: "warning",
  code: "parse.failed",
  message: "Eine Lokmeldung konnte nicht gelesen werden.",
  nextAction: "Verbindung prüfen und Daten erneut lesen.",
  createdAt: "2026-08-21T10:00:01Z"
}];

const writePreview: DigitalCenterWritePreview = {
  sessionId: readSession.id,
  itemId: "item-ok",
  provider: "ecos",
  objectId: "3",
  operation: "update",
  direction: "railkeeper_to_center",
  fields: ["name", "address"],
  changes: [
    { field: "name", current: "Alte Lok", desired: "BR 218" },
    { field: "address", current: "3", desired: "18" }
  ],
  token: "public-grant",
  expiresAt: "2099-08-21T10:10:00Z"
};

function capabilitiesFixture(enabled: boolean): DigitalCenterCapabilities {
  return {
    testConnection: enabled,
    readLocomotives: enabled,
    liveMonitor: enabled,
    writeLocomotives: enabled,
    writeCVs: false,
    diagnose: enabled
  };
}

function centerFixture(overrides: Partial<DigitalCenterSummary> = {}): DigitalCenterSummary {
  return {
    provider: "ecos",
    name: "ESU ECoS",
    active: true,
    selected: true,
    host: "center.local",
    port: 15471,
    capabilities: allCapabilities,
    transports: [{ id: "ecos_tcp", status: "available", capabilities: allCapabilities }],
    ...overrides
  };
}

function workItemFixture(overrides: Partial<DigitalCenterWorkItem> = {}): DigitalCenterWorkItem {
  return {
    id: "item-ok",
    sessionId: readSession.id,
    centerObjectId: "3",
    vehicleId: "vehicle-1",
    name: "BR 218",
    decoderAddress: 18,
    protocol: "DCC",
    compareStatus: "ok",
    stationStatus: "read",
    center: { objectId: 3, name: "Alte Lok", decoderAddress: 3, protocol: "DCC" },
    railkeeper: { vehicleId: "vehicle-1", name: "BR 218", decoderAddress: 18, protocol: "DCC" },
    proposed: {},
    conflicts: [],
    createdAt: "2026-08-21T10:00:01Z",
    updatedAt: "2026-08-21T10:00:01Z",
    ...overrides
  };
}

function liveStatusFixture(overrides: Partial<ECoSLiveStatus> = {}): ECoSLiveStatus {
  return {
    provider: "ecos",
    connected: false,
    state: "stopped",
    host: "center.local",
    port: 15471,
    startedAt: "2026-08-21T10:00:00Z",
    lastSeenAt: "2026-08-21T10:00:01Z",
    lastMessage: "Antwort empfangen",
    blocksReceived: 4,
    repliesReceived: 8,
    eventsReceived: 2,
    subscriptionCommands: ["request(1, view)"],
    pulseSamples: [{ at: "2026-08-21T10:00:01Z", repliesPerSecond: 8 }],
    recentEvents: [{ at: "2026-08-21T10:00:01Z", kind: "event", message: "Antwort", protocol: "ECoS" }],
    diagnosis: {
      connectionState: "stopped",
      lastSuccessfulCommunication: "2026-08-21T10:00:01Z",
      passive: true
    },
    message: "Live-Monitor gestoppt",
    ...overrides
  };
}

function confirmationFixture(
  overrides: Partial<DigitalCenterWriteConfirmation> = {}
): DigitalCenterWriteConfirmation {
  return {
    sessionId: readSession.id,
    itemId: "item-ok",
    provider: "ecos",
    objectId: "3",
    operation: "update",
    direction: "railkeeper_to_center",
    fields: ["name", "address"],
    applied: true,
    verified: true,
    result: "verified",
    message: "Änderungen wurden gelesen und bestätigt.",
	liveMonitor: { wasRunning: false, restarted: false },
    ...overrides
  };
}
