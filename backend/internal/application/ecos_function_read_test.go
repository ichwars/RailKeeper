package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestECoSFunctionReadChecksAllKeysAndNeverControlsLocomotive(t *testing.T) {
	commands := []string{}
	var mutex sync.Mutex
	listener := startECoSTestServer(t, func(command string) []string {
		mutex.Lock()
		commands = append(commands, command)
		mutex.Unlock()
		lines := []string{"<REPLY " + command + ">"}
		if strings.Contains(command, "funcexists[") {
			index, err := strconv.Atoi(strings.TrimSuffix(strings.Split(command, "[")[1], "])"))
			if err != nil {
				t.Error(err)
			}
			description := "-2"
			switch index {
			case 0:
				description = "35"
			case 31:
				description = "6,moment"
			}
			lines = append(lines, fmt.Sprintf("1002 funcexists[%d,%s]", index, description))
			lines = append([]string{"<EVENT 1002>", "1002 func[0,1]", "<END 0 (OK)>"}, lines...)
		}
		return append(lines, "<END 0 (OK)>")
	})
	defer func() { _ = listener.Close() }()
	host, port := splitTestAddress(t, listener.Addr().String())
	values, err := NewECoSService().ReadFunctions(t.Context(), ECoSConnectionInput{Host: host, Port: port}, 1002)
	if err != nil || len(values) != 2 || values[0].Description != 35 || values[1].Index != 31 || !values[1].Momentary {
		t.Fatal(values, err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if len(commands) != 34 || commands[0] != "request(1002, view)" || commands[33] != "release(1002, view)" {
		t.Fatal(commands)
	}
	for _, command := range commands {
		if strings.Contains(command, "control") || strings.HasPrefix(command, "set(") || strings.Contains(command, "cv[") {
			t.Fatal("active command", command)
		}
	}
}

func TestECoSFunctionReadRejectsUnconfirmedOrMalformedWholeSnapshot(t *testing.T) {
	for _, invalid := range []string{
		"1003 funcexists[0,3]", "1002 funcexists[1,3]", "1002 funcexists[0,abc]",
		"1002 funcexists[0,2147483648]", "1002 funcexists[0,-3]", "1002 funcexists[0,3,true]",
		"1002 funcexists[0,3]\n1002 funcexists[0,3]", "<END 22 (error)>",
	} {
		t.Run(invalid, func(t *testing.T) {
			listener := startECoSTestServer(t, func(command string) []string {
				if strings.Contains(command, "funcexists[") {
					if strings.HasPrefix(invalid, "<END") {
						return []string{"<REPLY " + command + ">", invalid}
					}
					return []string{"<REPLY " + command + ">", invalid, "<END 0 (OK)>"}
				}
				return []string{"<REPLY " + command + ">", "<END 0 (OK)>"}
			})
			defer func() { _ = listener.Close() }()
			host, port := splitTestAddress(t, listener.Addr().String())
			values, err := NewECoSService().ReadFunctions(t.Context(), ECoSConnectionInput{Host: host, Port: port}, 1002)
			if !errors.Is(err, ErrDigitalCenterDeviceOutput) || values != nil {
				t.Fatal("accepted invalid or partial snapshot", values, err)
			}
		})
	}
}

func TestECoSFunctionReadPreservesOpaqueDescriptionsOutsideTheSymbolCatalog(t *testing.T) {
	for _, code := range []int{4096, 7431, 65535, 2147483647} {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			listener := startECoSTestServer(t, func(command string) []string {
				lines := []string{"<REPLY " + command + ">"}
				if strings.Contains(command, "funcexists[") {
					index, err := strconv.Atoi(strings.TrimSuffix(strings.Split(command, "[")[1], "])"))
					if err != nil {
						t.Fatal(err)
					}
					description := -2
					if index == 3 {
						description = code
					}
					lines = append(lines, fmt.Sprintf("1002 funcexists[%d,%d]", index, description))
				}
				return append(lines, "<END 0 (OK)>")
			})
			defer func() { _ = listener.Close() }()
			host, port := splitTestAddress(t, listener.Addr().String())
			values, err := NewECoSService().ReadFunctions(t.Context(), ECoSConnectionInput{Host: host, Port: port}, 1002)
			if err != nil || len(values) != 1 || values[0].Index != 3 || values[0].Description != code {
				t.Fatal("unknown description was rejected or altered", values, err)
			}
		})
	}
}

type functionDeviceStub struct {
	*imageDeviceStub
	reads   int
	changed bool
}

func (d *functionDeviceStub) ReadFunctions(context.Context, ECoSConnectionInput, int) ([]ECoSFunctionValue, error) {
	d.reads++
	if d.changed {
		d.loco.Address++
	}
	return []ECoSFunctionValue{{Index: 0, Description: 3}}, nil
}

func TestECoSFunctionReadRequiresMappingAndStableIdentity(t *testing.T) {
	for _, mode := range []string{"ok", "unmapped", "ambiguous", "stale", "protocol", "changed"} {
		t.Run(mode, func(t *testing.T) {
			s, device, _, _, _ := imageTransferFixture(t)
			d := &functionDeviceStub{imageDeviceStub: device, changed: mode == "changed"}
			s.ecos = d
			mapping := VehicleExternalMap{Provider: "ecos", ExternalID: "1001", ExternalAddress: "003", ExternalProtocol: "DCC"}
			mappings := []VehicleExternalMap{mapping}
			if mode == "unmapped" {
				mappings = nil
			}
			if mode == "ambiguous" {
				mappings = append(mappings, mapping)
			}
			if mode == "stale" {
				mappings[0].ExternalAddress = "4"
			}
			if mode == "protocol" {
				mappings[0].ExternalProtocol = "MOTOROLA"
			}
			s.vehicles = &workspaceVehicleReaderStub{vehicles: []Vehicle{{ID: "vehicle-1", ExternalMappings: mappings}}}
			result, err := s.ReadVehicleECoSFunctions(t.Context(), "vehicle-1")
			if mode == "ok" {
				if err != nil || result.ObjectID != 1001 || len(result.Functions) != 1 || result.ReadAt == "" {
					t.Fatal(result, err)
				}
			} else if err == nil || result.Functions != nil {
				t.Fatal("accepted unverified mapping", result, err)
			}
			if (mode == "unmapped" || mode == "ambiguous" || mode == "stale" || mode == "protocol") && d.reads != 0 {
				t.Fatal("read wrong target")
			}
		})
	}
}
