package application

import (
	"fmt"
	"reflect"
	"testing"
)

func TestECoSStoredCVReadsWholeRangeIncludingHolesAndZero(t *testing.T) {
	want := []ECoSCVValue{{Number: 1, Value: 3}, {Number: 2, Value: 0},
		{Number: 257, Value: 128}, {Number: 512, Value: 255}, {Number: 1024, Value: 7}}
	seen := make(chan int, ECoSStoredCVLimit)
	listener := startECoSTestServer(t, func(command string) []string {
		var objectID, number int
		if _, err := fmt.Sscanf(command, "get(%d, cv[%d])", &objectID, &number); err != nil || objectID != 1002 {
			t.Errorf("unexpected command (writes forbidden): %q", command)
		}
		seen <- number
		for _, value := range want {
			if value.Number == number {
				return []string{"<REPLY " + command + ">", fmt.Sprintf("1002 cv[%d,%d]", number, value.Value), "<END 0 (OK)>"}
			}
		}
		return []string{"<REPLY " + command + ">", "<END 22 (not possible at 11)>"}
	})
	defer func() { _ = listener.Close() }()
	host, port := splitTestAddress(t, listener.Addr().String())
	s := NewECoSService()
	got := []ECoSCVValue{}
	unavailable := 0
	for start := 1; start <= ECoSStoredCVLimit; start += ECoSStoredCVBatchSize {
		batch, err := s.ReadStoredCVs(t.Context(), ECoSConnectionInput{Host: host, Port: port}, 1002, start)
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, batch.Values...)
		unavailable += batch.Unavailable
	}
	if !reflect.DeepEqual(got, want) || unavailable != 1019 || len(seen) != 1024 {
		t.Fatalf("values=%v unavailable=%d queries=%d", got, unavailable, len(seen))
	}
	for number := 1; number <= 1024; number++ {
		if actual := <-seen; actual != number {
			t.Fatalf("query %d: got CV%d", number, actual)
		}
	}
}

func TestECoSStoredCVRejectsInvalidReplies(t *testing.T) {
	for name, data := range map[string]string{
		"wrong object": "1003 cv[1,3]", "wrong CV": "1002 cv[8,151]",
		"non number": "1002 cv[1,unknown]", "byte overflow": "1002 cv[1,256]",
		"duplicate": "1002 cv[1,3]\n1002 cv[1,3]",
	} {
		t.Run(name, func(t *testing.T) {
			listener := startECoSTestServer(t, func(command string) []string {
				return []string{"<REPLY " + command + ">", data, "<END 0 (OK)>"}
			})
			defer func() { _ = listener.Close() }()
			host, port := splitTestAddress(t, listener.Addr().String())
			_, err := NewECoSService().ReadStoredCVs(t.Context(), ECoSConnectionInput{Host: host, Port: port}, 1002, 1)
			if err == nil {
				t.Fatal("invalid reply accepted")
			}
		})
	}
}

func TestECoSStoredCVNeverImportsNegativeReplyPayload(t *testing.T) {
	listener := startECoSTestServer(t, func(command string) []string {
		return []string{"<REPLY " + command + ">", "1002 cv[1,99]", "<END 22 (not possible)>"}
	})
	defer func() { _ = listener.Close() }()
	host, port := splitTestAddress(t, listener.Addr().String())
	batch, err := NewECoSService().ReadStoredCVs(t.Context(), ECoSConnectionInput{Host: host, Port: port}, 1002, 1)
	if err != nil || len(batch.Values) != 0 || batch.Unavailable != 32 {
		t.Fatalf("negative data imported: %+v %v", batch, err)
	}
}

func TestECoSCVParserDoesNotInventZeroForMalformedData(t *testing.T) {
	values := parseECoSCVValues(map[string][]string{"cv": {"1,unknown", "2,0", "3,256", "1025,1", "7:broken"}})
	if !reflect.DeepEqual(values, []ECoSCVValue{{Number: 2, Value: 0}}) {
		t.Fatalf("malformed data produced values: %#v", values)
	}
}
