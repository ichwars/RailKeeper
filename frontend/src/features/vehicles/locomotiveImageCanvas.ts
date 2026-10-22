import {
  encodeBMP24, fitLocomotiveImage, imageBackgroundColors, validImageSize,
  type ImageBackground, type ImageCrop, type LocomotiveImageTarget
} from "./locomotiveImage";
import { locomotiveImageBounds, quantizeECoSImage } from "./locomotiveImagePixels";

const maxSourceBytes = 20 * 1024 * 1024;

export async function loadLocomotiveImage(url: string, signal: AbortSignal) {
  const source = new URL(url, window.location.href);
  if (source.origin !== window.location.origin || !["http:", "https:", "blob:"].includes(source.protocol)) {
    throw new Error("localImage");
  }
  if (typeof createImageBitmap !== "function") throw new Error("unsupported");
  const response = await fetch(source.href, { signal, credentials: "same-origin" });
  if (!response.ok || !response.body) throw new Error("loadFailed");
  const reader = response.body.getReader();
  const chunks: ArrayBuffer[] = [];
  let bytes = 0;
  try {
    for (;;) {
      const { value, done } = await reader.read();
      if (done) break;
      bytes += value.byteLength;
      if (bytes > maxSourceBytes) throw new Error("tooLarge");
      chunks.push(value.slice().buffer);
    }
  } finally { await reader.cancel(); }
  const blob = new Blob(chunks);
  const magic = new Uint8Array(await blob.slice(0, 32).arrayBuffer());
  const png = magic[0] === 137 && magic[1] === 80 && magic[2] === 78 && magic[3] === 71;
  const jpeg = magic[0] === 255 && magic[1] === 216 && magic[2] === 255;
  const webp = String.fromCharCode(...magic.slice(0, 4)) === "RIFF" &&
    String.fromCharCode(...magic.slice(8, 12)) === "WEBP";
  if (!png && !jpeg && !webp) throw new Error("typeBlocked");
  if (png && magic.length >= 24) {
    const header = new DataView(magic.buffer);
    if (!validImageSize({ width: header.getUint32(16), height: header.getUint32(20) })) throw new Error("tooLarge");
  }
  const bitmap = await createImageBitmap(blob);
  if (signal.aborted || !validImageSize(bitmap)) { bitmap.close(); throw new Error("tooLarge"); }
  return bitmap;
}

export function suggestedImageBounds(bitmap: ImageBitmap) {
  const canvas = document.createElement("canvas");
  // Bound analysis memory and leave a margin around the proposed, user-reviewed trim.
  const scale = Math.min(1, 1200 / bitmap.width, 1200 / bitmap.height);
  canvas.width = Math.max(1, Math.ceil(bitmap.width * scale));
  canvas.height = Math.max(1, Math.ceil(bitmap.height * scale));
  const context = canvas.getContext("2d", { willReadFrequently: true });
  if (!context) throw new Error("unsupported");
  context.drawImage(bitmap, 0, 0, canvas.width, canvas.height);
  const bounds = locomotiveImageBounds(context.getImageData(0, 0, canvas.width, canvas.height).data, canvas.width, canvas.height);
  if (!bounds) return null;
  const x = Math.max(0, Math.floor(bounds.x / scale) - 1);
  const y = Math.max(0, Math.floor(bounds.y / scale) - 1);
  return { x, y, width: Math.min(bitmap.width - x, Math.ceil(bounds.width / scale) + 2),
    height: Math.min(bitmap.height - y, Math.ceil(bounds.height / scale) + 2) };
}

export function renderLocomotiveImage(
  canvas: HTMLCanvasElement, bitmap: ImageBitmap, crop: ImageCrop,
  target: LocomotiveImageTarget, background: ImageBackground
) {
  canvas.width = target.width; canvas.height = target.height;
  const context = canvas.getContext("2d");
  if (!context) throw new Error("unsupported");
  if (background !== "transparent" || target.format === "bmp") {
    context.fillStyle = imageBackgroundColors[background === "transparent" ? "grey" : background];
    context.fillRect(0, 0, canvas.width, canvas.height);
  }
  context.imageSmoothingEnabled = true; context.imageSmoothingQuality = "high";
  const fit = fitLocomotiveImage(crop, target);
  context.drawImage(bitmap, crop.x, crop.y, crop.width, crop.height, fit.x, fit.y, fit.width, fit.height);
  if (target.format === "bmp") {
    const pixels = context.getImageData(0, 0, canvas.width, canvas.height);
    pixels.data.set(quantizeECoSImage(pixels.data, canvas.width, canvas.height));
    context.putImageData(pixels, 0, 0);
    return Promise.resolve(new Blob([encodeBMP24(pixels.data, canvas.width, canvas.height)], { type: "image/bmp" }));
  }
  return new Promise<Blob>((resolve, reject) => canvas.toBlob(blob => {
    if (blob && blob.size < 5 * 1024 * 1024) resolve(blob);
    else reject(new Error("exportFailed"));
  }, "image/png"));
}
