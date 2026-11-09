import { Check, Circle, Minus } from "lucide-react";
import { forwardRef, InputHTMLAttributes } from "react";

type AppChoiceInputProps = Omit<InputHTMLAttributes<HTMLInputElement>, "type"> & {
  type: "checkbox" | "radio";
};

// Keep the native input's label, keyboard, form and indeterminate semantics.
export const AppChoiceInput = forwardRef<HTMLInputElement, AppChoiceInputProps>(function AppChoiceInput(
  { type, className = "", ...inputProps }, ref
) {
  return <span className="app-choice-control">
    <input {...inputProps} ref={ref} type={type} className={`app-choice-input ${className}`.trim()} />
    <span className="app-choice-mark" aria-hidden="true">
      {type === "checkbox" ? <><Check className="app-choice-check" size={13} />
        <Minus className="app-choice-mixed" size={13} /></> : <Circle size={8} fill="currentColor" />}
    </span>
  </span>;
});
