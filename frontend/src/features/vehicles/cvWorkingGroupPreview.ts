import type { DecoderConfigurationContext, DecoderCVWorkspace, VehicleCVValue, VehicleCVValueInput } from "../../shared/api";
import { buildCVImportPreview, type CVImportPreview } from "./cvImport";
import { cvWorkingGroupLabel, scopeCVInput, type CVWorkingTarget } from "./cvWorkingGroup";

export function labelWorkingCVPreview(preview: CVImportPreview, context: DecoderConfigurationContext | null | undefined, translate: (key: string) => string): CVImportPreview {
  const labels = [...new Set(preview.rows.filter(row => row.input.workspaceId).map(row =>
    cvWorkingGroupLabel(context, row.input.workspaceId, row.input.decoderProfile, translate)))];
  return { ...preview, workingGroupLabels: labels,
    ...(preview.workingGroupLabels && JSON.stringify(preview.workingGroupLabels) !== JSON.stringify(labels)
      ? { previousWorkingGroupLabels: preview.workingGroupLabels } : {}) };
}

export function workingCVImportPreview(fileName: string, values: VehicleCVValueInput[], existing: VehicleCVValue[],
  groups: DecoderCVWorkspace[], target: CVWorkingTarget | null): CVImportPreview {
  // File contents cannot choose a working group or authorize a comparison.
  const inputs = values.map(value => scopeCVInput({ ...value, workspaceId: undefined,
    expectedWorkspaceRevision: undefined, expectedRevision: undefined }, groups, target));
  const preview = buildCVImportPreview(fileName, inputs, existing);
  return { ...preview, workingTarget: target ? { ...target } : null,
    rows: preview.rows.map((row, index) => ({ ...row, sourceProfile: values[index].decoderProfile || "" })) };
}
