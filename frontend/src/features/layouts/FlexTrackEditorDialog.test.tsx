import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api, type FlexTrackPreview, type PlanTrackObject } from "../../shared/api";
import { FlexTrackEditorDialog } from "./FlexTrackEditorDialog";

const object = {
  id: "flex-1",
  lineageId: "flex-1",
  revisionId: "revision-1",
  geometryId: "tillig-tt-modellgleis-83125-v1",
  geometry: {
    id: "tillig-tt-modellgleis-83125-v1",
    libraryId: "tillig-tt-modellgleis-v1",
    manufacturer: "Tillig",
    articleNumber: "83125",
    name: "Flexgleis",
    kind: "flex",
    lengthMm: 664,
    minimumRadiusMm: 543,
    geometry: { schemaVersion: 1, ports: [], routes: [] },
    sourceUrl: "https://www.tillig.com/Produkte/produktinfo-83125.html",
    status: "verified",
    createdAt: "2026-08-10T00:00:00Z"
  },
  flexPath: {
    schemaVersion: 1,
    endXMm: 500,
    endYMm: 0,
    endDirectionDegrees: 0,
    startHandleMm: 166,
    endHandleMm: 166
  },
  effectiveGeometry: { schemaVersion: 1, ports: [], routes: [] },
  effectiveLengthMm: 500,
  effectiveMinimumRadiusMm: 900,
  positionXMm: 0,
  positionYMm: 0,
  rotationDegrees: 0,
  elevationStartMm: 0,
  elevationEndMm: 0,
  version: 3,
  createdAt: "2026-08-10T00:00:00Z",
  updatedAt: "2026-08-10T00:00:00Z"
} satisfies PlanTrackObject;

const preview: FlexTrackPreview = {
  path: { ...object.flexPath, endXMm: 620, startHandleMm: 210, endHandleMm: 210 },
  effectiveGeometry: {
    schemaVersion: 1,
    ports: [],
    routes: [{ id: "main", points: [{ xMm: 0, yMm: 0 }, { xMm: 620, yMm: 0 }] }]
  },
  effectiveLengthMm: 620,
  effectiveMinimumRadiusMm: 700,
  radiusLimitMm: 700,
  lengthExceeded: false,
  radiusBelowLimit: false,
  applicable: true
};

describe("FlexTrackEditorDialog", () => {
  afterEach(() => vi.restoreAllMocks());

  it("previews server geometry and applies it only after explicit confirmation", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "previewFlexTrackPath").mockResolvedValue(preview);
    const onApply = vi.fn();
    render(<FlexTrackEditorDialog object={object} saving={false} onApply={onApply} onClose={vi.fn()} />);

    expect(screen.getByRole("button", { name: "Endpunkt verschieben" })).toBeInTheDocument();
    const endX = screen.getByRole("spinbutton", { name: "Endpunkt X (mm)" });
    await user.clear(endX);
    await user.type(endX, "620");
    await user.click(screen.getByRole("button", { name: "Verlauf vorschlagen" }));

    await waitFor(() => expect(api.previewFlexTrackPath).toHaveBeenCalledWith(object.id, {
      endXMm: 620,
      endYMm: 0,
      endDirectionDegrees: 0,
      expectedVersion: 3
    }));
    expect(screen.getByText("620,00 mm")).toBeInTheDocument();
    expect(screen.getAllByText("700,00 mm")).toHaveLength(2);
    expect(onApply).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: "Verlauf übernehmen" }));
    expect(onApply).toHaveBeenCalledWith(preview.path);
  });

  it("shows radius and length warnings and blocks an overlong path", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "previewFlexTrackPath").mockResolvedValue({
      ...preview,
      effectiveLengthMm: 700,
      effectiveMinimumRadiusMm: 500,
      lengthExceeded: true,
      radiusBelowLimit: true,
      applicable: false
    });
    render(<FlexTrackEditorDialog object={object} saving={false} onApply={vi.fn()} onClose={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Verlauf vorschlagen" }));

    expect(await screen.findByText(/überschreitet die verfügbare Länge/)).toBeInTheDocument();
    expect(screen.getByText(/unterschreitet den Grenzwert/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Verlauf übernehmen" })).toBeDisabled();
  });
});
