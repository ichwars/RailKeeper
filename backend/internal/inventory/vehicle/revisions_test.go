package vehicle

import "testing"

// Tokens were computed independently from the pre-extraction record definitions.
func TestExistingDraftRevisionEncodingSurvivesTheBoundaryChange(t *testing.T) {
	value := Vehicle{ID: "vehicle", InventoryNumber: "RK-001", Name: "Recorded", PurchasePrice: "12,50",
		CreatedAt: "original", UpdatedAt: "original", VehicleSet: &VehicleSetSummary{ID: "set", Name: "Set",
			InventoryNumber: "SET-1", MemberCount: 2, MainImage: &VehicleSetMainImage{URL: "independent image"}}}
	if got := CoreRevision(value); got != "bb1f83b4b34c296e4dbabfe478490c331b5408c3d796f0c0ce0372f4f44f4ef5" {
		t.Fatal("existing vehicle draft token changed", got)
	}
	image := VehicleImage{ID: "image", VehicleID: "vehicle", URL: "/file", StoragePath: "private/original",
		BlobID: "blob", CreatedAt: "original"}
	if got := ImageRevision(image); got != "bfbc5fc345fc6a391ef42e039549936de2276602b36dedb401bc2f8065e88f65" {
		t.Fatal("existing image draft token changed", got)
	}
	attachment := VehicleAttachment{ID: "attachment", VehicleID: "vehicle", FileName: "manual.pdf",
		StoragePath: "private/manual", BlobID: "manual-blob", CreatedAt: "original"}
	if got := AttachmentRevision(attachment); got != "5d7f725a99942ed5ed20599654e890bce30cbbb20f3da9e616f6dcc1c768b0c4" {
		t.Fatal("existing attachment draft token changed", got)
	}
	previous := AttachmentRevision(attachment)
	attachment.Revision = "response-only revision"
	if AttachmentRevision(attachment) != previous {
		t.Fatal("response revision changes its own token")
	}
	attachment.StoragePath = "private/replacement"
	if AttachmentRevision(attachment) == previous {
		t.Fatal("private file replacement is invisible to conflict detection")
	}
}
