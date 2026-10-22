package application

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestECoSImageDescriptorRequiresMatchingConfirmedObject(t *testing.T) {
	for _, lines := range [][]string{{"1001 locodesc[LOCO_TYPE_DIESEL,IMAGE_TYPE_INT,118]", "<END 0 (OK)>"}, {"1002 locodesc[LOCO_TYPE_MISC,IMAGE_TYPE_USER,0]", "<END 0 (OK)>"}} {
		id := 1001
		if strings.HasPrefix(lines[0], "1002") {
			id = 1002
		}
		if _, err := parseECoSImageDescriptor(lines, id); err != nil {
			t.Fatal(err)
		}
		if _, err := parseECoSImageDescriptor(lines, 1003); err == nil {
			t.Fatal("accepted another object")
		}
		if _, err := parseECoSImageDescriptor(lines[:1], id); err == nil {
			t.Fatal("accepted incomplete reply")
		}
	}
}

func TestECoSImageAssignmentUsesControlWithoutForceAndOnlyLocodesc(t *testing.T) {
	for _, refuse := range []bool{false, true} {
		t.Run(fmt.Sprint(refuse), func(t *testing.T) {
			var mu sync.Mutex
			commands := []string{}
			listener := startECoSTestServer(t, func(command string) []string {
				mu.Lock()
				commands = append(commands, command)
				mu.Unlock()
				end := "<END 0 (OK)>"
				if refuse && strings.HasPrefix(command, "request(") {
					end = "<END 25 (couldn't get control)>"
				}
				return []string{"<REPLY " + command + ">", end}
			})
			defer func() { _ = listener.Close() }()
			host, port := splitTestAddress(t, listener.Addr().String())
			err := NewECoSService().AssignUserImage(t.Context(), ECoSConnectionInput{Host: host, Port: port}, 1001, 7)
			mu.Lock()
			defer mu.Unlock()
			if refuse {
				if !errors.Is(err, ErrECoSImageControlBusy) || len(commands) != 1 {
					t.Fatalf("refused: %v %v", err, commands)
				}
			} else if err != nil || strings.Join(commands, ";") != "request(1001, control);set(1001, locodesc[IMAGE_TYPE_USER,7]);release(1001, control)" {
				t.Fatalf("err=%v commands=%v", err, commands)
			}
		})
	}
}
