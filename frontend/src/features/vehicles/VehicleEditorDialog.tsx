import { useLayoutEffect, useRef, type ComponentProps, type FormEvent, type ReactNode } from "react";
import { Save, X } from "lucide-react";

import type { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { VehicleCVTab } from "./VehicleCVTab";
import { VehicleCreateWizard } from "./VehicleCreateWizard";
import type { VehicleSetMemberDraft } from "./vehicleCreateWizardState";
import { VehicleFunctionsTab } from "./VehicleFunctionsTab";
import { VehicleMaintenanceTab } from "./VehicleMaintenanceTab";
import { VehicleModelTab } from "./VehicleModelTab";
import { VehicleReadOnlyView } from "./VehicleReadOnlyView";
import { VehicleSparePartsTab } from "./VehicleSparePartsTab";
import { VehicleSpeedCurveTab } from "./VehicleSpeedCurveTab";
import { VehicleUploadsTab } from "./VehicleUploadsTab";
import { VehicleValuations } from "./VehicleValuations";
import { VehicleMediaConflicts } from "./VehicleMediaConflicts";
import type { MediaDraftConflict } from "./mediaDraftConflicts";
import { VehicleSaveState } from "./VehicleSaveState";
import type { VehicleSavedAction } from "./useVehicleEditorController";
import type { DraftRefreshConflict } from "./vehicleDraftRefresh";
import type { ModalMode, ModalTab } from "./vehicleViewModel";
import type { VehicleCreatePrefill } from "./vehicleSetDuplicate";
import type { ArticleSearchController } from "./useArticleSearchController";

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
  draftOwner: string;
  mode: ModalMode;
  selected: Vehicle | null;
  activeTab: ModalTab;
  saving: boolean;
  message: string;
  messageTone?: "error" | "status";
  mediaDialogs?: ReactNode;
  catalog?: ReactNode;
  refreshConflicts?: DraftRefreshConflict[];
  mediaConflicts?: MediaDraftConflict[];
  onResolveMediaConflict?: (conflict: MediaDraftConflict, choice: "draft" | "saved") => void;
  lastSavedAction?: VehicleSavedAction | null;
  onResolveRefreshConflicts?: (choice: "draft" | "saved") => void;
  tabs: EditorTabs;
  onSubmit: (event: FormEvent<HTMLFormElement>, creationId?: string) => void;
  onSubmitSet: (members: VehicleSetMemberDraft[], imageOwners: Record<string, number>, creationId?: string) => Promise<void>;
  onClose: () => void;
  onTabChange: (tab: ModalTab) => void;
  onEdit: () => void;
  onPrint: () => void;
  onQr: () => void;
  onPreviewImage: ComponentProps<typeof VehicleReadOnlyView>["onPreviewImage"];
  setCreationDisabled?: boolean;
	createPrefill?: VehicleCreatePrefill | null;
  articleSearchController: ArticleSearchController;
};

const editorTabs: Array<{ key: ModalTab; labelKey?: string; label?: string }> = [
  { key: "model", labelKey: "vehicles.tab.model" },
  { key: "control", labelKey: "vehicles.tab.control" },
  { key: "speedCurve", labelKey: "vehicles.tab.speedCurve" },
  { key: "cv", label: "CV" },
  { key: "uploads", labelKey: "vehicles.tab.uploads" },
  { key: "maintenance", labelKey: "vehicles.tab.maintenance" },
  { key: "valuations", labelKey: "valuation.title" },
  { key: "spareParts", labelKey: "vehicles.tab.spareParts" }
];

