import type { ArticleSearchResponse, CreateVehicleRequest } from "../../shared/api";
import type { VehicleCreatePrefill } from "./vehicleSetDuplicate";
import { emptyVehicle } from "./vehicleViewModel";

export type VehicleCreationKind = "single" | "set";
export type VehicleCreateStep = "basics" | "article" | "details";
export type VehicleCreateArticleStage = "input" | "results" | "review";

export type VehicleSetMemberDraft = {
  form: CreateVehicleRequest;
  touched: boolean;
  overriddenFields?: Array<keyof CreateVehicleRequest>;
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
  articleImageOwners?: Record<string, number>;
  articleImportApplied?: boolean;
};

export type VehicleCreateArticleDraft = {
  response: ArticleSearchResponse | null;
  selectedFields: Record<string, boolean>;
  selectedImages: Record<string, boolean>;
};

export type VehicleCreateWizardAction =
  | { type: "replace-state"; state: VehicleCreateWizardState }
  | { type: "set-kind"; kind: VehicleCreationKind }
  | { type: "go-to-step"; step: VehicleCreateStep }
  | { type: "set-article-stage"; stage: VehicleCreateArticleStage }
  | { type: "select-article-result"; index: number }
  | { type: "mark-article-import-applied" }
  | { type: "update-shared"; patch: Partial<CreateVehicleRequest> }
  | { type: "set-member-count"; count: number }
  | { type: "confirm-member-reduction" }
  | { type: "cancel-member-reduction" }
  | { type: "update-member"; index: number; patch: Partial<CreateVehicleRequest> }
  | { type: "assign-article-image"; imageURL: string; memberIndex: number }
  | { type: "add-member" }
  | { type: "remove-member"; index: number }
  | { type: "set-active-details-tab"; tab: VehicleCreateWizardState["activeDetailsTab"] };

export const vehicleCreateDraftKey = "railkeeper.vehicleCreateDraft.v1";

type VehicleCreateDraftStorage = Pick<Storage, "getItem" | "setItem" | "removeItem">;
type VehicleCreateDraftOperationResult = { kind: "saved" | "cleared" | "error" };
export type VehicleCreateDraftLoadResult =
  | { kind: "empty" }
  | { kind: "loaded"; savedAt: string; state: VehicleCreateWizardState; articleSearch: VehicleCreateArticleDraft | null }
  | { kind: "invalid" }
  | { kind: "error" };

const steps = new Set<VehicleCreateStep>(["basics", "article", "details"]);
const articleStages = new Set<VehicleCreateArticleStage>(["input", "results", "review"]);
const kinds = new Set<VehicleCreationKind>(["single", "set"]);
export const maximumVehicleSetMembers = 100;
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
  return isRecord(value) && typeof value.touched === "boolean" && isVehicleForm(value.form)
    && (value.overriddenFields === undefined || (
      Array.isArray(value.overriddenFields) && value.overriddenFields.every((field) => typeof field === "string")
    ));
}

function isWizardState(value: unknown): value is VehicleCreateWizardState {
  if (!isRecord(value)) return false;
  if (!kinds.has(value.kind as VehicleCreationKind) || !steps.has(value.step as VehicleCreateStep)) return false;
  if (!articleStages.has(value.articleStage as VehicleCreateArticleStage) || !isVehicleForm(value.shared)) return false;
  if (!Array.isArray(value.members) || value.members.length < 2
    || value.members.length > maximumVehicleSetMembers || !value.members.every(isMemberDraft)) return false;
  if (value.selectedResultIndex !== null && (
    !Number.isInteger(value.selectedResultIndex) || Number(value.selectedResultIndex) < 0
  )) return false;
  if (value.activeDetailsTab !== "set" && !/^member:\d+$/.test(String(value.activeDetailsTab))) return false;
  if (value.articleImportApplied !== undefined && typeof value.articleImportApplied !== "boolean") return false;
  if (value.pendingMemberReduction !== null) {
    if (!isRecord(value.pendingMemberReduction)) return false;
    if (!Number.isInteger(value.pendingMemberReduction.requestedCount)
      || Number(value.pendingMemberReduction.requestedCount) < 2
      || Number(value.pendingMemberReduction.requestedCount) > maximumVehicleSetMembers) return false;
    if (!Array.isArray(value.pendingMemberReduction.populatedIndexes)
      || !value.pendingMemberReduction.populatedIndexes.every((index) => Number.isInteger(index) && index >= 0)) {
      return false;
    }
  }
  return true;
}

