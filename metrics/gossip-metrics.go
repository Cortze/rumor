package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
    "strings"
    "net/http"

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
	GossipMetrics sync.Map
}

// Base Struct for the topic name and the received messages on the different topics
type PeerMetrics struct {
	PeerId     string
	NodeId     string
	ClientType string
	Pubkey     string
	Addrs     []string
	Ip         string
    Country    string
    City       string
    Latency    float64

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

// Function that Wraps/Marshals the content of the sync.Map to be exported as a json
func (c *GossipState) MarshalMetrics() ([]byte, error) {
    tmpMap := make(map[string]PeerMetrics)
    c.GossipMetrics.Range(func(k, v interface{}) bool {
        tmpMap[k.(string)] = v.(PeerMetrics)
        return true
    })
    return json.Marshal(tmpMap)
}

// Function that Wraps/Marshals the content of the Entire Peerstore into a json
func (c *GossipState) MarshalPeerStore(ep track.ExtendedPeerstore) ([]byte, error) {
    var peers []peer.ID
    peers = ep.Peers()
    peerData := make(map[string]*track.PeerAllData)
    for _, p := range peers {
        peerData[p.String()] = ep.GetAllData(p)
    }
    return json.Marshal(peerData)
}


// Function that Exports the entire Metrics to a .json file (lets see if in the future we can add websockets or other implementations)
func (c *GossipState) ExportMetrics(filePath string, peerstorePath string, ep track.ExtendedPeerstore) error {
    metrics, err := c.MarshalMetrics()
	if err != nil {
		fmt.Println("Error Marshalling the metrics")
	}
    fmt.Println("Loading Peerstore to Export them")
    peerstore, err := c.MarshalPeerStore(ep)
	if err != nil {
		fmt.Println("Error Marshalling the peerstore")
	}

    err = ioutil.WriteFile(filePath, metrics, 0644)
	if err != nil {
		fmt.Println("Error opening file: ", filePath)
		return err
	}
	err = ioutil.WriteFile(peerstorePath, peerstore, 0644)
	if err != nil {
		fmt.Println("Error opening file: ", peerstorePath)
		return err
	}
	return nil
}

// IP-API message structure
type IpApiMessage struct {
    Query       string
    Status      string
    Continent   string
    ContinentCode string
    Country     string
    CountryCode string
    Region      string
    RegionName  string
    City    string
    District    string
    Zip     string
    Lat         string
    Lon         string
    Timezone    string
    Offset      string
    Currency    string
    Isp     string
    Org     string
    As      string
    Asname  string
    Mobile  string
    Proxy   string
    Hosting string
}

// get IP, location country and City from the multiaddress of the peer on the peerstore
func getIpAndLocationFromAddrs(multiAddrs string) (ip string, country string, city string) {
    ip = strings.TrimPrefix(multiAddrs, "/ip4/")
    ipSlices := strings.Split(ip, "/")
    ip = ipSlices[0]
    url := "http://ip-api.com/json/" + ip
    resp, err := http.Get(url)
    if err != nil {
        fmt.Println("There was an error getting the Location of the peer from the IP-API, please, check that there 40requests/minute limit hasn't been exceed")
    }
    defer resp.Body.Close()
    bodyBytes, _ := ioutil.ReadAll(resp.Body)

    // Convert response body to Todo struct
    var ipApiResp IpApiMessage
    json.Unmarshal(bodyBytes, &ipApiResp)

    country = ipApiResp.Country
    city = ipApiResp.City
    // return the received values from the received message
    return ip, country, city

}

// Add new peer with all the information from the peerstore to the metrics peerstore
func (c *GossipState) AddNewPeer(peerId peer.ID, ep track.ExtendedPeerstore) {	// Check if the peer already exists
	_ , ok := c.GossipMetrics.Load(peerId.String())
    if !ok {
        peerData := ep.GetAllData(peerId)
        ip, country, city := getIpAndLocationFromAddrs(peerData.Addrs[0])
        peerMetrics := PeerMetrics {
            PeerId: peerId.String(),
            NodeId: peerData.NodeID.String(),
            ClientType: peerData.UserAgent,
            Pubkey: peerData.Pubkey,
            Addrs: peerData.Addrs,
            Ip: ip,
            Country: country,
            City: city,
            Latency: float64(peerData.Latency / 1*time.Millisecond),
        }

        // Temp
        fmt.Println(peerData.Latency)
        fmt.Println(peerMetrics.Latency)
		// Temp

        // Include the new PeerMetrics struct on the syn.Map
        c.GossipMetrics.Store(peerId.String(), peerMetrics)
    }
}

// Add a connection Event to the given peer
func (c *GossipState) AddConnectionEvent(peerId peer.ID, connectionType string) {
	newConnection := ConnectionEvents{
		ConnectionType: connectionType,
		TimeMili:       GetTimeMiliseconds(),
	}
	pMetrics , ok := c.GossipMetrics.Load(peerId.String())
    if ok {
        peerMetrics := pMetrics.(PeerMetrics)
        peerMetrics.ConnectionEvents = append(peerMetrics.ConnectionEvents, newConnection)
        c.GossipMetrics.Store(peerId.String(), peerMetrics)
    } else {
        // Might be possible to add 
        errors.New("Counld't add Event, Peer is not in the list")
	}
}


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
func (c *GossipState) IncomingMessageManager(peerId peer.ID, topicName string) error {
    pMetrics , _ := c.GossipMetrics.Load(peerId.String())
    peerMetrics := pMetrics.(PeerMetrics)
	messageMetrics, err := GetMessageMetrics(&peerMetrics, topicName)
	if err != nil {
		return errors.New("Topic Name no supported")
	}
	fmt.Println(err, topicName, messageMetrics)
	if messageMetrics.Cnt == 0 {
		messageMetrics.StampTime("first")
	}

	messageMetrics.IncrementCnt()
	messageMetrics.StampTime("last")

    // Store back the Loaded/Modified Variable
    c.GossipMetrics.Store(peerId.String(), peerMetrics)

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
	default: //TODO: - Not returning BeaconBlock as Default
		return &c.BeaconBlock, err
	}
}
