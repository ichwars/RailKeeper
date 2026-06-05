export type SetupStatus = {
  setupRequired: boolean;
};

export type CreateAdminRequest = {
  username: string;
  email: string;
  password: string;
};

export type LoginRequest = {
  username: string;
  password: string;
  twoFactorCode?: string;
};

export type Session = {
  username: string;
  roles: string[];
  csrfToken: string;
  twoFactorEnabled: boolean;
};

export type Role = {
  id: string;
  name: string;
};

export type UserAccount = {
  id: string;
  username: string;
  email?: string;
  roles: string[];
  createdAt: string;
  twoFactorEnabled: boolean;
};

export type UserAccountInput = {
  username: string;
  email?: string;
  password?: string;
  roles: string[];
};

export type PasswordResetRequest = {
  email: string;
};

export type PasswordResetResult = {
  status: string;
  message: string;
  expiresAt?: string;
};

export type PasswordResetConfirmRequest = {
  token: string;
  newPassword: string;
};

export type ChangePasswordInput = {
  currentPassword: string;
  newPassword: string;
};

export type TwoFactorStatus = {
  enabled: boolean;
  prepared: boolean;
  enabledAt?: string;
  username?: string;
  otpauthUrl?: string;
  secret?: string;
};

export type TwoFactorEnableInput = {
  code: string;
};

export type TwoFactorDisableInput = {
  currentPassword: string;
  code: string;
};

export type SMTPSettings = {
  enabled: boolean;
  publicUrl: string;
  host: string;
  port: string;
  username: string;
  from: string;
  tlsMode: string;
  passwordConfigured: boolean;
};

export type SMTPSettingsInput = {
  enabled: boolean;
  publicUrl: string;
  host: string;
  port: string;
  username: string;
  password?: string;
  from: string;
  tlsMode: string;
  clearPassword?: boolean;
};

export type SMTPTestRequest = {
  recipient: string;
};

export type SettingsPayload = {
  settings: Record<string, string>;
};

export type DigitalProviderSettings = {
  enabled: boolean;
  host: string;
  port: string;
};

export type DigitalCenterSettings = {
  provider: "ecos" | "z21" | "cs3";
  ecos: DigitalProviderSettings;
  z21: DigitalProviderSettings;
  cs3: DigitalProviderSettings;
};

export type SessionRecord = {
  id: string;
  userId: string;
  username: string;
  createdAt: string;
  expiresAt: string;
  revokedAt?: string;
  active: boolean;
};

export type Vehicle = {
  id: string;
  inventoryNumber: string;
  manufacturer: string;
  articleNumber?: string;
  articleSourceUrl?: string;
  name: string;
  gauge: string;
  epoch?: string;
  railwayCompany?: string;
  category?: string;
  gattung?: string;
  description?: string;
  series?: string;
  vehicleNumber?: string;
  digital: boolean;
  digitalDecoderNumber?: string;
  dtDecoder: boolean;
  dtDecoderNumber?: string;
  decoderType?: string;
  exhibitionReady: boolean;
  exhibition: boolean;
  abcBrakes: boolean;
  ean?: string;
  productionPeriod?: string;
  listPrice?: string;
  acquisitionType?: string;
  acquiredFrom?: string;
  purchasePrice?: string;
  purchaseDate?: string;
  storageLocation?: string;
  storageDetails?: string;
  condition?: string;
  conditionDetails?: string;
  packaging?: string;
  lengthMm?: string;
  weightG?: string;
  color?: string;
  lettering?: string;
  load?: string;
  interior?: string;
  axles?: string;
  axleCount?: string;
  tractionTireCount?: string;
  wheelset?: string;
  couplingSame: boolean;
  couplingFront?: string;
  couplingRear?: string;
  powerPickup?: string;
  adapter?: string;
  driveEnabled: boolean;
  driveDescription?: string;
  headlightsEnabled: boolean;
  headlightsDescription?: string;
  lightingEnabled: boolean;
  lightingDescription?: string;
  soundGeneratorEnabled: boolean;
  soundGeneratorDescription?: string;
  smokeGeneratorEnabled: boolean;
  smokeGeneratorDescription?: string;
  additionalInfo?: string;
  qrCodeEnabled: boolean;
  images?: VehicleImage[];
  attachments?: VehicleAttachment[];
  maintenance?: VehicleMaintenance[];
  spareParts?: VehicleSparePart[];
  functions?: VehicleFunction[];
  cvValues?: VehicleCVValue[];
  cvFiles?: VehicleCVFile[];
  externalMappings?: VehicleExternalMapping[];
  createdAt: string;
  updatedAt: string;
};

