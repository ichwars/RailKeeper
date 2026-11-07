import { MoreHorizontal } from "lucide-react";
import type { AppView } from "./App";
import type { SidebarItem } from "./sidebarNavigation";
import { useI18n } from "../shared/i18n";
import { followAppLink } from "./appNavigation";

export function SidebarAdditionalNavigation({ items, activeView, onNavigate, onSelect }: {
  items: SidebarItem[]; activeView: AppView; onNavigate: () => void; onSelect?: (view: AppView) => void;
}) {
  const { t } = useI18n();
  if (!items.length) return null;
  return <details className="nav-more" open={items.some((item) => item.view === activeView) || undefined}>
    <summary className="nav-entry" title={t("nav.moreFunctions")}>
      <MoreHorizontal size={16} aria-hidden="true" /><span>{t("nav.moreFunctions")}</span>
    </summary>
    {items.map((item) => {
      const Icon = item.icon;
      return <a key={item.view} href={item.href} title={t(item.labelKey)} onClick={event => {
        onNavigate();
        followAppLink(event, onSelect ? () => onSelect(item.view) : undefined);
      }}
        className={`nav-entry${activeView === item.view ? " active" : ""}`}
        aria-current={activeView === item.view ? "page" : undefined}>
        <Icon size={16} aria-hidden="true" /><span>{t(item.labelKey)}</span>
      </a>;
    })}
  </details>;
}
