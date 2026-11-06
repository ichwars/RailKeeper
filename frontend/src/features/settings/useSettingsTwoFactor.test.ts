import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { api } from "../../shared/api";
import { useSettingsTwoFactor } from "./useSettingsTwoFactor";

const options = (canManageUsers: boolean) => ({
  canManageUsers,
  loadCurrentSession: vi.fn(),
  loadUsersAndRoles: vi.fn(),
  loadSessions: vi.fn(),
  loadAuditLog: vi.fn()
});

describe("settings two-factor request boundaries", () => {
  it.each([false, true])("validates enable codes and refreshes only permitted admin data (%s)", async (admin) => {
    vi.spyOn(api, "enableTwoFactor").mockResolvedValue({ enabled: true, prepared: true });
    vi.spyOn(api, "twoFactorStatus").mockResolvedValue({ enabled: true, prepared: true });
    const reload = options(admin);
    const { result } = renderHook(() => useSettingsTwoFactor(reload));
    act(() => result.current.enableTwoFactor());
    expect(api.enableTwoFactor).not.toHaveBeenCalled();
    act(() => result.current.setTwoFactorCode(" 123456 "));
    act(() => result.current.enableTwoFactor());
    await waitFor(() => expect(result.current.twoFactorSaving).toBe(false));
    expect(api.enableTwoFactor).toHaveBeenCalledExactlyOnceWith({ code: "123456" });
    expect(result.current.twoFactorCode).toBe("");
    expect(reload.loadCurrentSession).toHaveBeenCalledOnce();
    expect(api.twoFactorStatus).toHaveBeenCalledOnce();
    expect(reload.loadUsersAndRoles).toHaveBeenCalledTimes(admin ? 1 : 0);
    expect(reload.loadAuditLog).toHaveBeenCalledTimes(admin ? 1 : 0);
  });

  it("requires the current password and retains inputs on failure before a successful disable", async () => {
    vi.spyOn(api, "disableTwoFactor")
      .mockRejectedValueOnce(new Error("Invalid credentials"))
      .mockResolvedValueOnce(undefined);
    const reload = options(true);
    const { result } = renderHook(() => useSettingsTwoFactor(reload));
    act(() => result.current.disableTwoFactor());
    expect(api.disableTwoFactor).not.toHaveBeenCalled();
    act(() => {
      result.current.setTwoFactorPassword("synthetic-password");
      result.current.setTwoFactorCode("654321");
    });
    act(() => result.current.disableTwoFactor());
    await waitFor(() => expect(result.current.twoFactorMessage).toBe("Invalid credentials"));
    expect(result.current.twoFactorPassword).toBe("synthetic-password");
    expect(result.current.twoFactorCode).toBe("654321");
    expect(reload.loadCurrentSession).not.toHaveBeenCalled();
    act(() => result.current.disableTwoFactor());
    await waitFor(() => expect(result.current.twoFactorPassword).toBe(""));
    expect(api.disableTwoFactor).toHaveBeenLastCalledWith({
      currentPassword: "synthetic-password", code: "654321"
    });
    expect(result.current.twoFactorCode).toBe("");
    expect(result.current.twoFactorStatus).toEqual({ enabled: false, prepared: false });
    expect(reload.loadCurrentSession).toHaveBeenCalledOnce();
    expect(reload.loadUsersAndRoles).toHaveBeenCalledOnce();
    expect(reload.loadSessions).toHaveBeenCalledOnce();
    expect(reload.loadAuditLog).toHaveBeenCalledOnce();
  });
});
