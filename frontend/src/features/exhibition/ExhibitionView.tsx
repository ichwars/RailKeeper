import { useEffect, useMemo, useState } from "react";
import {
  Edit3,
  Eye,
  Image as ImageIcon,
  Lock,
  LockOpen,
  Plus,
  Printer,
  Trash2,
  Upload
} from "lucide-react";
import { api, ExhibitionEntry, ExhibitionEntryInput, ExhibitionList, ExhibitionListInput, MasterDataEntry } from "../../shared/api";
import { FunctionSymbolPicker, functionSymbolIcon, functionSymbolMetadata } from "../../shared/functionSymbols";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";

type ListSortKey = "designation" | "date" | "entryCount" | "locked";
type EntrySortKey = "owner" | "locomotiveName" | "dtDecoder" | "decoderNumber" | "functionKeys";
type SortDirection = "asc" | "desc";
type EntryTab = "general" | "images" | "functions";
type PrintOptions = { includeImages: boolean };

type ExhibitionMasterDataOptions = {
  manufacturers: MasterDataEntry[];
  epochs: MasterDataEntry[];
  railwayCompanies: MasterDataEntry[];
  gattungen: MasterDataEntry[];
};

type ExhibitionFunction = {
  key: string;
  name: string;
  type: string;
  symbolKey?: string;
};

const emptyListForm: ExhibitionListInput = { designation: "", date: new Date().toISOString().slice(0, 10) };
const emptyEntryForm: ExhibitionEntryInput = {
  owner: "",
  imageUrl: "",
  locomotiveName: "",
  gattung: "",
  series: "",
  manufacturer: "",
  epoch: "",
  railwayCompany: "",
  dayScope: "all",
  dtDecoder: false,
  decoderNumber: "",
  decoderType: "",
  adapter: "",
  sxAddress: "",
  analog: false,
  functionKeys: "",
  notes: ""
};
const emptyExhibitionOptions: ExhibitionMasterDataOptions = {
  manufacturers: [],
  epochs: [],
  railwayCompanies: [],
  gattungen: []
};
const functionKeys = Array.from({ length: 32 }, (_, index) => `F${index}`);
const functionTypes = ["standard", "licht", "sound", "kupplung", "rauch", "sonderfunktion"];
const adapterOptions = ["NEM 651", "NEM 652", "PluX16", "PluX22", "MTC21", "Next18", "8-polig", "21-polig"];
const dayScopes = ["all", "day1", "day2", "day3", "day4"];
const selectableDayScopes = dayScopes.filter((scope) => scope !== "all");
const htmlEscapes: Record<string, string> = {
  "&": "&amp;",
  "<": "&lt;",
  ">": "&gt;",
  "\"": "&quot;",
  "'": "&#39;"
};
function defaultFunction(key: string): ExhibitionFunction {
  return {
    key,
    name: key === "F0" ? "Fahrlicht" : "",
    type: key === "F0" ? "licht" : "standard",
    symbolKey: key === "F0" ? "light" : ""
  };
}

const emptyFunctions = () => functionKeys.map(defaultFunction);

function isConfiguredFunction(item: ExhibitionFunction) {
  const fallback = defaultFunction(item.key);
  return Boolean(item.name.trim() || item.symbolKey || (item.type || fallback.type) !== fallback.type);
}

function functionDisplayName(item: ExhibitionFunction) {
  return item.name.trim() || item.symbolKey || item.type || item.key;
}

function hasAdmin(roles: string[]) {
  return roles.includes("Admin");
}

function formatDate(value: string, language: string) {
  if (!value) return "-";
  return new Intl.DateTimeFormat(language === "en" ? "en-US" : "de-DE", { day: "2-digit", month: "2-digit", year: "numeric" }).format(new Date(`${value}T00:00:00`));
}

function sortValue(value: unknown) {
  if (typeof value === "boolean") return value ? "1" : "0";
  if (typeof value === "number") return String(value).padStart(8, "0");
  return String(value || "").toLowerCase();
}

function sortEntries(entries: ExhibitionEntry[], sort: { key: EntrySortKey; direction: SortDirection }) {
  return [...entries].sort((a, b) => {
    const result = sortValue(a[sort.key]).localeCompare(sortValue(b[sort.key]), "de");
    return sort.direction === "asc" ? result : -result;
  });
}

function optionValue(entry: MasterDataEntry) {
  return entry.label;
}

function normalizeAddress(value?: string) {
  return String(value || "").trim().toLowerCase();
}

function parseFunctions(value?: string): ExhibitionFunction[] {
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
      if (match) byKey.set(match[1].toUpperCase(), { key: match[1].toUpperCase(), name: match[2].trim(), type: "standard" });
    }
    return emptyFunctions().map((item) => ({ ...item, ...(byKey.get(item.key) || {}) }));
  }
  return emptyFunctions();
}

function serializeFunctions(functions: ExhibitionFunction[]) {
  return JSON.stringify(functions.filter(isConfiguredFunction).map((item) => {
    const fallback = defaultFunction(item.key);
    return {
      key: item.key,
      name: item.name.trim(),
      type: item.type || fallback.type,
      symbolKey: item.symbolKey || ""
    };
  }));
}

function displayFunctions(value?: string) {
  const configured = parseFunctions(value).filter(isConfiguredFunction);
  if (configured.length === 0) return "-";
  return configured.map((item) => `${item.key} ${functionDisplayName(item)}`).join(", ");
}

function configuredFunctions(value?: string) {
  return parseFunctions(value).filter(isConfiguredFunction);
}

function selectedDayScopes(value: string | undefined) {
  const raw = String(value || "").split(",").map((scope) => scope.trim()).filter(Boolean);
  if (raw.length === 0 || raw.includes("all")) return ["all"];
  const selected = selectableDayScopes.filter((scope) => raw.includes(scope));
  if (selected.length === 0 || selected.length === selectableDayScopes.length) return ["all"];
  return selected;
}

function normalizeDayScopeSelection(scopes: string[]) {
  if (scopes.includes("all")) return "all";
  const selected = selectableDayScopes.filter((scope) => scopes.includes(scope));
  if (selected.length === 0 || selected.length === selectableDayScopes.length) return "all";
  return selected.join(",");
}

