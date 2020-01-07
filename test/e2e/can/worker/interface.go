package worker

type InitializingInputs struct {}
type InitializingOutputs struct {}

type CanaryWorker interface {
	CanaryInitializing(in *InitializingInputs) (out *InitializingOutputs, requeue bool, err error)
}
