import type { ReactNode } from "react";
import {
  AlertTriangle,
  BadgeCheck,
  Eye,
  Gauge,
  Grid2X2,
  Image,
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
import { formatDate } from "./vehicleFormat";
import { maintenanceReminderText } from "./vehicleMaintenance";
import { previewImageUrl, primaryImage, vehicleExhibitionEligible } from "./vehicleTransforms";

type SortKey = "inventoryNumber" | "manufacturer" | "articleNumber" | "name" | "gauge" | "epoch" | "category";
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
  inventoryView: InventoryViewMode;
  inventoryFilter: InventoryFilter;
  maintenanceFilter: MaintenanceFilter;
  manufacturerFilter: string;
  categoryFilter: string;
  gattungFilter: string;
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
  };
  hasActiveInventoryFilters: boolean;
  allVisibleSelected: boolean;
  selectedVehicleIDs: Set<string>;
  onCreate: () => void;
  onReload: () => void;
  onOpenReport: () => void;
  onQueryChange: (value: string) => void;
  onInventoryViewChange: (value: InventoryViewMode) => void;
  onInventoryFilterChange: (value: InventoryFilter) => void;
  onMaintenanceFilterChange: (value: MaintenanceFilter) => void;
  onManufacturerFilterChange: (value: string) => void;
  onCategoryFilterChange: (value: string) => void;
  onGattungFilterChange: (value: string) => void;
  onExhibitionReadyFilterChange: (value: boolean) => void;
  onResetFilters: () => void;
  onOpenDetail: (vehicle: Vehicle, tab?: "model" | "control" | "cv" | "uploads" | "maintenance") => void;
  onOpenEdit: (vehicle: Vehicle) => void;
  onDelete: (vehicle: Vehicle) => void;
  onToggleSelection: (vehicleID: string) => void;
  onToggleAllVisibleSelection: () => void;
  onToggleExhibition: (vehicle: Vehicle, exhibition: boolean) => void;
  renderSortHeader: (key: SortKey) => ReactNode;
  renderQuickMenu: (vehicle: Vehicle) => ReactNode;
};

