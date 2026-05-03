package sync

type State string

const (
	StateClean            State = "clean"
	StateLocalModified    State = "local_modified"
	StateRegistryModified State = "registry_modified"
	StateConflict         State = "conflict"
)

func evaluateState(currentRegistryHash, currentLocalHash, lockRegistryHash, lockLocalHash string) State {
	if lockRegistryHash == "" && lockLocalHash == "" {
		return StateRegistryModified
	}
	registryChanged := currentRegistryHash != lockRegistryHash
	localChanged := currentLocalHash != lockLocalHash
	switch {
	case registryChanged && localChanged:
		return StateConflict
	case registryChanged:
		return StateRegistryModified
	case localChanged:
		return StateLocalModified
	default:
		return StateClean
	}
}
