import type { DigitalCenterProvider, DigitalCenterWriteField } from "../features/digitalCenters/digitalCenterModel";

export type ECoSConnectionInput = {
  host: string;
  port?: number;
};

export type DigitalCenterConnectionInput = {
  host: string;
  port?: number;
};

export type DigitalCenterConnectionResult = {
  provider: DigitalCenterProvider;
  connected: boolean;
  host: string;
  port: number;
  status?: string;
  message: string;
  fields?: Record<string, string>;
};

export type DigitalCenterProbeCommandResult = {
  name: string;
  description: string;
  commandHex: string;
  request?: string;
  responseHex?: string;
  header?: string;
  payloadHex?: string;
  ok: boolean;
  error?: string;
  fields?: Record<string, string>;
};

export type DigitalCenterProbeResult = {
  provider: DigitalCenterProvider;
  connected: boolean;
  host: string;
  port: number;
  message: string;
  fields?: Record<string, string>;
  commands: DigitalCenterProbeCommandResult[];
};

export type ECoSConnectionResult = {
  connected: boolean;
  host: string;
  port: number;
  status?: string;
  protocolVersion?: string;
  applicationVersion?: string;
  hardwareVersion?: string;
  compatibilityStatus: "unreachable" | "rejected" | "partial" | "unverified" | "verified";
  missingFields: string[];
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
  functions?: Array<{
    index: number;
    description?: number;
  }>;
  cvs?: Array<{
    number: number;
    value: number;
  }>;
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

export type ECoSLocomotiveSyncInput = ECoSConnectionInput & {
  vehicleId: string;
  objectId?: number;
  dryRun?: boolean;
  confirm?: boolean;
};

export type ECoSLocomotiveSyncChange = {
  field: DigitalCenterWriteField;
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
