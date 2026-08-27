import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ComponentProps } from "react";
import { describe, expect, it, vi } from "vitest";

import { emptyVehicle } from "./vehicleViewModel";
import { VehicleCreateStepDetails } from "./VehicleCreateStepDetails";
import { VehicleModelTab } from "./VehicleModelTab";
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

    const selectOptions: ComponentProps<typeof VehicleModelTab>["selectOptions"] = (
      entries,
      currentValue,
      emptyLabel
    ) => <>
      <option value="">{emptyLabel}</option>
      {entries.map((entry) => <option key={entry.id} value={entry.key}>{entry.label}</option>)}
      {currentValue && !entries.some((entry) => entry.key === currentValue) &&
        <option value={currentValue}>{currentValue}</option>}
    </>;
    const model = {
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
      selectOptions,
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
    };
    const { rerender } = render(<VehicleCreateStepDetails state={state} dispatch={vi.fn()} model={model} />);

    expect(screen.getByText("Gattung")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Gattung Noch offen" }));
    expect(screen.getByRole("option", { name: "Reisezugwagen" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Erwerb" })).toBeEnabled();

    rerender(<VehicleCreateStepDetails state={{ ...state, activeDetailsTab: "member:0" }}
      dispatch={vi.fn()} model={model} />);
    expect(screen.getByText("Höchstgeschwindigkeit")).toBeInTheDocument();
    expect(screen.getByLabelText("Digital")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Kategorie Ausgefüllt" })).toBeEnabled();
    await user.click(screen.getByRole("button", { name: "Details" }));
    rerender(<VehicleCreateStepDetails state={{ ...state, activeDetailsTab: "member:0" }}
      dispatch={vi.fn()} model={{ ...model, openSections: { ...model.openSections, details: true } }} />);
    expect(screen.getByText("Radsatz")).toBeInTheDocument();

    rerender(<VehicleCreateStepDetails state={{ ...state, activeDetailsTab: "member:0" }}
      dispatch={vi.fn()} model={{ ...model, openSections: { ...model.openSections, vehicle: true } }} />);
    expect(screen.queryByText("Erwerb")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Zustand" })).toBeEnabled();
  });
});
