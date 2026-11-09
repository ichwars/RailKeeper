

export type ExhibitionList = {
  id: string;
  designation: string;
  date: string;
  endDate: string;
  location?: string;
  description?: string;
  organizationNotes?: string;
  status: ExhibitionStatus;
  revision: number;
  locked: boolean;
  lockReason?: string;
  lockedAt?: string;
  completedAt?: string;
  archivedAt?: string;
  entryCount: number;
  entries?: ExhibitionEntry[];
  createdAt: string;
  updatedAt: string;
};

export type ExhibitionListInput = {
  designation: string;
  date: string;
  endDate?: string;
  location?: string;
  description?: string;
  organizationNotes?: string;
  status?: ExhibitionStatus;
  expectedRevision?: number;
};

export type ExhibitionStatus = "draft" | "open" | "locked" | "running" | "completed" | "archived";

export type ExhibitionEntryStatus = "ready" | "addressConflict" | "missing" | "check" | "unavailable";

export type ExhibitionEntry = {
  id: string;
  listId: string;
  vehicleId?: string;
  owner: string;
  imageUrl?: string;
  locomotiveName: string;
  gattung?: string;
  series?: string;
  manufacturer?: string;
  epoch?: string;
  railwayCompany?: string;
  dayScope: string;
  dtDecoder: boolean;
  decoderNumber?: string;
  decoderType?: string;
  adapter?: string;
  interfaceName?: string;
  sxAddress?: string;
  analog: boolean;
  availability: "available" | "unavailable";
  revision: number;
  functionKeys?: string;
  notes?: string;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
};

export type ExhibitionEntryInput = {
  vehicleId?: string;
  owner: string;
  imageUrl?: string;
  locomotiveName: string;
  gattung?: string;
  series?: string;
  manufacturer?: string;
  epoch?: string;
  railwayCompany?: string;
  dayScope?: string;
  dtDecoder: boolean;
  decoderNumber?: string;
  decoderType?: string;
  adapter?: string;
  interfaceName?: string;
  sxAddress?: string;
  analog?: boolean;
  availability?: "available" | "unavailable";
  expectedRevision?: number;
  functionKeys?: string;
  notes?: string;
  sortOrder?: number;
};

export type ExhibitionSummary = {
  entryCount: number;
  ownerCount: number;
  conflictCount: number;
  readyCount: number;
};

export type ExhibitionReadiness = {
  total: number;
  addressesChecked: number;
  functionsDocumented: number;
  imagesPresent: number;
  problems: number;
};

export type ExhibitionConflict = {
  id: string;
  kind: "address" | "missing" | "duplicateVehicle";
  entryIds: string[];
  fields?: string[];
  interfaceName?: string;
  address?: string;
  dayScopes?: string[];
  excepted: boolean;
  exceptionReason?: string;
};

export type ExhibitionWorkspaceEntry = ExhibitionEntry & {
  status: ExhibitionEntryStatus;
  conflictIds: string[];
};

export type ExhibitionWorkspace = {
  list: ExhibitionList;
  summary: ExhibitionSummary;
  readiness: ExhibitionReadiness;
  entries: ExhibitionWorkspaceEntry[];
  conflicts: ExhibitionConflict[];
  dayScopes: string[];
};

export type ExhibitionStatusInput = {
  status: ExhibitionStatus;
  expectedRevision?: number;
  confirmConflicts?: boolean;
  reason?: string;
};
