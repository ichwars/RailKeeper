import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { CircleCheck, Download, ExternalLink, Image, TriangleAlert, X } from "lucide-react";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import type { PendingArticleImage } from "./vehicleTransforms";
import { LocomotiveImageCrop } from "./LocomotiveImageCrop";
import {
  imageSuitability, locomotiveImageFileName, locomotiveImageTargets, validImageCrop,
  type ImageBackground, type ImageCrop, type LocomotiveImageTarget
} from "./locomotiveImage";
import { loadLocomotiveImage, renderLocomotiveImage, transparentImageBounds } from "./locomotiveImageCanvas";

type ExportFile = { key: string; url: string; size: number };
const fullCrop = (image: ImageBitmap): ImageCrop => ({ x: 0, y: 0, width: image.width, height: image.height });
const errorKeys = new Set(["localImage", "loadFailed", "tooLarge", "typeBlocked", "unsupported"]);

export function LocomotiveImageExportDialog({ image, onClose }: { image: PendingArticleImage; onClose: () => void }) {
  const { t } = useI18n();
  const closeRef = useRef<HTMLButtonElement>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(onClose, closeRef);
  const resultRef = useRef<HTMLCanvasElement>(null);
  const [target, setTarget] = useState<LocomotiveImageTarget>(locomotiveImageTargets[0]);
  const [background, setBackground] = useState<ImageBackground>(target.background);
  const [bitmap, setBitmap] = useState<ImageBitmap | null>(null);
  const [crop, setCrop] = useState<ImageCrop>({ x: 0, y: 0, width: 0, height: 0 });
  const [trim, setTrim] = useState<ImageCrop | null>(null);
  const [loadError, setLoadError] = useState("");
  const [outputError, setOutputError] = useState(false);
  const [file, setFile] = useState<ExportFile | null>(null);
  const key = JSON.stringify([image.url, target.id, background, crop]);
  const usable = bitmap !== null && trim !== null && validImageCrop(crop, bitmap);
  const suitability = bitmap ? imageSuitability(crop, bitmap, target, trim || undefined) : null;
  const warnings = suitability ? [suitability.upscale && "upscale", suitability.narrow && "narrow",
    suitability.border && "border"].filter((value): value is string => Boolean(value)) : [];

  useEffect(() => {
    // Clicking noninteractive preview space can leave focus on the document body.
    const escape = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !event.defaultPrevented && document.activeElement === document.body) {
        event.preventDefault(); onClose();
      }
    };
    window.addEventListener("keydown", escape);
    return () => window.removeEventListener("keydown", escape);
  }, [onClose]);

  useEffect(() => {
    const controller = new AbortController();
    let loaded: ImageBitmap | undefined;
    setLoadError(""); setBitmap(null);
    void loadLocomotiveImage(image.url, controller.signal).then(value => {
      loaded = value;
      if (controller.signal.aborted) { value.close(); return; }
      setTrim(transparentImageBounds(value)); setCrop(fullCrop(value)); setBitmap(value);
    }).catch(error => {
      if (!controller.signal.aborted) setLoadError(error instanceof Error && errorKeys.has(error.message) ? error.message : "loadFailed");
    });
    return () => { controller.abort(); loaded?.close(); };
  }, [image.url]);

  useEffect(() => {
    let active = true, url = "";
    setFile(null); setOutputError(false);
    resultRef.current?.getContext("2d")?.clearRect(0, 0, target.width, target.height);
    if (usable && bitmap && resultRef.current) {
      const canvas = resultRef.current;
      void Promise.resolve().then(() => renderLocomotiveImage(canvas, bitmap, crop, target, background)).then(blob => {
        if (!active) return;
        url = URL.createObjectURL(blob); setFile({ key, url, size: blob.size });
      }).catch(() => { if (active) setOutputError(true); });
    }
    return () => { active = false; if (url) URL.revokeObjectURL(url); };
  }, [bitmap, crop, target, background, usable, key]);

  const filename = locomotiveImageFileName(image.title || "Lokbild", target);
  const ready = usable && !outputError && file?.key === key;
  const error = loadError || (outputError ? "exportFailed" : bitmap && !trim ? "empty" : bitmap && !usable ? "cropInvalid" : "");

  return <>
    <span ref={anchorRef} aria-hidden="true" />
    {createPortal(<div className="locomotive-image-layer" ref={layerRef} role="dialog" aria-modal="true"
      aria-label={t("locomotiveImage.title")} onKeyDown={onKeyDown}>
      <section className="locomotive-image-dialog">
        <header>
          <div><h2><Image size={20} aria-hidden="true" />{t("locomotiveImage.title")}</h2>
            <p>{image.title || t("locomotiveImage.original")}</p></div>
          <button type="button" className="icon-button" ref={closeRef} onClick={onClose}
            aria-label={t("vehicles.close")}><X size={19} aria-hidden="true" /></button>
        </header>
        <div className="locomotive-image-body">
          <div className="locomotive-image-target">
            <label>{t("locomotiveImage.target")}<AppSelect aria-label={t("locomotiveImage.target")} value={target.id} onChange={event => {
              const next = locomotiveImageTargets.find(item => item.id === event.target.value);
              if (next) { setTarget(next); setBackground(next.background); }
            }}>{locomotiveImageTargets.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}</AppSelect></label>
            <div><strong>{target.width} × {target.height} px · {target.format === "bmp" ? "BMP · 24 Bit RGB" : "PNG"}</strong>
              <span>{t(target.fixed ? "locomotiveImage.fixed" : "locomotiveImage.preset")}</span></div>
          </div>
          <p className="locomotive-image-target-help">{t(`locomotiveImage.target.${target.id}`)}{" "}
            <a href={target.source} target="_blank" rel="noreferrer">{t("locomotiveImage.source")} <ExternalLink size={13} aria-hidden="true" /></a></p>
          {error ? <p className="locomotive-image-error" role="alert"><TriangleAlert size={18} aria-hidden="true" />
            {t(`locomotiveImage.error.${error}`)}</p> : !bitmap ? <p role="status">{t("locomotiveImage.loading")}</p> : null}
          {bitmap && <div className="locomotive-image-workspace">
            <section>
              <h3>{t("locomotiveImage.crop")}</h3>
              <p>{t("locomotiveImage.cropHelp")}</p>
              <div className="locomotive-image-source"><LocomotiveImageCrop bitmap={bitmap} crop={crop} onChange={setCrop} /></div>
              <p className="locomotive-image-dimensions">{t("locomotiveImage.dimensions", { width: bitmap.width, height: bitmap.height })}</p>
              <div className="locomotive-image-tools">
                <button type="button" className="secondary-button" onClick={() => setCrop(fullCrop(bitmap))}>{t("locomotiveImage.reset")}</button>
                <button type="button" className="secondary-button" disabled={!trim ||
                  trim.x === 0 && trim.y === 0 && trim.width === bitmap.width && trim.height === bitmap.height}
                  onClick={() => { if (trim) setCrop(trim); }}>{t("locomotiveImage.trim")}</button>
              </div>
              <details className="locomotive-image-crop-fields"><summary>{t("locomotiveImage.cropFields")}</summary>
                <div>{(["x", "y", "width", "height"] as const).map(field => <AppNumberInput key={field}
                  label={t(`locomotiveImage.${field}`)} value={String(crop[field])} min={field === "x" || field === "y" ? 0 : 1}
                  step={1} onValueChange={value => setCrop(current => ({ ...current, [field]: Number(value) }))} />)}</div>
              </details>
            </section>
            <section>
              <h3>{t("locomotiveImage.preview")}</h3><p>{t("locomotiveImage.previewHelp")}</p>
              <div className="locomotive-image-result" tabIndex={0} role="region" aria-label={t("locomotiveImage.preview")}>
                <canvas ref={resultRef} width={target.width} height={target.height} role="img" aria-label={t("locomotiveImage.resultCanvas")} />
              </div>
              <label className="locomotive-image-background">{t("locomotiveImage.background")}<AppSelect aria-label={t("locomotiveImage.background")} value={background}
                onChange={event => {
                  const next = event.target.value;
                  if (next === "white" || next === "grey" || next === "black" || next === "transparent" && target.format === "png") setBackground(next);
                }}>
                {(["transparent", "white", "grey", "black"] as const).filter(item => target.format === "png" || item !== "transparent")
                  .map(item => <option key={item} value={item}>{t(`locomotiveImage.background.${item}`)}</option>)}
              </AppSelect></label>
              <p>{t("locomotiveImage.backgroundHelp")}</p>
              <div className={`locomotive-image-assessment ${error || warnings.length ? "warning" : ""}`} role="status">
                <strong>{error || warnings.length ? <TriangleAlert size={18} aria-hidden="true" /> : <CircleCheck size={18} aria-hidden="true" />}
                  {t(error ? "locomotiveImage.notReady" : warnings.length ? "locomotiveImage.check" : "locomotiveImage.usable")}</strong>
                {warnings.length > 0 && <ul>{warnings.map(warning => <li key={warning}>{t(`locomotiveImage.warning.${warning}`)}</li>)}</ul>}
                <p>{t("locomotiveImage.motifHelp")}</p>
              </div>
            </section>
          </div>}
        </div>
        <footer>
          <span>{ready ? `${filename} · ${Math.ceil(file.size / 1024)} KB` : t("locomotiveImage.originalSafe")}</span>
          <button type="button" className="secondary-button" onClick={onClose}>{t("vehicles.close")}</button>
          {ready ? <a className="primary-button" href={file.url} download={filename}>
            <Download size={16} aria-hidden="true" />{t("locomotiveImage.download")}</a> :
            <button type="button" className="primary-button" disabled><Download size={16} aria-hidden="true" />{t("locomotiveImage.download")}</button>}
        </footer>
      </section>
    </div>, document.body)}
  </>;
}
