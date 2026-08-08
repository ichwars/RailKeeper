import { api } from "../../shared/api";
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

export async function fetchArticleEditorResourcePatch(articleId: string): Promise<ArticleEditorResourceResult> {
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

  const related = await Promise.allSettled([
    api.accessoryStock(articleId),
    api.accessoryStockMovements(articleId),
    api.accessoryAssets(articleId),
    api.accessoryPurchases(articleId),
    api.accessoryDocuments(articleId),
    api.accessoryReservations(articleId),
    api.accessoryInstallations(articleId),
    api.accessoryUsageHistory(articleId)
  ] as const);
  include(related[0], "stock", patch, errors);
  include(related[1], "movements", patch, errors);
  include(related[2], "assets", patch, errors);
  include(related[3], "purchases", patch, errors);
  include(related[4], "documents", patch, errors);
  include(related[5], "reservations", patch, errors);
  include(related[6], "installations", patch, errors);
  include(related[7], "usageHistory", patch, errors);
  return { patch, errors };
}
