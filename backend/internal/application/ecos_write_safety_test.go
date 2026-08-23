package application

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	ecospkg "railkeeper/backend/internal/ecos"
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

func TestECoSReadLocomotiveAcceptsSuccessfulReplyWithCompatibilityWarning(t *testing.T) {
	listener := startECoSTestServer(t, func(command string) []string {
		switch command {
		case "request(1001, view)":
			return []string{"<REPLY request(1001, view)>", "<END 0 (OK)>"}
		case eCoSLocomotiveDetailCommand(1001):
			return []string{fmt.Sprintf("<REPLY %s>", command),
				`1001 name["Testlok A"] addr[3] protocol[DCC]`,
				"<END 0 (OK, but obsolete attribute at 11)>"}
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

func TestECoSListLocomotivesRejectsIncompleteAndNegativeReplies(t *testing.T) {
	tests := map[string][]string{
		"incomplete": {
			"<REPLY queryObjects(10, addr, name, protocol)>",
			`1001 addr[3] name["Testlok A"] protocol[DCC]`,
		},
		"rejected": {
			"<REPLY queryObjects(10, addr, name, protocol)>",
			`1001 addr[3] name["Testlok A"] protocol[DCC]`,
			"<END 11 (unsupported)>",
		},
	}
	for name, reply := range tests {
		t.Run(name, func(t *testing.T) {
			listener := startECoSTestServer(t, func(string) []string { return reply })
			defer func() { _ = listener.Close() }()
			host, port := splitTestAddress(t, listener.Addr().String())
			service := shortTimeoutECoSService()

			locomotives, err := service.ListLocomotives(
				t.Context(), ECoSConnectionInput{Host: host, Port: port},
			)
			if err == nil || len(locomotives) != 0 {
				t.Fatalf("locomotives=%#v err=%v", locomotives, err)
			}
		})
	}
}

func TestECoSReadLocomotiveRejectsIncompleteAndNegativeRepliesWithMatchingData(t *testing.T) {
	tests := map[string][]string{
		"incomplete": {
			"<REPLY get(1001, profile, protocol, name, addr, funcdesc)>",
			`1001 name["Testlok A"] addr[3] protocol[DCC]`,
		},
		"rejected": {
			"<REPLY get(1001, profile, protocol, name, addr, funcdesc)>",
			`1001 name["Testlok A"] addr[3] protocol[DCC]`,
			"<END 11 (unsupported)>",
		},
	}
	for name, detailReply := range tests {
		t.Run(name, func(t *testing.T) {
			listener := startECoSTestServer(t, func(command string) []string {
				switch command {
				case "request(1001, view)":
					return []string{"<REPLY request(1001, view)>", "<END 0 (OK)>"}
				case eCoSLocomotiveDetailCommand(1001):
					return detailReply
				case "release(1001, view)":
					return []string{"<REPLY release(1001, view)>", "<END 0 (OK)>"}
				default:
					return nil
				}
			})
			defer func() { _ = listener.Close() }()
			host, port := splitTestAddress(t, listener.Addr().String())
			service := shortTimeoutECoSService()

			locomotive, err := service.ReadLocomotive(
				t.Context(), ECoSConnectionInput{Host: host, Port: port}, 1001,
			)
			if err == nil || locomotive.ObjectID != 0 {
				t.Fatalf("locomotive=%#v err=%v", locomotive, err)
			}
		})
	}
}

func shortTimeoutECoSService() *ECoSService {
	timeout := 40 * time.Millisecond
	service := NewECoSService()
	service.timeout = timeout
	service.client = ecospkg.NewClient(timeout)
	return service
}

func TestECoSSyncAcquiresForcedControlOnlyForTheWrite(t *testing.T) {
	controlled := false
	listener := startECoSTestServer(t, func(command string) []string {
		switch {
		case command == "request(1001, view)":
			return []string{"<REPLY request(1001, view)>", "<END 0 (OK)>"}
		case command == eCoSLocomotiveDetailCommand(1001):
			return []string{fmt.Sprintf("<REPLY %s>", command),
				`1001 name["Old"] addr[3] protocol[DCC]`, "<END 0 (OK)>"}
		case command == "release(1001, view)":
			return []string{"<REPLY release(1001, view)>", "<END 0 (OK)>"}
		case command == "request(1001, view, control, force)":
			controlled = true
			return []string{fmt.Sprintf("<REPLY %s>", command), "<END 0 (OK)>"}
		case strings.HasPrefix(command, "set(1001,"):
			if !controlled {
				return []string{fmt.Sprintf("<REPLY %s>", command),
					"<END 25 (controlled by somebody else)>"}
			}
			return []string{fmt.Sprintf("<REPLY %s>", command), "<END 0 (OK)>"}
		case command == "release(1001, view, control)":
			controlled = false
			return []string{fmt.Sprintf("<REPLY %s>", command), "<END 0 (OK)>"}
		default:
			t.Fatalf("command = %q", command)
			return nil
		}
	})
	defer func() { _ = listener.Close() }()
	host, port := splitTestAddress(t, listener.Addr().String())

	result, err := NewECoSService().SyncLocomotive(t.Context(), ECoSLocomotiveSyncInput{
		Host: host, Port: port, ObjectID: 1001,
		Desired: ECoSLocomotiveSyncDesired{Name: "New"}, Confirm: true,
	})
	if err != nil || result == nil || !result.Applied || controlled {
		t.Fatalf("result=%#v err=%v controlled=%v", result, err, controlled)
	}
}

func TestECoSSyncSeparatesAddressFromOtherAttributes(t *testing.T) {
	written := []string{}
	listener := startECoSTestServer(t, func(command string) []string {
		switch {
		case command == "request(1001, view)":
			return []string{"<REPLY request(1001, view)>", "<END 0 (OK)>"}
		case command == eCoSLocomotiveDetailCommand(1001):
			return []string{fmt.Sprintf("<REPLY %s>", command),
				`1001 name["Old"] addr[3] protocol[DCC]`, "<END 0 (OK)>"}
		case command == "release(1001, view)":
			return []string{"<REPLY release(1001, view)>", "<END 0 (OK)>"}
		case command == "request(1001, view, control, force)":
			return []string{fmt.Sprintf("<REPLY %s>", command), "<END 0 (OK)>"}
		case command == `set(1001, name["New"])`, command == "set(1001, addr[4])":
			written = append(written, command)
			return []string{fmt.Sprintf("<REPLY %s>", command), "<END 0 (OK)>"}
		case strings.HasPrefix(command, "set(1001,"):
			return []string{fmt.Sprintf("<REPLY %s>", command),
				"<END 11 (protocol/addr with other attributes)>"}
		case command == "release(1001, view, control)":
			return []string{fmt.Sprintf("<REPLY %s>", command), "<END 0 (OK)>"}
		default:
			t.Fatalf("command = %q", command)
			return nil
		}
	})
	defer func() { _ = listener.Close() }()
	host, port := splitTestAddress(t, listener.Addr().String())

	result, err := NewECoSService().SyncLocomotive(t.Context(), ECoSLocomotiveSyncInput{
		Host: host, Port: port, ObjectID: 1001,
		Desired: ECoSLocomotiveSyncDesired{Name: "New", Address: 4}, Confirm: true,
	})
	if err != nil || result == nil || !result.Applied {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	want := []string{`set(1001, name["New"])`, "set(1001, addr[4])"}
	if fmt.Sprint(written) != fmt.Sprint(want) || fmt.Sprint(result.Commands) != fmt.Sprint(want) {
		t.Fatalf("written=%#v commands=%#v", written, result.Commands)
	}
}

func TestECoSSyncMarksMissingWriteReplyAsUnknown(t *testing.T) {
	listener := startECoSTestServer(t, func(command string) []string {
		switch {
		case command == "request(1001, view)":
			return []string{"<REPLY request(1001, view)>", "<END 0 (OK)>"}
		case command == "request(1001, view, control, force)":
			return []string{fmt.Sprintf("<REPLY %s>", command), "<END 0 (OK)>"}
		case command == eCoSLocomotiveDetailCommand(1001):
			return []string{fmt.Sprintf("<REPLY %s>", command),
				`1001 name["Old"] addr[3] protocol[DCC]`, "<END 0 (OK)>"}
		case strings.HasPrefix(command, "set(1001,"):
			return nil
		case command == "release(1001, view)":
			return []string{"<REPLY release(1001, view)>", "<END 0 (OK)>"}
		case command == "release(1001, view, control)":
			return []string{fmt.Sprintf("<REPLY %s>", command), "<END 0 (OK)>"}
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
		case command == "request(1001, view, control, force)":
			return []string{fmt.Sprintf("<REPLY %s>", command), "<END 0 (OK)>"}
		case command == eCoSLocomotiveDetailCommand(1001):
			return []string{fmt.Sprintf("<REPLY %s>", command),
				`1001 name["Old"] addr[3] protocol[DCC]`, "<END 0 (OK)>"}
		case strings.HasPrefix(command, "set(1001,"):
			return []string{fmt.Sprintf("<REPLY %s>", command), "<END 11 (unsupported)>"}
		case command == "release(1001, view)":
			return []string{"<REPLY release(1001, view)>", "<END 0 (OK)>"}
		case command == "release(1001, view, control)":
			return []string{fmt.Sprintf("<REPLY %s>", command), "<END 0 (OK)>"}
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
