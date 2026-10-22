import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { useState } from "react";
import { LocomotiveImageExportDialog } from "./LocomotiveImageExportDialog";
import { loadLocomotiveImage, renderLocomotiveImage, suggestedImageBounds } from "./locomotiveImageCanvas";

vi.mock("./locomotiveImageCanvas", () => ({
  loadLocomotiveImage: vi.fn(), renderLocomotiveImage: vi.fn(), suggestedImageBounds: vi.fn()
}));
const bitmap = { width: 1000, height: 200, close: vi.fn() };
const image = { id: "original", url: "/api/v1/images/original", title: "BR 86", source: "local" };
const revoke = vi.fn();
let sequence = 0;
beforeEach(() => {
  vi.mocked(loadLocomotiveImage).mockResolvedValue(bitmap);
  vi.mocked(suggestedImageBounds).mockReturnValue({ x: 0, y: 0, width: 1000, height: 200 });
  vi.mocked(renderLocomotiveImage).mockImplementation(async () => new Blob(["output"]));
  vi.stubGlobal("URL", class extends URL {
    static createObjectURL() { return `blob:export-${++sequence}`; }
    static revokeObjectURL = revoke;
  });
  // Canvas pixels are covered by binary/browser tests; these tests exercise dialog lifecycle and controls.
  const context = { clearRect: vi.fn(), drawImage: vi.fn(), fillRect: vi.fn(), strokeRect: vi.fn() };
  vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(context as unknown as CanvasRenderingContext2D);
});
afterEach(() => { vi.unstubAllGlobals(); revoke.mockClear(); bitmap.close.mockClear(); });

it("shows the output and all targets, protects the source, and updates filename/background together", async () => {
  const before = structuredClone(image);
  const { unmount } = render(<LocomotiveImageExportDialog image={image} onClose={vi.fn()} />);
  expect(await screen.findByRole("link", { name: "Bild herunterladen" })).toHaveAttribute("download", "BR-86_ecos.bmp");
  expect(screen.getByText("Technisch verwendbar")).toBeVisible();
  expect(screen.getAllByRole("button", { name: "Schließen" })[0]).toHaveFocus();
  expect(screen.queryByRole("option", { name: "Transparent", hidden: true })).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole("button", { name: "Zielgerät / Anwendung" }));
  fireEvent.click(screen.getByRole("option", { name: "Uhlenbrock Intellibox 3 · IB3 Konfiguration" }));
  await waitFor(() => expect(screen.getByRole("link", { name: "Bild herunterladen" })).toHaveAttribute("download", "BR-86_intellibox3.png"));
  expect(screen.getByText(/wandelt die Datei ins Geräteformat/)).toBeVisible();
  expect(screen.getByRole("button", { name: "Hintergrund" })).toHaveTextContent("Transparent");
  expect(image).toEqual(before); expect(revoke).toHaveBeenCalled();
  unmount(); expect(bitmap.close).toHaveBeenCalled();
});

it("blocks invalid crops, retains the editable controls and recovers without changing the source", async () => {
  render(<LocomotiveImageExportDialog image={image} onClose={vi.fn()} />);
  await screen.findByRole("link", { name: "Bild herunterladen" });
  fireEvent.click(screen.getByText("Ausschnitt in Pixeln einstellen"));
  fireEvent.change(screen.getByLabelText("Breite"), { target: { value: "1200" } });
  expect(screen.getByRole("alert")).toHaveTextContent("innerhalb des Originals");
  expect(screen.getByRole("button", { name: "Bild herunterladen" })).toBeDisabled();
  fireEvent.click(screen.getByRole("button", { name: "Ganzes Bild" }));
  expect(await screen.findByRole("link", { name: "Bild herunterladen" })).toBeVisible();
});

it("starts with the visible crop suggestion and lets the user restore the whole source", async () => {
  const suggestion = { x: 0, y: 60, width: 1000, height: 80 };
  vi.mocked(suggestedImageBounds).mockReturnValueOnce(suggestion);
  render(<LocomotiveImageExportDialog image={image} onClose={vi.fn()} />);
  await screen.findByRole("link", { name: "Bild herunterladen" });
  expect(vi.mocked(renderLocomotiveImage).mock.calls.at(-1)?.[2]).toEqual(suggestion);
  fireEvent.click(screen.getByRole("button", { name: "Ganzes Bild" }));
  await waitFor(() => expect(vi.mocked(renderLocomotiveImage).mock.calls.at(-1)?.[2]).toEqual({ x: 0, y: 0, width: 1000, height: 200 }));
  fireEvent.click(screen.getByRole("button", { name: "Außenränder entfernen" }));
  await waitFor(() => expect(vi.mocked(renderLocomotiveImage).mock.calls.at(-1)?.[2]).toEqual(suggestion));
});

