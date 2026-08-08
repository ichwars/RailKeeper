import { useState } from "react";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { api, type AccessoryDocument, type MasterDataEntry } from "../../shared/api";
import { emptyArticleEditorForm } from "./articleEditorModel";
import { ArticleEditorDialog, type ArticleEditorDialogProps } from "./ArticleEditorDialog";
import type { CustomArticleSubjectFieldDefinition } from "./articleTypeFields";

const customFields: CustomArticleSubjectFieldDefinition[] = [
  { key: "material", kind: "text", label: "Material" },
  { key: "lengthMm", kind: "number", label: "Länge", unit: "mm", step: 0.1 }
];
const subtypeEntries: MasterDataEntry[] = [
  { id: "straight", type: "accessory_subtype", key: "track:straight", label: "Straight", active: true,
    sortOrder: 10, metadata: {}, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z" },
  { id: "custom", type: "accessory_subtype", key: "track:custom_profile", label: "Vereinsprofil", active: true,
    sortOrder: 20, metadata: {}, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z" },
  { id: "signal", type: "accessory_subtype", key: "signal:main", label: "Main", active: true,
    sortOrder: 10, metadata: {}, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z" }
];
const articleTypeEntries: MasterDataEntry[] = [
  { id: "track", type: "article_type", key: "track", label: "Track", active: true,
    sortOrder: 10, metadata: {}, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z" },
  { id: "signal", type: "article_type", key: "signal", label: "Signal", active: true,
    sortOrder: 20, metadata: {}, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z" },
  { id: "decoder", type: "article_type", key: "decoder", label: "Decoder", active: true,
    sortOrder: 30, metadata: {}, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z" },
  { id: "electrical_control", type: "article_type", key: "electrical_control", label: "Electrical control",
    active: true, sortOrder: 40, metadata: {}, createdAt: "2026-08-08T08:00:00Z",
    updatedAt: "2026-08-08T08:00:00Z" },
  { id: "building_equipment", type: "article_type", key: "building_equipment", label: "Building equipment",
    active: true, sortOrder: 50, metadata: {}, createdAt: "2026-08-08T08:00:00Z",
    updatedAt: "2026-08-08T08:00:00Z" },
  { id: "landscape_consumable", type: "article_type", key: "landscape_consumable",
    label: "Landscape consumable", active: true, sortOrder: 60, metadata: {},
    createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z" },
  { id: "lighting", type: "article_type", key: "lighting", label: "Lighting", active: true,
    sortOrder: 70, metadata: {}, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z" },
  { id: "other", type: "article_type", key: "other", label: "Other", active: true,
    sortOrder: 80, metadata: {}, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z" }
];
const configuredArticleTypeEntries = articleTypeEntries.map((entry) => entry.key === "signal"
  ? { ...entry, label: "Formsignal" }
  : entry.key === "decoder" ? { ...entry, active: false } : entry);

const persistedArticle = {
  id: "article-1", manufacturer: "Tillig", articleNumber: "83101", name: "Gleis", category: "straight",
  trackingMode: "quantity" as const, manufacturerStatus: "available" as const, articleType: "track" as const,
  subtype: "straight", gauges: ["TT"], packageQuantity: 1, stockUnit: "piece", minimumStock: 0,
  inventoryStrategy: "quantity" as const, alternativeNumbers: [], keywords: [], archived: false, attributes: [],
  createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T09:00:00Z"
};

function props(overrides: Partial<ArticleEditorDialogProps> = {}): ArticleEditorDialogProps {
  return {
    mode: "create",
    form: emptyArticleEditorForm(),
    article: null,
    activeTab: "article",
    hasUsageHistory: false,
    saving: false,
    loading: false,
    error: "",
    resourceError: "",
    fieldErrors: {},
    tabErrors: {},
    duplicateCandidates: [],
    closeConfirmationOpen: false,
    permissions: { canEdit: true, canManageStock: true, canReserve: true, canInstall: true },
    resources: {
      locations: [], stock: null, movements: [], assets: [], purchases: [], documents: [],
      reservations: [], installations: [], usageHistory: null, vehicles: [], layouts: [], units: []
    },
    resourcesStale: false,
    customFields: [],
    customFieldsLoading: false,
    customFieldsError: "",
    articleTypeEntries,
    articleTypeEntriesLoading: false,
    articleTypeEntriesError: "",
    subtypeEntries,
    subtypeEntriesLoading: false,
    subtypeEntriesError: "",
    subjectFieldErrors: {},
    onChange: vi.fn(),
    onTabChange: vi.fn(),
    onSubmit: vi.fn(),
    onRequestClose: vi.fn(),
    onConfirmClose: vi.fn(),
    onCancelClose: vi.fn(),
    onConfirmDuplicate: vi.fn(),
    onCancelDuplicate: vi.fn(),
    onResourcesChanged: vi.fn(),
    onRetryResources: vi.fn(),
    onRetryCustomFields: vi.fn().mockResolvedValue(undefined),
    onRetryArticleTypeEntries: vi.fn().mockResolvedValue(undefined),
    onRetrySubtypeEntries: vi.fn().mockResolvedValue(undefined),
    onSubdraftDirty: vi.fn(),
    ...overrides
  };
}

describe("ArticleEditorDialog", () => {
  it("renders create, view, and edit modes through one shell and disables view controls", () => {
    const { rerender } = render(<ArticleEditorDialog {...props()} />);
    expect(screen.getByRole("dialog", { name: "Artikel anlegen" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Hersteller" })).toBeEnabled();

    rerender(<ArticleEditorDialog {...props({ mode: "edit" })} />);
    expect(screen.getByRole("dialog", { name: "Artikel bearbeiten" })).toBeInTheDocument();

    rerender(<ArticleEditorDialog {...props({ mode: "view", permissions: {
      canEdit: true, canManageStock: true, canReserve: true, canInstall: true
    } })} />);
    expect(screen.getByRole("dialog", { name: "Artikel ansehen" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Hersteller" })).toBeDisabled();
    expect(screen.queryByRole("button", { name: "Änderungen speichern" })).not.toBeInTheDocument();
  });

  it("uses whole-number browser constraints for package quantity and minimum stock", () => {
    const view = render(<ArticleEditorDialog {...props()} />);
    expect(screen.getByRole("spinbutton", { name: "Verpackungseinheit" })).toHaveAttribute("min", "1");
    expect(screen.getByRole("spinbutton", { name: "Verpackungseinheit" })).toHaveAttribute("step", "1");

    view.rerender(<ArticleEditorDialog {...props({ activeTab: "stock" })} />);
    expect(screen.getByRole("spinbutton", { name: "Mindestbestand" })).toHaveAttribute("min", "0");
    expect(screen.getByRole("spinbutton", { name: "Mindestbestand" })).toHaveAttribute("step", "1");
  });

  it("uses active configured article types and keeps only the current inactive historical type", async () => {
    const user = userEvent.setup();
    const view = render(<ArticleEditorDialog {...props({ articleTypeEntries: configuredArticleTypeEntries })} />);

    await user.click(screen.getByRole("button", { name: "Artikelart" }));
    expect(screen.getByRole("option", { name: "Gleis" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Formsignal" })).toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "Decoder" })).not.toBeInTheDocument();

    view.rerender(<ArticleEditorDialog {...props({
      mode: "edit",
      articleTypeEntries: configuredArticleTypeEntries,
      article: { ...persistedArticle, articleType: "decoder", subtype: "decoder:locomotive" },
      form: { ...emptyArticleEditorForm(), articleType: "decoder", subtype: "decoder:locomotive" },
      activeTab: "subject"
    })} />);
    expect(screen.getByRole("tab", { name: "Fachangaben: Decoder" })).toBeInTheDocument();
    await user.click(screen.getByRole("tab", { name: "Artikel" }));
    view.rerender(<ArticleEditorDialog {...props({
      mode: "edit",
      articleTypeEntries: configuredArticleTypeEntries,
      article: { ...persistedArticle, articleType: "decoder", subtype: "decoder:locomotive" },
      form: { ...emptyArticleEditorForm(), articleType: "decoder", subtype: "decoder:locomotive" }
    })} />);
    await user.click(screen.getByRole("button", { name: "Artikelart" }));
    expect(screen.getByRole("option", { name: "Decoder" })).toBeDisabled();

    view.rerender(<ArticleEditorDialog {...props({
      articleTypeEntries: configuredArticleTypeEntries,
      form: { ...emptyArticleEditorForm(), articleType: "signal" }, activeTab: "subject"
    })} />);
    expect(screen.getByRole("tab", { name: "Fachangaben: Formsignal" })).toBeInTheDocument();
  });

  it("uses a controlled localized subtype select filtered by article type and preserves custom labels", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<ArticleEditorDialog {...props({
      form: { ...emptyArticleEditorForm(), articleType: "track", subtype: "straight" }, onChange
    })} />);

    expect(screen.queryByRole("textbox", { name: "Unterart" })).not.toBeInTheDocument();
    const subtype = screen.getByRole("button", { name: "Unterart" });
    expect(subtype).toHaveTextContent("Gerade");
    await user.click(subtype);
    expect(screen.getByRole("option", { name: "Gerade" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Vereinsprofil" })).toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "Hauptsignal" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("option", { name: "Vereinsprofil" }));
    expect(onChange).toHaveBeenCalledWith({ subtype: "custom_profile" });
  });

  it("localizes the canonical subtype key returned by the backend", () => {
    render(<ArticleEditorDialog {...props({
      mode: "edit", form: { ...emptyArticleEditorForm(), articleType: "track", subtype: "track:straight" }
    })} />);

    expect(screen.getByRole("button", { name: "Unterart" })).toHaveTextContent("Gerade");
    expect(screen.queryByText("track:straight")).not.toBeInTheDocument();
  });

  it("honors an administrator-renamed built-in subtype label", () => {
    render(<ArticleEditorDialog {...props({
      form: { ...emptyArticleEditorForm(), articleType: "track", subtype: "straight" },
      subtypeEntries: [{ ...subtypeEntries[0]!, label: "Werkstattgerade" }, ...subtypeEntries.slice(1)]
    })} />);

    expect(screen.getByRole("button", { name: "Unterart" })).toHaveTextContent("Werkstattgerade");
  });

  it("connects the required subtype select to its stable validation message", () => {
    render(<ArticleEditorDialog {...props({
      form: { ...emptyArticleEditorForm(), articleType: "track", subtype: "" },
      fieldErrors: { subtype: "Unterart ist erforderlich." }
    })} />);

    const subtype = screen.getByRole("button", { name: "Unterart" });
    expect(subtype).toHaveAttribute("aria-required", "true");
    expect(subtype).toHaveAttribute("aria-invalid", "true");
    expect(subtype).toHaveAttribute("aria-describedby", "article-editor-subtype-error");
    expect(screen.getByRole("alert")).toHaveAttribute("id", "article-editor-subtype-error");
  });

  it("represents an inactive historical subtype without permitting arbitrary raw keys", async () => {
    const user = userEvent.setup();
    const historical = { ...subtypeEntries[0]!, id: "historical", key: "track:old_profile",
      label: "Altes Profil", active: false };
    render(<ArticleEditorDialog {...props({
      mode: "edit", form: { ...emptyArticleEditorForm(), articleType: "track", subtype: "old_profile" },
      subtypeEntries: [...subtypeEntries, historical]
    })} />);

    const subtype = screen.getByRole("button", { name: "Unterart" });
    expect(subtype).toHaveTextContent("Altes Profil");
    await user.click(subtype);
    expect(screen.getByRole("option", { name: /Altes Profil/ })).toBeInTheDocument();
    expect(screen.queryByRole("textbox", { name: "Unterart" })).not.toBeInTheDocument();
  });

  it("disables subtype selection while loading, on load failure, and in read-only mode with retry", async () => {
    const user = userEvent.setup();
    const retry = vi.fn().mockResolvedValue(undefined);
    const view = render(<ArticleEditorDialog {...props({ subtypeEntriesLoading: true })} />);
    expect(screen.getByRole("button", { name: "Unterart" })).toBeDisabled();

    view.rerender(<ArticleEditorDialog {...props({
      subtypeEntriesError: "Unterarten nicht verfügbar", onRetrySubtypeEntries: retry
    })} />);
    expect(screen.getByRole("button", { name: "Unterart" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "Unterarten erneut laden" }));
    expect(retry).toHaveBeenCalledOnce();

    view.rerender(<ArticleEditorDialog {...props({ mode: "view" })} />);
    expect(screen.getByRole("button", { name: "Unterart" })).toBeDisabled();
  });

  it("disables type-dependent create controls until authoritative article types are loaded", () => {
    const view = render(<ArticleEditorDialog {...props({
      form: { ...emptyArticleEditorForm(), articleType: "track" },
      articleTypeEntriesLoading: true
    })} />);

    expect(screen.getByRole("button", { name: "Unterart" })).toBeDisabled();
    view.rerender(<ArticleEditorDialog {...props({
      form: { ...emptyArticleEditorForm(), articleType: "track" },
      activeTab: "subject",
      articleTypeEntriesLoading: true
    })} />);
    expect(screen.getByRole("textbox", { name: "Gleissystem" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Artikel anlegen" })).toBeDisabled();
  });

  it("shows custom-field retry, allows switching to a standard type, and only blocks other save", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const onRetryCustomFields = vi.fn().mockResolvedValue(undefined);
    const view = render(<ArticleEditorDialog {...props({
      customFieldsError: "Zusatzfelder nicht verfügbar",
      onChange,
      onRetryCustomFields
    })} />);

    expect(screen.getByRole("alert")).toHaveTextContent("Zusatzfelder nicht verfügbar");
    expect(screen.getByRole("button", { name: "Artikel anlegen" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "Zusatzfelder erneut laden" }));
    expect(onRetryCustomFields).toHaveBeenCalledOnce();
    await user.click(screen.getByRole("button", { name: "Artikelart" }));
    await user.click(screen.getByRole("option", { name: "Gleis" }));
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ articleType: "track" }));

    view.rerender(<ArticleEditorDialog {...props({
      form: { ...emptyArticleEditorForm(), articleType: "track" },
      customFieldsError: "Zusatzfelder nicht verfügbar",
      onRetryCustomFields
    })} />);
    expect(screen.getByRole("button", { name: "Artikel anlegen" })).toBeEnabled();

    view.rerender(<ArticleEditorDialog {...props({
      activeTab: "subject",
      customFieldsError: "Zusatzfelder nicht verfügbar",
      onRetryCustomFields
    })} />);
    expect(screen.getAllByRole("alert")).toHaveLength(1);
  });

  it("keeps inactive historical custom attributes hidden until an explicit confirmed type change", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const historicalForm = { ...emptyArticleEditorForm(), articleType: "other" as const, subtype: "legacy",
      attributes: [{ key: "legacyMaterial", kind: "text" as const, textValue: "Holz" }] };
    const view = render(<ArticleEditorDialog {...props({
      mode: "edit",
      form: historicalForm,
      activeTab: "subject",
      onChange
    })} />);

    expect(screen.queryByRole("textbox", { name: "legacyMaterial" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("tab", { name: "Artikel" }));
    view.rerender(<ArticleEditorDialog {...props({
      mode: "edit", form: historicalForm, activeTab: "article", onChange
    })} />);
    await user.click(screen.getByRole("button", { name: "Artikelart" }));
    await user.click(screen.getByRole("option", { name: "Gleis" }));
    await user.click(screen.getByRole("button", { name: "Fachwerte verwerfen" }));
    expect(onChange).toHaveBeenLastCalledWith(expect.objectContaining({ articleType: "track", attributes: [] }));
  });

  it("keeps form values across unmounted tabs and renders exactly one subject seam", async () => {
    const user = userEvent.setup();
    const onTabChange = vi.fn();
    const form = { ...emptyArticleEditorForm(), manufacturer: "Tillig" };
    const view = render(<ArticleEditorDialog {...props({ form, onTabChange })} />);

    await user.click(screen.getByRole("tab", { name: "Bestand" }));
    expect(onTabChange).toHaveBeenCalledWith("stock");
    view.rerender(<ArticleEditorDialog {...props({ form, activeTab: "stock", onTabChange })} />);
    expect(screen.queryByRole("textbox", { name: "Hersteller" })).not.toBeInTheDocument();
    expect(screen.getAllByRole("tab").filter((tab) => tab.dataset.tabKind === "subject")).toHaveLength(1);

    view.rerender(<ArticleEditorDialog {...props({ form, activeTab: "article", onTabChange })} />);
    expect(screen.getByRole("textbox", { name: "Hersteller" })).toHaveValue("Tillig");
  });

  it("renders the selected type fields in the single dynamic subject tab", () => {
    render(<ArticleEditorDialog {...props({
      form: { ...emptyArticleEditorForm(), articleType: "track" }, activeTab: "subject"
    })} />);

    expect(screen.getAllByRole("tab").filter((tab) => tab.dataset.tabKind === "subject")).toHaveLength(1);
    expect(screen.getByRole("textbox", { name: "Gleissystem" })).toBeInTheDocument();
    expect(screen.queryByRole("textbox", { name: "Firmware" })).not.toBeInTheDocument();
  });

  it("confirms type changes before discarding only incompatible subject values", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const form = {
      ...emptyArticleEditorForm(), articleType: "track" as const, subtype: "straight",
      attributes: [
        { key: "trackSystem", kind: "text" as const, textValue: "Tillig TT Modellgleis" },
        { key: "mounting", kind: "single_select" as const, optionValues: ["surface"] as [string] }
      ]
    };
    render(<ArticleEditorDialog {...props({ form, onChange })} />);

    await user.click(screen.getByRole("button", { name: "Artikelart" }));
    await user.click(screen.getByRole("option", { name: "Signal" }));
    expect(onChange).not.toHaveBeenCalled();
    expect(screen.getByRole("dialog", { name: "Artikelart ändern" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Fachwerte verwerfen" }));
    expect(onChange).toHaveBeenCalledWith({
      articleType: "signal", subtype: "", attributes: [
        { key: "mounting", kind: "single_select", optionValues: ["surface"] }
      ], attributeNumberDrafts: {}
    });
  });

  it("keeps the original type and subject values when type-change discard is cancelled", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<ArticleEditorDialog {...props({
      form: { ...emptyArticleEditorForm(), articleType: "track", attributes: [
        { key: "trackSystem", kind: "text", textValue: "Tillig TT Modellgleis" }
      ] },
      onChange
    })} />);

    await user.click(screen.getByRole("button", { name: "Artikelart" }));
    await user.click(screen.getByRole("option", { name: "Signal" }));
    await user.click(screen.getByRole("button", { name: "Weiter bearbeiten" }));

    expect(onChange).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Artikelart" })).toHaveTextContent("Gleis");
  });

  it("prompts before clearing a non-empty subtype even without subject attributes", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<ArticleEditorDialog {...props({
      form: { ...emptyArticleEditorForm(), articleType: "track", subtype: "straight" },
      onChange
    })} />);

    await user.click(screen.getByRole("button", { name: "Artikelart" }));
    await user.click(screen.getByRole("option", { name: "Signal" }));
    expect(screen.getByRole("dialog", { name: "Artikelart ändern" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Weiter bearbeiten" }));
    expect(onChange).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Unterart" })).toHaveTextContent("Gerade");

    await user.click(screen.getByRole("button", { name: "Artikelart" }));
    await user.click(screen.getByRole("option", { name: "Signal" }));
    await user.click(screen.getByRole("button", { name: "Fachwerte verwerfen" }));
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ articleType: "signal", subtype: "" }));
  });

  it("prompts when a non-empty incompatible draft is hidden beside an empty compatible draft", async () => {
    const user = userEvent.setup();
    render(<ArticleEditorDialog {...props({
      form: { ...emptyArticleEditorForm(), articleType: "track",
        attributeNumberDrafts: { ledCount: "", radiusMm: "12" } }
    })} />);

    await user.click(screen.getByRole("button", { name: "Artikelart" }));
    await user.click(screen.getByRole("option", { name: "Beleuchtung" }));

    expect(screen.getByRole("dialog", { name: "Artikelart ändern" })).toBeInTheDocument();
  });

  it("preserves values compatible with loaded active custom fields when switching to other", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<ArticleEditorDialog {...props({
      customFields,
      form: {
        ...emptyArticleEditorForm(), articleType: "landscape_consumable", subtype: "grass",
        attributes: [
          { key: "material", kind: "text", textValue: "Naturfaser" },
          { key: "content", kind: "number", numberValue: 20 }
        ],
        attributeNumberDrafts: { content: "20", lengthMm: "166" }
      },
      onChange
    })} />);

    await user.click(screen.getByRole("button", { name: "Artikelart" }));
    await user.click(screen.getByRole("option", { name: "Sonstiger Artikel" }));
    await user.click(screen.getByRole("button", { name: "Fachwerte verwerfen" }));

    expect(onChange).toHaveBeenCalledWith({
      articleType: "other",
      subtype: "",
      attributes: [{ key: "material", kind: "text", textValue: "Naturfaser" }],
      attributeNumberDrafts: { lengthMm: "166" }
    });
  });

  it("shows tab error badges and mounts usage history only for a real signal", () => {
    const view = render(<ArticleEditorDialog {...props({ tabErrors: { stock: true } })} />);
    expect(screen.getByRole("tab", { name: /Bestand.*Fehler/ })).toBeInTheDocument();
    expect(screen.queryByRole("tab", { name: "Verwendung & Historie" })).not.toBeInTheDocument();

    view.rerender(<ArticleEditorDialog {...props({ hasUsageHistory: true })} />);
    expect(screen.getByRole("tab", { name: "Verwendung & Historie" })).toBeInTheDocument();
  });

  it("scrolls a programmatically activated validation tab into the mobile tab strip", () => {
    const scrollIntoView = vi.fn();
    HTMLElement.prototype.scrollIntoView = scrollIntoView;
    const view = render(<ArticleEditorDialog {...props({ activeTab: "article" })} />);
    scrollIntoView.mockClear();

    view.rerender(<ArticleEditorDialog {...props({ activeTab: "subject", tabErrors: { subject: true } })} />);

    expect(scrollIntoView).toHaveBeenCalledWith({ block: "nearest", inline: "nearest" });
  });

  it("sets initial focus, traps focus, and returns focus to the initiating action", async () => {
    const user = userEvent.setup();
    const trigger = document.createElement("button");
    trigger.textContent = "Öffnen";
    document.body.append(trigger);
    trigger.focus();
    const view = render(<ArticleEditorDialog {...props({ returnFocusTo: trigger })} />);

    expect(await screen.findByRole("textbox", { name: "Hersteller" })).toHaveFocus();
    const close = screen.getByRole("button", { name: "Dialog schließen" });
    close.focus();
    await user.tab();
    expect(screen.getByRole("textbox", { name: "Hersteller" })).toHaveFocus();

    view.unmount();
    expect(trigger).toHaveFocus();
    trigger.remove();
  });

  it("uses only enabled view controls for forward and reverse focus traversal", async () => {
    const user = userEvent.setup();
    render(<ArticleEditorDialog {...props({
      mode: "view",
      article: persistedArticle,
      permissions: { canEdit: false, canManageStock: false, canReserve: false, canInstall: false }
    })} />);

    const headerClose = screen.getByRole("button", { name: "Dialog schließen" });
    const footerClose = screen.getByRole("button", { name: "Schließen" });
    expect(headerClose).toHaveFocus();

    await user.tab();
    expect(screen.getByRole("tab", { name: "Artikel" })).toHaveFocus();
    expect(screen.getByRole("textbox", { name: "Hersteller" })).toBeDisabled();

    headerClose.focus();
    await user.tab({ shift: true });
    expect(footerClose).toHaveFocus();
  });

  it("keeps the Planner reservation workflow reachable through the view focus trap", async () => {
    const user = userEvent.setup();
    render(<ArticleEditorDialog {...props({
      mode: "view",
      article: persistedArticle,
      activeTab: "stock",
      permissions: { canEdit: false, canManageStock: false, canReserve: true, canInstall: false },
      resources: {
        ...props().resources,
        locations: [{ id: "location-1", name: "Lager", archived: false,
          createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z" }],
        layouts: [{ id: "layout-1", name: "Vereinsanlage", kind: "club", gauge: "TT", scale: "1:120",
          description: "", version: 1, archived: false,
          createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z" }]
      }
    })} />);

    const saveReservation = screen.getByRole("button", { name: "Reservierung anlegen" });
    for (let index = 0; index < 30 && document.activeElement !== saveReservation; index += 1) {
      await user.tab();
    }
    expect(saveReservation).toHaveFocus();
  });

  it("renders dirty-close and duplicate confirmations without replacing form values", async () => {
    const user = userEvent.setup();
    const onConfirmClose = vi.fn();
    const onConfirmDuplicate = vi.fn();
    const view = render(<ArticleEditorDialog {...props({ closeConfirmationOpen: true, onConfirmClose })} />);
    await user.click(screen.getByRole("button", { name: "Verwerfen" }));
    expect(onConfirmClose).toHaveBeenCalledOnce();

    view.rerender(<ArticleEditorDialog {...props({
      form: { ...emptyArticleEditorForm(), manufacturer: "Tillig" },
      duplicateCandidates: [{ id: "dup", manufacturer: "Tillig", articleNumber: "83101", name: "Gleis", articleType: "track", subtype: "straight" }],
      onConfirmDuplicate
    })} />);
    expect(screen.getByRole("textbox", { name: "Hersteller", hidden: true })).toHaveValue("Tillig");
    await user.click(screen.getByRole("button", { name: "Trotzdem speichern" }));
    expect(onConfirmDuplicate).toHaveBeenCalledOnce();
  });

  it("allows Planner reservation controls but no other mutation in the read-only article shell", () => {
    render(<ArticleEditorDialog {...props({
      mode: "view",
      article: persistedArticle,
      activeTab: "stock",
      hasUsageHistory: true,
      permissions: { canEdit: false, canManageStock: false, canReserve: true, canInstall: false }
    })} />);

    expect(screen.getByRole("button", { name: "Reservierung anlegen" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Einbau erfassen" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Änderungen speichern" })).not.toBeInTheDocument();
  });

  it("keeps purchase subdraft values mounted across fixed tab switches", async () => {
    const user = userEvent.setup();
    const view = render(<ArticleEditorDialog {...props({
      mode: "edit", article: persistedArticle, activeTab: "purchaseDocuments"
    })} />);
    await user.type(screen.getByRole("textbox", { name: "Bezugsquelle" }), "Modellbahnshop");

    view.rerender(<ArticleEditorDialog {...props({ mode: "edit", article: persistedArticle, activeTab: "article" })} />);
    view.rerender(<ArticleEditorDialog {...props({
      mode: "edit", article: persistedArticle, activeTab: "purchaseDocuments"
    })} />);

    expect(screen.getByRole("textbox", { name: "Bezugsquelle" })).toHaveValue("Modellbahnshop");
  });

  it("offers first reservation and installation actions in Stock before usage history exists", () => {
    render(<ArticleEditorDialog {...props({
      mode: "edit", article: persistedArticle, activeTab: "stock", hasUsageHistory: false
    })} />);

    expect(screen.queryByRole("tab", { name: "Verwendung & Historie" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Reservierung anlegen" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Einbau erfassen" })).toBeInTheDocument();
  });

  it("shows the primary uploaded image from refreshed document resources", () => {
    render(<ArticleEditorDialog {...props({
      mode: "edit",
      article: persistedArticle,
      resources: {
        ...props().resources,
        documents: [{ id: "image-1", productId: persistedArticle.id, originalName: "Front.png",
          fileName: "front.png", category: "image", mimeType: "image/png", sizeBytes: 100,
          isPrimary: true, createdBy: "admin", createdAt: "2026-08-08T09:00:00Z",
          updatedAt: "2026-08-08T09:00:00Z" }]
      }
    })} />);

    expect(screen.getByRole("img", { name: "Produktbild" })).toHaveAttribute(
      "src", "/api/v1/accessory-products/article-1/documents/image-1/download"
    );
  });

  it("prefers a refreshed primary document over the stale article image URL", () => {
    render(<ArticleEditorDialog {...props({
      mode: "edit",
      article: { ...persistedArticle, primaryImageUrl: "/old-primary.png" },
      resources: {
        ...props().resources,
        documents: [{ id: "image-2", productId: persistedArticle.id, originalName: "Neu.png",
          fileName: "new.png", category: "image", mimeType: "image/png", sizeBytes: 100,
          isPrimary: true, createdBy: "admin", createdAt: "2026-08-08T10:00:00Z",
          updatedAt: "2026-08-08T10:00:00Z" }]
      }
    })} />);

    expect(screen.getByRole("img", { name: "Produktbild" })).toHaveAttribute(
      "src", "/api/v1/accessory-products/article-1/documents/image-2/download"
    );
  });

  it("clears a deleted primary image and makes the next uploaded image primary", async () => {
    const user = userEvent.setup();
    const primaryA: AccessoryDocument = {
      id: "image-a", productId: persistedArticle.id, originalName: "A.png", fileName: "a.png",
      category: "image", mimeType: "image/png", sizeBytes: 100, isPrimary: true, createdBy: "admin",
      createdAt: "2026-08-08T09:00:00Z", updatedAt: "2026-08-08T09:00:00Z"
    };
    const primaryB: AccessoryDocument = {
      ...primaryA, id: "image-b", originalName: "B.png", fileName: "b.png",
      createdAt: "2026-08-08T10:00:00Z", updatedAt: "2026-08-08T10:00:00Z"
    };
    vi.spyOn(api, "deleteAccessoryDocument").mockResolvedValue(undefined);
    const upload = vi.spyOn(api, "uploadAccessoryDocument").mockResolvedValue(primaryB);

    function Harness() {
      const [activeTab, setActiveTab] = useState<ArticleEditorDialogProps["activeTab"]>("purchaseDocuments");
      const [documents, setDocuments] = useState<AccessoryDocument[]>([primaryA]);
      const refresh = async () => setDocuments((current) => current.length > 0 && current[0]?.id === "image-a"
        ? [] : [primaryB]);
      return <ArticleEditorDialog {...props({
        mode: "edit",
        article: { ...persistedArticle, primaryImageUrl: "/stale-image-a.png" },
        activeTab,
        resources: { ...props().resources, documents, documentsLoaded: true },
        onTabChange: setActiveTab,
        onResourcesChanged: refresh
      })} />;
    }
    render(<Harness />);
    const parentDialog = screen.getByRole("dialog", { name: "Artikel bearbeiten" });

    await user.click(screen.getByRole("button", { name: "Dokument löschen: A.png" }));
    const deleteConfirmation = screen.getByRole("dialog", { name: "Dokument löschen" });
    expect(parentDialog).toHaveAttribute("aria-hidden", "true");
    expect(parentDialog).toHaveAttribute("inert");
    expect(parentDialog).not.toContainElement(deleteConfirmation);
    expect(within(deleteConfirmation).getByRole("button", { name: "Abbrechen" })).toHaveFocus();
    await user.click(screen.getByRole("button", { name: "Löschen" }));
    await user.click(screen.getByRole("tab", { name: "Artikel" }));
    expect(screen.queryByRole("img", { name: "Produktbild" })).not.toBeInTheDocument();
    expect(screen.getByText("Kein Produktbild")).toBeInTheDocument();

    await user.click(screen.getByRole("tab", { name: "Kauf & Dokumente" }));
    const nextImage = new File(["image-b"], "B.png", { type: "image/png" });
    await user.upload(screen.getByLabelText("Datei", { selector: "input" }), nextImage);
    await user.click(screen.getByRole("button", { name: "Dokumentart" }));
    await user.click(screen.getByRole("option", { name: "Produktbild" }));
    await user.click(screen.getByRole("button", { name: "Dokument hochladen" }));
    expect(upload).toHaveBeenCalledWith(persistedArticle.id, {
      file: nextImage, category: "image", description: undefined, isPrimary: true
    });

    await user.click(screen.getByRole("tab", { name: "Artikel" }));
    expect(screen.getByRole("img", { name: "Produktbild" })).toHaveAttribute(
      "src", "/api/v1/accessory-products/article-1/documents/image-b/download"
    );
  });

  it("moves edit focus only after detail loading finishes", () => {
    const view = render(<ArticleEditorDialog {...props({ mode: "edit", loading: true })} />);
    expect(screen.queryByRole("textbox", { name: "Hersteller" })).not.toBeInTheDocument();

    view.rerender(<ArticleEditorDialog {...props({ mode: "edit", loading: false })} />);

    expect(screen.getByRole("textbox", { name: "Hersteller" })).toHaveFocus();
  });

  it("makes dirty confirmation own Escape and restores focus to its invoker", async () => {
    const user = userEvent.setup();
    const onCancelClose = vi.fn();
    const onRequestClose = vi.fn();
    const view = render(<ArticleEditorDialog {...props({ onCancelClose, onRequestClose })} />);
    const invoker = screen.getByRole("tab", { name: "Artikel" });
    const parentDialog = screen.getByRole("dialog", { name: "Artikel anlegen" });
    invoker.focus();
    view.rerender(<ArticleEditorDialog {...props({ closeConfirmationOpen: true, onCancelClose, onRequestClose })} />);
    const confirmation = screen.getByRole("dialog", { name: "Ungespeicherte Änderungen" });
    expect(parentDialog).toHaveAttribute("aria-hidden", "true");
    expect(parentDialog).toHaveAttribute("inert");
    expect(parentDialog).not.toContainElement(confirmation);
    expect(within(confirmation).getByRole("button", { name: "Weiter bearbeiten" })).toHaveFocus();

    await user.keyboard("{Escape}");
    expect(onCancelClose).toHaveBeenCalledOnce();
    expect(onRequestClose).not.toHaveBeenCalled();
    view.rerender(<ArticleEditorDialog {...props({ onCancelClose, onRequestClose })} />);
    expect(parentDialog).not.toHaveAttribute("aria-hidden");
    expect(parentDialog).not.toHaveAttribute("inert");
    expect(invoker).toHaveFocus();
  });

  it("makes the article form inert while duplicate confirmation is pending", () => {
    const view = render(<ArticleEditorDialog {...props()} />);
    const parentDialog = screen.getByRole("dialog", { name: "Artikel anlegen" });
    view.rerender(<ArticleEditorDialog {...props({ duplicateCandidates: [{ id: "dup", manufacturer: "Tillig",
      articleNumber: "83101", name: "Gleis", articleType: "track", subtype: "straight" }] })} />);

    const confirmation = screen.getByRole("dialog", { name: "Mögliche Dublette" });
    expect(parentDialog).toHaveAttribute("aria-hidden", "true");
    expect(parentDialog).toHaveAttribute("inert");
    expect(parentDialog).not.toContainElement(confirmation);
    expect(screen.getByRole("textbox", { name: "Hersteller", hidden: true })).toBeDisabled();
  });

  it("disables resource mutations while stale and offers an explicit retry", async () => {
    const user = userEvent.setup();
    const onRetryResources = vi.fn().mockRejectedValue(new Error("Weiterhin nicht verfügbar"));
    render(<ArticleEditorDialog {...props({
      mode: "edit", article: persistedArticle, activeTab: "stock", resourcesStale: true,
      error: "Bestand konnte nicht aktualisiert werden",
      resourceError: "Bestand konnte nicht aktualisiert werden", onRetryResources
    })} />);

    expect(screen.queryByRole("button", { name: "Reservierung anlegen" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Einbau erfassen" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Daten erneut laden" }));
    expect(onRetryResources).toHaveBeenCalledOnce();
  });
});
