import { FormEvent, ReactNode, RefObject, useEffect, useMemo, useRef, useState } from "react";
import {
  AlertTriangle,
  CalendarDays,
  CheckCircle2,
  ChevronRight,
  EllipsisVertical,
  ListChecks,
  Lock,
  LockOpen,
  MapPin,
  Pencil,
  Plus,
  Printer,
  Search,
  TrainFront,
  UsersRound,
  X
} from "lucide-react";

import {
  api,
  ApiError,
  ExhibitionConflict,
  ExhibitionEntryInput,
  ExhibitionList,
  ExhibitionListInput,
  ExhibitionStatus,
  ExhibitionWorkspace,
  ExhibitionWorkspaceEntry,
  MasterDataEntry,
  Vehicle
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppDateInput } from "../../shared/ui/AppDateInput";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import { ExhibitionEntryDialog, ExhibitionEntryMasterData } from "./ExhibitionEntryDialog";

type EntryFilter = "all" | "ready" | "check";
type ExhibitionDialog = "event" | "entry" | "conflicts" | "lock" | null;

const emptyEntryMasterData: ExhibitionEntryMasterData = {
  manufacturers: [],
  gattungen: [],
  epochs: [],
  railwayCompanies: []
};

function ExhibitionDialogLayer({
  children,
  onClose
}: {
  children: (closeRef: RefObject<HTMLButtonElement | null>) => ReactNode;
  onClose: () => void;
}) {
  const closeRef = useRef<HTMLButtonElement | null>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(onClose, closeRef);
  return (
    <div ref={layerRef} className="modal-layer" role="presentation" onKeyDown={onKeyDown}>
      <span ref={anchorRef} hidden />
      {children(closeRef)}
    </div>
  );
}

const statusLabels: Record<ExhibitionStatus, { de: string; en: string }> = {
  draft: { de: "Entwurf", en: "Draft" },
  open: { de: "Offen", en: "Open" },
  locked: { de: "Gesperrt", en: "Locked" },
  running: { de: "Läuft", en: "Running" },
  completed: { de: "Abgeschlossen", en: "Completed" },
  archived: { de: "Archiviert", en: "Archived" }
};

const entryStatusLabels = {
  ready: { de: "Bereit", en: "Ready" },
  addressConflict: { de: "Adresskonflikt", en: "Address conflict" },
  missing: { de: "Angaben fehlen", en: "Missing data" },
  check: { de: "Prüfen", en: "Check" },
  unavailable: { de: "Nicht verfügbar", en: "Unavailable" }
};

function localeDate(value: string, language: "de" | "en", options?: Intl.DateTimeFormatOptions) {
  if (!value) return "–";
  return new Intl.DateTimeFormat(language === "de" ? "de-DE" : "en-GB", options ?? {
    day: "2-digit", month: "long", year: "numeric"
  }).format(new Date(`${value}T12:00:00`));
}

function eventDateRange(event: ExhibitionList, language: "de" | "en") {
  const end = event.endDate || event.date;
  if (event.date === end) return localeDate(event.date, language);
  const startDate = new Date(`${event.date}T12:00:00`);
  const endDate = new Date(`${end}T12:00:00`);
  if (startDate.getMonth() === endDate.getMonth() && startDate.getFullYear() === endDate.getFullYear()) {
    return `${startDate.getDate()}.–${localeDate(end, language)}`;
  }
  return `${localeDate(event.date, language)} – ${localeDate(end, language)}`;
}

function relativeDate(event: ExhibitionList, language: "de" | "en") {
  const today = new Date();
  today.setHours(12, 0, 0, 0);
  const start = new Date(`${event.date}T12:00:00`);
  const days = Math.round((start.getTime() - today.getTime()) / 86_400_000);
  if (days === 0) return language === "de" ? "Heute" : "Today";
  if (days > 0) return language === "de" ? `In ${days} Tagen` : `In ${days} days`;
  return language === "de" ? `Vor ${Math.abs(days)} Tagen` : `${Math.abs(days)} days ago`;
}

function eventDayDate(event: ExhibitionList, index: number, language: "de" | "en") {
  const date = new Date(`${event.date}T12:00:00`);
  date.setDate(date.getDate() + index);
  return new Intl.DateTimeFormat(language === "de" ? "de-DE" : "en-GB", {
    day: "2-digit", month: "2-digit"
  }).format(date);
}

function dayScopeIncludes(value: string, scope: string) {
  return value === "all" || value.split(",").map((part) => part.trim()).includes(scope);
}

