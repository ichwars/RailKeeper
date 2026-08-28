import type { ReactNode } from "react";
import {
  AlertTriangle,
  BadgeCheck,
  Copy,
  Eye,
  Gauge,
  Grid2X2,
  Image,
  Layers3,
  PackageSearch,
  Pencil,
  Printer,
  RefreshCw,
  Search,
  Table2,
  Trash2,
  Wrench,
  X
} from "lucide-react";
import { Vehicle, VehicleMaintenance } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import type { TableColumnWidths } from "../../shared/tableColumnLayout";
import { formatDate } from "./vehicleFormat";
import { maintenanceReminderText } from "./vehicleMaintenance";
import { previewImageUrl, primaryImage } from "./vehicleTransforms";
import { AppSelect } from "../../shared/ui/AppSelect";
import { VehicleColumnPicker } from "./VehicleColumnPicker";
import { VehicleInventoryMobileList } from "./VehicleInventoryMobileList";
import { VehicleInventoryTable } from "./VehicleInventoryTable";
import { groupVehicleInventory } from "./vehicleSetGroups";
import type {
  InventoryQualityFilter,
  SortDirection,
  SortKey
} from "./vehicleViewModel";
import type {
  VehicleColumnMove,
  VehicleTableColumn
} from "./vehicleTableColumns";

type InventoryViewMode = "table" | "cards";
type InventoryFilter = "all" | "digital" | "analog" | "withImages" | "withoutImages";
type MaintenanceFilter = "all" | "due" | "none";

type InventoryFilterOption<T extends string> = {
  key: T;
  label: string;
  icon?: ReactNode;
};

type InventoryPanelProps = {
  vehicles: Vehicle[];
  sortedVehicles: Vehicle[];
  loading: boolean;
  message: string;
  query: string;
  columns: readonly VehicleTableColumn[];
  columnWidths: TableColumnWidths<VehicleTableColumn>;
  columnsLoading: boolean;
  sort: { key: SortKey; direction: SortDirection };
  inventoryView: InventoryViewMode;
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
  inventorySummary: {
    categories: number;
    digital: number;
    analog: number;
    withImages: number;
  };
  maintenanceReminderSummary: {
    due: number;
    upcoming: number;
  };
  nextMaintenanceReminder: {
    vehicle: Vehicle;
    entry: VehicleMaintenance;
    daysUntilDue: number;
  } | null;
  inventoryFilters: InventoryFilterOption<InventoryFilter>[];
  maintenanceFilters: InventoryFilterOption<MaintenanceFilter>[];
  inventoryFilterOptions: {
    manufacturers: string[];
    categories: string[];
    gattungen: string[];
    railwayCompanies: string[];
    epochs: string[];
    adapters: string[];
  };
  hasActiveInventoryFilters: boolean;
  allVisibleSelected: boolean;
  selectedVehicleIDs: Set<string>;
  onCreate: () => void;
  onReload: () => void;
  onOpenReport: () => void;
  onQueryChange: (value: string) => void;
  onToggleColumn: (column: VehicleTableColumn) => void;
  onMoveColumn: (column: VehicleTableColumn, direction: VehicleColumnMove) => void;
  onResetColumns: () => void;
  onPreviewColumnWidth: (column: VehicleTableColumn, width: number) => void;
  onCommitColumnWidth: (column: VehicleTableColumn, width: number) => void;
  onToggleSort: (key: SortKey) => void;
  onInventoryViewChange: (value: InventoryViewMode) => void;
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
  onOpenDetail: (vehicle: Vehicle, tab?: "model" | "control" | "cv" | "uploads" | "maintenance" | "spareParts") => void;
  onOpenEdit: (vehicle: Vehicle) => void;
  onDelete: (vehicle: Vehicle) => void;
  onToggleSelection: (vehicleID: string) => void;
	onToggleSetSelection: (vehicleIDs: string[]) => void;
  onToggleAllVisibleSelection: () => void;
  onToggleExhibition: (vehicle: Vehicle, exhibition: boolean) => void;
	onOpenSet: (setID: string) => void;
	onEditSet?: (setID: string) => void;
	onDuplicateSet?: (setID: string) => void;
  renderQuickMenu: (vehicle: Vehicle) => ReactNode;
};

