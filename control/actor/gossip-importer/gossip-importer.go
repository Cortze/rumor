package gossip-importer

import (
    "github.com/protolambda/ask"
    "github.com/protolambda/rumor/control/actor/base"
    "github.com/protolambda/rumor/blocks"
    "github.com/protolambda/rumor/states"
    "github.com/protolambda/rumor/chain"
    "github.com/protolambda/rumor/metrics"


)

type GossipImporterCmd struct {
    *base.Base
    *metrics.GossipState

    BlocksState *blocks.DBState
    StatesState *states.DBState

    ChainState  *chain.ChainState
}

func (c *GossipImporterCmd) Help() string{
    return "Import the Beacon Blocks and States to the DBs directly from the Gossip Messages (Make sure Gossip Topics are joined and logged."
}

func (c *GossipImporterCmd) Cmd(route string) (cmd interface, error) {
    switch route {
    case "import":
        // Pray for not needing to get and pass the Global State
        cmd = &ImportBlocksAndStatesCmd(Base: c.Base, GossipState: c.GossipState,
        BlockState: c.BlockState, StatesStates: c.StatesStates)
    case "cancel":
        cmd = &CancelCmd(Base: c.Base, GossipState: c.GossipState,
        BlockState: c.BlockState, StatesStates: c.StatesStates)
    default:
        return nil, ask.UnrecognizedErr
    }
    return cmd, nil
}

func (c *GossipImportedCmd) Routes() []String {
    return []string{"import", "cancel"}

}

var NoGossipErr     = errors.New("Must start gossip-sub first. Try 'gossip start'")
var NoBlockTopicErr = errors.New("Must join and log beacon_block topic on gossip")
