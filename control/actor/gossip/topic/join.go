package topic

import (
	"context"
	"fmt"

	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/p2p/gossip"
)

type TopicJoinCmd struct {
	*base.Base
	*gossip.GossipState
	*TopicState

	TopicName   string `ask:"--topic-name" help:"The name of the topic to join"`
	ForkVersion string `ask:"--fork-version" help:"The fork digest value of the network we want to join to"`
}

func (c *TopicJoinCmd) Help() string {
	return "Join a gossip topic. This only sets up the topic, it does not actively find peers. See `gossip log start` and `gossip publish`."
}

func (c *TopicJoinCmd) Default() {
	c.ForkVersion = "e7a75d5a"
}

func (c *TopicJoinCmd) Run(ctx context.Context, args ...string) (err error) {
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

	_, ok := c.GossipState.Topics.Load(c.TopicState.Eth2TopicName)
	if ok {
		return fmt.Errorf("already on gossip topic %s", c.TopicState.Eth2TopicName)
	}
	top, err := c.GossipState.GsNode.Join(c.TopicState.Eth2TopicName)
	if err != nil {
		return err
	}
	c.GossipState.Topics.Store(c.TopicState.Eth2TopicName, top)
	c.Log.Infof("joined topic %s", c.TopicState.Eth2TopicName)
	return nil
}
