import { Check, ChevronDown } from "lucide-react";
import { ButtonHTMLAttributes, CSSProperties, Children, Fragment, KeyboardEvent, ReactNode, isValidElement, useEffect, useId, useLayoutEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";

export type AppSelectChangeEvent = { target: { value: string } };

type OptionModel = {
  value: string;
  label: ReactNode;
  disabled: boolean;
};

type OptionProps = {
  value?: string | number;
  disabled?: boolean;
  children?: ReactNode;
};

type AppSelectProps = Omit<ButtonHTMLAttributes<HTMLButtonElement>, "children" | "onChange" | "value"> & {
  value: string;
  children: ReactNode;
  onChange?: (event: AppSelectChangeEvent) => void;
  required?: boolean;
};

export function AppSelect({ value, children, onChange, className = "", disabled, required, ...buttonProps }: AppSelectProps) {
  const listID = useId();
  const rootRef = useRef<HTMLDivElement | null>(null);
  const listRef = useRef<HTMLDivElement | null>(null);
  const typeaheadRef = useRef("");
  const typeaheadTimerRef = useRef<number | null>(null);
  const [open, setOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);
  const [floatingStyle, setFloatingStyle] = useState<CSSProperties>({});

  const options = useMemo<OptionModel[]>(() => {
    const collectOptions = (node: ReactNode): OptionModel[] => {
      return Children.toArray(node).flatMap((child) => {
        if (!isValidElement<OptionProps>(child)) return [];
        if (child.type === Fragment) return collectOptions(child.props.children);
        if (child.type !== "option") return [];
        return [{
          value: String(child.props.value ?? child.props.children ?? ""),
          label: child.props.children,
          disabled: Boolean(child.props.disabled)
        }];
      });
    };

    return collectOptions(children);
  }, [children]);

  const currentIndex = Math.max(0, options.findIndex((option) => option.value === String(value)));
  const selectedOption = options[currentIndex] || options[0];

  const optionLabelText = (label: ReactNode): string => {
    if (typeof label === "string" || typeof label === "number") return String(label);
    if (Array.isArray(label)) return label.map(optionLabelText).join(" ");
    if (isValidElement<OptionProps>(label)) return optionLabelText(label.props.children);
    return "";
  };

  useEffect(() => {
    if (!open) return;
    const close = (event: PointerEvent) => {
      const target = event.target as Node;
      if (rootRef.current?.contains(target) || listRef.current?.contains(target)) return;
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
      const optionGap = 4;
      const minWidth = 220;
      const spaceBelow = window.innerHeight - rect.bottom - viewportGap - optionGap;
      const spaceAbove = rect.top - viewportGap - optionGap;
      const opensAbove = spaceBelow < 180 && spaceAbove > spaceBelow;
      const availableHeight = opensAbove ? spaceAbove : spaceBelow;
      const maxHeight = Math.max(140, Math.min(260, availableHeight));
      setFloatingStyle({
        position: "fixed",
        top: opensAbove ? Math.max(viewportGap, rect.top - maxHeight - optionGap) : rect.bottom + optionGap,
        left: Math.max(viewportGap, Math.min(rect.left, window.innerWidth - Math.max(rect.width, minWidth) - viewportGap)),
        width: Math.max(rect.width, minWidth),
        maxHeight,
        zIndex: 130
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

  useEffect(() => {
    setActiveIndex(currentIndex);
  }, [currentIndex]);

  useEffect(() => {
    return () => {
      if (typeaheadTimerRef.current !== null) {
        window.clearTimeout(typeaheadTimerRef.current);
      }
    };
  }, []);

  const choose = (option: OptionModel) => {
    if (option.disabled || disabled) return;
    onChange?.({ target: { value: option.value } });
    setOpen(false);
  };

  const move = (direction: 1 | -1) => {
    if (options.length === 0) return;
    let next = activeIndex;
    for (let step = 0; step < options.length; step += 1) {
      next = (next + direction + options.length) % options.length;
      if (!options[next]?.disabled) break;
    }
    setActiveIndex(next);
  };

  const findTypeaheadMatch = (query: string) => {
    const normalizedQuery = query.trim().toLocaleLowerCase("de-DE");
    if (!normalizedQuery) return -1;
    const candidates = options.map((option, index) => ({
      index,
      option,
      label: optionLabelText(option.label).trim().toLocaleLowerCase("de-DE")
    }));
    const ordered = [...candidates.slice(activeIndex + 1), ...candidates.slice(0, activeIndex + 1)];
    return (
      ordered.find((candidate) => !candidate.option.disabled && candidate.label.startsWith(normalizedQuery)) ||
      ordered.find((candidate) => !candidate.option.disabled && candidate.label.includes(normalizedQuery))
    )?.index ?? -1;
  };

  const applyTypeahead = (key: string) => {
    if (disabled || options.length === 0) return;
    if (typeaheadTimerRef.current !== null) {
      window.clearTimeout(typeaheadTimerRef.current);
    }
    typeaheadRef.current = `${typeaheadRef.current}${key}`.slice(-30);
    typeaheadTimerRef.current = window.setTimeout(() => {
      typeaheadRef.current = "";
      typeaheadTimerRef.current = null;
    }, 700);
    const repeatedCharacterSearch = typeaheadRef.current.length > 1 && new Set(typeaheadRef.current).size === 1;
    const query = repeatedCharacterSearch ? key : typeaheadRef.current;
    const matchIndex = findTypeaheadMatch(query);
    if (matchIndex >= 0) {
      setActiveIndex(matchIndex);
      choose(options[matchIndex]);
    }
  };

  const handleKeyboard = (event: KeyboardEvent) => {
    if (event.key === "ArrowDown") {
      event.preventDefault();
      if (!open && options.length > 0) setOpen(true);
      move(1);
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      if (!open && options.length > 0) setOpen(true);
      move(-1);
    }
    if (event.key === "Home") {
      event.preventDefault();
      const first = options.findIndex((option) => !option.disabled);
      if (first >= 0) {
        setOpen(true);
        setActiveIndex(first);
      }
    }
    if (event.key === "End") {
      event.preventDefault();
      const last = options.map((option, index) => ({ option, index })).reverse().find((item) => !item.option.disabled)?.index ?? -1;
      if (last >= 0) {
        setOpen(true);
        setActiveIndex(last);
      }
    }
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      if (open && options[activeIndex]) choose(options[activeIndex]);
      else if (options.length > 0) setOpen(true);
    }
    if (event.key === "Escape") {
      event.preventDefault();
      setOpen(false);
    }
    if (event.key.length === 1 && !event.altKey && !event.ctrlKey && !event.metaKey && event.key !== " ") {
      event.preventDefault();
      applyTypeahead(event.key);
    }
  };

  return (
    <div ref={rootRef} className={`app-select ${open ? "open" : ""} ${disabled ? "disabled" : ""} ${className}`.trim()}>
      <button
        {...buttonProps}
        type="button"
        className="app-select-trigger"
        disabled={disabled}
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={listID}
        aria-required={required || undefined}
        onClick={() => options.length > 0 && setOpen((current) => !current)}
        onKeyDown={handleKeyboard}
      >
        <span>{selectedOption?.label || "-"}</span>
        <ChevronDown size={15} aria-hidden="true" />
      </button>
      {open && typeof document !== "undefined" && createPortal(
        <div ref={listRef} id={listID} className="app-select-options app-select-options-floating" role="listbox" style={floatingStyle} onKeyDown={handleKeyboard}>
          {options.map((option, index) => {
            const selected = option.value === String(value);
            return (
              <button
                key={`${option.value}-${index}`}
                type="button"
                className={`app-select-option ${selected ? "selected" : ""} ${index === activeIndex ? "active" : ""}`.trim()}
                role="option"
                aria-selected={selected}
                disabled={option.disabled}
                onMouseEnter={() => setActiveIndex(index)}
                onClick={() => choose(option)}
              >
                <span>{option.label}</span>
                {selected && <Check size={14} aria-hidden="true" />}
              </button>
            );
          })}
        </div>,
        document.body
      )}
    </div>
  );
}
