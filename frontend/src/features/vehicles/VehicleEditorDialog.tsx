import type { ComponentProps, FormEventHandler } from "react";
import { X } from "lucide-react";

import type { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { VehicleCVTab } from "./VehicleCVTab";
import { VehicleCreateWizard, type VehicleSetMemberDraft } from "./VehicleCreateWizard";
import { VehicleFunctionsTab } from "./VehicleFunctionsTab";
import { VehicleMaintenanceTab } from "./VehicleMaintenanceTab";
import { VehicleModelTab } from "./VehicleModelTab";
import { VehicleReadOnlyView } from "./VehicleReadOnlyView";
import { VehicleSparePartsTab } from "./VehicleSparePartsTab";
import { VehicleSpeedCurveTab } from "./VehicleSpeedCurveTab";
import { VehicleUploadsTab } from "./VehicleUploadsTab";
import type { ModalMode, ModalTab } from "./vehicleViewModel";

type EditorTabs = {
  model: ComponentProps<typeof VehicleModelTab>;
  functions: ComponentProps<typeof VehicleFunctionsTab>;
  speedCurve: ComponentProps<typeof VehicleSpeedCurveTab>;
  cv: ComponentProps<typeof VehicleCVTab>;
  uploads: ComponentProps<typeof VehicleUploadsTab>;
  maintenance: ComponentProps<typeof VehicleMaintenanceTab>;
  spareParts: ComponentProps<typeof VehicleSparePartsTab>;
};

type VehicleEditorDialogProps = {
  mode: ModalMode;
  selected: Vehicle | null;
  activeTab: ModalTab;
  saving: boolean;
  message: string;
  tabs: EditorTabs;
  onSubmit: FormEventHandler<HTMLFormElement>;
  onSubmitSet: (members: VehicleSetMemberDraft[]) => Promise<void>;
  onClose: () => void;
  onTabChange: (tab: ModalTab) => void;
  onEdit: () => void;
  onPrint: () => void;
  onQr: () => void;
  onPreviewImage: ComponentProps<typeof VehicleReadOnlyView>["onPreviewImage"];
  setCreationDisabled?: boolean;
};

const editorTabs: Array<{ key: ModalTab; labelKey?: string; label?: string }> = [
  { key: "model", labelKey: "vehicles.tab.model" },
  { key: "control", labelKey: "vehicles.tab.control" },
  { key: "speedCurve", labelKey: "vehicles.tab.speedCurve" },
  { key: "cv", label: "CV" },
  { key: "uploads", labelKey: "vehicles.tab.uploads" },
  { key: "maintenance", labelKey: "vehicles.tab.maintenance" },
  { key: "spareParts", labelKey: "vehicles.tab.spareParts" }
];

export function VehicleEditorDialog({
  mode,
  selected,
  activeTab,
  saving,
  message,
  tabs,
  onSubmit,
  onSubmitSet,
  onClose,
  onTabChange,
  onEdit,
  onPrint,
  onQr,
  onPreviewImage,
  setCreationDisabled = false
}: VehicleEditorDialogProps) {
  const { t } = useI18n();

  return (
    <div className="modal-layer" role="dialog" aria-modal="true" aria-label={t("vehicles.modal.aria")}>
      {mode === "create" ? (
        <VehicleCreateWizard
          model={tabs.model}
          saving={saving}
          message={message}
          onSubmitSingle={onSubmit}
          onSubmitSet={onSubmitSet}
          onClose={onClose}
          setCreationDisabled={setCreationDisabled}
        />
      ) : (
        <form
          key={`${mode}-${selected?.id || "new"}`}
          className={mode === "view" ? "vehicle-modal vehicle-read-modal-shell" : "vehicle-modal"}
          onSubmit={onSubmit}
        >
        <header className="modal-head">
          <h2>
            {mode === "edit"
                ? t("vehicles.modal.edit")
                : t("vehicles.modal.view")}
          </h2>
          <button
            type="button"
            className="icon-button"
            onClick={onClose}
            aria-label={t("vehicles.close")}
            title={t("vehicles.close")}
          >
            <X size={18} />
          </button>
        </header>

        {mode === "view" && selected ? (
          <VehicleReadOnlyView
            vehicle={selected}
            onEdit={onEdit}
            onPrint={onPrint}
            onQr={onQr}
            onPreviewImage={onPreviewImage}
          />
        ) : (
          <>
            <nav className="modal-tabs" aria-label={t("vehicles.modal.aria")}>
              {editorTabs.map((tab) => (
                <button
                  key={tab.key}
                  type="button"
                  className={activeTab === tab.key ? "active" : ""}
                  onClick={() => onTabChange(tab.key)}
                >
                  {tab.label || t(tab.labelKey || "")}
                </button>
              ))}
            </nav>

            <div className="modal-body">
              {activeTab === "model" && <VehicleModelTab {...tabs.model} />}
              {activeTab === "control" && <VehicleFunctionsTab {...tabs.functions} />}
              {activeTab === "speedCurve" && <VehicleSpeedCurveTab {...tabs.speedCurve} />}
              {activeTab === "cv" && <VehicleCVTab {...tabs.cv} />}
              {activeTab === "uploads" && <VehicleUploadsTab {...tabs.uploads} />}
              {activeTab === "maintenance" && <VehicleMaintenanceTab {...tabs.maintenance} />}
              {activeTab === "spareParts" && <VehicleSparePartsTab {...tabs.spareParts} />}
            </div>

            <footer className="modal-actions">
              {message && <p className="form-message">{message}</p>}
              <button type="button" className="secondary-button" onClick={onClose}>
                {t("vehicles.cancel")}
              </button>
              <button className="primary-button" disabled={saving}>
                {saving
                  ? t("vehicles.saving")
                  : t("vehicles.saveChanges")}
              </button>
            </footer>
          </>
        )}
        </form>
      )}
    </div>
  );
}
