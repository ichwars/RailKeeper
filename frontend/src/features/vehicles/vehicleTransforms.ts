import {
  ArticleSearchImage,
  CreateVehicleRequest,
  ExhibitionEntryInput,
  Vehicle,
  VehicleAttachment,
  VehicleFunction,
  VehicleFunctionInput,
  VehicleImage as VehicleImageRecord
} from "../../shared/api";

export type PendingArticleImage = ArticleSearchImage & {
  id: string;
  isPrimary?: boolean;
  persisted?: boolean;
  mimeType?: string;
  thumbnailUrl?: string;
  maintenanceId?: string;
};

export type AttachmentEditState = Record<string, { description: string; category: string; maintenanceId: string }>;
export type FunctionEditState = Record<string, VehicleFunctionInput & { persisted?: boolean }>;

export function vehicleToForm(vehicle: Vehicle): CreateVehicleRequest {
  return {
    inventoryNumber: vehicle.inventoryNumber,
    manufacturer: vehicle.manufacturer,
    articleNumber: vehicle.articleNumber || "",
    articleSourceUrl: vehicle.articleSourceUrl || "",
    name: vehicle.name,
    gauge: vehicle.gauge,
    epoch: vehicle.epoch || "",
    railwayCompany: vehicle.railwayCompany || "",
    category: vehicle.category || "",
    gattung: vehicle.gattung || "",
    description: vehicle.description || "",
    series: vehicle.series || "",
    vehicleNumber: vehicle.vehicleNumber || "",
    digital: vehicle.digital,
    digitalDecoderNumber: vehicle.digitalDecoderNumber || "",
    dtDecoder: vehicle.dtDecoder,
    dtDecoderNumber: vehicle.dtDecoderNumber || "",
    decoderType: vehicle.decoderType || "",
    exhibitionReady: vehicle.exhibitionReady,
    exhibition: vehicle.exhibition,
    abcBrakes: vehicle.abcBrakes,
    ean: vehicle.ean || "",
    productionPeriod: vehicle.productionPeriod || "",
    listPrice: vehicle.listPrice || "",
    acquisitionType: vehicle.acquisitionType || "",
    acquiredFrom: vehicle.acquiredFrom || "",
    purchasePrice: vehicle.purchasePrice || "",
    purchaseDate: vehicle.purchaseDate || "",
    storageLocation: vehicle.storageLocation || "",
    storageDetails: vehicle.storageDetails || "",
    condition: vehicle.condition || "",
    conditionDetails: vehicle.conditionDetails || "",
    packaging: vehicle.packaging || "",
    lengthMm: vehicle.lengthMm || "",
    weightG: vehicle.weightG || "",
    color: vehicle.color || "",
    lettering: vehicle.lettering || "",
    load: vehicle.load || "",
    interior: vehicle.interior || "",
    axles: vehicle.axles || "",
    axleCount: vehicle.axleCount || "",
    tractionTireCount: vehicle.tractionTireCount || "",
    wheelset: vehicle.wheelset || "",
    couplingSame: vehicle.couplingSame,
    couplingFront: vehicle.couplingFront || "",
    couplingRear: vehicle.couplingRear || "",
    powerPickup: vehicle.powerPickup || "",
    adapter: vehicle.adapter || "",
    driveEnabled: vehicle.driveEnabled,
    driveDescription: vehicle.driveDescription || "",
    headlightsEnabled: vehicle.headlightsEnabled,
    headlightsDescription: vehicle.headlightsDescription || "",
    lightingEnabled: vehicle.lightingEnabled,
    lightingDescription: vehicle.lightingDescription || "",
    soundGeneratorEnabled: vehicle.soundGeneratorEnabled,
    soundGeneratorDescription: vehicle.soundGeneratorDescription || "",
    smokeGeneratorEnabled: vehicle.smokeGeneratorEnabled,
    smokeGeneratorDescription: vehicle.smokeGeneratorDescription || "",
    additionalInfo: vehicle.additionalInfo || "",
    qrCodeEnabled: vehicle.qrCodeEnabled
  };
}

