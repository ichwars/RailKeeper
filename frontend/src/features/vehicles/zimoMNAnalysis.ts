// ZIMO MS/MN manual SW5.29, pp.31–33, 37–38. Called only by the reviewed model rule.
export function zimoMNAnalysis(read: (cv: number) => number | undefined) {
  const get = (cv: number) => {
    const value = read(cv);
    return value != null && Number.isInteger(value) && value >= 0 && value <= 255 ? value : undefined;
  };
  const cv19 = get(19), cv20 = get(20), mask1 = get(21), mask2 = get(22);
  const extensionValid = cv20 != null && (cv20 <= 102 || (cv20 >= 128 && cv20 <= 130));
  const combined = cv19 != null && cv19 !== 128 && cv20 != null && extensionValid
    ? ((cv20 & 127) * 100) + (cv19 & 127) : undefined;
  const consistAddress = combined != null && combined <= 10239 ? combined : undefined;
  const enabledFunctions: string[] = [];
  if (mask2 != null) {
    if (mask2 & 1) enabledFunctions.push("F0 →");
    if (mask2 & 2) enabledFunctions.push("F0 ←");
  }
  if (mask1 != null) for (let bit = 0; bit < 8; bit++) if (mask1 & (1 << bit)) enabledFunctions.push(`F${bit + 1}`);
  if (mask2 != null) {
    // Bit2 has contradictory manufacturer text. Never silently name it F9.
    for (let bit = 3; bit < 6; bit++) if (mask2 & (1 << bit)) enabledFunctions.push(`F${bit + 7}`);
    if (mask2 & 128) for (let key = 13; key <= 27; key++) enabledFunctions.push(`F${key}`);
  }
  const mode = get(140), distance = get(141);
  const brakeMode = mode != null && [0, 1, 2, 3, 11, 12, 13].includes(mode) ? mode : undefined;
  const pair = (highCV: number, lowCV: number) => {
    const high = get(highCV), low = get(lowCV);
    return high != null && low != null ? high * 256 + low : undefined;
  };
  const config = get(29), analog = get(12);
  return {
    consistAddress,
    consistReverse: consistAddress != null && cv19 != null ? Boolean(cv19 & 128) : undefined,
    autoConsist: mask2 != null ? Boolean(mask2 & 64) : undefined,
    enabledFunctions,
    functionsComplete: mask1 != null && mask2 != null && !(mask2 & 4),
    ambiguousFunctionBit: mask2 != null && Boolean(mask2 & 4),
    dcAnalog: config == null || config > 63 ? undefined : !(config & 4) ? false
      : analog == null || analog === 0 ? undefined : Boolean(analog & 1),
    brakeMode,
    constantBrake: brakeMode == null ? undefined : brakeMode !== 0,
    brakeDistance: {
      fine: distance === 0,
      forward: distance === 0 ? pair(830, 831) : distance,
      reverse: distance === 0 ? pair(832, 833) : distance,
      forwardCVs: distance === 0 ? [141, 830, 831] : [141],
      reverseCVs: distance === 0 ? [141, 832, 833] : [141]
    }
  };
}

export function zimoSpeedParameters(read: (cv: number) => number | undefined) {
  const maximum = read(5), middle = read(6);
  return {
    maximum: maximum != null && Number.isInteger(maximum) && maximum >= 0 && maximum <= 255
      ? maximum <= 1 ? 255 : maximum : null,
    middle: middle != null && Number.isInteger(middle) && middle >= 2 && middle <= 255 ? middle : null,
    defaultMiddle: middle === 1
  };
}
