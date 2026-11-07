import { act, fireEvent, render, renderHook, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { api, ApiError, type DecoderConfigurationContext } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { DecoderConfigurationPanel } from "./DecoderConfigurationPanel";
import { decoderConfigurationDraft, decoderDraftValid } from "./decoderConfigurationDraft";
import { useDecoderConfigurationController } from "./useDecoderConfigurationController";

function context(revision = "a"): DecoderConfigurationContext {
  return { vehicleId: "vehicle-1", vehicleName: "Reference", inventoryNumber: "REF", revision: revision.repeat(64),
    instances: [{ id: "known-unit", label: "Einzelnes Exemplar", manufacturer: "ESU", model: "LokSound", serialNumber: "",
      notes: "", createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" }],
    installations: [], snapshots: [], workspaces: [{ id: "group", vehicleId: "vehicle-1", profile: " Legacy ", origin: "legacy",
      installationId: null, valueCount: 2, sourceCount: 1, contentRevision: "b".repeat(64), bindingRevision: "c".repeat(64),
      createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" }] };
}
const selected = vehicleFixture();
function setup() {
  const setBusy = vi.fn();
  return { setBusy, ...renderHook(({ ownerKey }) => useDecoderConfigurationController({ ownerKey, selected,
    active: true, readonly: false, busy: false, setBusy, captureScope: () => () => true }), { initialProps: { ownerKey: "owner" } }) };
}

describe("confirmed decoder inventory commands", () => {
  beforeEach(() => setLanguage("de"));
  afterEach(() => vi.restoreAllMocks());

  it("keeps a stable identity and draft after conflict and requires explicit comparison", async () => {
    const latest = context("d");
    vi.spyOn(api, "decoderConfigurationContext").mockResolvedValueOnce(context()).mockResolvedValue(latest);
    const write = vi.spyOn(api, "createDecoderInstance").mockRejectedValueOnce(new ApiError("Changed", "decoder_configuration_conflict", 409)).mockResolvedValue(latest);
    const { result, setBusy } = setup();
    await waitFor(() => expect(result.current.context).not.toBeNull());
    act(() => result.current.start("createInstance"));
    act(() => result.current.update({ label: "My unit", model: "Identical model" }));
    const id = result.current.draft!.id;
    expect(write).not.toHaveBeenCalled();
    await act(() => result.current.confirm());
    expect(result.current.draft).toMatchObject({ id, label: "My unit", expectedRevision: "a".repeat(64) });
    await act(() => result.current.confirm()); expect(write).toHaveBeenCalledOnce();
    await act(() => result.current.load(true));
    expect(result.current.draft).toMatchObject({ id, label: "My unit", expectedRevision: latest.revision });
    expect(write).toHaveBeenCalledOnce();
    await act(() => result.current.confirm());
    expect(write).toHaveBeenLastCalledWith("vehicle-1", expect.objectContaining({ id, label: "My unit", expectedRevision: latest.revision }));
    expect(result.current.draft).toBeNull();
    expect(setBusy.mock.calls).toEqual([[true], [false], [true], [false]]);
  });

  it("recognizes a committed creation after a lost response without a second write", async () => {
    const read = vi.spyOn(api, "decoderConfigurationContext").mockResolvedValue(context());
    const write = vi.spyOn(api, "createDecoderCVWorkspace").mockRejectedValue(new TypeError("Response lost"));
    const { result } = setup(); await waitFor(() => expect(result.current.context).not.toBeNull());
    act(() => result.current.start("workspace")); act(() => result.current.update({ profile: "Second group" }));
    const id = result.current.draft!.id;
    await act(() => result.current.confirm());
    expect(result.current.needsReview).toBe(true);
    await act(() => result.current.confirm()); expect(write).toHaveBeenCalledOnce();
    const latest = context("d"); latest.workspaces.push({ ...latest.workspaces[0], id, profile: "Second group" });
    read.mockResolvedValue(latest);
    await act(() => result.current.load(true));
    expect(result.current.draft).toBeNull();
    expect(result.current.status).toContain("bereits gespeichert");
    await act(() => result.current.confirm()); expect(write).toHaveBeenCalledOnce();
  });

  it("keeps unknown dates null and clearing a hardware association explicit", async () => {
    vi.spyOn(api, "decoderConfigurationContext").mockResolvedValue(context());
    const install = vi.spyOn(api, "installDecoder").mockResolvedValue(context("d"));
    const assign = vi.spyOn(api, "assignDecoderCVWorkspace").mockResolvedValue(context("e"));
    const { result } = setup(); await waitFor(() => expect(result.current.context).not.toBeNull());
    act(() => result.current.start("install", "known-unit"));
    expect(result.current.draft?.installedAt).toBe("");
    await act(() => result.current.confirm());
    expect(install).toHaveBeenCalledWith("vehicle-1", expect.objectContaining({ decoderId: "known-unit", installedAt: null }));
    act(() => result.current.start("assign", "group"));
    const snapshotId = result.current.draft!.snapshotId;
    await act(() => result.current.confirm());
    expect(assign).toHaveBeenCalledWith("vehicle-1", "group", { expectedRevision: "d".repeat(64), installationId: null, snapshotId });
  });

  it("adopts untouched remote corrections and requires a choice for a field changed on both sides", async () => {
    const latest = context("d");
    latest.instances[0] = { ...latest.instances[0], label: "Remote designation", manufacturer: "ZIMO", serialNumber: "Recorded serial" };
    vi.spyOn(api, "decoderConfigurationContext").mockResolvedValueOnce(context()).mockResolvedValue(latest);
    const write = vi.spyOn(api, "updateDecoderInstance").mockResolvedValue(latest);
    const { result } = setup(); await waitFor(() => expect(result.current.context).not.toBeNull());
    act(() => result.current.start("editInstance", "known-unit"));
    act(() => result.current.update({ label: "My designation" }));
    await act(() => result.current.load(true));
    expect(result.current.draft).toMatchObject({ label: "My designation", manufacturer: "ZIMO", serialNumber: "Recorded serial" });
    expect(result.current.conflicts).toEqual([{ field: "label", local: "My designation", saved: "Remote designation" }]);
    await act(() => result.current.confirm()); expect(write).not.toHaveBeenCalled();
    // Loading the same comparison again must not silently resolve the conflict.
    await act(() => result.current.load(true)); expect(result.current.conflicts).toHaveLength(1);
    act(() => result.current.resolve("label", false));
    await act(() => result.current.confirm());
    expect(write).toHaveBeenCalledWith("vehicle-1", expect.objectContaining({ label: "My designation", manufacturer: "ZIMO", serialNumber: "Recorded serial" }));
  });

  it("preserves a concurrently recorded removal date while correcting an unrelated installation date", async () => {
    const before = context(), latest = context("d");
    const installation: DecoderConfigurationContext["installations"][number] = { id: "installation", decoderId: "known-unit", vehicleId: "vehicle-1", vehicleReferenceId: "vehicle-1",
      vehicleName: "Reference", vehicleInventoryNumber: "REF", state: "removed", installedAt: null, removedAt: null,
      notes: "Previous note", recordedAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" };
    before.installations = [installation]; latest.installations = [{ ...installation, removedAt: "2026-11-07" }];
    vi.spyOn(api, "decoderConfigurationContext").mockResolvedValueOnce(before).mockResolvedValue(latest);
    const write = vi.spyOn(api, "correctDecoderInstallationDates").mockResolvedValue(latest);
    const { result } = setup(); await waitFor(() => expect(result.current.context).not.toBeNull());
    act(() => result.current.start("dates", "installation"));
    act(() => result.current.update({ installedAt: "2026-10-01" }));
    await act(() => result.current.load(true));
    expect(result.current.draft?.removedAt).toBe("2026-11-07"); expect(result.current.conflicts).toEqual([]);
    await act(() => result.current.confirm());
    expect(write).toHaveBeenCalledWith("vehicle-1", "installation", { expectedRevision: latest.revision, installedAt: "2026-10-01", removedAt: "2026-11-07" });
  });

  it("ignores a late response from another owner", async () => {
    let finish!: (value: DecoderConfigurationContext) => void;
    vi.spyOn(api, "decoderConfigurationContext").mockImplementationOnce(() => new Promise(resolve => { finish = resolve; })).mockResolvedValue(context("e"));
    const { result, rerender } = setup();
    rerender({ ownerKey: "other-owner" });
    await waitFor(() => expect(result.current.context?.revision).toBe("e".repeat(64)));
    await act(async () => { finish(context()); });
    expect(result.current.context?.revision).toBe("e".repeat(64));
    expect(result.current.draft).toBeNull();
  });

  it("preserves invalid physical dates instead of inventing a valid date", () => {
    const state = context(), draft = decoderConfigurationDraft("install", state, "known-unit");
    expect(decoderDraftValid(draft, state)).toBe(true);
    expect(decoderDraftValid({ ...draft, installedAt: "2026-02-30" }, state)).toBe(false);
    expect(decoderDraftValid({ ...draft, installedAt: "2026-2-2" }, state)).toBe(false);
    expect(decoderDraftValid({ ...draft, installedAt: "2026-11-07", removedAt: "2026-11-06" }, state)).toBe(false);
  });

  it("does not submit the surrounding vehicle form and keeps independent required fields out of its validation", async () => {
    vi.spyOn(api, "decoderConfigurationContext").mockResolvedValue(context());
    const submit = vi.fn(), write = vi.spyOn(api, "createDecoderInstance");
    function Harness() {
      const model = useDecoderConfigurationController({ ownerKey: "owner", selected, active: true, readonly: false,
        busy: false, setBusy: vi.fn(), captureScope: () => () => true });
      return <form onSubmit={event => { event.preventDefault(); submit(); }}><input aria-label="Main draft" defaultValue="Untouched" />
        <DecoderConfigurationPanel controller={model} /><button type="submit">Save main</button></form>;
    }
    render(<Harness />);
    await waitFor(() => expect(screen.getByRole("button", { name: "Decoderexemplar erfassen" })).toBeEnabled());
    await userEvent.click(screen.getByRole("button", { name: "Decoderexemplar erfassen" }));
    const label = screen.getByRole("textbox", { name: "Bezeichnung des Exemplars" });
    expect(label).toHaveAttribute("aria-required", "true"); expect(label).not.toHaveAttribute("required");
    fireEvent.keyDown(label, { key: "Enter" });
    expect(submit).not.toHaveBeenCalled();
    await userEvent.click(screen.getByRole("button", { name: "Save main" }));
    expect(submit).toHaveBeenCalledOnce();
    expect(screen.getByLabelText("Main draft")).toHaveValue("Untouched");
    expect(write).not.toHaveBeenCalled();
  });
});
