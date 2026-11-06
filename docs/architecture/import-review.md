# Import comparison and selection

`ImportReviewContext` names the actual source, target and effect of a review. `ImportReviewStatus`
uses one vocabulary: new, different, identical, unclear and excluded. Status text remains visible
alongside icons. Comparison, selection and readiness are separate concepts; an unselected different
value remains different, and a ready action does not prove two source records identical.

The existing specialized parsers, supported formats, validation, selection keys and write paths remain
in their owning features. The shared components perform no requests and hold no adoption state.

| Review | Mapping and retained scope |
| --- | --- |
| CV files | Existing `new`, `changed` and `same` comparison is retained. Invalid/duplicate identities are excluded. Structured reason and changed-field keys localize explanations; legacy messages remain available to old local consumers. Filename, selected count, profile and nullable index identity remain visible. Original-file upload and function adoption stay separate. |
| Article search | Existing empty/equal/different comparison is retained. Conflicting external fields are unclear; protected fields are excluded. Source URL, score, trace and specialized field/image selectors remain. Vehicle and accessory drafts identify their respective targets. Only the later record save persists the draft. |
| CSV/JSON transfer | Source-field adoption is new for create/copy, different for replace, excluded for invalid/skipped input or reuse of an existing target, and unclear for unresolved review. Existing readiness severity and all conflict decisions remain alongside it. Excluded source fields do not disable the existing reuse/link action and do not claim the records identical. The job names its source and local inventory target. |
| Command-station comparison | Existing equal/deviation/new classifications map directly; missing or conflicting observations remain unclear. The context explicitly describes a read comparison. Existing hardware previews, grants, expiry checks, supported capabilities and explicit write consent remain separate. |

These components do not implement request idempotency, change transaction boundaries or authorize
device writes. Reviewed import jobs, CV identity/source checks and device-specific confirmation remain
the authority for writes. Article search keeps its existing synchronous loading state.
