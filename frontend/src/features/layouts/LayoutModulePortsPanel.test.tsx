import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type LayoutUnit, type LayoutUnitPort } from "../../shared/api";
import { LayoutModulePortsPanel } from "./LayoutModulePortsPanel";

const unit: LayoutUnit = {
  id: "unit-1", layoutId: "layout-1", name: "Bahnhofsmodul", kind: "module",
  widthMm: 1000, heightMm: 500, version: 1, archived: false,
  createdAt: "2026-08-10T10:00:00Z", updatedAt: "2026-08-10T10:00:00Z"
};
const port: LayoutUnitPort = {
  id: "port-1", layoutUnitId: unit.id, name: "West", kind: "track",
  interfaceKey: "track:tillig-tt-modellgleis-mit-einer-langen-dokumentierten-schnittstellenkennung",
  xMm: 0, yMm: 250, directionDegrees: 180, notes: "Hauptstrecke", version: 1,
  archived: false, createdAt: "2026-08-10T10:00:00Z", updatedAt: "2026-08-10T10:00:00Z"
};

describe("LayoutModulePortsPanel", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(api, "layoutUnitPorts").mockResolvedValue([port]);
  });

  it("loads the selected unit and creates a normalized module port with app controls", async () => {
    const user = userEvent.setup();
    const create = vi.spyOn(api, "createLayoutUnitPort").mockResolvedValue(port);
    render(<LayoutModulePortsPanel unit={unit} canPlan />);

    expect(await screen.findByText(port.interfaceKey)).toBeInTheDocument();
    expect(api.layoutUnitPorts).toHaveBeenCalledWith(unit.id);
    await user.click(screen.getByRole("button", { name: "Neuer Port" }));
    await user.type(screen.getByRole("textbox", { name: "Bezeichnung" }), "Ost");
    await user.click(screen.getByRole("button", { name: "Art" }));
    await user.click(screen.getByRole("option", { name: "Stromversorgung" }));
    await user.type(screen.getByRole("textbox", { name: "Schnittstelle" }), "power:16v-ac");
    await user.clear(screen.getByRole("spinbutton", { name: "X (mm)" }));
    await user.type(screen.getByRole("spinbutton", { name: "X (mm)" }), "1000");
    await user.clear(screen.getByRole("spinbutton", { name: "Y (mm)" }));
    await user.type(screen.getByRole("spinbutton", { name: "Y (mm)" }), "250");
    await user.click(screen.getByRole("button", { name: "Port speichern" }));

    await waitFor(() => expect(create).toHaveBeenCalledWith(unit.id, expect.objectContaining({
      name: "Ost", kind: "power", interfaceKey: "power:16v-ac", xMm: 1000, yMm: 250
    })));
  });

  it("updates and archives a selected port with its expected version", async () => {
    const user = userEvent.setup();
    const update = vi.spyOn(api, "updateLayoutUnitPort").mockResolvedValue({ ...port, archived: true, version: 2 });
    render(<LayoutModulePortsPanel unit={unit} canPlan />);

    await user.click(await screen.findByRole("button", { name: "West bearbeiten" }));
    await user.click(screen.getByRole("checkbox", { name: "Archiviert" }));
    await user.click(screen.getByRole("button", { name: "Port speichern" }));

    await waitFor(() => expect(update).toHaveBeenCalledWith(port.id, expect.objectContaining({
      expectedVersion: 1, archived: true
    })));
  });

  it("keeps the list read-only and shows the no-selection state", async () => {
    const { rerender } = render(<LayoutModulePortsPanel unit={unit} canPlan={false} />);
    expect(await screen.findByText("West")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Neuer Port" })).not.toBeInTheDocument();

    rerender(<LayoutModulePortsPanel unit={null} canPlan={false} />);
    expect(screen.getByText("Bitte zuerst eine Anlageneinheit auswählen.")).toBeInTheDocument();
  });
});
