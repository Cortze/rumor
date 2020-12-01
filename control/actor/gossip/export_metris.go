package gossip

import (
	"context"
    "fmt"
    "time"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
    "github.com/protolambda/rumor/p2p/track"
)

type GossipExportMetricsCmd struct {
	*base.Base
	*metrics.GossipState
    Store track.ExtendedPeerstore
    ExportPeriod time.Duration `ask:"--export-period" help:"Requets the frecuency in witch the Metrics will be exported to the files"`
	FilePath string `ask:"--file-path" help:"The path of the file where to export the metrics."`
	PeerstorePath string `ask:"--peerstore-path" help:"The path of the file where to export the peerstore."`

}

func (c *GossipExportMetricsCmd) Defaul() {
    c.ExportPeriod = 30 * time.Second
}

func (c *GossipExportMetricsCmd) Help() string {
	return "Exports the Gossip Metrics to the given file path"
}

func (c *GossipExportMetricsCmd) Run(ctx context.Context, args ...string) error {
    if c.GossipState.GsNode == nil {
        return NoGossipErr
    }
    stopping := false
	go func() {
		for {
            if stopping {
                fmt.Println("**************Aborting!!!!!!!!")
                return
            }
			start := time.Now()

            fmt.Println("------------->Exporting!!!!!!!!")
	        err := c.GossipState.ExportMetrics(c.FilePath, c.PeerstorePath, c.Store)
            if err != nil {
                 fmt.Println("Problems exporting the Metrics to the given file path")
            } else {
                fmt.Println("Metrics Exported")
            }
            exportStepDuration := time.Since(start)
			if exportStepDuration < c.ExportPeriod{
				time.Sleep(c.ExportPeriod - exportStepDuration)
			}
		}
	}()
	c.Control.RegisterStop(func(ctx context.Context) error {
		stopping = true
		c.Log.Infof("Stoped Exporting")
		return nil
	})

	return nil
}
