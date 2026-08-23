# Responsive ECoS Adoption Actions

## Goal

Keep every action in the locomotive comparison dialog fully visible at narrow desktop and mobile
widths. In particular, `Neues Fahrzeug anlegen` must no longer extend beyond the left dialog edge.

## Approved Design

For an ECoS locomotive without a RailKeeper assignment, the comparison footer shows only these
actions:

1. Create a new RailKeeper vehicle.
2. Assign the locomotive to an existing RailKeeper vehicle.
3. Close the dialog.

The disabled write-preview action is hidden in this state because writing cannot start before a
vehicle assignment exists. The explanatory text below the footer continues to state this boundary.

The footer may wrap its remaining actions when the available width is insufficient. Wrapped buttons
stay inside the dialog, keep their natural content width on larger screens, and allow long German or
English labels to wrap without clipping. Assigned locomotives retain the existing write-preview flow.

## Scope

- Adjust the conditional rendering in `LocomotiveComparisonDialog`.
- Add narrowly scoped responsive footer styles in `digital-centers.css`.
- Extend the comparison-dialog tests for hidden write preview and non-clipping responsive behavior.
- Keep all ECoS read, assignment, vehicle creation, and write APIs unchanged.

## Verification

- Component tests verify that an unassigned locomotive shows the two adoption actions and close,
  while omitting write preview.
- Existing assigned-locomotive write-preview tests remain green.
- The frontend production build succeeds.
- Browser QA checks the original narrow viewport and a normal desktop viewport, including console
  health and the comparison-to-assignment interaction.
