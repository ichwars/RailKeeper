import { describe, expect, it } from "vitest";
import {
  alphaBounds, encodeBMP24, fitLocomotiveImage, imageSuitability, locomotiveImageFileName,
  locomotiveImageTargets, validImageCrop, validImageSize
} from "./locomotiveImage";

describe("command station image compatibility", () => {
  it("includes all four supported providers and keeps ECoS fixed while distinguishing app/tool presets", () => {
    expect(locomotiveImageTargets.map(target => target.id)).toEqual(["ecos", "cs3", "intellibox3", "z21"]);
    expect(locomotiveImageTargets[0]).toMatchObject({ width: 190, height: 40, format: "bmp", fixed: true });
    expect(locomotiveImageTargets[2]).toMatchObject({ width: 340, height: 72, format: "png", fixed: false });
    expect(locomotiveImageTargets.slice(1).every(target => !target.fixed && target.format === "png")).toBe(true);
  });

  it("writes a Windows RGB24 BMP with bottom-up BGR pixels and correct row padding", () => {
    const pixels = new Uint8ClampedArray([
      255, 0, 0, 255, 0, 255, 0, 255,
      0, 0, 255, 255, 255, 255, 255, 255
    ]);
    const bmp = encodeBMP24(pixels, 2, 2), header = new DataView(bmp);
    expect(header.getUint16(0, true)).toBe(0x4d42);
    expect(header.getUint32(2, true)).toBe(70);
    expect(header.getUint32(10, true)).toBe(54);
    expect([header.getInt32(18, true), header.getInt32(22, true)]).toEqual([2, 2]);
    expect(header.getUint16(28, true)).toBe(24);
    expect(header.getUint32(30, true)).toBe(0);
    expect([...new Uint8Array(bmp).slice(54)]).toEqual([
      255, 0, 0, 255, 255, 255, 0, 0,
      0, 0, 255, 0, 255, 0, 0, 0
    ]);
    expect([...pixels.slice(0, 4)]).toEqual([255, 0, 0, 255]);
  });

  it("refuses implicit transparency loss, incomplete pixels and impossible allocations", () => {
    expect(() => encodeBMP24(new Uint8ClampedArray([0, 0, 0, 128]), 1, 1)).toThrow("background");
    expect(() => encodeBMP24(new Uint8ClampedArray(3), 1, 1)).toThrow("invalid");
    expect(validImageSize({ width: 40000, height: 1 })).toBe(false);
    expect(validImageSize({ width: 10000, height: 10000 })).toBe(false);
    expect(validImageCrop({ x: -1, y: 0, width: 10, height: 10 }, { width: 100, height: 100 })).toBe(false);
    expect(validImageCrop({ x: 0, y: 0, width: Infinity, height: 1 }, { width: 100, height: 100 })).toBe(false);
  });

  it("fits the entire selected locomotive without stretching or cropping it", () => {
    const crop = { x: 20, y: 30, width: 800, height: 100 };
    const fit = fitLocomotiveImage(crop, locomotiveImageTargets[0]);
    expect(fit.width / fit.height).toBeCloseTo(8);
    expect(fit.x).toBeGreaterThanOrEqual(1);
    expect(fit.y).toBeGreaterThanOrEqual(1);
    expect(fit.width + fit.x).toBeLessThan(190);
    expect(crop).toEqual({ x: 20, y: 30, width: 800, height: 100 });
  });

  it("warns about small, portrait and border-heavy selections without inventing subject recognition", () => {
    const target = locomotiveImageTargets[0];
    expect(imageSuitability({ x: 0, y: 0, width: 20, height: 5 }, { width: 20, height: 5 }, target).upscale).toBe(true);
    expect(imageSuitability({ x: 0, y: 0, width: 100, height: 800 }, { width: 100, height: 800 }, target).narrow).toBe(true);
    expect(imageSuitability({ x: 0, y: 0, width: 900, height: 300 }, { width: 900, height: 300 }, target,
      { x: 100, y: 100, width: 600, height: 80 }).border).toBe(true);
    expect(imageSuitability({ x: 800, y: 0, width: 900, height: 300 }, { width: 900, height: 300 }, target).invalid).toBe(true);
  });

  it("only trims truly transparent pixels and handles empty images explicitly", () => {
    const data = new Uint8ClampedArray(4 * 4 * 4);
    expect(alphaBounds(data, 4, 4)).toBeNull();
    data[(1 * 4 + 1) * 4 + 3] = 1;
    data[(2 * 4 + 3) * 4 + 3] = 255;
    expect(alphaBounds(data, 4, 4)).toEqual({ x: 1, y: 1, width: 3, height: 2 });
    data.fill(255);
    expect(alphaBounds(data, 4, 4)).toEqual({ x: 0, y: 0, width: 4, height: 4 });
  });

  it("creates safe, distinct filenames for each target", () => {
    expect(locomotiveImageFileName('../BR 86 – Göppingen.png', locomotiveImageTargets[0])).toBe('BR-86-Goppingen_ecos.bmp');
    expect(locomotiveImageFileName('', locomotiveImageTargets[3])).toBe('lokbild_z21.png');
  });
});
