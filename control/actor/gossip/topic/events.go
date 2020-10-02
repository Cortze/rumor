package topic

import (
	"context"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/p2p/gossip"

	"github.com/sirupsen/logrus"
)

type TopicEventsCmd struct {
	*base.Base
	*gossip.GossipState
	*TopicState

	TopicName   string `ask:"--topic-name" help:"The name of the topic to join"`
	ForkVersion string `ask:"--fork-version" help:"The fork digest value of the network we want to join to"`

	Eth2TopicName string
}

func (c *TopicEventsCmd) Help() string {
	return "Listen for events (not messages) on this topic. Events: 'join=<peer-ID>', 'leave=<peer-ID>'"
}

func (c *TopicEventsCmd) Default() {
	c.ForkVersion = "e7a75d5a"
}

func (c *TopicEventsCmd) Run(ctx context.Context, args ...string) (err error) {
	// Generate the full address of the eth2 topics
	c.Eth2TopicName, err = gossip.Eth2TopicBuilder(c.TopicName, c.ForkVersion)
	if err != nil {
		return fmt.Errorf("Error while generating the Full Eth2 Topic-Name")
	}

	// Temporal code
	fmt.Println("full address will be:", c.Eth2TopicName)
	// --- end Temporal code ---

	if c.GossipState.GsNode == nil {
		return gossip.NoGossipErr
	}
	top, ok := c.GossipState.Topics.Load(c.Eth2TopicName)
	if !ok {
		return fmt.Errorf("not on gossip topic %s", c.Eth2TopicName)
	}
	evHandler, err := top.(*pubsub.Topic).EventHandler()
	if err != nil {
		return err
	}
	ctx, cancelEvs := context.WithCancel(ctx)
	go func() {
		c.Log.Infof("Started listening for peer join/leave events for topic %s", c.Eth2TopicName)
		for {
			ev, err := evHandler.NextPeerEvent(ctx)
			if err != nil {
				c.Log.Infof("Stopped listening for peer join/leave events for topic %s", c.Eth2TopicName)
				return
			}
			switch ev.Type {
			case pubsub.PeerJoin:
				c.Log.WithFields(logrus.Fields{"peer_id": ev.Peer, "topic": c.Eth2TopicName}).Info("topic joined")
			case pubsub.PeerLeave:
				c.Log.WithFields(logrus.Fields{"peer_id": ev.Peer, "topic": c.Eth2TopicName}).Info("topic left")
			}
		}
	}()
	c.Control.RegisterStop(func(ctx context.Context) error {
		cancelEvs()
		return nil
	})
	return nil
}
