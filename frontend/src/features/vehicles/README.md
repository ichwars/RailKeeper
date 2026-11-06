# Vehicles Feature

Vehicle inventory is the central RailKeeper domain. This feature owns the inventory list, vehicle dialogs, article search review, QR-code workflow, images and attachments.

The inventory view composes the existing feature controllers. Collection loading belongs to
`useVehicleCollectionController`; form/selection state and aggregate refresh belong to
`useVehicleEditorController`. `useVehicleImageDraftController` owns local image drafts, while
`useVehicleMediaController` retains image and attachment drafts. `useVehicleUploads` and
`useVehicleMediaDeletion` own independent confirmed actions, bounded progress and revision guards.
Selecting local files alone does not upload them. Image persistence
and imported mapping/CV adoption run after the core save through `vehicleImagePersistence` and
`vehicleImportAdoption`. Set creation, duplication and dialog state belong to `useVehicleSetController`.

Accessories have their own product area and remain separate from vehicle editing. See the
[architecture](../../../../docs/architecture.md) for storage, transaction and integration boundaries.
