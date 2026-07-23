package application

import (
	"os"
	"strings"
	"testing"
)

func TestVehicleServiceCoreStaysFocused(t *testing.T) {
	data, err := os.ReadFile("vehicles.go")
	if err != nil {
		t.Fatalf("read vehicles.go: %v", err)
	}
	source := string(data)
	if lines := strings.Count(source, "\n") + 1; lines > 500 {
		t.Fatalf("vehicles.go has %d lines, want at most 500", lines)
	}
	for _, method := range []string{
		"func (s *VehicleService) Create(",
		"func (s *VehicleService) CreateAttachment(",
		"func (s *VehicleService) CreateMaintenance(",
		"func (s *VehicleService) CreateCVValue(",
	} {
		if strings.Contains(source, method) {
			t.Errorf("vehicles.go still contains method %q", method)
		}
	}
}
