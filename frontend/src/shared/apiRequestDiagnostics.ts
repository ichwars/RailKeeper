export type APIRequestFailureEvent = {
  category: "api" | "network" | "timeout";
  errorClass: string;
  problemCode: string;
  method: string;
  route: string;
  status: number | null;
};

let requestFailureObserver: ((event: APIRequestFailureEvent) => void) | null = null;

export function setAPIRequestFailureObserver(observer: ((event: APIRequestFailureEvent) => void) | null) {
  requestFailureObserver = observer;
}

export function notifyRequestFailure(event: APIRequestFailureEvent) {
  try {
    requestFailureObserver?.(event);
  } catch {
    // Error diagnostics must never interfere with the original request path.
  }
}

export function abstractRequestRoute(path: string) {
  const pathname = path.split(/[?#]/, 1)[0];
  return pathname.split("/").map((segment) => {
    let decoded = segment;
    try {
      decoded = decodeURIComponent(segment);
    } catch {
      return ":id";
    }
    if (/^[0-9a-f]{8}-[0-9a-f-]{27,}$/i.test(decoded) || /^\d{4,}$/.test(decoded) ||
      /^[a-z0-9_-]{20,}$/i.test(decoded)) return ":id";
    return segment;
  }).join("/").slice(0, 200);
}

