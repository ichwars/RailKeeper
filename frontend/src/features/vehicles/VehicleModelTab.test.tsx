import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { emptyVehicle } from "./vehicleViewModel";
import { VehicleModelTab } from "./VehicleModelTab";

describe("VehicleModelTab operational fields", () => {
  it("renders speed and home base directly as constrained model inputs", async () => {
    const user = userEvent.setup();
    const onUpdate = vi.fn();

    render(
      <VehicleModelTab
        form={{ ...emptyVehicle, maximumSpeedKmh: 120, homeBase: "Bw Leipzig-West" }}
        externalMappings={[]}
        readonly={false}
        articleSearchLoading={false}
        canRunArticleSearch={false}
        options={{
          manufacturers: [], gauges: [], epochs: [], railwayCompanies: [], categories: [],
          gattungen: [], symbols: [], categoryRelations: []
        }}
        filteredGattungen={[]}
        openSections={{ model: true, details: false, vehicle: false }}
        selectOptions={() => <option value="" />}
        ecosFieldClass={() => ""}
        showRequiredErrors={false}
        onToggleSection={vi.fn()}
        onOpenBarcodeSearch={vi.fn()}
        onRunArticleSearch={vi.fn()}
        onUpdate={onUpdate}
        onUpdateCategory={vi.fn()}
        onOpenQr={vi.fn()}
        canOpenQr={false}
        onUpdateCouplingFront={vi.fn()}
        onUpdateCouplingSame={vi.fn()}
      />
    );

    const speed = screen.getByRole("spinbutton", { name: /Höchstgeschwindigkeit/i });
    expect(speed).toHaveAttribute("min", "1");
    expect(speed).toHaveAttribute("max", "1000");
    expect(speed).toHaveAttribute("step", "1");
    expect(screen.getByText("km/h")).toBeInTheDocument();

    const homeBase = screen.getByRole("textbox", { name: "Heimat-Bw / Einsatzstelle" });
    expect(homeBase).toHaveAttribute("maxlength", "200");

    await user.clear(speed);
    expect(onUpdate).toHaveBeenCalledWith({ maximumSpeedKmh: undefined });
    fireEvent.change(speed, { target: { value: "120.5" } });
    expect(onUpdate).toHaveBeenLastCalledWith({ maximumSpeedKmh: 120.5 });
    fireEvent.change(speed, { target: { value: "1e2" } });
    expect(onUpdate).toHaveBeenLastCalledWith({ maximumSpeedKmh: 100 });
    fireEvent.change(homeBase, { target: { value: "Bw Dresden" } });
    expect(onUpdate).toHaveBeenLastCalledWith({ homeBase: "Bw Dresden" });
  });
});
