package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image/png"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/image/bmp"
	ecospkg "railkeeper/backend/internal/ecos"
)

var ErrECoSImageInput = errors.New("invalid ECoS image transfer input")

type ECoSImageTarget struct {
	ObjectID int    `json:"objectId"`
	Name     string `json:"name"`
	Address  int    `json:"address"`
	Protocol string `json:"protocol"`
}

type ECoSImageOptions struct {
	Host        string            `json:"host"`
	Locomotives []ECoSImageTarget `json:"locomotives"`
}

type ECoSImagePreviewInput struct {
	ObjectID     int
	Description  string
	BMP          []byte
	ExistingSlot *int
}

type ECoSImagePreview struct {
	Token               string              `json:"token"`
	ExpiresAt           string              `json:"expiresAt"`
	Host                string              `json:"host"`
	Target              ECoSImageTarget     `json:"target"`
	Current             ECoSImageDescriptor `json:"current"`
	Slot                int                 `json:"slot"`
	Description         string              `json:"description"`
	SHA256              string              `json:"sha256"`
	CurrentImage        string              `json:"currentImage"`
	CurrentImageMissing bool                `json:"currentImageMissing"`
	NewImage            string              `json:"newImage"`
	ExistingImage       bool                `json:"existingImage"`
}

type ECoSImageConfirmInput struct {
	Token   string `json:"token"`
	Confirm bool   `json:"confirm"`
}

type ECoSImageResult struct {
	Status             string                       `json:"status"` // verified, upload_blocked, upload_unknown, image_not_stored, image_unverified, assignment_unknown, control_busy
	Slot               int                          `json:"slot"`
	ObjectID           int                          `json:"objectId"`
	ImageVerified      bool                         `json:"imageVerified"`
	AssignmentVerified bool                         `json:"assignmentVerified"`
	LiveMonitor        DigitalCenterWriteLiveResult `json:"liveMonitor"`
	AuditRecorded      bool                         `json:"auditRecorded"`
}

type ecosImageGrant struct {
	preview    ECoSImagePreview
	actor      string
	vehicleID  string
	center     DigitalCenterSummary
	pinnedHost string
	data       []byte
	expires    time.Time
}

type ecosImageDevice interface {
	ListLocomotives(context.Context, ECoSConnectionInput) ([]ECoSLocomotive, error)
	ReadLocomotive(context.Context, ECoSConnectionInput, int) (ECoSLocomotive, error)
	ReadImageDescriptor(context.Context, ECoSConnectionInput, int) (ECoSImageDescriptor, error)
	AcquireImageControl(context.Context, ECoSConnectionInput, int) (ECoSImageControl, error)
}

type ecosImageWeb interface {
	Address() string
	FindImageSlot(context.Context, []byte) (int, bool, error)
	CheckUploadForm(context.Context, int) error
	CheckFreeSlot(context.Context, int) error
	CheckStoredSlot(context.Context, int) error
	ReadImage(context.Context, string, int) ([]byte, error)
	Upload(context.Context, int, string, string, []byte) error
}

func (service *DigitalCenterWorkspaceService) imageWeb(ctx context.Context, host string) (ecosImageWeb, error) {
	if service.imageWebFactory != nil {
		return service.imageWebFactory(ctx, host)
	}
	return ecospkg.NewImageWeb(ctx, host)
}

func (service *DigitalCenterWorkspaceService) imageCenter(ctx context.Context, vehicleID string) (DigitalCenterSummary, ecosImageDevice, error) {
	center, err := service.configuredCenter(ctx, "ecos")
	if err != nil {
		return center, nil, err
	}
	if !center.Active {
		return center, nil, ErrDigitalCenterInactive
	}
	device, ok := service.ecos.(ecosImageDevice)
	if !ok || service.vehicles == nil {
		return center, nil, ErrDigitalCenterWorkspaceUnavailable
	}
	vehicles, err := service.vehicles.ListReadOnly(ctx, "")
	if err != nil {
		return center, nil, err
	}
	for _, vehicle := range vehicles {
		if vehicle.ID == vehicleID {
			return center, device, nil
		}
	}
	return center, nil, sql.ErrNoRows
}

