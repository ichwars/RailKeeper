import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import type { MasterDataEntry } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

export type MasterDataBulkSelectionState = {
  selectedKeys: string[];
  selectedEntries: MasterDataEntry[];
  selectedCount: number;
  selectableVisibleCount: number;
  allVisibleSelected: boolean;
  someVisibleSelected: boolean;
  isSelected: (key: string) => boolean;
  toggle: (key: string, checked: boolean) => void;
  toggleVisible: (checked: boolean) => void;
  clear: () => void;
};

function canSelect(entry: MasterDataEntry) {
  return entry.active && Boolean(entry.capabilities?.canDeactivate);
}

export function useMasterDataBulkSelection(
  scopeKey: string,
  entries: MasterDataEntry[],
  visibleEntries: MasterDataEntry[]
): MasterDataBulkSelectionState {
  const [selection, setSelection] = useState<Set<string>>(() => new Set());
  const selectableKeys = useMemo(
    () => new Set(entries.filter(canSelect).map((entry) => entry.key)),
    [entries]
  );
  const visibleKeys = useMemo(
    () => visibleEntries.filter(canSelect).map((entry) => entry.key),
    [visibleEntries]
  );

  useEffect(() => {
    setSelection(new Set());
  }, [scopeKey]);

  useEffect(() => {
    setSelection((current) => {
      const next = new Set([...current].filter((key) => selectableKeys.has(key)));
      if (next.size === current.size && [...next].every((key) => current.has(key))) return current;
      return next;
    });
  }, [selectableKeys]);

  const toggle = useCallback((key: string, checked: boolean) => {
    if (checked && !selectableKeys.has(key)) return;
    setSelection((current) => {
      const next = new Set(current);
      if (checked) next.add(key);
      else next.delete(key);
      return next;
    });
  }, [selectableKeys]);

  const toggleVisible = useCallback((checked: boolean) => {
    setSelection((current) => {
      const next = new Set(current);
      for (const key of visibleKeys) {
        if (checked) next.add(key);
        else next.delete(key);
      }
      return next;
    });
  }, [visibleKeys]);

  const clear = useCallback(() => setSelection(new Set()), []);
  const selectedEntries = entries.filter((entry) => selection.has(entry.key) && canSelect(entry));
  const selectedKeys = selectedEntries.map((entry) => entry.key);
  const selectedVisibleCount = visibleKeys.filter((key) => selection.has(key)).length;
  const allVisibleSelected = visibleKeys.length > 0 && selectedVisibleCount === visibleKeys.length;

  return {
    selectedKeys,
    selectedEntries,
    selectedCount: selectedKeys.length,
    selectableVisibleCount: visibleKeys.length,
    allVisibleSelected,
    someVisibleSelected: selectedVisibleCount > 0 && !allVisibleSelected,
    isSelected: (key) => selection.has(key),
    toggle,
    toggleVisible,
    clear
  };
}

export function MasterDataBulkSelectAllCheckbox({
  selection,
  disabled = false
}: {
  selection: MasterDataBulkSelectionState;
  disabled?: boolean;
}) {
  const { t } = useI18n();
  const inputRef = useRef<HTMLInputElement | null>(null);
  useEffect(() => {
    if (inputRef.current) inputRef.current.indeterminate = selection.someVisibleSelected;
  }, [selection.someVisibleSelected]);

  return <input
    ref={inputRef}
    className="master-data-bulk-checkbox"
    type="checkbox"
    aria-label={t("settings.master.selectAllVisible")}
    checked={selection.allVisibleSelected}
    disabled={disabled || selection.selectableVisibleCount === 0}
    onChange={(event) => selection.toggleVisible(event.target.checked)}
  />;
}

export function MasterDataBulkToolbar({
  count,
  busy,
  onDeactivate
}: {
  count: number;
  busy: boolean;
  onDeactivate: () => void;
}) {
  const { t } = useI18n();
  if (count === 0) return null;
  return <div className="master-data-bulk-toolbar">
    <strong aria-live="polite">{t(
      count === 1 ? "settings.master.selectedOne" : "settings.master.selectedMany",
      { count }
    )}</strong>
    <button type="button" className="secondary-button compact-action" disabled={busy} onClick={onDeactivate}>
      {t("settings.master.deactivateSelected")}
    </button>
  </div>;
}
