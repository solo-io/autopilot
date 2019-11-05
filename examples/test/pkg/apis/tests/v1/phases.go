package v1

type TestPhase string

const (

	// Test has begun initializing
	TestPhaseInitializing TestPhase = "Initializing"

	// Test has begun processing
	TestPhaseProcessing TestPhase = "Processing"

	// Test has finished
	TestPhaseFinished TestPhase = "Finished"

	// Test has failed
	TestPhaseFailed TestPhase = "Failed"
)
