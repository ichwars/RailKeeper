import { useCallback, useEffect, useRef } from "react";

import type { OverviewTrendPoint } from "./overviewModel";

type CanvasDraw = (context: CanvasRenderingContext2D, width: number, height: number) => void;

function useOverviewCanvas(draw: CanvasDraw) {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const render = () => {
      const bounds = canvas.getBoundingClientRect();
      const scale = window.devicePixelRatio || 1;
      canvas.width = Math.max(1, Math.round(bounds.width * scale));
      canvas.height = Math.max(1, Math.round(bounds.height * scale));
      let context: CanvasRenderingContext2D | null = null;
      try {
        context = canvas.getContext("2d");
      } catch {
        return;
      }
      if (!context) return;
      context.setTransform(scale, 0, 0, scale, 0, 0);
      context.clearRect(0, 0, bounds.width, bounds.height);
      draw(context, bounds.width, bounds.height);
    };

    const resizeObserver = typeof ResizeObserver === "undefined" ? null : new ResizeObserver(render);
    const themeObserver = new MutationObserver(render);
    resizeObserver?.observe(canvas);
    if (!resizeObserver) window.addEventListener("resize", render);
    themeObserver.observe(document.documentElement, { attributes: true });
    render();
    return () => {
      resizeObserver?.disconnect();
      if (!resizeObserver) window.removeEventListener("resize", render);
      themeObserver.disconnect();
    };
  }, [draw]);

  return canvasRef;
}

function cssColor(name: string, fallback: string) {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim() || fallback;
}

export function InventoryTrendChart({ points }: { points: OverviewTrendPoint[] }) {
  const draw = useCallback<CanvasDraw>((context, width, height) => {
    const accent = cssColor("--accent", "#419310");
    const secondary = cssColor("--overview-secondary", "#289fea");
    const line = cssColor("--line", "#d5dfdc");
    const muted = cssColor("--muted", "#4f6869");
    const left = 36;
    const right = 10;
    const top = 10;
    const bottom = 27;
    const plotWidth = Math.max(1, width - left - right);
    const plotHeight = Math.max(1, height - top - bottom);
    const maximum = Math.max(1, ...points.flatMap((point) => [point.vehicles, point.accessories]));
    const roundedMaximum = Math.max(10, Math.ceil(maximum / 10) * 10);

    context.font = '10px "Segoe UI", sans-serif';
    context.lineWidth = 1;
    context.textBaseline = "middle";
    context.fillStyle = muted;
    for (let index = 0; index <= 4; index += 1) {
      const value = Math.round((roundedMaximum / 4) * (4 - index));
      const y = top + (plotHeight / 4) * index;
      context.strokeStyle = line;
      context.setLineDash([2, 3]);
      context.beginPath();
      context.moveTo(left, y);
      context.lineTo(width - right, y);
      context.stroke();
      context.fillText(String(value), 2, y);
    }
    context.setLineDash([]);

    const xFor = (index: number) => left + (points.length <= 1 ? plotWidth :
      (plotWidth / (points.length - 1)) * index);
    const yFor = (value: number) => top + plotHeight - (value / roundedMaximum) * plotHeight;
    const drawSeries = (key: "vehicles" | "accessories", color: string) => {
      const area = context.createLinearGradient(0, top, 0, top + plotHeight);
      area.addColorStop(0, color);
      area.addColorStop(1, "transparent");
      context.beginPath();
      points.forEach((point, index) => {
        const x = xFor(index);
        const y = yFor(point[key]);
        if (index === 0) context.moveTo(x, y);
        else context.lineTo(x, y);
      });
      context.lineTo(xFor(points.length - 1), top + plotHeight);
      context.lineTo(xFor(0), top + plotHeight);
      context.closePath();
      context.fillStyle = area;
      context.globalAlpha = 0.2;
      context.fill();
      context.globalAlpha = 1;

      context.strokeStyle = color;
      context.lineWidth = 2;
      context.beginPath();
      points.forEach((point, index) => {
        const x = xFor(index);
        const y = yFor(point[key]);
        if (index === 0) context.moveTo(x, y);
        else context.lineTo(x, y);
      });
      context.stroke();
      context.fillStyle = color;
      points.forEach((point, index) => {
        const x = xFor(index);
        const y = yFor(point[key]);
        context.beginPath();
        context.arc(x, y, 3, 0, Math.PI * 2);
        context.fill();
      });
    };
    drawSeries("vehicles", accent);
    drawSeries("accessories", secondary);

    context.fillStyle = muted;
    context.textAlign = "center";
    context.textBaseline = "top";
    const labelStep = points.length > 8 && width < 680 ? 2 : 1;
    points.forEach((point, index) => {
      if (index % labelStep !== 0 && index !== points.length - 1) return;
      context.fillText(point.label, xFor(index), height - bottom + 8);
    });
  }, [points]);
  const canvasRef = useOverviewCanvas(draw);

  return <canvas ref={canvasRef} className="overview-trend-canvas" aria-hidden="true" />;
}

export function ValueDistributionChart({ vehicleValue, accessoryValue }: {
  vehicleValue: number;
  accessoryValue: number;
}) {
  const draw = useCallback<CanvasDraw>((context, width, height) => {
    const accent = cssColor("--accent", "#419310");
    const secondary = cssColor("--overview-secondary", "#289fea");
    const line = cssColor("--line", "#d5dfdc");
    const total = vehicleValue + accessoryValue;
    const centerX = width / 2;
    const centerY = height / 2;
    const radius = Math.max(20, Math.min(width, height) / 2 - 8);
    const thickness = Math.max(12, radius * 0.34);
    context.lineWidth = thickness;
    context.lineCap = "butt";
    context.strokeStyle = line;
    context.beginPath();
    context.arc(centerX, centerY, radius - thickness / 2, 0, Math.PI * 2);
    context.stroke();
    if (total <= 0) return;
    const start = -Math.PI / 2;
    const vehicleEnd = start + (vehicleValue / total) * Math.PI * 2;
    context.strokeStyle = accent;
    context.beginPath();
    context.arc(centerX, centerY, radius - thickness / 2, start, vehicleEnd);
    context.stroke();
    context.strokeStyle = secondary;
    context.beginPath();
    context.arc(centerX, centerY, radius - thickness / 2, vehicleEnd, start + Math.PI * 2);
    context.stroke();
  }, [accessoryValue, vehicleValue]);
  const canvasRef = useOverviewCanvas(draw);

  return <canvas ref={canvasRef} className="overview-value-canvas" aria-hidden="true" />;
}
