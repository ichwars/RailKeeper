import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
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

  it("loads a set member that is absent from the filtered inventory", async () => {
    const user = userEvent.setup();
    const setSummary = {
      id: "set-1",
      inventoryNumber: "RK-SET-000001",
      name: "Rheingold",
      manufacturer: "Roco",
      articleNumber: "45923",
      gauge: "H0",
      memberCount: 2,
      position: 1
    };
    const visibleMember = vehicleFixture({ id: "visible-member", vehicleSet: setSummary });
    const hiddenMember = vehicleFixture({
      id: "hidden-member",
      inventoryNumber: "RK-WAG-000099",
      name: "Verdeckter Wagen",
      vehicleSet: { ...setSummary, position: 2 }
    });
    vi.mocked(api.vehicles).mockResolvedValue([visibleMember]);
    vi.spyOn(api, "vehicleSet").mockResolvedValue({
      ...setSummary,
      members: [visibleMember, hiddenMember],
      createdAt: "2026-08-17T00:00:00Z",
      updatedAt: "2026-08-17T00:00:00Z"
    });
    vi.spyOn(api, "vehicle").mockResolvedValue(hiddenMember);

    render(<VehiclesView username="tester" />);
    await user.click(await screen.findByRole("button", { name: "Rheingold" }));
    await user.click(await screen.findByRole("button", { name: /Verdeckter Wagen/ }));

    expect(api.vehicle).toHaveBeenCalledWith(hiddenMember.id);
  });
});
