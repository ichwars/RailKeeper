import { describe, expect, it } from "vitest";

import { createClientErrorBuffer } from "./clientErrorBuffer";

describe("client error buffer", () => {
  it("keeps only the newest bounded records in memory", () => {
    const buffer = createClientErrorBuffer(2, () => "2026-08-29T08:00:00.000Z");
    buffer.record({ category: "window", errorClass: "Error", message: "one" });
    buffer.record({ category: "window", errorClass: "Error", message: "two" });
    buffer.record({ category: "window", errorClass: "Error", message: "three" });

    expect(buffer.snapshot().map((entry) => entry.message)).toEqual([
      "Unexpected client error", "Unexpected client error"
    ]);
  });

  it("never stores arbitrary messages or machine and credential values", () => {
    const buffer = createClientErrorBuffer(30, () => "2026-08-29T08:00:00.000Z");
    buffer.record({
      category: "network",
      errorClass: "TypeError",
      message: "Bearer ghp_secret host.internal:8080 \\\\server\\share /etc/passwd ada@example.de",
      method: "GET",
      route: "/vehicles/8b1d6340-0f66-472e-9954-58f1e9d88c88?token=secret",
      problemCode: "unsafe value ghp_secret"
    });

    const [entry] = buffer.snapshot();
    expect(entry.message).toBe("Network request failed");
    expect(entry.problemCode).toBe("unknown_error");
    expect(entry.route).toBe("/vehicles");
    expect(entry).not.toHaveProperty("stack");
  });

  it("rejects regex-shaped problem codes that are not explicitly public", () => {
    const buffer = createClientErrorBuffer(30, () => "2026-08-29T08:00:00.000Z");
    buffer.record({
      category: "api",
      errorClass: "ApiError",
      problemCode: "bearer_secret_token",
      message: "API request failed",
      method: "POST",
      route: "/vehicles",
      status: 500
    });

    expect(buffer.snapshot()[0]?.problemCode).toBe("unknown_error");
  });

  it("omits a null status and normalizes fields to the public diagnostic contract", () => {
    const buffer = createClientErrorBuffer(30, () => "2026-08-29T08:00:00.000Z");
    buffer.record({
      category: "timeout", errorClass: "AbortError", problemCode: "request_timeout",
      message: "secret", method: "post secret", route: "/bug-reports/report-id", status: null
    });

    expect(buffer.snapshot()[0]).toEqual({
      timestamp: "2026-08-29T08:00:00.000Z", category: "timeout", errorClass: "AbortError",
      problemCode: "request_timeout", message: "Request timed out", method: "",
      route: "/bug-reports"
    });
  });
});
