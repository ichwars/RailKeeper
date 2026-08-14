import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type LayoutConfiguration, type ModulePortAnalysis } from "../../shared/api";
import { LayoutConfigurationPortAnalysis } from "./LayoutConfigurationPortAnalysis";

const configuration: LayoutConfiguration = {
  id: "configuration-1", layoutId: "layout-1", name: "Ausstellung", version: 1, archived: false,
  units: [], createdAt: "2026-08-10T10:00:00Z", updatedAt: "2026-08-10T10:00:00Z"
};

const analysis: ModulePortAnalysis = {
  connections: [{
    unitAId: "unit-a", unitAName: "Bahnhof", portAId: "port-a", portAName: "Ost",
    unitBId: "unit-b", unitBName: "Strecke", portBId: "port-b", portBName: "West"
  }],
  issues: [
    { code: "open_port", unitIds: ["unit-a"], unitNames: ["Bahnhof"],
      portIds: ["port-open"], portNames: ["West"] },
    { code: "incompatible_port", unitIds: ["unit-a", "unit-c"], unitNames: ["Bahnhof", "Abzweig"],
      portIds: ["port-power", "port-track"], portNames: ["Strom", "Gleis"] }
  ]
};

describe("LayoutConfigurationPortAnalysis", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(api, "layoutConfigurationPortAnalysis").mockResolvedValue(analysis);
  });

  it("loads and explains derived connections and issues", async () => {
    render(<LayoutConfigurationPortAnalysis configuration={configuration} />);

    expect(await screen.findByText("Bahnhof · Ost ↔ Strecke · West")).toBeInTheDocument();
    expect(screen.getByText("1 Verbindung")).toBeInTheDocument();
    expect(screen.getByText("1 offener Port")).toBeInTheDocument();
    expect(screen.getByText("1 inkompatibler Übergang")).toBeInTheDocument();
    expect(screen.getByText("Bahnhof · West")).toBeInTheDocument();
    expect(screen.getByText("Bahnhof · Strom ↔ Abzweig · Gleis")).toBeInTheDocument();
  });

  it("shows a stable no-selection state without loading", () => {
    render(<LayoutConfigurationPortAnalysis configuration={null} />);

    expect(screen.getByText("Aufbau auswählen, um die Modulports zu prüfen.")).toBeInTheDocument();
    expect(api.layoutConfigurationPortAnalysis).not.toHaveBeenCalled();
  });
});
