package common

const (
	PodQueue = "pod-queue"
)

// Event ...
type Event struct {
	Key          string
	EventType    string
	ResourceType string
}
