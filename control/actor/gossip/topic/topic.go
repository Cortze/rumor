package topic

import (
	"github.com/protolambda/ask"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/p2p/gossip"
)

type TopicState struct {
	Eth2TopicName string
}

type TopicCmd struct {
	*base.Base
	*gossip.GossipState
	*TopicState
}

func (c *TopicCmd) Help() string {
	return "Manage custom GossipSub topics"
}

func (c *TopicCmd) Cmd(route string) (cmd interface{}, err error) {

	switch route {
	case "create-db":
		cmd = &TopicCreateDBCmd{Base: c.Base, GossipState: c.GossipState, TopicState: c.TopicState}
	case "events":
		cmd = &TopicEventsCmd{Base: c.Base, GossipState: c.GossipState, TopicState: c.TopicState}
	case "join":
		cmd = &TopicJoinCmd{Base: c.Base, GossipState: c.GossipState, TopicState: c.TopicState}
	case "list_peers":
		cmd = &TopicListPeersCmd{Base: c.Base, GossipState: c.GossipState, TopicState: c.TopicState}
	case "leave":
		cmd = &TopicLeaveCmd{Base: c.Base, GossipState: c.GossipState, TopicState: c.TopicState}
	case "log":
		cmd = &TopicLogCmd{Base: c.Base, GossipState: c.GossipState, TopicState: c.TopicState}
	case "publish":
		cmd = &TopicPublishCmd{Base: c.Base, GossipState: c.GossipState, TopicState: c.TopicState}
	default:
		return nil, ask.UnrecognizedErr
	}
	return cmd, nil
}

func (c *TopicCmd) Routes() []string {
	return []string{"create-db", "join", "log", "events", "list_peers", "publish", "leave"}
}
