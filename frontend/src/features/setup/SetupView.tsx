import { FormEvent, useState } from "react";
import { api } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

export function SetupView({ onComplete }: { onComplete: () => void }) {
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [passwordRepeat, setPasswordRepeat] = useState("");
  const [message, setMessage] = useState("");
  const [saving, setSaving] = useState(false);
  const { t } = useI18n();

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setMessage("");

    if (password !== passwordRepeat) {
      setMessage(t("setup.passwordMismatch"));
      return;
    }

    setSaving(true);

    api
      .createAdmin({ username, email, password })
      .then(onComplete)
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
  };

  return (
    <main className="auth-page">
      <section className="auth-card" aria-labelledby="setup-title">
        <img className="auth-logo" src="/brand/railkeeper-logo.png" alt="RailKeeper" />
        <h1 id="setup-title">{t("setup.title")}</h1>
        <p>{t("setup.subtitle")}</p>

        <form className="auth-form" onSubmit={submit}>
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
