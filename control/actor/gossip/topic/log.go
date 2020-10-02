package topic

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/golang/snappy"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/p2p/gossip"

	"github.com/sirupsen/logrus"
)

type TopicLogCmd struct {
	*base.Base
	*gossip.GossipState
	*TopicState

	TopicName   string `ask:"--topic-name" help:"The name of the topic to join"`
	ForkVersion string `ask:"--fork-version" help:"The fork digest value of the network we want to join to"`

	Eth2TopicName string
}

func (c *TopicLogCmd) Help() string {
	return "Log the messages of a gossip topic. Messages are hex-encoded. Join a topic first."
}

func (c *TopicLogCmd) Default() {
	c.ForkVersion = "e7a75d5a"
}

func (c *TopicLogCmd) Run(ctx context.Context, args ...string) (err error) {
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
		sub, err := top.(*pubsub.Topic).Subscribe()
		if err != nil {
			return fmt.Errorf("cannot open subscription on topic %s: %v", c.Eth2TopicName, err)
		}
		ctx, cancelLog := context.WithCancel(ctx)
		go func() {
			defer sub.Cancel()
			for {
				msg, err := sub.Next(ctx)
				if err != nil {
					if err == ctx.Err() { // expected quit, context stopped.
						break
					}
					c.Log.WithError(err).WithField("topic", c.Eth2TopicName).Error("Gossip logging encountered error")
					return
				} else {
					var msgData []byte
					if strings.HasSuffix(c.Eth2TopicName, "_snappy") {
						msgData, err = snappy.Decode(nil, msg.Data)
						if err != nil {
							c.Log.WithError(err).WithField("topic", c.Eth2TopicName).Error("Cannot decompress snappy message")
							continue
						}
					} else {
						msgData = msg.Data
					}
					c.Log.WithFields(logrus.Fields{
						"from":      msg.GetFrom().String(),
						"data":      hex.EncodeToString(msgData),
						"signature": hex.EncodeToString(msg.Signature),
						"seq_no":    hex.EncodeToString(msg.Seqno),
					}).Infof("new message on %s", c.Eth2TopicName)
				}
			}
		}()

		c.Control.RegisterStop(func(ctx context.Context) error {
			cancelLog()
			c.Log.Info("Stopped gossip logger")
			return nil
		})
		return nil
	}
}
