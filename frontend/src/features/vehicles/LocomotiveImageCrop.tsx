import { useEffect, useRef, type PointerEvent } from "react";
import { useI18n } from "../../shared/i18n";
import { validImageCrop, type ImageCrop } from "./locomotiveImage";

export function LocomotiveImageCrop({ bitmap, crop, onChange }: {
  bitmap: ImageBitmap; crop: ImageCrop; onChange: (crop: ImageCrop) => void;
}) {
  const { t } = useI18n();
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const start = useRef<{ x: number; y: number } | null>(null);
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const scale = Math.min(1, 800 / bitmap.width, 240 / bitmap.height);
    canvas.width = Math.max(1, Math.round(bitmap.width * scale));
    canvas.height = Math.max(1, Math.round(bitmap.height * scale));
    const context = canvas.getContext("2d");
    if (!context) return;
    context.drawImage(bitmap, 0, 0, canvas.width, canvas.height);
    if (!validImageCrop(crop, bitmap)) return;
    const x = crop.x * scale, y = crop.y * scale, w = crop.width * scale, h = crop.height * scale;
    context.fillStyle = "#00000088";
    context.fillRect(0, 0, canvas.width, y); context.fillRect(0, y + h, canvas.width, canvas.height - y - h);
    context.fillRect(0, y, x, h); context.fillRect(x + w, y, canvas.width - x - w, h);
    context.strokeStyle = "#000000"; context.lineWidth = 3; context.strokeRect(x, y, w, h);
    context.strokeStyle = "#ffffff"; context.lineWidth = 1; context.strokeRect(x, y, w, h);
  }, [bitmap, crop]);
  function position(event: PointerEvent<HTMLCanvasElement>) {
    const box = event.currentTarget.getBoundingClientRect();
    return { x: Math.round(Math.max(0, Math.min(1, (event.clientX - box.left) / box.width)) * bitmap.width),
      y: Math.round(Math.max(0, Math.min(1, (event.clientY - box.top) / box.height)) * bitmap.height) };
  }
  return <canvas ref={canvasRef} className="locomotive-image-crop" role="img" tabIndex={0}
    style={{ width: bitmap.width / bitmap.height * 240 }}
    aria-label={t("locomotiveImage.cropCanvas")} onPointerDown={event => {
      if (!event.isPrimary || event.button !== 0) return;
      event.preventDefault(); event.currentTarget.focus();
      start.current = position(event); event.currentTarget.setPointerCapture(event.pointerId);
    }} onPointerMove={event => {
      if (!start.current) return;
      const end = position(event);
      const next = { x: Math.min(start.current.x, end.x), y: Math.min(start.current.y, end.y),
        width: Math.abs(end.x - start.current.x), height: Math.abs(end.y - start.current.y) };
      if (next.width >= 2 && next.height >= 2) onChange(next);
    }} onPointerUp={() => { start.current = null; }} onPointerCancel={() => { start.current = null; }}
    onLostPointerCapture={() => { start.current = null; }} />;
}
