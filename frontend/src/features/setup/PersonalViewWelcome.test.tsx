import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { defaultSidebarOrder, sidebarPrefsKey } from "../../app/sidebarPreferences";
import { PersonalViewWelcome } from "./PersonalViewWelcome";

const key = sidebarPrefsKey("first-owner");

describe("personal view offered after setup", () => {
  beforeEach(() => {
    setLanguage("de");
    vi.spyOn(api,"profileSettings").mockResolvedValue({ settings: {} });
    vi.spyOn(api,"updateProfileSettings").mockResolvedValue({ settings: {} });
  });

  it("offers Collection without writing until explicitly confirmed and preserves unrelated data", async () => {
    const complete = vi.fn();
    window.localStorage.setItem("railkeeper.vehicleColumns:first-owner", "retained columns");
    render(<PersonalViewWelcome username="first-owner" onComplete={complete} />);
    expect(screen.getByRole("radio", { name: /Sammlung/ })).toBeChecked();
    const start = screen.getByRole("button",{name:"Mit dieser Ansicht starten"});
    await waitFor(() => expect(start).toBeEnabled());
    expect(api.updateProfileSettings).not.toHaveBeenCalled();
    expect(window.localStorage.getItem(key)).toBeNull();
    await userEvent.click(start);
    await waitFor(() => expect(complete).toHaveBeenCalledOnce());
    expect(api.updateProfileSettings).toHaveBeenCalledExactlyOnceWith({[key]:JSON.stringify({order:defaultSidebarOrder,hidden:["exhibition","importExport","digitalCenters"]})});
    expect(window.localStorage.getItem("railkeeper.vehicleColumns:first-owner")).toBe("retained columns");
  });

  it("loads the server order before applying an explicit all-functions choice", async () => {
    const order=[...defaultSidebarOrder].reverse();
    vi.mocked(api.profileSettings).mockResolvedValue({settings:{[key]:JSON.stringify({order,hidden:["vehicles"]})}});
    const complete=vi.fn();
    render(<PersonalViewWelcome username="first-owner" onComplete={complete} />);
    await waitFor(() => expect(screen.getByRole("button",{name:"Mit dieser Ansicht starten"})).toBeEnabled());
    await userEvent.click(screen.getByRole("radio",{name:/Alle Funktionen/}));
    await userEvent.click(screen.getByRole("button",{name:"Mit dieser Ansicht starten"}));
    await waitFor(() => expect(complete).toHaveBeenCalledOnce());
    expect(api.updateProfileSettings).toHaveBeenCalledExactlyOnceWith({[key]:JSON.stringify({order,hidden:[]})});
  });

  it("retains the current view when deferred, including while preferences cannot be loaded", async () => {
    vi.mocked(api.profileSettings).mockRejectedValue(new Error("offline"));
    window.localStorage.setItem(key,"existing setting");
    const complete=vi.fn();
    render(<PersonalViewWelcome username="first-owner" onComplete={complete} />);
    await screen.findByRole("alert");
    expect(screen.getByRole("button",{name:"Mit dieser Ansicht starten"})).toBeDisabled();
    await userEvent.click(screen.getByRole("button",{name:"Vorerst bisherige Ansicht behalten"}));
    expect(complete).toHaveBeenCalledOnce();
    expect(api.updateProfileSettings).not.toHaveBeenCalled();
    expect(window.localStorage.getItem(key)).toBe("existing setting");
  });

  it("retains the old setting after a failed save and retries only the selected navigation", async () => {
    const previous=JSON.stringify({order:defaultSidebarOrder,hidden:["accessories"]});
    window.localStorage.setItem(key,previous);
    vi.mocked(api.updateProfileSettings).mockRejectedValueOnce(new Error("offline"));
    const complete=vi.fn();
    render(<PersonalViewWelcome username="first-owner" onComplete={complete} />);
    const start=screen.getByRole("button",{name:"Mit dieser Ansicht starten"});
    await waitFor(() => expect(start).toBeEnabled());
    await userEvent.click(start);
    await screen.findByRole("alert");
    expect(window.localStorage.getItem(key)).toBe(previous);
    expect(complete).not.toHaveBeenCalled();
    await userEvent.click(start);
    await waitFor(() => expect(complete).toHaveBeenCalledOnce());
    expect(api.updateProfileSettings).toHaveBeenCalledTimes(2);
  });

  it("reloads after a read error without overwriting a delayed server response", async () => {
    vi.mocked(api.profileSettings).mockRejectedValueOnce(new Error("offline"));
    const complete=vi.fn();
    render(<PersonalViewWelcome username="first-owner" onComplete={complete} />);
    await screen.findByRole("alert");
    await userEvent.click(screen.getByRole("button",{name:"Aktualisieren"}));
    await waitFor(() => expect(screen.getByRole("button",{name:"Mit dieser Ansicht starten"})).toBeEnabled());
    expect(api.profileSettings).toHaveBeenCalledTimes(2);
    expect(api.updateProfileSettings).not.toHaveBeenCalled();
  });

  it.each(["invalid JSON", "null", JSON.stringify({order:["removed",null],hidden:["settings","digitalCenters"]})])("keeps a usable selection with old malformed preferences (%s)",async value => {
    vi.mocked(api.profileSettings).mockResolvedValue({settings:{[key]:value}});
    render(<PersonalViewWelcome username="first-owner" onComplete={vi.fn()} />);
    await waitFor(() => expect(screen.getByRole("button",{name:"Mit dieser Ansicht starten"})).toBeEnabled());
    await userEvent.click(screen.getByRole("button",{name:"Mit dieser Ansicht starten"}));
    expect(api.updateProfileSettings).toHaveBeenCalledOnce();
  });

  it("does not finish an unmounted welcome after a delayed save",async () => {
    let finish!: () => void;
    vi.mocked(api.updateProfileSettings).mockReturnValue(new Promise(resolve=>{finish=()=>resolve({settings:{}});}));
    const complete=vi.fn();
    const {unmount}=render(<PersonalViewWelcome username="first-owner" onComplete={complete} />);
    const start=screen.getByRole("button",{name:"Mit dieser Ansicht starten"});
    await waitFor(() => expect(start).toBeEnabled());
    await userEvent.click(start);
    expect(screen.getByRole("button",{name:"Vorerst bisherige Ansicht behalten"})).toBeDisabled();
    unmount();
    await act(async()=>finish());
    expect(complete).not.toHaveBeenCalled();
  });

  it("offers the same explicit choices in English", async () => {
    setLanguage("en");
    render(<PersonalViewWelcome username="first-owner" onComplete={vi.fn()} />);
    expect(screen.getByRole("heading",{name:"Your start in RailKeeper"})).toBeInTheDocument();
    expect(screen.getByRole("radio",{name:/Collection/})).toBeChecked();
    await waitFor(() => expect(screen.getByRole("button",{name:"Start with this view"})).toBeEnabled());
  });
});
