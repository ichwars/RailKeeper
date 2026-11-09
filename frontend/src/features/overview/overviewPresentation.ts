import type { LucideIcon } from "lucide-react";
import {
  type DigitalCenterSettings
} from "../../shared/api";
import {
  type OverviewMetricID
} from "./overviewModel";

export const providerLabels: Record<string, string> = {
  ecos: "ESU ECoS",
  z21: "Roco Z21",
  intellibox3: "Uhlenbrock Intellibox 3",
  cs3: "Märklin CS3"
};

export const providerImages: Record<DigitalCenterSettings["provider"], string> = {
  ecos: "/assets/overview-ecos-50220.png",
  z21: "/assets/overview-z21-10820.png",
  intellibox3: "/assets/overview-intellibox-3-65300.jpg",
  cs3: "/assets/overview-cs3-60226.jpg"
};

export function sameDay(left: Date, right: Date) {
  return left.getFullYear() === right.getFullYear() && left.getMonth() === right.getMonth() &&
    left.getDate() === right.getDate();
}

export function formatRecentTime(value: string, language: string) {
  const date = new Date(value);
  const today = new Date();
  const yesterday = new Date(today);
  yesterday.setDate(today.getDate() - 1);
  const time = new Intl.DateTimeFormat(language === "en" ? "en-GB" : "de-DE", {
    hour: "2-digit",
    minute: "2-digit"
  }).format(date);
  if (sameDay(date, today)) return `${language === "en" ? "Today" : "Heute"}, ${time}`;
  if (sameDay(date, yesterday)) return `${language === "en" ? "Yesterday" : "Gestern"}, ${time}`;
  return new Intl.DateTimeFormat(language === "en" ? "en-GB" : "de-DE", {
    day: "2-digit",
    month: "2-digit",
    year: "2-digit"
  }).format(date);
}

export function maintenanceDistanceText(
  days: number,
  t: (key: string, values?: Record<string, string | number>) => string
) {
  if (days < 0) return t("overview.daysOverdue", { days: Math.abs(days) });
  if (days === 0) return t("overview.dueToday");
  if (days === 1) return t("overview.dueTomorrow");
  return t("overview.dueInDays", { days });
}

export function moveMetric(order: OverviewMetricID[], metric: OverviewMetricID, target: OverviewMetricID) {
  const sourceIndex = order.indexOf(metric);
  const targetIndex = order.indexOf(target);
  if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex) return order;
  const next = [...order];
  next.splice(sourceIndex, 1);
  next.splice(targetIndex, 0, metric);
  return next;
}

export function providerEnabled(settings: DigitalCenterSettings | null) {
  if (!settings) return false;
  return Boolean(settings[settings.provider]?.enabled && settings[settings.provider]?.host.trim());
}

export type OverviewViewProps = {
  username?: string;
  roles?: string[];
};

export type MetricCard = {
  icon: LucideIcon;
  value: string;
  detail: React.ReactNode;
  attention?: boolean;
  href?: string;
};
