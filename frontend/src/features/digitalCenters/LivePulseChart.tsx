import { useCallback, useEffect, useRef } from "react";

import { useI18n } from "../../shared/i18n";
import type { ECoSLivePulseSample } from "./digitalCenterModel";

export function drawPulseChart(
  context: CanvasRenderingContext2D,
  samples: ECoSLivePulseSample[],
  width: number,
  height: number,
  color: string
) {
  context.clearRect(0, 0, width, height);
  if (samples.length < 2) return;
  const maximum = Math.max(1, ...samples.map((sample) => sample.repliesPerSecond));
  context.strokeStyle = color;
  context.lineWidth = 2;
  context.beginPath();
  samples.forEach((sample, index) => {
    const x = index * (width / Math.max(1, samples.length - 1));
    const y = height - (sample.repliesPerSecond / maximum) * height;
    if (index === 0) context.moveTo(x, y);
    else context.lineTo(x, y);
  });
  context.stroke();
}

export function LivePulseChart({ samples }: { samples: ECoSLivePulseSample[] }) {
  const { t } = useI18n();
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const render = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
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
    const rootStyle = getComputedStyle(document.documentElement);
    const accent = rootStyle.getPropertyValue("--accent").trim() || rootStyle.color;
    drawPulseChart(context, samples, bounds.width, bounds.height, accent);
  }, [samples]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
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
  }, [render]);

  return (
    <div className="digital-live-chart">
      <strong>{t("digitalCenters.chart.title")}</strong>
      <canvas ref={canvasRef} role="img" aria-label={t("digitalCenters.chart.label")}
        data-sample-count={samples.length} />
    </div>
  );
}
