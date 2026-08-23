import { render, screen, waitFor, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../shared/api";
import { Shell } from "./Shell";

describe("Shell article navigation", () => {
  beforeEach(() => {
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
    vi.spyOn(api, "version").mockResolvedValue({
      version: "1.0.0",
      updateAvailable: false,
      checkedAt: "2026-08-08T10:00:00Z",
      status: "current",
      message: "current"
    });
  });

  it("uses the approved vehicle and accessory navigation labels", async () => {
    render(
      <Shell username="editor" roles={["Editor"]} activeView="accessories" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    const navigation = screen.getByRole("navigation", { name: "Hauptnavigation" });
    expect(navigation).toHaveTextContent("Fahrzeugbestand");
    expect(navigation).toHaveTextContent("Zubehör");
    expect(navigation).not.toHaveTextContent(/^Bestand$/);
    expect(navigation).not.toHaveTextContent("Artikelübersicht");
    await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
  });

  it("does not expose accessories to Messe users", () => {
    render(
      <Shell username="messe" roles={["Messe"]} activeView="exhibition" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    expect(screen.queryByRole("link", { name: "Zubehör" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Ausstellung" })).toBeInTheDocument();
  });

  it("exposes Import/Export to pure Messe users", () => {
    render(
      <Shell username="messe" roles={["Messe"]} activeView="importExport" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    expect(screen.getByRole("link", { name: "Import/Export" }))
      .toHaveAttribute("href", "/import-export");
  });

  it("uses the English accessories navigation label", async () => {
    window.localStorage.setItem("railkeeper.settings.language", "en");
    render(
      <Shell username="editor" roles={["Editor"]} activeView="accessories" onLogout={vi.fn()}>
        <p>Content</p>
      </Shell>
    );

    expect(screen.getByRole("link", { name: "Accessories" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Article overview" })).not.toBeInTheDocument();
    await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
  });

  it("does not expose the layout workspace in the navigation", async () => {
    render(
      <Shell username="editor" roles={["Editor"]} activeView="accessories" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    expect(screen.queryByRole("link", { name: "Anlage" })).not.toBeInTheDocument();
    await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
  });

  it("does not expose the layout workspace in the English navigation", async () => {
    window.localStorage.setItem("railkeeper.settings.language", "en");
    render(
      <Shell username="editor" roles={["Editor"]} activeView="accessories" onLogout={vi.fn()}>
        <p>Content</p>
      </Shell>
    );

    expect(screen.queryByRole("link", { name: "Layout" })).not.toBeInTheDocument();
    await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
  });

  it("links the Digital Centers navigation directly to its workspace", async () => {
    render(
      <Shell username="admin" roles={["Admin"]} activeView="digitalCenters" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    expect(screen.getByRole("link", { name: "Digitalzentralen" }))
      .toHaveAttribute("href", "/digital-centers");
    await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
  });

  it("places command stations directly before settings in the default navigation", () => {
    render(
      <Shell username="admin" roles={["Admin"]} activeView="overview" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    const navigation = screen.getByRole("navigation", { name: "Hauptnavigation" });
    const entries = Array.from(navigation.querySelectorAll(".nav-entry"));
    expect(entries.slice(-2)).toEqual([
      within(navigation).getByRole("link", { name: "Digitalzentralen" }),
      within(navigation).getByRole("link", { name: "Einstellungen" })
    ]);
  });

  it("migrates an older saved default order so command stations stay before settings", () => {
    window.localStorage.setItem("railkeeper.settings.sidebarPrefs:admin", JSON.stringify({
      order: ["overview", "vehicles", "accessories", "layouts", "exhibition", "importExport", "settings"],
      hidden: []
    }));

    render(
      <Shell username="admin" roles={["Admin"]} activeView="overview" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    const navigation = screen.getByRole("navigation", { name: "Hauptnavigation" });
    const entries = Array.from(navigation.querySelectorAll(".nav-entry"));
    expect(within(navigation).queryByRole("link", { name: "Anlage" })).not.toBeInTheDocument();
    expect(entries.slice(-2)).toEqual([
      within(navigation).getByRole("link", { name: "Digitalzentralen" }),
      within(navigation).getByRole("link", { name: "Einstellungen" })
    ]);
  });
});
