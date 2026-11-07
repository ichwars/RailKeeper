package transfer

import (
	"encoding/json"
	"testing"
)

func TestExistingPreviewFingerprintsRemainByteCompatible(t *testing.T) {
	vehicle := TransferVehicle{ID: "synthetic", InventoryNumber: "RK-001", Name: "Prüfung ä", Gauge: "H0", Digital: true, PurchasePrice: "123.45"}
	accessory := TransferAccessory{ID: "accessory", InventoryNumber: "A-001", Name: "Test"}
	cases := []struct {
		name     string
		value    any
		version  int
		expected string
	}{
		{"current", vehicle, 2, "8e067a706be6a55a0ebb5a13116af62b205476104443386e22ca2927b446739e"},
		{"legacy", vehicle, 1, "ba1d2a3c49b45964d895c2f67d14baa42d2b8330b9540f7298307f260cbd9428"},
		{"accessory", accessory, 2, "93974837216076d30b3740e42fd6ab429a51d68cc26745480e34c3c9bb06a651"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if actual := DataTransferTargetFingerprintForVersion(test.value, test.version); actual != test.expected {
				t.Fatalf("existing preview fingerprint changed: %s", actual)
			}
		})
	}
	if DataTransferTargetFingerprint(vehicle) != DataTransferTargetFingerprint(&vehicle) ||
		DataTransferTargetFingerprint(accessory) != DataTransferTargetFingerprint(&accessory) {
		t.Fatal("pointer/value preview identity differs")
	}
	if DataTransferTargetFingerprint(make(chan int)) != "" {
		t.Fatal("unsupported data received a fingerprint")
	}
}

func TestSelectedEmptyAreasRemainDistinctFromOmittedAreas(t *testing.T) {
	cases := []struct {
		areas DataTransferPackageAreas
		want  string
	}{
		{DataTransferPackageAreas{}, `{}`},
		{DataTransferPackageAreas{Vehicles: []TransferVehicle{}}, `{"vehicles":[]}`},
		{DataTransferPackageAreas{VehicleSets: []TransferVehicleSet{}}, `{"vehicleSets":[]}`},
		{DataTransferPackageAreas{Accessories: []TransferAccessory{}, ExhibitionLists: []TransferExhibitionList{}}, `{"accessories":[],"exhibitionLists":[]}`},
	}
	for _, test := range cases {
		data, err := json.Marshal(test.areas)
		if err != nil || string(data) != test.want {
			t.Fatalf("selected area lost: %s, %v", data, err)
		}
	}
}

func TestProfileNameIdentityRetainsUnicodeCaseFolding(t *testing.T) {
	for _, pair := range [][2]string{{" Straße ", "straße"}, {"K", "K"}, {"σ", "ς"}} {
		if DataTransferProfileNameKey(pair[0]) != DataTransferProfileNameKey(pair[1]) {
			t.Fatalf("profile identity changed: %q", pair)
		}
	}
}
