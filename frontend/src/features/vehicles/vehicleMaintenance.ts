import { VehicleMaintenance } from "../../shared/api";
import { formatDate } from "./vehicleFormat";

export function maintenanceOptionLabel(entry: VehicleMaintenance) {
  const due = entry.dueDate ? ` · ${formatDate(entry.dueDate)}` : "";
  const notes = entry.notes ? ` · ${entry.notes}` : "";
  return `${entry.kind}${due}${notes}`;
}

export function maintenanceIsDue(entry: VehicleMaintenance) {
  if (!entry.dueDate || entry.status === "erledigt") return false;
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const due = new Date(`${entry.dueDate}T00:00:00`);
  return !Number.isNaN(due.getTime()) && due <= today;
}

export function maintenanceDaysUntilDue(entry: VehicleMaintenance) {
  if (!entry.dueDate || entry.status === "erledigt") return null;
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const due = new Date(`${entry.dueDate}T00:00:00`);
  if (Number.isNaN(due.getTime())) return null;
  return Math.ceil((due.getTime() - today.getTime()) / 86400000);
}

export function maintenanceReminderText(daysUntilDue: number) {
  if (daysUntilDue < 0) return `seit ${Math.abs(daysUntilDue)} Tag${Math.abs(daysUntilDue) === 1 ? "" : "en"} überfällig`;
  if (daysUntilDue === 0) return "heute fällig";
  if (daysUntilDue === 1) return "morgen fällig";
  return `in ${daysUntilDue} Tagen fällig`;
}

export function normalizeMaintenanceStatus(status: string) {
  return status === "fällig" ? "faellig" : status;
}

export function maintenanceStatusClass(status: string) {
  return normalizeMaintenanceStatus(status);
}

export function todayISODate() {
  const today = new Date();
  today.setMinutes(today.getMinutes() - today.getTimezoneOffset());
  return today.toISOString().slice(0, 10);
}

export function formatMaintenanceCost(cost?: string) {
  if (!cost) return "-";
  const value = Number(cost.replace(",", "."));
  if (Number.isNaN(value)) return cost;
  return new Intl.NumberFormat("de-DE", { style: "currency", currency: "EUR" }).format(value);
}
