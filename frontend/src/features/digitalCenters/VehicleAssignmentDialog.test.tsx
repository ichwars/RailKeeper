import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { setLanguage } from "../../shared/i18n";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";
import { VehicleAssignmentDialog } from "./VehicleAssignmentDialog";

const item: DigitalCenterWorkItem = {
  id: "item-77", sessionId: "session-1", centerObjectId: "77", vehicleId: "",
  name: "BR 106", decoderAddress: 3, protocol: "DCC", compareStatus: "new", stationStatus: "read",
  center: { objectId: 77, name: "BR 106", decoderAddress: 3, protocol: "DCC" },
  railkeeper: {}, proposed: {}, conflicts: [],
  createdAt: "2026-08-23T08:00:00Z", updatedAt: "2026-08-23T08:00:00Z"
};

const vehicles = [
  vehicleFixture({
    id: "address-match", inventoryNumber: "RK-LOK-000001", name: "Andere Lok", digitalDecoderNumber: "3"
  }),
  vehicleFixture({
    id: "name-match", inventoryNumber: "RK-LOK-000002", name: "BR 106", digitalDecoderNumber: "8"
  }),
  vehicleFixture({ id: "ordinary", inventoryNumber: "RK-LOK-000003", name: "V 200" })
];

describe("VehicleAssignmentDialog", () => {
  beforeEach(() => setLanguage("de"));

  it("searches candidates and assigns only an explicitly selected vehicle", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    const onAssign = vi.fn();
    render(<VehicleAssignmentDialog item={item} provider="ecos" vehicles={vehicles} selectedVehicleId=""
      loading={false} saving={false} error="" onSelect={onSelect} onAssign={onAssign} onClose={vi.fn()} />);

    const search = screen.getByRole("textbox", { name: "Fahrzeuge durchsuchen" });
    expect(search).toHaveFocus();
    expect(screen.getByText("Decoderadresse stimmt überein")).toBeInTheDocument();
    expect(screen.getByText("Name stimmt überein")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Fahrzeug zuordnen" })).toBeDisabled();
    expect(screen.getAllByRole("radio").every((radio) => !(radio as HTMLInputElement).checked)).toBe(true);

    await user.type(search, "BR 106");
    expect(screen.getByText("RK-LOK-000002 · BR 106")).toBeInTheDocument();
    expect(screen.queryByText("RK-LOK-000003 · V 200")).not.toBeInTheDocument();

    await user.click(screen.getByRole("radio", { name: /RK-LOK-000002 · BR 106/ }));
    expect(onSelect).toHaveBeenCalledWith("name-match");
  });

  it("confirms the selected vehicle and exposes errors accessibly", async () => {
    const user = userEvent.setup();
    const onAssign = vi.fn();
    render(<VehicleAssignmentDialog item={item} provider="ecos" vehicles={vehicles} selectedVehicleId="address-match"
      loading={false} saving={false} error="Zuordnung fehlgeschlagen" onSelect={vi.fn()}
      onAssign={onAssign} onClose={vi.fn()} />);

    expect(screen.getByRole("alert")).toHaveTextContent("Zuordnung fehlgeschlagen");
    await user.click(screen.getByRole("button", { name: "Fahrzeug zuordnen" }));
    expect(onAssign).toHaveBeenCalledWith("address-match");
  });

  it("closes on Escape", () => {
    const onClose = vi.fn();
    render(<VehicleAssignmentDialog item={item} provider="ecos" vehicles={vehicles} selectedVehicleId=""
      loading={false} saving={false} error="" onSelect={vi.fn()} onAssign={vi.fn()} onClose={onClose} />);

    fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });
    expect(onClose).toHaveBeenCalledOnce();
  });
});
