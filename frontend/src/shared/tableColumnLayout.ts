export type TableColumnWidthDefinition = {
  defaultWidth: number;
  minWidth: number;
  maxWidth: number;
};

export type TableColumnWidths<Column extends string> = Partial<Record<Column, number>>;

export type TableColumnLayout<Column extends string> = {
  columns: Column[];
  widths: TableColumnWidths<Column>;
};

export type TableColumnLayoutOptions<Column extends string> = {
  columnKeys: readonly Column[];
  defaultColumns: readonly Column[];
  normalizeColumns: (values: Iterable<unknown>) => Column[];
  widthDefinitions: Record<Column, TableColumnWidthDefinition>;
};

type StoredTableColumnLayout = {
  version?: unknown;
  columns?: unknown;
  widths?: unknown;
};

function defaultLayout<Column extends string>(
  options: TableColumnLayoutOptions<Column>
): TableColumnLayout<Column> {
  return { columns: [...options.defaultColumns], widths: {} };
}

function normalizedWidth(value: unknown, definition: TableColumnWidthDefinition) {
  if (typeof value !== "number" || !Number.isFinite(value)) return undefined;
  return Math.min(definition.maxWidth, Math.max(definition.minWidth, Math.round(value)));
}

export function normalizeTableColumnWidths<Column extends string>(
  values: unknown,
  columnKeys: readonly Column[],
  definitions: Record<Column, TableColumnWidthDefinition>
): TableColumnWidths<Column> {
  if (!values || typeof values !== "object" || Array.isArray(values)) return {};
  const record = values as Record<string, unknown>;
  const widths: TableColumnWidths<Column> = {};
  for (const column of columnKeys) {
    const width = normalizedWidth(record[column], definitions[column]);
    if (width !== undefined && width !== definitions[column].defaultWidth) {
      widths[column] = width;
    }
  }
  return widths;
}

export function parseTableColumnLayout<Column extends string>(
  raw: string | undefined,
  options: TableColumnLayoutOptions<Column>
): TableColumnLayout<Column> {
  if (!raw) return defaultLayout(options);

  try {
    const parsed: unknown = JSON.parse(raw);
    if (Array.isArray(parsed)) {
      return { columns: options.normalizeColumns(parsed), widths: {} };
    }
    if (!parsed || typeof parsed !== "object") return defaultLayout(options);

    const stored = parsed as StoredTableColumnLayout;
    if (stored.version !== 1 || !Array.isArray(stored.columns)) return defaultLayout(options);
    return {
      columns: options.normalizeColumns(stored.columns),
      widths: normalizeTableColumnWidths(
        stored.widths,
        options.columnKeys,
        options.widthDefinitions
      )
    };
  } catch {
    return defaultLayout(options);
  }
}

export function serializeTableColumnLayout<Column extends string>(
  layout: TableColumnLayout<Column>,
  options: TableColumnLayoutOptions<Column>
) {
  return JSON.stringify({
    version: 1,
    columns: options.normalizeColumns(layout.columns),
    widths: normalizeTableColumnWidths(
      layout.widths,
      options.columnKeys,
      options.widthDefinitions
    )
  });
}

export function tableColumnWidth<Column extends string>(
  layout: TableColumnLayout<Column>,
  column: Column,
  definitions: Record<Column, TableColumnWidthDefinition>
) {
  return layout.widths[column] ?? definitions[column].defaultWidth;
}

export function setTableColumnWidth<Column extends string>(
  layout: TableColumnLayout<Column>,
  column: Column,
  value: number,
  definitions: Record<Column, TableColumnWidthDefinition>
): TableColumnLayout<Column> {
  const width = normalizedWidth(value, definitions[column]) ?? definitions[column].defaultWidth;
  const widths = { ...layout.widths };
  if (width === definitions[column].defaultWidth) delete widths[column];
  else widths[column] = width;
  return { columns: [...layout.columns], widths };
}

export function tableMinimumWidth<Column extends string>(
  layout: TableColumnLayout<Column>,
  definitions: Record<Column, TableColumnWidthDefinition>,
  fixedWidth: number
) {
  return layout.columns.reduce(
    (total, column) => total + tableColumnWidth(layout, column, definitions),
    fixedWidth
  );
}
