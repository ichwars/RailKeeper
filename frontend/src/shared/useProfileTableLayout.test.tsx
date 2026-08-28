import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api } from "./api";
import { useProfileTableLayout } from "./useProfileTableLayout";

type Layout = { value: number };

const parse = (raw: string | undefined): Layout => {
  if (!raw) return { value: 10 };
  try {
    const parsed = JSON.parse(raw) as { value?: unknown };
    return { value: typeof parsed.value === "number" ? parsed.value : 10 };
  } catch {
    return { value: 10 };
  }
};

describe("useProfileTableLayout", () => {
  afterEach(() => vi.restoreAllMocks());

  it("loads the profile and separates transient previews from persisted commits", async () => {
    vi.spyOn(api, "profileSettings").mockResolvedValue({
      settings: { "table.layout": '{"value":24}' }
    });
    const update = vi.spyOn(api, "updateProfileSettings").mockResolvedValue({ settings: {} });

    const { result } = renderHook(() => useProfileTableLayout<Layout>({
      settingKey: "table.layout",
      defaultLayout: { value: 10 },
      parse,
      serialize: JSON.stringify,
      onLoadError: vi.fn(),
      onSaveError: vi.fn()
    }));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.layout).toEqual({ value: 24 });

    act(() => result.current.preview((current) => ({ value: current.value + 1 })));
    expect(result.current.layout).toEqual({ value: 25 });
    expect(update).not.toHaveBeenCalled();

    act(() => result.current.commit(() => ({ value: 26 })));
    await waitFor(() => expect(update).toHaveBeenCalledWith({
      "table.layout": '{"value":26}'
    }));
  });

  it("replays a local commit onto the stored layout after loading", async () => {
    let resolveProfile: ((value: { settings: Record<string, string> }) => void) | undefined;
    vi.spyOn(api, "profileSettings").mockImplementation(() => new Promise((resolve) => {
      resolveProfile = resolve;
    }));
    const update = vi.spyOn(api, "updateProfileSettings").mockResolvedValue({ settings: {} });

    type ProfileLayout = { columns: string[]; widths: Record<string, number> };
    const parseProfileLayout = (raw: string | undefined): ProfileLayout => raw
      ? JSON.parse(raw) as ProfileLayout
      : { columns: ["default"], widths: {} };

    const { result } = renderHook(() => useProfileTableLayout<ProfileLayout>({
      settingKey: "table.layout",
      defaultLayout: { columns: ["default"], widths: {} },
      parse: parseProfileLayout,
      serialize: JSON.stringify,
      onLoadError: vi.fn(),
      onSaveError: vi.fn()
    }));

    act(() => result.current.commit((current) => ({
      ...current,
      widths: { ...current.widths, local: 44 }
    })));
    expect(result.current.layout).toEqual({ columns: ["default"], widths: { local: 44 } });
    expect(update).not.toHaveBeenCalled();

    act(() => resolveProfile?.({
      settings: {
        "table.layout": '{"columns":["stored"],"widths":{"stored":24}}'
      }
    }));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.layout).toEqual({
      columns: ["stored"],
      widths: { stored: 24, local: 44 }
    });
    await waitFor(() => expect(update).toHaveBeenCalledWith({
      "table.layout": '{"columns":["stored"],"widths":{"stored":24,"local":44}}'
    }));
  });

  it("migrates a legacy browser value only when the profile has no value", async () => {
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
    const update = vi.spyOn(api, "updateProfileSettings").mockResolvedValue({ settings: {} });

    const { result } = renderHook(() => useProfileTableLayout<Layout>({
      settingKey: "table.layout",
      defaultLayout: { value: 10 },
      parse,
      serialize: JSON.stringify,
      legacyValue: () => '{"value":33}',
      onLoadError: vi.fn(),
      onSaveError: vi.fn()
    }));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.layout).toEqual({ value: 33 });
    await waitFor(() => expect(update).toHaveBeenCalledWith({
      "table.layout": '{"value":33}'
    }));
  });

  it("reports load and queued save failures without losing local state", async () => {
    const onLoadError = vi.fn();
    const onSaveError = vi.fn();
    vi.spyOn(api, "profileSettings").mockRejectedValue(new Error("offline"));
    vi.spyOn(api, "updateProfileSettings").mockRejectedValue(new Error("offline"));

    const { result } = renderHook(() => useProfileTableLayout<Layout>({
      settingKey: "table.layout",
      defaultLayout: { value: 10 },
      parse,
      serialize: JSON.stringify,
      onLoadError,
      onSaveError
    }));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(onLoadError).toHaveBeenCalledOnce();

    act(() => result.current.commit(() => ({ value: 44 })));
    await waitFor(() => expect(onSaveError).toHaveBeenCalledOnce());
    expect(result.current.layout).toEqual({ value: 44 });
  });
});
