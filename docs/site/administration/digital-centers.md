---
title: Connect a command station
description: Connect a command station.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Connect a command station

A command station is optional. You can manage vehicles, pictures, and CV files without connected hardware. Setup and station operations require **Admin**.


<GuideFigure name="digital-settings" alt="Configure the station your RailKeeper server can actually reach." caption="Configure the station your RailKeeper server can actually reach. Synthetic demo data, not factory settings. Click for full size." />

## Configure a connection

1. Open **Settings > Digital command stations**.
2. Choose the actual station type.
3. Enter the station's IP address and its required port. Read the address from the station itself.
4. Save the settings and run the connection test.
5. Open **Digital command stations** in the sidebar for the supported workflows.

A successful ping does not establish that RailKeeper can reach the required services. Connectivity from the RailKeeper server matters, not only connectivity from your browser.

## What does each connection support?

| Station | Current scope |
| --- | --- |
| ESU ECoS | compare and adopt locomotive data, read stored CVs, and use confirmed write workflows; separate image transfer |
| Märklin CS3 | read and compare locomotive data, without a general write-back workflow |
| Z21 | connection testing and diagnostics, not a complete locomotive-list import |
| Intellibox 3 | Z21-compatible connection testing and diagnostics, not a complete locomotive-list import |

## ECoS: do not confuse the two connections

- **Port 15471:** TCP interface used by RailKeeper. It is not a browser web page.
- **Port 80:** web interface and image transfer. Normally open `http://YOUR-ECOS-IP/` in a browser, without `:15471`.

Separate networks need suitable routing, netmask, gateway, and any required firewall rules. Do not change station settings speculatively when only one application cannot connect.

## Next steps

- [Adopt ECoS locomotives and stored CVs](../guide/import-export/ecos-sync)
- [Prepare and confirm locomotive images](../guide/vehicles/locomotive-images)
- [Import CS3 locomotives](../guide/import-export/cs3-import)
