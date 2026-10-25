import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { useState } from "react";
import { api, type DigitalCenterSettings } from "../../shared/api";
import { LocomotiveImageExportDialog } from "./LocomotiveImageExportDialog";
import { loadLocomotiveImage, renderLocomotiveImage, analyzeLocomotiveImage } from "./locomotiveImageCanvas";

vi.mock("./locomotiveImageCanvas", () => ({
  loadLocomotiveImage: vi.fn(), renderLocomotiveImage: vi.fn(), analyzeLocomotiveImage: vi.fn()
}));
const bitmap = { width: 1000, height: 200, close: vi.fn() };
const image = { id: "original", url: "/api/v1/images/original", title: "BR 86", source: "local" };
const configuredSettings: DigitalCenterSettings = {
  provider: "cs3",
  ecos: { enabled: false, host: "", port: "15471" },
  cs3: { enabled: true, host: "192.168.1.20", port: "15731" },
  z21: { enabled: false, host: "", port: "21105" },
  intellibox3: { enabled: false, host: "", port: "5550" }
};
const revoke = vi.fn();
let sequence = 0;
beforeEach(() => {
  vi.spyOn(api, "digitalSettings").mockRejectedValue(new Error("not configured"));
  vi.mocked(loadLocomotiveImage).mockResolvedValue(bitmap);
  vi.mocked(analyzeLocomotiveImage).mockReturnValue({ bounds: { x: 0, y: 0, width: 1000, height: 200 }, backgroundMask: null });
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

it("uses the configured station and its automatic background without a device request", async () => {
  vi.mocked(api.digitalSettings).mockResolvedValue(configuredSettings);
  render(<LocomotiveImageExportDialog image={image} onClose={vi.fn()} transfer={{ vehicleId: "vehicle-1" }} />);
  await waitFor(() => expect(screen.getByRole("link", { name: "Bild herunterladen" }))
    .toHaveAttribute("download", "BR-86_cs3.png"));
  expect(screen.getByRole("button", { name: "Hintergrund" })).toHaveTextContent("Automatisch · Transparent");
});

it("keeps a manual choice when the configured station arrives late", async () => {
  let finish: ((settings: DigitalCenterSettings) => void) | undefined;
  vi.mocked(api.digitalSettings).mockImplementationOnce(() => new Promise(resolve => { finish = resolve; }));
  render(<LocomotiveImageExportDialog image={image} onClose={vi.fn()} transfer={{ vehicleId: "vehicle-1" }} />);
  await screen.findByRole("link", { name: "Bild herunterladen" });
  fireEvent.click(screen.getByRole("button", { name: "Hintergrund" }));
  fireEvent.click(screen.getByRole("option", { name: "Schwarz" }));
  await act(async () => { finish!(configuredSettings); });
  expect(screen.getByRole("button", { name: "Hintergrund" })).toHaveTextContent("Schwarz");
  expect(screen.getByRole("button", { name: "Zielgerät / Anwendung" })).toHaveTextContent("ESU ECoS");
});

it("replaces detected photo backgrounds by default and allows retaining the original or changing the colour", async () => {
  const mask = document.createElement("canvas");
  vi.mocked(analyzeLocomotiveImage).mockReturnValueOnce({
    bounds: { x: 0, y: 0, width: 1000, height: 200 }, backgroundMask: mask
  });
  render(<LocomotiveImageExportDialog image={image} onClose={vi.fn()} />);
  await screen.findByRole("link", { name: "Bild herunterladen" });
  expect(vi.mocked(renderLocomotiveImage).mock.calls.at(-1)?.slice(4)).toEqual(["grey", mask]);
  fireEvent.click(screen.getByRole("checkbox", { name: "Gleichmäßigen Fotohintergrund ersetzen" }));
  await waitFor(() => expect(vi.mocked(renderLocomotiveImage).mock.calls.at(-1)?.slice(4)).toEqual(["grey", null]));
  fireEvent.click(screen.getByRole("button", { name: "Hintergrund" }));
  fireEvent.click(screen.getByRole("option", { name: "Weiß" }));
  await waitFor(() => expect(vi.mocked(renderLocomotiveImage).mock.calls.at(-1)?.[4]).toBe("white"));
  expect(api.digitalSettings).not.toHaveBeenCalled();
});

it("enlarges the preview without changing exported pixels and offers a 1:1 view", async () => {
  render(<LocomotiveImageExportDialog image={image} onClose={vi.fn()} />);
  const link = await screen.findByRole("link", { name: "Bild herunterladen" });
  const url = link.getAttribute("href"), calls = vi.mocked(renderLocomotiveImage).mock.calls.length;
  expect(screen.getByRole("region", { name: "Ergebnis" })).toHaveAttribute("data-actual-size", "false");
  fireEvent.click(screen.getByRole("button", { name: "1:1 anzeigen" }));
  expect(screen.getByRole("region", { name: "Ergebnis" })).toHaveAttribute("data-actual-size", "true");
  expect(link).toHaveAttribute("href", url);
  expect(vi.mocked(renderLocomotiveImage).mock.calls.length).toBe(calls);
});

it("shows the output and all targets, protects the source, and updates filename/background together", async () => {
  const before = structuredClone(image);
  const { unmount } = render(<LocomotiveImageExportDialog image={image} onClose={vi.fn()} />);
  expect(await screen.findByRole("link", { name: "Bild herunterladen" })).toHaveAttribute("download", "BR-86_ecos.bmp");
  expect(screen.queryByText("Qualität prüfen")).not.toBeInTheDocument();
  expect(screen.getByRole("button", { name: "Hintergrund" })).toHaveTextContent("Automatisch · Grau");
  expect(screen.getAllByRole("button", { name: "Schließen" })[0]).toHaveFocus();
  expect(screen.queryByRole("option", { name: "Transparent", hidden: true })).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole("button", { name: "Zielgerät / Anwendung" }));
  fireEvent.click(screen.getByRole("option", { name: "Uhlenbrock Intellibox 3 · IB3 Konfiguration" }));
  await waitFor(() => expect(screen.getByRole("link", { name: "Bild herunterladen" })).toHaveAttribute("download", "BR-86_intellibox3.png"));
  expect(screen.getByText(/wandelt die Datei ins Geräteformat/)).toBeInTheDocument();
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
  vi.mocked(analyzeLocomotiveImage).mockReturnValueOnce({ bounds: suggestion, backgroundMask: null });
  render(<LocomotiveImageExportDialog image={image} onClose={vi.fn()} />);
  await screen.findByRole("link", { name: "Bild herunterladen" });
  expect(vi.mocked(renderLocomotiveImage).mock.calls.at(-1)?.[2]).toEqual(suggestion);
  fireEvent.click(screen.getByRole("button", { name: "Ganzes Bild" }));
  await waitFor(() => expect(vi.mocked(renderLocomotiveImage).mock.calls.at(-1)?.[2]).toEqual({ x: 0, y: 0, width: 1000, height: 200 }));
  fireEvent.click(screen.getByRole("button", { name: "Außenränder entfernen" }));
  await waitFor(() => expect(vi.mocked(renderLocomotiveImage).mock.calls.at(-1)?.[2]).toEqual(suggestion));
});

it("prepares small images automatically and blocks fully transparent sources", async () => {
  vi.mocked(loadLocomotiveImage).mockResolvedValueOnce({ ...bitmap, width: 20, height: 5 });
  vi.mocked(analyzeLocomotiveImage).mockReturnValueOnce({ bounds: { x: 0, y: 0, width: 20, height: 5 }, backgroundMask: null });
  const { unmount } = render(<LocomotiveImageExportDialog image={image} onClose={vi.fn()} />);
  await screen.findByRole("link", { name: "Bild herunterladen" });
  expect(screen.queryByText("Qualität prüfen")).not.toBeInTheDocument();
  unmount();
  vi.mocked(analyzeLocomotiveImage).mockReturnValueOnce({ bounds: null, backgroundMask: null });
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
