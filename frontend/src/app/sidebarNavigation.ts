import { BarChart3, Box, Boxes, CalendarDays, CircuitBoard, FileInput, Settings } from "lucide-react";
import type { SidebarPrefs } from "./sidebarPreferences";

const navItems = [
  { view: "overview", href: "/overview", labelKey: "nav.overview", icon: BarChart3 },
  { view: "vehicles", href: "/vehicles", labelKey: "nav.vehicles", icon: Box },
  { view: "accessories", href: "/accessories", labelKey: "nav.accessories", icon: Boxes },
  { view: "exhibition", href: "/exhibition", labelKey: "nav.exhibition", icon: CalendarDays },
  { view: "importExport", href: "/import-export", labelKey: "nav.importExport", icon: FileInput },
  { view: "digitalCenters", href: "/digital-centers", labelKey: "nav.digitalCenters", icon: CircuitBoard },
  { view: "settings", href: "/settings", labelKey: "nav.settings", icon: Settings }
] as const;

export type SidebarItem = (typeof navItems)[number];

export function sidebarNavigation(roles: readonly string[], prefs: SidebarPrefs) {
  const admin = roles.includes("Admin");
  const inventory = roles.includes("Editor") || roles.includes("Viewer");
  const exhibition = roles.includes("Messe");
  const available = navItems.filter((item) => {
    if (admin) return true;
    if (item.view === "exhibition") return exhibition;
    if (item.view === "importExport") return inventory || exhibition;
    return inventory;
  });
  const ordered = prefs.order.flatMap((view) => available.filter((item) => item.view === view));
  const items = [...ordered, ...available.filter((item) => !ordered.includes(item))];
  return {
    visible: items.filter((item) => item.view === "settings" || !prefs.hidden.includes(item.view)),
    additional: items.filter((item) => item.view !== "settings" && prefs.hidden.includes(item.view))
  };
}
