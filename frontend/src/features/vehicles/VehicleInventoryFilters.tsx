import { useId, useRef, useState, type ReactNode } from "react";
import { BadgeCheck, Check, ChevronDown, ChevronUp, SlidersHorizontal, X } from "lucide-react";

import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import type { InventoryQualityFilter } from "./vehicleViewModel";

type InventoryFilter = "all" | "digital" | "analog" | "withImages" | "withoutImages";
type MaintenanceFilter = "all" | "due" | "none";
type FilterOption<T extends string> = { key: T; label: string; icon?: ReactNode };

export type VehicleInventoryFilterProps = {
  inventoryFilter: InventoryFilter;
  maintenanceFilter: MaintenanceFilter;
  qualityFilter: InventoryQualityFilter;
  manufacturerFilter: string;
  categoryFilter: string;
  gattungFilter: string;
  railwayCompanyFilter: string;
  epochFilter: string;
  adapterFilter: string;
  exhibitionReadyFilter: boolean;
  inventoryFilters: FilterOption<InventoryFilter>[];
  maintenanceFilters: FilterOption<MaintenanceFilter>[];
  inventoryFilterOptions: {
    manufacturers: string[]; categories: string[]; gattungen: string[];
    railwayCompanies: string[]; epochs: string[]; adapters: string[];
  };
  resultCount: number;
  hasActiveInventoryFilters: boolean;
  onInventoryFilterChange: (value: InventoryFilter) => void;
  onMaintenanceFilterChange: (value: MaintenanceFilter) => void;
  onQualityFilterChange: (value: InventoryQualityFilter) => void;
  onManufacturerFilterChange: (value: string) => void;
  onCategoryFilterChange: (value: string) => void;
  onGattungFilterChange: (value: string) => void;
  onRailwayCompanyFilterChange: (value: string) => void;
  onEpochFilterChange: (value: string) => void;
  onAdapterFilterChange: (value: string) => void;
  onExhibitionReadyFilterChange: (value: boolean) => void;
  onResetFilters: () => void;
};

type StringFilterField = {
  key: string;
  group: "primary" | "additional";
  label: string;
  value: string;
  options: readonly string[];
  onChange: (value: string) => void;
};

function FilterSelect({ label, value, options, onChange }:
  Pick<StringFilterField, "label" | "value" | "options" | "onChange">) {
  const { t } = useI18n();
  const fieldID = useId();
  const choices = value && !options.includes(value) ? [value, ...options] : options;
  return <div className="inventory-filter-field">
    <label htmlFor={fieldID}>{label}</label>
    <AppSelect className="inventory-filter-select" value={value}
      id={fieldID} aria-label={label} onChange={(event) => onChange(event.target.value)}>
      <option value="">{t("vehicles.filter.all")}</option>
      {choices.map((option) => <option key={option} value={option}>{option}</option>)}
    </AppSelect>
  </div>;
}

function FilterButtons<T extends string>({ label, value, options, onChange }: {
  label: string; value: T; options: FilterOption<T>[]; onChange: (value: T) => void;
}) {
  return <div className="inventory-filter-button-field"><span>{label}</span>
    <div className="inventory-filter-group" role="group" aria-label={label}>
    {options.map((option) => <button key={option.key} type="button"
      className={`inventory-filter-pill${option.key === value ? " active" : ""}`}
      aria-pressed={option.key === value} onClick={() => onChange(option.key)}>
      {option.key === value && <Check size={14} aria-hidden="true" />}
      {option.icon}<span>{option.label}</span>
    </button>)}
    </div>
  </div>;
}

type ActiveFilter = { key: string; label: string; clear: () => void };
const qualityChoices = ["none", "missingArticleNumber", "missingEan", "digitalMissingDecoder"] as const;

function activeField(field: StringFilterField): ActiveFilter {
  return { key: field.key, label: `${field.label}: ${field.value}`, clear: () => field.onChange("") };
}

function inventoryFields(props: VehicleInventoryFilterProps, t: (key: string) => string): StringFilterField[] {
  return [
    {
      key: "manufacturer", group: "primary", label: t("vehicles.filter.manufacturer"),
      value: props.manufacturerFilter, options: props.inventoryFilterOptions.manufacturers,
      onChange: props.onManufacturerFilterChange
    },
    {
      key: "category", group: "primary", label: t("vehicles.filter.category"),
      value: props.categoryFilter, options: props.inventoryFilterOptions.categories,
      onChange: (value) => { props.onCategoryFilterChange(value); props.onGattungFilterChange(""); }
    },
    {
      key: "gattung", group: "additional", label: t("vehicles.filter.gattung"),
      value: props.gattungFilter, options: props.inventoryFilterOptions.gattungen,
      onChange: props.onGattungFilterChange
    },
    {
      key: "railway", group: "additional", label: t("vehicles.filter.railwayCompany"),
      value: props.railwayCompanyFilter, options: props.inventoryFilterOptions.railwayCompanies,
      onChange: props.onRailwayCompanyFilterChange
    },
    {
      key: "epoch", group: "additional", label: t("vehicles.filter.epoch"),
      value: props.epochFilter, options: props.inventoryFilterOptions.epochs,
      onChange: props.onEpochFilterChange
    },
    {
      key: "adapter", group: "additional", label: t("vehicles.filter.adapter"),
      value: props.adapterFilter, options: props.inventoryFilterOptions.adapters,
      onChange: props.onAdapterFilterChange
    }
  ];
}

