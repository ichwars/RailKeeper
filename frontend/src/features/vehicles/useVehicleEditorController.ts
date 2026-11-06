import { useCallback, useEffect, useRef, useState, type SetStateAction } from "react";

import { api, type CreateVehicleRequest, type Vehicle } from "../../shared/api";
import { emptyVehicle, optionValue } from "./vehicleViewModel";
import type { MasterDataOptions, ModalMode, ModalTab } from "./vehicleViewModel";
import { vehicleToForm } from "./vehicleTransforms";
import {
  coreSavedDetail, draftValue, reconcileDraft, retainRefreshConflicts, sameDraftValue, type DraftRefreshConflict
} from "./vehicleDraftRefresh";

export type VehicleSavedAction = "vehicle" | "media" | "cv" | "maintenance" | "spareParts" | "decoderFiles" | "documents";

type OpenSections = {
  model: boolean;
  details: boolean;
  vehicle: boolean;
};

type EditorResetReason = "create" | "close";

type UseVehicleEditorControllerOptions = {
  options: MasterDataOptions;
  onMessage: (message: string) => void;
  onReset: (reason: EditorResetReason) => void;
  onDetailLoaded: (detail: Vehicle) => void;
  onDetailRefreshed: (detail: Vehicle, previous: Vehicle) => void;
  onCoreSaved?: (detail: Vehicle, previous: Vehicle, phase: "core" | "complete") => void;
  onFormChange: (form: CreateVehicleRequest) => void;
  onCollectionChanged: () => void;
};

const initialOpenSections: OpenSections = {
  model: true,
  details: false,
  vehicle: false
};

