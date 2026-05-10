import type { ReactNode } from "react";
import { useState } from "react";
import { BarChart3, Box, Code2, FileInput, LogOut, Menu, Settings, X } from "lucide-react";
import type { AppView } from "./App";

const navItems = [
  { view: "overview", href: "/overview", label: "Übersicht", icon: BarChart3 },
  { view: "vehicles", href: "/", label: "Bestand", icon: Box },
  { view: "importExport", href: "/import-export", label: "Import/Export", icon: FileInput },
  { view: "settings", href: "/settings", label: "Einstellungen", icon: Settings }
] as const;

export function Shell({
  children,
  username,
  activeView,
  onLogout
}: {
  children: ReactNode;
  username: string;
  activeView: AppView;
  onLogout: () => void;
}) {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <div className={mobileMenuOpen ? "layout nav-open" : "layout"}>
      {mobileMenuOpen && <button type="button" className="mobile-nav-scrim" aria-label="Menü schließen" onClick={() => setMobileMenuOpen(false)} />}
      <aside className="sidebar">
        <button type="button" className="mobile-menu-button" aria-label={mobileMenuOpen ? "Menü schließen" : "Menü öffnen"} onClick={() => setMobileMenuOpen((open) => !open)}>
          {mobileMenuOpen ? <X size={19} aria-hidden="true" /> : <Menu size={19} aria-hidden="true" />}
        </button>
        <div className="brand">
          <img src="/brand/railkeeper-mark.svg" alt="" />
          <span>RailKeeper2</span>
        </div>

        <nav className="nav" aria-label="Hauptnavigation">
          {navItems.map((item) => {
            const Icon = item.icon;
            return (
              <a key={item.view} className={activeView === item.view ? "active" : ""} href={item.href} onClick={() => setMobileMenuOpen(false)}>
                <Icon size={16} aria-hidden="true" />
                {item.label}
              </a>
            );
          })}
          <a className="repo-link mobile-repo-link" href="https://github.com/ichwars/RailKeeper2" target="_blank" rel="noreferrer">
            <Code2 size={16} aria-hidden="true" />
            Repository
          </a>
        </nav>

        <a className="repo-link desktop-repo-link" href="https://github.com/ichwars/RailKeeper2" target="_blank" rel="noreferrer">
          <Code2 size={16} aria-hidden="true" />
          Repository
        </a>

        <div className="user-block">
          <span>{username}</span>
          <button onClick={onLogout} title="Abmelden" aria-label="Abmelden">
            <LogOut size={16} aria-hidden="true" />
          </button>
        </div>
      </aside>

      <main className="main">{children}</main>
    </div>
  );
}
