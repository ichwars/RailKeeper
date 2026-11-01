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
  it("shows the mutable RailComPlus name without changing the vehicle name", () => {
    const update = vi.fn();
    const name = [..."Decodername", "\0"].map((character, index) => cvValueFixture({
      id: `name-${index}`, cvNumber: 261 + index, value: character.charCodeAt(0),
      indexCV31: 1, indexCV32: 0, decoderProfile: "LokPilot 5 micro DCC"
    }));
    render(<VehicleDigitalFields form={{ ...emptyVehicle, name: "Meine Lok" }} values={[...values, ...name]}
      readonly={false} update={update} />);
    expect(screen.getByLabelText("RailComPlus-Anmeldename")).toHaveTextContent("Decodername");
    expect(update).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Aus CVs übernehmen" }));
    expect(update.mock.calls[0][0]).not.toHaveProperty("name");
  });
  it("applies known decoder and braking fields from a partial CV list without changing addresses", () => {
    const update = vi.fn();
    render(<VehicleDigitalFields form={{ ...emptyVehicle, digitalDecoderNumber: "77", dtDecoderNumber: "55" }}
      values={values.filter(value => value.cvNumber === 8 || value.cvNumber === 27)}
      readonly={false} update={update} />);
    const button = screen.getByRole("button", { name: "Aus CVs übernehmen" });
    expect(button).toBeEnabled();
    fireEvent.click(button);
    expect(update).toHaveBeenCalledExactlyOnceWith({ decoderType: "LokPilot 5 micro DCC", abcBrakes: false });
  });
  it("can apply a verified decoder type even when address and ABC state are unknown", () => {
    const update = vi.fn();
    render(<VehicleDigitalFields form={emptyVehicle} values={values.filter(value => value.cvNumber === 8)}
      readonly={false} update={update} />);
    fireEvent.click(screen.getByRole("button", { name: "Aus CVs übernehmen" }));
    expect(update).toHaveBeenCalledExactlyOnceWith({ decoderType: "LokPilot 5 micro DCC" });
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

  it("uses Lenz ABC evidence instead of ESU CV27 and displays the confirmed model", () => {
    const update = vi.fn();
    const lenz = [[8, 99], [27, 0], [51, 2]].map(([cvNumber, value]) =>
      cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "STANDARD+ V2" }));
    render(<VehicleDigitalFields form={emptyVehicle} values={lenz} readonly={false} update={update} />);
    expect(screen.getByText("CV51=2 · Bit 1")).toBeVisible();
    expect(screen.getByText("Lenz")).toBeVisible();
    expect(screen.getByText("ja")).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "Aus CVs übernehmen" }));
    expect(update).toHaveBeenCalledExactlyOnceWith({ decoderType: "STANDARD+ V2", abcBrakes: true });
  });

  it("shows independent Lenz constant braking without changing manual fields or inventing consist direction", () => {
    const update = vi.fn();
    const lenz = [[8, 99], [7, 79], [19, 42], [51, 128], [52, 50]].map(([cvNumber, value]) =>
      cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "STANDARD+ V2" }));
    const before = structuredClone(lenz);
    render(<VehicleDigitalFields form={{ ...emptyVehicle, digitalDecoderNumber: "77", dtDecoder: true,
      dtDecoderNumber: "55", abcBrakes: true }} values={lenz} readonly={false} update={update} />);
    expect(screen.getByLabelText("DCC-Verbundadresse")).toHaveTextContent(/^42$/);
    expect(screen.getByText("CV51=128")).toBeVisible();
    expect(screen.getByText("ja")).toBeVisible();
    expect(screen.getByText("nein")).toBeVisible();
    expect(screen.getByLabelText("Firmware")).toHaveTextContent("Unbekannt");
    expect(screen.getByDisplayValue("55")).toHaveValue("55");
    expect(update).not.toHaveBeenCalled();
    expect(lenz).toEqual(before);
  });

  it("does not borrow DH18A ABC rules for the catalogued PD06A", () => {
    const pd = [[8, 97], [27, 3]].map(([cvNumber, value]) =>
      cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "PD06A" }));
    render(<VehicleDigitalFields form={emptyVehicle} values={pd} readonly={false} update={vi.fn()} />);
    expect(screen.getByText("Doehler & Haass")).toBeVisible();
    expect(screen.getByRole("button", { name: "Aus CVs übernehmen" })).toBeDisabled();
    expect(screen.queryByText("ja")).not.toBeInTheDocument();
  });

  it("recognizes DH18A from raw CVs and changes the vehicle only after explicit application", () => {
    const update = vi.fn();
    const imported = [[8, 97], [261, 180], [27, 3]].map(([cvNumber, value]) =>
      cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "Zentralenexport" }));
    render(<VehicleDigitalFields form={emptyVehicle} values={imported} readonly={false} update={update} />);
    expect(screen.getByRole("status", { name: "Decoder-Typ" })).toHaveTextContent("DH18A");
    expect(screen.getByText("CV8=97 · CV261=180")).toBeVisible();
    expect(update).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Aus CVs übernehmen" }));
    expect(update).toHaveBeenCalledExactlyOnceWith({ decoderType: "DH18A", abcBrakes: true });
  });

  it("keeps an automatically identified ZIMO base type separate from interface and ABC rules", () => {
    const update = vi.fn();
    const imported = [[8, 145], [250, 127], [27, 3]].map(([cvNumber, value]) =>
      cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "Zentralenexport" }));
    render(<VehicleDigitalFields form={emptyVehicle} values={imported} readonly={false} update={update} />);
    expect(screen.getByRole("status", { name: "Decoder-Typ" })).toHaveTextContent("MN180");
    expect(screen.getByText(/Anschlussvariante und modellspezifische CV-Regeln/)).toBeVisible();
    expect(screen.queryByText("ja")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Aus CVs übernehmen" }));
    expect(update).toHaveBeenCalledExactlyOnceWith({ decoderType: "MN180" });
  });

  it("uses Lenz CV7 as a type identifier without inventing firmware", () => {
    const update = vi.fn();
    const imported = [[8, 99], [7, 79], [27, 0], [51, 2]].map(([cvNumber, value]) =>
      cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "Zentralenexport" }));
    render(<VehicleDigitalFields form={emptyVehicle} values={imported} readonly={false} update={update} />);
    expect(screen.getByLabelText("Firmware")).toHaveTextContent("Unbekannt");
    fireEvent.click(screen.getByRole("button", { name: "Aus CVs übernehmen" }));
    expect(update).toHaveBeenCalledExactlyOnceWith({ decoderType: "STANDARD+ V2", abcBrakes: true });
  });

});

