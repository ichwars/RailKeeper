import { Database, Folder, History, RefreshCw } from "lucide-react";

import { formatDateTime, type Language } from "../../shared/i18n";
import type { DataTransferSummary } from "./dataTransferModel";

type Translate = (key: string, values?: Record<string, string | number>) => string;

type TransferSummaryStripProps = {
  language: Language;
  summary: DataTransferSummary;
  t: Translate;
};

export function TransferSummaryStrip({ language, summary, t }: TransferSummaryStripProps) {
  const metrics = [
    {
      icon: RefreshCw,
      value: formatNumber(summary.openJobs, language),
      label: t("importExport.dashboard.summary.openJobs"),
      labelFirst: false
    },
    {
      icon: Database,
      value: formatNumber(summary.selectedRecords, language),
      label: t("importExport.dashboard.summary.records"),
      labelFirst: false
    },
    {
      icon: History,
      value: summary.lastExportAt ? formatTransferDate(summary.lastExportAt, language, t) : "–",
      label: t("importExport.dashboard.summary.lastExport"),
      labelFirst: true
    },
    {
      icon: Folder,
      value: t("importExport.dashboard.summary.local"),
      label: t("importExport.dashboard.summary.storage"),
      labelFirst: true
    }
  ];

  return (
    <section className="transfer-summary-strip" aria-label={t("importExport.dashboard.summary.label")}>
      {metrics.map(({ icon: Icon, value, label, labelFirst }) => (
        <div className="transfer-summary-metric" key={label}>
          <Icon size={23} strokeWidth={1.8} aria-hidden="true" />
          <span>
            {labelFirst ? <><small>{label}</small><strong>{value}</strong></> :
              <><strong>{value}</strong><small>{label}</small></>}
          </span>
        </div>
      ))}
    </section>
  );
}

function formatNumber(value: number, language: Language) {
  return new Intl.NumberFormat(language === "de" ? "de-DE" : "en-GB").format(value);
}

function formatTransferDate(value: string, language: Language, t: Translate) {
  const date = new Date(value);
  const now = new Date();
  if (date.toDateString() !== now.toDateString()) return formatDateTime(date, language);

  const time = new Intl.DateTimeFormat(language === "de" ? "de-DE" : "en-GB", {
    hour: "2-digit",
    minute: "2-digit"
  }).format(date);
  return t("importExport.dashboard.summary.today", { time });
}