function toggleDayScope(value: string | undefined, scope: string) {
  if (scope === "all") return "all";
  const current = selectedDayScopes(value).filter((item) => item !== "all");
  const next = current.includes(scope) ? current.filter((item) => item !== scope) : [...current, scope];
  return normalizeDayScopeSelection(next);
}

function isDayScopeActive(value: string | undefined, scope: string) {
  const selected = selectedDayScopes(value);
  return scope === "all" ? selected.includes("all") : selected.includes(scope);
}

function dayScopeLabel(value: string | undefined, t: (key: string, values?: Record<string, string | number>) => string) {
  const selected = selectedDayScopes(value);
  if (selected.includes("all")) return t("exhibition.dayScope.all");
  return selected.map((scope) => t(`exhibition.dayScope.${scope}`)).join(", ");
}

function locomotiveTitle(entry: Pick<ExhibitionEntry, "locomotiveName" | "railwayCompany">) {
  return [entry.locomotiveName, entry.railwayCompany ? `(${entry.railwayCompany})` : ""].filter(Boolean).join(" ");
}

function modelMeta(entry: Pick<ExhibitionEntry, "manufacturer" | "gattung" | "epoch">) {
  return [entry.manufacturer, entry.gattung, entry.epoch].filter(Boolean);
}

function controlRows(
  entry: Pick<ExhibitionEntry, "decoderNumber" | "sxAddress" | "analog" | "decoderType" | "adapter">,
  t: (key: string, values?: Record<string, string | number>) => string
) {
  const address = [
    entry.decoderNumber || "",
    entry.sxAddress ? `SX ${entry.sxAddress}` : ""
  ].filter(Boolean).join(" / ");
  return [
    [t("exhibition.address"), address || "-"],
    [t("exhibition.analog"), entry.analog ? t("exhibition.yes") : t("exhibition.no")],
    [t("exhibition.decoder"), entry.decoderType || "-"],
    [t("exhibition.interface"), entry.adapter || "-"]
  ];
}

function symbolImageDataFromMetadata(metadata?: Record<string, unknown>) {
  const value = metadata?.imageData || metadata?.activeImageData || metadata?.svgData;
  return typeof value === "string" ? value : "";
}

function printFunctionChips(value: string | undefined, symbols: MasterDataEntry[]) {
  const configured = parseFunctions(value).filter(isConfiguredFunction);
  if (configured.length === 0) return "-";
  return configured.map((item) => {
    const metadata = functionSymbolMetadata(symbols, item.symbolKey);
    const imageData = symbolImageDataFromMetadata(metadata);
    const label = functionDisplayName(item);
    return `<span class="function-chip">${imageData ? `<img src="${escapeHTML(imageData)}" alt="" />` : ""}<strong>${escapeHTML(item.key)}</strong> ${escapeHTML(label)}</span>`;
  }).join("");
}

function escapeHTML(value: unknown) {
  return String(value ?? "").replace(/[&<>"']/g, (char) => htmlEscapes[char] || char);
}

function fileToDataURL(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result || ""));
    reader.onerror = () => reject(reader.error || new Error("Bild konnte nicht gelesen werden."));
    reader.readAsDataURL(file);
  });
}

function printHTMLDocument(html: string) {
  const frame = document.createElement("iframe");
  frame.setAttribute("aria-hidden", "true");
  frame.style.position = "fixed";
  frame.style.right = "0";
  frame.style.bottom = "0";
  frame.style.width = "0";
  frame.style.height = "0";
  frame.style.border = "0";
  frame.style.opacity = "0";
  document.body.appendChild(frame);

  const printWindow = frame.contentWindow;
  if (!printWindow) {
    frame.remove();
    return;
  }

  let printed = false;
  const runPrint = () => {
    if (printed) return;
    printed = true;
    printWindow.focus();
    printWindow.print();
    window.setTimeout(() => frame.remove(), 1200);
  };

  frame.onload = runPrint;
  printWindow.document.open();
  printWindow.document.write(html);
  printWindow.document.close();
  window.setTimeout(runPrint, 250);
}

