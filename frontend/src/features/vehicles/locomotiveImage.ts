export type ImageCrop = { x: number; y: number; width: number; height: number };
export type ImageSize = { width: number; height: number };
export type ImageBackground = "transparent" | "white" | "grey" | "black";

export const locomotiveImageTargets = [
  { id: "ecos", name: "ESU ECoS 50200 / 50210 / 50220", width: 190, height: 40,
    format: "bmp", background: "grey", fixed: true,
    source: "https://www.esu.eu/uploads/tx_esudownloads/50200_ECoS_ESU_KG_Handbuch-Addendum_I_LokBilder_August_2010_eBook.pdf" },
  { id: "cs3", name: "Märklin CS3 / CS3 plus", width: 384, height: 132,
    format: "png", background: "transparent", fixed: false,
    source: "https://static.maerklin.de/damcontent/c1/ce/c1ce554ad6ad195c04e32ba1b4dd64511764068661.pdf" },
  { id: "intellibox3", name: "Uhlenbrock Intellibox 3 · IB3 Konfiguration", width: 340, height: 72,
    format: "png", background: "transparent", fixed: false,
    source: "https://www.uhlenbrock.de/de_DE/service/download/software/I4339F6A-001.htm!ArcEntryInfo=0004.24.I4339F6A" },
  { id: "z21", name: "Z21-App", width: 960, height: 400,
    format: "png", background: "white", fixed: false,
    source: "https://www.z21.eu/de/faq-support/fragen-antworten" }
] as const;

export type LocomotiveImageTarget = typeof locomotiveImageTargets[number];
export const imageBackgroundColors: Record<Exclude<ImageBackground, "transparent">, string> = {
  white: "#ffffff", grey: "#b6b6b6", black: "#000000"
};

export function validImageSize(size: ImageSize) {
  return Number.isInteger(size.width) && Number.isInteger(size.height) &&
    size.width > 0 && size.height > 0 && size.width <= 12000 && size.height <= 12000 &&
    size.width * size.height <= 40_000_000;
}

export function validImageCrop(crop: ImageCrop, size: ImageSize) {
  return [crop.x, crop.y, crop.width, crop.height].every(Number.isFinite) &&
    crop.x >= 0 && crop.y >= 0 && crop.width >= 1 && crop.height >= 1 &&
    crop.x + crop.width <= size.width && crop.y + crop.height <= size.height;
}

export function fitLocomotiveImage(crop: ImageCrop, target: ImageSize) {
  const padding = Math.max(1, Math.round(target.height * 0.04));
  const scale = Math.min((target.width - padding * 2) / crop.width, (target.height - padding * 2) / crop.height);
  const width = crop.width * scale;
  const height = crop.height * scale;
  return { x: (target.width - width) / 2, y: (target.height - height) / 2, width, height, scale };
}

export function imageSuitability(crop: ImageCrop, size: ImageSize, target: ImageSize, trim?: ImageCrop) {
  if (!validImageSize(size) || !validImageCrop(crop, size)) return { invalid: true, upscale: false, narrow: false, border: false };
  const fit = fitLocomotiveImage(crop, target);
  return { invalid: false, upscale: fit.scale > 1,
    narrow: fit.width < target.width * 0.55 || fit.height < target.height * 0.35,
    border: Boolean(trim && trim.width * trim.height < crop.width * crop.height * 0.6) };
}

// Only transparency is used for trimming; an opaque background can contain parts of the locomotive.
export function alphaBounds(data: Uint8ClampedArray, width: number, height: number): ImageCrop | null {
  let left = width, top = height, right = -1, bottom = -1;
  for (let y = 0; y < height; y++) for (let x = 0; x < width; x++) {
    if (data[(y * width + x) * 4 + 3] === 0) continue;
    left = Math.min(left, x); top = Math.min(top, y);
    right = Math.max(right, x); bottom = Math.max(bottom, y);
  }
  return right < 0 ? null : { x: left, y: top, width: right - left + 1, height: bottom - top + 1 };
}

export function locomotiveImageFileName(name: string, target: LocomotiveImageTarget) {
  const stem = name.replace(/\.(jpe?g|png|webp|bmp)$/i, "").normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "").replace(/[^a-zA-Z0-9_-]+/g, "-").replace(/^-+|-+$/g, "").slice(0, 80);
  return `${stem || "lokbild"}_${target.id}.${target.format}`;
}

export function encodeBMP24(data: Uint8ClampedArray, width: number, height: number): ArrayBuffer {
  if (!validImageSize({ width, height }) || data.length !== width * height * 4) throw new Error("invalid_bitmap");
  const stride = Math.ceil(width * 3 / 4) * 4;
  const buffer = new ArrayBuffer(54 + stride * height);
  const header = new DataView(buffer);
  header.setUint16(0, 0x4d42, true);
  header.setUint32(2, buffer.byteLength, true);
  header.setUint32(10, 54, true);
  header.setUint32(14, 40, true);
  header.setInt32(18, width, true);
  header.setInt32(22, height, true);
  header.setUint16(26, 1, true);
  header.setUint16(28, 24, true);
  header.setUint32(34, stride * height, true);
  const pixels = new Uint8Array(buffer);
  for (let y = 0; y < height; y++) for (let x = 0; x < width; x++) {
    const source = (y * width + x) * 4;
    if (data[source + 3] !== 255) throw new Error("bitmap_requires_background");
    const dest = 54 + (height - y - 1) * stride + x * 3;
    pixels[dest] = data[source + 2]; pixels[dest + 1] = data[source + 1]; pixels[dest + 2] = data[source];
  }
  return buffer;
}
