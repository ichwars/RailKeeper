---
title: Vehicle images and attachments
description: Upload, organize, preview, download, and safely remove vehicle images and attachments.
audience: user
status: stable
reviewedVersion: 0.1.19
lastReviewed: 2026-08-16
---

# Vehicle images and attachments

The **Uploads** tab keeps local vehicle images and general attachments with the vehicle record.
Admin, Editor, Viewer, and Planner can inspect stored media. Only Admin and Editor can upload,
change metadata, or delete it, and the server enforces that rule even if a control is visible.

## Open the Uploads tab

Open a vehicle from **Vehicle inventory**, choose **Edit**, then select **Uploads**. A vehicle must
be saved first: local images and attachments need its stored vehicle ID. Save the basic record,
then continue editing or reopen it to add media. The tab identifies empty image and attachment
areas and offers the appropriate upload action.

## Upload local images

Use **Upload image** to choose one or more local files. RailKeeper accepts JPG/JPEG, PNG, and
WebP. The default server limit is 10 MB per image; an operator can configure a stricter limit. The
browser checks the selected extension, and the server also checks the file size and detected MIME
type.

RailKeeper uploads selected images sequentially. When the vehicle initially has no image, every
request in a multi-file upload marks its image as main. The last successfully uploaded file is
therefore the main image. Earlier successful files stay stored if a later file fails; remaining
later files are not attempted in that run. Reload or reopen the vehicle before retrying. Compare
the refreshed image list and main image, then upload only files that are still missing. Image upload
is immediate, so it does not wait for the vehicle's **Save changes** action.

## Organize image metadata

Each image can be opened at original size, given an image description, moved up or down, or marked
as the main image. The main image is preferred in the vehicle inventory and compact views. You can
also associate the image with an existing maintenance entry.

Descriptions, order, main-image selection, and maintenance links are image metadata. They change
in the editor first and persist only when you use **Save changes** for the vehicle. If saving fails,
leave the editor open, read the error, and retry after checking the session and connection.

## Link an image to maintenance

Select a maintenance entry for an image that documents it. This chapter covers only the media link,
not editing the maintenance record itself. An image linked to maintenance cannot be removed
immediately. Select **No maintenance**, use **Save changes**, and then remove the image.

## Upload general attachments

Use **Upload attachment** to select one or more files, or drag files into **Drop files here** after
the vehicle has been saved. The category and note selected before uploading apply to every file in
that upload. With **Category automatically**, RailKeeper determines a category for each file.

Attachment batches are sequential. A successful earlier file remains stored if a later file fails,
and RailKeeper does not attempt later files in that run. Reload or reopen the vehicle, compare the
stored attachment list, and upload only files that are still missing. Stored attachments show their
original name, category, MIME type, and size.

## Attachment formats, limits, and categories

By default, RailKeeper accepts PDF, TXT, CSV, JSON, XML, ZIP, JPG/JPEG, PNG, and WebP attachments.
The default limit is 25 MB per file. An operator can configure a stricter size limit and restrict
which of these attachment extensions the server accepts. Empty, executable, disallowed,
content-blocked, and oversized files are rejected. Do not treat a selected file as stored until
RailKeeper reports a successful upload.

The stored category values are `Anleitung`, `Rechnung`, `Decoder-Datei`, `Dokumentation`,
`Ersatzteilliste`, `Zertifikat`, and `Sonstiges`.

Automatic categorization uses this priority, with the first matching rule winning:

1. `rechnung` or `invoice` in the file name: `Rechnung`.
2. `decoder` in the name, or JSON or XML: `Decoder-Datei`.
3. `ersatzteil` in the name: `Ersatzteilliste`.
4. `zertifikat` or `certificate` in the name: `Zertifikat`.
5. `anleitung`, `manual`, or `bedienung` in the name: `Anleitung`.
6. Another PDF: `Dokumentation`.
7. Any other allowed file: `Sonstiges`.

You can edit the category and note for each stored attachment. These changes persist only through
that row's save action, not through the vehicle's **Save changes** action. Saving an attachment row
reloads the vehicle and discards other unsaved media edits. Save pending image metadata first, then
edit and save one attachment row before changing another.

## Preview, open, and download attachments

**Preview** opens RailKeeper's integrated preview. PDFs and images, plus TXT, CSV, JSON, and XML,
are supported. ZIP files have no inline preview; download them and use a suitable local program.
**Open file** opens an appropriate inline representation in a new browser tab. **Download file**
downloads the stored original file. These distinct actions let you inspect in RailKeeper, open in
the browser, or keep a local copy.

## Delete images and attachments

Removing a stored image is immediate and has no extra confirmation. If it is the main image,
RailKeeper promotes the next image in sorted order. A linked image must first be unlinked and saved.
Removing an attachment opens a confirmation dialog. Confirm only after checking the original file
name. If deletion fails, reload the record and do not assume the file or metadata was removed.

## Roles, storage, and backup boundaries

Media is local and private RailKeeper data, not public website content. Application backups include
stored vehicle media. In v0.1.19, however, RailKeeper accepts backup files up to 250 MiB for
validation and restore. Before extensive cleanup, export a backup and ask an Admin to select that
file in the restore panel. Rely on it only after RailKeeper accepts it as compatible. This page
makes no promise about copies downloaded or created outside RailKeeper.

Admin, Editor, Viewer, and Planner may inspect existing media. Server-side writes require Admin or
Editor. **Open source** is the transition for imported image sources. See
[Article search, web documents, and spare parts](/guide/vehicles/search-and-spares) for remote
article images, **Found documents**, web-document import, and **Extract spare parts**. CV files and
editing maintenance records remain specialist boundaries that are not explained on this page.

## Empty, partial, and error states

| Situation | What to do |
| --- | --- |
| Vehicle is not saved | Save the basic record, then use **Uploads**. |
| No images or attachments | Use the empty-state upload action after checking the correct vehicle. |
| Format is not allowed | Choose one of the listed formats. The rejected file is not stored. |
| File is empty or too large | Correct or reduce the file, taking the server limit into account. |
| Batch upload partly fails | Earlier files remain stored. Reload the vehicle, compare the refreshed list, then upload only files that are still missing. |
| Image metadata does not save | Keep the editor open, read the error, and use **Save changes** again. |
| Linked image cannot be removed | Choose **No maintenance**, save the vehicle, then remove the image. |
| Preview is unavailable | Download the file or open it with an appropriate local program. |
| Opening or downloading fails | Check your session and connection, reload the record, then try again. |
| Attachment metadata or deletion fails | Read the error, reload the record, and do not assume the change succeeded. |

## Related pages

- [User Guide overview](/guide/)
- [First setup and sign-in](/guide/getting-started/)
- [Overview, metrics, and data quality](/guide/overview/)
- [Vehicle inventory and basic data](/guide/vehicles/)
- [Vehicle maintenance and condition](/guide/vehicles/maintenance)
- [Decoder, functions, and CV data](/guide/vehicles/decoder-cv)
- [Article search, web documents, and spare parts](/guide/vehicles/search-and-spares)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.19** and was last reviewed on 2026-08-16.
