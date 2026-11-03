import type { BugReportArea } from "../../shared/apiBugReports";

export function createBugReportID(
  source: Pick<Crypto, "getRandomValues"> & Partial<Pick<Crypto, "randomUUID">> = crypto
): string {
  if (source.randomUUID) return source.randomUUID();
  const bytes = source.getRandomValues(new Uint8Array(16));
  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  const hex = Array.from(bytes, (value) => value.toString(16).padStart(2, "0")).join("");
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

export type BugReportFormValues = {
  area: BugReportArea;
  description: string;
  expectedBehavior: string;
  reproductionSteps: string;
  email: string;
  screenshot: File | null;
  screenshotConsent: boolean;
};

export type BugReportFormErrors = Partial<Record<keyof BugReportFormValues, string>>;

const routeAreas: Array<[string, BugReportArea]> = [
  ["/overview", "overview"],
  ["/vehicles", "vehicles"],
  ["/accessories", "accessories"],
  ["/exhibition", "exhibition"],
  ["/import-export", "importExport"],
  ["/digital-centers", "digitalCenters"],
  ["/settings", "settings"]
];

export function areaForPathname(pathname: string): BugReportArea {
  const normalized = pathname.split(/[?#]/, 1)[0];
  return routeAreas.find(([prefix]) => normalized === prefix || normalized.startsWith(`${prefix}/`))?.[1] || "other";
}

export function validateBugReportForm(values: BugReportFormValues): BugReportFormErrors {
  const errors: BugReportFormErrors = {};
  const descriptionLength = [...values.description.trim()].length;
  if (descriptionLength < 20 || descriptionLength > 10_000) errors.description = "description_length";
  if ([...values.expectedBehavior.trim()].length > 5_000) errors.expectedBehavior = "expected_behavior_length";
  if ([...values.reproductionSteps.trim()].length > 10_000) errors.reproductionSteps = "reproduction_length";
  if (values.email.trim() && (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(values.email.trim()) || values.email.trim().length > 254)) {
    errors.email = "email_invalid";
  }
  if (values.screenshot && !values.screenshotConsent) errors.screenshotConsent = "screenshot_consent_required";
  return errors;
}
