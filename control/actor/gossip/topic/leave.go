package topic

import (
	"context"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/p2p/gossip"
)

type TopicLeaveCmd struct {
	*base.Base
	*gossip.GossipState
	*TopicState

	TopicName   string `ask:"--topic-name" help:"The name of the topic to join"`
	ForkVersion string `ask:"--fork-version" help:"The fork digest value of the network we want to join to"`
}

func (c *TopicLeaveCmd) Help() string {
	return "Leave a gossip topic."
}

func (c *TopicLeaveCmd) Default() {
	c.ForkVersion = "e7a75d5a"
}

func (c *TopicLeaveCmd) Run(ctx context.Context, args ...string) (err error) {
	if c.GossipState.GsNode == nil {
		return gossip.NoGossipErr
	}

	// Generate the full address of the eth2 topics
	c.TopicState.Eth2TopicName, err = gossip.Eth2TopicBuilder(c.TopicName, c.ForkVersion)
	if err != nil {
		return fmt.Errorf("Error while generating the Full Eth2 Topic-Name")
	}

	// Temporal code
	fmt.Println("full address will be:", c.TopicState.Eth2TopicName)
	// --- end Temporal code ---

	if top, ok := c.GossipState.Topics.Load(c.TopicState.Eth2TopicName); !ok {
		return fmt.Errorf("not on gossip topic %s", c.TopicState.Eth2TopicName)
	} else {
		err := top.(*pubsub.Topic).Close()
		if err != nil {
			return err
		}
		c.GossipState.Topics.Delete(c.TopicState.Eth2TopicName)
		return nil
	}
}
