// N45 revision b, pp.13–20. Applies only to the reviewed N45 rule, not other Kühn models.
export function kuehnN45Analysis(read: (cv: number) => number | undefined) {
  const get = (cv: number) => {
    const value = read(cv);
    return value != null && Number.isInteger(value) && value >= 0 && value <= 255 ? value : undefined;
  };
  const flag = (value: number | undefined, mask: number) => value == null ? undefined : Boolean(value & mask);
  const config = get(29), brake = get(27), motor = get(56), memory = get(112);
  const susi = flag(motor, 4);
  const keys = (cv: number, offset: number, count: number, first: number) => {
    const value = get(cv);
    return value == null ? undefined : Array.from({ length: count }, (_, i) => i)
      .filter(i => value & (1 << (offset + i))).map(i => `F${first + i}`);
  };
  const decimal = (cv: number) => {
    const value = get(cv);
    return value == null || value > 99 || (cv === 63 && value % 10 > 8)
      ? undefined : { tens: Math.floor(value / 10), units: value % 10 };
  };
  const outputEffect = (cv: number) => {
    const value = get(cv);
    if (value == null) return undefined;
    const code = value & 31;
    return { code: code <= 12 && ![3, 10, 11].includes(code) && !(value & 128) ? code : undefined,
      offReverse: Boolean(value & 32), offForward: Boolean(value & 64) };
  };
  const mappings = Array.from({ length: 14 }, (_, i) => i + 33).flatMap(cv => {
    const value = get(cv);
    if (value == null || value === 0) return [];
    const positions = cv <= 37 ? [0, 1, 2, 3, 4, 5] : cv <= 42 ? [5, 6, 7, 0, 1, 2] : [2, 3, 4, 5, 6, 7];
    const knownMask = positions.reduce((mask, bit) => mask | (1 << bit), 0);
    return [{ cv, key: cv <= 34 ? 0 : cv - 34,
      direction: cv === 33 ? "forward" as const : cv === 34 ? "reverse" as const : "both" as const,
      unknownBits: value & ~knownMask,
      outputs: [..."ABCDEF"].flatMap((output, i) => value & (1 << positions[i]) ? [{ output,
        available: i < 4 ? true : susi == null ? undefined : !susi,
        effect: outputEffect([49, 50, 51, 52, 47, 48][i]) }] : []) }];
  });
  const coupling = decimal(63), random = get(113), trim = get(95);
  const frequency = get(9);
  return {
    susi, mappings, effects: [49, 50, 51, 52, 47, 48].map(outputEffect),
    frequency: frequency == null ? undefined : [15600, 120, 32000][frequency],
    loadControl: flag(motor, 2), emkDivider: flag(motor, 1),
    memory: memory == null ? undefined : !(memory & 32),
    dcAnalog: config == null ? undefined : !(config & 4) ? false
      : brake == null ? undefined : (brake & 48) !== 48,
    dcOppositeRetained: flag(brake, 16), dcForward: flag(brake, 32),
    analogKeys: get(13) == null || get(14) == null ? undefined
      : ["F0", ...(keys(13, 0, 8, 1) ?? []), ...(keys(14, 2, 4, 9) ?? [])],
    shuntingKeys: keys(58, 4, 4, 1), noDelayKeys: keys(58, 0, 4, 1), dimmingKeys: keys(59, 0, 4, 1),
    shuntingLight: get(57) == null ? undefined : [..."ABCDEF"].filter((_, i) => (get(57) ?? 0) & (1 << i)),
    shuntingNeedsF0: flag(get(57), 64),
    effectCycle: decimal(55)?.tens, effectDimming: decimal(55)?.units,
    dimming: [60, 61, 62].map(decimal),
    pullInSeconds: coupling == null ? undefined : coupling.tens === 0 ? 0.5 : coupling.tens,
    holdingLevel: coupling?.units,
    randomSeconds: random == null ? undefined : [0.032, 0.064, 0.128, 0.256, 0.512, 1, 2, 4, 8, 16][random],
    reverseFactor: trim == null ? undefined : trim === 0 ? 1 : trim / 128
  };
}
