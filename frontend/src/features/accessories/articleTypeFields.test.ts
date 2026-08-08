import { describe, expect, it } from "vitest";

import type { AccessoryArticleType, MasterDataEntry } from "../../shared/api";
import {
  articleEditorWriteInput,
  articlePurchaseWriteInput,
  emptyArticleEditorForm
} from "./articleEditorModel";
import {
  articleTypeFieldRegistry,
  compatibleAttributesForType,
  customFieldDefinitions,
  subjectValuesAreValid
} from "./articleTypeFields";

const expectedFields: Record<Exclude<AccessoryArticleType, "other">, Record<string, string>> = {
  track: {
    trackSystem: "text", lengthMm: "number", radiusMm: "number", angleDegrees: "number",
    direction: "single_select", frogAngleDegrees: "number", sleeperType: "text",
    railHeightMm: "number", roadbed: "boolean", connectionCount: "number", digitalReady: "boolean"
  },
  signal: {
    prototype: "text", epoch: "multi_select", aspects: "multi_select", ledCount: "number",
    heightMm: "number", voltageAC: "number", voltageDC: "number", mounting: "single_select",
    driveType: "single_select", integratedDecoder: "boolean", controlModule: "text"
  },
  decoder: {
    interface: "single_select", protocols: "multi_select", functionOutputs: "number",
    motorCurrentMa: "number", outputCurrentMa: "number", totalCurrentMa: "number",
    railCom: "boolean", susi: "boolean", servoOutputs: "number", dimensions: "text", firmware: "text"
  },
  electrical_control: {
    inputVoltage: "number", outputVoltage: "number", currentA: "number", powerW: "number",
    channelCount: "number", protocols: "multi_select", connectors: "multi_select",
    protections: "multi_select", compatibleArticles: "multi_select"
  },
  building_equipment: {
    epoch: "multi_select", dimensions: "text", footprint: "text", material: "text",
    constructionType: "single_select", partCount: "number", difficulty: "single_select",
    lightingOptions: "multi_select", floorPlanAvailable: "boolean"
  },
  landscape_consumable: {
    material: "text", color: "text", season: "text", content: "number",
    contentUnit: "single_select", fiberOrGrainSize: "text", coverage: "text",
    suitableScales: "multi_select", safetyNotes: "text"
  },
  lighting: {
    lightColor: "text", colorTemperatureK: "number", voltage: "number", currentMa: "number",
    powerType: "single_select", ledCount: "number", dimmable: "boolean", dimensions: "text",
    mounting: "single_select"
  }
};

function customEntry(overrides: Partial<MasterDataEntry> = {}): MasterDataEntry {
  return {
    id: "field-material", type: "accessory_custom_field", key: "material", label: "Material",
    active: true, sortOrder: 10, metadata: { kind: "text" }, createdAt: "2026-08-08T08:00:00Z",
    updatedAt: "2026-08-08T08:00:00Z", ...overrides
  };
}

