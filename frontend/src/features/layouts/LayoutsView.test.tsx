import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  ApiError,
  api,
  type Layout,
  type LayoutConfiguration,
  type LayoutUnit,
  type PlanRevision,
  type PlanVariant
} from "../../shared/api";
import { Shell } from "../../app/Shell";
import { LayoutsView } from "./LayoutsView";

const layout: Layout = {
  id: "layout-1", name: "Clubanlage mit langem Bahnhofsnamen", kind: "club", gauge: "TT", scale: "1:120",
  description: "Modulare Clubanlage", version: 2, archived: false, createdAt: "2026-08-07T10:00:00Z",
  updatedAt: "2026-08-07T11:00:00Z"
};

const unit: LayoutUnit = {
  id: "unit-1", layoutId: layout.id, name: "Bahnhofsmodul", kind: "module", ownerLabel: "Daniel",
  widthMm: 1200, heightMm: 500, version: 1, archived: false, createdAt: "2026-08-07T10:00:00Z",
  updatedAt: "2026-08-07T10:00:00Z"
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
    vi.spyOn(api, "planVariants").mockResolvedValue([variant]);
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
    await user.click(within(dialog).getByRole("button", { name: "Anlage speichern" }));
    await waitFor(() => expect(api.createLayout).toHaveBeenCalledWith(expect.objectContaining({
      name: "Clubanlage", kind: "private", gauge: "TT", scale: "1:120"
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

  it("preserves local layout edits on a stale-version conflict", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "updateLayout").mockRejectedValue(new ApiError("Layout data has changed.", "layout_version_conflict", 409));
    render(<LayoutsView roles={["Planner"]} />);
    await screen.findAllByText(layout.name);

    const editPanel = (await screen.findByText("Anlage bearbeiten")).closest(".panel") as HTMLElement;
    const name = within(editPanel).getByLabelText("Bezeichnung");
    await user.clear(name);
    await user.type(name, "Lokaler Entwurf bleibt erhalten");
    await user.click(within(editPanel).getByRole("button", { name: "Änderungen speichern" }));

    expect(await screen.findByText("Die Anlage wurde zwischenzeitlich geändert. Deine Eingaben bleiben erhalten."))
      .toBeInTheDocument();
    expect(name).toHaveValue("Lokaler Entwurf bleibt erhalten");
    expect(screen.getByRole("button", { name: "Serverstand neu laden" })).toBeInTheDocument();
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
