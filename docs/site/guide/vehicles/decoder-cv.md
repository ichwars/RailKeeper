---
title: Decoder, functions, and CV data
description: Decoder, functions, and CV data.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Decoder, functions, and CV data

Choose the task you need. Practical guides use synthetic data and show expected results.
None of these exercises programs a decoder.

| Task | Guide |
| --- | --- |
| Import CVs and check index pages | [Try CV import](cv-import-example) |
| Check model, firmware and identification | [Identify a decoder](decoder-identification) |
| Fill F0/F1/F3 together and protect manual details | [Apply function keys](functions-walkthrough) |
| Look up a model or source | [Decoder library](decoder-library) |
| Understand manufacturer and firmware limits | [Model notes](decoder-model-notes) |
| Read formats, raw-value rules and edge cases | [Technical reference](decoder-cv-reference) |

<GuideFigure name="cv-analysis" alt="Plain CV explanations with readings and source links." caption="The actual demo shows interpreted readings; individual technical details open through a CV source link." />

## The combined workflow

1. **CV**: Import readings; check profile, firmware and index pages.
2. **Model > Digital**: Compare interpreted readings with stored vehicle details.
3. **Control**: Fill missing details together; choose missing symbols and your own names.
4. **Save changes**: Save the vehicle and function draft together.
5. Reopen the vehicle and check results and provenance.

CV imports and file actions save their own data and may reload the vehicle. Deliberately save
or discard other drafts first. **Cancel** does not undo an already completed CV import.

## When information remains unknown

The library interprets readings within documented model, firmware and index limits.
CV8 alone identifies the manufacturer. Missing index pages cannot be reconstructed.
A manufacturer reference does not automatically establish a calculated vehicle function.

## Existing detail links

Existing links remain accessible. Choose the corresponding reference section below.

<span id="decoder-library"></span>

- [Decoder library](decoder-cv-reference#decoder-library)

<span id="prerequisites-and-access-rights"></span>

- [Prerequisites and access rights](decoder-cv-reference#prerequisites-and-access-rights)

<span id="map-digital-functions-f0-f31"></span>

- [Map digital functions F0-F31](decoder-cv-reference#map-digital-functions-f0-f31)

<span id="import-and-export-functions"></span>

- [Import and export functions](decoder-cv-reference#import-and-export-functions)

<span id="read-the-speed-curve"></span>

- [Read the speed curve](decoder-cv-reference#read-the-speed-curve)

<span id="manage-cv-values-manually"></span>

- [Manage CV values manually](decoder-cv-reference#manage-cv-values-manually)

<span id="railcomplus-registration-name-and-function-symbols"></span>

- [RailComPlus registration name and function symbols](decoder-cv-reference#railcomplus-registration-name-and-function-symbols)

<span id="files-for-a-complete-import"></span>

- [Files for a complete import](decoder-cv-reference#files-for-a-complete-import)

<span id="import-and-export-cv-values"></span>

- [Import and export CV values](decoder-cv-reference#import-and-export-cv-values)

<span id="preview-apply-and-store-decoder-files"></span>

- [Preview, apply, and store decoder files](decoder-cv-reference#preview-apply-and-store-decoder-files)

<span id="use-an-ecos-preview-as-an-input-path"></span>

- [Use an ECoS preview as an input path](decoder-cv-reference#use-an-ecos-preview-as-an-input-path)

<span id="protect-data-during-writes"></span>

- [Protect data during writes](decoder-cv-reference#protect-data-during-writes)

<span id="troubleshoot-decoder-data"></span>

- [Troubleshoot decoder data](decoder-cv-reference#troubleshoot-decoder-data)

<span id="related-pages"></span>

- [Related pages](decoder-cv-reference#related-pages)

<span id="digital-settings-in-the-vehicle-dialog"></span>

- [Digital settings in the vehicle dialog](decoder-model-notes#digital-settings-in-the-vehicle-dialog)

<span id="kuhn-n45-reviewed-interpretation"></span>

- [Kühn N45: reviewed interpretation](decoder-model-notes#kuhn-n45-reviewed-interpretation)

<span id="uhlenbrock-intellidrive-76420"></span>

- [Uhlenbrock IntelliDrive 76420](decoder-model-notes#uhlenbrock-intellidrive-76420)

<span id="roco-10882-2012-manual-zimo"></span>

- [Roco 10882 (2012 manual, ZIMO)](decoder-model-notes#roco-10882-2012-manual-zimo)

<span id="piko-56500-plux22"></span>

- [PIKO 56500 PluX22](decoder-model-notes#piko-56500-plux22)

<span id="d-h-dh05c-and-dh10c"></span>

- [D&H DH05C and DH10C](decoder-model-notes#d-h-dh05c-and-dh10c)

<span id="d-h-dh16a"></span>

- [D&H DH16A](decoder-model-notes#d-h-dh16a)

<span id="d-h-dh12a-historical-plux12-decoder"></span>

- [D&H DH12A (historical PluX12 decoder)](decoder-model-notes#d-h-dh12a-historical-plux12-decoder)

<span id="d-h-dh14b-mtc14"></span>

- [D&H DH14B (mTc14)](decoder-model-notes#d-h-dh14b-mtc14)

<span id="d-h-dh21a-generations-and-connection-variants"></span>

- [D&H DH21A (generations and connection variants)](decoder-model-notes#d-h-dh21a-generations-and-connection-variants)

<span id="additional-d-h-models-and-manufacturer-references"></span>

- [Additional D&H models and manufacturer references](decoder-model-notes#additional-d-h-models-and-manufacturer-references)
