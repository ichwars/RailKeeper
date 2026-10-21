// Roco 8010884920 VI/2012 pp.2–7. Exact 10882 with technical CV8=145, not a ZIMO MN rule.
export function roco10882Analysis(read: (cv: number) => number | undefined) {
  const config = read(29), railcom = read(28), brake = read(124), brakeType = read(112);
  const flag = (value: number | undefined, mask: number) => value == null ? undefined : Boolean(value & mask);
  const analog = flag(config, 4), permitBrake = flag(brake, 32);
  const brakeEnabled = analog === true || permitBrake === false ? false
    : analog == null || permitBrake == null ? undefined : true;
  const outputs = ["frontLight", "rearLight", ...Array.from({ length: 7 }, (_, i) => `AUX${i + 1}`)];
  const mappings = Array.from({ length: 14 }, (_, i) => i + 33).flatMap(cv => {
    const value = read(cv);
    if (value == null || value === 0) return [];
    const start = cv <= 37 ? 0 : cv <= 40 ? 3 : 6;
    const count = Math.min(8, 9 - start), knownMask = (1 << count) - 1;
    return [{ cv, key: cv <= 34 ? 0 : cv - 34,
      direction: cv === 33 ? "forward" as const : cv === 34 ? "reverse" as const : "both" as const,
      outputs: Array.from({ length: count }, (_, bit) => bit).filter(bit => value & (1 << bit))
        .map(bit => ({ name: outputs[start + bit], available: start + bit < 4 })),
      unknownBits: value & ~knownMask }];
  });
  const key = (cv: number) => {
    const value = read(cv);
    // This edition does not establish whether zero means F0 or no assignment.
    return value == null || value === 0 ? undefined : `F${value}`;
  };
  const dimming = read(60), exceptions = read(114);
  return {
    mappings, analog, brakeEnabled,
    brakeType: brakeType == null ? undefined : brakeType & 64 ? "ac" as const : "dc" as const,
    railcomEnabled: config == null ? undefined : !(config & 8) ? false
      : railcom == null ? undefined : Boolean(railcom & 3),
    broadcast: flag(railcom, 1), cvData: flag(railcom, 2),
    dimmingEnabled: dimming == null ? undefined : dimming !== 0,
    dimmingExempt: exceptions == null ? undefined : outputs.slice(0, 4).filter((_, bit) => exceptions & (1 << bit)),
    unsupportedDimmingBits: exceptions == null ? undefined : exceptions & 240,
    shuntingKey: key(155), noDelayKey: key(156)
  };
}
