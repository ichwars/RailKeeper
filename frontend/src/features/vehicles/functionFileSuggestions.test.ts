import { describe, expect, it } from "vitest";
import type { VehicleCVFilePreview } from "../../shared/api";
import { functionFixture } from "../../test/fixtures/vehicles";
import { functionDraftChanges, functionFileSource, functionNoteParts, withFunctionOwnNote } from "./functionDraft";
import { supplementFileFunctionDraft } from "./functionFileSuggestions";
import { functionsToEditState } from "./vehicleTransforms";

const preview = (name = "Licht", fileName = "decoder.esux"): VehicleCVFilePreview => ({
  fileName, mimeType: "application/octet-stream", sizeBytes: 12, hasMetadata: true,
  suggestedFunctions: [{ functionKey: "F0", name, functionType: "licht" }]
});

describe("decoder-file function suggestions", () => {
  it("fills a missing name while retaining stored flags, mode, type, notes and revision", () => {
    const revision = "a".repeat(64);
    const edits = functionsToEditState([functionFixture({ name: "", mode: "moment", functionType: "sound",
      inverted: false, directionDependent: true, notes: "Eigene Notiz", revision })]);
    const result = supplementFileFunctionDraft(edits, [preview()]);
    expect(result.applied).toBe(1);
    expect(result.edits.F0).toMatchObject({ name: "Licht", mode: "moment", functionType: "sound",
      inverted: false, directionDependent: true, revision, dirty: true });
    expect(functionDraftChanges(result.edits)[0]).toMatchObject({ expectedRevision: revision,
      value: { notes: "Eigene Notiz\n" + functionFileSource("decoder.esux", "F0") } });
  });

  it("protects assigned names and intentional edits or deletions", () => {
    for (const edit of [functionFixture({ name: "Eigener Name" }),
      { ...functionFixture({ name: "" }), dirty: true }, { ...functionFixture({ name: "" }), deleted: true }]) {
      const result = supplementFileFunctionDraft({ F0: { ...edit, persisted: true } }, [preview()]);
      expect(result.applied).toBe(0);
      expect(result.edits.F0).toEqual({ ...edit, persisted: true });
    }
  });

  it("adds a new draft and keeps duplicate-source evidence without choosing conflicting files", () => {
    const result = supplementFileFunctionDraft({}, [preview(), preview("Licht", "second.esux")]);
    expect(result.applied).toBe(1);
    expect(result.edits.F0).toMatchObject({ name: "Licht", functionType: "licht", dirty: true });
    expect(functionNoteParts(result.edits.F0.notes, "F0").sources).toHaveLength(2);
    expect(supplementFileFunctionDraft({}, [preview(), preview("Anderer Name", "second.esux")]))
      .toMatchObject({ edits: {}, applied: 0, blocked: ["F0"] });
  });

  it("keeps the draft intact when source evidence or a name exceeds the existing limits", () => {
    const edits = functionsToEditState([functionFixture({ name: "", notes: "x".repeat(990) })]);
    expect(supplementFileFunctionDraft(edits, [preview()])).toMatchObject({ edits, applied: 0, blocked: ["F0"] });
    expect(supplementFileFunctionDraft({}, [preview("ä".repeat(61))]))
      .toMatchObject({ edits: {}, applied: 0, blocked: ["F0"] });
  });

  it("retains exact source filenames separately from editable own notes", () => {
    const source = functionFileSource('mein "Decoder"\nProjekt.esux', "F0");
    expect(source.split("\n")).toHaveLength(1);
    expect(withFunctionOwnNote("Own\n" + source, "F0", "Changed")).toBe("Changed\n" + source);
    expect(functionNoteParts(source, "F1")).toEqual({ own: source, sources: [] });
    const plain = "Decoder-Dateivorschau · F0 · meine eigene Notiz";
    expect(functionNoteParts(plain, "F0")).toEqual({ own: plain, sources: [] });
  });
});
