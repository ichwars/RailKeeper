import { ChangeEvent, Fragment, useEffect, useMemo, useState } from "react";
import { AlertTriangle, Check, ClipboardCheck, Database, Download, FileInput, Printer, Save, Upload } from "lucide-react";
import { api, CreateVehicleRequest, ECoSConnectionResult, ECoSRawLocomotive, ECoSRawProbe, MasterDataEntry, Vehicle, VehicleCVValueInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import {
  buildECoSVehicleDraftRow,
  ColumnMapping,
  cv8ManufacturersFromMasterData,
  defaultColumnMappings,
  detectDelimiter,
  displayImportValue,
  downloadText,
  ECoSBusyPhase,
  ECoSImportSession,
  ecosImportSessionStorageKey,
  ECoSVehicleDraftPayload,
  ecosExternalMapping,
  ecosRequiredFields,
  ecosVehicleDraftStorageKey,
  FunctionImportSuggestion,
  getImportChanges,
  ImportRow,
  ImportTablePreview,
  importRowsFromTable,
  mergeExternalMapping,
  mergeImportedVehicle,
  normalizeECoSProtocolForCV,
  optionValue,
  parseDelimited,
  parseXMLImport,
  printInventory,
  renderECoSRawProbe,
  VehicleImportField,
  vehicleImportFields,
  vehiclesToCSV
} from "./importExportHelpers";

export function ImportExportView() {
  const { language, t } = useI18n();
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [symbols, setSymbols] = useState<MasterDataEntry[]>([]);
  const [masterOptions, setMasterOptions] = useState<Record<string, MasterDataEntry[]>>({});
  const [rows, setRows] = useState<ImportRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [importTable, setImportTable] = useState<ImportTablePreview | null>(null);
  const [ecosHost, setEcosHost] = useState(window.localStorage.getItem("railkeeper.ecos.host") || "");
  const [ecosPort, setEcosPort] = useState(window.localStorage.getItem("railkeeper.ecos.port") || "15471");
  const [ecosBusy, setEcosBusy] = useState<ECoSBusyPhase>("idle");
  const [ecosResult, setEcosResult] = useState<ECoSConnectionResult | null>(null);
  const [ecosRawProbe, setEcosRawProbe] = useState<ECoSRawProbe | null>(null);
  const [ecosSession, setEcosSession] = useState<ECoSImportSession | null>(null);
  const [ecosMessage, setEcosMessage] = useState("");
  const fieldLabel = (key: VehicleImportField) => t(`vehicle.field.${key}`);
  const issueLabels = {
    missingManufacturer: t("importExport.issue.missingManufacturer"),
    missingName: t("importExport.issue.missingName"),
    missingGauge: t("importExport.issue.missingGauge"),
    missingCategory: t("importExport.issue.missingCategory"),
    missingGattung: t("importExport.issue.missingGattung"),
    duplicate: t("importExport.issue.duplicate"),
    ecosMatched: t("importExport.ecos.matched")
  };
  const cv8Manufacturers = useMemo(
    () => cv8ManufacturersFromMasterData(masterOptions.cv8_manufacturer || []),
    [masterOptions]
  );

  const rememberECoSImportSession = (session: ECoSImportSession) => {
    const next = { ...session, updatedAt: new Date().toISOString() };
    window.sessionStorage.setItem(ecosImportSessionStorageKey, JSON.stringify(next));
    setEcosSession(next);
    return next;
  };

  const createECoSImportSession = (probe: ECoSRawProbe): ECoSImportSession => rememberECoSImportSession({
    id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    rawProbe: probe,
    statuses: Object.fromEntries(probe.locomotives.map((locomotive) => [
      String(locomotive.objectId),
      { status: "open", updatedAt: new Date().toISOString() }
    ]))
  });

  const updateECoSImportSessionStatus = (objectId: number, status: ECoSImportSession["statuses"][string]["status"]) => {
    const base = ecosSession || (ecosRawProbe ? createECoSImportSession(ecosRawProbe) : null);
    if (!base) return null;
    return rememberECoSImportSession({
      ...base,
      statuses: {
        ...base.statuses,
        [String(objectId)]: {
          ...base.statuses[String(objectId)],
          status,
          updatedAt: new Date().toISOString()
        }
      }
    });
  };

  useEffect(() => {
    const rawSession = window.sessionStorage.getItem(ecosImportSessionStorageKey);
    if (!rawSession) return;
    try {
      const session = JSON.parse(rawSession) as ECoSImportSession;
      if (!session?.rawProbe?.locomotives) return;
      setEcosSession(session);
      setEcosRawProbe(session.rawProbe);
      setEcosHost(session.rawProbe.host || "");
      setEcosPort(String(session.rawProbe.port || 15471));
      setEcosResult({
        connected: true,
        host: session.rawProbe.host,
        port: session.rawProbe.port,
        message: t("importExport.ecos.sessionRestored")
      });
      const saved = Object.values(session.statuses || {}).filter((item) => item.status === "saved").length;
      setEcosMessage(t("importExport.ecos.sessionProgress", { saved, total: session.rawProbe.locomotives.length }));
    } catch {
      window.sessionStorage.removeItem(ecosImportSessionStorageKey);
    }
  }, [t]);

  useEffect(() => {
    api.vehicles().then(setVehicles).catch((error: Error) => setMessage(error.message)).finally(() => setLoading(false));
    api.masterDataAll(true)
      .then((entriesByType) => {
        setMasterOptions(entriesByType);
        setSymbols(entriesByType.symbols || []);
      })
      .catch(() => {
        setMasterOptions({});
        setSymbols([]);
      });
  }, []);

  const importSummary = useMemo(() => ({
    total: rows.length,
    selected: rows.filter((row) => row.selected && row.status !== "saved").length,
    errors: rows.filter((row) => row.status === "error").length,
    updates: rows.filter((row) => row.mode === "update" && row.status !== "saved").length,
    saved: rows.filter((row) => row.status === "saved").length
  }), [rows]);

  const mappingSummary = useMemo(() => {
    if (!importTable) {
      return { mapped: 0, unmapped: 0 };
    }
    const visibleMappings = importTable.mappings.filter((mapping) => mapping.header.trim());
    return {
      mapped: visibleMappings.filter((mapping) => mapping.key).length,
      unmapped: visibleMappings.filter((mapping) => !mapping.key).length
    };
  }, [importTable]);

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

  const setColumnMapping = (columnIndex: number, key: VehicleImportField | "") => {
    if (!importTable) {
      return;
    }
    const mappings: ColumnMapping[] = importTable.mappings.map((mapping) => {
      if (mapping.index === columnIndex) {
        return { ...mapping, key };
      }
      return key && mapping.key === key ? { ...mapping, key: "" } : mapping;
    });
    setImportTable({ ...importTable, mappings });
    setRows(importRowsFromTable(importTable.table, vehicles, mappings, issueLabels));
  };

  const updateRow = (rowID: string, patch: Partial<ImportRow["vehicle"]>) => {
    setRows((current) => current.map((row) => {
      if (row.id !== rowID) {
        return row;
      }
      const vehicle = { ...row.vehicle, ...patch };
      const issues: string[] = [];
      const duplicate = vehicle.inventoryNumber ? vehicles.find((existing) => existing.inventoryNumber.toLowerCase() === vehicle.inventoryNumber?.toLowerCase()) : undefined;
      const importedKeys = Array.from(new Set([...row.importedKeys, ...Object.keys(patch) as (keyof CreateVehicleRequest)[]]));
      if (duplicate) {
        issues.push(t("importExport.issue.duplicate"));
        return {
          ...row,
          vehicle,
          importedKeys,
          duplicateVehicleId: duplicate.id,
          mode: "update",
          issues,
          status: "warning",
          selected: row.selected
        };
      }
      if (!vehicle.manufacturer) issues.push(t("importExport.issue.missingManufacturer"));
      if (!vehicle.name) issues.push(t("importExport.issue.missingName"));
      if (!vehicle.gauge) issues.push(t("importExport.issue.missingGauge"));
      if (!vehicle.category) issues.push(t("importExport.issue.missingCategory"));
      if (!vehicle.gattung) issues.push(t("importExport.issue.missingGattung"));
      return {
        ...row,
        vehicle,
        importedKeys,
        duplicateVehicleId: undefined,
        mode: "create",
        issues,
        status: issues.length ? "error" : "ok",
        selected: issues.length ? false : row.selected
      };
    }));
  };

  const setRowSelected = (rowID: string, selected: boolean) => {
    setRows((current) => current.map((item) => item.id === rowID ? { ...item, selected } : item));
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

  const handleFile = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }
    setMessage("");
    const extension = file.name.split(".").pop()?.toLowerCase() || "";
    const text = await file.text();
    if (extension === "json") {
      const parsed = JSON.parse(text) as Vehicle[] | { vehicles?: Vehicle[] };
      const source = Array.isArray(parsed) ? parsed : parsed.vehicles || [];
      const table = [
        [
          fieldLabel("inventoryNumber"),
          fieldLabel("manufacturer"),
          fieldLabel("articleNumber"),
          fieldLabel("name"),
          fieldLabel("gauge"),
          fieldLabel("epoch"),
          fieldLabel("railwayCompany"),
          fieldLabel("category"),
          fieldLabel("gattung"),
          fieldLabel("digital"),
          fieldLabel("digitalDecoderNumber"),
          fieldLabel("listPrice")
        ],
        ...source.map((vehicle) => [vehicle.inventoryNumber, vehicle.manufacturer, vehicle.articleNumber || "", vehicle.name, vehicle.gauge, vehicle.epoch || "", vehicle.railwayCompany || "", vehicle.category || "", vehicle.gattung || "", vehicle.digital ? t("common.yes") : t("common.no"), vehicle.digitalDecoderNumber || "", vehicle.listPrice || ""])
      ];
      loadImportTable(table, file.name);
      return;
    }
    if (extension === "xml") {
      const table = parseXMLImport(text);
      if (table.length) {
        loadImportTable(table, file.name);
      } else {
        setRows([]);
        setImportTable(null);
        setMessage(t("importExport.error.emptyTable"));
      }
      return;
    }

    const delimiter = extension === "tsv" ? "\t" : detectDelimiter(text);
    loadImportTable(parseDelimited(text, delimiter), file.name);
  };

  const saveFunctionSuggestions = async (vehicleID: string, suggestions: FunctionImportSuggestion[] = []) => {
    for (const suggestion of suggestions) {
      await api.updateVehicleFunction(vehicleID, suggestion.functionKey, {
        name: suggestion.name || "",
        symbolKey: suggestion.symbolKey || "",
        functionType: suggestion.functionType || "standard",
        mode: suggestion.mode || "dauer",
        directionDependent: Boolean(suggestion.directionDependent),
        notes: suggestion.notes || ""
      });
    }
  };

  const saveCVSuggestions = async (vehicleID: string, suggestions: VehicleCVValueInput[] = []) => {
    if (suggestions.length === 0) {
      return;
    }
    const vehicle = await api.vehicle(vehicleID);
    for (const suggestion of suggestions) {
      const existing = (vehicle.cvValues || []).find((value) =>
        value.cvNumber === suggestion.cvNumber &&
        (value.decoderProfile || "") === (suggestion.decoderProfile || "")
      );
      if (existing) {
        await api.updateVehicleCVValue(vehicleID, existing.id, suggestion);
      } else {
        await api.createVehicleCVValue(vehicleID, suggestion);
      }
    }
  };

  const saveSelected = async () => {
    setSaving(true);
    setMessage("");
    for (const row of rows) {
      if (!row.selected || row.status === "saved" || row.status === "error") {
        continue;
      }
      try {
        const existing = row.duplicateVehicleId ? vehicles.find((vehicle) => vehicle.id === row.duplicateVehicleId) : undefined;
        let saved = row.mode === "update" && existing
          ? await api.updateVehicle(existing.id, mergeImportedVehicle(existing, row.vehicle, row.importedKeys))
          : await api.createVehicle(row.vehicle);
        if (row.externalMapping) {
          const mapping = await api.upsertVehicleExternalMapping(saved.id, row.externalMapping);
          saved = mergeExternalMapping(saved, mapping);
        }
        await saveFunctionSuggestions(saved.id, row.functionSuggestions);
        await saveCVSuggestions(saved.id, row.cvSuggestions);
        setVehicles((current) => {
          if (row.mode === "update") {
            return current.map((vehicle) => vehicle.id === saved.id ? saved : vehicle);
          }
          return [...current, saved];
        });
        setRows((current) => current.map((item) => item.id === row.id ? { ...item, selected: false, status: "saved", issues: [] } : item));
      } catch (error) {
        const message = error instanceof Error ? error.message : t("importExport.error.importFailed");
        setRows((current) => current.map((item) => item.id === row.id ? { ...item, status: "error", issues: [message] } : item));
      }
    }
    setSaving(false);
  };

  const ecosInput = () => ({
    host: ecosHost.trim(),
    port: Number(ecosPort) || 15471
  });

  const rememberECoSSettings = () => {
    window.localStorage.setItem("railkeeper.ecos.host", ecosHost.trim());
    window.localStorage.setItem("railkeeper.ecos.port", ecosPort.trim() || "15471");
  };

  const testECoSConnection = async () => {
    setEcosBusy("connecting");
    setEcosMessage("");
    setEcosResult(null);
    setEcosRawProbe(null);
    setEcosSession(null);
    window.sessionStorage.removeItem(ecosImportSessionStorageKey);
    try {
      rememberECoSSettings();
      const result = await api.testECoSConnection(ecosInput());
      setEcosResult(result);
      setEcosMessage(result.message);
    } catch (error) {
      setEcosMessage(error instanceof Error ? error.message : t("importExport.ecos.error"));
    } finally {
      setEcosBusy("idle");
    }
  };

  const probeECoSRawData = async () => {
    if (!ecosResult?.connected) {
      setEcosMessage(t("importExport.ecos.connectFirst"));
      return;
    }
    setEcosBusy("fetching");
    setEcosMessage("");
    setEcosRawProbe(null);
    try {
      rememberECoSSettings();
      const probe = await api.probeECoSLocomotiveRaw(ecosInput());
      setEcosRawProbe(probe);
      createECoSImportSession(probe);
      setEcosMessage(probe.message);
    } catch (error) {
      setEcosMessage(error instanceof Error ? error.message : t("importExport.ecos.error"));
    } finally {
      setEcosBusy("idle");
    }
  };

  const openECoSVehicleDraft = (locomotive: ECoSRawLocomotive) => {
    const row = buildECoSVehicleDraftRow(locomotive, vehicles, symbols, {
      matched: issueLabels.ecosMatched,
      missingManufacturer: issueLabels.missingManufacturer,
      missingName: issueLabels.missingName,
      missingGauge: issueLabels.missingGauge,
      missingCategory: issueLabels.missingCategory,
      missingGattung: issueLabels.missingGattung
    }, cv8Manufacturers);
    const unclearFields = ecosRequiredFields.filter((field) => !String(row.vehicle[field] ?? "").trim());
    const updatedSession = updateECoSImportSessionStatus(locomotive.objectId, "editing");
    const payload: ECoSVehicleDraftPayload = {
      source: "ecos",
      sourceSummary: {
        objectId: locomotive.objectId,
        name: locomotive.name || `ECoS ${locomotive.objectId}`,
        address: typeof locomotive.address === "number" ? String(locomotive.address) : "",
        protocol: normalizeECoSProtocolForCV(locomotive.protocol),
        profile: locomotive.profile || ""
      },
      vehicle: row.vehicle,
      externalMapping: row.externalMapping || ecosExternalMapping(locomotive),
      cvValues: row.cvSuggestions || [],
      functionValues: row.functionSuggestions || [],
      unclearFields,
      returnToEcos: updatedSession ? { sessionId: updatedSession.id, objectId: locomotive.objectId } : undefined
    };
    window.sessionStorage.setItem(ecosVehicleDraftStorageKey, JSON.stringify(payload));
    window.history.pushState(null, "", "/vehicles?source=ecos");
    window.dispatchEvent(new PopStateEvent("popstate"));
    setEcosMessage(t("importExport.ecos.handoff"));
  };

  const addECoSProbeToImportReview = () => {
    if (!ecosRawProbe || ecosRawProbe.locomotives.length === 0) {
      setEcosMessage(t("importExport.ecos.noImportRows"));
      return;
    }
    const ecosRows = ecosRawProbe.locomotives.map((locomotive) => buildECoSVehicleDraftRow(locomotive, vehicles, symbols, {
      matched: issueLabels.ecosMatched,
      missingManufacturer: issueLabels.missingManufacturer,
      missingName: issueLabels.missingName,
      missingGauge: issueLabels.missingGauge,
      missingCategory: issueLabels.missingCategory,
      missingGattung: issueLabels.missingGattung
    }, cv8Manufacturers));
    const nextIds = new Set(ecosRows.map((row) => row.id));
    setImportTable(null);
    setRows((current) => [...current.filter((row) => !nextIds.has(row.id)), ...ecosRows]);
    setEcosMessage(t("importExport.ecos.reviewAdded", { count: ecosRows.length }));
  };

  return (
    <>
      <section className="page-head">
        <p className="eyebrow">{t("importExport.eyebrow")}</p>
        <h1>{t("importExport.title")}</h1>
        <p>{t("importExport.subtitle")}</p>
      </section>

      {message && <p className="form-message">{message}</p>}

      <section className="import-export-grid">
        <article className="panel transfer-panel">
          <div className="panel-head">
            <div>
              <h2 className="panel-title-inline"><FileInput size={20} aria-hidden="true" />{t("importExport.import.title")}</h2>
              <p>{t("importExport.import.subtitle")}</p>
            </div>
          </div>
          <label className="file-drop compact-drop">
            <Upload size={18} aria-hidden="true" />
            {t("importExport.file.choose")}
            <input type="file" accept=".csv,.tsv,.xml,.json" onChange={handleFile} />
          </label>
          <div className="import-summary">
            <span>{t("importExport.summary.rows", { count: importSummary.total })}</span>
            <span>{previewLoading ? t("importExport.summary.reading") : t("importExport.summary.ready", { count: importSummary.selected })}</span>
            <span>{t("importExport.summary.updates", { count: importSummary.updates })}</span>
            <span className={importSummary.errors ? "danger" : ""}>{t("importExport.summary.notes", { count: importSummary.errors })}</span>
            <span>{t("importExport.summary.saved", { count: importSummary.saved })}</span>
            {importTable && <span>{t("importExport.summary.mapped", { count: mappingSummary.mapped })}</span>}
            {importTable && <span className={mappingSummary.unmapped ? "danger" : ""}>{t("importExport.summary.open", { count: mappingSummary.unmapped })}</span>}
          </div>
        </article>

        <article className="panel transfer-panel">
          <div className="panel-head">
            <div>
              <h2 className="panel-title-inline"><Download size={20} aria-hidden="true" />{t("importExport.export.title")}</h2>
              <p>{t("importExport.export.subtitle")}</p>
            </div>
          </div>
          <div className="export-actions">
            <button type="button" className="secondary-button" disabled={loading || vehicles.length === 0} onClick={() => downloadText("railkeeper-bestand.csv", `\uFEFF${vehiclesToCSV(vehicles, fieldLabel, t("common.yes"), t("common.no"))}`, "text/csv;charset=utf-8")}>
              <Download size={15} aria-hidden="true" />
              {t("importExport.export.csv")}
            </button>
            <button type="button" className="secondary-button" disabled={loading || vehicles.length === 0} onClick={() => downloadText("railkeeper-bestand.json", JSON.stringify({ format: "railkeeper-vehicles", version: 1, vehicles }, null, 2), "application/json;charset=utf-8")}>
              <Download size={15} aria-hidden="true" />
              {t("importExport.export.json")}
            </button>
            <button type="button" className="secondary-button" disabled={loading || vehicles.length === 0} onClick={() => printInventory(vehicles, fieldLabel, language, t)}>
              <Printer size={15} aria-hidden="true" />
              {t("importExport.export.print")}
            </button>
          </div>
        </article>
      </section>

      <section className="panel transfer-panel ecos-panel">
        <div className="panel-head">
          <div>
            <h2 className="panel-title-inline"><Database size={20} aria-hidden="true" />{t("importExport.ecos.title")}</h2>
            <p>{t("importExport.ecos.subtitle")}</p>
          </div>
        </div>
        <div className="ecos-connection-grid">
          <label>
            {t("importExport.ecos.host")}
            <input value={ecosHost} onChange={(event) => setEcosHost(event.target.value)} placeholder={t("importExport.ecos.hostPlaceholder")} />
          </label>
          <label>
            {t("importExport.ecos.port")}
            <input value={ecosPort} onChange={(event) => setEcosPort(event.target.value)} inputMode="numeric" placeholder="15471" />
          </label>
          <div className="ecos-actions">
            <button type="button" className="secondary-button" onClick={testECoSConnection} disabled={ecosBusy !== "idle" || !ecosHost.trim()}>
              <Check size={15} aria-hidden="true" />
              {t("importExport.ecos.connect")}
            </button>
            {ecosResult?.connected && (
              <button type="button" className="primary-button" onClick={probeECoSRawData} disabled={ecosBusy !== "idle" || !ecosHost.trim()}>
                <Download size={15} aria-hidden="true" />
                {t("importExport.ecos.fetchData")}
              </button>
            )}
          </div>
        </div>
        <div className="ecos-status-strip">
          <span className={ecosResult?.connected ? "status-ok" : ecosResult ? "status-error" : ""}>
            {ecosResult ? (ecosResult.connected ? t("importExport.ecos.connected") : t("importExport.ecos.notConnected")) : t("importExport.ecos.idle")}
          </span>
          {ecosResult?.status && <span>{t("importExport.ecos.status", { status: ecosResult.status })}</span>}
          {ecosResult?.protocolVersion && <span>{t("importExport.ecos.protocol", { version: ecosResult.protocolVersion })}</span>}
          {ecosResult?.applicationVersion && <span>{t("importExport.ecos.application", { version: ecosResult.applicationVersion })}</span>}
        </div>
        {ecosBusy !== "idle" && (
          <p className="ecos-busy-indicator">
            {ecosBusy === "connecting" ? t("importExport.ecos.connecting") : t("importExport.ecos.fetching")}
          </p>
        )}
        {ecosMessage && <p className="form-message">{ecosMessage}</p>}
        <p className="source-note backup-note">{t("importExport.ecos.note")}</p>
        <p className="source-note backup-note">{t("importExport.ecos.mappingNote")}</p>
        {ecosRawProbe && ecosRawProbe.locomotives.length > 0 && (
          <div className="ecos-preview-toolbar">
            <span>{t("importExport.ecos.reviewHint", { count: ecosRawProbe.locomotives.length })}</span>
            <button type="button" className="secondary-button" onClick={addECoSProbeToImportReview} disabled={saving}>
              <ClipboardCheck size={15} aria-hidden="true" />
              {t("importExport.ecos.addToReview")}
            </button>
          </div>
        )}
        {renderECoSRawProbe(ecosRawProbe, vehicles, t, cv8Manufacturers, openECoSVehicleDraft)}
      </section>


      {importTable && (
        <section className="panel column-mapping-panel">
          <div className="panel-head">
            <div>
              <h2>{t("importExport.mapping.title")}</h2>
              <p>{t("importExport.mapping.subtitle", { file: importTable.fileName })}</p>
            </div>
            <Database size={20} aria-hidden="true" />
          </div>
          <div className="column-mapping-grid">
            {importTable.mappings.map((mapping) => (
              <label key={mapping.index} className={mapping.key ? "" : "unmapped"}>
                <span>
                  <strong title={mapping.header}>{mapping.header || t("importExport.mapping.column", { number: mapping.index + 1 })}</strong>
                  <small>{mapping.key ? t("importExport.mapping.mapped") : t("importExport.mapping.unmapped")}</small>
                </span>
                <select value={mapping.key} onChange={(event) => setColumnMapping(mapping.index, event.target.value as VehicleImportField | "")}>
                  <option value="">{t("importExport.mapping.ignore")}</option>
                  {vehicleImportFields.map((field) => (
                    <option key={field.key} value={field.key}>{fieldLabel(field.key)}</option>
                  ))}
                </select>
              </label>
            ))}
          </div>
          <p className="source-note backup-note">{t("importExport.mapping.note")}</p>
        </section>
      )}

      {rows.length > 0 && (
      <section className="panel import-review-panel">
        <div className="panel-head">
          <div>
            <h2 className="panel-title-inline"><ClipboardCheck size={20} aria-hidden="true" />{t("importExport.review.title")}</h2>
            <p>{t("importExport.review.subtitle")}</p>
          </div>
          <button type="button" className="primary-button" disabled={saving || importSummary.selected === 0} onClick={saveSelected}>
            <Save size={15} aria-hidden="true" />
            {t("importExport.review.saveSelection")}
          </button>
        </div>

        {rows.length === 0 ? (
          <p className="empty-state">{t("importExport.review.empty")}</p>
        ) : (
          <div className="table-wrap import-table">
            <table>
              <thead>
                <tr>
                  <th>{t("importExport.review.apply")}</th>
                  <th>{t("importExport.review.action")}</th>
                  <th>{t("importExport.review.inventory")}</th>
                  <th>{fieldLabel("manufacturer")}</th>
                  <th>{t("importExport.review.article")}</th>
                  <th>{fieldLabel("name")}</th>
                  <th>{fieldLabel("gauge")}</th>
                  <th>{fieldLabel("category")}</th>
                  <th>{fieldLabel("gattung")}</th>
                  <th>{t("exhibition.status")}</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((row) => {
                  const existing = row.duplicateVehicleId ? vehicles.find((vehicle) => vehicle.id === row.duplicateVehicleId) : undefined;
                  const changes = getImportChanges(row, existing, fieldLabel, t("common.yes"), t("common.no"));
                  return (
                    <Fragment key={row.id}>
                      <tr className={row.status === "error" ? "import-row-error" : row.status === "warning" ? "import-row-warning" : row.status === "saved" ? "import-row-saved" : ""}>
                        <td><input type="checkbox" checked={row.selected} disabled={row.status === "saved" || row.status === "error"} onChange={(event) => setRowSelected(row.id, event.target.checked)} /></td>
                        <td>
                          <select value={row.mode} disabled={row.status === "saved"} onChange={(event) => setRowMode(row.id, event.target.value as ImportRow["mode"])}>
                            <option value="create">{t("importExport.review.create")}</option>
                            <option value="update" disabled={!row.duplicateVehicleId}>{t("importExport.review.update")}</option>
                          </select>
                        </td>
                        <td><input value={row.vehicle.inventoryNumber || ""} onChange={(event) => updateRow(row.id, { inventoryNumber: event.target.value })} /></td>
                        <td><input value={row.vehicle.manufacturer} onChange={(event) => updateRow(row.id, { manufacturer: event.target.value })} /></td>
                        <td><input value={row.vehicle.articleNumber || ""} onChange={(event) => updateRow(row.id, { articleNumber: event.target.value })} /></td>
                        <td><input value={row.vehicle.name} onChange={(event) => updateRow(row.id, { name: event.target.value })} /></td>
                        <td><input value={row.vehicle.gauge} onChange={(event) => updateRow(row.id, { gauge: event.target.value })} /></td>
                        <td><input value={row.vehicle.category} onChange={(event) => updateRow(row.id, { category: event.target.value })} /></td>
                        <td><input value={row.vehicle.gattung} onChange={(event) => updateRow(row.id, { gattung: event.target.value })} /></td>
                        <td>
                          <span className={`import-status ${row.status}`}>
                            {row.status === "saved" ? <Check size={14} /> : row.status === "error" || row.status === "warning" ? <AlertTriangle size={14} /> : <Check size={14} />}
                            {row.status === "saved" ? t("common.saved") : row.issues[0] || t("common.ready")}
                          </span>
                        </td>
                      </tr>
                      {((existing && row.mode === "update") || (row.functionSuggestions && row.functionSuggestions.length > 0) || (row.cvSuggestions && row.cvSuggestions.length > 0)) && (
                        <tr className="import-change-row">
                          <td colSpan={10}>
                            <div className="import-change-panel">
                              {existing && row.mode === "update" && (
                                <>
                                  <div>
                                    <strong>{t("importExport.review.updatePreview")}</strong>
                                    <span>{t("importExport.review.overwrites", { count: changes.filter((change) => change.status === "overwrite").length })}, {t("importExport.review.fills", { count: changes.filter((change) => change.status === "fill").length })}</span>
                                  </div>
                                  {changes.length === 0 ? (
                                    <p>{t("importExport.review.noValues")}</p>
                                  ) : (
                                    <table>
                                      <thead>
                                        <tr>
                                          <th>{t("importExport.review.field")}</th>
                                          <th>{t("importExport.review.current")}</th>
                                          <th>{t("importExport.review.import")}</th>
                                          <th>{t("exhibition.status")}</th>
                                        </tr>
                                      </thead>
                                      <tbody>
                                        {changes.map((change) => (
                                          <tr key={change.key} className={`change-${change.status}`}>
                                            <td>{change.label}</td>
                                            <td>{change.current}</td>
                                            <td>{change.incoming}</td>
                                            <td>{t(`importExport.review.status.${change.status}`)}</td>
                                          </tr>
                                        ))}
                                      </tbody>
                                    </table>
                                  )}
                                </>
                              )}
                              {row.functionSuggestions && row.functionSuggestions.length > 0 && (
                                <div className="import-function-panel">
                                  <div>
                                    <strong>{t("importExport.ecos.review.functions")}</strong>
                                    <span>{t("importExport.ecos.review.valuesApplied", { count: row.functionSuggestions.length })}</span>
                                  </div>
                                  <div className="import-function-grid">
                                    {row.functionSuggestions.map((fn) => (
                                      <span key={fn.functionKey} title={fn.notes || undefined}>
                                        <strong>{fn.functionKey}</strong>
                                        {fn.name || (typeof fn.ecosDescription === "number" ? `ECoS ${fn.ecosDescription}` : "ECoS")}
                                        {fn.active && <em>{t("common.active")}</em>}
                                      </span>
                                    ))}
                                  </div>
                                </div>
                              )}
                              {row.cvSuggestions && row.cvSuggestions.length > 0 && (
                                <div className="import-function-panel">
                                  <div>
                                    <strong>{t("importExport.ecos.review.cvValues")}</strong>
                                    <span>{t("importExport.ecos.review.valuesApplied", { count: row.cvSuggestions.length })}</span>
                                  </div>
                                  <div className="import-function-grid">
                                    {row.cvSuggestions.slice(0, 16).map((cv) => (
                                      <span key={`${cv.cvNumber}-${cv.decoderProfile || ""}`} title={cv.description || undefined}>
                                        <strong>CV {cv.cvNumber}</strong>
                                        {cv.value}
                                      </span>
                                    ))}
                                    {row.cvSuggestions.length > 16 && <span>{t("importExport.ecos.draft.moreCVs", { count: row.cvSuggestions.length - 16 })}</span>}
                                  </div>
                                </div>
                              )}
                            </div>
                          </td>
                        </tr>
                      )}
                    </Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>
      )}
    </>
  );
}
