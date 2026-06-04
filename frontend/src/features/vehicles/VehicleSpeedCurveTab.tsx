import { useEffect, useMemo, useState } from "react";
import { Gauge } from "lucide-react";
import type { Vehicle, VehicleCVValueInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { buildSpeedCurveProfiles, SpeedCurvePoint, SpeedCurveProfile } from "./speedCurve";

type ECoSDraftPreview = {
  cvValues: VehicleCVValueInput[];
  sourceSummary: {
    name: string;
    objectId: number;
  };
};

export function VehicleSpeedCurveTab({
  selected,
  ecosDraft
}: {
  selected: Vehicle | null;
  ecosDraft: ECoSDraftPreview | null;
}) {
  const { t } = useI18n();
  const sourceValues = selected?.cvValues || ecosDraft?.cvValues || [];
  const profiles = useMemo(
    () => buildSpeedCurveProfiles(sourceValues, t("vehicles.speedCurve.noProfile")),
    [sourceValues, t]
  );
  const [selectedProfileId, setSelectedProfileId] = useState("");

  useEffect(() => {
    if (profiles.length === 0) {
      setSelectedProfileId("");
      return;
    }
    if (!profiles.some((profile) => profile.id === selectedProfileId)) {
      setSelectedProfileId(profiles[0].id);
    }
  }, [profiles, selectedProfileId]);

  const activeProfile = profiles.find((profile) => profile.id === selectedProfileId) || profiles[0];
  const activePoints = activeProfile ? pointsForPrimaryCurve(activeProfile) : [];
  const hasCurve = Boolean(activeProfile && activePoints.length > 0);

  return (
    <section className="speed-curve-tab">
      <div className="upload-head">
        <div>
          <h3>{t("vehicles.speedCurve.title")}</h3>
          <p>{t("vehicles.speedCurve.subtitle")}</p>
        </div>
        <span className="speed-curve-source-badge">
          <Gauge size={15} aria-hidden="true" />
          {t("vehicles.speedCurve.readOnly")}
        </span>
      </div>

      {profiles.length === 0 && (
        <p className="empty-state compact">{t("vehicles.speedCurve.empty")}</p>
      )}

      {profiles.length > 0 && activeProfile && (
        <>
          <div className="speed-profile-list" aria-label={t("vehicles.speedCurve.profile")}>
            {profiles.map((profile) => (
              <button
                key={profile.id}
                type="button"
                className={profile.id === activeProfile.id ? "active" : ""}
                onClick={() => setSelectedProfileId(profile.id)}
              >
                <strong>{profile.label}</strong>
                <span>{t("vehicles.speedCurve.sourceCount", { count: profile.sourceCount })}</span>
              </button>
            ))}
          </div>

          <div className="speed-curve-summary">
            <div>
              <span>{t("vehicles.speedCurve.mode")}</span>
              <strong>{curveModeLabel(activeProfile, t)}</strong>
            </div>
            <div>
              <span>CV 29</span>
              <strong>{cv29Label(activeProfile, t)}</strong>
            </div>
            <div>
              <span>{t("vehicles.speedCurve.points")}</span>
              <strong>{activePoints.length}</strong>
            </div>
            <div>
              <span>{t("vehicles.speedCurve.trim")}</span>
              <strong>{trimLabel(activeProfile)}</strong>
            </div>
          </div>

          <section className={hasCurve ? "speed-chart-panel" : "speed-chart-panel empty"}>
            <div className="speed-chart-head">
              <div>
                <h4>{t("vehicles.speedCurve.chartTitle")}</h4>
                <p>{chartSubtitle(activeProfile, t)}</p>
              </div>
              <span>{t("vehicles.speedCurve.realValues")}</span>
            </div>
            {hasCurve ? (
              <SpeedCurveChart
                points={activePoints}
                ariaLabel={t("vehicles.speedCurve.chartAria")}
                stepLabel={t("vehicles.speedCurve.axisStep")}
                valueLabel={t("vehicles.speedCurve.axisValue")}
              />
            ) : (
              <p className="empty-state compact">{t("vehicles.speedCurve.noCurve")}</p>
            )}
          </section>

          <section className="speed-curve-detail-grid">
            <article>
              <h4>{t("vehicles.speedCurve.threePoint")}</h4>
              <SpeedPointList points={activeProfile.threePoint} emptyText={t("vehicles.speedCurve.noThreePoint")} />
              {activeProfile.missingThreePointCVs.length > 0 && (
                <p className="speed-missing">{t("vehicles.speedCurve.missingCVs", { cvs: activeProfile.missingThreePointCVs.join(", ") })}</p>
              )}
            </article>
            <article>
              <h4>{t("vehicles.speedCurve.speedTable")}</h4>
              <SpeedPointList points={activeProfile.speedTable} emptyText={t("vehicles.speedCurve.noSpeedTable")} compact />
              {activeProfile.missingSpeedTableCVs.length > 0 && activeProfile.speedTable.length > 0 && (
                <p className="speed-missing">{t("vehicles.speedCurve.missingCount", { count: activeProfile.missingSpeedTableCVs.length })}</p>
              )}
            </article>
          </section>

          {ecosDraft && !selected && (
            <p className="ecos-draft-mapping-preview">
              {t("vehicles.ecosDraft.mappingPreviewTitle")}: {ecosDraft.sourceSummary.name} · #{ecosDraft.sourceSummary.objectId}
            </p>
          )}
        </>
      )}
    </section>
  );
}

function pointsForPrimaryCurve(profile: SpeedCurveProfile) {
  if (profile.primaryCurve === "speedTable") return profile.speedTable;
  if (profile.primaryCurve === "threePoint") return profile.threePoint;
  return [];
}

function curveModeLabel(profile: SpeedCurveProfile, t: (key: string, values?: Record<string, string | number>) => string) {
  if (profile.primaryCurve === "speedTable") return t("vehicles.speedCurve.modeSpeedTable");
  if (profile.primaryCurve === "threePoint") return t("vehicles.speedCurve.modeThreePoint");
  return t("vehicles.speedCurve.modeUnknown");
}

function cv29Label(profile: SpeedCurveProfile, t: (key: string, values?: Record<string, string | number>) => string) {
  if (profile.cv29Value === null) return t("vehicles.speedCurve.cv29Unknown");
  return profile.speedTableActive
    ? t("vehicles.speedCurve.cv29SpeedTable", { value: profile.cv29Value })
    : t("vehicles.speedCurve.cv29ThreePoint", { value: profile.cv29Value });
}

function chartSubtitle(profile: SpeedCurveProfile, t: (key: string, values?: Record<string, string | number>) => string) {
  if (profile.primaryCurve === "speedTable") {
    return t("vehicles.speedCurve.speedTableSubtitle", { count: profile.speedTable.length });
  }
  if (profile.primaryCurve === "threePoint") {
    return t("vehicles.speedCurve.threePointSubtitle", { count: profile.threePoint.length });
  }
  return t("vehicles.speedCurve.noCurve");
}

function trimLabel(profile: SpeedCurveProfile) {
  const forward = profile.forwardTrim === null ? "-" : profile.forwardTrim;
  const reverse = profile.reverseTrim === null ? "-" : profile.reverseTrim;
  return `${forward} / ${reverse}`;
}

function SpeedCurveChart({
  points,
  ariaLabel,
  stepLabel,
  valueLabel
}: {
  points: SpeedCurvePoint[];
  ariaLabel: string;
  stepLabel: string;
  valueLabel: string;
}) {
  const width = 760;
  const height = 300;
  const pad = { top: 20, right: 18, bottom: 38, left: 46 };
  const plotWidth = width - pad.left - pad.right;
  const plotHeight = height - pad.top - pad.bottom;
  const xForStep = (step: number) => pad.left + ((step - 1) / 27) * plotWidth;
  const yForValue = (value: number) => pad.top + (1 - Math.max(0, Math.min(255, value)) / 255) * plotHeight;
  const plotted = points.map((point) => ({
    ...point,
    x: xForStep(point.step),
    y: yForValue(point.value)
  }));
  const linePath = plotted.map((point, index) => `${index === 0 ? "M" : "L"} ${point.x.toFixed(2)} ${point.y.toFixed(2)}`).join(" ");
  const areaPath = plotted.length > 1
    ? `${linePath} L ${plotted[plotted.length - 1].x.toFixed(2)} ${pad.top + plotHeight} L ${plotted[0].x.toFixed(2)} ${pad.top + plotHeight} Z`
    : "";
  const gridValues = [0, 64, 128, 192, 255];
  const stepLabels = [1, 7, 14, 21, 28];

  return (
    <svg className="speed-curve-chart" viewBox={`0 0 ${width} ${height}`} role="img" aria-label={ariaLabel}>
      <rect className="speed-chart-bg" x={pad.left} y={pad.top} width={plotWidth} height={plotHeight} rx="7" />
      {gridValues.map((value) => {
        const y = yForValue(value);
        return (
          <g key={value}>
            <line className="speed-chart-grid" x1={pad.left} x2={pad.left + plotWidth} y1={y} y2={y} />
            <text className="speed-chart-label" x={pad.left - 10} y={y + 4} textAnchor="end">{value}</text>
          </g>
        );
      })}
      {stepLabels.map((step) => {
        const x = xForStep(step);
        return (
          <g key={step}>
            <line className="speed-chart-step" x1={x} x2={x} y1={pad.top} y2={pad.top + plotHeight} />
            <text className="speed-chart-label" x={x} y={height - 12} textAnchor="middle">{step}</text>
          </g>
        );
      })}
      <text className="speed-chart-axis-title" x={pad.left + plotWidth / 2} y={height - 2} textAnchor="middle">{stepLabel}</text>
      <text className="speed-chart-axis-title" transform={`translate(12 ${pad.top + plotHeight / 2}) rotate(-90)`} textAnchor="middle">{valueLabel}</text>
      {areaPath && <path className="speed-chart-area" d={areaPath} />}
      {linePath && <path className="speed-chart-line" d={linePath} />}
      {plotted.map((point) => (
        <g key={`${point.cvNumber}-${point.step}`}>
          <circle className="speed-chart-point" cx={point.x} cy={point.y} r="5" />
          <title>{`${point.label}: ${point.value}`}</title>
        </g>
      ))}
    </svg>
  );
}

function SpeedPointList({
  points,
  emptyText,
  compact = false
}: {
  points: SpeedCurvePoint[];
  emptyText: string;
  compact?: boolean;
}) {
  if (points.length === 0) {
    return <p className="empty-state compact">{emptyText}</p>;
  }
  return (
    <div className={compact ? "speed-point-list compact" : "speed-point-list"}>
      {points.map((point) => (
        <span key={`${point.cvNumber}-${point.step}`}>
          <strong>{point.label}</strong>
          <em>{point.value}</em>
        </span>
      ))}
    </div>
  );
}
