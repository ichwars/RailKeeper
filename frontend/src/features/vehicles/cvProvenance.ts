// Exported wall-clock time has no timezone. Preserve it without a local/UTC conversion.
export function validCVExportTimestamp(value?: string) {
  if (!value) return true;
  if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$/.test(value) || value.startsWith("0000")) return false;
  const date = new Date(value + "Z");
  return Number.isFinite(date.getTime()) && date.toISOString().slice(0, 19) === value;
}

export function formatCVExportTimestamp(value: string, language: string) {
  if (!value || !validCVExportTimestamp(value)) return "—";
  const [date, time] = value.split("T"), [year, month, day] = date.split("-");
  return `${language === "de" ? `${day}.${month}.${year}` : `${month}/${day}/${year}`} ${time}`;
}
