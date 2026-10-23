---
title: Review search suggestions and extract spare parts
description: Review search suggestions and extract spare parts.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Review search suggestions and extract spare parts

**Goal:** Deliberately select suggested fields and try a local spare-parts list.
Applying information requires **Admin** or **Editor** and a saved test vehicle.
Search replies below are browser simulations; TXT extraction uses the actual local importer.

## 1. Start with suitable search criteria

Save other vehicle changes first. Article search needs suitable criteria such as manufacturer,
gauge and article number. Check that article web search is enabled under **Settings > General**
and choose relevant source classes.

Open the vehicle and select **Search article data**. For real results, inspect the original source,
exact article number, gauge and variant. Ranking and confidence do not replace this check.

<GuideFigure name="article-result-review" alt="Search dialog comparing an existing name with a different synthetic suggestion." caption="Actual search UI with controlled simulated results from example.invalid. This establishes comparison and selection, not search-provider quality." />

A differing existing name is a conflict and initially unselected. Select only fields you intend
to replace or fill. **Apply selected fields** changes the vehicle draft. **Save changes** persists
those vehicle details. **Cancel** discards the unsaved draft.

## 2. Use a barcode as a search aid

Select **Search barcode**. Type a code, use a keyboard scanner or transfer it from a scanner app.
Camera input needs browser support, a secure context, normally HTTPS or localhost, and your permission.

<GuideFigure name="article-barcode" alt="Barcode dialog containing a synthetic EAN before searching." caption="Manual entry was tested in the demo. This is not physical camera or scanner acceptance." />

Check the code before searching. It is used as **EAN**; article number remains separate.
Check the exact model variant even when an EAN matches.

## 3. No results or unsuitable sources

<GuideFigure name="article-no-results" alt="Search dialog with no results for the controlled example response." caption="Simulated empty state in the actual interface. An empty result does not prove the product does not exist." />

Check spelling, article number and source selection. Remove unsuitable extra terms.
For errors, retain the specific message. Do not replace reliable information with a merely similar product.

## 4. Keep a web document locally

Use the vehicle's **Web documents** tools under Uploads. Review source, title, article match and
category before importing. A saved link alone is not a locally available attachment.
After import, open the saved file and inspect its content.
The [detailed guide](search-and-spares#find-and-import-web-documents) covers selection, import,
errors and separate save timing. This picture sequence contains no actual web download.

## 5. Try local spare-parts extraction

<GuideDownload file="spare-parts-demo.txt">Download the synthetic local spare-parts list</GuideDownload>

1. Create a test vehicle with article number **99001**, an exercise number, not a product recommendation.
2. Open **Uploads** and upload the TXT as an attachment. Choose **Ersatzteilliste** as its category.
3. Check filename and contents. The list explicitly belongs to model article 99001.

<GuideFigure name="spare-document" alt="TXT spare-parts list saved as an attachment on the exercise vehicle." caption="Actual uploaded synthetic TXT. Internet, manufacturer downloads and OCR are not needed." />

4. Save or discard other vehicle drafts first.
5. Select **Extract spare parts** on this file.

::: warning This action saves immediately
The action saves attachment details, creates newly recognized spare parts and reloads the vehicle.
It does not wait for **Save changes**. After a partial failure, inspect already saved parts before retrying.
:::

<GuideFigure name="spare-extraction-result" alt="Two spare parts actually created from the TXT in the vehicle editor." caption="Actual local extraction created 99001-01 and 99001-02. No manufacturer facts or prices were invented." />

**Expected:** `99001-01` for the coupler and `99001-02` for the wheelset, without prices.
Review and correct descriptions where necessary. This exercise does not establish OCR quality for
scanned PDFs. Without a matching article number or readable text, there may be no useful suggestion.
Use manual spare-part entry when necessary.

[Detailed search, documents and spare-parts reference](search-and-spares).
