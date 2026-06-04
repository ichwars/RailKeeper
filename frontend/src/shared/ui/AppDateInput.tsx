import { Calendar, ChevronLeft, ChevronRight, X } from "lucide-react";
import { CSSProperties, InputHTMLAttributes, KeyboardEvent, useEffect, useId, useLayoutEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";

export type AppDateInputChangeEvent = { target: { value: string } };

type AppDateInputProps = Omit<InputHTMLAttributes<HTMLInputElement>, "onChange" | "type" | "value"> & {
  value?: string;
  onChange?: (event: AppDateInputChangeEvent) => void;
};

const dayNames = ["Mo", "Di", "Mi", "Do", "Fr", "Sa", "So"];
const monthNames = [
  "Januar",
  "Februar",
  "Maerz",
  "April",
  "Mai",
  "Juni",
  "Juli",
  "August",
  "September",
  "Oktober",
  "November",
  "Dezember"
];

function toISODate(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function fromISODate(value?: string) {
  if (!value || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return null;
  const [year, month, day] = value.split("-").map(Number);
  const date = new Date(year, month - 1, day);
  if (date.getFullYear() !== year || date.getMonth() !== month - 1 || date.getDate() !== day) return null;
  return date;
}

function displayDate(value?: string) {
  const date = fromISODate(value);
  if (!date) return value || "";
  return `${String(date.getDate()).padStart(2, "0")}.${String(date.getMonth() + 1).padStart(2, "0")}.${date.getFullYear()}`;
}

function parseDateInput(value: string) {
  const trimmed = value.trim();
  if (!trimmed) return "";
  const iso = fromISODate(trimmed);
  if (iso) return toISODate(iso);
  const match = /^(\d{1,2})\.(\d{1,2})\.(\d{4})$/.exec(trimmed);
  if (!match) return null;
  const [, dayText, monthText, yearText] = match;
  const day = Number(dayText);
  const month = Number(monthText);
  const year = Number(yearText);
  const date = new Date(year, month - 1, day);
  if (date.getFullYear() !== year || date.getMonth() !== month - 1 || date.getDate() !== day) return null;
  return toISODate(date);
}

function monthGrid(monthDate: Date) {
  const year = monthDate.getFullYear();
  const month = monthDate.getMonth();
  const first = new Date(year, month, 1);
  const offset = (first.getDay() + 6) % 7;
  const start = new Date(year, month, 1 - offset);
  return Array.from({ length: 42 }, (_, index) => new Date(start.getFullYear(), start.getMonth(), start.getDate() + index));
}

export function AppDateInput({ value = "", onChange, className = "", disabled, required, placeholder = "TT.MM.JJJJ", ...inputProps }: AppDateInputProps) {
  const calendarID = useId();
  const rootRef = useRef<HTMLDivElement | null>(null);
  const popupRef = useRef<HTMLDivElement | null>(null);
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState(displayDate(value));
  const [floatingStyle, setFloatingStyle] = useState<CSSProperties>({});
  const selectedDate = fromISODate(value);
  const [visibleMonth, setVisibleMonth] = useState(() => selectedDate || new Date());

  useEffect(() => {
    setDraft(displayDate(value));
    if (selectedDate) setVisibleMonth(selectedDate);
  }, [selectedDate?.getTime(), value]);

  useEffect(() => {
    if (!open) return;
    const close = (event: PointerEvent) => {
      const target = event.target as Node;
      if (rootRef.current?.contains(target) || popupRef.current?.contains(target)) return;
      setOpen(false);
    };
    document.addEventListener("pointerdown", close);
    return () => document.removeEventListener("pointerdown", close);
  }, [open]);

  useLayoutEffect(() => {
    if (!open) return;
    const updateFloatingPosition = () => {
      const root = rootRef.current;
      if (!root) return;
      const rect = root.getBoundingClientRect();
      const viewportGap = 8;
      const popupWidth = 284;
      const popupHeight = 330;
      const spaceBelow = window.innerHeight - rect.bottom - viewportGap;
      const opensAbove = spaceBelow < popupHeight && rect.top > spaceBelow;
      setFloatingStyle({
        position: "fixed",
        top: opensAbove ? Math.max(viewportGap, rect.top - popupHeight - 4) : rect.bottom + 4,
        left: Math.max(viewportGap, Math.min(rect.left, window.innerWidth - popupWidth - viewportGap)),
        width: popupWidth,
        zIndex: 140
      });
    };
    updateFloatingPosition();
    window.addEventListener("resize", updateFloatingPosition);
    window.addEventListener("scroll", updateFloatingPosition, true);
    return () => {
      window.removeEventListener("resize", updateFloatingPosition);
      window.removeEventListener("scroll", updateFloatingPosition, true);
    };
  }, [open]);

  const days = useMemo(() => monthGrid(visibleMonth), [visibleMonth]);

  const commit = (nextValue: string) => {
    onChange?.({ target: { value: nextValue } });
    setDraft(displayDate(nextValue));
  };

  const commitDraft = () => {
    const parsed = parseDateInput(draft);
    if (parsed === null) {
      setDraft(displayDate(value));
      return;
    }
    commit(parsed);
  };

  const chooseDay = (date: Date) => {
    commit(toISODate(date));
    setVisibleMonth(date);
    setOpen(false);
  };

  const changeMonth = (offset: number) => {
    setVisibleMonth((current) => new Date(current.getFullYear(), current.getMonth() + offset, 1));
  };

  const handleKeyboard = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter") {
      event.preventDefault();
      commitDraft();
      setOpen(false);
    }
    if (event.key === "Escape") {
      event.preventDefault();
      setDraft(displayDate(value));
      setOpen(false);
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setOpen(true);
    }
  };

  return (
    <div ref={rootRef} className={`app-date ${open ? "open" : ""} ${disabled ? "disabled" : ""} ${className}`.trim()}>
      <span className="app-date-field">
        <input
          {...inputProps}
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          onBlur={commitDraft}
          onKeyDown={handleKeyboard}
          disabled={disabled}
          required={required}
          placeholder={placeholder}
          inputMode="numeric"
          autoComplete="off"
          aria-controls={calendarID}
          aria-expanded={open}
        />
        {value && !disabled && (
          <button type="button" className="app-date-clear" onMouseDown={(event) => event.preventDefault()} onClick={() => commit("")} aria-label="Datum leeren" title="Datum leeren">
            <X size={14} aria-hidden="true" />
          </button>
        )}
        <button type="button" className="app-date-toggle" onMouseDown={(event) => event.preventDefault()} onClick={() => !disabled && setOpen((current) => !current)} disabled={disabled} aria-label="Kalender oeffnen" title="Kalender oeffnen">
          <Calendar size={16} aria-hidden="true" />
        </button>
      </span>
      {open && typeof document !== "undefined" && createPortal(
        <div ref={popupRef} id={calendarID} className="app-date-calendar" style={floatingStyle} role="dialog" aria-label="Datum auswaehlen">
          <div className="app-date-calendar-head">
            <button type="button" onClick={() => changeMonth(-1)} aria-label="Vorheriger Monat" title="Vorheriger Monat">
              <ChevronLeft size={16} aria-hidden="true" />
            </button>
            <strong>{monthNames[visibleMonth.getMonth()]} {visibleMonth.getFullYear()}</strong>
            <button type="button" onClick={() => changeMonth(1)} aria-label="Naechster Monat" title="Naechster Monat">
              <ChevronRight size={16} aria-hidden="true" />
            </button>
          </div>
          <div className="app-date-weekdays" aria-hidden="true">
            {dayNames.map((day) => <span key={day}>{day}</span>)}
          </div>
          <div className="app-date-days">
            {days.map((date) => {
              const iso = toISODate(date);
              const currentMonth = date.getMonth() === visibleMonth.getMonth();
              const selected = iso === value;
              const today = iso === toISODate(new Date());
              return (
                <button
                  key={iso}
                  type="button"
                  className={`${currentMonth ? "" : "muted"} ${selected ? "selected" : ""} ${today ? "today" : ""}`.trim()}
                  onClick={() => chooseDay(date)}
                >
                  {date.getDate()}
                </button>
              );
            })}
          </div>
          <footer className="app-date-calendar-actions">
            <button type="button" onClick={() => { commit(""); setOpen(false); }}>Leeren</button>
            <button type="button" onClick={() => chooseDay(new Date())}>Heute</button>
          </footer>
        </div>,
        document.body
      )}
    </div>
  );
}
