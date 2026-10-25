import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Download, ExternalLink, Image, TriangleAlert, X } from "lucide-react";
import { api } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { AppCheckbox } from "../../shared/ui/AppCheckbox";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import type { PendingArticleImage } from "./vehicleTransforms";
import { LocomotiveImageCrop } from "./LocomotiveImageCrop";
import { ECoSImageTransferPanel } from "./ECoSImageTransferPanel";
import {
  locomotiveImageFileName, locomotiveImageTargets, validImageCrop,
  type ImageBackground, type ImageCrop, type LocomotiveImageTarget
} from "./locomotiveImage";
import { analyzeLocomotiveImage, loadLocomotiveImage, renderLocomotiveImage } from "./locomotiveImageCanvas";

type ExportFile = { key: string; url: string; size: number; blob: Blob };
const fullCrop = (image: ImageBitmap): ImageCrop => ({ x: 0, y: 0, width: image.width, height: image.height });
const errorKeys = new Set(["localImage", "loadFailed", "tooLarge", "typeBlocked", "unsupported"]);

export function LocomotiveImageExportDialog({ image, onClose, transfer }: {
  image: PendingArticleImage; onClose: () => void;
  transfer?: { vehicleId: string; mappedObjectId?: string };
}) {
  const { t } = useI18n();
  const [transferBusy, setTransferBusy] = useState(false);
  const close = () => { if (!transferBusy) onClose(); };
  const closeRef = useRef<HTMLButtonElement>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(close, closeRef);
  const resultRef = useRef<HTMLCanvasElement>(null);
  const [target, setTarget] = useState<LocomotiveImageTarget>(locomotiveImageTargets[0]);
  const [backgroundChoice, setBackgroundChoice] = useState<ImageBackground | "auto">("auto");
  const background = backgroundChoice === "auto" ? target.background : backgroundChoice;
  const [backgroundMask, setBackgroundMask] = useState<HTMLCanvasElement | null>(null);
  const [replaceBackground, setReplaceBackground] = useState(true);
  const [enlarged, setEnlarged] = useState(true);
  const targetChosen = useRef(false);
  const canReadSettings = Boolean(transfer);
  const [bitmap, setBitmap] = useState<ImageBitmap | null>(null);
  const [crop, setCrop] = useState<ImageCrop>({ x: 0, y: 0, width: 0, height: 0 });
  const [trim, setTrim] = useState<ImageCrop | null>(null);
  const [loadError, setLoadError] = useState("");
  const [outputError, setOutputError] = useState(false);
  const [file, setFile] = useState<ExportFile | null>(null);
  const key = JSON.stringify([image.url, target.id, background, replaceBackground, crop]);
  const usable = bitmap !== null && trim !== null && validImageCrop(crop, bitmap);

  useEffect(() => {
    if (!canReadSettings) return;
    let active = true;
    void api.digitalSettings().then(settings => {
      const configured = locomotiveImageTargets.find(item => item.id === settings.provider);
      if (active && !targetChosen.current && configured) setTarget(configured);
    }).catch(() => { /* Preparation remains available with the manually selectable defaults. */ });
    return () => { active = false; };
  }, [canReadSettings]);

  useEffect(() => {
    // Clicking noninteractive preview space can leave focus on the document body.
    const escape = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !event.defaultPrevented && document.activeElement === document.body) {
        event.preventDefault(); if (!transferBusy) onClose();
      }
    };
    window.addEventListener("keydown", escape);
    return () => window.removeEventListener("keydown", escape);
  }, [onClose, transferBusy]);

  useEffect(() => {
    const controller = new AbortController();
    let loaded: ImageBitmap | undefined;
    setLoadError(""); setBitmap(null); setBackgroundMask(null);
    void loadLocomotiveImage(image.url, controller.signal).then(value => {
      loaded = value;
      if (controller.signal.aborted) { value.close(); return; }
      const analysis = analyzeLocomotiveImage(value);
      setTrim(analysis.bounds); setCrop(analysis.bounds || fullCrop(value));
      setBackgroundMask(analysis.backgroundMask); setBitmap(value);
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
      void Promise.resolve().then(() => renderLocomotiveImage(canvas, bitmap, crop, target, background,
        replaceBackground ? backgroundMask : null)).then(blob => {
        if (!active) return;
        url = URL.createObjectURL(blob); setFile({ key, url, size: blob.size, blob });
      }).catch(() => { if (active) setOutputError(true); });
    }
    return () => { active = false; if (url) URL.revokeObjectURL(url); };
  }, [bitmap, crop, target, background, backgroundMask, replaceBackground, usable, key]);

  const filename = locomotiveImageFileName(image.title || "Lokbild", target);
  const ready = usable && !outputError && file?.key === key;
  const error = loadError || (outputError ? "exportFailed" : bitmap && !trim ? "empty" : bitmap && !usable ? "cropInvalid" : "");

  return <>
    <span ref={anchorRef} aria-hidden="true" />
    {createPortal(<div className="locomotive-image-layer" ref={layerRef} role="dialog" aria-modal="true"
      aria-label={t("locomotiveImage.title")} onKeyDown={event => { targetChosen.current = true; onKeyDown(event); }}>
      <section className="locomotive-image-dialog" onPointerDown={() => { targetChosen.current = true; }}>
        <header>
          <div><h2><Image size={20} aria-hidden="true" />{t("locomotiveImage.title")}</h2>
            <p>{image.title || t("locomotiveImage.original")}</p></div>
          <button type="button" className="icon-button" ref={closeRef} onClick={close} disabled={transferBusy}
            aria-label={t("vehicles.close")}><X size={19} aria-hidden="true" /></button>
        </header>
        <div className="locomotive-image-body">
          <fieldset className="locomotive-image-controls" disabled={transferBusy} inert={transferBusy}>
          <div className="locomotive-image-target">
            <label>{t("locomotiveImage.target")}<AppSelect aria-label={t("locomotiveImage.target")} value={target.id} onChange={event => {
              const next = locomotiveImageTargets.find(item => item.id === event.target.value);
              if (next) {
                targetChosen.current = true; setTarget(next);
                if (next.format === "bmp" && backgroundChoice === "transparent") setBackgroundChoice("auto");
              }
            }}>{locomotiveImageTargets.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}</AppSelect></label>
            <label>{t("locomotiveImage.background")}<AppSelect aria-label={t("locomotiveImage.background")}
              value={backgroundChoice} onChange={event => {
                const next = event.target.value;
                if (next === "auto" || next === "white" || next === "grey" || next === "black" ||
                  next === "transparent" && target.format === "png") {
                  targetChosen.current = true; setBackgroundChoice(next);
                }
              }}>
              <option value="auto">{t("locomotiveImage.background.auto", { color: t(`locomotiveImage.background.${target.background}`) })}</option>
              {(["transparent", "white", "grey", "black"] as const).filter(item => target.format === "png" || item !== "transparent")
                .map(item => <option key={item} value={item}>{t(`locomotiveImage.background.${item}`)}</option>)}
            </AppSelect></label>
          </div>
          <details className="locomotive-image-target-help"><summary>
            {target.width} × {target.height} px · {target.format === "bmp" ? "BMP · 24 Bit RGB" : "PNG"}
            {" · "}{t(target.fixed ? "locomotiveImage.fixed" : "locomotiveImage.preset")}</summary>
            <p>{t(`locomotiveImage.target.${target.id}`)}{" "}
            <a href={target.source} target="_blank" rel="noreferrer">{t("locomotiveImage.source")} <ExternalLink size={13} aria-hidden="true" /></a></p>
          </details>
          {error ? <p className="locomotive-image-error" role="alert"><TriangleAlert size={18} aria-hidden="true" />
            {t(`locomotiveImage.error.${error}`)}</p> : !bitmap ? <p role="status">{t("locomotiveImage.loading")}</p> : null}
          {bitmap && <div className="locomotive-image-workspace">
            <section>
              <h3>{t("locomotiveImage.crop")}</h3>
              <div className="locomotive-image-source"><LocomotiveImageCrop bitmap={bitmap} crop={crop} onChange={setCrop} /></div>
              <p className="locomotive-image-dimensions">{t("locomotiveImage.dimensions", { width: bitmap.width, height: bitmap.height })}</p>
              <div className="locomotive-image-tools">
                <button type="button" className="secondary-button" onClick={() => setCrop(fullCrop(bitmap))}>{t("locomotiveImage.reset")}</button>
                <button type="button" className="secondary-button" disabled={!trim ||
                  trim.x === 0 && trim.y === 0 && trim.width === bitmap.width && trim.height === bitmap.height}
                  onClick={() => { if (trim) setCrop(trim); }}>{t("locomotiveImage.trim")}</button>
              </div>
              <details className="locomotive-image-crop-fields"><summary>{t("locomotiveImage.cropFields")}</summary>
                <p>{t("locomotiveImage.cropHelp")}</p>
                <div>{(["x", "y", "width", "height"] as const).map(field => <AppNumberInput key={field}
                  label={t(`locomotiveImage.${field}`)} value={String(crop[field])} min={field === "x" || field === "y" ? 0 : 1}
                  step={1} onValueChange={value => setCrop(current => ({ ...current, [field]: Number(value) }))} />)}</div>
              </details>
            </section>
            <section>
              <div className="locomotive-image-preview-heading">
                <h3>{t("locomotiveImage.preview")}</h3>
                <button type="button" className="secondary-button" aria-pressed={!enlarged}
                  onClick={() => setEnlarged(value => !value)}>{t(enlarged ? "locomotiveImage.actualSize" : "locomotiveImage.enlarge")}</button>
              </div>
              <div className="locomotive-image-result" data-actual-size={!enlarged} data-transparent={background === "transparent"}
                tabIndex={0} role="region"
                aria-label={t("locomotiveImage.preview")}>
                <canvas ref={resultRef} width={target.width} height={target.height} role="img"
                  aria-label={t("locomotiveImage.resultCanvas")} />
              </div>
              <p className="locomotive-image-dimensions">{t(enlarged ? "locomotiveImage.previewHelp" : "locomotiveImage.actualSizeHelp")}</p>
              <AppCheckbox className="locomotive-image-replace" checked={replaceBackground && Boolean(backgroundMask)} disabled={!backgroundMask}
                onChange={event => setReplaceBackground(event.target.checked)} label={t("locomotiveImage.replaceBackground")} />
              <p className="locomotive-image-background-help">{t(backgroundMask
                ? "locomotiveImage.backgroundHelp" : "locomotiveImage.backgroundUnchanged")}</p>
            </section>
          </div>}
          </fieldset>
          {target.id === "ecos" && transfer && <ECoSImageTransferPanel key={key}
            vehicleId={transfer.vehicleId} mappedObjectId={transfer.mappedObjectId} title={image.title || "Lokbild"}
            blob={ready ? file.blob : null} onBusyChange={setTransferBusy} />}
        </div>
        <footer>
          <span>{ready ? `${filename} · ${Math.ceil(file.size / 1024)} KB` : t("locomotiveImage.originalSafe")}</span>
          <button type="button" className="secondary-button" onClick={close} disabled={transferBusy}>{t("vehicles.close")}</button>
          {ready ? <a className="primary-button" href={file.url} download={filename}>
            <Download size={16} aria-hidden="true" />{t("locomotiveImage.download")}</a> :
            <button type="button" className="primary-button" disabled><Download size={16} aria-hidden="true" />{t("locomotiveImage.download")}</button>}
        </footer>
      </section>
    </div>, document.body)}
  </>;
}
