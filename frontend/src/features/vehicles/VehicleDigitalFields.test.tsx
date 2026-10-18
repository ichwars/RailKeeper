import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { VehicleDigitalFields } from "./VehicleDigitalFields";
import { cvValueFixture } from "../../test/fixtures/vehicles";
import { emptyVehicle } from "./vehicleViewModel";

const values = [[1, 3], [8, 151], [19, 0], [21, 0], [22, 0], [27, 24], [29, 14], [254, 0], [255, 0]]
  .map(([cvNumber, value]) => cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "LokPilot 5 micro DCC" }));

describe("vehicle digital settings", () => {
  it("shows unknowns for missing CVs while keeping DT editable independently", () => {
    const update = vi.fn();
    render(<VehicleDigitalFields form={{ ...emptyVehicle, dtDecoder: true, dtDecoderNumber: "55" }} values={[]}
      readonly={false} update={update} />);
    expect(screen.getByDisplayValue("55")).toBeEnabled();
    expect(screen.getAllByText("Unbekannt").length).toBeGreaterThan(2);
    fireEvent.change(screen.getByDisplayValue("55"), { target: { value: "66" } });
    expect(update).toHaveBeenCalledExactlyOnceWith({ dtDecoderNumber: "66" });
    expect(screen.getByRole("button", { name: "Aus CVs übernehmen" })).toBeDisabled();
  });
  it("derives consist and braking states and applies vehicle fields only explicitly", () => {
    const update = vi.fn();
    render(<VehicleDigitalFields form={emptyVehicle} values={values} readonly={false} update={update} />);
    expect(screen.getByText("Nicht aktiv")).toBeVisible();
    expect(screen.getByText("Keine freigegeben")).toBeVisible();
    expect(screen.getAllByText("nein")).toHaveLength(2);
    expect(update).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Aus CVs übernehmen" }));
    expect(update).toHaveBeenCalledExactlyOnceWith({ digital: true, digitalDecoderNumber: "3",
      decoderType: "LokPilot 5 micro DCC", abcBrakes: false });
  });
  it("keeps all vehicle write controls locked in read-only mode", () => {
    render(<VehicleDigitalFields form={emptyVehicle} values={values} readonly={true} update={vi.fn()} />);
    expect(screen.getByRole("button", { name: "Aus CVs übernehmen" })).toBeDisabled();
    for (const checkbox of screen.getAllByRole("checkbox")) expect(checkbox).toBeDisabled();
  });
  it("keeps identity values visible and additional stored fields behind the disclosure", () => {
    const open = vi.fn();
    render(<VehicleDigitalFields form={emptyVehicle} values={values} readonly={false} update={vi.fn()} onOpenCV={open} />);
    const more = screen.getByText("Weitere Digitalangaben").closest("details")!;
    expect(more).not.toHaveAttribute("open");
    const identity = screen.getByLabelText("Firmware").closest(".digital-metadata")!;
    expect(within(identity as HTMLElement).getByText("LokPilot 5 micro DCC")).toBeVisible();
    expect(within(identity as HTMLElement).getByText("ESU")).toBeVisible();
    expect(within(identity as HTMLElement).getByText("Unbekannt")).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "CV-Auswertung öffnen" }));
    expect(open).toHaveBeenCalledOnce();
  });

});