export type VehicleExternalMapping = {
  id: string;
  vehicleId: string;
  provider: string;
  externalId: string;
  externalName?: string;
  externalAddress?: string;
  externalProtocol?: string;
  syncStatus: string;
  lastSeenAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type VehicleExternalMappingInput = {
  provider: string;
  externalId: string;
  externalName?: string;
  externalAddress?: string;
  externalProtocol?: string;
  syncStatus?: string;
};

export type VehicleImage = {
  id: string;
  vehicleId: string;
  url: string;
  thumbnailUrl?: string;
  title?: string;
  sourceUrl?: string;
  fileName?: string;
  mimeType?: string;
  maintenanceId?: string;
  isPrimary: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt?: string;
};

export type VehicleImageInput = {
  id?: string;
  url: string;
  title?: string;
  sourceUrl?: string;
  maintenanceId?: string;
  isPrimary?: boolean;
  sortOrder?: number;
};

export type VehicleImageImportInput = {
  url: string;
  title?: string;
  sourceUrl?: string;
  maintenanceId?: string;
  isPrimary?: boolean;
  sortOrder?: number;
};

export type VehicleAttachment = {
  id: string;
  vehicleId: string;
  fileName: string;
  originalName: string;
  description?: string;
  category?: string;
  mimeType?: string;
  sizeBytes: number;
  maintenanceId?: string;
  createdAt: string;
  updatedAt: string;
};

export type VehicleAttachmentUpdateInput = {
  description?: string;
  category?: string;
  maintenanceId?: string;
};

export type VehicleMaintenance = {
  id: string;
  vehicleId: string;
  kind: string;
  status: string;
  conditionRating?: string;
  dueDate?: string;
  completedAt?: string;
  cost?: string;
  notes?: string;
  createdAt: string;
  updatedAt: string;
};

export type VehicleMaintenanceInput = {
  kind: string;
  status: string;
  conditionRating?: string;
  dueDate?: string;
  completedAt?: string;
  cost?: string;
  notes?: string;
};

export type VehicleSparePart = {
  id: string;
  vehicleId: string;
  articleNumber: string;
  description: string;
  price?: string;
  url?: string;
  createdAt: string;
  updatedAt: string;
};

export type VehicleSparePartInput = {
  articleNumber?: string;
  description?: string;
  price?: string;
  url?: string;
};


export type VehicleFunction = {
  id: string;
  vehicleId: string;
  functionKey: string;
  name?: string;
  symbolKey?: string;
  functionType: string;
  mode: string;
  directionDependent: boolean;
  notes?: string;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
};

export type VehicleFunctionInput = {
  name?: string;
  symbolKey?: string;
  functionType?: string;
  mode?: string;
  directionDependent?: boolean;
  notes?: string;
};

export type VehicleCVValue = {
  id: string;
  vehicleId: string;
  cvNumber: number;
  value: number;
  description?: string;
  category?: string;
  protocol?: string;
  decoderProfile?: string;
  sourceFileId?: string;
  createdAt: string;
  updatedAt: string;
  history?: VehicleCVValueHistory[];
};

export type VehicleCVValueHistory = {
  id: string;
  cvValueId: string;
  vehicleId: string;
  oldValue: number;
  newValue: number;
  changedAt: string;
};

export type VehicleCVValueInput = {
  cvNumber: number;
  value: number;
  description?: string;
  category?: string;
  protocol?: string;
  decoderProfile?: string;
  sourceFileId?: string;
};

export type VehicleCVFile = {
  id: string;
  vehicleId: string;
  fileName: string;
  originalName: string;
  description?: string;
  decoderProfile?: string;
  mimeType?: string;
  sizeBytes: number;
  createdAt: string;
  updatedAt: string;
};

export type VehicleCVFilePreview = {
  fileName: string;
  sizeBytes: number;
  mimeType: string;
  hasMetadata: boolean;
  projectName?: string;
  address?: string;
  type?: string;
  decoder?: string;
  manufacturer?: string;
  manufacturerId?: string;
  lokProgrammer?: string;
  suggestedDecoderProfile?: string;
  suggestedDescription?: string;
  suggestedCvValues?: {
    cvNumber: number;
    value: number;
    description?: string;
    category?: string;
    protocol?: string;
  }[];
  suggestedFunctions?: {
    functionKey: string;
    name: string;
    functionType?: string;
  }[];
  suggestedPreviewImage?: {
    mimeType: string;
    width: number;
    height: number;
    dataUrl: string;
  };
};

export type CreateVehicleRequest = {
  inventoryNumber?: string;
  manufacturer: string;
  articleNumber?: string;
  articleSourceUrl?: string;
  name: string;
  gauge: string;
  epoch?: string;
  railwayCompany?: string;
  category?: string;
  gattung?: string;
  description?: string;
  series?: string;
  vehicleNumber?: string;
  digital?: boolean;
  digitalDecoderNumber?: string;
  dtDecoder?: boolean;
  dtDecoderNumber?: string;
  decoderType?: string;
  exhibitionReady?: boolean;
  exhibition?: boolean;
  abcBrakes?: boolean;
  ean?: string;
  productionPeriod?: string;
  listPrice?: string;
  acquisitionType?: string;
  acquiredFrom?: string;
  purchasePrice?: string;
  purchaseDate?: string;
  storageLocation?: string;
  storageDetails?: string;
  condition?: string;
  conditionDetails?: string;
  packaging?: string;
  lengthMm?: string;
  weightG?: string;
  color?: string;
  lettering?: string;
  load?: string;
  interior?: string;
  axles?: string;
  axleCount?: string;
  tractionTireCount?: string;
  wheelset?: string;
  couplingSame?: boolean;
  couplingFront?: string;
  couplingRear?: string;
  powerPickup?: string;
  adapter?: string;
  driveEnabled?: boolean;
  driveDescription?: string;
  headlightsEnabled?: boolean;
  headlightsDescription?: string;
  lightingEnabled?: boolean;
  lightingDescription?: string;
  soundGeneratorEnabled?: boolean;
  soundGeneratorDescription?: string;
  smokeGeneratorEnabled?: boolean;
  smokeGeneratorDescription?: string;
  additionalInfo?: string;
  qrCodeEnabled?: boolean;
  images?: VehicleImageInput[];
};

export type MasterDataEntry = {
  id: string;
  type: string;
  key: string;
  label: string;
  active: boolean;
  sortOrder: number;
  sourceUrl?: string;
  metadata: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
};

export type MasterDataInput = {
  key?: string;
  label: string;
  active?: boolean;
  sortOrder?: number;
  sourceUrl?: string;
  metadata?: Record<string, unknown>;
};

export type MasterDataRelation = {
  id: string;
  parentType: string;
  parentKey: string;
  childType: string;
  childKey: string;
  sortOrder: number;
};

export type MasterDataDocument = {
  format: string;
  version: number;
  createdAt: string;
  entries: Record<string, MasterDataEntry[]>;
  relations: MasterDataRelation[];
};

export type MasterDataImportResult = {
  importedTypes: number;
  importedEntries: number;
  importedRelations: number;
};

export type InventoryNumberScheme = {
  id: string;
  category: string;
  prefix: string;
  nextNumber: number;
  padding: number;
  active: boolean;
  preview: string;
  createdAt: string;
  updatedAt: string;
};

export type InventoryNumberSchemeInput = {
  prefix: string;
  nextNumber: number;
  padding: number;
  active: boolean;
};

export type InventoryNumberSchemeCreateInput = InventoryNumberSchemeInput & {
  category: string;
};

export type ArticleSearchInput = {
  manufacturer?: string;
  articleNumber?: string;
  name?: string;
  gauge?: string;
  searchSources?: string[];
  fields?: Record<string, string>;
};

export type ArticleSearchField = {
  label: string;
  value: string;
  confidence: number;
};

export type ArticleSearchImage = {
  url: string;
  title: string;
  source: string;
};

export type ECoSImageCandidate = {
  key: string;
  value: string;
  kind: "url" | "data" | "base64" | "id" | "reference" | string;
  previewUrl?: string;
  mimeType?: string;
  transferable: boolean;
};

export type ECoSImageSuggestion = ArticleSearchImage & {
  id: string;
  thumbnailUrl?: string;
  mimeType?: string;
  ecosKey: string;
  ecosKind: string;
  rawValue: string;
  isPrimary?: boolean;
};

export type ArticleSearchSparePart = {
  articleNumber: string;
  description: string;
  price?: string;
  url?: string;
  source?: string;
  availability?: string;
};

export type ArticleSearchDocument = {
  title: string;
  url: string;
  source?: string;
  kind?: string;
};

export type VehicleAttachmentImportInput = {
  url: string;
  title?: string;
  description?: string;
  category?: string;
  maintenanceId?: string;
};

export type ArticleSearchResultTrace = {
  detailLoaded: boolean;
  detailFields: number;
  detailImages: number;
  detailSpareParts?: number;
  detailDocuments?: number;
  finalUrl?: string;
  error?: string;
};

export type ArticleSearchResult = {
  source: string;
  title: string;
  url: string;
  snippet: string;
  score: number;
  fields: Record<string, ArticleSearchField>;
  images?: ArticleSearchImage[];
  spareParts?: ArticleSearchSparePart[];
  documents?: ArticleSearchDocument[];
  trace?: ArticleSearchResultTrace;
  conflicts?: string[];
};

export type ArticleSearchQueryInfo = {
  source: string;
  query: string;
};

export type ArticleSearchResponse = {
  query: string;
  sources?: string[];
  manufacturerDomains?: string[];
  queries?: ArticleSearchQueryInfo[];
  results: ArticleSearchResult[];
};

export type ECoSConnectionInput = {
  host: string;
  port?: number;
};

export type DigitalCenterConnectionInput = {
  host: string;
  port?: number;
};

export type DigitalCenterConnectionResult = {
  provider: "z21" | "cs3" | string;
  connected: boolean;
  host: string;
  port: number;
  status?: string;
  message: string;
  fields?: Record<string, string>;
};

export type ECoSConnectionResult = {
  connected: boolean;
  host: string;
  port: number;
  status?: string;
  protocolVersion?: string;
  applicationVersion?: string;
  hardwareVersion?: string;
  message: string;
  rawLines?: string[];
  fields?: Record<string, string>;
};

export type ECoSRawCommandProbe = {
  command: string;
  fields: string[];
  ok: boolean;
  status?: string;
  error?: string;
  rawLines?: string[];
  attributes?: Record<string, string[]>;
};

export type ECoSRawLocomotive = {
  objectId: number;
  name?: string;
  address?: number;
  protocol?: string;
  profile?: string;
  speed?: number;
  speedStep?: number;
  direction?: number;
  functionSet?: string;
  numberOfFunctions?: number;
  functions?: Array<{
    index: number;
    active: boolean;
    description?: number;
  }>;
  cvs?: Array<{
    number: number;
    value: number;
  }>;
  imageCandidates?: ECoSImageCandidate[];
  attributes?: Record<string, string[]>;
  supportedFields?: string[];
  missingFields?: string[];
  interestingFields?: string[];
  probes?: ECoSRawCommandProbe[];
  detailError?: string;
};

export type ECoSRawProbe = {
  host: string;
  port: number;
  probeFields: string[];
  locomotives: ECoSRawLocomotive[];
  rawLines?: string[];
  message: string;
};

export type ECoSLocomotiveSummary = {
  host: string;
  port: number;
  count: number;
  message: string;
};

export type ECoSLiveStatus = {
  provider: string;
  connected: boolean;
  host?: string;
  port?: number;
  startedAt?: string;
  lastSeenAt?: string;
  lastMessage?: string;
  blocksReceived: number;
  repliesReceived: number;
  eventsReceived: number;
  subscriptionCommands?: string[];
  error?: string;
  message: string;
};

export type ECoSLocomotiveSyncInput = ECoSConnectionInput & {
  vehicleId: string;
  objectId?: number;
  dryRun?: boolean;
  confirm?: boolean;
};

export type ECoSLocomotiveSyncChange = {
  field: "name" | "address" | "protocol" | string;
  current: string;
  desired: string;
};

export type ECoSLocomotiveSyncResult = {
  host: string;
  port: number;
  objectId: number;
  dryRun: boolean;
  applied: boolean;
  current: {
    name?: string;
    address?: number;
    protocol?: string;
  };
  desired: {
    name?: string;
    address?: number;
    protocol?: string;
  };
  changes: ECoSLocomotiveSyncChange[];
  commands?: string[];
  rawLines?: string[];
  message: string;
};

export type BackupImportResult = {
  restoredTables: number;
  restoredRows: number;
  restoredFiles: number;
};

export type BackupValidationTable = {
  name: string;
  rows: number;
  missing: boolean;
  unknownColumns?: string[];
};

export type BackupValidationResult = {
  compatible: boolean;
  format?: string;
  version: number;
  createdAt?: string;
  tableCount: number;
  rowCount: number;
  fileCount: number;
  fileBytes: number;
  tables: BackupValidationTable[];
  warnings: string[];
  errors: string[];
};

export type VersionInfo = {
  version: string;
  latestVersion?: string;
  updateAvailable: boolean;
  sourceUrl?: string;
  releaseUrl?: string;
  releaseNotes?: string;
  assetUrl?: string;
  assetName?: string;
  checkedAt: string;
  status: "local" | "not_configured" | "current" | "update_available" | "unavailable" | "no_release";
  message: string;
};

export type StorageUsageCategory = {
  key: string;
  label: string;
  bytes: number;
  files: number;
};

export type StorageUsage = {
  totalBytes: number;
  categories: StorageUsageCategory[];
  updatedAt: string;
};

export type StorageOptimizeResult = {
  beforeBytes: number;
  afterBytes: number;
  reclaimedBytes: number;
  optimizedAt: string;
};

export type SystemPrinter = {
  id: string;
  name: string;
  isDefault: boolean;
};

export type SystemPrinters = {
  status: "available" | "configured" | "unavailable";
  message: string;
  defaultPrinter?: string;
  printers: SystemPrinter[];
};

export type AuditLogEntry = {
  id: string;
  actorUserId?: string;
  actorUsername?: string;
  action: string;
  targetType?: string;
  targetId?: string;
  createdAt: string;
  detailsJson: string;
};

export type AuditLogResponse = {
  entries: AuditLogEntry[];
};

export type ExhibitionList = {
  id: string;
  designation: string;
  date: string;
  locked: boolean;
  entryCount: number;
  entries?: ExhibitionEntry[];
  createdAt: string;
  updatedAt: string;
};

export type ExhibitionListInput = {
  designation: string;
  date: string;
};

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
  sxAddress?: string;
  analog: boolean;
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
  sxAddress?: string;
  analog?: boolean;
  functionKeys?: string;
  notes?: string;
  sortOrder?: number;
};

