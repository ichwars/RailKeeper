import {
  useLayoutEffect,
  useRef,
  useState,
  type Dispatch,
  type ReactNode,
  type SetStateAction
} from "react";
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
import { masterDataOptions } from "../../shared/masterDataOptions";
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

type QuickMenuPlacementInput = {
  anchorTop: number;
  anchorBottom: number;
  menuHeight: number;
  viewportHeight: number;
};

export function quickMenuShouldOpenAbove({
  anchorTop,
  anchorBottom,
  menuHeight,
  viewportHeight
}: QuickMenuPlacementInput) {
  const menuGap = 6;
  const spaceBelow = viewportHeight - anchorBottom;
  const spaceAbove = anchorTop;
  return spaceBelow < menuHeight + menuGap && spaceAbove > spaceBelow;
}

type VehicleQuickMenuProps = {
  vehicle: Vehicle;
  open: boolean;
  onToggle: () => void;
  onClose: () => void;
  openDetail: (vehicle: Vehicle) => void;
  openEdit: (vehicle: Vehicle, tab?: ModalTab) => void;
  openQr: (vehicle: Vehicle) => void;
  printVehicle: (vehicle: Vehicle) => void;
  setDeleteCandidate: Dispatch<SetStateAction<Vehicle | null>>;
  t: Translator;
};

function VehicleQuickMenu({
  vehicle,
  open,
  onToggle,
  onClose,
  openDetail,
  openEdit,
  openQr,
  printVehicle,
  setDeleteCandidate,
  t
}: VehicleQuickMenuProps) {
  const buttonRef = useRef<HTMLButtonElement | null>(null);
  const menuRef = useRef<HTMLDivElement | null>(null);
  const [openAbove, setOpenAbove] = useState(false);

  useLayoutEffect(() => {
    if (!open) return;
    const updatePlacement = () => {
      const anchor = buttonRef.current?.getBoundingClientRect();
      const menu = menuRef.current?.getBoundingClientRect();
      if (!anchor || !menu) return;
      setOpenAbove(quickMenuShouldOpenAbove({
        anchorTop: anchor.top,
        anchorBottom: anchor.bottom,
        menuHeight: menu.height,
        viewportHeight: window.innerHeight
      }));
    };
    updatePlacement();
    window.addEventListener("resize", updatePlacement);
    window.addEventListener("scroll", updatePlacement, true);
    return () => {
      window.removeEventListener("resize", updatePlacement);
      window.removeEventListener("scroll", updatePlacement, true);
    };
  }, [open]);

  return (
    <div className="quick-menu-wrap">
      <button
        ref={buttonRef}
        type="button"
        className={open ? "icon-button active" : "icon-button"}
        onClick={onToggle}
        aria-label={t("vehicles.quickMenu")}
        title={t("vehicles.quickMenu")}
      >
        <MoreVertical size={16} />
      </button>
      {open && (
        <div
          ref={menuRef}
          className={openAbove ? "quick-menu open-above" : "quick-menu"}
          role="menu"
        >
          <button type="button" role="menuitem" onClick={() => { onClose(); openDetail(vehicle); }}>
            <Eye size={14} />{t("vehicles.view")}
          </button>
          <button type="button" role="menuitem" onClick={() => { onClose(); openEdit(vehicle); }}>
            <Pencil size={14} />{t("vehicles.edit")}
          </button>
          <span className="quick-menu-separator" role="separator" />
          <button type="button" role="menuitem" onClick={() => { onClose(); openQr(vehicle); }}>
            <QrCode size={14} />QR-Code
          </button>
          <button type="button" role="menuitem" onClick={() => { onClose(); printVehicle(vehicle); }}>
            <Printer size={14} />{t("overview.print")}
          </button>
          <button type="button" role="menuitem" onClick={() => { onClose(); openEdit(vehicle, "uploads"); }}>
            <Upload size={14} />Uploads
          </button>
          <button type="button" role="menuitem" onClick={() => { onClose(); openEdit(vehicle, "maintenance"); }}>
            <Wrench size={14} />{t("vehicles.maintenance")}
          </button>
          <button type="button" role="menuitem" onClick={() => { onClose(); openEdit(vehicle, "spareParts"); }}>
            <PackageSearch size={14} />{t("vehicles.tab.spareParts")}
          </button>
          <span className="quick-menu-separator" role="separator" />
          <button type="button" role="menuitem" className="danger" onClick={() => { onClose(); setDeleteCandidate(vehicle); }}>
            <Trash2 size={14} />{t("vehicles.delete")}
          </button>
        </div>
      )}
    </div>
  );
}

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
    <VehicleQuickMenu
      vehicle={vehicle}
      open={quickMenuVehicleID === vehicle.id}
      onToggle={() => setQuickMenuVehicleID((current) => (
        current === vehicle.id ? "" : vehicle.id
      ))}
      onClose={() => setQuickMenuVehicleID("")}
      openDetail={openDetail}
      openEdit={openEdit}
      openQr={openQr}
      printVehicle={printVehicle}
      setDeleteCandidate={setDeleteCandidate}
      t={t}
    />
  );

  const selectOptions = (
    items: MasterDataEntry[],
    currentValue: string,
    emptyLabel = "Keine Auswahl"
  ): ReactNode => (
    <>
      <option value="">{emptyLabel}</option>
      {masterDataOptions(items, [currentValue], optionValue).map((option) => (
        <option key={option.id} value={option.value} disabled={!option.active}>
          {option.label}{option.active ? "" : ` (${t("common.inactive")})`}
        </option>
      ))}
    </>
  );

  return { sortHeader, vehicleQuickMenu, selectOptions };
}
