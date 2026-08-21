import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import type {
  DigitalCenterReadSession,
  DigitalCenterSummary,
  DigitalCenterWorkItem,
  DigitalCenterWorkItemPage,
  DigitalCenterWriteConfirmation,
  DigitalCenterWritePreview,
  ECoSLiveStatus
} from "./digitalCenterModel";
import { useDigitalCentersWorkspace } from "./useDigitalCentersWorkspace";

const centers: DigitalCenterSummary[] = [
  centerFixture({ provider: "ecos", selected: true }),
  centerFixture({
    provider: "z21",
    name: "Z21",
    selected: false,
    capabilities: capabilitiesFixture({ liveMonitor: true, writeLocomotives: false })
  })
];
const readySession: DigitalCenterReadSession = {
  id: "session-1",
  provider: "ecos",
  state: "ready",
  host: "center.local",
  port: 15471,
  capabilities: capabilitiesFixture(),
  readStartedAt: "2026-08-21T10:00:00Z",
  readCompletedAt: "2026-08-21T10:00:01Z",
  createdByUserId: "admin-1",
  createdAt: "2026-08-21T10:00:00Z",
  updatedAt: "2026-08-21T10:00:01Z"
};
const item = workItemFixture({ id: "item-1", name: "BR 218" });
const worklist = worklistFixture([item]);
const stoppedLive = liveStatusFixture({ state: "stopped", connected: false });

