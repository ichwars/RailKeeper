import type {
  Vehicle,
  VehicleCVFile,
  VehicleCVValue,
  VehicleFunction,
  VehicleImage,
  VehicleMaintenance
} from "../../shared/api";

const timestamp = "2026-07-23T08:00:00Z";

export function imageFixture(overrides: Partial<VehicleImage> = {}): VehicleImage {
  return {
    id: "image-1",
    vehicleId: "vehicle-1",
    url: "/api/v1/vehicles/vehicle-1/images/image-1/file",
    isPrimary: true,
    sortOrder: 0,
    createdAt: timestamp,
    ...overrides
  };
}

export function maintenanceFixture(overrides: Partial<VehicleMaintenance> = {}): VehicleMaintenance {
  return {
    id: "maintenance-1",
    vehicleId: "vehicle-1",
    kind: "Wartung",
    status: "offen",
    dueDate: "2026-07-24",
    createdAt: timestamp,
    updatedAt: timestamp,
    ...overrides
  };
}

export function functionFixture(overrides: Partial<VehicleFunction> = {}): VehicleFunction {
  return {
    id: "function-1",
    vehicleId: "vehicle-1",
    functionKey: "F0",
    name: "Licht",
    functionType: "licht",
    mode: "dauer",
    directionDependent: true,
    sortOrder: 0,
    createdAt: timestamp,
    updatedAt: timestamp,
    ...overrides
  };
}

export function cvValueFixture(overrides: Partial<VehicleCVValue> = {}): VehicleCVValue {
  return {
    id: "cv-1",
    vehicleId: "vehicle-1",
    cvNumber: 1,
    value: 3,
    description: "Adresse",
    decoderProfile: "ESU LokPilot 5",
    createdAt: timestamp,
    updatedAt: timestamp,
    ...overrides
  };
}

export function cvFileFixture(overrides: Partial<VehicleCVFile> = {}): VehicleCVFile {
  return {
    id: "cv-file-1",
    vehicleId: "vehicle-1",
    fileName: "decoder.esux",
    originalName: "decoder.esux",
    sizeBytes: 1024,
    createdAt: timestamp,
    updatedAt: timestamp,
    ...overrides
  };
}

export function vehicleFixture(overrides: Partial<Vehicle> = {}): Vehicle {
  return {
    id: "vehicle-1",
    inventoryNumber: "RK-LOK-000001",
    manufacturer: "ESU",
    articleNumber: "12345",
    name: "BR 106",
    gauge: "H0",
    category: "Lokomotive",
    gattung: "Diesellokomotive",
    digital: true,
    digitalDecoderNumber: "1001",
    dtDecoder: false,
    exhibitionReady: true,
    exhibition: false,
    abcBrakes: false,
    couplingSame: true,
    driveEnabled: true,
    headlightsEnabled: true,
    lightingEnabled: false,
    soundGeneratorEnabled: false,
    smokeGeneratorEnabled: false,
    qrCodeEnabled: true,
    images: [imageFixture()],
    maintenance: [maintenanceFixture()],
    functions: [functionFixture()],
    cvValues: [cvValueFixture()],
    cvFiles: [cvFileFixture()],
    createdAt: timestamp,
    updatedAt: timestamp,
    ...overrides
  };
}

export function analogVehicleFixture(overrides: Partial<Vehicle> = {}): Vehicle {
  return vehicleFixture({
    id: "vehicle-analog",
    inventoryNumber: "RK-LOK-000002",
    name: "BR 86",
    digital: false,
    digitalDecoderNumber: "",
    exhibitionReady: false,
    images: [],
    maintenance: [],
    functions: [],
    cvValues: [],
    cvFiles: [],
    ...overrides
  });
}
