# Roadmap

## Completed Foundation

- Go API and static frontend runtime
- SQLite migrations and seed loading
- first-run setup and authentication
- CSRF and role checks
- vehicle CRUD
- master data management
- inventory number schemes and history
- article data search review workflow
- master data import/export
- QR code generation
- article-search image suggestions and primary image selection
- local image uploads with drag and drop
- automatic thumbnails for JPEG/PNG/WebP images
- image deletion reference checks
- vehicle attachments
- backup and restore
- backup compatibility preflight
- overview dashboard
- vehicle CSV/TSV/XLSX/XLS/ODS/JSON import and export
- manual column mapping for unknown vehicle import headers
- field-level update preview for vehicle imports
- safe duplicate update mode for vehicle imports
- maintenance and condition history
- images and attachments per maintenance entry
- maintenance radar and dashboard summaries
- digital function mapping
- digital function icon picker
- digital function JSON import/export
- structured CV values and CV files
- CV import comparison preview
- decoder profile suggestions for CV values and files
- visible CV change history
- ESU/LokProgrammer project files as decoder attachments with safe metadata preview and extraction
- responsive inventory table/card switch
- compact mobile inventory layout

## Next Practical Milestones

1. ESU LokProgrammer import
   - add preview image extraction when safely available
   - evaluate LokProgrammer CSV/XML/text exports for CV values and function mappings
   - only reverse-engineer proprietary ESUX blocks if no supported export path exists
2. Inventory row quick menu
   - add a compact context menu for vehicle rows/cards
   - include RailKeeper actions such as details, edit, QR code, PDF entry, images, attachments, maintenance and delete
   - keep destructive actions separated and confirmed
3. Configurable overview dashboard
   - evaluate draggable/hideable widgets for RailKeeper statistics
   - include reset/export controls only where they provide real value
   - keep the dashboard informative without turning it into a configuration surface
4. Settings and system integration
   - wire default printer selection to real system printers where the host platform allows it
   - turn update checks from UI scaffolding into a real version check
   - decide which authentication options should become functional instead of informational
   - keep storage usage, backup and restore visible without making settings feel overloaded
5. Ongoing Bambuddy-inspired design polish
   - continue refining dense toolbar/table layouts without boxed hover effects
   - keep icon buttons transparent by default with color-only hover feedback
   - review mobile navigation after the collapsible desktop sidebar work
   - adapt dashboard widgets only when they add RailKeeper-specific value

## Explicitly Deferred

- accessories
- spare parts tab with targeted web search, images, prices, source, article numbers and update checks
- public sharing by default
- cloud sync
- multi-tenant hosting
