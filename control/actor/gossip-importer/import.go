package gossip-importer

import (
    "fmt"

    "github.com/protolambda/ask"
    "github.com/protolambda/rumor/control/actor/base"
    "github.com/protolambda/rumor/blocks"
    "github.com/protolambda/rumor/states"
    "github.com/protolambda/rumor/chain"
    "github.com/protolambda/rumor/metrics"

)

type ImportBlocksAndStatesCmd struct {
    *base.Base
    //TODO: Change the message DB from the metrics.GossipState 
    // to a new State/Struct defined on the p2p/gossip/database package 
    *metrics.GossipState

    BlocksState *blocks.DBState
    StatesState *states.DBState
    ChainState  *chain.ChainState

    ClientType  string `ask:"--client-type" help:"Give the client type from which the BeaconState will be obtained (prysm by default)"`
    //NOTE: So far it will be Hardcoded just for the Mainnet States and Blocks
    //      not really sure if more networks will need to be added  
}

func (c *ImportBlocksAndStatesCmd) Help() string {
    return "Imports the BeaconState for the blocks received from the GossipSub network protocol by requesting them to a local node (So far just working with Prysm Client, and incoming BeaconBlock import aswell, so we could store them and share them)"
}


func (c *ImportBlocksAndStatesCmd) Default() {
    c.ClientType = "prysm"
}

func (c *ImportBlocksAndStatesCmd) Cmd(ctx context.Context, args ...string) error {
    
    // Check if the beacon_block topic for the Mainnet is enrolled

    // Check if there is actually a MessageDB with the BeaconBlock Message Type ready

    // If both avobe are right now start reading blocks as soon as they are received
    // Generate the go routine to do it
    ctx, cancelImporter := context.WithCancel(ctx)
    go func (bs *blocks.DBState, ss *states.DBState, cs *chain.ChainState, gs *metrics.GossipState) {
        fmt.Println("Block and State Import has been Started")

        }(c.BlocksState, c.StatsState, c.ChainState, c.GossipState, )
    
    // Check just in case the context gets finished
    c.Control.RegisterStop(func(ctx context.Context) error {
        cancelImporter()
        c.Log.Info("Stopped gossip-importer")
        return nil
    })
    return nil
}


