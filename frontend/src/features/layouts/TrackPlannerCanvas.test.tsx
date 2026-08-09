import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  ApiError,
  api,
  type LayoutUnit,
  type PlanRevision,
  type PlanTrackObject,
  type TrackGeometryDefinition
} from "../../shared/api";
import { TrackPlannerCanvas } from "./TrackPlannerCanvas";

const unit: LayoutUnit = {
  id: "unit-1", layoutId: "layout-1", name: "Bahnhofsmodul", kind: "module",
  widthMm: 1200, heightMm: 500, version: 1, archived: false,
  createdAt: "2026-08-09T10:00:00Z", updatedAt: "2026-08-09T10:00:00Z"
};
const draft: PlanRevision = {
  id: "revision-1", variantId: "variant-1", revisionNumber: 1, status: "draft", version: 1,
  createdBy: "planner", createdAt: "2026-08-09T10:00:00Z", updatedAt: "2026-08-09T10:00:00Z"
};
const geometry: TrackGeometryDefinition = {
  id: "tillig-tt-modellgleis-83101-v1", libraryId: "tillig-tt-modellgleis-v1",
  articleNumber: "83101", name: "Gleisstück G1", kind: "straight", lengthMm: 166,
  sourceUrl: "https://www.tillig.com/Produkte/produktinfo-83101.html", status: "verified",
  createdAt: "2026-08-09T10:00:00Z",
  geometry: {
    schemaVersion: 1,
    ports: [
      { id: "a", xMm: 0, yMm: 0, directionDegrees: 180 },
      { id: "b", xMm: 166, yMm: 0, directionDegrees: 0 }
    ],
    routes: [{ id: "main", points: [{ xMm: 0, yMm: 0 }, { xMm: 166, yMm: 0 }] }]
  }
};
const trackObject: PlanTrackObject = {
  id: "track-1", revisionId: draft.id, geometryId: geometry.id, geometry,
  positionXMm: 517, positionYMm: 250, rotationDegrees: 0, version: 1,
  createdAt: "2026-08-09T10:00:00Z", updatedAt: "2026-08-09T10:00:00Z"
};

describe("TrackPlannerCanvas", () => {
  afterEach(() => vi.restoreAllMocks());

  beforeEach(() => {
    vi.spyOn(api, "trackGeometries").mockResolvedValue([geometry, { ...geometry, id: "draft-geometry", status: "draft" }]);
    vi.spyOn(api, "trackPlan").mockResolvedValue({ revisionId: draft.id, status: "draft", objects: [] });
  });

  it("places only verified TT geometry at the unit centre and renders exact SVG geometry", async () => {
    const user = userEvent.setup();
    const create = vi.spyOn(api, "createPlanTrackObject").mockResolvedValue(trackObject);
    render(<TrackPlannerCanvas unit={unit} gauge="TT" revision={draft} canPlan onClose={vi.fn()} />);

    const canvas = await screen.findByRole("img", { name: "Maßhaltiger Gleisplan für Bahnhofsmodul" });
    expect(canvas).toHaveAttribute("viewBox", "0 0 1200 500");
    expect(screen.getByRole("button", { name: "Tillig 83101 · Gleisstück G1" })).toBeInTheDocument();
    expect(screen.queryByText("draft-geometry")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Tillig 83101 · Gleisstück G1" }));
    expect(create).toHaveBeenCalledWith(draft.id, {
      geometryId: geometry.id, positionXMm: 517, positionYMm: 250, rotationDegrees: 0
    });
    const placed = await screen.findByRole("button", { name: "Gleis Tillig 83101 G1" });
    expect(placed).toHaveAttribute("transform", "translate(517 250) rotate(0)");
    expect(placed.querySelector("polyline")).toHaveAttribute("points", "0,0 166,0");
    expect(placed.querySelectorAll("circle")).toHaveLength(2);
  });

  it("moves on pointer release, rotates by 15 degrees and confirms deletion", async () => {
    const user = userEvent.setup();
    vi.mocked(api.trackPlan).mockResolvedValue({ revisionId: draft.id, status: "draft", objects: [trackObject] });
    const update = vi.spyOn(api, "updatePlanTrackObject").mockImplementation(async (_id, input) => ({
      ...trackObject, ...input, version: 2
    }));
    const remove = vi.spyOn(api, "deletePlanTrackObject").mockResolvedValue(undefined);
    render(<TrackPlannerCanvas unit={unit} gauge="TT" revision={draft} canPlan onClose={vi.fn()} />);

    const canvas = await screen.findByRole("img", { name: "Maßhaltiger Gleisplan für Bahnhofsmodul" });
    vi.spyOn(canvas, "getBoundingClientRect").mockReturnValue({
      left: 0, top: 0, width: 1200, height: 500, right: 1200, bottom: 500, x: 0, y: 0,
      toJSON: () => ({})
    });
    const placed = screen.getByRole("button", { name: "Gleis Tillig 83101 G1" });
    fireEvent.pointerDown(placed, { pointerId: 4, clientX: 517, clientY: 250 });
    fireEvent.pointerMove(canvas, { pointerId: 4, clientX: 540, clientY: 270 });
    fireEvent.pointerUp(canvas, { pointerId: 4, clientX: 540, clientY: 270 });
    await waitFor(() => expect(update).toHaveBeenCalledWith(trackObject.id,
      expect.objectContaining({ positionXMm: 540, positionYMm: 270, expectedVersion: 1 })));

    await user.click(screen.getByRole("button", { name: "+15°" }));
    await waitFor(() => expect(update).toHaveBeenLastCalledWith(trackObject.id,
      expect.objectContaining({ rotationDegrees: 15 })));
    await user.click(screen.getByRole("button", { name: "Gleis löschen" }));
    const dialog = screen.getByRole("dialog", { name: "Gleis löschen?" });
    expect(within(dialog).getByText(/Tillig 83101/)).toBeInTheDocument();
    await user.click(within(dialog).getByRole("button", { name: "Löschen" }));
    await waitFor(() => expect(remove).toHaveBeenCalledWith(trackObject.id, 2));
  });

  it("keeps published plans read-only and reports version conflicts", async () => {
    const user = userEvent.setup();
    const published = { ...draft, status: "published" as const };
    vi.mocked(api.trackPlan).mockResolvedValue({ revisionId: published.id, status: "published", objects: [trackObject] });
    const { rerender } = render(
      <TrackPlannerCanvas unit={unit} gauge="TT" revision={published} canPlan onClose={vi.fn()} />
    );
    await screen.findByRole("img", { name: "Maßhaltiger Gleisplan für Bahnhofsmodul" });
    expect(screen.queryByRole("button", { name: "+15°" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Tillig 83101 · Gleisstück G1" })).not.toBeInTheDocument();

    vi.mocked(api.trackPlan).mockResolvedValue({ revisionId: draft.id, status: "draft", objects: [trackObject] });
    vi.spyOn(api, "updatePlanTrackObject").mockRejectedValue(new ApiError("changed", "track_plan_conflict", 409));
    rerender(<TrackPlannerCanvas unit={unit} gauge="TT" revision={draft} canPlan onClose={vi.fn()} />);
    await user.click(await screen.findByRole("button", { name: "Gleis Tillig 83101 G1" }));
    await user.click(await screen.findByRole("button", { name: "+15°" }));
    expect(await screen.findByText("Der Gleisplan wurde zwischenzeitlich geändert.")).toBeInTheDocument();
  });
});
