import { FormEvent, useEffect, useState } from "react";
import { api, ApiError } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

export function SetupView({ onComplete }: { onComplete: (username: string) => void }) {
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [passwordRepeat, setPasswordRepeat] = useState("");
  const [setupCode, setSetupCode] = useState("");
  const [cookieMismatch, setCookieMismatch] = useState(false);
  const [message, setMessage] = useState("");
  const [saving, setSaving] = useState(false);
  const { t } = useI18n();

  useEffect(() => {
    let active = true;
    api.setupStatus().then((status) => {
      if (active) setCookieMismatch(status.cookieSecure !== (window.location.protocol === "https:"));
    }).catch((error: Error) => { if (active) setMessage(error.message); });
    return () => { active = false; };
  }, []);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setMessage("");

    if (password !== passwordRepeat) {
      setMessage(t("setup.passwordMismatch"));
      return;
    }

    setSaving(true);

    api
      .createAdmin({ username, email, password, setupCode: setupCode.trim() })
      .then(() => onComplete(username.trim()))
      .catch((error: Error) => {
        const key = error instanceof ApiError && error.code === "invalid_setup_code" ? "setup.invalidCode"
          : error instanceof ApiError && error.code === "already_setup" ? "setup.alreadyComplete" : null;
        setMessage(key ? t(key) : error.message);
      })
      .finally(() => setSaving(false));
  };

  return (
    <main className="auth-page">
      <section className="auth-card" aria-labelledby="setup-title">
        <img className="auth-logo" src="/brand/railkeeper-logo.png" alt="RailKeeper" />
        <h1 id="setup-title">{t("setup.title")}</h1>
        <p>{t("setup.subtitle")}</p>
        {cookieMismatch && <p className="form-message" role="alert">{t("setup.cookieMismatch")}</p>}

        <form className="auth-form" onSubmit={submit}>
          <label>
            <span id="setup-code-label">{t("setup.code")}</span>
            <input
              value={setupCode}
              autoComplete="off"
              spellCheck={false}
              aria-labelledby="setup-code-label"
              aria-describedby="setup-code-help"
              onChange={(event) => setSetupCode(event.target.value)}
              required
            />
            <small id="setup-code-help">{t("setup.codeHelp")}</small>
          </label>
          <label>
            {t("auth.username")}
            <input
              value={username}
              minLength={3}
              autoComplete="username"
              onChange={(event) => setUsername(event.target.value)}
              required
            />
          </label>

          <label>
            {t("auth.email")}
            <input
              type="email"
              value={email}
              autoComplete="email"
              onChange={(event) => setEmail(event.target.value)}
              required
            />
          </label>

          <label>
            {t("auth.password")}
            <input
              type="password"
              value={password}
              minLength={12}
              autoComplete="new-password"
              onChange={(event) => setPassword(event.target.value)}
              required
            />
          </label>

          <label>
            {t("setup.passwordRepeat")}
            <input
              type="password"
              value={passwordRepeat}
              minLength={12}
              autoComplete="new-password"
              onChange={(event) => setPasswordRepeat(event.target.value)}
              required
            />
          </label>

          <button className="primary-button" disabled={saving}>
            {saving ? t("setup.saving") : t("setup.submit")}
          </button>

          {message && <p className="form-message">{message}</p>}
        </form>
      </section>
    </main>
  );
}