let csrfToken = "";

export class ApiError extends Error {
  code: string;
  status: number;

  constructor(message: string, code: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
  }
}

type RequestOptions = {
  retries?: number;
  timeoutMs?: number;
};

function readCookie(name: string): string {
  const prefix = `${name}=`;
  const value = document.cookie
    .split("; ")
    .find((part) => part.startsWith(prefix))
    ?.slice(prefix.length);
  return value ? decodeURIComponent(value) : "";
}

async function request<T>(path: string, init: RequestInit = {}, options: RequestOptions = {}): Promise<T> {
  const method = (init.method || "GET").toUpperCase();
  const isFormData = typeof FormData !== "undefined" && init.body instanceof FormData;
  const headers: Record<string, string> = {
    ...(!["GET", "HEAD"].includes(method) && !isFormData ? { "Content-Type": "application/json" } : {}),
    ...((init.headers as Record<string, string>) || {})
  };
  const timeoutMs = options.timeoutMs || 12000;
  const attempts = 1 + (["GET", "HEAD"].includes(method) ? options.retries || 0 : 0);

  if (!["GET", "HEAD", "OPTIONS"].includes(method)) {
    const token = csrfToken || readCookie("rk_csrf");
    if (token) {
      headers["X-CSRF-Token"] = token;
    }
  }

  for (let attempt = 0; attempt < attempts; attempt += 1) {
    const controller = new AbortController();
    const timeout = window.setTimeout(() => controller.abort(), timeoutMs);

    try {
      const response = await fetch(`/api/v1${path}`, {
        credentials: "include",
        ...init,
        headers,
        signal: init.signal || controller.signal
      });

      if (!response.ok) {
        let message = response.statusText;
        let code = "request_failed";
        try {
          const body = await response.json();
          code = body.error || code;
          message = body.message || body.error || message;
        } catch {
          // Keep the HTTP status text when the server did not return JSON.
        }
        throw new ApiError(message, code, response.status);
      }

      if (response.status === 204) {
        return undefined as T;
      }

      return response.json() as Promise<T>;
    } catch (error) {
      if (attempt + 1 < attempts && error instanceof DOMException && error.name === "AbortError") {
        continue;
      }
      if (error instanceof DOMException && error.name === "AbortError") {
        throw new Error("Die Anfrage hat zu lange gedauert. Bitte erneut versuchen.");
      }
      throw error;
    } finally {
      window.clearTimeout(timeout);
    }
  }

  throw new Error("Die Anfrage konnte nicht verarbeitet werden.");
}

