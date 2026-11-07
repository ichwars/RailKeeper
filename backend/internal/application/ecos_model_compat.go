package application

import "railkeeper/backend/internal/ecosmodel"

type ECoSConnectionInput = ecosmodel.ECoSConnectionInput

type ECoSConnectionResult = ecosmodel.ECoSConnectionResult

type ECoSCompatibilityStatus = ecosmodel.ECoSCompatibilityStatus

const (
	ECoSCompatibilityUnreachable = ecosmodel.ECoSCompatibilityUnreachable
	ECoSCompatibilityRejected    = ecosmodel.ECoSCompatibilityRejected
	ECoSCompatibilityPartial     = ecosmodel.ECoSCompatibilityPartial
	ECoSCompatibilityUnverified  = ecosmodel.ECoSCompatibilityUnverified
	ECoSCompatibilityVerified    = ecosmodel.ECoSCompatibilityVerified
)

type ECoSLocomotive = ecosmodel.ECoSLocomotive

type ECoSFunction = ecosmodel.ECoSFunction

type ECoSRawProbe = ecosmodel.ECoSRawProbe

type ECoSLocomotiveSummary = ecosmodel.ECoSLocomotiveSummary

type ECoSRawLocomotive = ecosmodel.ECoSRawLocomotive

type ECoSRawCommandProbe = ecosmodel.ECoSRawCommandProbe

type ECoSCVValue = ecosmodel.ECoSCVValue

type ECoSLiveMonitorState = ecosmodel.ECoSLiveMonitorState

const (
	ECoSLiveStopped     = ecosmodel.ECoSLiveStopped
	ECoSLiveRunning     = ecosmodel.ECoSLiveRunning
	ECoSLiveInterrupted = ecosmodel.ECoSLiveInterrupted
)

type ECoSLivePulseSample = ecosmodel.ECoSLivePulseSample

type ECoSLiveEvent = ecosmodel.ECoSLiveEvent

type ECoSLiveDiagnosis = ecosmodel.ECoSLiveDiagnosis

type ECoSLiveStatus = ecosmodel.ECoSLiveStatus

type ECoSLocomotiveSyncInput = ecosmodel.ECoSLocomotiveSyncInput

type ECoSLocomotiveSyncDesired = ecosmodel.ECoSLocomotiveSyncDesired

type ECoSLocomotiveSyncSnapshot = ecosmodel.ECoSLocomotiveSyncSnapshot

type ECoSLocomotiveSyncChange = ecosmodel.ECoSLocomotiveSyncChange

type ECoSLocomotiveSyncResult = ecosmodel.ECoSLocomotiveSyncResult

func appendBoundedECoSLivePulse(
	samples []ECoSLivePulseSample,
	sample ECoSLivePulseSample,
) []ECoSLivePulseSample {
	return ecosmodel.AppendBoundedECoSLivePulse(samples, sample)
}

func appendBoundedECoSLiveEvent(events []ECoSLiveEvent, event ECoSLiveEvent) []ECoSLiveEvent {
	return ecosmodel.AppendBoundedECoSLiveEvent(events, event)
}

func cloneECoSLiveStatus(status ECoSLiveStatus) ECoSLiveStatus {
	return ecosmodel.CloneECoSLiveStatus(status)
}

func missingECoSConnectionFields(result *ECoSConnectionResult) []string {
	return ecosmodel.MissingECoSConnectionFields(result)
}

func eCoSLiveSubscriptionCommands() []string { return ecosmodel.ECoSLiveSubscriptionCommands() }

func eCoSLocomotiveDetailCommand(objectID int) string {
	return ecosmodel.ECoSLocomotiveDetailCommand(objectID)
}

func eCoSRawProbeFields() []string { return ecosmodel.ECoSRawProbeFields() }

func eCoSTargetedCVProbeFields() []string { return ecosmodel.ECoSTargetedCVProbeFields() }

func parseECoSFields(lines []string) map[string]string { return ecosmodel.ParseECoSFields(lines) }

func parseECoSAttributes(lines []string) map[string][]string {
	return ecosmodel.ParseECoSAttributes(lines)
}

func parseECoSEndStatus(lines []string) (string, bool) { return ecosmodel.ParseECoSEndStatus(lines) }

func sortedECoSFieldNames(fields map[string]bool) []string {
	return ecosmodel.SortedECoSFieldNames(fields)
}

func sortedMissingECoSFieldNames(missing map[string]bool, supported map[string]bool) []string {
	return ecosmodel.SortedMissingECoSFieldNames(missing, supported)
}

func interestingECoSFields(attributes map[string][]string) []string {
	return ecosmodel.InterestingECoSFields(attributes)
}

func parseECoSCVValues(attributes map[string][]string) []ECoSCVValue {
	return ecosmodel.ParseECoSCVValues(attributes)
}

func sortECoSCVValues(values []ECoSCVValue) { ecosmodel.SortECoSCVValues(values) }

func parseECoSLocomotives(lines []string) []ECoSLocomotive {
	return ecosmodel.ParseECoSLocomotives(lines)
}

func mergeECoSRawLocomotive(target *ECoSRawLocomotive, source ECoSLocomotive) {
	ecosmodel.MergeECoSRawLocomotive(target, source)
}

func cleanECoSLocomotiveSyncDesired(input ECoSLocomotiveSyncDesired) ECoSLocomotiveSyncDesired {
	return ecosmodel.CleanECoSLocomotiveSyncDesired(input)
}

func buildECoSLocomotiveSyncCommands(objectID int, current *ECoSLocomotive, desired ECoSLocomotiveSyncDesired) ([]ECoSLocomotiveSyncChange, []string, error) {
	return ecosmodel.BuildECoSLocomotiveSyncCommands(objectID, current, desired)
}

func eCoSLocomotiveSyncFields(changes []ECoSLocomotiveSyncChange) []string {
	return ecosmodel.ECoSLocomotiveSyncFields(changes)
}

func quoteECoSString(value string) string { return ecosmodel.QuoteECoSString(value) }

func firstNonEmpty(values ...string) string { return ecosmodel.FirstNonEmpty(values...) }
