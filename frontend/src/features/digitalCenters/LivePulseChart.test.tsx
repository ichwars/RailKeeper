import { describe, expect, it, vi } from "vitest";

import { drawPulseChart } from "./LivePulseChart";

describe("drawPulseChart", () => {
  it("draws every pulse sample as one native canvas line", () => {
    const context = canvasContext();

    drawPulseChart(context, [
      { at: "2026-08-21T14:35:08Z", repliesPerSecond: 0 },
      { at: "2026-08-21T14:35:38Z", repliesPerSecond: 20 },
      { at: "2026-08-21T14:36:08Z", repliesPerSecond: 10 }
    ], 240, 80, "#a5ec60");

    expect(context.clearRect).toHaveBeenCalledWith(0, 0, 240, 80);
    expect(context.moveTo).toHaveBeenCalledWith(0, 80);
    expect(context.lineTo).toHaveBeenNthCalledWith(1, 120, 0);
    expect(context.lineTo).toHaveBeenNthCalledWith(2, 240, 40);
    expect(context.strokeStyle).toBe("#a5ec60");
    expect(context.lineWidth).toBe(2);
    expect(context.stroke).toHaveBeenCalledOnce();
  });

  it("clears without drawing a misleading line for fewer than two samples", () => {
    const context = canvasContext();

    drawPulseChart(context, [{ at: "2026-08-21T14:35:08Z", repliesPerSecond: 4 }], 240, 80, "#a5ec60");

    expect(context.clearRect).toHaveBeenCalledWith(0, 0, 240, 80);
    expect(context.stroke).not.toHaveBeenCalled();
  });
});

function canvasContext() {
  return {
    clearRect: vi.fn(),
    beginPath: vi.fn(),
    moveTo: vi.fn(),
    lineTo: vi.fn(),
    stroke: vi.fn(),
    strokeStyle: "",
    lineWidth: 0
  } as unknown as CanvasRenderingContext2D;
}