export function VehicleInventoryPanel({
  vehicles,
  sortedVehicles,
  loading,
  message,
  query,
  columns,
  columnWidths,
  columnsLoading,
  sort,
  inventoryView,
  inventoryFilter,
  maintenanceFilter,
  qualityFilter,
  manufacturerFilter,
  categoryFilter,
  gattungFilter,
  railwayCompanyFilter,
  epochFilter,
  adapterFilter,
  exhibitionReadyFilter,
  inventorySummary,
  maintenanceReminderSummary,
  nextMaintenanceReminder,
  inventoryFilters,
  maintenanceFilters,
  inventoryFilterOptions,
  hasActiveInventoryFilters,
  allVisibleSelected,
  selectedVehicleIDs,
  onCreate,
  onReload,
  onOpenReport,
  onQueryChange,
  onToggleColumn,
  onMoveColumn,
  onResetColumns,
  onPreviewColumnWidth,
  onCommitColumnWidth,
  onToggleSort,
  onInventoryViewChange,
  onInventoryFilterChange,
  onMaintenanceFilterChange,
  onQualityFilterChange,
  onManufacturerFilterChange,
  onCategoryFilterChange,
  onGattungFilterChange,
  onRailwayCompanyFilterChange,
  onEpochFilterChange,
  onAdapterFilterChange,
  onExhibitionReadyFilterChange,
  onResetFilters,
  onOpenDetail,
  onOpenEdit,
  onDelete,
  onToggleSelection,
	onToggleSetSelection,
  onToggleAllVisibleSelection,
  onToggleExhibition,
	onOpenSet,
	onEditSet,
	onDuplicateSet,
  renderQuickMenu
}: InventoryPanelProps) {
  const { t } = useI18n();
  const gaugeCount = new Set(vehicles.map((vehicle) => vehicle.gauge).filter(Boolean)).size;
  const qualityFilterLabels: Record<InventoryQualityFilter, string> = {
    none: "",
    missingArticleNumber: t("vehicles.filter.missingArticleNumber"),
    missingEan: t("vehicles.filter.missingEan"),
    digitalMissingDecoder: t("vehicles.filter.digitalMissingDecoder")
  };

  return (
    <>
      <section className="inventory-head">
        <div>
          <h1>{t("vehicles.title")}</h1>
          <p>{t("vehicles.subtitle")}</p>
        </div>
        <button type="button" className="primary-button new-vehicle-button" onClick={onCreate}>
          <span aria-hidden="true">+</span>
          {t("vehicles.new")}
        </button>
      </section>

      <section className="inventory-status-row" aria-label={t("vehicles.status")}>
        <article className={inventoryFilter === "all" && maintenanceFilter === "all" && qualityFilter === "none" && !manufacturerFilter && !categoryFilter && !gattungFilter && !railwayCompanyFilter && !epochFilter && !adapterFilter && !exhibitionReadyFilter ? "inventory-status-card active" : "inventory-status-card"}>
          <button
            type="button"
            onClick={onResetFilters}
            aria-label={t("vehicles.status.allAria")}
          >
            <span><PackageSearch size={16} aria-hidden="true" /></span>
            <small>{t("vehicles.totalInventory")}</small>
            <strong>{vehicles.length}</strong>
            <em>{t("overview.categoriesGauges", { categories: inventorySummary.categories, gauges: gaugeCount })}</em>
          </button>
        </article>
        <article className={inventoryFilter === "digital" ? "inventory-status-card active" : "inventory-status-card"}>
          <button type="button" onClick={() => onInventoryFilterChange("digital")} aria-label={t("vehicles.status.digitalAria")}>
            <span><Gauge size={16} aria-hidden="true" /></span>
            <small>{t("vehicles.digitalization")}</small>
            <strong>{vehicles.length ? Math.round((inventorySummary.digital / vehicles.length) * 100) : 0}%</strong>
            <em>{t("vehicles.digitalAnalog", { digital: inventorySummary.digital, analog: inventorySummary.analog })}</em>
          </button>
        </article>
        <article className={[
          "inventory-status-card",
          maintenanceReminderSummary.due > 0 ? "attention" : "",
          maintenanceFilter === "due" ? "active" : ""
        ].filter(Boolean).join(" ")}>
          <button type="button" onClick={() => onMaintenanceFilterChange("due")} aria-label={t("vehicles.status.maintenanceAria")}>
            <span>{maintenanceReminderSummary.due > 0 ? <AlertTriangle size={16} aria-hidden="true" /> : <Wrench size={16} aria-hidden="true" />}</span>
            <small>{t("vehicles.maintenance")}</small>
            <strong>{maintenanceReminderSummary.due}</strong>
            <em>{maintenanceReminderSummary.upcoming} geplant</em>
          </button>
        </article>
        <article className={inventoryFilter === "withoutImages" ? "inventory-status-card active" : "inventory-status-card"}>
          <button type="button" onClick={() => onInventoryFilterChange("withoutImages")} aria-label={t("vehicles.status.imagesAria")}>
            <span><Image size={16} aria-hidden="true" /></span>
            <small>{t("vehicles.imageCare")}</small>
            <strong>{vehicles.length ? Math.round((inventorySummary.withImages / vehicles.length) * 100) : 0}%</strong>
            <em>{t("vehicles.withImage", { count: inventorySummary.withImages })}</em>
          </button>
        </article>
        <article className="inventory-status-card wide">
          {nextMaintenanceReminder ? (
            <button type="button" onClick={() => onOpenDetail(nextMaintenanceReminder.vehicle, "maintenance")}>
              <span><Wrench size={16} aria-hidden="true" /></span>
              <small>{t("vehicles.nextAppointment")}</small>
              <strong>{nextMaintenanceReminder.vehicle.inventoryNumber}</strong>
              <em>{nextMaintenanceReminder.entry.kind} · {maintenanceReminderText(nextMaintenanceReminder.daysUntilDue)} · {formatDate(nextMaintenanceReminder.entry.dueDate)}</em>
            </button>
          ) : (
            <>
              <span><Wrench size={16} aria-hidden="true" /></span>
              <small>{t("vehicles.nextAppointment")}</small>
              <strong>{t("vehicles.allQuiet")}</strong>
              <em>{t("vehicles.noDueMaintenance")}</em>
            </>
          )}
        </article>
      </section>

      <section className="panel inventory-panel">
        <div className="panel-head inventory-list-head">
          <div className="inventory-title-line">
            <div>
              <h2>{t("vehicles.list.title")}</h2>
              <p>{t("vehicles.list.count", { shown: sortedVehicles.length, total: vehicles.length })}</p>
            </div>
          </div>
          <div className="inventory-toolbar" aria-label={t("vehicles.tools")}>
            <label className="search-field inventory-search">
              <span>
                <Search size={16} aria-hidden="true" />
                <input
                  value={query}
                  onChange={(event) => onQueryChange(event.target.value)}
                  placeholder={t("vehicles.search.placeholder")}
                  aria-label={t("vehicles.search.aria")}
                />
              </span>
            </label>
            <div className="table-actions inventory-toolbar-actions">
              <span className="inventory-view-tools" aria-label="Ansicht wechseln">
                <button type="button" className={inventoryView === "table" ? "icon-button active" : "icon-button"} onClick={() => onInventoryViewChange("table")} aria-label="Tabellenansicht" title="Tabellenansicht">
                  <Table2 size={16} />
                </button>
                <button type="button" className={inventoryView === "cards" ? "icon-button active" : "icon-button"} onClick={() => onInventoryViewChange("cards")} aria-label="Kartenansicht" title="Kartenansicht">
                  <Grid2X2 size={16} />
                </button>
              </span>
              <VehicleColumnPicker
                columns={columns}
                loading={columnsLoading}
                onToggle={onToggleColumn}
                onMove={onMoveColumn}
                onReset={onResetColumns}
              />
              <button type="button" className="icon-button" onClick={onOpenReport} aria-label={t("vehicles.report.open")} title={t("vehicles.report.open")} disabled={loading || vehicles.length === 0}>
                <Printer size={16} />
              </button>
              <button type="button" className="icon-button" onClick={onReload} aria-label="Aktualisieren" title="Aktualisieren" disabled={loading}>
                <RefreshCw size={16} />
              </button>
            </div>
          </div>
          <div className="inventory-filter-row" aria-label={t("vehicles.filter")}>
            <div className="inventory-filter-group">
              {inventoryFilters.map((filter) => (
                <button
                  key={filter.key}
                  type="button"
                  className={inventoryFilter === filter.key ? "inventory-filter-pill active" : "inventory-filter-pill"}
                  onClick={() => onInventoryFilterChange(filter.key)}
                  aria-label={filter.label}
                  title={filter.label}
                  aria-pressed={inventoryFilter === filter.key}
                >
                  {filter.icon || <span>{filter.label}</span>}
                </button>
              ))}
            </div>

            <div className="inventory-filter-group">
              {maintenanceFilters.map((filter) => (
                <button
                  key={filter.key}
                  type="button"
                  className={maintenanceFilter === filter.key ? "inventory-filter-pill active" : "inventory-filter-pill"}
                  onClick={() => onMaintenanceFilterChange(filter.key)}
                  aria-label={filter.label}
                  title={filter.label}
                  aria-pressed={maintenanceFilter === filter.key}
                >
                  {filter.icon || <span>{filter.label}</span>}
                </button>
              ))}
            </div>

            <AppSelect className="inventory-filter-select" value={manufacturerFilter} onChange={(event) => onManufacturerFilterChange(event.target.value)} aria-label={t("vehicles.filter.manufacturer")}>
              <option value="">{t("vehicles.filter.manufacturer")}</option>
              {inventoryFilterOptions.manufacturers.map((manufacturer) => (
                <option key={manufacturer} value={manufacturer}>{manufacturer}</option>
              ))}
            </AppSelect>

            <AppSelect
              className="inventory-filter-select"
              value={categoryFilter}
              onChange={(event) => {
                onCategoryFilterChange(event.target.value);
                onGattungFilterChange("");
              }}
              aria-label={t("vehicles.filter.category")}
            >
              <option value="">{t("vehicles.filter.category")}</option>
              {inventoryFilterOptions.categories.map((category) => (
                <option key={category} value={category}>{category}</option>
              ))}
            </AppSelect>

            <AppSelect className="inventory-filter-select" value={gattungFilter} onChange={(event) => onGattungFilterChange(event.target.value)} aria-label={t("vehicles.filter.gattung")}>
              <option value="">{t("vehicles.filter.gattung")}</option>
              {inventoryFilterOptions.gattungen.map((gattung) => (
                <option key={gattung} value={gattung}>{gattung}</option>
              ))}
            </AppSelect>

            <AppSelect className="inventory-filter-select" value={railwayCompanyFilter} onChange={(event) => onRailwayCompanyFilterChange(event.target.value)} aria-label={t("vehicles.filter.railwayCompany")}>
              <option value="">{t("vehicles.filter.railwayCompany")}</option>
              {inventoryFilterOptions.railwayCompanies.map((railwayCompany) => (
                <option key={railwayCompany} value={railwayCompany}>{railwayCompany}</option>
              ))}
            </AppSelect>

            <AppSelect className="inventory-filter-select" value={epochFilter} onChange={(event) => onEpochFilterChange(event.target.value)} aria-label={t("vehicles.filter.epoch")}>
              <option value="">{t("vehicles.filter.epoch")}</option>
              {inventoryFilterOptions.epochs.map((epoch) => (
                <option key={epoch} value={epoch}>{epoch}</option>
              ))}
            </AppSelect>

            <AppSelect className="inventory-filter-select" value={adapterFilter} onChange={(event) => onAdapterFilterChange(event.target.value)} aria-label={t("vehicles.filter.adapter")}>
              <option value="">{t("vehicles.filter.adapter")}</option>
              {inventoryFilterOptions.adapters.map((adapter) => (
                <option key={adapter} value={adapter}>{adapter}</option>
              ))}
            </AppSelect>

            <button
              type="button"
              className={exhibitionReadyFilter ? "inventory-filter-pill inventory-filter-toggle active" : "inventory-filter-pill inventory-filter-toggle"}
              onClick={() => onExhibitionReadyFilterChange(!exhibitionReadyFilter)}
              aria-pressed={exhibitionReadyFilter}
              title={t("vehicles.filter.exhibitionReady")}
            >
              <BadgeCheck size={15} aria-hidden="true" />
              <span>{t("vehicles.filter.exhibitionReady")}</span>
            </button>

            {qualityFilter !== "none" && (
              <button
                type="button"
                className="inventory-filter-pill inventory-filter-toggle active"
                onClick={() => onQualityFilterChange("none")}
                aria-pressed="true"
                title={qualityFilterLabels[qualityFilter]}
              >
                <AlertTriangle size={15} aria-hidden="true" />
                <span>{qualityFilterLabels[qualityFilter]}</span>
              </button>
            )}

            {hasActiveInventoryFilters && (
              <>
                <span className="inventory-filter-divider" aria-hidden="true" />
                <button type="button" className="inventory-filter-clear" onClick={onResetFilters}>
                  <X size={14} aria-hidden="true" />
                  {t("vehicles.filter.clear")}
                </button>
              </>
            )}

            <span className="inventory-filter-result">
              {t("vehicles.filter.result", { count: sortedVehicles.length })}
            </span>
          </div>
        </div>

        {message && <p className="form-message">{message}</p>}

        {!loading && sortedVehicles.length > 0 ? (
          <VehicleInventoryMobileList
            vehicles={sortedVehicles}
            columns={columns}
            sort={sort}
            onOpenDetail={onOpenDetail}
            onOpenEdit={onOpenEdit}
            onOpenSet={onOpenSet}
            onEditSet={onEditSet}
            onDuplicateSet={onDuplicateSet}
            renderQuickMenu={renderQuickMenu}
          />
        ) : null}

        {loading && vehicles.length === 0 ? (
          <p className="empty-state">{t("vehicles.loading")}</p>
        ) : vehicles.length === 0 ? (
          <p className="empty-state">{t("vehicles.empty")}</p>
        ) : sortedVehicles.length === 0 ? (
          <p className="empty-state">{t("vehicles.emptyFilter")}</p>
        ) : (
          <div className="inventory-desktop-content">
            {inventoryView === "cards" ? (
              <div className="inventory-card-grid">
                {groupVehicleInventory(sortedVehicles, sort).map((group) => {
                  if (group.kind === "set") {
                    return (
                      <article key={group.id} className="inventory-card inventory-set-card">
                        <div className="inventory-set-card-head">
                          <span className="vehicle-type-badge set"><Layers3 size={15} />{t("vehicles.set.type")}</span>
                          <span className="inventory-card-gauge">{group.set.gauge || "-"}</span>
                        </div>
                        <div className="inventory-card-body">
                          <div className="inventory-card-title">
                            <div>
                              <strong>{group.set.inventoryNumber}</strong>
                              <span>{group.set.manufacturer || "-"}</span>
                            </div>
                          </div>
                          <h3>{group.set.name}</h3>
                          <dl>
                            <div><dt>{t("importExport.review.article")}</dt>
                              <dd>{group.set.articleNumber || "-"}</dd></div>
                            <div><dt>{t("vehicles.set.members")}</dt>
                              <dd>{t("vehicles.set.visibleOfTotal", {
                                visible: group.visibleMemberCount,
                                total: group.totalMemberCount
                              })}</dd></div>
                          </dl>
                          <div className="inventory-card-actions">
                            <button type="button" className="icon-button" onClick={() => onOpenSet(group.id)}
                              aria-label={t("vehicles.set.viewTitle")}><Eye size={16} /></button>
                            {onEditSet && <button type="button" className="icon-button" onClick={() => onEditSet(group.id)}
                              aria-label={`${t("vehicles.set.type")} ${t("common.edit")}`}><Pencil size={16} /></button>}
                            {onDuplicateSet && <button type="button" className="icon-button" onClick={() => onDuplicateSet(group.id)}
                              aria-label={`${t("vehicles.set.type")} ${t("common.duplicate")}`}><Copy size={16} /></button>}
                          </div>
                        </div>
                      </article>
                    );
                  }
                  const vehicle = group.vehicle;
                  const image = primaryImage(vehicle.images);
                  return (
                    <article key={vehicle.id} className="inventory-card">
                      <button type="button" className="inventory-card-media" onClick={() => onOpenDetail(vehicle)} aria-label={`${vehicle.inventoryNumber} anzeigen`}>
                        {image ? (
                          <img src={previewImageUrl(image)} alt="" />
                        ) : (
                          <div className="image-placeholder">{t("exhibition.noPreview")}</div>
                        )}
                      </button>
                      <div className="inventory-card-body">
                        <div className="inventory-card-title">
                          <div>
                            <strong>{vehicle.inventoryNumber}</strong>
                            <span>{vehicle.manufacturer || "-"}</span>
                          </div>
                          <span className="inventory-card-gauge">{vehicle.gauge || "-"}</span>
                        </div>
                        <h3>{vehicle.name}</h3>
                        <dl>
                          <div>
                            <dt>{t("importExport.review.article")}</dt>
                            <dd>{vehicle.articleNumber || "-"}</dd>
                          </div>
                          <div>
                            <dt>Epoche</dt>
                            <dd>{vehicle.epoch || "-"}</dd>
                          </div>
                          <div>
                            <dt>Kategorie</dt>
                            <dd>{vehicle.category || "-"}</dd>
                          </div>
                        </dl>
                        <div className="inventory-card-actions">
                          <button type="button" className="icon-button" onClick={() => onOpenDetail(vehicle)} aria-label={t("exhibition.view")} title={t("exhibition.view")}>
                            <Eye size={16} />
                          </button>
                          <button type="button" className="icon-button" onClick={() => onOpenEdit(vehicle)} aria-label={t("vehicles.edit")} title={t("vehicles.edit")}>
                            <Pencil size={16} />
                          </button>
                          <button type="button" className="icon-button danger" onClick={() => onDelete(vehicle)} aria-label={t("vehicles.delete")} title={t("vehicles.delete")}>
                            <Trash2 size={16} />
                          </button>
                          {renderQuickMenu(vehicle)}
                        </div>
                      </div>
                    </article>
                  );
                })}
              </div>
            ) : (
              <VehicleInventoryTable
                vehicles={sortedVehicles}
                columns={columns}
                columnWidths={columnWidths}
                allVisibleSelected={allVisibleSelected}
                selectedVehicleIDs={selectedVehicleIDs}
                sort={sort}
                onToggleSort={onToggleSort}
                onToggleSelection={onToggleSelection}
								onToggleSetSelection={onToggleSetSelection}
                onToggleAllVisibleSelection={onToggleAllVisibleSelection}
                onOpenDetail={onOpenDetail}
								onOpenSet={onOpenSet}
								onEditSet={onEditSet}
								onDuplicateSet={onDuplicateSet}
                onOpenEdit={onOpenEdit}
                onDelete={onDelete}
                onToggleExhibition={onToggleExhibition}
                onPreviewColumnWidth={onPreviewColumnWidth}
                onCommitColumnWidth={onCommitColumnWidth}
                renderQuickMenu={renderQuickMenu}
              />
            )}
          </div>
        )}
      </section>
    </>
  );
}
