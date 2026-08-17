import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { emptyVehicle } from "./vehicleViewModel";
import { VehicleCreateStepDetails } from "./VehicleCreateStepDetails";
import { createVehicleCreateWizardState } from "./vehicleCreateWizardState";

describe("VehicleCreateStepDetails", () => {
  it("keeps the required shared gattung editable on the set details tab", async () => {
    const user = userEvent.setup();
    const state = {
      ...createVehicleCreateWizardState({
        ...emptyVehicle,
        name: "Rekowagenset",
        manufacturer: "Roco",
        gauge: "TT",
        category: "Personenwagen",
        gattung: ""
      }),
      kind: "set" as const,
      step: "details" as const,
      activeDetailsTab: "set" as const
    };

    render(<VehicleCreateStepDetails state={state} dispatch={vi.fn()} model={{
      form: state.shared,
      externalMappings: [],
      readonly: false,
      articleSearchLoading: false,
      canRunArticleSearch: false,
      options: {
        manufacturers: [], gauges: [], epochs: [], railwayCompanies: [], categories: [],
        gattungen: [{ id: "gattung-1", type: "vehicle_gattung", key: "reisezugwagen",
          label: "Reisezugwagen", active: true, sortOrder: 1, metadata: {}, createdAt: "", updatedAt: "" }],
        symbols: [], categoryRelations: []
      },
      filteredGattungen: [{ id: "gattung-1", type: "vehicle_gattung", key: "reisezugwagen",
        label: "Reisezugwagen", active: true, sortOrder: 1, metadata: {}, createdAt: "", updatedAt: "" }],
      openSections: { model: true, details: false, vehicle: false },
      selectOptions: (entries, currentValue, emptyLabel) => <>
        <option value="">{emptyLabel}</option>
        {entries.map((entry) => <option key={entry.id} value={entry.key}>{entry.label}</option>)}
        {currentValue && !entries.some((entry) => entry.key === currentValue) &&
          <option value={currentValue}>{currentValue}</option>}
      </>,
      ecosFieldClass: () => "",
      showRequiredErrors: false,
      onToggleSection: vi.fn(),
      onOpenBarcodeSearch: vi.fn(),
      onRunArticleSearch: vi.fn(),
      onUpdate: vi.fn(),
      onUpdateCategory: vi.fn(),
      onOpenQr: vi.fn(),
      canOpenQr: false,
      onUpdateCouplingFront: vi.fn(),
      onUpdateCouplingSame: vi.fn()
    }} />);

    expect(screen.getByText("Gattung")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Gattung" }));
    expect(screen.getByRole("option", { name: "Reisezugwagen" })).toBeInTheDocument();
  });
});
