import { Save, Trash2, X } from "lucide-react";
import { useRef, useState, type FormEvent } from "react";
import { createPortal } from "react-dom";

import type { Language } from "../../shared/i18n";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import type {
  DataTransferArea,
  DataTransferDirection,
  DataTransferFormat,
  DataTransferProfile,
  DataTransferProfileInput
} from "./dataTransferModel";

type TransferProfileDialogProps = {
  availableAreas: DataTransferArea[];
  canDisable: boolean;
  language: Language;
  onClose: () => void;
  onCreate: (input: DataTransferProfileInput) => Promise<unknown>;
  onDisable: (profileId: string) => Promise<unknown>;
  onUpdate: (profileId: string, input: DataTransferProfileInput) => Promise<unknown>;
  profile?: DataTransferProfile;
};

export function TransferProfileDialog({
  availableAreas,
  canDisable,
  language,
  onClose,
  onCreate,
  onDisable,
  onUpdate,
  profile
}: TransferProfileDialogProps) {
  const copy = profileCopy(language);
  const [name, setName] = useState(profile?.name || "");
  const [direction, setDirection] = useState<DataTransferDirection>(profile?.direction || "export");
  const [format, setFormat] = useState<DataTransferFormat>(profile?.format || "railkeeper-json");
  const [areas, setAreas] = useState<DataTransferArea[]>(
    profile ? [...profile.areas] : availableAreas.slice(0, 1)
  );
  const [optionsText, setOptionsText] = useState(JSON.stringify(profile?.options || {}, null, 2));
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const closeRef = useRef<HTMLButtonElement | null>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(() => {
    if (!busy) onClose();
  }, closeRef);
  const csvUnavailable = areas.length !== 1 || areas[0] === "exhibitionLists";
  const exhibitionCSV = areas.includes("exhibitionLists");

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
    let options: Record<string, unknown>;
    try {
      const parsed: unknown = optionsText.trim() ? JSON.parse(optionsText) : {};
      if (!parsed || Array.isArray(parsed) || typeof parsed !== "object") throw new Error();
      options = { ...parsed as Record<string, unknown> };
    } catch {
      setError(copy.invalidOptions);
      return;
    }
    const input: DataTransferProfileInput = {
      name: name.trim(),
      direction,
      format,
      areas: [...areas],
      options
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
    if (!profile || !window.confirm(copy.disableConfirm.replace("{name}", profile.name))) return;
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
            <span>{copy.name}</span>
            <input value={name} onChange={(event) => setName(event.target.value)} autoComplete="off" />
          </label>

          <fieldset className="data-transfer-choice-group">
            <legend>{copy.direction}</legend>
            <label><input type="radio" name="transfer-direction" value="export" checked={direction === "export"}
              onChange={() => setDirection("export")} />{copy.export}</label>
            <label><input type="radio" name="transfer-direction" value="import" checked={direction === "import"}
              onChange={() => setDirection("import")} />{copy.import}</label>
          </fieldset>

          <fieldset className="data-transfer-choice-group transfer-area-choices">
            <legend>{copy.areas}</legend>
            {availableAreas.map((area) => (
              <label key={area}>
                <input type="checkbox" checked={areas.includes(area)} onChange={() => toggleArea(area)} />
                {areaLabel(area, language)}
              </label>
            ))}
          </fieldset>

          <fieldset className="data-transfer-choice-group">
            <legend>{copy.format}</legend>
            <label>
              <input type="radio" name="transfer-format" value="railkeeper-json"
                checked={format === "railkeeper-json"} onChange={() => setFormat("railkeeper-json")} />JSON
            </label>
            <label>
              <input type="radio" name="transfer-format" value="csv" disabled={csvUnavailable}
                checked={format === "csv"} onChange={() => setFormat("csv")} />CSV
            </label>
            {exhibitionCSV ? <p className="data-transfer-choice-reason">{copy.exhibitionCSV}</p> : null}
            {!exhibitionCSV && areas.length > 1
              ? <p className="data-transfer-choice-reason">{copy.multiAreaCSV}</p>
              : null}
          </fieldset>

          <label className="data-transfer-field">
            <span>{copy.options}</span>
            <textarea value={optionsText} onChange={(event) => setOptionsText(event.target.value)} rows={4}
              spellCheck={false} />
            <small>{copy.optionsHelp}</small>
          </label>

          {error ? <p className="form-message error" role="alert">{error}</p> : null}
          <footer className="data-transfer-dialog-actions">
            {profile && canDisable ? (
              <button type="button" className="danger-button" disabled={busy} onClick={() => void disable()}>
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
      </section>
    </div>
  );

  return <><span ref={anchorRef} hidden aria-hidden="true" />{createPortal(dialog, document.body)}</>;
}

function profileCopy(language: Language) {
  return language === "de" ? {
    eyebrow: "TRANSFERPROFIL", createTitle: "Transferprofil anlegen", editTitle: "Transferprofil bearbeiten",
    close: "Dialog schließen", name: "Profilname", direction: "Richtung", export: "Export", import: "Import",
    areas: "Bereiche", format: "Format", options: "Optionen (JSON)",
    optionsHelp: "Optionen werden unverändert im Profil und in neuen Auftragssnapshots gespeichert.",
    exhibitionCSV: "Ausstellungslisten sind nur als JSON verfügbar.",
    multiAreaCSV: "CSV unterstützt genau einen Bereich.", invalidOptions: "Die Optionen müssen ein gültiges JSON-Objekt sein.",
    cancel: "Abbrechen", create: "Profil anlegen", save: "Änderungen speichern", disable: "Profil deaktivieren",
    disableConfirm: "Profil „{name}“ deaktivieren? Bestehende Auftragssnapshots bleiben erhalten.",
    saveError: "Das Profil konnte nicht gespeichert werden.", disableError: "Das Profil konnte nicht deaktiviert werden."
  } : {
    eyebrow: "TRANSFER PROFILE", createTitle: "Create transfer profile", editTitle: "Edit transfer profile",
    close: "Close dialog", name: "Profile name", direction: "Direction", export: "Export", import: "Import",
    areas: "Areas", format: "Format", options: "Options (JSON)",
    optionsHelp: "Options are stored unchanged in the profile and in new job snapshots.",
    exhibitionCSV: "Exhibition lists are only available as JSON.",
    multiAreaCSV: "CSV supports exactly one area.", invalidOptions: "Options must be a valid JSON object.",
    cancel: "Cancel", create: "Create profile", save: "Save changes", disable: "Disable profile",
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
