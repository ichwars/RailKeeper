import { Bug, ExternalLink, RotateCw, Send, X } from "lucide-react";
import { useEffect, useMemo, useRef, useState, type FormEvent } from "react";
import { createPortal } from "react-dom";

import type { BugReportClientDiagnostics } from "../../shared/apiBugReports";
import { readDiagnosticPreference, saveDiagnosticPreference, observeDiagnosticPreference } from "./bugReportPreferences";
import { useI18n } from "../../shared/i18n";
import type {
  BugReportArea,
  BugReportCapability,
  BugReportSubmissionInput,
  BugReportSubmissionResult
} from "../../shared/apiBugReports";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppCheckbox } from "../../shared/ui/AppCheckbox";
import { AppTextArea } from "../../shared/ui/AppTextArea";
import { AppTextInput } from "../../shared/ui/AppTextInput";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import { areaForPathname, createBugReportID, validateBugReportForm, type BugReportFormErrors } from "./bugReportModel";
import { clientErrorBuffer } from "./clientErrorBuffer";
import {
  collectBugReportDiagnostics,
  collectBugReportDiagnosticsWithClientHints
} from "./bugReportDiagnostics";
import { BugReportScreenshotField } from "./BugReportScreenshotField";

const areas: BugReportArea[] = [
  "overview", "vehicles", "accessories", "exhibition",
  "importExport", "digitalCenters", "settings", "other"
];

