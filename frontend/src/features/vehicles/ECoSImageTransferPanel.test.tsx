import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { api, ApiError } from "../../shared/api";
import type { ECoSImagePreview, ECoSImageResult } from "../../shared/ecosImageModel";
import { ECoSImageTransferPanel } from "./ECoSImageTransferPanel";

const image = new Blob(["BMP"]);
const preview: ECoSImagePreview = {
  token: "once", expiresAt: new Date(Date.now() + 300000).toISOString(), host: "ecos.local",
  target: { objectId: 1001, name: "BR 218", address: 3, protocol: "DCC128" },
  current: { locomotiveType: "LOCO_TYPE_DIESEL", imageType: "IMAGE_TYPE_INT", index: 118 },
  slot: 7, description: "BR 218", sha256: "a".repeat(64), existingImage: false,
  currentImage: "data:image/png;base64,old", currentImageMissing: false, newImage: "data:image/png;base64,new"
};
const result: ECoSImageResult = {
  status: "verified", slot: 7, objectId: 1001, imageVerified: true, assignmentVerified: true,
  auditRecorded: true, liveMonitor: { wasRunning: false, restarted: false }
};
const busyChange = vi.fn();
beforeEach(() => {
  vi.spyOn(api, "ecosImageOptions").mockResolvedValue({ host: "ecos.local", locomotives: [preview.target] });
  vi.spyOn(api, "previewECoSImage").mockResolvedValue({ ...preview, expiresAt: new Date(Date.now() + 300000).toISOString() });
  vi.spyOn(api, "confirmECoSImage").mockResolvedValue(result);
  busyChange.mockClear();
});
afterEach(() => { vi.useRealTimers(); vi.restoreAllMocks(); });

function panel(mappedObjectId = "1001") {
  return <ECoSImageTransferPanel vehicleId="vehicle-1" mappedObjectId={mappedObjectId}
    title="BR 218" blob={image} onBusyChange={busyChange} />;
}
async function openPreview() {
  fireEvent.click(screen.getByRole("button", { name: "Direkt an ECoS übertragen …" }));
  fireEvent.click(await screen.findByRole("button", { name: "Übertragung prüfen und Vorschau laden" }));
  await screen.findByAltText("Neues Lokbild · freier Bildplatz 7");
}