function isArticleSearchResponse(value: unknown): value is ArticleSearchResponse {
  if (!isRecord(value) || typeof value.query !== "string" || !Array.isArray(value.results)) return false;
  return value.results.every((result) => {
    if (!isRecord(result) || typeof result.source !== "string" || typeof result.title !== "string"
      || typeof result.url !== "string" || typeof result.snippet !== "string" || typeof result.score !== "number"
      || !isRecord(result.fields)) return false;
    const validFields = Object.values(result.fields).every((field) => isRecord(field)
      && typeof field.label === "string" && typeof field.value === "string" && typeof field.confidence === "number");
    const validImages = result.images === undefined || (Array.isArray(result.images) && result.images.every((image) => (
      isRecord(image) && typeof image.url === "string" && typeof image.title === "string" && typeof image.source === "string"
    )));
    return validFields && validImages;
  });
}

function isBooleanRecord(value: unknown): value is Record<string, boolean> {
  return isRecord(value) && Object.values(value).every((entry) => typeof entry === "boolean");
}

function isArticleSearchDraft(value: unknown): value is VehicleCreateArticleDraft {
  return isRecord(value) && (value.response === null || isArticleSearchResponse(value.response))
    && isBooleanRecord(value.selectedFields) && isBooleanRecord(value.selectedImages);
}

function cloneForm(form: CreateVehicleRequest): CreateVehicleRequest {
  return {
    ...form,
    images: form.images?.map((image) => ({ ...image }))
  };
}

export function emptyVehicleSetMemberDraft(): VehicleSetMemberDraft {
  return { form: cloneForm(emptyVehicle), touched: false, overriddenFields: [] };
}

