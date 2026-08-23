package application

import "context"

func (service *DigitalCenterWorkspaceService) runECoSOperation(
	ctx context.Context,
	operation func() error,
) error {
	if service == nil {
		return ErrDigitalCenterWorkspaceUnavailable
	}
	service.operationMu.Lock()
	defer service.operationMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	return operation()
}

func (service *DigitalCenterWorkspaceService) StartReadSession(
	ctx context.Context,
	provider string,
	actor string,
) (session DigitalCenterReadSession, err error) {
	err = service.runECoSOperation(ctx, func() error {
		session, err = service.startReadSessionUnlocked(ctx, provider, actor)
		return err
	})
	return session, err
}

func (service *DigitalCenterWorkspaceService) PreviewWrite(
	ctx context.Context,
	sessionID string,
	itemID string,
	input DigitalCenterWritePreviewInput,
	actor string,
) (preview DigitalCenterWritePreview, err error) {
	err = service.runECoSOperation(ctx, func() error {
		preview, err = service.previewWriteUnlocked(ctx, sessionID, itemID, input, actor)
		return err
	})
	return preview, err
}

func (service *DigitalCenterWorkspaceService) ConfirmWrite(
	ctx context.Context,
	sessionID string,
	itemID string,
	input DigitalCenterWriteConfirmInput,
	actor string,
) (confirmation DigitalCenterWriteConfirmation, err error) {
	err = service.runECoSOperation(ctx, func() error {
		confirmation, err = service.confirmWriteUnlocked(ctx, sessionID, itemID, input, actor)
		return err
	})
	return confirmation, err
}

func (service *DigitalCenterWorkspaceService) StartLiveMonitor(
	ctx context.Context,
	provider string,
	sessionID string,
) (status *ECoSLiveStatus, err error) {
	err = service.runECoSOperation(ctx, func() error {
		status, err = service.startLiveMonitorUnlocked(ctx, provider, sessionID)
		return err
	})
	return status, err
}

func (service *DigitalCenterWorkspaceService) StopLiveMonitor(
	ctx context.Context,
	provider string,
	sessionID string,
) (status *ECoSLiveStatus, err error) {
	err = service.runECoSOperation(ctx, func() error {
		status, err = service.stopLiveMonitorUnlocked(ctx, provider, sessionID)
		return err
	})
	return status, err
}
