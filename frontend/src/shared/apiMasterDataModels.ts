import type { components } from "../generated/api";

export type MasterDataOrigin = "bundled" | "custom";

export type MasterDataCapabilities = {
  canDeactivate: boolean;
  canReactivate: boolean;
  canDelete: boolean;
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
  origin?: MasterDataOrigin;
  capabilities?: MasterDataCapabilities;
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

export type VehicleAttachmentImportInput = components["schemas"]["VehicleAttachmentImportInput"];

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
