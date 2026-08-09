import { api, type AccessoryArticle } from "../../shared/api";
import type { ArticleEditorResources } from "./useArticleEditorController";

export type ArticleEditorResourcePatch = Partial<ArticleEditorResources>;

export type ArticleEditorResourceResult = {
  patch: ArticleEditorResourcePatch;
  errors: Error[];
};

function asError(reason: unknown): Error {
  return reason instanceof Error ? reason : new Error("Die Artikeldaten konnten nicht vollständig geladen werden.");
}

function include<T, K extends keyof ArticleEditorResources>(
  result: PromiseSettledResult<T>,
  key: K,
  patch: ArticleEditorResourcePatch,
  errors: Error[]
) {
  if (result.status === "fulfilled") patch[key] = result.value as ArticleEditorResources[K];
  else errors.push(asError(result.reason));
}

export async function fetchArticleEditorResourcePatch(
  article: Pick<AccessoryArticle, "id" | "inventoryStrategy">
): Promise<ArticleEditorResourceResult> {
  const patch: ArticleEditorResourcePatch = {};
  const errors: Error[] = [];
  const shared = await Promise.allSettled([
    api.storageLocations(), api.vehicles(), api.layouts()
  ] as const);
  include(shared[0], "locations", patch, errors);
  include(shared[1], "vehicles", patch, errors);
  include(shared[2], "layouts", patch, errors);

  if (shared[2].status === "fulfilled") {
    const unitResults = await Promise.allSettled(shared[2].value.map((layout) => api.layoutUnits(layout.id)));
    const failedUnits = unitResults.filter((result) => result.status === "rejected");
    if (failedUnits.length === 0) {
      patch.units = unitResults.flatMap((result) => result.status === "fulfilled" ? result.value : []);
    } else {
      failedUnits.forEach((result) => {
        if (result.status === "rejected") errors.push(asError(result.reason));
      });
    }
  }

  const assets = article.inventoryStrategy === "quantity"
    ? Promise.resolve([])
    : api.accessoryAssets(article.id);
  const related = await Promise.allSettled([
    api.accessoryStock(article.id),
    api.accessoryStockMovements(article.id),
    assets,
    api.accessoryPurchases(article.id),
    api.accessoryDocuments(article.id),
    api.accessoryReservations(article.id),
    api.accessoryInstallations(article.id),
    api.accessoryUsageHistory(article.id)
  ] as const);
  include(related[0], "stock", patch, errors);
  include(related[1], "movements", patch, errors);
  include(related[2], "assets", patch, errors);
  include(related[3], "purchases", patch, errors);
  include(related[4], "documents", patch, errors);
  if (related[4].status === "fulfilled") patch.documentsLoaded = true;
  include(related[5], "reservations", patch, errors);
  include(related[6], "installations", patch, errors);
  include(related[7], "usageHistory", patch, errors);
  return { patch, errors };
}
