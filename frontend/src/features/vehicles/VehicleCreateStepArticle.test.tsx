import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { ArticleSearchResult } from "../../shared/api";
import { VehicleCreateStepArticle } from "./VehicleCreateStepArticle";
import {
  createVehicleCreateWizardState,
  vehicleCreateWizardReducer,
  type VehicleCreateWizardState
} from "./vehicleCreateWizardState";
import type { ArticleSearchController } from "./useArticleSearchController";

const result: ArticleSearchResult = {
  source: "manufacturer", title: "Roco 6280002", url: "https://roco.cc/6280002",
  snippet: "Rekowagenset mit Flicken", score: 435,
  fields: {
    manufacturer: { label: "Hersteller", value: "Roco", confidence: 1 },
    articleNumber: { label: "Artikel-Nr.", value: "6280002", confidence: 1 },
    name: { label: "Bezeichnung", value: "Rekowagenset mit Flicken", confidence: 0.9 }
  },
  conflicts: ["name"]
};

describe("VehicleCreateStepArticle", () => {
  it("keeps results and grouped review embedded in the creation dialog", async () => {
    const user = userEvent.setup();
    let state: VehicleCreateWizardState = {
      ...createVehicleCreateWizardState({ manufacturer: "Roco", articleNumber: "6280002", name: "Set", gauge: "H0" }),
      step: "article"
    };
    const controller = {
      state: { open: false, loading: false, response: { query: "Roco 6280002", results: [result] }, error: "",
        barcodeOpen: false, barcodeValue: "", selectedFields: {}, selectedImages: {} },
      setters: { setOpen: vi.fn(), setBarcodeOpen: vi.fn(), setBarcodeValue: vi.fn() },
      commands: { run: vi.fn(), openBarcode: vi.fn(), submitBarcode: vi.fn(), toggleField: vi.fn(),
        toggleImage: vi.fn(), applyResult: vi.fn() }
    } as unknown as ArticleSearchController;
    function Harness() {
      const dispatch = (action: Parameters<typeof vehicleCreateWizardReducer>[1]) => {
        state = vehicleCreateWizardReducer(state, action);
        rerender(<Harness />);
      };
      return <VehicleCreateStepArticle state={state} dispatch={dispatch} controller={controller}
        onUpdateShared={vi.fn()} />;
    }
    const { rerender } = render(<Harness />);

    await user.click(screen.getByRole("button", { name: /Artikeldaten suchen/ }));
    expect(screen.getByRole("heading", { name: /Suchergebnisse/ })).toBeVisible();
    await user.click(screen.getByRole("button", { name: /Roco 6280002 auswählen/ }));
    expect(screen.getByRole("heading", { name: /Datenübernahme prüfen/ })).toBeVisible();
    expect(screen.queryByRole("dialog", { name: /Artikeldaten-Websuche/ })).toBeNull();
    await user.click(screen.getByRole("button", { name: /Ausgewählte Felder übernehmen/ }));
    expect(controller.commands.applyResult).toHaveBeenCalledWith(result);
  });
});
