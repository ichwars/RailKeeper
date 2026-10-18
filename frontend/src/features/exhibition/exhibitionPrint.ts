import { ExhibitionEntry, ExhibitionEntryStatus, ExhibitionList, MasterDataEntry } from "../../shared/api";
import { functionSymbolImageData } from "../../shared/functionSymbolImages";
import { functionSymbolMetadata } from "../../shared/functionSymbols";

type Translate = (key: string, values?: Record<string, string | number>) => string;
type PrintEntry = ExhibitionEntry & { status?: ExhibitionEntryStatus };
type PrintOptions = { includeImages: boolean };

const htmlEscapes: Record<string, string> = {
  "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;"
};

function escapeHTML(value: unknown) {
  return String(value ?? "").replace(/[&<>"']/g, (char) => htmlEscapes[char]);
}

function imageHTML(source: string | undefined, alt = "") {
  if (!source) return "";
  // Attribute escaping alone does not make an imported image URL safe.
  try {
    const url = new URL(source, "https://railkeeper.invalid");
    if (!["http:", "https:"].includes(url.protocol) &&
      !/^data:image\/(png|jpeg|webp|gif|svg\+xml)[;,]/i.test(source)) return "";
  } catch {
    return "";
  }
  return `<img src="${escapeHTML(source)}" alt="${escapeHTML(alt)}" />`;
}

export function printFunctionChips(value: string | undefined, symbols: MasterDataEntry[]) {
  if (!value?.trim()) return "–";
  try {
    const parsed: unknown = JSON.parse(value);
    if (Array.isArray(parsed)) {
      const chips = parsed.map((item: unknown) => {
        if (!item || typeof item !== "object" || !("key" in item) || typeof item.key !== "string") {
          return "";
        }
        const name = "name" in item && typeof item.name === "string" ? item.name : "";
        const type = "type" in item && typeof item.type === "string" ? item.type : "";
        const symbolKey = "symbolKey" in item && typeof item.symbolKey === "string" ? item.symbolKey : "";
        const symbol = symbols.find((entry) => entry.key === symbolKey);
        const metadata = functionSymbolMetadata(symbols, symbolKey);
        const symbolLabel = symbol?.label || symbolKey;
        return `<span class="function-chip">${imageHTML(functionSymbolImageData(metadata, "print"))}
          <strong>${escapeHTML(item.key)}</strong> ${escapeHTML(name || symbolLabel || type)}
          <small>${escapeHTML([type, symbolLabel].filter(Boolean).join(" · "))}</small></span>`;
      }).filter(Boolean);
      if (chips.length > 0) return chips.join("");
      if (parsed.length === 0) return "–";
    }
  } catch {
    // Older entries store free text. Keep it verbatim instead of inventing default functions.
  }
  return `<span class="function-text">${escapeHTML(value)}</span>`;
}

export function buildExhibitionPrintHTML(
  list: ExhibitionList,
  entries: PrintEntry[],
  symbols: MasterDataEntry[],
  language: string,
  t: Translate,
  options: PrintOptions = { includeImages: true }
) {
  const de = language === "de";
  const yesNo = (value: boolean) => t(value ? "exhibition.yes" : "exhibition.no");
  const fields = (values: [string, string | undefined][]) => values.map(([label, value]) =>
    `<div class="field"><dt>${escapeHTML(label)}</dt><dd>${escapeHTML(value || "–")}</dd></div>`
  ).join("");
  const date = (value: string) => value
    ? new Intl.DateTimeFormat(de ? "de-DE" : "en-GB").format(new Date(`${value}T12:00:00`)) : "–";
  const statusLabels = de
    ? { draft: "Entwurf", open: "Offen", locked: "Gesperrt", running: "Läuft", completed: "Abgeschlossen", archived: "Archiviert" }
    : { draft: "Draft", open: "Open", locked: "Locked", running: "Running", completed: "Completed", archived: "Archived" };
  const entryStatusLabels = de
    ? { ready: "Bereit", addressConflict: "Adresskonflikt", missing: "Angaben fehlen", check: "Prüfen", unavailable: "Nicht verfügbar" }
    : { ready: "Ready", addressConflict: "Address conflict", missing: "Missing data", check: "Check", unavailable: "Unavailable" };
  const days = (value: string) => (value || "all").split(",").map((scope) => {
    const day = scope.trim();
    if (day === "all") return t("exhibition.dayScope.all");
    const number = /^day(\d+)$/.exec(day)?.[1];
    return number ? `${de ? "Tag" : "Day"} ${number}` : day;
  }).join(", ");
  const rows = entries.map((entry, index) => `<tbody class="entry">
    <tr>
      <td class="identity">
        <h2>${index + 1}. ${escapeHTML(entry.locomotiveName)}</h2>
        ${options.includeImages ? imageHTML(entry.imageUrl) : ""}
        <dl>${fields([
          [t("exhibition.owner"), entry.owner],
          [t("exhibition.dayScope"), days(entry.dayScope)],
          [de ? "Verfügbarkeit" : "Availability", entry.availability === "unavailable"
            ? (de ? "Nicht verfügbar" : "Unavailable") : (de ? "Verfügbar" : "Available")],
          ["Status", entry.status ? entryStatusLabels[entry.status] : undefined]
        ])}</dl>
      </td>
      <td><dl>${fields([
        [t("exhibition.manufacturer"), entry.manufacturer], [t("exhibition.series"), entry.series],
        [t("exhibition.gattung"), entry.gattung], [t("exhibition.epoch"), entry.epoch],
        [t("exhibition.railwayCompany"), entry.railwayCompany]
      ])}</dl></td>
      <td><dl>${fields([
        [de ? "Digitaldecoder" : "Digital decoder", yesNo(entry.dtDecoder)],
        [t("exhibition.decoderType"), entry.decoderType], [t("exhibition.adapter"), entry.adapter],
        [t("exhibition.dccAddress"), entry.decoderNumber], [t("exhibition.sxAddress"), entry.sxAddress],
        [de ? "Zentrale / Schnittstelle" : "Command station / interface", entry.interfaceName],
        [t("exhibition.analog"), yesNo(entry.analog)]
      ])}</dl></td>
    </tr>
    <tr><td colspan="3"><h3>${escapeHTML(t("exhibition.functionKeys"))}</h3>
      <div class="functions">${printFunctionChips(entry.functionKeys, symbols)}</div></td></tr>
    <tr><td colspan="3"><h3>${escapeHTML(t("exhibition.notesSection"))}</h3>
      <p class="notes">${escapeHTML(entry.notes || "–")}</p></td></tr>
  </tbody>`).join("");
  const dateRange = [date(list.date), list.endDate && list.endDate !== list.date ? date(list.endDate) : ""]
    .filter(Boolean).join(" – ");
  return `<!doctype html><html lang="${escapeHTML(language)}"><head><meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src http: https: data:; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'" />
    <title>${escapeHTML(list.designation)} · RailKeeper</title>
    <style>
      @page { size: A4 landscape; margin: 12mm; }
      * { box-sizing: border-box; }
      body { margin: 0; color: #111; background: #fff; font: 12px/1.4 Arial, sans-serif; }
      header { border-bottom: 2px solid #333; padding-bottom: 8px; margin-bottom: 12px; }
      h1 { font-size: 20px; margin: 4px 0; }
      h2 { font-size: 14px; margin: 0 0 8px; break-after: avoid; }
      h3 { font-size: 12px; margin: 0 0 4px; break-after: avoid; }
      p, dl, dd { margin: 0; }
      table { width: 100%; border-collapse: collapse; table-layout: fixed; }
      thead { display: table-header-group; }
      th, td { text-align: left; vertical-align: top; padding: 8px; overflow-wrap: anywhere; }
      th { border-block: 1px solid #555; }
      .entry { break-inside: avoid; }
      .entry + .entry { border-top: 2px solid #555; }
      .entry td { border-bottom: 1px solid #ccc; }
      .field { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1.2fr); gap: 8px; margin: 2px 0; }
      dt { color: #444; }
      dd { font-weight: 600; white-space: pre-wrap; }
      .identity img { width: 100%; height: 72px; object-fit: contain; object-position: left; margin-bottom: 4px; }
      .functions { display: flex; flex-wrap: wrap; gap: 4px 12px; }
      .function-chip { display: inline-flex; flex-wrap: wrap; align-items: center; gap: 4px; max-width: 100%; break-inside: avoid; }
      .function-chip img { width: 16px; height: 16px; object-fit: contain; }
      .function-chip small { color: #444; }
      .notes, .function-text { white-space: pre-wrap; overflow-wrap: anywhere; }
      .event-notes { margin-top: 8px; }
      @media screen and (max-width: 600px) {
        thead { display: none; }
        table, tbody, tr, td { display: block; width: 100%; }
      }
    </style></head><body>
    <header><p>RailKeeper · ${escapeHTML(t("exhibition.title"))}</p>
      <h1>${escapeHTML(list.designation)}</h1>
      <p>${escapeHTML(dateRange)} · ${escapeHTML(list.location || "–")} · ${escapeHTML(statusLabels[list.status])}
        · ${escapeHTML(t("exhibition.entriesCount", { count: entries.length }))}</p>
      ${list.description ? `<p class="event-notes notes">${escapeHTML(list.description)}</p>` : ""}
      ${list.organizationNotes ? `<div class="event-notes"><h3>${de ? "Organisationshinweise" : "Organisation notes"}</h3><p class="notes">${escapeHTML(list.organizationNotes)}</p></div>` : ""}
      ${list.lockReason ? `<div class="event-notes"><h3>${de ? "Sperrgrund" : "Lock reason"}</h3><p class="notes">${escapeHTML(list.lockReason)}</p></div>` : ""}
    </header>
    <table><thead><tr><th>${de ? "Fahrzeug / Betrieb" : "Vehicle / operation"}</th>
      <th>${de ? "Modelldaten" : "Model data"}</th><th>${escapeHTML(t("exhibition.controlData"))}</th></tr></thead>
      ${rows || `<tbody><tr><td colspan="3">${escapeHTML(t("exhibition.printEmpty"))}</td></tr></tbody>`}
    </table></body></html>`;
}

export function printList(
  list: ExhibitionList,
  entries: PrintEntry[],
  symbols: MasterDataEntry[],
  language: string,
  t: Translate,
  options: PrintOptions = { includeImages: true }
) {
  const html = buildExhibitionPrintHTML(list, entries, symbols, language, t, options);
  return new Promise<void>((resolve, reject) => {
    const frame = document.createElement("iframe");
    frame.title = list.designation;
    frame.setAttribute("aria-hidden", "true");
    frame.style.cssText = "position:fixed;width:0;height:0;border:0;opacity:0;pointer-events:none";
    let cleanupTimer: number;
    const cleanup = () => {
      window.clearTimeout(cleanupTimer);
      frame.remove();
    };
    cleanupTimer = window.setTimeout(() => {
      cleanup();
      reject(new Error(language === "de" ? "Druckansicht konnte nicht geladen werden." : "Print view could not be loaded."));
    }, 15_000);
    frame.onload = () => {
      const printWindow = frame.contentWindow;
      if (!printWindow) {
        cleanup();
        reject(new Error(language === "de" ? "Druckfenster ist nicht verfügbar." : "Print window is unavailable."));
        return;
      }
      window.clearTimeout(cleanupTimer);
      // Load includes images. Keep the frame until the native preview has closed.
      printWindow.addEventListener("afterprint", cleanup, { once: true });
      cleanupTimer = window.setTimeout(cleanup, 300_000);
      try {
        printWindow.focus();
        printWindow.print();
        resolve();
      } catch (error) {
        cleanup();
        reject(error);
      }
    };
    frame.srcdoc = html;
    document.body.appendChild(frame);
  });
}
