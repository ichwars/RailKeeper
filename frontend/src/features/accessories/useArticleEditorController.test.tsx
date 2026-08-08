import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type AccessoryArticle } from "../../shared/api";
import { emptyArticleEditorForm } from "./articleEditorModel";
import { useArticleEditorController } from "./useArticleEditorController";

const article: AccessoryArticle = {
  id: "article-1",
  manufacturer: "Tillig",
  articleNumber: "83101",
  name: "Gerades Modellgleis",
  category: "straight",
  trackingMode: "quantity",
  manufacturerStatus: "available",
  articleType: "track",
  subtype: "straight",
  gauges: ["TT"],
  packageQuantity: 1,
  stockUnit: "piece",
  minimumStock: 4,
  inventoryStrategy: "quantity_later_individual",
  alternativeNumbers: [],
  keywords: [],
  archived: false,
  attributes: [],
  createdAt: "2026-08-08T08:00:00Z",
  updatedAt: "2026-08-08T09:00:00Z"
};

describe("useArticleEditorController", () => {
  beforeEach(() => {
    vi.spyOn(api, "accessoryArticle").mockResolvedValue(article);
    vi.spyOn(api, "checkAccessoryArticleDuplicates").mockResolvedValue({ candidates: [] });
    vi.spyOn(api, "createAccessoryArticle").mockResolvedValue(article);
    vi.spyOn(api, "updateAccessoryArticle").mockResolvedValue(article);
  });

  it("validates the whole form, marks tabs, and navigates to the first invalid tab", async () => {
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openCreate());
    act(() => result.current.changeForm({ minimumStock: "-1" }));

    await act(async () => result.current.submit());

    expect(result.current.tabErrors).toEqual({ article: true, stock: true });
    expect(result.current.activeTab).toBe("article");
    expect(api.createAccessoryArticle).not.toHaveBeenCalled();
  });

  it("requires a deliberate duplicate confirmation before one create command", async () => {
    vi.mocked(api.checkAccessoryArticleDuplicates).mockResolvedValueOnce({ candidates: [{
      id: "duplicate-1",
      manufacturer: "Tillig",
      articleNumber: "83101",
      name: "Vorhandenes Gleis",
      articleType: "track",
      subtype: "straight"
    }] });
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openCreate());
    act(() => result.current.changeForm({
      ...emptyArticleEditorForm(),
      manufacturer: "Tillig",
      articleNumber: "83101",
      name: "Gerades Modellgleis",
      subtype: "straight"
    }));

    await act(async () => result.current.submit());
    expect(result.current.duplicateCandidates).toHaveLength(1);
    expect(api.createAccessoryArticle).not.toHaveBeenCalled();

    await act(async () => result.current.confirmDuplicateSave());
    expect(api.createAccessoryArticle).toHaveBeenCalledOnce();
  });

  it("preserves every value and exposes the error after a failed save", async () => {
    vi.mocked(api.createAccessoryArticle).mockRejectedValueOnce(new Error("Netzwerkfehler"));
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openCreate());
    act(() => result.current.changeForm({
      manufacturer: "Viessmann",
      articleNumber: "4011",
      name: "Signal",
      articleType: "signal",
      subtype: "main_signal",
      internalNotes: "Wert aus unmontiertem Reiter"
    }));

    await act(async () => result.current.submit());

    expect(result.current.isOpen).toBe(true);
    expect(result.current.form.internalNotes).toBe("Wert aus unmontiertem Reiter");
    expect(result.current.error).toBe("Netzwerkfehler");
  });

  it("loads edit data and asks before closing only when values really changed", async () => {
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openArticle("article-1", "edit", true));
    await waitFor(() => expect(result.current.article?.id).toBe("article-1"));

    act(() => result.current.requestClose());
    expect(result.current.isOpen).toBe(false);

    act(() => result.current.openArticle("article-1", "edit", true));
    await waitFor(() => expect(result.current.article?.id).toBe("article-1"));
    act(() => result.current.changeForm({ name: "Geändert" }));
    act(() => result.current.requestClose());
    expect(result.current.closeConfirmationOpen).toBe(true);
    expect(result.current.isOpen).toBe(true);
  });

  it("enforces editor, planner, viewer, and view-mode permissions", async () => {
    const editor = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    const planner = renderHook(() => useArticleEditorController({ roles: ["Planner"] }));
    const viewer = renderHook(() => useArticleEditorController({ roles: ["Viewer"] }));

    expect(editor.result.current.permissions).toEqual({ canEdit: true, canManageStock: true, canReserve: true, canInstall: true });
    expect(planner.result.current.permissions).toEqual({ canEdit: false, canManageStock: false, canReserve: true, canInstall: false });
    expect(viewer.result.current.permissions).toEqual({ canEdit: false, canManageStock: false, canReserve: false, canInstall: false });

    act(() => editor.result.current.openArticle("article-1", "view", false));
    await waitFor(() => expect(editor.result.current.article).not.toBeNull());
    expect(editor.result.current.isFormReadOnly).toBe(true);
  });
});
