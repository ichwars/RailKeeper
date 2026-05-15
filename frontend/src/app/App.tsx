import { useEffect, useState } from "react";
import { Shell } from "./Shell";
import { LoginView } from "../features/auth/LoginView";
import { ExhibitionView } from "../features/exhibition/ExhibitionView";
import { ImportExportView } from "../features/importExport/ImportExportView";
import { OverviewView } from "../features/overview/OverviewView";
import { SetupView } from "../features/setup/SetupView";
import { SettingsView } from "../features/settings/SettingsView";
import { VehiclesView } from "../features/vehicles/VehiclesView";
import { api, Session } from "../shared/api";
import { useI18n } from "../shared/i18n";
import { applyThemePreference, readThemePreference } from "../shared/theme";

export type AppView = "overview" | "vehicles" | "exhibition" | "importExport" | "settings";

const defaultViewSettingKey = "railkeeper.settings.defaultView";

function configuredStartView(): AppView {
  const stored = window.localStorage.getItem(defaultViewSettingKey);
  if (stored === "vehicles" || stored === "exhibition" || stored === "importExport" || stored === "settings" || stored === "overview") {
    return stored;
  }
  if (stored === "inventory") {
    return "vehicles";
  }
  return "overview";
}

function pathForView(nextView: AppView) {
  if (nextView === "overview") return "/overview";
  if (nextView === "vehicles") return "/vehicles";
  if (nextView === "exhibition") return "/exhibition";
  if (nextView === "importExport") return "/import-export";
  if (nextView === "settings") return "/settings";
  return "/";
}

function currentView(): AppView {
  if (window.location.pathname.startsWith("/overview")) {
    return "overview";
  }
  if (window.location.pathname.startsWith("/vehicles")) {
    return "vehicles";
  }
  if (window.location.pathname.startsWith("/exhibition")) {
    return "exhibition";
  }
  if (window.location.pathname.startsWith("/import-export")) {
    return "importExport";
  }
  if (window.location.pathname.startsWith("/settings")) {
    return "settings";
  }
  return configuredStartView();
}

function canAccessView(view: AppView, roles: string[]) {
  if (roles.includes("Admin")) return true;
  const canUseInventory = roles.includes("Editor") || roles.includes("Viewer");
  const canUseExhibition = roles.includes("Messe");
  if (view === "exhibition") return canUseExhibition;
  return canUseInventory;
}

function firstAllowedView(roles: string[]): AppView {
  if (roles.includes("Admin")) return "overview";
  if (roles.includes("Editor") || roles.includes("Viewer")) return "overview";
  if (roles.includes("Messe")) return "exhibition";
  return "overview";
}

export function App() {
  const [setupRequired, setSetupRequired] = useState<boolean | null>(null);
  const [session, setSession] = useState<Session | null | undefined>(undefined);
  const [loadError, setLoadError] = useState("");
  const [view, setView] = useState<AppView>(currentView);
  const { t } = useI18n();

  useEffect(() => {
    applyThemePreference(readThemePreference());
  }, []);

  useEffect(() => {
    const syncView = () => setView(currentView());

    window.addEventListener("popstate", syncView);
    return () => window.removeEventListener("popstate", syncView);
  }, []);

  useEffect(() => {
    api
      .setupStatus()
      .then((status) => {
        setSetupRequired(status.setupRequired);
        if (status.setupRequired) {
          setSession(null);
          return;
        }
        api.session().then(setSession).catch(() => setSession(null));
      })
      .catch((error: Error) => setLoadError(error.message));
  }, []);

  function handleLogin(nextSession: Session) {
    const nextView = firstAllowedView(nextSession.roles);
    window.history.replaceState(null, "", pathForView(nextView));
    setView(nextView);
    setSession(nextSession);
  }

  if (loadError) {
    return (
      <main className="auth-page">
        <section className="auth-card">
          <h1>RailKeeper2</h1>
          <p>{loadError}</p>
        </section>
      </main>
    );
  }

  if (setupRequired === null) {
    return (
      <main className="auth-page">
        <section className="auth-card">
          <h1>RailKeeper2</h1>
          <p>{t("app.init")}</p>
        </section>
      </main>
    );
  }

  if (setupRequired) {
    return (
      <SetupView
        onComplete={() => {
          setSetupRequired(false);
          setSession(null);
        }}
      />
    );
  }

  if (session === undefined) {
    return (
      <main className="auth-page">
        <section className="auth-card">
          <h1>RailKeeper2</h1>
          <p>{t("app.session")}</p>
        </section>
      </main>
    );
  }

  if (session === null) {
    return <LoginView onLogin={handleLogin} />;
  }

  const effectiveView = canAccessView(view, session.roles) ? view : firstAllowedView(session.roles);
  if (effectiveView !== view) {
    window.history.replaceState(null, "", pathForView(effectiveView));
  }

  return (
    <Shell
      username={session.username}
      roles={session.roles}
      activeView={effectiveView}
      onLogout={() => {
        api.logout().finally(() => setSession(null));
      }}
    >
      {effectiveView === "overview" && <OverviewView />}
      {effectiveView === "vehicles" && <VehiclesView />}
      {effectiveView === "exhibition" && <ExhibitionView roles={session.roles} />}
      {effectiveView === "importExport" && <ImportExportView />}
      {effectiveView === "settings" && <SettingsView username={session.username} />}
    </Shell>
  );
}
