---
title: Overview, metrics, and data quality
description: Read the RailKeeper dashboard, follow data gaps, and arrange its widgets.
audience: user
status: stable
reviewedVersion: 0.1.20.3
lastReviewed: 2026-08-16
---

# Overview, metrics, and data quality

The **Overview** is RailKeeper's working dashboard for inventory size, recorded value,
digitalization, maintenance, and data quality. This chapter explains what each number means and how
to continue from an indicator to the affected vehicles. It describes stable RailKeeper v0.1.20.3.

Admin, Editor, Viewer, and Planner users can open the overview. An account with only the Messe role
starts in **Exhibition** and cannot open this dashboard.

Only Admin and Editor users can change vehicle data. Viewer and Planner users can inspect the
dashboard and filtered results, but must ask an Admin or Editor to correct missing data.

## Open and refresh the overview

Select **Overview** in the sidebar. RailKeeper loads the vehicles available to the current session
and requests the current vehicle and accessory valuation totals from the server. The overview is a
current summary of stored inventory data, not a separately stored report.

Use the refresh icon in the page header after changing vehicle or maintenance data in another
view. The control is disabled while data is loading. If vehicle loading fails, RailKeeper shows the
error above the dashboard. If only the valuation request fails, the other metrics remain available
and the valuation area shows its own error. Previously loaded vehicle values remain visible when
available, so check the error before relying on them.

The Import/Export icon in the same header opens **Import/Export**.

## Read the four summary areas

| Metric | Meaning |
| --- | --- |
| **Total inventory** | Number of vehicles. The line below reports how many category and gauge groups occur among the five most frequent values in each list. Missing values form the **No category** and **No gauge** groups. It is not a count of every distinct value when more than five exist. |
| **Digitalization** | Digital vehicles divided by all vehicles, rounded to a whole percentage. The detail line shows the digital and analog counts. |
| **Recorded inventory values** | Four cent-exact euro totals: vehicle list value, vehicle purchase price, accessory list value, and accessory purchase costs. The values are kept separate rather than combined into one total. |
| **Maintenance** | Number of incomplete maintenance entries whose due date is today or earlier. The detail line separately shows incomplete entries due in the next 30 days and all incomplete entries, including entries without a due date. |

Vehicle totals add the maintained list and purchase prices once per vehicle. Accessory list value
multiplies each maintained unit list price by the currently owned quantity. Accessory purchase costs
use recorded euro purchase quantities and unit prices and also include manually maintained purchase
prices for individual assets that are not linked to a purchase. Purchases explicitly recorded in a
foreign currency are excluded and reported below the values.

RailKeeper accepts common comma and point number formats. Missing or unparseable prices contribute
zero. These totals describe recorded acquisition data and are not market-value estimates. With no
vehicles, vehicle inventory, vehicle values, and percentage metrics are zero; accessory values can
still be present.

## Use the dashboard widgets

The seven widgets below combine inventory signals and shortcuts. Their order can be changed and
each widget can be hidden.

### Inventory mix

**Inventory mix** lists up to five categories with the most vehicles. Missing category values are
grouped as **No category**. The bar shows the category's share of all vehicles, while the number is
the exact vehicle count. A non-zero bar has a minimum visible width of 8%, so small categories can
look wider than their exact percentage.

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

**Action needed** lists non-zero data gaps. Select a row to open **Vehicle inventory**, whose page
heading is **Inventory**, with the matching filter active.

| Gap | Vehicles selected |
| --- | --- |
| **Without main image** | Vehicles without any stored image |
| **Without article no.** | Vehicles without an article number |
| **Without EAN** | Vehicles without an EAN |
| **Digital without decoder no.** | Digital vehicles without a value in either decoder-number field |

In v0.1.20.3, **Without main image** technically checks whether any image exists. It does not
distinguish a specifically selected main image. When all four counts are zero, the widget reports
that no major data gaps were detected.

The overview itself has no general search or filter bar. Searching, combining filters, and editing
records happens in **Vehicle inventory** after following a gap or opening the inventory directly.

### Manufacturers

**Manufacturers** ranks up to five manufacturers by vehicle count. Empty manufacturer values are
grouped as **No manufacturer**.

### Quick actions

**Quick actions** opens the next work area without changing data:

- **Maintain inventory** opens **Vehicle inventory**.
- **Import/Export** opens the import, export, and print area.
- **Check master data** opens **Settings**, where authorized users can manage selection values and
  inventory-number settings.

The destination still applies the signed-in user's permissions. A shortcut does not grant an
additional role.

### Maintenance radar

**Maintenance radar** shows up to four incomplete maintenance entries with a due date, ordered by
the earliest due date. The oldest overdue entry therefore appears first, followed by later overdue,
today's, and upcoming entries. Each row contains the vehicle inventory number, vehicle name or
maintenance kind, maintenance kind, due-distance label, and date. Entries due today or overdue are
highlighted.

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
Creating, importing, or changing vehicles requires the Admin or Editor role.

## Work from data gaps

1. Open **Action needed** on the overview.
2. Select the relevant gap.
3. RailKeeper opens **Vehicle inventory** and activates the corresponding inventory or quality
   filter.
4. Open the affected records. As an Admin or Editor, add or correct the missing data. As a Viewer
   or Planner, use the result to identify the records and ask an Admin or Editor to update them.
5. Return to **Overview** and use refresh to recalculate the dashboard.

For an article-number, EAN, or decoder gap, clearing the active quality-filter pill removes the gap
parameter from the browser address. Changes in other filter groups leave that quality gap active and
retain the parameter. For **Without main image**, selecting another inventory filter replaces the
active image gap but leaves the now-stale `gap=no-main-image` parameter in the address. **Clear
filters** removes the parameter for every gap.

## Arrange the dashboard

Every widget header contains three controls:

- **Move forward** places the widget one position earlier.
- **Move backward** places it one position later.
- Hide removes the widget from the dashboard.

After at least one widget is hidden, the reset icon appears in the page header. **Reset layout**
shows all seven widgets again and restores their default order. If all widgets are hidden, an empty
dashboard panel also provides the reset button.

Widget order and hidden state are stored in the current browser's local storage. They are not
shared with other browsers or saved as account-wide dashboard settings.

## Empty and exceptional states

| Situation | What RailKeeper shows | What to do |
| --- | --- | --- |
| No vehicles | Zero summary values, empty lists, and the create/import recommendation | As an Admin or Editor, create vehicles or import a list. As a Viewer or Planner, ask one of those roles to populate the inventory. |
| No due-dated open maintenance | An empty maintenance-radar message | Add a due date if the work should appear in the radar |
| No major data gaps | A confirmation in **Action needed** | Continue with maintenance or other inventory details |
| All widgets hidden | **Dashboard empty** with **Reset layout** | Reset the layout to restore all widgets |
| Vehicle loading fails | Error text above the dashboard | Check the connection and session, then refresh again |
| Valuation loading fails | Error text in **Recorded inventory values** while the other metrics remain available | Check the connection and session, then refresh again |

## Related pages

- [User Guide overview](/guide/)
- [Accessories overview](/guide/accessories/)
- [Vehicle maintenance and condition](/guide/vehicles/maintenance)
- [First setup and sign-in](/guide/getting-started/)
- [Administration overview](/administration/)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.20.3** and was last reviewed on 2026-08-16.
