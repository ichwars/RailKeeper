// Exact 76420, German instructions 11.10 ru, pp.2,5–6. No firmware/model extrapolation.
export function uhlenbrock76420Analysis(read: (cv: number) => number | undefined) {
  const config = read(29), options = read(49), analog = read(51), timed = read(98), duration = read(99);
  const flag = (value: number | undefined, mask: number) => value == null ? undefined : Boolean(value & mask);
  const permitted = flag(config, 4);
  const analogAllowed = (mode: number) => permitted === false ? false
    : permitted == null || analog == null ? undefined : analog === mode || analog === 3;
  const targetSelected = flag(options, 4);
  const mappings = Array.from({ length: 10 }, (_, i) => i + 33).flatMap(cv => {
    const value = read(cv);
    if (value == null || value === 0) return [];
    return [{ cv, key: cv <= 34 ? 0 : cv - 34,
      direction: cv === 33 ? "forward" as const : cv === 34 ? "reverse" as const : "both" as const,
      actions: ["frontLight", "rearLight", "outputA1", "outputA2", "shunting", "delay"]
        .filter((_, bit) => value & (1 << bit)) }];
  });
  const minimum = read(2), middle = read(6), maximum = read(5);
  return {
    mappings, dcAnalog: analogAllowed(2), acAnalog: analogAllowed(1),
    loadControl: options == null ? undefined : !(options & 1),
    susi: options == null ? undefined : options & 2 ? "lissy" as const : "sound" as const,
    protocols: options == null ? undefined : (options & 24) === 24 ? "blocked" as const
      : options & 8 ? "dcc" as const : options & 16 ? "motorola" as const : "both" as const,
    // A false required condition is enough to disprove activation, without substituting missing CVs.
    dcBrake: permitted === false || flag(options, 128) === false ? false
      : permitted == null || options == null ? undefined : true,
    targetSelected, brakeTarget: targetSelected === false ? 0 : targetSelected ? read(52) : undefined,
    dimmingAux: flag(options, 32), lightsSwapped: flag(options, 64),
    timedOutputs: timed == null ? undefined : ["A1", "A2"].filter((_, bit) => timed & (1 << bit)),
    durationSeconds: duration == null ? undefined : duration / 10,
    shortCircuit: read(60) == null ? undefined : read(60) === 9,
    parameterOrder: minimum == null || middle == null || maximum == null ? undefined
      : minimum < middle && middle < maximum
  };
}
