import { useEffect, useId, useRef, type RefObject } from "react";
import { createPortal } from "react-dom";
import { CirclePause, Play, TriangleAlert } from "lucide-react";
import type { ECoSImagePreview, ECoSImageResult } from "../../shared/ecosImageModel";
import { useI18n } from "../../shared/i18n";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";

export function ECoSImagePauseDialog({ preview, result, checking, error, returnFocusRef, onCancel, onResume }: {
  preview: ECoSImagePreview; result: ECoSImageResult; checking: boolean; error: string;
  returnFocusRef: RefObject<HTMLButtonElement | null>; onCancel: () => void; onResume: () => void;
}) {
  const { t } = useI18n();
  const titleId = useId(), helpId = useId();
  const cancelRef = useRef<HTMLButtonElement>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(() => {
    if (!checking) onCancel();
  }, cancelRef);
  useEffect(() => () => { returnFocusRef.current?.focus(); }, [returnFocusRef]);

  return <>
    <span ref={anchorRef} hidden aria-hidden="true" />
    {createPortal(<div className="locomotive-image-layer ecos-image-pause-layer" ref={layerRef}
      role="dialog" aria-modal="true" aria-labelledby={titleId} aria-describedby={helpId}
      aria-busy={checking} onKeyDown={onKeyDown}>
      <section className="locomotive-image-dialog ecos-image-pause-dialog">
        <header><h2 id={titleId}><CirclePause size={22} aria-hidden="true" />{t("ecosImage.pause.title")}</h2></header>
        <div className="locomotive-image-body">
          <div className="ecos-image-pause-target"><strong>{preview.target.name}</strong>
            <span>{t("ecosImage.address", { address: preview.target.address })} · {preview.target.protocol}
              {" · "}#{preview.target.objectId} · ECoS {preview.host}</span></div>
          <p id={helpId}>{t("ecosImage.pause.help")}</p>
          <figure className="ecos-image-pause-preview"><figcaption>{preview.description}</figcaption>
            <img src={preview.newImage} width={190} height={40} alt={t("ecosImage.pause.image")} /></figure>
          <p>{t("ecosImage.pause.preserved")}</p>
          <p>{t("ecosImage.pause.recheck")}</p>
          {!result.auditRecorded && <p className="ecos-image-warning">{t("ecosImage.auditMissing")}</p>}
          {result.liveMonitor.wasRunning && !result.liveMonitor.restarted &&
            <p className="ecos-image-warning">{t("ecosImage.liveStopped")}</p>}
          {error && <p className="locomotive-image-error" role="alert"><TriangleAlert size={18} aria-hidden="true" />
            {t(`ecosImage.error.${error}`)}</p>}
          {checking && <p role="status">{t("ecosImage.previewing")}</p>}
        </div>
        <footer><button type="button" className="secondary-button" ref={cancelRef} disabled={checking}
          onClick={onCancel}>{t("ecosImage.pause.cancel")}</button>
          <button type="button" className="primary-button" disabled={checking} onClick={onResume}>
            <Play size={16} aria-hidden="true" />{t("ecosImage.pause.resume")}</button></footer>
      </section>
    </div>, document.body)}
  </>;
}
