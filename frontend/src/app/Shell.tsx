import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import { BarChart3, Box, Bug, CalendarDays, ChevronLeft, ChevronRight, FileInput, LogOut, Menu, Monitor, Moon, Settings, Sun, X } from "lucide-react";
import type { AppView } from "./App";
import { useI18n } from "../shared/i18n";
import { applyThemePreference, readThemePreference, type ThemePreference } from "../shared/theme";

const navItems = [
  { view: "overview", href: "/overview", labelKey: "nav.overview", icon: BarChart3 },
  { view: "vehicles", href: "/vehicles", labelKey: "nav.vehicles", icon: Box },
  { view: "exhibition", href: "/exhibition", labelKey: "nav.exhibition", icon: CalendarDays },
  { view: "importExport", href: "/import-export", labelKey: "nav.importExport", icon: FileInput },
  { view: "settings", href: "/settings", labelKey: "nav.settings", icon: Settings }
] as const;

const sidebarCollapsedKey = "railkeeper.sidebarCollapsed";
const sidebarOrderKey = "railkeeper.settings.sidebarOrder";
const sidebarPrefsBaseKey = "railkeeper.settings.sidebarPrefs";
const sidebarOrderChangedEvent = "railkeeper-sidebar-order-changed";

type SidebarPrefs = {
  order: AppView[];
  hidden: AppView[];
};

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

function allowedNavItems(roles: string[]) {
  if (roles.includes("Admin")) return [...navItems];
  const canUseInventory = roles.includes("Editor") || roles.includes("Viewer");
  const canUseExhibition = roles.includes("Messe");
  return navItems.filter((item) => {
    if (item.view === "exhibition") return canUseExhibition;
    return canUseInventory;
  });
}

function sidebarPrefsKey(username: string) {
  return `${sidebarPrefsBaseKey}:${username || "local"}`;
}

function readSidebarPrefs(username: string): SidebarPrefs {
  const fallback: SidebarPrefs = { order: [], hidden: [] };
  try {
    const stored = JSON.parse(window.localStorage.getItem(sidebarPrefsKey(username)) || "null") as Partial<SidebarPrefs> | null;
    if (stored) {
      return {
        order: Array.isArray(stored.order) ? stored.order.filter((view): view is AppView => navItems.some((item) => item.view === view)) : [],
        hidden: Array.isArray(stored.hidden) ? stored.hidden.filter((view): view is AppView => navItems.some((item) => item.view === view) && view !== "settings") : []
      };
    }
  } catch {
    return fallback;
  }
  try {
    const legacyOrder = JSON.parse(window.localStorage.getItem(sidebarOrderKey) || "[]") as AppView[];
    return { order: legacyOrder.filter((view): view is AppView => navItems.some((item) => item.view === view)), hidden: [] };
  } catch {
    return fallback;
  }
}

function readNavItems(roles: string[], username: string) {
  const available = allowedNavItems(roles);
  const prefs = readSidebarPrefs(username);
  const visible = available.filter((item) => item.view === "settings" || !prefs.hidden.includes(item.view));
  try {
    const ordered = prefs.order
      .map((view) => visible.find((item) => item.view === view))
      .filter((item): item is (typeof navItems)[number] => Boolean(item));
    const missing = visible.filter((item) => !ordered.some((orderedItem) => orderedItem.view === item.view));
    return [...ordered, ...missing];
  } catch {
    return visible;
  }
}

export function Shell({
  children,
  username,
  roles,
  activeView,
  onLogout
}: {
  children: ReactNode;
  username: string;
  roles: string[];
  activeView: AppView;
  onLogout: () => void;
}) {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(readSidebarCollapsed);
  const [theme, setTheme] = useState<ThemePreference>(readThemePreference);
  const [orderedNavItems, setOrderedNavItems] = useState(() => readNavItems(roles, username));
  const ThemeIcon = theme === "dark" ? Moon : theme === "light" ? Sun : Monitor;
  const { t } = useI18n();

  useEffect(() => {
    const syncOrder = () => setOrderedNavItems(readNavItems(roles, username));

    window.addEventListener(sidebarOrderChangedEvent, syncOrder);
    window.addEventListener("storage", syncOrder);
    return () => {
      window.removeEventListener(sidebarOrderChangedEvent, syncOrder);
      window.removeEventListener("storage", syncOrder);
    };
  }, [roles, username]);

  function toggleTheme() {
    const nextTheme: ThemePreference = theme === "dark" ? "light" : theme === "light" ? "system" : "dark";
    setTheme(nextTheme);
    applyThemePreference(nextTheme);
  }

  function toggleSidebarCollapsed() {
    setSidebarCollapsed((collapsed) => {
      const next = !collapsed;
      window.localStorage.setItem(sidebarCollapsedKey, String(next));
      return next;
    });
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
          <img className="brand-logo" src="/brand/railkeeper-logo.png" alt="RailKeeper2" />
          <img className="brand-mark" src="/brand/railkeeper-mark.png" alt="RailKeeper2" />
        </div>

        <nav id="main-navigation" className="nav" aria-label={t("nav.main")}>
          {orderedNavItems.map((item) => {
            const Icon = item.icon;
            return (
              <a key={item.view} className={activeView === item.view ? "active" : ""} href={item.href} onClick={() => setMobileMenuOpen(false)}>
                <Icon size={16} aria-hidden="true" />
                <span>{t(item.labelKey)}</span>
              </a>
            );
          })}

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
              <a href="/settings" title={t("nav.system")} aria-label={t("nav.system")}>
                <Settings size={17} aria-hidden="true" />
              </a>
              <a href="https://github.com/ichwars/RailKeeper2" target="_blank" rel="noreferrer" title={t("nav.repository")} aria-label={t("nav.repository")}>
                <GitHubMark size={17} />
              </a>
              <button type="button" onClick={toggleTheme} title={t("nav.theme")} aria-label={t("nav.theme")}>
                <ThemeIcon size={17} aria-hidden="true" />
              </button>
              <button type="button" onClick={onLogout} title={t("nav.logout", { username })} aria-label={t("nav.logout", { username })}>
                <LogOut size={17} aria-hidden="true" />
              </button>
            </div>
            <span className="sidebar-version">v0.1.7.1</span>
          </div>
        </nav>
      </aside>

      <main className="main">{children}</main>

      <a className="feedback-button" href="https://github.com/ichwars/RailKeeper2/issues/new" target="_blank" rel="noreferrer" title={t("nav.feedback")} aria-label={t("nav.feedback")}>
        <Bug size={20} aria-hidden="true" />
      </a>
    </div>
  );
}