function printList(
  list: ExhibitionList,
  entries: ExhibitionEntry[],
  symbols: MasterDataEntry[] = [],
  language = "de",
  t: (key: string, values?: Record<string, string | number>) => string = (key) => key,
  options: PrintOptions = { includeImages: true }
) {
  const rows = entries.map((entry) => {
    const modelParts = modelMeta(entry);
    const controlParts = controlRows(entry, t);
    return `
    <tr>
      ${options.includeImages ? `<td class="image-cell">${entry.imageUrl ? `<img src="${escapeHTML(entry.imageUrl)}" alt="" />` : `<span>-</span>`}</td>` : ""}
      <td class="owner-cell">${escapeHTML(entry.owner)}<small>${escapeHTML(dayScopeLabel(entry.dayScope, t))}</small></td>
      <td class="loco-cell">
        <strong>${escapeHTML(locomotiveTitle(entry))}</strong>
        ${modelParts.length > 0 ? `<small>${modelParts.map(escapeHTML).join(" | ")}</small>` : ""}
      </td>
      <td class="control-cell">${controlParts.map(([label, value]) => `<span class="control-row"><em>${escapeHTML(label)}:</em> <strong>${escapeHTML(value)}</strong></span>`).join("")}</td>
      <td class="function-cell"><div class="function-chip-grid">${printFunctionChips(entry.functionKeys, symbols)}</div></td>
      <td class="notes-cell">${entry.notes ? escapeHTML(entry.notes) : "-"}</td>
    </tr>
  `;
  }).join("");
  const emptyColSpan = options.includeImages ? 6 : 5;
  const colGroup = options.includeImages
    ? `<colgroup><col class="col-image" /><col class="col-owner" /><col class="col-loco" /><col class="col-control" /><col class="col-functions" /><col class="col-notes" /></colgroup>`
    : `<colgroup><col class="col-owner" /><col class="col-loco" /><col class="col-control" /><col class="col-functions" /><col class="col-notes" /></colgroup>`;
  printHTMLDocument(`<!doctype html>
    <html lang="${escapeHTML(language)}">
      <head>
        <meta charset="utf-8" />
        <title> </title>
        <style>
          @page { size: A4 landscape; margin: 13mm 12mm 12mm; }
          * { box-sizing: border-box; }
          body { font-family: Arial, sans-serif; margin: 0; color: #111; }
          .print-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 20px; margin-bottom: 12px; padding-bottom: 10px; border-bottom: 2px solid #78b943; }
          .print-eyebrow { margin: 0 0 4px; color: #4e7f27; font-size: 10px; font-weight: 700; letter-spacing: .04em; text-transform: uppercase; }
          h1 { margin: 0 0 5px; font-size: 22px; line-height: 1.12; }
          p { margin: 0; color: #555; }
          .print-meta { font-size: 11px; }
          .print-logo { width: 44px; height: 44px; object-fit: contain; flex: 0 0 auto; }
          table { width: 100%; table-layout: fixed; border-collapse: collapse; font-size: 10.5px; }
          .col-image { width: 74px; }
          .col-owner { width: 102px; }
          .col-loco { width: 200px; }
          .col-control { width: 180px; }
          .col-functions { width: 178px; }
          .col-notes { width: auto; }
          th, td { border-bottom: 1px solid #d7ddd9; padding: 7px 8px; text-align: left; vertical-align: top; overflow-wrap: anywhere; }
          th { background: #eef5ee; color: #26352a; font-size: 9px; text-transform: uppercase; }
          td strong, td small { display: block; }
          td small { margin-top: 3px; color: #666; line-height: 1.35; }
          .image-cell { width: 74px; }
          .image-cell img, .image-cell span { display: grid; width: 66px; height: 46px; place-items: center; border: 1px solid #d7ddd9; border-radius: 4px; object-fit: contain; }
          .control-row { display: grid; grid-template-columns: 76px minmax(0, 1fr); gap: 5px; margin-bottom: 3px; line-height: 1.25; }
          .control-row em { color: #666; font-style: normal; }
          .function-cell { min-width: 0; }
          .function-chip-grid { display: grid; grid-template-columns: repeat(2, max-content); gap: 5px 6px; align-content: start; min-width: 0; }
          .function-chip { display: inline-flex; align-items: center; gap: 4px; min-width: 54px; padding: 3px 6px; border: 1px solid #d9e4dc; border-radius: 4px; white-space: nowrap; }
          .function-chip img { width: 14px; height: 14px; object-fit: contain; }
          .function-chip strong { display: inline; }
        </style>
      </head>
      <body>
        <header class="print-head">
          <div>
            <p class="print-eyebrow">RailKeeper Ausstellung</p>
            <h1>${escapeHTML(list.designation)}</h1>
            <p class="print-meta">${escapeHTML(t("exhibition.entriesCountWithDate", { date: formatDate(list.date, language), count: entries.length }))}</p>
          </div>
          <img class="print-logo" src="/brand/railkeeper-mark.png" alt="RailKeeper" />
        </header>
        <table>
          ${colGroup}
          <thead>
            <tr>
              ${options.includeImages ? `<th>${escapeHTML(t("exhibition.image"))}</th>` : ""}
              <th>${escapeHTML(t("exhibition.owner"))}</th>
              <th>${escapeHTML(t("exhibition.locomotiveName"))}</th>
              <th>${escapeHTML(t("exhibition.controlData"))}</th>
              <th>${escapeHTML(t("exhibition.functionKeys"))}</th>
              <th>${escapeHTML(t("exhibition.notes"))}</th>
            </tr>
          </thead>
          <tbody>${rows || `<tr><td colspan="${emptyColSpan}">${escapeHTML(t("exhibition.printEmpty"))}</td></tr>`}</tbody>
        </table>
      </body>
    </html>`);
}

