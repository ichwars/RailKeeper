const key = "railkeeper.setup.completedUser";

export function readSetupWelcome() {
  try { return window.sessionStorage.getItem(key) || ""; } catch { return ""; }
}

export function markSetupWelcome(username: string) {
  try { window.sessionStorage.setItem(key, username); } catch { /* The current tab still offers the choice. */ }
}

export function clearSetupWelcome() {
  try { window.sessionStorage.removeItem(key); } catch { /* Storage may be unavailable. */ }
}
