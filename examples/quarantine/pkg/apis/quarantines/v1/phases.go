package v1

type QuarantinePhase string

const (

	// Quarantine has begun initializing
	QuarantinePhaseInitializing QuarantinePhase = "Initializing"

	// Quarantine has begun processing
	QuarantinePhaseProcessing QuarantinePhase = "Processing"

	// Quarantine has finished
	QuarantinePhaseFinished QuarantinePhase = "Finished"
)
