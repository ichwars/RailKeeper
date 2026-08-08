import { Check, ChevronDown } from "lucide-react";
import { Fragment, forwardRef, KeyboardEvent, ReactNode, useId, useState } from "react";

export type AppMultiSelectOption = {
  value: string;
  label: ReactNode;
  disabled?: boolean;
};

export type AppMultiSelectProps = {
  id?: string;
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
  placeholder?: string;
  "aria-describedby"?: string;
};

export const AppMultiSelect = forwardRef<HTMLButtonElement, AppMultiSelectProps>(function AppMultiSelect(
  {
    id,
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
    placeholder = "Auswählen",
    "aria-describedby": externalDescription
  },
  ref
) {
  const generatedId = useId();
  const inputId = id || generatedId;
  const labelId = `${inputId}-label`;
  const listboxId = `${inputId}-listbox`;
  const helpId = helpText ? `${inputId}-help` : undefined;
  const errorId = error ? `${inputId}-error` : undefined;
  const describedBy = [externalDescription, helpId, errorId].filter(Boolean).join(" ") || undefined;
  const [open, setOpen] = useState(false);
  const selectedOptions = options.filter((option) => value.includes(option.value));

  const toggle = (option: AppMultiSelectOption) => {
    if (disabled || readOnly || option.disabled) return;
    const next = value.includes(option.value)
      ? value.filter((selectedValue) => selectedValue !== option.value)
      : [...value, option.value];
    onValueChange?.(next);
  };

  const handleKeyboard = (event: KeyboardEvent<HTMLButtonElement>) => {
    if (event.key === "ArrowDown" || event.key === "ArrowUp") {
      event.preventDefault();
      if (!disabled) setOpen(true);
    }
    if (event.key === "Escape") {
      event.preventDefault();
      setOpen(false);
    }
  };

  return (
    <div className={`app-field app-multi-select ${error ? "has-error" : ""} ${className}`.trim()}>
      <span id={labelId} className="app-field-label">
        {label}{required ? <span aria-hidden="true"> *</span> : null}
      </span>
      <div className="app-multi-select-control">
        <button
          ref={ref}
          id={inputId}
          type="button"
          className="app-multi-select-trigger"
          disabled={disabled}
          aria-labelledby={labelId}
          aria-describedby={describedBy}
          aria-controls={listboxId}
          aria-expanded={open}
          aria-haspopup="listbox"
          aria-invalid={error ? true : undefined}
          aria-readonly={readOnly || undefined}
          aria-required={required || undefined}
          onClick={() => setOpen((current) => !current)}
          onKeyDown={handleKeyboard}
        >
          <span className={selectedOptions.length > 0 ? "" : "empty"}>
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
            {options.map((option) => {
              const selected = value.includes(option.value);
              return (
                <button
                  key={option.value}
                  type="button"
                  className={`app-multi-select-option ${selected ? "selected" : ""}`.trim()}
                  role="option"
                  aria-selected={selected}
                  disabled={disabled || option.disabled}
                  onClick={() => toggle(option)}
                >
                  <span>{option.label}</span>
                  {selected ? <Check size={14} aria-hidden="true" /> : null}
                </button>
              );
            })}
          </div>
        ) : null}
      </div>
      {helpText ? <span id={helpId} className="app-field-help">{helpText}</span> : null}
      {error ? <span id={errorId} className="app-field-error" role="alert">{error}</span> : null}
    </div>
  );
});
