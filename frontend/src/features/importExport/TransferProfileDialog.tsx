import { Save, Trash2, X } from "lucide-react";
import { useEffect, useRef, useState, type FormEvent } from "react";
import { createPortal } from "react-dom";

import type { Language } from "../../shared/i18n";
import { AppCheckbox } from "../../shared/ui/AppCheckbox";
import { AppSelect } from "../../shared/ui/AppSelect";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import type {
  DataTransferArea,
  DataTransferDirection,
  DataTransferFormat,
  DataTransferProfile,
  DataTransferProfileInput
} from "./dataTransferModel";
import { TransferConfirmDialog, type TransferPendingAction } from "./TransferConfirmDialog";

type TransferProfileDialogProps = {
  availableAreas: DataTransferArea[];
  canDisable: boolean;
  language: Language;
  onClose: () => void;
  onCreate: (input: DataTransferProfileInput) => Promise<unknown>;
  onDisable: (profileId: string) => Promise<unknown>;
  onUpdate: (profileId: string, input: DataTransferProfileInput) => Promise<unknown>;
  initialProfileId?: string;
  profiles: DataTransferProfile[];
};

export function TransferProfileDialog({
  availableAreas,
  canDisable,
  language,
  onClose,
  onCreate,
  onDisable,
  onUpdate,
  initialProfileId,
  profiles
}: TransferProfileDialogProps) {
  const copy = profileCopy(language);
  const [selectedId, setSelectedId] = useState(initialProfileId || "");
  const profile = profiles.find((item) => item.id === selectedId);
  const [name, setName] = useState(profile?.name || "");
  const [direction, setDirection] = useState<DataTransferDirection>(profile?.direction || "export");
  const [format, setFormat] = useState<DataTransferFormat>(profile?.format || "railkeeper-json");
  const [areas, setAreas] = useState<DataTransferArea[]>(
    profile ? [...profile.areas] : availableAreas.slice(0, 1)
  );
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [disableConfirmationOpen, setDisableConfirmationOpen] = useState(false);
  const closeRef = useRef<HTMLButtonElement | null>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(() => {
    if (!busy) onClose();
  }, closeRef);
  const csvUnavailable = areas.length !== 1 || areas[0] === "exhibitionLists";
  const exhibitionCSV = areas.includes("exhibitionLists");

  useEffect(() => {
    setName(profile?.name || "");
    setDirection(profile?.direction || "export");
    setFormat(profile?.format || "railkeeper-json");
    setAreas(profile ? [...profile.areas] : availableAreas.slice(0, 1));
    setError("");
  }, [availableAreas, profile]);

  function toggleArea(area: DataTransferArea) {
    setAreas((current) => current.includes(area)
      ? current.filter((selected) => selected !== area)
      : [...current, area]
    );
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    setError("");
    if (!name.trim() || areas.length === 0 || (format === "csv" && csvUnavailable)) return;
    const input: DataTransferProfileInput = {
      name: name.trim(),
      direction,
      format,
      areas: [...areas],
      options: { ...(profile?.options ?? {}) }
    };
    setBusy(true);
    try {
      if (profile) await onUpdate(profile.id, input);
      else await onCreate(input);
      onClose();
    } catch (reason) {
      setError(errorMessage(reason, copy.saveError));
    } finally {
      setBusy(false);
    }
  }

  async function disable() {
    if (!profile) return;
    setBusy(true);
    setError("");
    try {
      await onDisable(profile.id);
      onClose();
    } catch (reason) {
      setError(errorMessage(reason, copy.disableError));
    } finally {
      setBusy(false);
    }
  }

  const title = profile ? copy.editTitle : copy.createTitle;
  const disableAction: TransferPendingAction | null = profile && disableConfirmationOpen ? {
    title: copy.disableTitle,
    body: copy.disableConfirm.replace("{name}", profile.name),
    confirmLabel: copy.disable,
    dangerous: true,
    errorMessage: copy.disableError,
    run: disable
  } : null;
  const dialog = (
    <div
      ref={layerRef}
      className="confirm-layer data-transfer-dialog-layer"
      role="dialog"
      aria-modal="true"
      aria-label={title}
      onKeyDown={onKeyDown}
    >
      <section className="panel data-transfer-dialog transfer-profile-dialog">
        <header className="data-transfer-dialog-head">
          <div><p className="eyebrow">{copy.eyebrow}</p><h2>{title}</h2></div>
          <button ref={closeRef} type="button" className="icon-button" aria-label={copy.close} disabled={busy}
            onClick={onClose}>
            <X size={19} aria-hidden="true" />
          </button>
        </header>
        <form className="data-transfer-dialog-body" onSubmit={submit}>
          <label className="data-transfer-field">
            <span>{copy.manage}</span>
            <AppSelect aria-label={copy.manage} value={selectedId} disabled={busy}
              onChange={(event) => setSelectedId(event.target.value)}>
              <option value="">{copy.newProfile}</option>
              {profiles.map((item) => (
                <option key={item.id} value={item.id}>
                  {`${item.direction === "import" ? copy.import : copy.export}: ${item.name} (${item.enabled ? copy.enabled : copy.disabled})`}
                </option>
              ))}
            </AppSelect>
          </label>
          <label className="data-transfer-field">
            <span>{copy.name}</span>
            <input value={name} onChange={(event) => setName(event.target.value)} autoComplete="off" />
          </label>

          <fieldset className="data-transfer-choice-group data-transfer-radio-group">
            <legend>{copy.direction}</legend>
            <div role="radiogroup" aria-label={copy.direction}>
              <button type="button" role="radio" aria-checked={direction === "export"}
                className={direction === "export" ? "selected" : ""} onClick={() => setDirection("export")}>
                {copy.export}
              </button>
              <button type="button" role="radio" aria-checked={direction === "import"}
                className={direction === "import" ? "selected" : ""} onClick={() => setDirection("import")}>
                {copy.import}
              </button>
            </div>
          </fieldset>

          <fieldset className="data-transfer-choice-group transfer-area-choices">
            <legend>{copy.areas}</legend>
            {availableAreas.map((area) => (
              <AppCheckbox key={area} label={areaLabel(area, language)} checked={areas.includes(area)}
                onChange={() => toggleArea(area)} />
            ))}
          </fieldset>

          <fieldset className="data-transfer-choice-group data-transfer-radio-group">
            <legend>{copy.format}</legend>
            <div role="radiogroup" aria-label={copy.format}>
              <button type="button" role="radio" aria-checked={format === "railkeeper-json"}
                className={format === "railkeeper-json" ? "selected" : ""}
                onClick={() => setFormat("railkeeper-json")}>JSON</button>
              <button type="button" role="radio" aria-checked={format === "csv"} disabled={csvUnavailable}
                className={format === "csv" ? "selected" : ""} onClick={() => setFormat("csv")}>CSV</button>
            </div>
            {exhibitionCSV ? <p className="data-transfer-choice-reason">{copy.exhibitionCSV}</p> : null}
            {!exhibitionCSV && areas.length > 1
              ? <p className="data-transfer-choice-reason">{copy.multiAreaCSV}</p>
              : null}
          </fieldset>

          {error ? <p className="form-message error" role="alert">{error}</p> : null}
          <footer className="data-transfer-dialog-actions">
            {profile && canDisable ? (
              <button type="button" className="danger-button" disabled={busy}
                onClick={() => setDisableConfirmationOpen(true)}>
                <Trash2 size={16} aria-hidden="true" />{copy.disable}
              </button>
            ) : <span />}
            <span>
              <button type="button" className="secondary-button" disabled={busy} onClick={onClose}>{copy.cancel}</button>
              <button type="submit" className="primary-button"
                disabled={busy || !name.trim() || areas.length === 0 || (format === "csv" && csvUnavailable)}>
                <Save size={16} aria-hidden="true" />{profile ? copy.save : copy.create}
              </button>
            </span>
          </footer>
        </form>
        <TransferConfirmDialog action={disableAction} cancelLabel={copy.cancel}
          onClose={() => setDisableConfirmationOpen(false)} />
      </section>
    </div>
  );

  return <><span ref={anchorRef} hidden aria-hidden="true" />{createPortal(dialog, document.body)}</>;
}

