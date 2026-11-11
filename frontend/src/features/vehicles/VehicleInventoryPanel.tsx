import {
  AlertTriangle,
  Archive,
  Copy,
  Eye,
  Gauge,
  Grid2X2,
  Image,
  Layers3,
  PackageSearch,
  Pencil,
  Plus,
  Printer,
  RefreshCw,
  Search,
  Table2,
  Trash2,
  Wrench
} from "lucide-react";
import { useId, type ReactNode } from "react";
import { Vehicle, VehicleMaintenance } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import type { TableColumnWidths } from "../../shared/tableColumnLayout";
import { VehicleColumnPicker } from "./VehicleColumnPicker";
import { VehicleInventoryFilters, type VehicleInventoryFilterProps } from "./VehicleInventoryFilters";
import { VehicleInventoryMobileList } from "./VehicleInventoryMobileList";
import { VehicleInventoryTable } from "./VehicleInventoryTable";
import { formatDate } from "./vehicleFormat";
import { maintenanceReminderText } from "./vehicleMaintenance";
import { groupVehicleInventory } from "./vehicleSetGroups";
import type {
  VehicleColumnMove,
  VehicleTableColumn
} from "./vehicleTableColumns";
import { previewImageUrl, primaryImage } from "./vehicleTransforms";
import type {
  SortDirection,
  SortKey
} from "./vehicleViewModel";

