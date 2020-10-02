package gossip

import (
	"context"
	"fmt"

	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/p2p/gossip"
)

type GossipListCmd struct {
	*base.Base
	*gossip.GossipState
}

func (c *GossipListCmd) Help() string {
	return "List joined gossip topics"
}

func (c *GossipListCmd) Run(ctx context.Context, args ...string) error {
	if c.GossipState.GsNode == nil {
		return gossip.NoGossipErr
	}
	// Temporal
	fmt.Println(c.GossipState.Topics)
	//--- end temporal ---

	// ---- Original code ------
	//topics := make([]string, 0)
	//c.GossipState.Topics.Range(func(key, value interface{}) bool {
	//	topics = append(topics, key.(string))
	//	return false
	//})
	//c.Log.WithField("topics", topics).Infof("On %d topics.", len(topics))
	//return nil
	//----- End of original code ------

	topics := make([]string, 0)

	fmt.Println(c.GossipState.Topics)
	c.GossipState.Topics.Range(func(key, value interface{}) bool {
		fmt.Println("key inside the de sync.Map", key)
		topics = append(topics, key.(string))
		return false
	})
	c.Log.WithField("topics", topics).Infof("On %d topics.", len(topics))
	return nil
}
