import type { ReactNode } from "react";

type FallbackKey =
  | "light"
  | "sound"
  | "horn"
  | "coupling"
  | "smoke"
  | "drive"
  | "warning"
  | "standard";

type Props = { symbolKey?: string; functionType?: string };

const primary = "function-symbol-primary";
const accent = "function-symbol-accent";

const fallbackGeometry: Record<FallbackKey, ReactNode> = {
  light: (
    <>
      <path className={primary} d="M18 18h14c9 0 16 6 16 14s-7 14-16 14H18zM23 23v18" />
      <path className={accent} d="M12 23l-6-4M12 32H5M12 41l-6 4" />
    </>
  ),
  sound: (
    <>
      <path className={primary} d="M11 27h10l11-9v28l-11-9H11z" />
      <path className={accent} d="M39 24c4 4 4 12 0 16M45 18c8 8 8 20 0 28" />
    </>
  ),
  horn: (
    <>
      <path className={primary} d="M10 27h15l22-10v30L25 37H10zM14 37v8" />
      <path className={accent} d="M49 26c5 3 5 9 0 12" />
    </>
  ),
  coupling: (
    <>
      <path className={primary} d="M8 31h14l6-8v18l-6-8H8M56 31H42l-6-8v18l6-8h14" />
      <circle className={accent} cx="32" cy="32" r="4" />
    </>
  ),
  smoke: (
    <>
      <path className={primary} d="M17 48h30M21 48V29h22v19M27 29V18h10v11" />
      <path className={accent} d="M25 13c-4-5 4-6 0-11M34 13c-4-5 4-6 0-11M43 13c-4-5 4-6 0-11" />
    </>
  ),
  drive: (
    <>
      <circle className={primary} cx="24" cy="39" r="12" />
      <circle className={primary} cx="24" cy="39" r="4" />
      <path className={primary} d="M24 27v24M12 39h24" />
      <path className={accent} d="M36 18h19l-7-7M55 18l-7 7" />
    </>
  ),
  warning: (
    <>
      <path className={primary} d="M32 7 58 54H6z" />
      <path className={accent} d="M32 23v15M32 46v1" />
    </>
  ),
  standard: (
    <>
      <circle className={primary} cx="32" cy="32" r="22" />
      <path className={primary} d="M20 32h24M32 20v24" />
      <circle className={accent} cx="32" cy="32" r="5" />
    </>
  ),
};

function resolveFallbackKey(symbolKey?: string, functionType?: string): FallbackKey {
  const value = symbolKey || functionType || "standard";
  if (value === "licht") return "light";
  if (value === "kupplung") return "coupling";
  if (value === "rauch") return "smoke";
  return value in fallbackGeometry ? (value as FallbackKey) : "standard";
}

export function RailKeeperFunctionSymbolFallback({ symbolKey, functionType }: Props) {
  const key = resolveFallbackKey(symbolKey, functionType);
  return (
    <svg viewBox="0 0 64 64" data-rk-function-symbol={key} aria-hidden="true">
      {fallbackGeometry[key]}
    </svg>
  );
}
