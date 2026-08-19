import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { emptyVehicle } from "./vehicleViewModel";
import { VehicleCreateWizard, vehicleSetMembersFromForm } from "./VehicleCreateWizard";
import {
  createVehicleCreateWizardState,
  emptyVehicleSetMemberDraft,
  saveVehicleCreateDraft
} from "./vehicleCreateWizardState";
import type { VehicleSetMemberDraft } from "./vehicleCreateWizardState";
import type { ArticleSearchController } from "./useArticleSearchController";

const articleSearchController = {
  state: { open: false, loading: false, response: null, error: "", barcodeOpen: false,
    barcodeValue: "", selectedFields: {}, selectedImages: {} },
  setters: { setOpen: vi.fn(), setBarcodeOpen: vi.fn(), setBarcodeValue: vi.fn() },
  commands: { run: vi.fn(), openBarcode: vi.fn(), submitBarcode: vi.fn(), toggleField: vi.fn(),
    toggleImage: vi.fn(), applyResult: vi.fn(), restoreDraft: vi.fn() }
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
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
    vi.spyOn(api, "inventoryNumberSchemes").mockResolvedValue([]);
  });

  it("disables set creation for an ECoS draft", () => {
    render(<VehicleCreateWizard model={model} saving={false} message="" setCreationDisabled
      draftOwner="tester"
      articleSearchController={articleSearchController}
      onSubmitSingle={vi.fn()} onSubmitSet={vi.fn()} onClose={vi.fn()} />);

    expect(screen.getByRole("radio", { name: /Set/ })).toBeDisabled();
  });

  it("synchronizes a resumed draft with the parent form", async () => {
    const user = userEvent.setup();
    const draft = {
      ...createVehicleCreateWizardState({
        ...emptyVehicle,
        name: "Draft TEE",
        manufacturer: "Roco",
        gauge: "H0",
        category: "Triebzug"
      }),
      kind: "set" as const
    };
    saveVehicleCreateDraft(draft, "tester");

    render(<VehicleCreateWizard model={model} saving={false} message=""
      draftOwner="tester"
      articleSearchController={articleSearchController}
      onSubmitSingle={vi.fn()} onSubmitSet={vi.fn()} onClose={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /Entwurf fortsetzen/ }));
    expect(model.onUpdate).toHaveBeenCalledWith(expect.objectContaining({ name: "Draft TEE" }));
  });

  it("does not resume a set draft during an ECoS single-vehicle workflow", () => {
    const draft = { ...createVehicleCreateWizardState(model.form), kind: "set" as const };
    saveVehicleCreateDraft(draft, "tester");

    render(<VehicleCreateWizard model={model} saving={false} message="" setCreationDisabled
      draftOwner="tester" articleSearchController={articleSearchController}
      onSubmitSingle={vi.fn()} onSubmitSet={vi.fn()} onClose={vi.fn()} />);

    expect(screen.getByRole("button", { name: /Entwurf fortsetzen/ })).toBeDisabled();
  });

  it("loads the set inventory scheme only once per language", async () => {
    render(<VehicleCreateWizard model={model} saving={false} message=""
      draftOwner="tester"
      articleSearchController={articleSearchController}
      onSubmitSingle={vi.fn()} onSubmitSet={vi.fn()} onClose={vi.fn()} />);

    await waitFor(() => expect(api.inventoryNumberSchemes).toHaveBeenCalledTimes(1));
    await new Promise((resolve) => setTimeout(resolve, 25));
    expect(api.inventoryNumberSchemes).toHaveBeenCalledTimes(1);
  });

  it("opens the first invalid member instead of submitting the set", async () => {
    const user = userEvent.setup();
    const onSubmitSet = vi.fn();
    const draft = {
      ...createVehicleCreateWizardState(model.form),
      kind: "set" as const,
      step: "details" as const,
      activeDetailsTab: "set" as const,
      members: [
        {
          ...emptyVehicleSetMemberDraft(),
          form: { ...emptyVehicle, maximumSpeedKmh: 0 },
          touched: true,
          overriddenFields: ["maximumSpeedKmh" as const]
        },
        emptyVehicleSetMemberDraft()
      ]
    };
    saveVehicleCreateDraft(draft, "tester");

    render(<VehicleCreateWizard model={model} saving={false} message=""
      draftOwner="tester" articleSearchController={articleSearchController}
      onSubmitSingle={vi.fn()} onSubmitSet={onSubmitSet} onClose={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /Entwurf fortsetzen/ }));
    await user.click(screen.getByRole("button", { name: /Set anlegen/ }));

    expect(screen.getByRole("tab", { name: /Wagen 1/ })).toHaveAttribute("aria-selected", "true");
    expect(screen.getByText("Bitte prüfe die Angaben für Wagen 1.")).toBeInTheDocument();
    expect(onSubmitSet).not.toHaveBeenCalled();
  });

  it("preserves imported technical fields unless a member explicitly overrides them", () => {
    const shared = {
      ...model.form,
      series: "BR 601",
      lengthMm: "235",
      digital: true,
      decoderType: "DCC",
      category: "Triebzug",
      gattung: "Dieseltriebzug"
    };
    const first = emptyVehicleSetMemberDraft();
    const second: VehicleSetMemberDraft = {
      ...emptyVehicleSetMemberDraft(),
      form: { ...emptyVehicle, name: "B", lengthMm: "240", category: "Wagen", gattung: "Steuerwagen" },
      overriddenFields: ["name", "lengthMm", "category", "gattung"]
    };

    const members = vehicleSetMembersFromForm(shared, [first, second], [], {});

    expect(members[0]).toMatchObject({ series: "BR 601", lengthMm: "235", digital: true, decoderType: "DCC" });
    expect(members[1]).toMatchObject({
      series: "BR 601", lengthMm: "240", digital: true, decoderType: "DCC",
      category: "Wagen", gattung: "Steuerwagen"
    });
  });

  it("assigns selected article images to their chosen set members", () => {
    const members = vehicleSetMembersFromForm(model.form, [
      { ...emptyVehicleSetMemberDraft(), form: { ...emptyVehicle, inventoryNumber: "1", name: "A", vehicleNumber: "A-1" } },
      { ...emptyVehicleSetMemberDraft(), form: { ...emptyVehicle, inventoryNumber: "2", name: "B", vehicleNumber: "B-1" } }
    ], [{ id: "image-1", url: "https://example.test/image.jpg", title: "Front", source: "catalog", isPrimary: true }], {
      "https://example.test/image.jpg": 1
    });

    expect(members[0].images).toEqual([]);
    expect(members[1].images).toEqual([expect.objectContaining({
      url: "https://example.test/image.jpg", title: "Front", sourceUrl: "catalog", isPrimary: true, sortOrder: 0
    })]);
  });
});
