import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import type { TrackPlanChangePreview } from "../../shared/api";
import { TrackPlanChangePreviewPanel } from "./TrackPlanChangePreviewPanel";

const preview: TrackPlanChangePreview = {
  revisionId: "revision-2",
  baseRevisionId: "revision-1",
  objectChanges: [
    { type: "added", lineageId: "lineage-2" },
    { type: "changed", lineageId: "lineage-1" }
  ],
  materialDeltas: [{
    geometryId: "g1", libraryId: "tillig-v1", articleNumber: "83101", name: "Gleisstück G1",
    baseQuantity: 1, currentQuantity: 2, delta: 1
  }],
  issues: {
    added: [{ code: "open_end", severity: "warning", lineageIds: ["lineage-2"], portIds: ["b"] }],
    resolved: [{ code: "overlap", severity: "warning", lineageIds: ["lineage-1"] }]
  },
  affectedConfigurations: [{ id: "configuration-1", name: "Ausstellung" }]
};

describe("TrackPlanChangePreviewPanel", () => {
  it("shows compact revision, material, issue, and configuration deltas", () => {
    render(<TrackPlanChangePreviewPanel preview={preview} />);

    expect(screen.getByText("2 Objektänderungen")).toBeInTheDocument();
    expect(screen.getByText("+1 Tillig 83101")).toBeInTheDocument();
    expect(screen.getByText("1 neuer Prüfhinweis")).toBeInTheDocument();
    expect(screen.getByText("1 gelöst")).toBeInTheDocument();
    expect(screen.getByText("Ausstellung")).toBeInTheDocument();
  });
});
