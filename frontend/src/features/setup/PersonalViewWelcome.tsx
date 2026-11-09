import { AppChoiceInput } from "../../shared/ui/AppChoiceInput";
import { useEffect, useRef, useState, type FormEvent } from "react";
import {
  applySidebarProfile, defaultSidebarOrder, normalizeSidebarPrefs, readSidebarPrefs,
  sidebarOrderChangedEvent, sidebarPrefsKey, type SidebarPrefs, type SidebarProfile
} from "../../app/sidebarPreferences";
import { saveSidebarPreference, sidebarSaveState } from "../../app/sidebarSaving";
import { api, captureSessionGuard } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

function serverPreferences(value: string, fallback: SidebarPrefs): SidebarPrefs {
  try {
    const decoded: unknown = JSON.parse(value);
    if (!decoded || typeof decoded !== "object") return fallback;
    const order = "order" in decoded && Array.isArray(decoded.order)
      ? decoded.order.flatMap(item => defaultSidebarOrder.filter(view => view === item)) : [];
    const hidden = "hidden" in decoded && Array.isArray(decoded.hidden)
      ? decoded.hidden.flatMap(item => defaultSidebarOrder.filter(view => view === item)) : [];
    return normalizeSidebarPrefs({ order, hidden });
  } catch { return fallback; }
}

export function PersonalViewWelcome({ username, onComplete }: { username: string; onComplete: () => void }) {
  const { t } = useI18n();
  const [profile, setProfile] = useState<SidebarProfile>("collection");
  const [preferences, setPreferences] = useState<SidebarPrefs | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [failed, setFailed] = useState(false);
  const [loadRevision, setLoadRevision] = useState(0);
  const mounted = useRef(true);
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);
  useEffect(() => {
    let current = true;
    setLoading(true);
    setPreferences(null);
    setFailed(false);
    api.profileSettings().then(({ settings }) => {
      if (!current) return;
      const local = readSidebarPrefs(username);
      setPreferences(settings[sidebarPrefsKey(username)] ? serverPreferences(settings[sidebarPrefsKey(username)], local) : local);
    }).catch(() => { if (current) setFailed(true); }).finally(() => { if (current) setLoading(false); });
    return () => { current = false; };
  }, [username, loadRevision]);

  async function save(event: FormEvent) {
    event.preventDefault();
    if (!preferences || saving || loading) return;
    setSaving(true);
    setFailed(false);
    const current = captureSessionGuard();
    const next = applySidebarProfile(preferences, profile);
    const key = sidebarPrefsKey(username);
    await saveSidebarPreference(username, current, async () => {
      await api.updateProfileSettings({ [key]: JSON.stringify(next) });
      if (!current()) return;
      window.localStorage.setItem(key, JSON.stringify(next));
      window.dispatchEvent(new Event(sidebarOrderChangedEvent));
      if (mounted.current) onComplete();
    });
    if (mounted.current) {
      setSaving(false);
      setFailed(sidebarSaveState(username) === "failed");
    }
  }

  return <main className="auth-page"><section className="auth-card personal-view-welcome" aria-labelledby="personal-view-title">
    <h1 id="personal-view-title">{t("setup.personalView.title")}</h1>
    <p>{t("setup.personalView.intro")}</p>
    <form onSubmit={event => { void save(event); }}>
      <fieldset disabled={saving || loading}>
        <legend>{t("settings.sidebarProfile.title")}</legend>
        {(["collection", "all"] as const).map(value => <label className="personal-view-choice" key={value}>
          <AppChoiceInput type="radio" name="personalView" value={value} checked={profile === value}
            onChange={() => setProfile(value)} />
          <span><strong>{t(`settings.sidebarProfile.${value}`)}</strong><small>{t(`setup.personalView.${value}`)}</small></span>
        </label>)}
      </fieldset>
      <p>{t("setup.personalView.changeLater")}</p>
      {loading && <p role="status">{t("app.init")}</p>}
      {failed && <div role="alert"><p>{t("setup.personalView.failed")}</p>
        <button type="button" className="secondary-button" disabled={saving} onClick={() => setLoadRevision(value => value + 1)}>{t("common.refresh")}</button>
      </div>}
      <div className="personal-view-actions">
        <button type="submit" className="primary-button" disabled={saving || loading || !preferences}>{t(saving ? "settings.sidebarProfile.saving" : "setup.personalView.start")}</button>
        <button type="button" className="secondary-button" disabled={saving} onClick={onComplete}>{t("setup.personalView.later")}</button>
      </div>
    </form>
  </section></main>;
}