export function primaryImage(images?: { url: string; thumbnailUrl?: string; isPrimary?: boolean }[]) {
  return images?.find((image) => image.isPrimary) || images?.[0];
}

export function previewImageUrl(image?: { url: string; thumbnailUrl?: string }) {
  return image?.thumbnailUrl || image?.url || "";
}

export function vehicleImageToPending(image: VehicleImageRecord): PendingArticleImage {
  return {
    id: image.id || image.url,
    url: image.url,
    thumbnailUrl: image.thumbnailUrl,
    title: image.title || "",
    source: image.sourceUrl || image.url,
    isPrimary: image.isPrimary,
    persisted: true,
    mimeType: image.mimeType || "",
    maintenanceId: image.maintenanceId || ""
  };
}

export function vehicleImagesToPending(vehicle: Vehicle): PendingArticleImage[] {
  return (vehicle.images || []).map(vehicleImageToPending);
}

export function uploadedImageToPending(image: VehicleImageRecord): PendingArticleImage {
  return vehicleImageToPending(image);
}

export function attachmentsToEditState(attachments?: VehicleAttachment[]): AttachmentEditState {
  return Object.fromEntries(
    (attachments || []).map((attachment) => [
      attachment.id,
      {
        description: attachment.description || "",
        category: attachment.category || "",
        maintenanceId: attachment.maintenanceId || ""
      }
    ])
  );
}

export function functionsToEditState(functions?: VehicleFunction[]): FunctionEditState {
  return Object.fromEntries(
    (functions || []).map((item) => [
      item.functionKey,
      {
        name: item.name || "",
        symbolKey: item.symbolKey || "",
        functionType: item.functionType || "standard",
        mode: item.mode || "dauer",
        directionDependent: item.directionDependent,
        notes: item.notes || "",
        persisted: true
      }
    ])
  );
}

export function normalizedText(value?: string) {
  return String(value || "").trim().toLocaleLowerCase("de-DE");
}

export function vehicleExhibitionEligible(vehicle: Pick<Vehicle, "digital" | "digitalDecoderNumber">) {
  return Boolean(vehicle.digital && vehicle.digitalDecoderNumber?.trim());
}

function vehicleFunctionsToExhibitionFunctions(functions?: VehicleFunction[]) {
  const configured = (functions || [])
    .filter((item) => item.name?.trim() || item.symbolKey || item.notes?.trim())
    .map((item) => ({
      key: item.functionKey,
      name: item.name || "",
      type: item.functionType || "standard",
      symbolKey: item.symbolKey || ""
    }));
  return configured.length > 0 ? JSON.stringify(configured) : "";
}

export function vehicleToExhibitionEntry(vehicle: Vehicle, owner: string): ExhibitionEntryInput {
  const image = primaryImage(vehicle.images);
  return {
    vehicleId: vehicle.id,
    owner,
    imageUrl: image?.url || "",
    locomotiveName: vehicle.name,
    gattung: vehicle.gattung || "",
    series: vehicle.series || "",
    manufacturer: vehicle.manufacturer || "",
    epoch: vehicle.epoch || "",
    railwayCompany: vehicle.railwayCompany || "",
    dayScope: "all",
    dtDecoder: Boolean(vehicle.dtDecoder),
    decoderNumber: vehicle.digitalDecoderNumber || "",
    decoderType: vehicle.decoderType || "",
    adapter: vehicle.adapter || "",
    sxAddress: vehicle.dtDecoderNumber || "",
    analog: false,
    functionKeys: vehicleFunctionsToExhibitionFunctions(vehicle.functions),
    notes: [vehicle.vehicleNumber, vehicle.lettering, vehicle.additionalInfo].filter(Boolean).join(" · ")
  };
}
