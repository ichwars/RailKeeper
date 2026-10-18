import { createRef } from "react";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { cvValueFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import type { VehicleCVValueInput } from "../../shared/api";
import type { CVFileUploadPreview, CVImportPreview } from "./cvImport";
import { VehicleCVTab } from "./VehicleCVTab";

const cvForm: VehicleCVValueInput = {
  cvNumber: 1,
  value: 3,
  description: "",
  category: "",
  protocol: "",
  decoderProfile: "",
  sourceFileId: ""
};

function renderTab(overrides: Partial<React.ComponentProps<typeof VehicleCVTab>> = {}) {
  const props: React.ComponentProps<typeof VehicleCVTab> = {
    selected: vehicleFixture(),
    ecosDraft: null,
    readonly: false,
    saving: false,
    cvImportInputRef: createRef<HTMLInputElement>(),
    cvFileInputRef: createRef<HTMLInputElement>(),
    cvSummary: { values: 1, profiles: 1, files: 1 },
    cvImportPreview: null,
    cvImportStats: { selected: 0, new: 0, changed: 0, same: 0, invalid: 0 },
    cvForm,
    editingCVID: null,
    decoderProfileOptions: [],
    storedDecoderProfiles: [],
    cvFileProfile: "",
    cvFileDescription: "",
    cvFileUploadPreview: null,
    cvFilePreviewStats: { cvValues: 0, functions: 0 },
    importCVValues: vi.fn(),
    exportCVValues: vi.fn(),
    selectCVImportRows: vi.fn(),
    applyCVImportPreview: vi.fn(),
    discardCVImportPreview: vi.fn(),
    toggleCVImportRow: vi.fn(),
    updateCVForm: vi.fn(),
    resetCVForm: vi.fn(),
    saveCVValue: vi.fn(),
    editCVValue: vi.fn(),
    deleteCVValue: vi.fn(),
    uploadCVFiles: vi.fn(),
    setCVFileProfile: vi.fn(),
    setCVFileDescription: vi.fn(),
    applyFirstCVFileSuggestion: vi.fn(),
    previewCVFileValuesForImport: vi.fn(),
    applyCVFileFunctionSuggestions: vi.fn(),
    confirmCVFileUpload: vi.fn(),
    discardCVFileUploadPreview: vi.fn(),
    deleteCVFile: vi.fn(),
    ...overrides
  };
  render(<VehicleCVTab {...props} />);
  return props;
}

describe("VehicleCVTab", () => {
  beforeEach(() => window.localStorage.clear());

  it("applies selected CV values only after explicit confirmation", () => {
    const existing = cvValueFixture();
    const preview: CVImportPreview = {
      fileName: "decoder.csv",
      rows: [{
        id: "cv-1",
        input: { cvNumber: 1, value: 7 },
        existing,
        status: "changed",
        selected: true,
        message: "Geändert"
      }]
    };
    const applyCVImportPreview = vi.fn();
    const toggleCVImportRow = vi.fn();

    renderTab({
      cvImportPreview: preview,
      cvImportStats: { selected: 1, new: 0, changed: 1, same: 0, invalid: 0 },
      applyCVImportPreview,
      toggleCVImportRow
    });

    expect(screen.getByText("decoder.csv")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Ausgewählte Felder übernehmen" }));
    expect(applyCVImportPreview).toHaveBeenCalledOnce();

    fireEvent.click(screen.getByRole("checkbox", { name: "CV 1 übernehmen" }));
    expect(toggleCVImportRow).toHaveBeenCalledWith("cv-1", false);
  });

  it("opens a source in the side panel without inserting raw details into the list", () => {
    const address = cvValueFixture({ id: "address", cvNumber: 1, value: 3, description: "" });
    const config = cvValueFixture({ id: "config", cvNumber: 29, value: 14, description: "" });
    const props = renderTab({ selected: vehicleFixture({ cvValues: [address, config] }) });
    fireEvent.click(screen.getByRole("button", { name: "CV 1" }));
    const panel = screen.getByRole("dialog");
    expect(within(panel).getByText("0x03")).toBeVisible();
    expect(document.querySelector(".cv-explain-detail")).toBeNull();
    expect(within(panel).queryByText("Bit 0")).not.toBeInTheDocument();
    fireEvent.click(within(panel).getByRole("button", { name: "CV-Wert bearbeiten" }));
    expect(props.editCVValue).toHaveBeenCalledExactlyOnceWith(address);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Alle CV-Werte" })).toHaveAttribute("aria-pressed", "true");
  });

  it("confirms a decoder file from the upload preview", () => {
    const file = new File(["decoder"], "decoder.esux", { type: "application/octet-stream" });
    const uploadPreview: CVFileUploadPreview = {
      files: [file],
      previews: [{
        fileName: file.name,
        sizeBytes: file.size,
        mimeType: file.type,
        hasMetadata: true,
        projectName: "BR 106",
        decoder: "LokPilot 5"
      }]
    };
    const confirmCVFileUpload = vi.fn();

    renderTab({ cvFileUploadPreview: uploadPreview, confirmCVFileUpload });

    expect(screen.getByText("Upload-Vorschau")).toBeInTheDocument();
    expect(screen.getByText("BR 106")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Dateien speichern" }));
    expect(confirmCVFileUpload).toHaveBeenCalledOnce();
  });

  it("keeps write actions disabled in read-only mode", () => {
    renderTab({ readonly: true });

    expect(screen.getByRole("button", { name: "CV-Liste importieren" })).toBeDisabled();
    fireEvent.click(screen.getByRole("button", { name: "Alle CV-Werte" }));
    fireEvent.click(screen.getByText("Einzelnen CV-Wert hinzufügen oder bearbeiten"));
    expect(screen.getByRole("button", { name: "CV hinzufügen" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Decoderdatei hochladen" })).toBeDisabled();
  });
});