it("reads only until the target and before/after preview are explicitly confirmed, then writes once", async () => {
  render(panel());
  expect(api.ecosImageOptions).not.toHaveBeenCalled();
  await openPreview();
  expect(api.previewECoSImage).toHaveBeenCalledWith("vehicle-1", 1001, "BR 218", image);
  expect(api.confirmECoSImage).not.toHaveBeenCalled();
  expect(screen.getByAltText("Bisheriges Lokbild")).toBeVisible();
  const confirm = screen.getByRole("button", { name: "Jetzt hochladen und zuordnen" });
  expect(confirm).toBeDisabled();
  fireEvent.click(screen.getByRole("checkbox", { name: "Ziel-Lok und Bildvorschau geprüft" }));
  fireEvent.click(confirm);
  expect(await screen.findByText(/Bild auf Platz 7 geprüft und Lok #1001 zugeordnet/)).toBeVisible();
  expect(api.confirmECoSImage).toHaveBeenCalledExactlyOnceWith("vehicle-1", "once");
  expect(busyChange.mock.calls).toEqual([[true], [false]]);
  expect(screen.queryByRole("button", { name: "Jetzt hochladen und zuordnen" })).not.toBeInTheDocument();
});

it("requires a deliberate target choice when no saved ECoS mapping exists", async () => {
  render(panel(""));
  fireEvent.click(screen.getByRole("button", { name: "Direkt an ECoS übertragen …" }));
  const prepare = await screen.findByRole("button", { name: "Übertragung prüfen und Vorschau laden" });
  expect(prepare).toBeDisabled();
  fireEvent.click(screen.getByRole("button", { name: "Ziel-Lok auf der ECoS" }));
  fireEvent.click(screen.getByRole("option", { name: "BR 218 · 3 · DCC128 · #1001" }));
  expect(prepare).toBeEnabled();
  expect(api.previewECoSImage).not.toHaveBeenCalled();
});

it("invalidates confirmation when the description changes and ignores an unmounted preview response", async () => {
  const { unmount } = render(panel()); await openPreview();
  fireEvent.click(screen.getByRole("checkbox"));
  fireEvent.change(screen.getByLabelText("Bildbezeichnung auf der ECoS"), { target: { value: "BR 218 neu" } });
  expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
  let finish: ((value: ECoSImagePreview) => void) | undefined;
  vi.mocked(api.previewECoSImage).mockImplementationOnce(() => new Promise(resolve => { finish = resolve; }));
  fireEvent.click(screen.getByRole("button", { name: "Übertragung prüfen und Vorschau laden" }));
  unmount();
  await act(async () => { finish?.(preview); });
  expect(api.confirmECoSImage).not.toHaveBeenCalled();
});

it("expires the preview and blocks its confirmation", async () => {
  vi.mocked(api.previewECoSImage).mockResolvedValueOnce({ ...preview, expiresAt: new Date(Date.now() - 1000).toISOString() });
  render(panel()); await openPreview();
  fireEvent.click(screen.getByRole("checkbox"));
  await waitFor(() => expect(screen.getByRole("button", { name: "Jetzt hochladen und zuordnen" })).toBeDisabled());
  expect(screen.getByText(/Die Vorschau ist abgelaufen/)).toBeVisible();
});

it("keeps partial and ambiguous writes visible without offering an automatic retry", async () => {
  vi.mocked(api.confirmECoSImage).mockResolvedValueOnce({ ...result, status: "assignment_unknown", assignmentVerified: false });
  const { unmount } = render(panel()); await openPreview();
  fireEvent.click(screen.getByRole("checkbox")); fireEvent.click(screen.getByRole("button", { name: "Jetzt hochladen und zuordnen" }));
  expect(await screen.findByText(/Die Zuordnung zu Lok #1001 ist nicht bestätigt/)).toBeVisible();
  expect(screen.getByText(/ohne erneuten Upload/)).toBeVisible();
  expect(screen.queryByRole("button", { name: "Übertragung prüfen und Vorschau laden" })).not.toBeInTheDocument();
  unmount(); vi.mocked(api.confirmECoSImage).mockRejectedValueOnce(new Error("reply lost"));
  render(panel()); await openPreview(); fireEvent.click(screen.getByRole("checkbox"));
  fireEvent.click(screen.getByRole("button", { name: "Jetzt hochladen und zuordnen" }));
  expect(await screen.findByRole("alert")).toHaveTextContent("Die Antwort fehlt");
  expect(api.confirmECoSImage).toHaveBeenCalledTimes(2);
});

it("explains HTTP firewall failures", async () => {
  vi.mocked(api.previewECoSImage).mockRejectedValueOnce(new ApiError("gateway", "ecos_image_web", 502));
  render(panel());
  fireEvent.click(screen.getByRole("button", { name: "Direkt an ECoS übertragen …" }));
  fireEvent.click(await screen.findByRole("button", { name: "Übertragung prüfen und Vorschau laden" }));
  expect(await screen.findByRole("alert")).toHaveTextContent("Port 80");
});

it.each(["ecos_image_read", "ecos_image_inventory", "ecos_image_form"])("distinguishes %s from a failed HTTP connection", async code => {
  vi.mocked(api.previewECoSImage).mockRejectedValueOnce(new ApiError("image preparation failed",code,502));
  render(panel());
  fireEvent.click(screen.getByRole("button",{name:"Direkt an ECoS übertragen …"}));
  fireEvent.click(await screen.findByRole("button",{name:"Übertragung prüfen und Vorschau laden"}));
  expect(await screen.findByRole("alert")).toHaveTextContent("Es wurde nichts hochgeladen oder zugeordnet");
  expect(screen.getByRole("alert")).not.toHaveTextContent("Port 80");
  expect(api.confirmECoSImage).not.toHaveBeenCalled();
});

it("explains a deleted current user image and permits reviewed replacement without rendering a broken image", async () => {
  vi.mocked(api.previewECoSImage).mockResolvedValueOnce({...preview,currentImage:"",currentImageMissing:true,
    current:{...preview.current,imageType:"IMAGE_TYPE_USER",index:3}});
  render(panel());await openPreview();
  expect(screen.getByText(/Der Lok ist Bildplatz 3 zugeordnet, dieser ist aber leer/)).toBeVisible();
  expect(screen.queryByAltText("Bisheriges Lokbild")).not.toBeInTheDocument();
  expect(screen.getByRole("button",{name:"Jetzt hochladen und zuordnen"})).toBeDisabled();
  expect(api.confirmECoSImage).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("checkbox"));fireEvent.click(screen.getByRole("button",{name:"Jetzt hochladen und zuordnen"}));
  expect(await screen.findByText(/Die Zuordnung wurde zurückgelesen und geprüft/)).toBeVisible();
});

it("explains a slot still reported free without claiming the image was stored or retrying", async () => {
  vi.mocked(api.confirmECoSImage).mockResolvedValueOnce({
    ...result, status: "image_not_stored", imageVerified: false, assignmentVerified: false
  });
  render(panel()); await openPreview();
  fireEvent.click(screen.getByRole("checkbox"));
  fireEvent.click(screen.getByRole("button", { name: "Jetzt hochladen und zuordnen" }));
  expect(await screen.findByText(/Bildplatz 7 nach dem Upload weiterhin als frei/)).toBeVisible();
  expect(screen.getByText(/Ein weiterer Upload benötigt eine neue Vorschau und Bestätigung/)).toBeVisible();
  expect(screen.queryByText(/bereits hochgeladene Bild dort zuordnen/)).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "Jetzt hochladen und zuordnen" })).not.toBeInTheDocument();
  expect(api.confirmECoSImage).toHaveBeenCalledExactlyOnceWith("vehicle-1", "once");
});

