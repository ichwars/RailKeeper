import { render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { VehicleCreateWizardShell } from "./VehicleCreateWizardShell";

describe("VehicleCreateWizardShell", () => {
  it("renders one labelled dialog with an accessible ordered progress rail", () => {
    render(
      <VehicleCreateWizardShell
        step="article"
        summaries={{ basics: "Set · Roco", article: "Datenübernahme prüfen", details: "3 Fahrzeuge" }}
        onClose={vi.fn()}
        onSubmit={(event) => event.preventDefault()}
        footer={<><button type="button">Zurück</button><button type="submit">Weiter zu Fahrzeugdetails</button></>}
      >
        <h3>Artikeldaten</h3>
      </VehicleCreateWizardShell>
    );

    expect(screen.getAllByRole("dialog")).toHaveLength(1);
    const steps = screen.getAllByRole("listitem");
    expect(steps).toHaveLength(3);
    expect(steps[0]).toHaveClass("done");
    expect(within(steps[0]).getByTitle("Erledigt")).toBeInTheDocument();
    expect(steps[1]).toHaveAttribute("aria-current", "step");
    expect(screen.getByRole("button", { name: "Weiter zu Fahrzeugdetails" })).toBeVisible();
  });
});
