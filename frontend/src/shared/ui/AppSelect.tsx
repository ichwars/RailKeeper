import { Check, ChevronDown } from "lucide-react";
import { ButtonHTMLAttributes, Children, Fragment, ReactNode, isValidElement, useEffect, useId, useMemo, useRef, useState } from "react";

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
  const [open, setOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);

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

  useEffect(() => {
    if (!open) return;
    const close = (event: PointerEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false);
    };
    document.addEventListener("pointerdown", close);
    return () => document.removeEventListener("pointerdown", close);
  }, [open]);

  useEffect(() => {
    setActiveIndex(currentIndex);
  }, [currentIndex]);

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
        onKeyDown={(event) => {
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
          if (event.key === "Enter" || event.key === " ") {
            event.preventDefault();
            if (open && options[activeIndex]) choose(options[activeIndex]);
            else if (options.length > 0) setOpen(true);
          }
          if (event.key === "Escape") setOpen(false);
        }}
      >
        <span>{selectedOption?.label || "-"}</span>
        <ChevronDown size={15} aria-hidden="true" />
      </button>
      {open && (
        <div id={listID} className="app-select-options" role="listbox">
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
        </div>
      )}
    </div>
  );
}
