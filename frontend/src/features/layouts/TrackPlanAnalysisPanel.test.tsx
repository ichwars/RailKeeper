import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { TrackPlanAnalysis } from "../../shared/api";
import { TrackPlanAnalysisPanel } from "./TrackPlanAnalysisPanel";

const analysis: TrackPlanAnalysis = {
  revisionId: "revision-1",
  status: "draft",
  connections: [{ objectAId: "track-1", portAId: "b", objectBId: "track-2", portBId: "a" }],
  issues: [
    { code: "open_end", severity: "warning", objectIds: ["track-1"], portIds: ["a"] },
    { code: "open_end", severity: "warning", objectIds: ["track-2"], portIds: ["b"] },
    { code: "overlap", severity: "warning", objectIds: ["track-2", "track-3"] },
    {
      code: "elevation_mismatch", severity: "warning", objectIds: ["track-1", "track-2"],
      portIds: ["b", "a"], elevationDifferenceMm: 2
    }
  ],
  bom: [{
    geometryId: "g1", libraryId: "tillig-v1", articleNumber: "83101", name: "Gleisstück G1", quantity: 3
  }],
  grades: [],
  materials: [{
    geometryId: "g1", manufacturer: "Tillig", articleNumber: "83101", name: "Gleisstück G1",
    requiredQuantity: 3, productIds: ["product-1"], inventoryNumbers: ["RK-ART-000001"],
    physicalQuantity: 3, reservedQuantity: 1, availableQuantity: 2, missingQuantity: 1
  }],
  reservations: []
};

describe("TrackPlanAnalysisPanel", () => {
  it("shows connection health and material demand with explicit symbols and values", () => {
    render(<TrackPlanAnalysisPanel analysis={analysis} />);

    expect(screen.getByText("1 Verbindung")).toBeInTheDocument();
    expect(screen.getByText("2 offene Enden")).toBeInTheDocument();
    expect(screen.getByText("1 Überschneidung")).toBeInTheDocument();
    expect(screen.getByText("1 Höhenversatz")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Warnung: Überschneidung" })).toHaveTextContent("!");
    expect(screen.getByRole("button", {
      name: "Warnung: Höhenversatz an Gleisverbindung (2,00 mm)"
    })).toBeInTheDocument();

    const row = screen.getByRole("row", { name: /Tillig 83101/ });
    const cells = within(row).getAllByRole("cell");
    expect(cells[1]).toHaveTextContent("3");
    expect(cells[2]).toHaveTextContent("3");
    expect(cells[3]).toHaveTextContent("1");
    expect(cells[4]).toHaveTextContent("2");
    expect(cells[5]).toHaveTextContent("! 1");
    expect(within(row).getByText("RK-ART-000001")).toBeInTheDocument();
  });

  it("selects affected track objects from validation issues", async () => {
    const user = userEvent.setup();
    const selectObject = vi.fn();
    render(<TrackPlanAnalysisPanel analysis={analysis} selectedObjectId="track-2"
      onSelectObject={selectObject} />);

    const overlap = screen.getByRole("button", { name: "Warnung: Überschneidung" });
    expect(overlap).toHaveAttribute("aria-pressed", "true");
    await user.click(overlap);
    expect(selectObject).toHaveBeenCalledWith("track-2");
  });

  it("selects the first affected track from an elevation mismatch", async () => {
    const user = userEvent.setup();
    const selectObject = vi.fn();
    render(<TrackPlanAnalysisPanel analysis={analysis} onSelectObject={selectObject} />);

    await user.click(screen.getByRole("button", {
      name: "Warnung: Höhenversatz an Gleisverbindung (2,00 mm)"
    }));
    expect(selectObject).toHaveBeenCalledWith("track-1");
  });

  it("shows a valid empty state and an unmatched catalog requirement", () => {
    render(<TrackPlanAnalysisPanel analysis={{
      revisionId: "revision-1", status: "draft", connections: [], issues: [],
      bom: [{ geometryId: "g2", libraryId: "tillig-v1", articleNumber: "83102", name: "G2", quantity: 1 }],
      grades: [],
      materials: [{
        geometryId: "g2", manufacturer: "Tillig", articleNumber: "83102", name: "G2",
        requiredQuantity: 1, productIds: [], inventoryNumbers: [], physicalQuantity: 0,
        reservedQuantity: 0, availableQuantity: 0, missingQuantity: 1
      }],
      reservations: []
    }} />);

    expect(screen.getByText("0 Verbindungen")).toBeInTheDocument();
    expect(screen.getByText("0 Höhenversätze")).toBeInTheDocument();
    expect(screen.getByText("✓ Keine Prüfhinweise")).toBeInTheDocument();
    expect(screen.getByText("Kein Artikel zugeordnet")).toBeInTheDocument();
  });
});
