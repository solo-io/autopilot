// This package defines defaults which are built-into the system
package defaults

var (
	// configuration file for the autopilot CLI
	// this file will be used to generate and re-generate the autopilot operator
	// it is expected to live at the root of the project repository
	AutoPilotFile = "autopilot.yaml"

	// configuration file for the autopilot operator
	// these files will be loaded at boot time by the operator
	OperatorFile = "autopilot-operator.yaml"

	// Default installation namespace for Istio
	IstioNamespace = "istio-system"
)

const (
	// KubeConfigEnvVar defines the env variable KUBECONFIG which
	// contains the kubeconfig file path.
	KubeConfigEnvVar = "KUBECONFIG"

	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which is the namespace where the watch activity happens.
	// this value is empty if the operator is running with clusterScope.
	WatchNamespaceEnvVar = "WATCH_NAMESPACE"

	// OperatorNameEnvVar is the constant for env variable OPERATOR_NAME
	// which is the name of the current operator
	OperatorNameEnvVar = "OPERATOR_NAME"

	// PodNameEnvVar is the constant for env variable POD_NAME
	// which is the name of the current pod.
	PodNameEnvVar = "POD_NAME"
)
