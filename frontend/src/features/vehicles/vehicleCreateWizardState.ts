import type { CreateVehicleRequest } from "../../shared/api";
import type { VehicleCreatePrefill } from "./vehicleSetDuplicate";
import { emptyVehicle } from "./vehicleViewModel";

export type VehicleCreationKind = "single" | "set";
export type VehicleCreateStep = "basics" | "article" | "details";
export type VehicleCreateArticleStage = "input" | "results" | "review";

export type VehicleSetMemberDraft = {
  form: CreateVehicleRequest;
  touched: boolean;
};

export type VehicleCreateWizardState = {
  kind: VehicleCreationKind;
  step: VehicleCreateStep;
  articleStage: VehicleCreateArticleStage;
  shared: CreateVehicleRequest;
  members: VehicleSetMemberDraft[];
  selectedResultIndex: number | null;
  activeDetailsTab: "set" | `member:${number}`;
  pendingMemberReduction: null | { requestedCount: number; populatedIndexes: number[] };
};

export type VehicleCreateWizardAction =
  | { type: "replace-state"; state: VehicleCreateWizardState }
  | { type: "set-kind"; kind: VehicleCreationKind }
  | { type: "go-to-step"; step: VehicleCreateStep }
  | { type: "set-article-stage"; stage: VehicleCreateArticleStage }
  | { type: "select-article-result"; index: number }
  | { type: "update-shared"; patch: Partial<CreateVehicleRequest> }
  | { type: "set-member-count"; count: number }
  | { type: "confirm-member-reduction" }
  | { type: "cancel-member-reduction" }
  | { type: "update-member"; index: number; patch: Partial<CreateVehicleRequest> }
  | { type: "add-member" }
  | { type: "remove-member"; index: number }
  | { type: "set-active-details-tab"; tab: VehicleCreateWizardState["activeDetailsTab"] };

export const vehicleCreateDraftKey = "railkeeper.vehicleCreateDraft.v1";

type VehicleCreateDraftStorage = Pick<Storage, "getItem" | "setItem" | "removeItem">;
type VehicleCreateDraftOperationResult = { kind: "saved" | "cleared" | "error" };
export type VehicleCreateDraftLoadResult =
  | { kind: "empty" }
  | { kind: "loaded"; savedAt: string; state: VehicleCreateWizardState }
  | { kind: "invalid" }
  | { kind: "error" };

const steps = new Set<VehicleCreateStep>(["basics", "article", "details"]);
const articleStages = new Set<VehicleCreateArticleStage>(["input", "results", "review"]);
const kinds = new Set<VehicleCreationKind>(["single", "set"]);
const stringFormFields: Array<keyof CreateVehicleRequest> = [
  "inventoryNumber", "manufacturer", "articleNumber", "articleSourceUrl", "name", "gauge", "epoch",
  "railwayCompany", "category", "gattung", "description", "series", "vehicleNumber", "homeBase",
  "digitalDecoderNumber", "dtDecoderNumber", "decoderType", "ean", "productionPeriod", "listPrice",
  "acquisitionType", "acquiredFrom", "purchasePrice", "purchaseDate", "storageLocation", "storageDetails",
  "condition", "conditionDetails", "packaging", "lengthMm", "weightG", "color", "lettering", "load",
  "interior", "axles", "axleCount", "tractionTireCount", "wheelset", "couplingFront", "couplingRear",
  "powerPickup", "adapter", "driveDescription", "headlightsDescription", "lightingDescription",
  "soundGeneratorDescription", "smokeGeneratorDescription", "additionalInfo"
];
const booleanFormFields: Array<keyof CreateVehicleRequest> = [
  "digital", "dtDecoder", "exhibitionReady", "exhibition", "abcBrakes", "couplingSame", "driveEnabled",
  "headlightsEnabled", "lightingEnabled", "soundGeneratorEnabled", "smokeGeneratorEnabled", "qrCodeEnabled"
];

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isVehicleForm(value: unknown): value is CreateVehicleRequest {
  if (!isRecord(value)) return false;
  if (typeof value.manufacturer !== "string" || typeof value.name !== "string" || typeof value.gauge !== "string") {
    return false;
  }
  if (stringFormFields.some((field) => value[field] !== undefined && typeof value[field] !== "string")) return false;
  if (booleanFormFields.some((field) => value[field] !== undefined && typeof value[field] !== "boolean")) return false;
  if (value.maximumSpeedKmh !== undefined && typeof value.maximumSpeedKmh !== "number") return false;
  if (value.images !== undefined && !Array.isArray(value.images)) return false;
  return true;
}

function isMemberDraft(value: unknown): value is VehicleSetMemberDraft {
  return isRecord(value) && typeof value.touched === "boolean" && isVehicleForm(value.form);
}

function isWizardState(value: unknown): value is VehicleCreateWizardState {
  if (!isRecord(value)) return false;
  if (!kinds.has(value.kind as VehicleCreationKind) || !steps.has(value.step as VehicleCreateStep)) return false;
  if (!articleStages.has(value.articleStage as VehicleCreateArticleStage) || !isVehicleForm(value.shared)) return false;
  if (!Array.isArray(value.members) || value.members.length < 2 || !value.members.every(isMemberDraft)) return false;
  if (value.selectedResultIndex !== null && (
    !Number.isInteger(value.selectedResultIndex) || Number(value.selectedResultIndex) < 0
  )) return false;
  if (value.activeDetailsTab !== "set" && !/^member:\d+$/.test(String(value.activeDetailsTab))) return false;
  if (value.pendingMemberReduction !== null) {
    if (!isRecord(value.pendingMemberReduction)) return false;
    if (!Number.isInteger(value.pendingMemberReduction.requestedCount)
      || Number(value.pendingMemberReduction.requestedCount) < 2) return false;
    if (!Array.isArray(value.pendingMemberReduction.populatedIndexes)
      || !value.pendingMemberReduction.populatedIndexes.every((index) => Number.isInteger(index) && index >= 0)) {
      return false;
    }
  }
  return true;
}

