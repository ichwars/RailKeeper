import { FormEvent, useState } from "react";
import { api, Session } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

export function LoginView({ onLogin }: { onLogin: (session: Session) => void }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState("");
  const [recoveryMessage, setRecoveryMessage] = useState("");
  const [saving, setSaving] = useState(false);
  const { t } = useI18n();

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setSaving(true);
    setMessage("");
    setRecoveryMessage("");

    api
      .login({ username, password })
      .then(onLogin)
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
  };

  return (
    <main className="auth-page">
      <section className="auth-card" aria-labelledby="login-title">
        <img className="auth-logo" src="/brand/railkeeper-logo.png" alt="RailKeeper" />
        <h1 id="login-title">{t("auth.login.title")}</h1>
        <p>{t("auth.login.subtitle")}</p>

        <form className="auth-form" onSubmit={submit}>
          <label>
            {t("auth.username")}
            <input
              value={username}
              autoComplete="username"
              onChange={(event) => setUsername(event.target.value)}
              required
            />
          </label>

          <label>
            {t("auth.password")}
            <input
              type="password"
              value={password}
              autoComplete="current-password"
              onChange={(event) => setPassword(event.target.value)}
              required
            />
          </label>

          <button className="primary-button" disabled={saving}>
            {saving ? t("auth.login.saving") : t("auth.login.submit")}
          </button>

          <button
            type="button"
            className="forgot-password-button"
            onClick={() => setRecoveryMessage(t("auth.recovery"))}
          >
            {t("auth.forgot")}
          </button>

          {recoveryMessage && <p className="auth-hint">{recoveryMessage}</p>}
          {message && <p className="form-message">{message}</p>}
        </form>
      </section>
    </main>
  );
}
