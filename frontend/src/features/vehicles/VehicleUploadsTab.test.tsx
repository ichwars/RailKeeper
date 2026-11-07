import { fireEvent, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRef, type ComponentProps } from "react";
import { describe, expect, it, vi } from "vitest";

import type { VehicleAttachment } from "../../shared/api";
import { maintenanceFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { VehicleUploadsTab } from "./VehicleUploadsTab";
import { webDocumentKey } from "./vehicleDocuments";

type Props = ComponentProps<typeof VehicleUploadsTab>;
function props(overrides: Partial<Props> = {}): Props {
  return {
    selected: vehicleFixture(), readonly: false, saving: false, imageInputRef: createRef(), attachmentInputRef: createRef(),
    maintenanceEntries: [], imageUploadMaintenanceID: "", pendingArticleImages: [], attachmentDragActive: false,
    attachmentUploadCategory: "", attachmentUploadDescription: "", attachmentEdits: {}, documentSearchLoading: false,
    documentSearchError: "", documentSearchRan: false, foundDocuments: [], selectedDocuments: {},
    onImageUploadMaintenanceIDChange: vi.fn(), onUploadImages: vi.fn(), onPreviewImage: vi.fn(), onUpdatePendingImageTitle: vi.fn(),
    onUpdatePendingImageMaintenance: vi.fn(), onMovePendingImage: vi.fn(), onSetPrimaryPendingImage: vi.fn(), onRemovePendingImage: vi.fn(),
    onUploadAttachment: vi.fn(), onAttachmentDrag: vi.fn(), onAttachmentDrop: vi.fn(), onAttachmentUploadCategoryChange: vi.fn(),
    onAttachmentUploadDescriptionChange: vi.fn(), onUpdateAttachmentEdit: vi.fn(), onSaveAttachment: vi.fn(), onExtractAttachmentSpareParts: vi.fn(),
    onDeleteAttachment: vi.fn(), onSearchDocuments: vi.fn(), onImportDocument: vi.fn(), onImportSelectedDocuments: vi.fn(),
    onToggleDocument: vi.fn(), onToggleAllDocuments: vi.fn(), ...overrides
  };
}
function attachment(name: string, mimeType = ""): VehicleAttachment {
  return { id: "attachment", vehicleId: "vehicle-1", fileName: name, originalName: name, mimeType, sizeBytes: 4, createdAt: "", updatedAt: "" };
}

describe("Vehicle upload and source document actions", () => {
  it.each([
    ["manual.pdf", "application/pdf", "pdf"], ["picture.PNG", "", "image"],
    ["source.csv", "text/plain", "text"], ["opaque.bin", "application/octet-stream", "unsupported"]
  ])("previews %s without changing its stored file", async (name, mime, kind) => {
    const file = attachment(name, mime);
    const input = props({ selected: vehicleFixture({ attachments: [file] }) });
    render(<VehicleUploadsTab {...input} />);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Anzeigen" }));
    const dialog = screen.getByRole("dialog", { name: `Vorschau ${name}` });
    const frame = dialog.querySelector("iframe");
    if (kind === "text") expect(frame).toHaveAttribute("sandbox", "");
    if (kind === "pdf") expect(frame).toHaveAttribute("src", expect.stringContaining("?inline=true"));
    if (kind === "image") expect(within(dialog).getByRole("img")).toHaveAttribute("alt", name);
    if (kind === "unsupported") expect(within(dialog).getByText("Für diesen Dateityp ist keine Vorschau verfügbar.")).toBeInTheDocument();
    expect(within(dialog).getByRole("link", { name: "Download" })).toHaveAttribute("href", expect.stringContaining("attachment"));
    expect(input.onSaveAttachment).not.toHaveBeenCalled();
    expect(input.onDeleteAttachment).not.toHaveBeenCalled();
    await user.click(within(dialog).getByRole("button", { name: "Schließen" }));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("keeps imported documents read-only and selects only explicitly chosen new sources", async () => {
    const imported = { title: "Vorhanden", url: "https://example.invalid/stored.pdf" };
    const fresh = { title: "Neue Anleitung", url: "https://example.invalid/new.pdf", source: "Hersteller", kind: "Anleitung" };
    const absent = { title: "Ohne Download", url: "" };
    const input = props({
      selected: vehicleFixture({ attachments: [{ ...attachment("stored.pdf"), description: `Quelle: ${imported.url}` }] }),
      foundDocuments: [imported, fresh, absent], selectedDocuments: { [webDocumentKey(fresh, 1)]: true, [webDocumentKey(imported, 0)]: true }
    });
    render(<VehicleUploadsTab {...input} />);
    const user = userEvent.setup();
    expect(screen.getByRole("checkbox", { name: "Vorhanden auswählen" })).toBeDisabled();
    expect(screen.getByRole("checkbox", { name: "Ohne Download auswählen" })).toBeDisabled();
    expect(screen.getByRole("checkbox", { name: "Neue Anleitung auswählen" })).toBeChecked();
    await user.click(screen.getByRole("button", { name: "1 übernehmen" }));
    expect(input.onImportSelectedDocuments).toHaveBeenCalledOnce();
    await user.click(screen.getByRole("checkbox", { name: "Neue Anleitung auswählen" }));
    expect(input.onToggleDocument).toHaveBeenCalledWith(fresh, 1, false);
    await user.click(screen.getByRole("checkbox", { name: "Alle nicht importierten Dokumente auswählen" }));
    expect(input.onToggleAllDocuments).toHaveBeenCalledWith(false);
    const row = within(screen.getByRole("row", { name: /Neue Anleitung/ }));
    await user.click(row.getByRole("button", { name: "Als Beilage übernehmen" }));
    expect(input.onImportDocument).toHaveBeenCalledWith(fresh);
    expect(input.onSaveAttachment).not.toHaveBeenCalled();
  });

  it("passes the exact original selected files and keeps image reordering explicit", async () => {
    const image = { id: "image-a", url: "/image-a.png", source: "", title: "Vorher", isPrimary: true };
    const second = { ...image, id: "image-b", url: "/image-b.png", isPrimary: false };
    const input = props({ pendingArticleImages: [image, second], maintenanceEntries: [maintenanceFixture()] });
    render(<VehicleUploadsTab {...input} />);
    const user = userEvent.setup();
    const file = new File(["synthetic"], "original.png", { type: "image/png" });
    await user.upload(input.imageInputRef.current!, file);
    expect(vi.mocked(input.onUploadImages).mock.calls[0][0]?.item(0)).toBe(file);
    const document = new File(["synthetic"], "manual.txt", { type: "text/plain" });
    await user.upload(input.attachmentInputRef.current!, document);
    expect(vi.mocked(input.onUploadAttachment).mock.calls[0][0]?.item(0)).toBe(document);
    await user.click(screen.getAllByRole("button", { name: "Bild nach unten" })[0]);
    expect(input.onMovePendingImage).toHaveBeenCalledWith("image-a", 1);
    await user.click(screen.getAllByRole("button", { name: "Bild nach oben" })[1]);
    expect(input.onMovePendingImage).toHaveBeenCalledWith("image-b", -1);
    await user.click(screen.getAllByRole("button", { name: "Als Hauptbild markieren" })[1]);
    expect(input.onSetPrimaryPendingImage).toHaveBeenCalledWith("image-b");
    fireEvent.change(screen.getAllByRole("textbox", { name: "Bildbeschreibung" })[0], { target: { value: "Eigener Titel" } });
    expect(input.onUpdatePendingImageTitle).toHaveBeenCalledWith("image-a", "Eigener Titel");
    await user.click(screen.getAllByRole("button", { name: "Bild entfernen" })[1]);
    expect(input.onRemovePendingImage).toHaveBeenCalledWith(second);
    await user.click(screen.getAllByRole("button", { name: "Originalgröße anzeigen" })[0]);
    expect(input.onPreviewImage).toHaveBeenCalledWith(image);
  });

  it("prevents writes for viewers while retaining downloads and previews", async () => {
    const file = attachment("manual.pdf");
    const input = props({ readonly: true, selected: vehicleFixture({ attachments: [file] }) });
    render(<VehicleUploadsTab {...input} />);
    expect(screen.getByRole("button", { name: "Bild hochladen" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Beilage hochladen" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Speichern" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Löschen" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Anzeigen" })).toBeEnabled();
    expect(screen.getByRole("link", { name: "Datei herunterladen" })).toHaveAttribute("href", expect.stringContaining("attachment"));
    await userEvent.setup().click(screen.getByRole("button", { name: "Löschen" }));
    expect(input.onDeleteAttachment).not.toHaveBeenCalled();
  });

  it("shows document lookup failures without enabling an empty import", () => {
    const input = props({ documentSearchRan: true, documentSearchError: "Quelle nicht erreichbar" });
    render(<VehicleUploadsTab {...input} />);
    expect(screen.getByText("Quelle nicht erreichbar")).toBeInTheDocument();
    expect(screen.getByText("Keine Dokumente in den ausgewerteten Seiten gefunden.")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /übernehmen/ })).not.toBeInTheDocument();
  });
});
