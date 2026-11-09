import { ReactNode, RefObject, useRef } from "react";
import {
  ExhibitionConflict,
  ExhibitionEntryInput,
  ExhibitionList,
  ExhibitionListInput,
  ExhibitionStatus
} from "../../shared/api";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";

export function ExhibitionDialogLayer({
  children,
  onClose
}: {
  children: (closeRef: RefObject<HTMLButtonElement | null>) => ReactNode;
  onClose: () => void;
}) {
  const closeRef = useRef<HTMLButtonElement | null>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(onClose, closeRef);
  return (
    <div ref={layerRef} className="modal-layer" role="presentation" onKeyDown={onKeyDown}>
      <span ref={anchorRef} hidden />
      {children(closeRef)}
    </div>
  );
}

export const statusLabels: Record<ExhibitionStatus, { de: string; en: string }> = {
  draft: { de: "Entwurf", en: "Draft" },
  open: { de: "Offen", en: "Open" },
  locked: { de: "Gesperrt", en: "Locked" },
  running: { de: "Läuft", en: "Running" },
  completed: { de: "Abgeschlossen", en: "Completed" },
  archived: { de: "Archiviert", en: "Archived" }
};

export const entryStatusLabels = {
  ready: { de: "Bereit", en: "Ready" },
  addressConflict: { de: "Adresskonflikt", en: "Address conflict" },
  missing: { de: "Angaben fehlen", en: "Missing data" },
  check: { de: "Prüfen", en: "Check" },
  unavailable: { de: "Nicht verfügbar", en: "Unavailable" }
};

export function localeDate(value: string, language: "de" | "en", options?: Intl.DateTimeFormatOptions) {
  if (!value) return "–";
  return new Intl.DateTimeFormat(language === "de" ? "de-DE" : "en-GB", options ?? {
    day: "2-digit", month: "long", year: "numeric"
  }).format(new Date(`${value}T12:00:00`));
}

export function eventDateRange(event: ExhibitionList, language: "de" | "en") {
  const end = event.endDate || event.date;
  if (event.date === end) return localeDate(event.date, language);
  const startDate = new Date(`${event.date}T12:00:00`);
  const endDate = new Date(`${end}T12:00:00`);
  if (startDate.getMonth() === endDate.getMonth() && startDate.getFullYear() === endDate.getFullYear()) {
    return `${startDate.getDate()}.–${localeDate(end, language)}`;
  }
  return `${localeDate(event.date, language)} – ${localeDate(end, language)}`;
}

export function relativeDate(event: ExhibitionList, language: "de" | "en") {
  const today = new Date();
  today.setHours(12, 0, 0, 0);
  const start = new Date(`${event.date}T12:00:00`);
  const days = Math.round((start.getTime() - today.getTime()) / 86_400_000);
  if (days === 0) return language === "de" ? "Heute" : "Today";
  if (days > 0) return language === "de" ? `In ${days} Tagen` : `In ${days} days`;
  return language === "de" ? `Vor ${Math.abs(days)} Tagen` : `${Math.abs(days)} days ago`;
}

export function eventDayDate(event: ExhibitionList, index: number, language: "de" | "en") {
  const date = new Date(`${event.date}T12:00:00`);
  date.setDate(date.getDate() + index);
  return new Intl.DateTimeFormat(language === "de" ? "de-DE" : "en-GB", {
    day: "2-digit", month: "2-digit"
  }).format(date);
}

export function dayScopeIncludes(value: string, scope: string) {
  return value === "all" || value.split(",").map((part) => part.trim()).includes(scope);
}

export function hasAdmin(roles: string[]) {
  return roles.includes("Admin");
}

export function canReadInventory(roles: string[]) {
  return roles.some((role) => ["Admin", "Editor", "Viewer"].includes(role));
}

export function newEventInput(): ExhibitionListInput {
  const date = new Date().toISOString().slice(0, 10);
  return { designation: "", date, endDate: date, location: "", description: "", organizationNotes: "", status: "open" };
}

export function newEntryInput(): ExhibitionEntryInput {
  return {
    owner: "", locomotiveName: "", dayScope: "all", dtDecoder: true, decoderNumber: "",
    interfaceName: "", adapter: "", analog: false, availability: "available", imageUrl: "",
    functionKeys: "", notes: ""
  };
}

export function conflictDescription(conflict: ExhibitionConflict, language: "de" | "en") {
  if (conflict.kind === "address") {
    return language === "de"
      ? `Adresse ${conflict.address || "–"} auf ${conflict.interfaceName || "–"} überschneidet sich.`
      : `Address ${conflict.address || "–"} on ${conflict.interfaceName || "–"} overlaps.`;
  }
  if (conflict.kind === "duplicateVehicle") {
    return language === "de" ? "Fahrzeug ist mehrfach eingetragen." : "Vehicle is entered more than once.";
  }
  return language === "de"
    ? `Pflichtangaben fehlen: ${(conflict.fields || []).join(", ")}`
    : `Required data missing: ${(conflict.fields || []).join(", ")}`;
}
