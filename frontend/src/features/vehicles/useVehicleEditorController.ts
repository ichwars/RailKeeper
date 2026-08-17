import { useState } from "react";

import { api, type CreateVehicleRequest, type Vehicle } from "../../shared/api";
import { emptyVehicle, optionValue } from "./vehicleViewModel";
import type { MasterDataOptions, ModalMode, ModalTab } from "./vehicleViewModel";
import { vehicleToForm } from "./vehicleTransforms";

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
  onFormChange: (form: CreateVehicleRequest) => void;
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
  onFormChange
}: UseVehicleEditorControllerOptions) {
  const [form, setForm] = useState<CreateVehicleRequest>(emptyVehicle);
  const [saving, setSaving] = useState(false);
  const [selected, setSelected] = useState<Vehicle | null>(null);
  const [mode, setMode] = useState<ModalMode>("create");
  const [modalOpen, setModalOpen] = useState(false);
  const [activeTab, setActiveTab] = useState<ModalTab>("model");
  const [openSections, setOpenSections] = useState<OpenSections>(initialOpenSections);
  const [saveAttempted, setSaveAttempted] = useState(false);

  const update = (patch: Partial<CreateVehicleRequest>) => {
    setForm((current) => {
      const next = { ...current, ...patch };
      onFormChange(next);
      return next;
    });
  };

  const setSelectedDetail = (detail: Vehicle) => {
    setSelected(detail);
    setForm(vehicleToForm(detail));
    setSaveAttempted(false);
    onDetailLoaded(detail);
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
    setModalOpen(false);
    setSelected(null);
    setMode("create");
    setForm(emptyVehicle);
    setSaveAttempted(false);
    onReset("close");
    onMessage("");
  };

  const openVehicle = (vehicle: Vehicle, nextMode: Exclude<ModalMode, "create">, tab: ModalTab) => {
    api.vehicle(vehicle.id)
      .then((detail) => {
        setSelectedDetail(detail);
        setMode(nextMode);
        setActiveTab(tab);
        setOpenSections(initialOpenSections);
        setModalOpen(true);
        onMessage("");
      })
      .catch((error: Error) => onMessage(error.message));
  };

  const openDetail = (vehicle: Vehicle, tab: ModalTab = "model") => {
    openVehicle(vehicle, "view", tab);
  };

  const openEdit = (vehicle: Vehicle, tab: ModalTab = "model") => {
    openVehicle(vehicle, "edit", tab);
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
      updateCategory,
      updateCouplingFront,
      updateCouplingSame,
      openCreate,
      closeModal,
      openDetail,
      openEdit,
      toggleSection
    }
  };
}
