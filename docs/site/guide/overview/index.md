---
title: Overview, metrics, and data quality
description: Read the RailKeeper dashboard, follow data gaps, and arrange its widgets.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Overview, metrics, and data quality

The **Overview** is RailKeeper's working dashboard for inventory size, recorded value,
digitalization, maintenance, and data quality. This chapter explains what each number means and how
to continue from an indicator to the affected vehicles. It describes stable RailKeeper v0.1.17.6.

Admin, Editor, Viewer, and Planner users can open the overview. An account with only the Messe role
starts in **Exhibition** and cannot open this dashboard.

## Open and refresh the overview

Select **Overview** in the sidebar. RailKeeper loads the vehicles available to the current session
and calculates the dashboard values in the browser. The overview is therefore a current summary of
vehicle data, not a separately stored server report.

Use the refresh icon in the page header after changing vehicle or maintenance data in another
view. The control is disabled while data is loading. If loading fails, RailKeeper shows the error
above the dashboard. Previously loaded values remain visible when available, so check the error
before relying on them.

The Import/Export icon in the same header opens **Import/Export**.

## Read the four summary metrics

| Metric | Meaning |
| --- | --- |
| **Total inventory** | Number of vehicles. The line below reports how many category and gauge groups occur among the five most frequent values in each list. It is not a count of every distinct value when more than five exist. |
| **Digitalization** | Digital vehicles divided by all vehicles, rounded to a whole percentage. The detail line shows the digital and analog counts. |
| **Recorded list value** | Sum of parseable, maintained vehicle list prices. RailKeeper accepts common comma and point number formats and displays the euro total rounded to whole euros. Missing or unparseable prices contribute zero. This is not a market-value estimate. |
| **Maintenance** | Number of incomplete maintenance entries whose due date is today or earlier. The detail line separately shows incomplete entries due in the next 30 days and all incomplete entries, including entries without a due date. |

With no vehicles, inventory and value are zero and percentage metrics show 0%.

## Use the dashboard widgets

The seven widgets below combine inventory signals and shortcuts. Their order can be changed and
each widget can be hidden.

### Inventory mix

**Inventory mix** lists up to five categories with the most vehicles. Missing category values are
grouped as **No category**. The bar shows the category's share of all vehicles, while the number is
the exact vehicle count.

### Data quality

**Data quality** reports the share of all vehicles that meet five individual checks:

| Indicator | Counts as covered when |
| --- | --- |
| **Images** | At least one image is stored for the vehicle |
| **Decoder numbers** | Either decoder-number field contains a value |
| **Article numbers** | The article-number field contains a value |
| **EAN** | The EAN field contains a value |
| **Fully documented** | Article number, EAN, and at least one image are all present |

Every percentage uses the total vehicle count as its denominator. **Decoder numbers** therefore
also includes analog vehicles when they have a value in either decoder-number field. Conversely,
**Fully documented** does not require a decoder number and does not certify that every vehicle
field is complete. With no vehicles, all five percentages are 0%.

### Action needed

**Action needed** lists non-zero data gaps. Select a row to open **Vehicles** with the matching
filter active.

| Gap | Vehicles selected |
| --- | --- |
| **Without main image** | Vehicles without any stored image |
| **Without article no.** | Vehicles without an article number |
| **Without EAN** | Vehicles without an EAN |
| **Digital without decoder no.** | Digital vehicles without a value in either decoder-number field |

In v0.1.17.6, **Without main image** technically checks whether any image exists. It does not
distinguish a specifically selected main image. When all four counts are zero, the widget reports
that no major data gaps were detected.

The overview itself has no general search or filter bar. Searching, combining filters, and editing
records happens in **Vehicles** after following a gap or opening the inventory directly.

### Manufacturers

**Manufacturers** ranks up to five manufacturers by vehicle count. Empty manufacturer values are
grouped as **No manufacturer**.

### Quick actions

**Quick actions** opens the next work area without changing data:

- **Maintain inventory** opens **Vehicles**.
- **Import/Export** opens the import, export, and print area.
- **Check master data** opens **Settings**, where authorized users can manage selection values and
  inventory-number settings.

The destination still applies the signed-in user's permissions. A shortcut does not grant an
additional role.

### Maintenance radar

**Maintenance radar** shows up to four incomplete maintenance entries with a due date, ordered by
the closest due date. Each row contains the vehicle inventory number, vehicle name or maintenance
kind, maintenance kind, due-distance label, and date. Entries due today or overdue are highlighted.

The lower row summarizes:

- **Done**: every completed maintenance entry.
- **Cost**: the sum of parseable costs from all maintenance entries, including completed entries.
- **Conditions**: how many distinct condition ratings are represented among the five most frequent
  ratings, so the displayed value is capped at five.

If no incomplete maintenance entry has a due date, the radar shows its empty-state message. Open
maintenance without a due date still contributes to the summary metric at the top of the page.

### Next value

**Next value** chooses one recommendation from the current data in this order:

1. Create or import vehicles when the inventory is empty.
2. Add images while image coverage is below 70%.
3. Work through due maintenance when at least one item is due today or overdue.
4. Consider spare-parts and structured price/value maintenance when none of the earlier conditions
   applies.

This is a fixed rule-based hint, not an automated data assessment or external recommendation.

## Work from data gaps

1. Open **Action needed** on the overview.
2. Select the relevant gap.
3. RailKeeper opens **Vehicles** and activates the corresponding inventory or quality filter.
4. Open the affected records and add or correct the missing data.
5. Return to **Overview** and use refresh to recalculate the dashboard.

Resetting all filters in **Vehicles** removes the gap parameter from the browser address. Selecting
another quality filter also removes it. Other manual filter changes can leave the original
parameter in the address even though the visible selection has changed.

## Arrange the dashboard

Every widget header contains three controls:

- Move up places the widget one position earlier.
- Move down places it one position later.
- Hide removes the widget from the dashboard.

After at least one widget is hidden, the reset icon appears in the page header. **Reset layout**
shows all seven widgets again and restores their default order. If all widgets are hidden, an empty
dashboard panel also provides the reset button.

Widget order and hidden state are stored in the current browser's local storage. They are not
shared with other browsers or saved as account-wide dashboard settings.

## Empty and exceptional states

| Situation | What RailKeeper shows | What to do |
| --- | --- | --- |
| No vehicles | Zero summary values, empty lists, and the create/import recommendation | Create vehicles or import an inventory list |
| No due-dated open maintenance | An empty maintenance-radar message | Add a due date if the work should appear in the radar |
| No major data gaps | A confirmation in **Action needed** | Continue with maintenance or other inventory details |
| All widgets hidden | **Dashboard empty** with **Reset layout** | Reset the layout to restore all widgets |
| Vehicle loading fails | Error text above the dashboard | Check the connection and session, then refresh again |

## Related pages

- [User Guide overview](/guide/)
- [First setup and sign-in](/guide/getting-started/)
- [Administration overview](/administration/)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.17.6** and was last reviewed on 2026-08-16.
