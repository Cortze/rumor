package gossip

import (
	"context"
	"errors"
	"fmt"

	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
)

type ExportMetricsCmd struct {
	*base.Base
	*metrics.GossipState
	FilePath string `ask:"--file-path" help:"The path of the file where to export the metrics."`
}

func (c *ExportMetricsCmd) Help() string {
	return "Exports the Gossip Metrics to the given file path"
}

func (c *ExportMetricsCmd) Run(ctx context.Context, args ...string) error {
	if c.GossipState.GsNode == nil {
		return NoGossipErr
	}

	err := c.GossipState.GossipMetrics.ExportMetrics(c.FilePath)
	if err != nil {
		return errors.New("Problems exporting the Metrics to the given file path")
	}

	fmt.Printf("%+v\n", c.GossipState.GossipMetrics)
	return nil
}
