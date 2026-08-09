import { Check } from "lucide-react";
import { forwardRef, InputHTMLAttributes, ReactNode, useId } from "react";

export type AppCheckboxProps = Omit<InputHTMLAttributes<HTMLInputElement>, "type"> & {
  label: ReactNode;
};

export const AppCheckbox = forwardRef<HTMLInputElement, AppCheckboxProps>(function AppCheckbox(
  { label, className = "", id, ...inputProps },
  ref
) {
  const generatedId = useId();
  const inputId = id || generatedId;

  return (
    <label className={`app-checkbox ${className}`.trim()} htmlFor={inputId}>
      <span className="app-checkbox-control">
        <input {...inputProps} ref={ref} id={inputId} type="checkbox" />
        <span className="app-checkbox-mark" aria-hidden="true"><Check size={13} /></span>
      </span>
      <span>{label}</span>
    </label>
  );
});
