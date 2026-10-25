import { alphaBounds, type ImageCrop } from "./locomotiveImage";

// Only a uniform colour connected to the original image edge is background.
// Enclosed details of the same colour and ambiguous photographs stay intact.
export function locomotiveImageBackground(data: Uint8ClampedArray, width: number, height: number) {
  if (data.length !== width * height * 4 || width * height > 1_440_000) return null;
  for (let i = 3; i < data.length; i += 4) if (data[i] !== 255) return null;
  const matches = (pixel: number) => [0, 1, 2].every(channel => Math.abs(data[pixel * 4 + channel] - data[channel]) <= 16);
  if (![width - 1, (height - 1) * width, width * height - 1].every(matches)) return null;
  const edges: number[] = [];
  for (let x = 0; x < width; x++) edges.push(x, (height - 1) * width + x);
  for (let y = 1; y < height - 1; y++) edges.push(y * width, y * width + width - 1);
  if (edges.filter(matches).length < edges.length * 0.85) return null;
  const background = new Uint8Array(width * height), queue = new Uint32Array(width * height);
  let head = 0, tail = 0;
  const visit = (pixel: number) => {
    if (background[pixel] || !matches(pixel)) return;
    background[pixel] = 1; queue[tail++] = pixel;
  };
  edges.forEach(visit);
  while (head < tail) {
    const pixel = queue[head++], x = pixel % width;
    if (x > 0) visit(pixel - 1);
    if (x < width - 1) visit(pixel + 1);
    if (pixel >= width) visit(pixel - width);
    if (pixel + width < background.length) visit(pixel + width);
  }
  return tail === width * height ? null : background;
}

export function locomotiveImageBounds(
  data: Uint8ClampedArray, width: number, height: number,
  background = locomotiveImageBackground(data, width, height)
): ImageCrop | null {
  const visible = alphaBounds(data, width, height);
  if (!visible || !background) return visible;
  let left = width, top = height, right = -1, bottom = -1;
  for (let pixel = 0; pixel < background.length; pixel++) {
    if (background[pixel]) continue;
    const x = pixel % width, y = Math.floor(pixel / width);
    left = Math.min(left, x); right = Math.max(right, x); top = Math.min(top, y); bottom = Math.max(bottom, y);
  }
  return right < 0 ? visible : { x: left, y: top, width: right - left + 1, height: bottom - top + 1 };
}

const redGreen = [0, 36, 73, 109, 146, 182, 219, 255];
const blue = [0, 73, 182, 255];
const nearest = (value: number, levels: number[]) => levels.reduce((best, level) =>
  Math.abs(value - level) < Math.abs(value - best) ? level : best, levels[0]);

// Firmware 4.3.3's observed palette and integer error diffusion. Sending palette
// colours in RGB24 makes another device quantization idempotent. Preview uses
// these same pixels; the server still verifies the device readback independently.
export function quantizeECoSImage(data: Uint8ClampedArray, width: number, height: number) {
  if (!Number.isInteger(width) || !Number.isInteger(height) || width < 1 || height < 1 ||
    width > 190 || height > 40 || data.length !== width * height * 4) {
    throw new Error("invalid_bitmap");
  }
  const output = data.slice();
  let current = new Int32Array((width + 2) * 3);
  for (let y = 0; y < height; y++) {
    const next = new Int32Array(current.length);
    for (let x = 0; x < width; x++) {
      const pixel = (y * width + x) * 4, position = (x + 1) * 3;
      if (data[pixel + 3] !== 255) throw new Error("bitmap_requires_background");
      if (data[pixel] === 182 && data[pixel + 1] === 182 && data[pixel + 2] === 182) continue;
      for (let channel = 0; channel < 3; channel++) {
        const adjusted = Math.max(0, Math.min(255, data[pixel + channel] + (current[position + channel] >> 4)));
        const quantized = nearest(adjusted, channel === 2 ? blue : redGreen), error = adjusted - quantized;
        output[pixel + channel] = quantized;
        current[position + 3 + channel] += error * 7;
        next[position - 3 + channel] += error * 3;
        next[position + channel] += error * 5;
        next[position + 3 + channel] += error;
      }
    }
    current = next;
  }
  return output;
}
