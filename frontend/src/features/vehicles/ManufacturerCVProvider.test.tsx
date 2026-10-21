import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";
import { api } from "../../shared/api";
import { DecoderLibraryProvider, decoderLibraryChanged } from "../../shared/DecoderLibraryProvider";
import { bundledDecoderInventory, bundledDecoderProfiles, type DecoderCVReference } from "../../shared/decoderLibraryModel";
import { cvValueFixture } from "../../test/fixtures/vehicles";
import { ManufacturerCVProvider, ManufacturerCVStatus, useManufacturerCVHints } from "./ManufacturerCVProvider";
import { VehicleCVDetails } from "./VehicleCVDetails";
import { VehicleCVValuesTable } from "./VehicleCVValuesTable";

const values = [cvValueFixture({ id: "reading", cvNumber: 54, value: 42, decoderProfile: "Kühn N025",
  description: "Meine Beschreibung", sourceLine: 17, sourceExportedAt: "2026-10-17T17:41:59" }),
cvValueFixture({ id: "maker", cvNumber: 8, value: 157, decoderProfile: "Kühn N025" })];
const model = bundledDecoderInventory.models.find(model => model.name === "N025")!;
const ref: DecoderCVReference = {
  id: "n025", modelIds: [model.id], revision: "proof", active: true, status: "partial",
  title: { de: "Hersteller", en: "Manufacturer" }, note: { de: "Gilt vorbehaltlich der Firmware.", en: "Check firmware." },
  sources: [{ title: "Originalanleitung", version: "b", url: "https://example.test/manual.pdf", scope: "N025" }],
  rows: [{ cv: 54, label: "Herstellerbezeichnung", page: 4, sourceIndex: 0, indexScope: "base", applicability: "manual", factoryValue: "15" }]
};
beforeEach(() => { vi.spyOn(api, "decoderReferences").mockResolvedValue([structuredClone(ref)]); });

function Harness({ profile = "Kühn N025" }: { profile?: string }) {
  const hints = useManufacturerCVHints();
  return <><ManufacturerCVStatus /><span data-testid="hint">{hints.lookup({ ...values[0], decoderProfile: profile })?.row.label || "unbekannt"}</span></>;
}

it("loads one model on demand, retains manual descriptions and exposes qualified source/default without edits", async () => {
  const before = structuredClone(values);
  const edit = vi.fn();
  render(<ManufacturerCVProvider values={values} profile="Kühn N025" manualType="">
    <VehicleCVValuesTable values={values} readonly saving={false} edit={edit} remove={vi.fn()} />
    <VehicleCVDetails value={values[0]} values={values} files={[]} readonly saving={false} onClose={vi.fn()} onEdit={edit} />
  </ManufacturerCVProvider>);
  await screen.findByRole("heading", { name: "CV 54 · Herstellerbezeichnung" });
  expect(screen.getByText("Meine Beschreibung", { selector: "td" })).toBeInTheDocument();
  expect(screen.getByRole("link", { name: "Originalanleitung · b" })).toHaveAttribute("href", "https://example.test/manual.pdf#page=4");
  expect(screen.getByText("Handbuchreferenz; Modell- und Firmwarebedingungen beachten")).toBeInTheDocument();
  expect(screen.getByText("15")).toBeInTheDocument(); expect(screen.getByText("42", { selector: "strong" })).toBeInTheDocument();
  expect(screen.getByText("17.10.2026 17:41:59")).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "CV-Wert bearbeiten" })).toBeDisabled();
  expect(api.decoderReferences).toHaveBeenCalledTimes(1); expect(edit).not.toHaveBeenCalled(); expect(values).toEqual(before);
});

it("keeps loading/failure/gaps explicit and retries without fabricating a heading", async () => {
  vi.mocked(api.decoderReferences).mockRejectedValueOnce(new Error("offline"));
  render(<ManufacturerCVProvider values={values} profile="Kühn N025" manualType=""><Harness /></ManufacturerCVProvider>);
  expect(screen.getByRole("status")).toHaveTextContent("werden geladen");
  expect(await screen.findByRole("button", { name: "Erneut laden" })).toBeInTheDocument();
  expect(screen.getByTestId("hint")).toHaveTextContent("unbekannt");
  vi.mocked(api.decoderReferences).mockResolvedValueOnce([]);
  fireEvent.click(screen.getByRole("button", { name: "Erneut laden" }));
  await screen.findByText("Für dieses Modell ist noch keine belastbare CV-Tabelle hinterlegt.");
});

it("ignores late responses and never reuses references across models or CV profile groups", async () => {
  let resolve: (references: DecoderCVReference[]) => void = () => {};
  vi.mocked(api.decoderReferences).mockImplementationOnce(() => new Promise(done => { resolve = done; }));
  const { rerender } = render(<ManufacturerCVProvider values={values} profile="Kühn N025" manualType=""><Harness /></ManufacturerCVProvider>);
  const other = values.map(value => ({ ...value, decoderProfile: "N45" }));
  vi.mocked(api.decoderReferences).mockResolvedValueOnce([]);
  rerender(<ManufacturerCVProvider values={other} profile="N45" manualType=""><Harness profile="N45" /></ManufacturerCVProvider>);
  await screen.findByText("Für dieses Modell ist noch keine belastbare CV-Tabelle hinterlegt.");
  await act(async () => resolve([ref]));
  expect(screen.getByTestId("hint")).toHaveTextContent("unbekannt");
  expect(api.decoderReferences).toHaveBeenCalledTimes(2);
});

it("refreshes the references after library activation changes", async () => {
  vi.spyOn(api, "decoderLibrary").mockResolvedValue({ schemaVersion: 1, profiles: bundledDecoderProfiles, inventory: bundledDecoderInventory });
  render(<DecoderLibraryProvider><ManufacturerCVProvider values={values} profile="Kühn N025" manualType=""><Harness /></ManufacturerCVProvider></DecoderLibraryProvider>);
  await waitFor(() => expect(screen.getByTestId("hint")).toHaveTextContent("Herstellerbezeichnung"));
  vi.mocked(api.decoderReferences).mockResolvedValueOnce([]);
  act(() => window.dispatchEvent(new Event(decoderLibraryChanged)));
  await screen.findByText("Für dieses Modell ist noch keine belastbare CV-Tabelle hinterlegt.");
  expect(screen.getByTestId("hint")).toHaveTextContent("unbekannt");
});