export function BugReportDialog({
  capability,
  initialArea,
  roles,
  onClose,
  onLoadDiagnostics,
  onSubmit
}: {
  capability: BugReportCapability;
  initialArea?: BugReportArea;
  roles: string[];
  onClose: () => void;
  onLoadDiagnostics?: () => Promise<BugReportCapability>;
  onSubmit: (input: BugReportSubmissionInput) => Promise<BugReportSubmissionResult>;
}) {
  const { t, language } = useI18n();
  const [reportID] = useState(createBugReportID);
  const [area, setArea] = useState<BugReportArea>(initialArea || areaForPathname(window.location.pathname));
  const [description, setDescription] = useState("");
  const [expectedBehavior, setExpectedBehavior] = useState("");
  const [reproductionSteps, setReproductionSteps] = useState("");
  const [email, setEmail] = useState("");
  const [screenshot, setScreenshot] = useState<File | null>(null);
  const [screenshotConsent, setScreenshotConsent] = useState(false);
  const [errors, setErrors] = useState<BugReportFormErrors>({});
  const [busy, setBusy] = useState(false);
  const [submissionError, setSubmissionError] = useState("");
  const [result, setResult] = useState<BugReportSubmissionResult | null>(null);
  const descriptionRef = useRef<HTMLTextAreaElement | null>(null);
  const emailRef = useRef<HTMLInputElement | null>(null);
  const closeRef = useRef<HTMLButtonElement | null>(null);
  const [includeDiagnostics, setIncludeDiagnostics] = useState(readDiagnosticPreference);
  const [serverDiagnostics, setServerDiagnostics] = useState(capability.serverDiagnostics);
  const [diagnosticError, setDiagnosticError] = useState(false);
  const [clientDiagnostics, setClientDiagnostics] = useState<BugReportClientDiagnostics | null>(() =>
    includeDiagnostics ? collectBugReportDiagnostics({ area, roles, language, buffer: clientErrorBuffer }) : null
  );
  useEffect(() => observeDiagnosticPreference(setIncludeDiagnostics), []);
  useEffect(() => {
    if (!includeDiagnostics) {
      setClientDiagnostics(null);
      setDiagnosticError(false);
      return;
    }
    let cancelled = false;
    setDiagnosticError(false);
    const input = { area, roles, language, buffer: clientErrorBuffer };
    setClientDiagnostics(collectBugReportDiagnostics(input));
    void collectBugReportDiagnosticsWithClientHints(input).then((diagnostics) => {
      if (!cancelled) setClientDiagnostics(diagnostics);
    });
    if (onLoadDiagnostics) {
      void onLoadDiagnostics().then((loaded) => {
        if (!cancelled) setServerDiagnostics(loaded.serverDiagnostics);
      }).catch(() => {
        if (!cancelled) setDiagnosticError(true);
      });
    }
    return () => { cancelled = true; };
  }, [area, includeDiagnostics, language, roles, onLoadDiagnostics]);
  const diagnosticPreview = useMemo(
    () => includeDiagnostics ? JSON.stringify({ server: serverDiagnostics, browser: clientDiagnostics }, null, 2) : "",
    [includeDiagnostics, serverDiagnostics, clientDiagnostics]
  );

  const requestClose = () => {
    if (!busy) onClose();
  };
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(requestClose, descriptionRef);

  async function submit(event?: FormEvent) {
    event?.preventDefault();
    const formErrors = validateBugReportForm({
      area, description, expectedBehavior, reproductionSteps, email, screenshot, screenshotConsent
    });
    setErrors(formErrors);
    setSubmissionError("");
    if (Object.keys(formErrors).length > 0) {
      if (formErrors.description) descriptionRef.current?.focus();
      else if (formErrors.email) emailRef.current?.focus();
      return;
    }
    setBusy(true);
    try {
      const submitted = await onSubmit({
        reportId: reportID,
        area,
        description: description.trim(),
        expectedBehavior: expectedBehavior.trim(),
        reproductionSteps: reproductionSteps.trim(),
        email: email.trim(),
        includeDiagnostics,
        clientDiagnostics: includeDiagnostics ? clientDiagnostics : null,
        screenshotConsent,
        screenshot
      });
      setResult(submitted);
    } catch {
      setSubmissionError(t("bugReport.error.submit"));
    } finally {
      setBusy(false);
    }
  }

  const dialog = (
    <div
      ref={layerRef}
      className="confirm-layer bug-report-layer"
      role="dialog"
      aria-modal="true"
      aria-labelledby="bug-report-title"
      onKeyDown={onKeyDown}
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) requestClose();
      }}
    >
      <section className="panel bug-report-dialog">
        <header className="bug-report-dialog-head">
          <div className="bug-report-dialog-title">
            <Bug size={20} aria-hidden="true" />
            <h2 id="bug-report-title">{t("bugReport.title")}</h2>
          </div>
          <button ref={closeRef} type="button" className="icon-button" aria-label={t("bugReport.close")}
            disabled={busy} onClick={requestClose}>
            <X size={19} aria-hidden="true" />
          </button>
        </header>

        <form className="bug-report-form" onSubmit={submit}>
          <label className="app-field bug-report-area-field">
            <span className="app-field-label">{t("bugReport.area")} <span aria-hidden="true">*</span></span>
            <AppSelect aria-label={t("bugReport.area")} value={area} disabled={busy}
              onChange={(event) => setArea(event.target.value as BugReportArea)}>
              {areas.map((entry) => <option key={entry} value={entry}>{t(`bugReport.area.${entry}`)}</option>)}
            </AppSelect>
          </label>
          <AppTextArea
            ref={descriptionRef}
            className="bug-report-description"
            label={t("bugReport.description")}
            placeholder={t("bugReport.description.placeholder")}
            value={description}
            rows={4}
            maxLength={10_000}
            required
            disabled={busy}
            error={errors.description ? t("bugReport.error.description") : undefined}
            onChange={(event) => {
              setDescription(event.target.value);
              if (errors.description) setErrors((current) => ({ ...current, description: undefined }));
            }}
          />
          <AppTextArea
            className="bug-report-expected"
            label={t("bugReport.expected")}
            value={expectedBehavior}
            rows={2}
            maxLength={5_000}
            disabled={busy}
            error={errors.expectedBehavior ? t("bugReport.error.expected") : undefined}
            onChange={(event) => setExpectedBehavior(event.target.value)}
          />
          <AppTextArea
            className="bug-report-reproduction"
            label={t("bugReport.reproduction")}
            value={reproductionSteps}
            rows={3}
            maxLength={10_000}
            disabled={busy}
            error={errors.reproductionSteps ? t("bugReport.error.reproduction") : undefined}
            onChange={(event) => setReproductionSteps(event.target.value)}
          />
          {capability.contactAvailable ? (
            <AppTextInput
              ref={emailRef}
              label={t("bugReport.email")}
              value={email}
              inputMode="email"
              autoComplete="email"
              maxLength={254}
              disabled={busy}
              helpText={t("bugReport.email.help")}
              error={errors.email ? t("bugReport.error.email") : undefined}
              onChange={(event) => setEmail(event.target.value)}
            />
          ) : null}
          {capability.screenshotAvailable !== false ? <BugReportScreenshotField
            file={screenshot}
            onFileChange={setScreenshot}
            consent={screenshotConsent}
            onConsentChange={setScreenshotConsent}
            maxBytes={capability.maxScreenshotBytes}
            acceptedTypes={capability.acceptedScreenshotTypes}
            disabled={busy}
            error={errors.screenshotConsent ? t("bugReport.error.screenshotConsent") : undefined}
          /> : null}

          <AppCheckbox className="bug-report-diagnostic-toggle" checked={includeDiagnostics} disabled={busy}
            label={<><strong>{t("bugReport.diagnostics.include")}</strong>
              <span className="bug-report-diagnostic-help">{t("bugReport.diagnostics.help")}</span></>}
              onChange={(event) => {
                const enabled = event.target.checked;
                setIncludeDiagnostics(enabled);
                saveDiagnosticPreference(enabled);
              }} />
          {!includeDiagnostics ? <p className="bug-report-diagnostic-disabled" role="status">{t("bugReport.diagnostics.disabled")}</p> : null}
          {includeDiagnostics && diagnosticError ? <p role="alert">{t("bugReport.diagnostics.error")}</p> : null}
          {includeDiagnostics ? <details className="bug-report-disclosure">
            <summary>{t("bugReport.data.summary")}</summary>
            <div>
              <strong>{t("bugReport.data.included")}</strong>
              <p>{t("bugReport.data.includedText")}</p>
              <pre aria-label={t("bugReport.data.values")}>{diagnosticPreview}</pre>
              <strong>{t("bugReport.data.excluded")}</strong>
              <p>{t("bugReport.data.inventoryExcluded")}</p>
              <p>{t("bugReport.data.devicesExcluded")}</p>
              <p>{t("bugReport.data.secretsExcluded")}</p>
            </div>
          </details> : null}

          <p className="bug-report-security-notice">
            {t("bugReport.security.notice")} {" "}
            <a href="https://github.com/ichwars/RailKeeper/security/advisories/new" target="_blank" rel="noreferrer">
              {t("bugReport.security.link")}
            </a>
          </p>

          {submissionError ? <p className="form-message error" role="alert">{submissionError}</p> : null}
          {result ? (
            <section className={`bug-report-result ${result.status}`} aria-live="polite">
              <strong>{result.status === "complete" ? t("bugReport.success") : t("bugReport.partial")}</strong>
              {result.warning ? <p>{result.warning}</p> : null}
              <a href={result.issueUrl} target="_blank" rel="noreferrer">
                {t("bugReport.issue", { number: result.issueNumber })}<ExternalLink size={15} aria-hidden="true" />
              </a>
            </section>
          ) : null}

          <footer className="bug-report-actions">
            <button type="button" className="secondary-button" disabled={busy} onClick={requestClose}>
              {t("bugReport.cancel")}
            </button>
            {result?.status === "partial" ? (
              <button type="button" className="primary-button" disabled={busy} onClick={() => void submit()}>
                <RotateCw size={16} aria-hidden="true" />{t("bugReport.retry")}
              </button>
            ) : result?.status !== "complete" ? (
              <button type="submit" className="primary-button" disabled={busy}>
                {busy ? <RotateCw size={16} aria-hidden="true" /> : <Send size={16} aria-hidden="true" />}
                {busy ? t("bugReport.sending") : t("bugReport.send")}
              </button>
            ) : null}
          </footer>
        </form>
      </section>
    </div>
  );

  return <><span ref={anchorRef} aria-hidden="true" />{createPortal(dialog, document.body)}</>;
}
