import { FormEvent, useEffect, useState } from "react";
import { api, ApiError, Session } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

const initialResetToken = () => {
  if (typeof window === "undefined") {
    return "";
  }
  return new URLSearchParams(window.location.search).get("token") || "";
};

export function LoginView({ onLogin }: { onLogin: (session: Session) => void }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [twoFactorCode, setTwoFactorCode] = useState("");
  const [twoFactorRequired, setTwoFactorRequired] = useState(false);
  const [resetEmail, setResetEmail] = useState("");
  const [resetOpen, setResetOpen] = useState(false);
  const [resetToken, setResetToken] = useState(initialResetToken);
  const [resetPassword, setResetPassword] = useState("");
  const [resetPasswordRepeat, setResetPasswordRepeat] = useState("");
  const [message, setMessage] = useState("");
  const [recoveryMessage, setRecoveryMessage] = useState("");
  const [saving, setSaving] = useState(false);
  const { t } = useI18n();

  useEffect(() => {
    if (resetToken && window.location.search.includes("token=")) {
      window.history.replaceState(null, "", "/password-reset");
    }
  }, [resetToken]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setSaving(true);
    setMessage("");
    setRecoveryMessage("");

    api
      .login({ username, password, twoFactorCode: twoFactorRequired ? twoFactorCode : undefined })
      .then(onLogin)
      .catch((error: Error) => {
        if (error instanceof ApiError && error.code === "two_factor_required") {
          setTwoFactorRequired(true);
          setMessage(t("auth.twoFactor.required"));
          return;
        }
        setMessage(error.message);
      })
      .finally(() => setSaving(false));
  };

  const requestReset = () => {
    setSaving(true);
    setMessage("");
    setRecoveryMessage("");

    api
      .requestPasswordReset({ email: resetEmail })
      .then((result) => {
        setRecoveryMessage(result.message || t("auth.recovery.requested"));
      })
      .catch((error: Error) => setRecoveryMessage(error.message))
      .finally(() => setSaving(false));
  };

  const confirmReset = (event: FormEvent) => {
    event.preventDefault();
    setSaving(true);
    setMessage("");

    if (!resetToken.trim()) {
      setMessage(t("auth.reset.invalidToken"));
      setSaving(false);
      return;
    }
    if (resetPassword !== resetPasswordRepeat) {
      setMessage(t("auth.reset.mismatch"));
      setSaving(false);
      return;
    }

    api
      .confirmPasswordReset({ token: resetToken, newPassword: resetPassword })
      .then(() => {
        setRecoveryMessage(t("auth.reset.done"));
        setResetToken("");
        setResetPassword("");
        setResetPasswordRepeat("");
        window.history.replaceState(null, "", "/");
      })
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
  };

  if (resetToken) {
    return (
      <main className="auth-page">
        <section className="auth-card" aria-labelledby="password-reset-title">
          <img className="auth-logo" src="/brand/railkeeper-logo.png" alt="RailKeeper" />
          <h1 id="password-reset-title">{t("auth.reset.title")}</h1>
          <p>{t("auth.reset.subtitle")}</p>

          <form className="auth-form" onSubmit={confirmReset}>
            <label>
              {t("auth.reset.newPassword")}
              <input
                type="password"
                value={resetPassword}
                autoComplete="new-password"
                onChange={(event) => setResetPassword(event.target.value)}
                minLength={12}
                required
              />
            </label>

            <label>
              {t("auth.reset.confirmPassword")}
              <input
                type="password"
                value={resetPasswordRepeat}
                autoComplete="new-password"
                onChange={(event) => setResetPasswordRepeat(event.target.value)}
                minLength={12}
                required
              />
            </label>

            <button className="primary-button" disabled={saving}>
              {saving ? t("auth.reset.saving") : t("auth.reset.submit")}
            </button>

            <button
              type="button"
              className="forgot-password-button"
              onClick={() => {
                setResetToken("");
                setMessage("");
                window.history.replaceState(null, "", "/");
              }}
            >
              {t("auth.reset.backToLogin")}
            </button>

            {message && <p className="form-message">{message}</p>}
          </form>
        </section>
      </main>
    );
  }

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
              onChange={(event) => {
                setPassword(event.target.value);
                setTwoFactorRequired(false);
                setTwoFactorCode("");
              }}
              required
            />
          </label>

          {twoFactorRequired && (
            <label>
              {t("auth.twoFactor.code")}
              <input
                value={twoFactorCode}
                inputMode="numeric"
                autoComplete="one-time-code"
                pattern="[0-9 ]{6,8}"
                onChange={(event) => setTwoFactorCode(event.target.value)}
                required
              />
            </label>
          )}

          <button className="primary-button" disabled={saving}>
            {saving ? t("auth.login.saving") : t("auth.login.submit")}
          </button>

          <button
            type="button"
            className="forgot-password-button"
            onClick={() => {
              setResetOpen((current) => !current);
              setRecoveryMessage("");
            }}
          >
            {t("auth.forgot")}
          </button>

          {resetOpen && (
            <div className="password-reset-form">
              <label>
                {t("auth.email")}
                <input
                  type="email"
                  value={resetEmail}
                  autoComplete="email"
                  onChange={(event) => setResetEmail(event.target.value)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter") {
                      event.preventDefault();
                      requestReset();
                    }
                  }}
                  required
                />
              </label>
              <button type="button" className="secondary-button" onClick={requestReset} disabled={saving || !resetEmail.trim()}>
                {t("auth.recovery.submit")}
              </button>
            </div>
          )}

          {recoveryMessage && <p className="auth-hint">{recoveryMessage}</p>}
          {message && <p className="form-message">{message}</p>}
        </form>
      </section>
    </main>
  );
}
