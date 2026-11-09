// Expected request failures belong to their form/controller. Unexpected failures
// must still reach the browser's existing error/diagnostics path, never disappear.
export function reportUnexpectedAsyncError(error: unknown) {
  if (typeof window.reportError === "function") {
    window.reportError(error);
    return;
  }
  window.setTimeout(() => { throw error; }, 0);
}

export function observeAction(work: unknown): void {
  Promise.resolve(work).catch(reportUnexpectedAsyncError);
}
