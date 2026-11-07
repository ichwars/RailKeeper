import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, ApiError } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { SetupView } from "./SetupView";

describe("first administrator authorization", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    setLanguage("de");
    vi.spyOn(api, "setupStatus").mockResolvedValue({ setupRequired: true, cookieSecure: false });
  });

  function fillSetup() {
    fireEvent.change(screen.getByLabelText("Einrichtungscode"), { target: { value: " local-one-use-code " } });
    fireEvent.change(screen.getByLabelText("Benutzername"), { target: { value: "owner" } });
    fireEvent.change(screen.getByLabelText("E-Mail-Adresse"), { target: { value: "owner@example.test" } });
    fireEvent.change(screen.getByLabelText("Passwort", { exact: true }), { target: { value: "a secure password" } });
    fireEvent.change(screen.getByLabelText("Passwort wiederholen"), { target: { value: "a secure password" } });
  }

  it("submits the locally supplied code and retains it when setup is rejected", async () => {
    const submit = vi.spyOn(api, "createAdmin").mockRejectedValue(new Error("Invalid setup code"));
    const complete = vi.fn();
    const { container } = render(<SetupView onComplete={complete} />);
    fillSetup();
    fireEvent.submit(container.querySelector("form")!);
    expect(await screen.findByText("Invalid setup code")).toBeInTheDocument();
    expect(submit).toHaveBeenCalledExactlyOnceWith({
      username: "owner", email: "owner@example.test", password: "a secure password", setupCode: "local-one-use-code"
    });
    expect(screen.getByLabelText("Einrichtungscode")).toHaveValue(" local-one-use-code ");
    expect(complete).not.toHaveBeenCalled();
  });

  it("finishes setup only after the authorized request succeeds", async () => {
    vi.spyOn(api, "createAdmin").mockResolvedValue({ status: "created" });
    const complete = vi.fn();
    const { container } = render(<SetupView onComplete={complete} />);
    fillSetup();
    fireEvent.submit(container.querySelector("form")!);
    await waitFor(() => expect(complete).toHaveBeenCalledOnce());
  });

  it("explains a cookie setting that would prevent this HTTP session", async () => {
    vi.mocked(api.setupStatus).mockResolvedValue({ setupRequired: true, cookieSecure: true });
    render(<SetupView onComplete={vi.fn()} />);
    expect(await screen.findByRole("alert")).toHaveTextContent("Cookie-Einstellung passt nicht");
  });

  it("explains a rejected code in the selected language", async () => {
    vi.spyOn(api, "createAdmin").mockRejectedValue(new ApiError("Forbidden", "invalid_setup_code", 403));
    const { container } = render(<SetupView onComplete={vi.fn()} />);
    fillSetup();
    fireEvent.submit(container.querySelector("form")!);
    expect(await screen.findByText(/Der Einrichtungscode fehlt/)).toBeInTheDocument();
  });
});
