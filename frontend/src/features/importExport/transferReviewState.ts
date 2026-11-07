import type { ImportReviewState } from "../../shared/ImportReview";
import type { DataTransferIssue, DataTransferPreviewRecord } from "./dataTransferModel";

export function transferReviewState(record: DataTransferPreviewRecord, issues: DataTransferIssue[]): ImportReviewState {
  const relevant = issues.filter(issue => issue.area === record.area && issue.recordKey === record.recordKey);
  if (record.classification === "error" || relevant.some(issue => issue.selectedResolution === "skip")) return "excluded";
  if (relevant.some(issue => !issue.selectedResolution) || (record.classification === "warning" && relevant.length === 0)) return "unclear";
  // Reusing a target does not copy source fields. Its existing linking action remains available.
  if (record.proposedAction === "use_existing") return "excluded";
  return record.proposedAction === "replace" ? "changed" : "new";
}
