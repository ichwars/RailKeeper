package application

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ecospkg "railkeeper/backend/internal/ecos"
)

type rosterFailureWebStub struct {
	storedWebStub
	readError   error
	backupError error
	firstOnly   bool
}

func (web *rosterFailureWebStub) Read(_ context.Context, objectID int, _ string) ([]ecospkg.DatabaseCV, error) {
	web.reads++
	if web.readError != nil && (!web.firstOnly || objectID == 1002) {
		return nil, web.readError
	}
	return []ecospkg.DatabaseCV{{Number: 1, Value: 3}}, nil
}

func (web *rosterFailureWebStub) ReadBackup(context.Context) ([]byte, error) {
	web.backups++
	return nil, web.backupError
}

func TestStoredRosterBoundsHTTPTransportFailuresWithoutLosingTCPData(t *testing.T) {
	timeout := fmt.Errorf("%w: %w", ecospkg.ErrImageWeb,
		&net.OpError{Op: "dial", Net: "tcp", Err: os.ErrDeadlineExceeded})
	for _, test := range []struct {
		name         string
		backup       bool
		web          rosterFailureWebStub
		wantReads    int
		wantBackups  int
		wantComplete int
	}{
		{name: "database timeout", web: rosterFailureWebStub{readError: timeout}, wantReads: 1},
		{name: "backup timeout", backup: true, web: rosterFailureWebStub{backupError: timeout}, wantBackups: 1},
		{name: "one invalid workbook", web: rosterFailureWebStub{readError: ecospkg.ErrCVDatabase, firstOnly: true},
			wantReads: 23, wantComplete: 22},
	} {
		t.Run(test.name, func(t *testing.T) {
			listener := startECoSTestServer(t, func(command string) []string {
				lines := []string{"<REPLY " + command + ">"}
				switch {
				case command == eCoSLocomotiveListCommand:
					for objectID := 1002; objectID < 1025; objectID++ {
						lines = append(lines, fmt.Sprintf(`%d addr[3] name["Test"] protocol[DCC128]`, objectID))
					}
				case strings.Contains(command, "funcexists["):
					var objectID, index int
					_, _ = fmt.Sscanf(command, "get(%d, funcexists[%d])", &objectID, &index)
					lines = append(lines, fmt.Sprintf("%d funcexists[%d,3]", objectID, index))
				case strings.HasPrefix(command, "get("):
					var objectID int
					_, _ = fmt.Sscanf(command, "get(%d,", &objectID)
					lines = append(lines, fmt.Sprintf(`%d addr[3] name["Test"] protocol[DCC128] profile["Generic NMRA"]`, objectID))
				}
				return append(lines, "<END 0 (OK)>")
			})
			defer func() { _ = listener.Close() }()
			host, port := splitTestAddress(t, listener.Addr().String())
			service := NewECoSService()
			service.storedWebFactory = func(context.Context, string) (storedECoSWeb, error) { return &test.web, nil }
			if test.backup {
				dataDir := t.TempDir()
				private := filepath.Join(dataDir, "private", "ecos-backup")
				if err := os.MkdirAll(private, 0o700); err != nil {
					t.Fatal(err)
				}
				for _, name := range []string{"keyring.gpg", "passphrase"} {
					if err := os.WriteFile(filepath.Join(private, name), []byte("synthetic-only"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				service.SetBackupDataDir(dataDir)
			}
			probe, err := service.ReadStoredLocomotiveRoster(t.Context(), ECoSConnectionInput{Host: host, Port: port})
			if err != nil || probe == nil || len(probe.Locomotives) != 23 {
				t.Fatalf("optional HTTP failure discarded TCP roster: %+v %v", probe, err)
			}
			complete := 0
			for _, locomotive := range probe.Locomotives {
				if locomotive.FunctionSnapshot == nil || len(locomotive.FunctionSnapshot.Functions) != 32 {
					t.Fatal("HTTP failure lost verified function keys")
				}
				if locomotive.CVRead != nil {
					complete++
				} else if len(locomotive.CVs) != 0 || locomotive.CVSources == nil ||
					locomotive.CVSources.Warning == "" || locomotive.DetailError == "" {
					t.Fatal("failed HTTP source was presented as complete or without a warning")
				}
			}
			if test.web.reads != test.wantReads || test.web.backups != test.wantBackups || complete != test.wantComplete {
				t.Fatalf("unexpected source attempts or complete snapshots: %d/%d/%d", test.web.reads, test.web.backups, complete)
			}
			if test.wantComplete == 0 {
				previousReads := test.web.reads
				test.web.readError, test.web.backupError = nil, nil
				retry, err := service.ReadStoredLocomotiveRoster(t.Context(), ECoSConnectionInput{Host: host, Port: port})
				if err != nil || retry == nil || len(retry.Locomotives) != 23 || test.web.reads-previousReads != 23 {
					t.Fatalf("new roster did not retry the recovered HTTP transport: %+v %v", retry, err)
				}
				for _, locomotive := range retry.Locomotives {
					if locomotive.CVRead == nil {
						t.Fatal("old transport failure leaked into a new roster snapshot")
					}
				}
			}
		})
	}
}
