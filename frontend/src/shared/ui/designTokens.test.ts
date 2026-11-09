import { readdirSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const directory = resolve(import.meta.dirname, "../../styles");
const styles = readdirSync(directory).filter(name => name.endsWith(".css"))
  .map(name => ({ name, source: readFileSync(resolve(directory, name), "utf8") }));

describe("shared typography contract", () => {
  it("defines every referenced typography token", () => {
    const definitions = new Set(styles.flatMap(({ source }) =>
      [...source.matchAll(/(--(?:font|line-height)[\w-]+)\s*:/g)].map(match => match[1])));
    const missing = styles.flatMap(({ name, source }) =>
      [...source.matchAll(/var\((--(?:font|line-height)[\w-]+)/g)]
        .filter(match => !definitions.has(match[1])).map(match => `${name}: ${match[1]}`));
    expect(missing).toEqual([]);
  });

  it("uses the approved scale and families outside the token definitions", () => {
    const deviations = styles.filter(({ name }) => name !== "base.css").flatMap(({ name, source }) =>
      [...source.matchAll(/font-(?:size|family)\s*:\s*([^;\n}]+)/g)]
        .filter(match => !/^(?:var\(|inherit\b|0\b)/.test(match[1]))
        .map(match => `${name}: ${match[0]}`));
    expect(deviations).toEqual([]);
  });
});
