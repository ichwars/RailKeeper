package application

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	ecospkg "railkeeper/backend/internal/ecos"
)

type storedWebStub struct {
	reads, backups int
	fail           bool
	onRead         func()
}

func (*storedWebStub) Address() string { return "127.0.0.1" }
func (web *storedWebStub) Read(context.Context, int, string) ([]ecospkg.DatabaseCV, error) {
	web.reads++
	if web.onRead != nil {
		web.onRead()
	}
	if web.fail {
		return nil, errors.New("synthetic XML failure")
	}
	return []ecospkg.DatabaseCV{{Number: 1, Value: 3}}, nil
}
func (web *storedWebStub) ReadBackup(context.Context) ([]byte, error) {
	web.backups++
	return nil, fmt.Errorf("unexpected backup request without configured keys")
}

func TestStoredRosterNetworkPathRequestsOnlyMetadataAndFunctionKeys(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	var workers sync.WaitGroup
	var mu sync.Mutex
	commands := []string{}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			workers.Add(1)
			go func() {
				defer workers.Done()
				defer func() { _ = conn.Close() }()
				scanner := bufio.NewScanner(conn)
				for scanner.Scan() {
					command := strings.TrimSpace(scanner.Text())
					mu.Lock()
					commands = append(commands, command)
					mu.Unlock()
					body := ""
					switch {
					case command == eCoSLocomotiveListCommand:
						body = "1002 addr[3] name[\"Test\"] protocol[DCC128]\n"
					case command == eCoSLocomotiveDetailCommand(1002):
						body = "1002 addr[3] name[\"Test\"] protocol[DCC128] profile[\"LokPilot 5 micro DCC\"]\n"
					case strings.HasPrefix(command, "get(1002, funcexists["):
						var index int
						_, _ = fmt.Sscanf(command, "get(1002, funcexists[%d])", &index)
						if index == 1 {
							body = "1002 funcexists[1,12,moment]\n"
						} else {
							body = fmt.Sprintf("1002 funcexists[%d,-1]\n", index)
						}
					case command == "request(1002, view)" || command == "release(1002, view)":
					default:
						t.Errorf("unexpected station command in normal stored read: %s", command)
					}
					_, _ = fmt.Fprintf(conn, "<REPLY %s>\n%s<END 0 (OK)>\n", command, body)
				}
			}()
		}
	}()
	web := &storedWebStub{}
	service := NewECoSService()
	service.storedWebFactory = func(context.Context, string) (storedECoSWeb, error) { return web, nil }
	result, err := service.ReadStoredLocomotiveRoster(t.Context(), ECoSConnectionInput{Host: "127.0.0.1",
		Port: listener.Addr().(*net.TCPAddr).Port})
	var single ECoSCVDatabase
	if err == nil {
		single, err = service.ReadStoredCVDatabase(t.Context(), ECoSConnectionInput{Host: "127.0.0.1",
			Port: listener.Addr().(*net.TCPAddr).Port}, 1002)
	}
	dataDir := t.TempDir()
	private := filepath.Join(dataDir, "private", "ecos-backup")
	if err := os.MkdirAll(private, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"keyring.gpg", "passphrase"} {
		if err := os.WriteFile(filepath.Join(private, file), []byte("synthetic-only"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	service.SetBackupDataDir(dataDir)
	web.fail = true
	partial, partialErr := service.ReadStoredLocomotiveRoster(t.Context(), ECoSConnectionInput{Host: "127.0.0.1",
		Port: listener.Addr().(*net.TCPAddr).Port})
	canceledContext, cancel := context.WithCancel(t.Context())
	defer cancel()
	web.onRead = cancel
	aborted, abortErr := service.ReadStoredLocomotiveRoster(canceledContext, ECoSConnectionInput{Host: "127.0.0.1",
		Port: listener.Addr().(*net.TCPAddr).Port})
	_ = listener.Close()
	<-done
	workers.Wait()
	if err != nil || len(result.Locomotives) != 1 || web.reads != 4 || web.backups != 2 {
		t.Fatalf("stored read failed: %+v %v, HTTP counts %d/%d", result, err, web.reads, web.backups)
	}
	loco := result.Locomotives[0]
	if loco.CVRead == nil || loco.CVRead.NextCV != 1025 || loco.FunctionSnapshot == nil ||
		len(loco.FunctionSnapshot.Functions) != 1 || !loco.FunctionSnapshot.Functions[0].Momentary {
		t.Fatal("stored read lost complete CVs or exact momentary-key mode")
	}
	if len(single.Values) != 1 || single.Values[0].Value != 3 || single.Sources == nil || single.Functions == nil ||
		len(single.Functions.Functions) != 1 || !single.Functions.Functions[0].Momentary {
		t.Fatal("single-locomotive read did not use the same stored data and function pipeline")
	}
	if partialErr != nil || len(partial.Locomotives) != 1 {
		t.Fatalf("CV source failure discarded verified metadata/function keys: %+v %v", partial, partialErr)
	}
	partialLoco := partial.Locomotives[0]
	if partialLoco.CVSources == nil || partialLoco.CVSources.Warning == "" || partialLoco.CVSources.BackupStatus != "failed" ||
		partialLoco.CVRead != nil || len(partialLoco.CVs) != 0 || partialLoco.DetailError == "" ||
		partialLoco.FunctionSnapshot == nil || len(partialLoco.FunctionSnapshot.Functions) != 1 {
		t.Fatalf("partial source data was absent, claimed complete or lost functions: %+v", partialLoco)
	}
	partialItem := newDigitalCenterWorkItem(partialLoco)
	if partialItem.StationStatus != "incomplete" || partialItem.Center["cvSources"] == nil ||
		partialItem.Center["functionSnapshot"] == nil || partialItem.Center["cvRead"] != nil {
		t.Fatal("partial work item cannot distinguish failed sources from unattempted reads")
	}
	if aborted != nil || !errors.Is(abortErr, context.Canceled) {
		t.Fatalf("canceled read retained a fresh-looking partial snapshot: %+v %v", aborted, abortErr)
	}
	mu.Lock()
	defer mu.Unlock()
	for _, command := range commands {
		if strings.Contains(command, "cv[") || strings.Contains(command, ", cv") || strings.Contains(command, "control") || strings.HasPrefix(command, "set(") {
			t.Fatalf("normal read attempted programming/control or diagnostic CV request: %s", command)
		}
	}
}
