import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { api, ApiError } from "../../shared/api";
import type { VehicleValuation, VehicleValuationExport, VehicleValuationHistory } from "../../shared/apiValuations";
import { canonicalValuationAmount, valuationChangedEvent, type ValuationDraft } from "./valuationModel";
import { useVehicleValuations } from "./useVehicleValuations";

function saved(draft: ValuationDraft, extra: Partial<VehicleValuation> = {}): VehicleValuation {
  return {    
...draft, amount: canonicalValuationAmount(draft.amount) || "0.00", currency: draft.currency.trim().toUpperCase(),
    source: draft.source.trim(), notes: draft.notes.trim(), vehicleId: "a", recordedAt: "2026-11-07T12:00:00.000000000Z", replacedById: "", ...extra  
};
}
function exported(entries: VehicleValuation[]): VehicleValuationExport {
  return { format: "railkeeper-valuations", version: 1, vehicleId: "a", exportedAt: "2026-11-07T12:00:00.000000000Z", entries };
}
function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>(finish => { resolve = finish; });
  return { promise, resolve };
}

describe("valuation save acknowledgment and ownership", () => {
  beforeEach(() => {
    vi.spyOn(api, "vehicleValuations").mockResolvedValue({ entries: [], total: 0 });
    vi.spyOn(api, "recordVehicleValuation").mockImplementation(async (_vehicle, input) => saved({ notes: "", replacesId: "", withdrawn: false, ...input }));
    vi.spyOn(api, "vehicleValuationsExport").mockResolvedValue(exported([]));
  });
  async function ready() {
    const hook = renderHook(() => useVehicleValuations("a", true));
    await waitFor(() => expect(hook.result.current.loading).toBe(false));
    act(() => hook.result.current.setDraft(draft => ({ ...draft, amount: "123,456789", source: " Manual " })));
    return hook;
  }

  it("loads only the active view and retains input across tabs", async () => {
    const hook = renderHook(({ active }) => useVehicleValuations("a", active), { initialProps: { active: false } });
    expect(api.vehicleValuations).not.toHaveBeenCalled();
    hook.rerender({ active: true });
    await waitFor(() => expect(hook.result.current.loading).toBe(false));
    act(() => hook.result.current.setDraft(draft => ({ ...draft, source: "Uncommitted" })));
    const identity = hook.result.current.draft.id;
    hook.rerender({ active: false }); hook.rerender({ active: true });
    expect(hook.result.current.draft).toMatchObject({ id: identity, source: "Uncommitted" });
    await waitFor(() => expect(hook.result.current.loading).toBe(false));
  });

  it("commits once on repeated confirmation and resets only after acknowledgment", async () => {
    const hook = await ready(), pending = deferred<VehicleValuation>();
    const draft = hook.result.current.draft;
    vi.mocked(api.recordVehicleValuation).mockReturnValueOnce(pending.promise);
    let operation!: Promise<void>;
    act(() => { operation = hook.result.current.save(); void hook.result.current.save(); });
    expect(api.recordVehicleValuation).toHaveBeenCalledExactlyOnceWith("a", draft);
    expect(hook.result.current.draft.id).toBe(draft.id);
    await act(async () => { pending.resolve(saved(draft)); await operation; });
    expect(hook.result.current.status).toBe("valuation.saved");
    expect(hook.result.current.draft.id).not.toBe(draft.id);
  });

  it.each([[400, "valuation.invalid"], [409, "valuation.conflict"], [404, "valuation.notFound"], [403, "valuation.forbidden"], [401, "valuation.forbidden"]])(
    "retains editable input after known rejection %s", async (code, key) => {
      const hook = await ready(), draft = hook.result.current.draft;
      vi.mocked(api.recordVehicleValuation).mockRejectedValueOnce(new ApiError("rejected", "problem", code));
      await act(() => hook.result.current.save());
      expect(hook.result.current.error).toBe(key);
      expect(hook.result.current.draft).toEqual(draft);
      expect(hook.result.current.uncertain).toBeNull();
    });

  it("does not resubmit an uncertain operation and confirms it by reading", async () => {
    const hook = await ready(), draft = hook.result.current.draft;
    vi.mocked(api.recordVehicleValuation).mockRejectedValueOnce(new Error("response lost"));
    await act(() => hook.result.current.save());
    expect(hook.result.current.uncertain).toEqual(draft);
    await act(() => hook.result.current.save());
    act(() => hook.result.current.reset());
    expect(api.recordVehicleValuation).toHaveBeenCalledOnce();
    expect(hook.result.current.draft).toEqual(draft);
    vi.mocked(api.vehicleValuationsExport).mockResolvedValueOnce(exported([saved(draft)]));
    await act(() => hook.result.current.review());
    expect(hook.result.current.uncertain).toBeNull();
    expect(hook.result.current.status).toBe("valuation.saved");
    expect(api.recordVehicleValuation).toHaveBeenCalledOnce();
  });

  it("unlocks an absent record only after explicit review, retaining its identity", async () => {
    const hook = await ready(), draft = hook.result.current.draft;
    vi.mocked(api.recordVehicleValuation).mockRejectedValueOnce(new Error("offline"));
    await act(() => hook.result.current.save());
    await act(() => hook.result.current.review());
    expect(hook.result.current.status).toBe("valuation.reviewMissing");
    expect(hook.result.current.draft).toEqual(draft);
    await act(() => hook.result.current.save());
    expect(api.recordVehicleValuation).toHaveBeenCalledTimes(2);
    expect(vi.mocked(api.recordVehicleValuation).mock.calls[1][1].id).toBe(draft.id);
  });

  it("keeps an uncertain input locked if review fails or belongs to another vehicle", async () => {
    const hook = await ready(), draft = hook.result.current.draft;
    vi.mocked(api.recordVehicleValuation).mockRejectedValueOnce(new Error("offline"));
    await act(() => hook.result.current.save());
    vi.mocked(api.vehicleValuationsExport).mockRejectedValueOnce(new Error("offline"));
    await act(() => hook.result.current.review());
    expect(hook.result.current.error).toBe("valuation.reviewFailed");
    vi.mocked(api.vehicleValuationsExport).mockResolvedValueOnce(exported([saved(draft, { vehicleId: "other" })]));
    await act(() => hook.result.current.review());
    expect(hook.result.current.uncertain).toEqual(draft);
    expect(hook.result.current.error).toBe("valuation.reviewFailed");
  });

  it("requires comparison when a response or reviewed identity has different contents", async () => {
    const hook = await ready(), draft = hook.result.current.draft;
    vi.mocked(api.recordVehicleValuation).mockResolvedValueOnce(saved(draft, { amount: "1.00" }));
    await act(() => hook.result.current.save());
    expect(hook.result.current.uncertain).toEqual(draft);
    vi.mocked(api.vehicleValuationsExport).mockResolvedValueOnce(exported([saved(draft, { amount: "1.00" })]));
    await act(() => hook.result.current.review());
    expect(hook.result.current.error).toBe("valuation.conflict");
    expect(hook.result.current.uncertain).toBeNull();
    expect(hook.result.current.draft).toEqual(draft);
  });

  it("rejects late loads and preserves displayed history on load failure", async () => {
    const pending = deferred<VehicleValuationHistory>();
    vi.mocked(api.vehicleValuations).mockReturnValueOnce(pending.promise);
    const hook = renderHook(() => useVehicleValuations("a", true));
    const entry = saved(hook.result.current.draft, { amount: "1.00" });
    vi.mocked(api.vehicleValuations).mockResolvedValueOnce({ entries: [entry], total: 1 });
    await act(() => hook.result.current.load());
    await act(async () => pending.resolve({ entries: [], total: 0 }));
    expect(hook.result.current.history.entries).toEqual([entry]);
    vi.mocked(api.vehicleValuations).mockRejectedValueOnce(new Error("offline"));
    await act(() => hook.result.current.load());
    expect(hook.result.current.history.entries).toEqual([entry]);
    expect(hook.result.current.loadFailed).toBe(true);
  });

  it("keeps corrections explicit and starts withdrawal with a required reason", async () => {
    const hook = await ready(), original = saved(hook.result.current.draft);
    act(() => hook.result.current.startCorrection(original, false));
    expect(hook.result.current.draft.replacesId).toBe(original.id);
    expect(hook.result.current.draft.id).not.toBe(original.id);
    act(() => hook.result.current.startCorrection(original, true));
    expect(hook.result.current.draft).toMatchObject({ withdrawn: true, source: "", amount: original.amount });
    const identity = hook.result.current.draft.id;
    act(() => hook.result.current.startCorrection({ ...original, replacedById: "new" }, false));
    expect(hook.result.current.draft.id).toBe(identity);
  });

  it("does not publish another vehicle's late mutation after unmount", async () => {
    const hook = await ready(), draft = hook.result.current.draft, pending = deferred<VehicleValuation>();
    vi.mocked(api.recordVehicleValuation).mockReturnValueOnce(pending.promise);
    let operation!: Promise<void>;
    act(() => { operation = hook.result.current.save(); });
    const changed = vi.fn(); window.addEventListener(valuationChangedEvent, changed);
    hook.unmount();
    await act(async () => { pending.resolve(saved(draft)); await operation; });
    expect(changed).not.toHaveBeenCalled();
    window.removeEventListener(valuationChangedEvent, changed);
  });

  it.each([false, true])("retains my correction when explicitly comparing a new predecessor, withdrawal=%s", async withdrawn => {
    const hook = await ready();
    const original = saved(hook.result.current.draft, { id: "1".repeat(32) });
    act(() => hook.result.current.startCorrection(original, withdrawn));
    act(() => hook.result.current.setDraft(draft => ({ ...draft, amount: "150.00", source: "My reason", notes: "My notes" })));
    vi.mocked(api.recordVehicleValuation).mockRejectedValueOnce(new ApiError("changed", "valuation_conflict", 409));
    await act(() => hook.result.current.save());
    const pending = hook.result.current.draft;
    const peer = { ...original, id: "2".repeat(32), amount: "120.00", currency: "CHF", valuedOn: "2026-02-01" };
    act(() => hook.result.current.rebaseCorrection({ ...peer, vehicleId: "other" }));
    expect(hook.result.current.draft).toEqual(pending);
    act(() => hook.result.current.rebaseCorrection(peer));
    expect(hook.result.current.draft).toMatchObject({ source: "My reason", notes: "My notes", replacesId: peer.id, withdrawn });
    expect(hook.result.current.draft.id).not.toBe(pending.id);
    expect(hook.result.current.draft.amount).toBe(withdrawn ? peer.amount : "150.00");
    expect(hook.result.current.draft.currency).toBe(withdrawn ? "CHF" : pending.currency);
    expect(hook.result.current.status).toBe("valuation.rebased");
    expect(api.recordVehicleValuation).toHaveBeenCalledOnce();
  });

  it("uses complete export coverage when an uncertain record is beyond the first page", async () => {
    const hook = await ready();
    const draft = hook.result.current.draft;
    const rows = Array.from({ length: 100 }, (_, index) => saved(draft, { id: (index + 1).toString(16).padStart(32, "0") }));
    vi.mocked(api.vehicleValuations).mockResolvedValueOnce({ entries: rows, total: 101 });
    await act(() => hook.result.current.load());
    vi.mocked(api.recordVehicleValuation).mockRejectedValueOnce(new Error("response lost"));
    await act(() => hook.result.current.save());
    vi.mocked(api.vehicleValuationsExport).mockResolvedValueOnce(exported([...rows, saved(draft)]));
    vi.mocked(api.vehicleValuations).mockRejectedValueOnce(new Error("refresh failed"));
    await act(() => hook.result.current.review());
    expect(hook.result.current.history.total).toBe(101);
    expect(hook.result.current.history.entries).toHaveLength(100);
    expect(hook.result.current.status).toBe("valuation.saved");
    expect(hook.result.current.loadFailed).toBe(true);
  });
});
