import {
  parseTableColumnLayout,
  serializeTableColumnLayout,
  type TableColumnLayout,
  type TableColumnLayoutOptions,
  type TableColumnWidthDefinition
} from "../../shared/tableColumnLayout";

export const articleTableColumns = [
  "image",
  "inventoryNumber",
  "manufacturer",
  "articleNumber",
  "name",
  "type",
  "gauge",
  "listPrice",
  "stock",
  "storage"
] as const;

export type ArticleTableColumn = typeof articleTableColumns[number];
export type ArticleColumnMove = "up" | "down";

export const articleTableColumnSettingKey = "railkeeper.accessories.tableColumns";

export const defaultArticleTableColumns: ArticleTableColumn[] = [
  "image",
  "inventoryNumber",
  "manufacturer",
  "articleNumber",
  "name",
  "type",
  "gauge",
  "stock",
  "storage"
];

export const articleTableColumnWidthDefinitions: Record<
  ArticleTableColumn,
  TableColumnWidthDefinition
> = {
  image: { defaultWidth: 86, minWidth: 72, maxWidth: 160 },
  inventoryNumber: { defaultWidth: 142, minWidth: 112, maxWidth: 280 },
  manufacturer: { defaultWidth: 150, minWidth: 112, maxWidth: 360 },
  articleNumber: { defaultWidth: 126, minWidth: 104, maxWidth: 280 },
  name: { defaultWidth: 210, minWidth: 140, maxWidth: 480 },
  type: { defaultWidth: 175, minWidth: 128, maxWidth: 360 },
  gauge: { defaultWidth: 80, minWidth: 72, maxWidth: 180 },
  listPrice: { defaultWidth: 130, minWidth: 112, maxWidth: 240 },
  stock: { defaultWidth: 210, minWidth: 156, maxWidth: 360 },
  storage: { defaultWidth: 170, minWidth: 120, maxWidth: 420 }
};

export function isArticleTableColumn(value: unknown): value is ArticleTableColumn {
  return typeof value === "string" && articleTableColumns.some((column) => column === value);
}

export function normalizeArticleTableColumns(values: Iterable<unknown>) {
  const seen = new Set<ArticleTableColumn>();
  const columns = [...values].flatMap((value) => {
    if (!isArticleTableColumn(value) || seen.has(value)) return [];
    seen.add(value);
    return [value];
  });

  if (!columns.includes("inventoryNumber") && !columns.includes("name")) {
    columns.push("inventoryNumber");
  }
  return columns;
}

const articleTableLayoutOptions: TableColumnLayoutOptions<ArticleTableColumn> = {
  columnKeys: articleTableColumns,
  defaultColumns: defaultArticleTableColumns,
  normalizeColumns: normalizeArticleTableColumns,
  widthDefinitions: articleTableColumnWidthDefinitions
};

export const defaultArticleTableLayout: TableColumnLayout<ArticleTableColumn> = {
  columns: [...defaultArticleTableColumns],
  widths: {}
};

export function parseArticleTableLayout(raw: string | undefined) {
  return parseTableColumnLayout(raw, articleTableLayoutOptions);
}

export function serializeArticleTableLayout(layout: TableColumnLayout<ArticleTableColumn>) {
  return serializeTableColumnLayout(layout, articleTableLayoutOptions);
}

type ColumnStorage = Pick<Storage, "getItem" | "setItem">;

export function storedArticleTableColumns(
  storage: Pick<ColumnStorage, "getItem"> = window.localStorage
) {
  return parseArticleTableLayout(storage.getItem(articleTableColumnSettingKey) ?? undefined).columns;
}

export function persistArticleTableColumns(
  columns: readonly ArticleTableColumn[],
  storage: Pick<ColumnStorage, "setItem"> = window.localStorage
) {
  storage.setItem(
    articleTableColumnSettingKey,
    JSON.stringify(normalizeArticleTableColumns(columns))
  );
}

export function resetArticleTableColumns() {
  return [...defaultArticleTableColumns];
}

export function toggleArticleTableColumn(
  columns: readonly ArticleTableColumn[],
  column: ArticleTableColumn
) {
  const next = columns.filter((key) => key !== column);
  if (next.length === columns.length) next.push(column);

  const hidesFinalIdentity = (column === "inventoryNumber" && !next.includes("name")) ||
    (column === "name" && !next.includes("inventoryNumber"));
  return hidesFinalIdentity ? [...columns] : normalizeArticleTableColumns(next);
}

export function moveArticleTableColumn(
  columns: readonly ArticleTableColumn[],
  column: ArticleTableColumn,
  direction: ArticleColumnMove
) {
  const next = normalizeArticleTableColumns(columns);
  const from = next.indexOf(column);
  if (from < 0) return next;

  const to = direction === "up" ? from - 1 : from + 1;
  if (to < 0 || to >= next.length) return next;

  [next[from], next[to]] = [next[to], next[from]];
  return next;
}
