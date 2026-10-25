export function nearestPixels(seeds: Uint8Array, width: number, height: number, radius: number) {
  const nearest = new Int32Array(seeds.length).fill(-1);
  const distance = new Uint8Array(seeds.length).fill(255);
  const queue = new Uint32Array(seeds.length);
  let head = 0, tail = 0;
  for (let pixel = 0; pixel < seeds.length; pixel++) {
    if (!seeds[pixel]) continue;
    nearest[pixel] = pixel; distance[pixel] = 0; queue[tail++] = pixel;
  }
  while (head < tail) {
    const pixel = queue[head++], x = pixel % width;
    if (distance[pixel] === radius) continue;
    const visit = (next: number) => {
      if (nearest[next] !== -1) return;
      nearest[next] = nearest[pixel]; distance[next] = distance[pixel] + 1; queue[tail++] = next;
    };
    if (x > 0) visit(pixel - 1);
    if (x + 1 < width) visit(pixel + 1);
    if (pixel >= width) visit(pixel - width);
    if (pixel + width < width * height) visit(pixel + width);
  }
  return { nearest, distance };
}

function contrast(data: Uint8ClampedArray, pixel: number, other: number) {
  return Math.max(Math.abs(data[pixel * 4] - data[other * 4]),
    Math.abs(data[pixel * 4 + 1] - data[other * 4 + 1]),
    Math.abs(data[pixel * 4 + 2] - data[other * 4 + 2]));
}

function flatPaint(data: Uint8ClampedArray, pixel: number, width: number, height: number) {
  const x = pixel % width, y = Math.floor(pixel / width);
  if (x < 4 || y < 4 || x + 4 >= width || y + 4 >= height) return false;
  for (let row = y - 4; row <= y + 4; row++) for (let column = x - 4; column <= x + 4; column++) {
    for (let channel = 0; channel < 3; channel++) {
      if (Math.abs(data[(row * width + column) * 4 + channel] - data[pixel * 4 + channel]) > 3) return false;
    }
  }
  return true;
}

// Sum fourth-power colour steps along a path. This keeps a sharp object edge more
// expensive than a gradual shadow, without allowing distant weak edges to leak
// foreground ownership across the floor. Opposing seeds are fixed barriers.
function boundaryCosts(
  data: Uint8ClampedArray, seeds: Uint8Array, blocked: Uint8Array, width: number, height: number
) {
  // The processing-size limit keeps these integer path costs below 2^53.
  const costs = new Float64Array(seeds.length).fill(Infinity);
  const heap = new Int32Array(seeds.length), positions = new Int32Array(seeds.length).fill(-1);
  let size = 0;
  function insert(pixel: number, cost: number) {
    costs[pixel] = cost;
    let position = positions[pixel];
    if (position < 0) position = size++;
    while (position > 0) {
      const parent = (position - 1) >> 1, other = heap[parent];
      if (costs[other] <= cost) break;
      heap[position] = other; positions[other] = position; position = parent;
    }
    heap[position] = pixel; positions[pixel] = position;
  }
  function pop() {
    const pixel = heap[0], replacement = heap[--size];
    positions[pixel] = -1;
    if (size === 0) return pixel;
    let position = 0;
    while (position * 2 + 1 < size) {
      let child = position * 2 + 1;
      if (child + 1 < size && costs[heap[child + 1]] < costs[heap[child]]) child++;
      if (costs[replacement] <= costs[heap[child]]) break;
      heap[position] = heap[child]; positions[heap[position]] = position; position = child;
    }
    heap[position] = replacement; positions[replacement] = position;
    return pixel;
  }
  for (let pixel = 0; pixel < seeds.length; pixel++) if (seeds[pixel]) costs[pixel] = 0;
  for (let pixel = 0; pixel < seeds.length; pixel++) {
    if (!seeds[pixel]) continue;
    const x = pixel % width;
    if (x > 0 && !seeds[pixel - 1] || x + 1 < width && !seeds[pixel + 1] ||
      pixel >= width && !seeds[pixel - width] || pixel + width < seeds.length && !seeds[pixel + width]) insert(pixel, 0);
  }
  while (size > 0) {
    const pixel = pop(), x = pixel % width;
    const visit = (other: number) => {
      if (blocked[other]) return;
      const edge = contrast(data, pixel, other), squared = edge * edge;
      const candidate = costs[pixel] + 1 + squared * squared;
      if (candidate < costs[other]) insert(other, candidate);
    };
    if (x > 0) visit(pixel - 1);
    if (x + 1 < width) visit(pixel + 1);
    if (pixel >= width) visit(pixel - width);
    if (pixel + width < width * height) visit(pixel + width);
  }
  return costs;
}

