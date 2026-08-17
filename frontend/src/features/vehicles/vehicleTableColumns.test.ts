import { describe, expect, it } from "vitest";

import { vehicleFixture } from "../../test/fixtures/vehicles";
import {
  defaultVehicleTableColumns,
  moveVehicleTableColumn,
  normalizeVehicleTableColumns,
  parseVehicleTableColumns,
  serializeVehicleTableColumns,
	isVehicleDataColumn,
  toggleVehicleTableColumn,
  vehicleColumnSortValue
} from "./vehicleTableColumns";

describe("vehicle table columns", () => {
  it("keeps the current desktop columns as defaults", () => {
    expect(defaultVehicleTableColumns).toEqual([
	  "type",
      "inventoryNumber",
      "manufacturer",
      "articleNumber",
      "name",
      "gauge",
      "epoch",
      "exhibition"
    ]);
  });

	it("treats type as a presentation-only column", () => {
		expect(isVehicleDataColumn("type")).toBe(false);
		expect(parseVehicleTableColumns('["inventoryNumber","series"]'))
			.toEqual(["inventoryNumber", "series"]);
	});

  it("preserves saved order without appending unknown, duplicate, or new keys", () => {
    expect(normalizeVehicleTableColumns([
      "series",
      "manufacturer",
      "futureColumn",
      "series"
    ])).toEqual(["series", "manufacturer"]);
  });

  it("offers operational columns without enabling them by default or in old preferences", () => {
    expect(defaultVehicleTableColumns).not.toContain("maximumSpeedKmh");
    expect(defaultVehicleTableColumns).not.toContain("homeBase");
    expect(parseVehicleTableColumns('["inventoryNumber","series"]'))
      .toEqual(["inventoryNumber", "series"]);
  });

  it("restores inventory number when only presentation columns remain", () => {
    expect(normalizeVehicleTableColumns(["image", "exhibition"]))
      .toEqual(["image", "exhibition", "inventoryNumber"]);
  });

  it("uses defaults for missing, malformed, or non-array settings", () => {
    expect(parseVehicleTableColumns(undefined)).toEqual(defaultVehicleTableColumns);
    expect(parseVehicleTableColumns("not-json")).toEqual(defaultVehicleTableColumns);
    expect(parseVehicleTableColumns("{}")).toEqual(defaultVehicleTableColumns);
  });

  it("toggles, moves, and serializes a normalized ordered list", () => {
    const shown = toggleVehicleTableColumn(defaultVehicleTableColumns, "series");
    const moved = moveVehicleTableColumn(shown, "series", "up");

    expect(moved.at(-2)).toBe("series");
    expect(parseVehicleTableColumns(serializeVehicleTableColumns(moved))).toEqual(moved);
  });

  it("returns stable sortable values for booleans and text", () => {
    const vehicle = vehicleFixture({ digital: true, series: " BR 218 ", maximumSpeedKmh: 120 });

    expect(vehicleColumnSortValue(vehicle, "digital")).toBe("1");
    expect(vehicleColumnSortValue(vehicle, "series")).toBe("br 218");
    expect(vehicleColumnSortValue(vehicle, "maximumSpeedKmh")).toBe("120");
  });
});
