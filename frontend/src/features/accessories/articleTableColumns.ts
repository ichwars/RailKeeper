export const articleTableColumns = [
  "image",
  "inventoryNumber",
  "manufacturer",
  "articleNumber",
  "name",
  "type",
  "gauge",
  "stock",
  "storage"
] as const;

export type ArticleTableColumn = typeof articleTableColumns[number];

export const articleTableColumnSettingKey = "railkeeper.accessories.tableColumns";
export const defaultArticleTableColumns = new Set<ArticleTableColumn>(articleTableColumns);

export function resetArticleTableColumns() {
  return new Set<ArticleTableColumn>(articleTableColumns);
}

type ColumnStorage = Pick<Storage, "getItem" | "setItem">;

function isArticleTableColumn(value: unknown): value is ArticleTableColumn {
  return typeof value === "string" && articleTableColumns.some((column) => column === value);
}

export function normalizeArticleTableColumns(values: Iterable<unknown>) {
  const requested = new Set([...values].filter(isArticleTableColumn));
  if (!requested.has("inventoryNumber") && !requested.has("name")) {
    requested.add("inventoryNumber");
  }
  return new Set(articleTableColumns.filter((column) => requested.has(column)));
}

export function storedArticleTableColumns(
  storage: Pick<ColumnStorage, "getItem"> = window.localStorage
) {
  const raw = storage.getItem(articleTableColumnSettingKey);
  if (!raw) return new Set<ArticleTableColumn>(articleTableColumns);

  try {
    const parsed: unknown = JSON.parse(raw);
    return Array.isArray(parsed)
      ? normalizeArticleTableColumns(parsed)
      : new Set<ArticleTableColumn>(articleTableColumns);
  } catch {
    return new Set<ArticleTableColumn>(articleTableColumns);
  }
}

export function persistArticleTableColumns(
  columns: ReadonlySet<ArticleTableColumn>,
  storage: Pick<ColumnStorage, "setItem"> = window.localStorage
) {
  const normalized = normalizeArticleTableColumns(columns);
  storage.setItem(articleTableColumnSettingKey, JSON.stringify([...normalized]));
}

export function toggleArticleTableColumn(
  columns: ReadonlySet<ArticleTableColumn>,
  column: ArticleTableColumn
) {
  const next = new Set(columns);
  if (!next.has(column)) {
    next.add(column);
    return normalizeArticleTableColumns(next);
  }

  const hidesFinalIdentity = (column === "inventoryNumber" && !next.has("name")) ||
    (column === "name" && !next.has("inventoryNumber"));
  if (hidesFinalIdentity) return new Set(columns);

  next.delete(column);
  return normalizeArticleTableColumns(next);
}
