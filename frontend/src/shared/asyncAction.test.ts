import { afterEach, describe, expect, it, vi } from "vitest";
import { observeAction, reportUnexpectedAsyncError } from "./asyncAction";

afterEach(() => { vi.unstubAllGlobals(); vi.useRealTimers(); });

describe("asynchronous UI actions", () => {
  it("routes unexpected rejection to the existing browser diagnostics path", async () => {
    const report = vi.fn();
    vi.stubGlobal("reportError", report);
    const failure = new Error("Unexpected action failure");
    const work = vi.fn(async (value: { id: string }) => { expect(value.id).toBe("original"); throw failure; });
    const event = { id: "original" };
    expect(observeAction(work(event))).toBeUndefined();
    await Promise.resolve();
    expect(work).toHaveBeenCalledExactlyOnceWith(event);
    expect(report).toHaveBeenCalledExactlyOnceWith(failure);
  });

  it("retains controller-owned recovery and never reports a handled failure twice", async () => {
    const report = vi.fn();
    vi.stubGlobal("reportError", report);
    const draft = { text: "Unsaved original", error: "" };
    const save = async () => {
      try { await Promise.reject(new Error("Retry the save")); }
      catch (error) { draft.error = (error as Error).message; }
    };
    observeAction(save());
    await Promise.resolve();
    await Promise.resolve();
    expect(draft).toEqual({ text: "Unsaved original", error: "Retry the save" });
    expect(report).not.toHaveBeenCalled();
  });

  it("falls back to the normal uncaught-error path when reportError is unavailable", () => {
    vi.useFakeTimers();
    vi.stubGlobal("reportError", undefined);
    const failure = new Error("Visible to normal error capture");
    reportUnexpectedAsyncError(failure);
    expect(() => vi.runOnlyPendingTimers()).toThrow(failure);
  });
});
