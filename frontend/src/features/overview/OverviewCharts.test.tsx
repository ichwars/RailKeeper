import { act, render } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { InventoryTrendChart, ValueDistributionChart } from "./OverviewCharts";

function canvasFixture(width = 640, height = 240) {
  const context = {
    setTransform: vi.fn(), clearRect: vi.fn(), setLineDash: vi.fn(), beginPath: vi.fn(),
    moveTo: vi.fn(), lineTo: vi.fn(), stroke: vi.fn(), fillText: vi.fn(), closePath: vi.fn(),
    fill: vi.fn(), arc: vi.fn(), createLinearGradient: vi.fn(() => ({ addColorStop: vi.fn() }))
  };
  // This DOM adapter fixture supplies only the canvas drawing operations used by these charts.
  vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(context as unknown as CanvasRenderingContext2D);
  vi.spyOn(HTMLCanvasElement.prototype, "getBoundingClientRect").mockReturnValue(new DOMRect(0, 0, width, height));
  return context;
}

describe("overview chart rendering and lifecycle", () => {
  afterEach(() => document.documentElement.removeAttribute("style"));

  it("scales its backing pixels and plots both time series against a shared rounded scale", () => {
    vi.stubGlobal("devicePixelRatio", 2);
    const draw = canvasFixture();
    const { container } = render(<InventoryTrendChart points={[
      { label: "Before", vehicles: 0, accessories: 10 }, { label: "Now", vehicles: 20, accessories: 5 }
    ]} />);
    const canvas = container.querySelector("canvas")!;
    expect(canvas).toHaveAttribute("aria-hidden", "true");
    expect(canvas.width).toBe(1280); expect(canvas.height).toBe(480);
    expect(draw.setTransform).toHaveBeenCalledWith(2, 0, 0, 2, 0, 0);
    expect(draw.fillText).toHaveBeenCalledWith("20", 2, 10);
    expect(draw.fillText).toHaveBeenCalledWith("Before", 36, 221);
    expect(draw.fillText).toHaveBeenCalledWith("Now", 630, 221);
    expect(draw.arc).toHaveBeenCalledWith(630, 10, 3, 0, Math.PI * 2);
    expect(draw.arc).toHaveBeenCalledTimes(4);
    expect(draw.createLinearGradient).toHaveBeenCalledTimes(2);
  });

  it.each([{ points: [] }, { points: [{ label: "Only", vehicles: 0, accessories: 0 }] }])("keeps empty and single-point collections finite", ({ points }) => {
    const draw = canvasFixture(0, 0);
    render(<InventoryTrendChart points={points} />);
    for (const call of [...draw.lineTo.mock.calls, ...draw.moveTo.mock.calls, ...draw.arc.mock.calls]) {
      expect(call.every(Number.isFinite)).toBe(true);
    }
    expect(draw.fillText).toHaveBeenCalledWith("10", 2, 10);
  });

  it("keeps the final label while reducing narrow chart label density and redraws on theme/resize", async () => {
    vi.stubGlobal("ResizeObserver", undefined);
    const draw = canvasFixture(360);
    const points = Array.from({ length: 10 }, (_, i) => ({ label: `Month ${i}`, vehicles: i, accessories: i + 1 }));
    const { unmount } = render(<InventoryTrendChart points={points} />);
    expect(draw.fillText.mock.calls.filter(call => String(call[0]).startsWith("Month ")).map(call => call[0]))
      .toEqual(["Month 0", "Month 2", "Month 4", "Month 6", "Month 8", "Month 9"]);
    const first = draw.clearRect.mock.calls.length;
    act(() => window.dispatchEvent(new Event("resize")));
    expect(draw.clearRect.mock.calls.length).toBeGreaterThan(first);
    await act(async () => document.documentElement.style.setProperty("--accent", "#123456"));
    expect(draw.clearRect.mock.calls.length).toBeGreaterThan(first + 1);
    unmount();
    const last = draw.clearRect.mock.calls.length;
    act(() => window.dispatchEvent(new Event("resize")));
    await act(async () => document.documentElement.style.setProperty("--accent", "#654321"));
    expect(draw.clearRect).toHaveBeenCalledTimes(last);
  });

  it("represents a quarter vehicle and three-quarter accessory value without losing the neutral ring", () => {
    const draw = canvasFixture(200, 200);
    render(<ValueDistributionChart vehicleValue={25} accessoryValue={75} />);
    expect(draw.arc).toHaveBeenCalledTimes(3);
    const neutral = draw.arc.mock.calls[0], vehicles = draw.arc.mock.calls[1], accessories = draw.arc.mock.calls[2];
    expect(neutral.slice(0, 2)).toEqual([100, 100]);
    expect(neutral[4] - neutral[3]).toBeCloseTo(2 * Math.PI);
    expect(vehicles[4] - vehicles[3]).toBeCloseTo(Math.PI / 2);
    expect(accessories[4] - accessories[3]).toBeCloseTo(3 * Math.PI / 2);
    expect(vehicles[4]).toBe(accessories[3]);
  });

  it("draws only the neutral ring for an empty value distribution", () => {
    const draw = canvasFixture();
    render(<ValueDistributionChart vehicleValue={0} accessoryValue={0} />);
    expect(draw.arc).toHaveBeenCalledTimes(1);
  });

  it("tolerates a missing or failing canvas context", () => {
    vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockImplementationOnce(() => { throw new Error("canvas unavailable"); }).mockReturnValue(null);
    expect(() => render(<><InventoryTrendChart points={[]} /><ValueDistributionChart vehicleValue={1} accessoryValue={1} /></>)).not.toThrow();
  });
});
