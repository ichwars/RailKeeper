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

export function composeBrandedQrSvg(svg: string) {
  const mark = `<rect x="111" y="111" width="34" height="34" rx="8" fill="#fff"/><image href="/brand/railkeeper-mark.png" x="115" y="115" width="26" height="26" preserveAspectRatio="xMidYMid meet"/>`;
  return svg.replace("</svg>", `${mark}</svg>`);
}

export async function buildQrSvg(vehicle: Vehicle | null, formData: CreateVehicleRequest) {
  const QRCode = (await import("qrcode")).default;
  const svg = await QRCode.toString(qrPayload(vehicle, formData), {
    type: "svg",
    width: 256,
    margin: 2,
    color: { dark: "#0b1e26", light: "#ffffff" }
  });
  return composeBrandedQrSvg(svg);
}

export async function buildBrandedQrPngDataUrl(payload: string, width = 768) {
  const QRCode = (await import("qrcode")).default;
  const dataURL = await QRCode.toDataURL(payload, {
    width,
    margin: 2,
    color: { dark: "#0b1e26", light: "#ffffff" }
  });
  const canvas = document.createElement("canvas");
  canvas.width = width;
  canvas.height = width;
  const context = canvas.getContext("2d");
  if (!context) return dataURL;
  const qrImage = new window.Image();
  await new Promise<void>((resolve, reject) => {
    qrImage.onload = () => resolve();
    qrImage.onerror = () => reject(new Error("QR-Code konnte nicht geladen werden."));
    qrImage.src = dataURL;
  });
  context.drawImage(qrImage, 0, 0, width, width);
  const logoImage = new window.Image();
  await new Promise<void>((resolve) => {
    logoImage.onload = () => resolve();
    logoImage.onerror = () => resolve();
    logoImage.src = "/brand/railkeeper-mark.png";
  });
  const plateSize = Math.round(width * 0.14);
  const plateX = Math.round((width - plateSize) / 2);
  const plateRadius = Math.round(plateSize * 0.18);
  context.fillStyle = "#fff";
  context.roundRect(plateX, plateX, plateSize, plateSize, plateRadius);
  context.fill();
  if (logoImage.complete && logoImage.naturalWidth > 0) {
    const logoPadding = Math.round(plateSize * 0.12);
    context.drawImage(logoImage, plateX + logoPadding, plateX + logoPadding, plateSize - logoPadding * 2, plateSize - logoPadding * 2);
  }
  return canvas.toDataURL("image/png");
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
