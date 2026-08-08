import { Check, ChevronDown } from "lucide-react";
import {
  FocusEvent,
  forwardRef,
  Fragment,
  KeyboardEvent,
  ReactNode,
  useEffect,
  useId,
  useRef,
  useState
} from "react";

export type AppMultiSelectOption = {
  value: string;
  label: ReactNode;
  disabled?: boolean;
};

export type AppMultiSelectProps = {
  id?: string;
  name?: string;
  className?: string;
  label: ReactNode;
  helpText?: ReactNode;
  error?: ReactNode;
  options: AppMultiSelectOption[];
  value: string[];
  onValueChange?: (value: string[]) => void;
  disabled?: boolean;
  readOnly?: boolean;
  required?: boolean;
  placeholder: string;
  "aria-describedby"?: string;
};

export const AppMultiSelect = forwardRef<HTMLButtonElement, AppMultiSelectProps>(function AppMultiSelect(
  {
    id,
    name,
    className = "",
    label,
    helpText,
    error,
    options,
    value,
    onValueChange,
    disabled = false,
    readOnly = false,
    required = false,
    placeholder,
    "aria-describedby": externalDescription
  },
  forwardedRef
) {
  const generatedId = useId();
  const inputId = id || generatedId;
  const labelId = `${inputId}-label`;
  const valueId = `${inputId}-value`;
  const listboxId = `${inputId}-listbox`;
  const helpId = helpText ? `${inputId}-help` : undefined;
  const errorId = error ? `${inputId}-error` : undefined;
  const describedBy = [externalDescription, helpId, errorId].filter(Boolean).join(" ") || undefined;
  const rootRef = useRef<HTMLDivElement | null>(null);
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const optionRefs = useRef<Array<HTMLButtonElement | null>>([]);
  const firstEnabled = options.findIndex((option) => !option.disabled);
  const lastEnabled = options
    .map((option, index) => ({ option, index }))
    .reverse()
    .find(({ option }) => !option.disabled)?.index ?? -1;
  const selectedEnabled = options.findIndex((option) => value.includes(option.value) && !option.disabled);
  const initialActive = selectedEnabled >= 0 ? selectedEnabled : Math.max(0, firstEnabled);
  const [open, setOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(initialActive);
  const selectedOptions = options.filter((option) => value.includes(option.value));
  const invalid = Boolean(error) || Boolean(required && value.length === 0);

  useEffect(() => {
    if (!open) return;
    optionRefs.current[activeIndex]?.focus();
  }, [activeIndex, open]);

  useEffect(() => {
    if (!open) return;
    const closeOutside = (event: PointerEvent) => {
      if (event.target instanceof Node && rootRef.current?.contains(event.target)) return;
      setOpen(false);
    };
    document.addEventListener("pointerdown", closeOutside);
    return () => document.removeEventListener("pointerdown", closeOutside);
  }, [open]);

  const setTriggerRef = (node: HTMLButtonElement | null) => {
    triggerRef.current = node;
    if (typeof forwardedRef === "function") forwardedRef(node);
    else if (forwardedRef) forwardedRef.current = node;
  };

  const openList = (fromEnd = false) => {
    if (disabled || readOnly || firstEnabled < 0) return;
    setActiveIndex(fromEnd ? lastEnabled : initialActive);
    setOpen(true);
  };

  const closeAndFocusTrigger = () => {
    setOpen(false);
    triggerRef.current?.focus();
  };

  const toggle = (option: AppMultiSelectOption) => {
    if (disabled || readOnly || option.disabled) return;
    const next = value.includes(option.value)
      ? value.filter((selectedValue) => selectedValue !== option.value)
      : [...value, option.value];
    onValueChange?.(next);
  };

  const moveActive = (direction: 1 | -1) => {
    if (options.length === 0) return;
    let next = activeIndex;
    for (let count = 0; count < options.length; count += 1) {
      next = (next + direction + options.length) % options.length;
      if (!options[next]?.disabled) {
        setActiveIndex(next);
        return;
      }
    }
  };

  const handleTriggerKeyboard = (event: KeyboardEvent<HTMLButtonElement>) => {
    if (event.key === "ArrowDown" || event.key === "ArrowUp") {
      event.preventDefault();
      openList(event.key === "ArrowUp");
    }
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      openList();
    }
    if (event.key === "Escape") {
      event.preventDefault();
      setOpen(false);
    }
  };

  const handleOptionKeyboard = (event: KeyboardEvent<HTMLButtonElement>, option: AppMultiSelectOption) => {
    if (event.key === "ArrowDown" || event.key === "ArrowUp") {
      event.preventDefault();
      moveActive(event.key === "ArrowDown" ? 1 : -1);
    }
    if (event.key === "Home" || event.key === "End") {
      event.preventDefault();
      const boundary = event.key === "Home" ? firstEnabled : lastEnabled;
      if (boundary >= 0) setActiveIndex(boundary);
    }
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      toggle(option);
    }
    if (event.key === "Escape") {
      event.preventDefault();
      closeAndFocusTrigger();
    }
    if (event.key === "Tab") window.setTimeout(() => setOpen(false), 0);
  };

  const handleFocusLeave = (event: FocusEvent<HTMLDivElement>) => {
    if (event.relatedTarget instanceof Node && event.currentTarget.contains(event.relatedTarget)) return;
    setOpen(false);
  };

  return (
    <div
      ref={rootRef}
      className={`app-field app-multi-select ${invalid ? "has-error" : ""} ${className}`.trim()}
      onBlur={handleFocusLeave}
    >
      <span className="app-field-label">
        <span id={labelId}>{label}</span>{required ? <span aria-hidden="true"> *</span> : null}
      </span>
      <div className="app-multi-select-control">
        <button
          ref={setTriggerRef}
          id={inputId}
          type="button"
          className="app-multi-select-trigger"
          disabled={disabled}
          aria-labelledby={`${labelId} ${valueId}`}
          aria-describedby={describedBy}
          aria-controls={listboxId}
          aria-expanded={open}
          aria-haspopup="listbox"
          aria-invalid={invalid || undefined}
          aria-readonly={readOnly || undefined}
          aria-required={required || undefined}
          onClick={() => open ? setOpen(false) : openList()}
          onKeyDown={handleTriggerKeyboard}
        >
          <span id={valueId} className={selectedOptions.length > 0 ? "" : "empty"}>
            {selectedOptions.length > 0
              ? selectedOptions.map((option, index) => (
                  <Fragment key={option.value}>{index > 0 ? ", " : null}{option.label}</Fragment>
                ))
              : placeholder}
          </span>
          <ChevronDown size={15} aria-hidden="true" />
        </button>
        {open ? (
          <div
            id={listboxId}
            className="app-multi-select-options"
            role="listbox"
            aria-labelledby={labelId}
            aria-multiselectable="true"
          >
            {options.map((option, index) => {
              const selected = value.includes(option.value);
              return (
                <button
                  ref={(node) => { optionRefs.current[index] = node; }}
                  key={option.value}
                  type="button"
                  className={`app-multi-select-option ${selected ? "selected" : ""}`.trim()}
                  role="option"
                  aria-selected={selected}
                  aria-disabled={option.disabled || undefined}
                  disabled={option.disabled}
                  tabIndex={index === activeIndex && !option.disabled ? 0 : -1}
                  onFocus={() => setActiveIndex(index)}
                  onClick={() => toggle(option)}
                  onKeyDown={(event) => handleOptionKeyboard(event, option)}
                >
                  <span>{option.label}</span>
                  {selected ? <Check size={14} aria-hidden="true" /> : null}
                </button>
              );
            })}
          </div>
        ) : null}
      </div>
      <select
        className="visually-hidden"
        name={name}
        multiple
        value={value}
        required={required && value.length === 0}
        disabled={disabled}
        tabIndex={-1}
        aria-hidden="true"
        onChange={() => undefined}
        onInvalid={(event) => {
          event.preventDefault();
          triggerRef.current?.focus();
        }}
      >
        {options.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
      </select>
      {helpText ? <span id={helpId} className="app-field-help">{helpText}</span> : null}
      {error ? <span id={errorId} className="app-field-error" role="alert">{error}</span> : null}
    </div>
  );
});
