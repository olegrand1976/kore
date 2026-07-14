package kernel

type RequestPriority string

const (
	PriorityLow    RequestPriority = "low"
	PriorityNormal RequestPriority = "normal"
	PriorityHigh   RequestPriority = "high"
	PriorityUrgent RequestPriority = "urgent"
)

func NormalizeRequestPriority(raw string) RequestPriority {
	switch RequestPriority(raw) {
	case PriorityLow, PriorityHigh, PriorityUrgent:
		return RequestPriority(raw)
	default:
		return PriorityNormal
	}
}
