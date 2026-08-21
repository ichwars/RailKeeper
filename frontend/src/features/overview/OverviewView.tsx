import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  AlertTriangle,
  ArrowRight,
  Box,
  CircleDollarSign,
  Clock3,
  Cpu,
  Database,
  ExternalLink,
  FileInput,
  Flag,
  Gauge,
  ImageOff,
  LayoutGrid,
  PackagePlus,
  RefreshCw,
  Server,
  TrainFront,
  Upload,
  Wrench
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

import {
  api,
  type AccessoryArticleListResult,
  type DigitalCenterSettings,
  type ECoSLiveStatus,
  type OverviewValuation,
  type Vehicle
} from "../../shared/api";
import { translate, useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { InventoryTrendChart, ValueDistributionChart } from "./OverviewCharts";
import { OverviewMetricDialog } from "./OverviewMetricDialog";
import { formatOverviewMoney } from "./OverviewValuationCard";
import {
  accessoryCount,
  buildOverviewStats,
  buildOverviewTrend,
  defaultOverviewMetrics,
  overviewMetricIDs,
  overviewMetricLimitForWidth,
  overviewMetricProfileKey,
  parseMetricPreference,
  percentage,
  persistMetricPreference,
  primaryVehicleImage,
  readMetricPreference,
  safeNumber,
  type OverviewMetricID,
  type OverviewMetricPreference
} from "./overviewModel";

type OverviewViewProps = {
  username?: string;
  roles?: string[];
};

type MetricCard = {
  icon: LucideIcon;
  value: string;
  detail: React.ReactNode;
  attention?: boolean;
  href?: string;
};

const providerLabels: Record<string, string> = {
  ecos: "ESU ECoS",
  z21: "Roco Z21",
  intellibox3: "Uhlenbrock Intellibox 3",
  cs3: "Märklin CS3"
};

const providerImages: Record<DigitalCenterSettings["provider"], string> = {
  ecos: "/assets/overview-ecos-50220.png",
  z21: "/assets/overview-z21-10820.png",
  intellibox3: "/assets/overview-intellibox-3-65300.jpg",
  cs3: "/assets/overview-cs3-60226.jpg"
};

function sameDay(left: Date, right: Date) {
  return left.getFullYear() === right.getFullYear() && left.getMonth() === right.getMonth() &&
    left.getDate() === right.getDate();
}

function formatRecentTime(value: string, language: string) {
  const date = new Date(value);
  const today = new Date();
  const yesterday = new Date(today);
  yesterday.setDate(today.getDate() - 1);
  const time = new Intl.DateTimeFormat(language === "en" ? "en-GB" : "de-DE", {
    hour: "2-digit",
    minute: "2-digit"
  }).format(date);
  if (sameDay(date, today)) return `${language === "en" ? "Today" : "Heute"}, ${time}`;
  if (sameDay(date, yesterday)) return `${language === "en" ? "Yesterday" : "Gestern"}, ${time}`;
  return new Intl.DateTimeFormat(language === "en" ? "en-GB" : "de-DE", {
    day: "2-digit",
    month: "2-digit",
    year: "2-digit"
  }).format(date);
}

function maintenanceDistanceText(
  days: number,
  t: (key: string, values?: Record<string, string | number>) => string
) {
  if (days < 0) return t("overview.daysOverdue", { days: Math.abs(days) });
  if (days === 0) return t("overview.dueToday");
  if (days === 1) return t("overview.dueTomorrow");
  return t("overview.dueInDays", { days });
}

function moveMetric(order: OverviewMetricID[], metric: OverviewMetricID, target: OverviewMetricID) {
  const sourceIndex = order.indexOf(metric);
  const targetIndex = order.indexOf(target);
  if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex) return order;
  const next = [...order];
  next.splice(sourceIndex, 1);
  next.splice(targetIndex, 0, metric);
  return next;
}

function providerEnabled(settings: DigitalCenterSettings | null) {
  if (!settings) return false;
  return Boolean(settings[settings.provider]?.enabled && settings[settings.provider]?.host.trim());
}

