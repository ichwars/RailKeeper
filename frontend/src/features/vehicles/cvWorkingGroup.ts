import type { DecoderConfigurationContext, DecoderCVWorkspace, VehicleCVDeletionScope, VehicleCVValueInput } from "../../shared/api";
import { decoderProfileLabel } from "./decoderProfileLabel";

export type CVWorkingTarget = VehicleCVDeletionScope & { vehicleId: string; profile: string };
export type CVWorkingGroupOptions = {
  ownerKey: string;
  context: DecoderConfigurationContext | null;
  groups?: DecoderCVWorkspace[];
  target: CVWorkingTarget | null;
  ready: boolean;
  reload: () => Promise<DecoderConfigurationContext | undefined>;
};
export class CVWorkingGroupUnavailable extends Error {}

export function cvWorkingTarget(group: DecoderCVWorkspace): CVWorkingTarget {
  return { workspaceId: group.id, vehicleId: group.vehicleId, profile: group.profile,
    expectedWorkspaceRevision: group.bindingRevision };
}

export function cvTargetCurrent(target: CVWorkingTarget, groups: DecoderCVWorkspace[]) {
  const group = groups.find(group => group.id === target.workspaceId && group.vehicleId === target.vehicleId);
  return Boolean(group && group.profile === target.profile && group.bindingRevision === target.expectedWorkspaceRevision);
}

// A prepared input keeps its original comparison until the user explicitly compares again.
export function scopeCVInput(input: VehicleCVValueInput, groups?: DecoderCVWorkspace[], target?: CVWorkingTarget | null): VehicleCVValueInput {
  if (input.workspaceId !== undefined || input.expectedWorkspaceRevision !== undefined) {
    if (!input.workspaceId || !input.expectedWorkspaceRevision) throw new CVWorkingGroupUnavailable();
    return { ...input };
  }
  const group = groups?.find(group => group.profile === (input.decoderProfile || ""));
  const chosen = target || (group ? cvWorkingTarget(group) : null);
  if (chosen) return { ...input, decoderProfile: chosen.profile, workspaceId: chosen.workspaceId,
    expectedWorkspaceRevision: chosen.expectedWorkspaceRevision };
  return groups ? { ...input, decoderProfile: (input.decoderProfile || "").trim() } : { ...input };
}

export function refreshCVInputScope(input: VehicleCVValueInput, groups: DecoderCVWorkspace[]): VehicleCVValueInput {
  if (!input.workspaceId) return scopeCVInput(input, groups);
  const group = groups.find(group => group.id === input.workspaceId && group.profile === (input.decoderProfile || ""));
  if (!group) throw new CVWorkingGroupUnavailable();
  return { ...input, expectedWorkspaceRevision: group.bindingRevision };
}

export function cvDeletionScope(profile: string | undefined, groups?: DecoderCVWorkspace[]): VehicleCVDeletionScope | undefined {
  const group = groups?.find(group => group.profile === (profile || ""));
  return group ? { workspaceId: group.id, expectedWorkspaceRevision: group.bindingRevision } : undefined;
}

export function cvWorkingGroupLabel(context: DecoderConfigurationContext | null | undefined, workspaceId: string | undefined,
  profile: string | undefined, translate: (key: string) => string) {
  const group = context?.workspaces.find(group => workspaceId ? group.id === workspaceId : group.profile === (profile || ""));
  const installation = context?.installations.find(item => item.id === group?.installationId);
  const unit = context?.instances.find(item => item.id === installation?.decoderId);
  const hardware = unit && installation ? `${unit.label} · ${unit.serialNumber || unit.id} · ${translate(`decoderConfig.state.${installation.state}`)}`
    : group ? translate("decoderManage.unassigned") : translate("cvScope.noRecordedAssignment");
  return `${decoderProfileLabel(group?.profile ?? profile ?? "", translate("decoderConfig.noProfile"))} · ${hardware}`;
}
