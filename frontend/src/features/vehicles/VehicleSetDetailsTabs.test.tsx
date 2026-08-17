import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { createVehicleCreateWizardState } from "./vehicleCreateWizardState";
import { VehicleSetDetailsTabs } from "./VehicleSetDetailsTabs";

describe("VehicleSetDetailsTabs", () => {
  it("separates shared set data from isolated ordered member panels", async () => {
    const user = userEvent.setup();
    const state = {
      ...createVehicleCreateWizardState(), kind: "set" as const, activeDetailsTab: "set" as const,
      members: [
        { form: { manufacturer: "Roco", name: "Wagen 1", gauge: "H0", vehicleNumber: "1" }, touched: true },
        { form: { manufacturer: "Roco", name: "Wagen 2", gauge: "H0", vehicleNumber: "50 50 23-11 011-5" }, touched: true }
      ]
    };
    const dispatch = vi.fn();
    const { rerender } = render(<VehicleSetDetailsTabs state={state} dispatch={dispatch}
      setPanel={<p>Anschaffungsdaten</p>} memberPanel={(index) => <p>Fahrzeugnummer {state.members[index].form.vehicleNumber}</p>} />);

    expect(screen.getByRole("tab", { name: /Set & Anschaffung/ })).toHaveAttribute("aria-selected", "true");
    await user.click(screen.getByRole("tab", { name: /Wagen 2/ }));
    expect(dispatch).toHaveBeenCalledWith({ type: "set-active-details-tab", tab: "member:1" });

    rerender(<VehicleSetDetailsTabs state={{ ...state, activeDetailsTab: "member:1" }} dispatch={dispatch}
      setPanel={<p>Anschaffungsdaten</p>} memberPanel={(index) => <p>Fahrzeugnummer {state.members[index].form.vehicleNumber}</p>} />);
    expect(screen.getByRole("tabpanel")).toHaveTextContent("50 50 23-11 011-5");

		const secondTab = screen.getByRole("tab", { name: /Wagen 2/ });
		secondTab.focus();
		await user.keyboard("{ArrowLeft}");
		expect(dispatch).toHaveBeenCalledWith({ type: "set-active-details-tab", tab: "member:0" });
  });

  it("disables adding another member at the backend limit", () => {
    const state = {
      ...createVehicleCreateWizardState(),
      kind: "set" as const,
      activeDetailsTab: "set" as const,
      members: Array.from({ length: 100 }, (_, index) => ({
        form: { manufacturer: "Roco", name: `Wagen ${index + 1}`, gauge: "H0" },
        touched: false,
        overriddenFields: []
      }))
    };

    render(<VehicleSetDetailsTabs state={state} dispatch={vi.fn()}
      setPanel={<p>Set</p>} memberPanel={() => null} />);

    expect(screen.getByRole("button", { name: /Fahrzeug hinzufügen/ })).toBeDisabled();
  });
});
