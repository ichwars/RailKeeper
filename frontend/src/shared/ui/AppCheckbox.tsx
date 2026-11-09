import { forwardRef, InputHTMLAttributes, ReactNode, useId } from "react";
import { AppChoiceInput } from "./AppChoiceInput";

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
      <AppChoiceInput {...inputProps} ref={ref} id={inputId} type="checkbox" />
      <span>{label}</span>
    </label>
  );
});
