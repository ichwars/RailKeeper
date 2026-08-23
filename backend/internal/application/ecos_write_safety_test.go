package application

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestECoSListLocomotivesUsesOnlyApprovedMasterFields(t *testing.T) {
	listener := startECoSTestServer(t, func(command string) []string {
		if command != eCoSLocomotiveListCommand {
			t.Fatalf("command = %q", command)
		}
		return []string{
			"<REPLY queryObjects(10, addr, name, protocol)>",
			`1001 addr[3] name["Testlok A"] protocol[DCC]`,
			`1002 addr[4] name["Testlok B"] protocol[MM27]`,
			"<END 0 (OK)>",
		}
	})
	defer func() { _ = listener.Close() }()
	host, port := splitTestAddress(t, listener.Addr().String())

	locomotives, err := NewECoSService().ListLocomotives(
		t.Context(), ECoSConnectionInput{Host: host, Port: port},
	)
	if err != nil || len(locomotives) != 2 || locomotives[1].Address != 4 {
		t.Fatalf("locomotives=%#v err=%v", locomotives, err)
	}
}

func TestECoSReadLocomotiveUsesTargetedMasterGet(t *testing.T) {
	listener := startECoSTestServer(t, func(command string) []string {
		switch command {
		case "request(1001, view)":
			return []string{"<REPLY request(1001, view)>", "<END 0 (OK)>"}
		case eCoSLocomotiveDetailCommand(1001):
			return []string{fmt.Sprintf("<REPLY %s>", command),
				`1001 name["Testlok A"] addr[3] protocol[DCC]`, "<END 0 (OK)>"}
		case "release(1001, view)":
			return []string{"<REPLY release(1001, view)>", "<END 0 (OK)>"}
		default:
			t.Fatalf("command = %q", command)
			return nil
		}
	})
	defer func() { _ = listener.Close() }()
	host, port := splitTestAddress(t, listener.Addr().String())

	locomotive, err := NewECoSService().ReadLocomotive(
		t.Context(), ECoSConnectionInput{Host: host, Port: port}, 1001,
	)
	if err != nil || locomotive.ObjectID != 1001 || locomotive.Name != "Testlok A" {
		t.Fatalf("locomotive=%#v err=%v", locomotive, err)
	}
}

func TestECoSSyncMarksMissingWriteReplyAsUnknown(t *testing.T) {
	listener := startECoSTestServer(t, func(command string) []string {
		switch {
		case command == "request(1001, view)":
			return []string{"<REPLY request(1001, view)>", "<END 0 (OK)>"}
		case command == eCoSLocomotiveDetailCommand(1001):
			return []string{fmt.Sprintf("<REPLY %s>", command),
				`1001 name["Old"] addr[3] protocol[DCC]`, "<END 0 (OK)>"}
		case strings.HasPrefix(command, "set(1001,"):
			return nil
		case command == "release(1001, view)":
			return []string{"<REPLY release(1001, view)>", "<END 0 (OK)>"}
		default:
			t.Fatalf("command = %q", command)
			return nil
		}
	})
	defer func() { _ = listener.Close() }()
	host, port := splitTestAddress(t, listener.Addr().String())

	_, err := NewECoSService().SyncLocomotive(t.Context(), ECoSLocomotiveSyncInput{
		Host: host, Port: port, ObjectID: 1001,
		Desired: ECoSLocomotiveSyncDesired{Name: "New"}, Confirm: true,
	})
	if !errors.Is(err, ErrECoSWriteStateUnknown) {
		t.Fatalf("error=%v", err)
	}
}

func TestECoSSyncKeepsExplicitRejectionDefinite(t *testing.T) {
	listener := startECoSTestServer(t, func(command string) []string {
		switch {
		case command == "request(1001, view)":
			return []string{"<REPLY request(1001, view)>", "<END 0 (OK)>"}
		case command == eCoSLocomotiveDetailCommand(1001):
			return []string{fmt.Sprintf("<REPLY %s>", command),
				`1001 name["Old"] addr[3] protocol[DCC]`, "<END 0 (OK)>"}
		case strings.HasPrefix(command, "set(1001,"):
			return []string{fmt.Sprintf("<REPLY %s>", command), "<END 11 (unsupported)>"}
		case command == "release(1001, view)":
			return []string{"<REPLY release(1001, view)>", "<END 0 (OK)>"}
		default:
			t.Fatalf("command = %q", command)
			return nil
		}
	})
	defer func() { _ = listener.Close() }()
	host, port := splitTestAddress(t, listener.Addr().String())

	_, err := NewECoSService().SyncLocomotive(t.Context(), ECoSLocomotiveSyncInput{
		Host: host, Port: port, ObjectID: 1001,
		Desired: ECoSLocomotiveSyncDesired{Name: "New"}, Confirm: true,
	})
	if err == nil || errors.Is(err, ErrECoSWriteStateUnknown) {
		t.Fatalf("error=%v", err)
	}
}