it("blocks transfer when the station has no target locomotives", async () => {
  vi.mocked(api.ecosImageOptions).mockResolvedValueOnce({ host: "ecos.local", locomotives: [] });
  render(panel()); fireEvent.click(screen.getByRole("button", { name: "Direkt an ECoS übertragen …" }));
  expect(await screen.findByText(/Die ECoS hat keine Loks gemeldet/)).toBeVisible();
  expect(screen.getByRole("button", { name: "Übertragung prüfen und Vorschau laden" })).toBeDisabled();
  expect(api.previewECoSImage).not.toHaveBeenCalled(); expect(api.confirmECoSImage).not.toHaveBeenCalled();
});

it("requires a fresh checked preview for a verified partial assignment without another upload", async () => {
  vi.mocked(api.confirmECoSImage).mockResolvedValueOnce({ ...result, status: "assignment_unknown", assignmentVerified: false });
  render(panel()); await openPreview();
  fireEvent.click(screen.getByRole("checkbox"));
  fireEvent.click(screen.getByRole("button", { name: "Jetzt hochladen und zuordnen" }));
  expect(await screen.findByText(/Die Zuordnung zu Lok #1001 ist nicht bestätigt/)).toBeVisible();
  expect(api.previewECoSImage).toHaveBeenCalledTimes(1);
  vi.mocked(api.previewECoSImage).mockRejectedValueOnce(new ApiError("changed", "digital_center_write_preview_stale", 409));
  fireEvent.click(screen.getByRole("button", { name: "Nur Zuordnung erneut prüfen …" }));
  await screen.findByRole("alert");
  expect(screen.queryByRole("button", { name: "Übertragung prüfen und Vorschau laden" })).not.toBeInTheDocument();
  vi.mocked(api.previewECoSImage).mockResolvedValueOnce({ ...preview, token: "retry-assignment", existingImage: true });
  fireEvent.click(screen.getByRole("button", { name: "Nur Zuordnung erneut prüfen …" }));
  await screen.findByAltText("Bereits geprüftes Bild · Bildplatz 7");
  expect(api.previewECoSImage).toHaveBeenLastCalledWith("vehicle-1", 1001, "BR 218", image, 7);
  const assign = screen.getByRole("button", { name: "Vorhandenes Bild jetzt zuordnen" });
  expect(assign).toBeDisabled();
  expect(api.confirmECoSImage).toHaveBeenCalledTimes(1);
  expect(screen.getByText(/Es wird kein Bild hochgeladen/)).toBeVisible();
  fireEvent.click(screen.getByRole("checkbox")); fireEvent.click(assign);
  expect(await screen.findByText(/Die Zuordnung wurde zurückgelesen und geprüft/)).toBeVisible();
  expect(api.confirmECoSImage).toHaveBeenCalledTimes(2);
});

it.each(["upload_blocked"] as const)("explains %s before upload and requires another preview and confirmation", async status => {
  vi.mocked(api.confirmECoSImage).mockResolvedValueOnce({ ...result, status, imageVerified: false, assignmentVerified: false });
  render(panel()); await openPreview();
  fireEvent.click(screen.getByRole("checkbox")); fireEvent.click(screen.getByRole("button", { name: "Jetzt hochladen und zuordnen" }));
  expect(await screen.findByText(/Es wurde kein Bild hochgeladen und keine Bildzuordnung geschrieben/)).toBeVisible();
  expect(screen.queryByRole("button", { name: "Nur Zuordnung erneut prüfen …" })).not.toBeInTheDocument();
  expect(api.confirmECoSImage).toHaveBeenCalledTimes(1);
  fireEvent.click(screen.getByRole("button", { name: "Übertragung prüfen und Vorschau laden" }));
  await screen.findByAltText("Neues Lokbild · freier Bildplatz 7");
  expect(api.previewECoSImage).toHaveBeenLastCalledWith("vehicle-1", 1001, "BR 218", image);
  expect(screen.getByRole("button", { name: "Jetzt hochladen und zuordnen" })).toBeDisabled();
});

async function pauseTransfer() {
  vi.mocked(api.confirmECoSImage).mockResolvedValueOnce({
    ...result, status: "control_busy", imageVerified: false, assignmentVerified: false
  });
  await openPreview(); fireEvent.click(screen.getByRole("checkbox"));
  fireEvent.click(screen.getByRole("button", { name: "Jetzt hochladen und zuordnen" }));
  return screen.findByRole("dialog", { name: "Übertragung pausiert" });
}
const refreshed = (token = "resumed"): ECoSImagePreview => ({
  ...preview, token, expiresAt: new Date(Date.now() + 300000).toISOString()
});

it("pauses without polling and continues only after a click and a fresh unchanged review", async () => {
  render(panel());
  const dialog = await pauseTransfer();
  expect(dialog).toHaveTextContent("an beiden Fahrreglern");
  expect(dialog).toHaveTextContent("Bild und Ziel-Lok bleiben erhalten");
  expect(dialog).toHaveTextContent("BR 218");
  expect(screen.getByAltText("Vorgesehenes Lokbild")).toBeVisible();
  expect(busyChange.mock.calls).toEqual([[true]]);
  vi.useFakeTimers(); await act(async () => { vi.advanceTimersByTime(600000); }); vi.useRealTimers();
  expect(api.confirmECoSImage).toHaveBeenCalledTimes(1);
  expect(api.previewECoSImage).toHaveBeenCalledTimes(1);
  vi.mocked(api.previewECoSImage).mockResolvedValueOnce(refreshed());
  fireEvent.click(screen.getByRole("button", { name: "Fortsetzen" }));
  expect(await screen.findByText(/Die Zuordnung wurde zurückgelesen und geprüft/)).toBeVisible();
  expect(api.previewECoSImage).toHaveBeenLastCalledWith("vehicle-1", 1001, "BR 218", image);
  expect(vi.mocked(api.confirmECoSImage).mock.calls).toEqual([["vehicle-1", "once"], ["vehicle-1", "resumed"]]);
  expect(busyChange).toHaveBeenLastCalledWith(false);
});

it("cancels with Escape, restores parent focus and keeps image and target without another write", async () => {
  const parentKey = vi.fn();
  render(<div role="dialog" aria-label="Bildbearbeitung" onKeyDown={parentKey}>{panel()}</div>);
  const dialog = await pauseTransfer();
  const parent = screen.getByLabelText("Bildbearbeitung");
  await waitFor(() => expect(parent).toHaveAttribute("inert"));
  fireEvent.keyDown(dialog, { key: "Escape" });
  await waitFor(() => {
    expect(screen.queryByRole("dialog", { name: "Übertragung pausiert" })).not.toBeInTheDocument();
    expect(parent).not.toHaveAttribute("inert");
    expect(screen.getByRole("button", { name: "Übertragung prüfen und Vorschau laden" })).toHaveFocus();
  });
  expect(parentKey).not.toHaveBeenCalled();
  expect(screen.getByLabelText("Bildbezeichnung auf der ECoS")).toHaveValue("BR 218");
  expect(api.confirmECoSImage).toHaveBeenCalledTimes(1);
  expect(busyChange).toHaveBeenLastCalledWith(false);
});

it.each([
  { target: { ...preview.target, objectId: 1002 } },
  { target: { ...preview.target, name: "Andere Lok" } },
  { target: { ...preview.target, address: 4 } },
  { target: { ...preview.target, protocol: "DCC28" } },
  { current: { ...preview.current, index: 119 } },
  { current: { ...preview.current, locomotiveType: "LOCO_TYPE_STEAM" } },
  { current: { ...preview.current, imageType: "IMAGE_TYPE_USER" as const } },
  { currentImage: "data:image/png;base64,changed" },
  { currentImageMissing: true, currentImage: "" },
  { newImage: "data:image/png;base64,changed" },
  { sha256: "b".repeat(64) },
  { host: "other.local" },
  { description: "Anderes Bild" },
  { slot: 8 },
  { existingImage: true },
  { expiresAt: "2000-01-01T00:00:00Z" }
])("requires a new explicit review when pause state changes: %j", async changes => {
  render(panel()); await pauseTransfer();
  vi.mocked(api.previewECoSImage).mockResolvedValueOnce({ ...refreshed(), ...changes });
  fireEvent.click(screen.getByRole("button", { name: "Fortsetzen" }));
  expect(await screen.findByText(/Die Vorschau hat sich während der Pause geändert/)).toBeVisible();
  expect(screen.queryByRole("dialog", { name: "Übertragung pausiert" })).not.toBeInTheDocument();
  expect(screen.getByRole("checkbox")).not.toBeChecked();
  expect(api.confirmECoSImage).toHaveBeenCalledTimes(1);
});

it("retains the pause after a failed read and never resubmits a consumed token", async () => {
  render(panel()); await pauseTransfer();
  vi.mocked(api.previewECoSImage).mockRejectedValueOnce(new ApiError("read", "ecos_image_inventory", 502));
  fireEvent.click(screen.getByRole("button", { name: "Fortsetzen" }));
  expect(await screen.findByRole("alert")).toHaveTextContent("Der Bildbestand");
  expect(screen.getByRole("dialog", { name: "Übertragung pausiert" })).toBeVisible();
  vi.mocked(api.previewECoSImage).mockResolvedValueOnce(refreshed("once"));
  fireEvent.click(screen.getByRole("button", { name: "Fortsetzen" }));
  expect(await screen.findByRole("alert")).toHaveTextContent("bereits verwendet");
  expect(api.confirmECoSImage).toHaveBeenCalledTimes(1);
  fireEvent.click(screen.getByRole("button", { name: "Übertragung abbrechen" }));
  expect(screen.queryByRole("dialog", { name: "Übertragung pausiert" })).not.toBeInTheDocument();
});

it("blocks double clicks and cancellation during the fresh check and returns to pause when control stays busy", async () => {
  render(panel()); await pauseTransfer();
  let finish: ((value: ECoSImagePreview) => void) | undefined;
  vi.mocked(api.previewECoSImage).mockImplementationOnce(() => new Promise(resolve => { finish = resolve; }));
  vi.mocked(api.confirmECoSImage).mockResolvedValueOnce({ ...result, status: "control_busy", imageVerified: false, assignmentVerified: false });
  const button = screen.getByRole("button", { name: "Fortsetzen" });
  fireEvent.click(button); fireEvent.click(button);
  expect(button).toBeDisabled();
  expect(screen.getByRole("button", { name: "Übertragung abbrechen" })).toBeDisabled();
  fireEvent.keyDown(screen.getByRole("dialog", { name: "Übertragung pausiert" }), { key: "Escape" });
  expect(api.previewECoSImage).toHaveBeenCalledTimes(2);
  await act(async () => { finish?.(refreshed()); });
  expect(screen.getByRole("dialog", { name: "Übertragung pausiert" })).toBeVisible();
  expect(api.confirmECoSImage).toHaveBeenLastCalledWith("vehicle-1", "resumed");
  expect(busyChange).not.toHaveBeenCalledWith(false);
  vi.mocked(api.previewECoSImage).mockResolvedValueOnce(refreshed("once"));
  fireEvent.click(screen.getByRole("button", { name: "Fortsetzen" }));
  expect(await screen.findByRole("alert")).toHaveTextContent("bereits verwendet");
  expect(api.confirmECoSImage).toHaveBeenCalledTimes(2);
});

it("does not continue writing when the editor unmounts during the fresh preview", async () => {
  const { unmount } = render(panel()); await pauseTransfer();
  let finish: ((value: ECoSImagePreview) => void) | undefined;
  vi.mocked(api.previewECoSImage).mockImplementationOnce(() => new Promise(resolve => { finish = resolve; }));
  fireEvent.click(screen.getByRole("button", { name: "Fortsetzen" }));
  unmount();
  await act(async () => { finish?.(refreshed()); });
  expect(api.confirmECoSImage).toHaveBeenCalledTimes(1);
});

it("keeps assignment-only retries from uploading and never retries a lost continuation reply", async () => {
  vi.mocked(api.previewECoSImage).mockResolvedValueOnce({ ...refreshed("once"), existingImage: true });
  vi.mocked(api.confirmECoSImage).mockResolvedValueOnce({ ...result, status: "control_busy", assignmentVerified: false });
  render(panel());
  fireEvent.click(screen.getByRole("button", { name: "Direkt an ECoS übertragen …" }));
  fireEvent.click(await screen.findByRole("button", { name: "Übertragung prüfen und Vorschau laden" }));
  await screen.findByAltText("Bereits geprüftes Bild · Bildplatz 7");
  fireEvent.click(screen.getByRole("checkbox"));
  fireEvent.click(screen.getByRole("button", { name: "Vorhandenes Bild jetzt zuordnen" }));
  await screen.findByRole("dialog", { name: "Übertragung pausiert" });
  vi.mocked(api.previewECoSImage).mockResolvedValueOnce({ ...refreshed(), existingImage: true });
  vi.mocked(api.confirmECoSImage).mockRejectedValueOnce(new Error("reply lost"));
  fireEvent.click(screen.getByRole("button", { name: "Fortsetzen" }));
  expect(await screen.findByRole("alert")).toHaveTextContent("Die Antwort fehlt");
  expect(api.previewECoSImage).toHaveBeenLastCalledWith("vehicle-1", 1001, "BR 218", image, 7);
  expect(screen.queryByRole("dialog", { name: "Übertragung pausiert" })).not.toBeInTheDocument();
  expect(api.confirmECoSImage).toHaveBeenCalledTimes(2);
});

it("uses an automatically detected identical image in a new dialog without offering another upload", async () => {
  vi.mocked(api.previewECoSImage).mockResolvedValueOnce({ ...preview, existingImage: true });
  render(panel());
  fireEvent.click(screen.getByRole("button", { name: "Direkt an ECoS übertragen …" }));
  fireEvent.click(await screen.findByRole("button", { name: "Übertragung prüfen und Vorschau laden" }));
  await screen.findByAltText("Bereits geprüftes Bild · Bildplatz 7");
  expect(screen.getByText(/Die Bildbezeichnung auf der ECoS bleibt erhalten/)).toBeVisible();
  expect(screen.queryByRole("button", { name: "Jetzt hochladen und zuordnen" })).not.toBeInTheDocument();
  expect(screen.getByRole("button", { name: "Vorhandenes Bild jetzt zuordnen" })).toBeDisabled();
  expect(api.confirmECoSImage).not.toHaveBeenCalled();
});
