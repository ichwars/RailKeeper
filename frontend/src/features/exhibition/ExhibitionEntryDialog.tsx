import { FormEvent, RefObject, useMemo, useState } from "react";
import { FileText, Image as ImageIcon, ImageOff, Keyboard, Plus, Trash2, X } from "lucide-react";

import {
  ExhibitionEntryInput,
  ExhibitionWorkspace,
  ExhibitionWorkspaceEntry,
  MasterDataEntry,
  Vehicle
} from "../../shared/api";
import { FunctionSymbolPicker, functionSymbolIcon, functionSymbolMetadata } from "../../shared/functionSymbols";
import { AppFilePicker } from "../../shared/ui/AppFilePicker";
import { AppSelect } from "../../shared/ui/AppSelect";
import { adapterOptions } from "../vehicles/vehicleOptions";

type EntryTab = "general" | "image" | "functions";

export type ExhibitionEntryMasterData = {
  manufacturers: MasterDataEntry[];
  gattungen: MasterDataEntry[];
  epochs: MasterDataEntry[];
  railwayCompanies: MasterDataEntry[];
};

type ExhibitionFunction = {
  key: string;
  name: string;
  type: string;
  symbolKey?: string;
};

type ExhibitionEntryDialogProps = {
  admin: boolean;
  closeRef: RefObject<HTMLButtonElement | null>;
  de: boolean;
  editingEntry: ExhibitionWorkspaceEntry | null;
  form: ExhibitionEntryInput;
  masterData: ExhibitionEntryMasterData;
  saving: boolean;
  symbols: MasterDataEntry[];
  vehicles: Vehicle[];
  workspace: ExhibitionWorkspace;
  canReadInventory: boolean;
  onChange: (patch: Partial<ExhibitionEntryInput>) => void;
  onClose: () => void;
  onDelete: (entry: ExhibitionWorkspaceEntry) => void;
  onSelectVehicle: (vehicleID: string) => void;
  onSubmit: (event: FormEvent) => void;
};

const functionKeys = Array.from({ length: 32 }, (_, index) => `F${index}`);
const functionTypes = ["standard", "licht", "sound", "kupplung", "rauch", "sonderfunktion"];

function defaultFunction(key: string): ExhibitionFunction {
  return {
    key,
    name: key === "F0" ? "Fahrlicht" : "",
    type: key === "F0" ? "licht" : "standard",
    symbolKey: key === "F0" ? "light" : ""
  };
}

function emptyFunctions() {
  return functionKeys.map(defaultFunction);
}

function isConfiguredFunction(item: ExhibitionFunction) {
  const fallback = defaultFunction(item.key);
  return Boolean(item.name.trim() || item.symbolKey || item.type !== fallback.type);
}

function parseFunctions(value?: string) {
  if (!value) return emptyFunctions();
  try {
    const parsed = JSON.parse(value) as ExhibitionFunction[];
    if (Array.isArray(parsed)) {
      const byKey = new Map(parsed.map((item) => [item.key, item]));
      return emptyFunctions().map((item) => ({ ...item, ...(byKey.get(item.key) || {}) }));
    }
  } catch {
    const byKey = new Map<string, ExhibitionFunction>();
    for (const part of value.split(/[,;\n]/)) {
      const match = part.trim().match(/^(F\d{1,2})\s*[:=-]?\s*(.*)$/i);
      if (match) {
        const key = match[1].toUpperCase();
        byKey.set(key, {
          key,
          name: match[2].trim() || (key === "F0" ? "Fahrlicht" : ""),
          type: key === "F0" ? "licht" : "standard"
        });
      }
    }
    return emptyFunctions().map((item) => ({ ...item, ...(byKey.get(item.key) || {}) }));
  }
  return emptyFunctions();
}

function serializeFunctions(functions: ExhibitionFunction[]) {
  return JSON.stringify(functions.filter(isConfiguredFunction).map((item) => ({
    key: item.key,
    name: item.name.trim(),
    type: item.type,
    symbolKey: item.symbolKey || ""
  })));
}

