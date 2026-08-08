import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { emptyArticleEditorForm } from "./articleEditorModel";
import { ArticleEditorDialog, type ArticleEditorDialogProps } from "./ArticleEditorDialog";

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

  it("shows tab error badges and mounts usage history only for a real signal", () => {
    const view = render(<ArticleEditorDialog {...props({ tabErrors: { stock: true } })} />);
    expect(screen.getByRole("tab", { name: /Bestand.*Fehler/ })).toBeInTheDocument();
    expect(screen.queryByRole("tab", { name: "Verwendung & Historie" })).not.toBeInTheDocument();

    view.rerender(<ArticleEditorDialog {...props({ hasUsageHistory: true })} />);
    expect(screen.getByRole("tab", { name: "Verwendung & Historie" })).toBeInTheDocument();
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
    expect(screen.getByRole("textbox", { name: "Hersteller" })).toHaveValue("Tillig");
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
    invoker.focus();
    view.rerender(<ArticleEditorDialog {...props({ closeConfirmationOpen: true, onCancelClose, onRequestClose })} />);
    const confirmation = screen.getByRole("dialog", { name: "Ungespeicherte Änderungen" });
    expect(within(confirmation).getByRole("button", { name: "Weiter bearbeiten" })).toHaveFocus();

    await user.keyboard("{Escape}");
    expect(onCancelClose).toHaveBeenCalledOnce();
    expect(onRequestClose).not.toHaveBeenCalled();
    view.rerender(<ArticleEditorDialog {...props({ onCancelClose, onRequestClose })} />);
    expect(invoker).toHaveFocus();
  });

  it("makes the article form inert while duplicate confirmation is pending", () => {
    render(<ArticleEditorDialog {...props({ duplicateCandidates: [{ id: "dup", manufacturer: "Tillig",
      articleNumber: "83101", name: "Gleis", articleType: "track", subtype: "straight" }] })} />);

    expect(screen.getByRole("textbox", { name: "Hersteller" })).toBeDisabled();
  });

  it("disables resource mutations while stale and offers an explicit retry", async () => {
    const user = userEvent.setup();
    const onRetryResources = vi.fn().mockRejectedValue(new Error("Weiterhin nicht verfügbar"));
    render(<ArticleEditorDialog {...props({
      mode: "edit", article: persistedArticle, activeTab: "stock", resourcesStale: true,
      error: "Bestand konnte nicht aktualisiert werden", onRetryResources
    })} />);

    expect(screen.queryByRole("button", { name: "Reservierung anlegen" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Einbau erfassen" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Daten erneut laden" }));
    expect(onRetryResources).toHaveBeenCalledOnce();
  });
});
