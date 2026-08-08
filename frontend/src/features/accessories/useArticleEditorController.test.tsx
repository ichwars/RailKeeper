import { StrictMode, type PropsWithChildren } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type AccessoryArticle, type AccessoryStockSummary, type MasterDataEntry } from "../../shared/api";
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

const otherArticle: AccessoryArticle = {
  ...article,
  id: "article-2",
  manufacturer: "Viessmann",
  articleNumber: "4011",
  name: "Signal"
};

const historicalOtherArticle: AccessoryArticle = {
  ...article,
  id: "article-historical",
  articleType: "other",
  subtype: "legacy",
  attributes: [{ key: "legacyMaterial", kind: "text", textValue: "Holz" }]
};

const stock = (productId: string, totalQuantity: number): AccessoryStockSummary => ({
  productId,
  trackingMode: "quantity",
  totalQuantity,
  locations: []
});
const productionSubtype: MasterDataEntry = {
  id: "article-subtype-track-straight", type: "accessory_subtype", key: "track:straight", label: "Straight",
  active: true, sortOrder: 10, metadata: {}, createdAt: "2026-08-08T08:00:00Z",
  updatedAt: "2026-08-08T08:00:00Z"
};
const productionArticleType: MasterDataEntry = {
  id: "article-type-track", type: "article_type", key: "track", label: "Gleismaterial",
  active: false, sortOrder: 10, metadata: {}, createdAt: "2026-08-08T08:00:00Z",
  updatedAt: "2026-08-08T08:00:00Z"
};

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason: unknown) => void;
  const promise = new Promise<T>((next, fail) => { resolve = next; reject = fail; });
  return { promise, resolve, reject };
}

