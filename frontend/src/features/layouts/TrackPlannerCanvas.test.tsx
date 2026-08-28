import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  ApiError,
  api,
  type LayoutUnit,
  type PlanFreeObject,
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
  manufacturer: "Tillig",
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
  id: "track-1", lineageId: "track-1", revisionId: draft.id, geometryId: geometry.id, geometry,
  positionXMm: 517, positionYMm: 250, rotationDegrees: 0,
  elevationStartMm: 5, elevationEndMm: 9.15, version: 1,
  effectiveGeometry: geometry.geometry, effectiveLengthMm: 166,
  createdAt: "2026-08-09T10:00:00Z", updatedAt: "2026-08-09T10:00:00Z"
};
const freeObject: PlanFreeObject = {
  id: "free-1", lineageId: "free-1", revisionId: draft.id, name: "Bahnsteig",
  category: "platform", positionXMm: 600, positionYMm: 250, rotationDegrees: 0,
  shape: { schemaVersion: 1, kind: "rectangle", widthMm: 400, heightMm: 65 },
  version: 1, createdAt: "2026-08-10T00:00:00Z", updatedAt: "2026-08-10T00:00:00Z"
};

const flexGeometry: TrackGeometryDefinition = {
  ...geometry,
  id: "tillig-tt-modellgleis-83125-v1",
  articleNumber: "83125",
  name: "Flexgleis",
  kind: "flex",
  lengthMm: 664,
  minimumRadiusMm: 543,
  geometry: {
    schemaVersion: 1,
    ports: [
      { id: "a", xMm: 0, yMm: 0, directionDegrees: 180 },
      { id: "b", xMm: 664, yMm: 0, directionDegrees: 0 }
    ],
    routes: [{ id: "main", points: [{ xMm: 0, yMm: 0 }, { xMm: 664, yMm: 0 }] }]
  }
};
const flexObject: PlanTrackObject = {
  ...trackObject,
  id: "flex-1",
  geometryId: flexGeometry.id,
  geometry: flexGeometry,
  flexPath: {
    schemaVersion: 1,
    endXMm: 500,
    endYMm: 100,
    endDirectionDegrees: 20,
    startHandleMm: 180,
    endHandleMm: 170
  },
  effectiveGeometry: {
    schemaVersion: 1,
    ports: [
      { id: "a", xMm: 0, yMm: 0, directionDegrees: 180 },
      { id: "b", xMm: 500, yMm: 100, directionDegrees: 20 }
    ],
    routes: [{ id: "main", points: [{ xMm: 0, yMm: 0 }, { xMm: 250, yMm: 30 }, { xMm: 500, yMm: 100 }] }]
  },
  effectiveLengthMm: 515,
  effectiveMinimumRadiusMm: 700
};

