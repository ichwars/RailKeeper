package application

import (
	"context"
	"sync"
	"time"

	ecospkg "railkeeper/backend/internal/ecos"
)

const (
	defaultECoSPort           = ecospkg.DefaultPort
	eCoSLiveIdleTimeout       = 20 * time.Minute
	eCoSLiveReadDeadline      = 750 * time.Millisecond
	eCoSLocomotiveListCommand = "queryObjects(10, addr, name, protocol)"
)

type ECoSService struct {
	timeout                  time.Duration
	client                   ecospkg.Client
	backupMu                 sync.RWMutex
	backupDataDir            string
	storedWebFactory         func(context.Context, string) (storedECoSWeb, error)
	liveMu                   sync.Mutex
	liveCancel               context.CancelFunc
	liveDone                 chan struct{}
	liveStatus               ECoSLiveStatus
	livePulse                time.Time
	liveCount                int
	liveGeneration           uint64
	liveInterruptionNotified bool
	liveOnInterrupted        func()
}

func NewECoSService() *ECoSService {
	timeout := 8 * time.Second
	return &ECoSService{
		timeout: timeout,
		client:  ecospkg.NewClient(timeout),
		liveStatus: ECoSLiveStatus{
			Provider:  "ecos",
			State:     ECoSLiveStopped,
			Diagnosis: ECoSLiveDiagnosis{ConnectionState: ECoSLiveStopped, Passive: true},
			Message:   "Keine ECoS-Live-Verbindung aktiv.",
		},
	}
}
