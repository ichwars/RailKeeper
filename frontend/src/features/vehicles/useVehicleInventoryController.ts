import { useCallback, useEffect, useMemo, useState } from "react";

import type { Vehicle } from "../../shared/api";
import { maintenanceDaysUntilDue, maintenanceIsDue } from "./vehicleMaintenance";
import {
  inventoryViewMode,
  inventoryViewSettingKey,
  valueForSort
} from "./vehicleViewModel";
import type {
  InventoryFilter,
  InventoryQualityFilter,
  InventoryViewMode,
  MaintenanceFilter,
  MaintenanceReminder,
  SortDirection,
  SortKey
} from "./vehicleViewModel";

type InventorySort = {
  key: SortKey;
  direction: SortDirection;
};

type InventoryFilterPreset = {
  inventoryFilter?: InventoryFilter;
  maintenanceFilter?: MaintenanceFilter;
  qualityFilter?: InventoryQualityFilter;
};

function filterPresetFromLocation(): InventoryFilterPreset {
  if (typeof window === "undefined") {
    return {};
  }

  const gap = new URLSearchParams(window.location.search).get("gap");

  if (gap === "no-main-image") {
    return { inventoryFilter: "withoutImages" };
  }
  if (gap === "no-article-number") {
    return { qualityFilter: "missingArticleNumber" };
  }
  if (gap === "no-ean") {
    return { qualityFilter: "missingEan" };
  }
  if (gap === "digital-no-decoder") {
    return { qualityFilter: "digitalMissingDecoder" };
  }

  return {};
}

function clearVehicleFilterURL() {
  if (typeof window === "undefined" || !window.location.pathname.startsWith("/vehicles") || !window.location.search) {
    return;
  }

  window.history.replaceState(null, "", "/vehicles");
}

