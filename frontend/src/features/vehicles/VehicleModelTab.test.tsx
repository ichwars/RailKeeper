import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { emptyVehicle } from "./vehicleViewModel";
import { VehicleModelTab } from "./VehicleModelTab";
import { createVehicleInventoryRenderers } from "./vehicleInventoryRenderers";

describe("VehicleModelTab operational fields", () => {
  it("locks shared set fields while leaving member fields editable", () => {
    render(
      <VehicleModelTab
        {...({
          form: { ...emptyVehicle, manufacturer: "Roco", articleNumber: "123", name: "Wagen" },
          externalMappings: [], readonly: false, sharedFieldsReadonly: true,
          articleSearchLoading: false, canRunArticleSearch: true,
          options: { manufacturers: [], gauges: [], epochs: [], railwayCompanies: [], categories: [], gattungen: [], symbols: [], categoryRelations: [] },
          filteredGattungen: [], openSections: { model: true, details: false, vehicle: false },
          selectOptions: () => <option value="" />, ecosFieldClass: () => "", showRequiredErrors: false,
          onToggleSection: vi.fn(), onOpenBarcodeSearch: vi.fn(), onRunArticleSearch: vi.fn(),
          onUpdate: vi.fn(), onUpdateCategory: vi.fn(), onOpenQr: vi.fn(), canOpenQr: false,
          onUpdateCouplingFront: vi.fn(), onUpdateCouplingSame: vi.fn()
        })}
      />
    );

    expect(screen.getByRole("textbox", { name: "Artikel-Nr." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Hersteller: Ausgefüllt" })).toBeDisabled();
    expect(screen.getByRole("textbox", { name: "Bezeichnung: Ausgefüllt" })).toBeEnabled();
  });

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

  it("shows only the current inactive manufacturer with an inactive suffix", async () => {
    const user = userEvent.setup();
    const base = {
      id: "manufacturer:piko",
      type: "manufacturer",
      key: "piko",
      label: "Piko",
      active: true,
      sortOrder: 0,
      metadata: {},
      createdAt: "2026-08-16T00:00:00Z",
      updatedAt: "2026-08-16T00:00:00Z"
    };
    const renderers = createVehicleInventoryRenderers({
      sort: { key: "inventoryNumber", direction: "asc" },
      quickMenuVehicleID: "",
      setQuickMenuVehicleID: vi.fn(),
      toggleSort: vi.fn(),
      openDetail: vi.fn(),
      openEdit: vi.fn(),
      openQr: vi.fn(),
      printVehicle: vi.fn(),
      setDeleteCandidate: vi.fn(),
      t: (key) => key === "common.inactive" ? "inaktiv" : key
    });

    render(
      <VehicleModelTab
        form={{ ...emptyVehicle, manufacturer: "Roco" }}
        externalMappings={[]}
        readonly={false}
        articleSearchLoading={false}
        canRunArticleSearch={false}
        options={{
          manufacturers: [
            base,
            { ...base, id: "manufacturer:roco", key: "roco", label: "Roco", active: false },
            { ...base, id: "manufacturer:esu", key: "esu", label: "ESU", active: false }
          ],
          gauges: [], epochs: [], railwayCompanies: [], categories: [], gattungen: [],
          symbols: [], categoryRelations: []
        }}
        filteredGattungen={[]}
        openSections={{ model: true, details: false, vehicle: false }}
        selectOptions={renderers.selectOptions}
        ecosFieldClass={() => ""}
        showRequiredErrors={false}
        onToggleSection={vi.fn()}
        onOpenBarcodeSearch={vi.fn()}
        onRunArticleSearch={vi.fn()}
        onUpdate={vi.fn()}
        onUpdateCategory={vi.fn()}
        onOpenQr={vi.fn()}
        canOpenQr={false}
        onUpdateCouplingFront={vi.fn()}
        onUpdateCouplingSame={vi.fn()}
      />
    );

    await user.click(screen.getByRole("button", { name: "Hersteller: Ausgefüllt" }));
    expect(screen.getByRole("option", { name: "Roco (inaktiv)" })).toBeDisabled();
    expect(screen.getByRole("option", { name: "Piko" })).toBeEnabled();
    expect(screen.queryByRole("option", { name: "ESU" })).not.toBeInTheDocument();
  });
});
