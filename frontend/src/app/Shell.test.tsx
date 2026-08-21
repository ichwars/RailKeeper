import { render, screen, waitFor } from "@testing-library/react";
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
});
