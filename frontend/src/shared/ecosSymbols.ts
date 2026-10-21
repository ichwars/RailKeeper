import type { MasterDataEntry } from "./api";

// Never guess a numeric compatibility identifier from a label or filename.
export function exactECoSSymbol(code: number, symbols: MasterDataEntry[]) {
  if (!Number.isInteger(code) || code < 2 || code > 4095) return undefined;
  const matches = symbols.filter(symbol => {
    if (!symbol.active) return false;
    const codes = [symbol.metadata?.ecosCode, symbol.metadata?.esuCode]
      .filter(value => typeof value === "number" || (typeof value === "string" && /^\d+$/.test(value)))
      .map(Number);
    return codes.length > 0 && codes.every(value => value === code);
  });
  return matches.length === 1 ? matches[0] : undefined;
}