describe("useDigitalCentersWorkspace", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(api, "digitalCenterWorkspace").mockResolvedValue({ centers });
    vi.spyOn(api, "digitalCenterLiveStatus").mockResolvedValue(stoppedLive);
    vi.spyOn(api, "startDigitalCenterReadSession").mockResolvedValue(readySession);
    vi.spyOn(api, "digitalCenterWorkItems").mockResolvedValue(worklist);
    vi.spyOn(api, "digitalCenterWorkItem").mockResolvedValue(item);
    vi.spyOn(api, "digitalCenterSessionMessages").mockResolvedValue({ messages: [] });
    vi.spyOn(api, "startDigitalCenterLive").mockImplementation(async (provider) =>
      liveStatusFixture({ provider, state: "running" })
    );
    vi.spyOn(api, "stopDigitalCenterLive").mockResolvedValue(stoppedLive);
    vi.spyOn(api, "previewDigitalCenterWrite").mockResolvedValue(writePreviewFixture());
    vi.spyOn(api, "confirmDigitalCenterWrite").mockResolvedValue(writeConfirmationFixture());
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("preserves work-list filters and page across center and detail selection", async () => {
    const { result } = renderHook(() => useDigitalCentersWorkspace({ pollIntervalMs: 20 }));
    await waitFor(() => expect(result.current.loading.workspace).toBe(false));

    await act(async () => result.current.readData());
    await waitFor(() => expect(result.current.workItems.items).toEqual([item]));
    act(() => {
      result.current.setSearch("BR 218");
      result.current.setCompareStatus("deviation");
      result.current.setPage(3);
      result.current.selectItem(item.id);
    });
    await waitFor(() => expect(result.current.selectedItem?.id).toBe(item.id));

    act(() => {
      result.current.closeDetail();
      result.current.selectCenter("z21");
    });

    expect(result.current.search).toBe("BR 218");
    expect(result.current.compareStatus).toBe("deviation");
    expect(result.current.page).toBe(3);
    expect(result.current.selectedItemId).toBeNull();
  });

  it("derives actions from selected capabilities instead of provider names", async () => {
    const { result } = renderHook(() => useDigitalCentersWorkspace());
    await waitFor(() => expect(result.current.loading.workspace).toBe(false));

    expect(result.current.actions).toMatchObject({ canRead: true, canMonitor: true, canWrite: true });
    act(() => result.current.selectCenter("z21"));
    await waitFor(() => expect(result.current.selectedProvider).toBe("z21"));

    expect(result.current.actions).toMatchObject({ canRead: true, canMonitor: true, canWrite: false });
    await act(async () => result.current.startLive());
    expect(result.current.liveStatus?.provider).toBe("z21");
    expect(api.startDigitalCenterLive).toHaveBeenCalledWith("z21", undefined);
  });

  it("ignores stale live responses after the selected center changes", async () => {
    const ecos = deferred<ECoSLiveStatus>();
    const z21 = deferred<ECoSLiveStatus>();
    vi.mocked(api.digitalCenterLiveStatus).mockImplementation((provider) =>
      provider === "ecos" ? ecos.promise : z21.promise
    );
    const { result } = renderHook(() => useDigitalCentersWorkspace());
    await waitFor(() => expect(result.current.selectedProvider).toBe("ecos"));

    act(() => result.current.selectCenter("z21"));
    z21.resolve(liveStatusFixture({ provider: "z21", state: "stopped" }));
    await waitFor(() => expect(result.current.liveStatus?.provider).toBe("z21"));
    ecos.resolve(liveStatusFixture({ provider: "ecos", state: "running" }));
    await act(async () => Promise.resolve());

    expect(result.current.selectedProvider).toBe("z21");
    expect(result.current.liveStatus?.provider).toBe("z21");
  });

  it("polls only while passive monitoring runs and stops after stop and unmount", async () => {
    vi.mocked(api.digitalCenterLiveStatus).mockResolvedValue(liveStatusFixture({ state: "running" }));
    const { result, unmount } = renderHook(() => useDigitalCentersWorkspace({ pollIntervalMs: 10 }));
    await waitFor(() => expect(api.digitalCenterLiveStatus).toHaveBeenCalledTimes(2));

    await act(async () => result.current.stopLive());
    const callsAtStop = vi.mocked(api.digitalCenterLiveStatus).mock.calls.length;
    await new Promise((resolve) => setTimeout(resolve, 30));
    expect(api.digitalCenterLiveStatus).toHaveBeenCalledTimes(callsAtStop);

    unmount();
    const callsAtUnmount = vi.mocked(api.digitalCenterLiveStatus).mock.calls.length;
    await new Promise((resolve) => setTimeout(resolve, 40));

    expect(api.digitalCenterLiveStatus).toHaveBeenCalledTimes(callsAtUnmount);
  });

  it("pauses polling for a stopped monitor and isolates actionable errors", async () => {
    vi.mocked(api.digitalCenterLiveStatus).mockRejectedValueOnce(new Error("Live nicht erreichbar"));
    vi.mocked(api.startDigitalCenterReadSession).mockRejectedValueOnce(new Error("Lesen fehlgeschlagen"));
    const { result } = renderHook(() => useDigitalCentersWorkspace({ pollIntervalMs: 10 }));
    await waitFor(() => expect(result.current.errors.live).toBe("Live nicht erreichbar"));

    await act(async () => {
      await expect(result.current.readData()).rejects.toThrow("Lesen fehlgeschlagen");
    });

    expect(result.current.errors.live).toBe("Live nicht erreichbar");
    expect(result.current.errors.read).toBe("Lesen fehlgeschlagen");
    expect(result.current.errors.workspace).toBe("");
    await new Promise((resolve) => setTimeout(resolve, 30));
    expect(api.digitalCenterLiveStatus).toHaveBeenCalledOnce();
  });

  it("keeps session-message failures isolated from the work list", async () => {
    vi.mocked(api.digitalCenterSessionMessages).mockRejectedValueOnce(new Error("Meldungen nicht verfügbar"));
    const { result } = renderHook(() => useDigitalCentersWorkspace());
    await waitFor(() => expect(result.current.loading.workspace).toBe(false));

    await act(async () => result.current.readData());
    await waitFor(() => expect(result.current.errors.messages).toBe("Meldungen nicht verfügbar"));

    expect(result.current.workItems.items).toEqual([item]);
    expect(result.current.errors.worklist).toBe("");
  });

  it("keeps only the newest work-list response and owns preview-confirm state", async () => {
    const oldSearch = deferred<DigitalCenterWorkItemPage>();
    const newest = workItemFixture({ id: "item-new", name: "ICE 3" });
    vi.mocked(api.digitalCenterWorkItems).mockImplementation((_sessionID, filter) => {
      if (filter.query === "old") return oldSearch.promise;
      if (filter.query === "new") return Promise.resolve(worklistFixture([newest]));
      return Promise.resolve(worklist);
    });
    const { result } = renderHook(() => useDigitalCentersWorkspace());
    await waitFor(() => expect(result.current.loading.workspace).toBe(false));
    await act(async () => result.current.readData());

    act(() => result.current.setSearch("old"));
    await waitFor(() => expect(api.digitalCenterWorkItems).toHaveBeenCalledWith(
      readySession.id,
      expect.objectContaining({ query: "old" })
    ));
    act(() => result.current.setSearch("new"));
    await waitFor(() => expect(result.current.workItems.items).toEqual([newest]));
    oldSearch.resolve(worklistFixture([item]));
    await act(async () => Promise.resolve());
    expect(result.current.workItems.items).toEqual([newest]);

    act(() => result.current.selectItem(newest.id));
    await act(async () => result.current.previewWrite(["name"]));
    expect(result.current.writePreview?.token).toBe("public-grant");
    await act(async () => result.current.confirmWrite());
    expect(result.current.writeConfirmation?.result).toBe("verified");
    expect(api.confirmDigitalCenterWrite).toHaveBeenCalledWith(readySession.id, newest.id, {
      token: "public-grant",
      confirm: true,
      fields: ["name"]
    });
  });
});

