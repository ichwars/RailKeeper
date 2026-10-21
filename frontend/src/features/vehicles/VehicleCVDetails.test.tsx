import { useState } from "react";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { cvFileFixture, cvValueFixture } from "../../test/fixtures/vehicles";
import { VehicleCVDetails } from "./VehicleCVDetails";

const values = [cvValueFixture({ id: "address", cvNumber: 1, value: 3, description: "" }),
  cvValueFixture({ id: "config", cvNumber: 29, value: 14, description: "" })];

function renderDetails(value = values[0], readonly = false) {
  const close = vi.fn(), edit = vi.fn();
  render(<VehicleCVDetails value={value} values={values} files={[]} readonly={readonly} saving={false} onClose={close} onEdit={edit} />);
  return { dialog: within(screen.getByRole("dialog")), close, edit };
}

describe("CV detail side panel", () => {
  it("shows exactly the selected CV with decimal, hexadecimal and binary values", () => {
    const { dialog } = renderDetails();
    expect(dialog.getByRole("heading", { name: "CV 1 · Kurze DCC-Adresse" })).toBeVisible();
    expect(dialog.getByText("0x03")).toBeVisible();
    expect(dialog.getByText("00000011")).toBeVisible();
    expect(dialog.queryByText("00001110")).not.toBeInTheDocument();
    expect(dialog.queryByText("Bit 0")).not.toBeInTheDocument();
    expect(dialog.getAllByText("Nicht erfasst")).toHaveLength(2);
    expect(dialog.getByText(/Originaldatei nicht erfasst/)).toBeVisible();
  });
  it("explains CV29 bits only for the base CV29, preserving indexed identity", () => {
    const { dialog } = renderDetails({ ...values[1], indexCV31: 0, indexCV32: 0 });
    expect(dialog.getByText("CV31=0 / CV32=0")).toBeVisible();
    expect(dialog.queryByText("Bit 0")).not.toBeInTheDocument();
    expect(dialog.getByText(/noch nicht verifiziert/)).toBeVisible();
  });
  it("shows CV29's own six decoded settings", () => {
    const { dialog } = renderDetails(values[1]);
    expect(dialog.getByText("0x0E")).toBeVisible();
    expect(dialog.getByText("00001110")).toBeVisible();
    expect(dialog.getByText("Bit 5")).toBeVisible();
    expect(dialog.getByText("Kurze Adresse aktiv")).toBeVisible();
  });
  it("edits the exact indexed record", () => {
    const indexed = { ...values[0], id: "indexed", cvNumber: 257, indexCV31: 16, indexCV32: 3 };
    const { dialog, edit } = renderDetails(indexed);
    fireEvent.click(dialog.getByRole("button", { name: "CV-Wert bearbeiten" }));
    expect(edit).toHaveBeenCalledExactlyOnceWith(indexed);
  });
  it("keeps the raw value accessible in read-only mode", () => {
    const { dialog } = renderDetails(values[0], true);
    expect(dialog.getByRole("button", { name: "CV-Wert bearbeiten" })).toBeDisabled();
    expect(dialog.getByText("0x03")).toBeVisible();
  });
  it("uses only a linked file as provenance", () => {
    const file = cvFileFixture({ id: "linked", originalName: "Quelle.esux" });
    render(<VehicleCVDetails value={{ ...values[0], sourceFileId: file.id }} values={values} files={[file]}
      readonly={false} saving={false} onClose={vi.fn()} onEdit={vi.fn()} />);
    expect(screen.getByText("Quelle.esux")).toBeVisible();
    expect(screen.getAllByText("Nicht erfasst")).toHaveLength(2);
  });
  it("shows the original export timestamp without applying a timezone", () => {
    const { dialog } = renderDetails({ ...values[0], sourceLine: 4, sourceExportedAt: "2026-10-17T17:41:59" });
    expect(dialog.getByText("4")).toBeVisible();
    expect(dialog.getByText("17.10.2026 17:41:59")).toBeVisible();
    expect(dialog.getByText("Importierte CV-Liste")).toBeVisible();
    expect(dialog.queryByText("Nicht erfasst")).not.toBeInTheDocument();
  });
  it("locks the background and restores focus when Escape closes the panel", () => {
    function Harness() {
      const [open, setOpen] = useState(false);
      return <div role="dialog" aria-label="Fahrzeug"><button onClick={() => setOpen(true)}>CV öffnen</button>
        {open && <VehicleCVDetails value={values[0]} values={values} files={[]} readonly={false} saving={false}
          onClose={() => setOpen(false)} onEdit={vi.fn()} />}</div>;
    }
    render(<Harness />);
    const button = screen.getByRole("button", { name: "CV öffnen" });
    button.focus(); fireEvent.click(button);
    expect(button.closest("[role=dialog]")).toHaveAttribute("inert");
    const close = screen.getByRole("button", { name: "Details schließen" });
    expect(close).toHaveFocus();
    const metadata = screen.getByText("Datensatzangaben");
    fireEvent.keyDown(close, { key: "Tab", shiftKey: true });
    expect(metadata).toHaveFocus();
    fireEvent.keyDown(metadata, { key: "Tab" });
    expect(close).toHaveFocus();
    fireEvent.keyDown(close, { key: "Escape" });
    expect(button.closest("[role=dialog]")).not.toHaveAttribute("inert");
    expect(button).toHaveFocus();
  });

  it("keeps the exported reading separate from the factory reference and shows the exact source", () => {
    const decoderProfile = "LokPilot 5 micro DCC";
    const value = cvValueFixture({ cvNumber: 112, value: 42, decoderProfile,
      sourceLine: 72, sourceExportedAt: "2026-10-17T17:41:59" });
    const rows = [value, cvValueFixture({ cvNumber: 8, value: 151, decoderProfile })];
    const before = structuredClone(rows);
    const edit = vi.fn();
    render(<VehicleCVDetails value={value} values={rows} files={[]} readonly={false} saving={false}
      onClose={vi.fn()} onEdit={edit} />);
    expect(screen.getByText("0,065536 s je Schritt")).toBeVisible();
    expect(screen.getByText("20")).toBeVisible();
    expect(screen.getByText("Referenz laut Anleitung, kein ausgelesener Fahrzeugwert.")).toBeVisible();
    expect(screen.getByText("S.85, CV112")).toBeVisible();
    expect(screen.getByText("72")).toBeVisible();
    expect(screen.getByText("17.10.2026 17:41:59")).toBeVisible();
    expect(screen.getAllByRole("link")).toHaveLength(1);
    fireEvent.click(screen.getByRole("button", { name: "CV-Wert bearbeiten" }));
    expect(edit).toHaveBeenCalledExactlyOnceWith(value);
    expect(rows).toEqual(before);
  });

  it("explains reserved registers neutrally and never invents unit, default or bit meanings", () => {
    const decoderProfile = "LokPilot 5 micro DCC";
    const value = cvValueFixture({ cvNumber: 141, value: 25, decoderProfile });
    render(<VehicleCVDetails value={value} values={[value,
      cvValueFixture({ cvNumber: 8, value: 151, decoderProfile })]} files={[]}
      readonly={true} saving={false} onClose={vi.fn()} onEdit={vi.fn()} />);
    expect(screen.getByRole("heading", { name: "CV 141 · Reserviert laut ESU-Handbuch" })).toBeVisible();
    expect(screen.getByText(/Gespeicherte Werte bleiben erhalten/).closest("div")).toHaveClass("unassigned");
    expect(screen.queryByText("Dokumentierte Einheit")).not.toBeInTheDocument();
    expect(screen.queryByText("Dokumentangabe ab Werk")).not.toBeInTheDocument();
    expect(screen.queryByText("Bit 0")).not.toBeInTheDocument();
  });
});
