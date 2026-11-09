import { forwardRef, InputHTMLAttributes, ReactNode, useRef } from "react";

type AppFileButtonProps = Omit<InputHTMLAttributes<HTMLInputElement>, "type" | "value" | "children"> & {
  children: ReactNode;
  buttonClassName?: string;
};

// For immediate uploads: retain the original change event and the complete FileList.
export const AppFileButton = forwardRef<HTMLInputElement, AppFileButtonProps>(function AppFileButton(
  { children, buttonClassName = "secondary-button", className = "", disabled, ...inputProps }, forwardedRef
) {
  const inputRef = useRef<HTMLInputElement | null>(null);
  return <span className={`app-file-button ${className}`.trim()}>
    <input {...inputProps} type="file" className="visually-hidden" tabIndex={-1} disabled={disabled}
      ref={node => {
        inputRef.current = node;
        if (typeof forwardedRef === "function") forwardedRef(node);
        else if (forwardedRef) forwardedRef.current = node;
      }} />
    <button type="button" className={buttonClassName} disabled={disabled}
      aria-label={inputProps["aria-label"]} aria-describedby={inputProps["aria-describedby"]}
      onClick={() => inputRef.current?.click()}>{children}</button>
  </span>;
});