function capabilitiesFixture(overrides: Partial<DigitalCenterSummary["capabilities"]> = {}) {
  return {
    testConnection: true,
    readLocomotives: true,
    liveMonitor: true,
    writeLocomotives: true,
    writeCVs: false,
    diagnose: true,
    ...overrides
  };
}

function centerFixture(overrides: Partial<DigitalCenterSummary> = {}): DigitalCenterSummary {
  return {
    provider: "ecos",
    name: "ESU ECoS",
    active: true,
    selected: false,
    host: "center.local",
    port: 15471,
    capabilities: capabilitiesFixture(),
    ...overrides
  };
}

function liveStatusFixture(overrides: Partial<ECoSLiveStatus> = {}): ECoSLiveStatus {
  return {
    provider: "ecos",
    connected: true,
    state: "running",
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
      connectionState: "running",
      lastSuccessfulCommunication: "2026-08-21T10:00:01Z",
      passive: true
    },
    message: "Live-Monitor aktiv",
    ...overrides
  };
}

function workItemFixture(overrides: Partial<DigitalCenterWorkItem> = {}): DigitalCenterWorkItem {
  return {
    id: "item-1",
    sessionId: "session-1",
    centerObjectId: "3",
    vehicleId: "vehicle-1",
    name: "BR 218",
    decoderAddress: 18,
    protocol: "DCC",
    compareStatus: "deviation",
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

function worklistFixture(items: DigitalCenterWorkItem[]): DigitalCenterWorkItemPage {
  return { items, page: 1, pageSize: 10, total: items.length, totalPages: items.length ? 1 : 0 };
}

function writePreviewFixture(): DigitalCenterWritePreview {
  return {
    sessionId: "session-1",
    itemId: "item-new",
    provider: "ecos",
    objectId: "3",
    direction: "railkeeper_to_center",
    fields: ["name"],
    changes: [{ field: "name", current: "Alte Lok", desired: "ICE 3" }],
    token: "public-grant",
    expiresAt: "2026-08-21T10:10:00Z"
  };
}

function writeConfirmationFixture(): DigitalCenterWriteConfirmation {
  return {
    sessionId: "session-1",
    itemId: "item-new",
    provider: "ecos",
    objectId: "3",
    direction: "railkeeper_to_center",
    fields: ["name"],
    applied: true,
    verified: true,
    result: "verified",
    message: "Verifiziert"
  };
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((promiseResolve) => {
    resolve = promiseResolve;
  });
  return { promise, resolve };
}