type InventoryViewMode = "table" | "cards";
type InventoryPanelProps = Omit<VehicleInventoryFilterProps, "resultCount"> & {
  vehicles: Vehicle[];
  sortedVehicles: Vehicle[];
  loading: boolean;
  message: string;
  messageTone?: "error" | "status";
  query: string;
  columns: readonly VehicleTableColumn[];
  columnWidths: TableColumnWidths<VehicleTableColumn>;
  columnsLoading: boolean;
  sort: { key: SortKey; direction: SortDirection };
  inventoryView: InventoryViewMode;
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
  allVisibleSelected: boolean;
  selectedVehicleIDs: Set<string>;
  onCreate: () => void;
  onReload: () => void;
  onOpenReport: () => void;
  onOpenCVArchive?: () => void;
  onQueryChange: (value: string) => void;
  onToggleColumn: (column: VehicleTableColumn) => void;
  onMoveColumn: (column: VehicleTableColumn, direction: VehicleColumnMove) => void;
  onResetColumns: () => void;
  onApplyCompactColumns?: () => void;
  onPreviewColumnWidth?: (column: VehicleTableColumn, width: number) => void;
  onCommitColumnWidth?: (column: VehicleTableColumn, width: number) => void;
  onToggleSort: (key: SortKey) => void;
  onInventoryViewChange: (value: InventoryViewMode) => void;
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
  messageTone = "error",
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
  onOpenCVArchive,
  onQueryChange,
  onToggleColumn,
  onMoveColumn,
  onResetColumns,
  onApplyCompactColumns,
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
  const searchHelpId = useId();
  const gaugeCount = new Set(vehicles.map((vehicle) => vehicle.gauge).filter(Boolean)).size;


  return (
    <>
      <section className="inventory-head">
        <div>
          <h1>{t("vehicles.title")}</h1>
          <p>{t("vehicles.subtitle")}</p>
        </div>
        <button type="button" className="primary-button new-vehicle-button" onClick={onCreate}>
          <Plus size={16} aria-hidden="true" />
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
            <em>{t("vehicles.maintenance.plannedCount", { count: maintenanceReminderSummary.upcoming })}</em>
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
                  aria-describedby={searchHelpId}
                  title={t("vehicles.search.help")}
                />
              </span>
              <span className="visually-hidden" id={searchHelpId}>{t("vehicles.search.help")}</span>
            </label>
            <div className="table-actions inventory-toolbar-actions">
              <span className="inventory-view-tools" aria-label={t("vehicles.inventoryView.switch")}>
                <button type="button" className={`icon-button inventory-labeled-action${inventoryView === "table" ? " active" : ""}`} onClick={() => onInventoryViewChange("table")} aria-pressed={inventoryView === "table"} aria-label={t("vehicles.inventoryView.table")} title={t("vehicles.inventoryView.table")}>
                  <Table2 size={16} aria-hidden="true" /><span>{t("vehicles.inventoryView.tableShort")}</span>
                </button>
                <button type="button" className={`icon-button inventory-labeled-action${inventoryView === "cards" ? " active" : ""}`} onClick={() => onInventoryViewChange("cards")} aria-pressed={inventoryView === "cards"} aria-label={t("vehicles.inventoryView.cards")} title={t("vehicles.inventoryView.cards")}>
                  <Grid2X2 size={16} aria-hidden="true" /><span>{t("vehicles.inventoryView.cardsShort")}</span>
                </button>
              </span>
              <VehicleColumnPicker
                columns={columns}
                loading={columnsLoading}
                onToggle={onToggleColumn}
                onMove={onMoveColumn}
                onReset={onResetColumns}
                onApplyCompact={onApplyCompactColumns}
              />
              {onOpenCVArchive && <button type="button" className="icon-button inventory-labeled-action decoder-archive-open" onClick={onOpenCVArchive}
                aria-label={t("decoderConfig.archive")} title={t("decoderConfig.archive")}><Archive size={16} aria-hidden="true" /><span>{t("decoderConfig.archive")}</span></button>}
              <button type="button" className="icon-button inventory-labeled-action" onClick={onOpenReport} aria-label={t("vehicles.report.open")} title={t("vehicles.report.open")} disabled={loading || vehicles.length === 0}>
                <Printer size={16} aria-hidden="true" /><span>{t("vehicles.report.short")}</span>
              </button>
              <button type="button" className="icon-button inventory-labeled-action" onClick={onReload} aria-label={t("common.refresh")} title={t("common.refresh")} disabled={loading}>
                <RefreshCw size={16} aria-hidden="true" /><span>{t("common.refresh")}</span>
              </button>
            </div>
          </div>
          <VehicleInventoryFilters
            inventoryFilter={inventoryFilter}
            maintenanceFilter={maintenanceFilter}
            qualityFilter={qualityFilter}
            manufacturerFilter={manufacturerFilter}
            categoryFilter={categoryFilter}
            gattungFilter={gattungFilter}
            railwayCompanyFilter={railwayCompanyFilter}
            epochFilter={epochFilter}
            adapterFilter={adapterFilter}
            exhibitionReadyFilter={exhibitionReadyFilter}
            inventoryFilters={inventoryFilters}
            maintenanceFilters={maintenanceFilters}
            inventoryFilterOptions={inventoryFilterOptions}
            hasActiveInventoryFilters={hasActiveInventoryFilters}
            onInventoryFilterChange={onInventoryFilterChange}
            onMaintenanceFilterChange={onMaintenanceFilterChange}
            onQualityFilterChange={onQualityFilterChange}
            onManufacturerFilterChange={onManufacturerFilterChange}
            onCategoryFilterChange={onCategoryFilterChange}
            onGattungFilterChange={onGattungFilterChange}
            onRailwayCompanyFilterChange={onRailwayCompanyFilterChange}
            onEpochFilterChange={onEpochFilterChange}
            onAdapterFilterChange={onAdapterFilterChange}
            onExhibitionReadyFilterChange={onExhibitionReadyFilterChange}
            onResetFilters={onResetFilters}
            resultCount={sortedVehicles.length}
          />
        </div>

        {message && <p className={`form-message${messageTone === "status" ? " is-status" : ""}`}
          role={messageTone === "status" ? "status" : "alert"}>{message}</p>}

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
                      <button type="button" className="inventory-card-media" onClick={() => onOpenDetail(vehicle)} aria-label={`${vehicle.inventoryNumber} ${t("vehicles.view")}`}>
                        {image ? (
                          <img loading="lazy" src={previewImageUrl(image)} alt="" />
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
                columnWidthsLoading={columnsLoading}
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
