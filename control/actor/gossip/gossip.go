package gossip

import (
	"errors"

	"github.com/protolambda/ask"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/p2p/track"
)

type GossipCmd struct {
	*base.Base
	*metrics.GossipState
	Store track.ExtendedPeerstore
}

func (c *GossipCmd) Cmd(route string) (cmd interface{}, err error) {
	switch route {
	case "start":
		cmd = &GossipStartCmd{Base: c.Base, GossipState: c.GossipState}
	case "list":
		cmd = &GossipListCmd{Base: c.Base, GossipState: c.GossipState}
	case "join":
		cmd = &GossipJoinCmd{Base: c.Base, GossipState: c.GossipState}
	case "events":
		cmd = &GossipEventsCmd{Base: c.Base, GossipState: c.GossipState, Store: c.Store}
	case "list-peers":
		cmd = &GossipListPeersCmd{Base: c.Base, GossipState: c.GossipState}
	case "blacklist":
		cmd = &GossipBlacklistCmd{Base: c.Base, GossipState: c.GossipState}
	case "leave":
		cmd = &GossipLeaveCmd{Base: c.Base, GossipState: c.GossipState}
	case "log":
		cmd = &GossipLogCmd{Base: c.Base, GossipState: c.GossipState}
	case "publish":
		cmd = &GossipPublishCmd{Base: c.Base, GossipState: c.GossipState}
  case "export-metrics":
		cmd = &GossipExportMetricsCmd{Base: c.Base, GossipState: c.GossipState, Store: c.Store}
  default:
		return nil, ask.UnrecognizedErr
	}
	return cmd, nil
}

func (c *GossipCmd) Routes() []string {
	return []string{"start", "list", "join", "events", "list-peers", "blacklist", "leave", "log", "publish", "export_metrics"}
}

func (c *GossipCmd) Help() string {
	return "Manage Libp2p GossipSub"
}

var NoGossipErr = errors.New("Must start gossip-sub first. Try 'gossip start'")
