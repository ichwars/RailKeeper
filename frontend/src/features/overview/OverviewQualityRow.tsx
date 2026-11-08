import { useId, type CSSProperties } from "react";

export function OverviewQualityRow({ label, value }: { label: string; value: number }) {
  const tooltipID = useId();
  return <div className="overview-quality-row" tabIndex={0} aria-describedby={tooltipID}>
    <b aria-hidden="true" style={{ "--quality-value": value } as CSSProperties} />
    <span>{label}</span>
    <progress aria-label={label} value={value} max={100} />
    <span id={tooltipID} role="tooltip" className="overview-quality-tooltip">{label}: {value}%</span>
  </div>;
}
