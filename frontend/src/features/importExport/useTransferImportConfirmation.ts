import { useEffect, useRef, useState } from "react";
import { ApiError } from "../../shared/api";
import type { Language } from "../../shared/i18n";
import type { DataTransferJob, DataTransferJobDetails } from "./dataTransferModel";

type Outcome = "idle" | "unknown" | "completed" | "stopped";
type Options = {
  job: DataTransferJob | null; language: Language;
  confirm: (id: string, revision: number) => Promise<unknown>;
  refresh: (id: string) => Promise<DataTransferJobDetails>;
  recovered: (details: DataTransferJobDetails, reupload: boolean) => void;
  close: () => void; setBusy: (busy: boolean) => void; setError: (error: string) => void;
};

export function useTransferImportConfirmation(options: Options) {
  const opts = useRef(options); opts.current = options;
  const [outcome, setOutcome] = useState<Outcome>("idle");
  const busy = useRef(false), mounted = useRef(true);
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);
  useEffect(() => { setOutcome("idle"); }, [options.job?.id]);
  const active = (id: string) => mounted.current && opts.current.job?.id === id;
  const readOutcome = async (id: string) => {
    setOutcome("unknown");
    try {
      const details = await opts.current.refresh(id);
      if (!active(id)) return;
      if (details.job.id !== id) throw new Error("Unexpected import job");
      const state = details.job.state;
      const reupload = ["draft", "review_required", "ready"].includes(state);
      opts.current.recovered(details, reupload);
      setOutcome(state === "completed" ? "completed" : reupload ? "idle" : state === "running" ? "unknown" : "stopped");
    } catch {
      if (active(id)) setOutcome("unknown");
    }
  };
  const run = async (confirm: boolean) => {
    const job = opts.current.job;
    if (!job || busy.current || (confirm && (outcome !== "idle" || job.state !== "ready"))) return;
    busy.current = true; opts.current.setBusy(true); opts.current.setError("");
    try {
      if (!confirm) { await readOutcome(job.id); return; }
      try {
        await opts.current.confirm(job.id, job.revision);
        if (active(job.id)) opts.current.close();
      } catch (error) {
        if (!active(job.id)) return;
        if (!(error instanceof ApiError) || error.status === 409 || error.status >= 500 || [408, 499].includes(error.status)) {
          await readOutcome(job.id);
        } else opts.current.setError(error.message);
      }
    } finally {
      busy.current = false;
      if (active(job.id)) opts.current.setBusy(false);
    }
  };
  const de = options.language === "de";
  const message = outcome === "completed"
    ? (de ? "Dieser Importauftrag ist bereits abgeschlossen. Es wird nichts erneut übernommen. Prüfe das Ergebnis im Auftragsverlauf."
      : "This import job is already completed. Nothing will be adopted again. Review its result in job history.")
    : outcome === "stopped"
      ? (de ? "Dieser Importauftrag kann nicht mehr bestätigt werden. Prüfe seinen Status im Auftragsverlauf."
        : "This import job can no longer be confirmed. Review its status in job history.")
      : (de ? "Der Abschluss ist noch unklar. Prüfe den Auftragsstatus, bevor du einen weiteren Import beginnst."
        : "The outcome is still uncertain. Check the job status before starting another import.");
  return { outcome, blocked: outcome !== "idle", message,
    checkLabel: de ? "Auftragsstatus prüfen" : "Check job status",
    confirm: () => run(true), check: () => run(false) };
}
