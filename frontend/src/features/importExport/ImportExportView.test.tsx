import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { ImportExportView } from "./ImportExportView";
import { defaultColumnMappings, parseDelimited, vehicleImportFields } from "./importExportHelpers";

describe("ImportExportView", () => {
  beforeEach(() => {
    window.localStorage.clear();
    window.sessionStorage.clear();
    vi.spyOn(api, "vehicles").mockResolvedValue([]);
    vi.spyOn(api, "masterDataAll").mockResolvedValue({});
    vi.spyOn(api, "accessoryArticles").mockResolvedValue({
      items: [],
      metrics: { articleCount: 0, articleTypeCount: 0, available: 0, locationCount: 0, reserved: 0, installed: 0, careHintCount: 0 },
      filters: { manufacturers: [], articleTypes: [], gauges: [], storageLocations: [] }
    });
    vi.spyOn(api, "exhibitionLists").mockResolvedValue([]);
    vi.spyOn(window, "confirm").mockReturnValue(true);
  });

  it("does not expose master data in the main transfer workspace", async () => {
    render(<ImportExportView />);

    expect(await screen.findByRole("heading", { name: "Import/Export" })).toBeInTheDocument();
    expect(screen.queryByText("Stammdaten", { selector: "button, label, td" })).not.toBeInTheDocument();
    expect(api.masterDataAll).not.toHaveBeenCalled();
  });

  it("previews a mapped CSV and saves the selected vehicle", async () => {
    const saved = vehicleFixture({ id: "vehicle-imported", inventoryNumber: "RK-LOK-000003" });
    vi.spyOn(api, "createVehicle").mockResolvedValue(saved);
    const file = new File([
      "Inventarnummer;Hersteller;Bezeichnung;Spurweite;Kategorie;Gattung;Digital\n" +
        "RK-LOK-000003;ESU;BR 106;H0;Lokomotive;Diesellokomotive;ja"
    ], "bestand.csv", { type: "text/csv" });
    Object.defineProperty(file, "text", { value: () => Promise.resolve(fileContent()) });

    const { container } = render(<ImportExportView />);
    const input = container.querySelector<HTMLInputElement>('input[type="file"]');
    expect(input).not.toBeNull();
    fireEvent.change(input!, { target: { files: [file] } });

    expect(await screen.findByText("Spaltenzuordnung")).toBeInTheDocument();
    expect(screen.getByText("Importprüfung")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Auswahl speichern" }));

    await waitFor(() => expect(api.createVehicle).toHaveBeenCalledOnce());
    expect(api.createVehicle).toHaveBeenCalledWith(expect.objectContaining({
      inventoryNumber: "RK-LOK-000003",
      manufacturer: "ESU",
      name: "BR 106",
      digital: true
    }));
    expect(await screen.findByText("gespeichert")).toBeInTheDocument();
  });

  it("preserves operational fields when importing a RailKeeper JSON export", async () => {
    const vehicle = vehicleFixture({
      id: "vehicle-json-import",
      inventoryNumber: "RK-LOK-000004",
      maximumSpeedKmh: 120,
      homeBase: "Bw Leipzig-West"
    });
    vi.spyOn(api, "createVehicle").mockResolvedValue(vehicle);
    const content = JSON.stringify({ format: "railkeeper-vehicles", version: 1, vehicles: [vehicle] });
    const file = new File([content], "railkeeper-bestand.json", { type: "application/json" });
    Object.defineProperty(file, "text", { value: () => Promise.resolve(content) });

    const { container } = render(<ImportExportView />);
    const input = container.querySelector<HTMLInputElement>('input[type="file"]');
    expect(input).not.toBeNull();
    fireEvent.change(input!, { target: { files: [file] } });

    expect(await screen.findByText("Spaltenzuordnung")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Auswahl speichern" }));

    await waitFor(() => expect(api.createVehicle).toHaveBeenCalledOnce());
    expect(api.createVehicle).toHaveBeenCalledWith(expect.objectContaining({
      maximumSpeedKmh: 120,
      homeBase: "Bw Leipzig-West"
    }));
  });

  it("exports unambiguous CSV headers for ownership and storage fields", async () => {
    vi.mocked(api.vehicles).mockResolvedValue([vehicleFixture({
      acquisitionType: "Gebrauchtkauf",
      purchasePrice: "149.90",
      storageDetails: "Vitrine, Fach 3",
      conditionDetails: "Leichte Laufspuren"
    })]);
    const createObjectURL = vi.spyOn(URL, "createObjectURL").mockReturnValue("blob:railkeeper-csv");
    vi.spyOn(URL, "revokeObjectURL").mockImplementation(() => undefined);
    vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => undefined);

    render(<ImportExportView />);
    const exportButton = await screen.findByRole("button", { name: "Export erstellen" });
    await waitFor(() => expect(exportButton).toBeEnabled());
    fireEvent.click(exportButton);

    const blob = createObjectURL.mock.calls[0][0];
    if (!(blob instanceof Blob)) {
      throw new Error("CSV export did not create a Blob");
    }
    const csv = await blob.text();
    expect(csv).toContain("Erwerbsart");
    expect(csv).toContain("Kaufpreis");
    expect(csv).toContain("Lagerdetails");
    expect(csv).toContain("Zustandsdetails");
    expect(csv).not.toContain(";Details;Details;");
    expect(defaultColumnMappings(parseDelimited(csv, ";")).map((mapping) => mapping.key)).toEqual(
      vehicleImportFields.map((field) => field.key)
    );
  });
});

function fileContent() {
  return "Inventarnummer;Hersteller;Bezeichnung;Spurweite;Kategorie;Gattung;Digital\n" +
    "RK-LOK-000003;ESU;BR 106;H0;Lokomotive;Diesellokomotive;ja";
}