export function VehicleInventoryFilters(props: VehicleInventoryFilterProps) {
  const { t } = useI18n();
  const [expanded, setExpanded] = useState(false);
  const moreID = useId();
  const qualityID = useId();
  const moreButton = useRef<HTMLButtonElement>(null);
  const qualityLabels: Record<InventoryQualityFilter, string> = {
    none: t("vehicles.filter.all"), missingArticleNumber: t("vehicles.filter.missingArticleNumber"),
    missingEan: t("vehicles.filter.missingEan"), digitalMissingDecoder: t("vehicles.filter.digitalMissingDecoder")
  };
  const fields = inventoryFields(props, t);
  const primaryFields = fields.filter((field) => field.group === "primary");
  const additionalFields = fields.filter((field) => field.group === "additional");
  const additional: ActiveFilter[] = [
    ...additionalFields.filter((field) => field.value).map(activeField),
    ...(props.exhibitionReadyFilter ? [{ key: "exhibition", label: t("vehicles.filter.exhibitionReady"), clear: () => props.onExhibitionReadyFilterChange(false) }] : []),
    ...(props.qualityFilter !== "none" ? [{ key: "quality", label: qualityLabels[props.qualityFilter], clear: () => props.onQualityFilterChange("none") }] : [])
  ];
  const active: ActiveFilter[] = [
    ...(props.inventoryFilter !== "all" ? [{ key: "inventory", label: props.inventoryFilters.find((item) => item.key === props.inventoryFilter)?.label || t("vehicles.filter"), clear: () => props.onInventoryFilterChange("all") }] : []),
    ...(props.maintenanceFilter !== "all" ? [{ key: "maintenance", label: props.maintenanceFilters.find((item) => item.key === props.maintenanceFilter)?.label || t("vehicles.maintenance"), clear: () => props.onMaintenanceFilterChange("all") }] : []),
    ...primaryFields.filter((field) => field.value).map(activeField),
    ...additional
  ];

  return <div className="inventory-filter-row inventory-filters" role="group" aria-label={t("vehicles.filter")}>
    <div className="inventory-filter-primary">
      <FilterButtons label={t("vehicles.filter.inventoryGroup")} value={props.inventoryFilter}
        options={props.inventoryFilters} onChange={props.onInventoryFilterChange} />
      <FilterButtons label={t("vehicles.maintenance")} value={props.maintenanceFilter}
        options={props.maintenanceFilters} onChange={props.onMaintenanceFilterChange} />
      <div className="inventory-filter-main-fields">
      {primaryFields.map((field) => <FilterSelect key={field.key} label={field.label} value={field.value}
        options={field.options} onChange={field.onChange} />)}
      <button ref={moreButton} className="inventory-filter-pill inventory-more-toggle" type="button"
        aria-expanded={expanded} aria-controls={expanded ? moreID : undefined} onClick={() => setExpanded((value) => !value)}>
        <SlidersHorizontal size={15} aria-hidden="true" />
        <span>{additional.length ? t("vehicles.filter.moreActive", { count: additional.length }) : t("vehicles.filter.more")}</span>
        {expanded ? <ChevronUp size={15} aria-hidden="true" /> : <ChevronDown size={15} aria-hidden="true" />}
      </button>
      </div>
    </div>
    {expanded && <div id={moreID} className="inventory-more-filters" role="group" aria-label={t("vehicles.filter.more")}>
      {additionalFields.map((field) => <FilterSelect key={field.key} label={field.label} value={field.value}
        options={field.options} onChange={field.onChange} />)}
      <div className="inventory-filter-field"><label htmlFor={qualityID}>{t("vehicles.filter.quality")}</label>
        <AppSelect id={qualityID} className="inventory-filter-select" value={props.qualityFilter} aria-label={t("vehicles.filter.quality")}
          onChange={(event) => {
            const selected = qualityChoices.find((value) => value === event.target.value);
            if (selected) props.onQualityFilterChange(selected);
          }}>
          {qualityChoices.map((value) => <option key={value} value={value}>{qualityLabels[value]}</option>)}
        </AppSelect>
      </div>
      <button type="button" className={`inventory-filter-pill${props.exhibitionReadyFilter ? " active" : ""}`}
        aria-pressed={props.exhibitionReadyFilter} onClick={() => props.onExhibitionReadyFilterChange(!props.exhibitionReadyFilter)}>
        {props.exhibitionReadyFilter && <Check size={14} aria-hidden="true" />}
        <BadgeCheck size={15} aria-hidden="true" /><span>{t("vehicles.filter.exhibitionReady")}</span>
      </button>
    </div>}
    <div className="inventory-filter-summary">
      {active.length > 0 && <div className="inventory-active-filters" role="group" aria-label={t("vehicles.filter.active")}>
        {active.map((filter) => <button type="button" key={filter.key} className="inventory-active-filter"
          aria-label={t("vehicles.filter.remove", { label: filter.label })} title={filter.label}
          onClick={() => { filter.clear(); moreButton.current?.focus(); }}>
          <span>{filter.label}</span><X size={14} aria-hidden="true" />
        </button>)}
      </div>}
      {props.hasActiveInventoryFilters && <button type="button" className="inventory-filter-clear" onClick={props.onResetFilters}>
        <X size={14} aria-hidden="true" />{t("vehicles.filter.clear")}
      </button>}
      <span className="inventory-filter-result" role="status">{t("vehicles.filter.result", { count: props.resultCount })}</span>
    </div>
  </div>;
}
