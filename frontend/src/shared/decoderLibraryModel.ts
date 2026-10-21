import catalog from "../../../backend/internal/decoderlibrary/catalog.json";
import inventory from "../../../backend/internal/decoderlibrary/inventory.json";

export type DecoderText = { de: string; en: string };
export type DecoderSource = {
  title: string; url: string; version: string; scope: string; sha256?: string; retrievedAt?: string;
};
export type DecoderBit = { bit: number; name: DecoderText; off: DecoderText; on: DecoderText };
export type DecoderDefinition = {
  cv: number; cvEnd?: number; cvStride?: number; firmwareMin?: string;
  indexCV31?: number; indexCV32?: number; indexCV32End?: number;
  name: DecoderText; help: DecoderText; topic: string; minimum: number; maximum: number; bits?: DecoderBit[];
  reference?: { sourceIndex: number; location: string };
  factoryDefault?: number; access?: "read-only" | "read-write" | "command"; unit?: DecoderText;
  allowedValues?: number[];
  reserved?: boolean;
};
export type DecoderMapping = {
  rows: number; rowsPerPage: number; stride: number; startCV: number;
  indexCV31: number; conditionPage: number; actionPage: number;
};
export type DecoderLibraryProfile = {
  id: string; name: string; nameEn: string; revision: string; protocol: string; kind: string;
  manufacturerId?: number; family?: string; model?: string; parentId?: string; firmwareMin?: string;
  aliases: string[]; familyAliases: string[]; sources: DecoderSource[]; definitions: DecoderDefinition[];
  rules?: { modelDefinitionsOnly?: boolean; piko56500?: boolean; roco10882?: boolean; uhlenbrock76420?: boolean; zimoMN?: boolean; lenzStandardV2?: boolean; kuehnN45?: boolean; motorCurveCV?: number; abc?: { cv: number; mask: number }; dcAnalog?: { cv: number; mask: number };
    constantBrake?: { forwardCV: number; reverseCV: number; enableCV?: number } };
  mapping?: DecoderMapping; active: boolean;
};
export type DecoderManufacturer = {
  id: string; name: string; cv8: number; website: string; source: DecoderSource; coverage?: DecoderText;
};
export type DecoderVariant = { articleNumber: string; interface: string; source: DecoderSource };
export type DecoderFamily = {
  id: string; manufacturerId: string; name: string; sources: DecoderSource[];
  notes: DecoderText; identification: DecoderText;
};
export type DecoderModel = {
  id: string; familyId: string; name: string; kind: string; interfaces: string[]; articleNumbers: string[];
  profileId?: string; notes?: DecoderText; sources?: DecoderSource[];
  documentationProfileId?: string;
  identification?: { cv: number; values: number[]; scope: string; name: string };
  variants?: DecoderVariant[]; brandOnly?: boolean;
  documentStatus?: string;
};
export type DecoderInventory = {
  revision: string; reviewedAt: string; manufacturers: DecoderManufacturer[];
  families: DecoderFamily[]; models: DecoderModel[];
};
export type DecoderReferenceCoverage = {
  id: string; modelIds: string[]; status: "table" | "partial" | "unavailable"; rowCount: number; active: boolean;
};
export type DecoderCVReferenceRow = {
  cv: number; cvEnd?: number; label: string; valueRange?: string; factoryValue?: string;
  page: number; sourceIndex: number; indexScope: "base" | "indexed" | "unspecified";
  indexCV31?: number; indexCV32?: number; applicability: "model" | "manual";
};
export type DecoderCVReference = Omit<DecoderReferenceCoverage, "rowCount"> & {
  revision: string; title: DecoderText; note: DecoderText; sources: DecoderSource[]; rows: DecoderCVReferenceRow[];
};
export type DecoderLibrary = {
  schemaVersion: number; profiles: DecoderLibraryProfile[]; inventory?: DecoderInventory;
  referenceCoverage?: DecoderReferenceCoverage[];
};
export type DecoderLibraryStatus = "matched" | "identified" | "family" | "manufacturer" | "suggested" | "conflict" | "standard" | "unsupported" | "unavailable";
export type DecoderLibraryMatch = {
  status: DecoderLibraryStatus; profileId?: string; profileName?: string; revision?: string;
  manufacturerId?: number; firmware?: string; firmwareCompatible?: boolean;
  knownValues: number; totalValues: number;
  matchedBy?: "name" | "cv";
  identification?: { scope: string; name: string; cvNumber: number; value: number; modelId?: string };
};
export type DecoderReading = {
  cvNumber: number; value: number; indexCV31?: number | null; indexCV32?: number | null;
  decoderProfile?: string; description?: string; protocol?: string;
};

// One reviewed source is embedded in Go and used by standalone UI components.
// Application views use the persisted catalog from DecoderLibraryProvider.
function documentedAccess(value?: string): DecoderDefinition["access"] {
  if (value === undefined || value === "read-only" || value === "read-write" || value === "command") return value;
  throw new Error("Invalid bundled decoder register access");
}

export const bundledDecoderProfiles: DecoderLibraryProfile[] = catalog.profiles.map(profile => ({
  ...profile, active: true,
  definitions: profile.definitions.map(definition => ({
    ...definition, access: documentedAccess("access" in definition ? definition.access : undefined)
  }))
}));
export const bundledDecoderInventory: DecoderInventory = inventory;