export const api = {
  setupStatus: () => request<SetupStatus>("/setup/status"),
  createAdmin: (input: CreateAdminRequest) =>
    request<{ status: string }>("/setup/admin", {
      method: "POST",
      body: JSON.stringify(input)
    }),
  login: async (input: LoginRequest) => {
    const session = await request<Session>("/auth/login", {
      method: "POST",
      body: JSON.stringify(input)
    });
    csrfToken = session.csrfToken || readCookie("rk_csrf");
    return session;
  },
  requestPasswordReset: (input: PasswordResetRequest) =>
    request<PasswordResetResult>("/auth/password-reset", {
      method: "POST",
      body: JSON.stringify(input)
    }),
  confirmPasswordReset: (input: PasswordResetConfirmRequest) =>
    request<void>("/auth/password-reset/confirm", {
      method: "POST",
      body: JSON.stringify(input)
    }),
  session: async () => {
    const session = await request<Session>("/auth/session");
    csrfToken = session.csrfToken || readCookie("rk_csrf");
    return session;
  },
  profileSettings: () => request<SettingsPayload>("/profile/settings"),
  updateProfileSettings: (settings: Record<string, string>) =>
    request<SettingsPayload>("/profile/settings", {
      method: "PUT",
      body: JSON.stringify({ settings })
    }),
  logout: async () => {
    await request<void>("/auth/logout", { method: "POST" });
    csrfToken = "";
  },
  changePassword: (input: ChangePasswordInput) =>
    request<void>("/auth/password", {
      method: "PUT",
      body: JSON.stringify(input)
    }),
  twoFactorStatus: () => request<TwoFactorStatus>("/auth/two-factor"),
  setupTwoFactor: () =>
    request<TwoFactorStatus>("/auth/two-factor/setup", {
      method: "POST"
    }),
  enableTwoFactor: (input: TwoFactorEnableInput) =>
    request<TwoFactorStatus>("/auth/two-factor/enable", {
      method: "POST",
      body: JSON.stringify(input)
    }),
  disableTwoFactor: (input: TwoFactorDisableInput) =>
    request<void>("/auth/two-factor/disable", {
      method: "POST",
      body: JSON.stringify(input)
    }),
  roles: () => request<Role[]>("/roles"),
  users: () => request<UserAccount[]>("/users"),
  createUser: (input: UserAccountInput) =>
    request<UserAccount>("/users", {
      method: "POST",
      body: JSON.stringify(input)
    }),
  updateUser: (id: string, input: UserAccountInput) =>
    request<UserAccount>(`/users/${encodeURIComponent(id)}`, {
      method: "PUT",
      body: JSON.stringify(input)
    }),
  deleteUser: (id: string) =>
    request<void>(`/users/${encodeURIComponent(id)}`, {
      method: "DELETE"
    }),
  sessions: (limit = 5) => request<SessionRecord[]>(`/sessions?limit=${encodeURIComponent(String(limit))}`),
  revokeSession: (id: string) =>
    request<void>(`/sessions/${encodeURIComponent(id)}/revoke`, {
      method: "PUT"
    }),
  version: (check = false, includePrerelease = false) =>
    request<VersionInfo>(
      `/version${check ? `?check=true&prerelease=${includePrerelease ? "true" : "false"}` : ""}`,
      {},
      { timeoutMs: 10000 }
    ),
  storageUsage: () => request<StorageUsage>("/system/storage", {}, { timeoutMs: 30000 }),
  optimizeStorage: () => request<StorageOptimizeResult>("/system/storage/optimize", { method: "POST" }, { timeoutMs: 120000 }),
  systemPrinters: () => request<SystemPrinters>("/system/printers", {}, { timeoutMs: 10000 }),
  auditLog: (limit = 50) => request<AuditLogResponse>(`/system/audit-log?limit=${encodeURIComponent(String(limit))}`, {}, { timeoutMs: 10000 }),
  smtpSettings: () => request<SMTPSettings>("/system/smtp", {}, { timeoutMs: 10000 }),
  updateSMTPSettings: (input: SMTPSettingsInput) =>
    request<SMTPSettings>("/system/smtp", {
      method: "PUT",
      body: JSON.stringify(input)
    }),
  testSMTPSettings: (input: SMTPTestRequest) =>
    request<{ status: string }>("/system/smtp/test", {
      method: "POST",
      body: JSON.stringify(input)
    }, { timeoutMs: 30000 }),
  digitalSettings: () => request<DigitalCenterSettings>("/system/digital-settings", {}, { timeoutMs: 10000 }),
  updateDigitalSettings: (input: DigitalCenterSettings) =>
    request<DigitalCenterSettings>("/system/digital-settings", {
      method: "PUT",
      body: JSON.stringify(input)
    }),
  exhibitionLists: () => request<ExhibitionList[]>("/exhibition-lists"),
  exhibitionList: (id: string) => request<ExhibitionList>(`/exhibition-lists/${encodeURIComponent(id)}`),
  createExhibitionList: (input: ExhibitionListInput) =>
    request<ExhibitionList>("/exhibition-lists", {
      method: "POST",
      body: JSON.stringify(input)
    }),
  updateExhibitionList: (id: string, input: ExhibitionListInput) =>
    request<ExhibitionList>(`/exhibition-lists/${encodeURIComponent(id)}`, {
      method: "PUT",
      body: JSON.stringify(input)
    }),
  deleteExhibitionList: (id: string) =>
    request<void>(`/exhibition-lists/${encodeURIComponent(id)}`, {
      method: "DELETE"
    }),
  setExhibitionListLocked: (id: string, locked: boolean) =>
    request<ExhibitionList>(`/exhibition-lists/${encodeURIComponent(id)}/lock`, {
      method: "PUT",
      body: JSON.stringify({ locked })
    }),
  exhibitionEntries: (listId: string) => request<ExhibitionEntry[]>(`/exhibition-lists/${encodeURIComponent(listId)}/entries`),
  createExhibitionEntry: (listId: string, input: ExhibitionEntryInput) =>
    request<ExhibitionEntry>(`/exhibition-lists/${encodeURIComponent(listId)}/entries`, {
      method: "POST",
      body: JSON.stringify(input)
    }),
  updateExhibitionEntry: (listId: string, entryId: string, input: ExhibitionEntryInput) =>
    request<ExhibitionEntry>(`/exhibition-lists/${encodeURIComponent(listId)}/entries/${encodeURIComponent(entryId)}`, {
      method: "PUT",
      body: JSON.stringify(input)
    }),
  deleteExhibitionEntry: (listId: string, entryId: string) =>
    request<void>(`/exhibition-lists/${encodeURIComponent(listId)}/entries/${encodeURIComponent(entryId)}`, {
      method: "DELETE"
    }),
  vehicles: (query = "") =>
    request<Vehicle[]>(`/vehicles${query ? `?q=${encodeURIComponent(query)}` : ""}`),
  createVehicle: (input: CreateVehicleRequest) =>
    request<Vehicle>("/vehicles", {
      method: "POST",
      body: JSON.stringify(input)
    }),
  vehicle: (id: string) => request<Vehicle>(`/vehicles/${encodeURIComponent(id)}`),
  updateVehicle: (id: string, input: CreateVehicleRequest) =>
    request<Vehicle>(`/vehicles/${encodeURIComponent(id)}`, {
      method: "PUT",
      body: JSON.stringify(input)
    }),
  deleteVehicle: (id: string) =>
    request<void>(`/vehicles/${encodeURIComponent(id)}`, {
      method: "DELETE"
    }),
  uploadVehicleImage: (vehicleId: string, file: File, title = "", isPrimary = false, maintenanceId = "") => {
    const form = new FormData();
    form.append("file", file);
    form.append("title", title);
    form.append("isPrimary", String(isPrimary));
    form.append("maintenanceId", maintenanceId);
    return request<VehicleImage>(
      `/vehicles/${encodeURIComponent(vehicleId)}/images`,
      {
        method: "POST",
        body: form
      },
      { timeoutMs: 30000 }
    );
  },
  importVehicleImageFromUrl: (vehicleId: string, input: VehicleImageImportInput) =>
    request<VehicleImage>(
      `/vehicles/${encodeURIComponent(vehicleId)}/images/import-url`,
      {
        method: "POST",
        body: JSON.stringify(input)
      },
      { timeoutMs: 30000 }
    ),
  deleteVehicleImage: (vehicleId: string, imageId: string) =>
    request<void>(`/vehicles/${encodeURIComponent(vehicleId)}/images/${encodeURIComponent(imageId)}`, {
      method: "DELETE"
    }),
  uploadVehicleAttachment: (vehicleId: string, file: File, category = "", description = "", maintenanceId = "") => {
    const form = new FormData();
    form.append("file", file);
    form.append("category", category);
    form.append("description", description);
    form.append("maintenanceId", maintenanceId);
    return request<VehicleAttachment>(
      `/vehicles/${encodeURIComponent(vehicleId)}/attachments`,
      {
        method: "POST",
        body: form
      },
      { timeoutMs: 30000 }
    );
  },
  updateVehicleAttachment: (vehicleId: string, attachmentId: string, input: VehicleAttachmentUpdateInput) =>
    request<VehicleAttachment>(
      `/vehicles/${encodeURIComponent(vehicleId)}/attachments/${encodeURIComponent(attachmentId)}`,
      {
        method: "PUT",
        body: JSON.stringify(input)
      }
    ),
  deleteVehicleAttachment: (vehicleId: string, attachmentId: string) =>
    request<void>(`/vehicles/${encodeURIComponent(vehicleId)}/attachments/${encodeURIComponent(attachmentId)}`, {
      method: "DELETE"
    }),
  vehicleAttachmentDownloadUrl: (vehicleId: string, attachmentId: string) =>
    `/api/v1/vehicles/${encodeURIComponent(vehicleId)}/attachments/${encodeURIComponent(attachmentId)}/download`,
  importVehicleAttachmentFromUrl: (vehicleId: string, input: VehicleAttachmentImportInput) =>
    request<VehicleAttachment>(`/vehicles/${encodeURIComponent(vehicleId)}/attachments/import-url`, {
      method: "POST",
      body: JSON.stringify(input)
    }),
  vehicleMaintenance: (vehicleId: string) =>
    request<VehicleMaintenance[]>(`/vehicles/${encodeURIComponent(vehicleId)}/maintenance`),
  createVehicleMaintenance: (vehicleId: string, input: VehicleMaintenanceInput) =>
    request<VehicleMaintenance>(`/vehicles/${encodeURIComponent(vehicleId)}/maintenance`, {
      method: "POST",
      body: JSON.stringify(input)
    }),
  updateVehicleMaintenance: (vehicleId: string, maintenanceId: string, input: VehicleMaintenanceInput) =>
    request<VehicleMaintenance>(
      `/vehicles/${encodeURIComponent(vehicleId)}/maintenance/${encodeURIComponent(maintenanceId)}`,
      {
        method: "PUT",
        body: JSON.stringify(input)
      }
    ),
  deleteVehicleMaintenance: (vehicleId: string, maintenanceId: string) =>
    request<void>(`/vehicles/${encodeURIComponent(vehicleId)}/maintenance/${encodeURIComponent(maintenanceId)}`, {
      method: "DELETE"
    }),
  vehicleSpareParts: (vehicleId: string) =>
    request<VehicleSparePart[]>(`/vehicles/${encodeURIComponent(vehicleId)}/spare-parts`),
  vehicleSparePartSuggestions: (vehicleId: string, attachmentId = "") =>
    request<ArticleSearchSparePart[]>(
      `/vehicles/${encodeURIComponent(vehicleId)}/spare-parts/suggestions${attachmentId ? `?attachmentId=${encodeURIComponent(attachmentId)}` : ""}`,
      undefined,
      { timeoutMs: 15000 }
    ),
  createVehicleSparePart: (vehicleId: string, input: VehicleSparePartInput) =>
    request<VehicleSparePart>(`/vehicles/${encodeURIComponent(vehicleId)}/spare-parts`, {
      method: "POST",
      body: JSON.stringify(input)
    }),
  updateVehicleSparePart: (vehicleId: string, sparePartId: string, input: VehicleSparePartInput) =>
    request<VehicleSparePart>(
      `/vehicles/${encodeURIComponent(vehicleId)}/spare-parts/${encodeURIComponent(sparePartId)}`,
      {
        method: "PUT",
        body: JSON.stringify(input)
      }
    ),
  deleteVehicleSparePart: (vehicleId: string, sparePartId: string) =>
    request<void>(`/vehicles/${encodeURIComponent(vehicleId)}/spare-parts/${encodeURIComponent(sparePartId)}`, {
      method: "DELETE"
    }),
  vehicleFunctions: (vehicleId: string) =>
    request<VehicleFunction[]>(`/vehicles/${encodeURIComponent(vehicleId)}/functions`),
  updateVehicleFunction: (vehicleId: string, functionKey: string, input: VehicleFunctionInput) =>
    request<VehicleFunction>(`/vehicles/${encodeURIComponent(vehicleId)}/functions/${encodeURIComponent(functionKey)}`, {
      method: "PUT",
      body: JSON.stringify(input)
    }),
  deleteVehicleFunction: (vehicleId: string, functionKey: string) =>
    request<void>(`/vehicles/${encodeURIComponent(vehicleId)}/functions/${encodeURIComponent(functionKey)}`, {
      method: "DELETE"
    }),
  vehicleCVValues: (vehicleId: string) =>
    request<VehicleCVValue[]>(`/vehicles/${encodeURIComponent(vehicleId)}/cv-values`),
  createVehicleCVValue: (vehicleId: string, input: VehicleCVValueInput) =>
    request<VehicleCVValue>(`/vehicles/${encodeURIComponent(vehicleId)}/cv-values`, {
      method: "POST",
      body: JSON.stringify(input)
    }),
  updateVehicleCVValue: (vehicleId: string, cvValueId: string, input: VehicleCVValueInput) =>
    request<VehicleCVValue>(`/vehicles/${encodeURIComponent(vehicleId)}/cv-values/${encodeURIComponent(cvValueId)}`, {
      method: "PUT",
      body: JSON.stringify(input)
    }),
  deleteVehicleCVValue: (vehicleId: string, cvValueId: string) =>
    request<void>(`/vehicles/${encodeURIComponent(vehicleId)}/cv-values/${encodeURIComponent(cvValueId)}`, {
      method: "DELETE"
    }),
  uploadVehicleCVFile: (vehicleId: string, file: File, decoderProfile = "", description = "") => {
    const form = new FormData();
    form.append("file", file);
    form.append("decoderProfile", decoderProfile);
    form.append("description", description);
    return request<VehicleCVFile>(
      `/vehicles/${encodeURIComponent(vehicleId)}/cv-files`,
      {
        method: "POST",
        body: form
      },
      { timeoutMs: 30000 }
    );
  },
  previewVehicleCVFile: (file: File) => {
    const form = new FormData();
    form.append("file", file);
    return request<VehicleCVFilePreview>(
      "/cv-files/preview",
      {
        method: "POST",
        body: form
      },
      { timeoutMs: 30000 }
    );
  },
  deleteVehicleCVFile: (vehicleId: string, cvFileId: string) =>
    request<void>(`/vehicles/${encodeURIComponent(vehicleId)}/cv-files/${encodeURIComponent(cvFileId)}`, {
      method: "DELETE"
    }),
  vehicleCVFileDownloadUrl: (vehicleId: string, cvFileId: string) =>
    `/api/v1/vehicles/${encodeURIComponent(vehicleId)}/cv-files/${encodeURIComponent(cvFileId)}/download`,
  inventoryNumberSchemes: () => request<InventoryNumberScheme[]>("/inventory-number-schemes"),
  createInventoryNumberScheme: (input: InventoryNumberSchemeCreateInput) =>
    request<InventoryNumberScheme>("/inventory-number-schemes", {
      method: "POST",
      body: JSON.stringify(input)
    }),
  updateInventoryNumberScheme: (category: string, input: InventoryNumberSchemeInput) =>
    request<InventoryNumberScheme>(`/inventory-number-schemes/${encodeURIComponent(category)}`, {
      method: "PUT",
      body: JSON.stringify(input)
    }),
  upsertVehicleExternalMapping: (vehicleId: string, input: VehicleExternalMappingInput) =>
    request<VehicleExternalMapping>(`/vehicles/${encodeURIComponent(vehicleId)}/external-mappings`, {
      method: "POST",
      body: JSON.stringify(input)
    }),
  articleSearch: (input: ArticleSearchInput) =>
    request<ArticleSearchResponse>(
      "/article-search",
      {
        method: "POST",
        body: JSON.stringify(input)
      },
      { timeoutMs: 15000 }
    ),
  testECoSConnection: (input: ECoSConnectionInput) =>
    request<ECoSConnectionResult>(
      "/ecos/test",
      {
        method: "POST",
        body: JSON.stringify(input)
      },
      { timeoutMs: 10000 }
    ),
  probeECoSLocomotiveRaw: (input: ECoSConnectionInput) =>
    request<ECoSRawProbe>(
      "/ecos/locomotives/raw",
      {
        method: "POST",
        body: JSON.stringify(input)
      },
      { timeoutMs: 120000 }
    ),
  countECoSLocomotives: (input: ECoSConnectionInput) =>
    request<ECoSLocomotiveSummary>(
      "/ecos/locomotives/count",
      {
        method: "POST",
        body: JSON.stringify(input)
      },
      { timeoutMs: 10000 }
    ),
  getECoSLiveStatus: () => request<ECoSLiveStatus>("/digital-centers/ecos/live/status"),
  syncECoSLocomotive: (input: ECoSLocomotiveSyncInput) =>
    request<ECoSLocomotiveSyncResult>(
      "/digital-centers/ecos/locomotives/sync",
      {
        method: "POST",
        body: JSON.stringify(input)
      },
      { timeoutMs: 30000 }
    ),
  startECoSLive: (input: ECoSConnectionInput) =>
    request<ECoSLiveStatus>(
      "/digital-centers/ecos/live/start",
      {
        method: "POST",
        body: JSON.stringify(input)
      },
      { timeoutMs: 10000 }
    ),
  stopECoSLive: () =>
    request<ECoSLiveStatus>("/digital-centers/ecos/live/stop", {
      method: "POST"
    }),
  testZ21Connection: (input: DigitalCenterConnectionInput) =>
    request<DigitalCenterConnectionResult>(
      "/digital-centers/z21/test",
      {
        method: "POST",
        body: JSON.stringify(input)
      },
      { timeoutMs: 10000 }
    ),
  testCS3Connection: (input: DigitalCenterConnectionInput) =>
    request<DigitalCenterConnectionResult>(
      "/digital-centers/cs3/test",
      {
        method: "POST",
        body: JSON.stringify(input)
      },
      { timeoutMs: 10000 }
    ),
  backupExportUrl: () => "/api/v1/backup/export",
  validateBackup: (file: File) => {
    const form = new FormData();
    form.append("file", file);
    return request<BackupValidationResult>(
      "/backup/validate",
      {
        method: "POST",
        body: form
      },
      { timeoutMs: 120000 }
    );
  },
  restoreBackup: (file: File) => {
    const form = new FormData();
    form.append("file", file);
    return request<BackupImportResult>(
      "/backup/restore",
      {
        method: "POST",
        body: form
      },
      { timeoutMs: 120000 }
    );
  },
  masterData: (type: string, activeOnly = false) =>
    request<MasterDataEntry[]>(
      `/master-data/${encodeURIComponent(type)}${activeOnly ? "?active=true" : ""}`
    ),
  masterDataAll: (activeOnly = false) =>
    request<Record<string, MasterDataEntry[]>>(
      `/master-data-all${activeOnly ? "?active=true" : ""}`,
      {},
      { retries: 1, timeoutMs: 30000 }
    ),
  createMasterData: (type: string, input: MasterDataInput) =>
    request<MasterDataEntry>(`/master-data/${encodeURIComponent(type)}`, {
      method: "POST",
      body: JSON.stringify(input)
    }),
  updateMasterData: (type: string, key: string, input: MasterDataInput) =>
    request<MasterDataEntry>(`/master-data/${encodeURIComponent(type)}/${encodeURIComponent(key)}`, {
      method: "PUT",
      body: JSON.stringify(input)
    }),
  deleteMasterData: (type: string, key: string) =>
    request<void>(`/master-data/${encodeURIComponent(type)}/${encodeURIComponent(key)}`, {
      method: "DELETE"
    }),
  masterDataRelations: (parentType: string, childType: string) =>
    request<MasterDataRelation[]>(
      `/master-data-relations?parentType=${encodeURIComponent(parentType)}&childType=${encodeURIComponent(childType)}`
    ),
  masterDataExportUrl: () => "/api/v1/master-data/export",
  importMasterData: (file: File) => {
    const form = new FormData();
    form.append("file", file);
    return request<MasterDataImportResult>(
      "/master-data/import",
      {
        method: "POST",
        body: form
      },
      { timeoutMs: 120000 }
    );
  }
};
