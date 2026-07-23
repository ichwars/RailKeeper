import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type DigitalCenterSettings } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { ImportExportView } from "./ImportExportView";

const disabledDigitalSettings: DigitalCenterSettings = {
  provider: "ecos",
  ecos: { enabled: false, host: "", port: "15471" },
  z21: { enabled: false, host: "", port: "21105" },
  intellibox3: { enabled: false, host: "", port: "21105" },
  cs3: { enabled: false, host: "", port: "80" }
};

describe("ImportExportView", () => {
  beforeEach(() => {
    window.localStorage.clear();
    window.sessionStorage.clear();
    vi.spyOn(api, "vehicles").mockResolvedValue([]);
    vi.spyOn(api, "masterDataAll").mockResolvedValue({});
    vi.spyOn(api, "digitalSettings").mockResolvedValue(disabledDigitalSettings);
    vi.spyOn(api, "getECoSLiveStatus").mockResolvedValue({
      provider: "ecos",
      connected: false,
      blocksReceived: 0,
      repliesReceived: 0,
      eventsReceived: 0,
      message: "Nicht verbunden"
    });
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
});

function fileContent() {
  return "Inventarnummer;Hersteller;Bezeichnung;Spurweite;Kategorie;Gattung;Digital\n" +
    "RK-LOK-000003;ESU;BR 106;H0;Lokomotive;Diesellokomotive;ja";
}
