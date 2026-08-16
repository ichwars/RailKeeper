import type { Language } from "../../shared/i18n";

const digitsOnly = (value: string) => /^\d+$/.test(value);

function groupedInteger(value: string, separator: string): string | undefined {
  const parts = value.split(separator);
  if (parts.length < 2 || !/^\d{1,3}$/.test(parts[0] || "") ||
      parts.slice(1).some((part) => !/^\d{3}$/.test(part))) return undefined;
  return parts.join("");
}

function canonicalMoney(integer: string, fraction = ""): string {
  const canonicalInteger = integer.replace(/^0+(?=\d)/, "") || "0";
  return `${canonicalInteger}.${fraction.padEnd(2, "0")}`;
}

export function normalizeAccessoryMoney(value: string): string | undefined {
  const trimmed = value.trim();
  if (!trimmed) return "";
  if (!/^[\d.,]+$/.test(trimmed) || /^[.,]|[.,]$/.test(trimmed)) return undefined;

  const dotCount = (trimmed.match(/\./g) || []).length;
  const commaCount = (trimmed.match(/,/g) || []).length;
  if (dotCount === 0 && commaCount === 0) {
    return digitsOnly(trimmed) ? canonicalMoney(trimmed) : undefined;
  }

  if (dotCount > 0 && commaCount > 0) {
    const decimalSeparator = trimmed.lastIndexOf(",") > trimmed.lastIndexOf(".") ? "," : ".";
    const groupingSeparator = decimalSeparator === "," ? "." : ",";
    const parts = trimmed.split(decimalSeparator);
    if (parts.length !== 2 || !/^\d{1,2}$/.test(parts[1] || "")) return undefined;
    const integer = groupedInteger(parts[0] || "", groupingSeparator);
    return integer === undefined ? undefined : canonicalMoney(integer, parts[1]);
  }

  const separator = commaCount > 0 ? "," : ".";
  const parts = trimmed.split(separator);
  if (parts.length === 2 && /^\d+$/.test(parts[0] || "") && /^\d{1,2}$/.test(parts[1] || "")) {
    return canonicalMoney(parts[0] || "", parts[1]);
  }
  const integer = groupedInteger(trimmed, separator);
  return integer === undefined ? undefined : canonicalMoney(integer);
}

export function formatAccessoryMoney(value: string | undefined, language: Language): string {
  const canonical = normalizeAccessoryMoney(value || "");
  if (!canonical) return "";
  const [integer = "0", fraction = "00"] = canonical.split(".");
  const groupingSeparator = language === "de" ? "." : ",";
  const decimalSeparator = language === "de" ? "," : ".";
  const groupedInteger = integer.replace(/\B(?=(\d{3})+(?!\d))/g, groupingSeparator);
  return `${groupedInteger}${decimalSeparator}${fraction} €`;
}