func (service *DigitalCenterWorkspaceService) ECoSImageOptions(ctx context.Context, vehicleID string) (result ECoSImageOptions, err error) {
	err = service.runECoSOperation(ctx, func() error {
		center, device, err := service.imageCenter(ctx, vehicleID)
		if err != nil {
			return err
		}
		web, err := service.imageWeb(ctx, center.Host)
		if err != nil {
			return err
		}
		locomotives, err := device.ListLocomotives(ctx, ECoSConnectionInput{Host: web.Address(), Port: center.Port})
		if err != nil {
			return err
		}
		result = ECoSImageOptions{Host: center.Host, Locomotives: []ECoSImageTarget{}}
		for _, locomotive := range locomotives {
			result.Locomotives = append(result.Locomotives, imageTarget(locomotive))
		}
		return nil
	})
	return
}

func imageTarget(locomotive ECoSLocomotive) ECoSImageTarget {
	return ECoSImageTarget{ObjectID: locomotive.ObjectID, Name: locomotive.Name, Address: locomotive.Address, Protocol: locomotive.Protocol}
}

func (service *DigitalCenterWorkspaceService) PreviewECoSImage(ctx context.Context, vehicleID string, input ECoSImagePreviewInput, actor string) (result ECoSImagePreview, err error) {
	if strings.TrimSpace(actor) == "" || input.ObjectID < 1 || input.ObjectID > maxDigitalCenterObjectID ||
		strings.TrimSpace(input.Description) == "" || !utf8.ValidString(input.Description) || utf8.RuneCountInString(input.Description) > 80 || strings.IndexFunc(input.Description, unicode.IsControl) >= 0 {
		return result, ErrECoSImageInput
	}
	// Older device web pages may render descriptions without escaping markup.
	if strings.ContainsAny(input.Description, `<>"'&`) {
		return result, ErrECoSImageInput
	}
	if err = ecospkg.ValidateImageBMP(input.BMP); err != nil {
		return result, err
	}
	if input.ExistingSlot != nil && (*input.ExistingSlot < 0 || *input.ExistingSlot > 499) {
		return result, ErrECoSImageInput
	}
	err = service.runECoSOperation(ctx, func() error {
		center, device, err := service.imageCenter(ctx, vehicleID)
		if err != nil {
			return err
		}
		web, err := service.imageWeb(ctx, center.Host)
		if err != nil {
			return err
		}
		connection := ECoSConnectionInput{Host: web.Address(), Port: center.Port}
		locomotive, err := device.ReadLocomotive(ctx, connection, input.ObjectID)
		if err != nil {
			return err
		}
		if locomotive.ObjectID != input.ObjectID {
			return ErrDigitalCenterDeviceOutput
		}
		descriptor, err := device.ReadImageDescriptor(ctx, connection, input.ObjectID)
		if err != nil {
			return err
		}
		oldImage, currentImageMissing, err := currentECoSImage(ctx, web, descriptor)
		if err != nil {
			return err
		}
		slot, existingImage, err := previewECoSImageSlot(ctx, web, input)
		if err != nil {
			return err
		}
		token, hash, err := newDigitalCenterWriteToken()
		if err != nil {
			return err
		}
		now := service.digitalCenterWriteNow()
		if service.imageGrants == nil {
			service.imageGrants = map[string]ecosImageGrant{}
		}
		// Bound retained image bytes and expire grants lazily. A new preview replaces
		// that user's previous preview; restarts invalidate every confirmation token.
		for key, grant := range service.imageGrants {
			if !grant.expires.After(now) || grant.actor == actor {
				delete(service.imageGrants, key)
			}
		}
		if len(service.imageGrants) >= 64 {
			return ErrDigitalCenterWorkspaceUnavailable
		}
		digest := sha256.Sum256(input.BMP)
		decoded, _ := bmp.Decode(bytes.NewReader(input.BMP))
		var encoded bytes.Buffer
		if err := png.Encode(&encoded, decoded); err != nil {
			return err
		}
		result = ECoSImagePreview{Token: token, ExpiresAt: now.Add(5 * time.Minute).Format(time.RFC3339), Host: center.Host,
			Target: imageTarget(locomotive), Current: descriptor, Slot: slot, Description: strings.TrimSpace(input.Description),
			SHA256: hex.EncodeToString(digest[:]), CurrentImage: oldImage, CurrentImageMissing: currentImageMissing, NewImage: imageDataURL(encoded.Bytes()), ExistingImage: existingImage}
		service.imageGrants[hash] = ecosImageGrant{preview: result, actor: actor, vehicleID: vehicleID, center: center,
			pinnedHost: web.Address(), data: bytes.Clone(input.BMP), expires: now.Add(5 * time.Minute)}
		return nil
	})
	return
}

func imageDataURL(data []byte) string {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
}

