import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { setLanguage } from "../../shared/i18n";
import { ecosVehicleDraftStorageKey, emptyVehicle } from "../vehicles/vehicleViewModel";
import type {
  DigitalCenterReadSession,
  DigitalCenterSummary,
  DigitalCenterWorkItem,
  ECoSLiveStatus
} from "./digitalCenterModel";
import { DigitalCentersView } from "./DigitalCentersView";
import type { useDigitalCentersWorkspace } from "./useDigitalCentersWorkspace";

const digitalCentersCSS = readFileSync(resolve(process.cwd(), "src/styles/digital-centers.css"), "utf8");

const { adoptionHook, workspaceHook } = vi.hoisted(() => ({
  adoptionHook: vi.fn(),
  workspaceHook: vi.fn()
}));

vi.mock("./useDigitalCentersWorkspace", () => ({ useDigitalCentersWorkspace: workspaceHook }));
vi.mock("./useDigitalCenterVehicleAdoption", () => ({
  useDigitalCenterVehicleAdoption: adoptionHook
}));

type Workspace = ReturnType<typeof useDigitalCentersWorkspace>;

describe("DigitalCentersView", () => {
  beforeEach(() => {
    setLanguage("de");
    workspaceHook.mockReturnValue(workspaceFixture());
    adoptionHook.mockReturnValue(adoptionFixture());
    window.history.replaceState(null, "", "/digital-centers");
    window.sessionStorage.removeItem(ecosVehicleDraftStorageKey);
    vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockImplementation(() => null);
  });

  it("renders the reference header, toolbar, three columns, and locked-write panel", () => {
    const { container } = render(<DigitalCentersView roles={["Admin"]} />);

    expect(screen.getByText("DIGITALBETRIEB")).toHaveClass("eyebrow");
    expect(screen.getByRole("heading", { level: 1, name: "Digitalzentralen" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Digitalzentrale wählen" })).toHaveTextContent("ESU ECoS");
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

  it("wires toolbar, worklist, and tabs to the real workspace contract", async () => {
    const user = userEvent.setup();
    const workspace = workspaceFixture();
    workspaceHook.mockReturnValue(workspace);
    render(<DigitalCentersView roles={["Admin"]} />);

    await user.click(screen.getByRole("button", { name: "Digitalzentrale wählen" }));
    await user.click(screen.getByRole("option", { name: "Z21" }));
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

  it("shows an explicit read-in-progress state without an empty work list", () => {
    workspaceHook.mockReturnValue(workspaceFixture({ loading: loadingFixture({ read: true }) }));
    render(<DigitalCentersView roles={["Admin"]} />);

    expect(screen.getByRole("button", { name: "Daten werden gelesen" })).toBeDisabled();
    expect(screen.getByText("Lokdaten werden geladen")).toBeInTheDocument();
    expect(screen.queryByText("Noch keine Lokdaten gelesen")).not.toBeInTheDocument();
  });

  it("shows workspace and read failures distinctly without claiming configuration is empty", () => {
    const retry = vi.fn(async () => undefined);
    workspaceHook.mockReturnValue(workspaceFixture({
      centers: [],
      selectedProvider: null,
      selectedCenter: null,
      liveStatus: null,
      errors: {
        workspace: "Arbeitsbereich nicht erreichbar",
        live: "",
        read: "Lesen fehlgeschlagen",
        worklist: "",
        detail: "",
        messages: "",
        write: ""
      },
      refresh: retry
    }));
    render(<DigitalCentersView roles={["Admin"]} />);

    expect(screen.getByRole("alert", { name: "Arbeitsbereich konnte nicht geladen werden" }))
      .toHaveTextContent("Arbeitsbereich nicht erreichbar");
    expect(screen.queryByText("Keine Digitalzentrale konfiguriert")).not.toBeInTheDocument();
    expect(screen.getByRole("alert", { name: "Lesefehler" })).toHaveTextContent("Lesen fehlgeschlagen");
    fireEvent.click(screen.getByRole("button", { name: "Digitalzentralen erneut laden" }));
    expect(retry).toHaveBeenCalledOnce();
  });

  it("traps focus in the comparison dialog, closes on Escape, and restores focus", async () => {
    const user = userEvent.setup();
    const invoker = document.createElement("button");
    invoker.textContent = "Vergleich öffnen";
    document.body.append(invoker);
    invoker.focus();
    const workspace = workspaceFixture({
      selectedItem: workItem,
      selectedItemId: workItem.id,
      dialog: { kind: "comparison", itemId: workItem.id }
    });
    workspaceHook.mockReturnValue(workspace);
    const { unmount } = render(<DigitalCentersView roles={["Admin"]} />);

    expect(screen.getByRole("dialog", { name: "Lok-Vergleich BR 218 402-6" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "Digitalzentrale" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "RailKeeper" })).toBeInTheDocument();
    const first = screen.getByRole("button", { name: "Vergleich schließen" });
    const last = screen.getByRole("button", { name: "Schließen" });
    expect(first).toHaveFocus();
    last.focus();
    await user.tab();
    expect(first).toHaveFocus();
    await user.keyboard("{Escape}");
    expect(workspace.closeDialog).toHaveBeenCalledOnce();
    unmount();
    expect(invoker).toHaveFocus();
    invoker.remove();
  });

  it("explains why an ECoS-only locomotive cannot be written", () => {
    const unmatchedItem: DigitalCenterWorkItem = {
      ...workItem,
      id: "item-new",
      vehicleId: "",
      name: "LokPilot 5 micro",
      compareStatus: "new",
      railkeeper: {}
    };
    const workspace = workspaceFixture({
      selectedItemId: unmatchedItem.id,
      selectedItem: unmatchedItem,
      dialog: { kind: "comparison", itemId: unmatchedItem.id }
    });
    workspaceHook.mockReturnValue(workspace);

    render(<DigitalCentersView roles={["Admin"]} />);

    expect(screen.getByRole("button", { name: "Schreibvorschau erstellen" })).toBeDisabled();
    expect(screen.getByText(
      "Diese Lok ist noch keinem RailKeeper-Fahrzeug zugeordnet. Erst nach dem Anlegen oder Zuordnen " +
      "eines Fahrzeugs kann in die Digitalzentrale geschrieben werden."
    )).toBeInTheDocument();
    expect(workspace.previewWrite).not.toHaveBeenCalled();
  });

  it("offers the two explicit adoption paths for an ECoS-only locomotive", async () => {
    const user = userEvent.setup();
    const unmatchedItem = unmatchedWorkItem();
    const workspace = workspaceFixture({
      selectedItemId: unmatchedItem.id,
      selectedItem: unmatchedItem,
      dialog: { kind: "comparison", itemId: unmatchedItem.id }
    });
    const adoption = adoptionFixture();
    workspaceHook.mockReturnValue(workspace);
    adoptionHook.mockReturnValue(adoption);

    render(<DigitalCentersView roles={["Admin"]} />);

    await user.click(screen.getByRole("button", { name: "Bestehendem Fahrzeug zuordnen" }));
    expect(workspace.openDialog).toHaveBeenCalledWith("assignment", unmatchedItem.id);
    expect(adoption.commands.load).toHaveBeenCalledWith(unmatchedItem);

    await user.click(screen.getByRole("button", { name: "Neues Fahrzeug anlegen" }));
    expect(window.location.pathname + window.location.search).toBe("/vehicles?source=ecos");
    expect(JSON.parse(window.sessionStorage.getItem(ecosVehicleDraftStorageKey) || "null"))
      .toMatchObject({
        source: "ecos",
        mode: "create",
        vehicle: { name: "LokPilot 5 micro", digitalDecoderNumber: "1001" },
        returnToDigitalCenters: { sessionId: "session-1", objectId: "77" }
      });
  });

  it("does not expose vehicle adoption actions without the Admin role", () => {
    const unmatchedItem = unmatchedWorkItem();
    workspaceHook.mockReturnValue(workspaceFixture({
      selectedItemId: unmatchedItem.id,
      selectedItem: unmatchedItem,
      dialog: { kind: "comparison", itemId: unmatchedItem.id }
    }));

    render(<DigitalCentersView roles={["Viewer"]} />);

    expect(screen.queryByRole("button", { name: "Neues Fahrzeug anlegen" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Bestehendem Fahrzeug zuordnen" }))
      .not.toBeInTheDocument();
  });

  it("renders the explicit vehicle picker and forwards the confirmed assignment", async () => {
    const user = userEvent.setup();
    const unmatchedItem = unmatchedWorkItem();
    const adoption = adoptionFixture({
      state: {
        vehicles: [{
          ...emptyVehicle,
          id: "vehicle-2",
          inventoryNumber: "RK-002",
          manufacturer: "Roco",
          name: "LokPilot 5 micro",
          digital: true,
          digitalDecoderNumber: "1001",
          createdAt: "2026-08-23T08:00:00Z",
          updatedAt: "2026-08-23T08:00:00Z"
        }],
        selectedVehicleId: "vehicle-2",
        loading: false,
        saving: false,
        error: ""
      }
    });
    workspaceHook.mockReturnValue(workspaceFixture({
      selectedItemId: unmatchedItem.id,
      selectedItem: unmatchedItem,
      dialog: { kind: "assignment", itemId: unmatchedItem.id }
    }));
    adoptionHook.mockReturnValue(adoption);

    render(<DigitalCentersView roles={["Admin"]} />);

    expect(screen.getByRole("dialog", { name: "Fahrzeugzuordnung für LokPilot 5 micro" }))
      .toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /RK-002/ })).toBeChecked();
    await user.click(screen.getByRole("button", { name: "Fahrzeug zuordnen" }));
    expect(adoption.commands.assign).toHaveBeenCalledWith(unmatchedItem, "vehicle-2");
  });

  it("reads once after returning from the vehicle editor and removes the return marker", async () => {
    const workspace = workspaceFixture();
    workspaceHook.mockReturnValue(workspace);
    window.history.replaceState(null, "", "/digital-centers?sessionId=session-1&objectId=77");

    render(<DigitalCentersView roles={["Admin"]} />);

    await waitFor(() => expect(workspace.readData).toHaveBeenCalledOnce());
    expect(window.location.pathname + window.location.search).toBe("/digital-centers");
  });

  it("renders factual pagination without current-page quick-filter counts or text glyph controls", () => {
    workspaceHook.mockReturnValue(workspaceFixture({
      workItems: { items: [workItem], page: 2, pageSize: 25, total: 60, totalPages: 3 },
      page: 2,
      pageSize: 25
    }));
    render(<DigitalCentersView roles={["Admin"]} />);

    expect(screen.getByText("26–50 von 60")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Abweichung filtern" })).toHaveTextContent("Prüfen");
    expect(screen.getByRole("button", { name: "Neu filtern" })).toHaveTextContent("Neu");
    expect(screen.getByRole("button", { name: "Abweichung filtern" })).not.toHaveTextContent(/\d/);
    expect(screen.getByRole("button", { name: "Neu filtern" })).not.toHaveTextContent(/\d/);
    expect(screen.getByRole("button", { name: "Vorherige Seite" }).textContent).toBe("");
    expect(screen.getByRole("button", { name: "Nächste Seite" }).textContent).toBe("");
  });

  it("renders the complete operational hierarchy in English", () => {
    setLanguage("en");
    render(<DigitalCentersView roles={["Admin"]} />);

    expect(screen.getByRole("heading", { level: 1, name: "Command stations" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Read data" })).toBeEnabled();
    expect(screen.getByRole("heading", { name: "Stations" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Locomotive worklist" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Live status" })).toHaveAttribute("aria-selected", "true");
    expect(screen.getByRole("tab", { name: "Diagnostics" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Messages" })).toBeInTheDocument();
    expect(screen.getByText("Writing locked")).toBeInTheDocument();
  });
});

describe("Digital Centers responsive CSS contract", () => {
  it("switches the workspace to two columns before the 1500px shell can overflow", () => {
    const breakpoint = digitalCentersCSS
      .split("@media (max-width: 1500px)")[1]
      ?.split("@media (max-width: 1180px)")[0] ?? "";

    expect(breakpoint).toMatch(
      /\.digital-centers-layout\s*\{[\s\S]*grid-template-columns:\s*minmax\([^;]+\)\s+minmax\([^;]+\);/
    );
    expect(breakpoint).toMatch(/\.digital-centers-status\s*\{[\s\S]*grid-column:\s*1\s*\/\s*-1;/);

    const narrowBreakpoint = digitalCentersCSS.split("@media (max-width: 900px)")[1] ?? "";
    expect(narrowBreakpoint).toMatch(/\.digital-centers-layout\s*\{\s*grid-template-columns:\s*1fr;/);
  });

  it("uses the toolbar selector and local table scrolling at 900px", () => {
    const [desktopCSS, narrowBreakpoint = ""] = digitalCentersCSS.split("@media (max-width: 900px)");

    expect(narrowBreakpoint).toMatch(/\.digital-center-list\s*\{[^}]*display:\s*none;/s);
    expect(narrowBreakpoint).toMatch(/\.digital-status-tabs\s*\{[^}]*width:\s*100%;/s);
    expect(digitalCentersCSS).toMatch(
      /\.digital-centers-workspace\s*\{[^}]*overflow-x:\s*(?:hidden|clip);/s
    );
    expect(desktopCSS).not.toMatch(
      /\.digital-worklist-table\s*\{[^}]*min-width:\s*720px;/s
    );
    expect(narrowBreakpoint).toMatch(
      /\.digital-worklist-table\s*\{[^}]*min-width:\s*720px;/s
    );
    expect(desktopCSS).not.toMatch(
      /\.digital-worklist-table-wrap\s*\{[^}]*overflow-x:\s*auto;/s
    );
    expect(narrowBreakpoint).toMatch(
      /\.digital-worklist-table-wrap\s*\{[^}]*overflow-x:\s*auto;/s
    );
  });

  it("wraps long operational and capability messages inside the workspace", () => {
    expect(digitalCentersCSS).toMatch(
      /\.digital-center-capability-note\s*\{[^}]*overflow-wrap:\s*anywhere;/s
    );
    expect(digitalCentersCSS).toMatch(
      /\.digital-workspace-operation-error\s*\{[^}]*overflow-wrap:\s*anywhere;/s
    );
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

function unmatchedWorkItem(): DigitalCenterWorkItem {
  return {
    ...workItem,
    id: "item-new",
    centerObjectId: "77",
    vehicleId: "",
    name: "LokPilot 5 micro",
    decoderAddress: 1001,
    compareStatus: "new",
    center: {
      objectId: 77,
      name: "LokPilot 5 micro",
      decoderAddress: 1001,
      protocol: "DCC"
    },
    railkeeper: {}
  };
}

function adoptionFixture(overrides: Record<string, unknown> = {}) {
  return {
    state: { vehicles: [], selectedVehicleId: "", loading: false, saving: false, error: "" },
    setters: { setSelectedVehicleId: vi.fn() },
    commands: {
      load: vi.fn(async () => undefined),
      assign: vi.fn(async () => undefined),
      reset: vi.fn()
    },
    ...overrides
  };
}

function workspaceFixture(overrides: Partial<Workspace> = {}): Workspace {
  const liveStatus = liveStatusFixture();
  return {
    centers: [ecosCenter, z21Center], selectedProvider: "ecos", selectedCenter: ecosCenter,
    selectCenter: vi.fn(), liveStatus, connectionState: "running", readSession,
    readData: vi.fn(async () => readSession),
    workItems: { items: [workItem], page: 1, pageSize: 10, total: 42, totalPages: 5 },
    sessionTotal: 42,
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