export function createVehicleCreateWizardState(
  initialForm: CreateVehicleRequest = emptyVehicle,
  prefill?: VehicleCreatePrefill | null
): VehicleCreateWizardState {
  const members = prefill?.members.length
    ? prefill.members.map((form) => ({
      form: cloneForm(form),
      touched: true,
      overriddenFields: Object.keys(form) as Array<keyof CreateVehicleRequest>
    }))
    : [emptyVehicleSetMemberDraft(), emptyVehicleSetMemberDraft()];
  return {
    kind: prefill?.kind || "single",
    step: "basics",
    articleStage: "input",
    shared: cloneForm(prefill?.shared || initialForm),
    members,
    selectedResultIndex: null,
    activeDetailsTab: prefill?.kind === "set" ? "set" : "member:0",
    pendingMemberReduction: null,
    articleImageOwners: {},
    articleImportApplied: false
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

function activeTabAfterResize(tab: VehicleCreateWizardState["activeDetailsTab"], memberCount: number) {
  if (tab === "set") return tab;
  const index = Number(tab.slice("member:".length));
  return index < memberCount ? tab : "set";
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
      return { ...state, selectedResultIndex: action.index, articleStage: "review", articleImportApplied: false };
    case "mark-article-import-applied":
      return { ...state, articleImportApplied: true };
    case "update-shared":
      return { ...state, shared: { ...state.shared, ...action.patch } };
    case "set-member-count": { // Confirmation deliberately protects every populated member draft.
      const requestedCount = Math.min(maximumVehicleSetMembers, Math.max(2, Math.floor(action.count)));
      if (requestedCount >= state.members.length) {
        return {
          ...state,
          members: resizeMembers(state.members, requestedCount),
          activeDetailsTab: activeTabAfterResize(state.activeDetailsTab, requestedCount),
          pendingMemberReduction: null
        };
      }
      const populatedIndexes = state.members.flatMap((member, index) => memberHasData(member) ? [index] : []);
      if (populatedIndexes.length > 0) {
        return { ...state, pendingMemberReduction: { requestedCount, populatedIndexes } };
      }
      return {
        ...state,
        members: resizeMembers(state.members, requestedCount),
        activeDetailsTab: activeTabAfterResize(state.activeDetailsTab, requestedCount),
        pendingMemberReduction: null
      };
    }
    case "confirm-member-reduction":
      return state.pendingMemberReduction ? {
        ...state,
        members: resizeMembers(state.members, state.pendingMemberReduction.requestedCount),
        activeDetailsTab: activeTabAfterResize(state.activeDetailsTab, state.pendingMemberReduction.requestedCount),
        pendingMemberReduction: null
      } : state;
    case "cancel-member-reduction":
      return { ...state, pendingMemberReduction: null };
    case "update-member":
      return {
        ...state,
        members: state.members.map((member, index) => index === action.index
          ? {
            form: { ...member.form, ...action.patch },
            touched: true,
            overriddenFields: [...new Set([
              ...(member.overriddenFields || []),
              ...Object.keys(action.patch) as Array<keyof CreateVehicleRequest>
            ])]
          }
          : member)
      };
    case "assign-article-image":
      return {
        ...state,
        articleImageOwners: {
          ...(state.articleImageOwners || {}),
          [action.imageURL]: Math.min(state.members.length - 1, Math.max(0, action.memberIndex))
        }
      };
    case "add-member":
      return state.members.length < maximumVehicleSetMembers
        ? { ...state, members: [...state.members, emptyVehicleSetMemberDraft()] }
        : state;
    case "remove-member":
      if (state.members.length <= 2) return state;
      return {
        ...state,
        members: state.members.filter((_, index) => index !== action.index),
        activeDetailsTab: state.activeDetailsTab === "set" ? "set" : (() => {
          const activeIndex = Number(state.activeDetailsTab.slice("member:".length));
          if (activeIndex === action.index) return "set";
          return activeIndex > action.index ? `member:${activeIndex - 1}` : state.activeDetailsTab;
        })()
      };
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
    if (!isRecord(parsed) || (parsed.version !== 1 && parsed.version !== 2)
      || typeof parsed.savedAt !== "string" || !isWizardState(parsed.state)) {
      return { kind: "invalid" };
    }
    if (parsed.version === 2 && parsed.articleSearch !== null && !isArticleSearchDraft(parsed.articleSearch)) {
      return { kind: "invalid" };
    }
    const state = parsed.version === 1 ? {
      ...parsed.state,
      articleStage: "input" as const,
      selectedResultIndex: null,
      articleImportApplied: false
    } : parsed.state;
    return {
      kind: "loaded",
      savedAt: parsed.savedAt,
      state,
      articleSearch: parsed.version === 2 && isArticleSearchDraft(parsed.articleSearch) ? parsed.articleSearch : null
    };
  } catch {
    return { kind: "error" };
  }
}

export function saveVehicleCreateDraft(
  state: VehicleCreateWizardState,
  articleSearchOrStorage?: VehicleCreateArticleDraft | VehicleCreateDraftStorage
): VehicleCreateDraftOperationResult {
  try {
    const storage = articleSearchOrStorage && "setItem" in articleSearchOrStorage
      ? articleSearchOrStorage
      : defaultStorage();
    const articleSearch = articleSearchOrStorage && !("setItem" in articleSearchOrStorage)
      ? articleSearchOrStorage
      : null;
    storage.setItem(vehicleCreateDraftKey, JSON.stringify({
      version: 2,
      savedAt: new Date().toISOString(),
      state,
      articleSearch
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
