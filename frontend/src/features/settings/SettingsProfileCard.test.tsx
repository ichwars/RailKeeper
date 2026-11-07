import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { readSidebarPrefs, sidebarPrefsKey } from "../../app/sidebarPreferences";
import { SettingsProfileCard } from "./SettingsProfileCard";
import { useSettingsPreferences } from "./useSettingsPreferences";

function Profile({ username = "editor" }: { username?: string }) {
  return <SettingsProfileCard preferences={useSettingsPreferences(username)} />;
}

describe("explicit personal view selection", () => {
  beforeEach(() => {
    setLanguage("de");
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
    vi.spyOn(api, "updateProfileSettings").mockResolvedValue({ settings: {} });
  });

  it("retains existing settings until a preset is chosen and writes only navigation", async () => {
    const user = userEvent.setup();
    const prefs = { order: ["vehicles", "overview", "settings"], hidden: ["accessories"] };
    window.localStorage.setItem(sidebarPrefsKey("editor"), JSON.stringify(prefs));
    window.localStorage.setItem("railkeeper.vehicleColumns:editor", "existing-columns");
    render(<Profile />);
    await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
    expect(api.updateProfileSettings).not.toHaveBeenCalled();
    const original = readSidebarPrefs("editor");
    await user.click(screen.getByRole("button", { name: "Sammlung" }));
    expect(readSidebarPrefs("editor").order).toEqual(original.order);
    expect(readSidebarPrefs("editor").hidden).toEqual(["exhibition", "importExport", "digitalCenters"]);
    expect(window.localStorage.getItem("railkeeper.vehicleColumns:editor")).toBe("existing-columns");
    expect(screen.getByRole("button", { name: "Sammlung" })).toHaveAttribute("aria-pressed", "true");
    expect(api.updateProfileSettings).toHaveBeenCalledExactlyOnceWith({
      [sidebarPrefsKey("editor")]: JSON.stringify(readSidebarPrefs("editor"))
    });
    await user.click(screen.getByRole("button", { name: "Alle Funktionen" }));
    expect(readSidebarPrefs("editor")).toEqual({ order: original.order, hidden: [] });
  });

  it("does not overwrite a new choice with a delayed initial server response", async () => {
    let finish!: (value: { settings: Record<string, string> }) => void;
    vi.mocked(api.profileSettings).mockReturnValue(new Promise((resolve) => { finish = resolve; }));
    const user = userEvent.setup();
    render(<Profile />);
    await user.click(screen.getByRole("button", { name: "Sammlung" }));
    await act(async () => finish({ settings: { [sidebarPrefsKey("editor")]: JSON.stringify({ order: [], hidden: [] }) } }));
    await waitFor(() => expect(screen.getByRole("button", { name: "Sammlung" })).toHaveAttribute("aria-pressed", "true"));
    expect(readSidebarPrefs("editor").hidden).toHaveLength(3);
  });

  it("preserves the local view after a failed save and offers an explicit retry", async () => {
    vi.mocked(api.updateProfileSettings).mockRejectedValueOnce(new Error("offline"));
    const user = userEvent.setup();
    render(<Profile />);
    await user.click(screen.getByRole("button", { name: "Sammlung" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("lokal aktiv");
    expect(readSidebarPrefs("editor").hidden).toHaveLength(3);
    await user.click(screen.getByRole("button", { name: "Ansicht erneut speichern" }));
    await waitFor(() => expect(screen.queryByRole("alert")).not.toBeInTheDocument());
    expect(api.updateProfileSettings).toHaveBeenCalledTimes(2);
  });

  it("serializes rapid choices so a slow earlier save cannot win", async () => {
    let finish!: (value: { settings: Record<string, string> }) => void;
    vi.mocked(api.updateProfileSettings).mockReturnValueOnce(new Promise((resolve) => { finish = resolve; }));
    const user = userEvent.setup();
    render(<Profile />);
    await user.click(screen.getByRole("button", { name: "Sammlung" }));
    await user.click(screen.getByRole("button", { name: "Alle Funktionen" }));
    expect(api.updateProfileSettings).toHaveBeenCalledOnce();
    await act(async () => finish({ settings: {} }));
    await waitFor(() => expect(api.updateProfileSettings).toHaveBeenCalledTimes(2));
    const final = vi.mocked(api.updateProfileSettings).mock.calls[1][0];
    expect(JSON.parse(final[sidebarPrefsKey("editor")]).hidden).toEqual([]);
  });

  it("does not send an old queued choice after the signed-in user changes", async () => {
    let finish!: (value: { settings: Record<string, string> }) => void;
    vi.mocked(api.updateProfileSettings).mockReturnValueOnce(new Promise(resolve => { finish = resolve; }));
    const user = userEvent.setup();
    const view = render(<Profile />);
    await user.click(screen.getByRole("button", { name: "Sammlung" }));
    await user.click(screen.getByRole("button", { name: "Alle Funktionen" }));
    view.rerender(<Profile username="other-user" />);
    await user.click(screen.getByRole("button", { name: "Sammlung" }));
    await act(async () => finish({ settings: {} }));
    await waitFor(() => expect(api.updateProfileSettings).toHaveBeenCalledTimes(2));
    const last = vi.mocked(api.updateProfileSettings).mock.calls[1][0];
    expect(Object.keys(last)).toEqual([sidebarPrefsKey("other-user")]);
  });

  it("invalidates queued saves after unmount and a new login", async () => {
    let finish!: (value: { settings: Record<string, string> }) => void;
    vi.mocked(api.updateProfileSettings).mockReturnValueOnce(new Promise(resolve => { finish = resolve; }));
    const user = userEvent.setup();
    const view = render(<Profile />);
    await user.click(screen.getByRole("button", { name: "Sammlung" }));
    await user.click(screen.getByRole("button", { name: "Alle Funktionen" }));
    view.unmount();
    vi.spyOn(globalThis,"fetch").mockResolvedValue(new Response(JSON.stringify({ csrfToken:"different-synthetic-session" }), { status:200 }));
    await api.login({ username:"other-user", password:"synthetic-password" });
    await act(async () => finish({ settings:{} }));
    expect(api.updateProfileSettings).toHaveBeenCalledOnce();
  });
});
