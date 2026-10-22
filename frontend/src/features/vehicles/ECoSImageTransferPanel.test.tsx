import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { api, ApiError } from "../../shared/api";
import type { ECoSImagePreview, ECoSImageResult } from "../../shared/ecosImageModel";
import { ECoSImageTransferPanel } from "./ECoSImageTransferPanel";

const image = new Blob(["BMP"]);
const preview: ECoSImagePreview = {
  token: "once", expiresAt: "2099-01-01T00:00:00Z", host: "ecos.local",
  target: { objectId: 1001, name: "BR 218", address: 3, protocol: "DCC128" },
  current: { locomotiveType: "LOCO_TYPE_DIESEL", imageType: "IMAGE_TYPE_INT", index: 118 },
  slot: 7, description: "BR 218", sha256: "a".repeat(64), existingImage: false,
  currentImage: "data:image/png;base64,old", newImage: "data:image/png;base64,new"
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
afterEach(() => { vi.restoreAllMocks(); });

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
  expect(screen.getByText(/bereits hochgeladene Bild dort zuordnen/)).toBeVisible();
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

it("explains control refusal and requires a fresh checked preview for assignment without another upload", async () => {
  vi.mocked(api.confirmECoSImage).mockResolvedValueOnce({ ...result, status: "control_busy", assignmentVerified: false });
  render(panel()); await openPreview();
  fireEvent.click(screen.getByRole("checkbox"));
  fireEvent.click(screen.getByRole("button", { name: "Jetzt hochladen und zuordnen" }));
  expect(await screen.findByText(/ECoS verweigert die Loksteuerung/)).toBeVisible();
  expect(screen.getByText(/an beiden ECoS-Fahrreglern/)).toBeVisible();
  expect(api.previewECoSImage).toHaveBeenCalledTimes(1);
  vi.mocked(api.previewECoSImage).mockRejectedValueOnce(new ApiError("changed", "digital_center_write_preview_stale", 409));
  fireEvent.click(screen.getByRole("button", { name: "Nur Zuordnung erneut prüfen …" }));
  await screen.findByRole("alert");
  expect(screen.queryByRole("button", { name: "Übertragung prüfen und Vorschau laden" })).not.toBeInTheDocument();
  vi.mocked(api.previewECoSImage).mockResolvedValueOnce({ ...preview, existingImage: true });
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
