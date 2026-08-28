import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { ApiError, api, type Layout, type LayoutConfiguration, type LayoutTwin, type LayoutUnit } from "../../shared/api";
import { LayoutTwinPanel } from "./LayoutTwinPanel";

const layout: Layout = {
  id: "layout-1", name: "Clubanlage", kind: "club", gauge: "TT", scale: "1:120",
  version: 1, archived: false, createdAt: "2026-08-09T10:00:00Z", updatedAt: "2026-08-09T10:00:00Z"
};
const unit: LayoutUnit = {
  id: "unit-1", layoutId: layout.id, name: "Bahnhof", kind: "module", widthMm: 1000, heightMm: 400,
  version: 1, archived: false, createdAt: "2026-08-09T10:00:00Z", updatedAt: "2026-08-09T10:00:00Z"
};
const configuration: LayoutConfiguration = {
  id: "configuration-1", layoutId: layout.id, name: "Ausstellung", version: 1, archived: false,
  units: [{ unitId: unit.id, positionXMm: 0, positionYMm: 0, rotationDegrees: 0, sortOrder: 0 }],
  createdAt: "2026-08-09T10:00:00Z", updatedAt: "2026-08-09T10:00:00Z"
};
const twin: LayoutTwin = {
  layoutId: layout.id, configurationId: configuration.id, configurationName: configuration.name,
  bounds: { minXMm: 0, minYMm: 0, widthMm: 1000, heightMm: 400 }, hasGeometry: true,
  warnings: [{ code: "outline_fallback", unitId: unit.id }],
  units: [{
    id: unit.id, name: unit.name, kind: unit.kind, positionXMm: 0, positionYMm: 0, rotationDegrees: 0,
    version: unit.version,
    localOutline: [{ xMm: 0, yMm: 0 }, { xMm: 1000, yMm: 0 },
      { xMm: 1000, yMm: 400 }, { xMm: 0, yMm: 400 }],
    outline: [{ xMm: 0, yMm: 0 }, { xMm: 1000, yMm: 0 }, { xMm: 1000, yMm: 400 }, { xMm: 0, yMm: 400 }],
    positions: [{
      id: "position-reserved", layoutUnitId: unit.id, label: "Einfahrsignal A", kind: "signal",
      localXMm: 150, localYMm: 80, localRotationDegrees: 90,
      globalXMm: 150, globalYMm: 80, rotationDegrees: 90, version: 1, outsideOutline: false,
      productId: "product-1", inventoryNumber: "RK-ART-000001", manufacturer: "Tillig",
      articleNumber: "83101", productName: "Lichtsignal", description: "Gleis 1",
      statuses: ["reserved"], installations: [], reservations: [{
        id: "reservation-1", productId: "product-1", inventoryNumber: "RK-ART-000001",
        manufacturer: "Tillig", articleNumber: "83101", productName: "Lichtsignal", quantity: 1,
        reservationStatus: "active", digitalAddress: "17", connection: "J3"
      }]
    }, {
      id: "position-defective", layoutUnitId: unit.id, label: "Weiche West", kind: "turnout",
      localXMm: 500, localYMm: 120, localRotationDegrees: 0,
      globalXMm: 500, globalYMm: 120, rotationDegrees: 0, version: 1, outsideOutline: false,
      statuses: ["installed", "defective"], reservations: [], installations: [{
        id: "installation-1", productId: "product-2", inventoryNumber: "RK-ART-000002",
        manufacturer: "Tillig", productName: "Weichenantrieb", quantity: 1,
        installationCondition: "defective"
      }]
    }]
  }]
};

