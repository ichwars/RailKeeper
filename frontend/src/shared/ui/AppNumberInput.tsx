import { forwardRef, InputHTMLAttributes, ReactNode, useId } from "react";

export type AppNumberInputProps = Omit<InputHTMLAttributes<HTMLInputElement>, "onChange" | "type" | "value"> & {
  label: ReactNode;
  helpText?: ReactNode;
  error?: ReactNode;
  value: string;
  onValueChange?: (value: string) => void;
};

export const AppNumberInput = forwardRef<HTMLInputElement, AppNumberInputProps>(function AppNumberInput(
  { label, helpText, error, className = "", id, required, value, onValueChange, ...inputProps },
  ref
) {
  const generatedId = useId();
  const inputId = id || generatedId;
  const helpId = helpText ? `${inputId}-help` : undefined;
  const errorId = error ? `${inputId}-error` : undefined;
  const describedBy = [inputProps["aria-describedby"], helpId, errorId].filter(Boolean).join(" ") || undefined;

  return (
    <div className={`app-field app-number-input ${error ? "has-error" : ""} ${className}`.trim()}>
      <label className="app-field-label" htmlFor={inputId}>
        {label}{required ? <span aria-hidden="true"> *</span> : null}
      </label>
      <input
        {...inputProps}
        ref={ref}
        id={inputId}
        type="number"
        inputMode={inputProps.inputMode || "decimal"}
        required={required}
        value={value}
        onChange={(event) => onValueChange?.(event.target.value)}
        aria-describedby={describedBy}
        aria-invalid={error ? true : inputProps["aria-invalid"]}
      />
      {helpText ? <span id={helpId} className="app-field-help">{helpText}</span> : null}
      {error ? <span id={errorId} className="app-field-error" role="alert">{error}</span> : null}
    </div>
  );
});
