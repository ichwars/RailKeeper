import { useRef, type ReactNode } from "react";
import type { MasterDataEntry } from "./api";
import { RailKeeperFunctionSymbolFallback } from "./functionSymbolFallbackIcons";
import { functionSymbolImageData } from "./functionSymbolImages";
import { useI18n } from "./i18n";
import { masterDataOptions } from "./masterDataOptions";

const fallbackFunctionSymbols = [
  { key: "light", label: "Licht" },
  { key: "sound", label: "Sound" },
  { key: "horn", label: "Horn" },
  { key: "coupling", label: "Kupplung" },
  { key: "smoke", label: "Rauch" },
  { key: "drive", label: "Fahrt" },
  { key: "warning", label: "Warnung" }
];

function functionSymbolTile(content: ReactNode) {
  return (
    <span className="function-symbol-tile" aria-hidden="true">
      {content}
    </span>
  );
}

export function functionSymbolIcon(symbolKey?: string, functionType?: string, metadata?: Record<string, unknown>) {
  const imageData = functionSymbolImageData(metadata, "active");
  if (imageData) {
    return functionSymbolTile(<img className="function-symbol-image" src={imageData} alt="" />);
  }
  return functionSymbolTile(
    <RailKeeperFunctionSymbolFallback symbolKey={symbolKey} functionType={functionType} />,
  );
}

export function functionSymbolMetadata(symbols: MasterDataEntry[], key?: string) {
  if (!key) return undefined;
  return symbols.find((symbol) => symbol.key === key)?.metadata;
}

function functionSymbolOptions(symbols: MasterDataEntry[], currentKey?: string) {
  const source = symbols.length > 0
    ? symbols
    : fallbackFunctionSymbols.map((symbol, index): MasterDataEntry => ({
        id: `fallback:${symbol.key}`,
        type: "symbols",
        key: symbol.key,
        label: symbol.label,
        active: true,
        sortOrder: index,
        metadata: {},
        createdAt: "",
        updatedAt: ""
      }));
  return masterDataOptions(source, [currentKey || ""], (symbol) => symbol.key).map((option) => ({
    key: option.value,
    label: option.label,
    active: option.active,
    metadata: source.find((symbol) => symbol.id === option.id)?.metadata
  }));
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
  onChange: (value: string, label?: string) => void;
}) {
  const { t } = useI18n();
  const detailsRef = useRef<HTMLDetailsElement | null>(null);
  const options = functionSymbolOptions(symbols, value);
  const selected = options.find((symbol) => symbol.key === value);
  const optionLabel = (symbol: { label: string; active: boolean }) => (
    `${symbol.label}${symbol.active ? "" : ` (${t("common.inactive")})`}`
  );
  const selectSymbol = (nextValue: string, nextLabel?: string) => {
    onChange(nextValue, nextLabel);
    detailsRef.current?.removeAttribute("open");
  };
  return (
    <details ref={detailsRef} className="function-symbol-picker">
      <summary aria-label={label}>
        {functionSymbolIcon(value, functionType, selected?.metadata)}
        <span>{selected ? optionLabel(selected) : "Symbol"}</span>
      </summary>
      <div className="function-symbol-menu">
        <button type="button" className={!value ? "active" : ""} onClick={() => selectSymbol("")} disabled={disabled}>
          {functionSymbolTile(<RailKeeperFunctionSymbolFallback symbolKey="standard" />)}
          <span>Kein Symbol</span>
        </button>
        {options.map((symbol) => (
          <button type="button" key={symbol.key} className={value === symbol.key ? "active" : ""} onClick={() => selectSymbol(symbol.key, symbol.label)} disabled={disabled || !symbol.active} title={optionLabel(symbol)}>
            {functionSymbolIcon(symbol.key, functionType, symbol.metadata)}
            <span>{optionLabel(symbol)}</span>
          </button>
        ))}
      </div>
    </details>
  );
}