export function VehicleInventoryPanel({
  vehicles,
  sortedVehicles,
  loading,
  message,
  query,
  inventoryView,
  inventoryFilter,
  maintenanceFilter,
  manufacturerFilter,
  categoryFilter,
  gattungFilter,
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
  onInventoryViewChange,
  onInventoryFilterChange,
  onMaintenanceFilterChange,
  onManufacturerFilterChange,
  onCategoryFilterChange,
  onGattungFilterChange,
  onExhibitionReadyFilterChange,
  onResetFilters,
  onOpenDetail,
  onOpenEdit,
  onDelete,
  onToggleSelection,
  onToggleAllVisibleSelection,
  onToggleExhibition,
  renderSortHeader,
  renderQuickMenu
}: InventoryPanelProps) {
  const { t } = useI18n();
  const gaugeCount = new Set(vehicles.map((vehicle) => vehicle.gauge).filter(Boolean)).size;

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
        <article className={inventoryFilter === "all" && maintenanceFilter === "all" && !manufacturerFilter && !categoryFilter && !gattungFilter && !exhibitionReadyFilter ? "inventory-status-card active" : "inventory-status-card"}>
          <button
            type="button"
            onClick={() => {
              onInventoryFilterChange("all");
              onMaintenanceFilterChange("all");
              onManufacturerFilterChange("");
              onCategoryFilterChange("");
              onGattungFilterChange("");
              onExhibitionReadyFilterChange(false);
            }}
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
        <article className="inventory-status-card wide">
          <span><Wrench size={16} aria-hidden="true" /></span>
          <small>{t("vehicles.nextAppointment")}</small>
          {nextMaintenanceReminder ? (
            <button type="button" onClick={() => onOpenDetail(nextMaintenanceReminder.vehicle, "maintenance")}>
              <strong>{nextMaintenanceReminder.vehicle.inventoryNumber}</strong>
              <em>{nextMaintenanceReminder.entry.kind} · {maintenanceReminderText(nextMaintenanceReminder.daysUntilDue)} · {formatDate(nextMaintenanceReminder.entry.dueDate)}</em>
            </button>
          ) : (
            <>
              <strong>{t("vehicles.allQuiet")}</strong>
              <em>{t("vehicles.noDueMaintenance")}</em>
            </>
          )}
        </article>
        <article className={inventoryFilter === "withoutImages" ? "inventory-status-card active" : "inventory-status-card"}>
          <button type="button" onClick={() => onInventoryFilterChange("withoutImages")} aria-label={t("vehicles.status.imagesAria")}>
            <span><Image size={16} aria-hidden="true" /></span>
            <small>{t("vehicles.imageCare")}</small>
            <strong>{vehicles.length ? Math.round((inventorySummary.withImages / vehicles.length) * 100) : 0}%</strong>
            <em>{t("vehicles.withImage", { count: inventorySummary.withImages })}</em>
          </button>
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

            <select className="inventory-filter-select" value={manufacturerFilter} onChange={(event) => onManufacturerFilterChange(event.target.value)} aria-label={t("vehicles.filter.manufacturer")}>
              <option value="">{t("vehicles.filter.manufacturer")}</option>
              {inventoryFilterOptions.manufacturers.map((manufacturer) => (
                <option key={manufacturer} value={manufacturer}>{manufacturer}</option>
              ))}
            </select>

            <select
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
            </select>

            <select className="inventory-filter-select" value={gattungFilter} onChange={(event) => onGattungFilterChange(event.target.value)} aria-label={t("vehicles.filter.gattung")}>
              <option value="">{t("vehicles.filter.gattung")}</option>
              {inventoryFilterOptions.gattungen.map((gattung) => (
                <option key={gattung} value={gattung}>{gattung}</option>
              ))}
            </select>

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

        {!loading && sortedVehicles.length > 0 && (
          <div className="inventory-mobile-list" aria-label="Kompakte Fahrzeugliste">
            {sortedVehicles.map((vehicle) => {
              const image = primaryImage(vehicle.images);
              return (
                <article key={vehicle.id} className="inventory-mobile-item">
                  <button type="button" className="inventory-mobile-media" onClick={() => onOpenDetail(vehicle)} aria-label={`${vehicle.inventoryNumber} anzeigen`}>
                    {image ? (
                      <img src={previewImageUrl(image)} alt="" />
                    ) : (
                      <div className="image-placeholder">{t("exhibition.noPreview")}</div>
                    )}
                  </button>
                  <button type="button" className="inventory-mobile-main" onClick={() => onOpenDetail(vehicle)}>
                    <span>{vehicle.inventoryNumber}</span>
                    <strong>{vehicle.name}</strong>
                    <small>{vehicle.manufacturer || "-"} · {vehicle.articleNumber || "-"} · {vehicle.category || "-"}</small>
                  </button>
                  <div className="inventory-mobile-meta">
                    <span>{vehicle.gauge || "-"}</span>
                    <small>{vehicle.epoch || "-"}</small>
                  </div>
                  <div className="inventory-mobile-actions">
                    <button type="button" className="icon-button" onClick={() => onOpenEdit(vehicle)} aria-label={t("vehicles.edit")} title={t("vehicles.edit")}>
                      <Pencil size={16} />
                    </button>
                    <button type="button" className="icon-button danger" onClick={() => onDelete(vehicle)} aria-label={t("vehicles.delete")} title={t("vehicles.delete")}>
                      <Trash2 size={16} />
                    </button>
                    {renderQuickMenu(vehicle)}
                  </div>
                </article>
              );
            })}
          </div>
        )}

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
                {sortedVehicles.map((vehicle) => {
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
              <div className="table-wrap">
                <table className="inventory-table">
                  <thead>
                    <tr>
                      <th className="select-cell">
                        <label className="table-select-field" title={t("vehicles.report.selectAll")}>
                          <input
                            type="checkbox"
                            checked={allVisibleSelected}
                            onChange={onToggleAllVisibleSelection}
                            aria-label={t("vehicles.report.selectAll")}
                            disabled={sortedVehicles.length === 0}
                          />
                        </label>
                      </th>
                      <th>{t("vehicles.image")}</th>
                      <th>{renderSortHeader("inventoryNumber")}</th>
                      <th>{renderSortHeader("manufacturer")}</th>
                      <th>{renderSortHeader("articleNumber")}</th>
                      <th>{renderSortHeader("name")}</th>
                      <th>{renderSortHeader("gauge")}</th>
                      <th>{renderSortHeader("epoch")}</th>
                      <th>{t("vehicle.field.exhibition")}</th>
                      <th className="actions-cell">{t("vehicles.actions")}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {sortedVehicles.map((vehicle) => {
                      const image = primaryImage(vehicle.images);
                      return (
                        <tr key={vehicle.id} className={selectedVehicleIDs.has(vehicle.id) ? "selected-row" : ""}>
                          <td className="select-cell">
                            <label className="table-select-field" title={t("vehicles.report.selectVehicle")}>
                              <input
                                type="checkbox"
                                checked={selectedVehicleIDs.has(vehicle.id)}
                                onChange={() => onToggleSelection(vehicle.id)}
                                aria-label={`${vehicle.inventoryNumber} ${t("vehicles.report.selectVehicle")}`}
                              />
                            </label>
                          </td>
                          <td>
                            {image ? (
                              <img className="inventory-thumb" src={previewImageUrl(image)} alt="" />
                            ) : (
                              <div className="image-placeholder">{t("exhibition.noPreview")}</div>
                            )}
                          </td>
                          <td>{vehicle.inventoryNumber}</td>
                          <td>{vehicle.manufacturer}</td>
                          <td>{vehicle.articleNumber || "-"}</td>
                          <td>
                            <button type="button" className="inventory-name-link" onClick={() => onOpenDetail(vehicle)}>
                              {vehicle.name}
                            </button>
                          </td>
                          <td>{vehicle.gauge}</td>
                          <td>{vehicle.epoch || "-"}</td>
                          <td>
                            <label
                              className={vehicle.exhibition ? "inventory-inline-switch active" : "inventory-inline-switch"}
                              title={vehicle.exhibition || vehicleExhibitionEligible(vehicle) ? t("vehicles.exhibition.toggle") : t("vehicles.exhibition.requiresDecoder")}
                            >
                              <input
                                type="checkbox"
                                checked={Boolean(vehicle.exhibition)}
                                disabled={!vehicle.exhibition && !vehicleExhibitionEligible(vehicle)}
                                onChange={(event) => onToggleExhibition(vehicle, event.target.checked)}
                                aria-label={t("vehicles.exhibition.toggle")}
                              />
                              <span aria-hidden="true" />
                              <em>{vehicle.exhibition ? t("common.yes") : t("common.no")}</em>
                            </label>
                          </td>
                          <td className="actions-cell">
                            <div className="table-actions">
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
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}
      </section>
    </>
  );
}
