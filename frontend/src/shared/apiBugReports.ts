export type BugReportArea =
  | "overview"
  | "vehicles"
  | "accessories"
  | "exhibition"
  | "importExport"
  | "digitalCenters"
  | "settings"
  | "other";

export type BugReportClientError = {
  timestamp: string;
  category: "window" | "promise" | "api" | "network" | "timeout";
  errorClass: string;
  problemCode: string;
  message: string;
  method: string;
  route: string;
  status?: number;
};

export type BugReportClientDiagnostics = {
  browserFamily: string;
  browserMajor: string;
  osFamily: string;
  osMajor: string;
  viewportWidth: number;
  viewportHeight: number;
  language: string;
  timeZone: string;
  theme: "light" | "dark";
  pathname: string;
  area: BugReportArea;
  roles: string[];
  errors: BugReportClientError[];
};

export type BugReportServerDiagnostics = {
  collectedAt?: string;
  uptimeSeconds?: number;
  cpuCount?: number;
  goroutines?: number;
  heapBytes?: number;
  runtimeBytes?: number;
  sqliteVersion?: string;
  databaseBytes?: number;
  journalMode?: string;
  railkeeperVersion: string;
  deploymentType: "windows-portable" | "container" | "native";
  goVersion: string;
  operatingSystem: string;
  architecture: string;
  schemaVersion: number;
  enabledDigitalProviderTypes: string[];
};

export type BugReportCapability = {
  enabled: boolean;
  screenshotAvailable?: boolean;
  fallbackUrl: string;
  maxScreenshotBytes: number;
  acceptedScreenshotTypes: string[];
  includedDiagnosticCategories: string[];
  excludedDiagnosticCategories: string[];
  contactAvailable: boolean;
  serverDiagnostics?: BugReportServerDiagnostics;
};

export type BugReportSubmissionInput = {
  reportId: string;
  area: BugReportArea;
  description: string;
  expectedBehavior: string;
  reproductionSteps: string;
  email: string;
  includeDiagnostics?: boolean;
  clientDiagnostics?: BugReportClientDiagnostics | null;
  screenshotConsent: boolean;
  screenshot: File | null;
};

export type BugReportSubmissionResult = {
  reportId: string;
  issueNumber: number;
  issueUrl: string;
  status: "complete" | "partial";
  warning: string;
};

type RequestOptions = { retries?: number; timeoutMs?: number };
type APIRequest = <T>(path: string, init?: RequestInit, options?: RequestOptions) => Promise<T>;

export function createBugReportAPI(request: APIRequest) {
  return {
    bugReportConfig: (includeDiagnostics = false) => request<BugReportCapability>(`/bug-reports/config${includeDiagnostics ? "?diagnostics=true" : ""}`),
    submitBugReport: (input: BugReportSubmissionInput) => {
      const form = new FormData();
      form.set("reportId", input.reportId);
      form.set("area", input.area);
      form.set("description", input.description);
      if (input.expectedBehavior) form.set("expectedBehavior", input.expectedBehavior);
      if (input.reproductionSteps) form.set("reproductionSteps", input.reproductionSteps);
      if (input.email) form.set("email", input.email);
      form.set("includeDiagnostics", String(input.includeDiagnostics !== false));
      if (input.includeDiagnostics !== false && input.clientDiagnostics) {
        form.set("clientDiagnostics", JSON.stringify(input.clientDiagnostics));
      }
      form.set("screenshotConsent", String(input.screenshotConsent));
      if (input.screenshot) form.set("screenshot", input.screenshot, input.screenshot.name);
      return request<BugReportSubmissionResult>("/bug-reports", { method: "POST", body: form }, { timeoutMs: 30_000 });
    }
  };
}
