import { Suspense, lazy, useEffect, useRef, useState } from "react";
import { LoginView } from "../features/auth/LoginView";
import { PersonalViewWelcome } from "../features/setup/PersonalViewWelcome";
import { SetupView } from "../features/setup/SetupView";
import { DecoderLibraryProvider } from "../shared/DecoderLibraryProvider";
import { ApiError, Session, api } from "../shared/api";
import { observeAction } from "../shared/asyncAction";
import { useI18n } from "../shared/i18n";
import { applyThemePreference, readThemePreference } from "../shared/theme";
import { Shell } from "./Shell";
import { availableStartView } from "./navigationAvailability";
import { clearSetupWelcome, markSetupWelcome, readSetupWelcome } from "./setupWelcome";

export type AppView = "overview" | "vehicles" | "accessories" | "exhibition" | "importExport" | "digitalCenters" | "settings";

const defaultViewSettingKey = "railkeeper.settings.defaultView";
const OverviewView = lazy(() => import("../features/overview/OverviewView").then((module) => ({ default: module.OverviewView })));
const VehiclesView = lazy(() => import("../features/vehicles/VehiclesView").then((module) => ({ default: module.VehiclesView })));
const AccessoriesView = lazy(() => import("../features/accessories/AccessoriesView").then((module) => ({ default: module.AccessoriesView })));
const ExhibitionView = lazy(() => import("../features/exhibition/ExhibitionView").then((module) => ({ default: module.ExhibitionView })));
const ImportExportView = lazy(() => import("../features/importExport/ImportExportView").then((module) => ({ default: module.ImportExportView })));
const DigitalCentersView = lazy(() => import("../features/digitalCenters/DigitalCentersView").then((module) => ({ default: module.DigitalCentersView })));
const SettingsView = lazy(() => import("../features/settings/SettingsView").then((module) => ({ default: module.SettingsView })));

export function configuredStartView(): AppView {
  return availableStartView(window.localStorage.getItem(defaultViewSettingKey) || "");
}

export function pathForView(nextView: AppView) {
  if (nextView === "overview") return "/overview";
  if (nextView === "vehicles") return "/vehicles";
  if (nextView === "accessories") return "/accessories";
  if (nextView === "exhibition") return "/exhibition";
  if (nextView === "importExport") return "/import-export";
  if (nextView === "digitalCenters") return "/digital-centers";
  if (nextView === "settings") return "/settings";
  return "/";
}

export function currentView(): AppView {
  if (window.location.pathname.startsWith("/overview")) {
    return "overview";
  }
  if (window.location.pathname.startsWith("/vehicles")) {
    return "vehicles";
  }
  if (window.location.pathname.startsWith("/accessories")) {
    return "accessories";
  }
  if (window.location.pathname.startsWith("/exhibition")) {
    return "exhibition";
  }
  if (window.location.pathname.startsWith("/import-export")) {
    return "importExport";
  }
  if (window.location.pathname.startsWith("/digital-centers")) {
    return "digitalCenters";
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
  if (view === "importExport") return canUseInventory || canUseExhibition;
  return canUseInventory;
}

function firstAllowedView(roles: string[]): AppView {
  if (roles.includes("Admin")) return "overview";
  if (roles.includes("Editor") || roles.includes("Viewer")) return "overview";
  if (roles.includes("Messe")) return "exhibition";
  return "overview";
}

function ViewLoading() {
  const { t } = useI18n();
  return (
    <section className="panel">
      <p>{t("app.init")}</p>
    </section>
  );
}

export function App() {
  const [setupUsername, setSetupUsername] = useState(readSetupWelcome);
  const [setupRequired, setSetupRequired] = useState<boolean | null>(null);
  const [session, setSession] = useState<Session | null | undefined>(undefined);
  const [loadError, setLoadError] = useState("");
  const [logoutError, setLogoutError] = useState("");
  const logoutInFlight = useRef(false);
  const [view, setView] = useState<AppView>(currentView);
  const [navigationRevision, setNavigationRevision] = useState(0);
  const { t } = useI18n();

  useEffect(() => {
    applyThemePreference(readThemePreference());
  }, []);

  useEffect(() => {
    const syncView = () => {
      setView(currentView());
      setNavigationRevision(previous => previous + 1);
    };

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

  async function handleLogout() {
    if (logoutInFlight.current) return;
    logoutInFlight.current = true;
    setLogoutError("");
    try {
      await api.logout();
      setSession(null);
    } catch (error) {
      if (error instanceof ApiError && (error.status === 401 || error.status === 403)) {
        try {
          const current = await api.session();
          if (current.username !== session?.username) {
            setSession(null);
            return;
          }
          setSession(current);
        } catch (sessionError) {
          if (sessionError instanceof ApiError && sessionError.status === 401) {
            setSession(null);
            return;
          }
        }
      }
      setLogoutError(error instanceof Error ? error.message : String(error));
    } finally {
      logoutInFlight.current = false;
    }
  }

  if (loadError) {
    return (
      <main className="auth-page">
        <section className="auth-card">
          <h1>RailKeeper</h1>
          <p>{loadError}</p>
        </section>
      </main>
    );
  }

  if (setupRequired === null) {
    return (
      <main className="auth-page">
        <section className="auth-card">
          <h1>RailKeeper</h1>
          <p>{t("app.init")}</p>
        </section>
      </main>
    );
  }

  if (setupRequired) {
    return (
      <SetupView
        onComplete={username => {
          markSetupWelcome(username);
          setSetupUsername(username);
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
          <h1>RailKeeper</h1>
          <p>{t("app.session")}</p>
        </section>
      </main>
    );
  }

  if (session === null) {
    return <LoginView onLogin={handleLogin} />;
  }

  if (setupUsername === session.username && session.roles.some(role => ["Admin", "Editor", "Viewer"].includes(role))) {
    return <PersonalViewWelcome username={session.username} onComplete={() => {
      clearSetupWelcome();
      setSetupUsername("");
    }} />;
  }

  const effectiveView = canAccessView(view, session.roles) ? view : firstAllowedView(session.roles);
  if (effectiveView !== view) {
    window.history.replaceState(null, "", pathForView(effectiveView));
  }

  return (
    <DecoderLibraryProvider enabled={session.roles.some(role => ["Admin", "Editor", "Viewer"].includes(role))}><Shell
      username={session.username}
      roles={session.roles}
      activeView={effectiveView}
      onNavigate={nextView => {
        window.history.pushState(null, "", pathForView(nextView));
        setView(nextView);
        setNavigationRevision(previous => previous + 1);
      }}
      onLogout={() => observeAction(handleLogout())}
    >
      {logoutError && <p className="form-message error" role="alert">{logoutError}</p>}
      <Suspense key={navigationRevision} fallback={<ViewLoading />}>
        {effectiveView === "overview" && <OverviewView username={session.username} roles={session.roles} />}
        {effectiveView === "vehicles" && <VehiclesView username={session.username} roles={session.roles} />}
        {effectiveView === "accessories" && <AccessoriesView roles={session.roles} />}
        {effectiveView === "exhibition" && <ExhibitionView roles={session.roles} />}
        {effectiveView === "importExport" && <ImportExportView roles={session.roles} />}
        {effectiveView === "digitalCenters" && <DigitalCentersView roles={session.roles} />}
        {effectiveView === "settings" && <SettingsView username={session.username} />}
      </Suspense>
    </Shell></DecoderLibraryProvider>
  );
}
