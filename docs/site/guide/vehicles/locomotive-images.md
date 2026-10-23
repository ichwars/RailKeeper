---
title: Locomotive images for a command station
description: Locomotive images for a command station.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Locomotive images for a command station

**Goal:** Prepare a vehicle picture for your command station and optionally assign it directly
to an ECoS locomotive. The original picture is retained. Direct transfer requires **Admin**;
**Admin** and **Editor** can maintain vehicle pictures.

These captures show the actual RailKeeper interface and a synthetic locomotive drawing.
Station replies in the ECoS preview, pause and result pictures are simulated.
They explain the workflow and do not establish additional hardware acceptance.

## 1. Choose a suitable picture

Open a saved vehicle under **Edit > Uploads** and add a vehicle picture.
A side view with a calm background works well. Keep the whole locomotive visible and leave
little unused space. Enlarging cannot repair an unsuitable perspective or a blurred photograph.

Try these synthetic image files:

- <GuideDownload file="locomotive-demo.png">Demo locomotive at sufficient resolution</GuideDownload> (1200 × 400 pixels).
- <GuideDownload file="locomotive-low-resolution-demo.png">Low-resolution demo locomotive</GuideDownload> (40 × 14 pixels).

On the chosen vehicle picture, select **Create command station image**.

## 2. Choose the target and check crop and actual size

Choose **Target device / application**. Review both sides of the dialog: the source and crop
on the left, **Result at actual size** on the right.

<GuideFigure name="image-export" alt="Image dialog with crop on the left and the prepared ECoS result on the right." caption="The actual image dialog processes the larger demo file. The right preview shows the target's real output size. Synthetic drawing; click for full size." />

1. Check the suggested crop.
2. Use **Remove outer margins** when uniform empty borders get in the way.
3. Drag a new crop with mouse or touch if needed. Use **Set crop in pixels** for precise values.
   **Whole image** resets the selection.
4. Keep buffers, wheels and roof details visible while using as much width as possible.
5. Select a suitable background. This does not remove a photographed background.

RailKeeper scales within the target dimensions while preserving proportions.
A tall subject cannot fill the full width and remain entirely visible at the same time.
Improve the photograph or crop rather than stretching the locomotive.

| Target | Output and next step |
| --- | --- |
| ECoS 50200 / 50210 / 50220 | 190 × 40 pixels, 24-bit RGB BMP; fixed device format. Import manually or use confirmed direct transfer. |
| CS3 / CS3 plus | PNG, RailKeeper preset 384 × 132 pixels; import through the CS3 web interface. |
| Intellibox 3 | PNG, 340 × 72 preset; process and transfer using IB3 Configuration. |
| Z21 app | PNG, 960 × 400 preset; save on the mobile device and select for the vehicle in the app. |

PNG presets are not universal manufacturer limits. The Z21 station itself does not store these
app pictures. **Manufacturer information** in the dialog links to the relevant source.

## 3. Understand the suitability notice

**Technically usable** means the file can be processed for the selected target.
Still check that it shows the correct locomotive clearly. RailKeeper cannot reliably assess
perspective, sharpness or subject identity.

<GuideFigure name="image-low-resolution" alt="Warning for a very small source picture of the demo locomotive." caption="The 40 × 14 pixel example needs enlargement. The warning explains its limits; adding output pixels cannot recover missing detail." />

Warnings may concern enlargement, poor width use or excessive borders. They require review
and do not always block export. An invalid crop or unreadable image prevents output.
Use the larger original where resolution is insufficient.

## 4. Download and import manually

Select **Download image**. For ECoS, use the generated BMP rather than the original PNG.
Open its web interface, go to locomotive images and upload to a free user-image slot.
Then assign it to the correct locomotive on the ECoS. Uploading alone does not confirm assignment.

Use the table's workflow for CS3, Intellibox 3 and the Z21 app.
The direct-transfer procedure below is available only for ECoS.

## 5. Transfer directly to ECoS

The ECoS must be [configured and reachable](../../administration/digital-centers).
The protocol connection needs TCP port **15471**; image transfer also needs its web interface
on port **80**, both reachable from the RailKeeper server.

1. Select **Transfer directly to ECoS …**.
2. Choose the **Target locomotive on ECoS**. Check name, decoder address, protocol and object ID.
3. Review the description, then select **Check transfer and load preview**.
4. Compare the current and new pictures and the chosen slot. This preview does not upload
   a new picture yet.

<GuideFigure name="image-ecos-preview" alt="ECoS target selection with before/after pictures and an unconfirmed transfer." caption="Actual RailKeeper preview with simulated station replies and reserved example address 192.0.2.10. No real ECoS is contacted." />

5. Deselect the target on both ECoS driving displays and in other control applications.
6. Check **Target locomotive and image preview checked**, then select **Upload and assign now**.
7. Wait for the result. RailKeeper requests locomotive control first, then verifies the picture
   and assignment by reading them back after transfer.

If an identical picture already exists, the preview shows its slot and offers
**Assign existing image now**. It does not upload a second picture.
The existing ECoS description is retained.

## 6. Resume when locomotive control is busy

<GuideFigure name="image-ecos-pause" alt="Paused image transfer with a locomotive-control notice and Continue button." caption="Simulated busy-control case in the actual dialog. No new image or assignment has been written at this point." />

Deselect the locomotive on both driving displays and in other control applications.
Then deliberately select **Continue**. RailKeeper checks a fresh preview and continues the
previously confirmed operation if target, image and slot are unchanged. Changes require another
preview confirmation. It does not force locomotive-control takeover.

## 7. Check the outcome and handle uncertain replies

<GuideFigure name="image-ecos-result" alt="Success notice after verified image transfer and locomotive assignment." caption="Simulated success result in the actual RailKeeper dialog. Image and assignment are confirmed separately; this capture is not a real-device test." />

| Notice | Next step |
| --- | --- |
| Picture and assignment verified | Also check the picture on the ECoS. |
| Web interface unreachable | Check address, port 80 and reachability from the RailKeeper server. A successful protocol test alone is insufficient. |
| Preview expired or changed | Review again and confirm the current preview deliberately. |
| Picture verified, assignment unconfirmed | Check the slot and locomotive on ECoS. If offered, recheck only the assignment and reuse the existing image. |
| Missing reply or uncertain state | Do not blindly upload again. Check the image slot and locomotive assignment first. |

A late assignment failure does not automatically roll back an earlier upload.
Keep the downloaded file until the outcome is established.