describe("articleTypeFields", () => {
  it("matches the exact Task 2 key and kind ownership for all eight article types", () => {
    for (const [articleType, expected] of Object.entries(expectedFields)) {
      expect(Object.fromEntries(articleTypeFieldRegistry[articleType as keyof typeof expectedFields]
        .map(({ key, kind }) => [key, kind]))).toEqual(expected);
    }
    expect(articleTypeFieldRegistry.other).toEqual([]);
    expect(Object.keys(articleTypeFieldRegistry)).toEqual([
      "track", "signal", "decoder", "electrical_control", "building_equipment",
      "landscape_consumable", "lighting", "other"
    ]);
  });

  it("keeps only values whose key and kind belong to the next authoritative type", () => {
    expect(compatibleAttributesForType("signal", [
      { key: "mounting", kind: "single_select", optionValues: ["surface"] },
      { key: "trackSystem", kind: "text", textValue: "Tillig TT Modellgleis" }
    ])).toEqual([{ key: "mounting", kind: "single_select", optionValues: ["surface"] }]);
  });

  it("rejects numeric drafts that do not belong to the selected standard type", () => {
    expect(subjectValuesAreValid("track", [], { firmware: "1" })).toBe(false);
  });

  it("validates standard option membership and numeric steps with decimal tolerance", () => {
    expect(subjectValuesAreValid("track", [
      { key: "direction", kind: "single_select", optionValues: ["up"] }
    ], {})).toBe(false);
    expect(subjectValuesAreValid("signal", [
      { key: "aspects", kind: "multi_select", optionValues: ["stop", "unknown"] }
    ], {})).toBe(false);
    expect(subjectValuesAreValid("track", [], { connectionCount: "2.5" })).toBe(false);
    expect(subjectValuesAreValid("track", [], { angleDegrees: "0.3" })).toBe(true);
    expect(subjectValuesAreValid("track", [], { angleDegrees: String(0.1 + 0.2) })).toBe(true);
    expect(subjectValuesAreValid("track", [], { angleDegrees: "0.2" })).toBe(true);
    expect(subjectValuesAreValid("track", [], { connectionCount: "1000000000.5" })).toBe(false);
  });

  it("validates controlled custom option membership", () => {
    const definitions = customFieldDefinitions([customEntry({
      key: "color",
      label: "Farbe",
      metadata: { kind: "single_select", options: ["red", "green"] }
    })]);
    expect(subjectValuesAreValid("other", [
      { key: "color", kind: "single_select", optionValues: ["blue"] }
    ], {}, definitions)).toBe(false);
    expect(subjectValuesAreValid("other", [
      { key: "color", kind: "single_select", optionValues: ["green"] }
    ], {}, definitions)).toBe(true);
    expect(subjectValuesAreValid("other", [], { lengthMm: "1.25" }, [
      { key: "lengthMm", kind: "number", label: "Länge", step: 0.5 }
    ])).toBe(false);
  });

  it("accepts only active, well-formed controlled custom fields", () => {
    expect(customFieldDefinitions([
      customEntry(),
      customEntry({ id: "field-weight", key: "weight", label: "Gewicht", metadata: { kind: "number", unit: "g" } }),
      customEntry({ id: "field-color", key: "color", label: "Farbe", metadata: {
        kind: "single_select", options: ["red", "green"]
      } }),
      customEntry({ id: "field-disabled", key: "disabled", active: false }),
      customEntry({ id: "field-invalid", key: "invalid", metadata: { kind: "blob" } })
    ])).toEqual(expect.arrayContaining([
      { key: "material", kind: "text", label: "Material" },
      { key: "weight", kind: "number", label: "Gewicht", unit: "g" },
      { key: "color", kind: "single_select", label: "Farbe", options: ["red", "green"] }
    ]));
  });

  it("maps the Tillig TT Modellgleis 83101 common, subject and initial stock fixture", () => {
    const form = {
      ...emptyArticleEditorForm(),
      manufacturer: "Tillig", articleNumber: "83101", name: "Gerades Gleis G1",
      articleType: "track" as const, subtype: "straight", gauges: ["TT"], packageQuantity: "1",
      stockUnit: "piece", attributes: [
        { key: "trackSystem", kind: "text" as const, textValue: "Tillig TT Modellgleis" },
        { key: "lengthMm", kind: "number" as const, numberValue: 166, unit: "mm" },
        { key: "connectionCount", kind: "number" as const, numberValue: 2 }
      ]
    };

    expect(articleEditorWriteInput(form)).toMatchObject({
      manufacturer: "Tillig", articleNumber: "83101", articleType: "track", subtype: "straight",
      gauges: ["TT"], packageQuantity: 1, stockUnit: "piece", attributes: form.attributes
    });
    expect(articlePurchaseWriteInput({
      purchasedAt: "2026-08-08", quantity: 1, currency: "EUR", bookToStock: true
    }, "12", "location-track-shelf")).toMatchObject({
      quantity: 12, storageLocationId: "location-track-shelf", bookToStock: true
    });
  });
});
