package codegen

type TemplateData struct {
	Project

	ProjectPackage string // e.g. "github.com/solo-io/autopilot/examples/promoter"

	Group   string // e.g. "mesh.demos.io"
	Version string // e.g. "v1"

	TypesImportPath     string // e.g. "github.com/solo-io/autopilot/examples/promoter/pkg/apis/canaries/v1"
	SchedulerImportPath string // e.g. "github.com/solo-io/autopilot/examples/promoter/pkg/scheduler"
	ConfigImportPath    string // e.g. "github.com/solo-io/autopilot/examples/promoter/pkg/config"
	KindLowerCamel      string // e.g. "canary"
}
