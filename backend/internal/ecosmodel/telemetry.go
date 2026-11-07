package ecosmodel

func AppendBoundedECoSLivePulse(
	samples []ECoSLivePulseSample,
	sample ECoSLivePulseSample,
) []ECoSLivePulseSample {
	samples = append(samples, sample)
	if len(samples) > 60 {
		samples = append([]ECoSLivePulseSample(nil), samples[len(samples)-60:]...)
	}
	return samples
}

func AppendBoundedECoSLiveEvent(events []ECoSLiveEvent, event ECoSLiveEvent) []ECoSLiveEvent {
	events = append(events, event)
	if len(events) > 100 {
		events = append([]ECoSLiveEvent(nil), events[len(events)-100:]...)
	}
	return events
}

func CloneECoSLiveStatus(status ECoSLiveStatus) ECoSLiveStatus {
	status.SubscriptionCommands = append([]string(nil), status.SubscriptionCommands...)
	status.PulseSamples = append([]ECoSLivePulseSample{}, status.PulseSamples...)
	status.RecentEvents = append([]ECoSLiveEvent{}, status.RecentEvents...)
	return status
}

func ECoSLiveSubscriptionCommands() []string {
	return []string{
		"request(1, view)",
		"get(1, info, status)",
	}
}