export function useVehicleInventoryController(vehicles: Vehicle[]) {
  const initialPreset = filterPresetFromLocation();
  const [selectedVehicleIDs, setSelectedVehicleIDs] = useState<Set<string>>(() => new Set());
  const [inventoryView, setInventoryView] = useState<InventoryViewMode>(inventoryViewMode);
  const [inventoryFilter, setInventoryFilter] = useState<InventoryFilter>(initialPreset.inventoryFilter || "all");
  const [maintenanceFilter, setMaintenanceFilter] = useState<MaintenanceFilter>(initialPreset.maintenanceFilter || "all");
  const [qualityFilter, setQualityFilterState] = useState<InventoryQualityFilter>(initialPreset.qualityFilter || "none");
  const [manufacturerFilter, setManufacturerFilter] = useState("");
  const [categoryFilter, setCategoryFilter] = useState("");
  const [gattungFilter, setGattungFilter] = useState("");
  const [exhibitionReadyFilter, setExhibitionReadyFilter] = useState(false);
  const [sort, setSort] = useState<InventorySort>({
    key: "inventoryNumber",
    direction: "asc"
  });

  useEffect(() => {
    const availableIDs = new Set(vehicles.map((vehicle) => vehicle.id));

    setSelectedVehicleIDs((current) => {
      const next = new Set(Array.from(current).filter((id) => availableIDs.has(id)));
      return next.size === current.size ? current : next;
    });
  }, [vehicles]);

  const applyFilterPreset = useCallback((preset: InventoryFilterPreset) => {
    setInventoryFilter(preset.inventoryFilter || "all");
    setMaintenanceFilter(preset.maintenanceFilter || "all");
    setQualityFilterState(preset.qualityFilter || "none");
    setManufacturerFilter("");
    setCategoryFilter("");
    setGattungFilter("");
    setExhibitionReadyFilter(false);
  }, []);

  useEffect(() => {
    const applyURLPreset = () => {
      if (window.location.pathname.startsWith("/vehicles")) {
        applyFilterPreset(filterPresetFromLocation());
      }
    };

    window.addEventListener("popstate", applyURLPreset);
    return () => window.removeEventListener("popstate", applyURLPreset);
  }, [applyFilterPreset]);

  const inventoryFilterCounts = useMemo(() => {
    const withImages = vehicles.filter((vehicle) => (vehicle.images || []).length > 0).length;
    const digital = vehicles.filter((vehicle) => vehicle.digital).length;
    const maintenanceDue = vehicles.filter((vehicle) => (vehicle.maintenance || []).some(maintenanceIsDue)).length;
    const exhibitionReady = vehicles.filter((vehicle) => vehicle.exhibitionReady).length;

    return {
      all: vehicles.length,
      digital,
      analog: vehicles.length - digital,
      withImages,
      withoutImages: vehicles.length - withImages,
      maintenanceDue,
      withoutMaintenance: vehicles.length - maintenanceDue,
      exhibitionReady
    };
  }, [vehicles]);

  const inventoryFilterOptions = useMemo(() => {
    const unique = (values: string[]) =>
      Array.from(new Set(values.map((value) => value.trim()).filter(Boolean))).sort((left, right) =>
        left.localeCompare(right, "de-DE", { sensitivity: "base" })
      );

    const gattungSource = categoryFilter ? vehicles.filter((vehicle) => vehicle.category === categoryFilter) : vehicles;

    return {
      manufacturers: unique(vehicles.map((vehicle) => vehicle.manufacturer)),
      categories: unique(vehicles.map((vehicle) => vehicle.category || "")),
      gattungen: unique(gattungSource.map((vehicle) => vehicle.gattung || ""))
    };
  }, [categoryFilter, vehicles]);

  const filteredVehicles = useMemo(() => {
    return vehicles.filter((vehicle) => {
      const maintenanceDue = (vehicle.maintenance || []).some(maintenanceIsDue);

      if (inventoryFilter === "digital" && !vehicle.digital) return false;
      if (inventoryFilter === "analog" && vehicle.digital) return false;
      if (inventoryFilter === "withImages" && (vehicle.images || []).length === 0) return false;
      if (inventoryFilter === "withoutImages" && (vehicle.images || []).length > 0) return false;
      if (maintenanceFilter === "due" && !maintenanceDue) return false;
      if (maintenanceFilter === "none" && maintenanceDue) return false;
      if (qualityFilter === "missingArticleNumber" && vehicle.articleNumber?.trim()) return false;
      if (qualityFilter === "missingEan" && vehicle.ean?.trim()) return false;
      if (qualityFilter === "digitalMissingDecoder" && (!vehicle.digital || vehicle.digitalDecoderNumber?.trim() || vehicle.dtDecoderNumber?.trim())) return false;
      if (manufacturerFilter && vehicle.manufacturer !== manufacturerFilter) return false;
      if (categoryFilter && vehicle.category !== categoryFilter) return false;
      if (gattungFilter && vehicle.gattung !== gattungFilter) return false;
      if (exhibitionReadyFilter && !vehicle.exhibitionReady) return false;

      return true;
    });
  }, [categoryFilter, exhibitionReadyFilter, gattungFilter, inventoryFilter, maintenanceFilter, manufacturerFilter, qualityFilter, vehicles]);

  const sortedVehicles = useMemo(() => {
    return [...filteredVehicles].sort((left, right) => {
      const result = valueForSort(left, sort.key).localeCompare(valueForSort(right, sort.key), "de-DE", {
        numeric: true,
        sensitivity: "base"
      });

      return sort.direction === "asc" ? result : -result;
    });
  }, [filteredVehicles, sort]);

  const selectedVisibleVehicles = useMemo(
    () => sortedVehicles.filter((vehicle) => selectedVehicleIDs.has(vehicle.id)),
    [selectedVehicleIDs, sortedVehicles]
  );

  const allVisibleSelected = sortedVehicles.length > 0 && sortedVehicles.every((vehicle) => selectedVehicleIDs.has(vehicle.id));
  const someVisibleSelected = sortedVehicles.some((vehicle) => selectedVehicleIDs.has(vehicle.id));

  const maintenanceReminders = useMemo<MaintenanceReminder[]>(() => {
    return vehicles
      .flatMap((vehicle) =>
        (vehicle.maintenance || []).flatMap((entry) => {
          const daysUntilDue = maintenanceDaysUntilDue(entry);

          if (daysUntilDue === null || daysUntilDue > 14) return [];

          return [{ vehicle, entry, daysUntilDue }];
        })
      )
      .sort((left, right) => left.daysUntilDue - right.daysUntilDue || left.vehicle.inventoryNumber.localeCompare(right.vehicle.inventoryNumber, "de-DE"));
  }, [vehicles]);

  const maintenanceReminderSummary = {
    due: maintenanceReminders.filter((item) => item.daysUntilDue <= 0).length,
    upcoming: maintenanceReminders.filter((item) => item.daysUntilDue > 0).length
  };
  const nextMaintenanceReminder = maintenanceReminders[0] || null;

  const inventorySummary = useMemo(() => {
    const categories = new Set(vehicles.map((vehicle) => vehicle.category).filter(Boolean));
    const digital = vehicles.filter((vehicle) => vehicle.digital).length;
    const withImages = vehicles.filter((vehicle) => (vehicle.images || []).length > 0).length;

    return {
      categories: categories.size,
      digital,
      analog: vehicles.length - digital,
      withImages
    };
  }, [vehicles]);

  const hasActiveInventoryFilters =
    inventoryFilter !== "all" ||
    maintenanceFilter !== "all" ||
    qualityFilter !== "none" ||
    Boolean(manufacturerFilter || categoryFilter || gattungFilter || exhibitionReadyFilter);

  const resetInventoryFilters = () => {
    clearVehicleFilterURL();
    setInventoryFilter("all");
    setMaintenanceFilter("all");
    setQualityFilterState("none");
    setManufacturerFilter("");
    setCategoryFilter("");
    setGattungFilter("");
    setExhibitionReadyFilter(false);
  };

  const toggleVehicleSelection = (vehicleID: string) => {
    setSelectedVehicleIDs((current) => {
      const next = new Set(current);

      if (next.has(vehicleID)) {
        next.delete(vehicleID);
      } else {
        next.add(vehicleID);
      }

      return next;
    });
  };

  const toggleAllVisibleSelection = () => {
    setSelectedVehicleIDs((current) => {
      const next = new Set(current);

      if (allVisibleSelected) {
        sortedVehicles.forEach((vehicle) => next.delete(vehicle.id));
      } else {
        sortedVehicles.forEach((vehicle) => next.add(vehicle.id));
      }

      return next;
    });
  };

  const toggleSort = (key: SortKey) => {
    setSort((current) => ({
      key,
      direction: current.key === key && current.direction === "asc" ? "desc" : "asc"
    }));
  };

  const setInventoryViewMode = (modeName: InventoryViewMode) => {
    setInventoryView(modeName);
    window.localStorage.setItem(inventoryViewSettingKey, modeName);
  };

  const setQualityFilter = (filter: InventoryQualityFilter) => {
    if (filter === "none") {
      clearVehicleFilterURL();
    }
    setQualityFilterState(filter);
  };

  return {
    allVisibleSelected,
    categoryFilter,
    exhibitionReadyFilter,
    filteredVehicles,
    gattungFilter,
    hasActiveInventoryFilters,
    inventoryFilter,
    inventoryFilterCounts,
    inventoryFilterOptions,
    inventorySummary,
    inventoryView,
    maintenanceFilter,
    maintenanceReminderSummary,
    maintenanceReminders,
    manufacturerFilter,
    nextMaintenanceReminder,
    qualityFilter,
    resetInventoryFilters,
    selectedVehicleIDs,
    selectedVisibleVehicles,
    setCategoryFilter,
    setExhibitionReadyFilter,
    setGattungFilter,
    setInventoryFilter,
    setInventoryViewMode,
    setMaintenanceFilter,
    setManufacturerFilter,
    setQualityFilter,
    someVisibleSelected,
    sort,
    sortedVehicles,
    toggleAllVisibleSelection,
    toggleSort,
    toggleVehicleSelection
  };
}
