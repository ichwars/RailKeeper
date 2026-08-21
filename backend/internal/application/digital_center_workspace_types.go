package application

import (
	"context"
	"errors"
)

type DigitalCenterCompareStatus string
type DigitalCenterSessionState string
type DigitalCenterMessageSeverity string

const (
	DigitalCompareOK        DigitalCenterCompareStatus = "ok"
	DigitalCompareDeviation DigitalCenterCompareStatus = "deviation"
	DigitalCompareMissing   DigitalCenterCompareStatus = "missing"
	DigitalCompareNew       DigitalCenterCompareStatus = "new"
	DigitalCompareConflict  DigitalCenterCompareStatus = "conflict"

	DigitalCenterSessionReading     DigitalCenterSessionState = "reading"
	DigitalCenterSessionReady       DigitalCenterSessionState = "ready"
	DigitalCenterSessionInterrupted DigitalCenterSessionState = "interrupted"
	DigitalCenterSessionFailed      DigitalCenterSessionState = "failed"

	DigitalCenterMessageInfo    DigitalCenterMessageSeverity = "info"
	DigitalCenterMessageWarning DigitalCenterMessageSeverity = "warning"
	DigitalCenterMessageError   DigitalCenterMessageSeverity = "error"
)

var (
	ErrDigitalCenterGrantConsumed      = errors.New("digital center write grant already consumed")
	ErrDigitalCenterGrantExpired       = errors.New("digital center write grant expired")
	ErrDigitalCenterGrantActorMismatch = errors.New("digital center write grant actor mismatch")
)

type DigitalCenterCapabilities struct {
	TestConnection   bool `json:"testConnection"`
	ReadLocomotives  bool `json:"readLocomotives"`
	LiveMonitor      bool `json:"liveMonitor"`
	WriteLocomotives bool `json:"writeLocomotives"`
	WriteCVs         bool `json:"writeCVs"`
	Diagnose         bool `json:"diagnose"`
}

type DigitalCenterReadSession struct {
	ID              string                    `json:"id"`
	Provider        string                    `json:"provider"`
	State           DigitalCenterSessionState `json:"state"`
	Host            string                    `json:"host"`
	Port            int                       `json:"port"`
	Capabilities    DigitalCenterCapabilities `json:"capabilities"`
	ReadStartedAt   string                    `json:"readStartedAt"`
	ReadCompletedAt string                    `json:"readCompletedAt"`
	CreatedByUserID string                    `json:"createdByUserId"`
	CreatedAt       string                    `json:"createdAt"`
	UpdatedAt       string                    `json:"updatedAt"`
}

type DigitalCenterWorkItem struct {
	ID             string                     `json:"id"`
	SessionID      string                     `json:"sessionId"`
	CenterObjectID string                     `json:"centerObjectId"`
	VehicleID      string                     `json:"vehicleId"`
	Name           string                     `json:"name"`
	Address        int                        `json:"decoderAddress"`
	Protocol       string                     `json:"protocol"`
	CompareStatus  DigitalCenterCompareStatus `json:"compareStatus"`
	StationStatus  string                     `json:"stationStatus"`
	Center         map[string]any             `json:"center"`
	RailKeeper     map[string]any             `json:"railkeeper"`
	Proposed       map[string]any             `json:"proposed"`
	Conflicts      []map[string]any           `json:"conflicts"`
	CreatedAt      string                     `json:"createdAt"`
	UpdatedAt      string                     `json:"updatedAt"`
}

type DigitalCenterSessionMessage struct {
	ID         string                       `json:"id"`
	SessionID  string                       `json:"sessionId"`
	Severity   DigitalCenterMessageSeverity `json:"severity"`
	Code       string                       `json:"code"`
	Message    string                       `json:"message"`
	NextAction string                       `json:"nextAction"`
	CreatedAt  string                       `json:"createdAt"`
}

type DigitalCenterWriteGrant struct {
	ID          string `json:"id"`
	TokenHash   string `json:"-"`
	SessionID   string `json:"sessionId"`
	WorkItemID  string `json:"workItemId"`
	PreviewHash string `json:"previewHash"`
	ActorUserID string `json:"actorUserId"`
	ExpiresAt   string `json:"expiresAt"`
	ConsumedAt  string `json:"consumedAt"`
	CreatedAt   string `json:"createdAt"`
}

type DigitalCenterWorkspaceRepository interface {
	CreateSession(context.Context, DigitalCenterReadSession) (DigitalCenterReadSession, error)
	UpdateSession(context.Context, DigitalCenterReadSession) error
	GetSession(context.Context, string) (DigitalCenterReadSession, error)
	ReplaceWorkItems(context.Context, string, []DigitalCenterWorkItem) error
	ListWorkItems(context.Context, string) ([]DigitalCenterWorkItem, error)
	GetWorkItem(context.Context, string, string) (DigitalCenterWorkItem, error)
	AddMessage(context.Context, DigitalCenterSessionMessage) error
	ListMessages(context.Context, string) ([]DigitalCenterSessionMessage, error)
	CreateWriteGrant(context.Context, DigitalCenterWriteGrant) error
	ConsumeWriteGrant(context.Context, string, string) (DigitalCenterWriteGrant, error)
}
