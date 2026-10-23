---
title: Locomotive images for a command station
description: Locomotive images for a command station.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Locomotive images for a command station

A vehicle photo may be too large or have wide borders. RailKeeper prepares a separate file for the selected command station. Your original stays intact.


<GuideFigure name="media" alt="Start with the pictures under Uploads; add a suitable photo first." caption="Start with the pictures under Uploads; add a suitable photo first. Synthetic demo data, not factory settings. Click for full size." />

## Prepare an image

1. Open the saved vehicle under **Edit > Uploads**.
2. Select **Create command-station image** on an existing vehicle image.
3. Choose the target, for example **ECoS**.
4. Check the proposed crop. The locomotive should use the available width without being clipped or distorted.
5. Check the actual-size preview and download the prepared file.

RailKeeper checks file format, resolution, crop, and area usage. It does not reliably establish that the picture shows the right locomotive, is sharp, or uses a suitable perspective. Check those aspects yourself.

## Where does the file go?

| Target | Use |
| --- | --- |
| ECoS 50200 / 50210 / 50220 | BMP for manual import or the separately confirmed direct transfer |
| CS3 / CS3 plus | PNG for the web interface's image import |
| Intellibox 3 | PNG for further processing in IB3 Configuration |
| Z21 app | PNG for the app's image picker, not a protocol upload to the Z21 |

## Transfer directly to an ECoS

Only Admins can use **Transfer directly to ECoS**. Select the target locomotive and check its name, decoder address, object ID, and image preview. Confirm only after reviewing them. RailKeeper can reuse an identical existing image.

If locomotive control is occupied, the operation pauses. Deselect the locomotive on the ECoS, then deliberately choose **Continue**. Changed previews need a new confirmation. After an ambiguous reply, check the image slot and locomotive assignment rather than repeatedly uploading.

The ECoS connection uses port 15471; image transfer also needs the web interface on port 80. [Configure the station](../../administration/digital-centers).
