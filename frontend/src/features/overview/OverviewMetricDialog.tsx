import { useEffect, useRef, useState } from "react";
import { GripVertical } from "lucide-react";

import { AppCheckbox } from "../../shared/ui/AppCheckbox";
import type { OverviewMetricID } from "./overviewModel";

type OverviewMetricDialogProps = {
  open: boolean;
  order: OverviewMetricID[];
  active: OverviewMetricID[];
  maxActive: number;
  t: (key: string, values?: Record<string, string | number>) => string;
  onToggle: (metric: OverviewMetricID) => void;
  onMove: (metric: OverviewMetricID, target: OverviewMetricID) => void;
  onMoveBy: (metric: OverviewMetricID, direction: -1 | 1) => void;
  onReset: () => void;
  onDone: () => void;
  onClose: () => void;
};

export function OverviewMetricDialog({
  open,
  order,
  active,
  maxActive,
  t,
  onToggle,
  onMove,
  onMoveBy,
  onReset,
  onDone,
  onClose
}: OverviewMetricDialogProps) {
  const dialogRef = useRef<HTMLDivElement>(null);
  const [draggedMetric, setDraggedMetric] = useState<OverviewMetricID | null>(null);

  useEffect(() => {
    if (!open) return;
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    const closeOnPointerDown = (event: PointerEvent) => {
      if (!(event.target instanceof Element)) return;
      if (dialogRef.current?.contains(event.target) || event.target.closest(".overview-metric-trigger")) return;
      onClose();
    };
    document.addEventListener("keydown", closeOnEscape);
    document.addEventListener("pointerdown", closeOnPointerDown);
    return () => {
      document.removeEventListener("keydown", closeOnEscape);
      document.removeEventListener("pointerdown", closeOnPointerDown);
    };
  }, [onClose, open]);

  if (!open) return null;

  return (
    <div className="overview-metric-dialog" ref={dialogRef} role="dialog"
      aria-labelledby="overview-metric-dialog-title">
      <header>
        <h2 id="overview-metric-dialog-title">{t("overview.metrics.dialogTitle")}</h2>
        <p>{t("overview.metrics.dialogHelp", { count: maxActive })}</p>
      </header>
      <div className="overview-metric-options">
        {order.map((metric) => {
          const checked = active.includes(metric);
          const disabled = !checked && active.length >= maxActive;
          return (
            <div
              className={`overview-metric-option${draggedMetric === metric ? " dragging" : ""}`}
              key={metric}
              onDragOver={(event) => event.preventDefault()}
              onDrop={() => {
                if (draggedMetric && draggedMetric !== metric) onMove(draggedMetric, metric);
                setDraggedMetric(null);
              }}
            >
              <button
                type="button"
                className="overview-metric-grip"
                draggable
                aria-label={t("overview.metrics.reorder", { label: t(`overview.metric.${metric}`) })}
                title={t("overview.metrics.reorderHelp")}
                onDragStart={() => setDraggedMetric(metric)}
                onDragEnd={() => setDraggedMetric(null)}
                onKeyDown={(event) => {
                  if (event.key === "ArrowUp") {
                    event.preventDefault();
                    onMoveBy(metric, -1);
                  }
                  if (event.key === "ArrowDown") {
                    event.preventDefault();
                    onMoveBy(metric, 1);
                  }
                }}
              >
                <GripVertical size={17} aria-hidden="true" />
              </button>
              <AppCheckbox
                checked={checked}
                disabled={disabled}
                label={t(`overview.metric.${metric}`)}
                onChange={() => onToggle(metric)}
              />
            </div>
          );
        })}
      </div>
      <p className="overview-metric-order-help">{t("overview.metrics.orderHelp")}</p>
      <footer>
        <button type="button" className="secondary-button" onClick={onReset}>
          {t("overview.metrics.reset")}
        </button>
        <button type="button" className="primary-button" onClick={onDone}>
          {t("overview.metrics.done")}
        </button>
      </footer>
    </div>
  );
}
