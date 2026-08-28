import { describe, expect, it } from "vitest";

import {
  parseTableColumnLayout,
  serializeTableColumnLayout,
  setTableColumnWidth,
  tableColumnWidth,
  tableMinimumWidth,
  type TableColumnLayoutOptions
} from "./tableColumnLayout";

type Column = "first" | "second" | "third";

const options: TableColumnLayoutOptions<Column> = {
  columnKeys: ["first", "second", "third"],
  defaultColumns: ["first", "second"],
  normalizeColumns(values) {
    const known = new Set<Column>(["first", "second", "third"]);
    const seen = new Set<Column>();
    const columns = [...values].flatMap((value) => {
      if (typeof value !== "string" || !known.has(value as Column) || seen.has(value as Column)) {
        return [];
      }
      seen.add(value as Column);
      return [value as Column];
    });
    if (!columns.length) columns.push("first");
    return columns;
  },
  widthDefinitions: {
    first: { defaultWidth: 120, minWidth: 80, maxWidth: 240 },
    second: { defaultWidth: 160, minWidth: 100, maxWidth: 320 },
    third: { defaultWidth: 90, minWidth: 70, maxWidth: 180 }
  }
};

describe("table column layout", () => {
  it("loads legacy arrays and keeps default widths implicit", () => {
    expect(parseTableColumnLayout('["second","first"]', options)).toEqual({
      columns: ["second", "first"],
      widths: {}
    });
  });

  it("normalizes versioned layouts and retains widths for hidden known columns", () => {
    expect(parseTableColumnLayout(JSON.stringify({
      version: 1,
      columns: ["second", "second", "unknown"],
      widths: { first: 12, second: 241.6, third: 130, unknown: 200, extra: "wide" }
    }), options)).toEqual({
      columns: ["second"],
      widths: { first: 80, second: 242, third: 130 }
    });
  });

  it("falls back safely for malformed values and rejects non-finite widths", () => {
    expect(parseTableColumnLayout("not json", options)).toEqual({
      columns: ["first", "second"],
      widths: {}
    });
    expect(parseTableColumnLayout(JSON.stringify({
      version: 1,
      columns: ["first"],
      widths: { first: null, second: "180", third: Number.POSITIVE_INFINITY }
    }), options)).toEqual({ columns: ["first"], widths: {} });
  });

  it("stores only non-default widths and computes the bounded minimum width", () => {
    const initial = parseTableColumnLayout(undefined, options);
    const resized = setTableColumnWidth(initial, "first", 196.4, options.widthDefinitions);
    const reset = setTableColumnWidth(resized, "second", 160, options.widthDefinitions);

    expect(resized.widths).toEqual({ first: 196 });
    expect(reset.widths).toEqual({ first: 196 });
    expect(tableColumnWidth(resized, "first", options.widthDefinitions)).toBe(196);
    expect(tableColumnWidth(resized, "second", options.widthDefinitions)).toBe(160);
    expect(tableMinimumWidth(resized, options.widthDefinitions, 72)).toBe(428);
    expect(JSON.parse(serializeTableColumnLayout(resized, options))).toEqual({
      version: 1,
      columns: ["first", "second"],
      widths: { first: 196 }
    });
  });
});
