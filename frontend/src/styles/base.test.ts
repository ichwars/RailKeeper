import { readFileSync, readdirSync } from "node:fs";
import { basename, join, resolve } from "node:path";

import { describe, expect, it } from "vitest";

const sourceDirectory = resolve(process.cwd(), "src");
const stylesDirectory = join(sourceDirectory, "styles");
const baseCss = readFileSync(join(stylesDirectory, "base.css"), "utf8");

function cssFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) return cssFiles(path);
    return entry.isFile() && entry.name.endsWith(".css") ? [path] : [];
  });
}

describe("base design system", () => {
  it("defines the shared spacing, radius, duration, and easing scales", () => {
    const expectedTokens = [
      "--space-xs: 4px",
      "--space-sm: 8px",
      "--space-md: 12px",
      "--space-lg: 16px",
      "--space-xl: 24px",
      "--space-2xl: 32px",
      "--radius-xs: 4px",
      "--radius-sm: 6px",
      "--radius-md: 8px",
      "--radius-lg: 12px",
      "--radius-pill: 999px",
      "--duration-fast: 120ms",
      "--duration-normal: 160ms",
      "--duration-slow: 180ms",
      "--ease-standard: ease",
      "--ease-out: ease-out"
    ];

    for (const token of expectedTokens) expect(baseCss).toContain(token);
  });

  it("defines every radius token consumed by a frontend stylesheet", () => {
    const usedRadiusTokens = new Set(
      cssFiles(sourceDirectory).flatMap((file) =>
        [...readFileSync(file, "utf8").matchAll(/var\(--(radius-[a-z0-9-]+)\)/g)]
          .map((match) => match[1])
      )
    );
    const definedRadiusTokens = new Set(
      [...baseCss.matchAll(/--(radius-[a-z0-9-]+):/g)].map((match) => match[1])
    );

    expect([...usedRadiusTokens].filter((token) => !definedRadiusTokens.has(token))).toEqual([]);
  });

  it("globally reduces motion and smooth scrolling when requested", () => {
    expect(baseCss).toContain("@media (prefers-reduced-motion: reduce)");
    expect(baseCss).toMatch(/animation-duration:\s*0\.01ms\s*!important/);
    expect(baseCss).toMatch(/animation-iteration-count:\s*1\s*!important/);
    expect(baseCss).toMatch(/transition-duration:\s*0\.01ms\s*!important/);
    expect(baseCss).toMatch(/scroll-behavior:\s*auto\s*!important/);
  });

  it("bounds micro type to machine metadata and keeps user-facing small copy at caption size", () => {
    expect(baseCss).toContain("--font-size-micro: 11px");
    expect(baseCss).toContain("--font-size-caption: 12px");

    const styles = new Map(
      cssFiles(stylesDirectory).map((file) => [basename(file), readFileSync(file, "utf8")])
    );
    const combinedStyles = [...styles.entries()]
      .filter(([name]) => name !== "base.css")
      .map(([, css]) => css)
      .join("\n");

    expect(combinedStyles).not.toMatch(/font-size:\s*(?:11px|var\(--font-size-3xs\))/);
    expect(combinedStyles.match(/font-size:\s*var\(--font-size-micro\)/g)).toHaveLength(1);
    expect(styles.get("vehicle-uploads.css")).toMatch(
      /\.pending-image-card figcaption > span\s*\{[^}]*font-size:\s*var\(--font-size-micro\)/s
    );

    const captionRoles: Array<[string, RegExp]> = [
      ["forms-controls.css", /\.image-placeholder\s*\{[^}]*font-size:\s*var\(--font-size-caption\)/s],
      ["overview-dashboard.css", /\.overview-quality-list b\s*\{[^}]*font-size:\s*var\(--font-size-caption\)/s],
      ["exhibition.css", /\.event-list-state small\s*\{[^}]*font-size:\s*var\(--font-size-caption\)/s],
      ["exhibition.css", /\.exhibition-day-tabs small\s*\{[^}]*font-size:\s*var\(--font-size-caption\)/s],
      ["exhibition.css", /\.exhibition-quick-filters span\s*\{[^}]*font-size:\s*var\(--font-size-caption\)/s],
      ["vehicle-uploads.css", /\.pending-image-card input\s*\{[^}]*font-size:\s*var\(--font-size-caption\)/s]
    ];

    for (const [file, role] of captionRoles) expect(styles.get(file)).toMatch(role);
  });
});