describe("useArticleEditorController", () => {
  beforeEach(() => {
    vi.spyOn(api, "accessoryArticle").mockResolvedValue(article);
    vi.spyOn(api, "checkAccessoryArticleDuplicates").mockResolvedValue({ candidates: [] });
    vi.spyOn(api, "createAccessoryArticle").mockResolvedValue(article);
    vi.spyOn(api, "updateAccessoryArticle").mockResolvedValue(article);
    vi.spyOn(api, "storageLocations").mockResolvedValue([]);
    vi.spyOn(api, "masterData").mockResolvedValue([]);
    vi.spyOn(api, "vehicles").mockResolvedValue([]);
    vi.spyOn(api, "layouts").mockResolvedValue([]);
    vi.spyOn(api, "layoutUnits").mockResolvedValue([]);
    vi.spyOn(api, "accessoryStock").mockResolvedValue(stock(article.id, 0));
    vi.spyOn(api, "accessoryStockMovements").mockResolvedValue([]);
    vi.spyOn(api, "accessoryAssets").mockResolvedValue([]);
    vi.spyOn(api, "accessoryPurchases").mockResolvedValue([]);
    vi.spyOn(api, "accessoryDocuments").mockResolvedValue([]);
    vi.spyOn(api, "accessoryReservations").mockResolvedValue([]);
    vi.spyOn(api, "accessoryInstallations").mockResolvedValue([]);
    vi.spyOn(api, "accessoryUsageHistory").mockResolvedValue({ productId: article.id, events: [] });
  });

  it("validates the whole form, marks tabs, and navigates to the first invalid tab", async () => {
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openCreate());
    await waitFor(() => expect(result.current.customFieldsLoading).toBe(false));
    act(() => result.current.changeForm({ minimumStock: "-1" }));

    await act(async () => result.current.submit());

    expect(result.current.tabErrors).toEqual({ article: true, stock: true });
    expect(result.current.activeTab).toBe("article");
    expect(api.createAccessoryArticle).not.toHaveBeenCalled();
  });

  it("loads production-normalized accessory subtype master data and retries failures", async () => {
    vi.mocked(api.masterData)
      .mockImplementation(async (type) => type === "accessory_subtype" ? [productionSubtype] : []);
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openCreate());

    await waitFor(() => expect(result.current.subtypeEntriesLoading).toBe(false));
    expect(api.masterData).toHaveBeenCalledWith("accessory_subtype");
    expect(result.current.subtypeEntries).toEqual([productionSubtype]);

    vi.mocked(api.masterData).mockRejectedValueOnce(new Error("offline"));
    await act(async () => result.current.retrySubtypeEntries());
    expect(result.current.subtypeEntriesError).toBe("Unterarten konnten nicht geladen werden.");

    vi.mocked(api.masterData).mockResolvedValueOnce([productionSubtype]);
    await act(async () => result.current.retrySubtypeEntries());
    expect(result.current.subtypeEntriesError).toBe("");
    expect(result.current.subtypeEntries).toEqual([productionSubtype]);
  });

  it("loads article types including inactive entries and retries failures", async () => {
    vi.mocked(api.masterData)
      .mockImplementation(async (type) => type === "article_type" ? [productionArticleType] : []);
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openCreate());

    await waitFor(() => expect(result.current.articleTypeEntriesLoading).toBe(false));
    expect(api.masterData).toHaveBeenCalledWith("article_type");
    expect(result.current.articleTypeEntries).toEqual([productionArticleType]);

    vi.mocked(api.masterData).mockRejectedValueOnce(new Error("offline"));
    await act(async () => result.current.retryArticleTypeEntries());
    expect(result.current.articleTypeEntriesError).toBe("Artikelarten konnten nicht geladen werden.");

    vi.mocked(api.masterData).mockResolvedValueOnce([productionArticleType]);
    await act(async () => result.current.retryArticleTypeEntries());
    expect(result.current.articleTypeEntriesError).toBe("");
    expect(result.current.articleTypeEntries).toEqual([productionArticleType]);
  });

  it("does not request asset resources or stale the editor for a quantity article", async () => {
    vi.mocked(api.accessoryArticle).mockResolvedValueOnce({
      ...article,
      trackingMode: "quantity",
      inventoryStrategy: "quantity"
    });
    vi.mocked(api.accessoryAssets).mockRejectedValueOnce(new Error("Operation invalid for tracking mode"));
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));

    act(() => result.current.openArticle("article-1", "edit", false));

    await waitFor(() => expect(result.current.resources.stock).toEqual(stock("article-1", 0)));
    expect(api.accessoryAssets).not.toHaveBeenCalled();
    expect(result.current.resourcesStale).toBe(false);
    expect(result.current.resourceError).toBe("");
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
    await waitFor(() => expect(result.current.customFieldsLoading).toBe(false));
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

  it("preserves and resubmits hidden historical custom attributes after a failed edit save", async () => {
    vi.mocked(api.accessoryArticle).mockResolvedValueOnce(historicalOtherArticle);
    vi.mocked(api.updateAccessoryArticle).mockRejectedValueOnce(new Error("Speichern fehlgeschlagen"));
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openArticle(historicalOtherArticle.id, "edit", false));
    await waitFor(() => expect(result.current.article?.id).toBe(historicalOtherArticle.id));

    await act(async () => result.current.submit());

    expect(api.updateAccessoryArticle).toHaveBeenCalledWith(historicalOtherArticle.id,
      expect.objectContaining({ attributes: historicalOtherArticle.attributes }));
    expect(result.current.form.attributes).toEqual(historicalOtherArticle.attributes);
    expect(result.current.error).toBe("Speichern fehlgeschlagen");
    expect(result.current.isOpen).toBe(true);
  });

  it("does not grant historical custom ownership to attributes from a standard-type snapshot", async () => {
    const standardWithAttribute = { ...article, attributes: [
      { key: "trackSystem", kind: "text" as const, textValue: "Tillig TT Modellgleis" }
    ] };
    vi.mocked(api.accessoryArticle).mockResolvedValueOnce(standardWithAttribute);
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openArticle(standardWithAttribute.id, "edit", false));
    await waitFor(() => expect(result.current.article?.id).toBe(standardWithAttribute.id));
    await waitFor(() => expect(result.current.customFieldsLoading).toBe(false));
    act(() => result.current.changeForm({ articleType: "other", subtype: "legacy",
      attributes: standardWithAttribute.attributes }));

    await act(async () => result.current.submit());

    expect(api.updateAccessoryArticle).not.toHaveBeenCalled();
    expect(result.current.subjectFieldErrors.trackSystem).toEqual(expect.any(String));
  });

  it("owns custom-field load failure, blocks only other save, and preserves drafts across retry", async () => {
    vi.mocked(api.masterData)
      .mockRejectedValueOnce(new Error("Konfiguration nicht verfügbar"))
      .mockResolvedValueOnce([]);
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openCreate());
    act(() => result.current.changeForm({
      manufacturer: "Tillig", name: "Entwurf", subtype: "sonstig", internalNotes: "Bleibt erhalten"
    }));
    await waitFor(() => expect(result.current.customFieldsError).not.toBe(""));

    await act(async () => result.current.submit());
    expect(api.createAccessoryArticle).not.toHaveBeenCalled();
    expect(result.current.customFieldsError).not.toBe("");

    await act(async () => result.current.retryCustomFields());
    expect(result.current.customFieldsError).toBe("");
    expect(result.current.form.internalNotes).toBe("Bleibt erhalten");

    act(() => result.current.changeForm({ articleType: "track", subtype: "straight" }));
    await act(async () => result.current.submit());
    expect(api.createAccessoryArticle).toHaveBeenCalledOnce();
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

  it("ignores slow detail results from an older article request", async () => {
    const first = deferred<AccessoryArticle>();
    const second = deferred<AccessoryArticle>();
    vi.mocked(api.accessoryArticle)
      .mockImplementationOnce(() => first.promise)
      .mockImplementationOnce(() => second.promise);
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));

    act(() => result.current.openArticle("article-1", "edit", false));
    act(() => result.current.openArticle("article-2", "edit", false));
    await act(async () => second.resolve(otherArticle));
    await waitFor(() => expect(result.current.article?.id).toBe("article-2"));
    await act(async () => first.resolve(article));

    expect(result.current.article?.id).toBe("article-2");
    expect(result.current.form.name).toBe("Signal");
  });

  it("invalidates slow detail results when closing or starting create", async () => {
    const closing = deferred<AccessoryArticle>();
    const creating = deferred<AccessoryArticle>();
    vi.mocked(api.accessoryArticle)
      .mockImplementationOnce(() => closing.promise)
      .mockImplementationOnce(() => creating.promise);
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));

    act(() => result.current.openArticle("article-1", "edit", false));
    act(() => result.current.requestClose());
    await act(async () => closing.resolve(article));
    expect(result.current.isOpen).toBe(false);
    expect(result.current.article).toBeNull();

    act(() => result.current.openArticle("article-1", "edit", false));
    act(() => result.current.openCreate());
    await act(async () => creating.resolve(article));
    expect(result.current.mode).toBe("create");
    expect(result.current.article).toBeNull();
    expect(result.current.form).toEqual(emptyArticleEditorForm());
  });

  it("never overwrites the current article resources with an older session", async () => {
    const oldStock = deferred<AccessoryStockSummary>();
    vi.mocked(api.accessoryStock).mockImplementation((id) => id === "article-1"
      ? oldStock.promise : Promise.resolve(stock(id, 7)));
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));

    act(() => result.current.openArticle("article-1", "edit", false));
    await waitFor(() => expect(api.accessoryStock).toHaveBeenCalledWith("article-1"));
    act(() => result.current.openArticle("article-2", "edit", false));
    await waitFor(() => expect(result.current.resources.stock?.productId).toBe("article-2"));
    await act(async () => oldStock.resolve(stock("article-1", 99)));

    expect(result.current.resources.stock).toEqual(stock("article-2", 7));
  });

  it("blocks edit save after detail failure instead of falling back to create", async () => {
    vi.mocked(api.accessoryArticle).mockRejectedValueOnce(new Error("Detail nicht verfügbar"));
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));

    act(() => result.current.openArticle("missing", "edit", false));
    expect(result.current.form).toEqual(emptyArticleEditorForm());
    await waitFor(() => expect(result.current.error).toBe("Detail nicht verfügbar"));
    act(() => result.current.changeForm({ manufacturer: "Tillig", name: "Gleis", subtype: "straight" }));
    await act(async () => result.current.submit());

    expect(api.createAccessoryArticle).not.toHaveBeenCalled();
    expect(api.updateAccessoryArticle).not.toHaveBeenCalled();
    expect(result.current.isOpen).toBe(true);
  });

  it("preserves successful resource data and rejects a partial refresh failure", async () => {
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openArticle("article-1", "edit", false));
    await waitFor(() => expect(result.current.resources.stock).toEqual(stock("article-1", 0)));
    vi.mocked(api.accessoryStock).mockRejectedValueOnce(new Error("Bestand nicht verfügbar"));

    let refreshError: unknown;
    await act(async () => {
      try { await result.current.refreshResources(); } catch (reason) { refreshError = reason; }
    });
    expect(refreshError).toEqual(new Error("Bestand nicht verfügbar"));
    expect(result.current.resources.stock).toEqual(stock("article-1", 0));
    expect(result.current.error).toBe("Bestand nicht verfügbar");
    expect(result.current.resourcesStale).toBe(true);
  });

  it("keeps a mutation refresh newer than the still-pending initial resource load", async () => {
    const initialStock = deferred<AccessoryStockSummary>();
    vi.mocked(api.accessoryStock)
      .mockImplementationOnce(() => initialStock.promise)
      .mockResolvedValueOnce(stock("article-1", 8));
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));

    act(() => result.current.openArticle("article-1", "edit", false));
    await waitFor(() => expect(api.accessoryStock).toHaveBeenCalledTimes(1));
    await act(async () => result.current.refreshResources());
    expect(result.current.resources.stock).toEqual(stock("article-1", 8));

    await act(async () => initialStock.resolve(stock("article-1", 1)));
    expect(result.current.resources.stock).toEqual(stock("article-1", 8));
  });

  it("applies only the latest of two same-session resource refreshes", async () => {
    const firstRefresh = deferred<AccessoryStockSummary>();
    const secondRefresh = deferred<AccessoryStockSummary>();
    vi.mocked(api.accessoryStock)
      .mockResolvedValueOnce(stock("article-1", 0))
      .mockImplementationOnce(() => firstRefresh.promise)
      .mockImplementationOnce(() => secondRefresh.promise);
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));

    act(() => result.current.openArticle("article-1", "edit", false));
    await waitFor(() => expect(result.current.resources.stock).toEqual(stock("article-1", 0)));
    let older!: Promise<void>;
    let newer!: Promise<void>;
    act(() => { older = result.current.refreshResources(); });
    await waitFor(() => expect(api.accessoryStock).toHaveBeenCalledTimes(2));
    act(() => { newer = result.current.refreshResources(); });
    await waitFor(() => expect(api.accessoryStock).toHaveBeenCalledTimes(3));

    await act(async () => secondRefresh.resolve(stock("article-1", 9)));
    await newer;
    await act(async () => firstRefresh.resolve(stock("article-1", 2)));
    await older;

    expect(result.current.resources.stock).toEqual(stock("article-1", 9));
  });

  it("clears stale resource state only after an explicit successful retry", async () => {
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openArticle("article-1", "edit", false));
    await waitFor(() => expect(result.current.resources.stock).toEqual(stock("article-1", 0)));
    vi.mocked(api.accessoryStock).mockRejectedValueOnce(new Error("Bestand veraltet"));

    await act(async () => {
      try { await result.current.refreshResources(); } catch { /* expected */ }
    });
    expect(result.current.resourcesStale).toBe(true);

    await act(async () => result.current.retryResources());
    expect(result.current.resourcesStale).toBe(false);
    expect(result.current.error).toBe("");
  });

  it("keeps a later save error visible when retry clears only the stale resource error", async () => {
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openArticle("article-1", "edit", false));
    await waitFor(() => expect(result.current.resources.stock).toEqual(stock("article-1", 0)));
    vi.mocked(api.accessoryStock).mockRejectedValueOnce(new Error("Bestand veraltet"));
    await act(async () => {
      try { await result.current.refreshResources(); } catch { /* expected */ }
    });
    expect(result.current.error).toBe("Bestand veraltet");
    expect(result.current.resourceError).toBe("Bestand veraltet");
    expect(result.current.resourcesStale).toBe(true);

    vi.mocked(api.updateAccessoryArticle).mockRejectedValueOnce(new Error("Speichern fehlgeschlagen"));
    await act(async () => result.current.submit());
    expect(result.current.error).toBe("Speichern fehlgeschlagen");
    expect(result.current.resourceError).toBe("Bestand veraltet");

    await act(async () => result.current.retryResources());

    expect(result.current.resourcesStale).toBe(false);
    expect(result.current.resourceError).toBe("");
    expect(result.current.error).toBe("Speichern fehlgeschlagen");
  });

  it("binds duplicate confirmation to the checked immutable draft", async () => {
    const duplicateCheck = deferred<{ candidates: Array<{
      id: string; manufacturer: string; articleNumber: string; name: string; articleType: "track"; subtype: string;
    }> }>();
    vi.mocked(api.checkAccessoryArticleDuplicates).mockReturnValueOnce(duplicateCheck.promise);
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openCreate());
    act(() => result.current.changeForm({ manufacturer: "Tillig", articleNumber: "83101", name: "Gleis",
      articleType: "track", subtype: "straight" }));

    let submitPromise!: Promise<void>;
    act(() => { submitPromise = result.current.submit(); });
    await waitFor(() => expect(result.current.saving).toBe(true));
    act(() => result.current.changeForm({ name: "Ungeprüfter Name" }));
    expect(result.current.form.name).toBe("Gleis");
    await act(async () => duplicateCheck.resolve({ candidates: [{ id: "dup", manufacturer: "Tillig",
      articleNumber: "83101", name: "Alt", articleType: "track", subtype: "straight" }] }));
    await submitPromise;
    await act(async () => result.current.confirmDuplicateSave());

    expect(api.createAccessoryArticle).toHaveBeenCalledWith(expect.objectContaining({ name: "Gleis" }));
  });

  it("clears field and tab validation markers as soon as the value is corrected", async () => {
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openCreate());
    await waitFor(() => expect(result.current.customFieldsLoading).toBe(false));
    await act(async () => result.current.submit());
    expect(result.current.tabErrors.article).toBe(true);

    act(() => result.current.changeForm({ manufacturer: "Tillig", name: "Gleis", subtype: "straight" }));

    expect(result.current.fieldErrors.manufacturer).toBeUndefined();
    expect(result.current.tabErrors.article).toBeUndefined();
  });

  it("retains actionable errors for remaining invalid subject fields after one field is edited", async () => {
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    act(() => result.current.openCreate());
    act(() => result.current.changeForm({
      manufacturer: "Tillig",
      name: "Gleis",
      articleType: "track",
      subtype: "straight",
      attributes: [{ key: "direction", kind: "single_select", optionValues: ["up"] }],
      attributeNumberDrafts: { connectionCount: "1.5" }
    }));
    await act(async () => result.current.submit());
    expect(result.current.subjectFieldErrors).toMatchObject({
      direction: expect.any(String),
      connectionCount: expect.any(String)
    });

    act(() => result.current.changeForm({
      attributes: [{ key: "direction", kind: "single_select", optionValues: ["left"] }]
    }));

    expect(result.current.subjectFieldErrors.direction).toBeUndefined();
    expect(result.current.subjectFieldErrors.connectionCount).toEqual(expect.any(String));
    expect(result.current.fieldErrors.attributes).toEqual(expect.any(String));
    expect(result.current.tabErrors.subject).toBe(true);
  });

  it("treats non-empty tab subdrafts as dirty until their successful reset", async () => {
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    const markDirty = result.current.setSubdraftDirty;
    expect(markDirty).toBeTypeOf("function");
    if (!markDirty) return;
    act(() => result.current.openArticle("article-1", "edit", false));
    await waitFor(() => expect(result.current.article).not.toBeNull());

    act(() => markDirty("purchase", true));
    act(() => result.current.requestClose());
    expect(result.current.closeConfirmationOpen).toBe(true);

    act(() => result.current.cancelClose());
    act(() => markDirty("purchase", false));
    act(() => result.current.requestClose());
    expect(result.current.isOpen).toBe(false);
  });

  it("accepts current detail results after the React StrictMode effect replay", async () => {
    const wrapper = ({ children }: PropsWithChildren) => <StrictMode>{children}</StrictMode>;
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }), { wrapper });

    act(() => result.current.openArticle("article-1", "edit", false));

    await waitFor(() => expect(result.current.article?.id).toBe("article-1"));
  });

  it("advances the dialog session key for every new create or article request", () => {
    const { result } = renderHook(() => useArticleEditorController({ roles: ["Editor"] }));
    const initial = result.current.sessionKey;
    expect(initial).toBeTypeOf("number");

    act(() => result.current.openCreate());
    const createSession = result.current.sessionKey;
    act(() => result.current.openArticle("article-1", "edit", false));

    expect(createSession).toBeGreaterThan(initial);
    expect(result.current.sessionKey).toBeGreaterThan(createSession);
  });
});
