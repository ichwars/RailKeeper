import type { ReactNode } from "react";
import { lazy, Suspense, useCallback, useEffect, useRef, useState } from "react";
import { Bug, ChevronLeft, ChevronRight, LogOut, Menu, Monitor, Moon, Settings, Sun, X } from "lucide-react";
import type { AppView } from "./App";
import { api } from "../shared/api";
import { installClientErrorCapture } from "../features/bugReporting/clientErrorBuffer";
import type { BugReportCapability } from "../shared/apiBugReports";
import { useI18n } from "../shared/i18n";
import { applyThemePreference, readThemePreference, themePreferenceKey, type ThemePreference } from "../shared/theme";
import { sidebarNavigation } from "./sidebarNavigation";
import { SidebarAdditionalNavigation } from "./SidebarAdditionalNavigation";
import { followAppLink } from "./appNavigation";
import { captureSidebarRead } from "./sidebarSaving";
import {
  readSidebarPrefs,
  sidebarOrderChangedEvent,
  sidebarPrefsKey
} from "./sidebarPreferences";

const LazyBugReportDialog = lazy(() => import("../features/bugReporting/BugReportDialog").then((module) => ({
  default: module.BugReportDialog
})));
const bugReportFallbackURL = "https://github.com/ichwars/RailKeeper/issues/new";

const sidebarCollapsedKey = "railkeeper.sidebarCollapsed";
const betaUpdatesKey = "railkeeper.settings.betaUpdates";
export const updateStatusChangedEvent = "railkeeper-update-status-changed";

function GitHubMark({ size = 17 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" aria-hidden="true" focusable="false">
      <path
        fill="currentColor"
        d="M12 2C6.48 2 2 6.58 2 12.25c0 4.53 2.87 8.37 6.84 9.73.5.1.68-.22.68-.49l-.01-1.9c-2.78.62-3.37-1.22-3.37-1.22-.46-1.19-1.12-1.51-1.12-1.51-.91-.64.07-.63.07-.63 1.01.07 1.54 1.06 1.54 1.06.9 1.58 2.36 1.12 2.94.86.09-.67.35-1.12.64-1.38-2.22-.26-4.56-1.14-4.56-5.05 0-1.12.39-2.03 1.03-2.75-.1-.26-.45-1.3.1-2.71 0 0 .84-.28 2.75 1.05A9.29 9.29 0 0 1 12 6.97c.85 0 1.7.12 2.5.35 1.9-1.33 2.74-1.05 2.74-1.05.55 1.41.2 2.45.1 2.71.64.72 1.03 1.63 1.03 2.75 0 3.92-2.34 4.79-4.57 5.04.36.32.68.95.68 1.92l-.01 2.8c0 .27.18.59.69.49A10.16 10.16 0 0 0 22 12.25C22 6.58 17.52 2 12 2Z"
      />
    </svg>
  );
}

function readSidebarCollapsed() {
  return window.localStorage.getItem(sidebarCollapsedKey) === "true";
}

function readLocalBool(key: string, fallback: boolean) {
  const value = window.localStorage.getItem(key);
  if (value === null) return fallback;
  return value === "true";
}

function readNavItems(roles: string[], username: string) {
  return sidebarNavigation(roles, readSidebarPrefs(username));
}

