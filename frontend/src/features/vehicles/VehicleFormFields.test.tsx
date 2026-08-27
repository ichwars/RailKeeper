import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { languageSettingKey } from "../../shared/i18n";
import { RequiredLabel } from "./VehicleFormFields";

describe("RequiredLabel", () => {
  it.each([
    { filled: false, showError: false, state: "pending", text: "Noch offen", icon: "lucide-circle-dashed" },
    { filled: false, showError: true, state: "missing", text: "Eingabe fehlt", icon: "lucide-circle-alert" },
    { filled: true, showError: true, state: "filled", text: "Ausgefüllt", icon: "lucide-circle-check" },
  ])("renders the $state state with text and a distinct icon", ({ filled, showError, state, text, icon }) => {
    const { container } = render(<RequiredLabel label="Hersteller" filled={filled} showError={showError} />);

    expect(screen.getByRole("status", { name: text })).toHaveTextContent(text);
    expect(screen.getByRole("status", { name: text })).toHaveClass(state);
    expect(container.querySelector(`.${icon}`)).toBeInTheDocument();
  });

  it("associates the missing status with its required field", () => {
    render(
      <label>
        <RequiredLabel label="Hersteller" filled={false} showError />
        <input required />
      </label>,
    );

    expect(screen.getByRole("textbox", { name: "Hersteller Eingabe fehlt" })).toBeRequired();
  });

  it("provides English status text", () => {
    window.localStorage.setItem(languageSettingKey, "en");

    render(<RequiredLabel label="Manufacturer" filled={false} showError />);

    expect(screen.getByRole("status", { name: "Value missing" })).toHaveTextContent("Value missing");
  });
});
