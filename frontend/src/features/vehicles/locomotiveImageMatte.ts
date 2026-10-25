import { nearestPixels, segmentLocomotiveForeground } from "./locomotiveImageSegmentation";

function backgroundColour(data: Uint8ClampedArray, background: Uint8Array) {
  const histograms = [new Uint32Array(256), new Uint32Array(256), new Uint32Array(256)];
  let count = 0;
  for (let pixel = 0; pixel < background.length; pixel++) {
    if (!background[pixel]) continue;
    count++;
    for (let channel = 0; channel < 3; channel++) histograms[channel][data[pixel * 4 + channel]]++;
  }
  if (count === 0) throw new Error("invalid_bitmap");
  return histograms.map(histogram => {
    let total = 0;
    for (let value = 0; value < 256; value++) {
      total += histogram[value];
      if (total >= count / 2) return value;
    }
    return 255;
  });
}

function thinForeground(pixel: number, background: Uint8Array, width: number, height: number) {
  const x = pixel % width, y = Math.floor(pixel / width);
  return x >= 2 && x + 2 < width && background[pixel - 2] && background[pixel + 2] ||
    y >= 2 && y + 2 < height && background[pixel - 2 * width] && background[pixel + 2 * width];
}

function touchesForeground(pixel: number, mask: Uint8Array, width: number, height: number) {
  const x = pixel % width, y = Math.floor(pixel / width);
  for (let row = Math.max(0, y - 1); row <= Math.min(height - 1, y + 1); row++) {
    for (let column = Math.max(0, x - 1); column <= Math.min(width - 1, x + 1); column++) {
      if (mask[row * width + column]) return true;
    }
  }
  return false;
}

function paintBand(
  data: Uint8ClampedArray, pixel: number, object: Uint8Array, depth: Uint8Array, width: number, height: number
) {
  const x = pixel % width;
  const matches = (other: number) => object[other] && depth[other] !== depth[pixel] &&
    Math.max(Math.abs(data[pixel * 4] - data[other * 4]), Math.abs(data[pixel * 4 + 1] - data[other * 4 + 1]),
      Math.abs(data[pixel * 4 + 2] - data[other * 4 + 2])) <= 6;
  return x > 0 && matches(pixel - 1) || x + 1 < width && matches(pixel + 1) ||
    pixel >= width && matches(pixel - width) || pixel + width < width * height && matches(pixel + width);
}

// Preserve opaque details and coherent paint bands, including bright wheel rims.
// Background colour unmixing is restricted to the remaining contour pixels.
export function refineLocomotiveImageMatte(
  data: Uint8ClampedArray, width: number, height: number, background: Uint8Array
) {
  if (!Number.isInteger(width) || !Number.isInteger(height) || width < 1 || height < 1 ||
    data.length !== width * height * 4 || background.length !== width * height ||
    width * height > 1_440_000) throw new Error("invalid_bitmap");
  const colour = backgroundColour(data, background), count = background.length;
  const object = segmentLocomotiveForeground(data, width, height, background, colour);
  const outer = object.map(value => value ? 0 : 1);
  const border = nearestPixels(outer, width, height, 3);
  const foreground = new Uint8Array(count);
  for (let pixel = 0; pixel < count; pixel++) {
    foreground[pixel] = Number(object[pixel] &&
      (border.distance[pixel] > 2 || thinForeground(pixel, outer, width, height)));
  }
  const samples = nearestPixels(foreground, width, height, 8);
  const matte = new Uint8ClampedArray(data.length);
  for (let pixel = 0; pixel < count; pixel++) {
    const offset = pixel * 4, sample = samples.nearest[pixel];
    let alpha = object[pixel] ? 1 : 0;
    const preserved = foreground[pixel] || object[pixel] && paintBand(data, pixel, object, border.distance, width, height);
    if (!preserved && sample >= 0 && (object[pixel] || touchesForeground(pixel, object, width, height))) {
      let dot = 0, norm = 0;
      for (let channel = 0; channel < 3; channel++) {
        const direction = data[sample * 4 + channel] - colour[channel];
        dot += (data[offset + channel] - colour[channel]) * direction; norm += direction * direction;
      }
      alpha = norm > 12 ? Math.max(0, Math.min(1, dot / norm)) : alpha;
      let residual = 0;
      for (let channel = 0; channel < 3; channel++) {
        const predicted = colour[channel] + alpha * (data[sample * 4 + channel] - colour[channel]);
        residual = Math.max(residual, Math.abs(data[offset + channel] - predicted));
      }
      if (object[pixel] && residual > 12) alpha = 1;
    }
    matte.set(colour, offset); matte[offset + 3] = Math.round((1 - alpha) * 255);
  }
  return matte;
}

export function unmixLocomotiveImageBackground(data: Uint8ClampedArray, matte: Uint8ClampedArray) {
  if (data.length !== matte.length || data.length % 4 !== 0) throw new Error("invalid_bitmap");
  const output = data.slice();
  for (let offset = 0; offset < data.length; offset += 4) {
    const removed = matte[offset + 3] / 255, alpha = 1 - removed;
    for (let channel = 0; channel < 3; channel++) {
      output[offset + channel] = alpha > 0
        ? (data[offset + channel] - removed * matte[offset + channel]) / alpha : 0;
    }
    output[offset + 3] = Math.round(data[offset + 3] * alpha);
  }
  return output;
}
