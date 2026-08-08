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

  it("uses the approved vehicle and article navigation labels", async () => {
    render(
      <Shell username="editor" roles={["Editor"]} activeView="accessories" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    const navigation = screen.getByRole("navigation", { name: "Hauptnavigation" });
    expect(navigation).toHaveTextContent("Fahrzeugbestand");
    expect(navigation).toHaveTextContent("Artikelübersicht");
    expect(navigation).not.toHaveTextContent(/^Bestand$/);
    expect(navigation).not.toHaveTextContent(/^Zubehör$/);
    await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
  });

  it("does not expose the article overview to Messe users", () => {
    render(
      <Shell username="messe" roles={["Messe"]} activeView="exhibition" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    expect(screen.queryByRole("link", { name: "Artikelübersicht" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Ausstellung" })).toBeInTheDocument();
  });
});
