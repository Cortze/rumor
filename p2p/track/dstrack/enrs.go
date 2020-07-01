package dstrack

import (
	"github.com/ethereum/go-ethereum/p2p/enode"
	ds "github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/rumor/p2p/addrutil"
	"github.com/protolambda/rumor/p2p/track"
	"github.com/protolambda/rumor/p2p/types"
)

type dsENRBook struct {
	ds ds.Datastore

	// Latest ENR eth2 data
	enrEth2 *types.Eth2Data
	// Latest ENR attnets data
	enrAttnets types.AttnetBits

	// Track ENR
	n *enode.Node
}

var _ track.ENRBook = (*dsENRBook)(nil)

func NewENRBook(store ds.Datastore) (*dsENRBook, error) {
	return &dsENRBook{ds: store}, nil
}

// Update the record tracking of the peer,
// return updated=true if the node is new, or it overrides a previously seen node (by higher seq nr).
// and return eth2 and attnet data, if any.
func (eb *dsENRBook) UpdateENRMaybe(n *enode.Node) (updated bool, data *types.Eth2Data, attnetbits *types.AttnetBits, err error) {
	pi.Lock()
	defer pi.Unlock()
	if pi.n != nil {
		if pi.n.Seq() >= n.Seq() {
			return false, nil, nil, nil
		}
	}
	pi.n = n
	data, attnets, err := handleNewEnr(n)
	return true, data, attnets, err
}

// Latest fetches the latest ENR of the peer, nil if we have none. The returned ENR may not be mutated.
func (eb *dsENRBook) LatestENR() (n *enode.Node) {
	return pi.n
}

func (eb *dsENRBook) flush() error {
	var clErr error
	// store all statuses to datastore before exiting
	eb.data.Range(func(key, value interface{}) bool {
		id := key.(peer.ID)
		st := value.(*Status)
		if err := sb.storeStatus(id, st); err != nil {
			clErr = err
			return false
		}
		return true
	})
	return clErr
}

func (eb *dsENRBook) Close() error {
	return eb.flush()
}

func handleNewEnr(n *enode.Node) (data *types.Eth2Data, attnetbits *types.AttnetBits, err error) {
	var eth2 addrutil.Eth2ENREntry
	if err := n.Load(&eth2); err == nil {
		dat, err := eth2.Eth2Data()
		if err == nil {
			data = dat
		} else {
			return nil, nil, err
		}
	}
	var attnets addrutil.AttnetsENREntry
	if err := n.Load(&attnets); err == nil {
		dat, err := attnets.AttnetBits()
		if err == nil {
			attnetbits = &dat
		} else {
			return nil, nil, err
		}
	}
	return
}