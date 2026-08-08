import { KeyboardEvent, useEffect, useRef } from "react";

export type SettingsTabOption<T extends string> = {
  id: T;
  label: string;
};

type SettingsTabListProps<T extends string> = {
  ariaLabel: string;
  options: readonly SettingsTabOption<T>[];
  value: T;
  onChange: (value: T) => void;
  className?: string;
};

export function SettingsTabList<T extends string>({
  ariaLabel,
  options,
  value,
  onChange,
  className
}: SettingsTabListProps<T>) {
  const refs = useRef<Array<HTMLButtonElement | null>>([]);

  useEffect(() => {
    const selectedIndex = options.findIndex((option) => option.id === value);
    const selected = refs.current[selectedIndex];
    if (selected && typeof selected.scrollIntoView === "function") {
      selected.scrollIntoView({ block: "nearest", inline: "nearest" });
    }
  }, [options, value]);

  const selectAt = (index: number) => {
    const option = options[index];
    if (!option) return;
    onChange(option.id);
    refs.current[index]?.focus();
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLButtonElement>, index: number) => {
    let nextIndex: number | undefined;
    if (event.key === "ArrowRight") nextIndex = (index + 1) % options.length;
    if (event.key === "ArrowLeft") nextIndex = (index - 1 + options.length) % options.length;
    if (event.key === "Home") nextIndex = 0;
    if (event.key === "End") nextIndex = options.length - 1;
    if (nextIndex === undefined) return;
    event.preventDefault();
    selectAt(nextIndex);
  };

  return (
    <div
      className={`settings-secondary-tabs settings-tab-list ${className || ""}`.trim()}
      role="tablist"
      aria-label={ariaLabel}
    >
      {options.map((option, index) => {
        const selected = option.id === value;
        return (
          <button
            key={option.id}
            type="button"
            role="tab"
            ref={(element) => {
              refs.current[index] = element;
            }}
            aria-selected={selected}
            tabIndex={selected ? 0 : -1}
            className={selected ? "active" : ""}
            onClick={() => onChange(option.id)}
            onKeyDown={(event) => handleKeyDown(event, index)}
          >
            {option.label}
          </button>
        );
      })}
    </div>
  );
}
