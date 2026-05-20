import { Pencil, Printer, QrCode } from "lucide-react";
import { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { sourceShortLink } from "./articleSearch";
import { formatDate, formatEuro, formatFileSize } from "./vehicleFormat";
import {
  PendingArticleImage,
  previewImageUrl,
  primaryImage,
  vehicleImageToPending
} from "./vehicleTransforms";

type VehicleViewField = {
  label: string;
  value?: string | number | boolean;
  showFalse?: boolean;
  href?: string;
};

function viewValue(value?: string | number | boolean) {
  if (typeof value === "boolean") return value ? "Ja" : "Nein";
  if (value === 0) return "0";
  return String(value || "").trim();
}

function hasViewValue(field: VehicleViewField) {
  if (typeof field.value === "boolean") return field.value || field.showFalse;
  if (typeof field.value === "number") return true;
  return Boolean(String(field.value || "").trim());
}

function VehicleViewSection({ title, fields }: { title: string; fields: VehicleViewField[] }) {
  const visibleFields = fields.filter(hasViewValue);
  if (visibleFields.length === 0) return null;
  return (
    <section className="vehicle-view-section">
      <h3>{title}</h3>
      <dl>
        {visibleFields.map((field) => (
          <div key={`${title}-${field.label}`}>
            <dt>{field.label}</dt>
            <dd>
              {field.href ? (
                <a href={field.href} target="_blank" rel="noreferrer">
                  {viewValue(field.value)}
                </a>
              ) : (
                viewValue(field.value)
              )}
            </dd>
          </div>
        ))}
      </dl>
    </section>
  );
}

export function VehicleReadOnlyView({
  vehicle,
  onEdit,
  onPrint,
  onQr,
  onPreviewImage
}: {
  vehicle: Vehicle;
  onEdit: () => void;
  onPrint: () => void;
  onQr: () => void;
  onPreviewImage: (image: PendingArticleImage) => void;
}) {
  const { t } = useI18n();
  const image = primaryImage(vehicle.images);
  const configuredFunctions = (vehicle.functions || []).filter((item) => item.name || item.symbolKey || item.notes);
  const activeMaintenance = (vehicle.maintenance || []).filter((item) => item.kind || item.status || item.dueDate || item.completedAt || item.notes);
  const cvValues = vehicle.cvValues || [];
  const cvFiles = vehicle.cvFiles || [];
  const attachments = vehicle.attachments || [];
  const images = vehicle.images || [];
  const canShowQr = Boolean(String(vehicle.inventoryNumber || vehicle.name || "").trim());
  const qrButtonTitle = canShowQr ? t("vehicles.detail.qrShow") : t("vehicles.qr.missingInput");

  return (
    <div className="modal-body vehicle-read-view">
      <section className="vehicle-read-hero">
        <div className="vehicle-read-image">
          {image?.url ? <img src={previewImageUrl(image)} alt="" /> : <span>{t("exhibition.noPreview")}</span>}
        </div>
        <div className="vehicle-read-title">
          <p className="eyebrow">Fahrzeugansicht</p>
          <h2>{vehicle.name || vehicle.inventoryNumber}</h2>
          <p>{[vehicle.manufacturer, vehicle.articleNumber, vehicle.gauge, vehicle.epoch].filter(Boolean).join(" · ")}</p>
          <div className="vehicle-read-chips">
            <span>{vehicle.inventoryNumber}</span>
            {vehicle.category && <span>{vehicle.category}</span>}
            {vehicle.gattung && <span>{vehicle.gattung}</span>}
            <span>{vehicle.digital ? "Digital" : "Analog"}</span>
            {vehicle.exhibitionReady && <span>Messe tauglich</span>}
            {vehicle.exhibition && <span>Ausstellung</span>}
          </div>
        </div>
        <div className="vehicle-read-actions">
          <button type="button" className="icon-button" onClick={onEdit} aria-label={t("vehicles.edit")} title={t("vehicles.edit")}>
            <Pencil size={16} />
          </button>
          <button type="button" className="icon-button" onClick={onPrint} aria-label={t("vehicles.report.open")} title={t("vehicles.report.open")}>
            <Printer size={16} />
          </button>
          <button type="button" className="icon-button" onClick={onQr} aria-label={qrButtonTitle} title={qrButtonTitle} disabled={!canShowQr}>
            <QrCode size={16} />
          </button>
        </div>
      </section>

      <div className="vehicle-view-grid">
        <VehicleViewSection
          title="Produkt"
          fields={[
            { label: "Hersteller", value: vehicle.manufacturer },
            { label: "Artikel-Nr.", value: vehicle.articleNumber },
            { label: "EAN", value: vehicle.ean },
            { label: "Listenpreis", value: formatEuro(vehicle.listPrice) },
            { label: "Produktionszeit", value: vehicle.productionPeriod },
            { label: "Artikelquelle", value: sourceShortLink(vehicle.articleSourceUrl), href: vehicle.articleSourceUrl }
          ]}
        />
        <VehicleViewSection
          title="Modell"
          fields={[
            { label: "Bezeichnung", value: vehicle.name },
            { label: "Spurweite", value: vehicle.gauge },
            { label: "Epoche", value: vehicle.epoch },
            { label: "Bahngesellschaft", value: vehicle.railwayCompany },
            { label: "Kategorie", value: vehicle.category },
            { label: "Gattung", value: vehicle.gattung },
            { label: "Baureihe", value: vehicle.series },
            { label: "Fahrzeug-Nr.", value: vehicle.vehicleNumber },
            { label: "Beschreibung", value: vehicle.description },
            { label: "Messe tauglich", value: vehicle.exhibitionReady },
            { label: "Ausstellung", value: vehicle.exhibition }
          ]}
        />
        <VehicleViewSection
          title="Details"
          fields={[
            { label: "Länge", value: vehicle.lengthMm ? `${vehicle.lengthMm} mm` : "" },
            { label: "Gewicht", value: vehicle.weightG ? `${vehicle.weightG} g` : "" },
            { label: "Farbe", value: vehicle.color },
            { label: "Beschriftung", value: vehicle.lettering },
            { label: "Beladung", value: vehicle.load },
            { label: "Inneneinrichtung", value: vehicle.interior },
            { label: "Achsen", value: vehicle.axles },
            { label: "Anzahl Achsen", value: vehicle.axleCount },
            { label: "Haftreifen", value: vehicle.tractionTireCount },
            { label: "Radsatz", value: vehicle.wheelset },
            { label: "Kupplung", value: vehicle.couplingSame ? "Vorne und hinten gleich" : "" },
            { label: "Kupplung vorne", value: vehicle.couplingFront },
            { label: "Kupplung hinten", value: vehicle.couplingRear },
            { label: "Stromaufnahme", value: vehicle.powerPickup },
            { label: "Fahrlicht", value: vehicle.headlightsEnabled },
            { label: "Fahrlicht Beschreibung", value: vehicle.headlightsDescription },
            { label: "Antrieb", value: vehicle.driveEnabled },
            { label: "Antrieb Beschreibung", value: vehicle.driveDescription },
            { label: "Beleuchtung", value: vehicle.lightingEnabled },
            { label: "Beleuchtung Beschreibung", value: vehicle.lightingDescription },
            { label: "Soundgenerator", value: vehicle.soundGeneratorEnabled },
            { label: "Sound Beschreibung", value: vehicle.soundGeneratorDescription },
            { label: "Rauchgenerator", value: vehicle.smokeGeneratorEnabled },
            { label: "Rauch Beschreibung", value: vehicle.smokeGeneratorDescription }
          ]}
        />
        <VehicleViewSection
          title="Fahrzeug"
          fields={[
            { label: "Erwerb", value: vehicle.acquisitionType },
            { label: "von/bei", value: vehicle.acquiredFrom },
            { label: "Preis", value: formatEuro(vehicle.purchasePrice) },
            { label: "Datum", value: vehicle.purchaseDate ? formatDate(vehicle.purchaseDate) : "" },
            { label: "Standort", value: vehicle.storageLocation },
            { label: "Details", value: vehicle.storageDetails },
            { label: "Zustand", value: vehicle.condition },
            { label: "Zustand Details", value: vehicle.conditionDetails },
            { label: "Verpackung", value: vehicle.packaging },
            { label: "Zusatzinformationen", value: vehicle.additionalInfo }
          ]}
        />
        <VehicleViewSection
          title="Steuerung"
          fields={[
            { label: "Digital", value: vehicle.digital },
            { label: "Decoder-Nr.", value: vehicle.digitalDecoderNumber },
            { label: "Decoder-Typ", value: vehicle.decoderType },
            { label: "DT-Decoder", value: vehicle.dtDecoder },
            { label: "DT Decoder-Nr.", value: vehicle.dtDecoderNumber },
            { label: "ABC Bremsen", value: vehicle.abcBrakes },
            { label: "Adapter / Schnittstelle", value: vehicle.adapter },
            { label: "QR-Code aktiv", value: vehicle.qrCodeEnabled }
          ]}
        />
      </div>

      {configuredFunctions.length > 0 && (
        <section className="vehicle-view-section vehicle-view-wide">
          <h3>Funktionstasten</h3>
          <div className="vehicle-view-list">
            {configuredFunctions.map((item) => (
              <article key={item.functionKey}>
                <strong>{item.functionKey}</strong>
                <span>{[item.name, item.functionType, item.mode, item.notes].filter(Boolean).join(" · ")}</span>
              </article>
            ))}
          </div>
        </section>
      )}

      {images.length > 0 && (
        <section className="vehicle-view-section vehicle-view-wide">
          <h3>Bilder</h3>
          <div className="vehicle-view-gallery">
            {images.map((item) => (
              <figure key={item.id}>
                <button type="button" className="vehicle-view-image-button" onClick={() => onPreviewImage(vehicleImageToPending(item))}>
                  <img src={previewImageUrl(item)} alt="" />
                </button>
                <figcaption>{[item.isPrimary ? "Hauptbild" : "Alternativbild", item.title].filter(Boolean).join(" · ") || item.fileName || "Bild"}</figcaption>
              </figure>
            ))}
          </div>
        </section>
      )}

      {(attachments.length > 0 || activeMaintenance.length > 0 || cvValues.length > 0 || cvFiles.length > 0) && (
        <div className="vehicle-view-grid">
          {activeMaintenance.length > 0 && (
            <section className="vehicle-view-section">
              <h3>Wartung</h3>
              <div className="vehicle-view-list">
                {activeMaintenance.map((item) => (
                  <article key={item.id}>
                    <strong>{item.kind}</strong>
                    <span>{[item.status, item.dueDate && `Fällig ${formatDate(item.dueDate)}`, item.completedAt && `Erledigt ${formatDate(item.completedAt)}`, item.notes].filter(Boolean).join(" · ")}</span>
                  </article>
                ))}
              </div>
            </section>
          )}
          {attachments.length > 0 && (
            <section className="vehicle-view-section">
              <h3>Beilagen</h3>
              <div className="vehicle-view-list">
                {attachments.map((item) => (
                  <article key={item.id}>
                    <strong>{item.originalName || item.fileName}</strong>
                    <span>{[item.category, item.description, formatFileSize(item.sizeBytes)].filter(Boolean).join(" · ")}</span>
                  </article>
                ))}
              </div>
            </section>
          )}
          {cvValues.length > 0 && (
            <section className="vehicle-view-section">
              <h3>CV-Werte</h3>
              <div className="vehicle-view-list compact">
                {cvValues.slice(0, 12).map((item) => (
                  <article key={item.id}>
                    <strong>CV {item.cvNumber}</strong>
                    <span>{[String(item.value), item.category, item.protocol, item.decoderProfile, item.description].filter(Boolean).join(" · ")}</span>
                  </article>
                ))}
              </div>
            </section>
          )}
          {cvFiles.length > 0 && (
            <section className="vehicle-view-section">
              <h3>CV-Dateien</h3>
              <div className="vehicle-view-list">
                {cvFiles.map((item) => (
                  <article key={item.id}>
                    <strong>{item.originalName || item.fileName}</strong>
                    <span>{[item.decoderProfile, item.description, formatFileSize(item.sizeBytes)].filter(Boolean).join(" · ")}</span>
                  </article>
                ))}
              </div>
            </section>
          )}
        </div>
      )}
    </div>
  );
}
