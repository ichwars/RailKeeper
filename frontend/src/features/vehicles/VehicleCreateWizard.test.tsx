import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { emptyVehicle } from "./vehicleViewModel";
import { VehicleCreateWizard, vehicleSetMembersFromForm } from "./VehicleCreateWizard";
import { emptyVehicleSetMemberDraft } from "./vehicleCreateWizardState";
import type { ArticleSearchController } from "./useArticleSearchController";

const articleSearchController = {
  state: { open: false, loading: false, response: null, error: "", barcodeOpen: false,
    barcodeValue: "", selectedFields: {}, selectedImages: {} },
  setters: { setOpen: vi.fn(), setBarcodeOpen: vi.fn(), setBarcodeValue: vi.fn() },
  commands: { run: vi.fn(), openBarcode: vi.fn(), submitBarcode: vi.fn(), toggleField: vi.fn(),
    toggleImage: vi.fn(), applyResult: vi.fn() }
} as unknown as ArticleSearchController;

const model = {
  form: { ...emptyVehicle, name: "TEE", manufacturer: "Roco", gauge: "H0", category: "Triebzug", gattung: "ET" },
  externalMappings: [],
  readonly: false,
  articleSearchLoading: false,
  canRunArticleSearch: false,
  options: { manufacturers: [], gauges: [], epochs: [], railwayCompanies: [], categories: [], gattungen: [], symbols: [], categoryRelations: [] },
  filteredGattungen: [],
  openSections: { model: true, details: false, vehicle: false },
  selectOptions: () => <option value="" />,
  ecosFieldClass: () => "",
  showRequiredErrors: false,
  onToggleSection: vi.fn(),
  onOpenBarcodeSearch: vi.fn(),
  onRunArticleSearch: vi.fn(),
  onUpdate: vi.fn(),
  onUpdateCategory: vi.fn(),
  onOpenQr: vi.fn(),
  canOpenQr: false,
  onUpdateCouplingFront: vi.fn(),
  onUpdateCouplingSame: vi.fn()
};

describe("VehicleCreateWizard set safeguards", () => {
  it("disables set creation for an ECoS draft", () => {
    render(<VehicleCreateWizard model={model} saving={false} message="" setCreationDisabled
      articleSearchController={articleSearchController}
      onSubmitSingle={vi.fn()} onSubmitSet={vi.fn()} onClose={vi.fn()} />);

    expect(screen.getByRole("radio", { name: /Set/ })).toBeDisabled();
  });

  it("keeps selected article images on the first set member", () => {
    const members = vehicleSetMembersFromForm(model.form, [
      { ...emptyVehicleSetMemberDraft(), form: { ...emptyVehicle, inventoryNumber: "1", name: "A", vehicleNumber: "A-1" } },
      { ...emptyVehicleSetMemberDraft(), form: { ...emptyVehicle, inventoryNumber: "2", name: "B", vehicleNumber: "B-1" } }
    ], [{ id: "image-1", url: "https://example.test/image.jpg", title: "Front", source: "catalog", isPrimary: true }]);

    expect(members[0].images).toEqual([expect.objectContaining({
      url: "https://example.test/image.jpg", title: "Front", sourceUrl: "catalog", isPrimary: true, sortOrder: 0
    })]);
    expect(members[1].images).toEqual([]);
  });
});
