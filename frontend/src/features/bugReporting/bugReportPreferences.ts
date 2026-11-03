const preferenceKey = "railkeeper.bugReports.includeDiagnostics.v1";
const preferenceEvent = "railkeeper-bug-report-preference";

let sessionPreference: boolean | undefined;

export function readDiagnosticPreference(): boolean {
  if (sessionPreference !== undefined) return sessionPreference;
  try {
    return window.localStorage.getItem(preferenceKey) !== "false";
  } catch {
    return true;
  }
}

export function saveDiagnosticPreference(enabled: boolean) {
  sessionPreference = enabled;
  try {
    window.localStorage.setItem(preferenceKey, String(enabled));
    sessionPreference = undefined;
  } catch {
    // The current session still applies the choice when browser storage is unavailable.
  }
  window.dispatchEvent(new CustomEvent<boolean>(preferenceEvent, { detail: enabled }));
}

export function observeDiagnosticPreference(onChange: (enabled: boolean) => void) {
  const onPreference = (event: Event) => onChange((event as CustomEvent<boolean>).detail);
  const onStorage = (event: StorageEvent) => {
    if (event.key === preferenceKey || event.key === null) {
      sessionPreference = undefined;
      onChange(readDiagnosticPreference());
    }
  };
  window.addEventListener(preferenceEvent, onPreference);
  window.addEventListener("storage", onStorage);
  return () => {
    window.removeEventListener(preferenceEvent, onPreference);
    window.removeEventListener("storage", onStorage);
  };
}
