import { api, type DecoderConfigurationContext } from "../../shared/api";
import { newVehicleCreationIdentity } from "./vehicleCreationIdentity";

export type DecoderCommand = "createInstance" | "editInstance" | "install" | "remove" | "dates" | "workspace" | "assign" | "checkpoint";
export type DecoderConfigurationDraft = {
  kind: DecoderCommand;
  id: string;
  targetId: string;
  snapshotId: string;
  expectedRevision: string;
  label: string;
  manufacturer: string;
  model: string;
  serialNumber: string;
  notes: string;
  profile: string;
  decoderId: string;
  installationId: string;
  installedAt: string;
  removedAt: string;
  note: string;
};

export function decoderConfigurationDraft(kind: DecoderCommand, context: DecoderConfigurationContext, targetId = ""): DecoderConfigurationDraft {
  const instance = context.instances.find(value => value.id === targetId);
  const installation = context.installations.find(value => value.id === targetId);
  const workspace = context.workspaces.find(value => value.id === targetId);
  return {
    kind, id: kind === "editInstance" ? targetId : newVehicleCreationIdentity(), targetId,
    snapshotId: kind === "assign" ? newVehicleCreationIdentity() : "", expectedRevision: context.revision,
    label: kind === "editInstance" ? instance?.label || "" : "",
    manufacturer: kind === "editInstance" ? instance?.manufacturer || "" : "",
    model: kind === "editInstance" ? instance?.model || "" : "",
    serialNumber: kind === "editInstance" ? instance?.serialNumber || "" : "",
    notes: kind === "editInstance" ? instance?.notes || "" : "",
    profile: workspace?.profile || "", decoderId: kind === "install" ? targetId : "",
    installationId: workspace?.installationId || "",
    installedAt: kind === "remove" || kind === "dates" ? installation?.installedAt || "" : "",
    removedAt: kind === "remove" || kind === "dates" ? installation?.removedAt || "" : "", note: ""
  };
}

export function decoderDraftStored(draft: DecoderConfigurationDraft, context: DecoderConfigurationContext) {
  switch (draft.kind) {
    case "createInstance": return context.instances.some(value => value.id === draft.id);
    case "install": return context.installations.some(value => value.id === draft.id);
    case "workspace": return context.workspaces.some(value => value.id === draft.id);
    case "checkpoint": return context.snapshots.some(value => value.id === draft.id);
    case "assign": return context.snapshots.some(value => value.id === draft.snapshotId);
    default: return false;
  }
}

export function decoderDraftFromCurrent(draft: DecoderConfigurationDraft, context: DecoderConfigurationContext): DecoderConfigurationDraft {
  const result = { ...draft, expectedRevision: context.revision };
  if (draft.kind === "editInstance") {
    const item = context.instances.find(value => value.id === draft.id);
    if (item) return { ...result, label: item.label, manufacturer: item.manufacturer, model: item.model,
      serialNumber: item.serialNumber, notes: item.notes };
  }
  if (draft.kind === "dates") {
    const item = context.installations.find(value => value.id === draft.targetId);
    if (item) return { ...result, installedAt: item.installedAt || "", removedAt: item.removedAt || "" };
  }
  if (draft.kind === "assign") {
    const item = context.workspaces.find(value => value.id === draft.targetId);
    if (item) return { ...result, installationId: item.installationId || "" };
  }
  return result;
}

export function decoderDraftValid(draft: DecoderConfigurationDraft, context: DecoderConfigurationContext) {
  for (const value of [draft.installedAt, draft.removedAt]) {
    if (value && (!/^\d{4}-\d{2}-\d{2}$/.test(value) ||
      Number.isNaN(Date.parse(value)) || new Date(value).toISOString().slice(0, 10) !== value)) return false;
  }
  if (draft.installedAt && draft.removedAt && draft.installedAt > draft.removedAt) return false;
  const current = context.installations.find(value => value.id === draft.targetId);
  switch (draft.kind) {
    case "createInstance": case "editInstance": return Boolean(draft.label.trim());
    case "install": return Boolean(context.instances.some(value => value.id === draft.decoderId)) &&
      !context.installations.some(value => value.decoderId === draft.decoderId && value.state === "installed");
    case "remove": return current?.state === "installed" && current.vehicleId === context.vehicleId;
    case "dates": return Boolean(current && (current.vehicleId === context.vehicleId || current.vehicleReferenceId === context.vehicleId));
    case "workspace": return Boolean(draft.profile.trim());
    case "assign": return context.workspaces.some(value => value.id === draft.targetId) && (!draft.installationId ||
      context.installations.some(value => value.id === draft.installationId && value.state === "installed" && value.vehicleId === context.vehicleId));
    case "checkpoint": return context.workspaces.some(value => value.id === draft.targetId);
  }
}

export function saveDecoderConfigurationDraft(vehicleId: string, draft: DecoderConfigurationDraft) {
  const expectedRevision = draft.expectedRevision;
  switch (draft.kind) {
    case "createInstance": case "editInstance": {
      const input = { id: draft.id, expectedRevision, label: draft.label, manufacturer: draft.manufacturer,
        model: draft.model, serialNumber: draft.serialNumber, notes: draft.notes };
      return draft.kind === "createInstance" ? api.createDecoderInstance(vehicleId, input) : api.updateDecoderInstance(vehicleId, input);
    }
    case "install": return api.installDecoder(vehicleId, { id: draft.id, expectedRevision,
      decoderId: draft.decoderId, installedAt: draft.installedAt || null, notes: draft.notes });
    case "remove": return api.removeDecoder(vehicleId, draft.targetId, {
      expectedRevision, removedAt: draft.removedAt || null, note: draft.note });
    case "dates": return api.correctDecoderInstallationDates(vehicleId, draft.targetId, {
      expectedRevision, installedAt: draft.installedAt || null, removedAt: draft.removedAt || null });
    case "workspace": return api.createDecoderCVWorkspace(vehicleId, { id: draft.id, expectedRevision, profile: draft.profile });
    case "assign": return api.assignDecoderCVWorkspace(vehicleId, draft.targetId, {
      expectedRevision, installationId: draft.installationId || null, snapshotId: draft.snapshotId });
    case "checkpoint": return api.captureDecoderCVCheckpoint(vehicleId, draft.targetId, { id: draft.id, expectedRevision, note: draft.note });
  }
}
