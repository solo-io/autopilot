package constants

// Internal constants used for internal types

const (
	// this namespace is meant to be used
	// to pass events in-memory through the k8s controller-runtime system.
	// Objects are not intended to be written with this namespace
	InternalNamespace = "INTERNAL_NAMESPACE"
)
