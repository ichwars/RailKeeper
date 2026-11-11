import { FormEvent, useCallback, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Barcode, Camera, PackageSearch, X } from "lucide-react";
import { useI18n } from "../i18n";
import { useModalDialogLayer } from "../ui/useModalDialogLayer";

type BarcodeSearchDialogProps = {
  value: string;
  onValueChange: (value: string) => void;
  onClose: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

type ScannerControls = {
  stop: () => void;
};

export function BarcodeSearchDialog({ value, onValueChange, onClose, onSubmit }: BarcodeSearchDialogProps) {
  const { t } = useI18n();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(onClose, inputRef);
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const scannerControlsRef = useRef<ScannerControls | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const scanSequence = useRef(0);
  const [openScan, setOpenScan] = useState<number | null>(null);
  const [scannerMessage, setScannerMessage] = useState("");

  useEffect(() => {
    inputRef.current?.focus();
    inputRef.current?.select();
  }, []);

  const stopCameraScan = useCallback(() => {
    scanSequence.current += 1;
    const controls = scannerControlsRef.current;
    if (controls) controls.stop();
    else streamRef.current?.getTracks().forEach(track => track.stop());
    scannerControlsRef.current = null;
    streamRef.current = null;
    if (videoRef.current) {
      videoRef.current.srcObject = null;
    }
    setOpenScan(null);
  }, []);

  useEffect(() => () => stopCameraScan(), [stopCameraScan]);

  const startCameraScan = async () => {
    if (!window.isSecureContext) {
      setScannerMessage(t("vehicles.barcode.cameraSecureContext"));
      return;
    }

    if (!navigator.mediaDevices?.getUserMedia) {
      setScannerMessage(t("vehicles.barcode.cameraUnsupported"));
      return;
    }

    stopCameraScan();
    const scan = scanSequence.current;
    setScannerMessage(t("vehicles.barcode.cameraStarting"));
    setOpenScan(scan);

    try {
      await new Promise((resolve) => window.setTimeout(resolve, 0));
      if (scan !== scanSequence.current) return;

      const video = videoRef.current;
      if (!video) {
        setScannerMessage(t("vehicles.barcode.cameraUnavailable"));
        stopCameraScan();
        return;
      }

      const { BrowserMultiFormatReader } = await import("@zxing/browser");
      if (scan !== scanSequence.current) return;
      const stream = await navigator.mediaDevices.getUserMedia({
        audio: false,
        video: { facingMode: { ideal: "environment" }, height: { ideal: 720 }, width: { ideal: 1280 } }
      });
      if (scan !== scanSequence.current) {
        stream.getTracks().forEach(track => track.stop());
        return;
      }
      streamRef.current = stream;
      setScannerMessage(t("vehicles.barcode.cameraReady"));

      let detected = false;
      const controls = await new BrowserMultiFormatReader().decodeFromStream(
        stream,
        video,
        (result, _error, controlsFromCallback) => {
          if (scan !== scanSequence.current || !result) {
            return;
          }

          const ean = result.getText().replace(/[^\d]/g, "");
          if (ean.length < 8) {
            setScannerMessage(t("vehicles.barcode.cameraDetecting"));
            return;
          }

          detected = true;
          scannerControlsRef.current = controlsFromCallback;
          stopCameraScan();
          onValueChange(ean);
          setScannerMessage(t("vehicles.barcode.cameraDetected"));
          inputRef.current?.focus();
          inputRef.current?.select();
        }
      );
      if (scan !== scanSequence.current) {
        if (!detected) controls.stop();
        return;
      }
      scannerControlsRef.current = controls;
    } catch {
      if (scan !== scanSequence.current) return;
      stopCameraScan();
      setScannerMessage(t("vehicles.barcode.cameraPermission"));
    }
  };

  const handleClose = () => {
    stopCameraScan();
    onClose();
  };

  const dialog = (
    <div ref={layerRef} className="confirm-layer barcode-search-layer" role="dialog" aria-modal="true"
      aria-label={t("vehicles.barcode.title")} onKeyDown={onKeyDown}>
      <form className="barcode-search-dialog" onSubmit={onSubmit}>
        <header className="panel-head form-head">
          <div>
            <h2>{t("vehicles.barcode.title")}</h2>
            <p>{t("vehicles.barcode.help")}</p>
          </div>
          <button type="button" className="icon-button" onClick={handleClose}
            aria-label={t("common.cancel")} title={t("common.cancel")}>
            <X size={17} />
          </button>
        </header>

        <label className="barcode-input-label">
          Barcode / EAN
          <span className="barcode-input-row">
            <span className="barcode-input-shell">
              <Barcode size={18} aria-hidden="true" />
              <input
                ref={inputRef}
                value={value}
                onChange={(event) => onValueChange(event.target.value)}
                inputMode="numeric"
                autoComplete="off"
                placeholder={t("vehicles.barcode.placeholder")}
              />
            </span>
            <button type="button" className="secondary-button barcode-camera-button" onClick={startCameraScan}>
              <Camera size={15} aria-hidden="true" />
              {t("vehicles.barcode.camera")}
            </button>
          </span>
        </label>

        {openScan !== null && (
          <section className="barcode-camera-panel">
            <video key={openScan} ref={videoRef} muted playsInline />
            <div>
              <strong>{t("vehicles.barcode.cameraTitle")}</strong>
              <p>{scannerMessage}</p>
            </div>
            <button type="button" className="secondary-button" onClick={stopCameraScan}>
              {t("vehicles.barcode.cameraClose")}
            </button>
          </section>
        )}

        {openScan === null && scannerMessage && <p className="barcode-scan-message">{scannerMessage}</p>}

        <p className="barcode-hint">
          {t("vehicles.barcode.hint")}
        </p>

        <footer className="barcode-search-actions">
          <button type="button" className="secondary-button" onClick={handleClose}>
            {t("common.cancel")}
          </button>
          <button type="submit" className="primary-button">
            <PackageSearch size={15} aria-hidden="true" />
            {t("vehicles.articleSearch.search")}
          </button>
        </footer>
      </form>
    </div>
  );
  return <>
    <span ref={anchorRef} hidden aria-hidden="true" />
    {createPortal(dialog, document.body)}
  </>;
}