describe("LayoutTwinPanel", () => {
  afterEach(() => vi.restoreAllMocks());

  it("renders the transformed twin, filters statuses, and opens the inspector by click and keyboard", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "layoutTwin").mockResolvedValue(twin);
    const history = vi.spyOn(api, "accessoryUsageHistory").mockImplementation(async (productID) => ({
      productId: productID,
      events: []
    }));
    render(<LayoutTwinPanel layout={layout} units={[unit]} configurations={[configuration]} canPlan />);

    expect(await screen.findByRole("img", { name: "Grafische Anlagenübersicht mit technischen Positionen" }))
      .toBeInTheDocument();
    expect(api.layoutTwin).toHaveBeenCalledWith(layout.id, { configurationId: configuration.id });
    const reservedMarker = screen.getByRole("button", { name: /Einfahrsignal A, Reserviert/ });
    await user.click(reservedMarker);
    const inspector = screen.getByLabelText("Positionsinspektor");
    expect(within(inspector).getByText("RK-ART-000001")).toBeInTheDocument();
    expect(within(inspector).getByText("Digitaladresse: 17")).toBeInTheDocument();
    expect(within(inspector).getByText("Anschluss: J3")).toBeInTheDocument();

    await user.click(screen.getByRole("checkbox", { name: "Eingebaut (1)" }));
    await user.click(screen.getByRole("checkbox", { name: "Defekt (1)" }));
    expect(screen.queryByRole("button", { name: /Weiche West/ })).not.toBeInTheDocument();
    await user.click(screen.getByRole("checkbox", { name: "Eingebaut (1)" }));
    await user.click(screen.getByRole("checkbox", { name: "Defekt (1)" }));
    const defectiveMarker = screen.getByRole("button", { name: /Weiche West, Eingebaut, Defekt/ });
    defectiveMarker.focus();
    await user.keyboard("{Enter}");
    await waitFor(() => expect(screen.getByLabelText("Positionsinspektor")).toHaveTextContent("Weiche West"));
    await waitFor(() => expect(history).toHaveBeenCalledWith("product-2"));
  });

  it("uses the app-owned empty state when no unit or configuration exists", () => {
    const spy = vi.spyOn(api, "layoutTwin");
    render(<LayoutTwinPanel layout={layout} units={[]} configurations={[]} canPlan={false} />);
    expect(screen.getByText("Zuerst eine aktive Anlageneinheit oder Aufbaukonfiguration anlegen.")).toBeInTheDocument();
    expect(spy).not.toHaveBeenCalled();
  });

  it("keeps editing role-gated and autosaves keyboard and pointer changes", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "layoutTwin").mockResolvedValue(twin);
    const updatePosition = vi.spyOn(api, "updateLayoutTechnicalPosition").mockResolvedValue({
      id: "position-reserved", layoutUnitId: unit.id, label: "Einfahrsignal A", kind: "signal",
      positionXMm: 151, positionYMm: 80, rotationDegrees: 90, version: 2, archived: false,
      createdAt: "2026-08-09T10:00:00Z", updatedAt: "2026-08-09T11:00:00Z"
    });
    const updateOutline = vi.spyOn(api, "updateLayoutUnitOutline").mockResolvedValue({
      layoutUnitId: unit.id, points: twin.units[0].localOutline, version: 2
    });
    render(<LayoutTwinPanel layout={layout} units={[unit]} configurations={[configuration]} canPlan />);
    await screen.findByRole("img", { name: "Grafische Anlagenübersicht mit technischen Positionen" });
    await user.click(screen.getByRole("button", { name: "Bearbeiten" }));
    expect(screen.getByText(/Bearbeitungsmodus aktiv/)).toBeInTheDocument();
    expect(screen.getAllByLabelText(/Konturpunkt \d, Bahnhof/)).toHaveLength(4);

    const marker = screen.getByRole("button", { name: /Einfahrsignal A, Reserviert/ });
    marker.focus();
    await user.keyboard("{ArrowRight}");
    await waitFor(() => expect(updatePosition).toHaveBeenCalledWith("position-reserved",
      expect.objectContaining({ positionXMm: 151, positionYMm: 80, expectedVersion: 1 })), { timeout: 1500 });
    expect(api.layoutTwin).toHaveBeenCalledTimes(1);

    const canvas = screen.getByRole("img", { name: "Grafische Anlagenübersicht mit technischen Positionen" });
    vi.spyOn(canvas, "getBoundingClientRect").mockReturnValue({
      left: 0, top: 0, width: 1000, height: 400, right: 1000, bottom: 400, x: 0, y: 0,
      toJSON: () => ({})
    });
    const handle = screen.getByLabelText("Konturpunkt 1, Bahnhof");
    fireEvent.pointerDown(handle, { pointerId: 7, clientX: 40, clientY: 40 });
    fireEvent.pointerMove(canvas, { pointerId: 7, clientX: 55, clientY: 55 });
    fireEvent.pointerUp(canvas, { pointerId: 7, clientX: 55, clientY: 55 });
    await waitFor(() => expect(updateOutline).toHaveBeenCalledWith(unit.id,
      expect.objectContaining({ expectedVersion: 1 })), { timeout: 4000 });
  }, 10000);

  it("keeps independent autosaves when two positions are edited within the debounce window", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "layoutTwin").mockResolvedValue(twin);
    const updatePosition = vi.spyOn(api, "updateLayoutTechnicalPosition").mockResolvedValue({
      id: "position-reserved", layoutUnitId: unit.id, label: "Einfahrsignal A", kind: "signal",
      positionXMm: 151, positionYMm: 80, rotationDegrees: 90, version: 2, archived: false,
      createdAt: "2026-08-09T10:00:00Z", updatedAt: "2026-08-09T11:00:00Z"
    });
    render(<LayoutTwinPanel layout={layout} units={[unit]} configurations={[configuration]} canPlan />);
    await screen.findByRole("img", { name: "Grafische Anlagenübersicht mit technischen Positionen" });
    await user.click(screen.getByRole("button", { name: "Bearbeiten" }));

    const reserved = screen.getByRole("button", { name: /Einfahrsignal A, Reserviert/ });
    reserved.focus();
    await user.keyboard("{ArrowRight}");
    const defective = screen.getByRole("button", { name: /Weiche West, Eingebaut, Defekt/ });
    defective.focus();
    await user.keyboard("{ArrowDown}");

    await waitFor(() => expect(updatePosition).toHaveBeenCalledTimes(2), { timeout: 2000 });
    expect(updatePosition).toHaveBeenCalledWith("position-reserved", expect.objectContaining({
      positionXMm: 151, positionYMm: 80
    }));
    expect(updatePosition).toHaveBeenCalledWith("position-defective", expect.objectContaining({
      positionXMm: 500, positionYMm: 121
    }));
  });

  it("preserves the local edit on a version conflict and hides editing from viewers", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "layoutTwin").mockResolvedValue(twin);
    vi.spyOn(api, "updateLayoutTechnicalPosition").mockRejectedValue(
      new ApiError("changed", "layout_position_version_conflict", 409)
    );
    const { rerender } = render(
      <LayoutTwinPanel layout={layout} units={[unit]} configurations={[configuration]} canPlan />
    );
    await screen.findByRole("button", { name: "Bearbeiten" });
    await user.click(screen.getByRole("button", { name: "Bearbeiten" }));
    const marker = screen.getByRole("button", { name: /Einfahrsignal A, Reserviert/ });
    marker.focus();
    await user.keyboard("{ArrowRight}");
    expect(await screen.findByText(/Der Serverstand wurde zwischenzeitlich geändert/, {}, { timeout: 1500 }))
      .toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Lokalen Entwurf verwerfen" })).toBeInTheDocument();

    rerender(<LayoutTwinPanel layout={layout} units={[unit]} configurations={[configuration]} canPlan={false} />);
    expect(screen.queryByRole("button", { name: "Bearbeitung beenden" })).not.toBeInTheDocument();
  });
});
