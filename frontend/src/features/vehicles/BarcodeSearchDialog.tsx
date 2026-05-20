import { FormEvent, useCallback, useEffect, useRef, useState } from "react";
import { Barcode, Camera, PackageSearch, X } from "lucide-react";
import { useI18n } from "../../shared/i18n";

type BarcodeSearchDialogProps = {
  value: string;
  onValueChange: (value: string) => void;
  onClose: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

type ScannerControls = {
  stop: () => void;
};

type BarcodeReader = {
  decodeFromConstraints: (
    constraints: MediaStreamConstraints,
    previewElem: HTMLVideoElement,
    callbackFn: (result: { getText: () => string } | undefined, error: unknown, controls: ScannerControls) => void
  ) => Promise<ScannerControls>;
};

export function BarcodeSearchDialog({ value, onValueChange, onClose, onSubmit }: BarcodeSearchDialogProps) {
  const { t } = useI18n();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const scannerControlsRef = useRef<ScannerControls | null>(null);
  const readerRef = useRef<BarcodeReader | null>(null);
  const scannerActiveRef = useRef(false);
  const [scannerOpen, setScannerOpen] = useState(false);
  const [scannerMessage, setScannerMessage] = useState("");

  useEffect(() => {
    inputRef.current?.focus();
    inputRef.current?.select();
  }, []);

  const stopCameraScan = useCallback(() => {
    scannerActiveRef.current = false;
    scannerControlsRef.current?.stop();
    scannerControlsRef.current = null;
    if (videoRef.current) {
      videoRef.current.srcObject = null;
    }
    setScannerOpen(false);
  }, []);

  useEffect(() => () => stopCameraScan(), [stopCameraScan]);

  const startCameraScan = async () => {
    if (!window.isSecureContext) {
      setScannerMessage(t("vehicles.barcode.cameraSecureContext"));
      setScannerOpen(false);
      return;
    }

    if (!navigator.mediaDevices?.getUserMedia) {
      setScannerMessage(t("vehicles.barcode.cameraUnsupported"));
      setScannerOpen(false);
      return;
    }

    stopCameraScan();
    setScannerMessage(t("vehicles.barcode.cameraStarting"));
    setScannerOpen(true);

    try {
      await new Promise((resolve) => window.setTimeout(resolve, 0));

      const video = videoRef.current;
      if (!video) {
        setScannerMessage(t("vehicles.barcode.cameraUnavailable"));
        return;
      }

      const reader = readerRef.current ?? new (await import("@zxing/browser")).BrowserMultiFormatReader();
      readerRef.current = reader;
      scannerActiveRef.current = true;
      setScannerMessage(t("vehicles.barcode.cameraReady"));

      const controls = await reader.decodeFromConstraints(
        {
          audio: false,
          video: {
            facingMode: { ideal: "environment" },
            height: { ideal: 720 },
            width: { ideal: 1280 }
          }
        },
        video,
        (result, _error, controlsFromCallback) => {
          if (!scannerActiveRef.current || !result) {
            return;
          }

          const ean = result.getText().replace(/[^\d]/g, "");
          if (ean.length < 8) {
            setScannerMessage(t("vehicles.barcode.cameraDetecting"));
            return;
          }

          onValueChange(ean);
          setScannerMessage(t("vehicles.barcode.cameraDetected"));
          scannerActiveRef.current = false;
          controlsFromCallback.stop();
          scannerControlsRef.current = null;
          setScannerOpen(false);
          inputRef.current?.focus();
          inputRef.current?.select();
        }
      );
      scannerControlsRef.current = controls;
    } catch {
      scannerActiveRef.current = false;
      setScannerMessage(t("vehicles.barcode.cameraPermission"));
      setScannerOpen(false);
    }
  };

  const handleClose = () => {
    stopCameraScan();
    onClose();
  };

  return (
    <div className="confirm-layer barcode-search-layer" role="dialog" aria-modal="true" aria-label={t("vehicles.barcode.title")}>
      <form className="barcode-search-dialog" onSubmit={onSubmit}>
        <header className="panel-head form-head">
          <div>
            <h2>{t("vehicles.barcode.title")}</h2>
            <p>{t("vehicles.barcode.help")}</p>
          </div>
          <button type="button" className="icon-button" onClick={handleClose} aria-label={t("vehicles.close")} title={t("vehicles.close")}>
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

        {scannerOpen && (
          <section className="barcode-camera-panel">
            <video ref={videoRef} muted playsInline />
            <div>
              <strong>{t("vehicles.barcode.cameraTitle")}</strong>
              <p>{scannerMessage}</p>
            </div>
            <button type="button" className="secondary-button" onClick={stopCameraScan}>
              {t("vehicles.barcode.cameraClose")}
            </button>
          </section>
        )}

        {!scannerOpen && scannerMessage && <p className="barcode-scan-message">{scannerMessage}</p>}

        <p className="barcode-hint">
          {t("vehicles.barcode.hint")}
        </p>

        <footer className="barcode-search-actions">
          <button type="button" className="secondary-button" onClick={handleClose}>
            {t("vehicles.cancel")}
          </button>
          <button type="submit" className="primary-button">
            <PackageSearch size={15} aria-hidden="true" />
            {t("vehicles.articleSearch.search")}
          </button>
        </footer>
      </form>
    </div>
  );
}
