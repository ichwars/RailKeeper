import { setAPIRequestFailureObserver, type APIRequestFailureEvent } from "../../shared/apiRequestDiagnostics";
import type { BugReportClientError } from "../../shared/apiBugReports";
import { observeDiagnosticPreference, readDiagnosticPreference } from "./bugReportPreferences";
import { abstractClientErrorRoute } from "./bugReportDiagnostics";

type ClientErrorInput = Partial<Omit<BugReportClientError, "timestamp" | "status">> & {
  category: BugReportClientError["category"];
  status?: number | null;
};

export type ClientErrorBuffer = {
  record(input: ClientErrorInput): void;
  snapshot(): BugReportClientError[];
  clear(): void;
};

const safeProblemCodes = new Set([
  "network_error",
  "request_failed",
  "request_timeout",
  "unknown_error"
]);

export function createClientErrorBuffer(
  limit = 30,
  now: () => string = () => new Date().toISOString()
): ClientErrorBuffer {
  const records: BugReportClientError[] = [];
  return {
    record(input) {
      const category = input.category;
      const errorClasses = new Set([
        "AbortError", "ApiError", "DOMException", "Error", "RangeError", "ReferenceError", "SyntaxError",
        "TypeError", "UnhandledRejection"
      ]);
      const candidateProblemCode = input.problemCode || "";
      const problemCode = safeProblemCodes.has(candidateProblemCode)
        ? candidateProblemCode : category === "window" || category === "promise" ? "" : "unknown_error";
      const messages: Record<BugReportClientError["category"], string> = {
        window: "Unexpected client error",
        promise: "Unhandled promise rejection",
        api: "API request failed",
        network: "Network request failed",
        timeout: "Request timed out"
      };
      const status = input.status != null && input.status >= 100 && input.status <= 599 ? input.status : undefined;
      records.push({
        timestamp: now(),
        category,
        errorClass: errorClasses.has(input.errorClass || "") ? input.errorClass || "" : "Error",
        problemCode,
        message: messages[category],
        method: /^(DELETE|GET|HEAD|OPTIONS|PATCH|POST|PUT)$/.test((input.method || "").toUpperCase())
          ? (input.method || "").toUpperCase() : "",
        route: abstractClientErrorRoute(input.route || ""),
        ...(status === undefined ? {} : { status })
      });
      if (records.length > limit) records.splice(0, records.length - limit);
    },
    clear() { records.length = 0; },
    snapshot() {
      return records.map((record) => ({ ...record }));
    }
  };
}

export const clientErrorBuffer = createClientErrorBuffer(30);

function recordAPIError(event: APIRequestFailureEvent) {
  clientErrorBuffer.record({
    category: event.category,
    errorClass: event.errorClass,
    problemCode: event.problemCode,
    message: event.category === "api" ? "API request failed" : "Network request failed",
    method: event.method,
    route: event.route,
    status: event.status
  });
}

export function installClientErrorCapture() {
  let enabled = readDiagnosticPreference();
  const stopObserving = observeDiagnosticPreference((value) => {
    enabled = value;
    if (!enabled) clientErrorBuffer.clear();
  });
  const recordAPIErrorIfEnabled = (event: APIRequestFailureEvent) => {
    if (enabled) recordAPIError(event);
  };
  const onError = (event: ErrorEvent) => {
    if (!enabled) return;
    clientErrorBuffer.record({
      category: "window",
      errorClass: event.error?.name || "Error",
      message: "Unexpected client error"
    });
  };
  const onRejection = (event: PromiseRejectionEvent) => {
    if (!enabled) return;
    const reason = event.reason;
    clientErrorBuffer.record({
      category: "promise",
      errorClass: reason instanceof Error ? reason.name : "UnhandledRejection",
      message: "Unhandled promise rejection"
    });
  };
  window.addEventListener("error", onError);
  window.addEventListener("unhandledrejection", onRejection);
  setAPIRequestFailureObserver(recordAPIErrorIfEnabled);
  return () => {
    stopObserving();
    clientErrorBuffer.clear();
    window.removeEventListener("error", onError);
    window.removeEventListener("unhandledrejection", onRejection);
    setAPIRequestFailureObserver(null);
  };
}
