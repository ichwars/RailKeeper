import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  ApiError,
  api,
  type Layout,
  type LayoutConfiguration,
  type LayoutTechnicalPosition,
  type LayoutUnit,
  type PlanRevision,
  type PlanVariant
} from "../../shared/api";
import { Shell } from "../../app/Shell";
import { LayoutsView } from "./LayoutsView";

const layout: Layout = {
  id: "layout-1", name: "Clubanlage mit langem Bahnhofsnamen", kind: "club", gauge: "TT", scale: "1:120",
  description: "Modulare Clubanlage", maxGradePercent: 3.5, minimumTrackClearanceMm: 40,
  minimumFlexRadiusMm: 700,
  version: 2, archived: false,
  createdAt: "2026-08-07T10:00:00Z",
  updatedAt: "2026-08-07T11:00:00Z"
};

const unit: LayoutUnit = {
  id: "unit-1", layoutId: layout.id, name: "Bahnhofsmodul", kind: "module", ownerLabel: "Daniel",
  widthMm: 1200, heightMm: 500, version: 1, archived: false, createdAt: "2026-08-07T10:00:00Z",
  updatedAt: "2026-08-07T10:00:00Z"
};

const technicalPosition: LayoutTechnicalPosition = {
  id: "position-1", layoutUnitId: unit.id, label: "Einfahrsignal A", kind: "signal",
  positionXMm: 250, positionYMm: 80, rotationDegrees: 90, description: "Gleis 1",
  version: 1, archived: false, createdAt: "2026-08-09T10:00:00Z", updatedAt: "2026-08-09T10:00:00Z"
};

const reviewRevision: PlanRevision = {
  id: "revision-1", variantId: "variant-1", revisionNumber: 1, status: "review", version: 2,
  createdBy: "planner", createdAt: "2026-08-07T10:00:00Z", updatedAt: "2026-08-07T11:00:00Z"
};

const variant: PlanVariant = {
  id: "variant-1", layoutUnitId: unit.id, name: "Fahrbetrieb", archived: false, revisions: [reviewRevision],
  createdAt: "2026-08-07T10:00:00Z", updatedAt: "2026-08-07T11:00:00Z"
};