export function Shell({
  children,
  username,
  roles,
  activeView,
  onLogout,
  onNavigate
}: {
  children: ReactNode;
  username: string;
  roles: string[];
  activeView: AppView;
  onLogout: () => void;
  onNavigate?: (view: AppView) => void;
}) {
  const loadBugReportDiagnostics = useCallback(() => api.bugReportConfig(true), []);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(readSidebarCollapsed);
  const [theme, setTheme] = useState<ThemePreference>(readThemePreference);
  const [orderedNavItems, setOrderedNavItems] = useState(() => readNavItems(roles, username));
  const sidebarRevision = useRef(0);
  const mainRef = useRef<HTMLElement>(null);
  const navigate = onNavigate ? (view: AppView) => {
    onNavigate(view);
    const main = mainRef.current;
    if (main) {
      main.scrollTop = 0;
      main.scrollLeft = 0;
      main.focus({ preventScroll: true });
    }
  } : undefined;
  const [appVersion, setAppVersion] = useState("");
  const [updateAvailable, setUpdateAvailable] = useState(false);
  const [bugReportCapability, setBugReportCapability] = useState<BugReportCapability | null>(null);
  const [bugReportDialogCapability, setBugReportDialogCapability] = useState<BugReportCapability | null>(null);
  useEffect(() => {
    if (bugReportCapability?.enabled) return installClientErrorCapture();
  }, [bugReportCapability?.enabled]);
  const ThemeIcon = theme === "dark" ? Moon : theme === "light" ? Sun : Monitor;
  const { t } = useI18n();
  const refreshBugReportCapability = useCallback(async () => {
    try {
      setBugReportCapability(await api.bugReportConfig());
    } catch {
      setBugReportCapability(null);
    }
  }, []);

  useEffect(() => {
    const syncOrder = () => {
      sidebarRevision.current += 1;
      setOrderedNavItems(readNavItems(roles, username));
    };
    syncOrder();

    window.addEventListener(sidebarOrderChangedEvent, syncOrder);
    window.addEventListener("storage", syncOrder);
    return () => {
      window.removeEventListener(sidebarOrderChangedEvent, syncOrder);
      window.removeEventListener("storage", syncOrder);
    };
  }, [roles, username]);

  useEffect(() => {
    let cancelled = false;
    const requestedSidebarRevision = sidebarRevision.current;
    const sidebarReadCurrent = captureSidebarRead(username);
    api
      .profileSettings()
      .then(({ settings }) => {
        if (cancelled) return;
        const sidebarValue = settings[sidebarPrefsKey(username)];
        if (sidebarValue && sidebarRevision.current === requestedSidebarRevision && sidebarReadCurrent()) {
          window.localStorage.setItem(sidebarPrefsKey(username), sidebarValue);
          setOrderedNavItems(readNavItems(roles, username));
        }
        const themeValue = settings[themePreferenceKey];
        if (themeValue === "system" || themeValue === "light" || themeValue === "dark") {
          setTheme(themeValue);
          applyThemePreference(themeValue);
        }
      })
      .catch(() => undefined);
    return () => {
      cancelled = true;
    };
  }, [roles, username]);

  useEffect(() => {
    let mounted = true;

    const loadVersion = (forceCheck = false) => {
      const includeBeta = readLocalBool(betaUpdatesKey, false);
      api
        .version(forceCheck, includeBeta)
        .then((info) => {
          if (!mounted) return;
          setAppVersion(info.version);
          setUpdateAvailable(Boolean(forceCheck && info.updateAvailable && info.latestVersion));
        })
        .catch(() => {
          if (!mounted) return;
          setUpdateAvailable(false);
        });
    };
    const syncVersion = (event?: Event) => {
      const forceCheck = event instanceof CustomEvent && Boolean(event.detail?.forceCheck);
      loadVersion(forceCheck);
    };

    loadVersion();
    window.addEventListener(updateStatusChangedEvent, syncVersion);
    window.addEventListener("storage", syncVersion);
    return () => {
      mounted = false;
      window.removeEventListener(updateStatusChangedEvent, syncVersion);
      window.removeEventListener("storage", syncVersion);
    };
  }, []);

  useEffect(() => {
    if (appVersion) return;
    let mounted = true;
    api
      .version(false)
      .then((info) => {
        if (mounted) {
          setAppVersion(info.version);
        }
      })
      .catch(() => {
        if (mounted) {
          setAppVersion("");
        }
      });
    return () => {
      mounted = false;
    };
  }, [appVersion]);

  useEffect(() => {
    void refreshBugReportCapability();
  }, [refreshBugReportCapability]);

  useEffect(() => {
    if (bugReportCapability?.enabled) return;
    const refresh = () => void refreshBugReportCapability();
    const interval = window.setInterval(refresh, 30_000);
    window.addEventListener("focus", refresh);
    return () => {
      window.clearInterval(interval);
      window.removeEventListener("focus", refresh);
    };
  }, [bugReportCapability?.enabled, refreshBugReportCapability]);

  function toggleTheme() {
    const nextTheme: ThemePreference = theme === "dark" ? "light" : theme === "light" ? "system" : "dark";
    setTheme(nextTheme);
    applyThemePreference(nextTheme);
    void api.updateProfileSettings({ [themePreferenceKey]: nextTheme }).catch(() => undefined);
  }

  function toggleSidebarCollapsed() {
    setSidebarCollapsed((collapsed) => {
      const next = !collapsed;
      window.localStorage.setItem(sidebarCollapsedKey, String(next));
      return next;
    });
  }

  async function submitBugReport(input: Parameters<typeof api.submitBugReport>[0]) {
    try {
      const result = await api.submitBugReport(input);
      if (result.status === "partial") {
        void refreshBugReportCapability();
      }
      return result;
    } catch (error) {
      void refreshBugReportCapability();
      throw error;
    }
  }

  return (
    <div className={`layout${mobileMenuOpen ? " nav-open" : ""}${sidebarCollapsed ? " sidebar-collapsed" : ""}`}>
      {mobileMenuOpen && <button type="button" className="mobile-nav-scrim" aria-label={t("nav.menu.close")} onClick={() => setMobileMenuOpen(false)} />}
      <aside className={sidebarCollapsed ? "sidebar collapsed" : "sidebar"}>
        <button
          type="button"
          className="mobile-menu-button"
          aria-controls="main-navigation"
          aria-expanded={mobileMenuOpen}
          aria-label={mobileMenuOpen ? t("nav.menu.close") : t("nav.menu.open")}
          onClick={() => setMobileMenuOpen((open) => !open)}
        >
          {mobileMenuOpen ? <X size={19} aria-hidden="true" /> : <Menu size={19} aria-hidden="true" />}
        </button>
        <div className="brand">
          <img className="brand-logo" src="/brand/railkeeper-logo.png" alt="RailKeeper" />
          <img className="brand-mark" src="/brand/railkeeper-mark.png" alt="RailKeeper" />
        </div>

        <nav id="main-navigation" className="nav" aria-label={t("nav.main")}>
          {orderedNavItems.visible.map((item) => {
            const Icon = item.icon;
            const label = t(item.labelKey);
            return (
              <a
                key={item.view}
                className={`nav-entry${activeView === item.view ? " active" : ""}`}
                href={item.href}
                aria-current={activeView === item.view ? "page" : undefined}
                onClick={event => {
                  setMobileMenuOpen(false);
                  followAppLink(event, navigate ? () => navigate(item.view) : undefined);
                }}
              >
                <Icon size={16} aria-hidden="true" />
                <span>{label}</span>
              </a>
            );
          })}

          <SidebarAdditionalNavigation items={orderedNavItems.additional} activeView={activeView}
            onNavigate={() => setMobileMenuOpen(false)} onSelect={navigate} />

          <button
            type="button"
            className="sidebar-collapse"
            onClick={toggleSidebarCollapsed}
            aria-label={sidebarCollapsed ? t("nav.sidebar.expand") : t("nav.sidebar.collapse")}
            title={sidebarCollapsed ? t("nav.sidebar.expand") : t("nav.sidebar.collapse")}
          >
            {sidebarCollapsed ? <ChevronRight size={17} aria-hidden="true" /> : <ChevronLeft size={17} aria-hidden="true" />}
          </button>

          <div className="sidebar-footer" aria-label={t("nav.footerActions")}>
            <div className="sidebar-footer-actions">
              <a href="/settings" title={t("nav.system")} aria-label={t("nav.system")}
                onClick={event => {
                  setMobileMenuOpen(false);
                  followAppLink(event, navigate ? () => navigate("settings") : undefined);
                }}>
                <Settings size={17} aria-hidden="true" />
              </a>
              <a href="https://github.com/ichwars/RailKeeper" target="_blank" rel="noreferrer" title={t("nav.repository")} aria-label={t("nav.repository")}>
                <GitHubMark size={17} />
              </a>
              <button type="button" onClick={toggleTheme} title={t("nav.theme")} aria-label={t("nav.theme")}>
                <ThemeIcon size={17} aria-hidden="true" />
              </button>
              <button type="button" onClick={onLogout} title={t("nav.logout", { username })} aria-label={t("nav.logout", { username })}>
                <LogOut size={17} aria-hidden="true" />
              </button>
            </div>
            {appVersion && (
              <span className="sidebar-version-row">
                <span className="sidebar-version">v{appVersion}</span>
                {updateAvailable && (
                  <span className="sidebar-update-badge" title={t("nav.updateAvailable")} aria-label={t("nav.updateAvailable")}>
                    <span aria-hidden="true" />
                    Update
                  </span>
                )}
              </span>
            )}
          </div>
        </nav>
      </aside>

      <main className="main" ref={mainRef} tabIndex={-1}>{children}</main>

      {bugReportCapability?.enabled ? (
        <button
          type="button"
          className="feedback-button"
          title={t("nav.feedback")}
          aria-label={t("nav.feedback")}
          onClick={() => setBugReportDialogCapability(bugReportCapability)}
        >
          <Bug size={20} aria-hidden="true" />
        </button>
      ) : (
        <a
          className="feedback-button"
          href={bugReportCapability?.fallbackUrl || bugReportFallbackURL}
          target="_blank"
          rel="noreferrer"
          title={t("nav.feedback")}
          aria-label={t("nav.feedback")}
        >
          <Bug size={20} aria-hidden="true" />
        </a>
      )}
      {bugReportDialogCapability ? (
        <Suspense fallback={null}>
          <LazyBugReportDialog
            onLoadDiagnostics={loadBugReportDiagnostics}
            capability={bugReportDialogCapability}
            initialArea={activeView}
            roles={roles}
            onClose={() => setBugReportDialogCapability(null)}
            onSubmit={submitBugReport}
          />
        </Suspense>
      ) : null}
    </div>
  );
}
