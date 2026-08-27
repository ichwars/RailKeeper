import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { InventoryNumberScheme } from "../../shared/api";
import { emptyOptions } from "./vehicleViewModel";
import {
  createVehicleCreateWizardState,
  vehicleCreateWizardReducer,
  type VehicleCreateWizardState
} from "./vehicleCreateWizardState";
import { VehicleCreateStepBasics, inventoryNumberPreview } from "./VehicleCreateStepBasics";

const scheme: InventoryNumberScheme = {
  id: "scheme-set", category: "Set", prefix: "RK-SET", nextNumber: 1, padding: 6,
  active: true, preview: "RK-SET-000001", createdAt: "2026-08-17", updatedAt: "2026-08-17"
};

describe("VehicleCreateStepBasics", () => {
  it("shows the approved fields, provisional set number and safe member count", () => {
    let state: VehicleCreateWizardState = { ...createVehicleCreateWizardState(), kind: "set" };
    const dispatch = vi.fn((action) => { state = vehicleCreateWizardReducer(state, action); });
    const { rerender } = render(
      <VehicleCreateStepBasics state={state} dispatch={dispatch} options={emptyOptions}
        filteredGattungen={[]} selectOptions={() => <option value="" />}
        setScheme={scheme} setSchemeLoading={false} setSchemeError="" setCreationDisabled={false}
        onUpdateShared={vi.fn()} onUpdateCategory={vi.fn()} />
    );

    expect(screen.getByRole("radio", { name: /Set/ })).toHaveAttribute("aria-checked", "true");
    expect(screen.getByText(/RK-SET-000001/)).toHaveTextContent(/vorläufig/i);
    expect(inventoryNumberPreview(scheme)).toBe("RK-SET-000001");
    expect(screen.getByLabelText("Hersteller Noch offen")).toBeVisible();
    expect(screen.getByLabelText("Bezeichnung Noch offen")).toBeVisible();
    expect(screen.getByLabelText("Artikel-Nr.")).toBeVisible();
    expect(screen.getByLabelText("Spurweite Noch offen")).toBeVisible();
    expect(screen.getByLabelText("Kategorie Noch offen")).toBeVisible();
    const count = screen.getByLabelText(/Anzahl Fahrzeuge/);
    expect(count).toHaveAttribute("min", "2");

    state = {
      ...state,
      members: [
        state.members[0],
        { ...state.members[1], touched: true, form: { ...state.members[1].form, name: "Wagen 2" } },
        { ...state.members[0] }
      ]
    };
    rerender(<VehicleCreateStepBasics state={state} dispatch={dispatch} options={emptyOptions}
      filteredGattungen={[]} selectOptions={() => <option value="" />}
      setScheme={scheme} setSchemeLoading={false} setSchemeError="" setCreationDisabled={false}
      onUpdateShared={vi.fn()} onUpdateCategory={vi.fn()} />);
    fireEvent.change(screen.getByLabelText(/Anzahl Fahrzeuge/), { target: { value: "2" } });
    expect(dispatch).toHaveBeenCalledWith({ type: "set-member-count", count: 2 });
  });

  it("disables the set choice for ECoS and explains a missing Set scheme", () => {
    const state = createVehicleCreateWizardState();
    render(<VehicleCreateStepBasics state={state} dispatch={vi.fn()} options={emptyOptions}
      filteredGattungen={[]} selectOptions={() => <option value="" />}
      setScheme={null} setSchemeLoading={false} setSchemeError="Kein aktives Nummernschema für Sets."
      setCreationDisabled onUpdateShared={vi.fn()} onUpdateCategory={vi.fn()} />);

    expect(screen.getByRole("radio", { name: /Set/ })).toBeDisabled();
    expect(screen.getByText("Kein aktives Nummernschema für Sets.")).toBeVisible();
  });
});