function fileToDataURL(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result || ""));
    reader.onerror = () => reject(reader.error || new Error("Bild konnte nicht gelesen werden."));
    reader.readAsDataURL(file);
  });
}

function optionLabels(items: MasterDataEntry[], current?: string) {
  const labels = items.map((item) => item.label);
  if (current && !labels.includes(current)) labels.unshift(current);
  return labels;
}

function functionTypeLabel(type: string, de: boolean) {
  if (!de) return type.charAt(0).toUpperCase() + type.slice(1);
  return ({
    standard: "Standard",
    licht: "Licht",
    sound: "Sound",
    kupplung: "Kupplung",
    rauch: "Rauch",
    sonderfunktion: "Sonderfunktion"
  } as Record<string, string>)[type] || type;
}

export function ExhibitionEntryDialog({
  admin,
  closeRef,
  de,
  editingEntry,
  form,
  masterData,
  saving,
  symbols,
  vehicles,
  workspace,
  canReadInventory,
  onChange,
  onClose,
  onDelete,
  onSelectVehicle,
  onSubmit
}: ExhibitionEntryDialogProps) {
  const [activeTab, setActiveTab] = useState<EntryTab>("general");
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [functions, setFunctions] = useState<ExhibitionFunction[]>(() => parseFunctions(form.functionKeys));
  const [addedKeys, setAddedKeys] = useState<string[]>([]);

  const configuredFunctions = useMemo(() => functions.filter(isConfiguredFunction), [functions]);
  const visibleFunctions = useMemo(
    () => functions.filter((item) => isConfiguredFunction(item) || addedKeys.includes(item.key)),
    [addedKeys, functions]
  );

  const updateFunctions = (next: ExhibitionFunction[]) => {
    setFunctions(next);
    onChange({ functionKeys: serializeFunctions(next) });
  };

  const updateFunction = (key: string, patch: Partial<ExhibitionFunction>) => {
    updateFunctions(functions.map((item) => item.key === key ? { ...item, ...patch } : item));
  };

  const addFunction = () => {
    const next = functions.find((item) => !isConfiguredFunction(item) && !addedKeys.includes(item.key));
    if (!next) return;
    setAddedKeys((current) => [...current, next.key]);
  };

  const removeFunction = (key: string) => {
    const fallback = defaultFunction(key);
    updateFunctions(functions.map((item) => item.key === key ? {
      key,
      name: "",
      type: fallback.type,
      symbolKey: ""
    } : item));
    setAddedKeys((current) => current.filter((item) => item !== key));
  };

  const chooseImage = async (file: File | null) => {
    setSelectedFile(file);
    if (!file) return;
    onChange({ imageUrl: await fileToDataURL(file) });
  };

  const removeImage = () => {
    setSelectedFile(null);
    onChange({ imageUrl: "" });
  };

  const tabs: Array<{ id: EntryTab; label: string; count?: number; icon: typeof FileText }> = [
    { id: "general", label: de ? "Allgemein" : "General", icon: FileText },
    { id: "image", label: de ? "Fahrzeugbild" : "Vehicle image", count: form.imageUrl ? 1 : 0, icon: ImageIcon },
    { id: "functions", label: de ? "Funktionstasten" : "Function keys", count: configuredFunctions.length, icon: Keyboard }
  ];

  return (
    <form
      className="vehicle-modal exhibition-entry-workspace-modal"
      onSubmit={onSubmit}
      role="dialog"
      aria-modal="true"
      aria-labelledby="exhibition-entry-dialog-title"
    >
      <header>
        <div>
          <h2 id="exhibition-entry-dialog-title">{editingEntry ? (de ? "Eintrag bearbeiten" : "Edit entry") : (de ? "Eintrag hinzufügen" : "Add entry")}</h2>
          <p>{de ? "Veranstaltungsdaten ändern keine Fahrzeug-Stammdaten." : "Event data does not change vehicle master data."}</p>
        </div>
        <button ref={closeRef} type="button" className="icon-button" onClick={onClose} aria-label={de ? "Schließen" : "Close"}>
          <X size={19} />
        </button>
      </header>

      <div className="exhibition-entry-workspace-tabs" role="tablist" aria-label={de ? "Eintragsbereiche" : "Entry sections"}>
        {tabs.map((tab) => {
          const Icon = tab.icon;
          return (
            <button
              key={tab.id}
              type="button"
              role="tab"
              aria-selected={activeTab === tab.id}
              aria-controls={`exhibition-entry-panel-${tab.id}`}
              className={activeTab === tab.id ? "active" : ""}
              onClick={() => setActiveTab(tab.id)}
            >
              <Icon size={17} aria-hidden="true" />
              <span>{tab.label}</span>
              {typeof tab.count === "number" && <small>{tab.count}</small>}
            </button>
          );
        })}
      </div>

      <div className="exhibition-entry-workspace-body">
        {activeTab === "general" && (
          <section id="exhibition-entry-panel-general" role="tabpanel" className="exhibition-entry-general-tab">
            <div className="exhibition-entry-workspace-form">
              {canReadInventory && (
                <label className="wide">
                  <span>{de ? "Bestandsfahrzeug (optional)" : "Inventory vehicle (optional)"}</span>
                  <AppSelect value={form.vehicleId || ""} onChange={(event) => onSelectVehicle(event.target.value)}>
                    <option value="">{de ? "Gastfahrzeug / manuelle Eingabe" : "Guest vehicle / manual entry"}</option>
                    {vehicles.map((vehicle) => <option key={vehicle.id} value={vehicle.id}>{vehicle.inventoryNumber} · {vehicle.name}</option>)}
                  </AppSelect>
                </label>
              )}
              <label><span>{de ? "Besitzer" : "Owner"}</span><input value={form.owner} onChange={(event) => onChange({ owner: event.target.value })} /></label>
              <label><span>{de ? "Lok-Bezeichnung" : "Locomotive name"}</span><input required value={form.locomotiveName} onChange={(event) => onChange({ locomotiveName: event.target.value })} /></label>
              <label><span>{de ? "Hersteller" : "Manufacturer"}</span><AppSelect value={form.manufacturer || ""} onChange={(event) => onChange({ manufacturer: event.target.value })}><option value="">{de ? "Keine Auswahl" : "No selection"}</option>{optionLabels(masterData.manufacturers, form.manufacturer).map((label) => <option key={label} value={label}>{label}</option>)}</AppSelect></label>
              <label><span>{de ? "Baureihe" : "Series"}</span><input value={form.series || ""} onChange={(event) => onChange({ series: event.target.value })} /></label>
              <label><span>{de ? "Gattung" : "Category"}</span><AppSelect value={form.gattung || ""} onChange={(event) => onChange({ gattung: event.target.value })}><option value="">{de ? "Keine Auswahl" : "No selection"}</option>{optionLabels(masterData.gattungen, form.gattung).map((label) => <option key={label} value={label}>{label}</option>)}</AppSelect></label>
              <label><span>{de ? "Epoche" : "Epoch"}</span><AppSelect value={form.epoch || ""} onChange={(event) => onChange({ epoch: event.target.value })}><option value="">{de ? "Keine Auswahl" : "No selection"}</option>{optionLabels(masterData.epochs, form.epoch).map((label) => <option key={label} value={label}>{label}</option>)}</AppSelect></label>
              <label><span>{de ? "Bahnverwaltung" : "Railway company"}</span><AppSelect value={form.railwayCompany || ""} onChange={(event) => onChange({ railwayCompany: event.target.value })}><option value="">{de ? "Keine Auswahl" : "No selection"}</option>{optionLabels(masterData.railwayCompanies, form.railwayCompany).map((label) => <option key={label} value={label}>{label}</option>)}</AppSelect></label>
              <label><span>{de ? "Verfügbarkeit" : "Availability"}</span><AppSelect value={form.availability || "available"} onChange={(event) => onChange({ availability: event.target.value as "available" | "unavailable" })}><option value="available">{de ? "Verfügbar" : "Available"}</option><option value="unavailable">{de ? "Nicht verfügbar" : "Unavailable"}</option></AppSelect></label>
            </div>

            <div className="exhibition-entry-control-grid">
              <fieldset>
                <legend>{de ? "Steuerung" : "Control"}</legend>
                <div className="exhibition-entry-workspace-form">
                  <label><span>{de ? "Decoder-Typ" : "Decoder type"}</span><input value={form.decoderType || ""} onChange={(event) => onChange({ decoderType: event.target.value })} /></label>
                  <label><span>{de ? "Adapter / Schnittstelle" : "Adapter / interface"}</span><AppSelect value={form.adapter || ""} onChange={(event) => onChange({ adapter: event.target.value })}><option value="">{de ? "Keine Auswahl" : "No selection"}</option>{adapterOptions.map((option) => <option key={option} value={option}>{option}</option>)}</AppSelect></label>
                  <label><span>{de ? "Adresse DCC" : "DCC address"}</span><input value={form.decoderNumber || ""} onChange={(event) => onChange({ decoderNumber: event.target.value })} /></label>
                  <label><span>{de ? "Adresse SX" : "SX address"}</span><input value={form.sxAddress || ""} onChange={(event) => onChange({ sxAddress: event.target.value })} /></label>
                  <label><span>{de ? "Zentrale / Schnittstelle" : "Command station / interface"}</span><input value={form.interfaceName || ""} onChange={(event) => onChange({ interfaceName: event.target.value })} /></label>
                  <label className="switch-label exhibition-entry-analog"><span>{de ? "Analog" : "Analogue"}</span><span className="switch-field"><input type="checkbox" checked={Boolean(form.analog)} onChange={(event) => onChange({ analog: event.target.checked })} /><span /></span></label>
                </div>
              </fieldset>

              <label className="exhibition-entry-notes"><span>{de ? "Nr. / Beschriftung / Merkmale" : "No. / markings / features"}</span><textarea rows={6} value={form.notes || ""} onChange={(event) => onChange({ notes: event.target.value })} /></label>
            </div>

            <fieldset className="exhibition-form-days"><legend>{de ? "Fahrtage" : "Operating days"}</legend><button type="button" className={form.dayScope === "all" ? "active" : ""} onClick={() => onChange({ dayScope: "all" })}>{de ? "Alle Tage" : "All days"}</button>{workspace.dayScopes.map((scope, index) => {
              const selected = form.dayScope === "all" || String(form.dayScope || "").split(",").includes(scope);
              return <button type="button" key={scope} className={selected ? "active" : ""} onClick={() => {
                const current = form.dayScope === "all" ? [] : String(form.dayScope || "").split(",").filter(Boolean);
                const next = current.includes(scope) ? current.filter((item) => item !== scope) : [...current, scope];
                onChange({ dayScope: next.length ? next.join(",") : "all" });
              }}>{de ? `Tag ${index + 1}` : `Day ${index + 1}`}</button>;
            })}</fieldset>
          </section>
        )}

        {activeTab === "image" && (
          <section id="exhibition-entry-panel-image" role="tabpanel" className="exhibition-entry-image-tab">
            <div className="exhibition-entry-image-preview">
              {form.imageUrl ? <img src={form.imageUrl} alt="" /> : <div><ImageOff size={30} aria-hidden="true" /><span>{de ? "Kein Fahrzeugbild" : "No vehicle image"}</span></div>}
            </div>
            <div className="exhibition-entry-image-upload">
              <h3>{de ? "Fahrzeugbild" : "Vehicle image"}</h3>
              <p>{de ? "PNG, JPG oder WebP · max. 10 MB" : "PNG, JPG or WebP · max. 10 MB"}</p>
              <AppFilePicker
                label={de ? "Bilddatei" : "Image file"}
                accept="image/png,image/jpeg,image/webp"
                file={selectedFile}
                onFileChange={(file) => void chooseImage(file)}
                triggerLabel={de ? "Bild auswählen" : "Choose image"}
                clearLabel={de ? "Auswahl entfernen" : "Clear selection"}
                emptyLabel={form.imageUrl ? (de ? "Gespeichertes Bild" : "Saved image") : (de ? "Kein Bild ausgewählt" : "No image selected")}
              />
              {form.imageUrl && <button type="button" className="secondary-button danger" onClick={removeImage}><Trash2 size={16} />{de ? "Bild entfernen" : "Remove image"}</button>}
            </div>
          </section>
        )}

        {activeTab === "functions" && (
          <section id="exhibition-entry-panel-functions" role="tabpanel" className="exhibition-entry-functions-tab">
            <div className="exhibition-entry-function-head">
              <h3>{de ? "Funktionstasten" : "Function keys"}</h3>
              <span><strong>{configuredFunctions.length}</strong> {de ? "belegt" : "assigned"} · <strong>{configuredFunctions.filter((item) => item.type === "sound").length}</strong> Sound · <strong>{configuredFunctions.filter((item) => item.type === "licht").length}</strong> {de ? "Licht" : "Light"}</span>
            </div>
            <div className="exhibition-entry-function-table">
              <div className="exhibition-entry-function-columns" aria-hidden="true"><span>{de ? "Taste" : "Key"}</span><span>{de ? "Beschreibung" : "Description"}</span><span>{de ? "Typ" : "Type"}</span><span>{de ? "Symbol" : "Symbol"}</span><span /></div>
              {visibleFunctions.map((item) => <div key={item.key} className="exhibition-entry-function-row">
                <strong>{functionSymbolIcon(item.symbolKey, item.type, functionSymbolMetadata(symbols, item.symbolKey))}<span>{item.key}</span></strong>
                <input value={item.name} onChange={(event) => updateFunction(item.key, { name: event.target.value })} placeholder={de ? "Beschreibung" : "Description"} aria-label={`${item.key} ${de ? "Beschreibung" : "description"}`} />
                <AppSelect value={item.type} onChange={(event) => updateFunction(item.key, { type: event.target.value })} aria-label={`${item.key} ${de ? "Typ" : "type"}`}>{functionTypes.map((type) => <option key={type} value={type}>{functionTypeLabel(type, de)}</option>)}</AppSelect>
                <FunctionSymbolPicker value={item.symbolKey || ""} functionType={item.type} symbols={symbols} label={`${item.key} ${de ? "Symbol" : "symbol"}`} onChange={(symbolKey, symbolLabel) => updateFunction(item.key, { symbolKey, name: item.name || symbolLabel })} />
                <button type="button" className="icon-button" onClick={() => removeFunction(item.key)} aria-label={`${item.key} ${de ? "entfernen" : "remove"}`} title={`${item.key} ${de ? "entfernen" : "remove"}`}><Trash2 size={16} /></button>
              </div>)}
            </div>
            <button type="button" className="secondary-button exhibition-add-function" onClick={addFunction} disabled={visibleFunctions.length >= functionKeys.length}><Plus size={17} />{de ? "Funktion hinzufügen" : "Add function"}</button>
          </section>
        )}
      </div>

      <footer>
        {editingEntry && admin && <button type="button" className="secondary-button danger" onClick={() => onDelete(editingEntry)}><Trash2 size={16} />{de ? "Löschen" : "Delete"}</button>}
        <span />
        <button type="button" className="secondary-button" onClick={onClose}>{de ? "Abbrechen" : "Cancel"}</button>
        <button type="submit" className="primary-button" disabled={saving}>{de ? "Speichern" : "Save"}</button>
      </footer>
    </form>
  );
}
