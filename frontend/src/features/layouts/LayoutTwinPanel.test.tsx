import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api, type Layout, type LayoutConfiguration, type LayoutTwin, type LayoutUnit } from "../../shared/api";
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
      globalXMm: 150, globalYMm: 80, rotationDegrees: 90, version: 1,
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
      globalXMm: 500, globalYMm: 120, rotationDegrees: 0, version: 1,
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
    render(<LayoutTwinPanel layout={layout} units={[unit]} configurations={[configuration]} />);

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
  });

  it("uses the app-owned empty state when no unit or configuration exists", () => {
    const spy = vi.spyOn(api, "layoutTwin");
    render(<LayoutTwinPanel layout={layout} units={[]} configurations={[]} />);
    expect(screen.getByText("Zuerst eine aktive Anlageneinheit oder Aufbaukonfiguration anlegen.")).toBeInTheDocument();
    expect(spy).not.toHaveBeenCalled();
  });
});
