import { forwardRef, ReactNode, TextareaHTMLAttributes, useId } from "react";

export type AppTextAreaProps = TextareaHTMLAttributes<HTMLTextAreaElement> & {
  label: ReactNode;
  helpText?: ReactNode;
  error?: ReactNode;
};

export const AppTextArea = forwardRef<HTMLTextAreaElement, AppTextAreaProps>(function AppTextArea(
  { label, helpText, error, className = "", id, required, ...textareaProps },
  ref
) {
  const generatedId = useId();
  const inputId = id || generatedId;
  const helpId = helpText ? `${inputId}-help` : undefined;
  const errorId = error ? `${inputId}-error` : undefined;
  const describedBy = [textareaProps["aria-describedby"], helpId, errorId].filter(Boolean).join(" ") || undefined;

  return (
    <div className={`app-field app-text-area ${error ? "has-error" : ""} ${className}`.trim()}>
      <label className="app-field-label" htmlFor={inputId}>
        {label}{required ? <span aria-hidden="true"> *</span> : null}
      </label>
      <textarea
        {...textareaProps}
        ref={ref}
        id={inputId}
        required={required}
        aria-describedby={describedBy}
        aria-invalid={error ? true : textareaProps["aria-invalid"]}
      />
      {helpText ? <span id={helpId} className="app-field-help">{helpText}</span> : null}
      {error ? <span id={errorId} className="app-field-error" role="alert">{error}</span> : null}
    </div>
  );
});