export function ExhibitionView({ roles }: { roles: string[] }) {
  const { language, t } = useI18n();
  const canManageLists = hasAdmin(roles);
  const [lists, setLists] = useState<ExhibitionList[]>([]);
  const [selectedID, setSelectedID] = useState("");
  const [entries, setEntries] = useState<ExhibitionEntry[]>([]);
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [listSort, setListSort] = useState<{ key: ListSortKey; direction: SortDirection }>({ key: "date", direction: "desc" });
  const [entrySort, setEntrySort] = useState<{ key: EntrySortKey; direction: SortDirection }>({ key: "owner", direction: "asc" });
  const [listDialog, setListDialog] = useState<{ mode: "create" | "edit"; list?: ExhibitionList } | null>(null);
  const [entryDialog, setEntryDialog] = useState<{ mode: "create" | "edit"; entry?: ExhibitionEntry } | null>(null);
  const [viewDialog, setViewDialog] = useState<{ list: ExhibitionList; entries: ExhibitionEntry[] } | null>(null);
  const [printDialog, setPrintDialog] = useState<{ list: ExhibitionList; entries: ExhibitionEntry[] } | null>(null);
  const [printIncludeImages, setPrintIncludeImages] = useState(true);
  const [activeEntryTab, setActiveEntryTab] = useState<EntryTab>("general");
  const [entryFunctions, setEntryFunctions] = useState<ExhibitionFunction[]>(emptyFunctions);
  const [listForm, setListForm] = useState<ExhibitionListInput>(emptyListForm);
  const [entryForm, setEntryForm] = useState<ExhibitionEntryInput>(emptyEntryForm);
  const [symbols, setSymbols] = useState<MasterDataEntry[]>([]);
  const [masterDataOptions, setMasterDataOptions] = useState<ExhibitionMasterDataOptions>(emptyExhibitionOptions);

  const selectedList = lists.find((list) => list.id === selectedID) || null;
  const canEditEntries = Boolean(selectedList && !selectedList.locked);
  const canDeleteEntries = Boolean(canManageLists && canEditEntries);

  const sortedLists = useMemo(() => {
    return [...lists].sort((a, b) => {
      const result = sortValue(a[listSort.key]).localeCompare(sortValue(b[listSort.key]), "de");
      return listSort.direction === "asc" ? result : -result;
    });
  }, [listSort, lists]);

  const sortedEntries = useMemo(() => sortEntries(entries, entrySort), [entries, entrySort]);

  const duplicateAddress = useMemo(() => {
    const currentID = entryDialog?.entry?.id || "";
    const dcc = normalizeAddress(entryForm.decoderNumber);
    const sx = normalizeAddress(entryForm.sxAddress);
    return {
      dcc: dcc ? entries.find((entry) => entry.id !== currentID && normalizeAddress(entry.decoderNumber) === dcc) : undefined,
      sx: sx ? entries.find((entry) => entry.id !== currentID && normalizeAddress(entry.sxAddress) === sx) : undefined
    };
  }, [entries, entryDialog?.entry?.id, entryForm.decoderNumber, entryForm.sxAddress]);

  const hasDuplicateAddress = Boolean(duplicateAddress.dcc || duplicateAddress.sx);

  const load = () => {
    setLoading(true);
    setMessage("");
    api
      .exhibitionLists()
      .then((next) => {
        setLists(next);
        setSelectedID((current) => current || next[0]?.id || "");
      })
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    load();
  }, []);

  useEffect(() => {
    api.masterData("symbols", true).then(setSymbols).catch((error: Error) => setMessage(error.message));
  }, []);

  useEffect(() => {
    api.masterDataAll(true)
      .then((entriesByType) => setMasterDataOptions({
        manufacturers: entriesByType.manufacturer || [],
        epochs: entriesByType.epoch || [],
        railwayCompanies: entriesByType.railway_company || [],
        gattungen: entriesByType.vehicle_gattung || []
      }))
      .catch((error: Error) => setMessage(error.message));
  }, []);

  useEffect(() => {
    if (!selectedID) {
      setEntries([]);
      return;
    }
    api.exhibitionEntries(selectedID).then(setEntries).catch((error: Error) => setMessage(error.message));
  }, [selectedID]);

  const setListSortKey = (key: ListSortKey) => {
    setListSort((current) => ({ key, direction: current.key === key && current.direction === "asc" ? "desc" : "asc" }));
  };

  const setEntrySortKey = (key: EntrySortKey) => {
    setEntrySort((current) => ({ key, direction: current.key === key && current.direction === "asc" ? "desc" : "asc" }));
  };

  const openListDialog = (mode: "create" | "edit", list?: ExhibitionList) => {
    setListForm(list ? { designation: list.designation, date: list.date } : emptyListForm);
    setListDialog({ mode, list });
  };

  const saveList = async () => {
    if (!canManageLists) return;
    setSaving(true);
    setMessage("");
    try {
      const saved = listDialog?.mode === "edit" && listDialog.list
        ? await api.updateExhibitionList(listDialog.list.id, listForm)
        : await api.createExhibitionList(listForm);
      setListDialog(null);
      setSelectedID(saved.id);
      load();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : t("exhibition.saveListError"));
    } finally {
      setSaving(false);
    }
  };

  const deleteList = async (list: ExhibitionList) => {
    if (!canManageLists || !window.confirm(t("exhibition.deleteListConfirm", { name: list.designation }))) return;
    await api.deleteExhibitionList(list.id);
    if (selectedID === list.id) setSelectedID("");
    load();
  };

  const toggleLock = async (list: ExhibitionList) => {
    if (!canManageLists) return;
    const updated = await api.setExhibitionListLocked(list.id, !list.locked);
    setLists((current) => current.map((item) => (item.id === updated.id ? updated : item)));
  };

  const entriesForList = async (list: ExhibitionList) => {
    if (list.id === selectedID) return entries;
    return api.exhibitionEntries(list.id);
  };

  const openListView = async (list: ExhibitionList) => {
    setSelectedID(list.id);
    const nextEntries = await entriesForList(list);
    setViewDialog({ list, entries: sortEntries(nextEntries, entrySort) });
  };

  const printListByID = async (list: ExhibitionList) => {
    const nextEntries = await entriesForList(list);
    setPrintIncludeImages(true);
    setPrintDialog({ list, entries: sortEntries(nextEntries, entrySort) });
  };

  const runPrintDialog = () => {
    if (!printDialog) return;
    printList(printDialog.list, printDialog.entries, symbols, language, t, { includeImages: printIncludeImages });
    setPrintDialog(null);
  };

  const openEntryDialog = (mode: "create" | "edit", entry?: ExhibitionEntry) => {
    setActiveEntryTab("general");
    setEntryFunctions(parseFunctions(entry?.functionKeys));
    setEntryForm(entry ? {
      owner: entry.owner,
      imageUrl: entry.imageUrl || "",
      locomotiveName: entry.locomotiveName,
      gattung: entry.gattung || "",
      series: entry.series || "",
      manufacturer: entry.manufacturer || "",
      epoch: entry.epoch || "",
      railwayCompany: entry.railwayCompany || "",
      dayScope: entry.dayScope || "all",
      dtDecoder: entry.dtDecoder,
      decoderNumber: entry.decoderNumber || "",
      decoderType: entry.decoderType || "",
      adapter: entry.adapter || "",
      sxAddress: entry.sxAddress || "",
      analog: entry.analog,
      functionKeys: entry.functionKeys || "",
      notes: entry.notes || "",
      sortOrder: entry.sortOrder
    } : { ...emptyEntryForm, sortOrder: entries.length * 10 + 10 });
    setEntryDialog({ mode, entry });
  };

  const reloadEntries = async () => {
    if (!selectedID) return;
    const next = await api.exhibitionEntries(selectedID);
    setEntries(next);
    setLists((current) => current.map((list) => (list.id === selectedID ? { ...list, entryCount: next.length } : list)));
  };

  const saveEntry = async () => {
    if (!selectedID || !canEditEntries) return;
    if (hasDuplicateAddress) {
      setMessage(t("exhibition.duplicateAddressMessage"));
      return;
    }
    setSaving(true);
    setMessage("");
    const payload = {
      ...entryForm,
      dayScope: normalizeDayScopeSelection(selectedDayScopes(entryForm.dayScope)),
      functionKeys: serializeFunctions(entryFunctions)
    };
    try {
      if (entryDialog?.mode === "edit" && entryDialog.entry) {
        await api.updateExhibitionEntry(selectedID, entryDialog.entry.id, payload);
      } else {
        await api.createExhibitionEntry(selectedID, payload);
      }
      setEntryDialog(null);
      await reloadEntries();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : t("exhibition.saveEntryError"));
    } finally {
      setSaving(false);
    }
  };

  const deleteEntry = async (entry: ExhibitionEntry) => {
    if (!selectedID || !canDeleteEntries || !window.confirm(t("exhibition.deleteEntryConfirm", { name: entry.locomotiveName }))) return;
    await api.deleteExhibitionEntry(selectedID, entry.id);
    await reloadEntries();
  };

  const uploadEntryImage = async (files: FileList | null) => {
    const file = files?.[0];
    if (!file) return;
    const imageUrl = await fileToDataURL(file);
    setEntryForm((current) => ({ ...current, imageUrl }));
  };

  const updateEntryFunction = (key: string, patch: Partial<ExhibitionFunction>) => {
    setEntryFunctions((current) => current.map((item) => (item.key === key ? { ...item, ...patch } : item)));
  };

  const selectOptions = (items: MasterDataEntry[], emptyLabel = t("exhibition.noSelection")) => (
    <>
      <option value="">{emptyLabel}</option>
      {items.map((entry) => (
        <option key={entry.key} value={optionValue(entry)}>
          {entry.label}
        </option>
      ))}
    </>
  );

  const renderLocomotiveCell = (entry: ExhibitionEntry) => {
    const meta = modelMeta(entry);
    return (
      <div className="exhibition-loco-cell">
        <strong>{locomotiveTitle(entry)}</strong>
        {meta.length > 0 && <small>{meta.join(" | ")}</small>}
        {entry.notes && <small>{entry.notes}</small>}
      </div>
    );
  };

  const renderControlCell = (entry: ExhibitionEntry) => (
    <dl className="exhibition-control-stack">
      {controlRows(entry, t).map(([label, value]) => (
        <div key={label}>
          <dt>{label}:</dt>
          <dd>{value}</dd>
        </div>
      ))}
    </dl>
  );

  const renderFunctionGrid = (value?: string) => {
    const configured = configuredFunctions(value);
    if (configured.length === 0) return <span className="muted-inline">-</span>;
    return (
      <div className="exhibition-function-grid">
        {configured.map((item) => (
          <span key={item.key} className="exhibition-function-pill" title={functionDisplayName(item)} aria-label={`${item.key} ${functionDisplayName(item)}`}>
            <span className="exhibition-function-icon">
              {functionSymbolIcon(item.symbolKey, item.type, functionSymbolMetadata(symbols, item.symbolKey))}
            </span>
            <strong>{item.key}</strong>
          </span>
        ))}
      </div>
    );
  };

  return (
    <>
      <section className="inventory-head">
        <div>
          <p className="eyebrow">{t("exhibition.eyebrow")}</p>
          <h1>{t("exhibition.title")}</h1>
          <p>{t("exhibition.subtitle")}</p>
        </div>
        {canManageLists && (
          <button type="button" className="primary-button new-vehicle-button" onClick={() => openListDialog("create")}>
            <Plus size={16} aria-hidden="true" />
            {t("exhibition.newList")}
          </button>
        )}
      </section>

      {message && <p className="form-message">{message}</p>}

      <section className="exhibition-layout">
        <article className="panel exhibition-list-panel">
          <div className="inventory-list-head">
            <div>
              <h2>{t("exhibition.lists")}</h2>
              <p>{loading ? t("exhibition.loading") : t("exhibition.listCount", { count: lists.length })}</p>
            </div>
          </div>
          <div className="table-wrap">
            <table className="inventory-table exhibition-table">
              <thead>
                <tr>
                  <th><button type="button" className={listSort.key === "designation" ? "sort-button active" : "sort-button"} onClick={() => setListSortKey("designation")}>{t("exhibition.designation")}</button></th>
                  <th><button type="button" className={listSort.key === "date" ? "sort-button active" : "sort-button"} onClick={() => setListSortKey("date")}>{t("exhibition.date")}</button></th>
                  <th><button type="button" className={listSort.key === "entryCount" ? "sort-button active" : "sort-button"} onClick={() => setListSortKey("entryCount")}>{t("exhibition.entries")}</button></th>
                  <th><button type="button" className={listSort.key === "locked" ? "sort-button active" : "sort-button"} onClick={() => setListSortKey("locked")}>{t("exhibition.status")}</button></th>
                  <th>{t("exhibition.actions")}</th>
                </tr>
              </thead>
              <tbody>
                {sortedLists.map((list) => (
                  <tr key={list.id} className={selectedID === list.id ? "selected-row" : ""} onClick={() => setSelectedID(list.id)}>
                    <td><strong>{list.designation}</strong></td>
                    <td>{formatDate(list.date, language)}</td>
                    <td>{list.entryCount}</td>
                    <td><span className={list.locked ? "settings-pill muted" : "settings-pill active"}>{list.locked ? t("exhibition.locked") : t("exhibition.open")}</span></td>
                    <td>
                      <div className="table-actions">
                        <button type="button" className="icon-button" onClick={(event) => { event.stopPropagation(); openListView(list); }} aria-label={t("exhibition.view")} title={t("exhibition.view")}><Eye size={15} /></button>
                        {canManageLists && <button type="button" className="icon-button" onClick={(event) => { event.stopPropagation(); openListDialog("edit", list); }} aria-label={t("exhibition.edit")} title={t("exhibition.edit")}><Edit3 size={15} /></button>}
                        <button type="button" className="icon-button" onClick={(event) => { event.stopPropagation(); printListByID(list); }} aria-label={t("exhibition.print")} title={t("exhibition.print")}><Printer size={15} /></button>
                        {canManageLists && <button type="button" className="icon-button" onClick={(event) => { event.stopPropagation(); toggleLock(list); }} aria-label={list.locked ? t("exhibition.unlock") : t("exhibition.lock")} title={list.locked ? t("exhibition.unlock") : t("exhibition.lock")}>{list.locked ? <LockOpen size={15} /> : <Lock size={15} />}</button>}
                        {canManageLists && <button type="button" className="icon-button danger" onClick={(event) => { event.stopPropagation(); deleteList(list); }} aria-label={t("exhibition.delete")} title={t("exhibition.delete")}><Trash2 size={15} /></button>}
                      </div>
                    </td>
                  </tr>
                ))}
                {sortedLists.length === 0 && (
                  <tr><td colSpan={5}>{t("exhibition.noLists")}</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </article>

        <article className="panel exhibition-entry-panel">
          <div className="inventory-list-head">
            <div>
              <h2>{selectedList ? selectedList.designation : t("exhibition.entries")}</h2>
              <p>{selectedList ? t("exhibition.entriesCountWithDate", { date: formatDate(selectedList.date, language), count: entries.length }) : t("exhibition.selectList")}</p>
            </div>
            <div className="table-actions">
              {selectedList && <button type="button" className="icon-button" onClick={() => { setPrintIncludeImages(true); setPrintDialog({ list: selectedList, entries: sortedEntries }); }} aria-label={t("exhibition.printList")} title={t("exhibition.printList")}><Printer size={15} /></button>}
              {selectedList && <button type="button" className="primary-button" onClick={() => openEntryDialog("create")} disabled={!canEditEntries}>{t("exhibition.entry")}</button>}
            </div>
          </div>
          <div className="table-wrap">
            <table className="inventory-table exhibition-table">
              <thead>
                <tr>
                  <th>{t("exhibition.image")}</th>
                  <th><button type="button" className={entrySort.key === "owner" ? "sort-button active" : "sort-button"} onClick={() => setEntrySortKey("owner")}>{t("exhibition.owner")}</button></th>
                  <th><button type="button" className={entrySort.key === "locomotiveName" ? "sort-button active" : "sort-button"} onClick={() => setEntrySortKey("locomotiveName")}>{t("exhibition.locomotiveName")}</button></th>
                  <th><button type="button" className={entrySort.key === "decoderNumber" ? "sort-button active" : "sort-button"} onClick={() => setEntrySortKey("decoderNumber")}>{t("exhibition.controlData")}</button></th>
                  <th><button type="button" className={entrySort.key === "functionKeys" ? "sort-button active" : "sort-button"} onClick={() => setEntrySortKey("functionKeys")}>{t("exhibition.functionKeys")}</button></th>
                  <th>{t("exhibition.actions")}</th>
                </tr>
              </thead>
              <tbody>
                {sortedEntries.map((entry) => (
                  <tr key={entry.id}>
                    <td>{entry.imageUrl ? <img className="exhibition-thumb" src={entry.imageUrl} alt="" /> : <span className="image-placeholder mini">-</span>}</td>
                    <td><strong>{entry.owner}</strong><small>{dayScopeLabel(entry.dayScope, t)}</small></td>
                    <td>{renderLocomotiveCell(entry)}</td>
                    <td>{renderControlCell(entry)}</td>
                    <td>{renderFunctionGrid(entry.functionKeys)}</td>
                    <td>
                      <div className="table-actions">
                        <button type="button" className="icon-button" onClick={() => openEntryDialog("edit", entry)} disabled={!canEditEntries} aria-label={t("exhibition.edit")} title={t("exhibition.edit")}><Edit3 size={15} /></button>
                        {canManageLists && <button type="button" className="icon-button danger" onClick={() => deleteEntry(entry)} disabled={!canDeleteEntries} aria-label={t("exhibition.delete")} title={t("exhibition.delete")}><Trash2 size={15} /></button>}
                      </div>
                    </td>
                  </tr>
                ))}
                {selectedList && sortedEntries.length === 0 && (
                  <tr><td colSpan={6}>{t("exhibition.noEntries")}</td></tr>
                )}
                {!selectedList && (
                  <tr><td colSpan={6}>{t("exhibition.noList")}</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </article>
      </section>

      {listDialog && (
        <div className="modal-layer">
          <form className="vehicle-modal compact-modal" onSubmit={(event) => { event.preventDefault(); saveList(); }}>
            <div className="modal-head">
              <h2>{listDialog.mode === "edit" ? t("exhibition.listEdit") : t("exhibition.listCreate")}</h2>
              <button type="button" className="icon-button" onClick={() => setListDialog(null)} aria-label={t("exhibition.close")}>×</button>
            </div>
            <div className="modal-body simple-form">
              <label>
                <span>{t("exhibition.designation")}</span>
                <input value={listForm.designation} onChange={(event) => setListForm({ ...listForm, designation: event.target.value })} required />
              </label>
              <label>
                <span>{t("exhibition.date")}</span>
                <input type="date" value={listForm.date} onChange={(event) => setListForm({ ...listForm, date: event.target.value })} required />
              </label>
            </div>
            <div className="modal-actions">
              <button type="button" className="secondary-button" onClick={() => setListDialog(null)}>{t("exhibition.cancel")}</button>
              <button type="submit" className="primary-button" disabled={saving}>{t("exhibition.save")}</button>
            </div>
          </form>
        </div>
      )}

      {viewDialog && (
        <div className="modal-layer">
          <section className="vehicle-modal exhibition-view-modal">
            <div className="modal-head">
              <div>
                <h2>{viewDialog.list.designation}</h2>
                <p>{t("exhibition.entriesCountWithDate", { date: formatDate(viewDialog.list.date, language), count: viewDialog.entries.length })}</p>
              </div>
              <button type="button" className="icon-button" onClick={() => setViewDialog(null)} aria-label={t("exhibition.close")}>×</button>
            </div>
            <div className="modal-body">
              <div className="table-wrap">
                <table className="inventory-table exhibition-table">
                  <thead>
                    <tr>
                      <th>{t("exhibition.image")}</th>
                      <th>{t("exhibition.owner")}</th>
                      <th>{t("exhibition.locomotiveName")}</th>
                      <th>{t("exhibition.controlData")}</th>
                      <th>{t("exhibition.functionKeys")}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {viewDialog.entries.map((entry) => (
                      <tr key={entry.id}>
                        <td>{entry.imageUrl ? <img className="exhibition-thumb" src={entry.imageUrl} alt="" /> : <span className="image-placeholder mini">-</span>}</td>
                        <td><strong>{entry.owner}</strong><small>{dayScopeLabel(entry.dayScope, t)}</small></td>
                        <td>{renderLocomotiveCell(entry)}</td>
                        <td>{renderControlCell(entry)}</td>
                        <td>{renderFunctionGrid(entry.functionKeys)}</td>
                      </tr>
                    ))}
                    {viewDialog.entries.length === 0 && <tr><td colSpan={5}>{t("exhibition.noEntriesShort")}</td></tr>}
                  </tbody>
                </table>
              </div>
            </div>
            <div className="modal-actions">
              <button type="button" className="secondary-button" onClick={() => { setPrintIncludeImages(true); setPrintDialog({ list: viewDialog.list, entries: viewDialog.entries }); }}>{t("exhibition.print")}</button>
              <button type="button" className="primary-button" onClick={() => setViewDialog(null)}>{t("exhibition.close")}</button>
            </div>
          </section>
        </div>
      )}

      {printDialog && (
        <div className="modal-layer">
          <section className="vehicle-modal compact-modal">
            <div className="modal-head">
              <div>
                <h2>{t("exhibition.printDialogTitle")}</h2>
                <p>{printDialog.list.designation}</p>
              </div>
              <button type="button" className="icon-button" onClick={() => setPrintDialog(null)} aria-label={t("exhibition.close")}>×</button>
            </div>
            <div className="modal-body simple-form">
              <label className="checkbox-line">
                <input type="checkbox" checked={printIncludeImages} onChange={(event) => setPrintIncludeImages(event.target.checked)} />
                {t("exhibition.printIncludeImages")}
              </label>
            </div>
            <div className="modal-actions">
              <button type="button" className="secondary-button" onClick={() => setPrintDialog(null)}>{t("exhibition.cancel")}</button>
              <button type="button" className="primary-button" onClick={runPrintDialog}>{t("exhibition.print")}</button>
            </div>
          </section>
        </div>
      )}

      {entryDialog && (
        <div className="modal-layer">
          <form className="vehicle-modal exhibition-entry-modal" onSubmit={(event) => { event.preventDefault(); saveEntry(); }}>
            <div className="modal-head">
              <h2>{entryDialog.mode === "edit" ? t("exhibition.entryEdit") : t("exhibition.entryCreate")}</h2>
              <button type="button" className="icon-button" onClick={() => setEntryDialog(null)} aria-label={t("exhibition.close")}>×</button>
            </div>
            <div className="modal-tabs exhibition-entry-tabs" role="tablist" aria-label={t("exhibition.entryTabs")}>
              <button type="button" className={activeEntryTab === "general" ? "active" : ""} onClick={() => setActiveEntryTab("general")}>{t("exhibition.tab.general")}</button>
              <button type="button" className={activeEntryTab === "images" ? "active" : ""} onClick={() => setActiveEntryTab("images")}>{t("exhibition.tab.images")}</button>
              <button type="button" className={activeEntryTab === "functions" ? "active" : ""} onClick={() => setActiveEntryTab("functions")}>{t("exhibition.tab.functions")}</button>
            </div>
            <div className="modal-body">
              {activeEntryTab === "general" && (
                <div className="exhibition-entry-form">
                  <section className="exhibition-entry-section exhibition-entry-section-wide">
                    <h3>{t("exhibition.basicData")}</h3>
                    <div className="exhibition-day-selector" role="group" aria-label={t("exhibition.dayScope")}>
                      {dayScopes.map((scope) => (
                        <button
                          key={scope}
                          type="button"
                          className={isDayScopeActive(entryForm.dayScope, scope) ? "active" : ""}
                          aria-pressed={isDayScopeActive(entryForm.dayScope, scope)}
                          onClick={() => setEntryForm((current) => ({ ...current, dayScope: toggleDayScope(current.dayScope, scope) }))}
                        >
                          {dayScopeLabel(scope, t)}
                        </button>
                      ))}
                    </div>
                    <div className="exhibition-entry-grid">
                      <label>
                        <span>{t("exhibition.owner")}</span>
                        <input value={entryForm.owner} onChange={(event) => setEntryForm({ ...entryForm, owner: event.target.value })} required />
                      </label>
                      <label>
                        <span>{t("exhibition.locomotiveName")}</span>
                        <input value={entryForm.locomotiveName} onChange={(event) => setEntryForm({ ...entryForm, locomotiveName: event.target.value })} required />
                      </label>
                      <label>
                        <span>{t("exhibition.manufacturer")}</span>
                        <AppSelect value={entryForm.manufacturer || ""} onChange={(event) => setEntryForm({ ...entryForm, manufacturer: event.target.value })}>
                          {selectOptions(masterDataOptions.manufacturers)}
                        </AppSelect>
                      </label>
                      <label>
                        <span>{t("exhibition.series")}</span>
                        <input value={entryForm.series || ""} onChange={(event) => setEntryForm({ ...entryForm, series: event.target.value })} />
                      </label>
                      <label>
                        <span>{t("exhibition.gattung")}</span>
                        <AppSelect value={entryForm.gattung || ""} onChange={(event) => setEntryForm({ ...entryForm, gattung: event.target.value })}>
                          {selectOptions(masterDataOptions.gattungen)}
                        </AppSelect>
                      </label>
                      <label>
                        <span>{t("exhibition.epoch")}</span>
                        <AppSelect value={entryForm.epoch || ""} onChange={(event) => setEntryForm({ ...entryForm, epoch: event.target.value })}>
                          {selectOptions(masterDataOptions.epochs)}
                        </AppSelect>
                      </label>
                      <label>
                        <span>{t("exhibition.railwayCompany")}</span>
                        <AppSelect value={entryForm.railwayCompany || ""} onChange={(event) => setEntryForm({ ...entryForm, railwayCompany: event.target.value })}>
                          {selectOptions(masterDataOptions.railwayCompanies)}
                        </AppSelect>
                      </label>
                    </div>
                  </section>
                  <section className="exhibition-entry-section">
                    <h3>{t("exhibition.controlData")}</h3>
                    <div className="exhibition-entry-grid">
                      <label>
                        <span>{t("exhibition.decoderType")}</span>
                        <input value={entryForm.decoderType || ""} onChange={(event) => setEntryForm({ ...entryForm, decoderType: event.target.value })} />
                      </label>
                      <label>
                        <span>{t("exhibition.adapter")}</span>
                        <AppSelect value={entryForm.adapter || ""} onChange={(event) => setEntryForm({ ...entryForm, adapter: event.target.value })}>
                          <option value="">{t("exhibition.noSelection")}</option>
                          {adapterOptions.map((option) => <option key={option} value={option}>{option}</option>)}
                        </AppSelect>
                      </label>
                      <label>
                        <span>{t("exhibition.dccAddress")}</span>
                        <input value={entryForm.decoderNumber || ""} onChange={(event) => setEntryForm({ ...entryForm, decoderNumber: event.target.value })} />
                        {duplicateAddress.dcc && <small className="field-warning">{t("exhibition.duplicateDccAddress", { name: duplicateAddress.dcc.locomotiveName })}</small>}
                      </label>
                      <label>
                        <span>{t("exhibition.sxAddress")}</span>
                        <input value={entryForm.sxAddress || ""} onChange={(event) => setEntryForm({ ...entryForm, sxAddress: event.target.value })} />
                        {duplicateAddress.sx && <small className="field-warning">{t("exhibition.duplicateSxAddress", { name: duplicateAddress.sx.locomotiveName })}</small>}
                      </label>
                    </div>
                    <label className="switch-label exhibition-entry-switch">
                      {t("exhibition.analog")}
                      <span className="switch-field">
                        <input type="checkbox" checked={Boolean(entryForm.analog)} onChange={(event) => setEntryForm({ ...entryForm, analog: event.target.checked })} />
                        <span />
                      </span>
                    </label>
                  </section>
                  <section className="exhibition-entry-section">
                    <h3>{t("exhibition.notesSection")}</h3>
                    <label>
                      <span>{t("exhibition.notes")}</span>
                      <textarea value={entryForm.notes || ""} onChange={(event) => setEntryForm({ ...entryForm, notes: event.target.value })} rows={5} />
                    </label>
                  </section>
                </div>
              )}
              {activeEntryTab === "images" && (
                <section className="exhibition-image-tab">
                  <div className="upload-head">
                    <div>
                      <h3>{t("exhibition.images")}</h3>
                      <p>{t("exhibition.imagesHelp")}</p>
                    </div>
                  </div>
                  <div className="exhibition-image-editor">
                    {entryForm.imageUrl ? (
                      <img src={entryForm.imageUrl} alt="" />
                    ) : (
                      <div className="image-placeholder large"><ImageIcon size={24} aria-hidden="true" />{t("exhibition.noPreview")}</div>
                    )}
                    <div className="exhibition-image-source">
                      <label>
                        <span>{t("exhibition.imageSource")}</span>
                        <input value={entryForm.imageUrl || ""} onChange={(event) => setEntryForm({ ...entryForm, imageUrl: event.target.value })} placeholder="https://..." />
                      </label>
                      <div className="exhibition-image-source-actions">
                        <label className="secondary-button">
                          <Upload size={16} aria-hidden="true" />
                          {t("exhibition.uploadImage")}
                          <input type="file" accept="image/png,image/jpeg,image/webp" className="visually-hidden" onChange={(event) => uploadEntryImage(event.target.files)} />
                        </label>
                        {entryForm.imageUrl && <button type="button" className="secondary-button" onClick={() => setEntryForm({ ...entryForm, imageUrl: "" })}>{t("exhibition.removeImage")}</button>}
                      </div>
                    </div>
                  </div>
                </section>
              )}
              {activeEntryTab === "functions" && (
                <section className="functions-tab exhibition-functions-tab">
                  <div className="function-list">
                    <div className="function-toolbar">
                      <div className="function-summary">
                        <span><strong>{entryFunctions.filter(isConfiguredFunction).length}</strong> {t("exhibition.assigned")}</span>
                        <span><strong>{entryFunctions.filter((item) => item.type === "sound" && isConfiguredFunction(item)).length}</strong> {t("exhibition.sound")}</span>
                        <span><strong>{entryFunctions.filter((item) => item.type === "licht" && isConfiguredFunction(item)).length}</strong> {t("exhibition.light")}</span>
                      </div>
                    </div>
                    {entryFunctions.map((item) => (
                      <article key={item.key} className={isConfiguredFunction(item) ? "function-row exhibition-function-row persisted" : "function-row exhibition-function-row"}>
                        <strong className="function-key">
                          {functionSymbolIcon(item.symbolKey, item.type, functionSymbolMetadata(symbols, item.symbolKey))}
                          {item.key}
                        </strong>
                        <input value={item.name} onChange={(event) => updateEntryFunction(item.key, { name: event.target.value })} placeholder={t("exhibition.functionName")} aria-label={t("exhibition.functionNameAria", { key: item.key })} />
                        <FunctionSymbolPicker
                          value={item.symbolKey || ""}
                          functionType={item.type}
                          symbols={symbols}
                          label={t("exhibition.functionSymbolAria", { key: item.key })}
                          onChange={(symbolKey, symbolLabel) => updateEntryFunction(item.key, { symbolKey, name: symbolLabel || item.name })}
                        />
                        <AppSelect value={item.type} onChange={(event) => updateEntryFunction(item.key, { type: event.target.value })} aria-label={t("exhibition.functionTypeAria", { key: item.key })}>
                          {functionTypes.map((type) => <option key={type} value={type}>{type}</option>)}
                        </AppSelect>
                      </article>
                    ))}
                  </div>
                </section>
              )}
            </div>
            <div className="modal-actions">
              <button type="button" className="secondary-button" onClick={() => setEntryDialog(null)}>{t("exhibition.cancel")}</button>
              <button type="submit" className="primary-button" disabled={saving || hasDuplicateAddress}>{t("exhibition.save")}</button>
            </div>
          </form>
        </div>
      )}
    </>
  );
}
