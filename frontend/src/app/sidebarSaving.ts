type SaveState = "idle" | "saving" | "failed";
type Pending = { revision: number; current: () => boolean; state: Exclude<SaveState, "idle"> };

export const sidebarSaveChangedEvent = "railkeeper-sidebar-save-state-changed";
const pending = new Map<string, Pending>();
const queues = new Map<string, Promise<void>>();
const revisions = new Map<string, number>();
let nextRevision = 0;

export function sidebarSaveState(username: string): SaveState {
  const entry = pending.get(username);
  return entry?.current() ? entry.state : "idle";
}

export function captureSidebarRead(username: string) {
  const revision = revisions.get(username);
  const startedIdle = sidebarSaveState(username) === "idle";
  return () => startedIdle && revision === revisions.get(username) && sidebarSaveState(username) === "idle";
}

// The queue belongs to the authenticated account, beyond a settings-view mount.
// Native full-document exits still end this browser document's pending work.
export function saveSidebarPreference(username: string, current: () => boolean, save: () => Promise<unknown>) {
  const revision = ++nextRevision;
  revisions.set(username, revision);
  pending.set(username, { revision, current, state: "saving" });
  const notify = () => window.dispatchEvent(new Event(sidebarSaveChangedEvent));
  const finish = (state: "saved" | "failed" | "cancelled") => {
    const entry = pending.get(username);
    if (entry?.revision !== revision) return;
    if (state === "failed" && current()) entry.state = "failed";
    else pending.delete(username);
    notify();
  };
  notify();
  const work = (queues.get(username) || Promise.resolve()).then(async () => {
    try {
      if (!current()) { finish("cancelled"); return; }
      await save();
      finish("saved");
    } catch { finish("failed"); }
  });
  queues.set(username, work);
  const releaseQueue = () => { if (queues.get(username) === work) queues.delete(username); };
  work.then(releaseQueue, releaseQueue);
  return work;
}
