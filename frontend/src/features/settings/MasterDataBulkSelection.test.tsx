import { act, render, renderHook, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { MasterDataEntry } from "../../shared/api";
import {
  MasterDataBulkEntryCheckbox,
  MasterDataBulkSelectAllCheckbox,
  MasterDataBulkToolbar,
  useMasterDataBulkSelection
} from "./MasterDataBulkSelection";

function entry(key: string, active = true, canDeactivate = active): MasterDataEntry {
  return {
    id: `manufacturer:${key}`,
    type: "manufacturer",
    key,
    label: key[0].toUpperCase() + key.slice(1),
    active,
    sortOrder: 0,
    metadata: {},
    origin: "custom",
    capabilities: { canDeactivate, canReactivate: !active, canDelete: false },
    createdAt: "2026-08-28T00:00:00Z",
    updatedAt: "2026-08-28T00:00:00Z"
  };
}

describe("useMasterDataBulkSelection", () => {
  it("selects only eligible visible entries and prunes stale selections", async () => {
    const first = entry("first");
    const second = entry("second");
    const inactive = entry("inactive", false);
    const incapable = entry("incapable", true, false);
    const { result, rerender } = renderHook(
      ({ scope, entries, visible }) => useMasterDataBulkSelection(scope, entries, visible),
      { initialProps: {
        scope: "manufacturer|all",
        entries: [first, second, inactive, incapable],
        visible: [first, inactive, incapable]
      } }
    );

    act(() => result.current.toggle("inactive", true));
    expect(result.current.selectedKeys).toEqual([]);

    act(() => result.current.toggle("first", true));
    expect(result.current.selectedKeys).toEqual(["first"]);
    expect(result.current.someVisibleSelected).toBe(false);
    expect(result.current.allVisibleSelected).toBe(true);

    act(() => result.current.toggleVisible(false));
    expect(result.current.selectedKeys).toEqual([]);

    rerender({
      scope: "manufacturer|all",
      entries: [first, second, inactive, incapable],
      visible: [first, second, inactive, incapable]
    });
    act(() => result.current.toggleVisible(true));
    expect(result.current.selectedKeys).toEqual(["first", "second"]);

    rerender({
      scope: "manufacturer|all",
      entries: [entry("first", false), second, inactive, incapable],
      visible: [entry("first", false), second, inactive, incapable]
    });
    await waitFor(() => expect(result.current.selectedKeys).toEqual(["second"]));

    rerender({
      scope: "manufacturer|filtered",
      entries: [entry("first", false), second, inactive, incapable],
      visible: [second]
    });
    await waitFor(() => expect(result.current.selectedKeys).toEqual([]));
  });

  it("exposes an indeterminate select-all checkbox and a busy-aware toolbar", async () => {
    const user = userEvent.setup();
    const onDeactivate = vi.fn();
    const entries = [entry("first"), entry("second"), entry("incapable", true, false)];

    function Harness({ busy = false }: { busy?: boolean }) {
      const selection = useMasterDataBulkSelection("manufacturer", entries, entries);
      return <>
        <MasterDataBulkSelectAllCheckbox selection={selection} disabled={busy} />
        <MasterDataBulkEntryCheckbox
          selection={selection}
          entry={entries[0]}
          label={entries[0].label}
          disabled={busy}
        />
        <button type="button" onClick={() => selection.toggle("first", true)}>Ersten auswählen</button>
        <MasterDataBulkToolbar count={selection.selectedCount} busy={busy} onDeactivate={onDeactivate} />
      </>;
    }

    const { rerender } = render(<Harness />);
    const selectAll = screen.getByRole("checkbox", { name: "Alle sichtbaren aktiven Einträge auswählen" });
    const selectFirst = screen.getByRole("checkbox", { name: "First auswählen" });
    expect(selectAll.closest("label")).toHaveClass("master-data-bulk-checkbox-target");
    expect(selectFirst.closest("label")).toHaveClass("master-data-bulk-checkbox-target");
    expect(selectAll).not.toBeChecked();
    expect(selectAll).toHaveProperty("indeterminate", false);

    await user.click(screen.getByRole("button", { name: "Ersten auswählen" }));
    expect(selectAll).not.toBeChecked();
    expect(selectAll).toHaveProperty("indeterminate", true);
    expect(screen.getByText("1 Eintrag ausgewählt")).toBeVisible();

    await user.click(selectAll);
    expect(selectAll).toBeChecked();
    expect(selectAll).toHaveProperty("indeterminate", false);
    expect(screen.getByText("2 Einträge ausgewählt")).toBeVisible();

    await user.click(screen.getByRole("button", { name: "Ausgewählte deaktivieren" }));
    expect(onDeactivate).toHaveBeenCalledOnce();

    rerender(<Harness busy />);
    expect(screen.getByRole("checkbox", { name: "Alle sichtbaren aktiven Einträge auswählen" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Ausgewählte deaktivieren" })).toBeDisabled();
  });
});
