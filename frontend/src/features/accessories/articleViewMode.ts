export type ArticleViewMode = "table" | "cards";

export const articleViewSettingKey = "railkeeper.accessories.view";

type ViewStorage = Pick<Storage, "getItem" | "setItem">;

export function storedArticleViewMode(
  storage: ViewStorage = window.localStorage
): ArticleViewMode {
  return storage.getItem(articleViewSettingKey) === "cards" ? "cards" : "table";
}

export function persistArticleViewMode(
  mode: ArticleViewMode,
  storage: ViewStorage = window.localStorage
) {
  storage.setItem(articleViewSettingKey, mode);
}