it("keeps the manually named decoder when applying an exact cached ECoS profile without CV8", () => {
  const update = vi.fn();
  const source = { name: "LokPilot 5 micro DCC", objectId: "1002", host: "ecos.test", readAt: "2026-10-27T07:00:00Z" };
  const imported = values.filter(row => row.cvNumber !== 8).map(row => ({ ...row, decoderProfile: "ECoS" }));
  render(<VehicleDigitalFields form={{ ...emptyVehicle, decoderType: " ESU LokPilot 5 micro DCC " }} values={imported}
    ecosProfile={source} readonly={false} update={update} />);
  expect(screen.getByRole("status", { name: "Decoder-Typ" })).toHaveTextContent("LokPilot 5 micro DCC");
  expect(screen.getByLabelText("Firmware")).toHaveTextContent("Unbekannt");
  expect(update).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Aus CVs übernehmen" }));
  expect(update).toHaveBeenCalledExactlyOnceWith({ digital: true, digitalDecoderNumber: "3", abcBrakes: false });
});

it("does not present an unknown N45 consist direction as forward or copy unreviewed permissions", () => {
  const values = [[8, 157], [19, 42], [21, 255], [22, 255], [27, 3]]
    .map(([cvNumber, value]) => cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "N45" }));
  const before = structuredClone(values), update = vi.fn();
  render(<VehicleDigitalFields form={emptyVehicle} values={values} readonly={false} update={update} />);
  expect(screen.getByLabelText("DCC-Verbundadresse")).toHaveTextContent(/^42$/);
  expect(screen.getByText(/im geprüften N45-Handbuch nicht erläutert/)).toBeVisible();
  expect(screen.queryByText("F1")).not.toBeInTheDocument();
  expect(update).not.toHaveBeenCalled();
  expect(values).toEqual(before);
});

it("shows the extended ZIMO consist and configured brake mode without copying them into manual fields", () => {
  const zimo = [[8, 145], [19, 162], [20, 12], [21, 1], [22, 192], [140, 11]]
    .map(([cvNumber, value]) => cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "MN180N18" }));
  const before = structuredClone(zimo), update = vi.fn();
  render(<VehicleDigitalFields form={{ ...emptyVehicle, dtDecoder: true, dtDecoderNumber: "55" }}
    values={zimo} readonly={false} update={update} />);
  expect(screen.getByLabelText("DCC-Verbundadresse")).toHaveTextContent("1234 · Rückwärts");
  expect(screen.getByText("CV19=162 · CV20=12")).toBeVisible();
  expect(screen.getByText(/CV140=11 · Signalhalt, sofortiger Bremsbeginn/)).toBeVisible();
  expect(screen.getByText("F13")).toBeVisible();
  expect(screen.getByText("F27")).toBeVisible();
  expect(screen.queryByText("F9")).not.toBeInTheDocument();
  expect(screen.getByDisplayValue("55")).toHaveValue("55");
  expect(update).not.toHaveBeenCalled();
  expect(zimo).toEqual(before);
});
