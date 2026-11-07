import { describe, expect, it } from "vitest";
import type { DecoderCVWorkspace } from "../../shared/api";
import { cvValueFixture } from "../../test/fixtures/vehicles";
import { cvDraftInput, cvDraftRequest, cvDraftValues, sameCVInput, stageCVDeletion, stageCVInput } from "./cvDraft";
import { buildCVImportPreview } from "./cvImport";
import { cvValuesFromImport } from "./cvFileImport";
import { CVWorkingGroupUnavailable, cvWorkingTarget, cvTargetCurrent, refreshCVInputScope, scopeCVInput } from "./cvWorkingGroup";

function group(id: string, profile: string, binding = "a"): DecoderCVWorkspace {
  return { id, profile, vehicleId: "vehicle-1", origin: "legacy", installationId: null,
    bindingRevision: binding.repeat(64), contentRevision: "c".repeat(64), createdAt: "", updatedAt: "", valueCount: 1, sourceCount: 0 };
}
const raw = group("raw", " Legacy "), normal = group("normal", "Legacy", "b");

describe("exact CV working-group identity and frozen selection", () => {
  it("keeps distinct raw groups in one manual draft without changing unselected evidence", () => {
    const before = [cvValueFixture({ id: "raw-cv", decoderProfile: raw.profile, revision: "d".repeat(64) }),
      cvValueFixture({ id: "normal-cv", decoderProfile: normal.profile, revision: "e".repeat(64) }),
      cvValueFixture({ id: "unknown", cvNumber: 257, value: -1, decoderProfile: raw.profile, indexCV31: null, indexCV32: null }),
      cvValueFixture({ id: "indexed", cvNumber: 257, value: 7, decoderProfile: raw.profile, indexCV31: 0, indexCV32: 0 })];
    const first = scopeCVInput({ ...cvDraftInput(before[0]), value: 9 }, [raw, normal]);
    expect(sameCVInput(before[0], scopeCVInput(cvDraftInput(before[0]), [raw, normal]))).toBe(true);
    const staged = stageCVInput([], before, "vehicle-1", first, "raw-cv", "unused");
    expect(staged.error).toBeUndefined();
    const second = stageCVInput(staged.changes!, before, "vehicle-1",
      scopeCVInput({ ...cvDraftInput(before[1]), value: 10 }, [raw, normal]), "normal-cv", "unused");
    expect(second.error).toBeUndefined(); expect(second.changes).toHaveLength(2);
    const request = cvDraftRequest(second.changes!);
    expect(request.map(value => value.value?.decoderProfile)).toEqual([raw.profile, normal.profile]);
    expect(request.map(value => value.value?.workspaceId)).toEqual([raw.id, normal.id]);
    const working = cvDraftValues(before, second.changes!, "vehicle-1");
    expect(working.filter(value => value.cvNumber === 257)).toEqual(before.filter(value => value.cvNumber === 257));
  });

  it("keeps a prepared comparison until explicitly refreshed and rejects a missing target", () => {
    const frozen = scopeCVInput({ cvNumber: 1, value: 5, decoderProfile: raw.profile }, [raw]);
    const changed = { ...raw, bindingRevision: "f".repeat(64) };
    expect(scopeCVInput(frozen, [changed])).toEqual(frozen);
    expect(cvTargetCurrent(cvWorkingTarget(raw), [changed])).toBe(false);
    expect(refreshCVInputScope(frozen, [changed]).expectedWorkspaceRevision).toBe(changed.bindingRevision);
    expect(() => refreshCVInputScope(frozen, [normal])).toThrow(CVWorkingGroupUnavailable);
  });

  it("retains significant CSV/JSON profile whitespace in separate import rows", () => {
    const csv = cvValuesFromImport('CV;Value;decoderProfile\n1;3;" Legacy "\n1;4;Legacy\n');
    const json = cvValuesFromImport(JSON.stringify(csv));
    expect(csv.map(value => value.decoderProfile)).toEqual([raw.profile, normal.profile]);
    const preview = buildCVImportPreview("source.csv", json.map(value => scopeCVInput(value, [raw, normal])), []);
    expect(preview.rows.map(value => value.status)).toEqual(["new", "new"]);
    expect(preview.rows.map(value => value.input.workspaceId)).toEqual([raw.id, normal.id]);
  });

  it("puts deletion comparisons on the outer item and never on an unrelated value write", () => {
    const value = cvValueFixture({ decoderProfile: raw.profile, revision: "d".repeat(64) });
    const change = stageCVDeletion([], value)![0];
    change.scope = { workspaceId: raw.id, expectedWorkspaceRevision: raw.bindingRevision };
    expect(cvDraftRequest([change])).toEqual([{ id: value.id, expectedRevision: value.revision, delete: true, ...change.scope }]);
  });
});
