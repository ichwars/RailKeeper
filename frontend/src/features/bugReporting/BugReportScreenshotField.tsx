import { useEffect, useState, type ClipboardEvent, type DragEvent } from "react";

import { useI18n } from "../../shared/i18n";
import { AppCheckbox } from "../../shared/ui/AppCheckbox";
import { AppFilePicker } from "../../shared/ui/AppFilePicker";

export function BugReportScreenshotField({
  file,
  onFileChange,
  consent,
  onConsentChange,
  maxBytes,
  acceptedTypes,
  error,
  disabled = false
}: {
  file: File | null;
  onFileChange: (file: File | null) => void;
  consent: boolean;
  onConsentChange: (consent: boolean) => void;
  maxBytes: number;
  acceptedTypes: string[];
  error?: string;
  disabled?: boolean;
}) {
  const { t } = useI18n();
  const [fileError, setFileError] = useState("");
  const [previewURL, setPreviewURL] = useState("");

  const formattedSize = file ? file.size < 1024
    ? `${file.size} B`
    : `${(file.size / 1024).toFixed(file.size < 10 * 1024 ? 1 : 0)} KB` : "";

  useEffect(() => {
    if (!file) {
      setPreviewURL("");
      return;
    }
    const url = URL.createObjectURL(file);
    setPreviewURL(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);

  function choose(candidate: File | null) {
    setFileError("");
    if (!candidate) {
      onFileChange(null);
      onConsentChange(false);
      return;
    }
    if (!acceptedTypes.includes(candidate.type)) {
      setFileError(t("bugReport.screenshot.typeError"));
      return;
    }
    if (candidate.size > maxBytes) {
      setFileError(t("bugReport.screenshot.sizeError"));
      return;
    }
    onFileChange(candidate);
  }

  function dropped(event: DragEvent<HTMLDivElement>) {
    event.preventDefault();
    if (!disabled) choose(event.dataTransfer.files?.[0] || null);
  }

  function pasted(event: ClipboardEvent<HTMLDivElement>) {
    if (disabled) return;
    const pastedFile = event.clipboardData.files?.[0];
    if (!pastedFile) return;
    event.preventDefault();
    choose(pastedFile);
  }

  return (
    <div
      className="bug-report-screenshot-field"
      role="group"
      aria-label={t("bugReport.screenshot.label")}
      tabIndex={disabled ? -1 : 0}
      onDragOver={(event) => event.preventDefault()}
      onDrop={dropped}
      onPaste={pasted}
    >
      <AppFilePicker
        label={t("bugReport.screenshot.label")}
        helpText={t("bugReport.screenshot.help")}
        error={fileError || error}
        file={file}
        onFileChange={choose}
        accept={acceptedTypes.join(",")}
        capture="environment"
        disabled={disabled}
        triggerLabel={t("bugReport.screenshot.choose")}
        clearLabel={t("bugReport.screenshot.remove")}
        emptyLabel={t("bugReport.screenshot.empty")}
      />
      {previewURL ? (
        <img className="bug-report-screenshot-preview" src={previewURL} alt={t("bugReport.screenshot.preview")} />
      ) : null}
      {file ? <p className="bug-report-screenshot-size">{t("bugReport.screenshot.size", { size: formattedSize })}</p> : null}
      {file ? (
        <AppCheckbox
          checked={consent}
          disabled={disabled}
          aria-invalid={Boolean(error) || undefined}
          label={t("bugReport.screenshot.consent")}
          onChange={(event) => onConsentChange(event.target.checked)}
        />
      ) : null}
    </div>
  );
}