describe("LayoutsView", () => {
  afterEach(() => vi.restoreAllMocks());

  beforeEach(() => {
    vi.spyOn(api, "layouts").mockResolvedValue([layout]);
    vi.spyOn(api, "layout").mockResolvedValue(layout);
    vi.spyOn(api, "layoutUnits").mockResolvedValue([unit]);
    vi.spyOn(api, "layoutConfigurations").mockResolvedValue([]);
    vi.spyOn(api, "layoutConfigurationPortAnalysis").mockResolvedValue({ connections: [], issues: [] });
    vi.spyOn(api, "layoutTwin").mockResolvedValue({
      layoutId: layout.id, unitId: unit.id,
      bounds: { minXMm: 0, minYMm: 0, widthMm: unit.widthMm, heightMm: unit.heightMm },
      hasGeometry: true,
      units: [{ id: unit.id, name: unit.name, kind: unit.kind, positionXMm: 0, positionYMm: 0,
        rotationDegrees: 0, version: unit.version,
        localOutline: [{ xMm: 0, yMm: 0 }, { xMm: unit.widthMm, yMm: 0 },
          { xMm: unit.widthMm, yMm: unit.heightMm }, { xMm: 0, yMm: unit.heightMm }],
        outline: [{ xMm: 0, yMm: 0 }, { xMm: unit.widthMm, yMm: 0 },
          { xMm: unit.widthMm, yMm: unit.heightMm }, { xMm: 0, yMm: unit.heightMm }], positions: [] }],
      warnings: [{ code: "outline_fallback", unitId: unit.id }]
    });
    vi.spyOn(api, "planVariants").mockResolvedValue([variant]);
    vi.spyOn(api, "layoutTechnicalPositions").mockResolvedValue([technicalPosition]);
    vi.spyOn(api, "accessoryArticles").mockResolvedValue({
      items: [],
      metrics: { articleCount: 0, articleTypeCount: 0, available: 0, locationCount: 0,
        reserved: 0, installed: 0, careHintCount: 0 },
      filters: { manufacturers: [], articleTypes: [], gauges: [], storageLocations: [] }
    });
  });

  it("renders private and club layouts read-only for viewers", async () => {
    vi.mocked(api.layouts).mockResolvedValue([{ ...layout, id: "layout-private", kind: "private", name: "Heimanlage" }, layout]);
    render(<LayoutsView roles={["Viewer"]} />);

    expect((await screen.findAllByText("Heimanlage")).length).toBeGreaterThan(0);
    expect(screen.getByText(layout.name)).toBeInTheDocument();
    expect(screen.queryByText("Anlage anlegen")).not.toBeInTheDocument();
    expect(screen.queryByText("Anlage bearbeiten")).not.toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Module" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Aufbauten" })).toBeInTheDocument();
    const privateLayout = screen.getByRole("button", { name: /Heimanlage/ });
    expect(privateLayout).toHaveTextContent("Private Anlage");
    expect(privateLayout).toHaveTextContent("TT");
    expect(privateLayout).toHaveTextContent("1:120");
    expect(privateLayout).toHaveTextContent("Version 2");
    expect(privateLayout).toHaveTextContent(new Date(layout.updatedAt).toLocaleString());
  });

  it("creates a layout and a module for planners", async () => {
    const user = userEvent.setup();
    vi.mocked(api.layouts).mockResolvedValueOnce([]).mockResolvedValue([layout]);
    vi.spyOn(api, "createLayout").mockResolvedValue(layout);
    vi.spyOn(api, "createLayoutUnit").mockResolvedValue(unit);
    render(<LayoutsView roles={["Planner"]} />);

    await screen.findByText("Noch keine Anlage erfasst.");
    await user.click(screen.getByRole("button", { name: "Anlage anlegen" }));
    const dialog = screen.getByRole("dialog", { name: "Anlage anlegen" });
    await user.type(within(dialog).getByRole("textbox", { name: "Bezeichnung" }), "Clubanlage");
    await user.type(within(dialog).getByRole("spinbutton", { name: "Maximale Steigung (%)" }), "3.5");
    await user.type(within(dialog).getByRole("spinbutton", {
      name: "Mindestabstand kreuzender Gleise (mm)"
    }), "40");
    await user.type(within(dialog).getByRole("spinbutton", {
      name: "Mindest-Flexgleisradius (mm)"
    }), "700");
    await user.click(within(dialog).getByRole("button", { name: "Anlage speichern" }));
    await waitFor(() => expect(api.createLayout).toHaveBeenCalledWith(expect.objectContaining({
      name: "Clubanlage", kind: "private", gauge: "TT", scale: "1:120", maxGradePercent: 3.5,
      minimumTrackClearanceMm: 40, minimumFlexRadiusMm: 700
    })));

    await screen.findAllByText(layout.name);
    await user.click(screen.getByRole("tab", { name: "Module" }));
    const modulePanel = screen.getByText("Einheit anlegen").closest(".panel") as HTMLElement;
    await user.type(within(modulePanel).getByLabelText("Bezeichnung"), "Schattenbahnhof");
    await user.clear(within(modulePanel).getByLabelText("Breite (mm)"));
    await user.type(within(modulePanel).getByLabelText("Breite (mm)"), "900");
    await user.click(within(modulePanel).getByRole("button", { name: "Einheit speichern" }));
    await waitFor(() => expect(api.createLayoutUnit).toHaveBeenCalledWith(layout.id, expect.objectContaining({
      name: "Schattenbahnhof", kind: "module", widthMm: 900
    })));
  });

  it("creates structured setup configurations with positions", async () => {
    const user = userEvent.setup();
    const configuration: LayoutConfiguration = {
      id: "setup-1", layoutId: layout.id, name: "Ausstellung 2026", version: 1, archived: false,
      units: [{ unitId: unit.id, positionXMm: 0, positionYMm: 0, rotationDegrees: 0, sortOrder: 0 }],
      createdAt: "2026-08-07T10:00:00Z", updatedAt: "2026-08-07T10:00:00Z"
    };
    vi.spyOn(api, "createLayoutConfiguration").mockResolvedValue(configuration);
    render(<LayoutsView roles={["Planner"]} />);
    await screen.findAllByText(layout.name);

    await user.click(screen.getByRole("tab", { name: "Aufbauten" }));
    const setupPanel = (await screen.findByText("Aufbau anlegen")).closest(".panel") as HTMLElement;
    await user.type(within(setupPanel).getByLabelText("Bezeichnung"), "Ausstellung 2026");
    await user.click(within(setupPanel).getByRole("checkbox", { name: /Bahnhofsmodul/ }));
    await user.clear(within(setupPanel).getByLabelText("X (mm)"));
    await user.type(within(setupPanel).getByLabelText("X (mm)"), "1500");
    await user.click(within(setupPanel).getByRole("button", { name: "Aufbau speichern" }));

    await waitFor(() => expect(api.createLayoutConfiguration).toHaveBeenCalledWith(layout.id,
      expect.objectContaining({ name: "Ausstellung 2026", units: [expect.objectContaining({
        unitId: unit.id, positionXMm: 1500
      })] })));
  });

  it("previews module-port alignment without saving the configuration", async () => {
    const user = userEvent.setup();
    const targetUnit: LayoutUnit = {
      ...unit, id: "unit-2", name: "Streckenmodul", ownerLabel: "Verein"
    };
    const configuration: LayoutConfiguration = {
      id: "setup-1", layoutId: layout.id, name: "Ausstellung 2026", version: 2, archived: false,
      units: [
        { unitId: unit.id, positionXMm: 5, positionYMm: 10, rotationDegrees: 2, sortOrder: 0 },
        { unitId: targetUnit.id, positionXMm: 112, positionYMm: 0, rotationDegrees: 0, sortOrder: 1 }
      ],
      createdAt: "2026-08-10T10:00:00Z", updatedAt: "2026-08-10T10:00:00Z"
    };
    vi.mocked(api.layoutUnits).mockResolvedValue([unit, targetUnit]);
    vi.mocked(api.layoutConfigurations).mockResolvedValue([configuration]);
    vi.spyOn(api, "previewLayoutConfigurationUnitSnap").mockResolvedValue({
      snapped: true,
      pose: { positionXMm: 12, positionYMm: 20, rotationDegrees: 0 },
      movingPortId: "port-a", targetUnitId: "unit-b", targetPortId: "port-b", distanceMm: 7
    });
    const update = vi.spyOn(api, "updateLayoutConfiguration");
    render(<LayoutsView roles={["Planner"]} />);
    await screen.findAllByText(layout.name);

    await user.click(screen.getByRole("tab", { name: "Aufbauten" }));
    await user.click(await screen.findByRole("button", { name: /Ausstellung 2026/ }));
    const xInputs = screen.getAllByLabelText("X (mm)");
    await user.clear(xInputs[1]);
    await user.type(xInputs[1], "150");
    await user.click(screen.getByRole("button", { name: "Bahnhofsmodul an Ports ausrichten" }));

    await waitFor(() => expect(api.previewLayoutConfigurationUnitSnap).toHaveBeenCalledWith(configuration.id, {
      unitId: unit.id, positionXMm: 5, positionYMm: 10, rotationDegrees: 2,
      units: [
        { unitId: unit.id, positionXMm: 5, positionYMm: 10, rotationDegrees: 2 },
        { unitId: targetUnit.id, positionXMm: 150, positionYMm: 0, rotationDegrees: 0 }
      ]
    }));
    expect(screen.getAllByLabelText("X (mm)")[0]).toHaveValue(12);
    expect(screen.getAllByLabelText("Y (mm)")[0]).toHaveValue(20);
    expect(screen.getAllByLabelText("Drehung (Grad)")[0]).toHaveValue(0);
    expect(screen.getByText("Ausrichtung übernommen. Aufbau noch speichern.")).toBeInTheDocument();
    expect(update).not.toHaveBeenCalled();
  });

  it("lists and creates technical positions with app-owned controls", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "createLayoutTechnicalPosition").mockResolvedValue({
      ...technicalPosition, id: "position-2", label: "Weiche 1", kind: "turnout", positionXMm: 120,
      positionYMm: 45, rotationDegrees: 0
    });
    render(<LayoutsView roles={["Planner"]} />);
    await screen.findAllByText(layout.name);

    await user.click(screen.getByRole("tab", { name: "Technik" }));
    expect(await screen.findByText("Einfahrsignal A")).toBeInTheDocument();
    expect(screen.getByText("250 / 80 mm")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Technische Position anlegen" }));

    const dialog = screen.getByRole("dialog", { name: "Technische Position anlegen" });
    await user.type(within(dialog).getByRole("textbox", { name: "Bezeichnung" }), "Weiche 1");
    await user.clear(within(dialog).getByRole("spinbutton", { name: "X-Position (mm)" }));
    await user.type(within(dialog).getByRole("spinbutton", { name: "X-Position (mm)" }), "120");
    await user.clear(within(dialog).getByRole("spinbutton", { name: "Y-Position (mm)" }));
    await user.type(within(dialog).getByRole("spinbutton", { name: "Y-Position (mm)" }), "45");
    await user.click(within(dialog).getByRole("button", { name: "Position speichern" }));

    await waitFor(() => expect(api.createLayoutTechnicalPosition).toHaveBeenCalledWith(unit.id,
      expect.objectContaining({ label: "Weiche 1", kind: "turnout", positionXMm: 120, positionYMm: 45 })));
  });

  it("replaces technical position edits with the selected server version after a conflict", async () => {
    const user = userEvent.setup();
    const serverPosition = { ...technicalPosition, version: 2, positionXMm: 275 };
    vi.mocked(api.layoutTechnicalPositions)
      .mockResolvedValueOnce([technicalPosition])
      .mockResolvedValue([serverPosition]);
    vi.spyOn(api, "updateLayoutTechnicalPosition")
      .mockRejectedValueOnce(new ApiError("Position changed.", "layout_position_version_conflict", 409))
      .mockResolvedValue({ ...serverPosition, label: "Lokaler Signalname", version: 3 });
    render(<LayoutsView roles={["Planner"]} />);
    await screen.findAllByText(layout.name);

    await user.click(screen.getByRole("tab", { name: "Technik" }));
    await screen.findByText("Einfahrsignal A");
    await user.click(screen.getByRole("button", { name: "Einfahrsignal A bearbeiten" }));
    const dialog = screen.getByRole("dialog", { name: "Technische Position bearbeiten" });
    const name = within(dialog).getByRole("textbox", { name: "Bezeichnung" });
    await user.clear(name);
    await user.type(name, "Lokaler Signalname");
    await user.click(within(dialog).getByRole("button", { name: "Position speichern" }));

    expect(await within(dialog).findByText(
      "Die Position wurde zwischenzeitlich geändert. Deine Eingaben bleiben erhalten."
    )).toBeInTheDocument();
    expect(name).toHaveValue("Lokaler Signalname");
    await user.click(within(dialog).getByRole("button", { name: "Serverstand neu laden" }));
    expect(await within(dialog).findByText(
      "Aktueller Serverstand geladen. Der lokale Entwurf wurde ersetzt."
    )).toBeInTheDocument();
    expect(name).toHaveValue("Einfahrsignal A");
    expect(within(dialog).getByRole("spinbutton", { name: "X-Position (mm)" })).toHaveValue(275);
    await user.click(within(dialog).getByRole("button", { name: "Position speichern" }));
    await waitFor(() => expect(api.updateLayoutTechnicalPosition).toHaveBeenLastCalledWith(technicalPosition.id,
      expect.objectContaining({ label: "Einfahrsignal A", positionXMm: 275, expectedVersion: 2 })));
  });

  it.each(["Viewer", "Editor"])("keeps technical positions read-only for %s users", async (role) => {
    const user = userEvent.setup();
    render(<LayoutsView roles={[role]} />);
    await screen.findAllByText(layout.name);

    await user.click(screen.getByRole("tab", { name: "Technik" }));
    expect(await screen.findByText("Einfahrsignal A")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Technische Position anlegen" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Einfahrsignal A bearbeiten" })).not.toBeInTheDocument();
  });

  it("publishes reviewed revisions only after confirmation", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "publishPlanRevision").mockResolvedValue({ ...reviewRevision, status: "published", version: 3,
      publishedBy: "planner", publishedAt: "2026-08-07T12:00:00Z" });
    render(<LayoutsView roles={["Planner"]} />);
    await screen.findAllByText(layout.name);

    await user.click(screen.getByRole("tab", { name: "Planer" }));
    await screen.findByText("Fahrbetrieb");
    await user.click(screen.getByRole("button", { name: "Veröffentlichen" }));
    expect(screen.getByRole("dialog", { name: "Planrevision veröffentlichen" })).toHaveTextContent("Revision 1");
    expect(api.publishPlanRevision).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));
    await waitFor(() => expect(api.publishPlanRevision).toHaveBeenCalledWith(reviewRevision.id, 2));
  });

  it("loads plan variants once while the selected unit is unchanged", async () => {
    const user = userEvent.setup();
    render(<LayoutsView roles={["Viewer"]} />);
    await screen.findAllByText(layout.name);

    await user.click(screen.getByRole("tab", { name: "Planer" }));
    await screen.findByText("Fahrbetrieb");
    await new Promise((resolve) => window.setTimeout(resolve, 50));
    expect(api.planVariants).toHaveBeenCalledTimes(1);
  });

  it("creates variants and drafts and submits drafts for review", async () => {
    const user = userEvent.setup();
    const draft = { ...reviewRevision, status: "draft" as const, version: 1 };
    const draftVariant = { ...variant, revisions: [draft] };
    vi.mocked(api.planVariants).mockResolvedValue([draftVariant]);
    vi.spyOn(api, "createPlanVariant").mockResolvedValue({ ...variant, id: "variant-2", name: "Rangierbetrieb",
      revisions: [] });
    vi.spyOn(api, "createPlanRevision").mockResolvedValue(draft);
    vi.spyOn(api, "submitPlanRevision").mockResolvedValue({ ...draft, status: "review", version: 2 });
    render(<LayoutsView roles={["Planner"]} />);
    await screen.findAllByText(layout.name);

    await user.click(screen.getByRole("tab", { name: "Planer" }));
    const createPanel = screen.getByText("Planvariante anlegen").closest(".panel") as HTMLElement;
    await user.type(within(createPanel).getByLabelText("Bezeichnung"), "Rangierbetrieb");
    await user.click(within(createPanel).getByRole("button", { name: "Planvariante speichern" }));
    await waitFor(() => expect(api.createPlanVariant).toHaveBeenCalledWith(unit.id,
      { name: "Rangierbetrieb", description: undefined }));

    await user.click(screen.getByRole("button", { name: "Neuen Entwurf anlegen" }));
    await waitFor(() => expect(api.createPlanRevision).toHaveBeenCalledWith(variant.id,
      { baseRevisionId: undefined }));
    await user.click(screen.getByRole("button", { name: "Zur Prüfung geben" }));
    await waitFor(() => expect(api.submitPlanRevision).toHaveBeenCalledWith(draft.id, 1));
  });

  it("shows the complete profile and edits through the dialog", async () => {
    const user = userEvent.setup();
    const updated = { ...layout, name: "Neue Bezeichnung", version: 3 };
    vi.spyOn(api, "updateLayout").mockResolvedValue(updated);
    render(<LayoutsView roles={["Planner"]} />);
    await screen.findAllByText(layout.name);

    const profile = (await screen.findByText("Anlagenprofil")).closest(".panel") as HTMLElement;
    expect(profile).toHaveTextContent("Clubanlage");
    expect(profile).toHaveTextContent("Aktiv");
    expect(profile).toHaveTextContent("TT");
    expect(profile).toHaveTextContent("1:120");
    expect(profile).toHaveTextContent("3,50 %");
    expect(profile).toHaveTextContent("40,00 mm");
    expect(profile).toHaveTextContent("700,00 mm");
    expect(profile).toHaveTextContent("Version 2");
    expect(profile).toHaveTextContent(new Date(layout.createdAt).toLocaleString());
    expect(profile).toHaveTextContent(new Date(layout.updatedAt).toLocaleString());

    await user.click(within(profile).getByRole("button", { name: "Bearbeiten" }));
    const dialog = screen.getByRole("dialog", { name: "Anlage bearbeiten" });
    const name = within(dialog).getByRole("textbox", { name: "Bezeichnung" });
    const grade = within(dialog).getByRole("spinbutton", { name: "Maximale Steigung (%)" });
    const clearance = within(dialog).getByRole("spinbutton", {
      name: "Mindestabstand kreuzender Gleise (mm)"
    });
    const flexRadius = within(dialog).getByRole("spinbutton", { name: "Mindest-Flexgleisradius (mm)" });
    expect(name).toHaveValue(layout.name);
    expect(grade).toHaveValue(3.5);
    expect(clearance).toHaveValue(40);
    expect(flexRadius).toHaveValue(700);
    await user.clear(name);
    await user.type(name, updated.name);
    await user.clear(grade);
    await user.type(grade, "2.5");
    await user.clear(clearance);
    await user.type(clearance, "25.5");
    await user.clear(flexRadius);
    await user.type(flexRadius, "650");
    await user.click(within(dialog).getByRole("button", { name: "Änderungen speichern" }));

    await waitFor(() => expect(api.updateLayout).toHaveBeenCalledWith(layout.id, expect.objectContaining({
      name: updated.name, maxGradePercent: 2.5, minimumTrackClearanceMm: 25.5,
      minimumFlexRadiusMm: 650,
      expectedVersion: layout.version
    })));
    expect((await screen.findAllByText(updated.name)).length).toBeGreaterThan(0);
  });

  it("preserves local layout edits on a stale-version conflict", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "updateLayout").mockRejectedValue(new ApiError("Layout data has changed.", "layout_version_conflict", 409));
    render(<LayoutsView roles={["Planner"]} />);
    await screen.findAllByText(layout.name);

    const profile = (await screen.findByText("Anlagenprofil")).closest(".panel") as HTMLElement;
    await user.click(within(profile).getByRole("button", { name: "Bearbeiten" }));
    const dialog = screen.getByRole("dialog", { name: "Anlage bearbeiten" });
    const name = within(dialog).getByRole("textbox", { name: "Bezeichnung" });
    await user.clear(name);
    await user.type(name, "Lokaler Entwurf bleibt erhalten");
    await user.click(within(dialog).getByRole("button", { name: "Änderungen speichern" }));

    expect(await within(dialog).findByText(
      "Die Anlage wurde zwischenzeitlich geändert. Deine Eingaben bleiben erhalten."
    )).toBeInTheDocument();
    expect(name).toHaveValue("Lokaler Entwurf bleibt erhalten");
    await user.click(within(dialog).getByRole("button", { name: "Serverstand neu laden" }));
    await waitFor(() => expect(api.layout).toHaveBeenCalledWith(layout.id));
    await waitFor(() => expect(name).toHaveValue(layout.name));
  });

  it("keeps the layout navigation hidden from Messe users", async () => {
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
    vi.spyOn(api, "version").mockResolvedValue({
      version: "0.1.15", updateAvailable: false, checkedAt: "2026-08-07T10:00:00Z",
      status: "local", message: "local"
    });
    render(<Shell username="messe" roles={["Messe"]} activeView="exhibition" onLogout={() => undefined}>
      <div>Ausstellungsansicht</div>
    </Shell>);

    expect(screen.queryByRole("link", { name: "Anlagen" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Ausstellung" })).toBeInTheDocument();
  });
});
