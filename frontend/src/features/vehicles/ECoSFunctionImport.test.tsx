import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { api } from "../../shared/api";
import { decoderAnalysis } from "./decoderAnalysis";
import { ECoSFunctionImport } from "./ECoSFunctionImport";
import { emptyFunctionEdit } from "./vehicleViewModel";

const snapshot = { objectId: 1002, host: "192.0.2.10", readAt: "2026-10-26T08:00:00Z",
  functions: [{ index: 0, description: 3, momentary: true }] };
const props = { vehicleID: "vehicle-1", objectID: "1002", analysis: decoderAnalysis([], ""), mappings: [], symbols: [],
  locked: false, libraryUnavailable: false, functionEdit: emptyFunctionEdit };
afterEach(() => vi.restoreAllMocks());

it("shows a large unknown descriptor and keeps the other keys available", async () => {
  vi.spyOn(api, "readVehicleECoSFunctions").mockResolvedValue({ ...snapshot,
    functions: [...snapshot.functions, { index: 3, description: 7431, momentary: false }] });
  const update = vi.fn();
  render(<ECoSFunctionImport {...props} updateFunctionEdit={update} />);
  fireEvent.click(screen.getByText("Funktionstasten aus ECoS übernehmen"));
  fireEvent.click(screen.getByRole("button", { name: "Aus ECoS lesen" }));
  expect(await screen.findByRole("checkbox", { name: "F3" })).toBeEnabled();
  expect(screen.getByText(/7431/)).toBeVisible();
  expect(screen.getByRole("checkbox", { name: "F0" })).toBeEnabled();
  expect(update).not.toHaveBeenCalled();
});

it("requires a read and explicit selection; applying changes only the draft", async () => {
  const read = vi.spyOn(api, "readVehicleECoSFunctions").mockResolvedValue(snapshot), update = vi.fn();
  const save = vi.spyOn(api, "updateVehicleFunction");
  render(<ECoSFunctionImport {...props} updateFunctionEdit={update} />);
  fireEvent.click(screen.getByText("Funktionstasten aus ECoS übernehmen"));
  expect(read).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Aus ECoS lesen" }));
  expect(await screen.findByRole("checkbox", { name: "F0" })).not.toBeChecked();
  expect(screen.getByRole("button", { name: "Auswahl in Entwurf übernehmen" })).toBeDisabled();
  fireEvent.click(screen.getByRole("checkbox", { name: "F0" }));
  fireEvent.click(screen.getByRole("button", { name: "Auswahl in Entwurf übernehmen" }));
  expect(update).toHaveBeenCalledWith("F0", expect.objectContaining({ mode: "moment" }));
  expect(save).not.toHaveBeenCalled();
  expect(screen.getByRole("status")).toHaveTextContent("Speichere das Fahrzeug");
});

it("rechecks manual edits made during the read, and blocks a mismatched object", async () => {
  const update = vi.fn(); let finish!: (value: typeof snapshot) => void;
  vi.spyOn(api, "readVehicleECoSFunctions").mockImplementation(() => new Promise(resolve => { finish = resolve; }));
  const { rerender } = render(<ECoSFunctionImport {...props} updateFunctionEdit={update} />);
  fireEvent.click(screen.getByText("Funktionstasten aus ECoS übernehmen"));
  fireEvent.click(screen.getByRole("button", { name: "Aus ECoS lesen" }));
  rerender(<ECoSFunctionImport {...props} updateFunctionEdit={update}
    functionEdit={() => ({ persisted: true, mode: "dauer", name: "Manual", symbolKey: "manual" })} />);
  finish(snapshot);
  expect(await screen.findByRole("checkbox", { name: "F0" })).toBeDisabled();
  expect(update).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Aus ECoS lesen" }));
  finish({ ...snapshot, objectId: 1003 });
  expect(await screen.findByRole("alert")).toHaveTextContent("Zuordnung");
  expect(screen.queryByRole("checkbox", { name: "F0" })).not.toBeInTheDocument();
});

it("requires a fresh selection when decoder-library evidence changes", async () => {
  vi.spyOn(api, "readVehicleECoSFunctions").mockResolvedValue(snapshot);
  const update = vi.fn(), { rerender } = render(<ECoSFunctionImport {...props} updateFunctionEdit={update} />);
  fireEvent.click(screen.getByText("Funktionstasten aus ECoS übernehmen"));
  fireEvent.click(screen.getByRole("button", { name: "Aus ECoS lesen" }));
  fireEvent.click(await screen.findByRole("checkbox", { name: "F0" }));
  rerender(<ECoSFunctionImport {...props} analysis={decoderAnalysis([], "changed")}
    updateFunctionEdit={update} />);
  expect(screen.getByRole("checkbox", { name: "F0" })).not.toBeChecked();
  expect(update).not.toHaveBeenCalled();
});

it("blocks unavailable libraries, readonly modes and missing mappings; read failures retain no preview", async () => {
  vi.spyOn(api, "readVehicleECoSFunctions").mockRejectedValue(new Error("offline"));
  const update = vi.fn(), { rerender } = render(<ECoSFunctionImport {...props} objectID={undefined} updateFunctionEdit={update} />);
  fireEvent.click(screen.getByText("Funktionstasten aus ECoS übernehmen"));
  expect(screen.getByRole("button", { name: "Aus ECoS lesen" })).toBeDisabled();
  for (const blocked of [{ libraryUnavailable: true }, { locked: true }]) {
    rerender(<ECoSFunctionImport {...props} {...blocked} updateFunctionEdit={update} />);
    expect(screen.getByRole("button", { name: "Aus ECoS lesen" })).toBeDisabled();
  }
  rerender(<ECoSFunctionImport {...props} updateFunctionEdit={update} />);
  fireEvent.click(screen.getByRole("button", { name: "Aus ECoS lesen" }));
  await waitFor(() => expect(screen.getByRole("alert")).toHaveTextContent("nicht vollständig"));
  expect(update).not.toHaveBeenCalled();
});