it("warns about enlargement and never presents an empty transparent image as usable", async () => {
  vi.mocked(loadLocomotiveImage).mockResolvedValueOnce({ ...bitmap, width: 20, height: 5 });
  vi.mocked(suggestedImageBounds).mockReturnValueOnce({ x: 0, y: 0, width: 20, height: 5 });
  const { unmount } = render(<LocomotiveImageExportDialog image={image} onClose={vi.fn()} />);
  await screen.findByRole("link", { name: "Bild herunterladen" });
  expect(screen.getByText("Qualität prüfen")).toBeVisible();
  expect(screen.getByText(/muss vergrößert werden/)).toBeVisible();
  unmount();
  vi.mocked(suggestedImageBounds).mockReturnValueOnce(null);
  render(<LocomotiveImageExportDialog image={image} onClose={vi.fn()} />);
  expect(await screen.findByRole("alert")).toHaveTextContent("vollständig transparent");
  expect(screen.getByRole("button", { name: "Bild herunterladen" })).toBeDisabled();
});

it("reports inaccessible sources and cancels a pending load when closed", async () => {
  vi.mocked(loadLocomotiveImage).mockRejectedValueOnce(new Error("localImage"));
  const { unmount } = render(<LocomotiveImageExportDialog image={image} onClose={vi.fn()} />);
  expect(await screen.findByRole("alert")).toHaveTextContent("extern verlinkt");
  expect(screen.queryByRole("link", { name: "Bild herunterladen" })).not.toBeInTheDocument();
  const signal = vi.mocked(loadLocomotiveImage).mock.calls.at(-1)![1];
  unmount(); expect(signal.aborted).toBe(true);
});

it("contains synchronous canvas failures and allows another attempt without leaving the editor", async () => {
  vi.mocked(renderLocomotiveImage).mockImplementationOnce(() => { throw new Error("canvas_lost"); });
  render(<LocomotiveImageExportDialog image={image} onClose={vi.fn()} />);
  expect(await screen.findByRole("alert")).toHaveTextContent("Ausgabedatei konnte nicht erstellt");
  expect(screen.getByRole("button", { name: "Bild herunterladen" })).toBeDisabled();
  fireEvent.click(screen.getByRole("button", { name: "Ganzes Bild" }));
  expect(await screen.findByRole("link", { name: "Bild herunterladen" })).toBeVisible();
});

it("discards stale PNG encodes and keeps focus and Escape inside the nested dialog", async () => {
  let finish: ((blob: Blob) => void) | undefined;
  vi.mocked(renderLocomotiveImage).mockImplementationOnce(() => new Promise(resolve => { finish = resolve; }));
  function Harness() {
    const [open, setOpen] = useState(false);
    return <div role="dialog" aria-label="Fahrzeugdaten bearbeiten"><button type="button" onClick={() => setOpen(true)}>Aufbereiten</button>
      {open && <LocomotiveImageExportDialog image={image} onClose={() => setOpen(false)} />}</div>;
  }
  render(<Harness />);
  const trigger = screen.getByRole("button", { name: "Aufbereiten" }); trigger.focus(); fireEvent.click(trigger);
  await waitFor(() => expect(finish).toBeDefined());
  expect(trigger.closest("[role=dialog]")).toHaveAttribute("inert");
  const dialog = screen.getByRole("dialog", { name: "Lokbild für die Digitalzentrale" });
  fireEvent.click(within(dialog).getByRole("button", { name: "Zielgerät / Anwendung" }));
  fireEvent.click(screen.getByRole("option", { name: "Märklin CS3 / CS3 plus" }));
  const link = await within(dialog).findByRole("link", { name: "Bild herunterladen" });
  const current = link.getAttribute("href");
  await act(async () => { finish!(new Blob(["stale"])); });
  expect(link).toHaveAttribute("href", current!);
  fireEvent.keyDown(dialog, { key: "Escape" });
  expect(screen.queryByRole("dialog", { name: "Lokbild für die Digitalzentrale" })).not.toBeInTheDocument();
  expect(trigger).toHaveFocus(); expect(trigger.closest("[role=dialog]")).not.toHaveAttribute("inert");
});
