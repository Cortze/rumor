package topic

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/snappy"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/p2p/gossip"
)

type TopicPublishCmd struct {
	*base.Base
	*gossip.GossipState
	*TopicState

	TopicName   string `ask:"--topic-name" help:"The name of the topic to join"`
	ForkVersion string `ask:"--fork-version" help:"The fork digest value of the network we want to join to"`
	Message     []byte `ask:"<message>" help:"The uncompressed message bytes, hex-encoded"`

	Eth2TopicName string
}

func (c *TopicPublishCmd) Help() string {
	return "Publish a message to the topic. The message should be hex-encoded."
}

func (c *TopicPublishCmd) Default() {
	c.ForkVersion = "e7a75d5a"
}

func (c *TopicPublishCmd) Run(ctx context.Context, args ...string) (err error) {
	if c.GossipState.GsNode == nil {
		return gossip.NoGossipErr
	}

	// Generate the full address of the eth2 topics
	c.Eth2TopicName, err = gossip.Eth2TopicBuilder(c.TopicName, c.ForkVersion)
	if err != nil {
		return fmt.Errorf("Error while generating the Full Eth2 Topic-Name")
	}

	// Temporal code
	fmt.Println("full address will be:", c.Eth2TopicName)
	// --- end Temporal code ---

	if top, ok := c.GossipState.Topics.Load(c.Eth2TopicName); !ok {
		return fmt.Errorf("not on gossip topic %s", c.Eth2TopicName)
	} else {
		data := c.Message
		if strings.HasSuffix(c.Eth2TopicName, "_snappy") {
			data = snappy.Encode(nil, data)
		}
		if err := top.(*pubsub.Topic).Publish(ctx, data); err != nil {
			return fmt.Errorf("failed to publish message, err: %v", err)
		}
		return nil
	}
}