function hasAdmin(roles: string[]) {
  return roles.includes("Admin");
}

function canReadInventory(roles: string[]) {
  return roles.some((role) => ["Admin", "Editor", "Viewer", "Planner"].includes(role));
}

function newEventInput(): ExhibitionListInput {
  const date = new Date().toISOString().slice(0, 10);
  return { designation: "", date, endDate: date, location: "", description: "", organizationNotes: "", status: "open" };
}

function newEntryInput(): ExhibitionEntryInput {
  return {
    owner: "", locomotiveName: "", dayScope: "all", dtDecoder: true, decoderNumber: "",
    interfaceName: "", adapter: "", analog: false, availability: "available", imageUrl: "",
    functionKeys: "", notes: ""
  };
}

function conflictDescription(conflict: ExhibitionConflict, language: "de" | "en") {
  if (conflict.kind === "address") {
    return language === "de"
      ? `Adresse ${conflict.address || "–"} auf ${conflict.interfaceName || "–"} überschneidet sich.`
      : `Address ${conflict.address || "–"} on ${conflict.interfaceName || "–"} overlaps.`;
  }
  if (conflict.kind === "duplicateVehicle") {
    return language === "de" ? "Fahrzeug ist mehrfach eingetragen." : "Vehicle is entered more than once.";
  }
  return language === "de"
    ? `Pflichtangaben fehlen: ${(conflict.fields || []).join(", ")}`
    : `Required data missing: ${(conflict.fields || []).join(", ")}`;
}

