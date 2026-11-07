import type { VehicleCatalogContext, VehicleCatalogFields, VehicleCatalogRevision } from "../../shared/api";

export const catalogFieldNames = [
  "manufacturer", "articleNumber", "name", "gauge", "epoch", "railwayCompany", "category", "gattung",
  "description", "articleSourceUrl", "ean", "productionPeriod", "listPrice"
] as const satisfies readonly (keyof VehicleCatalogFields)[];
export type CatalogField = typeof catalogFieldNames[number];
export const catalogRequiredFields = ["manufacturer", "name", "gauge", "category", "gattung"] as const;
export type CatalogDraft = {
  fields: VehicleCatalogFields;
  selected: CatalogField[];
  revisionID: string;
  creating: boolean;
  independent: boolean;
  note: string;
  identity: string;
};

export function catalogFieldsFrom(value: Partial<VehicleCatalogFields>): VehicleCatalogFields {
  return {
    manufacturer: value.manufacturer ?? "", articleNumber: value.articleNumber ?? "", name: value.name ?? "",
    gauge: value.gauge ?? "", epoch: value.epoch ?? "", railwayCompany: value.railwayCompany ?? "",
    category: value.category ?? "", gattung: value.gattung ?? "", description: value.description ?? "",
    articleSourceUrl: value.articleSourceUrl ?? "", ean: value.ean ?? "",
    productionPeriod: value.productionPeriod ?? "", listPrice: value.listPrice ?? ""
  };
}

export function sameCatalogFields(left: VehicleCatalogFields, right: VehicleCatalogFields) {
  return catalogFieldNames.every(field => left[field] === right[field]);
}

export function catalogSelection(context: VehicleCatalogContext, fields: VehicleCatalogFields): CatalogField[] {
  return catalogFieldNames.filter(field => fields[field] !== context.currentFields[field] && !Object.hasOwn(context.overrides, field));
}

export function catalogDraftFrom(context: VehicleCatalogContext, revision?: VehicleCatalogRevision): CatalogDraft {
  const source = revision ?? context.pinned;
  const fields = { ...(source?.fields ?? context.currentFields) };
  return { fields, selected: catalogSelection(context, fields), revisionID: source?.id ?? "", creating: false,
    independent: false, note: "", identity: "" };
}

export function missingCatalogFields(fields: VehicleCatalogFields) {
  return catalogRequiredFields.filter(field => !fields[field].trim());
}