describe("TrackPlannerCanvas", () => {
  afterEach(() => vi.restoreAllMocks());

  beforeEach(() => {
    vi.spyOn(api, "trackGeometries").mockResolvedValue([geometry, { ...geometry, id: "draft-geometry", status: "draft" }]);
    vi.spyOn(api, "trackPlan").mockResolvedValue({
      revisionId: draft.id, status: "draft", objects: [], freeObjects: []
    });
    vi.spyOn(api, "trackPlanAnalysis").mockResolvedValue({
      revisionId: draft.id, status: "draft", connections: [], issues: [], bom: [], grades: [],
      materials: [], reservations: []
    });
    vi.spyOn(api, "trackPlanChangePreview").mockResolvedValue({
      revisionId: draft.id, baseRevisionId: "", objectChanges: [], freeObjectChanges: [], materialDeltas: [],
      issues: { added: [], resolved: [] }, affectedConfigurations: []
    });
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
      geometryId: geometry.id, positionXMm: 517, positionYMm: 250, rotationDegrees: 0,
      elevationStartMm: 0, elevationEndMm: 0
    });
    const placed = await screen.findByRole("button", { name: "Gleis Tillig 83101 · Gleisstück G1" });
    expect(placed).toHaveAttribute("transform", "translate(517 250) rotate(0)");
    expect(placed.querySelector("polyline")).toHaveAttribute("points", "0,0 166,0");
    expect(placed.querySelectorAll("circle")).toHaveLength(2);
  });

  it("uses the imported library manufacturer in the planner palette", async () => {
    vi.mocked(api.trackGeometries).mockResolvedValue([{
      ...geometry,
      id: "kuehn-72620",
      manufacturer: "Kühn",
      articleNumber: "72620",
      name: "Gerades Gleis"
    }]);

    render(<TrackPlannerCanvas unit={unit} gauge="TT" revision={draft} canPlan onClose={vi.fn()} />);

    expect(await screen.findByRole("button", { name: "Kühn 72620 · Gerades Gleis" }))
      .toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Tillig 72620 · Gerades Gleis" }))
      .not.toBeInTheDocument();
  });

  it("creates, drags, rotates, edits and deletes free plan objects", async () => {
    const user = userEvent.setup();
    const create = vi.spyOn(api, "createFreePlanObject").mockResolvedValue(freeObject);
    const update = vi.spyOn(api, "updateFreePlanObject").mockImplementation(async (_id, input) => ({
      ...freeObject, ...input, version: input.expectedVersion + 1
    }));
    const remove = vi.spyOn(api, "deleteFreePlanObject").mockResolvedValue(undefined);
    render(<TrackPlannerCanvas unit={unit} gauge="TT" revision={draft} canPlan onClose={vi.fn()} />);

    const canvas = await screen.findByRole("img", { name: "Maßhaltiger Gleisplan für Bahnhofsmodul" });
    await user.click(screen.getByRole("button", { name: "Planobjekt hinzufügen" }));
    await user.type(screen.getByRole("textbox", { name: "Name" }), "Bahnsteig");
    await user.click(screen.getByRole("button", { name: "Planobjekt speichern" }));
    await waitFor(() => expect(create).toHaveBeenCalledWith(draft.id, expect.objectContaining({
      name: "Bahnsteig", positionXMm: 600, positionYMm: 250
    })));

    vi.spyOn(canvas, "getBoundingClientRect").mockReturnValue({
      left: 0, top: 0, width: 1200, height: 500, right: 1200, bottom: 500, x: 0, y: 0,
      toJSON: () => ({})
    });
    const placed = await screen.findByRole("button", { name: "Bahnsteig" });
    const capturePointer = vi.fn();
    Object.defineProperty(placed, "setPointerCapture", { value: capturePointer });
    fireEvent.pointerDown(placed, { pointerId: 7, clientX: 600, clientY: 250 });
    expect(capturePointer).toHaveBeenCalledWith(7);
    fireEvent.pointerMove(canvas, { pointerId: 7, clientX: 700, clientY: 300 });
    fireEvent.pointerUp(canvas, { pointerId: 7, clientX: 700, clientY: 300 });
    await waitFor(() => expect(update).toHaveBeenCalledTimes(1));
    expect(update).toHaveBeenNthCalledWith(1, freeObject.id, expect.objectContaining({
      positionXMm: 700, positionYMm: 300, expectedVersion: 1
    }));

    await user.click(screen.getByRole("button", { name: "+15°" }));
    await waitFor(() => expect(update).toHaveBeenCalledTimes(2));
    expect(update).toHaveBeenNthCalledWith(2, freeObject.id, expect.objectContaining({
      rotationDegrees: 15, expectedVersion: 2
    }));
    await user.click(screen.getByRole("button", { name: "Planobjekt bearbeiten" }));
    const name = screen.getByRole("textbox", { name: "Name" });
    await user.clear(name);
    await user.type(name, "Bahnsteig 1");
    await user.click(screen.getByRole("button", { name: "Planobjekt speichern" }));
    await waitFor(() => expect(update).toHaveBeenCalledTimes(3));
    expect(update).toHaveBeenNthCalledWith(3, freeObject.id, expect.objectContaining({
      name: "Bahnsteig 1", expectedVersion: 3
    }));

    await user.click(screen.getByRole("button", { name: "Planobjekt löschen" }));
    const dialog = screen.getByRole("dialog", { name: "Planobjekt löschen?" });
    await user.click(within(dialog).getByRole("button", { name: "Löschen" }));
    await waitFor(() => expect(remove).toHaveBeenCalledWith(freeObject.id, 4));
  });

  it("keeps track and free-object selection exclusive and reloads free-object conflicts", async () => {
    const user = userEvent.setup();
    vi.mocked(api.trackPlan).mockResolvedValue({
      revisionId: draft.id, status: "draft", objects: [trackObject], freeObjects: [freeObject]
    });
    vi.spyOn(api, "updateFreePlanObject").mockRejectedValue(
      new ApiError("changed", "track_plan_conflict", 409)
    );
    render(<TrackPlannerCanvas unit={unit} gauge="TT" revision={draft} canPlan onClose={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "Bahnsteig" }));
    expect(screen.getByRole("heading", { name: "Bahnsteig" })).toBeInTheDocument();
    const track = screen.getByRole("button", { name: "Gleis Tillig 83101 · Gleisstück G1" });
    await user.click(track);
    expect(screen.queryByRole("heading", { name: "Bahnsteig" })).not.toBeInTheDocument();
    expect(track).toHaveClass("selected");
    expect(screen.getByRole("button", { name: "Bahnsteig" })).toHaveAttribute("aria-pressed", "false");
    await user.click(screen.getByRole("button", { name: "Bahnsteig" }));
    await user.click(screen.getByRole("button", { name: "+15°" }));
    expect(await screen.findByText("Der Gleisplan wurde zwischenzeitlich geändert.")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Serverstand neu laden" }));
    await waitFor(() => expect(api.trackPlan).toHaveBeenCalledTimes(2));
  });

  it("moves on pointer release, rotates by 15 degrees and confirms deletion", async () => {
    const user = userEvent.setup();
    vi.mocked(api.trackPlan).mockResolvedValue({
      revisionId: draft.id, status: "draft", objects: [trackObject], freeObjects: []
    });
    const update = vi.spyOn(api, "updatePlanTrackObject").mockImplementation(async (_id, input) => ({
      ...trackObject, ...input,
      flexPath: input.flexPath ?? undefined,
      transitionPath: input.transitionPath ?? undefined,
      version: 2
    }));
    const remove = vi.spyOn(api, "deletePlanTrackObject").mockResolvedValue(undefined);
    render(<TrackPlannerCanvas unit={unit} gauge="TT" revision={draft} canPlan onClose={vi.fn()} />);

    const canvas = await screen.findByRole("img", { name: "Maßhaltiger Gleisplan für Bahnhofsmodul" });
    vi.spyOn(canvas, "getBoundingClientRect").mockReturnValue({
      left: 0, top: 0, width: 1200, height: 500, right: 1200, bottom: 500, x: 0, y: 0,
      toJSON: () => ({})
    });
    const placed = screen.getByRole("button", { name: "Gleis Tillig 83101 · Gleisstück G1" });
    const capturePointer = vi.fn();
    Object.defineProperty(placed, "setPointerCapture", { value: capturePointer });
    fireEvent.pointerDown(placed, { pointerId: 4, clientX: 517, clientY: 250 });
    expect(capturePointer).toHaveBeenCalledWith(4);
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

  it("previews a compatible endpoint snap and submits the snapped pose", async () => {
    const target = { ...trackObject, id: "target", positionXMm: 0, positionYMm: 0 };
    const moving = { ...trackObject, id: "moving", positionXMm: 200, positionYMm: 0 };
    vi.mocked(api.trackPlan).mockResolvedValue({
      revisionId: draft.id, status: "draft", objects: [target, moving], freeObjects: []
    });
    const update = vi.spyOn(api, "updatePlanTrackObject").mockResolvedValue({
      ...moving, positionXMm: 165.5, version: 2
    });
    render(<TrackPlannerCanvas unit={unit} gauge="TT" revision={draft} canPlan onClose={vi.fn()} />);

    const canvas = await screen.findByRole("img", { name: "Maßhaltiger Gleisplan für Bahnhofsmodul" });
    vi.spyOn(canvas, "getBoundingClientRect").mockReturnValue({
      left: 0, top: 0, width: 1200, height: 500, right: 1200, bottom: 500, x: 0, y: 0,
      toJSON: () => ({})
    });
    const placed = screen.getAllByRole("button", { name: "Gleis Tillig 83101 · Gleisstück G1" })[1];
    fireEvent.pointerDown(placed, { pointerId: 5, clientX: 200, clientY: 0 });
    fireEvent.pointerMove(canvas, { pointerId: 5, clientX: 172, clientY: 2 });
    expect(placed).toHaveAttribute("transform", "translate(166 0) rotate(0)");
    fireEvent.pointerUp(canvas, { pointerId: 5, clientX: 172, clientY: 2 });

    await waitFor(() => expect(update).toHaveBeenCalledWith("moving", {
      positionXMm: 166, positionYMm: 0, rotationDegrees: 0,
      elevationStartMm: 5, elevationEndMm: 9.15, expectedVersion: 1
    }));
    await waitFor(() => expect(placed).toHaveAttribute("transform", "translate(165.5 0) rotate(0)"));
  });

  it("renders explicit connected, open, and incompatible endpoint marker classes", async () => {
    const target = { ...trackObject, id: "target", positionXMm: 0 };
    const moving = { ...trackObject, id: "moving", positionXMm: 166 };
    const conflicting = { ...trackObject, id: "conflicting", positionXMm: 500 };
    vi.mocked(api.trackPlan).mockResolvedValue({
      revisionId: draft.id, status: "draft", objects: [target, moving, conflicting], freeObjects: []
    });
    vi.mocked(api.trackPlanAnalysis).mockResolvedValue({
      revisionId: draft.id,
      status: "draft",
      connections: [{ objectAId: "target", portAId: "b", objectBId: "moving", portBId: "a" }],
      issues: [
        { code: "open_end", severity: "warning", objectIds: ["target"], portIds: ["a"] },
        { code: "incompatible_connection", severity: "warning",
          objectIds: ["moving", "conflicting"], portIds: ["b", "a"] },
        { code: "open_end", severity: "warning", objectIds: ["conflicting"], portIds: ["b"] }
      ],
      bom: [],
      grades: [],
      materials: [],
      reservations: []
    });
    render(<TrackPlannerCanvas unit={unit} gauge="TT" revision={draft} canPlan onClose={vi.fn()} />);

    const canvas = await screen.findByRole("img", { name: "Maßhaltiger Gleisplan für Bahnhofsmodul" });
    expect(canvas.querySelectorAll(".track-port.status-connected")).toHaveLength(2);
    expect(canvas.querySelectorAll(".track-port.status-open")).toHaveLength(2);
    expect(canvas.querySelectorAll(".track-port.status-incompatible")).toHaveLength(2);
    expect(canvas.querySelectorAll(".track-port text")).toHaveLength(6);
  });

  it("edits an elevation profile with app inputs and displays the derived grade", async () => {
    const user = userEvent.setup();
    vi.mocked(api.trackPlan).mockResolvedValue({
      revisionId: draft.id, status: "draft", objects: [trackObject], freeObjects: []
    });
    vi.mocked(api.trackPlanAnalysis).mockResolvedValue({
      revisionId: draft.id, status: "draft", connections: [], issues: [], bom: [],
      grades: [{
        objectId: trackObject.id, elevationStartMm: 5, elevationEndMm: 9.15,
        lengthMm: 166, gradePercent: 2.5
      }],
      materials: [], reservations: []
    });
    const update = vi.spyOn(api, "updatePlanTrackObject").mockImplementation(async (_id, input) => ({
      ...trackObject, ...input,
      flexPath: input.flexPath ?? undefined,
      transitionPath: input.transitionPath ?? undefined,
      version: 2
    }));
    render(<TrackPlannerCanvas unit={unit} gauge="TT" revision={draft} canPlan onClose={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "Gleis Tillig 83101 · Gleisstück G1" }));
    expect(screen.getByText("2,50 %")).toBeInTheDocument();
    const start = screen.getByRole("spinbutton", { name: "Anfangshöhe (mm)" });
    const end = screen.getByRole("spinbutton", { name: "Endhöhe (mm)" });
    expect(start).toHaveValue(5);
    expect(end).toHaveValue(9.15);

    await user.clear(start);
    await user.type(start, "7");
    await user.clear(end);
    await user.type(end, "11.98");
    await user.click(screen.getByRole("button", { name: "Höhenprofil speichern" }));
    await waitFor(() => expect(update).toHaveBeenCalledWith(trackObject.id, {
      positionXMm: 517, positionYMm: 250, rotationDegrees: 0,
      elevationStartMm: 7, elevationEndMm: 11.98, expectedVersion: 1
    }));
  });

  it("renders effective flex geometry and saves an explicitly accepted preview", async () => {
    const user = userEvent.setup();
    vi.mocked(api.trackGeometries).mockResolvedValue([geometry, flexGeometry]);
    vi.mocked(api.trackPlan).mockResolvedValue({
      revisionId: draft.id, status: "draft", objects: [flexObject], freeObjects: []
    });
    vi.spyOn(api, "previewFlexTrackPath").mockResolvedValue({
      path: { ...flexObject.flexPath!, endXMm: 520 },
      effectiveGeometry: flexObject.effectiveGeometry!,
      effectiveLengthMm: 530,
      effectiveMinimumRadiusMm: 680,
      radiusLimitMm: 700,
      lengthExceeded: false,
      radiusBelowLimit: true,
      applicable: true
    });
    const update = vi.spyOn(api, "updatePlanTrackObject").mockImplementation(async (_id, input) => ({
      ...flexObject, ...input,
      flexPath: input.flexPath ?? undefined,
      transitionPath: input.transitionPath ?? undefined,
      version: 2
    }));
    render(<TrackPlannerCanvas unit={unit} gauge="TT" revision={draft} canPlan onClose={vi.fn()} />);

    const placed = await screen.findByRole("button", { name: "Gleis Tillig 83125 · Flexgleis" });
    expect(placed.querySelector("polyline")).toHaveAttribute("points", "0,0 250,30 500,100");
    await user.click(placed);
    expect(screen.getByText("515,00 mm")).toBeInTheDocument();
    expect(screen.getByText("700,00 mm")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Flexverlauf bearbeiten" }));
    await user.clear(screen.getByRole("spinbutton", { name: "Endpunkt X (mm)" }));
    await user.type(screen.getByRole("spinbutton", { name: "Endpunkt X (mm)" }), "520");
    await user.click(screen.getByRole("button", { name: "Verlauf vorschlagen" }));
    await user.click(await screen.findByRole("button", { name: "Verlauf übernehmen" }));

    await waitFor(() => expect(update).toHaveBeenCalledWith(flexObject.id, expect.objectContaining({
      flexPath: expect.objectContaining({ endXMm: 520 }), transitionPath: null, expectedVersion: 1
    })));
  });

  it("converts a free flex path to a transition curve and preserves it on pose updates", async () => {
    const user = userEvent.setup();
    vi.mocked(api.trackGeometries).mockResolvedValue([geometry, flexGeometry]);
    vi.mocked(api.trackPlan).mockResolvedValue({
      revisionId: draft.id, status: "draft", objects: [flexObject], freeObjects: []
    });
    vi.spyOn(api, "previewTransitionCurvePath").mockResolvedValue({
      path: { schemaVersion: 1, lengthMm: 500, endRadiusMm: 700, direction: "left" },
      effectiveGeometry: flexObject.effectiveGeometry,
      effectiveLengthMm: 500,
      effectiveMinimumRadiusMm: 700,
      radiusLimitMm: 700,
      lengthExceeded: false,
      radiusBelowLimit: false,
      applicable: true
    });
    const transitionObject: PlanTrackObject = {
      ...flexObject,
      flexPath: undefined,
      transitionPath: { schemaVersion: 1, lengthMm: 500, endRadiusMm: 700, direction: "left" },
      effectiveLengthMm: 500,
      version: 2
    };
    const update = vi.spyOn(api, "updatePlanTrackObject")
      .mockResolvedValueOnce(transitionObject)
      .mockResolvedValueOnce({ ...transitionObject, rotationDegrees: 15, version: 3 });
    render(<TrackPlannerCanvas unit={unit} gauge="TT" revision={draft} canPlan onClose={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "Gleis Tillig 83125 · Flexgleis" }));
    await user.click(screen.getByRole("button", { name: "Übergangsbogen" }));
    await user.clear(screen.getByRole("spinbutton", { name: "Länge (mm)" }));
    await user.type(screen.getByRole("spinbutton", { name: "Länge (mm)" }), "500");
    await user.click(screen.getByRole("button", { name: "Übergangsbogen vorschlagen" }));
    await user.click(await screen.findByRole("button", { name: "Übergangsbogen übernehmen" }));
    await waitFor(() => expect(update).toHaveBeenNthCalledWith(1, flexObject.id, expect.objectContaining({
      flexPath: null,
      transitionPath: { schemaVersion: 1, lengthMm: 500, endRadiusMm: 700, direction: "left" },
      expectedVersion: 1
    })));

    await user.click(screen.getByRole("button", { name: "+15°" }));
    await waitFor(() => expect(update).toHaveBeenNthCalledWith(2, flexObject.id, expect.objectContaining({
      flexPath: null,
      transitionPath: transitionObject.transitionPath,
      rotationDegrees: 15,
      expectedVersion: 2
    })));
  });

  it("keeps published plans read-only and reports version conflicts", async () => {
    const user = userEvent.setup();
    const published = { ...draft, status: "published" as const };
    vi.mocked(api.trackPlan).mockResolvedValue({
      revisionId: published.id, status: "published", objects: [trackObject], freeObjects: []
    });
    const { rerender } = render(
      <TrackPlannerCanvas unit={unit} gauge="TT" revision={published} canPlan onClose={vi.fn()} />
    );
    await screen.findByRole("img", { name: "Maßhaltiger Gleisplan für Bahnhofsmodul" });
    await user.click(screen.getByRole("button", { name: "Gleis Tillig 83101 · Gleisstück G1" }));
    expect(screen.getByText("5,00 mm")).toBeInTheDocument();
    expect(screen.getByText("9,15 mm")).toBeInTheDocument();
    expect(screen.queryByRole("spinbutton", { name: "Anfangshöhe (mm)" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "+15°" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Tillig 83101 · Gleisstück G1" })).not.toBeInTheDocument();

    vi.mocked(api.trackPlan).mockResolvedValue({
      revisionId: draft.id, status: "draft", objects: [trackObject], freeObjects: []
    });
    vi.spyOn(api, "updatePlanTrackObject").mockRejectedValue(new ApiError("changed", "track_plan_conflict", 409));
    rerender(<TrackPlannerCanvas unit={unit} gauge="TT" revision={draft} canPlan onClose={vi.fn()} />);
    await user.click(await screen.findByRole("button", { name: "Gleis Tillig 83101 · Gleisstück G1" }));
    await user.click(await screen.findByRole("button", { name: "+15°" }));
    expect(await screen.findByText("Der Gleisplan wurde zwischenzeitlich geändert.")).toBeInTheDocument();
  });
});
