import { useState } from "react";
import { api, type CreateVehicleRequest, type Vehicle } from "../../shared/api";
import { vehicleSetInputFromForm, vehicleSetMembersFromForm } from "./VehicleCreateWizard";
import { clearVehicleCreateDraft, type VehicleSetMemberDraft } from "./vehicleCreateWizardState";
import { vehicleSetDuplicatePrefill, type VehicleCreatePrefill } from "./vehicleSetDuplicate";
import type { PendingArticleImage } from "./vehicleTransforms";
import type { ModalMode, ModalTab } from "./vehicleViewModel";

type VehicleSetControllerOptions = {
  username: string;
  form: CreateVehicleRequest;
  vehicles: Vehicle[];
  openDetail: (vehicle: Vehicle) => void;
  pendingArticleImages: PendingArticleImage[];
  setSaving: (saving: boolean) => void;
  setSelectedDetail: (vehicle: Vehicle) => void;
  setMode: (mode: ModalMode) => void;
  setActiveTab: (tab: ModalTab) => void;
  closeModal: () => void;
  openCreate: (form?: CreateVehicleRequest) => void;
  load: () => void;
  onMessage: (message: string) => void;
  t: (key: string, values?: Record<string, string | number>) => string;
};

export function useVehicleSetController({
  username, form, pendingArticleImages, setSaving, setSelectedDetail, setMode,
  setActiveTab, closeModal, openCreate, openDetail, vehicles, load, onMessage, t
}: VehicleSetControllerOptions) {
	const [selectedVehicleSetID, setSelectedVehicleSetID] = useState("");
	const [vehicleSetMode, setVehicleSetMode] = useState<"view" | "edit" | null>(null);
	const [createPrefill, setCreatePrefill] = useState<VehicleCreatePrefill | null>(null);
  const submitVehicleSet = async (members: VehicleSetMemberDraft[], imageOwners: Record<string, number>) => {
    setSaving(true);
    onMessage("");
    try {
      const created = await api.createVehicleSet({
        set: vehicleSetInputFromForm(form),
        members: vehicleSetMembersFromForm(form, members, pendingArticleImages, imageOwners)
      });
      clearVehicleCreateDraft(username);
      const firstMember = created.members[0];
      if (firstMember) {
        setSelectedDetail(firstMember);
        setMode("edit");
        setActiveTab("model");
      } else {
        closeModal();
      }
      await load();
			setCreatePrefill(null);
      onMessage(t("vehicles.wizard.created", { count: created.members.length }));
    } catch (error) {
      onMessage(error instanceof Error ? error.message : t("vehicles.wizard.createFailed"));
    } finally {
      setSaving(false);
    }
  };
	const closeVehicleSetDialog = () => {
		setSelectedVehicleSetID("");
		setVehicleSetMode(null);
	};
	const openVehicleSet = (setID: string, nextMode: "view" | "edit") => {
		setSelectedVehicleSetID(setID);
		setVehicleSetMode(nextMode);
	};
	const duplicateVehicleSet = (prefill: VehicleCreatePrefill) => {
		closeVehicleSetDialog();
		setCreatePrefill(prefill);
		openCreate(prefill.shared);
	};
	const closeVehicleCreateDialog = () => {
		setCreatePrefill(null);
		closeModal();
	};
	const duplicateVehicleSetByID = async (setID: string) => {
		try {
			const vehicleSet = await api.vehicleSet(setID);
			duplicateVehicleSet(vehicleSetDuplicatePrefill(vehicleSet));
		} catch (error) {
			onMessage(error instanceof Error ? error.message : t("vehicles.set.loadFailed"));
		}
	};
	const handleVehicleSetUpdated = async () => {
		await load();
		setVehicleSetMode("view");
	};
	const handleVehicleSetImageChanged = async () => {
		await load();
	};
  const openSetMember = async (vehicleID: string) => {
					try {
						const vehicle = vehicles.find((entry) => entry.id === vehicleID) || await api.vehicle(vehicleID);
						closeVehicleSetDialog();
						openDetail(vehicle);
					} catch (error) {
						onMessage(error instanceof Error ? error.message : t("vehicles.set.loadFailed"));
					}
				  };

  return {
    selectedVehicleSetID, vehicleSetMode, createPrefill, setCreatePrefill,
    submitVehicleSet, closeVehicleSetDialog, openVehicleSet, duplicateVehicleSet,
    closeVehicleCreateDialog, duplicateVehicleSetByID, handleVehicleSetUpdated, handleVehicleSetImageChanged, openSetMember
  };
}
