package application_test

import (
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestRocoMX622EvidencePersistsAndTransfersWithoutExtendingThe10734Gap(t *testing.T) {
	db := testDB(t)
	for range 2 {
		if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
			t.Fatal(err)
		}
	}
	service := application.NewMasterDataService(db)
	doc, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	target := application.NewMasterDataService(testDB(t))
	if _, err := target.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	for _, model := range []string{
		"roco-motor-685404-6-poliger-decoder-mit-stecker-nem-651",
		"roco-motor-685504-6-poliger-decoder-mit-stecker-nem-651",
		"roco-motor-686101-6-poliger-decoder-gerade-stifte-nem-651",
		"roco-motor-686201-8-poliger-decoder-nem-652",
	} {
		refs, err := service.DecoderReferences(t.Context(), model, true)
		if err != nil || len(refs) != 1 || refs[0].ID != "cv-ref-roco-mx622-oem-2022" || len(refs[0].Rows) != 40 {
			t.Fatalf("scoped OEM reference missing for %s: %v", model, err)
		}
		after, err := target.DecoderReferences(t.Context(), model, true)
		if err != nil || !reflect.DeepEqual(refs, after) {
			t.Fatalf("OEM evidence or source pages changed during transfer: %s, %v", model, err)
		}
		for _, row := range after[0].Rows {
			if row.SourceIndex != 1 || row.Page != 69 || row.FactoryValue != "" {
				t.Fatal("source index or absent OEM defaults changed during persistence")
			}
		}
	}
	refs, err := target.DecoderReferences(t.Context(), "roco-motor-10734-h0e-n-dcc-decoder-mit-geraden-kontaktstiften", true)
	if err != nil || len(refs) != 1 || refs[0].ID != "cv-gap-roco-motor" || len(refs[0].Rows) != 0 {
		t.Fatal("10734 inherited the MX622 evidence")
	}
}
