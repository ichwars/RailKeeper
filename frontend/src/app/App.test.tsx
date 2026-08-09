import { beforeEach, describe, expect, it } from "vitest";

import { configuredStartView, currentView } from "./App";

describe("App navigation availability", () => {
  beforeEach(() => {
    window.history.replaceState(null, "", "/");
  });

  it("uses a stored layout start view", () => {
    window.localStorage.setItem("railkeeper.settings.defaultView", "layouts");

    expect(configuredStartView()).toBe("layouts");
    expect(window.localStorage.getItem("railkeeper.settings.defaultView")).toBe("layouts");
  });

  it("keeps the direct layout route available", () => {
    window.history.replaceState(null, "", "/layouts");
    expect(currentView()).toBe("layouts");
  });
});
