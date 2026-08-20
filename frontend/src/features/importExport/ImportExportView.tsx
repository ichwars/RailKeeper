import { ChangeEvent, DragEvent, Fragment, useEffect, useMemo, useRef, useState } from "react";
import {
  Boxes,
  Check,
  CheckCircle2,
  Circle,
  ClipboardCheck,
  Database,
  Download,
  FileJson,
  FileSpreadsheet,
  FileText,
  History,
  Info,
  MoreVertical,
  Printer,
  RotateCcw,
  Save,
  TrainFront,
  Upload,
  XCircle
} from "lucide-react";
import {
  api,
  CreateVehicleRequest,
  type AccessoryArticleListItem,
  type ExhibitionList,
  Vehicle
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import {
  ColumnMapping,
  defaultColumnMappings,
  detectDelimiter,
  displayImportValue,
  downloadText,
  getImportChanges,
  ImportRow,
  ImportTablePreview,
  importRowsFromTable,
  mergeImportedVehicle,
  parseDelimited,
  parseXMLImport,
  printInventory,
  VehicleImportField,
  vehicleImportFields,
  vehiclesToCSV
} from "./importExportHelpers";
import { useDataTransferWorkspace } from "./useDataTransferWorkspace";

type DataArea = "vehicles" | "accessories" | "exhibitionLists";
type ExportFormat = "csv" | "json" | "print";
type TransferResult = "success" | "warning" | "failed";

type TransferEntry = {
  id: string;
  timestamp: Date;
  operation: "import" | "export";
  areas: DataArea[];
  fileName: string;
  result: TransferResult;
  format?: ExportFormat;
};

const csvFieldLabelKeys: Partial<Record<VehicleImportField, string>> = {
  decoderType: "importExport.csvField.decoderType",
  acquisitionType: "importExport.csvField.acquisitionType",
  acquiredFrom: "importExport.csvField.acquiredFrom",
  purchasePrice: "importExport.csvField.purchasePrice",
  purchaseDate: "importExport.csvField.purchaseDate",
  storageLocation: "importExport.csvField.storageLocation",
  storageDetails: "importExport.csvField.storageDetails",
  condition: "importExport.csvField.condition",
  conditionDetails: "importExport.csvField.conditionDetails",
  packaging: "importExport.csvField.packaging"
};

const areaOrder: DataArea[] = ["vehicles", "accessories", "exhibitionLists"];

function csvEscape(value: unknown) {
  const text = Array.isArray(value) ? value.join(", ") : String(value ?? "");
  return /[;"\r\n]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text;
}

function accessoriesToCSV(items: AccessoryArticleListItem[]) {
  const rows = items.map((item) => [
    item.inventoryNumber,
    item.manufacturer,
    item.articleNumber,
    item.name,
    item.articleType,
    item.subtype,
    item.gauges,
    item.owned,
    item.available,
    item.reserved,
    item.installed,
    item.locationNames
  ]);
  return [
    ["Inventarnummer", "Hersteller", "Artikelnummer", "Bezeichnung", "Artikeltyp", "Untertyp", "Spurweiten", "Bestand", "Verfügbar", "Reserviert", "Verbaut", "Lagerorte"],
    ...rows
  ].map((row) => row.map(csvEscape).join(";")).join("\n");
}

function fileBaseName(format: ExportFormat, areas: DataArea[]) {
  const scope = areas.length === 1
    ? areas[0] === "vehicles" ? "fahrzeuge" : areas[0] === "accessories" ? "zubehoer" : "ausstellungslisten"
    : "bestand";
  const extension = format === "print" ? "pdf" : format;
  return `railkeeper-${scope}.${extension}`;
}

export function ImportExportView({ roles }: { roles: string[] }) {
  useDataTransferWorkspace(roles);
  const { language, t } = useI18n();
  const reviewRef = useRef<HTMLElement>(null);
  const historyRef = useRef<HTMLElement>(null);
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [accessories, setAccessories] = useState<AccessoryArticleListItem[]>([]);
  const [exhibitionLists, setExhibitionLists] = useState<ExhibitionList[]>([]);
  const [rows, setRows] = useState<ImportRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [importTable, setImportTable] = useState<ImportTablePreview | null>(null);
  const [importArea, setImportArea] = useState<DataArea>("vehicles");
  const [stagedFile, setStagedFile] = useState<File | null>(null);
  const [exportAreas, setExportAreas] = useState<Record<DataArea, boolean>>({
    vehicles: true,
    accessories: false,
    exhibitionLists: false
  });
  const [exportFormat, setExportFormat] = useState<ExportFormat>("csv");
  const [history, setHistory] = useState<TransferEntry[]>([]);

  const fieldLabel = (key: VehicleImportField) => t(`vehicle.field.${key}`);
  const csvFieldLabel = (key: VehicleImportField) => {
    const translationKey = csvFieldLabelKeys[key];
    return translationKey ? t(translationKey) : fieldLabel(key);
  };
  const issueLabels = {
    missingManufacturer: t("importExport.issue.missingManufacturer"),
    missingName: t("importExport.issue.missingName"),
    missingGauge: t("importExport.issue.missingGauge"),
    missingCategory: t("importExport.issue.missingCategory"),
    missingGattung: t("importExport.issue.missingGattung"),
    invalidMaximumSpeed: t("importExport.issue.invalidMaximumSpeed"),
    invalidHomeBase: t("importExport.issue.invalidHomeBase"),
    duplicate: t("importExport.issue.duplicate"),
    ecosMatched: t("importExport.ecos.matched")
  };

  useEffect(() => {
    let cancelled = false;
    const loadTransferData = async () => {
      try {
        const [vehicles, accessories, exhibitionLists] = await Promise.all([
          api.vehicles(),
          api.accessoryArticles(),
          api.exhibitionLists()
        ]);
        if (cancelled) return;
        setVehicles(vehicles);
        setAccessories(accessories.items);
        setExhibitionLists(exhibitionLists);
      } catch (error) {
        if (!cancelled) setMessage(error instanceof Error ? error.message : t("importExport.loadPartial"));
      } finally {
        if (!cancelled) setLoading(false);
      }
    };
    void loadTransferData();
    return () => { cancelled = true; };
  }, [t]);

  const areaCounts: Record<DataArea, number> = {
    vehicles: vehicles.length,
    accessories: accessories.length,
    exhibitionLists: exhibitionLists.length
  };
  const importSummary = useMemo(() => ({
    total: rows.length,
    selected: rows.filter((row) => row.selected && row.status !== "saved").length,
    errors: rows.filter((row) => row.status === "error").length,
    updates: rows.filter((row) => row.mode === "update" && row.status !== "saved").length,
    saved: rows.filter((row) => row.status === "saved").length
  }), [rows]);
  const mappingSummary = useMemo(() => {
    if (!importTable) return { mapped: 0, unmapped: 0 };
    const visibleMappings = importTable.mappings.filter((mapping) => mapping.header.trim());
    return {
      mapped: visibleMappings.filter((mapping) => mapping.key).length,
      unmapped: visibleMappings.filter((mapping) => !mapping.key).length
    };
  }, [importTable]);
  const selectedExportAreas = areaOrder.filter((area) => exportAreas[area]);
  const selectedExportCount = selectedExportAreas.reduce((total, area) => total + areaCounts[area], 0);
  const formatAvailability: Record<ExportFormat, boolean> = {
    csv: selectedExportAreas.length === 1
      && (selectedExportAreas[0] === "vehicles" || selectedExportAreas[0] === "accessories"),
    json: selectedExportAreas.length > 0,
    print: selectedExportAreas.length === 1 && selectedExportAreas[0] === "vehicles"
  };
  const exportValid = selectedExportAreas.length > 0 && formatAvailability[exportFormat] && selectedExportCount > 0;

  const addHistory = (entry: Omit<TransferEntry, "id" | "timestamp">) => {
    setHistory((current) => [{ ...entry, id: crypto.randomUUID(), timestamp: new Date() }, ...current].slice(0, 8));
  };

  const resetImport = () => {
    setRows([]);
    setImportTable(null);
    setStagedFile(null);
    setMessage("");
  };

  const selectImportArea = (area: DataArea) => {
    if (area === importArea) return;
    if (stagedFile && !window.confirm(t("importExport.areaChangeConfirm"))) return;
    resetImport();
    setImportArea(area);
  };

  const loadImportTable = (table: string[][], fileName: string) => {
    if (table.length === 0) {
      setImportTable(null);
      setRows([]);
      setMessage(t("importExport.error.emptyFile"));
      return;
    }
    const mappings = defaultColumnMappings(table);
    const importedRows = importRowsFromTable(table, vehicles, mappings, issueLabels);
    const unmapped = mappings.filter((mapping) => !mapping.key && mapping.header.trim()).length;
    setImportTable({ fileName, table, mappings });
    setRows(importedRows);
    setMessage(unmapped > 0 ? t("importExport.message.unmapped", { count: unmapped }) : "");
  };

  const readVehicleFile = async (file: File) => {
    setPreviewLoading(true);
    setMessage("");
    try {
      const extension = file.name.split(".").pop()?.toLowerCase() || "";
      const text = await file.text();
      if (extension === "json") {
        const parsed = JSON.parse(text) as Vehicle[] | { vehicles?: Vehicle[] };
        const source = Array.isArray(parsed) ? parsed : parsed.vehicles || [];
        const table = [
          [
            fieldLabel("inventoryNumber"), fieldLabel("manufacturer"), fieldLabel("articleNumber"),
            fieldLabel("name"), fieldLabel("gauge"), fieldLabel("epoch"), fieldLabel("railwayCompany"),
            fieldLabel("category"), fieldLabel("gattung"), fieldLabel("maximumSpeedKmh"),
            fieldLabel("homeBase"), fieldLabel("digital"), fieldLabel("digitalDecoderNumber"), fieldLabel("listPrice")
          ],
          ...source.map((vehicle) => [
            vehicle.inventoryNumber, vehicle.manufacturer, vehicle.articleNumber || "", vehicle.name,
            vehicle.gauge, vehicle.epoch || "", vehicle.railwayCompany || "", vehicle.category || "",
            vehicle.gattung || "", vehicle.maximumSpeedKmh ? String(vehicle.maximumSpeedKmh) : "",
            vehicle.homeBase || "", vehicle.digital ? t("common.yes") : t("common.no"),
            vehicle.digitalDecoderNumber || "", vehicle.listPrice || ""
          ])
        ];
        loadImportTable(table, file.name);
        return;
      }
      if (extension === "xml") {
        const table = parseXMLImport(text);
        if (table.length) loadImportTable(table, file.name);
        else setMessage(t("importExport.error.emptyTable"));
        return;
      }
      const delimiter = extension === "tsv" ? "\t" : detectDelimiter(text);
      loadImportTable(parseDelimited(text, delimiter), file.name);
    } catch (error) {
      setRows([]);
      setImportTable(null);
      setMessage(error instanceof Error ? error.message : t("importExport.error.emptyTable"));
    } finally {
      setPreviewLoading(false);
    }
  };

  const stageFile = (file: File) => {
    setStagedFile(file);
    if (importArea !== "vehicles") {
      setRows([]);
      setImportTable(null);
      setMessage(t("importExport.areaPreviewPending", { area: t(`importExport.area.${importArea}`) }));
      return;
    }
    void readVehicleFile(file);
  };

  const handleFile = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) stageFile(file);
    event.target.value = "";
  };

  const handleDrop = (event: DragEvent<HTMLLabelElement>) => {
    event.preventDefault();
    const file = event.dataTransfer.files?.[0];
    if (file) stageFile(file);
  };

  const setColumnMapping = (columnIndex: number, key: VehicleImportField | "") => {
    if (!importTable) return;
    const mappings: ColumnMapping[] = importTable.mappings.map((mapping) => {
      if (mapping.index === columnIndex) return { ...mapping, key };
      return key && mapping.key === key ? { ...mapping, key: "" } : mapping;
    });
    setImportTable({ ...importTable, mappings });
    setRows(importRowsFromTable(importTable.table, vehicles, mappings, issueLabels));
  };

  const updateRow = (rowID: string, patch: Partial<ImportRow["vehicle"]>) => {
    setRows((current) => current.map((row) => {
      if (row.id !== rowID) return row;
      const vehicle = { ...row.vehicle, ...patch };
      const duplicate = vehicle.inventoryNumber
        ? vehicles.find((existing) => existing.inventoryNumber.toLowerCase() === vehicle.inventoryNumber?.toLowerCase())
        : undefined;
      const issues: string[] = [];
      if (duplicate && duplicate.id !== row.duplicateVehicleId) issues.push(t("importExport.issue.duplicate"));
      if (!vehicle.manufacturer) issues.push(t("importExport.issue.missingManufacturer"));
      if (!vehicle.name) issues.push(t("importExport.issue.missingName"));
      if (!vehicle.gauge) issues.push(t("importExport.issue.missingGauge"));
      if (!vehicle.category) issues.push(t("importExport.issue.missingCategory"));
      if (!vehicle.gattung) issues.push(t("importExport.issue.missingGattung"));
      return {
        ...row,
        vehicle,
        importedKeys: Array.from(new Set([...row.importedKeys, ...Object.keys(patch) as (keyof CreateVehicleRequest)[]])),
        issues,
        status: issues.length ? "error" : row.mode === "update" ? "warning" : "ok"
      };
    }));
  };

  const setRowSelected = (rowID: string, selected: boolean) => {
    setRows((current) => current.map((row) => row.id === rowID ? { ...row, selected } : row));
  };

  const setRowMode = (rowID: string, mode: ImportRow["mode"]) => {
    setRows((current) => current.map((row) => {
      if (row.id !== rowID) return row;
      if (mode === "create" && row.duplicateVehicleId) {
        return { ...row, mode, selected: false, status: "error", issues: [t("importExport.issue.inventoryExists")] };
      }
      if (mode === "update" && row.duplicateVehicleId) {
        return { ...row, mode, status: "warning", issues: [t("importExport.issue.duplicate")] };
      }
      return { ...row, mode };
    }));
  };

  const saveSelected = async () => {
    if (!window.confirm(t("importExport.review.confirmSave", { count: importSummary.selected }))) return;
    setSaving(true);
    setMessage("");
    let failed = 0;
    for (const row of rows) {
      if (!row.selected || row.status === "saved" || row.status === "error") continue;
      try {
        const existing = row.duplicateVehicleId
          ? vehicles.find((vehicle) => vehicle.id === row.duplicateVehicleId)
          : undefined;
        const saved = row.mode === "update" && existing
          ? await api.updateVehicle(existing.id, mergeImportedVehicle(existing, row.vehicle, row.importedKeys))
          : await api.createVehicle(row.vehicle);
        setVehicles((current) => row.mode === "update"
          ? current.map((vehicle) => vehicle.id === saved.id ? saved : vehicle)
          : [...current, saved]);
        setRows((current) => current.map((item) => item.id === row.id
          ? { ...item, selected: false, status: "saved", issues: [] }
          : item));
      } catch (error) {
        failed += 1;
        const errorMessage = error instanceof Error ? error.message : t("importExport.error.importFailed");
        setRows((current) => current.map((item) => item.id === row.id
          ? { ...item, status: "error", issues: [errorMessage] }
          : item));
      }
    }
    addHistory({
      operation: "import",
      areas: ["vehicles"],
      fileName: stagedFile?.name || importTable?.fileName || t("importExport.history.unknownFile"),
      result: failed > 0 ? "warning" : "success"
    });
    setSaving(false);
  };

  const toggleExportArea = (area: DataArea) => {
    setExportAreas((current) => ({ ...current, [area]: !current[area] }));
  };

  const chooseFormat = (format: ExportFormat) => {
    if (formatAvailability[format]) setExportFormat(format);
  };

  const createExport = () => {
    if (!exportValid) return;
    const fileName = fileBaseName(exportFormat, selectedExportAreas);
    if (exportFormat === "csv") {
      if (selectedExportAreas[0] === "vehicles") {
        downloadText(fileName, `\uFEFF${vehiclesToCSV(vehicles, csvFieldLabel, t("common.yes"), t("common.no"))}`, "text/csv;charset=utf-8");
      } else {
        downloadText(fileName, `\uFEFF${accessoriesToCSV(accessories)}`, "text/csv;charset=utf-8");
      }
    } else if (exportFormat === "print") {
      printInventory(vehicles, fieldLabel, language, t);
    } else {
      downloadText(fileName, JSON.stringify({
        format: "railkeeper-export",
        version: 1,
        createdAt: new Date().toISOString(),
        ...(exportAreas.vehicles ? { vehicles } : {}),
        ...(exportAreas.accessories ? { accessories } : {}),
        ...(exportAreas.exhibitionLists ? { exhibitionLists } : {})
      }, null, 2), "application/json;charset=utf-8");
    }
    addHistory({ operation: "export", areas: selectedExportAreas, fileName, result: "success", format: exportFormat });
  };

  const reviewImport = () => {
    if (importArea !== "vehicles") {
      setMessage(t("importExport.areaPreviewPending", { area: t(`importExport.area.${importArea}`) }));
      return;
    }
    reviewRef.current?.scrollIntoView({ behavior: "smooth", block: "start" });
  };

  const retryTransfer = (entry: TransferEntry) => {
    if (entry.operation === "import") {
      selectImportArea(entry.areas[0]);
      return;
    }
    setExportAreas({
      vehicles: entry.areas.includes("vehicles"),
      accessories: entry.areas.includes("accessories"),
      exhibitionLists: entry.areas.includes("exhibitionLists")
    });
    if (entry.format) setExportFormat(entry.format);
  };

  const areaMeta: Record<DataArea, { icon: typeof TrainFront; formats: string }> = {
    vehicles: { icon: TrainFront, formats: "CSV · XML · JSON" },
    accessories: { icon: Boxes, formats: "CSV · JSON" },
    exhibitionLists: { icon: ClipboardCheck, formats: "JSON" }
  };
  const areaLabel = (area: DataArea) => area === "exhibitionLists"
    ? t("nav.exhibition")
    : t(`importExport.area.${area}`);
  const accept = importArea === "vehicles" ? ".csv,.tsv,.xml,.json" : importArea === "accessories" ? ".csv,.json" : ".json";

  return (
    <div className="transfer-workspace">
      <header className="transfer-page-head">
        <div>
          <p className="eyebrow">{t("importExport.eyebrow")}</p>
          <h1>{t("importExport.title")}</h1>
          <p>{t("importExport.workspaceSubtitle")}</p>
        </div>
        <button type="button" className="secondary-button transfer-history-button" onClick={() => historyRef.current?.scrollIntoView({ behavior: "smooth" })}>
          <History size={17} aria-hidden="true" />
          {t("importExport.showHistory")}
        </button>
      </header>

      {message && <p className="form-message transfer-message">{message}</p>}

      <section className="transfer-area-grid" aria-label={t("importExport.areaSelection")}>
        {areaOrder.map((area) => {
          const Icon = areaMeta[area].icon;
          const selected = importArea === area;
          return (
            <button
              key={area}
              type="button"
              className={`transfer-area-card${selected ? " selected" : ""}`}
              aria-pressed={selected}
              onClick={() => selectImportArea(area)}
            >
              <Icon size={34} aria-hidden="true" />
              <span>
                <strong>{areaLabel(area)}</strong>
                <b>{loading ? "…" : areaCounts[area].toLocaleString(language === "en" ? "en-US" : "de-DE")}</b>
                <small>{areaMeta[area].formats}</small>
              </span>
              {selected ? <CheckCircle2 className="area-selection-mark" size={22} aria-hidden="true" /> : <Circle className="area-selection-mark" size={22} aria-hidden="true" />}
            </button>
          );
        })}
      </section>

      <section className="transfer-work-grid">
        <article className="panel transfer-work-panel import-work-panel">
          <header className="transfer-panel-head">
            <h2><Download size={21} aria-hidden="true" />{t("importExport.importing")}</h2>
            <span className="transfer-area-badge"><i aria-hidden="true" />{areaLabel(importArea)}</span>
          </header>
          <label className="transfer-file-drop" onDragOver={(event) => event.preventDefault()} onDrop={handleDrop}>
            <Upload size={25} aria-hidden="true" />
            <span><strong>{t("importExport.file.select")}</strong> {t("importExport.file.orDrop")}</span>
            <small>{importArea === "vehicles" ? "CSV, TSV, XML oder RailKeeper-JSON" : areaMeta[importArea].formats}</small>
            <input type="file" accept={accept} onChange={handleFile} />
          </label>
          <div className="transfer-safety-row" aria-label={t("importExport.safety.title")}>
            <span><CheckCircle2 size={15} aria-hidden="true" />{t("importExport.safety.preview")}</span>
            <span><CheckCircle2 size={15} aria-hidden="true" />{t("importExport.safety.conflicts")}</span>
            <span><CheckCircle2 size={15} aria-hidden="true" />{t("importExport.safety.noAutomatic")}</span>
          </div>
          {stagedFile && (
            <div className="transfer-file-row">
              <FileText size={23} aria-hidden="true" />
              <span className="transfer-file-name"><strong title={stagedFile.name}>{stagedFile.name}</strong><small>{importSummary.total || "–"} {t("importExport.rows")}</small></span>
              <span className="transfer-count ready"><Check size={13} aria-hidden="true" />{importSummary.selected} {t("importExport.readyShort")}</span>
              <span className="transfer-count review">{importSummary.updates} {t("importExport.reviewShort")}</span>
              <span className="transfer-count error"><XCircle size={13} aria-hidden="true" />{importSummary.errors} {t("importExport.errorShort")}</span>
              <button type="button" className="secondary-button" onClick={reviewImport} disabled={previewLoading}>
                {previewLoading ? t("importExport.summary.reading") : t("importExport.reviewImport")}
              </button>
            </div>
          )}
        </article>

        <article className="panel transfer-work-panel export-work-panel">
          <header className="transfer-panel-head">
            <h2><Upload size={21} aria-hidden="true" />{t("importExport.exporting")}</h2>
          </header>
          <fieldset className="transfer-fieldset">
            <legend>{t("importExport.selectArea")}</legend>
            <div className="transfer-checkbox-row">
              {areaOrder.map((area) => (
                <label key={area}>
                  <input type="checkbox" checked={exportAreas[area]} onChange={() => toggleExportArea(area)} />
                  <span>{areaLabel(area)}</span>
                </label>
              ))}
            </div>
          </fieldset>
          <fieldset className="transfer-fieldset">
            <legend>{t("importExport.selectFormat")}</legend>
            <div className="transfer-format-grid">
              <button type="button" className={exportFormat === "csv" ? "selected" : ""} disabled={!formatAvailability.csv} onClick={() => chooseFormat("csv")}><FileSpreadsheet size={18} aria-hidden="true" />CSV</button>
              <button type="button" className={exportFormat === "json" ? "selected" : ""} disabled={!formatAvailability.json} onClick={() => chooseFormat("json")}><FileJson size={18} aria-hidden="true" />JSON</button>
              <button type="button" className={exportFormat === "print" ? "selected" : ""} disabled={!formatAvailability.print} onClick={() => chooseFormat("print")}><Printer size={18} aria-hidden="true" />PDF/{t("importExport.printShort")}</button>
            </div>
          </fieldset>
          <div className="transfer-export-footer">
            <div className="transfer-option-list">
              <label title={t("importExport.imagesUnavailable")}><input type="checkbox" disabled /><span>{t("importExport.includeImages")}</span><Info size={14} aria-hidden="true" /></label>
              <label title={t("importExport.filterUnavailable")}><input type="checkbox" disabled /><span>{t("importExport.filteredOnly")}</span><Info size={14} aria-hidden="true" /></label>
            </div>
            <div className="transfer-export-action">
              <span>{selectedExportCount.toLocaleString(language === "en" ? "en-US" : "de-DE")} {t("importExport.recordsSelected")}</span>
              <button type="button" className="primary-button" disabled={!exportValid} onClick={createExport}><Upload size={16} aria-hidden="true" />{t("importExport.createExport")}</button>
            </div>
          </div>
        </article>
      </section>

      <section ref={historyRef} className="panel transfer-history-panel">
        <header className="transfer-history-head"><h2><History size={21} aria-hidden="true" />{t("importExport.recentTransfers")}</h2></header>
        <div className="table-wrap transfer-history-table">
          <table>
            <thead><tr><th>{t("importExport.history.time")}</th><th>{t("importExport.history.operation")}</th><th>{t("importExport.history.area")}</th><th>{t("importExport.history.file")}</th><th>{t("importExport.history.result")}</th><th>{t("importExport.history.action")}</th></tr></thead>
            <tbody>
              {history.length === 0 ? (
                <tr><td colSpan={6} className="transfer-history-empty">{t("importExport.history.empty")}</td></tr>
              ) : history.map((entry) => (
                <tr key={entry.id}>
                  <td>{entry.timestamp.toLocaleString(language === "en" ? "en-US" : "de-DE", { dateStyle: "short", timeStyle: "short" })}</td>
                  <td><span className="transfer-operation"><Upload size={15} aria-hidden="true" />{t(`importExport.history.${entry.operation}`)}</span></td>
                  <td>{entry.areas.map(areaLabel).join(", ")}</td>
                  <td><span className="transfer-history-file" title={entry.fileName}>{entry.fileName}</span></td>
                  <td><span className={`transfer-result ${entry.result}`}>{entry.result === "success" ? <CheckCircle2 size={14} aria-hidden="true" /> : <Info size={14} aria-hidden="true" />}{t(`importExport.history.${entry.result}`)}</span></td>
                  <td><div className="transfer-history-actions"><button type="button" className="secondary-button">{t("importExport.history.details")}</button><button type="button" className="secondary-button" onClick={() => retryTransfer(entry)}><RotateCcw size={14} aria-hidden="true" />{t("importExport.history.retry")}</button><button type="button" className="icon-button" aria-label={t("importExport.history.more")}><MoreVertical size={16} aria-hidden="true" /></button></div></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      {importTable && (
        <section className="panel column-mapping-panel">
          <div className="panel-head">
            <div><h2>{t("importExport.mapping.title")}</h2><p>{t("importExport.mapping.subtitle", { file: importTable.fileName })}</p></div>
            <Database size={20} aria-hidden="true" />
          </div>
          <div className="column-mapping-grid">
            {importTable.mappings.map((mapping) => (
              <label key={mapping.index} className={mapping.key ? "" : "unmapped"}>
                <span><strong title={mapping.header}>{mapping.header || t("importExport.mapping.column", { number: mapping.index + 1 })}</strong><small>{mapping.key ? t("importExport.mapping.mapped") : t("importExport.mapping.unmapped")}</small></span>
                <AppSelect value={mapping.key} onChange={(event) => setColumnMapping(mapping.index, event.target.value as VehicleImportField | "")}>
                  <option value="">{t("importExport.mapping.ignore")}</option>
                  {vehicleImportFields.map((field) => <option key={field.key} value={field.key}>{csvFieldLabel(field.key)}</option>)}
                </AppSelect>
              </label>
            ))}
          </div>
          <p className="source-note backup-note">{t("importExport.mapping.note")}</p>
        </section>
      )}

      {rows.length > 0 && (
        <section ref={reviewRef} className="panel import-review-panel">
          <div className="panel-head">
            <div><h2 className="panel-title-inline"><ClipboardCheck size={20} aria-hidden="true" />{t("importExport.review.title")}</h2><p>{t("importExport.review.subtitle")}</p></div>
            <button type="button" className="primary-button" disabled={saving || importSummary.selected === 0} onClick={saveSelected}><Save size={15} aria-hidden="true" />{t("importExport.review.saveSelection")}</button>
          </div>
          <div className="table-wrap import-table">
            <table>
              <thead><tr><th>{t("importExport.review.apply")}</th><th>{t("importExport.review.action")}</th><th>{t("importExport.review.inventory")}</th><th>{csvFieldLabel("manufacturer")}</th><th>{t("importExport.review.article")}</th><th>{csvFieldLabel("name")}</th><th>{csvFieldLabel("gauge")}</th><th>{csvFieldLabel("category")}</th><th>{csvFieldLabel("gattung")}</th><th>{t("settings.master.status")}</th></tr></thead>
              <tbody>
                {rows.map((row) => {
                  const existing = row.duplicateVehicleId ? vehicles.find((vehicle) => vehicle.id === row.duplicateVehicleId) : undefined;
                  const changes = getImportChanges(row, existing, csvFieldLabel, t("common.yes"), t("common.no"));
                  return (
                    <Fragment key={row.id}>
                      <tr className={row.status === "error" ? "import-row-error" : row.status === "warning" ? "import-row-warning" : row.status === "saved" ? "import-row-saved" : ""}>
                        <td><input type="checkbox" checked={row.selected} disabled={row.status === "saved" || row.status === "error"} onChange={(event) => setRowSelected(row.id, event.target.checked)} /></td>
                        <td><AppSelect value={row.mode} disabled={row.status === "saved"} onChange={(event) => setRowMode(row.id, event.target.value as ImportRow["mode"])}><option value="create">{t("importExport.review.create")}</option><option value="update" disabled={!row.duplicateVehicleId}>{t("importExport.review.update")}</option></AppSelect></td>
                        <td><input value={row.vehicle.inventoryNumber || ""} onChange={(event) => updateRow(row.id, { inventoryNumber: event.target.value })} /></td>
                        <td><input value={row.vehicle.manufacturer} onChange={(event) => updateRow(row.id, { manufacturer: event.target.value })} /></td>
                        <td><input value={row.vehicle.articleNumber || ""} onChange={(event) => updateRow(row.id, { articleNumber: event.target.value })} /></td>
                        <td><input value={row.vehicle.name} onChange={(event) => updateRow(row.id, { name: event.target.value })} /></td>
                        <td><input value={row.vehicle.gauge} onChange={(event) => updateRow(row.id, { gauge: event.target.value })} /></td>
                        <td><input value={row.vehicle.category} onChange={(event) => updateRow(row.id, { category: event.target.value })} /></td>
                        <td><input value={row.vehicle.gattung} onChange={(event) => updateRow(row.id, { gattung: event.target.value })} /></td>
                        <td><span className={`import-status ${row.status}`}><strong>{row.status === "saved" ? t("importExport.savedShort") : row.status}</strong><small>{row.issues.join(" · ") || "-"}</small></span></td>
                      </tr>
                      {row.mode === "update" && existing && changes.length > 0 && (
                        <tr className="import-change-row"><td colSpan={10}><div className="import-change-panel"><div className="import-change-head"><strong>{t("importExport.review.updatePreview")}</strong><span>{t("importExport.review.overwrites", { count: changes.filter((change) => change.status === "overwrite").length })}, {t("importExport.review.fills", { count: changes.filter((change) => change.status === "fill").length })}</span></div><div className="table-wrap"><table><thead><tr><th>{t("importExport.review.field")}</th><th>{t("importExport.review.current")}</th><th>{t("importExport.review.import")}</th><th>{t("settings.master.status")}</th></tr></thead><tbody>{changes.map((change) => <tr key={change.key} className={`change-${change.status}`}><td>{change.label}</td><td>{displayImportValue(change.current)}</td><td>{displayImportValue(change.incoming)}</td><td>{t(`importExport.review.status.${change.status}`)}</td></tr>)}</tbody></table></div></div></td></tr>
                      )}
                    </Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
        </section>
      )}
    </div>
  );
}
