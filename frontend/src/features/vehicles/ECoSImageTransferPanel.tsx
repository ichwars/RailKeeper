import { useEffect, useRef, useState } from "react";
import { CircleCheck, Send, TriangleAlert } from "lucide-react";
import { api, ApiError } from "../../shared/api";
import type { ECoSImageOptions, ECoSImagePreview, ECoSImageResult } from "../../shared/ecosImageModel";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppTextInput } from "../../shared/ui/AppTextInput";
import { AppCheckbox } from "../../shared/ui/AppCheckbox";
import { ECoSImagePauseDialog } from "./ECoSImagePauseDialog";
import { sameECoSImageReview } from "./ecosImageReview";

type Props = {
  vehicleId: string;
  mappedObjectId?: string;
  title: string;
  blob: Blob | null;
  onBusyChange: (busy: boolean) => void;
};

const knownErrors = new Set(["ecos_image_web", "ecos_image_read", "ecos_image_inventory", "ecos_image_form", "ecos_image_slots_full", "ecos_image_slot_occupied",
  "ecos_image_input", "digital_center_write_preview_stale", "digital_center_write_grant_conflict",
  "digital_center_not_configured", "digital_center_inactive"]);

export function ECoSImageTransferPanel({ vehicleId, mappedObjectId, title, blob, onBusyChange }: Props) {
  const { t } = useI18n();
  const [open, setOpen] = useState(false);
  const [options, setOptions] = useState<ECoSImageOptions | null>(null);
  const [objectId, setObjectId] = useState("");
  const [description, setDescription] = useState(title.replace(/[<>"'&]/g, "").slice(0, 80));
  const [phase, setPhase] = useState<"idle" | "loading" | "previewing" | "sending">("idle");
  const [preview, setPreview] = useState<ECoSImagePreview | null>(null);
  const [checked, setChecked] = useState(false);
  const [expired, setExpired] = useState(false);
  const [result, setResult] = useState<ECoSImageResult | null>(null);
  const [storedSlot, setStoredSlot] = useState<number>();
  const [error, setError] = useState("");
  const [uncertain, setUncertain] = useState(false);
  const [paused, setPaused] = useState<{ preview: ECoSImagePreview; result: ECoSImageResult } | null>(null);
  const [reviewChanged, setReviewChanged] = useState(false);
  const revision = useRef(0);
  const alive = useRef(true);
  const inFlight = useRef(false);
  const submittedTokens = useRef(new Set<string>());
  const confirmationRef = useRef<HTMLDivElement>(null);
  const prepareRef = useRef<HTMLButtonElement>(null);
  const busy = phase === "sending";

  useEffect(() => {
    alive.current = true;
    return () => { alive.current = false; revision.current++; };
  }, []);
  useEffect(() => {
    if (!preview) return;
    confirmationRef.current?.focus({ preventScroll: true });
    confirmationRef.current?.scrollIntoView?.({ block: "nearest" });
    const timer = window.setTimeout(() => setExpired(true), Math.max(0, Date.parse(preview.expiresAt) - Date.now()));
    return () => window.clearTimeout(timer);
  }, [preview]);

  function invalidate() {
    revision.current++; setPreview(null); setChecked(false); setResult(null); setError("");
    setReviewChanged(false);
  }
  function report(error: unknown) {
    const code = error instanceof ApiError ? error.code : "";
    setError(knownErrors.has(code) ? code : "failed");
  }
  async function loadOptions() {
    setOpen(true); setPhase("loading"); setError("");
    try {
      const value = await api.ecosImageOptions(vehicleId);
      if (!alive.current) return;
      setOptions(value);
      setObjectId(value.locomotives.some(item => String(item.objectId) === mappedObjectId) ? mappedObjectId || "" : "");
    } catch (error) { if (alive.current) report(error); }
    finally { if (alive.current) setPhase("idle"); }
  }
  async function prepare(existingSlot = storedSlot) {
    if (!blob || !objectId || !description.trim() || paused || phase !== "idle" || inFlight.current) return;
    invalidate(); const version = revision.current; setPhase("previewing"); setUncertain(false);
    try {
      const value = existingSlot === undefined
        ? await api.previewECoSImage(vehicleId, Number(objectId), description, blob)
        : await api.previewECoSImage(vehicleId, Number(objectId), description, blob, existingSlot);
      if (alive.current && version === revision.current) { setPreview(value); setExpired(false); }
    } catch (error) { if (alive.current && version === revision.current) report(error); }
    finally { if (alive.current) setPhase("idle"); }
  }
  async function confirm() {
    if (!preview || !checked || expired || phase !== "idle" || inFlight.current || !(Date.parse(preview.expiresAt) > Date.now())) return;
    inFlight.current = true;
    try { await send(preview); }
    finally { inFlight.current = false; }
  }
  async function send(review: ECoSImagePreview) {
    if (submittedTokens.current.has(review.token)) { setError("digital_center_write_grant_conflict"); return; }
    submittedTokens.current.add(review.token);
    setPhase("sending"); onBusyChange(true); setError("");
    setPaused(null); setReviewChanged(false);
    // The same confirmation can never be submitted again, even after a lost reply.
    setPreview(null); setChecked(false);
    let waiting = false;
    try {
      const value = await api.confirmECoSImage(vehicleId, review.token);
      if (alive.current) {
        if (value.status === "control_busy") {
          waiting = true; setPaused({ preview: review, result: value }); setResult(null);
        } else setResult(value);
        if (value.imageVerified) setStoredSlot(value.slot);
      }
    } catch (error) {
      if (alive.current) {
        if (error instanceof ApiError && error.status >= 400 && error.status < 500) report(error);
        else setUncertain(true);
      }
    } finally { if (alive.current) { setPhase("idle"); if (!waiting) onBusyChange(false); } }
  }
  async function resume() {
    if (!paused || !blob || phase !== "idle" || inFlight.current) return;
    inFlight.current = true; setPhase("previewing"); setError("");
    const review = paused.preview;
    try {
      const value = review.existingImage
        ? await api.previewECoSImage(vehicleId, review.target.objectId, review.description, blob, review.slot)
        : await api.previewECoSImage(vehicleId, review.target.objectId, review.description, blob);
      if (!alive.current) return;
      if (submittedTokens.current.has(value.token) || !Number.isFinite(Date.parse(value.expiresAt))) {
        setError("digital_center_write_grant_conflict"); return;
      }
      if (sameECoSImageReview(review, value) && Date.parse(value.expiresAt) > Date.now()) {
        await send(value);
      } else {
        setPaused(null); setPreview(value); setChecked(false); setExpired(false);
        setReviewChanged(true); onBusyChange(false);
      }
    } catch (error) { if (alive.current) report(error); }
    finally { inFlight.current = false; if (alive.current) setPhase("idle"); }
  }
  function cancelPause() {
    if (phase !== "idle" || inFlight.current) return;
    setPaused(null); invalidate(); onBusyChange(false);
  }

  if (!open) return <div className="ecos-image-start">
    <button type="button" className="secondary-button" disabled={!blob} onClick={() => void loadOptions()}>
      <Send size={16} aria-hidden="true" />{t("ecosImage.open")}</button>
  </div>;
  return <><section className="ecos-image-transfer" aria-labelledby="ecos-image-heading" aria-busy={phase !== "idle"}
    inert={Boolean(paused)} aria-hidden={paused ? true : undefined}>
    <h3 id="ecos-image-heading">{t("ecosImage.title")}</h3>
    <p>{t("ecosImage.help")}</p>
    {phase === "loading" && <p role="status">{t("ecosImage.loading")}</p>}
    {error && <p className="locomotive-image-error" role="alert"><TriangleAlert size={18} aria-hidden="true" />
      {t(`ecosImage.error.${error}`)}</p>}
    {!options && phase === "idle" && <button type="button" className="secondary-button" onClick={() => void loadOptions()}>{t("ecosImage.reload")}</button>}
    {options && <>
      <p className="ecos-image-host">ECoS · {options.host}</p>
      {options.locomotives.length === 0 ? <p>{t("ecosImage.empty")}</p> : <div className="ecos-image-fields">
        <label>{t("ecosImage.target")}<AppSelect aria-label={t("ecosImage.target")} value={objectId}
          disabled={phase !== "idle"} onChange={event => { invalidate(); setObjectId(event.target.value); }}>
          <option value="">{t("ecosImage.choose")}</option>
          {options.locomotives.map(item => <option key={item.objectId} value={item.objectId}>
            {item.name} · {item.address} · {item.protocol} · #{item.objectId}</option>)}
        </AppSelect></label>
        <AppTextInput label={t("ecosImage.description")} value={description} maxLength={80} disabled={phase !== "idle"}
          onChange={event => { invalidate(); setDescription(event.target.value); }} />
      </div>}
      {!result && !uncertain && !preview && <button type="button" className="secondary-button" ref={prepareRef}
        disabled={phase !== "idle" || !objectId || !description.trim() || !blob} onClick={() => void prepare()}>
        {t(phase === "previewing" ? "ecosImage.previewing" : storedSlot !== undefined ? "ecosImage.retryAssignment" : "ecosImage.prepare")}</button>}
    </>}
    {preview && <div className="ecos-image-confirm" ref={confirmationRef} tabIndex={-1} role="region"
      aria-label={t("ecosImage.review")}>
      {reviewChanged && <p className="ecos-image-warning" role="status">{t("ecosImage.pause.changed")}</p>}
      <strong>{preview.target.name} · {t("ecosImage.address", { address: preview.target.address })} · {preview.target.protocol} · #{preview.target.objectId}</strong>
      <div className="ecos-image-comparison">
        <figure><figcaption>{t("ecosImage.before")}</figcaption>
          {preview.currentImageMissing ? <p className="ecos-image-warning" role="status">
            <TriangleAlert size={18} aria-hidden="true" /> {t("ecosImage.currentMissing", { slot: preview.current.index })}
          </p> : <img src={preview.currentImage} alt={t("ecosImage.before")} width={190} height={40} />}</figure>
        <figure><figcaption>{t(preview.existingImage ? "ecosImage.existing" : "ecosImage.after", { slot: preview.slot })}</figcaption>
          <img src={preview.newImage} alt={t(preview.existingImage ? "ecosImage.existing" : "ecosImage.after", { slot: preview.slot })} width={190} height={40} /></figure>
      </div>
      <p>{t(preview.existingImage ? "ecosImage.assignHelp" : "ecosImage.confirmHelp", { description: preview.description, slot: preview.slot })}</p>
      <AppCheckbox className="ecos-image-check" checked={checked} onChange={event => setChecked(event.target.checked)}
        label={t("ecosImage.checked")} />
      {expired && <p role="status">{t("ecosImage.expired")}</p>}
      <div className="ecos-image-actions">
        <button type="button" className="secondary-button" onClick={invalidate}>{t("ecosImage.back")}</button>
        <button type="button" className="primary-button" disabled={!checked || expired} onClick={() => void confirm()}>
          <Send size={16} aria-hidden="true" />{t(preview.existingImage ? "ecosImage.assign" : "ecosImage.confirm")}</button>
      </div>
    </div>}
    {busy && <p role="status">{t("ecosImage.sending")}</p>}
    {uncertain && <p className="ecos-image-warning" role="alert">{t("ecosImage.uncertain")}</p>}
    {result && <div className={result.status === "verified" ? "ecos-image-success" : "ecos-image-warning"} role="status">
      <strong>{result.status === "verified" ? <CircleCheck size={18} aria-hidden="true" /> : <TriangleAlert size={18} aria-hidden="true" />}
        {t(`ecosImage.result.${result.status}`, { slot: result.slot, objectId: result.objectId })}</strong>
      {result.status !== "verified" && <p>{t(result.status === "upload_blocked" ? "ecosImage.blockedHelp"
        : result.status === "image_not_stored" ? "ecosImage.notStoredHelp" : "ecosImage.partialHelp")}</p>}
      {result.imageVerified && !result.assignmentVerified && <button type="button" className="secondary-button"
        disabled={phase !== "idle" || !blob} onClick={() => void prepare(result.slot)}>{t("ecosImage.retryAssignment")}</button>}
      {!result.imageVerified && result.status === "upload_blocked" &&
        <button type="button" className="secondary-button" disabled={phase !== "idle" || !blob}
          onClick={() => void prepare()}>{t("ecosImage.prepare")}</button>}
      {!result.auditRecorded && <p>{t("ecosImage.auditMissing")}</p>}
      {result.liveMonitor.wasRunning && !result.liveMonitor.restarted && <p>{t("ecosImage.liveStopped")}</p>}
    </div>}
  </section>{paused && <ECoSImagePauseDialog preview={paused.preview} result={paused.result}
    checking={phase !== "idle"} error={error} returnFocusRef={prepareRef}
    onCancel={cancelPause} onResume={() => void resume()} />}</>;
}