function profileCopy(language: Language) {
  return language === "de" ? {
    eyebrow: "TRANSFERPROFIL", createTitle: "Transferprofil anlegen", editTitle: "Transferprofil bearbeiten",
    close: "Dialog schließen", name: "Profilname", direction: "Richtung", export: "Export", import: "Import",
    manage: "Profil verwalten", newProfile: "Neues Profil", enabled: "aktiv", disabled: "deaktiviert",
    areas: "Bereiche", format: "Format",
    exhibitionCSV: "Ausstellungslisten sind nur als JSON verfügbar.",
    multiAreaCSV: "CSV unterstützt genau einen Bereich.",
    cancel: "Abbrechen", create: "Profil anlegen", save: "Änderungen speichern", disable: "Profil deaktivieren",
    disableTitle: "Profil deaktivieren?",
    disableConfirm: "Profil „{name}“ deaktivieren? Bestehende Auftragssnapshots bleiben erhalten.",
    saveError: "Das Profil konnte nicht gespeichert werden.", disableError: "Das Profil konnte nicht deaktiviert werden."
  } : {
    eyebrow: "TRANSFER PROFILE", createTitle: "Create transfer profile", editTitle: "Edit transfer profile",
    close: "Close dialog", name: "Profile name", direction: "Direction", export: "Export", import: "Import",
    manage: "Manage profile", newProfile: "New profile", enabled: "enabled", disabled: "disabled",
    areas: "Areas", format: "Format",
    exhibitionCSV: "Exhibition lists are only available as JSON.",
    multiAreaCSV: "CSV supports exactly one area.",
    cancel: "Cancel", create: "Create profile", save: "Save changes", disable: "Disable profile",
    disableTitle: "Disable profile?",
    disableConfirm: "Disable profile “{name}”? Existing job snapshots remain unchanged.",
    saveError: "The profile could not be saved.", disableError: "The profile could not be disabled."
  };
}

function areaLabel(area: DataTransferArea, language: Language) {
  const labels = language === "de"
    ? { vehicles: "Fahrzeuge", accessories: "Zubehör", exhibitionLists: "Ausstellungslisten" }
    : { vehicles: "Vehicles", accessories: "Accessories", exhibitionLists: "Exhibition lists" };
  return labels[area];
}

function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}
