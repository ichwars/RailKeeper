import { render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import type { TrackPlanAnalysis } from "../../shared/api";
import { TrackPlanAnalysisPanel } from "./TrackPlanAnalysisPanel";

const analysis: TrackPlanAnalysis = {
  revisionId: "revision-1",
  status: "draft",
  connections: [{ objectAId: "track-1", portAId: "b", objectBId: "track-2", portBId: "a" }],
  issues: [
    { code: "open_end", severity: "warning", objectIds: ["track-1"], portIds: ["a"] },
    { code: "open_end", severity: "warning", objectIds: ["track-2"], portIds: ["b"] },
    { code: "overlap", severity: "warning", objectIds: ["track-2", "track-3"] }
  ],
  bom: [{
    geometryId: "g1", libraryId: "tillig-v1", articleNumber: "83101", name: "Gleisstück G1", quantity: 3
  }],
  materials: [{
    geometryId: "g1", manufacturer: "Tillig", articleNumber: "83101", name: "Gleisstück G1",
    requiredQuantity: 3, productIds: ["product-1"], inventoryNumbers: ["RK-ART-000001"],
    physicalQuantity: 3, reservedQuantity: 1, availableQuantity: 2, missingQuantity: 1
  }]
};

describe("TrackPlanAnalysisPanel", () => {
  it("shows connection health and material demand with explicit symbols and values", () => {
    render(<TrackPlanAnalysisPanel analysis={analysis} />);

    expect(screen.getByText("1 Verbindung")).toBeInTheDocument();
    expect(screen.getByText("2 offene Enden")).toBeInTheDocument();
    expect(screen.getByText("1 Überschneidung")).toBeInTheDocument();
    expect(screen.getByLabelText("Warnung: Überschneidung")).toHaveTextContent("!");

    const row = screen.getByRole("row", { name: /Tillig 83101/ });
    expect(within(row).getByText("3")).toBeInTheDocument();
    expect(within(row).getByText("2")).toBeInTheDocument();
    expect(within(row).getByText("1")).toBeInTheDocument();
    expect(within(row).getByText("RK-ART-000001")).toBeInTheDocument();
  });
});
