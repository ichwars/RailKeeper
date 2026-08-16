import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { MasterDataEntry } from "../../shared/api";
import { MasterDataLifecycleActions } from "./MasterDataLifecycleActions";

const entry = (active: boolean, capabilities: MasterDataEntry["capabilities"]): MasterDataEntry => ({
  id: "manufacturer-custom",
  type: "manufacturer",
  key: "custom",
  label: "Eigener Hersteller",
  active,
  sortOrder: 10,
  metadata: {},
  origin: "custom",
  capabilities,
  createdAt: "2026-08-16T10:00:00Z",
  updatedAt: "2026-08-16T10:00:00Z"
});

describe("MasterDataLifecycleActions", () => {
  it("uses backend capabilities as the only source for lifecycle actions", async () => {
    const user = userEvent.setup();
    const onEdit = vi.fn();
    const onDeactivate = vi.fn();
    const onDelete = vi.fn();
    render(
      <MasterDataLifecycleActions
        entry={entry(true, { canDeactivate: true, canReactivate: false, canDelete: true })}
        displayLabel="Eigener Hersteller"
        onEdit={onEdit}
        onDeactivate={onDeactivate}
        onReactivate={vi.fn()}
        onDelete={onDelete}
      />
    );

    await user.click(screen.getByRole("button", { name: "Eigener Hersteller bearbeiten" }));
    await user.click(screen.getByRole("button", { name: "Eigener Hersteller deaktivieren" }));
    await user.click(screen.getByRole("button", { name: "Eigener Hersteller endgültig löschen" }));

    expect(onEdit).toHaveBeenCalledOnce();
    expect(onDeactivate).toHaveBeenCalledOnce();
    expect(onDelete).toHaveBeenCalledOnce();
    expect(screen.queryByRole("button", { name: /reaktivieren/i })).not.toBeInTheDocument();
  });

  it("does not infer deletion from a custom origin", () => {
    render(
      <MasterDataLifecycleActions
        entry={entry(false, { canDeactivate: false, canReactivate: true, canDelete: false })}
        displayLabel="Verwendeter Hersteller"
        onEdit={vi.fn()}
        onDeactivate={vi.fn()}
        onReactivate={vi.fn()}
        onDelete={vi.fn()}
      />
    );

    expect(screen.getByRole("button", { name: "Verwendeter Hersteller reaktivieren" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /endgültig löschen/i })).not.toBeInTheDocument();
  });
});
