// Exact PIKO56500, detailed manual28620 pp.2–11. No rules for other XP hardware.
const effectCodes = new Set([0, 1, 2, 3, 4, 10, 11, 12, 13, 14, 15, 16, 17, 18,
  30, 31, 32, 33, 34, 35, 36, 37, 40, 41, 50, 51, 52, 53, 255]);

export function piko56500Analysis(read: (cv: number) => number | undefined,
  indexed: (cv: number, cv31: number, cv32: number) => number | undefined) {
  const flag = (value: number | undefined, mask: number) => value == null ? undefined : Boolean(value & mask);
  const any = (values: (boolean | undefined)[]) => values.includes(true) ? true : values.includes(undefined) ? undefined : false;
  const config = read(29), protocols = read(12), brake = read(27), triggers = read(145), test = read(142);
  const gate = (mask: number, permission: boolean | undefined) => triggers == null ? undefined
    : !(triggers & mask) ? false : permission;
  const constantBrake = any([
    gate(1, flag(brake, 3)), gate(2, flag(brake, 48)), gate(12, true),
    triggers == null || triggers & 48 ? undefined : false,
    test == null ? undefined : test > 0
  ]);
  const mappings = read(96) !== 1 ? [] : Array.from({ length: 14 }, (_, i) => i + 33).flatMap(cv => {
    const value = read(cv);
    if (value == null || value === 0) return [];
    // Table p.4: CV33–37 start at front light,38–42 at A2,43–46 at A5.
    const start = cv <= 37 ? 0 : cv <= 42 ? 3 : 6;
    return [{ cv, key: cv <= 34 ? 0 : cv - 34,
      direction: cv === 33 ? "forward" as const : cv === 34 ? "reverse" as const : "both" as const,
      outputs: Array.from({ length: 8 }, (_, bit) => bit).filter(bit => value & (1 << bit))
        .map(bit => ({ name: start + bit, available: start + bit < 9 })) }];
  });
  const outputSets = Array.from({ length: 18 }, (_, i) => {
    const output = Math.floor(i / 2), set = i % 2 === 0 ? "A" : "B", start = 257 + output * 10 + (i % 2) * 5;
    const raw = Array.from({ length: 5 }, (_, offset) => indexed(start + offset, 18, 0));
    const byte = (value: number | undefined, max = 255) => value != null && Number.isInteger(value) && value >= 0 && value <= max ? value : undefined;
    const effect = byte(raw[0]), known = effect != null && effectCodes.has(effect);
    const flags = effect === 255 ? undefined : byte(raw[2]), smoke = effect === 40 || effect === 41;
    const autoOff = !known || smoke || flags == null ? undefined : Boolean((flags & 3) && (flags & 28));
    const timerCV = 180 + output, timer = read(timerCV);
    return { output, set, start, raw, effect: known ? effect : undefined, pwm: byte(raw[1], 64),
      flags: known && !smoke ? flags : undefined, smoke, autoOff,
      autoOffSeconds: autoOff && timer != null ? timer * 0.5 : undefined,
      fadeIn: known && !smoke ? flag(flags, 64) : undefined,
      fadeOut: known && !smoke ? flag(flags, 128) : undefined,
      timerCV, sourceCV: start + Math.max(0, raw.findIndex(v => v != null)) };
  }).filter(output => output.raw.some(value => value != null));
  return {
    mappings, outputSets, constantBrake,
    mappingMode: read(96), triggers, testKey: test == null || test === 0 ? undefined : `F${test}`,
    brakeDistanceCm: read(141), abcThresholdV: read(48) == null ? undefined : 1.6 + read(48)! * 0.1,
    acAnalog: config == null ? undefined : !(config & 4) ? false : flag(protocols, 16),
    railcomEnabled: flag(config, 8),
    railcomPlus: config == null ? undefined : !(config & 8) ? false : flag(read(28), 128),
    protocolNames: protocols == null ? undefined : [[0, "DC"], [2, "DCC"], [4, "AC"], [5, "Motorola"], [6, "mfx"]]
      .filter(([bit]) => protocols & (1 << Number(bit))).map(([, name]) => String(name)),
    lockMatches: read(15) == null || read(16) == null ? undefined : read(15) === read(16)
  };
}
