# RailKeeper Documentation Logo Design

## Goal

Replace the stylized placeholder logos on both documentation landing pages with the existing
official RailKeeper assets. The change must preserve the current bilingual structure and responsive
VitePress layout.

## Considered approaches

1. Use the full logo everywhere. This is visually consistent, but the wordmark becomes unreadable
   in the compact navigation slot and duplicates the adjacent site title.
2. Use only the train mark everywhere. This fits the navigation, but the prominent hero area would
   still omit the official RailKeeper wordmark.
3. Use each official asset according to its role. This keeps the navigation compact and gives the
   hero the complete brand presentation.

The third approach is selected.

## Design

- Navigation uses `/brand/railkeeper-mark.png`, the official train signet without the wordmark.
- English and German hero sections use `/brand/railkeeper-logo.png`, the complete official logo.
- The favicon follows the navigation mark for consistent browser branding.
- Existing headings, buttons, navigation, translations, colors, and layout remain unchanged.
- Existing responsive image constraints remain in place unless visual verification shows clipping
  or distortion. Any sizing adjustment must stay limited to the documentation theme.

## Verification

- Run the complete documentation check and production build.
- Inspect English and German landing pages in light and dark mode.
- Check desktop and mobile widths for correct proportions, no clipping, and no horizontal overflow.
- Confirm both assets load from the repository and no external image request is introduced.
