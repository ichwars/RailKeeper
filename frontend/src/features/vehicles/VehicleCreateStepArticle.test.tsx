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
const resultWithImage: ArticleSearchResult = {
  ...result,
  images: [{ url: "https://roco.cc/set.jpg", title: "Set", source: "manufacturer" }]
};
const resultWithExtendedFields: ArticleSearchResult = {
  ...result,
  fields: {
    ...result.fields,
    driveDescription: { label: "Antrieb Beschreibung", value: "Mittelmotor", confidence: 0.8 },
    soundGeneratorEnabled: { label: "Soundgenerator", value: "Ja", confidence: 0.8 },
    wheelset: { label: "Radsatz", value: "AC", confidence: 0.8 },
    couplingFront: { label: "Kupplung vorne", value: "Kurzkupplung", confidence: 0.8 }
  }
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
    await user.click(screen.getByRole("button", { name: /Suche ändern/ }));
    expect(screen.getByRole("heading", { name: /Artikeldaten finden/ })).toBeVisible();
    await user.click(screen.getByRole("button", { name: /Artikeldaten suchen/ }));
    expect(document.querySelector(".vehicle-create-result-card.without-image .vehicle-create-result-copy"))
      .toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /Roco 6280002 auswählen/ }));
    expect(screen.getByRole("heading", { name: /Datenübernahme prüfen/ })).toBeVisible();
    expect(screen.getByText("Gefundener Wert")).toBeVisible();
    expect(screen.getByText("Übernehmen")).toBeVisible();
    expect(screen.getByText(/Felder gefunden/)).toBeVisible();
    expect(screen.getByText(/ausgewählt/)).toBeVisible();
    expect(screen.queryByText("Aktuell")).toBeNull();
    expect(screen.queryByRole("dialog", { name: /Artikeldaten-Websuche/ })).toBeNull();
    expect(screen.queryByRole("button", { name: /Ausgewählte Felder übernehmen/ })).toBeNull();
  });

  it("lets set images be assigned to a specific member", async () => {
    const user = userEvent.setup();
    let state: VehicleCreateWizardState = {
      ...createVehicleCreateWizardState({ manufacturer: "Roco", articleNumber: "6280002", name: "Set", gauge: "H0" }),
      kind: "set",
      step: "article",
      articleStage: "review",
      selectedResultIndex: 0
    };
    const controller = {
      state: { open: false, loading: false, response: { query: "Roco 6280002", results: [resultWithImage] }, error: "",
        barcodeOpen: false, barcodeValue: "", selectedFields: {}, selectedImages: {} },
      setters: { setOpen: vi.fn(), setBarcodeOpen: vi.fn(), setBarcodeValue: vi.fn() },
      commands: { run: vi.fn(), openBarcode: vi.fn(), submitBarcode: vi.fn(), toggleField: vi.fn(),
        toggleImage: vi.fn(), applyResult: vi.fn() }
    } as unknown as ArticleSearchController;
    const dispatch = vi.fn((action) => { state = vehicleCreateWizardReducer(state, action); });

    render(<VehicleCreateStepArticle state={state} dispatch={dispatch} controller={controller}
      onUpdateShared={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /Set-Bild zuordnen/ }));
    await user.click(screen.getByRole("option", { name: /Wagen 2/ }));
    expect(dispatch).toHaveBeenCalledWith({
      type: "assign-article-image",
      imageURL: "https://roco.cc/set.jpg",
      memberIndex: 1
    });
  });

  it("shows every recognized article field before applying it", () => {
    const state: VehicleCreateWizardState = {
      ...createVehicleCreateWizardState({ manufacturer: "Roco", articleNumber: "6280002", name: "Set", gauge: "H0" }),
      step: "article",
      articleStage: "review",
      selectedResultIndex: 0
    };
    const controller = {
      state: { open: false, loading: false, response: { query: "Roco 6280002", results: [resultWithExtendedFields] },
        error: "", barcodeOpen: false, barcodeValue: "", selectedFields: {}, selectedImages: {} },
      setters: { setOpen: vi.fn(), setBarcodeOpen: vi.fn(), setBarcodeValue: vi.fn() },
      commands: { run: vi.fn(), openBarcode: vi.fn(), submitBarcode: vi.fn(), toggleField: vi.fn(),
        toggleImage: vi.fn(), applyResult: vi.fn() }
    } as unknown as ArticleSearchController;

    render(<VehicleCreateStepArticle state={state} dispatch={vi.fn()} controller={controller}
      onUpdateShared={vi.fn()} />);

    expect(screen.getByLabelText("Antrieb Beschreibung: Mittelmotor")).toBeInTheDocument();
    expect(screen.getByLabelText("Soundgenerator: Ja")).toBeInTheDocument();
    expect(screen.getByLabelText("Radsatz: AC")).toBeInTheDocument();
    expect(screen.getByLabelText("Kupplung vorne: Kurzkupplung")).toBeInTheDocument();
  });
});
