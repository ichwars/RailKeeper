import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { emptyVehicle } from "./vehicleViewModel";
import { VehicleCreateWizard, vehicleSetMembersFromForm } from "./VehicleCreateWizard";
import {
  createVehicleCreateWizardState,
  emptyVehicleSetMemberDraft,
  saveVehicleCreateDraft,
  loadVehicleCreateDraft
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

  it("reuses one creation identity on retry, persists it with the draft and gives a new draft another identity", async () => {
    const user = userEvent.setup(), submit = vi.fn();
    const props = { model, saving: false, message: "", draftOwner: "tester", articleSearchController,
      onSubmitSingle: submit, onSubmitSet: vi.fn(), onClose: vi.fn() };
    const first = render(<VehicleCreateWizard {...props} />);
    await user.click(screen.getByRole("button", { name: "Weiter zu Artikeldaten" }));
    await user.click(screen.getByRole("button", { name: "Weiter zu Fahrzeugdetails" }));
    await user.click(screen.getByRole("button", { name: "Anlegen" }));
    await user.click(screen.getByRole("button", { name: "Anlegen" }));
    const id = submit.mock.calls[0][1];
    expect(id).toMatch(/^[0-9a-f]{32}$/); expect(submit.mock.calls[1][1]).toBe(id);
    await user.click(screen.getByRole("button", { name: "Entwurf speichern" }));
    const saved = loadVehicleCreateDraft("tester");
    expect(saved.kind).toBe("loaded"); if (saved.kind === "loaded") expect(saved.state.creationId).toBe(id);
    first.unmount();
    render(<VehicleCreateWizard {...props} />);
    await user.click(screen.getByRole("button", { name: "Weiter zu Artikeldaten" }));
    await user.click(screen.getByRole("button", { name: "Weiter zu Fahrzeugdetails" }));
    await user.click(screen.getByRole("button", { name: "Anlegen" }));
    expect(submit.mock.calls[2][1]).not.toBe(id);
  });

  it("disables set creation for an ECoS draft", () => {
    render(<VehicleCreateWizard model={model} saving={false} message="" setCreationDisabled
      draftOwner="tester"
      articleSearchController={articleSearchController}
      onSubmitSingle={vi.fn()} onSubmitSet={vi.fn()} onClose={vi.fn()} />);

    expect(screen.getByRole("radio", { name: /Set/ })).toBeDisabled();
  });

  it("keeps the restored set creation identity across repeated confirmations", async () => {
    const id = "c".repeat(32), submitSet = vi.fn().mockResolvedValue(undefined), user = userEvent.setup();
    saveVehicleCreateDraft({ ...createVehicleCreateWizardState(model.form), kind: "set", step: "details", activeDetailsTab: "set", creationId: id }, "tester");
    render(<VehicleCreateWizard model={model} saving={false} message="" draftOwner="tester"
      articleSearchController={articleSearchController} onSubmitSingle={vi.fn()}
      onSubmitSet={submitSet} onClose={vi.fn()} />);
    await user.click(screen.getByRole("button", { name: /Entwurf fortsetzen/ }));
    await user.click(screen.getByRole("button", { name: "Set anlegen" }));
    await user.click(screen.getByRole("button", { name: "Set anlegen" }));
    expect(submitSet).toHaveBeenCalledTimes(2);
    expect(submitSet.mock.calls.map(call => call[2])).toEqual([id, id]);
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
