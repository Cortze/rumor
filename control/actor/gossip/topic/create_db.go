package topic

import (
	"context"
	"fmt"

	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/p2p/gossip"
)

type TopicCreateDBCmd struct {
	*base.Base
	*gossip.GossipState
	*TopicState

	// Variables might be usable to see if it already exists a db for the given topic
	TopicName   string `ask:"--topic-name" help:"The name of the topic to join"`
	ForkVersion string `ask:"--fork-version" help:"The fork digest value of the network we want to join to"`

	StoreType string `ask:"--store-type" help:"The type of datastore to use. Options: 'mem', 'leveldb', 'badger'"`
	StorePath string `ask:"--store-path" help:"The path of the datastore, must be empty for memory store."`
}

func (c *TopicCreateDBCmd) Default() {
	c.StoreType = "mem"
	c.ForkVersion = "e7a75d5a"
}

func (c *TopicCreateDBCmd) Help() string {
	return "Creates a Database where all the  received messages on the given topics will be stored"
}

func (c *TopicCreateDBCmd) Run(ctx context.Context, args ...string) error {

	fmt.Println("Creating DB for the topic:", c.TopicState.Eth2TopicName)
	// TODO:
	// - Check if a db exits on the same topic
	// - if not generate one
	// - if there is already one, print Error
	// - check also which is the best way of generating the db
	return nil
}
