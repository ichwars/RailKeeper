import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type {
  DigitalCenterReadSession,
  DigitalCenterSummary,
  DigitalCenterWorkItem,
  ECoSLiveStatus
} from "./digitalCenterModel";
import { DigitalCentersView } from "./DigitalCentersView";
import type { useDigitalCentersWorkspace } from "./useDigitalCentersWorkspace";

const { workspaceHook } = vi.hoisted(() => ({ workspaceHook: vi.fn() }));

vi.mock("./useDigitalCentersWorkspace", () => ({ useDigitalCentersWorkspace: workspaceHook }));

type Workspace = ReturnType<typeof useDigitalCentersWorkspace>;

describe("DigitalCentersView", () => {
  beforeEach(() => {
    workspaceHook.mockReturnValue(workspaceFixture());
    window.history.replaceState(null, "", "/digital-centers");
    vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockImplementation(() => null);
  });

  it("renders the reference header, toolbar, three columns, and locked-write panel", () => {
    const { container } = render(<DigitalCentersView roles={["Admin"]} />);

    expect(screen.getByText("DIGITALBETRIEB")).toHaveClass("eyebrow");
    expect(screen.getByRole("heading", { level: 1, name: "Digitalzentralen" })).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "Digitalzentrale wählen" })).toHaveValue("ecos");
    expect(screen.getByRole("button", { name: "Daten lesen" })).toBeEnabled();
    expect(screen.getByRole("heading", { name: "Zentralen" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Lok-Arbeitsliste" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Live-Status" })).toHaveAttribute("aria-selected", "true");
    expect(screen.getByRole("tab", { name: "Diagnose" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Meldungen" })).toBeInTheDocument();
    expect(screen.getByRole("img", { name: "Live-Puls der Digitalzentrale" }).tagName).toBe("CANVAS");
    expect(screen.getByText("Schreiben gesperrt")).toBeInTheDocument();
    expect(screen.getByText("In dieser Zentrale sind Schreibbefehle gesperrt.")).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "BR 218 402-6" })).toHaveAttribute("title", "BR 218 402-6");
    expect(container.querySelector(".digital-centers-layout"))
      .toHaveAttribute("data-testid", "digital-centers-layout");
  });

  it("routes plus and configuration actions to Settings without a local configuration dialog", () => {
    render(<DigitalCentersView roles={["Admin"]} />);

    fireEvent.click(screen.getByRole("button", { name: "Digitalzentrale hinzufügen" }));
    expect(window.location.pathname + window.location.search).toBe("/settings?tab=digital");
    expect(screen.queryByRole("dialog", { name: /konfigur/i })).not.toBeInTheDocument();

    window.history.replaceState(null, "", "/digital-centers");
    fireEvent.click(screen.getByRole("button", { name: "Digitalzentrale konfigurieren" }));
    expect(window.location.pathname + window.location.search).toBe("/settings?tab=digital");
  });

  it("wires toolbar, worklist, and tabs to the real workspace contract", () => {
    const workspace = workspaceFixture();
    workspaceHook.mockReturnValue(workspace);
    render(<DigitalCentersView roles={["Admin"]} />);

    fireEvent.change(screen.getByRole("combobox", { name: "Digitalzentrale wählen" }), {
      target: { value: "z21" }
    });
    fireEvent.click(screen.getByRole("button", { name: "Daten lesen" }));
    fireEvent.click(screen.getByRole("button", { name: "BR 218 402-6 vergleichen" }));
    fireEvent.click(screen.getByRole("tab", { name: "Diagnose" }));

    expect(workspace.selectCenter).toHaveBeenCalledWith("z21");
    expect(workspace.readData).toHaveBeenCalledOnce();
    expect(workspace.selectItem).toHaveBeenCalledWith("item-1");
    expect(workspace.openDialog).toHaveBeenCalledWith("comparison", "item-1");
    expect(workspace.setTab).toHaveBeenCalledWith("diagnosis");
  });

  it("keeps topology stable for loading, empty, unavailable, and disconnected states", () => {
    const { rerender } = render(<DigitalCentersView roles={["Admin"]} />);
    const states: Array<[Workspace, string]> = [
      [workspaceFixture({
        centers: [], selectedProvider: null, selectedCenter: null,
        loading: loadingFixture({ workspace: true })
      }), "Digitalzentralen werden geladen"],
      [workspaceFixture({
        centers: [], selectedProvider: null, selectedCenter: null, liveStatus: null, readSession: null,
        workItems: { items: [], page: 1, pageSize: 10, total: 0, totalPages: 0 }
      }), "Keine Digitalzentrale konfiguriert"],
      [workspaceFixture({
        centers: [centerFixture({ active: false, capabilities: capabilitiesFixture(false) })],
        selectedCenter: centerFixture({ active: false, capabilities: capabilitiesFixture(false) }),
        liveStatus: null, actions: actionsFixture(false)
      }), "Live-Monitor nicht verfügbar"],
      [workspaceFixture({ liveStatus: liveStatusFixture({ connected: false, state: "stopped" }) }),
        "Nicht verbunden"]
    ];

    for (const [workspace, message] of states) {
      workspaceHook.mockReturnValue(workspace);
      rerender(<DigitalCentersView roles={["Admin"]} />);
      expect(screen.getAllByText(message).length).toBeGreaterThan(0);
      expect(screen.getByRole("heading", { name: "Zentralen" })).toBeInTheDocument();
      expect(screen.getByRole("heading", { name: "Lok-Arbeitsliste" })).toBeInTheDocument();
      expect(screen.getByRole("tablist", { name: "Digitalzentralen-Status" })).toBeInTheDocument();
    }
  });

  it("renders the focused locomotive comparison dialog", () => {
    const workspace = workspaceFixture({
      selectedItem: workItem,
      selectedItemId: workItem.id,
      dialog: { kind: "comparison", itemId: workItem.id }
    });
    workspaceHook.mockReturnValue(workspace);
    render(<DigitalCentersView roles={["Admin"]} />);

    expect(screen.getByRole("dialog", { name: "Lok-Vergleich BR 218 402-6" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "Digitalzentrale" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "RailKeeper" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Vergleich schließen" }));
    expect(workspace.closeDialog).toHaveBeenCalledOnce();
  });
});

const capabilities = capabilitiesFixture(true);
const ecosCenter = centerFixture();
const z21Center = centerFixture({
  provider: "z21", name: "Z21", active: false, selected: false,
  host: "192.168.2.152", port: 21105, capabilities: capabilitiesFixture(false)
});
const readSession: DigitalCenterReadSession = {
  id: "session-1", provider: "ecos", state: "ready", host: "192.168.2.151", port: 15471,
  capabilities, readStartedAt: "2026-08-21T14:35:00Z", readCompletedAt: "2026-08-21T14:35:01Z",
  createdByUserId: "admin-1", createdAt: "2026-08-21T14:35:00Z", updatedAt: "2026-08-21T14:35:01Z"
};
const workItem: DigitalCenterWorkItem = {
  id: "item-1", sessionId: readSession.id, centerObjectId: "3", vehicleId: "vehicle-1",
  name: "BR 218 402-6", decoderAddress: 3, protocol: "DCC", compareStatus: "ok", stationStatus: "read",
  center: { objectId: 3, name: "BR 218 402-6", decoderAddress: 3, protocol: "DCC" },
  railkeeper: { vehicleId: "vehicle-1", name: "BR 218 402-6", decoderAddress: 3, protocol: "DCC" },
  proposed: {}, conflicts: [], createdAt: "2026-08-21T14:35:01Z", updatedAt: "2026-08-21T14:35:01Z"
};

function workspaceFixture(overrides: Partial<Workspace> = {}): Workspace {
  const liveStatus = liveStatusFixture();
  return {
    centers: [ecosCenter, z21Center], selectedProvider: "ecos", selectedCenter: ecosCenter,
    selectCenter: vi.fn(), liveStatus, connectionState: "running", readSession,
    readData: vi.fn(async () => readSession),
    workItems: { items: [workItem], page: 1, pageSize: 10, total: 42, totalPages: 5 },
    messages: [], search: "", setSearch: vi.fn(), compareStatus: "all", setCompareStatus: vi.fn(),
    page: 1, setPage: vi.fn(), pageSize: 10, setPageSize: vi.fn(), selectedItemId: null,
    selectedItem: null, selectItem: vi.fn(), closeDetail: vi.fn(), tab: "live", setTab: vi.fn(),
    dialog: null, openDialog: vi.fn(), closeDialog: vi.fn(), writePreview: null, writeConfirmation: null,
    previewWrite: vi.fn(async () => writePreview), confirmWrite: vi.fn(async () => writeConfirmation),
    startLive: vi.fn(async () => liveStatus),
    stopLive: vi.fn(async () => liveStatusFixture({ connected: false, state: "stopped" })),
    actions: actionsFixture(true), loading: loadingFixture(),
    errors: { workspace: "", live: "", read: "", worklist: "", detail: "", messages: "", write: "" },
    refresh: vi.fn(async () => undefined), ...overrides
  };
}

const writePreview: Awaited<ReturnType<Workspace["previewWrite"]>> = {
  sessionId: readSession.id, itemId: workItem.id, provider: "ecos", objectId: "3",
  direction: "railkeeper_to_center", fields: ["name"],
  changes: [{ field: "name", current: "Alt", desired: workItem.name }],
  token: "preview-token", expiresAt: "2026-08-21T14:45:00Z"
};
const writeConfirmation: Awaited<ReturnType<Workspace["confirmWrite"]>> = {
  sessionId: readSession.id, itemId: workItem.id, provider: "ecos", objectId: "3",
  direction: "railkeeper_to_center", fields: ["name"], applied: true, verified: true,
  result: "verified", message: "Verifiziert"
};

function capabilitiesFixture(enabled: boolean): DigitalCenterSummary["capabilities"] {
  return {
    testConnection: enabled, readLocomotives: enabled, liveMonitor: enabled,
    writeLocomotives: enabled, writeCVs: false, diagnose: enabled
  };
}

function centerFixture(overrides: Partial<DigitalCenterSummary> = {}): DigitalCenterSummary {
  return {
    provider: "ecos", name: "ESU ECoS", active: true, selected: true,
    host: "192.168.2.151", port: 15471, capabilities: capabilitiesFixture(true), ...overrides
  };
}

function actionsFixture(enabled: boolean): Workspace["actions"] {
  return {
    canTestConnection: enabled, canRead: enabled, canMonitor: enabled,
    canWrite: enabled, canWriteCVs: false, canDiagnose: enabled
  };
}

function loadingFixture(overrides: Partial<Workspace["loading"]> = {}): Workspace["loading"] {
  return {
    workspace: false, live: false, read: false, worklist: false, detail: false, write: false, ...overrides
  };
}

function liveStatusFixture(overrides: Partial<ECoSLiveStatus> = {}): ECoSLiveStatus {
  return {
    provider: "ecos", connected: true, state: "running", host: "192.168.2.151", port: 15471,
    startedAt: "2026-08-21T14:35:00Z", lastSeenAt: "2026-08-21T14:36:08Z",
    lastMessage: "Lok 3 – Geschwindigkeit 63", blocksReceived: 1284, repliesReceived: 1279,
    eventsReceived: 5, subscriptionCommands: ["get(1, info)"],
    pulseSamples: [
      { at: "2026-08-21T14:35:08Z", repliesPerSecond: 0 },
      { at: "2026-08-21T14:35:18Z", repliesPerSecond: 28 },
      { at: "2026-08-21T14:36:08Z", repliesPerSecond: 22 }
    ],
    recentEvents: [
      { at: "2026-08-21T14:36:08Z", kind: "event", message: "Lok 3 – Geschwindigkeit 63", protocol: "DCC" },
      { at: "2026-08-21T14:36:07Z", kind: "event", message: "Weiche 24 – Gerade", protocol: "DCC" }
    ],
    diagnosis: {
      connectionState: "running", lastSuccessfulCommunication: "2026-08-21T14:36:08Z", passive: true
    },
    message: "Live-Monitor aktiv", ...overrides
  };
}