export function OverviewView({ username = "local", roles = ["Editor"] }: OverviewViewProps) {
  const { language, t } = useI18n();
  const canEdit = roles.includes("Admin") || roles.includes("Editor");
  const canManageDigital = roles.includes("Admin");
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [accessories, setAccessories] = useState<AccessoryArticleListResult | null>(null);
  const [valuation, setValuation] = useState<OverviewValuation | null>(null);
  const [digitalSettings, setDigitalSettings] = useState<DigitalCenterSettings | null>(null);
  const [liveStatus, setLiveStatus] = useState<ECoSLiveStatus | null>(null);
  const [vehicleLoading, setVehicleLoading] = useState(true);
  const [accessoryLoading, setAccessoryLoading] = useState(true);
  const [valuationLoading, setValuationLoading] = useState(true);
  const [digitalLoading, setDigitalLoading] = useState(canManageDigital);
  const [vehicleError, setVehicleError] = useState("");
  const [accessoryError, setAccessoryError] = useState("");
  const [valuationError, setValuationError] = useState("");
  const [digitalError, setDigitalError] = useState("");
  const [period, setPeriod] = useState(12);
  const [valueMode, setValueMode] = useState<"list" | "purchase">("list");
  const [metricDialogOpen, setMetricDialogOpen] = useState(false);
  const [metricPreference, setMetricPreference] = useState<OverviewMetricPreference>(() =>
    readMetricPreference(username));
  const [draftMetrics, setDraftMetrics] = useState<OverviewMetricPreference>(metricPreference);
  const [metricLimit, setMetricLimit] = useState(4);
  const [preferenceError, setPreferenceError] = useState("");
  const valuationRequestID = useRef(0);
  const metricsRef = useRef<HTMLElement>(null);

  const loadVehicles = useCallback(() => {
    setVehicleLoading(true);
    setVehicleError("");
    return api.vehicles().then(setVehicles).catch((error: Error) => setVehicleError(error.message))
      .finally(() => setVehicleLoading(false));
  }, []);

  const loadAccessories = useCallback(() => {
    setAccessoryLoading(true);
    setAccessoryError("");
    return api.accessoryArticles().then(setAccessories)
      .catch((error: Error) => setAccessoryError(error.message))
      .finally(() => setAccessoryLoading(false));
  }, []);

  const loadValuation = useCallback(() => {
    setValuationLoading(true);
    setValuationError("");
    const requestID = valuationRequestID.current + 1;
    valuationRequestID.current = requestID;
    return api.overviewValuation().then((nextValuation) => {
      if (requestID === valuationRequestID.current) setValuation(nextValuation);
    }).catch(() => {
      if (requestID === valuationRequestID.current) {
        setValuationError(translate(language, "overview.valuation.error"));
      }
    }).finally(() => {
      if (requestID === valuationRequestID.current) setValuationLoading(false);
    });
  }, [language]);

  const loadDigital = useCallback(() => {
    if (!canManageDigital) return Promise.resolve();
    setDigitalLoading(true);
    setDigitalError("");
    return api.digitalSettings().then(async (settings) => {
      setDigitalSettings(settings);
      if (settings.provider === "ecos" && settings.ecos.enabled && settings.ecos.host.trim()) {
        try {
          setLiveStatus(await api.getECoSLiveStatus());
        } catch {
          setLiveStatus(null);
        }
      } else {
        setLiveStatus(null);
      }
    }).catch((error: Error) => setDigitalError(error.message))
      .finally(() => setDigitalLoading(false));
  }, [canManageDigital]);

  useEffect(() => {
    void loadVehicles();
    void loadAccessories();
  }, [loadAccessories, loadVehicles]);

  useEffect(() => {
    void loadValuation();
  }, [loadValuation]);

  useEffect(() => {
    void loadDigital();
  }, [loadDigital]);

  useEffect(() => {
    let active = true;
    api.profileSettings().then(({ settings }) => {
      if (!active) return;
      const stored = parseMetricPreference(settings[overviewMetricProfileKey] || null);
      if (!stored) return;
      setMetricPreference(stored);
      setDraftMetrics(stored);
      persistMetricPreference(username, stored);
    }).catch(() => undefined);
    return () => { active = false; };
  }, [username]);

  useEffect(() => {
    const metricsElement = metricsRef.current;
    if (!metricsElement) return;
    const updateLimit = () => setMetricLimit(overviewMetricLimitForWidth(
      metricsElement.getBoundingClientRect().width));
    const observer = typeof ResizeObserver === "undefined" ? null : new ResizeObserver(updateLimit);
    observer?.observe(metricsElement);
    if (!observer) window.addEventListener("resize", updateLimit);
    updateLimit();
    return () => {
      observer?.disconnect();
      if (!observer) window.removeEventListener("resize", updateLimit);
    };
  }, []);

  const stats = useMemo(() => buildOverviewStats(vehicles), [vehicles]);
  const accessoryTotal = accessoryCount(accessories);
  const imageShare = percentage(stats.withImages, vehicles.length);
  const articleShare = percentage(stats.withArticleNumbers, vehicles.length);
  const decoderShare = percentage(stats.withDecoderAddresses, vehicles.length);
  const documentedShare = percentage(stats.documented, vehicles.length);
  const dataQuality = Math.round((imageShare + articleShare + decoderShare + documentedShare) / 4);
  const trend = useMemo(() => buildOverviewTrend(vehicles, accessories, period, language),
    [accessories, language, period, vehicles]);
  const valueVehicle = safeNumber(valueMode === "list" ? valuation?.vehicleListValue :
    valuation?.vehiclePurchaseValue);
  const valueAccessory = safeNumber(valueMode === "list" ? valuation?.accessoryListValue :
    valuation?.accessoryPurchaseCost);
  const provider = digitalSettings?.provider || "ecos";
  const lastSynchronization = liveStatus?.lastSeenAt || liveStatus?.startedAt;
  const managedLocomotives = vehicles.filter((vehicle) => vehicle.externalMappings
    ?.some((mapping) => mapping.provider.toLowerCase() === provider)).length;

  const money = (value: number) => formatOverviewMoney(value.toFixed(2), language);
  const metricCards: Record<OverviewMetricID, MetricCard> = {
    vehicles: {
      icon: TrainFront,
      value: vehicleLoading ? "…" : String(vehicles.length),
      detail: t("overview.metric.vehiclesDetail", { digital: stats.digital, analog: stats.analog }),
      href: "/vehicles"
    },
    accessories: {
      icon: Box,
      value: accessoryLoading ? "…" : String(accessoryTotal),
      detail: t("overview.metric.accessoriesDetail", { reserved: accessories?.metrics.reserved || 0 }),
      href: "/accessories"
    },
    inventoryValue: {
      icon: Database,
      value: valuationLoading ? "…" : valuation ? money(valueVehicle + valueAccessory) : "–",
      detail: valuationError || t("overview.metric.inventoryValueDetail")
    },
    maintenance: {
      icon: Wrench,
      value: vehicleLoading ? "…" : String(stats.due + stats.upcoming),
      detail: <><strong>{t("overview.metric.overdue", { count: stats.due })}</strong>
        {` · ${t("overview.metric.planned", { count: stats.upcoming })}`}</>,
      attention: stats.due > 0,
      href: "/vehicles?maintenance=due"
    },
    digitalized: {
      icon: Cpu,
      value: vehicleLoading ? "…" : String(stats.digital),
      detail: t("overview.metric.digitalizedDetail", { share: percentage(stats.digital, vehicles.length) }),
      href: "/vehicles?digital=true"
    },
    dataQuality: {
      icon: Gauge,
      value: vehicleLoading ? "…" : `${dataQuality}%`,
      detail: t("overview.metric.dataQualityDetail")
    },
    vehicleListValue: {
      icon: CircleDollarSign,
      value: valuationLoading ? "…" : valuation ? formatOverviewMoney(valuation.vehicleListValue, language) : "–",
      detail: t("overview.metric.vehicleListValueDetail")
    },
    vehiclePurchaseValue: {
      icon: CircleDollarSign,
      value: valuationLoading ? "…" : valuation ? formatOverviewMoney(valuation.vehiclePurchaseValue, language) : "–",
      detail: t("overview.metric.vehiclePurchaseValueDetail")
    },
    accessoryListValue: {
      icon: CircleDollarSign,
      value: valuationLoading ? "…" : valuation ? formatOverviewMoney(valuation.accessoryListValue, language) : "–",
      detail: t("overview.metric.accessoryListValueDetail")
    },
    accessoryPurchaseValue: {
      icon: CircleDollarSign,
      value: valuationLoading ? "…" : valuation ? formatOverviewMoney(valuation.accessoryPurchaseCost, language) : "–",
      detail: t("overview.metric.accessoryPurchaseValueDetail")
    }
  };

  const displayedMetricPreference = metricDialogOpen ? draftMetrics : metricPreference;
  const visibleMetrics = displayedMetricPreference.order.filter((metric) =>
    displayedMetricPreference.active.includes(metric));
  const refreshOverview = () => {
    void loadVehicles();
    void loadAccessories();
    void loadValuation();
    void loadDigital();
  };
  const openMetricDialog = () => {
    setDraftMetrics(metricPreference);
    setPreferenceError("");
    setMetricDialogOpen(true);
  };
  const closeMetricDialog = () => {
    setDraftMetrics(metricPreference);
    setMetricDialogOpen(false);
  };
  const saveMetricPreference = () => {
    setMetricPreference(draftMetrics);
    persistMetricPreference(username, draftMetrics);
    setMetricDialogOpen(false);
    const settingsUpdate = {
      [overviewMetricProfileKey]: JSON.stringify(draftMetrics)
    };
    void api.updateProfileSettings(settingsUpdate)
      .catch(() => api.updateProfileSettings(settingsUpdate))
      .catch(() => setPreferenceError(t("overview.metrics.saveError")));
  };

  const vehicleDependentState = vehicleError || (vehicleLoading ? t("overview.state.loading") : "");

  return (
    <div className="overview-page">
      <section className="overview-page-head">
        <div>
          <p className="eyebrow">{t("overview.eyebrow")}</p>
          <h1>{t("overview.title")}</h1>
          <p>{t("overview.subtitle")}</p>
        </div>
        <div className="overview-head-tools" aria-label={t("overview.tools")}>
          <div className="overview-period">
            <AppSelect value={String(period)} onChange={(event) => setPeriod(Number(event.target.value))}
              aria-label={t("overview.period")}>
              <option value={6}>{t("overview.periodMonths", { count: 6 })}</option>
              <option value={12}>{t("overview.periodMonths", { count: 12 })}</option>
              <option value={24}>{t("overview.periodMonths", { count: 24 })}</option>
            </AppSelect>
          </div>
          <button type="button" className="icon-button" onClick={refreshOverview}
            disabled={vehicleLoading || accessoryLoading || valuationLoading || digitalLoading}
            aria-label={t("overview.refresh")} title={t("overview.refresh")}>
            <RefreshCw size={17} aria-hidden="true" />
          </button>
          <button type="button" className={`icon-button overview-metric-trigger${metricDialogOpen ? " active" : ""}`}
            onClick={metricDialogOpen ? closeMetricDialog : openMetricDialog} aria-expanded={metricDialogOpen}
            aria-controls="overview-metric-dialog-title" aria-label={t("overview.metrics.dialogTitle")}
            title={t("overview.metrics.dialogTitle")}>
            <LayoutGrid size={18} aria-hidden="true" />
          </button>
        </div>
      </section>

      {preferenceError ? <p className="overview-preference-error" role="alert">{preferenceError}</p> : null}

      <section className="overview-metrics" aria-label={t("overview.metrics.label")} ref={metricsRef}
        style={{ "--overview-metric-columns": Math.min(visibleMetrics.length, metricLimit) } as React.CSSProperties}>
        {visibleMetrics.map((metric) => {
          const card = metricCards[metric];
          const Icon = card.icon;
          const content = <>
            <div className="overview-metric-title"><Icon size={20} aria-hidden="true" />
              <span>{t(`overview.metric.${metric}`)}</span></div>
            <strong>{card.value}</strong>
            <small>{card.detail}</small>
          </>;
          return card.href ? (
            <a className={`overview-metric-card${card.attention ? " attention" : ""}`}
              href={card.href} key={metric}>{content}</a>
          ) : (
            <article className={`overview-metric-card${card.attention ? " attention" : ""}`}
              key={metric}>{content}</article>
          );
        })}
      </section>

      <section className="overview-operational" aria-label={t("overview.operational.label")}>
        <article className="overview-card overview-priority-card">
          <header className="overview-card-head"><Flag size={19} aria-hidden="true" />
            <h2>{t("overview.priority.title")}</h2></header>
          {vehicleDependentState ? <p className={vehicleError ? "overview-module-state error" : "overview-module-state"}
            role={vehicleError ? "alert" : "status"}>{vehicleDependentState}</p> : (
            <div className="overview-priority-list">
              <a href="/vehicles?maintenance=due">
                <AlertTriangle className="warning" size={19} aria-hidden="true" />
                <span>{t("overview.priority.overdue", { count: stats.due })}</span>
                <small>{t("overview.priority.vehicles", { count: stats.due })}</small>
                <ArrowRight size={17} aria-hidden="true" />
              </a>
              <a href="/vehicles?gap=no-main-image">
                <ImageOff className="info" size={19} aria-hidden="true" />
                <span>{t("overview.priority.noImage", { count: vehicles.length - stats.withImages })}</span>
                <small>{t("overview.priority.vehicles", { count: vehicles.length - stats.withImages })}</small>
                <ArrowRight size={17} aria-hidden="true" />
              </a>
              <a href="/vehicles?gap=digital-no-decoder">
                <Cpu className="decoder" size={19} aria-hidden="true" />
                <span>{t("overview.priority.noAddress", { count: stats.digitalWithoutAddress })}</span>
                <small>{t("overview.priority.decoders", { count: stats.digitalWithoutAddress })}</small>
                <ArrowRight size={17} aria-hidden="true" />
              </a>
            </div>
          )}
        </article>

        <article className="overview-card overview-digital-card">
          <header className="overview-card-head"><Server size={19} aria-hidden="true" />
            <h2>{t("overview.digitalCenters.title")}</h2></header>
          {!canManageDigital ? (
            <p className="overview-module-state">{t("overview.digitalCenters.noAccess")}</p>
          ) : digitalLoading ? (
            <p className="overview-module-state" role="status">{t("overview.state.loading")}</p>
          ) : digitalError ? (
            <p className="overview-module-state error" role="alert">{digitalError}</p>
          ) : (
            <div className="overview-digital-body">
              <div className="overview-device-art" aria-hidden="true">
                <img src={providerImages[provider]} alt="" />
              </div>
              <div className="overview-digital-main">
                <div><strong>{providerLabels[provider] || provider}</strong>
                  <span className={liveStatus?.connected ? "status connected" : "status"}>
                    {liveStatus?.connected ? t("overview.digitalCenters.connected") :
                      providerEnabled(digitalSettings) ? t("overview.digitalCenters.configured") :
                        t("overview.digitalCenters.notConfigured")}
                  </span></div>
                <dl>
                  <div><dt>{t("overview.digitalCenters.managed")}</dt><dd>{managedLocomotives}</dd></div>
                  <div><dt>{t("overview.digitalCenters.liveMonitor")}</dt><dd>{liveStatus?.connected ?
                    t("overview.digitalCenters.active") : t("overview.digitalCenters.inactive")}</dd></div>
                </dl>
                <p className="overview-digital-sync"><span>{t("overview.digitalCenters.lastSync")}</span>
                  {lastSynchronization ? <time dateTime={lastSynchronization}>
                    {formatRecentTime(lastSynchronization, language)}</time> :
                    <strong>{t("overview.digitalCenters.neverSynced")}</strong>}</p>
              </div>
              <a href="/digital-centers" className="overview-open-action"><ExternalLink size={15} aria-hidden="true" />
                {t("overview.digitalCenters.open")}</a>
            </div>
          )}
        </article>

        <article className="overview-card overview-recent-card">
          <header className="overview-card-head"><Clock3 size={19} aria-hidden="true" />
            <h2>{t("overview.recent.title")}</h2></header>
          {vehicleDependentState ? <p className={vehicleError ? "overview-module-state error" : "overview-module-state"}
            role={vehicleError ? "alert" : "status"}>{vehicleDependentState}</p> : stats.recentVehicles.length === 0 ?
            <p className="overview-module-state">{t("overview.recent.empty")}</p> : (
              <div className="overview-vehicle-list">
                {stats.recentVehicles.map((vehicle) => <a href={`/vehicles?id=${encodeURIComponent(vehicle.id)}`}
                  key={vehicle.id}>
                  <span className="overview-vehicle-thumb">
                    {primaryVehicleImage(vehicle) ? <img src={primaryVehicleImage(vehicle)} alt="" /> :
                      <TrainFront size={24} aria-hidden="true" />}
                  </span>
                  <strong>{vehicle.name || vehicle.inventoryNumber}</strong>
                  <span className="overview-vehicle-type">{vehicle.gattung || vehicle.category || "–"}</span>
                  <time dateTime={vehicle.updatedAt}>{formatRecentTime(vehicle.updatedAt, language)}</time>
                  <ArrowRight size={17} aria-hidden="true" />
                </a>)}
              </div>
            )}
        </article>

        <article className="overview-card overview-maintenance-card">
          <header className="overview-card-head"><Wrench size={19} aria-hidden="true" />
            <h2>{t("overview.nextMaintenance.title")}</h2></header>
          {vehicleDependentState ? <p className={vehicleError ? "overview-module-state error" : "overview-module-state"}
            role={vehicleError ? "alert" : "status"}>{vehicleDependentState}</p> : stats.nextMaintenance.length === 0 ?
            <p className="overview-module-state">{t("overview.nextMaintenance.empty")}</p> : (
              <div className="overview-maintenance-list">
                {stats.nextMaintenance.map(({ vehicle, entry, days }) => <a
                  href={`/vehicles?id=${encodeURIComponent(vehicle.id)}&tab=maintenance`} key={entry.id}
                  className={days <= 0 ? "overdue" : ""}>
                  <span className="overview-vehicle-thumb">
                    {primaryVehicleImage(vehicle) ? <img src={primaryVehicleImage(vehicle)} alt="" /> :
                      <TrainFront size={24} aria-hidden="true" />}
                  </span>
                  <span><strong>{vehicle.name || vehicle.inventoryNumber}</strong>
                    <small>{vehicle.gattung || vehicle.category || entry.kind}</small></span>
                  <b>{maintenanceDistanceText(days, t)}</b>
                </a>)}
              </div>
            )}
        </article>
      </section>

      <section className="overview-analysis-primary" aria-label={t("overview.analysis.label")}>
        <article className="overview-card overview-trend-card">
          <header className="overview-analysis-head"><h2>{t("overview.trend.title")}</h2>
            <div className="overview-chart-legend"><span className="vehicle">{t("overview.metric.vehicles")}</span>
              <span className="accessory">{t("overview.metric.accessories")}</span></div></header>
          {vehicleError || accessoryError ? <p className="overview-module-state error" role="alert">
            {vehicleError || accessoryError}</p> : vehicleLoading || accessoryLoading ?
            <p className="overview-module-state" role="status">{t("overview.state.loading")}</p> :
            <InventoryTrendChart points={trend} />}
        </article>

        <article className="overview-card overview-value-card">
          <header className="overview-analysis-head"><h2>{t("overview.valueDistribution.title")}</h2>
            <div className="overview-segmented" aria-label={t("overview.valueDistribution.mode")}>
              <button type="button" className={valueMode === "list" ? "active" : ""}
                onClick={() => setValueMode("list")}>{t("overview.valueDistribution.list")}</button>
              <button type="button" className={valueMode === "purchase" ? "active" : ""}
                onClick={() => setValueMode("purchase")}>{t("overview.valueDistribution.purchase")}</button>
            </div></header>
          {valuationError ? <p className="overview-module-state error" role="alert">{valuationError}</p> :
            valuationLoading ? <p className="overview-module-state" role="status">{t("overview.state.loading")}</p> :
              <div className="overview-value-body">
                <ValueDistributionChart vehicleValue={valueVehicle} accessoryValue={valueAccessory} />
                <dl>
                  <div><dt><i className="vehicle" />{t("overview.metric.vehicles")}</dt>
                    <dd>{money(valueVehicle)}</dd></div>
                  <div><dt><i className="accessory" />{t("overview.metric.accessories")}</dt>
                    <dd>{money(valueAccessory)}</dd></div>
                </dl>
              </div>}
        </article>
      </section>

      <section className="overview-analysis-secondary">
        <article className="overview-card overview-ranking-card">
          <header className="overview-analysis-head"><h2>{t("overview.structure.title")}</h2></header>
          {vehicleDependentState ? <p className={vehicleError ? "overview-module-state error" : "overview-module-state"}
            role={vehicleError ? "alert" : "status"}>{vehicleDependentState}</p> : stats.categories.length === 0 ?
            <p className="overview-module-state">{t("overview.state.empty")}</p> :
            <div className="overview-progress-list">{stats.categories.map(([label, count]) => <div key={label}>
              <span>{label}</span><progress value={count} max={Math.max(...stats.categories.map((entry) => entry[1]))} />
              <strong>{count}</strong></div>)}</div>}
        </article>
        <article className="overview-card overview-ranking-card">
          <header className="overview-analysis-head"><h2>{t("overview.manufacturers.title")}</h2></header>
          {vehicleDependentState ? <p className={vehicleError ? "overview-module-state error" : "overview-module-state"}
            role={vehicleError ? "alert" : "status"}>{vehicleDependentState}</p> : stats.manufacturers.length === 0 ?
            <p className="overview-module-state">{t("overview.state.empty")}</p> :
            <div className="overview-progress-list">{stats.manufacturers.map(([label, count]) => <div key={label}>
              <span>{label}</span><progress value={count} max={Math.max(...stats.manufacturers.map((entry) => entry[1]))} />
              <strong>{count}</strong></div>)}</div>}
        </article>
        <article className="overview-card overview-quality-card">
          <header className="overview-analysis-head"><h2>{t("overview.quality.title")}</h2></header>
          {vehicleDependentState ? <p className={vehicleError ? "overview-module-state error" : "overview-module-state"}
            role={vehicleError ? "alert" : "status"}>{vehicleDependentState}</p> :
            <div className="overview-quality-list">
              {[
                [t("overview.quality.images"), imageShare],
                [t("overview.quality.articleNumbers"), articleShare],
                [t("overview.quality.decoderAddresses"), decoderShare],
                [t("overview.quality.documented"), documentedShare]
              ].map(([label, value]) => <div key={String(label)}><b
                style={{ "--quality-value": Number(value) } as React.CSSProperties}><span>{value}%</span></b><span>{label}</span>
                <progress value={Number(value)} max={100} /></div>)}
            </div>}
        </article>
      </section>

      <OverviewMetricDialog
        open={metricDialogOpen}
        order={draftMetrics.order}
        active={draftMetrics.active}
        maxActive={metricLimit}
        t={t}
        onToggle={(metric) => setDraftMetrics((current) => ({
          ...current,
          active: current.active.includes(metric) ? current.active.filter((item) => item !== metric) :
            current.active.length < metricLimit ? [...current.active, metric] : current.active
        }))}
        onMove={(metric, target) => setDraftMetrics((current) => ({
          ...current,
          order: moveMetric(current.order, metric, target)
        }))}
        onMoveBy={(metric, direction) => setDraftMetrics((current) => {
          const index = current.order.indexOf(metric);
          const target = current.order[index + direction];
          return target ? { ...current, order: moveMetric(current.order, metric, target) } : current;
        })}
        onReset={() => setDraftMetrics({ active: defaultOverviewMetrics, order: [...overviewMetricIDs] })}
        onDone={saveMetricPreference}
        onClose={closeMetricDialog}
      />

      <footer className="overview-action-footer">
        {canEdit ? <a href="/vehicles?create=1"><TrainFront size={23} aria-hidden="true" /><span>
          <strong>{t("overview.footer.createVehicle")}</strong><small>{t("overview.footer.createVehicleHelp")}</small>
        </span></a> : <span className="overview-footer-action disabled" aria-disabled="true"><TrainFront size={23}
          aria-hidden="true" /><span><strong>{t("overview.footer.createVehicle")}</strong>
            <small>{t("overview.footer.noPermission")}</small></span></span>}
        {canEdit ? <a href="/accessories?create=1"><PackagePlus size={23} aria-hidden="true" /><span>
          <strong>{t("overview.footer.createAccessory")}</strong><small>{t("overview.footer.createAccessoryHelp")}</small>
        </span></a> : <span className="overview-footer-action disabled" aria-disabled="true"><PackagePlus size={23}
          aria-hidden="true" /><span><strong>{t("overview.footer.createAccessory")}</strong>
            <small>{t("overview.footer.noPermission")}</small></span></span>}
        {canEdit ? <a href="/import-export"><Upload size={23} aria-hidden="true" /><span>
          <strong>{t("overview.footer.import")}</strong><small>{t("overview.footer.importHelp")}</small>
        </span></a> : <span className="overview-footer-action disabled" aria-disabled="true"><FileInput size={23}
          aria-hidden="true" /><span><strong>{t("overview.footer.import")}</strong>
            <small>{t("overview.footer.noPermission")}</small></span></span>}
      </footer>
    </div>
  );
}
