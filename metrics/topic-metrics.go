package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pgossip "github.com/protolambda/rumor/p2p/gossip"
	"github.com/protolambda/rumor/p2p/track"
)

type GossipState struct {
	GsNode  pgossip.GossipSub
	CloseGS context.CancelFunc
	// string -> *pubsub.Topic
	Topics sync.Map
	// Metrics for Gossip Messages
	GossipMetrics GossipMetrics
}

// GossipMetrics struct will contain sync.Map (an array) of PeersMetrics sync.Maps.Store("PeerId", []PeerMetrics)
type GossipMetrics struct {
	PeersMetrics        []PeerMetrics
	MetricsStartingTime int64
}

// Base Struct for the topic name and the received messages on the different topics
type PeerMetrics struct {
	PeerId     string
	NodeId     string
	ClientType string
	Pubkey     string
	Addres     []string
	Latency    time.Duration

	ConnectionEvents []ConnectionEvents
	// Counters for the different topics
	BeaconBlock          MessageMetrics
	BeaconAggregateProof MessageMetrics
	VoluntaryExit        MessageMetrics
	ProposerSlashing     MessageMetrics
	AttesterSlashing     MessageMetrics
	// Variables related to the SubNets (only needed for when Shards will be implemented)
}

// Connection event model
type ConnectionEvents struct {
	ConnectionType string
	TimeMili       int64
}

// Information regarding the messages received on the beacon_lock topic
type MessageMetrics struct {
	Cnt              int64
	FirstMessageTime int64
	LastMessageTime  int64
}

// Function that Exports the entire Metrics to a .json file (lets see if in the future we can add websockets or other implementations)
func (c *GossipMetrics) ExportMetrics(filePath string) error {
	metrics, err := json.Marshal(c)
	if err != nil {
		fmt.Println("Error Marshalling the metrics")
	}

	err = ioutil.WriteFile("output.json", metrics, 0644)
	if err != nil {
		fmt.Println("Error opening file: ", filePath)
		return err
	}
	return nil
}

// Staps the time when the metrics started to get taken
func (c *GossipMetrics) StampStartingTime() {
	unixMillis := GetTimeMiliseconds()
	c.MetricsStartingTime = unixMillis
}

func (c *GossipMetrics) AddNewPeer(peerId string) []PeerMetrics {
	// Check if the peer already exists
	_, err := c.GetPeerIndex(peerId)
	if err != nil { // Only add it if it doesn't exist
		metrics := PeerMetrics{PeerId: peerId}
		c.PeersMetrics = append(c.PeersMetrics, metrics)
	}
	return c.PeersMetrics
}

// Returns the index of the peer and a false error, and a 0 index and error
func (c *GossipMetrics) GetPeerIndex(peerId string) (n int, err error) {
	for i := 0; i < len(c.PeersMetrics); i++ {
		if c.PeersMetrics[i].PeerId == peerId {
			return i, nil
		}
	}
	return 0, errors.New("Peer not found on the list")
}

func (c *GossipMetrics) AddConnectionEvent(peerId string, connectionType string) {
	newConnection := ConnectionEvents{
		ConnectionType: connectionType,
		TimeMili:       GetTimeMiliseconds(),
	}
	peerIndex, err := c.GetPeerIndex(peerId)
	if err != nil {
		errors.New("Counld't add Event, Peer is not in the list")
	}
	c.PeersMetrics[peerIndex].ConnectionEvents = append(c.PeersMetrics[peerIndex].ConnectionEvents, newConnection)
}

func (c *GossipMetrics) ParseDataFromPeer(ep track.ExtendedPeerstore, peerId peer.ID) {
	peerIndex, _ := c.GetPeerIndex(peerId.String())

	// get peer.id functionalities from the id pased
	// Get all the data fromt he peer
	peerData := ep.GetAllData(peerId)
	// Assignation of the Values from the peerstore to the Local Peer Metrics
	c.PeersMetrics[peerIndex].PeerId = peerData.PeerID.String()
	c.PeersMetrics[peerIndex].NodeId = peerData.NodeID.String()
	c.PeersMetrics[peerIndex].Pubkey = peerData.Pubkey
	copy(peerData.Addrs, c.PeersMetrics[peerIndex].Addres)
	c.PeersMetrics[peerIndex].Latency = peerData.Latency * time.Microsecond

	c.PeersMetrics[peerIndex].ClientType = peerData.ProtocolVersion

}

// -------------------- TODO ------------------------
//
//    func (GossipMetrics)get_protocol_version_from_peer(){}
//	  func (GossipMetrics)get_ip_address(){}
//	  func test_ping_message_for_checking_delays(){}
//  MAYBE: also try to see if I can get any API to ask the location of the IPaddres
//
// --------------------------------------------------

// Increments the counter of the topic
func (c *MessageMetrics) IncrementCnt() int64 {
	c.Cnt++
	return c.Cnt
}

// Stamps linux_time(millis) on the FirstMessageTime/LastMessageTime from given args: time (int64), flag string("first"/"last")
func (c *MessageMetrics) StampTime(flag string) {
	unixMillis := GetTimeMiliseconds()

	switch flag {
	case "first":
		c.FirstMessageTime = unixMillis
	case "last":
		c.LastMessageTime = unixMillis
	default:
		fmt.Println("Metrics Package -> StampTime.flag wrongly parsed")
	}
}

func GetTimeMiliseconds() int64 {
	now := time.Now()
	//secs := now.Unix()
	nanos := now.UnixNano()
	millis := nanos / 1000000

	return millis
}

// Function that Manages the metrics updates for the incoming messages
func IncomingMessageManager(c *GossipMetrics, peerId string, topicName string) error {
	// Load and delete
	peerIndex, err := c.GetPeerIndex(peerId)
	if err != nil {
		c.AddNewPeer(peerId)
		peerIndex, err = c.GetPeerIndex(peerId)
	}
	if err != nil {
		return errors.New("Something goes wrong with: AddNewPeer - GetPeerIndex")
	}

	peerMetrics := &c.PeersMetrics[peerIndex]
	messageMetrics, err := GetMessageMetrics(peerMetrics, topicName)
	if err != nil {
		return errors.New("Topic Name no supported")
	}

	if messageMetrics.Cnt == 0 {
		messageMetrics.StampTime("first")
	}

	messageMetrics.IncrementCnt()
	messageMetrics.StampTime("last")

	return nil
}

func GetMessageMetrics(c *PeerMetrics, topicName string) (mesMetr *MessageMetrics, err error) {
	// All this could be inside a different function
	switch topicName {
	case pgossip.BeaconBlock:
		return &c.BeaconBlock, nil
	case pgossip.BeaconAggregateProof:
		return &c.BeaconAggregateProof, nil
	case pgossip.VoluntaryExit:
		return &c.VoluntaryExit, nil
	case pgossip.ProposerSlashing:
		return &c.ProposerSlashing, nil
	case pgossip.AttesterSlashing:
		return &c.AttesterSlashing, nil
	default:
		return &c.BeaconBlock, err
	}
}
