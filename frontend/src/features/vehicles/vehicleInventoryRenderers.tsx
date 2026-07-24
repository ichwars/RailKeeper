import type { Dispatch, ReactNode, SetStateAction } from "react";
import {
  ArrowUpDown,
  ChevronDown,
  ChevronUp,
  Circle,
  CircleOff,
  Cpu,
  Eye,
  Image,
  ImageOff,
  MoreVertical,
  PackageSearch,
  Pencil,
  Printer,
  QrCode,
  Trash2,
  Upload,
  Wrench
} from "lucide-react";

import type { MasterDataEntry, Vehicle } from "../../shared/api";
import { optionValue, type ModalTab, type SortDirection, type SortKey } from "./vehicleViewModel";

type Translator = (key: string, values?: Record<string, string | number>) => string;

type VehicleInventoryRenderersOptions = {
  sort: { key: SortKey; direction: SortDirection };
  quickMenuVehicleID: string;
  setQuickMenuVehicleID: Dispatch<SetStateAction<string>>;
  toggleSort: (key: SortKey) => void;
  openDetail: (vehicle: Vehicle) => void;
  openEdit: (vehicle: Vehicle, tab?: ModalTab) => void;
  openQr: (vehicle: Vehicle) => void;
  printVehicle: (vehicle: Vehicle) => void;
  setDeleteCandidate: Dispatch<SetStateAction<Vehicle | null>>;
  t: Translator;
};

export function createVehicleFilterDefinitions({
  vehicleCount,
  counts,
  t
}: {
  vehicleCount: number;
  counts: {
    all: number;
    digital: number;
    analog: number;
    withImages: number;
    withoutImages: number;
    maintenanceDue: number;
    withoutMaintenance: number;
  };
  t: Translator;
}) {
  return {
    inventoryFilters: [
      { key: "all" as const, label: t("vehicles.filter.all"), count: counts.all },
      { key: "digital" as const, label: t("vehicles.filter.digital"), count: counts.digital, icon: <Cpu size={15} aria-hidden="true" /> },
      { key: "analog" as const, label: t("vehicles.filter.analog"), count: counts.analog, icon: <Circle size={15} aria-hidden="true" /> },
      { key: "withImages" as const, label: t("vehicles.filter.withImages"), count: counts.withImages, icon: <Image size={15} aria-hidden="true" /> },
      { key: "withoutImages" as const, label: t("vehicles.filter.withoutImages"), count: counts.withoutImages, icon: <ImageOff size={15} aria-hidden="true" /> }
    ],
    maintenanceFilters: [
      { key: "all" as const, label: t("vehicles.filter.all"), count: vehicleCount },
      { key: "due" as const, label: t("vehicles.filter.maintenanceDue"), count: counts.maintenanceDue, icon: <Wrench size={15} aria-hidden="true" /> },
      { key: "none" as const, label: t("vehicles.filter.withoutMaintenance"), count: counts.withoutMaintenance, icon: <CircleOff size={15} aria-hidden="true" /> }
    ]
  };
}

export function createVehicleInventoryRenderers({
  sort,
  quickMenuVehicleID,
  setQuickMenuVehicleID,
  toggleSort,
  openDetail,
  openEdit,
  openQr,
  printVehicle,
  setDeleteCandidate,
  t
}: VehicleInventoryRenderersOptions) {
  const sortHeader = (key: SortKey) => (
    <button
      type="button"
      className={`sort-button ${sort.key === key ? "active" : ""}`}
      onClick={() => toggleSort(key)}
      title={t("common.sort", { label: t(`vehicle.field.${key}`) })}
    >
      {t(`vehicle.field.${key}`)}
      {sort.key === key
        ? sort.direction === "asc"
          ? <ChevronUp size={14} />
          : <ChevronDown size={14} />
        : <ArrowUpDown size={13} />}
    </button>
  );

  const vehicleQuickMenu = (vehicle: Vehicle) => (
    <div className="quick-menu-wrap">
      <button
        type="button"
        className={quickMenuVehicleID === vehicle.id ? "icon-button active" : "icon-button"}
        onClick={() => setQuickMenuVehicleID((current) => current === vehicle.id ? "" : vehicle.id)}
        aria-label={t("vehicles.quickMenu")}
        title={t("vehicles.quickMenu")}
      >
        <MoreVertical size={16} />
      </button>
      {quickMenuVehicleID === vehicle.id && (
        <div className="quick-menu" role="menu">
          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); openDetail(vehicle); }}>
            <Eye size={14} />{t("vehicles.view")}
          </button>
          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); openEdit(vehicle); }}>
            <Pencil size={14} />{t("vehicles.edit")}
          </button>
          <span className="quick-menu-separator" role="separator" />
          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); openQr(vehicle); }}>
            <QrCode size={14} />QR-Code
          </button>
          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); printVehicle(vehicle); }}>
            <Printer size={14} />{t("overview.print")}
          </button>
          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); openEdit(vehicle, "uploads"); }}>
            <Upload size={14} />Uploads
          </button>
          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); openEdit(vehicle, "maintenance"); }}>
            <Wrench size={14} />{t("vehicles.maintenance")}
          </button>
          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); openEdit(vehicle, "spareParts"); }}>
            <PackageSearch size={14} />{t("vehicles.tab.spareParts")}
          </button>
          <span className="quick-menu-separator" role="separator" />
          <button type="button" role="menuitem" className="danger" onClick={() => { setQuickMenuVehicleID(""); setDeleteCandidate(vehicle); }}>
            <Trash2 size={14} />{t("vehicles.delete")}
          </button>
        </div>
      )}
    </div>
  );

  const selectOptions = (items: MasterDataEntry[], emptyLabel = "Keine Auswahl"): ReactNode => (
    <>
      <option value="">{emptyLabel}</option>
      {items.map((entry) => (
        <option key={entry.key} value={optionValue(entry)}>{entry.label}</option>
      ))}
    </>
  );

  return { sortHeader, vehicleQuickMenu, selectOptions };
}
