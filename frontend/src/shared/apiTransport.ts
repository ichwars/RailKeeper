import { abstractRequestRoute, notifyRequestFailure } from "./apiRequestDiagnostics";

type RequestOptions = {
  retries?: number;
  timeoutMs?: number;
};

let csrfToken = "";

// Keep queued work bound to the session that created it, across view unmounts.
export function captureSessionGuard() {
  const owner = csrfToken;
  return () => csrfToken === owner;
}

export class ApiError extends Error {
  code: string;
  status: number;
  details: unknown;

  constructor(message: string, code: string, status: number, details: unknown = null) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
    this.details = details;
  }
}

export function readCookie(name: string): string {
  const prefix = `${name}=`;
  const value = document.cookie
    .split("; ")
    .find((part) => part.startsWith(prefix))
    ?.slice(prefix.length);
  return value ? decodeURIComponent(value) : "";
}

export async function request<T>(path: string, init: RequestInit = {}, options: RequestOptions = {}): Promise<T> {
  const method = (init.method || "GET").toUpperCase();
  const isFormData = typeof FormData !== "undefined" && init.body instanceof FormData;
  const headers: Record<string, string> = {
    ...(!["GET", "HEAD"].includes(method) && !isFormData ? { "Content-Type": "application/json" } : {}),
    ...((init.headers as Record<string, string>) || {})
  };
  const timeoutMs = options.timeoutMs || 12000;
  const attempts = 1 + (["GET", "HEAD"].includes(method) ? options.retries || 0 : 0);

  if (!["GET", "HEAD", "OPTIONS"].includes(method)) {
    const token = csrfToken || readCookie("rk_csrf");
    if (token) {
      headers["X-CSRF-Token"] = token;
    }
  }

  for (let attempt = 0; attempt < attempts; attempt += 1) {
    const controller = new AbortController();
    const timeout = window.setTimeout(() => controller.abort(), timeoutMs);

    try {
      const response = await fetch(`/api/v1${path}`, {
        credentials: "include",
        ...init,
        headers,
        signal: init.signal || controller.signal
      });

      if (!response.ok) {
        let message = response.statusText;
        let code = "request_failed";
        let details: unknown = null;
        try {
          const body = await response.json();
          code = body.error || code;
          message = body.message || body.error || message;
          details = body.details ?? null;
        } catch {
          // Keep the HTTP status text when the server did not return JSON.
        }
        const apiError = new ApiError(message, code, response.status, details);
        notifyRequestFailure({
          category: "api", errorClass: apiError.name, problemCode: code, method,
          route: abstractRequestRoute(path), status: response.status
        });
        throw apiError;
      }

      if (response.status === 204) {
        return undefined as T;
      }

      return response.json() as Promise<T>;
    } catch (error) {
      if (attempt + 1 < attempts && error instanceof DOMException && error.name === "AbortError") {
        continue;
      }
      if (error instanceof DOMException && error.name === "AbortError") {
        notifyRequestFailure({
          category: "timeout", errorClass: error.name, problemCode: "request_timeout", method,
          route: abstractRequestRoute(path), status: null
        });
        throw new Error("Die Anfrage hat zu lange gedauert. Bitte erneut versuchen.");
      }
      if (!(error instanceof ApiError)) {
        notifyRequestFailure({
          category: "network", errorClass: error instanceof Error ? error.name : "Error",
          problemCode: "network_error", method, route: abstractRequestRoute(path), status: null
        });
      }
      throw error;
    } finally {
      window.clearTimeout(timeout);
    }
  }

  throw new Error("Die Anfrage konnte nicht verarbeitet werden.");
}

export type APIRequest = typeof request;

export function setCSRFToken(token: string): void {
  csrfToken = token;
}
