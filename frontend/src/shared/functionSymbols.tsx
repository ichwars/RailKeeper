import { AlertTriangle, Circle, Cloud, Gauge, Lightbulb, Link, Megaphone, Volume2 } from "lucide-react";
import type { ReactNode } from "react";
import type { MasterDataEntry } from "./api";

const fallbackFunctionSymbols = [
  { key: "light", label: "Licht" },
  { key: "sound", label: "Sound" },
  { key: "horn", label: "Horn" },
  { key: "coupling", label: "Kupplung" },
  { key: "smoke", label: "Rauch" },
  { key: "drive", label: "Fahrt" },
  { key: "warning", label: "Warnung" }
];

function symbolImageFromMetadata(metadata?: Record<string, unknown>) {
  const value = metadata?.imageData || metadata?.activeImageData || metadata?.svgData;
  return typeof value === "string" ? value : "";
}

function functionSymbolTile(content: ReactNode) {
  return (
    <span className="function-symbol-tile" aria-hidden="true">
      {content}
    </span>
  );
}

export function functionSymbolIcon(symbolKey?: string, functionType?: string, metadata?: Record<string, unknown>) {
  const imageData = symbolImageFromMetadata(metadata);
  if (imageData) {
    return functionSymbolTile(<img className="function-symbol-image" src={imageData} alt="" />);
  }

  const key = symbolKey || functionType || "standard";
  const props = { size: 18, "aria-hidden": true };
  switch (key) {
    case "light":
    case "licht":
      return functionSymbolTile(<Lightbulb {...props} />);
    case "sound":
      return functionSymbolTile(<Volume2 {...props} />);
    case "horn":
      return functionSymbolTile(<Megaphone {...props} />);
    case "coupling":
    case "kupplung":
      return functionSymbolTile(<Link {...props} />);
    case "smoke":
    case "rauch":
      return functionSymbolTile(<Cloud {...props} />);
    case "drive":
      return functionSymbolTile(<Gauge {...props} />);
    case "warning":
      return functionSymbolTile(<AlertTriangle {...props} />);
    default:
      return functionSymbolTile(<Circle {...props} />);
  }
}

export function functionSymbolMetadata(symbols: MasterDataEntry[], key?: string) {
  if (!key) return undefined;
  return symbols.find((symbol) => symbol.key === key && symbol.active)?.metadata;
}

function functionSymbolOptions(symbols: MasterDataEntry[]) {
  const merged = new Map<string, { key: string; label: string; metadata?: Record<string, unknown> }>();
  for (const symbol of fallbackFunctionSymbols) {
    merged.set(symbol.key, symbol);
  }
  for (const symbol of symbols) {
    if (symbol.active) {
      merged.set(symbol.key, { key: symbol.key, label: symbol.label, metadata: symbol.metadata });
    }
  }
  return [...merged.values()];
}

export function FunctionSymbolPicker({
  value,
  functionType,
  symbols,
  disabled,
  label,
  onChange
}: {
  value?: string;
  functionType?: string;
  symbols: MasterDataEntry[];
  disabled?: boolean;
  label: string;
  onChange: (value: string) => void;
}) {
  const options = functionSymbolOptions(symbols);
  const selected = options.find((symbol) => symbol.key === value);
  return (
    <details className="function-symbol-picker">
      <summary aria-label={label}>
        {functionSymbolIcon(value, functionType, selected?.metadata)}
        <span>{selected?.label || "Symbol"}</span>
      </summary>
      <div className="function-symbol-menu">
        <button type="button" className={!value ? "active" : ""} onClick={() => onChange("")} disabled={disabled}>
          {functionSymbolTile(<Circle size={18} aria-hidden="true" />)}
          <span>Kein Symbol</span>
        </button>
        {options.map((symbol) => (
          <button type="button" key={symbol.key} className={value === symbol.key ? "active" : ""} onClick={() => onChange(symbol.key)} disabled={disabled} title={symbol.label}>
            {functionSymbolIcon(symbol.key, functionType, symbol.metadata)}
            <span>{symbol.label}</span>
          </button>
        ))}
      </div>
    </details>
  );
}