export function useVehicleEditorController({
  options,
  onMessage,
  onReset,
  onDetailLoaded,
  onDetailRefreshed,
  onCoreSaved,
  onFormChange,
  onCollectionChanged
}: UseVehicleEditorControllerOptions) {
  const [form, setFormState] = useState<CreateVehicleRequest>(emptyVehicle);
  const formRef = useRef(form);
  const [saving, setSaving] = useState(false);
  const [selected, setSelectedState] = useState<Vehicle | null>(null);
  const selectedRef = useRef(selected);
  const [mode, setMode] = useState<ModalMode>("create");
  const [modalOpen, setModalOpenState] = useState(false);
  const modalOpenRef = useRef(false);
  const [activeTab, setActiveTab] = useState<ModalTab>("model");
  const [openSections, setOpenSections] = useState<OpenSections>(initialOpenSections);
  const [saveAttempted, setSaveAttempted] = useState(false);
  const [refreshConflicts, setRefreshConflictsState] = useState<DraftRefreshConflict[]>([]);
  const conflictsRef = useRef(refreshConflicts);
  const [lastSavedAction, setLastSavedAction] = useState<VehicleSavedAction | null>(null);
  const requestVersion = useRef(0);
  useEffect(() => () => {
    requestVersion.current++;
    selectedRef.current = null;
    modalOpenRef.current = false;
  }, []);

  const setRefreshConflicts = useCallback((value: DraftRefreshConflict[]) => {
    conflictsRef.current = value;
    setRefreshConflictsState(value);
  }, []);
  // Keep event-time values available even when an API response and typing are batched together.
  const setForm = useCallback((action: SetStateAction<CreateVehicleRequest>) => {
    const next = typeof action === "function" ? action(formRef.current) : action;
    formRef.current = next;
    setFormState(next);
    if (conflictsRef.current.length) setRefreshConflicts(conflictsRef.current.map(conflict => ({
      ...conflict, local: draftValue(next, conflict.field)
    })).filter(conflict => !sameDraftValue(conflict.local, conflict.saved)));
  }, [setRefreshConflicts]);
  const setSelected = useCallback((action: SetStateAction<Vehicle | null>) => {
    selectedRef.current = typeof action === "function" ? action(selectedRef.current) : action;
    setSelectedState(selectedRef.current);
  }, []);
  const setModalOpen = useCallback((action: SetStateAction<boolean>) => {
    modalOpenRef.current = typeof action === "function" ? action(modalOpenRef.current) : action;
    setModalOpenState(modalOpenRef.current);
  }, []);

  const mergeForm = (detail: Vehicle, previous: CreateVehicleRequest) => {
    const incoming = vehicleToForm(detail);
    const merged = reconcileDraft(formRef.current, previous, incoming);
    const conflicts = retainRefreshConflicts(conflictsRef.current, merged.value, incoming, merged.conflicts);
    setForm(merged.value);
    setRefreshConflicts(conflicts);
  };

  const update = (patch: Partial<CreateVehicleRequest>) => {
    setForm((current) => {
      const next = { ...current, ...patch };
      onFormChange(next);
      return next;
    });
  };

  const setSelectedDetail = (detail: Vehicle) => {
    requestVersion.current++;
    setRefreshConflicts([]);
    setSelected(detail);
    setForm(vehicleToForm(detail));
    setSaveAttempted(false);
    onDetailLoaded(detail);
  };

  const applyIndependentDetail = (detail: Vehicle) => {
    if (!modalOpenRef.current) { setSelectedDetail(detail); return; }
    const previous = selectedRef.current;
    if (!previous || previous.id !== detail.id) return;
    requestVersion.current++;
    mergeForm(detail, vehicleToForm(previous));
    setSelected(detail);
    onDetailRefreshed(detail, previous);
  };

  const acknowledgeSavedVehicle = (detail: Vehicle, submitted: CreateVehicleRequest, phase: "core" | "complete") => {
    if (!modalOpenRef.current || (selectedRef.current && selectedRef.current.id !== detail.id)) return;
    requestVersion.current++;
    const previous = selectedRef.current || detail;
    const saved = phase === "core" && previous.id === detail.id ? coreSavedDetail(detail, previous) : detail;
    mergeForm(saved, submitted);
    setSelected(saved);
    setSaveAttempted(false);
    setLastSavedAction("vehicle");
    if (onCoreSaved) onCoreSaved(saved, previous, phase);
    else onDetailRefreshed(saved, previous);
  };

  const resolveRefreshConflicts = (choice: "draft" | "saved") => {
    if (choice === "saved") {
      let next = { ...formRef.current };
      for (const conflict of conflictsRef.current) next = { ...next, [conflict.field]: conflict.saved };
      setForm(next);
    }
    setRefreshConflicts([]);
  };

  const updateCategory = (category: string) => {
    const categoryKey = options.categories.find((entry) => optionValue(entry) === category)?.key;
    const allowed = new Set(
      options.categoryRelations
        .filter((relation) => relation.parentKey === categoryKey)
        .map((relation) => relation.childKey)
    );
    const currentGattung = options.gattungen.find((entry) => optionValue(entry) === form.gattung);

    update({
      category,
      gattung: currentGattung && allowed.has(currentGattung.key) ? form.gattung : ""
    });
  };

  const updateCouplingFront = (couplingFront: string) => {
    update({
      couplingFront,
      couplingRear: form.couplingSame ? couplingFront : form.couplingRear
    });
  };

  const updateCouplingSame = (couplingSame: boolean) => {
    update({
      couplingSame,
      couplingRear: couplingSame ? form.couplingFront : form.couplingRear
    });
  };

	const openCreate = (initialForm: CreateVehicleRequest = emptyVehicle) => {
    requestVersion.current++;
    setRefreshConflicts([]);
    setLastSavedAction(null);
    setSelected(null);
    setMode("create");
		setForm(initialForm);
    setSaveAttempted(false);
    setActiveTab("model");
    setOpenSections(initialOpenSections);
    setModalOpen(true);
    onReset("create");
    onMessage("");
  };

  const closeModal = () => {
    requestVersion.current++;
    setRefreshConflicts([]);
    setLastSavedAction(null);
    setModalOpen(false);
    setSelected(null);
    setMode("create");
    setForm(emptyVehicle);
    setSaveAttempted(false);
    onReset("close");
    onMessage("");
  };

  const openVehicle = (vehicle: Vehicle, nextMode: Exclude<ModalMode, "create">, tab: ModalTab) => {
    const request = ++requestVersion.current;
    api.vehicle(vehicle.id)
      .then((detail) => {
        if (request !== requestVersion.current) return;
        setSelectedDetail(detail);
        setLastSavedAction(null);
        setMode(nextMode);
        setActiveTab(tab);
        setOpenSections(initialOpenSections);
        setModalOpen(true);
        onMessage("");
      })
      .catch((error: Error) => {
        if (request === requestVersion.current) onMessage(error.message);
      });
  };

  const openDetail = (vehicle: Vehicle, tab: ModalTab = "model") => {
    openVehicle(vehicle, "view", tab);
  };

  const openEdit = (vehicle: Vehicle, tab: ModalTab = "model") => {
    openVehicle(vehicle, "edit", tab);
  };

  const refreshSelectedVehicle = (vehicleID = selectedRef.current?.id, action?: VehicleSavedAction, propagateError = false) => {
    if (!vehicleID || selectedRef.current?.id !== vehicleID) return Promise.resolve();
    if (action) setLastSavedAction(action);
    const request = ++requestVersion.current;
    return api
      .vehicle(vehicleID)
      .then((detail) => {
        if (request !== requestVersion.current || selectedRef.current?.id !== detail.id) return;
        applyIndependentDetail(detail);
        onCollectionChanged();
      })
      .catch((error: Error) => {
        if (request === requestVersion.current) onMessage(error.message);
        if (propagateError) throw error;
      });
  };

  const toggleSection = (section: keyof OpenSections) => {
    setOpenSections((current) => ({ ...current, [section]: !current[section] }));
  };

  return {
    state: {
      form,
      saving,
      selected,
      mode,
      modalOpen,
      activeTab,
      openSections,
      saveAttempted,
      refreshConflicts,
      lastSavedAction,
      readonly: mode === "view"
    },
    setters: {
      setForm,
      setSaving,
      setSelected,
      setMode,
      setModalOpen,
      setActiveTab,
      setOpenSections,
      setSaveAttempted
    },
    commands: {
      update,
      setSelectedDetail,
      applyIndependentDetail,
      acknowledgeSavedVehicle,
      resolveRefreshConflicts,
      updateCategory,
      updateCouplingFront,
      updateCouplingSame,
      openCreate,
      closeModal,
      openDetail,
      openEdit,
      toggleSection,
      refreshSelectedVehicle
    }
  };
}
