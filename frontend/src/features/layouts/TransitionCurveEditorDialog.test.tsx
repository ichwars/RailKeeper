import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api, type PlanTrackObject, type TransitionCurvePreview } from "../../shared/api";
import { TransitionCurveEditorDialog } from "./TransitionCurveEditorDialog";

const object = {
  id: "flex-1",
  lineageId: "flex-1",
  revisionId: "revision-1",
  geometryId: "tillig-tt-modellgleis-83125-v1",
  geometry: {
    id: "tillig-tt-modellgleis-83125-v1",
    libraryId: "tillig-tt-modellgleis-v1",
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
    endYMm: 100,
    endDirectionDegrees: 20,
    startHandleMm: 166,
    endHandleMm: 166
  },
  effectiveGeometry: {
    schemaVersion: 1,
    ports: [],
    routes: [{ id: "main", points: [{ xMm: 0, yMm: 0 }, { xMm: 500, yMm: 100 }] }]
  },
  effectiveLengthMm: 515,
  effectiveMinimumRadiusMm: 700,
  positionXMm: 0,
  positionYMm: 0,
  rotationDegrees: 0,
  elevationStartMm: 0,
  elevationEndMm: 0,
  version: 3,
  createdAt: "2026-08-10T00:00:00Z",
  updatedAt: "2026-08-10T00:00:00Z"
} satisfies PlanTrackObject;

const preview: TransitionCurvePreview = {
  path: { schemaVersion: 1, lengthMm: 500, endRadiusMm: 700, direction: "left" },
  effectiveGeometry: {
    schemaVersion: 1,
    ports: [],
    routes: [{ id: "main", points: [{ xMm: 0, yMm: 0 }, { xMm: 499, yMm: 30 }] }]
  },
  effectiveLengthMm: 500,
  effectiveMinimumRadiusMm: 700,
  radiusLimitMm: 700,
  lengthExceeded: false,
  radiusBelowLimit: false,
  applicable: true
};

describe("TransitionCurveEditorDialog", () => {
  afterEach(() => vi.restoreAllMocks());

  it("previews server geometry and applies it only after explicit confirmation", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "previewTransitionCurvePath").mockResolvedValue(preview);
    const onApply = vi.fn();
    render(<TransitionCurveEditorDialog object={object} saving={false}
      onApply={onApply} onClose={vi.fn()} />);

    expect(screen.getByRole("button", { name: "Schließen" })).toHaveFocus();
    expect(screen.getByRole("spinbutton", { name: "Länge (mm)" })).toHaveValue(515);
    expect(screen.getByRole("spinbutton", { name: "Endradius (mm)" })).toHaveValue(700);
    expect(screen.getByRole("button", { name: "Richtung" })).toHaveTextContent("Links");
    await user.clear(screen.getByRole("spinbutton", { name: "Länge (mm)" }));
    await user.type(screen.getByRole("spinbutton", { name: "Länge (mm)" }), "500");
    await user.click(screen.getByRole("button", { name: "Übergangsbogen vorschlagen" }));

    await waitFor(() => expect(api.previewTransitionCurvePath).toHaveBeenCalledWith(object.id, {
      lengthMm: 500, endRadiusMm: 700, direction: "left", expectedVersion: 3
    }));
    expect(screen.getAllByText("500,00 mm")).toHaveLength(1);
    expect(screen.getAllByText("700,00 mm")).toHaveLength(2);
    expect(onApply).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Übergangsbogen übernehmen" }));
    expect(onApply).toHaveBeenCalledWith(preview.path);
  });

  it("shows warnings, blocks overlength and cancels without writing", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "previewTransitionCurvePath").mockResolvedValue({
      ...preview,
      effectiveLengthMm: 665,
      effectiveMinimumRadiusMm: 600,
      lengthExceeded: true,
      radiusBelowLimit: true,
      applicable: false
    });
    const onApply = vi.fn();
    const onClose = vi.fn();
    render(<TransitionCurveEditorDialog object={object} saving={false}
      onApply={onApply} onClose={onClose} />);

    await user.click(screen.getByRole("button", { name: "Übergangsbogen vorschlagen" }));
    expect(await screen.findByText(/überschreitet die verfügbare Länge/)).toBeInTheDocument();
    expect(screen.getByText(/unterschreitet den Grenzwert/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Übergangsbogen übernehmen" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "Abbrechen" }));
    expect(onClose).toHaveBeenCalledOnce();
    expect(onApply).not.toHaveBeenCalled();
  });
});
