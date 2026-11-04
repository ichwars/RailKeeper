---
title: Try RailKeeper
description: Start RailKeeper with Windows or Docker, import sample data and add a vehicle of your own with a picture.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-11-04
---

# Try RailKeeper

**Goal:** Import two sample vehicles, add a third and attach a picture.
You need a Windows PC or Docker, a browser and an empty test collection.
This exercise does not need a command station or your real collection data.

## 1. Choose how to start

### Windows

1. Open the [latest release](https://github.com/ichwars/RailKeeper/releases/latest).
   Under **Assets**, download the Windows x64 ZIP and extract it completely.
2. Open PowerShell in the extracted folder. Start with a separate demo data directory:

```powershell
$env:RAILKEEPER_DATA_DIR = Join-Path $env:LOCALAPPDATA 'RailKeeper-Demo\data'
.\start-railkeeper.bat
```

3. Keep the program window open. The browser opens automatically; alternatively use the local
   address shown in the window.

RailKeeper needs no additional runtime on Windows. Your sample data stays in
`%LOCALAPPDATA%\RailKeeper-Demo\data`. A new program folder alone does not create a new collection.
[Windows operation and updates](administration/deployment-configuration).

### Docker

With Docker running, enter this command in your terminal:

```bash
docker run -d --name railkeeper-demo --publish 127.0.0.1:18080:8080 --mount source=railkeeper_demo,target=/data ghcr.io/ichwars/railkeeper:latest
```

Open `http://localhost:18080` on the same computer.
The container and volume deliberately use different names from a regular installation.
This startup is only accessible through the local address. For a remote home server,
follow the [installation guide](administration/deployment-configuration).

After trying it, `docker stop railkeeper-demo` stops the program and keeps your data.
Use `docker start railkeeper-demo` to resume the same collection.
Images support AMD64 and ARM64; a Raspberry Pi needs a 64-bit operating system.

## 2. Create your local account

On first opening, complete the setup with your own administrator account and sign in.
There is no default password. This local trial does not need a mail server.
Check the data directory under **Settings > General** before continuing.

## 3. Use the sample data {#sample-data}

Download these two files. All vehicles and picture contents are synthetic examples.

| File | Use |
| --- | --- |
| <GuideDownload file="vehicle-import-demo.json">Two sample vehicles as JSON</GuideDownload> | Read through vehicle import, not backup restore. |
| <GuideDownload file="locomotive-demo.png">Sample locomotive picture</GuideDownload> | Upload it to the third vehicle you create. |

1. Open **Import/Export > Create profile**. Name: **First try**, direction **Import**,
   areas **Vehicles**, format **JSON**. Save with **Create profile**.
2. Choose **New import**, this profile and the downloaded JSON file.
3. Review the preview: `DEMO-JSON-001` and `DEMO-JSON-002`. Confirm the import.
4. Open **Vehicle inventory** and search for `DEMO-JSON`. Both vehicles should appear.

The import creates records. Use the empty demo collection from step 1.
Importing the same inventory numbers again produces conflicts; you can cancel the import.
[Detailed import exercise with CSV and conflicts](guide/import-export/examples).

## 4. Add your own vehicle and picture

1. In **Vehicle inventory**, choose **New vehicle**.
2. Choose **Individual vehicle** and enter **My first demo locomotive** as its name.
   Choose a manufacturer, gauge and locomotive category.
3. Choose **Continue to article data**, then **Continue to vehicle details** without searching.
   Select a subtype there and confirm with **Create**.
4. The vehicle editor opens. Choose **Uploads > Upload image**, then `locomotive-demo.png`
   and save your changes.
5. Close the editor with **X**. Search for **My first demo locomotive** and open it again.
   The picture should still be there.

<GuideFigure name="vehicles" alt="Vehicle inventory with a search field and New vehicle button." caption="Find the vehicles you imported or created here. The picture shows a different synthetic demo collection." />

## 5. Check the result and share your experience {#feedback}

- [ ] Setup and sign-in work.
- [ ] Both imported vehicles can be found.
- [ ] Your own demo locomotive and its picture are present after reopening it.

Share your experience in [Discussions](https://github.com/ichwars/RailKeeper/discussions/categories/general).
These details help even if something did not work:

```text
System and RailKeeper version:
How I managed my collection before:
What worked:
Where I got stuck:
Would I keep using RailKeeper? Why?
```

Find the version under **Settings > General > Updates**. For an error, describe the last
step and the error message. Use your demo data in the report and leave out passwords,
credentials and personal collection data.

Does RailKeeper help you? Support the project with a
[star on GitHub](https://github.com/ichwars/RailKeeper).
Continue with the [handbook](guide/) or [your first vehicle](guide/quick-start).