export function VehicleEditorDialog({
  draftOwner,
  mode,
  selected,
  activeTab,
  saving,
  message,
  messageTone = "error",
  mediaDialogs,
  catalog,
  refreshConflicts = [],
  mediaConflicts = [],
  onResolveMediaConflict = () => undefined,
  lastSavedAction = null,
  onResolveRefreshConflicts = () => undefined,
  tabs,
  onSubmit,
  onSubmitSet,
  onClose,
  onTabChange,
  onEdit,
  onPrint,
  onQr,
  onPreviewImage,
	setCreationDisabled = false,
	createPrefill,
  articleSearchController
}: VehicleEditorDialogProps) {
  const { t, language } = useI18n();
  const bodyRef = useRef<HTMLDivElement>(null);
  const tabNavigationRef = useRef<HTMLElement>(null);
  useLayoutEffect(() => { if (bodyRef.current) bodyRef.current.scrollTop = 0; }, [activeTab]);
  useLayoutEffect(() => {
    const revealActiveTab = () => {
      const navigation = tabNavigationRef.current;
      const active = navigation?.querySelector<HTMLElement>("button.active");
      if (!navigation || !active) return;
      const bounds = navigation.getBoundingClientRect(), target = active.getBoundingClientRect();
      if (target.left < bounds.left) navigation.scrollLeft += target.left - bounds.left;
      else if (target.right > bounds.right) navigation.scrollLeft += target.right - bounds.right;
    };
    revealActiveTab();
    window.addEventListener("resize", revealActiveTab);
    return () => window.removeEventListener("resize", revealActiveTab);
  }, [activeTab, language, mode]);

  return (
    <div className="modal-layer" onKeyDown={(event) => {
      if (event.key === "Escape" && !event.defaultPrevented && mode !== "create" && !saving) {
        event.stopPropagation();
        onClose();
      }
    }} role={mode === "create" ? undefined : "dialog"}
      aria-modal={mode === "create" ? undefined : "true"}
      aria-label={mode === "create" ? undefined : t("vehicles.modal.aria")}>
      {mode === "create" ? (
        <VehicleCreateWizard
          draftOwner={draftOwner}
          model={tabs.model}
          saving={saving}
          message={message}
          onSubmitSingle={onSubmit}
          onSubmitSet={onSubmitSet}
          onClose={onClose}
          setCreationDisabled={setCreationDisabled}
					prefill={createPrefill}
          articleSearchController={articleSearchController}
        />
      ) : (
        <form
          key={`${mode}-${selected?.id || "new"}`}
          className={mode === "view" ? "vehicle-modal vehicle-read-modal-shell" : "vehicle-modal vehicle-edit-modal"}
          onSubmit={event => { if (activeTab === "valuations") event.preventDefault(); else onSubmit(event); }}
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
            disabled={saving}
            title={t("vehicles.close")}
          >
            <X size={18} />
          </button>
        </header>

        {mode === "view" && selected ? (
          <VehicleReadOnlyView
            vehicle={selected}
            catalog={catalog}
            onEdit={onEdit}
            onPrint={onPrint}
            onQr={onQr}
            onPreviewImage={onPreviewImage}
          />
        ) : (
          <>
            <nav ref={tabNavigationRef} className="modal-tabs" aria-label={t("vehicles.modal.aria")}>
              {editorTabs.map((tab) => (
                <button
                  key={tab.key}
                  type="button"
                  className={activeTab === tab.key ? "active" : ""}
                  aria-pressed={activeTab === tab.key}
                  onClick={() => onTabChange(tab.key)}
                >
                  {tab.label || t(tab.labelKey || "")}
                </button>
              ))}
            </nav>

            <div className="modal-body" ref={bodyRef}>
              <VehicleSaveState tab={activeTab} saved={lastSavedAction} conflicts={refreshConflicts}
                onResolve={onResolveRefreshConflicts} />
              <VehicleMediaConflicts vehicle={selected} conflicts={mediaConflicts} saving={saving} onResolve={onResolveMediaConflict} />
              <div className="vehicle-catalog-slot" hidden={activeTab !== "model"}>{catalog}</div>
              {activeTab === "model" && <VehicleModelTab {...tabs.model} />}
              {activeTab === "control" && <VehicleFunctionsTab {...tabs.functions} />}
              {activeTab === "speedCurve" && <VehicleSpeedCurveTab {...tabs.speedCurve} />}
              {activeTab === "cv" && <VehicleCVTab {...tabs.cv} />}
              {activeTab === "uploads" && <VehicleUploadsTab {...tabs.uploads} />}
              {activeTab === "maintenance" && <VehicleMaintenanceTab {...tabs.maintenance} />}
              {activeTab === "spareParts" && <VehicleSparePartsTab {...tabs.spareParts} />}
              {selected && <VehicleValuations key={`${draftOwner}:${selected.id}`} vehicleID={selected.id}
                canEdit journalOwner={draftOwner} active={activeTab === "valuations"} />}
            </div>

            <footer className={`modal-actions${activeTab === "control" ? " function-editor-footer" : ""}`}>
              {message && <p className={`form-message${messageTone === "status" ? " is-status" : ""}`}
                role={messageTone === "status" ? "status" : "alert"}>{message}</p>}
              <button type="button" className="secondary-button" onClick={onClose} disabled={saving}>
                {t("vehicles.cancel")}
              </button>
              {activeTab !== "valuations" && <button className="primary-button" disabled={saving || refreshConflicts.length > 0 || mediaConflicts.some(conflict => conflict.kind === "image")}>
                <Save size={16} aria-hidden="true" />
                {saving
                  ? t("vehicles.saving")
                  : t("vehicles.saveChanges")}
              </button>}
            </footer>
          </>
        )}
        </form>
      )}
      {mediaDialogs}
    </div>
  );
}