export function segmentLocomotiveForeground(
  data: Uint8ClampedArray, width: number, height: number, background: Uint8Array, colour: number[]
) {
  if (!Number.isInteger(width) || !Number.isInteger(height) || width < 1 || height < 1 ||
    width * height > 1_440_000 || data.length !== width * height * 4 || background.length !== width * height ||
    colour.length !== 3 || !colour.every(value => Number.isFinite(value) && value >= 0 && value <= 255)) {
    throw new Error("invalid_bitmap");
  }
  let left = width, right = -1, top = height, bottom = -1;
  for (let pixel = 0; pixel < background.length; pixel++) {
    if (background[pixel]) continue;
    left = Math.min(left, pixel % width); right = Math.max(right, pixel % width);
    top = Math.min(top, Math.floor(pixel / width)); bottom = Math.max(bottom, Math.floor(pixel / width));
  }
  const inset = Math.max(2, Math.min(32, Math.round(Math.min(right - left + 1, bottom - top + 1) / 10)));
  const depth = nearestPixels(background, width, height, inset).distance;
  const seeds = new Uint8Array(background.length);
  let seedCount = 0;
  for (let pixel = 0; pixel < background.length; pixel++) {
    if (background[pixel]) continue;
    const x = pixel % width;
    const difference = Math.max(Math.abs(data[pixel * 4] - colour[0]),
      Math.abs(data[pixel * 4 + 1] - colour[1]), Math.abs(data[pixel * 4 + 2] - colour[2]));
    const chroma = Math.max(data[pixel * 4], data[pixel * 4 + 1], data[pixel * 4 + 2]) -
      Math.min(data[pixel * 4], data[pixel * 4 + 1], data[pixel * 4 + 2]);
    if (difference <= 24) continue;
    const supportedEdge = (other: number) => {
      if (contrast(data, pixel, other) < 32) return false;
      const otherDifference = Math.max(Math.abs(data[other * 4] - colour[0]),
        Math.abs(data[other * 4 + 1] - colour[1]), Math.abs(data[other * 4 + 2] - colour[2]));
      const otherChroma = Math.max(data[other * 4], data[other * 4 + 1], data[other * 4 + 2]) -
        Math.min(data[other * 4], data[other * 4 + 1], data[other * 4 + 2]);
      // Seed the object side of an outline, not the bright floor next to it.
      // Light paint adjoining a coloured body still needs its own anchor.
      return chroma >= 6 && difference >= otherDifference + 16 || depth[pixel] >= 2 && otherChroma >= 24;
    };
    const sharp = x > 0 && supportedEdge(pixel - 1) || x + 1 < width && supportedEdge(pixel + 1) ||
      pixel >= width && supportedEdge(pixel - width) ||
      pixel + width < background.length && supportedEdge(pixel + width);
    const y = Math.floor(pixel / width);
    const thin = difference >= 64 && (x >= 2 && x + 2 < width && background[pixel - 2] && background[pixel + 2] ||
      y >= 2 && y + 2 < height && background[pixel - 2 * width] && background[pixel + 2 * width]);
    const core = depth[pixel] > inset && (chroma >= 6 || difference >= 200 || flatPaint(data, pixel, width, height));
    if (core || sharp || thin) { seeds[pixel] = 1; seedCount++; }
  }
  if (!seedCount) return background.map(value => value ? 0 : 1);
  const fromBackground = boundaryCosts(data, background, seeds, width, height);
  const fromForeground = boundaryCosts(data, seeds, background, width, height);
  return background.map((value, pixel) => Number(!value && fromForeground[pixel] <= fromBackground[pixel]));
}
