# Accessory Search Enrichment Design

## Goal

Extend the accessory editor with structured subject-data suggestions from article web search,
automatic scale selection from gauge master data, and useful keyword defaults. The behavior must
preserve manual user input and remain local to the accessory search fix branch.

## Scope

### Structured track data from web search

For articles of type `track`, the article search request includes the article type, subtype, and
the current track subject values as search context. Search enrichment may return the following
canonical fields:

- `trackSystem`
- `lengthMm`
- `radiusMm`
- `angleDegrees`
- `direction`
- `frogAngleDegrees`
- `sleeperType`
- `railHeightMm`
- `roadbed`
- `connectionCount`
- `digitalReady`

The article search dialog presents these values in a separate `Fachangaben: Gleis` field group.
Every value remains a suggestion and must be selected before it is applied. The frontend validates
the returned value against the existing track field definition before converting it to an
`AccessoryAttributeValue` or numeric draft. Invalid numbers, booleans, or option values cannot be
selected or imported. Existing subject values are shown as conflicts and are not selected by
default.

Search extraction should use structured page data and clearly labelled product specifications.
It must not infer unsupported values from unrelated prose. Existing URL, timeout, redirect, MIME,
and size protections remain unchanged.

## Automatic scale behavior

Gauge master data already contains a `scale` metadata value. Selecting the first gauge fills the
article scale from that active gauge entry.

The scale remains auto-managed while it is empty or still equals the last automatically derived
value. A later gauge change updates an auto-managed scale. As soon as the user edits the scale
field manually, subsequent gauge changes leave it unchanged. When several gauges are selected,
the first selected gauge determines the automatic scale.

## Automatic keyword behavior

For a new article with an initially empty keyword field, RailKeeper builds a comma-separated,
case-insensitively deduplicated list from:

- the complete article designation,
- the manufacturer label,
- the localized article-type label,
- the localized subtype label.

The generated list stays synchronized while those source fields change. The first manual edit of
the keyword field disables automatic synchronization for the current editor session. Existing
articles with stored keywords start in manual mode. An existing article with no keywords starts in
automatic mode, so missing keywords can be populated without overwriting stored content.

## Component boundaries

- Pure accessory editor helpers derive scale and keyword defaults and convert search fields to
  typed subject attributes.
- `ArticleCoreTab` reports user changes without owning synchronization state.
- `ArticleEditorDialog` owns per-editor-session auto-management state and composes patches before
  forwarding them to the existing controller.
- The accessory article-search adapter supplies request context, field groups, selection
  validation, current-value comparison, and result application.
- Backend article-search enrichment extracts only canonical track fields and returns them through
  the existing `fields` map, so the public response shape stays compatible.

## Error handling and compatibility

- Missing gauge scale metadata leaves the scale unchanged.
- Unknown or inactive gauges remain non-importable, matching existing master-data behavior.
- Unsupported, malformed, or ambiguous subject values remain visible only when safe to display,
  but cannot be applied.
- Manual scale and keyword values always win over automatic defaults.
- Non-track accessory types keep their current article-search behavior.
- No database migration or API response schema change is required.

## Verification

Tests cover:

- extraction and normalization of representative Tillig track specifications,
- search request context and the additional track field group,
- safe conversion of text, number, boolean, and direction suggestions,
- rejection of malformed or unsupported subject values,
- scale derivation, gauge changes, and manual scale preservation,
- keyword synchronization, deduplication, and manual override,
- the existing article-search popup and keyboard interaction regressions,
- the complete frontend suite, backend suite, frontend production build, and `git diff --check`.
