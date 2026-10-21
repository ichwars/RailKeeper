import { afterEach, expect, it, vi } from "vitest";
import { loadLocomotiveImage } from "./locomotiveImageCanvas";

afterEach(() => vi.unstubAllGlobals());
const signal = () => new AbortController().signal;
function decoder(width = 400, height = 80) {
  const bitmap = { width, height, close: vi.fn() };
  const decode = vi.fn().mockResolvedValue(bitmap);
  vi.stubGlobal("createImageBitmap", decode);
  return { bitmap, decode };
}

it("never fetches external image URLs, data/SVG sources or lookalike origins", async () => {
  const fetch = vi.fn(); vi.stubGlobal("fetch", fetch);
  for (const url of ["https://example.test/x.png", "data:image/svg+xml,<svg/>", "javascript:alert(1)",
    window.location.origin + ".example.test/x.png"]) {
    await expect(loadLocomotiveImage(url, signal())).rejects.toThrow("localImage");
  }
  expect(fetch).not.toHaveBeenCalled();
});

it("sniffs image data instead of trusting an image MIME type and enforces the size bound", async () => {
  const { decode } = decoder();
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response("<svg/>", { headers: { "Content-Type": "image/png" } })));
  await expect(loadLocomotiveImage('/api/v1/images/source', signal())).rejects.toThrow("typeBlocked");
  expect(decode).not.toHaveBeenCalled();
  const cancel = vi.fn();
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(new ReadableStream({
    start(controller) { controller.enqueue(new Uint8Array(20 * 1024 * 1024 + 1)); }, cancel
  }))));
  await expect(loadLocomotiveImage('/api/v1/images/source', signal())).rejects.toThrow("tooLarge");
  expect(cancel).toHaveBeenCalled(); expect(decode).not.toHaveBeenCalled();
});

it("checks PNG dimensions before decoding and closes oversized decoded images", async () => {
  const { bitmap, decode } = decoder(12001, 40);
  const png = new Uint8Array(32); png.set([137, 80, 78, 71]);
  new DataView(png.buffer).setUint32(16, 15000); new DataView(png.buffer).setUint32(20, 40);
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(png)));
  await expect(loadLocomotiveImage('/api/v1/images/source', signal())).rejects.toThrow("tooLarge");
  expect(decode).not.toHaveBeenCalled();
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(new Uint8Array([255,216,255,0]))));
  await expect(loadLocomotiveImage('/api/v1/images/source', signal())).rejects.toThrow("tooLarge");
  expect(bitmap.close).toHaveBeenCalled();
});

it("uses authenticated same-origin reads, returns the original dimensions and honors cancellation", async () => {
  const { bitmap } = decoder();
  const fetch = vi.fn().mockImplementation(() => Promise.resolve(new Response(new Uint8Array([255,216,255,0]))));
  vi.stubGlobal("fetch", fetch);
  const controller = new AbortController();
  expect(await loadLocomotiveImage('/api/v1/images/source', controller.signal)).toBe(bitmap);
  expect(fetch).toHaveBeenCalledWith(window.location.origin + '/api/v1/images/source', { signal: controller.signal, credentials: "same-origin" });
  controller.abort();
  await expect(loadLocomotiveImage('/api/v1/images/source', controller.signal)).rejects.toThrow();
  expect(bitmap.close).toHaveBeenCalled();
});
