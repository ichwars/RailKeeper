import { ChevronDown, ChevronUp } from "lucide-react";
import { forwardRef, InputHTMLAttributes, useRef } from "react";
import { useI18n } from "../i18n";

type AppNumberControlProps = Omit<InputHTMLAttributes<HTMLInputElement>, "type">;

export const AppNumberControl = forwardRef<HTMLInputElement, AppNumberControlProps>(function AppNumberControl(
  { className = "", disabled, readOnly, step, ...inputProps }, forwardedRef
) {
  const { t } = useI18n();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const setRef = (node: HTMLInputElement | null) => {
    inputRef.current = node;
    if (typeof forwardedRef === "function") forwardedRef(node);
    else if (forwardedRef) forwardedRef.current = node;
  };
  const change = (direction: number) => {
    const input = inputRef.current;
    if (!input || disabled || readOnly) return;
    const previous = input.value;
    if (direction > 0) input.stepUp();
    else input.stepDown();
    if (previous !== input.value) input.dispatchEvent(new Event("input", { bubbles: true }));
    input.focus();
  };

  return <span className="app-number-control">
    <input {...inputProps} ref={setRef} type="number" className={className}
      inputMode={inputProps.inputMode || "decimal"} disabled={disabled} readOnly={readOnly} step={step} />
    {step !== "any" && <span className="app-number-steps">
      <button type="button" tabIndex={-1} disabled={disabled || readOnly}
        aria-label={t("controls.number.increase")} title={t("controls.number.increase")}
        onClick={() => change(1)}><ChevronUp size={14} aria-hidden="true" /></button>
      <button type="button" tabIndex={-1} disabled={disabled || readOnly}
        aria-label={t("controls.number.decrease")} title={t("controls.number.decrease")}
        onClick={() => change(-1)}><ChevronDown size={14} aria-hidden="true" /></button>
    </span>}
  </span>;
});
