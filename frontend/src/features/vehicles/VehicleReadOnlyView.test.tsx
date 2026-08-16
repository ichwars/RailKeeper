import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { vehicleFixture } from "../../test/fixtures/vehicles";
import { VehicleReadOnlyView } from "./VehicleReadOnlyView";

describe("VehicleReadOnlyView operational fields", () => {
  it("shows prototype operation separately from the storage location", () => {
    render(
      <VehicleReadOnlyView
        vehicle={vehicleFixture({
          maximumSpeedKmh: 120,
          homeBase: "Bw Leipzig-West",
          storageLocation: "Vitrine 1"
        })}
        onEdit={vi.fn()}
        onPrint={vi.fn()}
        onQr={vi.fn()}
        onPreviewImage={vi.fn()}
      />
    );

    const section = screen.getByRole("heading", { name: "Vorbild & Betrieb" }).closest("section");
    expect(section).toHaveTextContent("Höchstgeschwindigkeit");
    expect(section).toHaveTextContent("120 km/h");
    expect(section).toHaveTextContent("Bw Leipzig-West");
    expect(section).not.toHaveTextContent("Vitrine 1");
  });
});
