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

  it("shows the singular layout item but does not expose it as a link", async () => {
    render(
      <Shell username="editor" roles={["Editor"]} activeView="accessories" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    const label = screen.getByText("Anlage");
    const disabledItem = label.closest("[aria-disabled='true']");
    expect(disabledItem).toHaveClass("disabled");
    expect(disabledItem).toHaveAttribute("title", "Anlage ist vorübergehend nicht verfügbar.");
    expect(screen.queryByRole("link", { name: "Anlage" })).not.toBeInTheDocument();
    await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
  });

  it("uses the singular English layout label and hint", async () => {
    window.localStorage.setItem("railkeeper.settings.language", "en");
    render(
      <Shell username="editor" roles={["Editor"]} activeView="accessories" onLogout={vi.fn()}>
        <p>Content</p>
      </Shell>
    );

    const disabledItem = screen.getByText("Layout").closest("[aria-disabled='true']");
    expect(disabledItem).toHaveAttribute("title", "Layout is temporarily unavailable.");
    expect(screen.queryByRole("link", { name: "Layout" })).not.toBeInTheDocument();
    await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
  });
});
