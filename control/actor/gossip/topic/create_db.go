package topic

import (
    "context"
	"fmt"

	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/p2p/gossip"

)


type TopicCreateDBCmd struct {
    *base.Base
	GossipState *gossip.GossipState

	// Variables might be usable to see if it already exists a db for the given topic
	TopicName   string `ask:"--topic-name" help:"The name of the topic to join"`
	ForkVersion string `ask:"--fork-version" help:"The fork digest value of the network we want to join to"`

	Eth2TopicName string `ask:"--eth-topic" help:"The name of the eth2 topics"`
	StoreType     string `ask:"--store-type" help:"The type of datastore to use. Options: 'mem', 'leveldb', 'badger'"`
	StorePath     string `ask:"--store-path" help:"The path of the datastore, must be empty for memory store."`
}

func (c *TopicCreateDBCmd) Default() {
    fmt.Printls("Default settings for the generated database")
}



func (c *TopicCreateDBCmd) Help() string {
    return "Creates a Database where all the  received messages on the given topics will be stored"
}



func (c *TopicCreateDBCmd) Cmd() string {

    fmt.Println("Creating the database ")



}