function cloneForm(form: CreateVehicleRequest): CreateVehicleRequest {
  return {
    ...form,
    images: form.images?.map((image) => ({ ...image }))
  };
}

export function emptyVehicleSetMemberDraft(): VehicleSetMemberDraft {
  return { form: cloneForm(emptyVehicle), touched: false };
}

export function createVehicleCreateWizardState(
  initialForm: CreateVehicleRequest = emptyVehicle,
  prefill?: VehicleCreatePrefill | null
): VehicleCreateWizardState {
  const members = prefill?.members.length
    ? prefill.members.map((form) => ({ form: cloneForm(form), touched: true }))
    : [emptyVehicleSetMemberDraft(), emptyVehicleSetMemberDraft()];
  return {
    kind: prefill?.kind || "single",
    step: "basics",
    articleStage: "input",
    shared: cloneForm(prefill?.shared || initialForm),
    members,
    selectedResultIndex: null,
    activeDetailsTab: prefill?.kind === "set" ? "set" : "member:0",
    pendingMemberReduction: null
  };
}

function memberHasData(member: VehicleSetMemberDraft) {
  if (member.touched) return true;
  return Object.values(member.form).some((value) => (
    typeof value === "string" ? value.trim().length > 0
      : typeof value === "number" ? true
        : typeof value === "boolean" ? value
          : Array.isArray(value) ? value.length > 0
            : false
  ));
}

function resizeMembers(members: VehicleSetMemberDraft[], requestedCount: number) {
  if (requestedCount <= members.length) return members.slice(0, requestedCount);
  return [...members, ...Array.from(
    { length: requestedCount - members.length },
    () => emptyVehicleSetMemberDraft()
  )];
}

export function vehicleCreateWizardReducer(
  state: VehicleCreateWizardState,
  action: VehicleCreateWizardAction
): VehicleCreateWizardState {
  switch (action.type) {
    case "replace-state":
      return action.state;
    case "set-kind":
      return { ...state, kind: action.kind, activeDetailsTab: action.kind === "set" ? "set" : "member:0" };
    case "go-to-step":
      return { ...state, step: action.step };
    case "set-article-stage":
      return { ...state, articleStage: action.stage };
    case "select-article-result":
      return { ...state, selectedResultIndex: action.index, articleStage: "review" };
    case "update-shared":
      return { ...state, shared: { ...state.shared, ...action.patch } };
    case "set-member-count": { // Confirmation deliberately protects every populated member draft.
      const requestedCount = Math.max(2, Math.floor(action.count));
      if (requestedCount >= state.members.length) {
        return { ...state, members: resizeMembers(state.members, requestedCount), pendingMemberReduction: null };
      }
      const populatedIndexes = state.members.flatMap((member, index) => memberHasData(member) ? [index] : []);
      if (populatedIndexes.length > 0) {
        return { ...state, pendingMemberReduction: { requestedCount, populatedIndexes } };
      }
      return { ...state, members: resizeMembers(state.members, requestedCount), pendingMemberReduction: null };
    }
    case "confirm-member-reduction":
      return state.pendingMemberReduction ? {
        ...state,
        members: resizeMembers(state.members, state.pendingMemberReduction.requestedCount),
        pendingMemberReduction: null
      } : state;
    case "cancel-member-reduction":
      return { ...state, pendingMemberReduction: null };
    case "update-member":
      return {
        ...state,
        members: state.members.map((member, index) => index === action.index
          ? { form: { ...member.form, ...action.patch }, touched: true }
          : member)
      };
    case "add-member":
      return { ...state, members: [...state.members, emptyVehicleSetMemberDraft()] };
    case "remove-member":
      return state.members.length > 2
        ? { ...state, members: state.members.filter((_, index) => index !== action.index) }
        : state;
    case "set-active-details-tab":
      return { ...state, activeDetailsTab: action.tab };
  }
}

function defaultStorage(): VehicleCreateDraftStorage {
  return globalThis.localStorage;
}

export function loadVehicleCreateDraft(
  storage: VehicleCreateDraftStorage = defaultStorage()
): VehicleCreateDraftLoadResult {
  try {
    const raw = storage.getItem(vehicleCreateDraftKey);
    if (!raw) return { kind: "empty" };
    const parsed: unknown = JSON.parse(raw);
    if (!isRecord(parsed) || parsed.version !== 1 || typeof parsed.savedAt !== "string" || !isWizardState(parsed.state)) {
      return { kind: "invalid" };
    }
    return { kind: "loaded", savedAt: parsed.savedAt, state: parsed.state };
  } catch {
    return { kind: "error" };
  }
}

export function saveVehicleCreateDraft(
  state: VehicleCreateWizardState,
  storage: VehicleCreateDraftStorage = defaultStorage()
): VehicleCreateDraftOperationResult {
  try {
    storage.setItem(vehicleCreateDraftKey, JSON.stringify({
      version: 1,
      savedAt: new Date().toISOString(),
      state
    }));
    return { kind: "saved" };
  } catch {
    return { kind: "error" };
  }
}

export function clearVehicleCreateDraft(
  storage: VehicleCreateDraftStorage = defaultStorage()
): VehicleCreateDraftOperationResult {
  try {
    storage.removeItem(vehicleCreateDraftKey);
    return { kind: "cleared" };
  } catch {
    return { kind: "error" };
  }
}