export function ExhibitionWorkspacePage({ roles }: { roles: string[] }) {
  const { language } = useI18n();
  const de = language === "de";
  const admin = hasAdmin(roles);
  const manager = admin || roles.includes("Editor");
  const entryManager = manager || roles.includes("Messe");
  const [lists, setLists] = useState<ExhibitionList[]>([]);
  const [selectedID, setSelectedID] = useState("");
  const [workspace, setWorkspace] = useState<ExhibitionWorkspace | null>(null);
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [entryMasterData, setEntryMasterData] = useState<ExhibitionEntryMasterData>(emptyEntryMasterData);
  const [symbols, setSymbols] = useState<MasterDataEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [dialog, setDialog] = useState<ExhibitionDialog>(null);
  const [editingEvent, setEditingEvent] = useState<ExhibitionList | null>(null);
  const [editingEntry, setEditingEntry] = useState<ExhibitionWorkspaceEntry | null>(null);
  const [eventForm, setEventForm] = useState<ExhibitionListInput>(newEventInput);
  const [entryForm, setEntryForm] = useState<ExhibitionEntryInput>(newEntryInput);
  const [activeDay, setActiveDay] = useState("all");
  const [entryFilter, setEntryFilter] = useState<EntryFilter>("all");
  const [search, setSearch] = useState("");
  const [showArchived, setShowArchived] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [lockReason, setLockReason] = useState("");
  const [exceptionReasons, setExceptionReasons] = useState<Record<string, string>>({});

  const selectedList = lists.find((item) => item.id === selectedID) || workspace?.list || null;
  const locked = Boolean(selectedList?.locked);
  const planningOpen = selectedList?.status === "draft" || selectedList?.status === "open";
  const canEditEntries = entryManager && planningOpen && !locked;

  const loadLists = async (preferredID?: string) => {
    setLoading(true);
    try {
      const next = await api.exhibitionLists();
      setLists(next);
      setSelectedID((current) => preferredID || current || next.find((item) => item.status !== "archived")?.id || next[0]?.id || "");
      setMessage("");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : de ? "Veranstaltungen konnten nicht geladen werden." : "Events could not be loaded.");
    } finally {
      setLoading(false);
    }
  };

  const loadWorkspace = async (id: string) => {
    try {
      const next = await api.exhibitionWorkspace(id);
      setWorkspace(next);
      setLists((current) => current.map((item) => item.id === id ? { ...next.list, entryCount: next.summary.entryCount } : item));
      setMessage("");
    } catch (error) {
      setWorkspace(null);
      setMessage(error instanceof Error ? error.message : de ? "Arbeitsansicht konnte nicht geladen werden." : "Workspace could not be loaded.");
    }
  };

  useEffect(() => { void loadLists(); }, []);
  useEffect(() => {
    if (selectedID) void loadWorkspace(selectedID);
    else setWorkspace(null);
  }, [selectedID]);
  useEffect(() => {
    if (!canReadInventory(roles)) return;
    api.vehicles().then(setVehicles).catch(() => setVehicles([]));
  }, [roles]);
  useEffect(() => {
    api.masterData("symbols").then(setSymbols).catch(() => setSymbols([]));
    if (!canReadInventory(roles)) {
      setEntryMasterData(emptyEntryMasterData);
      return;
    }
    api.masterDataAll(true).then((entriesByType) => setEntryMasterData({
      manufacturers: entriesByType.manufacturer || [],
      gattungen: entriesByType.vehicle_gattung || [],
      epochs: entriesByType.epoch || [],
      railwayCompanies: entriesByType.railway_company || []
    })).catch(() => setEntryMasterData(emptyEntryMasterData));
  }, [roles]);

  const visibleLists = useMemo(() => lists.filter((item) => showArchived || item.status !== "archived"), [lists, showArchived]);
  const filteredEntries = useMemo(() => {
    const query = search.trim().toLocaleLowerCase(language === "de" ? "de" : "en");
    return (workspace?.entries || []).filter((entry) => {
      if (activeDay !== "all" && !dayScopeIncludes(entry.dayScope, activeDay)) return false;
      if (entryFilter === "ready" && entry.status !== "ready") return false;
      if (entryFilter === "check" && entry.status === "ready") return false;
      if (!query) return true;
      return [entry.locomotiveName, entry.owner, entry.decoderNumber, entry.interfaceName]
        .some((value) => String(value || "").toLocaleLowerCase(language === "de" ? "de" : "en").includes(query));
    });
  }, [activeDay, entryFilter, language, search, workspace?.entries]);

  const openEventDialog = (event?: ExhibitionList) => {
    setEditingEvent(event || null);
    setEventForm(event ? {
      designation: event.designation, date: event.date, endDate: event.endDate || event.date,
      location: event.location || "", description: event.description || "",
      organizationNotes: event.organizationNotes || "", status: event.status,
      expectedRevision: event.revision
    } : newEventInput());
    setDialog("event");
    setMenuOpen(false);
  };

  const saveEvent = async (event: FormEvent) => {
    event.preventDefault();
    if (!manager) return;
    setSaving(true);
    try {
      const saved = editingEvent
        ? await api.updateExhibitionList(editingEvent.id, eventForm)
        : await api.createExhibitionList(eventForm);
      setDialog(null);
      await loadLists(saved.id);
      await loadWorkspace(saved.id);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : de ? "Veranstaltung konnte nicht gespeichert werden." : "Event could not be saved.");
    } finally {
      setSaving(false);
    }
  };

  const openEntryDialog = (entry?: ExhibitionWorkspaceEntry) => {
    setEditingEntry(entry || null);
    setEntryForm(entry ? {
      vehicleId: entry.vehicleId, owner: entry.owner, locomotiveName: entry.locomotiveName,
      imageUrl: entry.imageUrl || "", gattung: entry.gattung || "", series: entry.series || "",
      manufacturer: entry.manufacturer || "", epoch: entry.epoch || "",
      railwayCompany: entry.railwayCompany || "", dayScope: entry.dayScope,
      dtDecoder: entry.dtDecoder, decoderNumber: entry.decoderNumber || "",
      decoderType: entry.decoderType || "", interfaceName: entry.interfaceName || "",
      adapter: entry.adapter || "", sxAddress: entry.sxAddress || "", analog: entry.analog,
      availability: entry.availability, functionKeys: entry.functionKeys || "", notes: entry.notes || "",
      sortOrder: entry.sortOrder, expectedRevision: entry.revision
    } : newEntryInput());
    setDialog("entry");
  };

  const selectVehicle = (vehicleID: string) => {
    const vehicle = vehicles.find((item) => item.id === vehicleID);
    if (!vehicle) {
      setEntryForm((current) => ({ ...current, vehicleId: "" }));
      return;
    }
    setEntryForm((current) => ({
      ...current,
      vehicleId: vehicle.id,
      locomotiveName: vehicle.name,
      manufacturer: vehicle.manufacturer,
      series: vehicle.series || "",
      gattung: vehicle.gattung || vehicle.category || "",
      epoch: vehicle.epoch || "",
      railwayCompany: vehicle.railwayCompany || "",
      dtDecoder: vehicle.dtDecoder || vehicle.digital,
      decoderNumber: vehicle.dtDecoderNumber || vehicle.digitalDecoderNumber || "",
      decoderType: vehicle.decoderType || ""
    }));
  };

  const saveEntry = async (event: FormEvent) => {
    event.preventDefault();
    if (!selectedID || !canEditEntries) return;
    setSaving(true);
    try {
      if (editingEntry) await api.updateExhibitionEntry(selectedID, editingEntry.id, entryForm);
      else await api.createExhibitionEntry(selectedID, entryForm);
      setDialog(null);
      await loadWorkspace(selectedID);
      await loadLists(selectedID);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : de ? "Eintrag konnte nicht gespeichert werden." : "Entry could not be saved.");
    } finally {
      setSaving(false);
    }
  };

  const toggleDay = (scope: string) => {
    if (scope === "all") {
      setEntryForm((current) => ({ ...current, dayScope: "all" }));
      return;
    }
    const current = entryForm.dayScope === "all" ? [] : (entryForm.dayScope || "").split(",").filter(Boolean);
    const next = current.includes(scope) ? current.filter((item) => item !== scope) : [...current, scope];
    setEntryForm((value) => ({ ...value, dayScope: next.length ? next.join(",") : "all" }));
  };

  const updateStatus = async (status: ExhibitionStatus, reason = "", confirmConflicts = false) => {
    if (!selectedList || !manager) return;
    setSaving(true);
    try {
      await api.setExhibitionStatus(selectedList.id, {
        status, expectedRevision: workspace?.list.revision || selectedList.revision, reason, confirmConflicts
      });
      setDialog(null);
      setLockReason("");
      await loadWorkspace(selectedList.id);
      await loadLists(selectedList.id);
    } catch (error) {
      if (error instanceof ApiError && error.code === "exhibition_conflicts") {
        setDialog("lock");
      } else {
        setMessage(error instanceof Error ? error.message : de ? "Status konnte nicht geändert werden." : "Status could not be changed.");
      }
    } finally {
      setSaving(false);
    }
  };

  const deleteEntry = async (entry: ExhibitionWorkspaceEntry) => {
    if (!selectedID || !admin || locked) return;
    const confirmed = window.confirm(de ? `Eintrag „${entry.locomotiveName}“ löschen?` : `Delete “${entry.locomotiveName}”?`);
    if (!confirmed) return;
    await api.deleteExhibitionEntry(selectedID, entry.id);
    await loadWorkspace(selectedID);
    await loadLists(selectedID);
  };

  const deleteEvent = async () => {
    if (!selectedList || !admin) return;
    if (!window.confirm(de ? `Veranstaltung „${selectedList.designation}“ löschen?` : `Delete “${selectedList.designation}”?`)) return;
    await api.deleteExhibitionList(selectedList.id);
    setMenuOpen(false);
    setSelectedID("");
    await loadLists();
  };

  const saveException = async (conflict: ExhibitionConflict) => {
    if (!selectedList || !workspace) return;
    const reason = (exceptionReasons[conflict.id] || "").trim();
    if (!reason) return;
    const next = await api.setExhibitionConflictException(selectedList.id, conflict.id, reason, workspace.list.revision);
    setWorkspace(next);
    setExceptionReasons((current) => ({ ...current, [conflict.id]: "" }));
  };

  if (loading && lists.length === 0) {
    return <section className="exhibition-workspace-page"><p>{de ? "Ausstellung wird geladen…" : "Loading exhibition…"}</p></section>;
  }

  return (
    <section className="exhibition-workspace-page">
      <header className="exhibition-page-head">
        <div>
          <p className="eyebrow">{de ? "MESSEBETRIEB" : "EXHIBITION OPERATIONS"}</p>
          <h1>{de ? "Ausstellung" : "Exhibition"}</h1>
          <p>{de ? "Veranstaltungen vorbereiten, Fahrzeuge koordinieren und Fahrtage sicher durchführen." : "Prepare events, coordinate vehicles and run operating days safely."}</p>
        </div>
        <div className="exhibition-page-actions">
          {manager && <button type="button" className="primary-button" onClick={() => openEventDialog()}><Plus size={18} />{de ? "Neue Veranstaltung" : "New event"}</button>}
          <div className="exhibition-menu-wrap">
            <button type="button" className="secondary-button icon-only" aria-label={de ? "Weitere Aktionen" : "More actions"} onClick={() => setMenuOpen((value) => !value)}><EllipsisVertical size={20} /></button>
            {menuOpen && <div className="exhibition-overflow-menu">
              <button type="button" onClick={() => setShowArchived((value) => !value)}>{showArchived ? (de ? "Archive ausblenden" : "Hide archive") : (de ? "Archive anzeigen" : "Show archive")}</button>
              {manager && selectedList && planningOpen && <button type="button" onClick={() => openEventDialog(selectedList)}>{de ? "Veranstaltung bearbeiten" : "Edit event"}</button>}
              {manager && selectedList?.status === "draft" && <button type="button" onClick={() => void updateStatus("open")}>{de ? "Planung öffnen" : "Open planning"}</button>}
              {manager && selectedList?.status === "locked" && <button type="button" onClick={() => void updateStatus("running")}>{de ? "Veranstaltung starten" : "Start event"}</button>}
              {manager && selectedList?.status === "running" && <button type="button" onClick={() => void updateStatus("completed")}>{de ? "Veranstaltung abschließen" : "Complete event"}</button>}
              {manager && selectedList?.status === "completed" && <button type="button" onClick={() => void updateStatus("open")}>{de ? "Planung wieder öffnen" : "Reopen planning"}</button>}
              {manager && selectedList?.status === "archived" && <button type="button" onClick={() => void updateStatus("open")}>{de ? "Aus Archiv holen" : "Restore from archive"}</button>}
              {manager && selectedList && ["draft", "open", "locked", "completed"].includes(selectedList.status) && <button type="button" onClick={() => void updateStatus("archived")}>{de ? "Archivieren" : "Archive"}</button>}
              {admin && selectedList && <button type="button" className="danger" onClick={() => void deleteEvent()}>{de ? "Löschen" : "Delete"}</button>}
            </div>}
          </div>
        </div>
      </header>

      {message && <div className="exhibition-alert" role="alert">{message}<button type="button" onClick={() => setMessage("")} aria-label={de ? "Meldung schließen" : "Dismiss message"}><X size={16} /></button></div>}

      {workspace && selectedList ? <>
        <section className="exhibition-event-summary" aria-label={de ? "Veranstaltungsübersicht" : "Event overview"}>
          <div className="exhibition-event-title-row">
            <div className="exhibition-event-heading">
              <span className="exhibition-event-icon"><CalendarDays size={26} /></span>
              <div><h2>{selectedList.designation}</h2><p>{eventDateRange(selectedList, language)}</p></div>
              <span className="exhibition-pill info">{relativeDate(selectedList, language)}</span>
              <span className={`exhibition-pill status-${selectedList.status}`}>{statusLabels[selectedList.status][language]}</span>
              {selectedList.location && <span className="exhibition-event-location"><MapPin size={14} />{selectedList.location}</span>}
            </div>
            <div className="exhibition-summary-actions">
              <button type="button" className="secondary-button" onClick={() => window.print()}><Printer size={17} />{de ? "Drucken" : "Print"}</button>
              {manager && (selectedList.status === "open" || selectedList.status === "locked") && <button type="button" className="secondary-button" onClick={() => selectedList.locked ? void updateStatus("open", de ? "Liste wieder geöffnet" : "List reopened") : void updateStatus("locked")}>
                {selectedList.locked ? <LockOpen size={17} /> : <Lock size={17} />}{selectedList.locked ? (de ? "Liste entsperren" : "Unlock list") : (de ? "Liste sperren" : "Lock list")}
              </button>}
            </div>
          </div>
          <div className="exhibition-metric-grid">
            <article><ListChecks size={28} /><div><strong>{workspace.summary.entryCount}</strong><span>{de ? "Einträge" : "Entries"}</span></div></article>
            <article><UsersRound size={28} /><div><strong>{workspace.summary.ownerCount}</strong><span>{de ? "Besitzer" : "Owners"}</span></div></article>
            <article className="warning"><AlertTriangle size={28} /><div><strong>{workspace.summary.conflictCount}</strong><span>{de ? "Konflikte" : "Conflicts"}</span></div></article>
            <article><CheckCircle2 size={28} /><div><strong>{workspace.summary.readyCount}</strong><span>{de ? "bereit" : "ready"}</span></div></article>
          </div>
        </section>

        <div className="exhibition-main-grid">
          <section className="exhibition-events-panel" aria-label={de ? "Veranstaltungen" : "Events"}>
            <div className="exhibition-panel-title"><h2>{de ? "Veranstaltungen" : "Events"}</h2>{manager && <button type="button" className="icon-button" onClick={() => openEventDialog()} aria-label={de ? "Veranstaltung anlegen" : "Create event"}><Plus size={18} /></button>}</div>
            <div className="exhibition-event-list">
              {visibleLists.map((item) => <button type="button" key={item.id} className={item.id === selectedID ? "active" : ""} onClick={() => { setSelectedID(item.id); setActiveDay("all"); setEntryFilter("all"); }}>
                <span className="event-list-icon"><CalendarDays size={20} /></span>
                <span className="event-list-copy"><strong>{item.designation}</strong><span>{eventDateRange(item, language)}</span><span>{item.entryCount} {de ? "Einträge" : "entries"}</span></span>
                <span className="event-list-state"><span className={`exhibition-pill status-${item.status}`}>{statusLabels[item.status][language]}</span><small>{relativeDate(item, language)}</small></span>
              </button>)}
            </div>
            <button type="button" className="exhibition-all-events" onClick={() => setShowArchived(true)}><ListChecks size={17} />{de ? "Alle Veranstaltungen anzeigen" : "Show all events"}<ChevronRight size={17} /></button>
          </section>

          <section className="exhibition-operations-panel" aria-label={de ? "Teilnehmer und Fahrzeuge" : "Participants and vehicles"}>
            <div className="exhibition-operations-title"><h2>{de ? "Teilnehmer und Fahrzeuge" : "Participants and vehicles"}</h2>{canEditEntries && <button type="button" className="icon-button" onClick={() => openEntryDialog()} aria-label={de ? "Eintrag hinzufügen" : "Add entry"}><Plus size={18} /></button>}</div>
            <div className="exhibition-day-tabs" role="tablist" aria-label={de ? "Veranstaltungstage" : "Event days"}>
              <button type="button" className={activeDay === "all" ? "active" : ""} onClick={() => setActiveDay("all")}>{de ? "Alle Tage" : "All days"}</button>
              {workspace.dayScopes.map((scope, index) => <button type="button" key={scope} className={activeDay === scope ? "active" : ""} onClick={() => setActiveDay(scope)}>
                {de ? `Tag ${index + 1}` : `Day ${index + 1}`}<small>{eventDayDate(selectedList, index, language)}</small>
              </button>)}
            </div>
            <div className="exhibition-table-toolbar">
              <div className="exhibition-quick-filters">
                <button type="button" className={entryFilter === "all" ? "active" : ""} onClick={() => setEntryFilter("all")}>{de ? "Alle" : "All"}<span>{workspace.summary.entryCount}</span></button>
                <button type="button" className={entryFilter === "ready" ? "active" : ""} onClick={() => setEntryFilter("ready")}>{de ? "Bereit" : "Ready"} <span>{workspace.summary.readyCount}</span></button>
                <button type="button" className={entryFilter === "check" ? "active" : ""} onClick={() => setEntryFilter("check")}>{de ? "Prüfen" : "Check"} <span>{workspace.summary.entryCount - workspace.summary.readyCount}</span></button>
              </div>
              <label className="exhibition-search"><Search size={17} /><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder={de ? "Suchen…" : "Search…"} aria-label={de ? "Einträge durchsuchen" : "Search entries"} /></label>
              <button type="button" className="secondary-button" onClick={() => setDialog("conflicts")}><AlertTriangle size={17} />{de ? "Konflikte prüfen" : "Check conflicts"}</button>
            </div>
            <div className="exhibition-table-wrap">
              <table className="exhibition-operations-table">
                <thead><tr><th>{de ? "Fahrzeug" : "Vehicle"}</th><th>{de ? "Besitzer" : "Owner"}</th><th>{de ? "Adresse" : "Address"}</th><th>{de ? "Schnittstelle" : "Interface"}</th><th>{de ? "Tage" : "Days"}</th><th>Status</th><th>{de ? "Aktion" : "Action"}</th></tr></thead>
                <tbody>
                  {filteredEntries.map((entry) => <tr key={entry.id}>
                    <td><div className="exhibition-vehicle-cell">{entry.imageUrl ? <img src={entry.imageUrl} alt="" /> : <span><TrainFront size={23} /></span>}<strong>{entry.locomotiveName}</strong></div></td>
                    <td>{entry.owner || "–"}</td><td>{entry.analog ? (de ? "Analog" : "Analogue") : entry.decoderNumber || entry.sxAddress || "–"}</td><td>{entry.interfaceName || entry.adapter || "–"}</td>
                    <td><div className="exhibition-day-badges">{workspace.dayScopes.map((scope, index) => dayScopeIncludes(entry.dayScope, scope) && <span key={scope}>{index + 1}</span>)}</div></td>
                    <td><span className={`exhibition-entry-status status-${entry.status}`}>{entry.status === "ready" ? <CheckCircle2 size={14} /> : <AlertTriangle size={14} />}{entryStatusLabels[entry.status][language]}</span></td>
                    <td><div className="exhibition-row-menu">{canEditEntries && <button type="button" className="icon-button" onClick={() => openEntryDialog(entry)} aria-label={de ? `${entry.locomotiveName} bearbeiten` : `Edit ${entry.locomotiveName}`} title={de ? `${entry.locomotiveName} bearbeiten` : `Edit ${entry.locomotiveName}`}><Pencil size={17} /></button>}{admin && canEditEntries && <button type="button" className="visually-hidden-action" onClick={() => void deleteEntry(entry)}>{de ? "Löschen" : "Delete"}</button>}</div></td>
                  </tr>)}
                  {filteredEntries.length === 0 && <tr><td colSpan={7} className="exhibition-empty-row">{de ? "Keine Einträge für diese Auswahl." : "No entries match this selection."}</td></tr>}
                </tbody>
              </table>
            </div>
          </section>
        </div>

        <section className="exhibition-readiness" aria-label={de ? "Bereitschaft" : "Readiness"}>
          <h2>{de ? "Bereitschaft" : "Readiness"}</h2>
          <div><button type="button" onClick={() => setEntryFilter("check")}><CheckCircle2 size={20} /><span>{de ? "Adressen geprüft" : "Addresses checked"}</span><strong>{workspace.readiness.addressesChecked} / {workspace.readiness.total}</strong></button>
            <button type="button" onClick={() => setEntryFilter("check")}><CheckCircle2 size={20} /><span>{de ? "Funktionstasten" : "Function keys"}</span><strong>{workspace.readiness.functionsDocumented} / {workspace.readiness.total}</strong></button>
            <button type="button" onClick={() => setEntryFilter("check")}><CheckCircle2 size={20} /><span>{de ? "Bilder" : "Images"}</span><strong>{workspace.readiness.imagesPresent} / {workspace.readiness.total}</strong></button>
            <button type="button" className="problems" onClick={() => setDialog("conflicts")}><AlertTriangle size={20} /><strong>{workspace.readiness.problems} {de ? "Probleme öffnen" : "open problems"}</strong><ChevronRight size={18} /></button></div>
        </section>
      </> : <section className="exhibition-empty-state"><CalendarDays size={34} /><h2>{de ? "Noch keine Veranstaltung" : "No event yet"}</h2><p>{de ? "Legen Sie die erste Veranstaltung an." : "Create the first event."}</p>{manager && <button type="button" className="primary-button" onClick={() => openEventDialog()}><Plus size={17} />{de ? "Neue Veranstaltung" : "New event"}</button>}</section>}

      {dialog === "event" && <ExhibitionDialogLayer onClose={() => setDialog(null)}>{(closeRef) => <form className="vehicle-modal exhibition-event-modal" onSubmit={saveEvent} role="dialog" aria-modal="true" aria-labelledby="exhibition-event-dialog-title">
        <header><div><h2 id="exhibition-event-dialog-title">{editingEvent ? (de ? "Veranstaltung bearbeiten" : "Edit event") : (de ? "Neue Veranstaltung" : "New event")}</h2><p>{de ? "Termin, Ort und organisatorische Angaben." : "Dates, location and organisational details."}</p></div><button ref={closeRef} type="button" className="icon-button" onClick={() => setDialog(null)} aria-label={de ? "Schließen" : "Close"}><X size={19} /></button></header>
        <div className="exhibition-form-grid"><label className="wide"><span>{de ? "Bezeichnung" : "Name"}</span><input required value={eventForm.designation} onChange={(event) => setEventForm({ ...eventForm, designation: event.target.value })} /></label>
          <label><span>{de ? "Beginn" : "Start"}</span><AppDateInput required value={eventForm.date} onChange={(event) => setEventForm({ ...eventForm, date: event.target.value })} /></label>
          <label><span>{de ? "Ende" : "End"}</span><AppDateInput required value={eventForm.endDate} onChange={(event) => setEventForm({ ...eventForm, endDate: event.target.value })} /></label>
          <label className="wide"><span>{de ? "Ort" : "Location"}</span><input value={eventForm.location} onChange={(event) => setEventForm({ ...eventForm, location: event.target.value })} /></label>
          <label className="wide"><span>{de ? "Beschreibung" : "Description"}</span><textarea rows={3} value={eventForm.description} onChange={(event) => setEventForm({ ...eventForm, description: event.target.value })} /></label>
          <label className="wide"><span>{de ? "Organisation (intern)" : "Organisation (internal)"}</span><textarea rows={3} value={eventForm.organizationNotes} onChange={(event) => setEventForm({ ...eventForm, organizationNotes: event.target.value })} /></label></div>
        <footer><button type="button" className="secondary-button" onClick={() => setDialog(null)}>{de ? "Abbrechen" : "Cancel"}</button><button type="submit" className="primary-button" disabled={saving}>{de ? "Speichern" : "Save"}</button></footer>
      </form>}</ExhibitionDialogLayer>}

      {dialog === "entry" && workspace && <ExhibitionDialogLayer onClose={() => setDialog(null)}>{(closeRef) => <ExhibitionEntryDialog
        admin={admin}
        closeRef={closeRef}
        de={de}
        editingEntry={editingEntry}
        form={entryForm}
        masterData={entryMasterData}
        saving={saving}
        symbols={symbols}
        vehicles={vehicles}
        workspace={workspace}
        canReadInventory={canReadInventory(roles)}
        onChange={(patch) => setEntryForm((current) => ({ ...current, ...patch }))}
        onClose={() => setDialog(null)}
        onDelete={(entry) => { setDialog(null); void deleteEntry(entry); }}
        onSelectVehicle={selectVehicle}
        onSubmit={saveEntry}
      />}</ExhibitionDialogLayer>}

      {dialog === "conflicts" && workspace && <ExhibitionDialogLayer onClose={() => setDialog(null)}>{(closeRef) => <section className="vehicle-modal exhibition-conflict-modal" role="dialog" aria-modal="true" aria-labelledby="exhibition-conflicts-title">
        <header><div><h2 id="exhibition-conflicts-title">{de ? "Konflikte und Prüfungen" : "Conflicts and checks"}</h2><p>{de ? "Ausnahmen bleiben sichtbar und gelten nicht als vollständig gelöst." : "Exceptions remain visible and are not considered fully resolved."}</p></div><button ref={closeRef} type="button" className="icon-button" onClick={() => setDialog(null)} aria-label={de ? "Schließen" : "Close"}><X size={19} /></button></header>
        <div className="exhibition-conflict-list">{workspace.conflicts.map((conflict) => <article key={conflict.id} className={conflict.excepted ? "excepted" : ""}><AlertTriangle size={20} /><div><strong>{conflictDescription(conflict, language)}</strong><span>{conflict.dayScopes?.join(", ")}</span>{conflict.excepted && <small>{de ? "Dokumentierte Ausnahme" : "Documented exception"}: {conflict.exceptionReason}</small>}</div>{!conflict.excepted && <div className="exhibition-exception-form"><input value={exceptionReasons[conflict.id] || ""} onChange={(event) => setExceptionReasons({ ...exceptionReasons, [conflict.id]: event.target.value })} placeholder={de ? "Begründung der Ausnahme" : "Exception reason"} /><button type="button" className="secondary-button" onClick={() => void saveException(conflict)}>{de ? "Ausnahme speichern" : "Save exception"}</button></div>}</article>)}{workspace.conflicts.length === 0 && <div className="exhibition-no-conflicts"><CheckCircle2 size={30} /><strong>{de ? "Keine Konflikte erkannt" : "No conflicts detected"}</strong></div>}</div>
        <footer><button type="button" className="primary-button" onClick={() => setDialog(null)}>{de ? "Schließen" : "Close"}</button></footer>
      </section>}</ExhibitionDialogLayer>}

      {dialog === "lock" && workspace && <ExhibitionDialogLayer onClose={() => setDialog(null)}>{(closeRef) => <form className="vehicle-modal exhibition-lock-modal" onSubmit={(event) => { event.preventDefault(); void updateStatus("locked", lockReason, true); }} role="dialog" aria-modal="true" aria-labelledby="exhibition-lock-title">
        <header><div><h2 id="exhibition-lock-title">{de ? "Liste trotz Konflikten sperren?" : "Lock list despite conflicts?"}</h2><p>{de ? "Die Konflikte bleiben im Arbeitsbereich sichtbar." : "Conflicts remain visible in the workspace."}</p></div><button ref={closeRef} type="button" className="icon-button" onClick={() => setDialog(null)} aria-label={de ? "Schließen" : "Close"}><X size={19} /></button></header>
        <label><span>{de ? "Begründung" : "Reason"}</span><textarea required rows={4} value={lockReason} onChange={(event) => setLockReason(event.target.value)} /></label>
        <footer><button type="button" className="secondary-button" onClick={() => setDialog(null)}>{de ? "Abbrechen" : "Cancel"}</button><button type="submit" className="primary-button" disabled={!lockReason.trim() || saving}><Lock size={17} />{de ? "Trotzdem sperren" : "Lock anyway"}</button></footer>
      </form>}</ExhibitionDialogLayer>}
    </section>
  );
}
