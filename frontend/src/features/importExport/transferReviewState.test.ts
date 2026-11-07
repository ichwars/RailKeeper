import { describe, expect, it } from "vitest";
import type { DataTransferIssue, DataTransferPreviewRecord } from "./dataTransferModel";
import { transferReviewState } from "./transferReviewState";

const record: DataTransferPreviewRecord = { area: "vehicles", recordKey: "RK-1", classification: "ready",
  proposedAction: "create", data: {} };
const issue: DataTransferIssue = { id: "issue", jobId: "job", area: "vehicles", recordKey: "RK-1",
  field: "name", rowNumber: null, severity: "warning", code: "matching_manufacturer_article_number", message: "Review",
  proposedResolution: "create", selectedResolution: "", createdAt: "", updatedAt: "" };

describe("source-field adoption status", () => {
  it("keeps unresolved comparison, exclusion and actual writes distinct", () => {
    expect(transferReviewState(record, [])).toBe("new");
    expect(transferReviewState({ ...record, proposedAction: "copy" }, [])).toBe("new");
    expect(transferReviewState({ ...record, proposedAction: "replace" }, [])).toBe("changed");
    expect(transferReviewState(record, [issue])).toBe("unclear");
    expect(transferReviewState(record, [{ ...issue, selectedResolution: "skip" }])).toBe("excluded");
    expect(transferReviewState({ ...record, classification: "error" }, [])).toBe("excluded");
  });
  it("does not infer identical data when an existing target is reused", () => {
    expect(transferReviewState({ ...record, proposedAction: "use_existing" }, [])).toBe("excluded");
    expect(transferReviewState(record, [{ ...issue, selectedResolution: "create" }])).toBe("new");
    expect(transferReviewState(record, [{ ...issue, area: "accessories", selectedResolution: "skip" }])).toBe("new");
    expect(transferReviewState(record, [{ ...issue, recordKey: "RK-2", selectedResolution: "skip" }])).toBe("new");
  });
});
