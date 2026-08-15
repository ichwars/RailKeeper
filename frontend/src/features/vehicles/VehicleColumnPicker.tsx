import { useEffect, useRef, useState } from "react";
import { ChevronDown, ChevronUp, Columns3, RotateCcw } from "lucide-react";

import { useI18n } from "../../shared/i18n";
import {
  vehicleColumnLabel,
  vehicleTableColumns,
  type VehicleColumnGroup,
  type VehicleColumnMove,
  type VehicleTableColumn
} from "./vehicleTableColumns";

const groups: VehicleColumnGroup[] = [
  "identity",
  "digital",
  "ownership",
  "technical",
  "equipment"
];

type VehicleColumnPickerProps = {
  columns: readonly VehicleTableColumn[];
  loading: boolean;
  onToggle: (column: VehicleTableColumn) => void;
  onMove: (column: VehicleTableColumn, direction: VehicleColumnMove) => void;
  onReset: () => void;
};

export function VehicleColumnPicker({
  columns,
  loading,
  onToggle,
  onMove,
  onReset
}: VehicleColumnPickerProps) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement | null>(null);
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const { t } = useI18n();

  useEffect(() => {
    if (!open) return;

    const closeOutside = (event: PointerEvent) => {
      if (event.target instanceof Node && rootRef.current?.contains(event.target)) return;
      setOpen(false);
    };
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key !== "Escape") return;
      event.preventDefault();
      setOpen(false);
      triggerRef.current?.focus();
    };

    document.addEventListener("pointerdown", closeOutside);
    document.addEventListener("keydown", closeOnEscape);
    return () => {
      document.removeEventListener("pointerdown", closeOutside);
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [open]);

  return (
    <div ref={rootRef} className="vehicle-column-picker">
      <button
        ref={triggerRef}
        type="button"
        className={`icon-button${open ? " active" : ""}`}
        aria-label={t("vehicles.columns.choose")}
        title={t("vehicles.columns.choose")}
        aria-haspopup="dialog"
        aria-expanded={open}
        disabled={loading}
        onClick={() => setOpen((current) => !current)}
      >
        <Columns3 size={16} aria-hidden="true" />
      </button>
      {open ? (
        <div
          className="vehicle-column-picker-popover"
          role="dialog"
          aria-label={t("vehicles.columns.dialogTitle")}
        >
          <section className="vehicle-column-order" aria-label={t("vehicles.columns.visibleOrder")}>
            <h3>{t("vehicles.columns.visibleOrder")}</h3>
            {columns.map((column, index) => {
              const label = vehicleColumnLabel(column, t);
              return (
                <div key={column} className="vehicle-column-order-row">
                  <span>{label}</span>
                  <button
                    type="button"
                    className="icon-button"
                    disabled={index === 0}
                    aria-label={t("vehicles.columns.moveUp", { label })}
                    title={t("vehicles.columns.moveUp", { label })}
                    onClick={() => onMove(column, "up")}
                  >
                    <ChevronUp size={15} aria-hidden="true" />
                  </button>
                  <button
                    type="button"
                    className="icon-button"
                    disabled={index === columns.length - 1}
                    aria-label={t("vehicles.columns.moveDown", { label })}
                    title={t("vehicles.columns.moveDown", { label })}
                    onClick={() => onMove(column, "down")}
                  >
                    <ChevronDown size={15} aria-hidden="true" />
                  </button>
                </div>
              );
            })}
          </section>
          <section aria-label={t("vehicles.columns.available")}>
            <h3>{t("vehicles.columns.available")}</h3>
            <div className="vehicle-column-groups">
              {groups.map((group) => (
                <fieldset key={group}>
                  <legend>{t(`vehicles.columns.group.${group}`)}</legend>
                  {vehicleTableColumns
                    .filter((definition) => definition.group === group)
                    .map(({ key }) => (
                      <label key={key}>
                        <input
                          type="checkbox"
                          checked={columns.includes(key)}
                          onChange={() => onToggle(key)}
                        />
                        <span>{vehicleColumnLabel(key, t)}</span>
                      </label>
                    ))}
                </fieldset>
              ))}
            </div>
          </section>
          <button type="button" className="vehicle-column-reset" onClick={onReset}>
            <RotateCcw size={15} aria-hidden="true" />
            {t("vehicles.columns.reset")}
          </button>
        </div>
      ) : null}
    </div>
  );
}