func (service *DigitalCenterWorkspaceService) ConfirmECoSImage(ctx context.Context, vehicleID string, input ECoSImageConfirmInput, actor string) (result ECoSImageResult, err error) {
	if !input.Confirm {
		return result, ErrDigitalCenterConfirmationRequired
	}
	err = service.runECoSOperation(ctx, func() error {
		digest := sha256.Sum256([]byte(input.Token))
		hash := hex.EncodeToString(digest[:])
		grant, ok := service.imageGrants[hash]
		if !ok || actor == "" || grant.actor != actor || grant.vehicleID != vehicleID || !grant.expires.After(service.digitalCenterWriteNow()) {
			return ErrDigitalCenterGrantMismatch
		}
		delete(service.imageGrants, hash) // One attempt, including stale/failed attempts.
		center, device, err := service.imageCenter(ctx, vehicleID)
		if err != nil {
			return err
		}
		if center.Host != grant.center.Host || center.Port != grant.center.Port {
			return ErrDigitalCenterPreviewStale
		}
		web, err := service.imageWeb(ctx, center.Host)
		if err != nil {
			return err
		}
		if web.Address() != grant.pinnedHost {
			return ErrDigitalCenterPreviewStale
		}
		return service.confirmECoSImageGrant(ctx, device, web, grant, &result)
	})
	return
}

func (service *DigitalCenterWorkspaceService) confirmECoSImageGrant(ctx context.Context, device ecosImageDevice, web ecosImageWeb, grant ecosImageGrant, result *ECoSImageResult) error {
	connection := ECoSConnectionInput{Host: web.Address(), Port: grant.center.Port}
	preview := grant.preview
	live, err := service.pauseDigitalCenterLive(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if live.WasRunning {
			if monitor, ok := service.ecos.(digitalCenterECoSLiveMonitor); ok {
				resume, cancel := context.WithTimeout(context.WithoutCancel(ctx), 8*time.Second)
				defer cancel()
				start := monitor.StartLive
				if resumer, ok := service.ecos.(interface {
					ResumePausedLive(context.Context, ECoSConnectionInput) (*ECoSLiveStatus, error)
				}); ok {
					start = resumer.ResumePausedLive
				}
				_, err := start(resume, connection)
				live.Restarted = err == nil
			}
		}
		result.LiveMonitor = live
	}()
	locomotive, err := device.ReadLocomotive(ctx, connection, preview.Target.ObjectID)
	if err != nil {
		return err
	}
	descriptor, err := device.ReadImageDescriptor(ctx, connection, preview.Target.ObjectID)
	if err != nil {
		return err
	}
	if imageTarget(locomotive) != preview.Target || descriptor != preview.Current {
		return ErrDigitalCenterPreviewStale
	}
	if err := checkCurrentECoSImage(ctx, web, preview); err != nil {
		return err
	}
	if preview.ExistingImage {
		if err := verifyExistingECoSImage(ctx, web, preview.Slot, grant.data); err != nil {
			return err
		}
	} else {
		if err := web.CheckFreeSlot(ctx, preview.Slot); err != nil {
			return err
		}
	}
	// Persist intent before the first device write. If audit storage is unavailable,
	// do not upload. The token itself and image bytes never enter the audit log.
	if err := service.auditECoSImage(ctx, grant, "confirmed"); err != nil {
		return err
	}
	*result = ECoSImageResult{Status: "upload_blocked", Slot: preview.Slot, ObjectID: preview.Target.ObjectID, ImageVerified: preview.ExistingImage}
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 45*time.Second)
	defer cancel()
	defer func() { result.AuditRecorded = service.auditECoSImage(writeCtx, grant, result.Status) == nil }()
	return completeECoSImageTransfer(writeCtx, device, web, grant, result)
}

func imageUploadKind(kind string) string {
	switch kind {
	case "LOCO_TYPE_E":
		return "e"
	case "LOCO_TYPE_DIESEL":
		return "d"
	case "LOCO_TYPE_STEAM":
		return "s"
	default:
		return "m"
	}
}

func (service *DigitalCenterWorkspaceService) auditECoSImage(ctx context.Context, grant ecosImageGrant, status string) error {
	if service.auth == nil {
		return ErrDigitalCenterWorkspaceUnavailable
	}
	details, _ := json.Marshal(map[string]any{"status": status, "objectId": grant.preview.Target.ObjectID, "slot": grant.preview.Slot, "sha256": grant.preview.SHA256, "existingImage": grant.preview.ExistingImage})
	return service.auth.RecordAudit(ctx, grant.actor, "ECoSImageTransfer", "vehicle", grant.vehicleID, string(details))
}
