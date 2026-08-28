export type DigitalCenterProvider = "ecos" | "z21" | "intellibox3" | "cs3";
export type DigitalCenterCompareStatus = "ok" | "deviation" | "missing" | "new" | "conflict";
export type DigitalCenterCompareFilter = DigitalCenterCompareStatus | "all";
export type DigitalCenterWorkspaceTab = "live" | "diagnosis" | "messages";
export type DigitalCenterSessionState = "reading" | "ready" | "interrupted" | "failed";
export type DigitalCenterStationStatus = "read" | "incomplete" | "missing";
export type DigitalCenterMessageSeverity = "info" | "warning" | "error";
export type DigitalCenterMessageCode =
  | "connection.succeeded"
  | "connection.failed"
  | "connection.interrupted"
  | "read.started"
  | "read.completed"
  | "read.failed"
  | "parse.failed"
  | "capability.unavailable"
  | "live.started"
  | "live.stopped"
  | "live.interrupted"
  | "write.preview_failed"
  | "write.failed"
	| "write.unknown"
  | "write.verified"
	| "write.verification_failed"
	| "live.restart_failed";
export type ECoSLiveMonitorState = "stopped" | "running" | "interrupted";
export type DigitalCenterWriteField = "address" | "name" | "protocol";
export type DigitalCenterWriteDirection = "railkeeper_to_center";
export type DigitalCenterWriteOperation = "update" | "create";
export type DigitalCenterWriteResult = "verified" | "verification_failed" | "failed" | "unknown";

export type DigitalCenterCapabilities = {
  testConnection: boolean;
  readLocomotives: boolean;
  liveMonitor: boolean;
  writeLocomotives: boolean;
  writeCVs: boolean;
  diagnose: boolean;
};

export type DigitalCenterTransport = {
  id: "ecos_tcp" | "z21_udp" | "loconet_tcp" | "cs3_http";
  status: "available" | "planned";
  capabilities: DigitalCenterCapabilities;
};

export type DigitalCenterSummary = {
  provider: DigitalCenterProvider;
  name: string;
  active: boolean;
  selected: boolean;
  host: string;
  port: number;
  capabilities: DigitalCenterCapabilities;
  transports: DigitalCenterTransport[];
};

export type DigitalCenterWorkspaceSummary = {
  centers: DigitalCenterSummary[];
};

export type DigitalCenterReadSession = {
  id: string;
  provider: DigitalCenterProvider;
  state: DigitalCenterSessionState;
  host: string;
  port: number;
  capabilities: DigitalCenterCapabilities;
  readStartedAt: string;
  readCompletedAt: string;
  createdByUserId: string;
  createdAt: string;
  updatedAt: string;
};

export type DigitalCenterStationSnapshot = {
  objectId?: number;
  name?: string;
  decoderAddress?: number;
  protocol?: string;
};

export type DigitalCenterVehicleSnapshot = {
  vehicleId?: string;
  name?: string;
  decoderAddress?: number;
  protocol?: string;
};

export type DigitalCenterProposedMatch = {
  vehicleId?: string;
  match?: "address_protocol" | "address";
};

export type DigitalCenterWorkItem = {
  id: string;
  sessionId: string;
  centerObjectId: string;
  vehicleId: string;
  name: string;
  decoderAddress: number;
  protocol: string;
  compareStatus: DigitalCenterCompareStatus;
  stationStatus: DigitalCenterStationStatus;
  center: DigitalCenterStationSnapshot;
  railkeeper: DigitalCenterVehicleSnapshot;
  proposed: DigitalCenterProposedMatch;
  conflicts: DigitalCenterVehicleSnapshot[];
  createdAt: string;
  updatedAt: string;
};

export type DigitalCenterWorkItemFilter = {
  query: string;
  compareStatus: DigitalCenterCompareFilter;
  page: number;
  pageSize: number;
};

export type DigitalCenterWorkItemPage = {
  items: DigitalCenterWorkItem[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
};

export type DigitalCenterSessionMessage = {
  id: string;
  sessionId: string;
  severity: DigitalCenterMessageSeverity;
  code: DigitalCenterMessageCode;
  message: string;
  nextAction: string;
  createdAt: string;
};

export type DigitalCenterMessagesResponse = {
  messages: DigitalCenterSessionMessage[];
};

export type ECoSLivePulseSample = {
  at: string;
  repliesPerSecond: number;
};

export type ECoSLiveEvent = {
  at: string;
  kind: "event";
  message: string;
  protocol: string;
};

export type ECoSLiveDiagnosis = {
  connectionState: ECoSLiveMonitorState;
  lastSuccessfulCommunication?: string;
  lastError?: string;
  passive: boolean;
};

export type ECoSLiveStatus = {
  provider: DigitalCenterProvider;
  connected: boolean;
  state: ECoSLiveMonitorState;
  host?: string;
  port?: number;
  startedAt?: string;
  lastSeenAt?: string;
  lastMessage?: string;
  blocksReceived: number;
  repliesReceived: number;
  eventsReceived: number;
  subscriptionCommands?: string[];
  pulseSamples: ECoSLivePulseSample[];
  recentEvents: ECoSLiveEvent[];
  diagnosis: ECoSLiveDiagnosis;
  error?: string;
  message: string;
};

export type DigitalCenterWriteChange = {
  field: DigitalCenterWriteField;
  current: string;
  desired: string;
};

export type DigitalCenterWritePreviewInput = {
  operation?: DigitalCenterWriteOperation;
  fields?: DigitalCenterWriteField[];
};

export type DigitalCenterWritePreview = {
  sessionId: string;
  itemId: string;
  provider: DigitalCenterProvider;
  objectId: string;
  operation: DigitalCenterWriteOperation;
  direction: DigitalCenterWriteDirection;
  fields: DigitalCenterWriteField[];
  changes: DigitalCenterWriteChange[];
  token: string;
  expiresAt: string;
};

export type DigitalCenterWriteConfirmInput = {
  operation?: DigitalCenterWriteOperation;
  token: string;
  confirm: true;
  fields?: DigitalCenterWriteField[];
};

export type DigitalCenterWriteConfirmation = {
  sessionId: string;
  itemId: string;
  provider: DigitalCenterProvider;
  objectId: string;
  operation: DigitalCenterWriteOperation;
  direction: DigitalCenterWriteDirection;
  fields: DigitalCenterWriteField[];
  applied: boolean;
  verified: boolean;
  result: DigitalCenterWriteResult;
  message: string;
	verifiedValues?: {
	  name?: string;
	  address?: number;
	  protocol?: string;
	};
	liveMonitor: {
	  wasRunning: boolean;
	  restarted: boolean;
	};
	workItem?: DigitalCenterWorkItem;
};

export type DigitalCenterWorkspaceDialog =
  | { kind: "comparison"; itemId: string }
  | { kind: "write-preview"; itemId: string }
  | { kind: "assignment"; itemId: string };

export type DigitalCenterWorkspaceErrors = {
  workspace: string;
  live: string;
  read: string;
  worklist: string;
  detail: string;
  messages: string;
  write: string;
};

export const emptyDigitalCenterWorkItemPage: DigitalCenterWorkItemPage = {
  items: [],
  page: 1,
  pageSize: 10,
  total: 0,
  totalPages: 0
};
