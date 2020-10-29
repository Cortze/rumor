package gossip

import (
	"context"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/p2p/track"
	"github.com/sirupsen/logrus"
)

type GossipEventsCmd struct {
	*base.Base
	*metrics.GossipState
	Store     track.ExtendedPeerstore
	TopicName string `ask:"<topic>" help:"The name of the topic to track events of"`
}

func (c *GossipEventsCmd) Help() string {
	return "Listen for events (not messages) on this topic. Events: 'join=<peer-ID>', 'leave=<peer-ID>'"
}

func (c *GossipEventsCmd) Run(ctx context.Context, args ...string) error {
	if c.GossipState.GsNode == nil {
		return NoGossipErr
	}
	top, ok := c.GossipState.Topics.Load(c.TopicName)
	if !ok {
		return fmt.Errorf("not on gossip topic %s", c.TopicName)
	}
	evHandler, err := top.(*pubsub.Topic).EventHandler()
	if err != nil {
		return err
	}
	ctx, cancelEvs := context.WithCancel(ctx)
	go func() {
		c.Log.Infof("Started listening for peer join/leave events for topic %s", c.TopicName)
		for {
			ev, err := evHandler.NextPeerEvent(ctx)
			if err != nil {
				c.Log.Infof("Stopped listening for peer join/leave events for topic %s", c.TopicName)
				return
			}
			switch ev.Type {
			case pubsub.PeerJoin:
				c.GossipState.GossipMetrics.AddNewPeer(ev.Peer.String())
				c.GossipState.GossipMetrics.AddConnectionEvent(ev.Peer.String(), "Connection")
				c.Log.WithFields(logrus.Fields{"peer_id": ev.Peer, "topic": c.TopicName}).Info("topic joined")
				c.GossipState.GossipMetrics.ParseDataFromPeer(c.Store, ev.Peer)

				// TODO: add here the protocolVersion and Ip address
				// 		 add also tha ping to the peer?

			case pubsub.PeerLeave:
				c.GossipState.GossipMetrics.AddConnectionEvent(ev.Peer.String(), "Disconnection")
				c.Log.WithFields(logrus.Fields{"peer_id": ev.Peer, "topic": c.TopicName}).Info("topic left")
				// TODO: add here the protocolVersion and Ip address
				// 		 add also tha ping to the peer?
			}
		}
	}()
	c.Control.RegisterStop(func(ctx context.Context) error {
		cancelEvs()
		return nil
	})
	return nil
}
