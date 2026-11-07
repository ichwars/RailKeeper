package decoder

import "testing"

func TestCVNumberAndByteBoundaries(t *testing.T) {
	for _, test := range []struct {
		number int
		valid  bool
	}{{-1, false}, {0, false}, {1, true}, {1024, true}, {1025, false}} {
		if ValidCVNumber(test.number) != test.valid {
			t.Fatalf("CV number boundary %d", test.number)
		}
	}
	for _, test := range []struct {
		value int
		valid bool
	}{{-1, false}, {0, true}, {255, true}, {256, false}} {
		if ValidCVValue(test.value) != test.valid {
			t.Fatalf("CV byte boundary %d", test.value)
		}
	}
}
