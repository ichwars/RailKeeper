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

type Translate = (key: string, values?: Record<string, string | number>) => string;

function viewValue(value: string | number | boolean | undefined, t: Translate) {
  if (typeof value === "boolean") return value ? t("common.yes") : t("common.no");
  if (value === 0) return "0";
  return String(value || "").trim();
}

function hasViewValue(field: VehicleViewField) {
  if (typeof field.value === "boolean") return field.value || field.showFalse;
  if (typeof field.value === "number") return true;
  return Boolean(String(field.value || "").trim());
}

function VehicleViewSection({ title, fields, t }: { title: string; fields: VehicleViewField[]; t: Translate }) {
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
                  {viewValue(field.value, t)}
                </a>
              ) : (
                viewValue(field.value, t)
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
  const separator = " · ";

  return (
    <div className="modal-body vehicle-read-view">
      <section className="vehicle-read-hero">
        <div className="vehicle-read-image">
          {image?.url ? <img src={previewImageUrl(image)} alt="" /> : <span>{t("exhibition.noPreview")}</span>}
        </div>
        <div className="vehicle-read-title">
          <p className="eyebrow">{t("vehicles.modal.view")}</p>
          <h2>{vehicle.name || vehicle.inventoryNumber}</h2>
          <p>{[vehicle.manufacturer, vehicle.articleNumber, vehicle.gauge, vehicle.epoch].filter(Boolean).join(separator)}</p>
          <div className="vehicle-read-chips">
            <span>{vehicle.inventoryNumber}</span>
            {vehicle.category && <span>{vehicle.category}</span>}
            {vehicle.gattung && <span>{vehicle.gattung}</span>}
            <span>{vehicle.digital ? "Digital" : "Analog"}</span>
            {vehicle.exhibitionReady && <span>{t("vehicle.field.exhibitionReady")}</span>}
            {vehicle.exhibition && <span>{t("vehicle.field.exhibition")}</span>}
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
          title={t("vehicles.product.title")}
          t={t}
          fields={[
            { label: t("vehicle.field.manufacturer"), value: vehicle.manufacturer },
            { label: t("vehicle.field.articleNumber"), value: vehicle.articleNumber },
            { label: t("vehicle.field.ean"), value: vehicle.ean },
            { label: t("vehicle.field.listPrice"), value: formatEuro(vehicle.listPrice) },
            { label: t("vehicle.field.productionPeriod"), value: vehicle.productionPeriod },
            { label: t("vehicles.source"), value: sourceShortLink(vehicle.articleSourceUrl), href: vehicle.articleSourceUrl }
          ]}
        />
        <VehicleViewSection
          title={t("vehicles.tab.model")}
          t={t}
          fields={[
            { label: t("vehicle.field.name"), value: vehicle.name },
            { label: t("vehicle.field.gauge"), value: vehicle.gauge },
            { label: t("vehicle.field.epoch"), value: vehicle.epoch },
            { label: t("vehicle.field.railwayCompany"), value: vehicle.railwayCompany },
            { label: t("vehicle.field.category"), value: vehicle.category },
            { label: t("vehicle.field.gattung"), value: vehicle.gattung },
            { label: t("vehicle.field.series"), value: vehicle.series },
            { label: t("vehicle.field.vehicleNumber"), value: vehicle.vehicleNumber },
            { label: t("vehicle.field.description"), value: vehicle.description },
            { label: t("vehicle.field.exhibitionReady"), value: vehicle.exhibitionReady },
            { label: t("vehicle.field.exhibition"), value: vehicle.exhibition }
          ]}
        />
        <VehicleViewSection
          title={t("vehicles.details.title")}
          t={t}
          fields={[
            { label: t("vehicle.field.lengthMm"), value: vehicle.lengthMm ? `${vehicle.lengthMm} mm` : "" },
            { label: t("vehicle.field.weightG"), value: vehicle.weightG ? `${vehicle.weightG} g` : "" },
            { label: t("vehicle.field.color"), value: vehicle.color },
            { label: t("vehicle.field.lettering"), value: vehicle.lettering },
            { label: t("vehicle.field.load"), value: vehicle.load },
            { label: t("vehicle.field.interior"), value: vehicle.interior },
            { label: t("vehicle.field.axles"), value: vehicle.axles },
            { label: t("vehicle.field.axleCount"), value: vehicle.axleCount },
            { label: t("vehicle.field.tractionTireCount"), value: vehicle.tractionTireCount },
            { label: t("vehicle.field.wheelset"), value: vehicle.wheelset },
            { label: t("vehicle.field.couplingSame"), value: vehicle.couplingSame ? t("vehicles.detail.couplingSameValue") : "" },
            { label: t("vehicle.field.couplingFront"), value: vehicle.couplingFront },
            { label: t("vehicle.field.couplingRear"), value: vehicle.couplingRear },
            { label: t("vehicle.field.powerPickup"), value: vehicle.powerPickup },
            { label: t("vehicle.field.headlightsEnabled"), value: vehicle.headlightsEnabled },
            { label: t("vehicle.field.headlightsDescription"), value: vehicle.headlightsDescription },
            { label: t("vehicle.field.driveEnabled"), value: vehicle.driveEnabled },
            { label: t("vehicle.field.driveDescription"), value: vehicle.driveDescription },
            { label: t("vehicle.field.lightingEnabled"), value: vehicle.lightingEnabled },
            { label: t("vehicle.field.lightingDescription"), value: vehicle.lightingDescription },
            { label: t("vehicle.field.soundGeneratorEnabled"), value: vehicle.soundGeneratorEnabled },
            { label: t("vehicle.field.soundGeneratorDescription"), value: vehicle.soundGeneratorDescription },
            { label: t("vehicle.field.smokeGeneratorEnabled"), value: vehicle.smokeGeneratorEnabled },
            { label: t("vehicle.field.smokeGeneratorDescription"), value: vehicle.smokeGeneratorDescription }
          ]}
        />
        <VehicleViewSection
          title={t("vehicles.vehicle.title")}
          t={t}
          fields={[
            { label: t("vehicle.field.acquisitionType"), value: vehicle.acquisitionType },
            { label: t("vehicle.field.acquiredFrom"), value: vehicle.acquiredFrom },
            { label: t("vehicle.field.purchasePrice"), value: formatEuro(vehicle.purchasePrice) },
            { label: t("vehicle.field.purchaseDate"), value: vehicle.purchaseDate ? formatDate(vehicle.purchaseDate) : "" },
            { label: t("vehicle.field.storageLocation"), value: vehicle.storageLocation },
            { label: t("vehicle.field.storageDetails"), value: vehicle.storageDetails },
            { label: t("vehicle.field.condition"), value: vehicle.condition },
            { label: t("vehicle.field.conditionDetails"), value: vehicle.conditionDetails },
            { label: t("vehicle.field.packaging"), value: vehicle.packaging },
            { label: t("vehicle.field.additionalInfo"), value: vehicle.additionalInfo }
          ]}
        />
        <VehicleViewSection
          title={t("vehicles.tab.control")}
          t={t}
          fields={[
            { label: t("vehicle.field.digital"), value: vehicle.digital },
            { label: t("vehicle.field.digitalDecoderNumber"), value: vehicle.digitalDecoderNumber },
            { label: t("vehicle.field.decoderType"), value: vehicle.decoderType },
            { label: t("vehicle.field.dtDecoder"), value: vehicle.dtDecoder },
            { label: t("vehicle.field.dtDecoderNumber"), value: vehicle.dtDecoderNumber },
            { label: t("vehicle.field.abcBrakes"), value: vehicle.abcBrakes },
            { label: t("vehicle.field.adapter"), value: vehicle.adapter },
            { label: t("vehicle.field.qrCodeEnabled"), value: vehicle.qrCodeEnabled }
          ]}
        />
      </div>

      {configuredFunctions.length > 0 && (
        <section className="vehicle-view-section vehicle-view-wide">
          <h3>{t("vehicles.functions.title")}</h3>
          <div className="vehicle-view-list">
            {configuredFunctions.map((item) => (
              <article key={item.functionKey}>
                <strong>{item.functionKey}</strong>
                <span>{[item.name, item.functionType, item.mode, item.notes].filter(Boolean).join(separator)}</span>
              </article>
            ))}
          </div>
        </section>
      )}

      {images.length > 0 && (
        <section className="vehicle-view-section vehicle-view-wide">
          <h3>{t("vehicles.uploads.imagesTitle")}</h3>
          <div className="vehicle-view-gallery">
            {images.map((item) => (
              <figure key={item.id}>
                <button type="button" className="vehicle-view-image-button" onClick={() => onPreviewImage(vehicleImageToPending(item))}>
                  <img src={previewImageUrl(item)} alt="" />
                </button>
                <figcaption>{[item.isPrimary ? t("vehicles.uploads.primary") : t("vehicles.uploads.alternative"), item.title].filter(Boolean).join(separator) || item.fileName || t("vehicles.image")}</figcaption>
              </figure>
            ))}
          </div>
        </section>
      )}

      {(attachments.length > 0 || activeMaintenance.length > 0 || cvValues.length > 0 || cvFiles.length > 0) && (
        <div className="vehicle-view-grid">
          {activeMaintenance.length > 0 && (
            <section className="vehicle-view-section">
              <h3>{t("vehicles.tab.maintenance")}</h3>
              <div className="vehicle-view-list">
                {activeMaintenance.map((item) => (
                  <article key={item.id}>
                    <strong>{item.kind}</strong>
                    <span>{[item.status, item.dueDate && t("vehicles.maintenance.dueWithDate", { date: formatDate(item.dueDate) }), item.completedAt && t("vehicles.maintenance.completedWithDate", { date: formatDate(item.completedAt) }), item.notes].filter(Boolean).join(separator)}</span>
                  </article>
                ))}
              </div>
            </section>
          )}
          {attachments.length > 0 && (
            <section className="vehicle-view-section">
              <h3>{t("vehicles.uploads.attachmentsTitle")}</h3>
              <div className="vehicle-view-list">
                {attachments.map((item) => (
                  <article key={item.id}>
                    <strong>{item.originalName || item.fileName}</strong>
                    <span>{[item.category, item.description, formatFileSize(item.sizeBytes)].filter(Boolean).join(separator)}</span>
                  </article>
                ))}
              </div>
            </section>
          )}
          {cvValues.length > 0 && (
            <section className="vehicle-view-section">
              <h3>{t("vehicles.cv.values")}</h3>
              <div className="vehicle-view-list compact">
                {cvValues.slice(0, 12).map((item) => (
                  <article key={item.id}>
                    <strong>CV {item.cvNumber}</strong>
                    <span>{[String(item.value), item.category, item.protocol, item.decoderProfile, item.description].filter(Boolean).join(separator)}</span>
                  </article>
                ))}
              </div>
            </section>
          )}
          {cvFiles.length > 0 && (
            <section className="vehicle-view-section">
              <h3>{t("vehicles.cv.filesTitle")}</h3>
              <div className="vehicle-view-list">
                {cvFiles.map((item) => (
                  <article key={item.id}>
                    <strong>{item.originalName || item.fileName}</strong>
                    <span>{[item.decoderProfile, item.description, formatFileSize(item.sizeBytes)].filter(Boolean).join(separator)}</span>
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
