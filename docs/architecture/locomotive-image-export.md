# Locomotive images for command stations

The vehicle Uploads tab exposes one preparation dialog per source image. The browser reads the
local authenticated original, checks technical suitability and downloads a separate file. Original
images and vehicle data are unchanged. Existing source-image authorization applies. A separate,
Admin-only [ECoS transfer flow](ecos-image-transfer.md) offers a confirmed device write after preview;
creating/downloading an image alone never writes to a device. No backup fields or runtime libraries
are added.

## Target evidence

| Target | Output | Evidence and scope |
| --- | --- | --- |
| ECoS 50200 / 50210 / 50220 | 190 × 40, uncompressed RGB24 BMP | [Original ESU image addendum](https://www.esu.eu/uploads/tx_esudownloads/50200_ECoS_ESU_KG_Handbuch-Addendum_I_LokBilder_August_2010_eBook.pdf), August 2010, p.3; [manufacturer download directory](https://www.esu.eu/downloads/instruction-manuals/digital-systems/) groups all three models and links the addendum. 50200 needs firmware 3.3.0 onward. Older monochrome 50000/CS1 Reloaded is outside this requested scope. |
| CS3 / CS3 plus | PNG, RailKeeper preset 384 × 132 | [Original quick guide for software 2.5 onward](https://static.maerklin.de/damcontent/c1/ce/c1ce554ad6ad195c04e32ba1b4dd64511764068661.pdf), p.37: web image upload, 5 MB maximum and an editor. No fixed input dimensions; our preset is not a manufacturer limit. |
| Intellibox 3 | PNG, 340 × 72, for IB3 Configuration | [Manufacturer software page](https://www.uhlenbrock.de/de_DE/service/download/software/I4339F6A-001.htm!ArcEntryInfo=0004.24.I4339F6A), package 1.3.03 of 22 September 2026. Original release notes document an image editor, scaling, transparency, native conversion and upload. Its image picker accepts PNG/JPG/BMP; all 135 supplied locomotive PNGs are 340 × 72. This is a preparation preset based on examples, not a direct native IB3 file or a proven hardware size limit. The manufacturer tool converts and transfers it. |
| Z21 app | PNG, RailKeeper preset 960 × 400 | [Manufacturer FAQ, section 1.18](https://www.z21.eu/de/faq-support/fragen-antworten) locates vehicle data in the app; [original z21 start guide](https://www.z21.eu/media/Kwc_Basic_DownloadTag_Trl_Component.Kwc_Basic_DownloadTag_Component/root-en-main_51-1065-1662-downloadTag/default/f09b1787/1734536724/8051282921-xii-2024-web.pdf), pp.25–26, describes own pictures and gallery/camera selection. This is a gallery image, not a Z21-protocol upload or manufacturer pixel requirement. |

Sources reviewed on 21 October 2026. Uhlenbrock's package was extracted without executing Windows
software. SHA-256: `9ca92d91788ba636fa73824664d361ff3ecb9fb0a19d1ab096df97ff1256cdd2`.
The original `Config/IB3_Release Notes.txt` inside that package supplies the release notes. A direct
notes link returned another binary and was not treated as text. Vendor software and pictures remain
ignored research artifacts and are not bundled. The user has confirmed manual ECoS upload of a
RailKeeper export. Direct RailKeeper-to-ECoS transfer requires its own hardware acceptance.

## Processing and limits

- Only same-origin HTTP(S) and same-origin blob URLs are fetched. External article links first need
  local import/upload. There is no new remote-fetch proxy or external image/AI service.
- Source streams are bounded to 20 MiB and PNG/JPEG/WebP signatures are checked. Existing backend
  upload checks remain. Browser preparation caps each side at 12,000 pixels and the image at 40 million
  pixels. PNG dimensions are additionally checked before bitmap decode.
- Crop analysis is bounded to 1200 × 1200 pixels. Alpha bounds or uniform outer-background
  bounds produce the initial visible crop, with a one-source-pixel safety margin. Opaque
  trimming requires agreeing corners and at least 85% matching edge pixels (per-channel
  tolerance 16), then floods only matching pixels connected to the edge. Ambiguous
  photographic backgrounds and blank opaque images retain the full image. “Whole image”
  restores the original crop; the crop suggestion is reviewable and manually adjustable.
  Fully transparent images cannot export. No subject, sharpness or perspective detection is claimed.
- Pointer selection and keyboard pixel fields define a bounded crop. Aspect ratio is retained and
  the whole crop is fitted into the full available target width/height without an extra
  export margin, distortion or clipping. Tall crops remain height-limited. Enlarging,
  poor target-area use and large outer borders produce specific warnings. Technical
  usability does not certify the subject.
- Actual-size preview scrolls inside its bounded frame. Background fills transparency/spare margins;
  BMP always has an opaque background. Grey is a RailKeeper display choice, not a mandatory vendor colour.
- BMP uses a 40-byte Windows DIB header, bottom-up BGR pixels and four-byte row alignment, with
  24-bit colour and no compression/alpha. ECoS pixels are prepared with the observed
  firmware 4.3.3 palette and integer diffusion described in the transfer document; the
  actual-size preview shows those exported pixels. Other PNG targets keep their normal
  colour processing. Browser-encoded PNG is bounded below 5 MiB.
- Target/crop/background changes invalidate the previous download. Stale async encodes are discarded;
  object URLs and bitmaps are released on replacement/close. Filenames are sanitized and target-suffixed.

The existing portal/modal focus helper makes the vehicle dialog temporarily inert, contains keyboard
focus and Escape, and restores focus to the initiating action. Download never submits the vehicle form
or persists unsaved edits. Original image metadata, blob storage and backup behavior are unchanged.
