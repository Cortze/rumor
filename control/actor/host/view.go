package host

import (
	"context"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/p2p/addrutil"
)

type HostViewCmd struct {
	*base.Base
}

func (c *HostViewCmd) Help() string {
	return "View local peer ID, listening addresses, etc."
}

func (c *HostViewCmd) Run(ctx context.Context, args ...string) error {
	h, err := c.Host()
	if err != nil {
		return err
	}
	c.Log.WithField("peer_id", h.ID()).Info("Peer ID")
	for i, a := range h.Addrs() {
		c.Log.WithField("multi_addr", a.String()+"/p2p/"+h.ID().String()).Infof("Listening address %d", i)
	}
	enr, err := addrutil.EnrToString(c.GetEnr())
	if err != nil {
		return err
	}
	c.Log.WithField("enr", enr).Info("ENR")
	return nil
}
