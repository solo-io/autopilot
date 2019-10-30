package progressing
import metrics "github.com/solo-io/autopilot/pkg/metrics"
import aliases "github.com/solo-io/autopilot/pkg/aliases"

type Inputs struct {
    Metrics metrics.Metrics
}

type Outputs struct {
    TrafficSplits aliases.TrafficSplits
}
