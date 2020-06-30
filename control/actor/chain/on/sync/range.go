package sync

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/protolambda/rumor/chain"
	bdb "github.com/protolambda/rumor/chain/db/blocks"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/control/actor/flags"
	"github.com/protolambda/rumor/p2p/rpc/methods"
	"github.com/protolambda/rumor/p2p/rpc/reqresp"
	"github.com/protolambda/zrnt/eth2/beacon"
	"time"
)

type ByRangeCmd struct {
	*base.Base

	Blocks bdb.DB
	Chain  chain.FullChain

	PeerID      flags.PeerIDFlag      `ask:"--peer" help:"Peers to make blocks-by-range request to."`
	StartSlot   beacon.Slot           `ask:"--start" help:"Start slot of request"`
	Count       uint64                `ask:"--count" help:"Count of blocks of request"`
	Step        uint64                `ask:"--step" help:"Step between slots of blocks of request"`
	Timeout     time.Duration         `ask:"--timeout" help:"Timeout for full request and response. 0 to disable"`
	Compression flags.CompressionFlag `ask:"--compression" help:"Compression. 'none' to disable, 'snappy' for streaming-snappy"`
	Store       bool                  `ask:"--store" help:"If the blocks should be stored in the blocks DB"`
	Process     bool                  `ask:"--process" help:"If the blocks should be added to the current chain view, ignored otherwise"`
}

func (c *ByRangeCmd) Default() {
	c.Timeout = 20 * time.Second
	c.Compression.Compression = reqresp.SnappyCompression{}
	c.Store = true
	c.Process = true
}

func (c *ByRangeCmd) Help() string {
	return "Sync the chain by slot range."
}

func (c *ByRangeCmd) Run(ctx context.Context, args ...string) error {
	h, err := c.Host()
	if err != nil {
		return err
	}
	sFn := reqresp.NewStreamFn(h.NewStream)

	reqCtx := ctx
	if c.Timeout != 0 {
		reqCtx, _ = context.WithTimeout(reqCtx, c.Timeout)
	}

	method := &methods.BlocksByRangeRPCv1
	peerId := c.PeerID.PeerID

	protocolId := method.Protocol
	if c.Compression.Compression != nil {
		protocolId += protocol.ID("_" + c.Compression.Compression.Name())
	}

	pstore := h.Peerstore()
	if protocols, err := pstore.SupportsProtocols(peerId, string(protocolId)); err != nil {
		return fmt.Errorf("failed to check protocol support of peer %s: %v", peerId.String(), err)
	} else if len(protocols) == 0 {
		return fmt.Errorf("peer %s does not support protocol %s", peerId.String(), protocolId)
	}

	req := methods.BlocksByRangeReqV1{
		StartSlot: c.StartSlot,
		Count:     c.Count,
		Step:      c.Step,
	}
	var block beacon.SignedBeaconBlock
	return method.RunRequest(reqCtx, sFn, peerId, c.Compression.Compression, reqresp.RequestSSZInput{Obj: &req}, req.Count,
		func(chunk reqresp.ChunkedResponseHandler) error {
			resultCode := chunk.ResultCode()
			f := map[string]interface{}{
				"from":        peerId.String(),
				"chunk_index": chunk.ChunkIndex(),
				"chunk_size":  chunk.ChunkSize(),
				"result_code": resultCode,
			}
			switch resultCode {
			case reqresp.ServerErrCode, reqresp.InvalidReqCode:
				msg, err := chunk.ReadErrMsg()
				if err != nil {
					return err
				}
				f["msg"] = msg
				c.Log.WithField("chunk", f).Warn("Received error response")
				return fmt.Errorf("got error response %d on chunk %d: %s", resultCode, chunk.ChunkIndex(), msg)
			case reqresp.SuccessCode:
				// re-use the allocated block for each chunk.
				if err := chunk.ReadObj(&block); err != nil {
					return err
				}
				expectedSlot := beacon.Slot(chunk.ChunkIndex()*req.Step) + req.StartSlot
				if block.Message.Slot != expectedSlot {
					return fmt.Errorf("bad block, expected slot %d, got %d", expectedSlot, block.Message.Slot)
				}
				if c.Store {
					exists, err := c.Blocks.Store(ctx, bdb.WithRoot(&block))
					if err != nil {
						return fmt.Errorf("failed to store block: %v", err)
					}
					f["known"] = exists
				}
				if c.Process {
					if err := c.Chain.AddBlock(ctx, &block); err != nil {
						return fmt.Errorf("failed to process block: %v", err)
					}
				}
				c.Log.WithField("chunk", f).Debug("Received block")
				return nil
			default:
				return fmt.Errorf("received chunk (index %d, size %d) with unknown result code %d", chunk.ChunkIndex(), chunk.ChunkSize(), resultCode)
			}
		})
}
