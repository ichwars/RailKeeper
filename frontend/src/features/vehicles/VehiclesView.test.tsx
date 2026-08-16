import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { VehiclesView } from "./VehiclesView";

describe("VehiclesView", () => {
  beforeEach(() => {
    window.localStorage.clear();
    window.sessionStorage.clear();
    vi.spyOn(api, "vehicles").mockResolvedValue([vehicleFixture()]);
    vi.spyOn(api, "masterDataAll").mockResolvedValue({});
    vi.spyOn(api, "masterDataRelations").mockResolvedValue([]);
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
  });

  it("loads and renders the inventory", async () => {
    render(<VehiclesView username="tester" />);

    expect((await screen.findAllByText("BR 106")).length).toBeGreaterThan(0);
    expect(api.vehicles).toHaveBeenCalledWith("");
    expect(api.masterDataAll).toHaveBeenCalledWith();
    expect(api.masterDataRelations).toHaveBeenCalledWith("vehicle_category", "vehicle_gattung");
  });

  it("shows inventory loading failures", async () => {
    vi.mocked(api.vehicles).mockRejectedValue(new Error("Bestand nicht erreichbar"));

    render(<VehiclesView username="tester" />);

    expect(await screen.findByText("Bestand nicht erreichbar")).toBeInTheDocument();
  });
});
