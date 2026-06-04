import type { CreateVehicleRequest, Vehicle } from "../../shared/api";

export function qrPayload(vehicle: Vehicle | null, form: CreateVehicleRequest) {
  const inventory = form.inventoryNumber || vehicle?.inventoryNumber || "";
  const name = form.name || vehicle?.name || "";
  const decoder = form.digitalDecoderNumber || form.dtDecoderNumber || vehicle?.digitalDecoderNumber || vehicle?.dtDecoderNumber || "";
  const lines = [
    `Inventar-Nr.: ${inventory || "-"}`,
    `Bezeichnung: ${name || "-"}`
  ];
  if (decoder) {
    lines.push(`Decoder-Nr.: ${decoder}`);
  }
  return lines.join("\n");
}

export async function buildQrSvg(vehicle: Vehicle | null, formData: CreateVehicleRequest) {
  const QRCode = (await import("qrcode")).default;
  return QRCode.toString(qrPayload(vehicle, formData), {
    type: "svg",
    width: 256,
    margin: 2,
    color: { dark: "#0b1e26", light: "#ffffff" }
  });
}

export async function buildBrandedQrPngDataUrl(payload: string, width = 768) {
  const QRCode = (await import("qrcode")).default;
  return QRCode.toDataURL(payload, {
    width,
    margin: 2,
    color: { dark: "#0b1e26", light: "#ffffff" }
  });
}

export function downloadQrSvgFile(qrSvg: string, fileBase: string) {
  if (!qrSvg) return;
  const blob = new Blob([qrSvg], { type: "image/svg+xml" });
  const link = document.createElement("a");
  link.href = URL.createObjectURL(blob);
  link.download = `${fileBase || "railkeeper"}-qr.svg`;
  link.click();
  URL.revokeObjectURL(link.href);
}

export async function downloadQrPngFile(payload: string, fileBase: string) {
  const link = document.createElement("a");
  link.href = await buildBrandedQrPngDataUrl(payload);
  link.download = `${fileBase || "railkeeper"}-qr.png`;
  link.click();
}

export function printQrSvgLabel(qrSvg: string, form: CreateVehicleRequest) {
  if (!qrSvg) return;
  const printWindow = window.open("", "railkeeper-qr-print", "width=520,height=680");
  if (!printWindow) throw new Error("Druckfenster konnte nicht geoeffnet werden.");
  printWindow.document.write([
    "<!doctype html>",
    "<html><head><title>RailKeeper QR-Code</title>",
    "<style>",
    "body { font-family: system-ui, sans-serif; margin: 24px; color: #0b1e26; }",
    ".label { width: 62mm; min-height: 38mm; border: 1px solid #d7e1dc; padding: 5mm; display: grid; grid-template-columns: 26mm 1fr; gap: 4mm; align-items: center; }",
    "svg { width: 26mm; height: 26mm; }",
    "strong { display: block; font-size: 12pt; }",
    "span { display: block; font-size: 9pt; margin-top: 2mm; }",
    "@media print { body { margin: 0; } .label { border: 0; } }",
    "</style></head><body><div class=\"label\">",
    qrSvg,
    "<div><strong>", form.inventoryNumber || "", "</strong><span>", form.name || "", "</span><span>", form.digitalDecoderNumber || form.dtDecoderNumber || "", "</span></div>",
    "</div></body></html>"
  ].join(""));
  printWindow.document.close();
  printWindow.focus();
  printWindow.print();
}
