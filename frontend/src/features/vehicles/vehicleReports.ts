import { Vehicle } from "../../shared/api";

export type InventoryReportAssets = Record<string, { qrCode?: string }>;

const sortLabels: Record<string, string> = {
  inventoryNumber: "Inventarnummer",
  manufacturer: "Hersteller",
  articleNumber: "Artikel-Nr.",
  name: "Bezeichnung",
  gauge: "Spurweite",
  epoch: "Epoche",
  category: "Kategorie"
};

function reportValue(value?: string | number | boolean) {
  if (typeof value === "boolean") return value ? "Ja" : "Nein";
  if (value === 0) return "0";
  return String(value || "-");
}

function escapeHtml(value?: string | number | boolean) {
  return reportValue(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function hasReportValue(value?: string | number | boolean) {
  if (typeof value === "boolean") return true;
  if (typeof value === "number") return true;
  return Boolean(String(value || "").trim());
}

function reportField(label: string, value?: string | number | boolean) {
  if (!hasReportValue(value)) return "";
  const labelText = label.trim().endsWith(":") ? label.trim() : `${label.trim()}:`;
  return `<div class="field"><span>${escapeHtml(labelText)}</span><strong>${escapeHtml(value)}</strong></div>`;
}

function primaryImage(images?: { url: string; thumbnailUrl?: string; isPrimary?: boolean }[]) {
  return images?.find((image) => image.isPrimary) || images?.[0];
}

function previewImageUrl(image?: { url: string; thumbnailUrl?: string }) {
  return image?.thumbnailUrl || image?.url || "";
}

function sourceDisplayName(rawUrl: string) {
  try {
    const parsed = new URL(rawUrl);
    return parsed.hostname.replace(/^www\./, "") || rawUrl;
  } catch {
    return rawUrl;
  }
}

function sourceShortLink(rawUrl?: string) {
  if (!rawUrl) return "";
  return sourceDisplayName(rawUrl);
}

function formatEuro(value?: string | number) {
  const text = String(value || "").trim();
  if (!text) return "";
  const normalized = Number(text.replace(",", "."));
  if (!Number.isFinite(normalized)) return text;
  return new Intl.NumberFormat("de-DE", { style: "currency", currency: "EUR" }).format(normalized);
}

function formatFileSize(size: number) {
  if (!Number.isFinite(size) || size <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  let value = size;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit += 1;
  }
  return `${value.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`;
}

function formatDate(value?: string) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleDateString("de-DE");
}

function formatMaintenanceCost(cost?: string) {
  if (!cost) return "";
  return formatEuro(cost);
}

function reportImage(vehicle: Vehicle, includeImages: boolean) {
  if (!includeImages) return "";
  const image = primaryImage(vehicle.images);
  if (!image?.url) {
    return "";
  }
  return `<img class="vehicle-image" src="${escapeHtml(previewImageUrl(image))}" alt="">`;
}

function reportSection(title: string, fields: string[]) {
  const body = fields.filter(Boolean).join("");
  if (!body) return "";
  return `<section class="detail-section"><h3>${escapeHtml(title)}</h3><div class="field-grid">${body}</div></section>`;
}

function reportJoined(values: Array<string | undefined | false>) {
  return values.filter(Boolean).join(" · ");
}

function reportImageGallery(vehicle: Vehicle, includeImages: boolean) {
  if (!includeImages || !vehicle.images?.length) return "";
  const images = vehicle.images
    .map((image) => {
      const meta = reportJoined([
        image.isPrimary ? "Hauptbild" : "Alternativbild",
        image.title,
        image.fileName,
        image.mimeType,
        image.sourceUrl
      ]);
      return `
        <figure class="image-tile">
          <img src="${escapeHtml(previewImageUrl(image))}" alt="">
          ${meta ? `<figcaption>${escapeHtml(meta)}</figcaption>` : ""}
        </figure>
      `;
    })
    .join("");
  return `<section class="detail-section image-section"><h3>Bilder</h3><div class="image-grid">${images}</div></section>`;
}

function vehicleOverviewRow(vehicle: Vehicle, assets: InventoryReportAssets, includeImages: boolean, includeQRCode: boolean) {
  const image = reportImage(vehicle, includeImages);
  const qrCode = includeQRCode && assets[vehicle.id]?.qrCode
    ? `<img class="qr-code" src="${escapeHtml(assets[vehicle.id].qrCode)}" alt="">`
    : "";
  return `
    <section class="overview-row">
      <div class="overview-media">${image || `<div class="image-spacer"></div>`}</div>
      <div class="overview-main">
        <h3>${escapeHtml(vehicle.name || vehicle.inventoryNumber)}</h3>
        <div class="overview-fields">
          ${reportField("Hersteller", vehicle.manufacturer)}
          ${reportField("Artikel-Nr.", vehicle.articleNumber)}
          ${reportField("Gattung", vehicle.gattung || vehicle.category)}
          ${reportField("Baureihe", vehicle.series)}
          ${reportField("Betriebs-Nr.", vehicle.vehicleNumber)}
        </div>
      </div>
      <div class="overview-identity">
        ${reportField("Inventar-Nr.", vehicle.inventoryNumber)}
        ${reportField("Decoder-Nr.", vehicle.digitalDecoderNumber || vehicle.dtDecoderNumber)}
      </div>
      <div class="overview-qr">${qrCode}</div>
    </section>
  `;
}

function vehicleDetailReport(vehicle: Vehicle, assets: InventoryReportAssets, includeImages: boolean, includeQRCode: boolean) {
  const image = reportImage(vehicle, includeImages);
  const qrCode = includeQRCode && assets[vehicle.id]?.qrCode
    ? `<img class="qr-code" src="${escapeHtml(assets[vehicle.id].qrCode)}" alt="">`
    : "";
  const functions = (vehicle.functions || [])
    .filter((item) => item.name || item.symbolKey || item.notes)
    .map((item) => reportField(item.functionKey, [item.name, item.functionType, item.mode, item.notes].filter(Boolean).join(" · ")));
  const maintenance = (vehicle.maintenance || [])
    .filter((item) => item.kind || item.status || item.dueDate || item.completedAt || item.cost || item.notes)
    .map((item) => reportField(item.kind || "Wartung", [item.status, item.dueDate && `Fällig: ${formatDate(item.dueDate)}`, item.completedAt && `Erledigt: ${formatDate(item.completedAt)}`, item.cost && formatMaintenanceCost(item.cost), item.notes].filter(Boolean).join(" · ")));
  const attachments = (vehicle.attachments || [])
    .map((item) => reportField(item.originalName || item.fileName, [item.category, item.description, formatFileSize(item.sizeBytes)].filter(Boolean).join(" · ")));
  const cvValues = (vehicle.cvValues || [])
    .map((item) => reportField(`CV ${item.cvNumber}`, [String(item.value), item.category, item.protocol, item.decoderProfile, item.description].filter(Boolean).join(" · ")));
  const cvFiles = (vehicle.cvFiles || [])
    .map((item) => reportField(item.originalName || item.fileName, [item.decoderProfile, item.description, formatFileSize(item.sizeBytes)].filter(Boolean).join(" · ")));
  const externalMappings = (vehicle.externalMappings || [])
    .map((item) => reportField(item.provider, reportJoined([item.externalId, item.externalName, item.externalAddress, item.externalProtocol, item.syncStatus, item.lastSeenAt && formatDate(item.lastSeenAt)])));

  return `
    <article class="detail-card">
      <header class="detail-card-head">
        <div class="detail-media">${image}</div>
        <div>
          <h2>${escapeHtml(vehicle.name || vehicle.inventoryNumber)}</h2>
          <p>${escapeHtml([vehicle.manufacturer, vehicle.articleNumber, vehicle.gauge, vehicle.epoch].filter(Boolean).join(" · "))}</p>
        </div>
        <div class="detail-identity">
          ${reportField("Inventar-Nr.", vehicle.inventoryNumber)}
          ${reportField("Decoder-Nr.", vehicle.digitalDecoderNumber || vehicle.dtDecoderNumber)}
        </div>
        <div class="detail-qr">${qrCode}</div>
      </header>
      ${reportSection("Produkt", [
        reportField("Hersteller", vehicle.manufacturer),
        reportField("Artikel-Nr.", vehicle.articleNumber),
        reportField("Artikelquelle", sourceShortLink(vehicle.articleSourceUrl)),
        reportField("EAN", vehicle.ean),
        reportField("Listenpreis", formatEuro(vehicle.listPrice)),
        reportField("Produktionszeit", vehicle.productionPeriod),
        reportField("Erfasst am", formatDate(vehicle.createdAt)),
        reportField("Aktualisiert am", formatDate(vehicle.updatedAt))
      ])}
      ${reportSection("Modell", [
        reportField("Bezeichnung", vehicle.name),
        reportField("Spurweite / Epoche", [vehicle.gauge, vehicle.epoch].filter(Boolean).join(" / ")),
        reportField("Bahngesellschaft", vehicle.railwayCompany),
        reportField("Kategorie", vehicle.category),
        reportField("Gattung", vehicle.gattung),
        reportField("Baureihe", vehicle.series),
        reportField("Betriebs-Nr.", vehicle.vehicleNumber),
        reportField("Messe tauglich", vehicle.exhibitionReady),
        reportField("Ausstellung", vehicle.exhibition),
        reportField("QR-Code aktiv", vehicle.qrCodeEnabled)
      ])}
      ${reportSection("Details", [
        reportField("Länge", vehicle.lengthMm ? `${vehicle.lengthMm} mm` : ""),
        reportField("Gewicht", vehicle.weightG ? `${vehicle.weightG} g` : ""),
        reportField("Farbe", vehicle.color),
        reportField("Beschriftung", vehicle.lettering),
        reportField("Beladung", vehicle.load),
        reportField("Inneneinrichtung", vehicle.interior),
        reportField("Achsen", vehicle.axles),
        reportField("Anzahl Achsen", vehicle.axleCount),
        reportField("Haftreifen", vehicle.tractionTireCount),
        reportField("Radsatz", vehicle.wheelset),
        reportField("Kupplung", vehicle.couplingSame ? "Vorne und hinten gleich" : ""),
        reportField("Kupplung vorne", vehicle.couplingFront),
        reportField("Kupplung hinten", vehicle.couplingRear),
        reportField("Stromaufnahme", vehicle.powerPickup),
        reportField("Fahrlicht", vehicle.headlightsEnabled),
        reportField("Fahrlicht Beschreibung", vehicle.headlightsDescription),
        reportField("Antrieb", vehicle.driveEnabled),
        reportField("Antrieb Beschreibung", vehicle.driveDescription),
        reportField("Beleuchtung", vehicle.lightingEnabled),
        reportField("Beleuchtung Beschreibung", vehicle.lightingDescription),
        reportField("Soundgenerator", vehicle.soundGeneratorEnabled),
        reportField("Sound Beschreibung", vehicle.soundGeneratorDescription),
        reportField("Rauchgenerator", vehicle.smokeGeneratorEnabled),
        reportField("Rauch Beschreibung", vehicle.smokeGeneratorDescription)
      ])}
      ${reportSection("Fahrzeug", [
        reportField("Erwerb", vehicle.acquisitionType),
        reportField("von/bei", vehicle.acquiredFrom),
        reportField("Preis", formatEuro(vehicle.purchasePrice)),
        reportField("Datum", formatDate(vehicle.purchaseDate)),
        reportField("Standort", vehicle.storageLocation),
        reportField("Details", vehicle.storageDetails),
        reportField("Zustand", vehicle.condition),
        reportField("Zustand Details", vehicle.conditionDetails),
        reportField("Verpackung", vehicle.packaging),
        reportField("Zusatzinformationen", vehicle.additionalInfo)
      ])}
      ${reportSection("Steuerung", [
        reportField("Digital", vehicle.digital),
        reportField("Decoder-Nr.", vehicle.digitalDecoderNumber),
        reportField("Decoder-Typ", vehicle.decoderType),
        reportField("DT-Decoder", vehicle.dtDecoder),
        reportField("DT Decoder-Nr.", vehicle.dtDecoderNumber),
        reportField("ABC Bremsen", vehicle.abcBrakes),
        reportField("Adapter / Schnittstelle", vehicle.adapter)
      ])}
      ${reportSection("Funktionstasten", functions)}
      ${reportSection("Wartung", maintenance)}
      ${reportSection("CV-Werte", cvValues)}
      ${reportSection("CV-Dateien", cvFiles)}
      ${reportImageGallery(vehicle, includeImages)}
      ${reportSection("Beilagen", attachments)}
      ${reportSection("Externe Zuordnung", externalMappings)}
    </article>
  `;
}

export function inventoryReportHtml(
  vehicles: Vehicle[],
  query: string,
  sort: { key: string; direction: string },
  options: { mode: string; title: string; includeQRCode: boolean; includeImages: boolean },
  assets: InventoryReportAssets
) {
  const now = new Date();
  const modeTitle = options.mode === "details" ? "Detailliste" : "Übersichtsliste";
  const overviewRows = vehicles.map((vehicle) => vehicleOverviewRow(vehicle, assets, options.includeImages, options.includeQRCode)).join("");
  const detailRows = options.mode === "details"
    ? vehicles.map((vehicle) => vehicleDetailReport(vehicle, assets, options.includeImages, options.includeQRCode)).join("")
    : "";
  const sortLabel = sortLabels[sort.key] || sort.key;

  return `
<!doctype html>
<html lang="de">
  <head>
    <meta charset="utf-8">
    <title>RailKeeper Bestand</title>
    <style>
      :root { color: #101820; font-family: "Segoe UI", Arial, sans-serif; font-size: 11px; }
      * { box-sizing: border-box; }
      body { margin: 18mm 14mm 20mm; background: #fff; }
      header { display: grid; grid-template-columns: 1fr 62px; gap: 18px; align-items: start; padding-bottom: 10px; border-bottom: 1.5px solid #67b532; }
      h1 { margin: 0; font-size: 14px; letter-spacing: .02em; }
      h2 { margin: 0; font-size: 15px; }
      h3 { margin: 0; font-size: 12px; }
      p { margin: 4px 0 0; color: #4a6268; }
      .brand-mark { width: 48px; justify-self: end; }
      .report-subtitle { margin-top: 3px; font-size: 13px; font-weight: 800; color: #101820; }
      .report-meta { display: flex; gap: 14px; margin: 8px 0 14px; color: #60747b; font-size: 10px; }
      .overview-row { display: grid; grid-template-columns: 88px 1fr 110px 66px; gap: 14px; align-items: start; padding: 10px 0; border-bottom: 1px solid #101820; page-break-inside: avoid; }
      .overview-main h3 { margin-bottom: 10px; }
      .overview-fields { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 4px 18px; }
      .overview-identity .field { margin-bottom: 8px; }
      .overview-qr, .detail-qr { display: grid; place-items: start center; min-height: 58px; }
      .qr-code { width: 58px; height: 58px; object-fit: contain; }
      .vehicle-image { width: 78px; max-height: 54px; object-fit: contain; }
      .image-spacer { width: 78px; height: 1px; }
      .field { display: grid; grid-template-columns: minmax(88px, max-content) minmax(0, 1fr); gap: 6px; align-items: baseline; min-width: 0; }
      .field span { display: inline; color: #344a50; font-size: 9px; font-weight: 650; line-height: 1.25; }
      .field strong { display: inline; min-width: 0; font-size: 10px; font-weight: 800; line-height: 1.25; overflow-wrap: anywhere; }
      .overview-row .field { display: block; }
      .overview-row .field span,
      .overview-row .field strong { display: block; }
      .overview-row .field strong { margin-top: 2px; }
      .detail-card { page-break-inside: avoid; padding: 12px 0 16px; border-bottom: 1px solid #101820; }
      .detail-card-head { display: grid; grid-template-columns: 132px 1fr 120px 72px; gap: 16px; align-items: start; margin-bottom: 14px; }
      .detail-card-head .vehicle-image { width: 118px; max-height: 74px; }
      .detail-identity .field { display: grid; grid-template-columns: 1fr; gap: 2px; margin-bottom: 8px; text-align: center; }
      .detail-identity .field span,
      .detail-identity .field strong { display: block; }
      .detail-section { display: grid; grid-template-columns: 78px 1fr; gap: 18px; margin: 7px 0; }
      .detail-section h3 { font-size: 10px; }
      .field-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 3px 24px; }
      .image-section { align-items: start; }
      .image-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 10px; }
      .image-tile { margin: 0; page-break-inside: avoid; }
      .image-tile img { width: 100%; height: 88px; object-fit: contain; border: 1px solid #d9e4df; border-radius: 4px; padding: 3px; }
      .image-tile figcaption { margin-top: 4px; color: #4a6268; font-size: 8px; line-height: 1.25; overflow-wrap: anywhere; }
      .description { border-left: 2px solid #67b532; padding-left: 8px; margin-top: 10px; color: #101820; white-space: pre-wrap; }
      .footer { position: fixed; left: 14mm; right: 14mm; bottom: 8mm; display: grid; grid-template-columns: 1fr 1fr 1fr; align-items: center; border-top: 1px solid #101820; padding-top: 6px; font-size: 9px; color: #101820; }
      .footer-center { text-align: center; font-weight: 700; }
      .footer-right { text-align: right; }
      .screen-actions { position: sticky; top: 0; display: flex; justify-content: flex-end; gap: 8px; margin-bottom: 12px; }
      button { border: 0; border-radius: 7px; padding: 10px 14px; background: #3c8eff; color: white; font-weight: 800; cursor: pointer; }
      @page { margin: 14mm; }
      @media print {
        body { margin: 0; }
        .screen-actions { display: none; }
        header { break-after: avoid; }
      }
    </style>
  </head>
  <body>
    <div class="screen-actions">
      <button onclick="window.print()">Drucken / Als PDF speichern</button>
    </div>
    <header>
      <div>
        <h1>${escapeHtml(options.title || "Fahrzeugsammlung")}</h1>
        <div class="report-subtitle">${escapeHtml(modeTitle)} / Allgemein</div>
      </div>
      <img class="brand-mark" src="/brand/railkeeper-mark.png" alt="">
    </header>
    <div class="report-meta">
      <span>${escapeHtml(vehicles.length)} Fahrzeuge</span>
      <span>${query.trim() ? `Filter: ${escapeHtml(query.trim())}` : "Alle Fahrzeuge"}</span>
      <span>Sortierung: ${escapeHtml(sortLabel)} ${sort.direction === "asc" ? "aufsteigend" : "absteigend"}</span>
    </div>
    ${options.mode === "details" ? detailRows : overviewRows}
    <footer class="footer">
      <span>${escapeHtml(now.toLocaleDateString("de-DE"))}</span>
      <span class="footer-center">RailKeeper</span>
      <span class="footer-right">${escapeHtml(modeTitle)}</span>
    </footer>
  </body>
</html>
`;
}

function writePrintWindow(printWindow: Window, html: string) {
  printWindow.document.open();
  printWindow.document.write(html);
  printWindow.document.close();
  printWindow.focus();
}

function printHtmlFallback(html: string) {
  const iframe = document.createElement("iframe");
  iframe.title = "RailKeeper PDF Report";
  iframe.style.position = "fixed";
  iframe.style.right = "0";
  iframe.style.bottom = "0";
  iframe.style.width = "0";
  iframe.style.height = "0";
  iframe.style.border = "0";
  iframe.style.opacity = "0";
  iframe.setAttribute("aria-hidden", "true");
  document.body.appendChild(iframe);

  const cleanup = () => {
    window.setTimeout(() => iframe.remove(), 1000);
  };
  iframe.onload = () => {
    iframe.contentWindow?.focus();
    iframe.contentWindow?.print();
    cleanup();
  };
  iframe.srcdoc = html;
}

export function openPrintDocument(html: string, name: string) {
  const printWindow = window.open("", name, "width=1180,height=860");
  if (printWindow) {
    writePrintWindow(printWindow, html);
    return;
  }
  printHtmlFallback(html);
}
