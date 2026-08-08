import { ChangeEvent, forwardRef, InputHTMLAttributes, ReactNode, useId, useRef } from "react";
import { FileUp, X } from "lucide-react";

export type AppFilePickerProps = Omit<InputHTMLAttributes<HTMLInputElement>, "onChange" | "type" | "value"> & {
  label: ReactNode;
  helpText?: ReactNode;
  error?: ReactNode;
  file?: File | null;
  onFileChange?: (file: File | null) => void;
  readOnly?: boolean;
  triggerLabel?: string;
  clearLabel?: string;
  emptyLabel?: string;
};

export const AppFilePicker = forwardRef<HTMLInputElement, AppFilePickerProps>(function AppFilePicker(
  {
    label,
    helpText,
    error,
    file,
    onFileChange,
    readOnly = false,
    triggerLabel = "Datei auswählen",
    clearLabel = "Datei entfernen",
    emptyLabel = "Keine Datei ausgewählt",
    className = "",
    id,
    required,
    disabled,
    ...inputProps
  },
  forwardedRef
) {
  const generatedId = useId();
  const inputId = id || generatedId;
  const labelId = `${inputId}-label`;
  const helpId = helpText ? `${inputId}-help` : undefined;
  const errorId = error ? `${inputId}-error` : undefined;
  const describedBy = [inputProps["aria-describedby"], helpId, errorId].filter(Boolean).join(" ") || undefined;
  const inputRef = useRef<HTMLInputElement | null>(null);
  const unavailable = disabled || readOnly;

  const setInputRef = (node: HTMLInputElement | null) => {
    inputRef.current = node;
    if (typeof forwardedRef === "function") forwardedRef(node);
    else if (forwardedRef) forwardedRef.current = node;
  };

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    onFileChange?.(event.target.files?.[0] || null);
  };

  const clear = () => {
    if (unavailable) return;
    if (inputRef.current) inputRef.current.value = "";
    onFileChange?.(null);
  };

  return (
    <div
      className={`app-field app-file-picker ${error ? "has-error" : ""} ${className}`.trim()}
      aria-readonly={readOnly || undefined}
    >
      <span id={labelId} className="app-field-label">
        {label}{required ? <span aria-hidden="true"> *</span> : null}
      </span>
      <input
        {...inputProps}
        ref={setInputRef}
        id={inputId}
        className="visually-hidden"
        type="file"
        tabIndex={-1}
        required={required}
        disabled={disabled}
        onChange={handleChange}
        aria-labelledby={labelId}
        aria-describedby={describedBy}
        aria-invalid={error ? true : inputProps["aria-invalid"]}
      />
      <div className="app-file-picker-control">
        <button
          type="button"
          className="app-file-picker-trigger"
          disabled={unavailable}
          onClick={() => inputRef.current?.click()}
          aria-describedby={describedBy}
        >
          <FileUp size={15} aria-hidden="true" />
          {triggerLabel}
        </button>
        <span className={`app-file-picker-name ${file ? "" : "empty"}`.trim()}>{file?.name || emptyLabel}</span>
        {file ? (
          <button
            type="button"
            className="app-file-picker-clear"
            disabled={unavailable}
            onClick={clear}
            aria-label={clearLabel}
            title={clearLabel}
          >
            <X size={15} aria-hidden="true" />
          </button>
        ) : null}
      </div>
      {helpText ? <span id={helpId} className="app-field-help">{helpText}</span> : null}
      {error ? <span id={errorId} className="app-field-error" role="alert">{error}</span> : null}
    </div>
  );
});
