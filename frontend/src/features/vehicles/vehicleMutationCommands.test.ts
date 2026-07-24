import type { FormEvent } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api, type Vehicle } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { emptyVehicle } from "./vehicleViewModel";
import { createVehicleMutationCommands } from "./vehicleMutationCommands";

function commandOptions() {
  const selected: Vehicle | null = null;
  return {
    editor: {
      form: { ...emptyVehicle, manufacturer: "Piko", name: "BR 106" },
      selected: selected as Vehicle | null,
      mode: "create" as const,
      setSaving: vi.fn(),
      setSelectedDetail: vi.fn(),
      setMode: vi.fn(),
      setSaveAttempted: vi.fn(),
      setActiveTab: vi.fn(),
      openModelSection: vi.fn(),
      close: vi.fn()
    },
    validation: { missingRequiredLabels: [] as string[] },
    media: { pendingImages: [] },
    spareParts: { selectedInputs: vi.fn(() => []), clearSelected: vi.fn() },
    functions: { configuredKeys: [], edit: vi.fn() },
    ecos: {
      draft: null,
      unclearFieldCount: 0,
      markSaved: vi.fn(),
      clear: vi.fn(),
      returnToSession: vi.fn()
    },
    deletion: { candidate: null as Vehicle | null, setCandidate: vi.fn() },
    reloadVehicles: vi.fn(),
    onMessage: vi.fn(),
    t: (key: string) => key
  };
}

const submitEvent = () => ({ preventDefault: vi.fn() }) as unknown as FormEvent;

describe("createVehicleMutationCommands", () => {
  afterEach(() => vi.restoreAllMocks());

  it("stops at required model fields before writing", async () => {
    const options = commandOptions();
    options.validation.missingRequiredLabels = ["Spurweite"];
    const createVehicle = vi.spyOn(api, "createVehicle");

    await createVehicleMutationCommands(options).submit(submitEvent());

    expect(createVehicle).not.toHaveBeenCalled();
    expect(options.editor.setActiveTab).toHaveBeenCalledWith("model");
    expect(options.editor.openModelSection).toHaveBeenCalledOnce();
    expect(options.onMessage).toHaveBeenCalledWith("vehicles.requiredMissing");
  });

  it("creates and reloads a vehicle", async () => {
    const options = commandOptions();
    const created = vehicleFixture({ id: "created" });
    vi.spyOn(api, "createVehicle").mockResolvedValue(created);
    vi.spyOn(api, "vehicle").mockResolvedValue(created);

    await createVehicleMutationCommands(options).submit(submitEvent());

    expect(api.createVehicle).toHaveBeenCalledWith(expect.objectContaining({ name: "BR 106", images: [] }));
    expect(options.editor.setSelectedDetail).toHaveBeenCalledWith(created);
    expect(options.editor.setMode).toHaveBeenCalledWith("edit");
    expect(options.reloadVehicles).toHaveBeenCalledOnce();
    expect(options.onMessage).toHaveBeenCalledWith("vehicles.createdContinue");
    expect(options.editor.setSaving).toHaveBeenLastCalledWith(false);
  });

  it("deletes the selected candidate and closes its editor", async () => {
    const candidate = vehicleFixture({ id: "delete-me" });
    const options = commandOptions();
    options.editor.selected = candidate;
    options.deletion.candidate = candidate;
    vi.spyOn(api, "deleteVehicle").mockResolvedValue(undefined);

    createVehicleMutationCommands(options).confirmDelete();
    await vi.waitFor(() => expect(options.reloadVehicles).toHaveBeenCalledOnce());

    expect(api.deleteVehicle).toHaveBeenCalledWith(candidate.id);
    expect(options.editor.close).toHaveBeenCalledOnce();
    expect(options.deletion.setCandidate).toHaveBeenCalledWith(null);
  });
});
