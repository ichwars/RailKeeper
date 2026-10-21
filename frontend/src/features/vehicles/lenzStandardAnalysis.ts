// Lenz STANDARD+ V2, edition5 pp.19–26; Plus manual edition4 pp.50–51.
// Called only for the reviewed model rule. No defaults replace missing readings.
export function lenzStandardAnalysis(read: (cv: number) => number | undefined) {
  const get = (cv: number) => {
    const value = read(cv);
    return value != null && Number.isInteger(value) && value >= 0 && value <= 255 ? value : undefined;
  };
  const config = get(29), brake = get(51), motor = get(50), channels = get(28);
  const flag = (value: number | undefined, mask: number) => value == null ? undefined : Boolean(value & mask);
  const abc = flag(brake, 2), analog = flag(config, 4);
  const motorType = motor == null || (motor & 15) > 5 ? undefined : motor & 15;
  const constantABC = brake == null ? undefined : Boolean(brake & 1) && Boolean(abc);
  const constantZero = flag(brake, 128);
  const masks = [33, 34, 35, ...Array.from({ length: 11 }, (_, i) => i + 36), 47,
    ...Array.from({ length: 16 }, (_, i) => i + 129)];
  const mappings = masks.flatMap(cv => {
    const value = get(cv);
    if (value == null || value === 0 || value > 31) return [];
    const key = cv === 33 || cv === 34 ? 0 : cv === 35 || cv === 47 ? 1 : cv < 129 ? cv - 34 : cv - 116;
    const direction = cv === 33 || cv === 35 ? "forward" as const : cv === 34 || cv === 47 ? "reverse" as const : "both" as const;
    return [{ cv, key, direction, outputs: [..."ABCD"].filter((_, bit) => value & (1 << bit)), unknownBit: Boolean(value & 16) }];
  });
  const keys = (cv: number) => {
    const mask = get(cv);
    return mask == null ? undefined : Array.from({ length: 8 }, (_, bit) => bit + 1)
      .filter(key => mask & (1 << (key - 1))).map(key => `F${key}`);
  };
  const effect = (cv: number, tens: boolean, maximum: number) => {
    const value = get(cv);
    if (value == null) return undefined;
    if (Math.floor(value / 10) > 4 || value % 10 > (cv === 62 ? 2 : 4)) return undefined;
    const digit = tens ? Math.floor(value / 10) : value % 10;
    return digit <= maximum ? digit : undefined;
  };
  const frequency = get(63);
  return {
    constantABC, constantZero,
    constantBrake: brake == null ? undefined : Boolean(constantABC || constantZero),
    abcDirectionIndependent: flag(brake, 4),
    shuttleMode: brake == null ? undefined : (brake & 24) === 24 ? "conflict" : !(brake & 24) ? "off"
      : !abc ? "requiresABC" : brake & 16 ? "withStop" : "withoutStop",
    dcStop: brake == null || analog == null ? undefined : !analog && Boolean(brake & 32),
    railcom: config == null ? undefined : !(config & 8) ? false : channels == null ? undefined : Boolean(channels & 3),
    motorType, functionDecoder: flag(motor, 16), emkDivider: flag(motor, 32),
    loadControl: motor == null ? undefined : !(motor & 64),
    lowFrequency: flag(motor, 128), mappings,
    dimmingKeys: keys(57), shuntingKeys: keys(58), noDelayKeys: keys(59), effectsABKeys: keys(61), effectsCDKeys: keys(64),
    effectA: effect(60, false, 4), effectB: effect(60, true, 4),
    effectC: effect(62, false, 2), effectD: effect(62, true, 4),
    flashHz: frequency == null ? undefined : 1 / (0.03 * (1 + frequency)),
    faultBits: get(30) == null ? undefined : [0, 1, 2].filter(bit => (get(30) ?? 0) & (1 << bit))
  };
}
