package gossip

import (
	"github.com/protolambda/ask"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/control/actor/gossip/topic"
	"github.com/protolambda/rumor/p2p/gossip"
)

type GossipCmd struct {
	*base.Base
	*gossip.GossipState
	*topic.TopicState
}

func (c *GossipCmd) Cmd(route string) (cmd interface{}, err error) {
	switch route {
	case "start":
		cmd = &GossipStartCmd{Base: c.Base, GossipState: c.GossipState}
	case "list":
		cmd = &GossipListCmd{Base: c.Base, GossipState: c.GossipState}
	case "blacklist":
		cmd = &GossipBlacklistCmd{Base: c.Base, GossipState: c.GossipState}
	case "topic":
		cmd = &topic.TopicCmd{Base: c.Base, GossipState: c.GossipState, TopicState: c.TopicState}
	default:
		return nil, ask.UnrecognizedErr
	}
	return cmd, nil
}

func (c *GossipCmd) Routes() []string {
	return []string{"start", "topic", "list", "blacklist"}
}

func (c *GossipCmd) Help() string {
	return "Manage Libp2p GossipSub"
}
