package topic


import(

    "github.com/protolambda/ask"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/p2p/gossip"

    "github.com/protolambda/rumor/p2p/track"
    "github.com/protolambda/rumor/metrics"
)

type TopicCmd struct {
    *base.Base
    *gossip.GossipState
    Store track.ExtendedPeerstore
}


func (c *TopicCmd) Help() string {
    return "Manage custom GossipSub topics"
}


func (c *TopicCmd) Cmd(route string) (cmd interface{}, err error) {

    switch route {
    case "create-db":
        cmd = &TopicCreateDBCmd{Base: c.Base, GossipState: c.GossipState, GossipState: c.GossipState}
	case "events":
		cmd = &TopicEventsCmd{Base: c.Base, GossipState: c.GossipState, GossipState: c.GossipState, Store: c.Store}
	case "join":
		cmd = &TopicJoinCmd{Base: c.Base, GossipState: c.GossipState, GossipState: c.GossipState}
	case "list_peers":
		cmd = &TopicListPeersCmd{Base: c.Base, GossipState: c.GossipState, GossipState: c.GossipState}
	case "leave":
		cmd = &TopicLeaveCmd{Base: c.Base, GossipState: c.GossipState, GossipState: c.GossipState}
	case "log":
		cmd = &TopicLogCmd{Base: c.Base, GossipState: c.GossipState, GossipState: c.GossipState}
	case "publish":
		cmd = &TopicPublishCmd{Base: c.Base, GossipState: c.GossipState, GossipState: c.GossipState}
	case "export-metrics":
        cmd = &TopicExportMetricsCmd{Base: c.Base, GossipState: c.GossipState, GossipState: c.GossipState, Store: c.Store}
    default:
		return nil, ask.UnrecognizedErr
	}
	return cmd, nil
}

func (c *TopicCmd) Routes() []string {
    return []string{"create-db", "join", "log", "events", "list_peers", "publish", "leave", "export-metrics"}
}

var NoGossipErr = errors.New("Must start gossip-sub first. Try 'gossip start'")
